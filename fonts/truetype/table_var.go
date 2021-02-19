package truetype

import (
	"encoding/binary"
	"errors"
	"fmt"

	"github.com/benoitkugler/textlayout/fonts"
)

type fvarHeader struct {
	majorVersion    uint16 // Major version number of the font variations table — set to 1.
	minorVersion    uint16 // Minor version number of the font variations table — set to 0.
	axesArrayOffset uint16 // Offset in bytes from the beginning of the table to the start of the VariationAxisRecord array.
	reserved        uint16 // This field is permanently reserved. Set to 2.
	axisCount       uint16 // The number of variation axes in the font (the number of records in the axes array).
	axisSize        uint16 // The size in bytes of each VariationAxisRecord — set to 20 (0x0014) for this version.
	instanceCount   uint16 // The number of named instances defined in the font (the number of records in the instances array).
	instanceSize    uint16 // The size in bytes of each InstanceRecord — set to either axisCount * sizeof(Fixed) + 4, or to axisCount * sizeof(Fixed) + 6.
}

func fixed1616ToFloat(fi uint32) float32 {
	// value are actually signed integers
	return float32(int32(fi)) / (1 << 16)
}

func fixed214ToFloat(fi uint16) float32 {
	// value are actually signed integers
	return float32(int16(fi)) / (1 << 14)
}

func parseTableFvar(table []byte, names TableName) (*TableFvar, error) {
	const headerSize = 8 * 2
	if len(table) < headerSize {
		return nil, errors.New("invalid 'fvar' table header")
	}
	// majorVersion := binary.BigEndian.Uint16(table)
	// minorVersion := binary.BigEndian.Uint16(table[2:])
	axesArrayOffset := binary.BigEndian.Uint16(table[4:])
	// reserved := binary.BigEndian.Uint16(table[6:])
	axisCount := binary.BigEndian.Uint16(table[8:])
	axisSize := binary.BigEndian.Uint16(table[10:])
	instanceCount := binary.BigEndian.Uint16(table[12:])
	instanceSize := binary.BigEndian.Uint16(table[14:])

	axis, instanceOffset, err := parseVarAxis(table, int(axesArrayOffset), int(axisSize), axisCount)
	if err != nil {
		return nil, err
	}
	// the instance offset is at the end of the axis
	instances, err := parseVarInstance(table, instanceOffset, int(instanceSize), instanceCount, axisCount)
	if err != nil {
		return nil, err
	}

	out := TableFvar{Axis: axis, Instances: instances}
	out.checkDefaultInstance(names)
	return &out, nil
}

func parseVarAxis(table []byte, offset, size int, count uint16) ([]VarAxis, int, error) {
	// we need at least 20 byte per axis ....
	if size < 20 {
		return nil, 0, errors.New("invalid 'fvar' table axis")
	}
	// ...but "implementations must use the axisSize and instanceSize fields
	// to determine the start of each record".
	end := offset + int(count)*size
	if len(table) < end {
		return nil, 0, errors.New("invalid 'fvar' table axis")
	}

	out := make([]VarAxis, count) // limited by 16 bit type
	for i := range out {
		out[i] = parseOneVarAxis(table[offset+i*size:])
	}

	return out, end, nil
}

// do not check the size of data
func parseOneVarAxis(axis []byte) VarAxis {
	var out VarAxis
	out.Tag = Tag(binary.BigEndian.Uint32(axis))

	// convert from 16.16 to float64
	out.Minimum = fixed1616ToFloat(binary.BigEndian.Uint32(axis[4:]))
	out.Default = fixed1616ToFloat(binary.BigEndian.Uint32(axis[8:]))
	out.Maximum = fixed1616ToFloat(binary.BigEndian.Uint32(axis[12:]))

	out.flags = binary.BigEndian.Uint16(axis[16:])
	out.strid = NameID(binary.BigEndian.Uint16(axis[18:]))

	return out
}

func parseVarInstance(table []byte, offset, size int, count, axisCount uint16) ([]VarInstance, error) {
	// we need at least 4+4*axisCount byte per instance ....
	if size < 4+4*int(axisCount) {
		return nil, errors.New("invalid 'fvar' table instance")
	}
	withPs := size >= 4+4*int(axisCount)+2

	// ...but "implementations must use the axisSize and instanceSize fields
	// to determine the start of each record".
	if len(table) < offset+int(count)*size {
		return nil, errors.New("invalid 'fvar' table axis")
	}

	out := make([]VarInstance, count) // limited by 16 bit type
	for i := range out {
		out[i] = parseOneVarInstance(table[offset+i*size:], axisCount, withPs)
	}

	return out, nil
}

// do not check the size of data
func parseOneVarInstance(data []byte, axisCount uint16, withPs bool) VarInstance {
	var out VarInstance
	out.Subfamily = NameID(binary.BigEndian.Uint16(data))
	// _ = binary.BigEndian.Uint16(data[2:]) reserved flags
	out.Coords = make([]float32, axisCount)
	for i := range out.Coords {
		out.Coords[i] = fixed1616ToFloat(binary.BigEndian.Uint32(data[4+i*4:]))
	}
	// optional PostscriptName id
	if withPs {
		out.PSStringID = NameID(binary.BigEndian.Uint16(data[4+axisCount*4:]))
	}

	return out
}

// -------------------------- avar table --------------------------

type tableAvar struct {
	majorVersion uint16 // Major version number of the axis variations table — set to 1.
	minorVersion uint16 // Minor version number of the axis variations table — set to 0.
	// <reserved> uint16	// Permanently reserved; set to zero.
	// axisCount       uint16           // The number of variation axes for this font. This must be the same number as axisCount in the 'fvar' table.
	axisSegmentMaps [][]axisValueMap //	The segment maps array — one segment map for each axis, in the order of axes specified in the 'fvar' table.
}

type axisValueMap struct {
	from, to float32 // found as int16 fixed point 2.14
}

func parseTableAvar(data []byte) (*tableAvar, error) {
	const avarHeaderSize = 2 * 4
	if len(data) < avarHeaderSize {
		return nil, errors.New("invalid 'avar' table")
	}
	var table tableAvar
	table.majorVersion = binary.BigEndian.Uint16(data)
	table.minorVersion = binary.BigEndian.Uint16(data[2:])
	// reserved
	axisCount := binary.BigEndian.Uint16(data[6:])
	table.axisSegmentMaps = make([][]axisValueMap, axisCount) // guarded by 16-bit constraint

	var err error
	data = data[avarHeaderSize:] // start at the first segment list
	for i := range table.axisSegmentMaps {
		table.axisSegmentMaps[i], data, err = parseSegmentList(data)
		if err != nil {
			return nil, err
		}
	}
	return &table, nil
}

// data is at the start of the segment, return value at the start of the next
func parseSegmentList(data []byte) ([]axisValueMap, []byte, error) {
	const mapSize = 4
	if len(data) < 2 {
		return nil, nil, errors.New("invalid segment in 'avar' table")
	}
	count := binary.BigEndian.Uint16(data)
	fmt.Println(count)
	size := int(count) * mapSize
	if len(data) < 2+size {
		return nil, nil, errors.New("invalid segment in 'avar' table")
	}
	out := make([]axisValueMap, count) // guarded by 16-bit constraint
	for i := range out {
		out[i].from = fixed214ToFloat(binary.BigEndian.Uint16(data[2+i*mapSize:]))
		out[i].to = fixed214ToFloat(binary.BigEndian.Uint16(data[2+i*mapSize+2:]))
	}
	data = data[2+size:]
	return out, data, nil
}

// VariationStoreIndex reference an item in the variation store
type VariationStoreIndex struct {
	DeltaSetOuter, DeltaSetInner uint16
}

// TODO: sanitize array length using FVar
// After successful parsing, every region indexes in `Datas` elements are valid.
type VariationStore struct {
	Regions [][]VariationRegion // for each region, for each axis
	Datas   []ItemVariationData
}

// GetDelta uses the variation store and the selected instance coordinates
// to compute the value at `index`.
func (store VariationStore) GetDelta(index VariationStoreIndex, coords []float32) float32 {
	if int(index.DeltaSetOuter) >= len(store.Datas) {
		return 0
	}
	varData := store.Datas[index.DeltaSetOuter]
	if int(index.DeltaSetInner) >= len(varData.Deltas) {
		return 0
	}
	deltaSet := varData.Deltas[index.DeltaSetInner]
	var delta float32
	for i, regionIndex := range varData.RegionIndexes {
		region := store.Regions[regionIndex]
		v := float32(1)
		for axis, coord := range coords {
			factor := region[axis].evaluate(coord)
			v *= factor
		}
		delta += float32(deltaSet[i]) * v
	}
	return delta
}

func parseVariationStore(data []byte, offset uint32) (out VariationStore, err error) {
	if len(data) < int(offset)+8 {
		return out, errors.New("invalid item variation store (EOF)")
	}
	data = data[offset:]
	// format is ignored
	regionsOffset := binary.BigEndian.Uint32(data[2:])
	count := binary.BigEndian.Uint16(data[6:])

	out.Regions, err = parseItemVariationRegions(data, regionsOffset)
	if err != nil {
		return out, err
	}

	if len(data) < 8+4*int(count) {
		return out, errors.New("invalid item variation store (EOF)")
	}
	out.Datas = make([]ItemVariationData, count)
	for i := range out.Datas {
		subtableOffset := binary.BigEndian.Uint32(data[8+4*i:])
		out.Datas[i], err = parseItemVariationData(data, subtableOffset, uint16(len(out.Regions)))
		if err != nil {
			return out, err
		}
	}
	return out, nil
}

func parseItemVariationRegions(data []byte, offset uint32) ([][]VariationRegion, error) {
	if len(data) < int(offset)+4 {
		return nil, errors.New("invalid item variation regions list (EOF)")
	}
	data = data[offset:]
	axisCount := int(binary.BigEndian.Uint16(data))
	regionCount := int(binary.BigEndian.Uint16(data[2:]))

	if len(data) < 4+6*axisCount*regionCount {
		return nil, errors.New("invalid item variation regions list (EOF)")
	}
	regions := make([][]VariationRegion, regionCount)
	for i := range regions {
		ri := make([]VariationRegion, axisCount)
		for j := range ri {
			start := fixed214ToFloat(binary.BigEndian.Uint16(data[4+(i*axisCount+j)*6:]))
			peak := fixed214ToFloat(binary.BigEndian.Uint16(data[4+(i*axisCount+j)*6+2:]))
			end := fixed214ToFloat(binary.BigEndian.Uint16(data[4+(i*axisCount+j)*6+4:]))

			if start > peak || peak > end {
				return nil, errors.New("invalid item variation regions list")
			}
			if start < 0 && end > 0 && peak != 0 {
				return nil, errors.New("invalid item variation regions list")
			}
			ri[j] = VariationRegion{start, peak, end}
		}
		regions[i] = ri
	}
	return regions, nil
}

type ItemVariationData struct {
	RegionIndexes []uint16  // Array of indices into the variation region list for the regions referenced by this item variation data table.
	Deltas        [][]int16 // Each row as the same length as `RegionIndexes`
}

func parseItemVariationData(data []byte, offset uint32, nbRegions uint16) (out ItemVariationData, err error) {
	if len(data) < int(offset)+6 {
		return out, errors.New("invalid item variation data subtable (EOF)")
	}
	data = data[offset:]
	itemCount := int(binary.BigEndian.Uint16(data))
	shortDeltaCount := int(binary.BigEndian.Uint16(data[2:]))
	regionIndexCount := int(binary.BigEndian.Uint16(data[4:]))

	out.RegionIndexes, err = parseUint16s(data[6:], regionIndexCount)
	if err != nil {
		return out, fmt.Errorf("invalid item variation data subtable: %s", err)
	}
	// sanitize the indexes
	for _, regionIndex := range out.RegionIndexes {
		if regionIndex >= nbRegions {
			return out, fmt.Errorf("invalid item variation region index: %d (for size %d)", regionIndex, nbRegions)
		}
	}

	data = data[6+2*regionIndexCount:] // length checked by the previous `parseUint16s` call
	rowLength := shortDeltaCount + regionIndexCount
	if len(data) < itemCount*rowLength {
		return out, errors.New("invalid item variation data subtable (EOF)")
	}
	if shortDeltaCount > regionIndexCount {
		return out, errors.New("invalid item variation data subtable")
	}
	out.Deltas = make([][]int16, itemCount)
	for i := range out.Deltas {
		vi := make([]int16, regionIndexCount)
		j := 0
		for ; j < shortDeltaCount; j++ {
			vi[j] = int16(binary.BigEndian.Uint16(data[2*j:]))
		}
		for ; j < regionIndexCount; j++ {
			vi[j] = int16(int8(data[shortDeltaCount+j]))
		}
		out.Deltas[i] = vi
		data = data[rowLength:]
	}
	return out, nil
}

// ---------------------------------- mvar table ----------------------------------

type TableMvar struct {
	Values []VarValueRecord // sorted by tag
	Store  VariationStore
}

// return 0 if `tag` is not found
func (t TableMvar) getVar(tag Tag, coords []float32) float32 {
	// binary search
	for i, j := 0, len(t.Values); i < j; {
		h := i + (j-i)/2
		entry := t.Values[h]
		if tag < entry.Tag {
			j = h
		} else if entry.Tag < tag {
			i = h + 1
		} else {
			return t.Store.GetDelta(entry.Index, coords)
		}
	}
	return 0
}

type VarValueRecord struct {
	Tag   Tag
	Index VariationStoreIndex
}

func parseTableMvar(data []byte) (out TableMvar, err error) {
	if len(data) < 12 {
		return out, errors.New("invalid 'mvar' table (EOF)")
	}
	recordSize := int(binary.BigEndian.Uint16(data[6:]))
	recordCount := binary.BigEndian.Uint16(data[8:])
	storeOffset := uint32(binary.BigEndian.Uint16(data[10:]))

	if recordSize < 8 {
		return out, fmt.Errorf("invalid 'mvar' table record size: %d", recordSize)
	}

	out.Store, err = parseVariationStore(data, storeOffset)
	if err != nil {
		return out, err
	}

	if len(data) < 12+recordSize*int(recordCount) {
		return out, errors.New("invalid 'mvar' table (EOF)")
	}
	out.Values = make([]VarValueRecord, recordCount)
	for i := range out.Values {
		out.Values[i].Tag = Tag(binary.BigEndian.Uint32(data[12+recordSize*i:]))
		out.Values[i].Index.DeltaSetOuter = binary.BigEndian.Uint16(data[12+recordSize*i+4:])
		out.Values[i].Index.DeltaSetInner = binary.BigEndian.Uint16(data[12+recordSize*i+6:])
	}

	return out, nil
}

// ---------------------------------- HVAR/VVAR ----------------------------------

type tableHVvar struct {
	store VariationStore
	// optional
	advanceMapping deltaSetMapping
}

func (t tableHVvar) getAdvanceVar(glyph fonts.GlyphIndex, coords []float32) float32 {
	index := t.advanceMapping.getIndex(glyph)
	return t.store.GetDelta(index, coords)
}

func parseTableHVvar(data []byte) (out tableHVvar, err error) {
	if len(data) < 20 {
		return out, errors.New("invalid metrics variation table (EOF)")
	}
	storeOffset := binary.BigEndian.Uint32(data[4:])
	advanceOffset := binary.BigEndian.Uint32(data[8:])
	out.store, err = parseVariationStore(data, storeOffset)
	if err != nil {
		return out, err
	}
	if advanceOffset != 0 {
		out.advanceMapping, err = parseDeltaSetMapping(data, advanceOffset)
		if err != nil {
			return out, err
		}
	}

	return out, nil
}

// may have a length < numGlyph
type deltaSetMapping []VariationStoreIndex

func (m deltaSetMapping) getIndex(glyph fonts.GlyphIndex) VariationStoreIndex {
	// If a mapping table is not provided, glyph indices are used as implicit delta-set indices.
	// [...] the delta-set outer-level index is zero, and the glyph ID is used as the inner-level index.
	if len(m) == 0 {
		return VariationStoreIndex{DeltaSetInner: uint16(glyph)}
	}

	// If a given glyph ID is greater than mapCount - 1, then the last entry is used.
	if int(glyph) >= len(m) {
		glyph = fonts.GlyphIndex(len(m) - 1)
	}

	return m[glyph]
}

func parseDeltaSetMapping(data []byte, offset uint32) (deltaSetMapping, error) {
	if len(data) < int(offset)+4 {
		return nil, errors.New("invalid delta-set mapping (EOF)")
	}
	format := binary.BigEndian.Uint16(data[offset:])
	count := int(binary.BigEndian.Uint16(data[offset+2:]))
	data = data[offset+4:]

	entrySize := int((format&0x0030)>>4 + 1)
	innerBitSize := format&0x0F + 1
	if entrySize > 4 || len(data) < entrySize*count {
		return nil, errors.New("invalid delta-set mapping (EOF)")
	}
	out := make(deltaSetMapping, count)
	for i := range out {
		var v uint32
		for _, b := range data[entrySize*i : entrySize*(i+1)] { // 1 to 4 bytes
			v = v<<8 + uint32(b)
		}
		out[i].DeltaSetOuter = uint16(v >> innerBitSize)
		out[i].DeltaSetInner = uint16(v & (1<<innerBitSize - 1))
	}

	return out, nil
}
