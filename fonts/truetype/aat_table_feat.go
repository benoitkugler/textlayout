package truetype

import (
	"encoding/binary"
	"errors"
)

type TableFeat []AATFeatureName

// GetFeature performs a binary seach into the names, using `Feature` as key,
// returning `nil` if not found.
func (t TableFeat) GetFeature(feature uint16) *AATFeatureName {
	for i, j := 0, len(t); i < j; {
		h := i + (j-i)/2
		entry := t[h].Feature
		if feature < entry {
			j = h
		} else if entry < feature {
			i = h + 1
		} else {
			return &t[h]
		}
	}
	return nil
}

func parseTableFeat(data []byte) (TableFeat, error) {
	if len(data) < 12 {
		return nil, errors.New("invalid feat table (EOF)")
	}
	featureNameCount := binary.BigEndian.Uint16(data[4:])
	if len(data) < 12+12*int(featureNameCount) {
		return nil, errors.New("invalid feat table (EOF)")
	}
	out := make(TableFeat, featureNameCount)
	var err error
	for i := range out {
		out[i].Feature = binary.BigEndian.Uint16(data[12+12*i:])
		nSettings := binary.BigEndian.Uint16(data[12+12*i+2:])
		offsetSetting := binary.BigEndian.Uint32(data[12+12*i+4:])
		out[i].Flags = binary.BigEndian.Uint16(data[12+12*i+8:])
		out[i].NameIndex = NameID(binary.BigEndian.Uint16(data[12+12*i+10:]))
		out[i].Settings, err = parseAATSettingNames(data, offsetSetting, nSettings)
		if err != nil {
			return nil, err
		}
	}

	return out, nil
}

type AATFeatureName struct {
	Feature   uint16
	Flags     uint16
	NameIndex NameID
	Settings  []AATSettingName
}

// IsExclusive returns true if the feature settings are mutually exclusive.
func (feature *AATFeatureName) IsExclusive() bool {
	const Exclusive = 0x8000
	return feature.Flags&Exclusive != 0
}

type AATSettingName struct {
	Setting uint16
	Name    NameID
}

func parseAATSettingNames(data []byte, offset uint32, count uint16) ([]AATSettingName, error) {
	if len(data) < int(offset)+4*int(count) {
		return nil, errors.New("invalid feat table settings names (EOF)")
	}

	out := make([]AATSettingName, count)
	data = data[offset:]
	for i := range out {
		out[i].Setting = binary.BigEndian.Uint16(data[4*i:])
		out[i].Name = NameID(binary.BigEndian.Uint16(data[4*i+2:]))
	}
	return out, nil
}
