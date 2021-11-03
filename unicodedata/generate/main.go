// Generate lookup function for Unicode properties not
// covered by the standard package unicode.
package main

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
)

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	fetch := flag.Bool("fetch", false, "download the datas and save them locally (required at first usage)")
	flag.Parse()

	if *fetch {
		fetchData(urlXML)
		fetchData(urlUnicode)
		fetchData(urlEmoji)
		fetchData(urlEmojiTest)
		fetchData(urlMirroring)
		fetchData(urlArabic)
		fetchData(urlScripts)
		fetchData(urlIndic1)
		fetchData(urlIndic2)
		fetchData(urlBlocks)
		fetchData(urlLineBreak)
		fetchData(urlSentenceBreak)
		fetchData(urlDerivedCore)
	}

	// parse
	fmt.Println("Parsing Unicode files...")

	b, err := ioutil.ReadFile("UnicodeData.txt")
	check(err)
	err = parseUnicodeDatabase(b)
	check(err)

	b, err = ioutil.ReadFile("emoji-data.txt")
	check(err)
	emojis, err := parseAnnexTables(b)
	check(err)

	b, err = ioutil.ReadFile("emoji-test.txt")
	check(err)
	emojisTests := parseEmojisTest(b)

	b, err = ioutil.ReadFile("BidiMirroring.txt")
	check(err)
	mirrors, err := parseMirroring(b)
	check(err)

	dms, compEx := parseXML("ucd.nounihan.grouped.zip")

	b, err = ioutil.ReadFile("ArabicShaping.txt")
	check(err)
	joiningTypes := parseArabicShaping(b)

	b, err = ioutil.ReadFile("Scripts.txt")
	check(err)
	scripts, err := parseAnnexTables(b)
	check(err)

	b, err = ioutil.ReadFile("Blocks.txt")
	check(err)
	blocks, err := parseAnnexTables(b)
	check(err)

	b, err = ioutil.ReadFile("IndicSyllabicCategory.txt")
	check(err)
	indicS, err := parseAnnexTables(b)
	check(err)

	b, err = ioutil.ReadFile("IndicPositionalCategory.txt")
	check(err)
	indicP, err := parseAnnexTables(b)
	check(err)

	b, err = ioutil.ReadFile("ms-use/IndicSyllabicCategory-Additional.txt")
	check(err)
	indicSAdd, err := parseAnnexTables(b)
	check(err)

	b, err = ioutil.ReadFile("ms-use/IndicPositionalCategory-Additional.txt")
	check(err)
	indicPAdd, err := parseAnnexTables(b)
	check(err)

	b, err = ioutil.ReadFile("ms-use/IndicShapingInvalidCluster.txt")
	check(err)
	vowelsConstraints := parseUSEInvalidCluster(b)

	b, err = ioutil.ReadFile("LineBreak.txt")
	check(err)
	lineBreak, err := parseAnnexTables(b)
	check(err)

	b, err = ioutil.ReadFile("SentenceBreakProperty.txt")
	check(err)
	sentenceBreaks, err := parseAnnexTables(b)
	check(err)

	b, err = ioutil.ReadFile("Scripts.txt")
	check(err)
	scriptsRanges, err := parseAnnexTablesAsRanges(b)
	check(err)

	b, err = ioutil.ReadFile("Scripts-iso15924.txt")
	check(err)
	scriptNames, err := parseScriptNames(b)
	check(err)

	b, err = ioutil.ReadFile("DerivedCoreProperties.txt")
	check(err)
	derivedCore, err := parseAnnexTables(b)
	check(err)

	// generate
	process("../combining_classes.go", func(w io.Writer) {
		generateCombiningClasses(combiningClasses, w)
	})
	process("../emojis.go", func(w io.Writer) {
		generateEmojis(emojis, w)
	})
	process("../../harfbuzz/emojis_list_test.go", func(w io.Writer) {
		generateEmojisTest(emojisTests, w)
	})
	process("../mirroring.go", func(w io.Writer) {
		generateMirroring(mirrors, w)
	})
	process("../decomposition.go", func(w io.Writer) {
		generateDecomposition(dms, compEx, w)
	})
	process("../arabic.go", func(w io.Writer) {
		generateArabicShaping(joiningTypes, w)
		generateHasArabicJoining(joiningTypes, scripts, w)
	})
	process("../../harfbuzz/ot_use_table.go", func(w io.Writer) {
		generateUSETable(indicS, indicP, blocks, indicSAdd, indicPAdd, derivedCore, scripts, joiningTypes, w)
	})
	process("../../harfbuzz/ot_vowels_constraints.go", func(w io.Writer) {
		generateVowelConstraints(scripts, vowelsConstraints, w)
	})
	process("../../harfbuzz/ot_indic_table.go", func(w io.Writer) {
		generateIndicTable(indicS, indicP, blocks, w)
	})
	process("../linebreak.go", func(w io.Writer) {
		generateLineBreak(lineBreak, w)
	})
	process("../indic.go", func(w io.Writer) {
		generateIndicCategories(indicS, w)
	})
	process("../sentenceBreak.go", func(w io.Writer) {
		generateSTermProperty(sentenceBreaks, w)
	})
	process("../../language/scripts_table.go", func(w io.Writer) {
		generateScriptLookupTable(scriptsRanges, scriptNames, w)
	})
	fmt.Println("Done.")
}

// write into filename
func process(filename string, generator func(w io.Writer)) {
	fmt.Println("Generating", filename, "...")
	file, err := os.Create(filename)
	check(err)

	generator(file)

	err = file.Close()
	check(err)

	cmd := exec.Command("goimports", "-w", filename)
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	check(err)
}
