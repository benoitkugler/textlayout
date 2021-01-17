package truetype

import "testing"

func TestParseVariations(t *testing.T) {
	datas := [...]struct {
		input    string
		expected Variation
	}{
		{" frea=45.78", Variation{Tag: MustNewTag("frea"), Value: 45.78}},
		{"G45E=45", Variation{Tag: MustNewTag("G45E"), Value: 45}},
		{"fAAD 45.78", Variation{Tag: MustNewTag("fAAD"), Value: 45.78}},
		{"fr 45.78", Variation{Tag: MustNewTag("fr  "), Value: 45.78}},
		{"fr=45.78", Variation{Tag: MustNewTag("fr  "), Value: 45.78}},
		{"fr=-45.4", Variation{Tag: MustNewTag("fr  "), Value: -45.4}},
		{"'fr45'=-45.4", Variation{Tag: MustNewTag("fr45"), Value: -45.4}}, // with quotes
		{`"frZD"=-45.4`, Variation{Tag: MustNewTag("frZD"), Value: -45.4}}, // with quotes
	}
	for _, data := range datas {
		out, err := NewVariation(data.input)
		if err != nil {
			t.Fatalf("error on %s: %s", data.input, err)
		}
		if out != data.expected {
			t.Fatalf("for %s, expected %v, got %v", data.input, data.expected, out)
		}
	}
}
