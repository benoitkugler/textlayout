package main

import (
	"archive/zip"
	"bytes"
	"encoding/binary"
	"encoding/xml"
	"errors"
	"fmt"
	"strconv"
	"strings"

	ucd "github.com/benoitkugler/textlayout/unicodedata"
)

// we do not support surrogates yet
const maxUnicode = 0x110000

func parseRune(s string) rune {
	i, err := strconv.ParseUint(s, 16, 64)
	check(err)
	return rune(i)
}

// parse a space separated list of runes
func parseRunes(s string) []rune {
	var out []rune
	for _, s := range strings.Fields(s) {
		out = append(out, parseRune(s))
	}
	return out
}

type runeRange struct {
	Start, End rune
}

func (r runeRange) runes() []rune {
	var out []rune
	for ru := r.Start; ru <= r.End; ru++ {
		out = append(out, ru)
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
		cs := strings.Split(line, ";")
		for i, s := range cs {
			cs[i] = strings.TrimSpace(s)
		}
		out = append(out, cs)
	}
	return
}

// filled by `parseUnicodeDatabase`
var (
	generalCategory = map[rune]string{}

	shapingTable struct {
		table    [maxUnicode][4]rune
		min, max rune
	}
	equivTable [maxUnicode]rune

	combiningClasses = map[uint8][]rune{} // class -> runes

	ligatures = map[[2]rune][4]rune{}
)

// rune;comment;General_Category;Canonical_Combining_Class;Bidi_Class;Decomposition_Mapping;...;Bidi_Mirrored
func parseUnicodeDatabase(b []byte) error {
	// initialisation
	for c := range shapingTable.table {
		for i := range shapingTable.table[c] {
			shapingTable.table[c][i] = rune(c)
		}
	}

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
			unshaped rune
		)

		// Rune
		ci, err := strconv.ParseInt(chunks[0], 16, 32)
		if err != nil {
			return fmt.Errorf("invalid line %s: %s", chunks[0], err)
		}
		c = rune(ci)
		if c >= maxUnicode || unshaped >= maxUnicode {
			return fmt.Errorf("invalid rune value: %s", chunks[0])
		}

		// general category
		generalCategory[c] = strings.TrimSpace(chunks[2])

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

		if chunks[5][0] != '<' {
			_, err = fmt.Sscanf(chunks[5], "%x", &unshaped)
			check(err)

			// equiv table
			equivTable[c] = unshaped

			continue
		}

		items := strings.Split(chunks[5], " ")
		if len(items) < 2 {
			check(fmt.Errorf("invalid line %v", chunks))
		}

		unshaped = parseRune(items[1])
		equivTable[c] = unshaped // equiv table

		shape := isShape(items[0])
		if shape == -1 {
			continue
		}

		if len(items) == 3 { // ligatures
			r2 := parseRune(items[2])
			// we only care about lam-alef ligatures
			if unshaped != 0x0644 || !(r2 == 0x0622 || r2 == 0x0623 || r2 == 0x0625 || r2 == 0x0627) {
				continue
			}
			// save ligature
			// names[c] = fields[1]
			v := ligatures[[2]rune{unshaped, r2}]
			v[shape] = c
			ligatures[[2]rune{unshaped, r2}] = v
		}

		// shape table: only single unshaped rune are considered
		if len(items) == 2 {
			shapingTable.table[unshaped][shape] = c
			if unshaped < min {
				min = unshaped
			}
			if unshaped > max {
				max = unshaped
			}
		}
	}
	shapingTable.min, shapingTable.max = min, max

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

func parseAnnexTablesAsRanges(b []byte) (map[string][]runeRange, error) {
	outRanges := map[string][]runeRange{}
	for _, parts := range splitLines(b) {
		if len(parts) < 2 {
			return nil, fmt.Errorf("invalid line: %s", parts)
		}
		rang, typ := strings.TrimSpace(parts[0]), strings.TrimSpace(strings.Split(parts[1], "#")[0])
		rangS := strings.Split(rang, "..")
		start := parseRune(rangS[0])
		end := start
		if len(rangS) > 1 {
			end = parseRune(rangS[1])
		}
		outRanges[typ] = append(outRanges[typ], runeRange{Start: start, End: end})
	}
	return outRanges, nil
}

func parseAnnexTables(b []byte) (map[string][]rune, error) {
	tmp, err := parseAnnexTablesAsRanges(b)
	if err != nil {
		return nil, err
	}
	outRanges := map[string][]rune{}
	for k, v := range tmp {
		for _, r := range v {
			outRanges[k] = append(outRanges[k], r.runes()...)
		}
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
		startRune, endRune := parseRune(start), parseRune(end)
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
	Dm        string `xml:"dm,attr"`
	Dt        string `xml:"dt,attr"`
	CompEx    string `xml:"Comp_Ex,attr"`
	Chars     []char `xml:"char"`
	Reserved  []char `xml:"reserved"`
	NonChar   []char `xml:"noncharacter"`
	Surrogate []char `xml:"surrogate"`
}

type char struct {
	Cp      string `xml:"cp,attr"`
	FirstCp string `xml:"first-cp,attr"`
	LastCp  string `xml:"last-cp,attr"`
	Dm      string `xml:"dm,attr"`
	Dt      string `xml:"dt,attr"`
	CompEx  string `xml:"Comp_Ex,attr"`
}

func parseXML(filename string) (map[rune][]rune, map[rune]bool) {
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

	parseDm := func(dm string) (runes []rune) {
		if dm == "#" {
			return nil
		}
		return parseRunes(dm)
	}

	dms := map[rune][]rune{}
	compEx := map[rune]bool{}
	handleRunes := func(l []char, gr group) {
		for _, ch := range l {
			if ch.Dm == "" {
				ch.Dm = gr.Dm
			}
			if ch.Dt == "" {
				ch.Dt = gr.Dt
			}
			if ch.CompEx == "" {
				ch.CompEx = gr.CompEx
			}
			if ch.Dt != "can" {
				continue
			}

			runes := parseDm(ch.Dm)

			if ch.Cp != "" {
				ru, err := strconv.ParseInt(ch.Cp, 16, 32)
				check(err)
				dms[rune(ru)] = runes
				if ch.CompEx == "Y" {
					compEx[rune(ru)] = true
				}
			} else {
				firstRune, err := strconv.ParseInt(ch.FirstCp, 16, 32)
				check(err)
				lastRune, err := strconv.ParseInt(ch.LastCp, 16, 32)
				check(err)
				for ru := firstRune; ru <= lastRune; ru++ {
					dms[rune(ru)] = runes
					if ch.CompEx == "Y" {
						compEx[rune(ru)] = true
					}
				}
			}
		}
	}

	for _, group := range out.Reps {
		handleRunes(group.Chars, group)
		handleRunes(group.Reserved, group)
		handleRunes(group.NonChar, group)
		handleRunes(group.Surrogate, group)
	}

	// remove unused runes
	for i := 0xAC00; i < 0xAC00+11172; i++ {
		delete(dms, rune(i))
	}

	return dms, compEx
}

// return the joining type and joining group
func parseArabicShaping(b []byte) map[rune]ucd.ArabicJoining {
	out := make(map[rune]ucd.ArabicJoining)
	for _, fields := range splitLines(b) {
		if len(fields) < 2 {
			check(fmt.Errorf("invalid line %v", fields))
		}

		var c rune
		_, err := fmt.Sscanf(fields[0], "%x", &c)
		if err != nil {
			check(fmt.Errorf("invalid line %v: %s", fields, err))
		}

		if c >= maxUnicode {
			check(fmt.Errorf("to high rune value: %d", c))
		}

		if fields[2] == "" {
			check(fmt.Errorf("invalid line %v", fields))
		}

		joiningType := ucd.ArabicJoining(fields[2][0])
		if len(fields) >= 4 {
			switch fields[3] {
			case "ALAPH":
				joiningType = 'a'
			case "DALATH RISH":
				joiningType = 'd'
			}
		}

		switch joiningType {
		case ucd.U, ucd.R, ucd.Alaph, ucd.DalathRish, ucd.D, ucd.C, ucd.L, ucd.T, ucd.G:
		default:
			check(fmt.Errorf("invalid joining type %s", string(joiningType)))
		}

		out[c] = joiningType
	}

	return out
}

func parseUSEInvalidCluster(b []byte) [][]rune {
	var constraints [][]rune
	for _, parts := range splitLines(b) {
		if len(parts) < 1 {
			check(fmt.Errorf("invalid line: %s", parts))
		}

		constraint := parseRunes(parts[0])
		if len(constraint) == 0 {
			continue
		}
		if len(constraint) == 1 {
			check(fmt.Errorf("prohibited sequence is too short: %v", constraint))
		}
		constraints = append(constraints, constraint)
	}
	return constraints
}

func parseEmojisTest(b []byte) (sequences [][]rune) {
	for _, line := range splitLines(b) {
		if len(line) == 0 {
			continue
		}
		runes := parseRunes(line[0])
		sequences = append(sequences, runes)
	}
	return sequences
}

func parseScriptNames(b []byte) (map[string]uint32, error) {
	m := map[string]uint32{}
	for _, chunks := range splitLines(b) {
		code := chunks[0]
		if len(code) != 4 {
			return nil, fmt.Errorf("invalid code %s ", code)
		}

		if code == "Geok" {
			continue // special case: duplicate tag
		}
		tag := binary.BigEndian.Uint32([]byte(strings.ToLower(code)))

		pva := chunks[4]
		if pva == "" {
			continue
		}
		m[pva] = tag
	}
	return m, nil
}
