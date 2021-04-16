package main

import (
	"fmt"
	"io"
)

/* indic_category_t Cateories used in the OpenType spec:
 * https://docs.microsoft.com/en-us/typography/script-development/devanagari
 */
/* Note: This enum is duplicated in the -machine.rl source file.
 * Not sure how to avoid duplication. */
const (
	otX = iota
	otC
	otV
	otN
	otH
	otZWNJ
	otZWJ
	otM
	otSM
	_ // otVD: UNUSED; we use otA instead.
	otA
	otPLACEHOLDER
	otDOTTEDCIRCLE
	otRS    /* Register Shifter, used in Khmer OT spec. */
	otCoeng /* Khmer-style Virama. */
	otRepha /* Atomically-encoded logical or visual repha. */
	otRa
	otCM     /* Consonant-Medial. */
	otSymbol /* Avagraha, etc that take marks (SM,A,VD). */
	otCS

	/* The following are used by Khmer & Myanmar shapers.  Defined
	 * here for them to share. */
	otVAbv = 26
	otVBlw = 27
	otVPre = 28
	otVPst = 29
)

/* indic_position_t Visual positions in a syllable from left to right. */
const (
	posStart = iota

	posRaToBecomeReph
	posPreM
	posPreC

	posBaseC
	posAfterMain

	posAboveC

	posBeforeSub
	posBelowC
	posAfterSub

	posBeforePost
	posPostC
	posAfterPost

	posFinalC
	posSmvd

	posEnd
)

// we keep only on source of truth, here,
// and also generate the enumerations.
var (
	otNames = [...]string{
		otX:            "otX",
		otC:            "otC",
		otV:            "otV",
		otN:            "otN",
		otH:            "otH",
		otZWNJ:         "otZWNJ",
		otZWJ:          "otZWJ",
		otM:            "otM",
		otSM:           "otSM",
		otA:            "otA",
		otPLACEHOLDER:  "otPLACEHOLDER",
		otDOTTEDCIRCLE: "otDOTTEDCIRCLE",
		otRS:           "otRS",
		otCoeng:        "otCoeng",
		otRepha:        "otRepha",
		otRa:           "otRa",
		otCM:           "otCM",
		otSymbol:       "otSymbol",
		otCS:           "otCS",
		otVAbv:         "otVAbv",
		otVBlw:         "otVBlw",
		otVPre:         "otVPre",
		otVPst:         "otVPst",
	}

	posNames = [...]string{
		posStart: "posStart",

		posRaToBecomeReph: "posRaToBecomeReph",
		posPreM:           "posPreM",
		posPreC:           "posPreC",

		posBaseC:     "posBaseC",
		posAfterMain: "posAfterMain",

		posAboveC: "posAboveC",

		posBeforeSub: "posBeforeSub",
		posBelowC:    "posBelowC",
		posAfterSub:  "posAfterSub",

		posBeforePost: "posBeforePost",
		posPostC:      "posPostC",
		posAfterPost:  "posAfterPost",

		posFinalC: "posFinalC",
		posSmvd:   "posSmvd",

		posEnd: "posEnd",
	}
)

/* indic_syllabic_category_t Categories used in IndicSyllabicCategory.txt from UCD. */
var indicSyllabicCategories = map[string]uint8{
	"Other": otX,

	"Avagraha":                   otSymbol,
	"Bindu":                      otSM,
	"Brahmi_Joining_Number":      otPLACEHOLDER, /* Don't care. */
	"Cantillation_Mark":          otA,
	"Consonant":                  otC,
	"Consonant_Dead":             otC,
	"Consonant_Final":            otCM,
	"Consonant_Head_Letter":      otC,
	"Consonant_Killer":           otM, /* U+17CD only. */
	"Consonant_Medial":           otCM,
	"Consonant_Placeholder":      otPLACEHOLDER,
	"Consonant_Preceding_Repha":  otRepha,
	"Consonant_Prefixed":         otX, /* Don't care. */
	"Consonant_Subjoined":        otCM,
	"Consonant_Succeeding_Repha": otCM,
	"Consonant_With_Stacker":     otCS,
	"Gemination_Mark":            otSM, /* https://github.com/harfbuzz/harfbuzz/issues/552 */
	"Invisible_Stacker":          otCoeng,
	"Joiner":                     otZWJ,
	"Modifying_Letter":           otX,
	"Non_Joiner":                 otZWNJ,
	"Nukta":                      otN,
	"Number":                     otPLACEHOLDER,
	"Number_Joiner":              otPLACEHOLDER, /* Don't care. */
	"Pure_Killer":                otM,           /* Is like a vowel matra. */
	"Register_Shifter":           otRS,
	"Syllable_Modifier":          otSM,
	"Tone_Letter":                otX,
	"Tone_Mark":                  otN,
	"Virama":                     otH,
	"Visarga":                    otSM,
	"Vowel":                      otV,
	"Vowel_Dependent":            otM,
	"Vowel_Independent":          otV,
}

/* indic_matra_category_t Categories used in IndicSMatraCategory.txt from UCD */
const (
	indicMatraCategoryLeft   = posPreC
	indicMatraCategoryTop    = posAboveC
	indicMatraCategoryBottom = posBelowC
	indicMatraCategoryRight  = posPostC
)

var indicMatraCategory = map[string]uint8{
	"Not_Applicable": posEnd,

	"Left":   posPreC,
	"Top":    posAboveC,
	"Bottom": posBelowC,
	"Right":  posPostC,

	/* These should resolve to the position of the last part of the split sequence. */
	"Bottom_And_Right":         indicMatraCategoryRight,
	"Left_And_Right":           indicMatraCategoryRight,
	"Top_And_Bottom":           indicMatraCategoryBottom,
	"Top_And_Bottom_And_Left":  indicMatraCategoryBottom,
	"Top_And_Bottom_And_Right": indicMatraCategoryRight,
	"Top_And_Left":             indicMatraCategoryTop,
	"Top_And_Left_And_Right":   indicMatraCategoryRight,
	"Top_And_Right":            indicMatraCategoryRight,

	"Overstruck":        posAfterMain,
	"Visual_Order_Left": posPreM,
}

// resolve the numerical value for syllabic and mantra and combine
// them in one uint16
func indicCombineCategories(Ss, Ms string) uint16 {
	S, ok := indicSyllabicCategories[Ss]
	if !ok {
		check(fmt.Errorf("unknown syllabic category <%s>", Ss))
	}

	if !(S == indicSyllabicCategories["Consonant_Medial"] ||
		S == indicSyllabicCategories["Gemination_Mark"] ||
		S == indicSyllabicCategories["Register_Shifter"] ||
		S == indicSyllabicCategories["Consonant_Succeeding_Repha"] ||
		S == indicSyllabicCategories["Virama"] ||
		S == indicSyllabicCategories["Vowel_Dependent"]) {
		Ms = "Not_Applicable"
	}

	M, ok := indicMatraCategory[Ms]
	if !ok {
		check(fmt.Errorf("unknown matra category <%s>", Ms))
	}
	return uint16(S) | uint16(M)<<8
}

var (
	allowedSingles = []rune{0x00A0, 0x25CC}
	allowedBlocks  = []string{
		"Basic Latin",
		"Latin-1 Supplement",
		"Devanagari",
		"Bengali",
		"Gurmukhi",
		"Gujarati",
		"Oriya",
		"Tamil",
		"Telugu",
		"Kannada",
		"Malayalam",
		"Sinhala",
		"Myanmar",
		"Khmer",
		"Vedic Extensions",
		"General Punctuation",
		"Superscripts and Subscripts",
		"Devanagari Extended",
		"Myanmar Extended-B",
		"Myanmar Extended-A",
	}
)

var defaultsIndic = [3]string{"Other", "Not_Applicable", "No_Block"}

func generateIndicTable(indicS, indicP, blocks map[string][]rune, w io.Writer) (starts, ends []rune) {
	data, singles := aggregateIndicTable(indicS, indicP, blocks)

	fmt.Fprintln(w, `
	package harfbuzz

	// Code generated by unicodedata/generate/main.go DO NOT EDIT.

	`)

	// enumerations

	fmt.Fprintln(w, "const (")
	for i, name := range otNames {
		if name == "" {
			continue
		}
		fmt.Fprintf(w, "%s = %d\n", name, i)
	}
	fmt.Fprintln(w, ")")

	fmt.Fprintln(w, "const (")
	for i, name := range posNames {
		if name == "" {
			continue
		}
		fmt.Fprintf(w, "%s = %d\n", name, i)
	}
	fmt.Fprintln(w, ")")

	total := 0
	used := 0
	lastBlock := ""
	printBlock := func(block string, start, end rune) {
		if block != "" && block != lastBlock {
			fmt.Fprintln(w)
			fmt.Fprintln(w)
			fmt.Fprintf(w, "  /* %s */\n", block)
		}
		num := 0
		if start%8 != 0 {
			check(fmt.Errorf("in printBlock, expected start%%8 == 0, got %d", start))
		}
		if (end+1)%8 != 0 {
			check(fmt.Errorf("in printBlock, expected (end+1)%%8 == 0, got %d", end+1))
		}
		for u := start; u <= end; u++ {
			if u%16 == 0 {
				fmt.Fprintln(w)
				fmt.Fprintf(w, "  /* %04X */", u)
			}
			d, in := data[u]
			if in {
				num += 1
			} else {
				d = defaultsIndic
			}
			fmt.Fprintf(w, "0x%x,", indicCombineCategories(d[0], d[1]))
		}
		total += int(end - start + 1)
		used += num
		if block != "" {
			lastBlock = block
		}
	}
	var uu []rune
	for u := range data {
		uu = append(uu, u)
	}
	sortRunes(uu)

	last := rune(-100000)
	offset := 0
	fmt.Fprintln(w, "var indicTable = [...]uint16{")
	var offsetsDef string
	for _, u := range uu {
		if u <= last {
			continue
		}

		block := data[u][2]

		start := u / 8 * 8
		end := start + 1
		for inR(end, uu...) && block == data[end][2] {
			end += 1
		}
		end = (end-1)/8*8 + 7

		if start != last+1 {
			if start-last <= 1+16*3 {
				printBlock("", last+1, start-1)
			} else {
				if last >= 0 {
					ends = append(ends, last+1)
					offset += int(ends[len(ends)-1] - starts[len(starts)-1])
				}
				fmt.Fprintln(w)
				fmt.Fprintln(w)
				offsetsDef += fmt.Sprintf("offsetIndic0x%04xu = %d \n", start, offset)
				starts = append(starts, start)
			}
		}

		printBlock(block, start, end)
		last = end
	}

	ends = append(ends, last+1)
	offset += int(ends[len(ends)-1] - starts[len(starts)-1])
	fmt.Fprintln(w)
	fmt.Fprintln(w)
	occupancy := used * 100. / total
	pageBits := 12
	fmt.Fprintf(w, "}; /* Table items: %d; occupancy: %d%% */\n", offset, occupancy)
	fmt.Fprintln(w)

	fmt.Fprintln(w, "const (")
	fmt.Fprintln(w, offsetsDef)
	fmt.Fprintln(w, ")")

	fmt.Fprintln(w, "func indicGetCategories (u rune) uint16 {")
	fmt.Fprintf(w, "  switch u >> %d { \n", pageBits)

	pagesSet := map[rune]bool{}
	for _, u := range append(starts, ends...) {
		pagesSet[u>>pageBits] = true
	}
	for k := range singles {
		pagesSet[k>>pageBits] = true
	}
	var pages []rune
	for p := range pagesSet {
		pages = append(pages, p)
	}
	sortRunes(pages)
	for _, p := range pages {
		fmt.Fprintf(w, "    case 0x%0X:\n", p)
		for u, d := range singles {
			if p != u>>pageBits {
				continue
			}
			fmt.Fprintf(w, "      if u == 0x%04X {return 0x%x};\n", u, indicCombineCategories(d[0], d[1]))
		}
		for i, start := range starts {
			end := ends[i]
			if p != start>>pageBits && p != end>>pageBits {
				continue
			}
			offset := fmt.Sprintf("offsetIndic0x%04xu", start)
			fmt.Fprintf(w, "      if  0x%04X <= u && u <= 0x%04X {return indicTable[u - 0x%04X + %s]};\n", start, end-1, start, offset)
		}

		fmt.Fprintln(w, "")
	}
	fmt.Fprintln(w, "  }")
	fmt.Fprintf(w, "  return 0x%x\n", indicCombineCategories("Other", "Not_Applicable"))
	fmt.Fprintln(w, "}")
	fmt.Fprintln(w)

	// Maintain at least 30% occupancy in the table */
	if occupancy < 30 {
		check(fmt.Errorf("table too sparse, please investigate: %d", occupancy))
	}

	return starts, ends // to do some basic tests
}

func aggregateIndicTable(indicS, indicP, blocks map[string][]rune) (map[rune][3]string, map[rune][3]string) {
	// Merge data into one dict:
	data := [3]map[rune]string{{}, {}, {}}

	for t, rs := range indicS {
		for _, r := range rs {
			data[0][r] = t
		}
	}
	for t, rs := range indicP {
		for _, r := range rs {
			data[1][r] = t
		}
	}
	for t, rs := range blocks {
		for _, r := range rs {
			data[2][r] = t
		}
	}

	combined := map[rune][3]string{}
	for i, d := range data {
		for u, v := range d {
			vals, ok := combined[u]
			if i == 2 && !ok {
				continue
			}
			if !ok {
				vals = defaultsIndic
			}
			vals[i] = v
			combined[u] = vals
		}
	}
	for k, v := range combined {
		if !(inR(k, allowedSingles...) || in(v[2], allowedBlocks...)) {
			delete(combined, k)
		}
	}

	// Move the outliers NO-BREAK SPACE and DOTTED CIRCLE out
	singles := map[rune][3]string{}
	for _, u := range allowedSingles {
		singles[u] = combined[u]
		delete(combined, u)
	}

	return combined, singles
}
