package graphite

import (
	"bytes"
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// use a reference library to extensively test the shaping process

const referenceDir = "<XXX>/graphite/build/gr2fonttest"

// return stdout
func referenceShaping(t *testing.T, input shapingInput) []byte {
	fontFile, err := filepath.Abs(filepath.Join("testdata", input.fontfile))
	if err != nil {
		t.Fatal(err)
	}
	args := []string{fontFile, "-codes"}
	for _, r := range input.text {
		args = append(args, fmt.Sprintf("%04x", r))
	}
	if input.features != "" {
		args = append(args, "-feat", input.features)
	}
	if input.rtl {
		args = append(args, "-rtl")
	}
	cmd := exec.Command("./gr2fonttest", args...)
	cmd.Dir = referenceDir
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		t.Fatal(err)
	}
	return out.Bytes()
}

type aggregatedInput struct {
	runes    []rune
	features []string
}

func aggregateInputs(inputs []shapingInput) map[string]aggregatedInput {
	out := make(map[string]aggregatedInput)
	for _, input := range inputs {
		l := out[input.fontfile]
		l.runes = append(l.runes, input.text...)
		if input.features != "" {
			l.features = append(l.features, input.features)
		}
		out[input.fontfile] = l
	}
	return out
}

func randText(possible []rune, maxSize int) []rune {
	L := rand.Int31n(int32(maxSize)) + 1
	out := make([]rune, L)
	for i := range out {
		index := rand.Intn(len(possible))
		out[i] = possible[index]
	}
	return out
}

func fuzzReferenceShaping(possibles map[string]aggregatedInput, nbTry, maxInputSize int, t *testing.T) {
	var (
		failures  []shapingInput
		expecteds [][]byte
	)
	for fontFile, possible := range possibles {
		fmt.Print("shaping with font", fontFile, "...")
		for _, feature := range append(possible.features, "") {
			for _, rtl := range [2]bool{true, false} {
				in := shapingInput{
					fontfile: fontFile,
					features: feature,
					rtl:      rtl,
				}
				for i := 0; i < nbTry; i++ {
					in.text = randText(possible.runes, maxInputSize)
					expected := referenceShaping(t, in)

					if err := in.test(t, expected); err != nil {
						failures = append(failures, in)
						expecteds = append(expecteds, expected)
					}
				}
			}
		}
		fmt.Println(" done.")
	}

	L := len(fuzzTestInput)
	for i := range failures {
		failures[i].name = fmt.Sprintf("fuzz_%d", L+i)

		// store the expected output on disk
		expectedPath := "testdata/shape_refs/fuzz/" + failures[i].name + ".log"
		if err := os.WriteFile(expectedPath, expecteds[i], os.ModePerm); err != nil {
			t.Fatalf("can't write expected outputs: %s", err)
		}
	}
	if len(failures) != 0 {
		// dump the cases for further study
		fmt.Printf("%v\n", failures)
		t.Error("failures")
	}
}

// func TestGenerateFuzz(t *testing.T) {
// 	// Running this test use a reference binary to
// 	// generate output against random inputs,
// 	// and reports an error if our output is incorect.
// 	// It also records the corresponding .log file in testdata/shape_refs/fuzz

// 	// possibles := aggregateInputs(referenceFonttestInput)

// 	// extracted from testdata/texts/inputs/awami_tests.txt
// 	possibles := map[string]aggregatedInput{
// 		"AwamiNastaliq-Regular.ttf": {
// 			runes: []rune{
// 				10, 32, 45, 124, 1548, 1556, 1563, 1567, 1568, 1570, 1571, 1572, 1573,
// 				1574, 1575, 1576, 1577, 1578, 1579, 1580, 1581, 1582, 1583, 1584, 1585, 1586, 1587,
// 				1588, 1589, 1590, 1591, 1592, 1593, 1594, 1601, 1602, 1603, 1604, 1605, 1606, 1607,
// 				1608, 1610, 1611, 1612, 1613, 1614, 1615, 1616, 1617, 1618, 1619, 1620, 1643, 1644,
// 				1648, 1650, 1651, 1653, 1657, 1658, 1659, 1660, 1661, 1662, 1665, 1667, 1668, 1669,
// 				1670, 1671, 1672, 1673, 1674, 1680, 1681, 1683, 1686, 1687, 1688, 1689, 1690, 1691,
// 				1705, 1707, 1711, 1715, 1719, 1722, 1724, 1726, 1728, 1729, 1730, 1731, 1732, 1733,
// 				1734, 1735, 1740, 1741, 1744, 1746, 1747, 1748, 1757, 1770, 1776, 1777, 1778, 1779,
// 				1780, 1781, 1782, 1783, 1784, 1785, 1881, 1896, 1897, 1898, 1900, 1901, 8205, 8206, 64830, 64831,
// 			},
// 		},
// 	}
// 	fuzzReferenceShaping(possibles, 30, 30, t)
// }
