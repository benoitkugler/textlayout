package fribidi

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"
	"testing"
	"time"
)

func parseOrdering(line string) ([]int, error) {
	fields := strings.Fields(line)
	out := make([]int, len(fields))
	for i, posLit := range fields {
		pos, err := strconv.Atoi(posLit)
		if err != nil {
			return nil, fmt.Errorf("invalid position %s: %s", posLit, err)
		}
		out[i] = pos
	}
	return out, nil
}

func parseLevels(line string) ([]Level, error) {
	fields := strings.Fields(line)
	out := make([]Level, len(fields))
	for i, f := range fields {
		if f == "x" {
			out[i] = -1
		} else {
			lev, err := strconv.Atoi(f)
			if err != nil {
				return nil, fmt.Errorf("invalid level %s: %s", f, err)
			}
			out[i] = Level(lev)
		}
	}
	return out, nil
}

type testData struct {
	codePoints       []rune
	expectedLevels   []Level
	visualOrdering   []int
	parDir           int
	resolvedParLevel int
}

func parseTestLine(line []byte) (out testData, err error) {
	fields := strings.Split(string(line), ";")
	if len(fields) < 5 {
		return out, fmt.Errorf("invalid line %s", line)
	}

	//  Field 0. Code points
	for _, runeLit := range strings.Fields(fields[0]) {
		var c rune
		if _, err = fmt.Sscanf(runeLit, "%04x", &c); err != nil {
			return out, fmt.Errorf("invalid rune %s: %s", runeLit, err)
		}
		out.codePoints = append(out.codePoints, c)
	}

	// Field 1. Paragraph direction
	out.parDir, err = strconv.Atoi(fields[1])
	if err != nil {
		return out, fmt.Errorf("invalid paragraph direction %s: %s", fields[1], err)
	}

	// Field 2. resolved paragraph_dir
	out.resolvedParLevel, err = strconv.Atoi(fields[2])
	if err != nil {
		return out, fmt.Errorf("invalid resolved paragraph embedding level %s: %s", fields[2], err)
	}

	// Field 3. resolved levels (or -1)
	out.expectedLevels, err = parseLevels(fields[3])
	if err != nil {
		return out, err
	}

	if len(out.expectedLevels) != len(out.codePoints) {
		return out, errors.New("different lengths for levels and codepoints")
	}

	//  Field 4 - resulting visual ordering
	out.visualOrdering, err = parseOrdering(fields[4])

	return out, err
}

func parseSimpleBidi() ([]testData, error) {
	const filename = "test/unicode-conformance/BidiCharacterTest.txt"

	b, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	var out []testData
	for lineNumber, line := range bytes.Split(b, []byte{'\n'}) {
		if len(line) == 0 || line[0] == '#' || line[0] == '\n' {
			continue
		}

		lineData, err := parseTestLine(line)
		if err != nil {
			return nil, fmt.Errorf("invalid line %d: %s", lineNumber+1, err)
		}
		out = append(out, lineData)
	}
	return out, nil
}

func runOneSimpleBidi(lineData testData) ([]Level, []int) {
	types := getBidiTypes(lineData.codePoints)
	bracketTypes := getBracketTypes(lineData.codePoints, types)

	var baseDir ParType
	switch lineData.parDir {
	case 0:
		baseDir = LTR
	case 1:
		baseDir = RTL
	case 2:
		baseDir = ON
	}

	levels, _ := GetParEmbeddingLevels(types, bracketTypes, &baseDir)

	ltor := make([]int, len(lineData.codePoints))
	for i := range ltor {
		ltor[i] = i
	}

	ReorderLine(0 /*FRIBIDI_FLAG_REORDER_NSM*/, types, len(types), 0, baseDir, levels, nil, ltor)

	j := 0
	for _, lr := range ltor {
		if !types[lr].isExplicitOrBn() {
			ltor[j] = lr
			j++
		}
	}
	ltor = ltor[0:j] // slice to length

	return levels, ltor
}

func TestBidiCharacters(t *testing.T) {
	ti := time.Now()

	datas, err := parseSimpleBidi()
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println("input data parsed in", time.Since(ti))
	ti = time.Now()

	for _, lineData := range datas {
		levels, ltor := runOneSimpleBidi(lineData)

		/* Compare */
		for i, level := range levels {
			if exp := lineData.expectedLevels[i]; level != exp && exp != -1 {
				t.Fatalf("failure: levels[%d]: expected %d, got %d", i, exp, level)
				break
			}
		}

		if len(lineData.visualOrdering) != len(ltor) {
			t.Fatalf("failure visual ordering: got %v, expected %v", ltor, lineData.visualOrdering)
		}
		for i := range ltor {
			if lineData.visualOrdering[i] != ltor[i] {
				t.Fatalf("failure visual ordering: got %v, expected %v", ltor, lineData.visualOrdering)
			}
		}
	}

	fmt.Println("test run in", time.Since(ti))
}

func parseCharType(s string) (CharType, error) {
	switch s {
	case "L":
		return LTR, nil
	case "R":
		return RTL, nil
	case "AL":
		return AL, nil
	case "EN":
		return EN, nil
	case "AN":
		return AN, nil
	case "ES":
		return ES, nil
	case "ET":
		return ET, nil
	case "CS":
		return CS, nil
	case "NSM":
		return NSM, nil
	case "BN":
		return BN, nil
	case "B":
		return BS, nil
	case "S":
		return SS, nil
	case "WS":
		return WS, nil
	case "ON":
		return ON, nil
	case "LRE":
		return LRE, nil
	case "RLE":
		return RLE, nil
	case "LRO":
		return LRO, nil
	case "RLO":
		return RLO, nil
	case "PDF":
		return PDF, nil
	case "LRI":
		return LRI, nil
	case "RLI":
		return RLI, nil
	case "FSI":
		return FSI, nil
	case "PDI":
		return PDI, nil
	default:
		return 0, fmt.Errorf("invalid char type %s", s)
	}
}

func parseLevelsLine(line string) ([]Level, error) {
	line = strings.TrimPrefix(line, "@Levels:")
	return parseLevels(line)
}

func parseReorderLine(line string) ([]int, error) {
	line = strings.TrimPrefix(line, "@Reorder:")
	return parseOrdering(line)
}

func parseCharsLine(line string) ([]CharType, int, error) {
	fields := strings.Split(line, ";")
	if len(fields) != 2 {
		return nil, 0, fmt.Errorf("invalid line: %s", line)
	}
	var err error
	chars := strings.Fields(fields[0])
	out := make([]CharType, len(chars))
	for i, cs := range chars {
		out[i], err = parseCharType(cs)
		if err != nil {
			return nil, 0, err
		}
	}
	baseDirFlags, err := strconv.Atoi(strings.TrimSpace(fields[1]))
	return out, baseDirFlags, err
}

type oneBidiData struct {
	types       []CharType
	baseDirFlag int
}

type bidiTest struct {
	ltor   []int
	levels []Level
	data   []oneBidiData
}

func parseBidi() ([]bidiTest, error) {
	const filename = "test/unicode-conformance/BidiTest.txt"

	b, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	var (
		out     []bidiTest
		current bidiTest
	)
	for lineNumber, lineB := range bytes.Split(b, []byte{'\n'}) {
		line := string(lineB)
		if len(line) == 0 || line[0] == '#' {
			// flush the current datas
			if len(current.data) != 0 {
				out = append(out, current)
				current.data = nil
			}
			continue
		}

		if strings.HasPrefix(line, "@Reorder:") {
			current.ltor, err = parseReorderLine(line)
			if err != nil {
				return nil, fmt.Errorf("invalid  line %d: %s", lineNumber+1, err)
			}
			continue
		} else if strings.HasPrefix(line, "@Levels:") {
			current.levels, err = parseLevelsLine(line)
			if err != nil {
				return nil, fmt.Errorf("invalid line %d: %s", lineNumber+1, err)
			}
			continue
		}

		/* Test line */
		var lineData oneBidiData
		lineData.types, lineData.baseDirFlag, err = parseCharsLine(line)
		if err != nil {
			return nil, fmt.Errorf("invalid line %d: %s", lineNumber+1, err)
		}
		current.data = append(current.data, lineData)
	}
	return out, nil
}

func runOneComplexBidi(data bidiTest) (levelsList [][]Level, ltorList [][]int) {
	for _, line := range data.data {
		for baseDirMode := 0; baseDirMode < 3; baseDirMode++ {

			if (line.baseDirFlag & (1 << baseDirMode)) == 0 {
				continue
			}

			var baseDir ParType
			switch baseDirMode {
			case 0:
				baseDir = ON
			case 1:
				baseDir = LTR
			case 2:
				baseDir = RTL
			}

			// Brackets are not used in the BidiTest.txt file
			levels, _ := GetParEmbeddingLevels(line.types, nil, &baseDir)

			ltor := make([]int, len(levels))
			for i := range ltor {
				ltor[i] = i
			}

			ReorderLine(0 /*FRIBIDI_FLAG_REORDER_NSM*/, line.types, len(line.types),
				0, baseDir, levels,
				nil, ltor)

			j := 0
			for _, lr := range ltor {
				if !line.types[lr].isExplicitOrBn() {
					ltor[j] = lr
					j++
				}
			}
			ltor = ltor[0:j] // slice to length

			levelsList = append(levelsList, levels)
			ltorList = append(ltorList, ltor)
		}
	}
	return
}

func TestBidi(t *testing.T) {
	ti := time.Now()
	datas, err := parseBidi()
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println("parsed BidiTest.txt in", time.Since(ti))
	ti = time.Now()

	for index, data := range datas {
		/* Test it */
		levelsList, ltorList := runOneComplexBidi(data)

		/* Compare */
		for j := range levelsList {
			levels, ltor := levelsList[j], ltorList[j]

			for i, level := range levels {
				if exp := data.levels[i]; level != exp && exp != -1 {
					t.Fatalf("failure on test %d: levels[%d]: expected %d, got %d", index+1, i, exp, level)
					break
				}
			}

			if len(data.ltor) != len(ltor) {
				t.Fatalf("failure on test %d: visual ordering: got %v, expected %v", index+1, ltor, data.ltor)
			}
			for i := range ltor {
				if data.ltor[i] != ltor[i] {
					t.Fatalf("failure on test %d: visual ordering: got %v, expected %v", index+1, ltor, data.ltor)
				}
			}
		}
	}

	fmt.Println("test run in", time.Since(ti))
}

func BenchmarkSimple(b *testing.B) {
	datas, err := parseSimpleBidi()
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		for _, lineData := range datas {
			runOneSimpleBidi(lineData)
		}
	}
}

func BenchmarkComplex(b *testing.B) {
	datas, err := parseBidi()
	if err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, lineData := range datas {
			runOneComplexBidi(lineData)
		}
	}
}
