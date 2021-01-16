package type1

import (
	"encoding/hex"
	"errors"
	"fmt"

	tk "github.com/benoitkugler/pstokenizer"
	"github.com/benoitkugler/textlayout/fonts/simpleencodings"
)

// constants for encryption
const (
	EEXEC_KEY      = 55665
	CHARSTRING_KEY = 4330
)

var none = tk.Token{} // null token

type parser struct {
	// state
	lexer lexer
	font  PFBFont
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
	t, err := l.Tokenizer.PeekToken()
	if err != nil {
		return none
	}
	return t
}

// Encoding is either the standard encoding, or defined by the font
type Encoding struct {
	Standard bool
	Custom   simpleencodings.Encoding
}

/*
 Parses an Adobe Type 1 (.pfb) font, composed of `segment1` (ASCII) and `segment2` (Binary).
 It is used exclusively in Type1 font.

 The Type 1 font format is a free-text format which is somewhat difficult
 to parse. This is made worse by the fact that many Type 1 font files do
 not conform to the specification, especially those embedded in PDFs. This
 parser therefore tries to be as forgiving as possible.

 See "Adobe Type 1 Font Format, Adobe Systems (1999)"

 Ported from the code from John Hewson

 For now, only the parsing of the first segment is implemented.
*/
func Parse(segment1 []byte) (PFBFont, error) {
	p := parser{}
	err := p.parseASCII(segment1)
	if err != nil {
		return PFBFont{}, err
	}
	// TODO:
	// if len(segment2) > 0 {
	// parser.parseBinary(segment2)
	// }
	return p.font, nil
}

// Parses the ASCII portion of a Type 1 font.
func (p *parser) parseASCII(bytes []byte) error {
	if len(bytes) == 0 {
		return errors.New("bytes is empty")
	}

	// %!FontType1-1.0
	// %!PS-AdobeFont-1.0
	if len(bytes) < 2 || (bytes[0] != '%' && bytes[1] != '!') {
		return errors.New("Invalid start of ASCII segment")
	}

	p.lexer = newLexer(bytes)

	// (corrupt?) synthetic font
	if p.lexer.peekToken().Value == "FontDirectory" {
		if err := p.readWithName(tk.Other, "FontDirectory"); err != nil {
			return err
		}
		if _, err := p.read(tk.Name); err != nil { // font name;
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
	}

	// font dict
	lengthT, err := p.read(tk.Integer)
	if err != nil {
		return err
	}
	length, _ := lengthT.Int()
	if err := p.readWithName(tk.Other, "dict"); err != nil {
		return err
	}
	// found in some TeX fonts
	if _, err := p.readMaybe(tk.Other, "dup"); err != nil {
		return err
	}
	// if present, the "currentdict" is not required
	if err := p.readWithName(tk.Other, "begin"); err != nil {
		return err
	}

	for i := 0; i < length; i++ {
		// premature end
		token := p.lexer.peekToken()
		if token == none {
			break
		}
		if token.Kind == tk.Other && ("currentdict" == token.Value || "end" == token.Value) {
			break
		}

		// key/value
		keyT, err := p.read(tk.Name)
		if err != nil {
			return err
		}
		switch key := keyT.Value; key {
		case "FontInfo", "Fontinfo":
			dict, err := p.readSimpleDict()
			if err != nil {
				return err
			}
			p.readFontInfo(dict)
		case "Metrics":
			_, err = p.readSimpleDict()
		case "Encoding":
			err = p.readEncoding()
		default:
			err = p.readSimpleValue(key)
		}
		if err != nil {
			return err
		}
	}

	if _, err := p.readMaybe(tk.Other, "currentdict"); err != nil {
		return err
	}
	if err := p.readWithName(tk.Other, "end"); err != nil {
		return err
	}
	if err := p.readWithName(tk.Other, "currentfile"); err != nil {
		return err
	}
	if err := p.readWithName(tk.Other, "eexec"); err != nil {
		return err
	}
	return nil
}

func (p *parser) readSimpleValue(key string) error {
	value, err := p.readDictValue()
	if err != nil {
		return err
	}
	switch key {
	case "FontName", "PaintType", "FontType", "UniqueID", "StrokeWidth", "FID":
		if len(value) == 0 {
			return errors.New("missing value")
		}
	}
	switch key {
	case "FontName":
		p.font.FontName = value[0].Value
	case "PaintType":
		p.font.PaintType, _ = value[0].Int()
	case "FontType":
		p.font.FontType, _ = value[0].Int()
	case "UniqueID":
		p.font.UniqueID, _ = value[0].Int()
	case "StrokeWidth":
		p.font.StrokeWidth, _ = value[0].Float()
	case "FID":
		p.font.FontID = value[0].Value
	case "FontMatrix":
		p.font.FontMatrix, err = p.arrayToNumbers(value)
	case "FontBBox":
		p.font.FontBBox, err = p.arrayToNumbers(value)
	}
	return err
}

func (p *parser) readEncoding() error {
	if p.lexer.peekToken().Kind == tk.Other {
		nameT, err := p.lexer.nextToken()
		if err != nil {
			return err
		}
		name_ := nameT.Value
		if name_ == "StandardEncoding" {
			p.font.Encoding.Standard = true
		} else {
			return errors.New("Unknown encoding: " + name_)
		}
		if _, err := p.readMaybe(tk.Other, "readonly"); err != nil {
			return err
		}
		if err := p.readWithName(tk.Other, "def"); err != nil {
			return err
		}
	} else {
		if _, err := p.read(tk.Integer); err != nil {
			return err
		}
		if _, err := p.readMaybe(tk.Other, "array"); err != nil {
			return err
		}

		// 0 1 255 {1 index exch /.notdef put } for
		// we have to check "readonly" and "def" too
		// as some fonts don't provide any dup-values, see PDFBOX-2134
		for !(p.lexer.peekToken().Kind == tk.Other &&
			(p.lexer.peekToken().Value == "dup" ||
				p.lexer.peekToken().Value == "readonly" ||
				p.lexer.peekToken().Value == "def")) {
			_, err := p.lexer.nextToken()
			if err != nil {
				return err
			}
		}

		for p.lexer.peekToken().Kind == tk.Other &&
			p.lexer.peekToken().Value == "dup" {
			if err := p.readWithName(tk.Other, "dup"); err != nil {
				return err
			}
			codeT, err := p.read(tk.Integer)
			if err != nil {
				return err
			}
			code, _ := codeT.Int()
			nameT, err := p.read(tk.Name)
			if err != nil {
				return err
			}
			if err := p.readWithName(tk.Other, "put"); err != nil {
				return err
			}
			p.font.Encoding.Custom[byte(code)] = nameT.Value
		}
		if _, err := p.readMaybe(tk.Other, "readonly"); err != nil {
			return err
		}
		if err := p.readWithName(tk.Other, "def"); err != nil {
			return err
		}
	}
	return nil
}

// Extracts values from an array as numbers.
func (p *parser) arrayToNumbers(value []tk.Token) ([]float64, error) {
	var numbers []float64
	for i, size := 1, len(value)-1; i < size; i++ {
		token := value[i]
		if token.Kind == tk.Float || token.Kind == tk.Integer {
			f, _ := token.Float()
			numbers = append(numbers, f)
		} else {
			return nil, fmt.Errorf("Expected INTEGER or REAL but got %s", token.Kind)
		}
	}
	return numbers, nil
}

// Extracts values from the /FontInfo dictionary.
func (p *parser) readFontInfo(fontInfo map[string][]tk.Token) {
	for key, value := range fontInfo {
		switch key {
		case "version":
			p.font.Version = value[0].Value
		case "Notice":
			p.font.Notice = value[0].Value
		case "FullName":
			p.font.FullName = value[0].Value
		case "FamilyName":
			p.font.FamilyName = value[0].Value
		case "Weight":
			p.font.Weight = value[0].Value
		case "isFixedPitch":
			p.font.IsFixedPitch = value[0].Value == "true"
		case "ItalicAngle":
			p.font.ItalicAngle, _ = value[0].Int()
		case "UnderlinePosition":
			p.font.UnderlinePosition, _ = value[0].Int()
		case "UnderlineThickness":
			p.font.UnderlineThickness, _ = value[0].Int()
		}
	}
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
		if p.lexer.peekToken() == none {
			break
		}
		if p.lexer.peekToken().Kind == tk.Other &&
			!(p.lexer.peekToken().Value == "end") {
			if _, err := p.read(tk.Other); err != nil {
				return nil, err
			}
		}
		// premature end
		if p.lexer.peekToken() == none {
			break
		}
		if p.lexer.peekToken().Kind == tk.Other &&
			p.lexer.peekToken().Value == "end" {
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
		dict[keyT.Value] = value
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
	if p.lexer.peekToken() == none {
		return value, nil
	}
	value = append(value, token)

	switch token.Kind {
	case tk.StartArray:
		openArray := 1
		for {
			if p.lexer.peekToken() == none {
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
		if _, err := p.read(tk.EndDic); err != nil {
			return nil, err
		}
		return value, nil
	}
	err = p.readPostScriptWrapper(value)
	return value, err
}

func (p *parser) readPostScriptWrapper(value []tk.Token) error {
	// postscript wrapper (not in the Type 1 spec)
	if p.lexer.peekToken().Value != "systemdict" {
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
	if executeonly != none {
		value = append(value, executeonly)
	}

	return value, nil
}

// // Parses the binary portion of a Type 1 font.
// func (p *Parser) parseBinary(bytes []byte) error {
// 	var decrypted []byte
// 	// Sometimes, fonts use the hex format, so this needs to be converted before decryption
// 	if isBinary(bytes) {
// 		decrypted = decrypt(bytes, EEXEC_KEY, 4)
// 	} else {
// 		decrypted = decrypt(hexToBinary(bytes), EEXEC_KEY, 4)
// 	}
// 	lexer := lexer{data: decrypted}

// 	// find /Private dict
// 	peekToken := lexer.peekToken()
// 	for peekToken != none && !peekToken.Value == "Private" {
// 		// for a more thorough validation, the presence of "begin" before Private
// 		// determines how code before and following charstrings should look
// 		// it is not currently checked anyway
// 		lexer.nextToken()
// 		peekToken = lexer.peekToken()
// 	}
// 	if peekToken == none {
// 		return errors.New("/Private token not found")
// 	}

// 	// Private dict
// 	read(pt.Name, "Private")
// 	length := read(pt.Integer).intValue()
// 	read(pt.Other, "dict")
// 	// actually could also be "/Private 10 dict def Private begin"
// 	// instead of the "dup"
// 	p.readMaybe(pt.Other, "dup")
// 	read(pt.Other, "begin")

// 	lenIV := 4 // number of random bytes at start of charstring

// 	for i := 0; i < length; i++ {
// 		// premature end
// 		if lexer.peekToken() == none || lexer.peekToken().Kind != pt.Name {
// 			break
// 		}

// 		// key/value
// 		key := read(pt.Name).Value

// 		switch key {
// 		case "Subrs":
// 			readSubrs(lenIV)

// 		case "OtherSubrs":
// 			readOtherSubrs()

// 		case "lenIV":
// 			lenIV = readDictValue()[0].intValue()

// 		case "ND":
// 			read(pt.StartProc)
// 			// the access restrictions are not mandatory
// 			p.readMaybe(pt.Other, "noaccess")
// 			read(pt.Other, "def")
// 			read(token.END_PROC)
// 			p.readMaybe(pt.Other, "executeonly")
// 			read(pt.Other, "def")

// 		case "NP":
// 			read(pt.StartProc)
// 			p.readMaybe(pt.Other, "noaccess")
// 			read(pt.Other)
// 			read(token.END_PROC)
// 			p.readMaybe(pt.Other, "executeonly")
// 			read(pt.Other, "def")

// 		case "RD":
// 			// /RD {string currentfile exch readstring pop} bind executeonly def
// 			read(pt.StartProc)
// 			readProc()
// 			p.readMaybe(pt.Other, "bind")
// 			p.readMaybe(pt.Other, "executeonly")
// 			read(pt.Other, "def")

// 		default:
// 			readPrivate(key, readDictValue())

// 		}
// 	}

// 	// some fonts have "2 index" here, others have "end noaccess put"
// 	// sometimes followed by "put". Either way, we just skip until
// 	// the /CharStrings dict is found
// 	for !(lexer.peekToken().Kind == pt.Name &&
// 		lexer.peekToken().Value == "CharStrings") {
// 		lexer.nextToken()
// 	}

// 	// CharStrings dict
// 	read(pt.Name, "CharStrings")
// 	readCharStrings(lenIV)
// }

// 	 /**
// 	  * Extracts values from the /Private dictionary.
// 	  */
//func (p *Parser) void readPrivate(String key, List<token> value) error
// 	 {
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
// 	 }

// 	 /**
// 	  * Reads the /Subrs array.
// 	  * @param lenIV The number of random bytes used in charstring encryption.
// 	  */
//func (p *Parser) void readSubrs(int lenIV) error
// 	 {
// 		 // allocate size (array indexes may not be in-order)
// 		  length := read(pt.Integer).intValue();
// 		 for (int i = 0; i < length; i++)
// 		 {
// 			 font.subrs.add(none);
// 		 }
// 		 read(pt.Other, "array");

// 		 for (int i = 0; i < length; i++)
// 		 {
// 			 // premature end
// 			 if (lexer.peekToken() == none)
// 			 {
// 				 break;
// 			 }
// 			 if (!(lexer.peekToken().Kind == pt.Other &&
// 				   lexer.peekToken().Value == "dup")))
// 			 {
// 				 break;
// 			 }

// 			 read(pt.Other, "dup");
// 			 token index = read(pt.Integer);
// 			 read(pt.Integer);

// 			 // RD
// 			 token charstring = read(token.CHARSTRING);
// 			 font.subrs.set(index.intValue(), decrypt(charstring.getData(), CHARSTRING_KEY, lenIV));
// 			 readPut();
// 		 }
// 		 readDef();
// 	 }

// 	 // OtherSubrs are embedded PostScript procedures which we can safely ignore
//func (p *Parser) void readOtherSubrs() error
// 	 {
// 		 if (lexer.peekToken().Kind == token.START_ARRAY)
// 		 {
// 			 readValue();
// 			 readDef();
// 		 }
// 		 else
// 		 {
// 			  length := read(pt.Integer).intValue();
// 			 read(pt.Other, "array");

// 			 for (int i = 0; i < length; i++)
// 			 {
// 				 read(pt.Other, "dup");
// 				 read(pt.Integer); // index
// 				 readValue(); // PostScript
// 				 readPut();
// 			 }
// 			 readDef();
// 		 }
// 	 }

// 	 /**
// 	  * Reads the /CharStrings dictionary.
// 	  * @param lenIV The number of random bytes used in charstring encryption.
// 	  */
//func (p *Parser) void readCharStrings(int lenIV) error
// 	 {
// 		  length := read(pt.Integer).intValue();
// 		 read(pt.Other, "dict");
// 		 // could actually be a sequence ending in "CharStrings begin", too
// 		 // instead of the "dup begin"
// 		 read(pt.Other, "dup");
// 		 read(pt.Other, "begin");

// 		 for (int i = 0; i < length; i++)
// 		 {
// 			 // premature end
// 			 if (lexer.peekToken() == none)
// 			 {
// 				 break;
// 			 }
// 			 if (lexer.peekToken().Kind == pt.Other &&
// 				 lexer.peekToken().Value == "end"))
// 			 {
// 				 break;
// 			 }
// 			 // key/value
// 			 name := read(pt.Name).Value;

// 			 // RD
// 			 read(pt.Integer);
// 			 token charstring = read(token.CHARSTRING);
// 			 font.charstrings.put(name, decrypt(charstring.getData(), CHARSTRING_KEY, lenIV));
// 			 readDef();
// 		 }

// 		 // some fonts have one "end", others two
// 		 read(pt.Other, "end");
// 		 // since checking ends here, this does not matter ....
// 		 // more thorough checking would see whether there is "begin" before /Private
// 		 // and expect a "def" somewhere, otherwise a "put"
// 	 }

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
	switch token.Value {
	case "ND", "|-":
		return nil
	case "noaccess":
		token, err = p.read(tk.Other)
		if err != nil {
			return err
		}
	}
	if token.Value == "def" {
		return nil
	}
	return fmt.Errorf("Found %s but expected ND", token.Value)
}

// 	 /**
// 	  * Reads the sequence "noaccess put" or equivalent.
// 	  */
//func (p *Parser) void readPut() error
// 	 {
// 		 p.readMaybe(pt.Other, "readonly");

// 		 token := read(pt.Other);
// 		 switch (token.Value)
// 		 {
// 			 case "NP":
// 			 case "|":
// 				 return;
// 			 case "noaccess":
// 				 token = read(pt.Other);
// 				 break;
// 			 default:
// 				 break;
// 		 }

// 		 if (token.Value == "put"))
// 		 {
// 			 return;
// 		 }
// 		 return errors.New("Found " + token + " but expected NP");
// 	 }

/// Reads the next token and throws an error if it is not of the given kind.
func (p *parser) read(kind tk.Kind) (tk.Token, error) {
	token, err := p.lexer.nextToken()
	if err != nil {
		return none, err
	}
	if token.Kind != kind {
		return none, fmt.Errorf("found token %s (%s) but expected token %s", token.Kind, token.Value, kind)
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
	if token.Value != name {
		return fmt.Errorf("found %s but expected %s", token.Value, name)
	}
	return nil
}

// Reads the next token if and only if it is of the given kind and
// has the given value.
func (p *parser) readMaybe(kind tk.Kind, name string) (tk.Token, error) {
	token := p.lexer.peekToken()
	if token.Kind == kind && token.Value == name {
		return p.lexer.nextToken()
	}
	return none, nil
}

func decryptSegment(crypted []byte) ([]byte, error) {
	// Sometimes, fonts use the hex format, so this needs to be converted before decryption
	if isBinary(crypted) {
		return decrypt(crypted, EEXEC_KEY, 4), nil
	} else {
		dl := hex.DecodedLen(len(crypted))
		tmp := make([]byte, dl)
		_, err := hex.Decode(tmp, crypted)
		if err != nil {
			return nil, err
		}
		return decrypt(tmp, EEXEC_KEY, 4), nil
	}
}

// Type 1 Decryption (eexec, charstring).
// `r` is the key and `n` the number of random bytes (lenIV)
func decrypt(cipherBytes []byte, r, n int) []byte {
	// lenIV of -1 means no encryption (not documented)
	if n == -1 {
		return cipherBytes
	}
	// empty charstrings and charstrings of insufficient length
	if len(cipherBytes) == 0 || n > len(cipherBytes) {
		return nil
	}
	// decrypt
	c1 := 52845
	c2 := 22719
	plainBytes := make([]byte, len(cipherBytes)-n)
	for i := 0; i < len(cipherBytes); i++ {
		cipher := int(cipherBytes[i] & 0xFF)
		plain := int(cipher ^ r>>8)
		if i >= n {
			plainBytes[i-n] = byte(plain)
		}
		r = (cipher+r)*c1 + c2&0xffff
	}
	return plainBytes
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

		if _, isHex := tk.IsHexChar(by); by != 0x0a && by != 0x0d && by != 0x20 && by != '\t' && !isHex {
			return true
		}
	}
	return false
}

//func (p *Parser) byte[] hexToBinary(byte[] bytes)
// 	 {
// 		 // calculate needed length
// 		 int len = 0;
// 		 for (byte by : bytes)
// 		 {
// 			 if (Character.digit((char) by, 16) != -1)
// 			 {
// 				 ++len;
// 			 }
// 		 }
// 		 byte[] res = new byte[len / 2];
// 		 int r = 0;
// 		 int prev = -1;
// 		 for (byte by : bytes)
// 		 {
// 			 int digit = Character.digit((char) by, 16);
// 			 if (digit != -1)
// 			 {
// 				 if (prev == -1)
// 				 {
// 					 prev = digit;
// 				 }
// 				 else
// 				 {
// 					 res[r++] = (byte) (prev * 16 + digit);
// 					 prev = -1;
// 				 }
// 			 }
// 		 }
// 		 return res;
// 	 }
//  }
