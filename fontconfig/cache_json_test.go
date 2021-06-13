package fontconfig

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"testing"
)

// serialisation of pattern using JSON

// add type information
func (v valueElt) MarshalJSON() ([]byte, error) {
	tmp := struct {
		Value   interface{}  `json:"v"`
		Binding valueBinding `json:"b"`
		Type    uint8        `json:"t"`
	}{Value: v.Value, Binding: v.Binding, Type: v.dataType()}
	return json.Marshal(tmp)
}

func (v *valueElt) UnmarshalJSON(data []byte) error {
	var tmp struct {
		Value   json.RawMessage `json:"v"`
		Binding valueBinding    `json:"b"`
		Type    uint8           `json:"t"`
	}
	err := json.Unmarshal(data, &tmp)
	if err != nil {
		return err
	}
	switch tmp.Type {
	case tInt:
		var out Int
		err = json.Unmarshal([]byte(tmp.Value), &out)
		v.Value = out
	case tFloat:
		var out Float
		err = json.Unmarshal([]byte(tmp.Value), &out)
		v.Value = out
	case tString:
		var out String
		err = json.Unmarshal([]byte(tmp.Value), &out)
		v.Value = out
	case tBool:
		var out Bool
		err = json.Unmarshal([]byte(tmp.Value), &out)
		v.Value = out
	case tCharset:
		var out Charset
		err = json.Unmarshal([]byte(tmp.Value), &out)
		v.Value = out
	case tLangset:
		var out Langset
		err = json.Unmarshal([]byte(tmp.Value), &out)
		v.Value = out
	case tMatrix:
		var out Matrix
		err = json.Unmarshal([]byte(tmp.Value), &out)
		v.Value = out
	case tRange:
		var out Range
		err = json.Unmarshal([]byte(tmp.Value), &out)
		v.Value = out
	default:
		return nil
	}

	v.Binding = tmp.Binding
	return err
}

type publicCharsetJSON struct {
	PageNumbers []uint16
	Pages       []charPage
}

func (c Charset) MarshalJSON() ([]byte, error) {
	pc := publicCharsetJSON{PageNumbers: c.pageNumbers, Pages: c.pages}
	return json.Marshal(pc)
}

func (c *Charset) UnmarshalJSON(data []byte) error {
	var pc publicCharsetJSON
	err := json.Unmarshal(data, &pc)
	c.pageNumbers = pc.PageNumbers
	c.pages = pc.Pages
	return err
}

type publicLangsetJSON struct {
	Extra strSet
	Page  [langPageSize]uint32
}

func (c Langset) MarshalJSON() ([]byte, error) {
	pc := publicLangsetJSON{Extra: c.extra, Page: c.page}
	return json.Marshal(pc)
}

func (c *Langset) UnmarshalJSON(data []byte) error {
	var pc publicLangsetJSON
	err := json.Unmarshal(data, &pc)
	c.extra = pc.Extra
	c.page = pc.Page
	return err
}

// Serialize serialise the content of the font set (using gob and gzip).
// Since scanning the fonts is rather slow, this methods can be used in
// conjonction with `LoadFontset` to cache the result of a scan.
func (fs Fontset) SerializeJSON(w io.Writer) error {
	gzipWr := gzip.NewWriter(w)
	gw := json.NewEncoder(gzipWr)
	err := gw.Encode(fs)
	if err != nil {
		return fmt.Errorf("internal error: can't encode fontset: %s", err)
	}
	err = gzipWr.Close()
	if err != nil {
		return fmt.Errorf("internal error: can't compress fonset dump: %s", err)
	}
	return nil
}

// LoadFontset reads a cache file, exported by the `Fontset.Serialize` method,
// and constructs the associated font set.
func LoadFontsetJSON(r io.Reader) (Fontset, error) {
	gzipR, err := gzip.NewReader(r)
	if err != nil {
		return nil, fmt.Errorf("invalid fontconfig compressed dump file: %s", err)
	}
	gr := json.NewDecoder(gzipR)
	var out Fontset
	err = gr.Decode(&out)
	if err != nil {
		return nil, fmt.Errorf("invalid fontconfig dump file format: %s", err)
	}
	return out, nil
}

func TestCacheJSON(t *testing.T) {
	fs := randPatterns(100)

	b, err := json.Marshal(fs)
	if err != nil {
		t.Fatal(err)
	}

	var back Fontset
	err = json.Unmarshal(b, &back)
	if err != nil {
		t.Fatal(err)
	}

	for i := range back {
		if fs[i].Hash() != back[i].Hash() {
			t.Fatal("hash not preserved")
		}
	}
}
