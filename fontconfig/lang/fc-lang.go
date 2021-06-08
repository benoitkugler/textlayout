// Read a set of language orthographies and build Go declarations for
// charsets which can then be used to identify which languages are
// supported by a given font.
package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

// port from fontconfig/fc-lang/fc-lang.py
//
// Copyright © 2001-2002 Keith Packard
// Copyright © 2019 Tim-Philipp Müller

func assert(b bool) {
	if !b {
		log.Fatal("assertion error")
	}
}

// we just store the leaves in a dict, we can order the leaves later if needed
type charset map[uint16][8]uint32 // leaf_number -> leaf data (= 16 uint32)

// Build a single charset from a source file
//
// The file format is quite simple, either
// a single hex value or a pair separated with a dash
func parseOrthFile(fileName string, lines []lineData) charset {
	charset := make(charset)
	for _, l := range lines {
		fn, num, line := l.fileName, l.num, l.line
		deleteChar := strings.HasPrefix(line, "-")
		if deleteChar {
			line = line[1:]
		}
		var parts []string
		if strings.IndexByte(line, '-') != -1 {
			parts = strings.Split(line, "-")
		} else if strings.Index(line, "..") != -1 {
			parts = strings.Split(line, "..")
		} else {
			parts = []string{line}
		}

		var startString, endString string

		startString, parts = strings.TrimSpace(parts[0]), parts[1:]
		start, err := strconv.ParseInt(strings.TrimPrefix(startString, "0x"), 16, 32)
		if err != nil {
			log.Fatal("can't parse ", startString, " : ", err)
		}

		end := start
		if len(parts) != 0 {
			endString, parts = strings.TrimSpace(parts[0]), parts[1:]
			end, err = strconv.ParseInt(strings.TrimPrefix(endString, "0x"), 16, 32)
			if err != nil {
				log.Fatal("can't parse ", endString, " : ", err)
			}
		}
		if len(parts) != 0 {
			log.Fatalf("%s line %d: parse error (too many parts)", fn, num)
		}

		for ucs4 := start; ucs4 <= end; ucs4++ {
			if deleteChar {
				charset.delChar(int(ucs4))
			} else {
				charset.addChar(int(ucs4))
			}
		}
	}
	assert(charset.equals(charset)) // sanity check for the equals function

	return charset
}

func (cs charset) addChar(ucs4 int) {
	assert(ucs4 < 0x01000000)
	leafNum := uint16(ucs4 >> 8)
	leaf := cs[leafNum]
	leaf[(ucs4&0xff)>>5] |= (1 << (ucs4 & 0x1f))
	cs[leafNum] = leaf
}

func (cs charset) delChar(ucs4 int) {
	assert(ucs4 < 0x01000000)
	leafNum := uint16(ucs4 >> 8)
	if leaf, ok := cs[leafNum]; ok {
		leaf[(ucs4&0xff)>>5] &= ^(1 << (ucs4 & 0x1f))
		cs[leafNum] = leaf
		// We don't bother removing the leaf if it's empty
	}
}

func (cs charset) sortedKeys() []int {
	var keys []int
	for k := range cs {
		keys = append(keys, int(k))
	}
	sort.Ints(keys)
	return keys
}

func (cs charset) equals(otherCs charset) bool {
	if len(cs) != len(otherCs) {
		return false
	}

	for i := range cs {
		if cs[i] != otherCs[i] {
			return false
		}
	}
	return true
}

// Convert a file name into a name suitable for Go declarations
func getName(fileName string) string {
	return strings.Split(fileName, ".")[0]
}

// Convert a C name into a language name
func getLang(cName string) string {
	cName = strings.ReplaceAll(cName, "_", "-")
	cName = strings.ReplaceAll(cName, " ", "")
	return strings.ToLower(cName)
}

type lineData struct {
	fileName string
	num      int
	line     string
}

func readOrthFile(fileName string) []lineData {
	var lines []lineData

	b, err := ioutil.ReadFile("orth/" + fileName)
	if err != nil {
		log.Fatal("can't read ", fileName, err)
	}
	linesString := strings.Split(string(b), "\n")
	for num, line := range linesString {
		if strings.HasPrefix(line, "include ") {
			includeFn := strings.TrimSpace(line[8:])
			lines = append(lines, readOrthFile(includeFn)...)
		} else {
			// remove comments and strip whitespaces
			line = strings.TrimSpace(strings.Split(line, "#")[0])
			line = strings.TrimSpace(strings.Split(line, "\t")[0])
			// skip empty lines
			if line != "" {
				lines = append(lines, lineData{fileName, num, line})
			}
		}
	}
	return lines
}

func main() {
	table := flag.String("table", "", "output table file")
	conf := flag.String("conf", "", "output conf dir")
	flag.Parse()

	generateLangTable(*table)
	generate35Conf(*conf)
}

func generate35Conf(confDir string) {
	var orthList []string
	for _, o := range orthFiles {
		o = strings.Split(o, ".")[0] // strip filename suffix
		if strings.IndexByte(o, '_') == -1 {
			orthList = append(orthList, o) // ignore those with an underscore
		}
	}

	sort.Strings(orthList)

	output, err := os.Create(filepath.Join(confDir, "35-lang-normalize.conf"))
	if err != nil {
		log.Fatal(err)
	}
	defer output.Close()

	fmt.Fprintln(output, `<?xml version="1.0"?>\`)
	fmt.Fprintln(output, `<!DOCTYPE fontconfig SYSTEM "urn:fontconfig:fonts.dtd">`)
	fmt.Fprintln(output, "<fontconfig>")

	for _, o := range orthList {
		fmt.Fprintf(output, "  <!-- %s* -> %s -->\n", o, o)
		fmt.Fprintln(output, "  <match>")
		fmt.Fprintf(output, `    <test name="lang" compare="contains"><string>%s</string></test>`, o)
		fmt.Fprintf(output, `    <edit name="lang" mode="assign" binding="same"><string>%s</string></edit>`, o)
		fmt.Fprintln(output, "  </match>")
	}

	fmt.Fprintln(output, "</fontconfig>")
}

func generateLangTable(output string) {
	var (
		err             error
		sets            []charset
		country         []int
		names, langs    []string
		langCountrySets = map[string][]int{}
	)

	outputFile := os.Stdout
	// Open output file
	if output != "" {
		outputFile, err = os.Create(output)
		if err != nil {
			log.Fatal(err)
		}
	}

	orthEntries := map[string]int{}
	var sortedKeys []string
	for i, fn := range orthFiles {
		orthEntries[fn] = i
		sortedKeys = append(sortedKeys, fn)
	}
	sort.Strings(sortedKeys)

	for _, fn := range sortedKeys {
		lines := readOrthFile(fn)
		charset := parseOrthFile(fn, lines)

		sets = append(sets, charset)

		name := getName(fn)
		names = append(names, name)

		lang := getLang(name)
		langs = append(langs, lang)
		if strings.Index(lang, "-") != -1 {
			country = append(country, orthEntries[fn]) // maps to original index
			languageFamily := strings.Split(lang, "-")[0]
			langCountrySets[languageFamily] = append(langCountrySets[languageFamily], orthEntries[fn])
		}
	}

	// ----------------------------------- output -----------------------------------

	fmt.Fprintln(outputFile, "package fontconfig")
	fmt.Fprintln(outputFile, "// Code auto-generated by lang/fc-lang.go DO NOT EDIT")
	fmt.Fprintln(outputFile)
	fmt.Fprintf(outputFile, "// Number of charsets: %d\n\n", len(sets))

	// If this fails, the lang indices will need to be 16-bit, instead of a single byte.
	assert(len(sets) < 256)

	// Serialize leaves
	dumpLeaves := func(leaves [][8]uint32) string {
		var chunks []string
		for _, leaf := range leaves {
			chunks = append(chunks, fmt.Sprintf("%#v", leaf))
		}
		return strings.Join(chunks, ",\n")
	}

	fmt.Fprintf(outputFile, `
	var fcLangCharSets = [...]langToCharset{`)
	fmt.Fprintln(outputFile)
	// Serialize sets
	for i, set := range sets {
		numbers := make([]uint16, len(set))
		leaves := make([][8]uint32, len(set))
		for pos, k := range set.sortedKeys() {
			numbers[pos] = uint16(k)
			leaves[pos] = set[uint16(k)]
		}

		fmt.Fprintf(outputFile, "    { %q, Charset{\n%#v, \n[]charPage{\n%s,\n},\n} }, // %d ",
			langs[i], numbers, dumpLeaves(leaves), len(set))
		fmt.Fprintln(outputFile)
	}
	fmt.Fprintln(outputFile, "}")
	fmt.Fprintln(outputFile)

	// langIndices
	fmt.Fprintln(outputFile, "var fcLangCharSetIndices = [...]byte{")
	for i := range sets {
		fn := fmt.Sprintf("%s.orth", names[i])
		fmt.Fprintf(outputFile, "    %d, /* %s */\n", orthEntries[fn], names[i])
	}
	fmt.Fprintln(outputFile, "}")

	// langIndicesInv
	fmt.Fprintln(outputFile, "var fcLangCharSetIndicesInv = [...]byte{")
	var invLines []string // to sort
	for k, pos := range orthEntries {
		name := getName(k)
		idx := -1
		for i, s := range names {
			if s == name {
				idx = i
				break
			}
		}
		invLines = append(invLines, fmt.Sprintf("    %d: %d, /* %s */\n", pos, idx, name))
	}
	sort.Strings(invLines)
	fmt.Fprintln(outputFile, strings.Join(invLines, ""))
	fmt.Fprintln(outputFile, "}")

	num_lang_set_map := (len(sets) + 31) / 32
	fmt.Fprintf(outputFile, "const langPageSize	= %d \n", num_lang_set_map)

	// Serialize indices with country codes
	assert(len(country) > 0)
	assert(len(langCountrySets) > 0)
	fmt.Fprintln(outputFile)
	fmt.Fprintln(outputFile, "var fcLangCountrySets = [][langPageSize]uint32 {")
	var langCSSortedkeys []string
	for k := range langCountrySets {
		langCSSortedkeys = append(langCSSortedkeys, k)
	}
	sort.Strings(langCSSortedkeys)
	for _, k := range langCSSortedkeys {
		langset_map := make([]int, num_lang_set_map) // initialise all zeros
		for _, entries_id := range langCountrySets[k] {
			langset_map[entries_id>>5] |= (1 << (entries_id & 0x1f))
		}
		fmt.Fprintf(outputFile, "    {")
		for _, v := range langset_map {
			fmt.Fprintf(outputFile, " 0x%08x,", v)
		}
		fmt.Fprintf(outputFile, " }, /* %s */\n", k)
	}
	fmt.Fprintln(outputFile, "};")

	// Find ranges for each letter for faster searching
	// Dumps sets start/finish for the fastpath
	fmt.Fprintln(outputFile, "var fcLangCharSetRanges = []langCharsetRange{")
	for c := 'a'; c <= 'z'; c++ {
		start := 9999
		stop := -1
		for i := range sets {
			if strings.HasPrefix(names[i], string(c)) {
				start = min(start, i)
				stop = max(stop, i)
			}
		}
		fmt.Fprintf(outputFile, "    { %d, %d }, /* %d */\n", start, stop, c)
	}
	fmt.Fprintln(outputFile, "};")

	if output != "" {
		outputFile.Close()
		exec.Command("goimports", "-w", output).Run()
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// Do not reorder, magic
var orthFiles = []string{
	"aa.orth",
	"ab.orth",
	"af.orth",
	"am.orth",
	"ar.orth",
	"as.orth",
	"ast.orth",
	"av.orth",
	"ay.orth",
	"az_az.orth",
	"az_ir.orth",
	"ba.orth",
	"bm.orth",
	"be.orth",
	"bg.orth",
	"bh.orth",
	"bho.orth",
	"bi.orth",
	"bin.orth",
	"bn.orth",
	"bo.orth",
	"br.orth",
	"bs.orth",
	"bua.orth",
	"ca.orth",
	"ce.orth",
	"ch.orth",
	"chm.orth",
	"chr.orth",
	"co.orth",
	"cs.orth",
	"cu.orth",
	"cv.orth",
	"cy.orth",
	"da.orth",
	"de.orth",
	"dz.orth",
	"el.orth",
	"en.orth",
	"eo.orth",
	"es.orth",
	"et.orth",
	"eu.orth",
	"fa.orth",
	"fi.orth",
	"fj.orth",
	"fo.orth",
	"fr.orth",
	"ff.orth",
	"fur.orth",
	"fy.orth",
	"ga.orth",
	"gd.orth",
	"gez.orth",
	"gl.orth",
	"gn.orth",
	"gu.orth",
	"gv.orth",
	"ha.orth",
	"haw.orth",
	"he.orth",
	"hi.orth",
	"ho.orth",
	"hr.orth",
	"hu.orth",
	"hy.orth",
	"ia.orth",
	"ig.orth",
	"id.orth",
	"ie.orth",
	"ik.orth",
	"io.orth",
	"is.orth",
	"it.orth",
	"iu.orth",
	"ja.orth",
	"ka.orth",
	"kaa.orth",
	"ki.orth",
	"kk.orth",
	"kl.orth",
	"km.orth",
	"kn.orth",
	"ko.orth",
	"kok.orth",
	"ks.orth",
	"ku_am.orth",
	"ku_ir.orth",
	"kum.orth",
	"kv.orth",
	"kw.orth",
	"ky.orth",
	"la.orth",
	"lb.orth",
	"lez.orth",
	"ln.orth",
	"lo.orth",
	"lt.orth",
	"lv.orth",
	"mg.orth",
	"mh.orth",
	"mi.orth",
	"mk.orth",
	"ml.orth",
	"mn_cn.orth",
	"mo.orth",
	"mr.orth",
	"mt.orth",
	"my.orth",
	"nb.orth",
	"nds.orth",
	"ne.orth",
	"nl.orth",
	"nn.orth",
	"no.orth",
	"nr.orth",
	"nso.orth",
	"ny.orth",
	"oc.orth",
	"om.orth",
	"or.orth",
	"os.orth",
	"pa.orth",
	"pl.orth",
	"ps_af.orth",
	"ps_pk.orth",
	"pt.orth",
	"rm.orth",
	"ro.orth",
	"ru.orth",
	"sa.orth",
	"sah.orth",
	"sco.orth",
	"se.orth",
	"sel.orth",
	"sh.orth",
	"shs.orth",
	"si.orth",
	"sk.orth",
	"sl.orth",
	"sm.orth",
	"sma.orth",
	"smj.orth",
	"smn.orth",
	"sms.orth",
	"so.orth",
	"sq.orth",
	"sr.orth",
	"ss.orth",
	"st.orth",
	"sv.orth",
	"sw.orth",
	"syr.orth",
	"ta.orth",
	"te.orth",
	"tg.orth",
	"th.orth",
	"ti_er.orth",
	"ti_et.orth",
	"tig.orth",
	"tk.orth",
	"tl.orth",
	"tn.orth",
	"to.orth",
	"tr.orth",
	"ts.orth",
	"tt.orth",
	"tw.orth",
	"tyv.orth",
	"ug.orth",
	"uk.orth",
	"ur.orth",
	"uz.orth",
	"ve.orth",
	"vi.orth",
	"vo.orth",
	"vot.orth",
	"wa.orth",
	"wen.orth",
	"wo.orth",
	"xh.orth",
	"yap.orth",
	"yi.orth",
	"yo.orth",
	"zh_cn.orth",
	"zh_hk.orth",
	"zh_mo.orth",
	"zh_sg.orth",
	"zh_tw.orth",
	"zu.orth",
	"ak.orth",
	"an.orth",
	"ber_dz.orth",
	"ber_ma.orth",
	"byn.orth",
	"crh.orth",
	"csb.orth",
	"dv.orth",
	"ee.orth",
	"fat.orth",
	"fil.orth",
	"hne.orth",
	"hsb.orth",
	"ht.orth",
	"hz.orth",
	"ii.orth",
	"jv.orth",
	"kab.orth",
	"kj.orth",
	"kr.orth",
	"ku_iq.orth",
	"ku_tr.orth",
	"kwm.orth",
	"lg.orth",
	"li.orth",
	"mai.orth",
	"mn_mn.orth",
	"ms.orth",
	"na.orth",
	"ng.orth",
	"nv.orth",
	"ota.orth",
	"pa_pk.orth",
	"pap_an.orth",
	"pap_aw.orth",
	"qu.orth",
	"quz.orth",
	"rn.orth",
	"rw.orth",
	"sc.orth",
	"sd.orth",
	"sg.orth",
	"sid.orth",
	"sn.orth",
	"su.orth",
	"ty.orth",
	"wal.orth",
	"za.orth",
	"lah.orth",
	"nqo.orth",
	"brx.orth",
	"sat.orth",
	"doi.orth",
	"mni.orth",
	"und_zsye.orth",
	"und_zmth.orth",
}
