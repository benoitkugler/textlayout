package fontconfig

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

// ported from fontconfig/src/fcname.c Copyright © 2000 Keith Packard

// used to identify a type
type typeMeta interface {
	parse(str string, object Object) (Value, error)
}

type objectType struct {
	typeInfo typeMeta
	object   Object
}

var objects = map[string]objectType{
	objectNames[FAMILY]:          {object: FAMILY, typeInfo: typeString{}},          // String
	objectNames[FAMILYLANG]:      {object: FAMILYLANG, typeInfo: typeString{}},      // String
	objectNames[STYLE]:           {object: STYLE, typeInfo: typeString{}},           // String
	objectNames[STYLELANG]:       {object: STYLELANG, typeInfo: typeString{}},       // String
	objectNames[FULLNAME]:        {object: FULLNAME, typeInfo: typeString{}},        // String
	objectNames[FULLNAMELANG]:    {object: FULLNAMELANG, typeInfo: typeString{}},    // String
	objectNames[SLANT]:           {object: SLANT, typeInfo: typeInt{}},              // Int
	objectNames[WEIGHT]:          {object: WEIGHT, typeInfo: typeRange{}},           // Range
	objectNames[WIDTH]:           {object: WIDTH, typeInfo: typeRange{}},            // Range
	objectNames[SIZE]:            {object: SIZE, typeInfo: typeRange{}},             // Range
	objectNames[ASPECT]:          {object: ASPECT, typeInfo: typeFloat{}},           // Double
	objectNames[PIXEL_SIZE]:      {object: PIXEL_SIZE, typeInfo: typeFloat{}},       // Double
	objectNames[SPACING]:         {object: SPACING, typeInfo: typeInt{}},            // Int
	objectNames[FOUNDRY]:         {object: FOUNDRY, typeInfo: typeString{}},         // String
	objectNames[ANTIALIAS]:       {object: ANTIALIAS, typeInfo: typeBool{}},         // Bool
	objectNames[HINT_STYLE]:      {object: HINT_STYLE, typeInfo: typeInt{}},         // Int
	objectNames[HINTING]:         {object: HINTING, typeInfo: typeBool{}},           // Bool
	objectNames[VERTICAL_LAYOUT]: {object: VERTICAL_LAYOUT, typeInfo: typeBool{}},   // Bool
	objectNames[AUTOHINT]:        {object: AUTOHINT, typeInfo: typeBool{}},          // Bool
	objectNames[GLOBAL_ADVANCE]:  {object: GLOBAL_ADVANCE, typeInfo: typeBool{}},    // Bool
	objectNames[FILE]:            {object: FILE, typeInfo: typeString{}},            // String
	objectNames[INDEX]:           {object: INDEX, typeInfo: typeInt{}},              // Int
	objectNames[RASTERIZER]:      {object: RASTERIZER, typeInfo: typeString{}},      // String
	objectNames[OUTLINE]:         {object: OUTLINE, typeInfo: typeBool{}},           // Bool
	objectNames[SCALABLE]:        {object: SCALABLE, typeInfo: typeBool{}},          // Bool
	objectNames[DPI]:             {object: DPI, typeInfo: typeFloat{}},              // Double
	objectNames[RGBA]:            {object: RGBA, typeInfo: typeInt{}},               // Int
	objectNames[SCALE]:           {object: SCALE, typeInfo: typeFloat{}},            // Double
	objectNames[MINSPACE]:        {object: MINSPACE, typeInfo: typeBool{}},          // Bool
	objectNames[CHARWIDTH]:       {object: CHARWIDTH, typeInfo: typeInt{}},          // Int
	objectNames[CHAR_HEIGHT]:     {object: CHAR_HEIGHT, typeInfo: typeInt{}},        // Int
	objectNames[MATRIX]:          {object: MATRIX, typeInfo: typeMatrix{}},          // Matrix
	objectNames[CHARSET]:         {object: CHARSET, typeInfo: typeCharSet{}},        // CharSet
	objectNames[LANG]:            {object: LANG, typeInfo: typeLangSet{}},           // LangSet
	objectNames[FONTVERSION]:     {object: FONTVERSION, typeInfo: typeInt{}},        // Int
	objectNames[CAPABILITY]:      {object: CAPABILITY, typeInfo: typeString{}},      // String
	objectNames[FONTFORMAT]:      {object: FONTFORMAT, typeInfo: typeString{}},      // String
	objectNames[EMBOLDEN]:        {object: EMBOLDEN, typeInfo: typeBool{}},          // Bool
	objectNames[EMBEDDED_BITMAP]: {object: EMBEDDED_BITMAP, typeInfo: typeBool{}},   // Bool
	objectNames[DECORATIVE]:      {object: DECORATIVE, typeInfo: typeBool{}},        // Bool
	objectNames[LCD_FILTER]:      {object: LCD_FILTER, typeInfo: typeInt{}},         // Int
	objectNames[NAMELANG]:        {object: NAMELANG, typeInfo: typeString{}},        // String
	objectNames[FONT_FEATURES]:   {object: FONT_FEATURES, typeInfo: typeString{}},   // String
	objectNames[PRGNAME]:         {object: PRGNAME, typeInfo: typeString{}},         // String
	objectNames[HASH]:            {object: HASH, typeInfo: typeString{}},            // String
	objectNames[POSTSCRIPT_NAME]: {object: POSTSCRIPT_NAME, typeInfo: typeString{}}, // String
	objectNames[COLOR]:           {object: COLOR, typeInfo: typeBool{}},             // Bool
	objectNames[SYMBOL]:          {object: SYMBOL, typeInfo: typeBool{}},            // Bool
	objectNames[FONT_VARIATIONS]: {object: FONT_VARIATIONS, typeInfo: typeString{}}, // String
	objectNames[VARIABLE]:        {object: VARIABLE, typeInfo: typeBool{}},          // Bool
	objectNames[FONT_HAS_HINT]:   {object: FONT_HAS_HINT, typeInfo: typeBool{}},     // Bool
	objectNames[ORDER]:           {object: ORDER, typeInfo: typeInt{}},              // Int
}

var objectNames = [...]string{
	invalid:         "<invalid>",
	FAMILY:          "family",
	FAMILYLANG:      "familylang",
	STYLE:           "style",
	STYLELANG:       "stylelang",
	FULLNAME:        "fullname",
	FULLNAMELANG:    "fullnamelang",
	SLANT:           "slant",
	WEIGHT:          "weight",
	WIDTH:           "width",
	SIZE:            "size",
	ASPECT:          "aspect",
	PIXEL_SIZE:      "pixelsize",
	SPACING:         "spacing",
	FOUNDRY:         "foundry",
	ANTIALIAS:       "antialias",
	HINT_STYLE:      "hintstyle",
	HINTING:         "hinting",
	VERTICAL_LAYOUT: "verticallayout",
	AUTOHINT:        "autohint",
	GLOBAL_ADVANCE:  "globaladvance",
	FILE:            "file",
	INDEX:           "index",
	RASTERIZER:      "rasterizer",
	OUTLINE:         "outline",
	SCALABLE:        "scalable",
	DPI:             "dpi",
	RGBA:            "rgba",
	SCALE:           "scale",
	MINSPACE:        "minspace",
	CHARWIDTH:       "charwidth",
	CHAR_HEIGHT:     "charheight",
	MATRIX:          "matrix",
	CHARSET:         "charset",
	LANG:            "lang",
	FONTVERSION:     "fontversion",
	CAPABILITY:      "capability",
	FONTFORMAT:      "fontformat",
	EMBOLDEN:        "embolden",
	EMBEDDED_BITMAP: "embeddedbitmap",
	DECORATIVE:      "decorative",
	LCD_FILTER:      "lcdfilter",
	NAMELANG:        "namelang",
	FONT_FEATURES:   "fontfeatures",
	PRGNAME:         "prgname",
	HASH:            "hash",
	POSTSCRIPT_NAME: "postscriptname",
	COLOR:           "color",
	SYMBOL:          "symbol",
	FONT_VARIATIONS: "fontvariations",
	VARIABLE:        "variable",
	FONT_HAS_HINT:   "fonthashint",
	ORDER:           "order",
}

func (object Object) String() string {
	if int(object) < len(objectNames) { // common case for buitlin objects
		return objectNames[object]
	}
	return fmt.Sprintf("<custom_object_%d>", object)
}

// // FromString lookup an object from its string value,
// // both for builtin and custom objects.
// // The zero value is returned for unknown objects.
// func FromString(object string) Object {
// 	if builtin, ok := objects[object]; ok {
// 		return builtin.object
// 	}
// 	if o, ok := customObjects[object]; ok {
// 		return o
// 	}
// 	return invalid
// }

// the + 20 is to leave some room for future added internal objects
const nextId = FirstCustomObject + 20

func (c *Config) lookupCustomObject(object string) objectType {
	if o, ok := c.customObjects[object]; ok {
		return objectType{object: o} // parser is nil for unknown type
	}

	// we add new objects
	id := nextId
	for _, o := range c.customObjects {
		if o > id {
			id = o
		}
	}
	if id+1 == 0 {
		panic("implementation limit for the number of custom objects reached")
	}
	c.customObjects[object] = id + 1
	return objectType{object: id + 1}
}

// Return the object type for the pattern element named object
// Add a custom object if not found
func (c *Config) getRegisterObjectType(object string) objectType {
	if builtin, ok := objects[object]; ok {
		return builtin
	}
	return c.lookupCustomObject(object)
}

type constant struct {
	name   string
	object Object
	value  int
}

var baseConstants = [...]constant{
	{"thin", WEIGHT, WEIGHT_THIN},
	{"extralight", WEIGHT, WEIGHT_EXTRALIGHT},
	{"ultralight", WEIGHT, WEIGHT_EXTRALIGHT},
	{"demilight", WEIGHT, WEIGHT_DEMILIGHT},
	{"semilight", WEIGHT, WEIGHT_DEMILIGHT},
	{"light", WEIGHT, WEIGHT_LIGHT},
	{"book", WEIGHT, WEIGHT_BOOK},
	{"regular", WEIGHT, WEIGHT_REGULAR},
	{"medium", WEIGHT, WEIGHT_MEDIUM},
	{"demibold", WEIGHT, WEIGHT_DEMIBOLD},
	{"semibold", WEIGHT, WEIGHT_DEMIBOLD},
	{"bold", WEIGHT, WEIGHT_BOLD},
	{"extrabold", WEIGHT, WEIGHT_EXTRABOLD},
	{"ultrabold", WEIGHT, WEIGHT_EXTRABOLD},
	{"black", WEIGHT, WEIGHT_BLACK},
	{"heavy", WEIGHT, WEIGHT_HEAVY},

	{"roman", SLANT, SLANT_ROMAN},
	{"italic", SLANT, SLANT_ITALIC},
	{"oblique", SLANT, SLANT_OBLIQUE},

	{"ultracondensed", WIDTH, WIDTH_ULTRACONDENSED},
	{"extracondensed", WIDTH, WIDTH_EXTRACONDENSED},
	{"condensed", WIDTH, WIDTH_CONDENSED},
	{"semicondensed", WIDTH, WIDTH_SEMICONDENSED},
	{"normal", WIDTH, WIDTH_NORMAL},
	{"semiexpanded", WIDTH, WIDTH_SEMIEXPANDED},
	{"expanded", WIDTH, WIDTH_EXPANDED},
	{"extraexpanded", WIDTH, WIDTH_EXTRAEXPANDED},
	{"ultraexpanded", WIDTH, WIDTH_ULTRAEXPANDED},

	{"proportional", SPACING, PROPORTIONAL},
	{"dual", SPACING, DUAL},
	{"mono", SPACING, MONO},
	{"charcell", SPACING, CHARCELL},

	{"unknown", RGBA, RGBA_UNKNOWN},
	{"rgb", RGBA, RGBA_RGB},
	{"bgr", RGBA, RGBA_BGR},
	{"vrgb", RGBA, RGBA_VRGB},
	{"vbgr", RGBA, RGBA_VBGR},
	{"none", RGBA, RGBA_NONE},

	{"hintnone", HINT_STYLE, HINT_NONE},
	{"hintslight", HINT_STYLE, HINT_SLIGHT},
	{"hintmedium", HINT_STYLE, HINT_MEDIUM},
	{"hintfull", HINT_STYLE, HINT_FULL},

	{"antialias", ANTIALIAS, 1},
	{"hinting", HINTING, 1},
	{"verticallayout", VERTICAL_LAYOUT, 1},
	{"autohint", AUTOHINT, 1},
	{"globaladvance", GLOBAL_ADVANCE, 1}, /* deprecated */
	{"outline", OUTLINE, 1},
	{"scalable", SCALABLE, 1},
	{"minspace", MINSPACE, 1},
	{"embolden", EMBOLDEN, 1},
	{"embeddedbitmap", EMBEDDED_BITMAP, 1},
	{"decorative", DECORATIVE, 1},

	{"lcdnone", LCD_FILTER, LCD_NONE},
	{"lcddefault", LCD_FILTER, LCD_DEFAULT},
	{"lcdlight", LCD_FILTER, LCD_LIGHT},
	{"lcdlegacy", LCD_FILTER, LCD_LEGACY},
}

func nameGetConstant(str string) *constant {
	for i := range baseConstants {
		if cmpIgnoreCase(str, baseConstants[i].name) == 0 {
			return &baseConstants[i]
		}
	}
	return nil
}

func nameConstant(str String) (int, bool) {
	if c := nameGetConstant(string(str)); c != nil {
		return c.value, true
	}
	return 0, false
}

// parseName converts `name` from the standard text format Described above into a pattern.
func (c *Config) parseName(name []byte) (Pattern, error) {
	var (
		delim byte
		save  string
		pat   = NewPattern()
	)

	for {
		delim, name, save = nameFindNext(name, "-,:")
		if len(save) != 0 {
			pat.Add(FAMILY, String(save), true)
		}
		if delim != ',' {
			break
		}
	}
	if delim == '-' {
		for {
			delim, name, save = nameFindNext(name, "-,:")
			d, err := strconv.ParseFloat(save, 64)
			if err == nil {
				pat.Add(SIZE, Float(d), true)
			}
			if delim != ',' {
				break
			}
		}
	}
	for delim == ':' {
		delim, name, save = nameFindNext(name, "=_:")
		if len(save) != 0 {
			if delim == '=' || delim == '_' {
				t := c.getRegisterObjectType(save)
				for {
					delim, name, save = nameFindNext(name, ":,")
					v, err := t.typeInfo.parse(save, t.object)
					if err != nil {
						return nil, err
					}
					pat.Add(t.object, v, true)
					if delim != ',' {
						break
					}
				}
			} else {
				if co := nameGetConstant(save); c != nil {
					t := c.getRegisterObjectType(objectNames[co.object])

					switch t.typeInfo.(type) {
					case typeInt, typeFloat, typeRange:
						pat.Add(co.object, Int(co.value), true)
					case typeBool:
						pat.Add(co.object, Bool(co.value), true)
					}
				}
			}
		}
	}

	return pat, nil
}

func nameFindNext(cur []byte, delim string) (byte, []byte, string) {
	cur = bytes.TrimLeftFunc(cur, unicode.IsSpace)
	i := 0
	var save []byte
	for i < len(cur) {
		if cur[i] == '\\' {
			i++
			if i == len(cur) {
				break
			}
		} else if strings.IndexByte(delim, cur[i]) != -1 {
			break
		}
		save = append(save, cur[i])
		i++
	}
	var last byte
	if i < len(cur) {
		last = cur[i]
		i++
	}
	return last, cur[i:], string(save)
}

func constantWithObjectCheck(str string, object Object) (int, bool, error) {
	c := nameGetConstant(str)
	if c != nil {
		if c.object != object {
			return 0, false, fmt.Errorf("fontconfig : unexpected constant name %s used for object %s: should be %s\n", str, object, c.object)
		}
		return c.value, true, nil
	}
	return 0, false, nil
}

func nameBool(v string) (Bool, error) {
	c0 := toLower(v)
	if c0 == 't' || c0 == 'y' || c0 == '1' {
		return True, nil
	}
	if c0 == 'f' || c0 == 'n' || c0 == '0' {
		return False, nil
	}
	if c0 == 'd' || c0 == 'x' || c0 == '2' {
		return DontCare, nil
	}
	if c0 == 'o' {
		c1 := toLower(v[1:])
		if c1 == 'n' {
			return True, nil
		}
		if c1 == 'f' {
			return False, nil
		}
		if c1 == 'r' {
			return DontCare, nil
		}
	}
	return 0, fmt.Errorf("fontconfig: unknown boolean %s", v)
}

type typeInt struct{}

func (typeInt) parse(str string, object Object) (Value, error) {
	v, builtin, err := constantWithObjectCheck(str, object)
	if err != nil {
		return nil, err
	}
	if !builtin {
		v, err = strconv.Atoi(str)
	}
	return Int(v), err
}

type typeString struct{}

func (typeString) parse(str string, object Object) (Value, error) { return String(str), nil }

type typeBool struct{}

func (typeBool) parse(str string, object Object) (Value, error) { return nameBool(str) }

type typeFloat struct{}

func (typeFloat) parse(str string, object Object) (Value, error) {
	d, err := strconv.ParseFloat(str, 64)
	return Float(d), err
}

type typeMatrix struct{}

func (typeMatrix) parse(str string, object Object) (Value, error) {
	var m Matrix
	_, err := fmt.Sscanf(str, "%g %g %g %g", &m.Xx, &m.Xy, &m.Yx, &m.Yy)
	return m, err
}

type typeCharSet struct{}

func (typeCharSet) parse(str string, object Object) (Value, error) {
	return parseCharSet(str)
}

type typeLangSet struct{}

func (typeLangSet) parse(str string, object Object) (Value, error) {
	return NewLangset(str), nil
}

type typeRange struct{}

func (typeRange) parse(str string, object Object) (Value, error) {
	var b, e float32
	n, _ := fmt.Sscanf(str, "[%g %g]", &b, &e)
	if n == 2 {
		return Range{Begin: b, End: e}, nil
	}

	var sc, ec string
	n, _ = fmt.Sscanf(strings.TrimSuffix(str, "]"), "[%s %s", &sc, &ec)
	if n == 2 {
		si, oks, err := constantWithObjectCheck(sc, object)
		if err != nil {
			return nil, err
		}
		ei, oke, err := constantWithObjectCheck(ec, object)
		if err != nil {
			return nil, err
		}
		if oks && oke {
			return Range{Begin: float32(si), End: float32(ei)}, nil
		}
	}

	si, ok, err := constantWithObjectCheck(str, object)
	if err != nil {
		return nil, err
	}
	if ok {
		return Float(si), nil
	}
	v, err := strconv.ParseFloat(str, 64)
	if err != nil {
		return nil, err
	}
	return Float(v), nil
}
