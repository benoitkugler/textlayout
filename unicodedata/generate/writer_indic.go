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
	OT_X = iota
	OT_C
	OT_V
	OT_N
	OT_H
	OT_ZWNJ
	OT_ZWJ
	OT_M
	OT_SM
	_ // OT_VD: UNUSED; we use OT_A instead.
	OT_A
	OT_PLACEHOLDER
	OT_DOTTEDCIRCLE
	OT_RS    /* Register Shifter, used in Khmer OT spec. */
	OT_Coeng /* Khmer-style Virama. */
	OT_Repha /* Atomically-encoded logical or visual repha. */
	OT_Ra
	OT_CM     /* Consonant-Medial. */
	OT_Symbol /* Avagraha, etc that take marks (SM,A,VD). */
	OT_CS

	/* The following are used by Khmer & Myanmar shapers.  Defined
	 * here for them to share. */
	OT_VAbv = 26
	OT_VBlw = 27
	OT_VPre = 28
	OT_VPst = 29
)

/* indic_position_t Visual positions in a syllable from left to right. */
const (
	POS_START = iota

	POS_RA_TO_BECOME_REPH
	POS_PRE_M
	POS_PRE_C

	POS_BASE_C
	POS_AFTER_MAIN

	POS_ABOVE_C

	POS_BEFORE_SUB
	POS_BELOW_C
	POS_AFTER_SUB

	POS_BEFORE_POST
	POS_POST_C
	POS_AFTER_POST

	POS_FINAL_C
	POS_SMVD

	POS_END
)

// we keep only on source of truth, here,
// and also generate the enumerations.
var (
	otNames = [...]string{
		OT_X:            "OT_X",
		OT_C:            "OT_C",
		OT_V:            "OT_V",
		OT_N:            "OT_N",
		OT_H:            "OT_H",
		OT_ZWNJ:         "OT_ZWNJ",
		OT_ZWJ:          "OT_ZWJ",
		OT_M:            "OT_M",
		OT_SM:           "OT_SM",
		OT_A:            "OT_A",
		OT_PLACEHOLDER:  "OT_PLACEHOLDER",
		OT_DOTTEDCIRCLE: "OT_DOTTEDCIRCLE",
		OT_RS:           "OT_RS",
		OT_Coeng:        "OT_Coeng",
		OT_Repha:        "OT_Repha",
		OT_Ra:           "OT_Ra",
		OT_CM:           "OT_CM",
		OT_Symbol:       "OT_Symbol",
		OT_CS:           "OT_CS",
		OT_VAbv:         "OT_VAbv",
		OT_VBlw:         "OT_VBlw",
		OT_VPre:         "OT_VPre",
		OT_VPst:         "OT_VPst",
	}

	posNames = [...]string{
		POS_START: "POS_START",

		POS_RA_TO_BECOME_REPH: "POS_RA_TO_BECOME_REPH",
		POS_PRE_M:             "POS_PRE_M",
		POS_PRE_C:             "POS_PRE_C",

		POS_BASE_C:     "POS_BASE_C",
		POS_AFTER_MAIN: "POS_AFTER_MAIN",

		POS_ABOVE_C: "POS_ABOVE_C",

		POS_BEFORE_SUB: "POS_BEFORE_SUB",
		POS_BELOW_C:    "POS_BELOW_C",
		POS_AFTER_SUB:  "POS_AFTER_SUB",

		POS_BEFORE_POST: "POS_BEFORE_POST",
		POS_POST_C:      "POS_POST_C",
		POS_AFTER_POST:  "POS_AFTER_POST",

		POS_FINAL_C: "POS_FINAL_C",
		POS_SMVD:    "POS_SMVD",

		POS_END: "POS_END",
	}
)

/* indic_syllabic_category_t Categories used in IndicSyllabicCategory.txt from UCD. */
var indicSyllabicCategories = map[string]uint8{
	"Other": OT_X,

	"Avagraha":                   OT_Symbol,
	"Bindu":                      OT_SM,
	"Brahmi_Joining_Number":      OT_PLACEHOLDER, /* Don't care. */
	"Cantillation_Mark":          OT_A,
	"Consonant":                  OT_C,
	"Consonant_Dead":             OT_C,
	"Consonant_Final":            OT_CM,
	"Consonant_Head_Letter":      OT_C,
	"Consonant_Killer":           OT_M, /* U+17CD only. */
	"Consonant_Medial":           OT_CM,
	"Consonant_Placeholder":      OT_PLACEHOLDER,
	"Consonant_Preceding_Repha":  OT_Repha,
	"Consonant_Prefixed":         OT_X, /* Don't care. */
	"Consonant_Subjoined":        OT_CM,
	"Consonant_Succeeding_Repha": OT_CM,
	"Consonant_With_Stacker":     OT_CS,
	"Gemination_Mark":            OT_SM, /* https://github.com/harfbuzz/harfbuzz/issues/552 */
	"Invisible_Stacker":          OT_Coeng,
	"Joiner":                     OT_ZWJ,
	"Modifying_Letter":           OT_X,
	"Non_Joiner":                 OT_ZWNJ,
	"Nukta":                      OT_N,
	"Number":                     OT_PLACEHOLDER,
	"Number_Joiner":              OT_PLACEHOLDER, /* Don't care. */
	"Pure_Killer":                OT_M,           /* Is like a vowel matra. */
	"Register_Shifter":           OT_RS,
	"Syllable_Modifier":          OT_SM,
	"Tone_Letter":                OT_X,
	"Tone_Mark":                  OT_N,
	"Virama":                     OT_H,
	"Visarga":                    OT_SM,
	"Vowel":                      OT_V,
	"Vowel_Dependent":            OT_M,
	"Vowel_Independent":          OT_V,
}

/* indic_matra_category_t Categories used in IndicSMatraCategory.txt from UCD */
const (
	INDIC_MATRA_CATEGORY_LEFT   = POS_PRE_C
	INDIC_MATRA_CATEGORY_TOP    = POS_ABOVE_C
	INDIC_MATRA_CATEGORY_BOTTOM = POS_BELOW_C
	INDIC_MATRA_CATEGORY_RIGHT  = POS_POST_C
)

var indicMatraCategory = map[string]uint8{
	"Not_Applicable": POS_END,

	"Left":   POS_PRE_C,
	"Top":    POS_ABOVE_C,
	"Bottom": POS_BELOW_C,
	"Right":  POS_POST_C,

	/* These should resolve to the position of the last part of the split sequence. */
	"Bottom_And_Right":         INDIC_MATRA_CATEGORY_RIGHT,
	"Left_And_Right":           INDIC_MATRA_CATEGORY_RIGHT,
	"Top_And_Bottom":           INDIC_MATRA_CATEGORY_BOTTOM,
	"Top_And_Bottom_And_Left":  INDIC_MATRA_CATEGORY_BOTTOM,
	"Top_And_Bottom_And_Right": INDIC_MATRA_CATEGORY_RIGHT,
	"Top_And_Left":             INDIC_MATRA_CATEGORY_TOP,
	"Top_And_Left_And_Right":   INDIC_MATRA_CATEGORY_RIGHT,
	"Top_And_Right":            INDIC_MATRA_CATEGORY_RIGHT,

	"Overstruck":        POS_AFTER_MAIN,
	"Visual_Order_Left": POS_PRE_M,
}

// resolve the numerical value for syllabic and mantra and combine
// them in one uint16
func indicCombineCategories(Ss, Ms string) uint16 {
	if !(Ss == "Consonant_Medial" ||
		Ss == "Gemination_Mark" ||
		Ss == "Register_Shifter" ||
		Ss == "Consonant_Succeeding_Repha" ||
		Ss == "Virama" ||
		Ss == "Vowel_Dependent") {
		Ms = "Not_Applicable"
	}
	S, ok := indicSyllabicCategories[Ss]
	if !ok {
		check(fmt.Errorf("unknown syllabic category <%s>", Ss))
	}
	M, ok := indicMatraCategory[Ms]
	if !ok {
		check(fmt.Errorf("unknown matra category <%s>", Ms))
	}
	return uint16(S) | uint16(M)<<8
}

var (
	ALLOWED_SINGLES = []rune{0x00A0, 0x25CC}
	ALLOWED_BLOCKS  = []string{
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

func generateIndicTable(indicS, indicP, blocks map[string][]rune, w io.Writer) {
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
	var starts, ends []rune
	fmt.Fprintln(w, "var indicTable = [...]uint16{")
	var offsetsDef string
	for _, u := range uu {
		if u <= last {
			continue
		}

		block := data[u][2]

		start := u / 8 * 8
		end := start + 1
		for inR(end, uu...) && block == data[end][1] {
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

			printBlock(block, start, end)
			last = end
		}
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
		if !(inR(k, ALLOWED_SINGLES...) || in(v[2], ALLOWED_BLOCKS...)) {
			delete(combined, k)
		}
	}

	// Move the outliers NO-BREAK SPACE and DOTTED CIRCLE out
	singles := map[rune][3]string{}
	for _, u := range ALLOWED_SINGLES {
		singles[u] = combined[u]
		delete(combined, u)
	}

	return combined, singles
}
