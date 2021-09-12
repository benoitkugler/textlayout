package fontconfig

import (
	"bytes"
	"compress/gzip"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
)

// encoded value type
const (
	tUnknown = iota
	tInt
	tFloat
	tString
	tBool
	tCharset
	tLangset
	tMatrix
	tRange
)

func (v valueElt) dataType() uint8 {
	switch v.Value.(type) {
	case Int:
		return tInt
	case Float:
		return tFloat
	case String:
		return tString
	case Bool:
		return tBool
	case Charset:
		return tCharset
	case Langset:
		return tLangset
	case Matrix:
		return tMatrix
	case Range:
		return tRange
	default:
		return tUnknown
	}
}

// return the pattern slice without its size
func (p Pattern) serializeBin() ([]byte, error) {
	var (
		w   bytes.Buffer
		buf [2]byte
	)
	binary.BigEndian.PutUint16(buf[:], uint16(len(p)))
	_, err := w.Write(buf[:])
	if err != nil {
		return nil, err
	}
	for obj, list := range p {
		binary.BigEndian.PutUint16(buf[:], uint16(obj))
		_, err := w.Write(buf[:])
		if err != nil {
			return nil, err
		}
		if err := list.serializeBin(&w); err != nil {
			return nil, err
		}
	}
	return w.Bytes(), nil
}

func (l valueList) serializeBin(w io.Writer) error {
	var buf [2]byte
	binary.BigEndian.PutUint16(buf[:], uint16(len(l)))
	_, err := w.Write(buf[:])
	if err != nil {
		return err
	}
	for _, elem := range l {
		if err := elem.serializeBin(w); err != nil {
			return err
		}
	}
	return nil
}

func (e valueElt) serializeBin(w io.Writer) error {
	buf := [2]byte{byte(e.Binding), e.dataType()}
	_, err := w.Write(buf[:])
	if err != nil {
		return err
	}
	err = e.Value.serializeBin(w)
	return err
}

func (v Int) serializeBin(w io.Writer) error {
	var buf [4]byte
	binary.BigEndian.PutUint32(buf[:], uint32(v))
	_, err := w.Write(buf[:])
	return err
}

func (v Float) serializeBin(w io.Writer) error {
	var buf [4]byte
	binary.BigEndian.PutUint32(buf[:], math.Float32bits(float32(v)))
	_, err := w.Write(buf[:])
	return err
}

func (v String) serializeBin(w io.Writer) error {
	buffer := make([]byte, 4+len(v)) // len as uint32 + data
	binary.BigEndian.PutUint32(buffer, uint32(len(v)))
	copy(buffer[4:], v)
	_, err := w.Write(buffer)
	return err
}

func (v Bool) serializeBin(w io.Writer) error {
	_, err := w.Write([]byte{byte(v)})
	return err
}

func (v Charset) serializeBin(w io.Writer) error {
	buffer := make([]byte, 2+charsetRecordSize*len(v.pageNumbers))
	binary.BigEndian.PutUint16(buffer, uint16(len(v.pageNumbers)))
	for i, nb := range v.pageNumbers {
		binary.BigEndian.PutUint16(buffer[2+charsetRecordSize*i:], nb)
		slice := buffer[2+charsetRecordSize*i+2:]
		for j, k := range v.pages[i] {
			binary.BigEndian.PutUint32(slice[4*j:], k)
		}
	}
	_, err := w.Write(buffer)
	return err
}

func (v Langset) serializeBin(w io.Writer) error {
	// like font config, we dont serialize extra languages
	// which should not been found in fonts
	buffer := make([]byte, 4*langPageSize)
	for j, v := range v.page {
		binary.BigEndian.PutUint32(buffer[4*j:], v)
	}
	_, err := w.Write(buffer)
	return err
}

func (v Matrix) serializeBin(w io.Writer) error {
	var buf [4 * 4]byte
	binary.BigEndian.PutUint32(buf[:], math.Float32bits(v.Xx))
	binary.BigEndian.PutUint32(buf[4:], math.Float32bits(v.Xy))
	binary.BigEndian.PutUint32(buf[8:], math.Float32bits(v.Yx))
	binary.BigEndian.PutUint32(buf[12:], math.Float32bits(v.Yy))
	_, err := w.Write(buf[:])
	return err
}

func (v Range) serializeBin(w io.Writer) error {
	var buf [2 * 4]byte
	binary.BigEndian.PutUint32(buf[:], math.Float32bits(v.Begin))
	binary.BigEndian.PutUint32(buf[4:], math.Float32bits(v.End))
	_, err := w.Write(buf[:])
	return err
}

func (v *Int) deserializeBin(data []byte) (int, error) {
	if len(data) < 4 {
		return 0, errors.New("invalid Int (EOF)")
	}
	*v = Int(binary.BigEndian.Uint32(data))
	return 4, nil
}

func (v *Float) deserializeBin(data []byte) (int, error) {
	if len(data) < 4 {
		return 0, errors.New("invalid Float (EOF)")
	}
	*v = Float(math.Float32frombits(binary.BigEndian.Uint32(data)))
	return 4, nil
}

func (v *String) deserializeBin(data []byte) (int, error) {
	if len(data) < 4 {
		return 0, errors.New("invalid String (EOF)")
	}
	L := int(binary.BigEndian.Uint32(data))
	if len(data) < 4+L {
		return 0, errors.New("invalid String length (EOF)")
	}
	*v = String(data[4 : 4+L])
	return 4 + L, nil
}

func (v *Bool) deserializeBin(data []byte) (int, error) {
	if len(data) < 1 {
		return 0, errors.New("invalid Bool (EOF)")
	}
	*v = Bool(data[0])
	return 1, nil
}

const charsetRecordSize = 2 + 8*4

func (v *Charset) deserializeBin(data []byte) (int, error) {
	if len(data) < 2 {
		return 0, errors.New("invalid Charset (EOF)")
	}
	L := int(binary.BigEndian.Uint16(data))
	if len(data) < 2+charsetRecordSize*L {
		return 0, errors.New("invalid Charset size (EOF)")
	}
	v.pageNumbers = make([]uint16, L)
	v.pages = make([]charPage, L)

	for i := range v.pageNumbers {
		v.pageNumbers[i] = binary.BigEndian.Uint16(data[2+charsetRecordSize*i:])
		slice := data[2+charsetRecordSize*i+2:]
		for j := range v.pages[i] {
			v.pages[i][j] = binary.BigEndian.Uint32(slice[4*j:])
		}
	}

	return 2 + charsetRecordSize*L, nil
}

func (v *Langset) deserializeBin(data []byte) (int, error) {
	// like font config, we dont serialize extra languages
	// which should not been found in fonts
	if len(data) < 4*langPageSize {
		return 0, errors.New("invalid Langset (EOF)")
	}
	for j := range v.page {
		v.page[j] = binary.BigEndian.Uint32(data[4*j:])
	}
	return 4 * langPageSize, nil
}

func (v *Matrix) deserializeBin(data []byte) (int, error) {
	if len(data) < 4*4 {
		return 0, errors.New("invalid Matrix (EOF)")
	}
	v.Xx = math.Float32frombits(binary.BigEndian.Uint32(data[:]))
	v.Xy = math.Float32frombits(binary.BigEndian.Uint32(data[4:]))
	v.Yx = math.Float32frombits(binary.BigEndian.Uint32(data[8:]))
	v.Yy = math.Float32frombits(binary.BigEndian.Uint32(data[12:]))
	return 4 * 4, nil
}

func (v *Range) deserializeBin(data []byte) (int, error) {
	if len(data) < 2*4 {
		return 0, errors.New("invalid Range (EOF)")
	}
	v.Begin = math.Float32frombits(binary.BigEndian.Uint32(data[:]))
	v.End = math.Float32frombits(binary.BigEndian.Uint32(data[4:]))
	return 2 * 4, nil
}

func (v *valueElt) deserializeBin(data []byte) (int, error) {
	if len(data) < 2 {
		return 0, errors.New("invalid value (EOF)")
	}
	v.Binding = valueBinding(data[0])
	type_ := data[1]
	var (
		valueSize int
		err       error
	)
	switch type_ { // note that we dont want to store a pointer as Value
	case tInt:
		var out Int
		valueSize, err = out.deserializeBin(data[2:])
		v.Value = out
	case tFloat:
		var out Float
		valueSize, err = out.deserializeBin(data[2:])
		v.Value = out
	case tString:
		var out String
		valueSize, err = out.deserializeBin(data[2:])
		v.Value = out
	case tBool:
		var out Bool
		valueSize, err = out.deserializeBin(data[2:])
		v.Value = out
	case tCharset:
		var out Charset
		valueSize, err = out.deserializeBin(data[2:])
		v.Value = out
	case tLangset:
		var out Langset
		valueSize, err = out.deserializeBin(data[2:])
		v.Value = out
	case tMatrix:
		var out Matrix
		valueSize, err = out.deserializeBin(data[2:])
		v.Value = out
	case tRange:
		var out Range
		valueSize, err = out.deserializeBin(data[2:])
		v.Value = out
	}

	return 2 + valueSize, err
}

func deserializeValueListBin(data []byte) (valueList, int, error) {
	if len(data) < 2 {
		return nil, 0, errors.New("invalid value list (EOF)")
	}
	L := int(binary.BigEndian.Uint16(data))
	list := make(valueList, L)
	offset := 2
	for i := range list {
		if len(data) < offset {
			return nil, 0, errors.New("invalid value list size (EOF)")
		}
		read, err := list[i].deserializeBin(data[offset:])
		if err != nil {
			return nil, 0, fmt.Errorf("invalid value list: %s", err)
		}
		offset += read
	}

	return list, offset, nil
}

func deserializePatternBin(data []byte) (Pattern, error) {
	if len(data) < 2 {
		return nil, errors.New("invalid pattern (EOF)")
	}
	L := int(binary.BigEndian.Uint16(data))
	pat := make(Pattern, L)
	offset := 2
	for i := 0; i < L; i++ {
		if len(data) < offset+2 {
			return nil, errors.New("invalid pattern entry (EOF)")
		}
		obj := Object(binary.BigEndian.Uint16(data[offset:]))
		offset += 2
		list, read, err := deserializeValueListBin(data[offset:])
		if err != nil {
			return nil, fmt.Errorf("invalid pattern: %s", err)
		}
		(pat)[obj] = &list
		offset += read
	}
	return pat, nil
}

// Serialize serialise the content of the font set (using a custom binary encoding and gzip).
// Since scanning the fonts is rather slow, this methods can be used in
// conjonction with `LoadFontset` to cache the result of a scan.
func (fs Fontset) Serialize(dst io.Writer) error {
	w := gzip.NewWriter(dst)

	var buf [4]byte
	binary.BigEndian.PutUint32(buf[:], uint32(len(fs)))
	_, err := w.Write(buf[:])
	if err != nil {
		return err
	}
	for _, p := range fs {
		var data []byte
		data, err = p.serializeBin()
		if err != nil {
			return err
		}
		// we add the encoded length for decoding
		binary.BigEndian.PutUint32(buf[:], uint32(len(data)))
		_, err = w.Write(buf[:])
		if err != nil {
			return err
		}
		_, err = w.Write(data)
		if err != nil {
			return err
		}
	}

	err = w.Close()
	if err != nil {
		return fmt.Errorf("internal error: can't compress fonset dump: %s", err)
	}
	return nil
}

// LoadFontset reads a cache file, exported by the `Fontset.Serialize` method,
// and constructs the associated font set.
func LoadFontset(src io.Reader) (Fontset, error) {
	r, err := gzip.NewReader(src)
	if err != nil {
		return nil, fmt.Errorf("invalid fontconfig compressed dump file: %s", err)
	}
	defer r.Close()

	var buf [4]byte
	if _, err := io.ReadFull(r, buf[:]); err != nil {
		return nil, fmt.Errorf("invalid fontset format: %s", err)
	}
	// guard against malicious files with hard limit
	L := binary.BigEndian.Uint32(buf[:])
	if L > 1e6 {
		return nil, fmt.Errorf("unsupported fontset length: %d", L)
	}

	fs := make(Fontset, L)
	var buffer bytes.Buffer
	for i := range fs {
		// size of the encoded pattern
		if _, err := io.ReadFull(r, buf[:]); err != nil {
			return nil, fmt.Errorf("invalid fontset: %s", err)
		}
		size := int64(binary.BigEndian.Uint32(buf[:]))
		buffer.Reset()
		_, err := io.CopyN(&buffer, r, size)
		if err != nil {
			return nil, fmt.Errorf("invalid fontset: %s", err)
		}

		fs[i], err = deserializePatternBin(buffer.Bytes())
		if err != nil {
			return nil, fmt.Errorf("invalid fontset: %s", err)
		}
	}
	return fs, nil
}

// LoadFontsetFile is a convenience wrapper of `LoadFontset` for files.
func LoadFontsetFile(file string) (Fontset, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, fmt.Errorf("loading font set: %s", err)
	}
	defer f.Close()

	return LoadFontset(f)
}

// ScanAndCache uses the standard config, scans
// the fonts on disk from the default directories,
// and caches the result into `fontsFileCache`.
func ScanAndCache(fontsFileCache string) (Fontset, error) {
	// launch the scan
	dirs, err := DefaultFontDirs()
	if err != nil {
		return nil, err
	}

	out, err := Standard.ScanFontDirectories(dirs...)
	if err != nil {
		return nil, err
	}

	// create the cache
	f, err := os.Create(fontsFileCache)
	if err != nil {
		return nil, err
	}

	if err = out.Serialize(f); err != nil {
		return nil, err
	}

	err = f.Close()
	return out, err
}
