package bitmap

import (
	"bytes"
	"io/ioutil"
	"os"
	"testing"
)

func TestParse(t *testing.T) {
	for _, file := range []string{
		"test/4x6.pcf",
		"test/8x16.pcf.gz",
		"test/charB18.pcf.gz",
		"test/courB18.pcf.gz",
		"test/hanglg16.pcf.gz",
		"test/helvB18.pcf.gz",
		"test/lubB18.pcf.gz",
		"test/ncenB18.pcf.gz",
		"test/orp-italic.pcf.gz",
		"test/timB18.pcf.gz",
		"test/timR24-ISO8859-1.pcf.gz",
		"test/timR24.pcf.gz",
	} {
		fi, err := os.Open(file)
		if err != nil {
			t.Fatal("can't read test file", err)
		}

		font, err := Parse(fi)
		if err != nil {
			t.Fatal(file, err)
		}

		font.LoadSummary()

		fs, err := Load(fi)
		if err != nil {
			t.Fatal(err)
		}
		if len(fs) != 1 {
			t.Error("expected one font")
		}
		fi.Close()
	}
}

func BenchmarkParse(b *testing.B) {
	for i := 0; i < b.N; i++ {
		for _, file := range []string{
			"test/4x6.pcf",
			"test/8x16.pcf.gz",
			"test/charB18.pcf.gz",
			"test/courB18.pcf.gz",
			"test/hanglg16.pcf.gz",
			"test/helvB18.pcf.gz",
			"test/lubB18.pcf.gz",
			"test/ncenB18.pcf.gz",
			"test/orp-italic.pcf.gz",
			"test/timB18.pcf.gz",
			"test/timR24-ISO8859-1.pcf.gz",
			"test/timR24.pcf.gz",
		} {
			fi, err := ioutil.ReadFile(file)
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
