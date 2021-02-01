package main

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// we do not support surrogates yet
const maxUnicode = 0x110000

func convertHexa(s string) uint32 {
	i, err := strconv.ParseUint(s, 16, 64)
	check(err)
	return uint32(i)
}

type runeRange struct {
	Start, End uint32
}

func (r runeRange) runes() []rune {
	var out []rune
	for ru := r.Start; ru <= r.End; ru++ {
		out = append(out, rune(ru))
	}
	return out
}

// split the file by line, ignore comments, and split each line by ';'
func splitLines(b []byte) (out [][]string) {
	for _, l := range bytes.Split(b, []byte{'\n'}) {
		line := string(bytes.TrimSpace(l))
		if line == "" || line[0] == '#' { // reading header or comment
			continue
		}
		out = append(out, strings.Split(line, ";"))
	}
	return
}

// filled by `parseUnicodeDatabase`
var (
	shapingTable struct {
		table    [maxUnicode][4]rune
		min, max rune
	}
	equivTable [maxUnicode]rune

	combiningClasses = map[uint8][]rune{} // class -> runes
)

// rune;comment;General_Category;Canonical_Combining_Class;Bidi_Class;Decomposition_Mapping;...;Bidi_Mirrored
func parseUnicodeDatabase(b []byte) error {
	var (
		min rune = maxUnicode
		max rune
	)

	for _, chunks := range splitLines(b) {

		if len(chunks) < 6 {
			continue
		}
		var (
			c        rune
			tag      string
			unshaped rune
		)

		// Rune
		_, err := fmt.Sscanf(chunks[0], "%04x", &c)
		if err != nil {
			return fmt.Errorf("invalid line %s: %s", chunks[0], err)
		}
		if c >= maxUnicode || unshaped >= maxUnicode {
			return fmt.Errorf("invalid rune value: %s", chunks[0])
		}

		// Combining class
		cc, err := strconv.Atoi(chunks[3])
		if err != nil {
			return fmt.Errorf("invalid combining class %s: %s", chunks[3], err)
		}
		if cc >= 256 {
			return fmt.Errorf("combining class too high %d", cc)
		}
		combiningClasses[uint8(cc)] = append(combiningClasses[uint8(cc)], c)

		// we are now looking for <...> XXXX
		if chunks[5] == "" {
			continue
		}
		fmt.Println(chunks[5])
		if chunks[5][0] == '<' {
			_, err = fmt.Sscanf(chunks[5], "%s %04x", &tag, &unshaped)
		} else {
			_, err = fmt.Sscanf(chunks[5], "%04x", &unshaped)
		}
		if err != nil {
			return fmt.Errorf("invalid shape %s: %s", chunks[5], err)
		}

		// shape table: only single unshaped rune are considered
		if shape := isShape(tag); shape >= 0 && len(chunks[5]) == len(tag)+5 {
			shapingTable.table[unshaped][shape] = c
			if unshaped < min {
				min = unshaped
			}
			if unshaped > max {
				max = unshaped
			}
		}

		// equiv table
		equivTable[c] = unshaped
	}
	return nil
}

func isShape(s string) int {
	for i, tag := range [...]string{
		"<isolated>",
		"<final>",
		"<initial>",
		"<medial>",
	} {
		if tag == s {
			return i
		}
	}
	return -1
}

func parseAnnexTables(b []byte) (map[string][]rune, error) {
	outRanges := map[string][]rune{}
	for _, parts := range splitLines(b) {
		if len(parts) < 2 {
			return nil, fmt.Errorf("invalid line: %s", parts)
		}
		rang, typ := strings.TrimSpace(parts[0]), strings.TrimSpace(strings.Split(parts[1], "#")[0])
		rangS := strings.Split(rang, "..")
		start := convertHexa(rangS[0])
		end := start
		if len(rangS) > 1 {
			end = convertHexa(rangS[1])
		}
		outRanges[typ] = append(outRanges[typ], runeRange{Start: start, End: end}.runes()...)
	}
	return outRanges, nil
}

func parseMirroring(b []byte) (map[uint16]uint16, error) {
	out := make(map[uint16]uint16)
	for _, parts := range splitLines(b) {
		if len(parts) < 2 {
			return nil, fmt.Errorf("invalid line: %s", parts)
		}
		start, end := strings.TrimSpace(parts[0]), strings.TrimSpace(strings.Split(parts[1], "#")[0])
		startRune, endRune := convertHexa(start), convertHexa(end)
		if startRune > 0xFFFF {
			return nil, fmt.Errorf("rune %d overflows implementation limit", startRune)
		}
		if endRune > 0xFFFF {
			return nil, fmt.Errorf("rune %d overflows implementation limit", endRune)
		}
		out[uint16(startRune)] = uint16(endRune)
	}
	return out, nil
}

type ucdXML struct {
	XMLName xml.Name `xml:"ucd"`
	Reps    []group  `xml:"repertoire>group"`
}

type group struct {
	Dm string `xml:"dm,attr"`
	Dt string `xml:"dt,attr"`
	// Chars    []char   `xml:"char"`
	Reserved []string `xml:"reserved"`
}

type char struct {
	Cp      string `xml:"cp,attr"`
	FirstCp string `xml:"first-cp,attr"`
	LastCp  string `xml:"last-cp,attr"`
}

func parseXML(filename string) {
	f, err := zip.OpenReader(filename)
	check(err)
	if len(f.File) != 1 {
		check(errors.New("invalid zip file"))
	}
	content, err := f.File[0].Open()
	check(err)

	var out ucdXML
	dec := xml.NewDecoder(content)
	err = dec.Decode(&out)
	check(err)
	fmt.Println(out)
}
