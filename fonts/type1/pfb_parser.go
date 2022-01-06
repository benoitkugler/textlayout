package type1

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"strings"

	tk "github.com/benoitkugler/pstokenizer"
	"github.com/benoitkugler/textlayout/fonts"
	"github.com/benoitkugler/textlayout/fonts/simpleencodings"
)

const (
	// constants for encryption
	eexecKey       = 55665
	CHARSTRING_KEY = 4330

	headerT11 = "%!FontType"
	headerT12 = "%!PS-AdobeFont"

	// start marker of a segment
	startMarker = 0x80

	// marker of the ascii segment
	asciiMarker = 0x01

	// marker of the binary segment
	binaryMarker = 0x02
)

func readOneRecord(pfb fonts.Resource, expectedMarker byte, totalSize int64) ([]byte, error) {
	var buffer [6]byte

	_, err := io.ReadFull(pfb, buffer[:])
	if err != nil {
		return nil, fmt.Errorf("invalid .pfb file: missing record marker")
	}
	if buffer[0] != startMarker {
		return nil, errors.New("invalid .pfb file: start marker missing")
	}

	if buffer[1] != expectedMarker {
		return nil, errors.New("invalid .pfb file: incorrect record type")
	}

	size := int64(binary.LittleEndian.Uint32(buffer[2:]))
	if size >= totalSize {
		return nil, errors.New("corrupted .pfb file")
	}
	out := make([]byte, size)
	_, err = io.ReadFull(pfb, out)
	if err != nil {
		return nil, fmt.Errorf("invalid .pfb file: %s", err)
	}
	return out, nil
}

// fetchs the segments of a .pfb font file.
// see https://www.adobe.com/content/dam/acom/en/devnet/font/pdfs/5040.Download_Fonts.pdf
// IBM PC format
func openPfb(pfb fonts.Resource) (segment1, segment2 []byte, err error) {
	totalSize, err := pfb.Seek(0, io.SeekEnd)
	if err != nil {
		return nil, nil, err
	}
	_, err = pfb.Seek(0, io.SeekStart)
	if err != nil {
		return nil, nil, err
	}

	// ascii record
	segment1, err = readOneRecord(pfb, asciiMarker, totalSize)
	if err != nil {
		// try with the brute force approach for file who have no tag
		segment1, segment2, err = seekMarkers(pfb)
		if err == nil {
			return segment1, segment2, nil
		}
		return nil, nil, err
	}

	// binary record
	segment2, err = readOneRecord(pfb, binaryMarker, totalSize)
	if err != nil {
		return nil, nil, err
	}
	// ignore the last segment, which is not needed

	return segment1, segment2, nil
}

// fallback when no binary marker are present:
// we look for the currentfile exec pattern, then for the cleartomark
func seekMarkers(pfb fonts.Resource) (segment1, segment2 []byte, err error) {
	_, err = pfb.Seek(0, io.SeekStart)
	if err != nil {
		return nil, nil, err
	}

	// quickly return for invalid files
	var buffer [len(headerT12)]byte
	io.ReadFull(pfb, buffer[:])
	if h := string(buffer[:]); !(strings.HasPrefix(h, headerT11) || strings.HasPrefix(h, headerT12)) {
		return nil, nil, errors.New("not a Type1 font file")
	}

	_, err = pfb.Seek(0, io.SeekStart)
	if err != nil {
		return nil, nil, err
	}
	data, err := ioutil.ReadAll(pfb)
	if err != nil {
		return nil, nil, err
	}
	const exec = "currentfile eexec"
	index := bytes.Index(data, []byte(exec))
	if index == -1 {
		return nil, nil, errors.New("not a Type1 font file")
	}
	segment1 = data[:index+len(exec)]
	segment2 = data[index+len(exec):]
	if len(segment2) != 0 && tk.IsAsciiWhitespace(segment2[0]) { // end of line
		segment2 = segment2[1:]
	}
	return segment1, segment2, nil
}

type parser struct {
	lexer lexer
}

type lexer struct {
	tk.Tokenizer
}

// constructs a new lexer given a header-less .pfb segment
func newLexer(data []byte) lexer {
	return lexer{*tk.NewTokenizer(data)}
}

func (l *lexer) nextToken() (tk.Token, error) {
	return l.Tokenizer.NextToken()
}

func (l lexer) peekToken() tk.Token {
	t, _ := l.Tokenizer.PeekToken()
	return t
}

// Encoding is either the standard encoding, or defined by the font
type Encoding struct {
	Custom   simpleencodings.Encoding
	Standard bool
}

// The Type 1 font format is a free-text format, composed of `segment1` (ASCII) and `segment2` (Binary),
// which is somewhat difficult to parse. This is made worse by the fact that many Type 1 font files do
// not conform to the specification, especially those embedded in PDFs. This
// parser therefore tries to be as forgiving as possible.
//
// See "Adobe Type 1 Font Format, Adobe Systems (1999)"
//
// Ported from the code from John Hewson
func parse(segment1, segment2 []byte) (Font, error) {
	p := parser{}
	out, err := p.parseASCII(segment1)
	if err != nil {
		return Font{}, err
	}
	if len(segment2) > 0 {
		p.parseBinary(segment2, &out)
	}
	return out, nil
}

// Parses the ASCII portion of a Type 1 font.
func (p *parser) parseASCII(bytes []byte) (Font, error) {
	if len(bytes) == 0 {
		return Font{}, errors.New("bytes is empty")
	}

	// %!FontType1-1.0
	// %!PS-AdobeFont-1.0
	if len(bytes) < 2 || (bytes[0] != '%' && bytes[1] != '!') {
		return Font{}, errors.New("Invalid start of ASCII segment")
	}

	var out Font
	p.lexer = newLexer(bytes)

	// (corrupt?) synthetic font
	if string(p.lexer.peekToken().Value) == "FontDirectory" {
		if err := p.readWithName(tk.Other, "FontDirectory"); err != nil {
			return out, err
		}
		if _, err := p.read(tk.Name); err != nil { // font name;
			return out, err
		}
		if err := p.readWithName(tk.Other, "known"); err != nil {
			return out, err
		}
		if _, err := p.read(tk.StartProc); err != nil {
			return out, err
		}
		if _, err := p.readProc(); err != nil {
			return out, err
		}
		if _, err := p.read(tk.StartProc); err != nil {
			return out, err
		}
		if _, err := p.readProc(); err != nil {
			return out, err
		}
		if err := p.readWithName(tk.Other, "ifelse"); err != nil {
			return out, err
		}
	}

	// font dict
	lengthT, err := p.read(tk.Integer)
	if err != nil {
		return out, err
	}
	length, _ := lengthT.Int()
	if err := p.readWithName(tk.Other, "dict"); err != nil {
		return out, err
	}
	// found in some TeX fonts
	if _, err := p.readMaybe(tk.Other, "dup"); err != nil {
		return out, err
	}
	// if present, the "currentdict" is not required
	if err := p.readWithName(tk.Other, "begin"); err != nil {
		return out, err
	}

	for i := 0; i < length; i++ {
		token := p.lexer.peekToken()
		if token.Kind == 0 { // premature end
			break
		}
		if token.IsOther("currentdict") || token.IsOther("end") {
			break
		}

		// key/value
		keyT, err := p.read(tk.Name)
		if err != nil {
			return out, err
		}
		switch key := string(keyT.Value); key {
		case "FontInfo", "Fontinfo":
			dict, err := p.readSimpleDict()
			if err != nil {
				return out, err
			}
			out.PSInfo = p.readFontInfo(dict)
		case "Metrics":
			_, err = p.readSimpleDict()
		case "Encoding":
			out.Encoding, err = p.readEncoding()
		default:
			err = p.readSimpleValue(key, &out)
		}
		if err != nil {
			return out, err
		}
	}

	if _, err := p.readMaybe(tk.Other, "currentdict"); err != nil {
		return out, err
	}
	if err := p.readWithName(tk.Other, "end"); err != nil {
		return out, err
	}
	if err := p.readWithName(tk.Other, "currentfile"); err != nil {
		return out, err
	}
	if err := p.readWithName(tk.Other, "eexec"); err != nil {
		return out, err
	}
	return out, nil
}

func (p *parser) readSimpleValue(key string, font *Font) error {
	value, err := p.readDictValue()
	if err != nil {
		return err
	}
	switch key {
	case "FontName", "PaintType", "FontType", "UniqueID", "StrokeWidth", "FID":
		if len(value) == 0 {
			return fmt.Errorf("missing value for key %s", key)
		}
	}
	switch key {
	case "FontName":
		font.FontName = string(value[0].Value)
	case "PaintType":
		font.PaintType, _ = value[0].Int()
	case "FontType":
		font.FontType, _ = value[0].Int()
	case "UniqueID":
		font.UniqueID, _ = value[0].Int()
	case "StrokeWidth":
		f, _ := value[0].Float()
		font.StrokeWidth = float32(f)
	case "FID":
		font.FontID = string(value[0].Value)
	case "FontMatrix":
		font.FontMatrix, err = p.arrayToNumbers(value)
	case "FontBBox":
		font.FontBBox, err = p.arrayToNumbers(value)
	}
	return err
}

func (p *parser) readEncoding() (*simpleencodings.Encoding, error) {
	var out *simpleencodings.Encoding
	if p.lexer.peekToken().Kind == tk.Other {
		nameT, err := p.lexer.nextToken()
		if err != nil {
			return nil, err
		}
		name_ := string(nameT.Value)
		if name_ == "StandardEncoding" {
			out = &simpleencodings.AdobeStandard
		} else {
			return nil, errors.New("Unknown encoding: " + name_)
		}
		if _, err := p.readMaybe(tk.Other, "readonly"); err != nil {
			return nil, err
		}
		if err := p.readWithName(tk.Other, "def"); err != nil {
			return nil, err
		}
	} else {
		if _, err := p.read(tk.Integer); err != nil {
			return nil, err
		}
		if _, err := p.readMaybe(tk.Other, "array"); err != nil {
			return nil, err
		}

		// 0 1 255 {1 index exch /.notdef put } for
		// we have to check "readonly" and "def" too
		// as some fonts don't provide any dup-values, see PDFBOX-2134
		for {
			n := p.lexer.peekToken()
			if n.IsOther("dup") || n.IsOther("readonly") || n.IsOther("def") {
				break
			}
			_, err := p.lexer.nextToken()
			if err != nil {
				return nil, err
			}
		}

		out = new(simpleencodings.Encoding)
		for p.lexer.peekToken().IsOther("dup") {
			if err := p.readWithName(tk.Other, "dup"); err != nil {
				return nil, err
			}
			codeT, err := p.read(tk.Integer)
			if err != nil {
				return nil, err
			}
			code, _ := codeT.Int()
			nameT, err := p.read(tk.Name)
			if err != nil {
				return nil, err
			}
			if err := p.readWithName(tk.Other, "put"); err != nil {
				return nil, err
			}
			out[byte(code)] = string(nameT.Value)
		}
		if _, err := p.readMaybe(tk.Other, "readonly"); err != nil {
			return nil, err
		}
		if err := p.readWithName(tk.Other, "def"); err != nil {
			return nil, err
		}
	}
	return out, nil
}

// Extracts values from an array as numbers.
func (p *parser) arrayToNumbers(value []tk.Token) ([]Fl, error) {
	var numbers []Fl
	for i, size := 1, len(value)-1; i < size; i++ {
		token := value[i]
		if token.Kind == tk.Float || token.Kind == tk.Integer {
			f, _ := token.Float()
			numbers = append(numbers, Fl(f))
		} else {
			return nil, fmt.Errorf("Expected INTEGER or REAL but got %s", token.Kind)
		}
	}
	return numbers, nil
}

// Extracts values from the /FontInfo dictionary.
func (p *parser) readFontInfo(fontInfo map[string][]tk.Token) fonts.PSInfo {
	var out fonts.PSInfo
	for key, value := range fontInfo {
		switch key {
		case "version":
			out.Version = string(value[0].Value)
		case "Notice":
			out.Notice = string(value[0].Value)
		case "FullName":
			out.FullName = string(value[0].Value)
		case "FamilyName":
			out.FamilyName = string(value[0].Value)
		case "Weight":
			out.Weight = string(value[0].Value)
		case "isFixedPitch":
			out.IsFixedPitch = string(value[0].Value) == "true"
		case "ItalicAngle":
			out.ItalicAngle, _ = value[0].Int()
		case "UnderlinePosition":
			out.UnderlinePosition, _ = value[0].Int()
		case "UnderlineThickness":
			out.UnderlineThickness, _ = value[0].Int()
		}
	}
	return out
}

// Reads a dictionary whose values are simple, i.e., do not contain nested dictionaries.
func (p *parser) readSimpleDict() (map[string][]tk.Token, error) {
	dict := map[string][]tk.Token{}

	lengthT, err := p.read(tk.Integer)
	if err != nil {
		return nil, err
	}
	length, _ := lengthT.Int()
	if err := p.readWithName(tk.Other, "dict"); err != nil {
		return nil, err
	}
	if _, err := p.readMaybe(tk.Other, "dup"); err != nil {
		return nil, err
	}
	if err := p.readWithName(tk.Other, "begin"); err != nil {
		return nil, err
	}

	for i := 0; i < length; i++ {
		if p.lexer.peekToken().Kind == 0 {
			break
		}
		if p.lexer.peekToken().Kind == tk.Other &&
			!(string(p.lexer.peekToken().Value) == "end") {
			if _, err := p.read(tk.Other); err != nil {
				return nil, err
			}
		}
		// premature end
		if p.lexer.peekToken().Kind == 0 {
			break
		}
		if p.lexer.peekToken().IsOther("end") {
			break
		}

		// simple value
		keyT, err := p.read(tk.Name)
		if err != nil {
			return nil, err
		}
		value, err := p.readDictValue()
		if err != nil {
			return nil, err
		}
		dict[string(keyT.Value)] = value
	}

	if err := p.readWithName(tk.Other, "end"); err != nil {
		return nil, err
	}
	if _, err := p.readMaybe(tk.Other, "readonly"); err != nil {
		return nil, err
	}
	if err := p.readWithName(tk.Other, "def"); err != nil {
		return nil, err
	}

	return dict, nil
}

// Reads a simple value from a dictionary.
func (p *parser) readDictValue() ([]tk.Token, error) {
	value, err := p.readValue()
	if err != nil {
		return nil, err
	}
	err = p.readDef()
	return value, err
}

// Reads a simple value. This is either a number, a string,
// a name, a literal name, an array, a procedure, or a charstring.
// This method does not support reading nested dictionaries unless they're empty.
func (p *parser) readValue() ([]tk.Token, error) {
	var value []tk.Token
	token, err := p.lexer.nextToken()
	if err != nil {
		return nil, err
	}
	if p.lexer.peekToken().Kind == 0 {
		return value, nil
	}
	value = append(value, token)

	switch token.Kind {
	case tk.StartArray:
		openArray := 1
		for {
			if p.lexer.peekToken().Kind == 0 {
				return value, nil
			}
			if p.lexer.peekToken().Kind == tk.StartArray {
				openArray++
			}

			token, err = p.lexer.nextToken()
			if err != nil {
				return nil, err
			}
			value = append(value, token)

			if token.Kind == tk.EndArray {
				openArray--
				if openArray == 0 {
					break
				}
			}
		}
	case tk.StartProc:
		proc, err := p.readProc()
		if err != nil {
			return nil, err
		}
		value = append(value, proc...)
	case tk.StartDic:
		// skip "/GlyphNames2HostCode << >> def"
		if _, err = p.read(tk.EndDic); err != nil {
			return nil, err
		}
		return value, nil
	}
	err = p.readPostScriptWrapper(value)
	return value, err
}

func (p *parser) readPostScriptWrapper(value []tk.Token) error {
	// postscript wrapper (not in the Type 1 spec)
	if string(p.lexer.peekToken().Value) != "systemdict" {
		return nil
	}
	if err := p.readWithName(tk.Other, "systemdict"); err != nil {
		return err
	}
	if err := p.readWithName(tk.Name, "internaldict"); err != nil {
		return err
	}
	if err := p.readWithName(tk.Other, "known"); err != nil {
		return err
	}

	if _, err := p.read(tk.StartProc); err != nil {
		return err
	}
	if _, err := p.readProc(); err != nil {
		return err
	}

	if _, err := p.read(tk.StartProc); err != nil {
		return err
	}
	if _, err := p.readProc(); err != nil {
		return err
	}

	if err := p.readWithName(tk.Other, "ifelse"); err != nil {
		return err
	}

	// replace value
	if _, err := p.read(tk.StartProc); err != nil {
		return err
	}
	if err := p.readWithName(tk.Other, "pop"); err != nil {
		return err
	}
	value = nil
	other, err := p.readValue()
	if err != nil {
		return err
	}
	value = append(value, other...)
	if _, err := p.read(tk.EndProc); err != nil {
		return err
	}

	if err := p.readWithName(tk.Other, "if"); err != nil {
		return err
	}
	return nil
}

// Reads a procedure.
func (p *parser) readProc() ([]tk.Token, error) {
	var value []tk.Token
	openProc := 1
	for {
		if p.lexer.peekToken().Kind == tk.StartProc {
			openProc++
		}

		token, err := p.lexer.nextToken()
		if err != nil {
			return nil, err
		}
		value = append(value, token)

		if token.Kind == tk.EndProc {
			openProc--
			if openProc == 0 {
				break
			}
		}
	}
	executeonly, err := p.readMaybe(tk.Other, "executeonly")
	if err != nil {
		return nil, err
	}
	if executeonly.Kind != 0 {
		value = append(value, executeonly)
	}

	return value, nil
}

// Parses the binary portion of a Type 1 font.
func (p *parser) parseBinary(bytes []byte, font *Font) error {
	decrypted := decryptSegment(bytes)

	p.lexer = newLexer(decrypted)

	// find /Private dict
	peekToken := p.lexer.peekToken()
	for string(peekToken.Value) != "Private" {
		// for a more thorough validation, the presence of "begin" before Private
		// determines how code before and following charstrings should look
		// it is not currently checked anyway
		_, err := p.lexer.nextToken()
		if err != nil {
			return err
		}
		peekToken = p.lexer.peekToken()
	}
	if peekToken.Kind == 0 {
		return errors.New("/Private token not found")
	}

	// Private dict
	if err := p.readWithName(tk.Name, "Private"); err != nil {
		return err
	}
	lengthT, err := p.read(tk.Integer)
	if err != nil {
		return err
	}
	length, _ := lengthT.Int()
	if err = p.readWithName(tk.Other, "dict"); err != nil {
		return err
	}
	// actually could also be "/Private 10 dict def Private begin"
	// instead of the "dup"
	if _, err = p.readMaybe(tk.Other, "dup"); err != nil {
		return err
	}
	if err = p.readWithName(tk.Other, "begin"); err != nil {
		return err
	}

	lenIV := 4 // number of random bytes at start of charstring

	for i := 0; i < length; i++ {
		// premature end
		if p.lexer.peekToken().Kind != tk.Name {
			break
		}

		// key/value
		key, err := p.read(tk.Name)
		if err != nil {
			return err
		}

		switch string(key.Value) {
		case "Subrs":
			font.subrs, err = p.readSubrs(lenIV)
		case "OtherSubrs":
			err = p.readOtherSubrs()
		case "lenIV":
			vs, err := p.readDictValue()
			if err != nil {
				return err
			}
			lenIV, err = vs[0].Int()
		case "ND":
			if _, err = p.read(tk.StartProc); err != nil {
				return err
			}
			// the access restrictions are not mandatory
			if _, err = p.readMaybe(tk.Other, "noaccess"); err != nil {
				return err
			}
			if err = p.readWithName(tk.Other, "def"); err != nil {
				return err
			}
			if _, err = p.read(tk.EndProc); err != nil {
				return err
			}
			if _, err = p.readMaybe(tk.Other, "executeonly"); err != nil {
				return err
			}
			if err = p.readWithName(tk.Other, "def"); err != nil {
				return err
			}
		case "NP":
			if _, err = p.read(tk.StartProc); err != nil {
				return err
			}
			if _, err = p.readMaybe(tk.Other, "noaccess"); err != nil {
				return err
			}
			if _, err = p.read(tk.Other); err != nil {
				return err
			}
			if _, err = p.read(tk.EndProc); err != nil {
				return err
			}
			if _, err = p.readMaybe(tk.Other, "executeonly"); err != nil {
				return err
			}
			if err = p.readWithName(tk.Other, "def"); err != nil {
				return err
			}
		case "RD":
			// /RD {string currentfile exch readstring pop} bind executeonly def
			if _, err = p.read(tk.StartProc); err != nil {
				return err
			}
			if _, err = p.readProc(); err != nil {
				return err
			}
			if _, err = p.readMaybe(tk.Other, "bind"); err != nil {
				return err
			}
			if _, err = p.readMaybe(tk.Other, "executeonly"); err != nil {
				return err
			}
			if err = p.readWithName(tk.Other, "def"); err != nil {
				return err
			}
		default:
			var vs []tk.Token
			vs, err = p.readDictValue()
			if err != nil {
				return err
			}
			err = p.readPrivate(key.Value, vs)
		}

		if err != nil {
			return err
		}
	}

	// some fonts have "2 index" here, others have "end noaccess put"
	// sometimes followed by "put". Either way, we just skip until
	// the /CharStrings dict is found
	for {
		n := p.lexer.peekToken()
		if n.Kind == tk.Name && string(n.Value) == "CharStrings" {
			break
		}
		_, err := p.lexer.nextToken()
		if err != nil {
			return err
		}
	}

	// CharStrings dict
	if err = p.readWithName(tk.Name, "CharStrings"); err != nil {
		return err
	}
	font.charstrings, err = p.readCharStrings(lenIV)
	return err
}

// Extracts values from the /Private dictionary.
func (p *parser) readPrivate(key []byte, value []tk.Token) error {
	// TODO: complete if needed
	// 		 switch (key)
	// 		 {
	// 			 case "BlueValues":
	// 				 font.blueValues = arrayToNumbers(value);
	// 				 break;
	// 			 case "OtherBlues":
	// 				 font.otherBlues = arrayToNumbers(value);
	// 				 break;
	// 			 case "FamilyBlues":
	// 				 font.familyBlues = arrayToNumbers(value);
	// 				 break;
	// 			 case "FamilyOtherBlues":
	// 				 font.familyOtherBlues = arrayToNumbers(value);
	// 				 break;
	// 			 case "BlueScale":
	// 				 font.blueScale = value[0].floatValue();
	// 				 break;
	// 			 case "BlueShift":
	// 				 font.blueShift = value[0].intValue();
	// 				 break;
	// 			 case "BlueFuzz":
	// 				 font.blueFuzz = value[0].intValue();
	// 				 break;
	// 			 case "StdHW":
	// 				 font.stdHW = arrayToNumbers(value);
	// 				 break;
	// 			 case "StdVW":
	// 				 font.stdVW = arrayToNumbers(value);
	// 				 break;
	// 			 case "StemSnapH":
	// 				 font.stemSnapH = arrayToNumbers(value);
	// 				 break;
	// 			 case "StemSnapV":
	// 				 font.stemSnapV = arrayToNumbers(value);
	// 				 break;
	// 			 case "ForceBold":
	// 				 font.forceBold = value[0].booleanValue();
	// 				 break;
	// 			 case "LanguageGroup":
	// 				 font.languageGroup = value[0].intValue();
	// 				 break;
	// 			 default:
	// 				 break;
	// 		 }
	return nil
}

// Reads the /Subrs array.
// `lenIV` is he number of random bytes used in charstring encryption.
func (p *parser) readSubrs(lenIV int) ([][]byte, error) {
	// allocate size (array indexes may not be in-order)
	lengthT, err := p.read(tk.Integer)
	if err != nil {
		return nil, err
	}
	length, _ := lengthT.Int()
	subrs := make([][]byte, length)
	if err = p.readWithName(tk.Other, "array"); err != nil {
		return nil, err
	}

	for i := 0; i < length; i++ {
		// premature end
		if !p.lexer.peekToken().IsOther("dup") {
			break
		}

		if err = p.readWithName(tk.Other, "dup"); err != nil {
			return nil, err
		}
		indexT, err := p.read(tk.Integer)
		if err != nil {
			return nil, err
		}
		index, _ := indexT.Int()
		if _, err = p.read(tk.Integer); err != nil {
			return nil, err
		}
		if index >= length {
			return nil, fmt.Errorf("out of range charstring index %d (for %d)", index, length)
		}

		// RD
		charstring, err := p.read(tk.CharString)
		if err != nil {
			return nil, err
		}
		subrs[index] = decrypt([]byte(charstring.Value), CHARSTRING_KEY, lenIV)
		err = p.readPut()
		if err != nil {
			return nil, err
		}
	}
	err = p.readDef()
	return subrs, err
}

// OtherSubrs are embedded PostScript procedures which we can safely ignore
func (p *parser) readOtherSubrs() error {
	if p.lexer.peekToken().Kind == tk.StartArray {
		if _, err := p.readValue(); err != nil {
			return err
		}
		err := p.readDef()
		return err
	}
	lengthT, err := p.read(tk.Integer)
	if err != nil {
		return err
	}
	length, _ := lengthT.Int()
	if err = p.readWithName(tk.Other, "array"); err != nil {
		return err
	}

	for i := 0; i < length; i++ {
		if err = p.readWithName(tk.Other, "dup"); err != nil {
			return err
		}
		if _, err = p.read(tk.Integer); err != nil { // index
			return err
		}
		if _, err = p.readValue(); err != nil { // PostScript
			return err
		}
		if err = p.readPut(); err != nil {
			return err
		}
	}
	err = p.readDef()
	return err
}

// Reads the /CharStrings dictionary.
// `lenIV` is the number of random bytes used in charstring encryption.
func (p *parser) readCharStrings(lenIV int) ([]charstring, error) {
	lengthT, err := p.read(tk.Integer)
	if err != nil {
		return nil, err
	}
	length, _ := lengthT.Int()
	if err = p.readWithName(tk.Other, "dict"); err != nil {
		return nil, err
	}
	// could actually be a sequence ending in "CharStrings begin", too
	// instead of the "dup begin"
	if err = p.readWithName(tk.Other, "dup"); err != nil {
		return nil, err
	}
	if err = p.readWithName(tk.Other, "begin"); err != nil {
		return nil, err
	}

	charstrings := make([]charstring, length)
	for i := range charstrings {
		// premature end
		if tok := p.lexer.peekToken(); tok.Kind == 0 || tok.IsOther("end") {
			break
		}
		// key/value
		nameT, err := p.read(tk.Name)
		if err != nil {
			return nil, err
		}

		// RD
		_, err = p.read(tk.Integer)
		if err != nil {
			return nil, err
		}
		charstring, err := p.read(tk.CharString)
		if err != nil {
			return nil, err
		}

		charstrings[i].name = string(nameT.Value)
		charstrings[i].data = decrypt(charstring.Value, CHARSTRING_KEY, lenIV)

		err = p.readDef()
		if err != nil {
			return nil, err
		}
	}

	// some fonts have one "end", others two
	err = p.readWithName(tk.Other, "end")
	// since checking ends here, this does not matter ....
	// more thorough checking would see whether there is "begin" before /Private
	// and expect a "def" somewhere, otherwise a "put"
	return charstrings, err
}

// Reads the sequence "noaccess def" or equivalent.
func (p *parser) readDef() error {
	if _, err := p.readMaybe(tk.Other, "readonly"); err != nil {
		return err
	}
	// allows "noaccess ND" (not in the Type 1 spec)
	if _, err := p.readMaybe(tk.Other, "noaccess"); err != nil {
		return err
	}

	token, err := p.read(tk.Other)
	if err != nil {
		return err
	}
	switch string(token.Value) {
	case "ND", "|-":
		return nil
	case "noaccess":
		token, err = p.read(tk.Other)
		if err != nil {
			return err
		}
	}
	if string(token.Value) == "def" {
		return nil
	}
	return fmt.Errorf("Found %s but expected ND", token.Value)
}

// Reads the sequence "noaccess put" or equivalent.
func (p *parser) readPut() error {
	_, err := p.readMaybe(tk.Other, "readonly")
	if err != nil {
		return err
	}

	token, err := p.read(tk.Other)
	if err != nil {
		return err
	}
	switch string(token.Value) {
	case "NP", "|":
		return nil
	case "noaccess":
		token, err = p.read(tk.Other)
		if err != nil {
			return err
		}
	}

	if string(token.Value) == "put" {
		return nil
	}
	return fmt.Errorf("found %s but expected NP", token.Value)
}

/// Reads the next token and throws an error if it is not of the given kind.
func (p *parser) read(kind tk.Kind) (tk.Token, error) {
	token, err := p.lexer.nextToken()
	if err != nil {
		return tk.Token{}, err
	}
	if token.Kind != kind {
		return tk.Token{}, fmt.Errorf("found token %s (%s) but expected token %s", token.Kind, token.Value, kind)
	}
	return token, nil
}

// Reads the next token and throws an error if it is not of the given kind
// and does not have the given value.
func (p *parser) readWithName(kind tk.Kind, name string) error {
	token, err := p.read(kind)
	if err != nil {
		return err
	}
	if string(token.Value) != name {
		return fmt.Errorf("found %s but expected %s", token.Value, name)
	}
	return nil
}

// Reads the next token if and only if it is of the given kind and
// has the given value.
func (p *parser) readMaybe(kind tk.Kind, name string) (tk.Token, error) {
	token := p.lexer.peekToken()
	if token.Kind == kind && string(token.Value) == name {
		return p.lexer.nextToken()
	}
	return tk.Token{}, nil
}

func decryptSegment(crypted []byte) []byte {
	// Sometimes, fonts use the hex format, so this needs to be converted before decryption
	if isBinary(crypted) {
		return decrypt(crypted, eexecKey, 4)
	}
	return decrypt(hexToBinary(crypted), eexecKey, 4)
}

// Type 1 Decryption (eexec, charstring).
// `r` is the key and `n` the number of random bytes (lenIV)
// the input is modified (the return slice share its storage)
func decrypt(cipherBytes []byte, r uint16, n int) []byte {
	// lenIV of -1 means no encryption (not documented)
	if n == -1 {
		return cipherBytes
	}
	// empty charstrings and charstrings of insufficient length
	if len(cipherBytes) == 0 || len(cipherBytes) < n {
		return nil
	}
	// decrypt
	const (
		c1 uint16 = 52845
		c2 uint16 = 22719
	)
	for i, c := range cipherBytes {
		cipherBytes[i] = c ^ byte(r>>8)
		r = (uint16(c)+r)*c1 + c2
	}
	return cipherBytes[n:]
}

// Check whether binary or hex encoded. See Adobe Type 1 Font Format specification
// 7.2 eexec encryption
func isBinary(bytes []byte) bool {
	if len(bytes) < 4 {
		return true
	}
	// "At least one of the first 4 ciphertext bytes must not be one of
	// the ASCII hexadecimal character codes (a code for 0-9, A-F, or a-f)."
	for i := 0; i < 4; i++ {
		by := bytes[i]
		isSpace := tk.IsAsciiWhitespace(by)
		_, isHex := tk.IsHexChar(by)
		if !isSpace && !isHex {
			return true
		}
	}
	return false
}

func hexToBinary(data []byte) []byte {
	// white space characters may be interspersed
	tmp := make([]byte, 0, len(data))
	for _, c := range data {
		if _, ok := tk.IsHexChar(c); ok {
			tmp = append(tmp, c)
		}
	}
	out := make([]byte, hex.DecodedLen(len(tmp)))
	hex.Decode(out, tmp)
	return out
}
