package fontconfig

import (
	"bytes"
	"compress/gzip"
	"encoding/gob"
	"fmt"
	"io"
)

func init() {
	gob.Register(Float(1))
	gob.Register(Int(1))
	gob.Register(String(""))
	gob.Register(Bool(0))
	gob.Register(Charset{})
	gob.Register(Langset{})
	gob.Register(Matrix{})
	gob.Register(Range{})
}

type publicCharset struct {
	PageNumbers []uint16
	Pages       []charPage
}

func (c Charset) GobEncode() ([]byte, error) {
	pc := publicCharset{PageNumbers: c.pageNumbers, Pages: c.pages}
	var b bytes.Buffer
	err := gob.NewEncoder(&b).Encode(pc)
	return b.Bytes(), err
}

func (c *Charset) GobDecode(data []byte) error {
	var pc publicCharset
	err := gob.NewDecoder(bytes.NewReader(data)).Decode(&pc)
	c.pageNumbers = pc.PageNumbers
	c.pages = pc.Pages
	return err
}

type publicLangset struct {
	Extra strSet
	Page  [langPageSize]uint32
}

func (c Langset) GobEncode() ([]byte, error) {
	pc := publicLangset{Extra: c.extra, Page: c.page}
	var b bytes.Buffer
	err := gob.NewEncoder(&b).Encode(pc)
	return b.Bytes(), err
}

func (c *Langset) GobDecode(data []byte) error {
	var pc publicLangset
	err := gob.NewDecoder(bytes.NewReader(data)).Decode(&pc)
	c.extra = pc.Extra
	c.page = pc.Page
	return err
}

// Dump serialise the content of the font set (using gob and gzip),
// making caching possible.
func (fs FontSet) Dump(w io.Writer) error {
	gzipWr := gzip.NewWriter(w)
	gw := gob.NewEncoder(gzipWr)
	err := gw.Encode(fs)
	if err != nil {
		return fmt.Errorf("internal error: can't encode fontset: %s", err)
	}
	err = gzipWr.Close()
	if err != nil {
		return fmt.Errorf("internal error: can't compress dump: %s", err)
	}
	return nil
}

// LoadFontSet reads a cache file, exported by the `Dump` method,
// and constructs the associated font set.
func LoadFontSet(r io.Reader) (FontSet, error) {
	gzipR, err := gzip.NewReader(r)
	if err != nil {
		return nil, fmt.Errorf("invalid fontconfig compressed dump file: %s", err)
	}
	gr := gob.NewDecoder(gzipR)
	var out FontSet
	err = gr.Decode(&out)
	if err != nil {
		return nil, fmt.Errorf("invalid fontconfig dump file format: %s", err)
	}
	return out, nil
}
