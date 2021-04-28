package graphite

import (
	"errors"
	"fmt"
	"sort"

	"github.com/benoitkugler/textlayout/fonts/binaryreader"
)

// sorted by id
type TableFeat []feature

type feature struct {
	settings []featureSetting
	id       uint32
	flags    uint16
	label    uint16
}

type featureSetting struct {
	Value int16
	Label uint16
}

// FeatureValued is the result of choosing a feature
// and applying a value to it.
type FeatureValued struct {
	Id    Tag
	Flags uint16
	Value int16
	// Label uint16
}

// return the feature with their first setting selected (or 0)
func (tf TableFeat) defaultFeatures() []FeatureValued {
	out := make([]FeatureValued, len(tf))
	for i, f := range tf {
		out[i].Id = zeroToSpace(f.id)
		out[i].Flags = f.flags
		if len(f.settings) != 0 {
			out[i].Value = f.settings[0].Value
		}
	}
	return out
}

func (tf TableFeat) findFeature(id Tag) (feature, bool) {
	// binary search
	for i, j := 0, len(tf); i < j; {
		h := i + (j-i)/2
		entry := tf[h]
		if id < entry.id {
			j = h
		} else if entry.id < id {
			i = h + 1
		} else {
			return entry, true
		}
	}
	return feature{}, false
}

func parseTableFeat(data []byte) (TableFeat, error) {
	const headerSize = 12
	if len(data) < headerSize {
		return nil, errors.New("invalid Feat table (EOF)")
	}
	r := binaryreader.NewReader(data)
	version_, _ := r.Uint32()
	version := version_ >> 16
	numFeat, _ := r.Uint16()
	r.Skip(6)

	recordSize := 12
	if version >= 2 {
		recordSize = 16
	}
	featSlice, err := r.FixedSizes(int(numFeat), recordSize)
	if err != nil {
		return nil, fmt.Errorf("invalid Feat table: %s", err)
	}

	rFeat := binaryreader.NewReader(featSlice)
	out := make(TableFeat, numFeat)
	tmpIndexes := make([][2]int, numFeat)
	var maxSettingsLength int
	for i := range out {
		if version >= 2 {
			out[i].id, _ = rFeat.Uint32()
		} else {
			id_, _ := rFeat.Uint16()
			out[i].id = uint32(id_)
		}
		numSettings, _ := rFeat.Uint16()
		if version >= 2 {
			rFeat.Skip(2)
		}
		offset, _ := rFeat.Uint32()
		out[i].flags, _ = rFeat.Uint16()
		out[i].label, _ = rFeat.Uint16()

		// convert from offset to index
		index := (int(offset) - headerSize - len(featSlice)) / 4
		end := index + int(numSettings)
		if numSettings != 0 && end > maxSettingsLength {
			maxSettingsLength = end
		}

		tmpIndexes[i] = [2]int{index, int(numSettings)}
	}

	// parse the settings array
	allSettings := make([]featureSetting, maxSettingsLength)
	err = r.ReadStruct(allSettings)
	if err != nil {
		return nil, fmt.Errorf("invalid Feat table: %s", err)
	}

	for i, indexes := range tmpIndexes {
		index, length := indexes[0], indexes[1]
		if length == 0 {
			continue
		}
		out[i].settings = allSettings[index : index+length]
	}

	// sort by id
	sort.Slice(out, func(i, j int) bool { return out[i].id < out[j].id })

	return out, nil
}