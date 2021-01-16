package bitmap

import (
	"bytes"
	"compress/gzip"
	"io/ioutil"
	"testing"
)

func TestParse(t *testing.T) {
	for _, file := range []string{
		"test/4x6.pcf.gz",
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
		b, err := ioutil.ReadFile(file)
		if err != nil {
			t.Fatal("can't read test file", err)
		}

		r, err := gzip.NewReader(bytes.NewReader(b))
		if err != nil {
			t.Fatal("invalid gzip file", err)
		}

		b, err = ioutil.ReadAll(r)
		if err != nil {
			t.Fatal("invalid gzip file", err)
		}

		font, err := parse(b)
		if err != nil {
			t.Fatal(err)
		}

		font.Style()
	}
}
