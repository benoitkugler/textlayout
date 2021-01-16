package truetype

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
)

// TableLayout represents the common layout table used by GPOS and GSUB.
// The Features field contains all the features for this layout. However,
// the script and language determines which feature is used.
//
// See https://www.microsoft.com/typography/otspec/chapter2.htm#organization
// See https://www.microsoft.com/typography/otspec/GPOS.htm
// See https://www.microsoft.com/typography/otspec/GSUB.htm
type TableLayout struct {
	version versionHeader
	header  layoutHeader11

	Scripts  []Script  // Scripts contains all the scripts in this layout.
	Features []Feature // Features contains all the features in this layout.
	Lookups  []Lookup  // Lookups contains all the lookups in this layout.
}

// Script represents a single script (i.e "latn" (Latin), "cyrl" (Cyrillic), etc).
type Script struct {
	Tag             TableTag   // Tag for this script.
	DefaultLanguage *LangSys   // DefaultLanguage used by this script.
	Languages       []*LangSys // Languages within this script.
}

// LangSys represents the language system for a script.
type LangSys struct {
	Tag      TableTag  // Tag for this language.
	Features []Feature // Features contains the features for this language.
}

// Feature represents a glyph substitution or glyph positioning features.
type Feature struct {
	Tag TableTag // Tag for this feature
}

// Lookup represents a feature lookup table.
type Lookup struct {
	Type uint16 // Different enumerations for GSUB and GPOS.
	Flag uint16 // Lookup qualifiers.

	subtableOffsets []uint16 // Array of offsets to lookup subtables, from beginning of Lookup table
	data            []byte   // input data of the lookup table
	// markFilteringSet uint16 // Index (base 0) into GDEF mark glyph sets structure. This field is only present if bit useMarkFilteringSet of lookup flags is set.
}

// versionHeader is the beginning of on-disk format of the GPOS/GSUB version header.
// See https://www.microsoft.com/typography/otspec/GPOS.htm
// See https://www.microsoft.com/typography/otspec/GSUB.htm
type versionHeader struct {
	Major uint16 // Major version of the GPOS/GSUB table.
	Minor uint16 // Minor version of the GPOS/GSUB table.
}

// layoutHeader10 is the on-disk format of GPOS/GSUB version header when major=1 and minor=0.
type layoutHeader10 struct {
	ScriptListOffset  uint16 // offset to ScriptList table, from beginning of GPOS/GSUB table.
	FeatureListOffset uint16 // offset to FeatureList table, from beginning of GPOS/GSUB table.
	LookupListOffset  uint16 // offset to LookupList table, from beginning of GPOS/GSUB table.
}

// layoutHeader11 is the on-disk format of GPOS/GSUB version header when major=1 and minor=1.
type layoutHeader11 struct {
	layoutHeader10
	FeatureVariationsOffset uint32 // offset to FeatureVariations table, from beginning of GPOS/GSUB table (may be NULL).
}

// tagOffsetRecord is a on-disk format of a Tag and Offset record, commonly used thoughout this table.
type tagOffsetRecord struct {
	Tag    TableTag // 4-byte script tag identifier
	Offset uint16   // Offset to object from beginning of list
}
type scriptRecord tagOffsetRecord
type featureRecord tagOffsetRecord
type lookupRecord tagOffsetRecord
type langSysRecord tagOffsetRecord

type scriptTable struct {
	DefaultLangSys uint16 // Offset to default LangSys table, from beginning of Script table — may be NULL
	LangSysCount   uint16 // Number of LangSysRecords for this script — excluding the default LangSys
	// langSysRecords[langSysCount] langSysRecord // Array of LangSysRecords, listed alphabetically by LangSys tag
}

type featureTable struct {
	FeatureParams    uint16 // = NULL (reserved for offset to FeatureParams)
	LookupIndexCount uint16 // Number of LookupList indices for this feature
	// lookupListIndices [lookupIndexCount]uint16 // Array of indices into the LookupList — zero-based (first lookup is LookupListIndex = 0)}
}

type langSysTable struct {
	LookupOrder          uint16 // = NULL (reserved for an offset to a reordering table)
	RequiredFeatureIndex uint16 // Index of a feature required for this language system; if no required features = 0xFFFF
	FeatureIndexCount    uint16 // Number of feature index values for this language system — excludes the required feature
	// featureIndices[featureIndexCount] uint16 // Array of indices into the FeatureList, in arbitrary order
}

// parseLangSys parses a single Language System table. b expected to be the beginning of Script table.
// See https://www.microsoft.com/typography/otspec/chapter2.htm#langSysTbl
func (t *TableLayout) parseLangSys(b []byte, record langSysRecord) (*LangSys, error) {
	if int(record.Offset) >= len(b) {
		return nil, io.ErrUnexpectedEOF
	}

	r := bytes.NewReader(b[record.Offset:])

	var lang langSysTable
	if err := binary.Read(r, binary.BigEndian, &lang); err != nil {
		return nil, fmt.Errorf("reading langSysTable: %s", err)
	}

	featureIndices := make([]uint16, lang.FeatureIndexCount, lang.FeatureIndexCount)
	if err := binary.Read(r, binary.BigEndian, &featureIndices); err != nil {
		return nil, fmt.Errorf("reading langSysTable featureIndices[%d]: %s", lang.FeatureIndexCount, err)
	}

	var features []Feature
	for i := 0; i < len(featureIndices); i++ {
		if int(featureIndices[i]) >= len(t.Features) {
			return nil, fmt.Errorf("invalid featureIndices[%d] = %d", i, featureIndices[i])
		}
		features = append(features, t.Features[featureIndices[i]])
	}

	return &LangSys{
		Tag:      record.Tag,
		Features: features,
	}, nil
}

// parseScript parses a single Script table. b expected to be the beginning of ScriptList.
// See https://www.microsoft.com/typography/otspec/chapter2.htm#sTbl_lsRec
func (t *TableLayout) parseScript(b []byte, record scriptRecord) (Script, error) {
	if int(record.Offset) >= len(b) {
		return Script{}, io.ErrUnexpectedEOF
	}

	b = b[record.Offset:]
	r := bytes.NewReader(b)

	var script scriptTable
	if err := binary.Read(r, binary.BigEndian, &script); err != nil {
		return Script{}, fmt.Errorf("reading scriptTable: %s", err)
	}

	var defaultLang *LangSys
	var langs []*LangSys

	if script.DefaultLangSys > 0 {
		var err error
		defaultLang, err = t.parseLangSys(b, langSysRecord{Offset: script.DefaultLangSys})
		if err != nil {
			return Script{}, err
		}
	}

	for i := 0; i < int(script.LangSysCount); i++ {
		var record langSysRecord
		if err := binary.Read(r, binary.BigEndian, &record); err != nil {
			return Script{}, fmt.Errorf("reading langSysRecord[%d]: %s", i, err)
		}

		if record.Offset == script.DefaultLangSys {
			// Don't process the same language twice
			continue
		}

		lang, err := t.parseLangSys(b, record)
		if err != nil {
			return Script{}, err
		}

		langs = append(langs, lang)
	}

	return Script{
		Tag:             record.Tag,
		DefaultLanguage: defaultLang,
		Languages:       langs,
	}, nil
}

// parseScriptList parses the ScriptList.
// See https://www.microsoft.com/typography/otspec/chapter2.htm#slTbl_sRec
func (t *TableLayout) parseScriptList(buf []byte) error {
	offset := int(t.header.ScriptListOffset)
	if offset >= len(buf) {
		return io.ErrUnexpectedEOF
	}

	b := buf[offset:]
	r := bytes.NewReader(b)

	var count uint16
	if err := binary.Read(r, binary.BigEndian, &count); err != nil {
		return fmt.Errorf("reading scriptCount: %s", err)
	}

	t.Scripts = make([]Script, count)
	for i := 0; i < int(count); i++ {
		var record scriptRecord
		if err := binary.Read(r, binary.BigEndian, &record); err != nil {
			return fmt.Errorf("reading scriptRecord[%d]: %s", i, err)
		}

		script, err := t.parseScript(b, record)
		if err != nil {
			return err
		}

		t.Scripts[i] = script
	}

	return nil
}

// parseFeature parses a single Feature table. b expected to be the beginning of FeatureList.
// See https://www.microsoft.com/typography/otspec/chapter2.htm#featTbl
func (t *TableLayout) parseFeature(b []byte, record featureRecord) (Feature, error) {
	if int(record.Offset) >= len(b) {
		return Feature{}, io.ErrUnexpectedEOF
	}

	r := bytes.NewReader(b[record.Offset:])

	var feature featureTable
	if err := binary.Read(r, binary.BigEndian, &feature); err != nil {
		return Feature{}, fmt.Errorf("reading featureTable: %s", err)
	}

	// TODO Read feature.FeatureParams and feature.LookupIndexCount

	return Feature{
		Tag: record.Tag,
	}, nil
}

// parseFeatureList parses the FeatureList.
// See https://www.microsoft.com/typography/otspec/chapter2.htm#flTbl
func (t *TableLayout) parseFeatureList(buf []byte) error {
	offset := int(t.header.FeatureListOffset)
	if offset >= len(buf) {
		return io.ErrUnexpectedEOF
	}

	b := buf[offset:]
	r := bytes.NewReader(b)

	var count uint16
	if err := binary.Read(r, binary.BigEndian, &count); err != nil {
		return fmt.Errorf("reading featureCount: %s", err)
	}

	t.Features = make([]Feature, count)
	for i := 0; i < int(count); i++ {
		var record featureRecord
		if err := binary.Read(r, binary.BigEndian, &record); err != nil {
			return fmt.Errorf("reading featureRecord[%d]: %s", i, err)
		}

		feature, err := t.parseFeature(b, record)
		if err != nil {
			return err
		}

		t.Features[i] = feature
	}

	return nil
}

// parseLookup parses a single Lookup table. b expected to be the beginning of LookupList.
// See https://www.microsoft.com/typography/otspec/chapter2.htm#featTbl
func (t *TableLayout) parseLookup(b []byte, lookupTableOffset uint16) (Lookup, error) {
	if int(lookupTableOffset) >= len(b) {
		return Lookup{}, io.ErrUnexpectedEOF
	}

	b = b[lookupTableOffset:]
	const tableHeaderSize = 6
	if len(b) < tableHeaderSize {
		return Lookup{}, io.ErrUnexpectedEOF
	}

	type_ := be.Uint16(b)
	flag := be.Uint16(b[2:])
	subTableCount := be.Uint16(b[4:])

	if len(b) < tableHeaderSize+2*int(subTableCount) {
		return Lookup{}, io.ErrUnexpectedEOF
	}

	subtableOffsets := make([]uint16, subTableCount)
	for i := range subtableOffsets {
		subtableOffsets[i] = be.Uint16(b[tableHeaderSize+2*i:])
	}

	// TODO Read lookup.MarkFilteringSet

	return Lookup{
		Type:            type_,
		Flag:            flag, // TODO Parse the type Enum
		subtableOffsets: subtableOffsets,
		data:            b,
	}, nil
}

// parseLookupList parses the LookupList.
// See https://www.microsoft.com/typography/otspec/chapter2.htm#lulTbl
func (t *TableLayout) parseLookupList(buf []byte) error {
	offset := int(t.header.LookupListOffset)
	if offset >= len(buf) {
		return io.ErrUnexpectedEOF
	}

	b := buf[offset:]
	r := bytes.NewReader(b)

	var count uint16
	if err := binary.Read(r, binary.BigEndian, &count); err != nil {
		return fmt.Errorf("reading lookupCount: %s", err)
	}

	t.Lookups = make([]Lookup, count)
	for i := 0; i < int(count); i++ {
		var lookupTableOffset uint16
		if err := binary.Read(r, binary.BigEndian, &lookupTableOffset); err != nil {
			return fmt.Errorf("reading lookupRecord[%d]: %s", i, err)
		}

		lookup, err := t.parseLookup(b, lookupTableOffset)
		if err != nil {
			return err
		}

		t.Lookups[i] = lookup
	}

	return nil
}

// parseTableLayout parses a common Layout Table used by GPOS and GSUB.
func parseTableLayout(buf []byte) (*TableLayout, error) {
	t := &TableLayout{}

	r := bytes.NewReader(buf)
	if err := binary.Read(r, binary.BigEndian, &t.version); err != nil {
		return nil, fmt.Errorf("reading layout version header: %s", err)
	}

	if t.version.Major != 1 || (t.version.Minor != 0 && t.version.Minor != 1) {
		return nil, fmt.Errorf("unsupported layout version (major: %d, minor: %d)", t.version.Major, t.version.Minor)
	}

	switch t.version.Minor {
	case 0:
		if err := binary.Read(r, binary.BigEndian, &t.header.layoutHeader10); err != nil {
			return nil, fmt.Errorf("reading layout header: %s", err)
		}
	case 1:
		if err := binary.Read(r, binary.BigEndian, &t.header); err != nil {
			return nil, fmt.Errorf("reading layout header: %s", err)
		}
	default:
		// Should never get here, because we are gated by a earlier check.
		panic("unsupported minor version")
	}

	if err := t.parseLookupList(buf); err != nil {
		return nil, err
	}

	if err := t.parseFeatureList(buf); err != nil {
		return nil, err
	}

	if err := t.parseScriptList(buf); err != nil {
		return nil, err
	}

	return t, nil
}
