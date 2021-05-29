package graphite

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// use a reference library to extensively test the shaping process

const referenceDir = "/home/benoit/Téléchargements/graphite/gr2fonttest"

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

func randText(possible []rune) []rune {
	const maxSize = 5
	L := rand.Int31n(maxSize) + 1
	out := make([]rune, L)
	for i := range out {
		index := rand.Intn(len(possible))
		out[i] = possible[index]
	}
	return out
}

func fuzzReferenceShaping(t *testing.T) {
	possibles := aggregateInputs(referenceFonttestInput)

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
				for range [2]bool{} {
					in.text = randText(possible.runes)
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
		if err := ioutil.WriteFile(expectedPath, expecteds[i], os.ModePerm); err != nil {
			t.Fatalf("can't write expected outputs: %s", err)
		}
	}
	if len(failures) != 0 {
		// dump the cases for further study
		fmt.Printf("%v\n", failures)
		t.Error("failures")
	}
}

func TestGenerateFuzz(t *testing.T) {
	fuzzReferenceShaping(t)
}
