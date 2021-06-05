package harfbuzz

// This file use a reference Harfbuzz binary to compare outputs and record fails

import (
	"bytes"
	"fmt"
	"math/rand"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// use a reference library to extensively test the shaping process

const referenceDir = "<XXX>/harfbuzz"

type fontID struct {
	file  string
	index int
}

type shapingInput struct {
	font     fontID
	features string
	text     []rune
}

func (sh shapingInput) testOptions() testOptions {
	// var out testOptions
	out := parseOptions("") // default values
	out.input.text = sh.text
	out.shaper.features = sh.features
	out.fontOpts.fontRef = sh.font
	return out
}

// return stdout
func referenceShaping(t *testing.T, input shapingInput) string {
	fontFile, err := filepath.Abs(input.font.file)
	if err != nil {
		t.Fatal(err)
	}
	args := []string{fontFile, fmt.Sprintf("--face-index=%d", input.font.index), "-u"}
	var runes []string
	for _, r := range input.text {
		runes = append(runes, fmt.Sprintf("0x%04x", r))
	}
	args = append(args, strings.Join(runes, ","))
	if input.features != "" {
		args = append(args, "--features="+input.features)
	}
	cmd := exec.Command("builddir/util/hb-shape", args...)
	cmd.Dir = referenceDir
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		t.Fatal(err)
	}
	return strings.TrimSpace(out.String())
}

type aggregatedInput struct {
	runes    []rune   // all possibles runes
	features []string // all possibles features
}

func aggregateInputs(t *testing.T) map[fontID]aggregatedInput {
	out := make(map[fontID]aggregatedInput)

	walkShapeTests(t, func(_ *testing.T, driver testOptions, _, _, glyphsExpected string) {
		if glyphsExpected == "*" {
			return
		}
		l := out[driver.fontOpts.fontRef]
		l.runes = append(l.runes, driver.input.text...)
		if driver.shaper.features != "" {
			l.features = append(l.features, driver.shaper.features)
		}
		out[driver.fontOpts.fontRef] = l
	})

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

func fuzzReferenceShaping(possibles map[fontID]aggregatedInput, nbTry, maxInputSize int, t *testing.T) {
	var (
		failures  []shapingInput
		expecteds []string
		gots      []string
	)
	for fontFile, possible := range possibles {
		fmt.Print("Shaping with font", fontFile, "...")
		for _, feature := range append(possible.features, "") {
			in := shapingInput{
				font:     fontFile,
				features: feature,
			}
			for i := 0; i < nbTry; i++ {
				in.text = randText(possible.runes, maxInputSize)
				expected := referenceShaping(t, in)

				// some tests font pass the verify
				// since we compare to harfbuzz output it is redondant anyway
				got, err := in.testOptions().shape(false)
				if err != nil {
					t.Fatal(err)
				}

				if expected != got {
					failures = append(failures, in)
					expecteds = append(expecteds, expected)
					gots = append(gots, got)
				}
			}
		}
		fmt.Println(" done.")
	}

	if len(failures) != 0 {
		// dump the cases for further study
		fmt.Printf("\n%#v\n", failures)
		fmt.Printf("\n%#v\n", expecteds)
		fmt.Printf("\n%#v\n", gots)
		t.Errorf("%d failures happened", len(failures))
	}
}

// func TestReference(t *testing.T) {
// 	out := referenceShaping(t, shapingInput{fontID{"testdata/harfbuzz_reference/aots/fonts/gsub4_1_multiple_ligsets_f1.otf", 0}, "", []rune{21, 21, 22, 19}})
// 	fmt.Println(out)
// }

// func TestGenerateFuzz(t *testing.T) {
// 	// Running this test use a reference binary to
// 	// generate output against random inputs,
// 	// and reports an error if our output is incorect.

// 	possibles := aggregateInputs(t)

// 	fuzzReferenceShaping(possibles, 10, 20, t)
// }
