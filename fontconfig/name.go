package fontconfig

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"unicode"
)

// ported from fontconfig/src/fcname.c Copyright © 2000 Keith Packard

// used to identify a type
type typeMeta interface {
	parse(str string, object Object) (Value, error)
}

type objectType struct {
	object Object
	parser typeMeta
}

var objects = map[string]objectType{
	objectNames[FC_FAMILY]:          {object: FC_FAMILY, parser: typeString{}},          // String
	objectNames[FC_FAMILYLANG]:      {object: FC_FAMILYLANG, parser: typeString{}},      // String
	objectNames[FC_STYLE]:           {object: FC_STYLE, parser: typeString{}},           // String
	objectNames[FC_STYLELANG]:       {object: FC_STYLELANG, parser: typeString{}},       // String
	objectNames[FC_FULLNAME]:        {object: FC_FULLNAME, parser: typeString{}},        // String
	objectNames[FC_FULLNAMELANG]:    {object: FC_FULLNAMELANG, parser: typeString{}},    // String
	objectNames[FC_SLANT]:           {object: FC_SLANT, parser: typeInteger{}},          // Integer
	objectNames[FC_WEIGHT]:          {object: FC_WEIGHT, parser: typeRange{}},           // Range
	objectNames[FC_WIDTH]:           {object: FC_WIDTH, parser: typeRange{}},            // Range
	objectNames[FC_SIZE]:            {object: FC_SIZE, parser: typeRange{}},             // Range
	objectNames[FC_ASPECT]:          {object: FC_ASPECT, parser: typeFloat{}},           // Double
	objectNames[FC_PIXEL_SIZE]:      {object: FC_PIXEL_SIZE, parser: typeFloat{}},       // Double
	objectNames[FC_SPACING]:         {object: FC_SPACING, parser: typeInteger{}},        // Integer
	objectNames[FC_FOUNDRY]:         {object: FC_FOUNDRY, parser: typeString{}},         // String
	objectNames[FC_ANTIALIAS]:       {object: FC_ANTIALIAS, parser: typeBool{}},         // Bool
	objectNames[FC_HINT_STYLE]:      {object: FC_HINT_STYLE, parser: typeInteger{}},     // Integer
	objectNames[FC_HINTING]:         {object: FC_HINTING, parser: typeBool{}},           // Bool
	objectNames[FC_VERTICAL_LAYOUT]: {object: FC_VERTICAL_LAYOUT, parser: typeBool{}},   // Bool
	objectNames[FC_AUTOHINT]:        {object: FC_AUTOHINT, parser: typeBool{}},          // Bool
	objectNames[FC_GLOBAL_ADVANCE]:  {object: FC_GLOBAL_ADVANCE, parser: typeBool{}},    // Bool
	objectNames[FC_FILE]:            {object: FC_FILE, parser: typeString{}},            // String
	objectNames[FC_INDEX]:           {object: FC_INDEX, parser: typeInteger{}},          // Integer
	objectNames[FC_RASTERIZER]:      {object: FC_RASTERIZER, parser: typeString{}},      // String
	objectNames[FC_OUTLINE]:         {object: FC_OUTLINE, parser: typeBool{}},           // Bool
	objectNames[FC_SCALABLE]:        {object: FC_SCALABLE, parser: typeBool{}},          // Bool
	objectNames[FC_DPI]:             {object: FC_DPI, parser: typeFloat{}},              // Double
	objectNames[FC_RGBA]:            {object: FC_RGBA, parser: typeInteger{}},           // Integer
	objectNames[FC_SCALE]:           {object: FC_SCALE, parser: typeFloat{}},            // Double
	objectNames[FC_MINSPACE]:        {object: FC_MINSPACE, parser: typeBool{}},          // Bool
	objectNames[FC_CHARWIDTH]:       {object: FC_CHARWIDTH, parser: typeInteger{}},      // Integer
	objectNames[FC_CHAR_HEIGHT]:     {object: FC_CHAR_HEIGHT, parser: typeInteger{}},    // Integer
	objectNames[FC_MATRIX]:          {object: FC_MATRIX, parser: typeMatrix{}},          // Matrix
	objectNames[FC_CHARSET]:         {object: FC_CHARSET, parser: typeCharSet{}},        // CharSet
	objectNames[FC_LANG]:            {object: FC_LANG, parser: typeLangSet{}},           // LangSet
	objectNames[FC_FONTVERSION]:     {object: FC_FONTVERSION, parser: typeInteger{}},    // Integer
	objectNames[FC_CAPABILITY]:      {object: FC_CAPABILITY, parser: typeString{}},      // String
	objectNames[FC_FONTFORMAT]:      {object: FC_FONTFORMAT, parser: typeString{}},      // String
	objectNames[FC_EMBOLDEN]:        {object: FC_EMBOLDEN, parser: typeBool{}},          // Bool
	objectNames[FC_EMBEDDED_BITMAP]: {object: FC_EMBEDDED_BITMAP, parser: typeBool{}},   // Bool
	objectNames[FC_DECORATIVE]:      {object: FC_DECORATIVE, parser: typeBool{}},        // Bool
	objectNames[FC_LCD_FILTER]:      {object: FC_LCD_FILTER, parser: typeInteger{}},     // Integer
	objectNames[FC_NAMELANG]:        {object: FC_NAMELANG, parser: typeString{}},        // String
	objectNames[FC_FONT_FEATURES]:   {object: FC_FONT_FEATURES, parser: typeString{}},   // String
	objectNames[FC_PRGNAME]:         {object: FC_PRGNAME, parser: typeString{}},         // String
	objectNames[FC_HASH]:            {object: FC_HASH, parser: typeString{}},            // String
	objectNames[FC_POSTSCRIPT_NAME]: {object: FC_POSTSCRIPT_NAME, parser: typeString{}}, // String
	objectNames[FC_COLOR]:           {object: FC_COLOR, parser: typeBool{}},             // Bool
	objectNames[FC_SYMBOL]:          {object: FC_SYMBOL, parser: typeBool{}},            // Bool
	objectNames[FC_FONT_VARIATIONS]: {object: FC_FONT_VARIATIONS, parser: typeString{}}, // String
	objectNames[FC_VARIABLE]:        {object: FC_VARIABLE, parser: typeBool{}},          // Bool
	objectNames[FC_FONT_HAS_HINT]:   {object: FC_FONT_HAS_HINT, parser: typeBool{}},     // Bool
	objectNames[FC_ORDER]:           {object: FC_ORDER, parser: typeInteger{}},          // Integer
}

var objectNames = [...]string{
	FC_FAMILY:          "family",
	FC_FAMILYLANG:      "familylang",
	FC_STYLE:           "style",
	FC_STYLELANG:       "stylelang",
	FC_FULLNAME:        "fullname",
	FC_FULLNAMELANG:    "fullnamelang",
	FC_SLANT:           "slant",
	FC_WEIGHT:          "weight",
	FC_WIDTH:           "width",
	FC_SIZE:            "size",
	FC_ASPECT:          "aspect",
	FC_PIXEL_SIZE:      "pixelsize",
	FC_SPACING:         "spacing",
	FC_FOUNDRY:         "foundry",
	FC_ANTIALIAS:       "antialias",
	FC_HINT_STYLE:      "hintstyle",
	FC_HINTING:         "hinting",
	FC_VERTICAL_LAYOUT: "verticallayout",
	FC_AUTOHINT:        "autohint",
	FC_GLOBAL_ADVANCE:  "globaladvance",
	FC_FILE:            "file",
	FC_INDEX:           "index",
	FC_RASTERIZER:      "rasterizer",
	FC_OUTLINE:         "outline",
	FC_SCALABLE:        "scalable",
	FC_DPI:             "dpi",
	FC_RGBA:            "rgba",
	FC_SCALE:           "scale",
	FC_MINSPACE:        "minspace",
	FC_CHARWIDTH:       "charwidth",
	FC_CHAR_HEIGHT:     "charheight",
	FC_MATRIX:          "matrix",
	FC_CHARSET:         "charset",
	FC_LANG:            "lang",
	FC_FONTVERSION:     "fontversion",
	FC_CAPABILITY:      "capability",
	FC_FONTFORMAT:      "fontformat",
	FC_EMBOLDEN:        "embolden",
	FC_EMBEDDED_BITMAP: "embeddedbitmap",
	FC_DECORATIVE:      "decorative",
	FC_LCD_FILTER:      "lcdfilter",
	FC_NAMELANG:        "namelang",
	FC_FONT_FEATURES:   "fontfeatures",
	FC_PRGNAME:         "prgname",
	FC_HASH:            "hash",
	FC_POSTSCRIPT_NAME: "postscriptname",
	FC_COLOR:           "color",
	FC_SYMBOL:          "symbol",
	FC_FONT_VARIATIONS: "fontvariations",
	FC_VARIABLE:        "variable",
	FC_FONT_HAS_HINT:   "fonthashint",
	FC_ORDER:           "order",
}

func (object Object) String() string {
	if int(object) < len(objectNames) { // common case for buitlin objects
		return objectNames[object]
	}
	customObjectsLock.Lock()
	defer customObjectsLock.Unlock()
	for name, o := range customObjects {
		if o == object {
			return name
		}
	}
	return fmt.Sprintf("invalid_object_%d", object)
}

// FromString lookup an object from its string value,
// both for builtin and custom objects.
// FC_INVALID is returned for unknown objects
func FromString(object string) Object {
	if builtin, ok := objects[object]; ok {
		return builtin.object
	}
	if o, ok := customObjects[object]; ok {
		return o
	}
	return FC_INVALID
}

// the + 100 is to leave some room for future added internal objects
const firstCustomObject = FirstCustomObject + 100

var (
	// the name is used defined, and the object assigned by the library
	customObjects     = map[string]Object{}
	customObjectsLock sync.Mutex
)

func lookupCustomObject(object string) objectType {
	customObjectsLock.Lock()
	defer customObjectsLock.Unlock()

	if o, ok := customObjects[object]; ok {
		return objectType{object: o} // parser is nil for unknown type
	}

	// we add new objects
	id := firstCustomObject
	for _, o := range customObjects {
		if o > id {
			id = o
		}
	}
	if id+1 == 0 {
		panic("implementation limit for the number of custom objects reached")
	}
	customObjects[object] = id + 1
	return objectType{object: id + 1}
}

// Return the object type for the pattern element named object
// Add a custom object if not found
func getRegisterObjectType(object string) objectType {
	if builtin, ok := objects[object]; ok {
		return builtin
	}
	return lookupCustomObject(object)
}

//  Bool
//  hasValidType (FcObject object, FcType type)
//  {
// 	 const FcObjectType    *t = FcObjectFindById (object);

// 	 if (t) {
// 	 switch ((int) t.type) {
// 	 case FcTypeUnknown:
// 		 return true;
// 	 case FcTypeDouble:
// 	 case FcTypeInteger:
// 		 if (type == FcTypeDouble || type == FcTypeInteger)
// 		 return true;
// 		 break;
// 	 case FcTypeLangSet:
// 		 if (type == FcTypeLangSet || type == FcTypeString)
// 		 return true;
// 		 break;
// 	 case FcTypeRange:
// 		 if (type == FcTypeRange ||
// 		 type == FcTypeDouble ||
// 		 type == FcTypeInteger)
// 		 return true;
// 		 break;
// 	 default:
// 		 if (type == t.type)
// 		 return true;
// 		 break;
// 	 }
// 	 return false;
// 	 }
// 	 return true;
//  }

//  FcObject
//  FcObjectFromName (const char * name)
//  {
// 	 return FcObjectLookupIdByName (name);
//  }

//  FcObjectSet *
//  FcObjectGetSet (void)
//  {
// 	 int		i;
// 	 FcObjectSet	*os = NULL;

// 	 os = FcObjectSetCreate ();
// 	 for (i = 0; i < NUM_OBJECT_TYPES; i++)
// 	 FcObjectSetAdd (os, FcObjects[i].object);

// 	 return os;
//  }

//  const char *
//  FcObjectName (FcObject object)
//  {
// 	 const FcObjectType   *o = FcObjectFindById (object);

// 	 if (o)
// 	 return o.object;

// 	 return FcObjectLookupOtherNameById (object);
//  }

type constant struct {
	name   string
	object Object
	value  int
}

var baseConstants = [...]constant{
	{"thin", FC_WEIGHT, WEIGHT_THIN},
	{"extralight", FC_WEIGHT, WEIGHT_EXTRALIGHT},
	{"ultralight", FC_WEIGHT, WEIGHT_EXTRALIGHT},
	{"demilight", FC_WEIGHT, WEIGHT_DEMILIGHT},
	{"semilight", FC_WEIGHT, WEIGHT_DEMILIGHT},
	{"light", FC_WEIGHT, WEIGHT_LIGHT},
	{"book", FC_WEIGHT, WEIGHT_BOOK},
	{"regular", FC_WEIGHT, WEIGHT_REGULAR},
	{"medium", FC_WEIGHT, WEIGHT_MEDIUM},
	{"demibold", FC_WEIGHT, WEIGHT_DEMIBOLD},
	{"semibold", FC_WEIGHT, WEIGHT_DEMIBOLD},
	{"bold", FC_WEIGHT, WEIGHT_BOLD},
	{"extrabold", FC_WEIGHT, WEIGHT_EXTRABOLD},
	{"ultrabold", FC_WEIGHT, WEIGHT_EXTRABOLD},
	{"black", FC_WEIGHT, WEIGHT_BLACK},
	{"heavy", FC_WEIGHT, WEIGHT_HEAVY},

	{"roman", FC_SLANT, SLANT_ROMAN},
	{"italic", FC_SLANT, SLANT_ITALIC},
	{"oblique", FC_SLANT, SLANT_OBLIQUE},

	{"ultracondensed", FC_WIDTH, WIDTH_ULTRACONDENSED},
	{"extracondensed", FC_WIDTH, WIDTH_EXTRACONDENSED},
	{"condensed", FC_WIDTH, WIDTH_CONDENSED},
	{"semicondensed", FC_WIDTH, WIDTH_SEMICONDENSED},
	{"normal", FC_WIDTH, WIDTH_NORMAL},
	{"semiexpanded", FC_WIDTH, WIDTH_SEMIEXPANDED},
	{"expanded", FC_WIDTH, WIDTH_EXPANDED},
	{"extraexpanded", FC_WIDTH, WIDTH_EXTRAEXPANDED},
	{"ultraexpanded", FC_WIDTH, WIDTH_ULTRAEXPANDED},

	{"proportional", FC_SPACING, FC_PROPORTIONAL},
	{"dual", FC_SPACING, FC_DUAL},
	{"mono", FC_SPACING, FC_MONO},
	{"charcell", FC_SPACING, FC_CHARCELL},

	{"unknown", FC_RGBA, FC_RGBA_UNKNOWN},
	{"rgb", FC_RGBA, FC_RGBA_RGB},
	{"bgr", FC_RGBA, FC_RGBA_BGR},
	{"vrgb", FC_RGBA, FC_RGBA_VRGB},
	{"vbgr", FC_RGBA, FC_RGBA_VBGR},
	{"none", FC_RGBA, FC_RGBA_NONE},

	{"hintnone", FC_HINT_STYLE, FC_HINT_NONE},
	{"hintslight", FC_HINT_STYLE, FC_HINT_SLIGHT},
	{"hintmedium", FC_HINT_STYLE, FC_HINT_MEDIUM},
	{"hintfull", FC_HINT_STYLE, FC_HINT_FULL},

	{"antialias", FC_ANTIALIAS, 1},
	{"hinting", FC_HINTING, 1},
	{"verticallayout", FC_VERTICAL_LAYOUT, 1},
	{"autohint", FC_AUTOHINT, 1},
	{"globaladvance", FC_GLOBAL_ADVANCE, 1}, /* deprecated */
	{"outline", FC_OUTLINE, 1},
	{"scalable", FC_SCALABLE, 1},
	{"minspace", FC_MINSPACE, 1},
	{"embolden", FC_EMBOLDEN, 1},
	{"embeddedbitmap", FC_EMBEDDED_BITMAP, 1},
	{"decorative", FC_DECORATIVE, 1},

	{"lcdnone", FC_LCD_FILTER, FC_LCD_NONE},
	{"lcddefault", FC_LCD_FILTER, FC_LCD_DEFAULT},
	{"lcdlight", FC_LCD_FILTER, FC_LCD_LIGHT},
	{"lcdlegacy", FC_LCD_FILTER, FC_LCD_LEGACY},
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

// FcNameParse converts `name` from the standard text format Described above into a pattern.
func FcNameParse(name []byte) (Pattern, error) {
	var (
		delim byte
		save  string
		pat   = NewPattern()
	)

	for {
		delim, name, save = nameFindNext(name, "-,:")
		if len(save) != 0 {
			pat.Add(FC_FAMILY, String(save), true)
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
				pat.Add(FC_SIZE, Float(d), true)
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
				t := getRegisterObjectType(save)
				for {
					delim, name, save = nameFindNext(name, ":,")
					v, err := t.parser.parse(save, t.object)
					if err != nil {
						return nil, err
					}
					pat.Add(t.object, v, true)
					if delim != ',' {
						break
					}
				}
			} else {
				if c := nameGetConstant(save); c != nil {
					t := getRegisterObjectType(objectNames[c.object])

					switch t.parser.(type) {
					case typeInteger, typeFloat, typeRange:
						pat.Add(c.object, Int(c.value), true)
					case typeBool:
						pat.Add(c.object, Bool(c.value), true)
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
	c0 := FcToLower(v)
	if c0 == 't' || c0 == 'y' || c0 == '1' {
		return FcTrue, nil
	}
	if c0 == 'f' || c0 == 'n' || c0 == '0' {
		return FcFalse, nil
	}
	if c0 == 'd' || c0 == 'x' || c0 == '2' {
		return FcDontCare, nil
	}
	if c0 == 'o' {
		c1 := FcToLower(v[1:])
		if c1 == 'n' {
			return FcTrue, nil
		}
		if c1 == 'f' {
			return FcFalse, nil
		}
		if c1 == 'r' {
			return FcDontCare, nil
		}
	}
	return 0, fmt.Errorf("fontconfig: unknown boolean %s", v)
}

type typeInteger struct{}

func (typeInteger) parse(str string, object Object) (Value, error) {
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
	return FcNameParseCharSet(str)
}

type typeLangSet struct{}

func (typeLangSet) parse(str string, object Object) (Value, error) {
	return parseLangset(str), nil
}

type typeRange struct{}

func (typeRange) parse(str string, object Object) (Value, error) {
	var b, e float64
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
			return Range{Begin: float64(si), End: float64(ei)}, nil
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

//  static FcValue
//  FcNameConvert (FcType type, const char *object, FcChar8 *string)
//  {
// 	 FcValue	v;
// 	 FcMatrix	m;
// 	 double	b, e;
// 	 char	*p;

// 	 v.type = type;
// 	 switch ((int) v.type) {
// 	 case FcTypeInteger:
// 	 if (!constantWithObjectCheck (string, object, &v.u.i))
// 		 v.u.i = atoi ((char *) string);
// 	 break;
// 	 case FcTypeString:
// 	 v.u.s = FcStrdup (string);
// 	 if (!v.u.s)
// 		 v.type = FcTypeVoid;
// 	 break;
// 	 case FcTypeBool:
// 	 if (!nameBool (string, &v.u.b))
// 		 v.u.b = false;
// 	 break;
// 	 case FcTypeDouble:
// 	 v.u.d = strtod ((char *) string, 0);
// 	 break;
// 	 case FcTypeMatrix:
// 	 FcMatrixInit (&m);
// 	 sscanf ((char *) string, "%lg %lg %lg %lg", &m.xx, &m.xy, &m.yx, &m.yy);
// 	 v.u.m = FcMatrixCopy (&m);
// 	 break;
// 	 case FcTypeCharSet:
// 	 v.u.c = FcNameParseCharSet (string);
// 	 if (!v.u.c)
// 		 v.type = FcTypeVoid;
// 	 break;
// 	 case FcTypeLangSet:
// 	 v.u.l = FcNameParseLangSet (string);
// 	 if (!v.u.l)
// 		 v.type = FcTypeVoid;
// 	 break;
// 	 case FcTypeRange:
// 	 if (sscanf ((char *) string, "[%lg %lg]", &b, &e) != 2)
// 	 {
// 		 char *sc, *ec;
// 		 size_t len = strlen ((const char *) string);
// 		 int si, ei;

// 		 sc = malloc (len + 1);
// 		 ec = malloc (len + 1);
// 		 if (sc && ec && sscanf ((char *) string, "[%s %[^]]]", sc, ec) == 2)
// 		 {
// 		 if (constantWithObjectCheck ((const FcChar8 *) sc, object, &si) &&
// 			 constantWithObjectCheck ((const FcChar8 *) ec, object, &ei))
// 			 v.u.r =  FcRangeCreateDouble (si, ei);
// 		 else
// 			 goto bail1;
// 		 }
// 		 else
// 		 {
// 		 bail1:
// 		 v.type = FcTypeDouble;
// 		 if (constantWithObjectCheck (string, object, &si))
// 		 {
// 			 v.u.d = (double) si;
// 		 } else {
// 			 v.u.d = strtod ((char *) string, &p);
// 			 if (p != NULL && p[0] != 0)
// 			 v.type = FcTypeVoid;
// 		 }
// 		 }
// 		 if (sc)
// 		 free (sc);
// 		 if (ec)
// 		 free (ec);
// 	 }
// 	 else
// 		 v.u.r = FcRangeCreateDouble (b, e);
// 	 break;
// 	 default:
// 	 break;
// 	 }
// 	 return v;
//  }

//  static Bool
//  FcNameUnparseString (FcStrBuf	    *buf,
// 			  const FcChar8  *string,
// 			  const FcChar8  *escape)
//  {
// 	 FcChar8 c;
// 	 for ((c = *string++))
// 	 {
// 	 if (escape && strchr ((char *) escape, (char) c))
// 	 {
// 		 if (!FcStrBufChar (buf, escape[0]))
// 		 return false;
// 	 }
// 	 if (!FcStrBufChar (buf, c))
// 		 return false;
// 	 }
// 	 return true;
//  }

//  Bool
//  FcNameUnparseValue (FcStrBuf	*buf,
// 			 FcValue	*v0,
// 			 FcChar8	*escape)
//  {
// 	 FcChar8	temp[1024];
// 	 FcValue v = FcValueCanonicalize(v0);

// 	 switch (v.type) {
// 	 case FcTypeUnknown:
// 	 case FcTypeVoid:
// 	 return true;
// 	 case FcTypeInteger:
// 	 sprintf ((char *) temp, "%d", v.u.i);
// 	 return FcNameUnparseString (buf, temp, 0);
// 	 case FcTypeDouble:
// 	 sprintf ((char *) temp, "%g", v.u.d);
// 	 return FcNameUnparseString (buf, temp, 0);
// 	 case FcTypeString:
// 	 return FcNameUnparseString (buf, v.u.s, escape);
// 	 case FcTypeBool:
// 	 return FcNameUnparseString (buf,
// 					 v.u.b == true  ? (FcChar8 *) "True" :
// 					 v.u.b == false ? (FcChar8 *) "False" :
// 										(FcChar8 *) "DontCare", 0);
// 	 case FcTypeMatrix:
// 	 sprintf ((char *) temp, "%g %g %g %g",
// 		  v.u.m.xx, v.u.m.xy, v.u.m.yx, v.u.m.yy);
// 	 return FcNameUnparseString (buf, temp, 0);
// 	 case FcTypeCharSet:
// 	 return FcNameUnparseCharSet (buf, v.u.c);
// 	 case FcTypeLangSet:
// 	 return FcNameUnparseLangSet (buf, v.u.l);
// 	 case FcTypeFTFace:
// 	 return true;
// 	 case FcTypeRange:
// 	 sprintf ((char *) temp, "[%g %g]", v.u.r.begin, v.u.r.end);
// 	 return FcNameUnparseString (buf, temp, 0);
// 	 }
// 	 return false;
//  }

//  Bool
//  FcNameUnparseValueList (FcStrBuf	*buf,
// 			 FcValueListPtr	v,
// 			 FcChar8		*escape)
//  {
// 	 for (v)
// 	 {
// 	 if (!FcNameUnparseValue (buf, &v.value, escape))
// 		 return false;
// 	 if ((v = FcValueListNext(v)) != NULL)
// 		 if (!FcNameUnparseString (buf, (FcChar8 *) ",", 0))
// 		 return false;
// 	 }
// 	 return true;
//  }

//  #define FC_ESCAPE_FIXED    "\\-:,"
//  #define FC_ESCAPE_VARIABLE "\\=_:,"

//  FcChar8 *
//  FcNameUnparse (FcPattern *pat)
//  {
// 	 return FcNameUnparseEscaped (pat, true);
//  }

//  FcChar8 *
//  FcNameUnparseEscaped (FcPattern *pat, Bool escape)
//  {
// 	 FcStrBuf		    buf, buf2;
// 	 FcChar8		    buf_static[8192], buf2_static[256];
// 	 int			    i;
// 	 FcPatternElt	    *e;

// 	 FcStrBufInit (&buf, buf_static, sizeof (buf_static));
// 	 FcStrBufInit (&buf2, buf2_static, sizeof (buf2_static));
// 	 e = FcPatternObjectFindElt (pat, FC_FAMILY_OBJECT);
// 	 if (e)
// 	 {
// 		 if (!FcNameUnparseValueList (&buf, FcPatternEltValues(e), escape ? (FcChar8 *) FC_ESCAPE_FIXED : 0))
// 		 goto bail0;
// 	 }
// 	 e = FcPatternObjectFindElt (pat, FC_SIZE_OBJECT);
// 	 if (e)
// 	 {
// 	 FcChar8 *p;

// 	 if (!FcNameUnparseString (&buf2, (FcChar8 *) "-", 0))
// 		 goto bail0;
// 	 if (!FcNameUnparseValueList (&buf2, FcPatternEltValues(e), escape ? (FcChar8 *) FC_ESCAPE_FIXED : 0))
// 		 goto bail0;
// 	 p = FcStrBufDoneStatic (&buf2);
// 	 FcStrBufDestroy (&buf2);
// 	 if (strlen ((const char *)p) > 1)
// 		 if (!FcStrBufString (&buf, p))
// 		 goto bail0;
// 	 }
// 	 for (i = 0; i < NUM_OBJECT_TYPES; i++)
// 	 {
// 	 FcObject id = i + 1;
// 	 const FcObjectType	    *o;
// 	 o = &FcObjects[i];
// 	 if (!strcmp (o.object, FC_FAMILY) ||
// 		 !strcmp (o.object, FC_SIZE))
// 		 continue;

// 	 e = FcPatternObjectFindElt (pat, id);
// 	 if (e)
// 	 {
// 		 if (!FcNameUnparseString (&buf, (FcChar8 *) ":", 0))
// 		 goto bail0;
// 		 if (!FcNameUnparseString (&buf, (FcChar8 *) o.object, escape ? (FcChar8 *) FC_ESCAPE_VARIABLE : 0))
// 		 goto bail0;
// 		 if (!FcNameUnparseString (&buf, (FcChar8 *) "=", 0))
// 		 goto bail0;
// 		 if (!FcNameUnparseValueList (&buf, FcPatternEltValues(e), escape ?
// 					  (FcChar8 *) FC_ESCAPE_VARIABLE : 0))
// 		 goto bail0;
// 	 }
// 	 }
// 	 return FcStrBufDone (&buf);
//  bail0:
// 	 FcStrBufDestroy (&buf);
// 	 return 0;
//  }
