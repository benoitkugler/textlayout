package bitmap

import (
	"bytes"
	"testing"

	testdata "github.com/benoitkugler/textlayout-testdata/bitmap"
)

func TestParse(t *testing.T) {
	for _, file := range []string{
		"4x6.pcf",
		"8x16.pcf.gz",
		"charB18.pcf.gz",
		"courB18.pcf.gz",
		"hanglg16.pcf.gz",
		"helvB18.pcf.gz",
		"lubB18.pcf.gz",
		"ncenB18.pcf.gz",
		"orp-italic.pcf.gz",
		"timB18.pcf.gz",
		"timR24-ISO8859-1.pcf.gz",
		"timR24.pcf.gz",
	} {
		fi, err := testdata.Files.ReadFile(file)
		if err != nil {
			t.Fatal("can't read test file", err)
		}

		font, err := Parse(bytes.NewReader(fi))
		if err != nil {
			t.Fatal(file, err)
		}

		font.LoadSummary()

		fs, err := Load(bytes.NewReader(fi))
		if err != nil {
			t.Fatal(err)
		}
		if len(fs) != 1 {
			t.Error("expected one font")
		}
	}
}

func BenchmarkParse(b *testing.B) {
	for i := 0; i < b.N; i++ {
		for _, file := range []string{
			"4x6.pcf",
			"8x16.pcf.gz",
			"charB18.pcf.gz",
			"courB18.pcf.gz",
			"hanglg16.pcf.gz",
			"helvB18.pcf.gz",
			"lubB18.pcf.gz",
			"ncenB18.pcf.gz",
			"orp-italic.pcf.gz",
			"timB18.pcf.gz",
			"timR24-ISO8859-1.pcf.gz",
			"timR24.pcf.gz",
		} {
			fi, err := testdata.Files.ReadFile(file)
			if err != nil {
				b.Fatal("can't read test file", err)
			}

			_, err = Parse(bytes.NewReader(fi))
			if err != nil {
				b.Fatal(file, err)
			}

		}
	}
}
