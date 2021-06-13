package fontconfig

import (
	"bytes"
	"fmt"
	"sort"
	"strings"

	"github.com/benoitkugler/textlayout/fonts"
	"github.com/benoitkugler/textlayout/fonts/bitmap"
	"github.com/benoitkugler/textlayout/fonts/truetype"
	"github.com/benoitkugler/textlayout/fonts/type1"
	type1c "github.com/benoitkugler/textlayout/fonts/type1C"
	"golang.org/x/image/math/fixed"
	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/encoding/korean"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/encoding/traditionalchinese"
	"golang.org/x/text/encoding/unicode"
)

// ported from fontconfig/src/fcdir.c and fcfreetype.c   2000 Keith Packard

var loaders = [...]struct {
	loader fonts.FontLoader
	format string
}{
	{truetype.Loader, "TrueType"},
	{bitmap.Loader, "PCF"},
	{type1.Loader, "Type 1"},
	{type1c.Loader, "CFF"},
}

type FontFormat string

const (
	TrueType FontFormat = "TrueType"
	PCF      FontFormat = "PCF"
	Type1    FontFormat = "Type 1"
	CFF      FontFormat = "CFF"
)

// Loader returns the loader for the font format.
func (ff FontFormat) Loader() fonts.FontLoader {
	switch ff {
	case "TrueType":
		return truetype.Loader
	case "PCF":
		return bitmap.Loader
	case "Type 1":
		return type1.Loader
	case "CFF":
		return type1c.Loader
	default:
		return nil
	}
}

// constructs patterns found in 'file'. `fileID` is included
// in the returned patterns, as well as the index of each face.
// invalid files are simply ignored, return an empty font set
func scanOneFontFile(file fonts.Resource, fileID string, config *Config) Fontset {
	if debugMode {
		fmt.Printf("Scanning file %s...\n", file)
	}

	faces, ok := readFontFile(file)
	if !ok {
		return nil
	}

	var set Fontset

	for faceNum, face := range faces {
		// basic face
		// TODO: share the sets
		pat, _, _, _ := queryFace(face, fileID, uint32(faceNum))
		if pat != nil {
			set = append(set, pat)
		}

		// optional variable instances
		if variable, ok := face.(truetype.VariableFont); ok {
			vars := variable.Variations()
			for instNum, instance := range vars.Instances {
				// skip named-instance that coincides with base instance.
				if vars.IsDefaultInstance(instance) {
					continue
				}
				// for variable fonts, the id contains both
				// the face number (16 lower bits) and the instance number
				// (16 higher bits, starting at 1)
				id := uint32(faceNum) | uint32(instNum+1)<<19
				// TODO: share the sets
				pat, _, _, _ = queryFace(face, fileID, id)
				if pat != nil {
					set = append(set, pat)
				}
			}
		}
	}

	for _, font := range set {
		/*
		 * Get rid of sysroot here so that targeting scan rule may contains FILE pattern
		 * and they should usually expect without sysroot.
		 */
		// if sysroot != "" {
		// 	f, res := font.GetAtString(FILE, 0)
		// 	if res == ResultMatch && strings.HasPrefix(f, sysroot) {
		// 		font.Del(FILE)
		// 		s := filepath.Clean(strings.TrimPrefix(f, sysroot))
		// 		font.Add(FILE, String(s), true)
		// 	}
		// }

		// Edit pattern with user-defined rules
		config.Substitute(font, nil, MatchScan)

		font.addFullname()

		if debugMode {
			fmt.Println("Final font pattern:", font)
			fmt.Println()
		}
	}

	if debugMode {
		fmt.Printf("Done. (found %d faces)\n\n", len(faces))
	}

	return set
}

func FT_Set_Var_Design_Coordinates(face FT_Face, id int, coors []float32) {}

// readFontFile tries for every possible format, returning true if one match
func readFontFile(file fonts.Resource) (fonts.Faces, bool) {
	for _, loader := range loaders {
		out, err := loader.loader.Load(file)
		if err == nil {
			return out, true
		}
	}
	return nil, false
}

// constructs patterns found in 'file'. `fileID` is included
// in the returned pattern, as well as the index of each face.
// The number of faces in 'file' is also returned, which
// will differ from the size of the returned font set for variable fonts,
// or when some faces are invalid.
func scanFontRessource(file fonts.Resource, fileID string) (nbFaces int, set Fontset) {
	faces, ok := readFontFile(file)
	if !ok {
		return 0, nil
	}

	for faceNum, face := range faces {
		// basic face
		// TODO: share the sets
		pat, _, _, _ := queryFace(face, fileID, uint32(faceNum))
		if pat != nil {
			set = append(set, pat)
		}

		// optional variable instances
		if variable, ok := face.(truetype.VariableFont); ok {
			vars := variable.Variations()
			for instNum, instance := range vars.Instances {
				// skip named-instance that coincides with base instance.
				if vars.IsDefaultInstance(instance) {
					continue
				}
				// for variable fonts, the id contains both
				// the face number (16 lower bits) and the instance number
				// (16 higher bits, starting at 1)
				id := uint32(faceNum) | uint32(instNum+1)<<19
				// TODO: share the sets
				pat, _, _, _ = queryFace(face, fileID, id)
				if pat != nil {
					set = append(set, pat)
				}
			}
		}
	}

	return len(faces), set
}

const lang_DONT_CARE = 0xffff

const (
	macLangID_ENGLISH = iota
	macLangID_FRENCH
	macLangID_GERMAN
	macLangID_ITALIAN
	macLangID_DUTCH
	macLangID_SWEDISH
	macLangID_SPANISH
	macLangID_DANISH
	macLangID_PORTUGUESE
	macLangID_NORWEGIAN
	macLangID_HEBREW
	macLangID_JAPANESE
	macLangID_ARABIC
	macLangID_FINNISH
	macLangID_GREEK
	macLangID_ICELANDIC
	macLangID_MALTESE
	macLangID_TURKISH
	macLangID_CROATIAN
	macLangID_CHINESE_TRADITIONAL
	macLangID_URDU
	macLangID_HINDI
	macLangID_THAI
	macLangID_KOREAN
	macLangID_LITHUANIAN
	macLangID_POLISH
	macLangID_HUNGARIAN
	macLangID_ESTONIAN
	macLangID_LETTISH
	macLangID_SAAMISK
	macLangID_FAEROESE
	macLangID_FARSI
	macLangID_RUSSIAN
	macLangID_CHINESE_SIMPLIFIED
	macLangID_FLEMISH
	macLangID_IRISH
	macLangID_ALBANIAN
	macLangID_ROMANIAN
	macLangID_CZECH
	macLangID_SLOVAK
	macLangID_SLOVENIAN
	macLangID_YIDDISH
	macLangID_SERBIAN
	macLangID_MACEDONIAN
	macLangID_BULGARIAN
	macLangID_UKRAINIAN
	macLangID_BYELORUSSIAN
	macLangID_UZBEK
	macLangID_KAZAKH
	//  macLangID_AZERBAIJANI
	macLangID_AZERBAIJANI_CYRILLIC_SCRIPT
	macLangID_AZERBAIJANI_ARABIC_SCRIPT
	macLangID_ARMENIAN
	macLangID_GEORGIAN
	macLangID_MOLDAVIAN
	macLangID_KIRGHIZ
	macLangID_TAJIKI
	macLangID_TURKMEN
	macLangID_MONGOLIAN
	//  macLangID_MONGOLIAN_MONGOLIAN_SCRIPT
	macLangID_MONGOLIAN_CYRILLIC_SCRIPT
	macLangID_PASHTO
	macLangID_KURDISH
	macLangID_KASHMIRI
	macLangID_SINDHI
	macLangID_TIBETAN
	macLangID_NEPALI
	macLangID_SANSKRIT
	macLangID_MARATHI
	macLangID_BENGALI
	macLangID_ASSAMESE
	macLangID_GUJARATI
	macLangID_PUNJABI
	macLangID_ORIYA
	macLangID_MALAYALAM
	macLangID_KANNADA
	macLangID_TAMIL
	macLangID_TELUGU
	macLangID_SINHALESE
	macLangID_BURMESE
	macLangID_KHMER
	macLangID_LAO
	macLangID_VIETNAMESE
	macLangID_INDONESIAN
	macLangID_TAGALOG
	macLangID_MALAY_ROMAN_SCRIPT
	macLangID_MALAY_ARABIC_SCRIPT
	macLangID_AMHARIC
	macLangID_TIGRINYA
	macLangID_GALLA
	macLangID_SOMALI
	macLangID_SWAHILI
	macLangID_RUANDA
	macLangID_RUNDI
	macLangID_CHEWA
	macLangID_MALAGASY
	macLangID_ESPERANTO
)

const (
	macLangID_WELSH = 128 + iota
	macLangID_BASQUE
	macLangID_CATALAN
	macLangID_LATIN
	macLangID_QUECHUA
	macLangID_GUARANI
	macLangID_AYMARA
	macLangID_TATAR
	macLangID_UIGHUR
	macLangID_DZONGKHA
	macLangID_JAVANESE
	macLangID_SUNDANESE

	/* The following codes are new as of 2000-03-10 */
	macLangID_GALICIAN
	macLangID_AFRIKAANS
	macLangID_BRETON
	macLangID_INUKTITUT
	macLangID_SCOTTISH_GAELIC
	macLangID_MANX_GAELIC
	macLangID_IRISH_GAELIC
	macLangID_TONGAN
	macLangID_GREEK_POLYTONIC
	macLangID_GREELANDIC
	macLangID_AZERBAIJANI_ROMAN_SCRIPT
)

const (
	msLangID_ARABIC_GENERAL                  = 0x0001
	msLangID_CHINESE_GENERAL                 = 0x0004
	msLangID_ENGLISH_GENERAL                 = 0x0009
	msLangID_FRENCH_WEST_INDIES              = 0x1C0C
	msLangID_FRENCH_REUNION                  = 0x200C
	msLangID_FRENCH_CONGO                    = 0x240C
	msLangID_FRENCH_SENEGAL                  = 0x280C
	msLangID_FRENCH_CAMEROON                 = 0x2C0C
	msLangID_FRENCH_COTE_D_IVOIRE            = 0x300C
	msLangID_FRENCH_MALI                     = 0x340C
	msLangID_ARABIC_SAUDI_ARABIA             = 0x0401
	msLangID_ARABIC_IRAQ                     = 0x0801
	msLangID_ARABIC_EGYPT                    = 0x0C01
	msLangID_ARABIC_LIBYA                    = 0x1001
	msLangID_ARABIC_ALGERIA                  = 0x1401
	msLangID_ARABIC_MOROCCO                  = 0x1801
	msLangID_ARABIC_TUNISIA                  = 0x1C01
	msLangID_ARABIC_OMAN                     = 0x2001
	msLangID_ARABIC_YEMEN                    = 0x2401
	msLangID_ARABIC_SYRIA                    = 0x2801
	msLangID_ARABIC_JORDAN                   = 0x2C01
	msLangID_ARABIC_LEBANON                  = 0x3001
	msLangID_ARABIC_KUWAIT                   = 0x3401
	msLangID_ARABIC_UAE                      = 0x3801
	msLangID_ARABIC_BAHRAIN                  = 0x3C01
	msLangID_ARABIC_QATAR                    = 0x4001
	msLangID_BULGARIAN_BULGARIA              = 0x0402
	msLangID_CATALAN_CATALAN                 = 0x0403
	msLangID_CHINESE_TAIWAN                  = 0x0404
	msLangID_CHINESE_PRC                     = 0x0804
	msLangID_CHINESE_HONG_KONG               = 0x0C04
	msLangID_CHINESE_SINGAPORE               = 0x1004
	msLangID_CHINESE_MACAO                   = 0x1404
	msLangID_CZECH_CZECH_REPUBLIC            = 0x0405
	msLangID_DANISH_DENMARK                  = 0x0406
	msLangID_GERMAN_GERMANY                  = 0x0407
	msLangID_GERMAN_SWITZERLAND              = 0x0807
	msLangID_GERMAN_AUSTRIA                  = 0x0C07
	msLangID_GERMAN_LUXEMBOURG               = 0x1007
	msLangID_GERMAN_LIECHTENSTEIN            = 0x1407
	msLangID_GREEK_GREECE                    = 0x0408
	msLangID_ENGLISH_UNITED_STATES           = 0x0409
	msLangID_ENGLISH_UNITED_KINGDOM          = 0x0809
	msLangID_ENGLISH_AUSTRALIA               = 0x0C09
	msLangID_ENGLISH_CANADA                  = 0x1009
	msLangID_ENGLISH_NEW_ZEALAND             = 0x1409
	msLangID_ENGLISH_IRELAND                 = 0x1809
	msLangID_ENGLISH_SOUTH_AFRICA            = 0x1C09
	msLangID_ENGLISH_JAMAICA                 = 0x2009
	msLangID_ENGLISH_CARIBBEAN               = 0x2409
	msLangID_ENGLISH_BELIZE                  = 0x2809
	msLangID_ENGLISH_TRINIDAD                = 0x2C09
	msLangID_ENGLISH_ZIMBABWE                = 0x3009
	msLangID_ENGLISH_PHILIPPINES             = 0x3409
	msLangID_ENGLISH_HONG_KONG               = 0x3C09
	msLangID_ENGLISH_INDIA                   = 0x4009
	msLangID_ENGLISH_MALAYSIA                = 0x4409
	msLangID_ENGLISH_SINGAPORE               = 0x4809
	msLangID_SPANISH_SPAIN_TRADITIONAL_SORT  = 0x040A
	msLangID_SPANISH_MEXICO                  = 0x080A
	msLangID_SPANISH_SPAIN_MODERN_SORT       = 0x0C0A
	msLangID_SPANISH_GUATEMALA               = 0x100A
	msLangID_SPANISH_COSTA_RICA              = 0x140A
	msLangID_SPANISH_PANAMA                  = 0x180A
	msLangID_SPANISH_DOMINICAN_REPUBLIC      = 0x1C0A
	msLangID_SPANISH_VENEZUELA               = 0x200A
	msLangID_SPANISH_COLOMBIA                = 0x240A
	msLangID_SPANISH_PERU                    = 0x280A
	msLangID_SPANISH_ARGENTINA               = 0x2C0A
	msLangID_SPANISH_ECUADOR                 = 0x300A
	msLangID_SPANISH_CHILE                   = 0x340A
	msLangID_SPANISH_URUGUAY                 = 0x380A
	msLangID_SPANISH_PARAGUAY                = 0x3C0A
	msLangID_SPANISH_BOLIVIA                 = 0x400A
	msLangID_SPANISH_EL_SALVADOR             = 0x440A
	msLangID_SPANISH_HONDURAS                = 0x480A
	msLangID_SPANISH_NICARAGUA               = 0x4C0A
	msLangID_SPANISH_PUERTO_RICO             = 0x500A
	msLangID_SPANISH_UNITED_STATES           = 0x540A
	msLangID_SPANISH_LATIN_AMERICA           = 0xE40A
	msLangID_FRENCH_NORTH_AFRICA             = 0xE40C
	msLangID_FRENCH_MOROCCO                  = 0x380C
	msLangID_FRENCH_HAITI                    = 0x3C0C
	msLangID_FINNISH_FINLAND                 = 0x040B
	msLangID_FRENCH_FRANCE                   = 0x040C
	msLangID_FRENCH_BELGIUM                  = 0x080C
	msLangID_FRENCH_CANADA                   = 0x0C0C
	msLangID_FRENCH_SWITZERLAND              = 0x100C
	msLangID_FRENCH_LUXEMBOURG               = 0x140C
	msLangID_FRENCH_MONACO                   = 0x180C
	msLangID_HEBREW_ISRAEL                   = 0x040D
	msLangID_HUNGARIAN_HUNGARY               = 0x040E
	msLangID_ICELANDIC_ICELAND               = 0x040F
	msLangID_ITALIAN_ITALY                   = 0x0410
	msLangID_ITALIAN_SWITZERLAND             = 0x0810
	msLangID_JAPANESE_JAPAN                  = 0x0411
	msLangID_KOREAN_KOREA                    = 0x0412
	msLangID_KOREAN_JOHAB_KOREA              = 0x0812 // legacy
	msLangID_DUTCH_NETHERLANDS               = 0x0413
	msLangID_DUTCH_BELGIUM                   = 0x0813
	msLangID_NORWEGIAN_NORWAY_BOKMAL         = 0x0414
	msLangID_NORWEGIAN_NORWAY_NYNORSK        = 0x0814
	msLangID_POLISH_POLAND                   = 0x0415
	msLangID_PORTUGUESE_BRAZIL               = 0x0416
	msLangID_PORTUGUESE_PORTUGAL             = 0x0816
	msLangID_ROMANSH_SWITZERLAND             = 0x0417
	msLangID_ROMANIAN_ROMANIA                = 0x0418
	msLangID_MOLDAVIAN_MOLDAVIA              = 0x0818 // legacy
	msLangID_RUSSIAN_MOLDAVIA                = 0x0819 // legacy
	msLangID_RUSSIAN_RUSSIA                  = 0x0419
	msLangID_CROATIAN_CROATIA                = 0x041A
	msLangID_SERBIAN_SERBIA_LATIN            = 0x081A
	msLangID_SERBIAN_SERBIA_CYRILLIC         = 0x0C1A
	msLangID_CROATIAN_BOSNIA_HERZEGOVINA     = 0x101A
	msLangID_BOSNIAN_BOSNIA_HERZEGOVINA      = 0x141A
	msLangID_SERBIAN_BOSNIA_HERZ_LATIN       = 0x181A
	msLangID_SERBIAN_BOSNIA_HERZ_CYRILLIC    = 0x1C1A
	msLangID_BOSNIAN_BOSNIA_HERZ_CYRILLIC    = 0x201A
	msLangID_URDU_INDIA                      = 0x0820
	msLangID_SLOVAK_SLOVAKIA                 = 0x041B
	msLangID_ALBANIAN_ALBANIA                = 0x041C
	msLangID_SWEDISH_SWEDEN                  = 0x041D
	msLangID_SWEDISH_FINLAND                 = 0x081D
	msLangID_THAI_THAILAND                   = 0x041E
	msLangID_TURKISH_TURKEY                  = 0x041F
	msLangID_URDU_PAKISTAN                   = 0x0420
	msLangID_INDONESIAN_INDONESIA            = 0x0421
	msLangID_UKRAINIAN_UKRAINE               = 0x0422
	msLangID_BELARUSIAN_BELARUS              = 0x0423
	msLangID_SLOVENIAN_SLOVENIA              = 0x0424
	msLangID_ESTONIAN_ESTONIA                = 0x0425
	msLangID_LATVIAN_LATVIA                  = 0x0426
	msLangID_LITHUANIAN_LITHUANIA            = 0x0427
	msLangID_CLASSIC_LITHUANIAN_LITHUANIA    = 0x0827 // legacy
	msLangID_TAJIK_TAJIKISTAN                = 0x0428
	msLangID_YIDDISH_GERMANY                 = 0x043D
	msLangID_VIETNAMESE_VIET_NAM             = 0x042A
	msLangID_ARMENIAN_ARMENIA                = 0x042B
	msLangID_AZERI_AZERBAIJAN_LATIN          = 0x042C
	msLangID_AZERI_AZERBAIJAN_CYRILLIC       = 0x082C
	msLangID_BASQUE_BASQUE                   = 0x042D
	msLangID_UPPER_SORBIAN_GERMANY           = 0x042E
	msLangID_LOWER_SORBIAN_GERMANY           = 0x082E
	msLangID_MACEDONIAN_MACEDONIA            = 0x042F
	msLangID_SUTU_SOUTH_AFRICA               = 0x0430
	msLangID_TSONGA_SOUTH_AFRICA             = 0x0431
	msLangID_SETSWANA_SOUTH_AFRICA           = 0x0432
	msLangID_VENDA_SOUTH_AFRICA              = 0x0433
	msLangID_ISIXHOSA_SOUTH_AFRICA           = 0x0434
	msLangID_ISIZULU_SOUTH_AFRICA            = 0x0435
	msLangID_AFRIKAANS_SOUTH_AFRICA          = 0x0436
	msLangID_GEORGIAN_GEORGIA                = 0x0437
	msLangID_FAEROESE_FAEROE_ISLANDS         = 0x0438
	msLangID_HINDI_INDIA                     = 0x0439
	msLangID_MALTESE_MALTA                   = 0x043A
	msLangID_SAAMI_LAPONIA                   = 0x043B
	msLangID_SAMI_NORTHERN_NORWAY            = 0x043B
	msLangID_SAMI_NORTHERN_SWEDEN            = 0x083B
	msLangID_SAMI_NORTHERN_FINLAND           = 0x0C3B
	msLangID_SAMI_LULE_NORWAY                = 0x103B
	msLangID_SAMI_LULE_SWEDEN                = 0x143B
	msLangID_SAMI_SOUTHERN_NORWAY            = 0x183B
	msLangID_SAMI_SOUTHERN_SWEDEN            = 0x1C3B
	msLangID_SAMI_SKOLT_FINLAND              = 0x203B
	msLangID_SAMI_INARI_FINLAND              = 0x243B
	msLangID_IRISH_GAELIC_IRELAND            = 0x043C // legacy
	msLangID_SCOTTISH_GAELIC_UNITED_KINGDOM  = 0x083C // legacy
	msLangID_IRISH_IRELAND                   = 0x083C
	msLangID_MALAY_MALAYSIA                  = 0x043E
	msLangID_MALAY_BRUNEI_DARUSSALAM         = 0x083E
	msLangID_KAZAKH_KAZAKHSTAN               = 0x043F
	msLangID_KYRGYZ_KYRGYZSTAN               = /* Cyrillic*/ 0x0440
	msLangID_KISWAHILI_KENYA                 = 0x0441
	msLangID_TURKMEN_TURKMENISTAN            = 0x0442
	msLangID_UZBEK_UZBEKISTAN_LATIN          = 0x0443
	msLangID_UZBEK_UZBEKISTAN_CYRILLIC       = 0x0843
	msLangID_TATAR_RUSSIA                    = 0x0444
	msLangID_BENGALI_INDIA                   = 0x0445
	msLangID_BENGALI_BANGLADESH              = 0x0845
	msLangID_PUNJABI_INDIA                   = 0x0446
	msLangID_PUNJABI_ARABIC_PAKISTAN         = 0x0846
	msLangID_GUJARATI_INDIA                  = 0x0447
	msLangID_ODIA_INDIA                      = 0x0448
	msLangID_TAMIL_INDIA                     = 0x0449
	msLangID_TELUGU_INDIA                    = 0x044A
	msLangID_KANNADA_INDIA                   = 0x044B
	msLangID_MALAYALAM_INDIA                 = 0x044C
	msLangID_ASSAMESE_INDIA                  = 0x044D
	msLangID_MARATHI_INDIA                   = 0x044E
	msLangID_SANSKRIT_INDIA                  = 0x044F
	msLangID_MONGOLIAN_MONGOLIA              = /* Cyrillic */ 0x0450
	msLangID_MONGOLIAN_PRC                   = 0x0850
	msLangID_TIBETAN_PRC                     = 0x0451
	msLangID_DZONGHKA_BHUTAN                 = 0x0851
	msLangID_WELSH_UNITED_KINGDOM            = 0x0452
	msLangID_KHMER_CAMBODIA                  = 0x0453
	msLangID_LAO_LAOS                        = 0x0454
	msLangID_BURMESE_MYANMAR                 = 0x0455
	msLangID_GALICIAN_GALICIAN               = 0x0456
	msLangID_MANIPURI_INDIA                  = /* Bengali */ 0x0458
	msLangID_SINDHI_INDIA                    = /* Arabic */ 0x0459
	msLangID_KONKANI_INDIA                   = 0x0457
	msLangID_KASHMIRI_PAKISTAN               = /* Arabic */ 0x0460
	msLangID_KASHMIRI_SASIA                  = 0x0860
	msLangID_SYRIAC_SYRIA                    = 0x045A
	msLangID_SINHALA_SRI_LANKA               = 0x045B
	msLangID_CHEROKEE_UNITED_STATES          = 0x045C
	msLangID_INUKTITUT_CANADA                = 0x045D
	msLangID_INUKTITUT_CANADA_LATIN          = 0x085D
	msLangID_AMHARIC_ETHIOPIA                = 0x045E
	msLangID_TAMAZIGHT_ALGERIA               = 0x085F
	msLangID_NEPALI_NEPAL                    = 0x0461
	msLangID_FRISIAN_NETHERLANDS             = 0x0462
	msLangID_PASHTO_AFGHANISTAN              = 0x0463
	msLangID_FILIPINO_PHILIPPINES            = 0x0464
	msLangID_DHIVEHI_MALDIVES                = 0x0465
	msLangID_OROMO_ETHIOPIA                  = 0x0472
	msLangID_TIGRIGNA_ETHIOPIA               = 0x0473
	msLangID_TIGRIGNA_ERYTHREA               = 0x0873
	msLangID_HAUSA_NIGERIA                   = 0x0468
	msLangID_YORUBA_NIGERIA                  = 0x046A
	msLangID_QUECHUA_BOLIVIA                 = 0x046B
	msLangID_QUECHUA_ECUADOR                 = 0x086B
	msLangID_QUECHUA_PERU                    = 0x0C6B
	msLangID_SESOTHO_SA_LEBOA_SOUTH_AFRICA   = 0x046C
	msLangID_BASHKIR_RUSSIA                  = 0x046D
	msLangID_LUXEMBOURGISH_LUXEMBOURG        = 0x046E
	msLangID_GREENLANDIC_GREENLAND           = 0x046F
	msLangID_IGBO_NIGERIA                    = 0x0470
	msLangID_KANURI_NIGERIA                  = 0x0471
	msLangID_GUARANI_PARAGUAY                = 0x0474
	msLangID_HAWAIIAN_UNITED_STATES          = 0x0475
	msLangID_LATIN                           = 0x0476
	msLangID_SOMALI_SOMALIA                  = 0x0477
	msLangID_YI_PRC                          = 0x0478
	msLangID_MAPUDUNGUN_CHILE                = 0x047A
	msLangID_MOHAWK_MOHAWK                   = 0x047C
	msLangID_BRETON_FRANCE                   = 0x047E
	msLangID_UIGHUR_PRC                      = 0x0480
	msLangID_MAORI_NEW_ZEALAND               = 0x0481
	msLangID_FARSI_IRAN                      = 0x0429
	msLangID_OCCITAN_FRANCE                  = 0x0482
	msLangID_CORSICAN_FRANCE                 = 0x0483
	msLangID_ALSATIAN_FRANCE                 = 0x0484
	msLangID_YAKUT_RUSSIA                    = 0x0485
	msLangID_KICHE_GUATEMALA                 = 0x0486
	msLangID_KINYARWANDA_RWANDA              = 0x0487
	msLangID_WOLOF_SENEGAL                   = 0x0488
	msLangID_DARI_AFGHANISTAN                = 0x048C
	msLangID_PAPIAMENTU_NETHERLANDS_ANTILLES = 0x0479
)

var fcFtLanguage = [...]struct {
	PlatformID truetype.PlatformID
	LanguageID truetype.PlatformLanguageID
	lang       string
}{
	{truetype.PlatformUnicode, lang_DONT_CARE, ""},
	{truetype.PlatformMac, macLangID_ENGLISH, "en"},
	{truetype.PlatformMac, macLangID_FRENCH, "fr"},
	{truetype.PlatformMac, macLangID_GERMAN, "de"},
	{truetype.PlatformMac, macLangID_ITALIAN, "it"},
	{truetype.PlatformMac, macLangID_DUTCH, "nl"},
	{truetype.PlatformMac, macLangID_SWEDISH, "sv"},
	{truetype.PlatformMac, macLangID_SPANISH, "es"},
	{truetype.PlatformMac, macLangID_DANISH, "da"},
	{truetype.PlatformMac, macLangID_PORTUGUESE, "pt"},
	{truetype.PlatformMac, macLangID_NORWEGIAN, "no"},
	{truetype.PlatformMac, macLangID_HEBREW, "he"},
	{truetype.PlatformMac, macLangID_JAPANESE, "ja"},
	{truetype.PlatformMac, macLangID_ARABIC, "ar"},
	{truetype.PlatformMac, macLangID_FINNISH, "fi"},
	{truetype.PlatformMac, macLangID_GREEK, "el"},
	{truetype.PlatformMac, macLangID_ICELANDIC, "is"},
	{truetype.PlatformMac, macLangID_MALTESE, "mt"},
	{truetype.PlatformMac, macLangID_TURKISH, "tr"},
	{truetype.PlatformMac, macLangID_CROATIAN, "hr"},
	{truetype.PlatformMac, macLangID_CHINESE_TRADITIONAL, "zh-tw"},
	{truetype.PlatformMac, macLangID_URDU, "ur"},
	{truetype.PlatformMac, macLangID_HINDI, "hi"},
	{truetype.PlatformMac, macLangID_THAI, "th"},
	{truetype.PlatformMac, macLangID_KOREAN, "ko"},
	{truetype.PlatformMac, macLangID_LITHUANIAN, "lt"},
	{truetype.PlatformMac, macLangID_POLISH, "pl"},
	{truetype.PlatformMac, macLangID_HUNGARIAN, "hu"},
	{truetype.PlatformMac, macLangID_ESTONIAN, "et"},
	{truetype.PlatformMac, macLangID_LETTISH, "lv"},

	{truetype.PlatformMac, macLangID_FAEROESE, "fo"},
	{truetype.PlatformMac, macLangID_FARSI, "fa"},
	{truetype.PlatformMac, macLangID_RUSSIAN, "ru"},
	{truetype.PlatformMac, macLangID_CHINESE_SIMPLIFIED, "zh-cn"},
	{truetype.PlatformMac, macLangID_FLEMISH, "nl"},
	{truetype.PlatformMac, macLangID_IRISH, "ga"},
	{truetype.PlatformMac, macLangID_ALBANIAN, "sq"},
	{truetype.PlatformMac, macLangID_ROMANIAN, "ro"},
	{truetype.PlatformMac, macLangID_CZECH, "cs"},
	{truetype.PlatformMac, macLangID_SLOVAK, "sk"},
	{truetype.PlatformMac, macLangID_SLOVENIAN, "sl"},
	{truetype.PlatformMac, macLangID_YIDDISH, "yi"},
	{truetype.PlatformMac, macLangID_SERBIAN, "sr"},
	{truetype.PlatformMac, macLangID_MACEDONIAN, "mk"},
	{truetype.PlatformMac, macLangID_BULGARIAN, "bg"},
	{truetype.PlatformMac, macLangID_UKRAINIAN, "uk"},
	{truetype.PlatformMac, macLangID_BYELORUSSIAN, "be"},
	{truetype.PlatformMac, macLangID_UZBEK, "uz"},
	{truetype.PlatformMac, macLangID_KAZAKH, "kk"},
	{truetype.PlatformMac, macLangID_AZERBAIJANI_CYRILLIC_SCRIPT, "az"},
	{truetype.PlatformMac, macLangID_AZERBAIJANI_ARABIC_SCRIPT, "ar"},
	{truetype.PlatformMac, macLangID_ARMENIAN, "hy"},
	{truetype.PlatformMac, macLangID_GEORGIAN, "ka"},
	{truetype.PlatformMac, macLangID_MOLDAVIAN, "mo"},
	{truetype.PlatformMac, macLangID_KIRGHIZ, "ky"},
	{truetype.PlatformMac, macLangID_TAJIKI, "tg"},
	{truetype.PlatformMac, macLangID_TURKMEN, "tk"},
	{truetype.PlatformMac, macLangID_MONGOLIAN, "mn"},
	{truetype.PlatformMac, macLangID_MONGOLIAN_CYRILLIC_SCRIPT, "mn"},
	{truetype.PlatformMac, macLangID_PASHTO, "ps"},
	{truetype.PlatformMac, macLangID_KURDISH, "ku"},
	{truetype.PlatformMac, macLangID_KASHMIRI, "ks"},
	{truetype.PlatformMac, macLangID_SINDHI, "sd"},
	{truetype.PlatformMac, macLangID_TIBETAN, "bo"},
	{truetype.PlatformMac, macLangID_NEPALI, "ne"},
	{truetype.PlatformMac, macLangID_SANSKRIT, "sa"},
	{truetype.PlatformMac, macLangID_MARATHI, "mr"},
	{truetype.PlatformMac, macLangID_BENGALI, "bn"},
	{truetype.PlatformMac, macLangID_ASSAMESE, "as"},
	{truetype.PlatformMac, macLangID_GUJARATI, "gu"},
	{truetype.PlatformMac, macLangID_PUNJABI, "pa"},
	{truetype.PlatformMac, macLangID_ORIYA, "or"},
	{truetype.PlatformMac, macLangID_MALAYALAM, "ml"},
	{truetype.PlatformMac, macLangID_KANNADA, "kn"},
	{truetype.PlatformMac, macLangID_TAMIL, "ta"},
	{truetype.PlatformMac, macLangID_TELUGU, "te"},
	{truetype.PlatformMac, macLangID_SINHALESE, "si"},
	{truetype.PlatformMac, macLangID_BURMESE, "my"},
	{truetype.PlatformMac, macLangID_KHMER, "km"},
	{truetype.PlatformMac, macLangID_LAO, "lo"},
	{truetype.PlatformMac, macLangID_VIETNAMESE, "vi"},
	{truetype.PlatformMac, macLangID_INDONESIAN, "id"},
	{truetype.PlatformMac, macLangID_TAGALOG, "tl"},
	{truetype.PlatformMac, macLangID_MALAY_ROMAN_SCRIPT, "ms"},
	{truetype.PlatformMac, macLangID_MALAY_ARABIC_SCRIPT, "ms"},
	{truetype.PlatformMac, macLangID_AMHARIC, "am"},
	{truetype.PlatformMac, macLangID_TIGRINYA, "ti"},
	{truetype.PlatformMac, macLangID_GALLA, "om"},
	{truetype.PlatformMac, macLangID_SOMALI, "so"},
	{truetype.PlatformMac, macLangID_SWAHILI, "sw"},
	{truetype.PlatformMac, macLangID_RUANDA, "rw"},
	{truetype.PlatformMac, macLangID_RUNDI, "rn"},
	{truetype.PlatformMac, macLangID_CHEWA, "ny"},
	{truetype.PlatformMac, macLangID_MALAGASY, "mg"},
	{truetype.PlatformMac, macLangID_ESPERANTO, "eo"},
	{truetype.PlatformMac, macLangID_WELSH, "cy"},
	{truetype.PlatformMac, macLangID_BASQUE, "eu"},
	{truetype.PlatformMac, macLangID_CATALAN, "ca"},
	{truetype.PlatformMac, macLangID_LATIN, "la"},
	{truetype.PlatformMac, macLangID_QUECHUA, "qu"},
	{truetype.PlatformMac, macLangID_GUARANI, "gn"},
	{truetype.PlatformMac, macLangID_AYMARA, "ay"},
	{truetype.PlatformMac, macLangID_TATAR, "tt"},
	{truetype.PlatformMac, macLangID_UIGHUR, "ug"},
	{truetype.PlatformMac, macLangID_DZONGKHA, "dz"},
	{truetype.PlatformMac, macLangID_JAVANESE, "jw"},
	{truetype.PlatformMac, macLangID_SUNDANESE, "su"},

	/* The following codes are new as of 2000-03-10 */
	{truetype.PlatformMac, macLangID_GALICIAN, "gl"},
	{truetype.PlatformMac, macLangID_AFRIKAANS, "af"},
	{truetype.PlatformMac, macLangID_BRETON, "br"},
	{truetype.PlatformMac, macLangID_INUKTITUT, "iu"},
	{truetype.PlatformMac, macLangID_SCOTTISH_GAELIC, "gd"},
	{truetype.PlatformMac, macLangID_MANX_GAELIC, "gv"},
	{truetype.PlatformMac, macLangID_IRISH_GAELIC, "ga"},
	{truetype.PlatformMac, macLangID_TONGAN, "to"},
	{truetype.PlatformMac, macLangID_GREEK_POLYTONIC, "el"},
	{truetype.PlatformMac, macLangID_GREELANDIC, "ik"},
	{truetype.PlatformMac, macLangID_AZERBAIJANI_ROMAN_SCRIPT, "az"},

	{truetype.PlatformMicrosoft, msLangID_ARABIC_SAUDI_ARABIA, "ar"},
	{truetype.PlatformMicrosoft, msLangID_ARABIC_IRAQ, "ar"},
	{truetype.PlatformMicrosoft, msLangID_ARABIC_EGYPT, "ar"},
	{truetype.PlatformMicrosoft, msLangID_ARABIC_LIBYA, "ar"},
	{truetype.PlatformMicrosoft, msLangID_ARABIC_ALGERIA, "ar"},
	{truetype.PlatformMicrosoft, msLangID_ARABIC_MOROCCO, "ar"},
	{truetype.PlatformMicrosoft, msLangID_ARABIC_TUNISIA, "ar"},
	{truetype.PlatformMicrosoft, msLangID_ARABIC_OMAN, "ar"},
	{truetype.PlatformMicrosoft, msLangID_ARABIC_YEMEN, "ar"},
	{truetype.PlatformMicrosoft, msLangID_ARABIC_SYRIA, "ar"},
	{truetype.PlatformMicrosoft, msLangID_ARABIC_JORDAN, "ar"},
	{truetype.PlatformMicrosoft, msLangID_ARABIC_LEBANON, "ar"},
	{truetype.PlatformMicrosoft, msLangID_ARABIC_KUWAIT, "ar"},
	{truetype.PlatformMicrosoft, msLangID_ARABIC_UAE, "ar"},
	{truetype.PlatformMicrosoft, msLangID_ARABIC_BAHRAIN, "ar"},
	{truetype.PlatformMicrosoft, msLangID_ARABIC_QATAR, "ar"},
	{truetype.PlatformMicrosoft, msLangID_BULGARIAN_BULGARIA, "bg"},
	{truetype.PlatformMicrosoft, msLangID_CATALAN_CATALAN, "ca"},
	{truetype.PlatformMicrosoft, msLangID_CHINESE_TAIWAN, "zh-tw"},
	{truetype.PlatformMicrosoft, msLangID_CHINESE_PRC, "zh-cn"},
	{truetype.PlatformMicrosoft, msLangID_CHINESE_HONG_KONG, "zh-hk"},
	{truetype.PlatformMicrosoft, msLangID_CHINESE_SINGAPORE, "zh-sg"},

	{truetype.PlatformMicrosoft, msLangID_CHINESE_MACAO, "zh-mo"},

	{truetype.PlatformMicrosoft, msLangID_CZECH_CZECH_REPUBLIC, "cs"},
	{truetype.PlatformMicrosoft, msLangID_DANISH_DENMARK, "da"},
	{truetype.PlatformMicrosoft, msLangID_GERMAN_GERMANY, "de"},
	{truetype.PlatformMicrosoft, msLangID_GERMAN_SWITZERLAND, "de"},
	{truetype.PlatformMicrosoft, msLangID_GERMAN_AUSTRIA, "de"},
	{truetype.PlatformMicrosoft, msLangID_GERMAN_LUXEMBOURG, "de"},
	{truetype.PlatformMicrosoft, msLangID_GERMAN_LIECHTENSTEIN, "de"},
	{truetype.PlatformMicrosoft, msLangID_GREEK_GREECE, "el"},
	{truetype.PlatformMicrosoft, msLangID_ENGLISH_UNITED_STATES, "en"},
	{truetype.PlatformMicrosoft, msLangID_ENGLISH_UNITED_KINGDOM, "en"},
	{truetype.PlatformMicrosoft, msLangID_ENGLISH_AUSTRALIA, "en"},
	{truetype.PlatformMicrosoft, msLangID_ENGLISH_CANADA, "en"},
	{truetype.PlatformMicrosoft, msLangID_ENGLISH_NEW_ZEALAND, "en"},
	{truetype.PlatformMicrosoft, msLangID_ENGLISH_IRELAND, "en"},
	{truetype.PlatformMicrosoft, msLangID_ENGLISH_SOUTH_AFRICA, "en"},
	{truetype.PlatformMicrosoft, msLangID_ENGLISH_JAMAICA, "en"},
	{truetype.PlatformMicrosoft, msLangID_ENGLISH_CARIBBEAN, "en"},
	{truetype.PlatformMicrosoft, msLangID_ENGLISH_BELIZE, "en"},
	{truetype.PlatformMicrosoft, msLangID_ENGLISH_TRINIDAD, "en"},
	{truetype.PlatformMicrosoft, msLangID_ENGLISH_ZIMBABWE, "en"},
	{truetype.PlatformMicrosoft, msLangID_ENGLISH_PHILIPPINES, "en"},
	{truetype.PlatformMicrosoft, msLangID_SPANISH_SPAIN_TRADITIONAL_SORT, "es"},
	{truetype.PlatformMicrosoft, msLangID_SPANISH_MEXICO, "es"},
	{truetype.PlatformMicrosoft, msLangID_SPANISH_SPAIN_MODERN_SORT, "es"},
	{truetype.PlatformMicrosoft, msLangID_SPANISH_GUATEMALA, "es"},
	{truetype.PlatformMicrosoft, msLangID_SPANISH_COSTA_RICA, "es"},
	{truetype.PlatformMicrosoft, msLangID_SPANISH_PANAMA, "es"},
	{truetype.PlatformMicrosoft, msLangID_SPANISH_DOMINICAN_REPUBLIC, "es"},
	{truetype.PlatformMicrosoft, msLangID_SPANISH_VENEZUELA, "es"},
	{truetype.PlatformMicrosoft, msLangID_SPANISH_COLOMBIA, "es"},
	{truetype.PlatformMicrosoft, msLangID_SPANISH_PERU, "es"},
	{truetype.PlatformMicrosoft, msLangID_SPANISH_ARGENTINA, "es"},
	{truetype.PlatformMicrosoft, msLangID_SPANISH_ECUADOR, "es"},
	{truetype.PlatformMicrosoft, msLangID_SPANISH_CHILE, "es"},
	{truetype.PlatformMicrosoft, msLangID_SPANISH_URUGUAY, "es"},
	{truetype.PlatformMicrosoft, msLangID_SPANISH_PARAGUAY, "es"},
	{truetype.PlatformMicrosoft, msLangID_SPANISH_BOLIVIA, "es"},
	{truetype.PlatformMicrosoft, msLangID_SPANISH_EL_SALVADOR, "es"},
	{truetype.PlatformMicrosoft, msLangID_SPANISH_HONDURAS, "es"},
	{truetype.PlatformMicrosoft, msLangID_SPANISH_NICARAGUA, "es"},
	{truetype.PlatformMicrosoft, msLangID_SPANISH_PUERTO_RICO, "es"},
	{truetype.PlatformMicrosoft, msLangID_FINNISH_FINLAND, "fi"},
	{truetype.PlatformMicrosoft, msLangID_FRENCH_FRANCE, "fr"},
	{truetype.PlatformMicrosoft, msLangID_FRENCH_BELGIUM, "fr"},
	{truetype.PlatformMicrosoft, msLangID_FRENCH_CANADA, "fr"},
	{truetype.PlatformMicrosoft, msLangID_FRENCH_SWITZERLAND, "fr"},
	{truetype.PlatformMicrosoft, msLangID_FRENCH_LUXEMBOURG, "fr"},
	{truetype.PlatformMicrosoft, msLangID_FRENCH_MONACO, "fr"},
	{truetype.PlatformMicrosoft, msLangID_HEBREW_ISRAEL, "he"},
	{truetype.PlatformMicrosoft, msLangID_HUNGARIAN_HUNGARY, "hu"},
	{truetype.PlatformMicrosoft, msLangID_ICELANDIC_ICELAND, "is"},
	{truetype.PlatformMicrosoft, msLangID_ITALIAN_ITALY, "it"},
	{truetype.PlatformMicrosoft, msLangID_ITALIAN_SWITZERLAND, "it"},
	{truetype.PlatformMicrosoft, msLangID_JAPANESE_JAPAN, "ja"},
	{truetype.PlatformMicrosoft, msLangID_KOREAN_KOREA, "ko"},
	{truetype.PlatformMicrosoft, msLangID_KOREAN_JOHAB_KOREA, "ko"},
	{truetype.PlatformMicrosoft, msLangID_DUTCH_NETHERLANDS, "nl"},
	{truetype.PlatformMicrosoft, msLangID_DUTCH_BELGIUM, "nl"},
	{truetype.PlatformMicrosoft, msLangID_NORWEGIAN_NORWAY_BOKMAL, "no"},
	{truetype.PlatformMicrosoft, msLangID_NORWEGIAN_NORWAY_NYNORSK, "nn"},
	{truetype.PlatformMicrosoft, msLangID_POLISH_POLAND, "pl"},
	{truetype.PlatformMicrosoft, msLangID_PORTUGUESE_BRAZIL, "pt"},
	{truetype.PlatformMicrosoft, msLangID_PORTUGUESE_PORTUGAL, "pt"},
	{truetype.PlatformMicrosoft, msLangID_ROMANSH_SWITZERLAND, "rm"},
	{truetype.PlatformMicrosoft, msLangID_ROMANIAN_ROMANIA, "ro"},
	{truetype.PlatformMicrosoft, msLangID_MOLDAVIAN_MOLDAVIA, "mo"},
	{truetype.PlatformMicrosoft, msLangID_RUSSIAN_RUSSIA, "ru"},
	{truetype.PlatformMicrosoft, msLangID_RUSSIAN_MOLDAVIA, "ru"},
	{truetype.PlatformMicrosoft, msLangID_CROATIAN_CROATIA, "hr"},
	{truetype.PlatformMicrosoft, msLangID_SERBIAN_SERBIA_LATIN, "sr"},
	{truetype.PlatformMicrosoft, msLangID_SERBIAN_SERBIA_CYRILLIC, "sr"},
	{truetype.PlatformMicrosoft, msLangID_SLOVAK_SLOVAKIA, "sk"},
	{truetype.PlatformMicrosoft, msLangID_ALBANIAN_ALBANIA, "sq"},
	{truetype.PlatformMicrosoft, msLangID_SWEDISH_SWEDEN, "sv"},
	{truetype.PlatformMicrosoft, msLangID_SWEDISH_FINLAND, "sv"},
	{truetype.PlatformMicrosoft, msLangID_THAI_THAILAND, "th"},
	{truetype.PlatformMicrosoft, msLangID_TURKISH_TURKEY, "tr"},
	{truetype.PlatformMicrosoft, msLangID_URDU_PAKISTAN, "ur"},
	{truetype.PlatformMicrosoft, msLangID_INDONESIAN_INDONESIA, "id"},
	{truetype.PlatformMicrosoft, msLangID_UKRAINIAN_UKRAINE, "uk"},
	{truetype.PlatformMicrosoft, msLangID_BELARUSIAN_BELARUS, "be"},
	{truetype.PlatformMicrosoft, msLangID_SLOVENIAN_SLOVENIA, "sl"},
	{truetype.PlatformMicrosoft, msLangID_ESTONIAN_ESTONIA, "et"},
	{truetype.PlatformMicrosoft, msLangID_LATVIAN_LATVIA, "lv"},
	{truetype.PlatformMicrosoft, msLangID_LITHUANIAN_LITHUANIA, "lt"},
	{truetype.PlatformMicrosoft, msLangID_CLASSIC_LITHUANIAN_LITHUANIA, "lt"},
	{truetype.PlatformMicrosoft, msLangID_MAORI_NEW_ZEALAND, "mi"},
	{truetype.PlatformMicrosoft, msLangID_FARSI_IRAN, "fa"},
	{truetype.PlatformMicrosoft, msLangID_VIETNAMESE_VIET_NAM, "vi"},
	{truetype.PlatformMicrosoft, msLangID_ARMENIAN_ARMENIA, "hy"},
	{truetype.PlatformMicrosoft, msLangID_AZERI_AZERBAIJAN_LATIN, "az"},
	{truetype.PlatformMicrosoft, msLangID_AZERI_AZERBAIJAN_CYRILLIC, "az"},
	{truetype.PlatformMicrosoft, msLangID_BASQUE_BASQUE, "eu"},
	{truetype.PlatformMicrosoft, msLangID_UPPER_SORBIAN_GERMANY, "wen"},
	{truetype.PlatformMicrosoft, msLangID_MACEDONIAN_MACEDONIA, "mk"},
	{truetype.PlatformMicrosoft, msLangID_SUTU_SOUTH_AFRICA, "st"},
	{truetype.PlatformMicrosoft, msLangID_TSONGA_SOUTH_AFRICA, "ts"},
	{truetype.PlatformMicrosoft, msLangID_SETSWANA_SOUTH_AFRICA, "tn"},
	{truetype.PlatformMicrosoft, msLangID_VENDA_SOUTH_AFRICA, "ven"},
	{truetype.PlatformMicrosoft, msLangID_ISIXHOSA_SOUTH_AFRICA, "xh"},
	{truetype.PlatformMicrosoft, msLangID_ISIZULU_SOUTH_AFRICA, "zu"},
	{truetype.PlatformMicrosoft, msLangID_AFRIKAANS_SOUTH_AFRICA, "af"},
	{truetype.PlatformMicrosoft, msLangID_GEORGIAN_GEORGIA, "ka"},
	{truetype.PlatformMicrosoft, msLangID_FAEROESE_FAEROE_ISLANDS, "fo"},
	{truetype.PlatformMicrosoft, msLangID_HINDI_INDIA, "hi"},
	{truetype.PlatformMicrosoft, msLangID_MALTESE_MALTA, "mt"},
	{truetype.PlatformMicrosoft, msLangID_SAAMI_LAPONIA, "se"},

	{truetype.PlatformMicrosoft, msLangID_SCOTTISH_GAELIC_UNITED_KINGDOM, "gd"},
	{truetype.PlatformMicrosoft, msLangID_IRISH_GAELIC_IRELAND, "ga"},

	{truetype.PlatformMicrosoft, msLangID_MALAY_MALAYSIA, "ms"},
	{truetype.PlatformMicrosoft, msLangID_MALAY_BRUNEI_DARUSSALAM, "ms"},
	{truetype.PlatformMicrosoft, msLangID_KAZAKH_KAZAKHSTAN, "kk"},
	{truetype.PlatformMicrosoft, msLangID_KISWAHILI_KENYA, "sw"},
	{truetype.PlatformMicrosoft, msLangID_UZBEK_UZBEKISTAN_LATIN, "uz"},
	{truetype.PlatformMicrosoft, msLangID_UZBEK_UZBEKISTAN_CYRILLIC, "uz"},
	{truetype.PlatformMicrosoft, msLangID_TATAR_RUSSIA, "tt"},
	{truetype.PlatformMicrosoft, msLangID_BENGALI_INDIA, "bn"},
	{truetype.PlatformMicrosoft, msLangID_PUNJABI_INDIA, "pa"},
	{truetype.PlatformMicrosoft, msLangID_GUJARATI_INDIA, "gu"},
	{truetype.PlatformMicrosoft, msLangID_ODIA_INDIA, "or"},
	{truetype.PlatformMicrosoft, msLangID_TAMIL_INDIA, "ta"},
	{truetype.PlatformMicrosoft, msLangID_TELUGU_INDIA, "te"},
	{truetype.PlatformMicrosoft, msLangID_KANNADA_INDIA, "kn"},
	{truetype.PlatformMicrosoft, msLangID_MALAYALAM_INDIA, "ml"},
	{truetype.PlatformMicrosoft, msLangID_ASSAMESE_INDIA, "as"},
	{truetype.PlatformMicrosoft, msLangID_MARATHI_INDIA, "mr"},
	{truetype.PlatformMicrosoft, msLangID_SANSKRIT_INDIA, "sa"},
	{truetype.PlatformMicrosoft, msLangID_KONKANI_INDIA, "kok"},

	/* new as of 2001-01-01 */
	{truetype.PlatformMicrosoft, msLangID_ARABIC_GENERAL, "ar"},
	{truetype.PlatformMicrosoft, msLangID_CHINESE_GENERAL, "zh"},
	{truetype.PlatformMicrosoft, msLangID_ENGLISH_GENERAL, "en"},
	{truetype.PlatformMicrosoft, msLangID_FRENCH_WEST_INDIES, "fr"},
	{truetype.PlatformMicrosoft, msLangID_FRENCH_REUNION, "fr"},
	{truetype.PlatformMicrosoft, msLangID_FRENCH_CONGO, "fr"},

	{truetype.PlatformMicrosoft, msLangID_FRENCH_SENEGAL, "fr"},
	{truetype.PlatformMicrosoft, msLangID_FRENCH_CAMEROON, "fr"},
	{truetype.PlatformMicrosoft, msLangID_FRENCH_COTE_D_IVOIRE, "fr"},
	{truetype.PlatformMicrosoft, msLangID_FRENCH_MALI, "fr"},
	{truetype.PlatformMicrosoft, msLangID_BOSNIAN_BOSNIA_HERZEGOVINA, "bs"},
	{truetype.PlatformMicrosoft, msLangID_URDU_INDIA, "ur"},
	{truetype.PlatformMicrosoft, msLangID_TAJIK_TAJIKISTAN, "tg"},
	{truetype.PlatformMicrosoft, msLangID_YIDDISH_GERMANY, "yi"},
	{truetype.PlatformMicrosoft, msLangID_KYRGYZ_KYRGYZSTAN, "ky"},

	{truetype.PlatformMicrosoft, msLangID_TURKMEN_TURKMENISTAN, "tk"},
	{truetype.PlatformMicrosoft, msLangID_MONGOLIAN_MONGOLIA, "mn"},

	// the following seems to be inconsistent;   here is the current "official" way:
	{truetype.PlatformMicrosoft, msLangID_DZONGHKA_BHUTAN, "bo"},
	/* and here is what is used by Passport SDK */
	{truetype.PlatformMicrosoft, msLangID_TIBETAN_PRC, "bo"},
	{truetype.PlatformMicrosoft, msLangID_DZONGHKA_BHUTAN, "dz"},
	/* end of inconsistency */

	{truetype.PlatformMicrosoft, msLangID_WELSH_UNITED_KINGDOM, "cy"},
	{truetype.PlatformMicrosoft, msLangID_KHMER_CAMBODIA, "km"},
	{truetype.PlatformMicrosoft, msLangID_LAO_LAOS, "lo"},
	{truetype.PlatformMicrosoft, msLangID_BURMESE_MYANMAR, "my"},
	{truetype.PlatformMicrosoft, msLangID_GALICIAN_GALICIAN, "gl"},
	{truetype.PlatformMicrosoft, msLangID_MANIPURI_INDIA, "mni"},
	{truetype.PlatformMicrosoft, msLangID_SINDHI_INDIA, "sd"},
	// the following one is only encountered in Microsoft RTF specification
	{truetype.PlatformMicrosoft, msLangID_KASHMIRI_PAKISTAN, "ks"},
	// the following one is not in the Passport list, looks like an omission
	{truetype.PlatformMicrosoft, msLangID_KASHMIRI_SASIA, "ks"},
	{truetype.PlatformMicrosoft, msLangID_NEPALI_NEPAL, "ne"},
	// {truetype.PlatformMicrosoft, msLangID_NEPALI_INDIA, "ne"},
	{truetype.PlatformMicrosoft, msLangID_FRISIAN_NETHERLANDS, "fy"},

	// new as of 2001-03-01 (from Office Xp)
	{truetype.PlatformMicrosoft, msLangID_ENGLISH_HONG_KONG, "en"},
	{truetype.PlatformMicrosoft, msLangID_ENGLISH_INDIA, "en"},
	{truetype.PlatformMicrosoft, msLangID_ENGLISH_MALAYSIA, "en"},
	{truetype.PlatformMicrosoft, msLangID_ENGLISH_SINGAPORE, "en"},
	{truetype.PlatformMicrosoft, msLangID_SYRIAC_SYRIA, "syr"},
	{truetype.PlatformMicrosoft, msLangID_SINHALA_SRI_LANKA, "si"},
	{truetype.PlatformMicrosoft, msLangID_CHEROKEE_UNITED_STATES, "chr"},
	{truetype.PlatformMicrosoft, msLangID_INUKTITUT_CANADA, "iu"},
	{truetype.PlatformMicrosoft, msLangID_AMHARIC_ETHIOPIA, "am"},

	{truetype.PlatformMicrosoft, msLangID_PASHTO_AFGHANISTAN, "ps"},
	{truetype.PlatformMicrosoft, msLangID_FILIPINO_PHILIPPINES, "phi"},
	{truetype.PlatformMicrosoft, msLangID_DHIVEHI_MALDIVES, "div"},

	{truetype.PlatformMicrosoft, msLangID_OROMO_ETHIOPIA, "om"},
	{truetype.PlatformMicrosoft, msLangID_TIGRIGNA_ETHIOPIA, "ti"},
	{truetype.PlatformMicrosoft, msLangID_TIGRIGNA_ERYTHREA, "ti"},

	/* New additions from Windows Xp/Passport SDK 2001-11-10. */

	{truetype.PlatformMicrosoft, msLangID_SPANISH_UNITED_STATES, "es"},
	// The following two IDs blatantly violate MS specs by using a sublanguage >,.                                         */
	{truetype.PlatformMicrosoft, msLangID_SPANISH_LATIN_AMERICA, "es"},
	{truetype.PlatformMicrosoft, msLangID_FRENCH_NORTH_AFRICA, "fr"},

	{truetype.PlatformMicrosoft, msLangID_FRENCH_MOROCCO, "fr"},
	{truetype.PlatformMicrosoft, msLangID_FRENCH_HAITI, "fr"},
	{truetype.PlatformMicrosoft, msLangID_BENGALI_BANGLADESH, "bn"},
	{truetype.PlatformMicrosoft, msLangID_PUNJABI_ARABIC_PAKISTAN, "ar"},
	{truetype.PlatformMicrosoft, msLangID_MONGOLIAN_PRC, "mn"},
	{truetype.PlatformMicrosoft, msLangID_HAUSA_NIGERIA, "ha"},
	{truetype.PlatformMicrosoft, msLangID_YORUBA_NIGERIA, "yo"},
	/* language codes from, to, are (still) unknown. */
	{truetype.PlatformMicrosoft, msLangID_IGBO_NIGERIA, "ibo"},
	{truetype.PlatformMicrosoft, msLangID_KANURI_NIGERIA, "kau"},
	{truetype.PlatformMicrosoft, msLangID_GUARANI_PARAGUAY, "gn"},
	{truetype.PlatformMicrosoft, msLangID_HAWAIIAN_UNITED_STATES, "haw"},
	{truetype.PlatformMicrosoft, msLangID_LATIN, "la"},
	{truetype.PlatformMicrosoft, msLangID_SOMALI_SOMALIA, "so"},

	/* Note: Yi does not have a (proper) ISO 639-2 code, since it is mostly */
	/*       not written (but OTOH the peculiar writing system is worth     */
	/*       studying).                                                     */
	//   {  truetype.PlatformMicrosoft,	msLangID_YI_CHINA },

	{truetype.PlatformMicrosoft, msLangID_PAPIAMENTU_NETHERLANDS_ANTILLES, "pap"},
}

const (
	encMacRoman      = "MACINTOSH"
	encodingDontCare = 0xffff

	macIdJapanese = truetype.PlatformEncodingID(1)

	msIdSjis    = truetype.PlatformEncodingID(2)
	msIdPrc     = truetype.PlatformEncodingID(3)
	msIdBig_5   = truetype.PlatformEncodingID(4)
	msIdWansung = truetype.PlatformEncodingID(5)
	msIdJohab   = truetype.PlatformEncodingID(6)

	isoId_7BitAscii = truetype.PlatformEncodingID(0)
	isoId_10646     = truetype.PlatformEncodingID(1)
	isoId_8859_1    = truetype.PlatformEncodingID(2)
)

var fcFtEncoding = [...]struct {
	PlatformID truetype.PlatformID
	EncodingID truetype.PlatformEncodingID
	fromcode   encoding.Encoding
}{
	{truetype.PlatformMac, encodingDontCare, unicode.UTF16(unicode.BigEndian, unicode.IgnoreBOM)},
	{truetype.PlatformMac, truetype.PEMacRoman, charmap.Macintosh},
	{truetype.PlatformMac, macIdJapanese, japanese.ShiftJIS},
	{truetype.PlatformMicrosoft, truetype.PEMicrosoftSymbolCs, unicode.UTF16(unicode.BigEndian, unicode.IgnoreBOM)},
	{truetype.PlatformMicrosoft, truetype.PEMicrosoftUnicodeCs, unicode.UTF16(unicode.BigEndian, unicode.IgnoreBOM)},
	{truetype.PlatformMicrosoft, msIdSjis, japanese.ShiftJIS},
	{truetype.PlatformMicrosoft, msIdPrc, simplifiedchinese.HZGB2312},
	{truetype.PlatformMicrosoft, msIdBig_5, traditionalchinese.Big5},
	{truetype.PlatformMicrosoft, msIdWansung, korean.EUCKR},
	// {truetype.PlatformMicrosoft, msIdJohab, "Johab"}, // Johab is not supported by golang.x/text/encoding
	{truetype.PlatformMicrosoft, truetype.PEMicrosoftUcs4, unicode.UTF16(unicode.BigEndian, unicode.IgnoreBOM)},
	{truetype.PlatformIso, isoId_7BitAscii, charmap.ISO8859_1},
	{truetype.PlatformIso, isoId_10646, unicode.UTF16(unicode.BigEndian, unicode.IgnoreBOM)},
	{truetype.PlatformIso, isoId_8859_1, charmap.ISO8859_1},
}

var fcMacRomanFake = [...]struct {
	fromcode encoding.Encoding
	language truetype.PlatformLanguageID
}{
	{japanese.ShiftJIS, msLangID_JAPANESE_JAPAN},
	{charmap.ISO8859_1, msLangID_ENGLISH_UNITED_STATES},
}

// A shift-JIS will have many high bits turned on
func looksLikeSJIS(str []byte) bool {
	var nhigh, nlow int

	for _, b := range str {
		if (b & 0x80) != 0 {
			nhigh++
		} else {
			nlow++
		}
	}
	/* Heuristic -- if more than 1/3 of the bytes have the high-bit set,
	 * this is likely to be SJIS and not ROMAN  */
	return nhigh*2 > nlow
}

func nameTranscode(sname truetype.NameEntry) string {
	var fromcode encoding.Encoding
	for _, fcEnc := range fcFtEncoding {
		if fcEnc.PlatformID == sname.PlatformID &&
			(fcEnc.EncodingID == encodingDontCare || fcEnc.EncodingID == sname.EncodingID) {
			fromcode = fcEnc.fromcode
			break
		}
	}
	if fromcode == nil {
		// can't find an encoding, just return the raw bytes
		return string(sname.Value)
	}

	/*  Many names encoded for truetype.PlatformMac are broken
	 * in various ways. Kludge around them. */
	if fromcode == charmap.Macintosh {
		if sname.LanguageID == macLangID_ENGLISH && looksLikeSJIS(sname.Value) {
			fromcode = japanese.ShiftJIS
		} else if sname.LanguageID >= 0x100 {
			/* "real" Mac language IDs are all less than 150.
			 * Names using one of the MS language IDs are assumed
			 * to use an associated encoding (Yes, this is a kludge) */

			fromcode = nil
			for _, macFake := range fcMacRomanFake {
				if macFake.language == sname.LanguageID {
					fromcode = macFake.fromcode
					break
				}
			}
			if fromcode == nil {
				// can't find an encoding, just return the raw bytes
				return string(sname.Value)
			}
		}
	}

	out, _ := fromcode.NewDecoder().Bytes(sname.Value)
	return string(out)
}

func nameLanguage(sname truetype.NameEntry) string {
	platformID := sname.PlatformID
	languageID := sname.LanguageID

	/* Many names encoded for truetype.PlatformMac are broken
	 * in various ways. Kludge around them. */
	if platformID == truetype.PlatformMac && sname.EncodingID == truetype.PEMacRoman &&
		looksLikeSJIS(sname.Value) {
		languageID = macLangID_JAPANESE
	}

	for _, langEntry := range fcFtLanguage {
		if langEntry.PlatformID == platformID &&
			(langEntry.LanguageID == lang_DONT_CARE || langEntry.LanguageID == languageID) {
			return langEntry.lang
		}
	}
	return ""
}

/* Order is significant.  For example, some B&H fonts are hinted by
   URW++, and both strings appear in the notice. */
var noticeFoundries = [...][2]string{
	{"Adobe", "adobe"},
	{"Bigelow", "b&h"},
	{"Bitstream", "bitstream"},
	{"Gnat", "culmus"},
	{"Iorsh", "culmus"},
	{"HanYang System", "hanyang"},
	{"Font21", "hwan"},
	{"IBM", "ibm"},
	{"International Typeface Corporation", "itc"},
	{"Linotype", "linotype"},
	{"LINOTYPE-HELL", "linotype"},
	{"Microsoft", "microsoft"},
	{"Monotype", "monotype"},
	{"Omega", "omega"},
	{"Tiro Typeworks", "tiro"},
	{"URW", "urw"},
	{"XFree86", "xfree86"},
	{"Xorg", "xorg"},
}

func noticeFoundry(notice string) string {
	for _, entry := range noticeFoundries {
		if strings.Contains(notice, entry[0]) {
			return entry[1]
		}
	}
	return ""
}

type stringConst struct {
	name  string
	value int
}

func stringIsConst(str string, c []stringConst) int {
	for _, v := range c {
		if cmpIgnoreBlanksAndCase(str, v.name) == 0 {
			return v.value
		}
	}
	return -1
}

func isPunct(c byte) bool {
	switch {
	case c < '0':
		return true
	case c <= '9':
		return false
	case c < 'A':
		return true
	case c <= 'Z':
		return false
	case c < 'a':
		return true
	case c <= 'z':
		return false
	case c <= '~':
		return true
	default:
		return false
	}
}

// Does s1 contain an instance of s2 on a word boundary (ignoring case)?
func containsWord(s1, s2 []byte) bool {
	wordStart := true
	s1len := len(s1)
	s2len := len(s2)

	for len(s1) >= len(s2) {
		if wordStart && bytes.HasPrefix(bytes.ToLower(s1), bytes.ToLower(s2)) &&
			(s1len == s2len || isPunct(s1[s2len])) {
			return true
		}
		wordStart = false
		if isPunct(s1[0]) {
			wordStart = true
		}
		s1 = s1[1:]
	}
	return false
}

// Does s1 contain an instance of s2 (ignoring blanks and case)?
func containsIgnoreBlanksAndCase(s1 []byte, s2 string) bool {
	for len(s1) != 0 {
		if strings.HasPrefix(ignoreBlanksAndCase(string(s1)), ignoreBlanksAndCase(s2)) {
			return true
		}
		s1 = s1[1:]
	}
	return false
}

func stringContainsConst(str string, c []stringConst) int {
	for _, v := range c {
		if v.name[0] == '<' {
			if containsWord([]byte(str), []byte(v.name[1:])) {
				return v.value
			}
		} else {
			if containsIgnoreBlanksAndCase([]byte(str), v.name) {
				return v.value
			}
		}
	}
	return -1
}

var weightConsts = [...]stringConst{
	{"thin", WEIGHT_THIN},
	{"extralight", WEIGHT_EXTRALIGHT},
	{"ultralight", WEIGHT_ULTRALIGHT},
	{"demilight", WEIGHT_DEMILIGHT},
	{"semilight", WEIGHT_SEMILIGHT},
	{"light", WEIGHT_LIGHT},
	{"book", WEIGHT_BOOK},
	{"regular", WEIGHT_REGULAR},
	{"normal", WEIGHT_NORMAL},
	{"medium", WEIGHT_MEDIUM},
	{"demibold", WEIGHT_DEMIBOLD},
	{"demi", WEIGHT_DEMIBOLD},
	{"semibold", WEIGHT_SEMIBOLD},
	{"extrabold", WEIGHT_EXTRABOLD},
	{"superbold", WEIGHT_EXTRABOLD},
	{"ultrabold", WEIGHT_ULTRABOLD},
	{"bold", WEIGHT_BOLD},
	{"ultrablack", WEIGHT_ULTRABLACK},
	{"superblack", WEIGHT_EXTRABLACK},
	{"extrablack", WEIGHT_EXTRABLACK},
	{"<ultra", WEIGHT_ULTRABOLD}, /* only if a word */
	{"black", WEIGHT_BLACK},
	{"heavy", WEIGHT_HEAVY},
}

var widthConsts = [...]stringConst{
	{"ultracondensed", WIDTH_ULTRACONDENSED},
	{"extracondensed", WIDTH_EXTRACONDENSED},
	{"semicondensed", WIDTH_SEMICONDENSED},
	{"condensed", WIDTH_CONDENSED}, /* must be after *condensed */
	{"normal", WIDTH_NORMAL},
	{"semiexpanded", WIDTH_SEMIEXPANDED},
	{"extraexpanded", WIDTH_EXTRAEXPANDED},
	{"ultraexpanded", WIDTH_ULTRAEXPANDED},
	{"expanded", WIDTH_EXPANDED}, /* must be after *expanded */
	{"extended", WIDTH_EXPANDED},
}

var slantConsts = [...]stringConst{
	{"italic", SLANT_ITALIC},
	{"kursiv", SLANT_ITALIC},
	{"oblique", SLANT_OBLIQUE},
}

var decorativeConsts = [...]stringConst{
	{"shadow", 1},
	{"caps", 1},
	{"antiqua", 1},
	{"romansc", 1},
	{"embosed", 1},
	{"dunhill", 1},
}

// return true if `str` is at `obj`, ignoring blank and case
func (pat Pattern) hasString(obj Object, str string) bool {
	for _, v := range pat.getVals(obj) {
		vs, ok := v.Value.(String)
		if ok && cmpIgnoreBlanksAndCase(string(vs), str) == 0 {
			return true
		}
	}
	return false
}

var platformOrder = [...]truetype.PlatformID{
	truetype.PlatformMicrosoft,
	truetype.PlatformUnicode,
	truetype.PlatformMac,
	truetype.PlatformIso,
}

var nameidOrder = [...]truetype.NameID{
	truetype.NameWWSFamily,
	truetype.NamePreferredFamily, // Typographic
	truetype.NameFontFamily,
	truetype.NameCompatibleFull, // MacFullname
	truetype.NameFull,
	truetype.NameWWSSubfamily,
	truetype.NamePreferredSubfamily, // TypographicSub
	truetype.NameFontSubfamily,
	truetype.NameTrademark,
	truetype.NameManufacturer,
}

type nameMapping struct {
	truetype.NameEntry
	idx uint
}

func isEnglish(platform truetype.PlatformID, language truetype.PlatformLanguageID) bool {
	switch platform {
	case truetype.PlatformMac:
		return language == truetype.PLMacEnglish
	case truetype.PlatformMicrosoft:
		return language == truetype.PLMicrosoftEnglish
	}
	return false
}

func isLess(a, b nameMapping) bool {
	if a.PlatformID != b.PlatformID {
		return a.PlatformID < b.PlatformID
	}
	if a.NameID != b.NameID {
		return a.NameID < b.NameID
	}
	if a.EncodingID != b.EncodingID {
		return a.EncodingID < b.EncodingID
	}
	if a.LanguageID != b.LanguageID {
		if isEnglish(a.PlatformID, a.LanguageID) {
			return true
		}
		if isEnglish(b.PlatformID, b.LanguageID) {
			return false
		}
		return a.LanguageID < b.LanguageID
	}
	return a.idx < b.idx
}

// return -1 if not found
func getFirstName(nameTable truetype.TableName, platform truetype.PlatformID, nameid truetype.NameID,
	mapping []nameMapping) (int, truetype.NameEntry) {
	min, max := 0, len(mapping)-1

	for min <= max {
		mid := (min + max) / 2

		sname := nameTable[mapping[mid].idx]

		if platform < sname.PlatformID ||
			(platform == sname.PlatformID &&
				(nameid < sname.NameID ||
					(nameid == sname.NameID &&
						(mid != 0 && platform == mapping[mid-1].PlatformID &&
							nameid == mapping[mid-1].NameID)))) {
			max = mid - 1
		} else if platform > sname.PlatformID ||
			(platform == sname.PlatformID &&
				nameid > sname.NameID) {
			min = mid + 1
		} else {
			return mid, sname
		}
	}

	return -1, truetype.NameEntry{}
}

type FT_Face struct {
	num_faces int
	// face_index int

	face_flags  int
	style_flags int

	// num_glyphs int

	family_name string
	style_name  string

	available_sizes []fonts.BitmapSize // length num_fixed_sizes

	// charmaps []FT_CharMap // length num_charmaps

	// FT_Generic generic

	/*# The following member variables (down to `underline_thickness`) */
	/*# are only relevant to scalable outlines; cf. @FT_Bitmap_Size    */
	/*# for bitmap fonts.                                              */
	// FT_BBox bbox

	// units_per_EM uint16
	// ascender     int16
	// descender    int16
	// height       int16

	// max_advance_width  int16
	// max_advance_height int16

	// underline_position  int16
	// underline_thickness int16

	// FT_GlyphSlot glyph
	// FT_Size      size
	// FT_CharMap   charmap
}

var (
	wght = truetype.MustNewTag("wght")
	wdth = truetype.MustNewTag("wdth")
	opsz = truetype.MustNewTag("opsz")
)

func hasHint(face *truetype.Font) bool { return face.HasTable(truetype.TagPrep) }

// TODO: implements all these methods
// TODO:
func FT_Get_MM_Var(face FT_Face) *truetype.TableFvar { return nil }

type FT_Glyph_Format uint8

const (
	FT_GLYPH_FORMAT_NONE FT_Glyph_Format = iota
	FT_GLYPH_FORMAT_COMPOSITE
	FT_GLYPH_FORMAT_BITMAP
	FT_GLYPH_FORMAT_OUTLINE
	FT_GLYPH_FORMAT_PLOTTER
)

type FT_Outline struct {
	points []fixed.Rectangle26_6 /* the outline's points, length n_points               */
	// tags []byte                 /* the points flags                   */
	contours []int16 /* the contour end points, length n_contours            */

	// int flags /* outline masks                      */
}

type LoadFlags uint32

const (
	FT_LOAD_DEFAULT  LoadFlags = 0x0
	FT_LOAD_NO_SCALE LoadFlags = 1 << iota
	FT_LOAD_NO_HINTING
	FT_LOAD_RENDER
	FT_LOAD_NO_BITMAP
	FT_LOAD_VERTICAL_LAYOUT
	FT_LOAD_FORCE_AUTOHINT
	FT_LOAD_CROP_BITMAP
	FT_LOAD_PEDANTIC
	FT_LOAD_IGNORE_GLOBAL_ADVANCE_WIDTH
	FT_LOAD_NO_RECURSE
	FT_LOAD_IGNORE_TRANSFORM
	FT_LOAD_MONOCHROME
	FT_LOAD_LINEAR_DESIGN
	FT_LOAD_NO_AUTOHINT
	/* Bits 16-19 are used by `FT_LOAD_TARGET_` */
	FT_LOAD_COLOR
	FT_LOAD_COMPUTE_METRICS
	FT_LOAD_BITMAP_METRICS_ONLY
)

// TODO:
func FT_Get_Advance(face FT_Face, glyph fonts.GID, loadFlags LoadFlags) (int32, bool) {
	return 0, false
}

// TODO:
func FT_Select_Size(face FT_Face, strikeIndex int) {}

// see `loaders`
func getFontFormat(face fonts.Face) string {
	switch face.(type) {
	case *truetype.Font:
		return "TrueType"
	case *bitmap.Font:
		return "PCF"
	case *type1.Font:
		return "Type 1"
	case *type1c.Font:
		return "CFF"
	default:
		return ""
	}
}

// load various information from the font to build the pattern
// this is the core of the library
func queryFace(face fonts.Face, file string, id uint32) (Pattern, []nameMapping, Charset, Langset) {
	var (
		variableWeight, variableWidth, variableSize, variable bool
		weight, width                                         float32 = -1., -1.

		// Support for glyph-variation named-instances.
		instance              *truetype.VarInstance
		weightMult, widthMult float32 = 1., 1.

		foundry                                                                       string
		nameCount, nfamily, nfamilyLang, nstyle, nstyleLang, nfullname, nfullnameLang int
		exclusiveLang                                                                 string
		slant                                                                         int32
		decorative                                                                    bool
	)

	pat := NewPattern()

	summary, _ := face.LoadSummary()

	pat.AddBool(OUTLINE, summary.HasScalableGlyphs)
	pat.AddBool(COLOR, summary.HasColorGlyphs)

	/* All color fonts are designed to be scaled, even if they only have
	 * bitmap strikes.  Client is responsible to scale the bitmaps.  This
	 * is in contrast to non-color strikes... */
	pat.AddBool(SCALABLE, summary.HasScalableGlyphs || summary.HasColorGlyphs)

	if id>>16 != 0 {
		master := FT_Get_MM_Var(FT_Face{}) // TODO:
		if master == nil {
			return nil, nil, Charset{}, Langset{}
		}

		if id>>16 == 0x8000 {
			// Query variable font itself.

			for _, axis := range master.Axis {
				defValue := float32(axis.Default / (1 << 16))
				minValue := float32(axis.Minimum / (1 << 16))
				maxValue := float32(axis.Maximum / (1 << 16))

				if minValue > defValue || defValue > maxValue || minValue == maxValue {
					continue
				}

				var obj Object
				switch axis.Tag {
				case wght:
					obj = WEIGHT
					minValue = WeightFromOT(minValue)
					maxValue = WeightFromOT(maxValue)
					variableWeight = true
					weight = 0 // To stop looking for weight.

				case wdth:
					obj = WIDTH
					// Values in 'wdth' match Fontconfig WIDTH_* scheme directly.
					variableWidth = true
					width = 0 // To stop looking for width.

				case opsz:
					obj = SIZE
					// Values in 'opsz' match Fontconfig SIZE, both are in points.
					variableSize = true
				}

				if obj != invalid {
					r := Range{Begin: minValue, End: maxValue}
					pat.Add(obj, r, true)
					variable = true
				}
			}

			if !variable {
				return nil, nil, Charset{}, Langset{}
			}

			id &= 0xFFFF
		} else if index := (id >> 16) - 1; int(index) < len(master.Instances) {
			// Pull out weight and width from named-instance.

			instance = &master.Instances[index]

			for i, axis := range master.Axis {
				value := instance.Coords[i] / (1 << 16)
				defaultValue := axis.Default / (1 << 16)
				mult := float32(1.)
				if defaultValue != 0 {
					mult = float32(value / defaultValue)
				}
				switch axis.Tag {
				case wght:
					weightMult = mult

				case wdth:
					widthMult = mult

				case opsz:
					pat.AddFloat(SIZE, float32(value))
				}
			}
		} else {
			return nil, nil, Charset{}, Langset{}
		}
	}

	pat.AddBool(VARIABLE, variable)

	var (
		os2   *truetype.TableOS2
		names truetype.TableName
		head  *truetype.TableHead
	)
	// Get the OS/2 table
	if ttf, ok := face.(*truetype.Font); ok {
		os2, _ = ttf.OS2Table()
		names = ttf.Names
		head = &ttf.Head
	}

	/*
	 * Look first in the OS/2 table for the foundry, if
	 * not found here, the various notices will be searched for
	 * that information, either from the sfnt name tables or
	 * the Postscript FontInfo dictionary.  Finally, the
	 * BDF properties will be queried.
	 */

	if os2 != nil && os2.Version >= 0x0001 && os2.Version != 0xffff {
		if os2.AchVendID != 0 {
			foundry = os2.AchVendID.String()
		}
	}

	/* Grub through the name table looking for family
	 * and style names. FreeType makes quite a hash of them */
	nameMappings := make([]nameMapping, len(names))
	for i, p := range names {
		nameMappings[i] = nameMapping{NameEntry: p, idx: uint(i)}
	}

	sort.Slice(nameMappings, func(i, j int) bool { return isLess(nameMappings[i], nameMappings[j]) })

	for _, platform := range platformOrder {
		// Order nameids so preferred names appear first in the resulting list
		for _, nameid := range nameidOrder {
			obj, objlang := invalid, invalid

			lookupid := nameid

			if instance != nil {
				/* For named-instances, we skip regular style nameIDs,
				 * and treat the instance's nameid as FONT_SUBFAMILY.
				 * Postscript name is automatically handled by FreeType. */
				if nameid == truetype.NameWWSSubfamily ||
					nameid == truetype.NamePreferredSubfamily ||
					nameid == truetype.NameFull {
					continue
				}

				if nameid == truetype.NameFontSubfamily {
					lookupid = instance.PSStringID
				}
			}

			nameidx, sname := getFirstName(names, platform, lookupid, nameMappings)
			if nameidx == -1 {
				continue
			}

			var (
				np, nlangp *int
				lang       string
			)

			for do := true; do; {
				sname = names[nameMappings[nameidx].idx]
				do = nameidx < nameCount && platform == sname.PlatformID && lookupid == sname.NameID
				nameidx++

				switch nameid {
				case truetype.NameWWSFamily, truetype.NamePreferredFamily, truetype.NameFontFamily:
					if debugMode {
						fmt.Printf("\tfound family (n %2d p %d e %d l 0x%04x)\n",
							sname.NameID, sname.PlatformID,
							sname.EncodingID, sname.LanguageID)
					}

					obj = FAMILY
					objlang = FAMILYLANG
					np = &nfamily
					nlangp = &nfamilyLang
				case truetype.NameCompatibleFull, truetype.NameFull:
					if variable {
						break
					}
					if debugMode {
						fmt.Printf("\tfound full   (n %2d p %d e %d l 0x%04x)\n",
							sname.NameID, sname.PlatformID,
							sname.EncodingID, sname.LanguageID)
					}

					obj = FULLNAME
					objlang = FULLNAMELANG
					np = &nfullname
					nlangp = &nfullnameLang
				case truetype.NameWWSSubfamily, truetype.NamePreferredSubfamily, truetype.NameFontSubfamily:
					if variable {
						break
					}
					if debugMode {
						fmt.Printf("\tfound style  (n %2d p %d e %d l 0x%04x)\n",
							sname.NameID, sname.PlatformID,
							sname.EncodingID, sname.LanguageID)
					}

					obj = STYLE
					objlang = STYLELANG
					np = &nstyle
					nlangp = &nstyleLang
				case truetype.NameTrademark, truetype.NameManufacturer:
					// If the foundry wasn't found in the OS/2 table, look here
					if foundry == "" {
						utf8 := nameTranscode(sname)
						foundry = noticeFoundry(utf8)
					}
				}
				if obj != invalid {
					utf8 := nameTranscode(sname)
					lang = nameLanguage(sname)

					if debugMode {
						fmt.Println("\ttranscoded: ", utf8)
					}
					if utf8 == "" {
						continue
					}

					// Trim surrounding whitespace.
					utf8 = strings.TrimSpace(utf8)

					if pat.hasString(obj, utf8) {
						continue
					}

					// add new element
					pat.AddString(obj, utf8)

					if lang != "" {
						// pad lang list with 'und' to line up with elt
						for *nlangp < *np {
							pat.AddString(objlang, "und")
							*nlangp++
						}
						pat.AddString(objlang, lang)
						*nlangp++
					}
					*np++
				}
			}

		}
	}

	// if !nm_share {
	// 	free(nameMapping)
	// 	nameMapping = nil
	// }

	if nfamily == 0 && cmpIgnoreBlanksAndCase(summary.Familly, "") != 0 {
		if debugMode {
			fmt.Printf("\tusing FreeType family \"%s\"\n", summary.Familly)
		}
		pat.AddString(FAMILY, summary.Familly)
		pat.AddString(FAMILYLANG, "en")
		nfamily++
	}

	if !variable && nstyle == 0 {
		if debugMode {
			fmt.Printf("\tusing FreeType style \"%s\"\n", summary.Style)
		}
		pat.AddString(STYLE, summary.Style)
		pat.AddString(STYLELANG, "en")
		nstyle++
	}

	if nfamily == 0 && file != "" {
		//  FcChar8	*start, *end;
		//  FcChar8	*family;

		start := strings.IndexByte(file, '/')
		end := strings.IndexByte(file, '.')
		if end == -1 {
			end = len(file)
		}
		family := file[start+1 : end]
		if debugMode {
			fmt.Printf("\tusing filename for family %s\n", family)
		}
		pat.AddString(FAMILY, family)
		pat.AddString(FAMILYLANG, "en")
		nfamily++
	}

	// Add the PostScript name into the cache
	if !variable {
		psname := face.PoscriptName()
		if psname == "" {
			/* Workaround when PoscriptName didn't give any name.
			* try to find out the English family name and convert. */
			n := 0
			familylang, res := pat.GetAtString(FAMILYLANG, n)
			for ; res == ResultMatch; familylang, res = pat.GetAtString(FAMILYLANG, n) {
				if familylang == "en" {
					break
				}
				n++
			}
			if familylang == "" {
				n = 0
			}

			family, res := pat.GetAtString(FAMILY, n)
			if res != ResultMatch {
				return nil, nil, Charset{}, Langset{}
			}
			psname = strings.Map(func(r rune) rune {
				switch r {
				// those characters are not allowed to be the literal name in PostScript
				case '\x04', '(', ')', '/', '<', '>', '[', ']', '{', '}', '\t', '\f', '\r', '\n', ' ':
					return '-'
				default:
					return r
				}
			}, family)
		}
		pat.AddString(POSTSCRIPT_NAME, psname)
	}

	if file != "" {
		pat.AddString(FILE, file)
	}
	pat.AddInt(INDEX, int32(id))

	// don't even try using FT_FACE_FLAG_FIXED_WIDTH -- CJK 'monospace' fonts are really
	// dual width, and most other fonts don't bother to set
	// the attribute.  Sigh.

	// Find the font revision (if available)
	if head != nil {
		pat.AddInt(FONTVERSION, int32(head.FontRevision))
	} else {
		pat.AddInt(FONTVERSION, 0)
	}
	pat.AddInt(ORDER, 0)

	if os2 != nil && os2.Version >= 0x0001 && os2.Version != 0xffff {
		for _, codePage := range codePageRange {
			var (
				bit  byte
				bits uint32
			)
			if codePage.bit < 32 {
				bits = os2.UlCodePageRange1
				bit = codePage.bit
			} else {
				bits = os2.UlCodePageRange2
				bit = codePage.bit - 32
			}
			if bits&(1<<bit) != 0 {
				/*
				 * If the font advertises support for multiple
				 * "exclusive" languages, then include support
				 * for any language found to have coverage
				 */
				if exclusiveLang != "" {
					exclusiveLang = ""
					break
				}
				exclusiveLang = codePage.lang
			}
		}
	}

	if os2 != nil && os2.Version != 0xffff {
		weight = float32(os2.USWeightClass)
		weight = WeightFromOT(weight * weightMult)
		if debugMode && weight != -1 {
			fmt.Printf("\tos2 weight class %d multiplier %g maps to weight %g\n",
				os2.USWeightClass, weightMult, weight)
		}

		switch os2.USWidthClass {
		case 1:
			width = WIDTH_ULTRACONDENSED
		case 2:
			width = WIDTH_EXTRACONDENSED
		case 3:
			width = WIDTH_CONDENSED
		case 4:
			width = WIDTH_SEMICONDENSED
		case 5:
			width = WIDTH_NORMAL
		case 6:
			width = WIDTH_SEMIEXPANDED
		case 7:
			width = WIDTH_EXPANDED
		case 8:
			width = WIDTH_EXTRAEXPANDED
		case 9:
			width = WIDTH_ULTRAEXPANDED
		}
		width *= widthMult
		if debugMode && width != -1 {
			fmt.Printf("\tos2 width class %d multiplier %g maps to width %g\n",
				os2.USWidthClass, widthMult, width)
		}
	}

	if face, ok := face.(*truetype.Font); ok {
		if complexFeats := fontCapabilities(face); os2 != nil && complexFeats != "" {
			if debugMode {
				fmt.Printf("\tcomplex features in this font: %s\n", complexFeats)
			}
			pat.AddString(CAPABILITY, complexFeats)
		}

		pat.AddBool(FONT_HAS_HINT, hasHint(face))
	}

	if !variableSize && os2 != nil && os2.Version >= 0x0005 && os2.Version != 0xffff {
		// usLowerPointSize and usUpperPointSize is actually twips
		lowerSize := float32(os2.UsLowerPointSize) / 20.0
		upperSize := float32(os2.UsUpperPointSize) / 20.0

		if lowerSize == upperSize {
			pat.AddFloat(SIZE, lowerSize)
		} else {
			pat.Add(SIZE, Range{Begin: lowerSize, End: upperSize}, true)
		}
	}

	/* Type 1: Check for FontInfo dictionary information
	 * Code from g2@magestudios.net (Gerard Escalante) */
	if psfontinfo, ok := face.PostscriptInfo(); ok {
		if weight == -1 && psfontinfo.Weight != "" {
			weight = float32(stringIsConst(psfontinfo.Weight, weightConsts[:]))
			if debugMode {
				fmt.Printf("\tType1 weight %s maps to %g\n", psfontinfo.Weight, weight)
			}
		}
		if foundry == "" {
			foundry = noticeFoundry(psfontinfo.Notice)
		}
	}

	// Finally, look for a FOUNDRY BDF property if no other mechanism has managed to locate a foundry
	if face, isBDF := face.(*bitmap.Font); isBDF {
		if foundry == "" {
			prop := face.GetBDFProperty("FOUNDRY")
			if atom, ok := prop.(bitmap.Atom); ok {
				foundry = string(atom)
			}
		}
		if width == -1 {
			if propInt, isInt := face.GetBDFProperty("RELATIVE_SETWIDTH").(bitmap.Int); isInt {
				width = weightFromBFD(int32(propInt))
			}
			if width == -1 {
				if atom, _ := face.GetBDFProperty("SETWIDTH_NAME").(bitmap.Atom); atom != "" {
					width = float32(stringIsConst(string(atom), widthConsts[:]))
					if debugMode {
						fmt.Printf("\tsetwidth %s maps to %g\n", atom, width)
					}
				}
			}
		}
	}

	// Look for weight, width and slant names in the style value
	st := 0
	style, res := pat.GetAtString(STYLE, st)
	for ; res == ResultMatch; st++ {
		style, res = pat.GetAtString(STYLE, st)

		if weight == -1 {
			weight = float32(stringContainsConst(style, weightConsts[:]))
			if debugMode {
				fmt.Printf("\tStyle %s maps to weight %g\n", style, weight)
			}
		}
		if width == -1 {
			width = float32(stringContainsConst(style, widthConsts[:]))
			if debugMode {
				fmt.Printf("\tStyle %s maps to width %g\n", style, width)
			}
		}
		if slant == -1 {
			slant = int32(stringContainsConst(style, slantConsts[:]))
			if debugMode {
				fmt.Printf("\tStyle %s maps to slant %d\n", style, slant)
			}
		}
		if decorative == false {
			decorative = stringContainsConst(style, decorativeConsts[:]) > 0
			if debugMode {
				fmt.Printf("\tStyle %s maps to decorative %v\n", style, decorative)
			}
		}
	}

	// Pull default values from the summary if more specific values not found above
	if slant == -1 {
		slant = SLANT_ROMAN
		if summary.IsItalic {
			slant = SLANT_ITALIC
		}
	}

	if weight == -1 {
		weight = WEIGHT_MEDIUM
		if summary.IsBold {
			weight = WEIGHT_BOLD
		}
	}

	if width == -1 {
		width = WIDTH_NORMAL
	}

	if foundry == "" {
		foundry = "unknown"
	}

	pat.AddInt(SLANT, slant)

	if !variableWeight {
		pat.AddFloat(WEIGHT, weight)
	}

	if !variableWidth {
		pat.AddFloat(WIDTH, width)
	}

	pat.AddString(FOUNDRY, foundry)

	pat.AddBool(DECORATIVE, decorative)

	//  Compute the unicode coverage for the font
	cs, enc := getCharSet(face)
	if enc == fonts.EncOther {
		return nil, nil, Charset{}, Langset{}
	}

	// getCharSet() chose the encoding; test it for symbol.
	symbol := enc == fonts.EncSymbol
	pat.AddBool(SYMBOL, symbol)
	spacing := getSpacing(face, head, nil) // TODO:

	// For PCF fonts, override the computed spacing with the one from the property
	if face, isBDF := face.(*bitmap.Font); isBDF {
		if prop, _ := face.GetBDFProperty("SPACING").(bitmap.Atom); prop != "" {
			switch prop {
			case "c", "C":
				spacing = CHARCELL
			case "m", "M":
				spacing = MONO
			case "p", "P":
				spacing = PROPORTIONAL
			}
		}

		// Skip over PCF fonts that have no encoded characters; they're
		// usually just Unicode fonts transcoded to some legacy encoding
		if cs.Len() == 0 {
			return nil, nil, Charset{}, Langset{}
		}
	}

	pat.Add(CHARSET, cs, true)

	var ls Langset
	// Symbol fonts don't cover any language, even though they
	// claim to support Latin1 range.
	if !symbol {
		if debugMode {
			fmt.Printf("\tfont charset: %v \n", cs)
		}
		ls = buildLangSet(cs, exclusiveLang)
	}

	pat.Add(LANG, ls, true)

	if spacing != PROPORTIONAL {
		pat.AddInt(SPACING, spacing)
	}

	if !summary.HasScalableGlyphs {
		for _, size := range face.LoadBitmaps() {
			pat.AddFloat(PIXEL_SIZE, float32(size.YPpem))
		}
		pat.AddBool(ANTIALIAS, false)
	}

	if fontFormat := getFontFormat(face); fontFormat != "" {
		pat.AddString(FONTFORMAT, fontFormat)
	}

	return pat, nameMappings, cs, ls
}

func weightFromBFD(value int32) float32 {
	switch (value + 5) / 10 {
	case 1:
		return WIDTH_ULTRACONDENSED
	case 2:
		return WIDTH_EXTRACONDENSED
	case 3:
		return WIDTH_CONDENSED
	case 4:
		return WIDTH_SEMICONDENSED
	case 5:
		return WIDTH_NORMAL
	case 6:
		return WIDTH_SEMIEXPANDED
	case 7:
		return WIDTH_EXPANDED
	case 8:
		return WIDTH_EXTRAEXPANDED
	case 9:
		return WIDTH_ULTRAEXPANDED
	default:
		return -1
	}
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func approximatelyEqual(x, y int) bool { return abs(x-y)*33 <= max(abs(x), abs(y)) }

func getSpacing(face fonts.Face, head *truetype.TableHead, coords []float32) int32 {
	// if face.face_flags&FT_FACE_FLAG_SCALABLE == 0 && len(face.available_sizes) > 0 && head != nil {
	// 	var strikeIndex int
	// 	// Select the face closest to 16 pixels tall
	// 	for i := 1; i < len(face.available_sizes); i++ {
	// 		if abs(int(face.available_sizes[i].Height-16)) < abs(int(face.available_sizes[strikeIndex].Height-16)) {
	// 			strikeIndex = i
	// 		}
	// 	}

	// 	// TODO: this influence the later Get_Advance call
	// 	FT_Select_Size(face, strikeIndex)
	// }

	cmap, enc := face.Cmap()
	if enc != fonts.EncUnicode && enc != fonts.EncSymbol {
		return MONO
	}

	return PROPORTIONAL // TODO
	metrics := face.LoadMetrics()
	advances := make([]int, 0, 3)
	iter := cmap.Iter()
	for iter.Next() && len(advances) < 3 {
		_, glyph := iter.Char()
		advance := int(metrics.HorizontalAdvance(glyph, coords))
		if advance != 0 {
			// add if not already found
			var j int
			for j = 0; j < len(advances); j++ {
				if approximatelyEqual(int(advance), advances[j]) {
					break
				}
			}
			if j == len(advances) {
				advances = append(advances, int(advance))
			}
		}
	}

	if len(advances) <= 1 {
		return MONO
	} else if len(advances) == 2 && approximatelyEqual(min(advances[0], advances[1])*2,
		max(advances[0], advances[1])) {
		return DUAL
	}
	return PROPORTIONAL
}

// also returns the selected encoding
func getCharSet(face fonts.Face) (Charset, fonts.CmapEncoding) {
	var fcs Charset

	cmap, enc := face.Cmap()
	if enc != fonts.EncUnicode && enc != fonts.EncSymbol {
		return fcs, enc
	}

	var (
		leaf *charPage
		page = ^uint16(0)
		off  uint32
	)
	iter := cmap.Iter()
	for iter.Next() {
		ucs4, _ := iter.Char()

		/* CID fonts built by Adobe used to make ASCII control chars to cid1
		 * (space glyph). As such, always check contour for those characters. */
		// if ucs4 <= 0x001F {
		// 	glyphMetric := FT_Load_Glyph(face, glyph, loadFlags)

		// 	if glyphMetric == nil ||
		// 		(glyphMetric.format == FT_GLYPH_FORMAT_OUTLINE && len(glyphMetric.outline.contours) == 0) {
		// 		continue
		// 	}
		// }

		fcs.AddChar(ucs4)
		if pa := uint16(ucs4 >> 8); pa != page {
			page = pa
			leaf = fcs.findLeafCreate(pa)
		}
		off = uint32(ucs4) & 0xff
		leaf[off>>5] |= (1 << (off & 0x1f))
	}
	if enc == fonts.EncSymbol {
		/* For symbol-encoded OpenType fonts, we duplicate the
		 * U+F000..F0FF range at U+0000..U+00FF.  That's what
		 * Windows seems to do, and that's hinted about at:
		 * http://www.microsoft.com/typography/otspec/recom.htm
		 * under "Non-Standard (Symbol) Fonts".
		 *
		 * See thread with subject "Webdings and other MS symbol
		 * fonts don't display" on mailing list from May 2015.
		 */
		for ucs4 := rune(0xF000); ucs4 < 0xF100; ucs4++ {
			if fcs.HasChar(ucs4) {
				fcs.AddChar(ucs4 - 0xF000)
			}
		}
	}
	return fcs, enc
}

// This is a bit generous; the registry has only lower case and space  except for 'DFLT'.
func isValidScript(x byte) bool {
	return (0101 <= x && x <= 0132) || (0141 <= x && x <= 0172) ||
		('0' <= x && x <= '9') || (040 == x)
}

func addtag(complexFeats []byte, tag truetype.Tag) []byte {
	tagString := tag.String()

	/* skip tags which aren't alphanumeric, under the assumption that
	 * they're probably broken  */
	if !isValidScript(tagString[0]) || !isValidScript(tagString[1]) ||
		!isValidScript(tagString[2]) || !isValidScript(tagString[3]) {
		return complexFeats
	}

	if len(complexFeats) != 0 {
		complexFeats = append(complexFeats, ' ')
	}
	complexFeats = append(complexFeats, "otlayout:"+tagString...)
	return complexFeats
}

func fontCapabilities(face *truetype.Font) string {
	var complexFeats []byte

	if isSil, _ := face.IsGraphite(); isSil {
		complexFeats = []byte("ttable:Silf ")
	}

	var gposScripts, gsubScripts []truetype.Script
	if gpos, err := face.GPOSTable(); err == nil {
		gposScripts = gpos.Scripts
	}
	if gsub, err := face.GSUBTable(); err == nil {
		gsubScripts = gsub.Scripts
	}
	gsubCount, gposCount := len(gsubScripts), len(gposScripts)

	for indx1, indx2 := 0, 0; indx1 < gsubCount || indx2 < gposCount; {
		if indx1 == gsubCount {
			complexFeats = addtag(complexFeats, gposScripts[indx2].Tag)
			indx2++
		} else if (indx2 == gposCount) || (gsubScripts[indx1].Tag < gposScripts[indx2].Tag) {
			complexFeats = addtag(complexFeats, gsubScripts[indx1].Tag)
			indx1++
		} else if gsubScripts[indx1].Tag == gposScripts[indx2].Tag {
			complexFeats = addtag(complexFeats, gsubScripts[indx1].Tag)
			indx1++
			indx2++
		} else {
			complexFeats = addtag(complexFeats, gposScripts[indx2].Tag)
			indx2++
		}
	}

	return string(complexFeats)
}

//  FcCharset *
//  getCharSetAndSpacing (FT_Face face, FcBlanks *blanks UNUSED, int *spacing)
//  {

// 	 if (spacing)
// 	 *spacing = getSpacing (face);

// 	 return getCharSet (face, blanks);
//  }

//  #define TTAG_GPOS  FT_MAKE_TAG( 'G', 'P', 'O', 'S' )
//  #define TTAG_GSUB  FT_MAKE_TAG( 'G', 'S', 'U', 'B' )
//  #define TTAG_SILF  FT_MAKE_TAG( 'S', 'i', 'l', 'f')
//  #define TTAG_prep  FT_MAKE_TAG( 'p', 'r', 'e', 'p' )

//  static int
//  compareulong (const void *a, const void *b)
//  {
// 	 const FT_ULong *ua = (const FT_ULong *) a;
// 	 const FT_ULong *ub = (const FT_ULong *) b;
// 	 return *ua - *ub;
//  }

//  static Bool
//  FindTable (FT_Face face, FT_ULong tabletag)
//  {
// 	 FT_Stream  stream = face.stream;
// 	 FT_Error   error;

// 	 if (!stream)
// 		 return false;

// 	 if (( error = ftglue_face_goto_table( face, tabletag, stream ) ))
// 	 return false;

// 	 return true;
//  }

//  /*
//   * Map a UCS4 glyph to a glyph index.  Use all available encoding
//   * tables to try and find one that works.  This information is expected
//   * to be cached by higher levels, so performance isn't critical
//   */

//  FT_UInt
//  FcFreeTypeCharIndex (FT_Face face, FcChar32 ucs4)
//  {
// 	 int		    initial, offset, decode;
// 	 FT_UInt	    glyphindex;

// 	 initial = 0;

// 	 if (!face)
// 		 return 0;

// 	 /*
// 	  * Find the current encoding
// 	  */
// 	 if (face.charmap)
// 	 {
// 	 for (; initial < NUM_DECODE; initial++)
// 		 if (fcFontEncodings[initial] == face.charmap.encoding)
// 		 break;
// 	 if (initial == NUM_DECODE)
// 		 initial = 0;
// 	 }
// 	 /*
// 	  * Check each encoding for the glyph, starting with the current one
// 	  */
// 	 for (offset = 0; offset < NUM_DECODE; offset++)
// 	 {
// 	 decode = (initial + offset) % NUM_DECODE;
// 	 if (!face.charmap || face.charmap.encoding != fcFontEncodings[decode])
// 		 if (FT_Select_Charmap (face, fcFontEncodings[decode]) != 0)
// 		 continue;
// 	 glyphindex = FT_Get_Char_Index (face, (FT_ULong) ucs4);
// 	 if (glyphindex)
// 		 return glyphindex;
// 	 if (ucs4 < 0x100 && face.charmap &&
// 		 face.charmap.encoding == FT_ENCODING_MS_SYMBOL)
// 	 {
// 		 /* For symbol-encoded OpenType fonts, we duplicate the
// 		  * U+F000..F0FF range at U+0000..U+00FF.  That's what
// 		  * Windows seems to do, and that's hinted about at:
// 		  * http://www.microsoft.com/typography/otspec/recom.htm
// 		  * under "Non-Standard (Symbol) Fonts".
// 		  *
// 		  * See thread with subject "Webdings and other MS symbol
// 		  * fonts don't display" on mailing list from May 2015.
// 		  */
// 		 glyphindex = FT_Get_Char_Index (face, (FT_ULong) ucs4 + 0xF000);
// 		 if (glyphindex)
// 		 return glyphindex;
// 	 }
// 	 }
// 	 return 0;
//  }

//  Pattern *
//  FcFreeTypeQueryFace (const FT_Face  face,
// 			  const FcChar8  *file,
// 			  unsigned int   id,
// 			  FcBlanks	    *blanks UNUSED)
//  {
// 	 return queryFace (face, file, id, nil, nil, nil);
//  }

//  Pattern *
//  FcFreeTypeQuery(const FcChar8	*file,
// 		 unsigned int	id,
// 		 FcBlanks	*blanks UNUSED,
// 		 int		*count)
//  {
// 	 FT_Face	    face;
// 	 FT_Library	    ftLibrary;
// 	 Pattern	    *pat = nil;

// 	 if (FT_Init_FreeType (&ftLibrary))
// 	 return nil;

// 	 if (FT_New_Face (ftLibrary, (char *) file, id & 0x7FFFFFFF, &face))
// 	 goto bail;

// 	 if (count)
// 	   *count = face.num_faces;

// 	 pat = queryFace (face, file, id, nil, nil, nil);

// 	 FT_Done_Face (face);
//  bail:
// 	 FT_Done_FreeType (ftLibrary);
// 	 return pat;
//  }

//  unsigned int
//  scanFontRessource(const FcChar8	*file,
// 			unsigned int		id,
// 			FcBlanks		*blanks,
// 			int			*count,
// 			Fontset            *set)
//  {
// 	 FT_Face face = nil;
// 	 FT_Library ftLibrary = nil;
// 	 FcCharset *cs = nil;
// 	 langSet *ls = nil;
// 	 nameMapping  *nm = nil;
// 	 FT_MM_Var *mm_var = nil;
// 	 Bool index_set = id != (unsigned int) -1;
// 	 unsigned int set_face_num = index_set ? id & 0xFFFF : 0;
// 	 unsigned int set_instance_num = index_set ? id >> 16 : 0;
// 	 unsigned int face_num = set_face_num;
// 	 unsigned int instance_num = set_instance_num;
// 	 unsigned int num_faces = 0;
// 	 unsigned int num_instances = 0;
// 	 unsigned int ret = 0;
// 	 int err = 0;

// 	 if (count)
// 	 *count = 0;

// 	 if (FT_Init_FreeType (&ftLibrary))
// 	 return 0;

// 	 if (FT_New_Face (ftLibrary, (const char *) file, face_num, &face))
// 	 goto bail;

// 	 num_faces = face.num_faces;
// 	 num_instances = face.style_flags >> 16;
// 	 if (num_instances && (!index_set || instance_num))
// 	 {
// 	 FT_Get_MM_Var (face, &mm_var);
// 	 if (!mm_var)
// 	   num_instances = 0;
// 	 }

// 	 if (count)
// 	   *count = num_faces;

// 	 do {
// 	 Pattern *pat = nil;

// 	 if (instance_num == 0x8000 || instance_num > num_instances)
// 		 FT_Set_Var_Design_Coordinates (face, 0, nil); /* Reset variations. */
// 	 else if (instance_num)
// 	 {
// 		 FT_Var_Named_Style *instance = &mm_var.namedstyle[instance_num - 1];
// 		 FT_Fixed *coords = instance.coords;
// 		 Bool nonzero;
// 		 unsigned int i;

// 		 /* Skip named-instance that coincides with base instance. */
// 		 nonzero = false;
// 		 for (i = 0; i < mm_var.num_axis; i++)
// 		 if (coords[i] != mm_var.axis[i].def)
// 		 {
// 			 nonzero = true;
// 			 break;
// 		 }
// 		 if (!nonzero)
// 		 goto skip;

// 		 FT_Set_Var_Design_Coordinates (face, mm_var.num_axis, coords);
// 	 }

// 	 id = ((instance_num << 16) + face_num);
// 	 pat = queryFace (face, (const FcChar8 *) file, id, &cs, &ls, &nm);

// 	 if (pat)
// 	 {

// 		 ret++;
// 		 if (!set || ! FontsetAdd (set, pat))
// 		 PatternDestroy (pat);
// 	 }
// 	 else if (instance_num != 0x8000)
// 		 err = 1;

//  skip:
// 	 if (!index_set && instance_num < num_instances)
// 		 instance_num++;
// 	 else if (!index_set && instance_num == num_instances)
// 		 instance_num = 0x8000; /* variable font */
// 	 else
// 	 {
// 		 free (nm);
// 		 nm = nil;
// 		 langSetDestroy (ls);
// 		 ls = nil;
// 		 FcCharsetDestroy (cs);
// 		 cs = nil;
// 		 FT_Done_Face (face);
// 		 face = nil;

// 		 face_num++;
// 		 instance_num = set_instance_num;

// 		 if (FT_New_Face (ftLibrary, (const char *) file, face_num, &face))
// 		   break;
// 	 }
// 	 } while (!err && (!index_set || face_num == set_face_num) && face_num < num_faces);

//  bail:
//  #ifdef HAVE_FT_DONE_MM_VAR
// 	 FT_Done_MM_Var (ftLibrary, mm_var);
//  #else
// 	 free (mm_var);
//  #endif
// 	 langSetDestroy (ls);
// 	 FcCharsetDestroy (cs);
// 	 if (face)
// 	 FT_Done_Face (face);
// 	 FT_Done_FreeType (ftLibrary);
// 	 if (nm)
// 	 free (nm);

// 	 return ret;
//  }

// func FcFileScanConfig(set *Fontset, dirs strSet, file string, config *Config) bool {
// 	if isDir(file) {
// 		sysroot := config.getSysRoot()
// 		d := file
// 		if sysroot != "" {
// 			if strings.HasPrefix(file, sysroot) {
// 				d = filepath.Clean(strings.TrimPrefix(file, sysroot))
// 			}
// 		}
// 		dirs[d] = true
// 		return true
// 	}

// 	return scanFontConfig(set, file, config)
// }
