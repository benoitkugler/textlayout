package truetype

import (
	"errors"
	"fmt"
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

func fixed1616ToFloat(fi uint32) float64 {
	// value are actually signed integers
	return float64(int32(fi)) / (1 << 16)
}

func fixed214ToFloat(fi uint16) float64 {
	// value are actually signed integers
	return float64(int16(fi)) / (1 << 14)
}

func parseTableFvar(table []byte, names TableName) (*TableFvar, error) {
	const headerSize = 8 * 2
	if len(table) < headerSize {
		return nil, errors.New("invalid 'fvar' table header")
	}
	// majorVersion := be.Uint16(table)
	// minorVersion := be.Uint16(table[2:])
	axesArrayOffset := be.Uint16(table[4:])
	// reserved := be.Uint16(table[6:])
	axisCount := be.Uint16(table[8:])
	axisSize := be.Uint16(table[10:])
	instanceCount := be.Uint16(table[12:])
	instanceSize := be.Uint16(table[14:])

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
	out.Tag = Tag(be.Uint32(axis))

	// convert from 16.16 to float64
	out.Minimum = fixed1616ToFloat(be.Uint32(axis[4:]))
	out.Default = fixed1616ToFloat(be.Uint32(axis[8:]))
	out.Maximum = fixed1616ToFloat(be.Uint32(axis[12:]))

	out.flags = be.Uint16(axis[16:])
	out.strid = NameID(be.Uint16(axis[18:]))

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
	out.Subfamily = NameID(be.Uint16(data))
	// _ = be.Uint16(data[2:]) reserved flags
	out.Coords = make([]float64, axisCount)
	for i := range out.Coords {
		out.Coords[i] = fixed1616ToFloat(be.Uint32(data[4+i*4:]))
	}
	// optional PostscriptName id
	if withPs {
		out.psid = NameID(be.Uint16(data[4+axisCount*4:]))
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
	from, to float64 // found as int16 fixed point 2.14
}

func parseTableAvar(data []byte) (*tableAvar, error) {
	const avarHeaderSize = 2 * 4
	if len(data) < avarHeaderSize {
		return nil, errors.New("invalid 'avar' table")
	}
	var table tableAvar
	table.majorVersion = be.Uint16(data)
	table.minorVersion = be.Uint16(data[2:])
	// reserved
	axisCount := be.Uint16(data[6:])
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
	count := be.Uint16(data)
	fmt.Println(count)
	size := int(count) * mapSize
	if len(data) < 2+size {
		return nil, nil, errors.New("invalid segment in 'avar' table")
	}
	out := make([]axisValueMap, count) // guarded by 16-bit constraint
	for i := range out {
		out[i].from = fixed214ToFloat(be.Uint16(data[2+i*mapSize:]))
		out[i].to = fixed214ToFloat(be.Uint16(data[2+i*mapSize+2:]))
	}
	data = data[2+size:]
	return out, data, nil
}
