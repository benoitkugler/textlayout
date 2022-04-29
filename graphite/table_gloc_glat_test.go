package graphite

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"log"
	"math"
	"reflect"
	"strconv"
	"strings"
	"testing"

	testdata "github.com/benoitkugler/textlayout-testdata/graphite"
	"github.com/benoitkugler/textlayout/fonts/truetype"
)

func (oc *octaboxMetrics) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	var tmp struct {
		Bitmap   string          `xml:"bitmap,attr"`
		Subboxes []subboxMetrics `xml:"octabox"`
	}
	if err := d.DecodeElement(&tmp, &start); err != nil {
		return err
	}

	oc.subBbox = tmp.Subboxes

	if tmp.Bitmap == "" {
		return nil
	}
	bm, err := strconv.ParseUint(tmp.Bitmap, 16, 16)
	if err != nil {
		return fmt.Errorf("invalid bitmap attribute %s: %s", tmp.Bitmap, err)
	}
	oc.bitmap = uint16(bm)

	if err := readByteAttrAsPercent(start, &oc.diagPosMin, "diagPosMin"); err != nil {
		return err
	}
	if err := readByteAttrAsPercent(start, &oc.diagPosMax, "diagPosMax"); err != nil {
		return err
	}
	if err := readByteAttrAsPercent(start, &oc.diagNegMin, "diagNegMin"); err != nil {
		return err
	}
	if err := readByteAttrAsPercent(start, &oc.diagNegMax, "diagNegMax"); err != nil {
		return err
	}

	return nil
}

func readByteAttrAsPercent(start xml.StartElement, dst *byte, target string) error {
	for _, a := range start.Attr {
		if a.Name.Local == target {
			var f float64
			if _, err := fmt.Sscanf(a.Value, "%f%%", &f); err != nil {
				return err
			}
			*dst = byte(math.Round(f * 255 / 100))
		}
	}
	return nil
}

func (oc *subboxMetrics) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	if err := readByteAttrAsPercent(start, &oc.Left, "left"); err != nil {
		return err
	}
	if err := readByteAttrAsPercent(start, &oc.Right, "right"); err != nil {
		return err
	}
	if err := readByteAttrAsPercent(start, &oc.Bottom, "bottom"); err != nil {
		return err
	}
	if err := readByteAttrAsPercent(start, &oc.Top, "top"); err != nil {
		return err
	}
	if err := readByteAttrAsPercent(start, &oc.DiagPosMin, "diagPosMin"); err != nil {
		return err
	}
	if err := readByteAttrAsPercent(start, &oc.DiagPosMax, "diagPosMax"); err != nil {
		return err
	}
	if err := readByteAttrAsPercent(start, &oc.DiagNegMin, "diagNegMin"); err != nil {
		return err
	}
	if err := readByteAttrAsPercent(start, &oc.DiagNegMax, "diagNegMax"); err != nil {
		return err
	}

	return d.Skip()
}

// load a .ttx file, produced by fonttools
func readExpectedGlat(filename string) []struct {
	attributes map[uint16]int16
	metrics    *octaboxMetrics
} {
	data, err := testdata.Files.ReadFile(strings.ReplaceAll(filename, ".ttf", ".ttx"))
	if err != nil {
		log.Fatal(err)
	}

	type xmlDoc struct {
		Glyphs []struct {
			Name      string          `xml:"name,attr"`
			Octaboxes *octaboxMetrics `xml:"octaboxes"`
			Attrs     []struct {
				Index uint16 `xml:"index,attr"`
				Value int16  `xml:"value,attr"`
			} `xml:"attribute"`
		} `xml:"Glat>glyph"`
	}
	var doc xmlDoc
	err = xml.Unmarshal(data, &doc)
	if err != nil {
		log.Fatal(err)
	}

	out := make([]struct {
		attributes map[uint16]int16
		metrics    *octaboxMetrics
	}, len(doc.Glyphs))
	for i, glyph := range doc.Glyphs {
		m := make(map[uint16]int16)
		for _, attr := range glyph.Attrs {
			m[attr.Index] = attr.Value
		}
		out[i].attributes = m
		out[i].metrics = glyph.Octaboxes
	}

	return out
}

// graphite
var (
	tagGloc = truetype.MustNewTag("Gloc")
	tagGlat = truetype.MustNewTag("Glat")
)

func TestGlat(t *testing.T) {
	for _, filename := range []string{
		"Annapurnarc2.ttf",
		"Awami_test.ttf",
	} {
		f, err := testdata.Files.ReadFile(filename)
		if err != nil {
			t.Fatal(err)
		}

		font, err := truetype.NewFontParser(bytes.NewReader(f))
		if err != nil {
			t.Fatal(err)
		}
		ng, err := font.NumGlyphs()
		if err != nil {
			t.Fatal(err)
		}

		ta, err := font.GetRawTable(tagGloc)
		if err != nil {
			t.Fatal(err)
		}

		locations, _, err := parseTableGloc(ta, ng)
		if err != nil {
			t.Fatal(err)
		}

		ta, err = font.GetRawTable(tagGlat)
		if err != nil {
			t.Fatal(err)
		}

		attrs, err := parseTableGlat(ta, locations)
		if err != nil {
			t.Fatal(err)
		}
		expected := readExpectedGlat(filename)
		if len(attrs) != len(expected) {
			t.Errorf("wrong length")
		}

		for i, m := range expected {
			attrSet := attrs[i].attributes
			for k, v := range m.attributes {
				if g := attrSet.get(k); g != v {
					t.Errorf("expected %d, got %d", v, g)
				}
				if _, in := m.attributes[k+1]; !in {
					if attrSet.get(k+1) != 0 {
						t.Errorf("expected not found")
					}
				}
			}

			exp, got := m.metrics, attrs[i].octaboxMetrics
			if exp == nil {
				continue
			}
			if got == nil {
				t.Fatalf("missing metrics for glyph %d", i)
			}

			if len(exp.subBbox) == 0 {
				exp.subBbox = nil
			}
			if len(got.subBbox) == 0 {
				got.subBbox = nil
			}
			if !reflect.DeepEqual(*got, *exp) {
				t.Errorf("expected %v, got %v", *exp, *exp)
			}
		}
	}
}
