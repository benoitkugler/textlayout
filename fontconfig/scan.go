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
	format FontFormat
}{
	{truetype.Load, "TrueType"},
	{bitmap.Load, "PCF"},
	{type1.Load, "Type 1"},
}

// FontFormat identifies the supported font file types.
type FontFormat string

const (
	TrueType FontFormat = "TrueType"
	PCF      FontFormat = "PCF"
	Type1    FontFormat = "Type 1"
)

// Loader returns the loader for the font format.
func (ff FontFormat) Loader() fonts.FontLoader {
	switch ff {
	case "TrueType":
		return truetype.Load
	case "PCF":
		return bitmap.Load
	case "Type 1":
		return type1.Load
	default:
		return nil
	}
}

// constructs patterns found in 'file'. `fileID` is included
// in the returned patterns, as well as the index of each face.
// invalid files are simply ignored, returning an empty font set
func scanOneFontFile(file fonts.Resource, fileID string, config *Config) Fontset {
	if debugMode {
		fmt.Printf("Scanning file %s...\n", file)
	}

	faces, format := ReadFontFile(file)
	if format == "" {
		return nil
	}

	var set Fontset

	for faceNum, face := range faces {
		var pat Pattern

		// basic face
		pat, shared := queryFace(face, fileID, uint32(faceNum), nil)
		if pat != nil {
			set = append(set, pat)
		}

		// optional variable instances
		if variable, ok := face.(truetype.FaceVariable); ok {
			vars := variable.Variations()
			for instNum, instance := range vars.Instances {
				// skip named-instance that coincides with base instance.
				if vars.IsDefaultInstance(instance) {
					continue
				}
				// for variable fonts, the id contains both
				// the face number (16 lower bits) and the instance number
				// (16 higher bits, starting at 1)
				id := uint32(faceNum) | uint32(instNum+1)<<16
				pat, shared = queryFace(face, fileID, id, shared)
				if pat != nil {
					set = append(set, pat)
				}
			}
		}
	}

	for _, font := range set {
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

// ReadFontFile tries for every supported font format, returning a valid font format if one matches.
func ReadFontFile(file fonts.Resource) (fonts.Faces, FontFormat) {
	for _, loader := range loaders {
		out, err := loader.loader(file)
		if err == nil {
			return out, loader.format
		}
	}
	return nil, ""
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
	lang       string
	PlatformID truetype.PlatformID
	LanguageID truetype.PlatformLanguageID
}{
	{"", truetype.PlatformUnicode, lang_DONT_CARE},
	{"en", truetype.PlatformMac, macLangID_ENGLISH},
	{"fr", truetype.PlatformMac, macLangID_FRENCH},
	{"de", truetype.PlatformMac, macLangID_GERMAN},
	{"it", truetype.PlatformMac, macLangID_ITALIAN},
	{"nl", truetype.PlatformMac, macLangID_DUTCH},
	{"sv", truetype.PlatformMac, macLangID_SWEDISH},
	{"es", truetype.PlatformMac, macLangID_SPANISH},
	{"da", truetype.PlatformMac, macLangID_DANISH},
	{"pt", truetype.PlatformMac, macLangID_PORTUGUESE},
	{"no", truetype.PlatformMac, macLangID_NORWEGIAN},
	{"he", truetype.PlatformMac, macLangID_HEBREW},
	{"ja", truetype.PlatformMac, macLangID_JAPANESE},
	{"ar", truetype.PlatformMac, macLangID_ARABIC},
	{"fi", truetype.PlatformMac, macLangID_FINNISH},
	{"el", truetype.PlatformMac, macLangID_GREEK},
	{"is", truetype.PlatformMac, macLangID_ICELANDIC},
	{"mt", truetype.PlatformMac, macLangID_MALTESE},
	{"tr", truetype.PlatformMac, macLangID_TURKISH},
	{"hr", truetype.PlatformMac, macLangID_CROATIAN},
	{"zh-tw", truetype.PlatformMac, macLangID_CHINESE_TRADITIONAL},
	{"ur", truetype.PlatformMac, macLangID_URDU},
	{"hi", truetype.PlatformMac, macLangID_HINDI},
	{"th", truetype.PlatformMac, macLangID_THAI},
	{"ko", truetype.PlatformMac, macLangID_KOREAN},
	{"lt", truetype.PlatformMac, macLangID_LITHUANIAN},
	{"pl", truetype.PlatformMac, macLangID_POLISH},
	{"hu", truetype.PlatformMac, macLangID_HUNGARIAN},
	{"et", truetype.PlatformMac, macLangID_ESTONIAN},
	{"lv", truetype.PlatformMac, macLangID_LETTISH},

	{"fo", truetype.PlatformMac, macLangID_FAEROESE},
	{"fa", truetype.PlatformMac, macLangID_FARSI},
	{"ru", truetype.PlatformMac, macLangID_RUSSIAN},
	{"zh-cn", truetype.PlatformMac, macLangID_CHINESE_SIMPLIFIED},
	{"nl", truetype.PlatformMac, macLangID_FLEMISH},
	{"ga", truetype.PlatformMac, macLangID_IRISH},
	{"sq", truetype.PlatformMac, macLangID_ALBANIAN},
	{"ro", truetype.PlatformMac, macLangID_ROMANIAN},
	{"cs", truetype.PlatformMac, macLangID_CZECH},
	{"sk", truetype.PlatformMac, macLangID_SLOVAK},
	{"sl", truetype.PlatformMac, macLangID_SLOVENIAN},
	{"yi", truetype.PlatformMac, macLangID_YIDDISH},
	{"sr", truetype.PlatformMac, macLangID_SERBIAN},
	{"mk", truetype.PlatformMac, macLangID_MACEDONIAN},
	{"bg", truetype.PlatformMac, macLangID_BULGARIAN},
	{"uk", truetype.PlatformMac, macLangID_UKRAINIAN},
	{"be", truetype.PlatformMac, macLangID_BYELORUSSIAN},
	{"uz", truetype.PlatformMac, macLangID_UZBEK},
	{"kk", truetype.PlatformMac, macLangID_KAZAKH},
	{"az", truetype.PlatformMac, macLangID_AZERBAIJANI_CYRILLIC_SCRIPT},
	{"ar", truetype.PlatformMac, macLangID_AZERBAIJANI_ARABIC_SCRIPT},
	{"hy", truetype.PlatformMac, macLangID_ARMENIAN},
	{"ka", truetype.PlatformMac, macLangID_GEORGIAN},
	{"mo", truetype.PlatformMac, macLangID_MOLDAVIAN},
	{"ky", truetype.PlatformMac, macLangID_KIRGHIZ},
	{"tg", truetype.PlatformMac, macLangID_TAJIKI},
	{"tk", truetype.PlatformMac, macLangID_TURKMEN},
	{"mn", truetype.PlatformMac, macLangID_MONGOLIAN},
	{"mn", truetype.PlatformMac, macLangID_MONGOLIAN_CYRILLIC_SCRIPT},
	{"ps", truetype.PlatformMac, macLangID_PASHTO},
	{"ku", truetype.PlatformMac, macLangID_KURDISH},
	{"ks", truetype.PlatformMac, macLangID_KASHMIRI},
	{"sd", truetype.PlatformMac, macLangID_SINDHI},
	{"bo", truetype.PlatformMac, macLangID_TIBETAN},
	{"ne", truetype.PlatformMac, macLangID_NEPALI},
	{"sa", truetype.PlatformMac, macLangID_SANSKRIT},
	{"mr", truetype.PlatformMac, macLangID_MARATHI},
	{"bn", truetype.PlatformMac, macLangID_BENGALI},
	{"as", truetype.PlatformMac, macLangID_ASSAMESE},
	{"gu", truetype.PlatformMac, macLangID_GUJARATI},
	{"pa", truetype.PlatformMac, macLangID_PUNJABI},
	{"or", truetype.PlatformMac, macLangID_ORIYA},
	{"ml", truetype.PlatformMac, macLangID_MALAYALAM},
	{"kn", truetype.PlatformMac, macLangID_KANNADA},
	{"ta", truetype.PlatformMac, macLangID_TAMIL},
	{"te", truetype.PlatformMac, macLangID_TELUGU},
	{"si", truetype.PlatformMac, macLangID_SINHALESE},
	{"my", truetype.PlatformMac, macLangID_BURMESE},
	{"km", truetype.PlatformMac, macLangID_KHMER},
	{"lo", truetype.PlatformMac, macLangID_LAO},
	{"vi", truetype.PlatformMac, macLangID_VIETNAMESE},
	{"id", truetype.PlatformMac, macLangID_INDONESIAN},
	{"tl", truetype.PlatformMac, macLangID_TAGALOG},
	{"ms", truetype.PlatformMac, macLangID_MALAY_ROMAN_SCRIPT},
	{"ms", truetype.PlatformMac, macLangID_MALAY_ARABIC_SCRIPT},
	{"am", truetype.PlatformMac, macLangID_AMHARIC},
	{"ti", truetype.PlatformMac, macLangID_TIGRINYA},
	{"om", truetype.PlatformMac, macLangID_GALLA},
	{"so", truetype.PlatformMac, macLangID_SOMALI},
	{"sw", truetype.PlatformMac, macLangID_SWAHILI},
	{"rw", truetype.PlatformMac, macLangID_RUANDA},
	{"rn", truetype.PlatformMac, macLangID_RUNDI},
	{"ny", truetype.PlatformMac, macLangID_CHEWA},
	{"mg", truetype.PlatformMac, macLangID_MALAGASY},
	{"eo", truetype.PlatformMac, macLangID_ESPERANTO},
	{"cy", truetype.PlatformMac, macLangID_WELSH},
	{"eu", truetype.PlatformMac, macLangID_BASQUE},
	{"ca", truetype.PlatformMac, macLangID_CATALAN},
	{"la", truetype.PlatformMac, macLangID_LATIN},
	{"qu", truetype.PlatformMac, macLangID_QUECHUA},
	{"gn", truetype.PlatformMac, macLangID_GUARANI},
	{"ay", truetype.PlatformMac, macLangID_AYMARA},
	{"tt", truetype.PlatformMac, macLangID_TATAR},
	{"ug", truetype.PlatformMac, macLangID_UIGHUR},
	{"dz", truetype.PlatformMac, macLangID_DZONGKHA},
	{"jw", truetype.PlatformMac, macLangID_JAVANESE},
	{"su", truetype.PlatformMac, macLangID_SUNDANESE},

	/* The following codes are new as of 2000-03-10 */
	{"gl", truetype.PlatformMac, macLangID_GALICIAN},
	{"af", truetype.PlatformMac, macLangID_AFRIKAANS},
	{"br", truetype.PlatformMac, macLangID_BRETON},
	{"iu", truetype.PlatformMac, macLangID_INUKTITUT},
	{"gd", truetype.PlatformMac, macLangID_SCOTTISH_GAELIC},
	{"gv", truetype.PlatformMac, macLangID_MANX_GAELIC},
	{"ga", truetype.PlatformMac, macLangID_IRISH_GAELIC},
	{"to", truetype.PlatformMac, macLangID_TONGAN},
	{"el", truetype.PlatformMac, macLangID_GREEK_POLYTONIC},
	{"ik", truetype.PlatformMac, macLangID_GREELANDIC},
	{"az", truetype.PlatformMac, macLangID_AZERBAIJANI_ROMAN_SCRIPT},

	{"ar", truetype.PlatformMicrosoft, msLangID_ARABIC_SAUDI_ARABIA},
	{"ar", truetype.PlatformMicrosoft, msLangID_ARABIC_IRAQ},
	{"ar", truetype.PlatformMicrosoft, msLangID_ARABIC_EGYPT},
	{"ar", truetype.PlatformMicrosoft, msLangID_ARABIC_LIBYA},
	{"ar", truetype.PlatformMicrosoft, msLangID_ARABIC_ALGERIA},
	{"ar", truetype.PlatformMicrosoft, msLangID_ARABIC_MOROCCO},
	{"ar", truetype.PlatformMicrosoft, msLangID_ARABIC_TUNISIA},
	{"ar", truetype.PlatformMicrosoft, msLangID_ARABIC_OMAN},
	{"ar", truetype.PlatformMicrosoft, msLangID_ARABIC_YEMEN},
	{"ar", truetype.PlatformMicrosoft, msLangID_ARABIC_SYRIA},
	{"ar", truetype.PlatformMicrosoft, msLangID_ARABIC_JORDAN},
	{"ar", truetype.PlatformMicrosoft, msLangID_ARABIC_LEBANON},
	{"ar", truetype.PlatformMicrosoft, msLangID_ARABIC_KUWAIT},
	{"ar", truetype.PlatformMicrosoft, msLangID_ARABIC_UAE},
	{"ar", truetype.PlatformMicrosoft, msLangID_ARABIC_BAHRAIN},
	{"ar", truetype.PlatformMicrosoft, msLangID_ARABIC_QATAR},
	{"bg", truetype.PlatformMicrosoft, msLangID_BULGARIAN_BULGARIA},
	{"ca", truetype.PlatformMicrosoft, msLangID_CATALAN_CATALAN},
	{"zh-tw", truetype.PlatformMicrosoft, msLangID_CHINESE_TAIWAN},
	{"zh-cn", truetype.PlatformMicrosoft, msLangID_CHINESE_PRC},
	{"zh-hk", truetype.PlatformMicrosoft, msLangID_CHINESE_HONG_KONG},
	{"zh-sg", truetype.PlatformMicrosoft, msLangID_CHINESE_SINGAPORE},

	{"zh-mo", truetype.PlatformMicrosoft, msLangID_CHINESE_MACAO},

	{"cs", truetype.PlatformMicrosoft, msLangID_CZECH_CZECH_REPUBLIC},
	{"da", truetype.PlatformMicrosoft, msLangID_DANISH_DENMARK},
	{"de", truetype.PlatformMicrosoft, msLangID_GERMAN_GERMANY},
	{"de", truetype.PlatformMicrosoft, msLangID_GERMAN_SWITZERLAND},
	{"de", truetype.PlatformMicrosoft, msLangID_GERMAN_AUSTRIA},
	{"de", truetype.PlatformMicrosoft, msLangID_GERMAN_LUXEMBOURG},
	{"de", truetype.PlatformMicrosoft, msLangID_GERMAN_LIECHTENSTEIN},
	{"el", truetype.PlatformMicrosoft, msLangID_GREEK_GREECE},
	{"en", truetype.PlatformMicrosoft, msLangID_ENGLISH_UNITED_STATES},
	{"en", truetype.PlatformMicrosoft, msLangID_ENGLISH_UNITED_KINGDOM},
	{"en", truetype.PlatformMicrosoft, msLangID_ENGLISH_AUSTRALIA},
	{"en", truetype.PlatformMicrosoft, msLangID_ENGLISH_CANADA},
	{"en", truetype.PlatformMicrosoft, msLangID_ENGLISH_NEW_ZEALAND},
	{"en", truetype.PlatformMicrosoft, msLangID_ENGLISH_IRELAND},
	{"en", truetype.PlatformMicrosoft, msLangID_ENGLISH_SOUTH_AFRICA},
	{"en", truetype.PlatformMicrosoft, msLangID_ENGLISH_JAMAICA},
	{"en", truetype.PlatformMicrosoft, msLangID_ENGLISH_CARIBBEAN},
	{"en", truetype.PlatformMicrosoft, msLangID_ENGLISH_BELIZE},
	{"en", truetype.PlatformMicrosoft, msLangID_ENGLISH_TRINIDAD},
	{"en", truetype.PlatformMicrosoft, msLangID_ENGLISH_ZIMBABWE},
	{"en", truetype.PlatformMicrosoft, msLangID_ENGLISH_PHILIPPINES},
	{"es", truetype.PlatformMicrosoft, msLangID_SPANISH_SPAIN_TRADITIONAL_SORT},
	{"es", truetype.PlatformMicrosoft, msLangID_SPANISH_MEXICO},
	{"es", truetype.PlatformMicrosoft, msLangID_SPANISH_SPAIN_MODERN_SORT},
	{"es", truetype.PlatformMicrosoft, msLangID_SPANISH_GUATEMALA},
	{"es", truetype.PlatformMicrosoft, msLangID_SPANISH_COSTA_RICA},
	{"es", truetype.PlatformMicrosoft, msLangID_SPANISH_PANAMA},
	{"es", truetype.PlatformMicrosoft, msLangID_SPANISH_DOMINICAN_REPUBLIC},
	{"es", truetype.PlatformMicrosoft, msLangID_SPANISH_VENEZUELA},
	{"es", truetype.PlatformMicrosoft, msLangID_SPANISH_COLOMBIA},
	{"es", truetype.PlatformMicrosoft, msLangID_SPANISH_PERU},
	{"es", truetype.PlatformMicrosoft, msLangID_SPANISH_ARGENTINA},
	{"es", truetype.PlatformMicrosoft, msLangID_SPANISH_ECUADOR},
	{"es", truetype.PlatformMicrosoft, msLangID_SPANISH_CHILE},
	{"es", truetype.PlatformMicrosoft, msLangID_SPANISH_URUGUAY},
	{"es", truetype.PlatformMicrosoft, msLangID_SPANISH_PARAGUAY},
	{"es", truetype.PlatformMicrosoft, msLangID_SPANISH_BOLIVIA},
	{"es", truetype.PlatformMicrosoft, msLangID_SPANISH_EL_SALVADOR},
	{"es", truetype.PlatformMicrosoft, msLangID_SPANISH_HONDURAS},
	{"es", truetype.PlatformMicrosoft, msLangID_SPANISH_NICARAGUA},
	{"es", truetype.PlatformMicrosoft, msLangID_SPANISH_PUERTO_RICO},
	{"fi", truetype.PlatformMicrosoft, msLangID_FINNISH_FINLAND},
	{"fr", truetype.PlatformMicrosoft, msLangID_FRENCH_FRANCE},
	{"fr", truetype.PlatformMicrosoft, msLangID_FRENCH_BELGIUM},
	{"fr", truetype.PlatformMicrosoft, msLangID_FRENCH_CANADA},
	{"fr", truetype.PlatformMicrosoft, msLangID_FRENCH_SWITZERLAND},
	{"fr", truetype.PlatformMicrosoft, msLangID_FRENCH_LUXEMBOURG},
	{"fr", truetype.PlatformMicrosoft, msLangID_FRENCH_MONACO},
	{"he", truetype.PlatformMicrosoft, msLangID_HEBREW_ISRAEL},
	{"hu", truetype.PlatformMicrosoft, msLangID_HUNGARIAN_HUNGARY},
	{"is", truetype.PlatformMicrosoft, msLangID_ICELANDIC_ICELAND},
	{"it", truetype.PlatformMicrosoft, msLangID_ITALIAN_ITALY},
	{"it", truetype.PlatformMicrosoft, msLangID_ITALIAN_SWITZERLAND},
	{"ja", truetype.PlatformMicrosoft, msLangID_JAPANESE_JAPAN},
	{"ko", truetype.PlatformMicrosoft, msLangID_KOREAN_KOREA},
	{"ko", truetype.PlatformMicrosoft, msLangID_KOREAN_JOHAB_KOREA},
	{"nl", truetype.PlatformMicrosoft, msLangID_DUTCH_NETHERLANDS},
	{"nl", truetype.PlatformMicrosoft, msLangID_DUTCH_BELGIUM},
	{"no", truetype.PlatformMicrosoft, msLangID_NORWEGIAN_NORWAY_BOKMAL},
	{"nn", truetype.PlatformMicrosoft, msLangID_NORWEGIAN_NORWAY_NYNORSK},
	{"pl", truetype.PlatformMicrosoft, msLangID_POLISH_POLAND},
	{"pt", truetype.PlatformMicrosoft, msLangID_PORTUGUESE_BRAZIL},
	{"pt", truetype.PlatformMicrosoft, msLangID_PORTUGUESE_PORTUGAL},
	{"rm", truetype.PlatformMicrosoft, msLangID_ROMANSH_SWITZERLAND},
	{"ro", truetype.PlatformMicrosoft, msLangID_ROMANIAN_ROMANIA},
	{"mo", truetype.PlatformMicrosoft, msLangID_MOLDAVIAN_MOLDAVIA},
	{"ru", truetype.PlatformMicrosoft, msLangID_RUSSIAN_RUSSIA},
	{"ru", truetype.PlatformMicrosoft, msLangID_RUSSIAN_MOLDAVIA},
	{"hr", truetype.PlatformMicrosoft, msLangID_CROATIAN_CROATIA},
	{"sr", truetype.PlatformMicrosoft, msLangID_SERBIAN_SERBIA_LATIN},
	{"sr", truetype.PlatformMicrosoft, msLangID_SERBIAN_SERBIA_CYRILLIC},
	{"sk", truetype.PlatformMicrosoft, msLangID_SLOVAK_SLOVAKIA},
	{"sq", truetype.PlatformMicrosoft, msLangID_ALBANIAN_ALBANIA},
	{"sv", truetype.PlatformMicrosoft, msLangID_SWEDISH_SWEDEN},
	{"sv", truetype.PlatformMicrosoft, msLangID_SWEDISH_FINLAND},
	{"th", truetype.PlatformMicrosoft, msLangID_THAI_THAILAND},
	{"tr", truetype.PlatformMicrosoft, msLangID_TURKISH_TURKEY},
	{"ur", truetype.PlatformMicrosoft, msLangID_URDU_PAKISTAN},
	{"id", truetype.PlatformMicrosoft, msLangID_INDONESIAN_INDONESIA},
	{"uk", truetype.PlatformMicrosoft, msLangID_UKRAINIAN_UKRAINE},
	{"be", truetype.PlatformMicrosoft, msLangID_BELARUSIAN_BELARUS},
	{"sl", truetype.PlatformMicrosoft, msLangID_SLOVENIAN_SLOVENIA},
	{"et", truetype.PlatformMicrosoft, msLangID_ESTONIAN_ESTONIA},
	{"lv", truetype.PlatformMicrosoft, msLangID_LATVIAN_LATVIA},
	{"lt", truetype.PlatformMicrosoft, msLangID_LITHUANIAN_LITHUANIA},
	{"lt", truetype.PlatformMicrosoft, msLangID_CLASSIC_LITHUANIAN_LITHUANIA},
	{"mi", truetype.PlatformMicrosoft, msLangID_MAORI_NEW_ZEALAND},
	{"fa", truetype.PlatformMicrosoft, msLangID_FARSI_IRAN},
	{"vi", truetype.PlatformMicrosoft, msLangID_VIETNAMESE_VIET_NAM},
	{"hy", truetype.PlatformMicrosoft, msLangID_ARMENIAN_ARMENIA},
	{"az", truetype.PlatformMicrosoft, msLangID_AZERI_AZERBAIJAN_LATIN},
	{"az", truetype.PlatformMicrosoft, msLangID_AZERI_AZERBAIJAN_CYRILLIC},
	{"eu", truetype.PlatformMicrosoft, msLangID_BASQUE_BASQUE},
	{"wen", truetype.PlatformMicrosoft, msLangID_UPPER_SORBIAN_GERMANY},
	{"mk", truetype.PlatformMicrosoft, msLangID_MACEDONIAN_MACEDONIA},
	{"st", truetype.PlatformMicrosoft, msLangID_SUTU_SOUTH_AFRICA},
	{"ts", truetype.PlatformMicrosoft, msLangID_TSONGA_SOUTH_AFRICA},
	{"tn", truetype.PlatformMicrosoft, msLangID_SETSWANA_SOUTH_AFRICA},
	{"ven", truetype.PlatformMicrosoft, msLangID_VENDA_SOUTH_AFRICA},
	{"xh", truetype.PlatformMicrosoft, msLangID_ISIXHOSA_SOUTH_AFRICA},
	{"zu", truetype.PlatformMicrosoft, msLangID_ISIZULU_SOUTH_AFRICA},
	{"af", truetype.PlatformMicrosoft, msLangID_AFRIKAANS_SOUTH_AFRICA},
	{"ka", truetype.PlatformMicrosoft, msLangID_GEORGIAN_GEORGIA},
	{"fo", truetype.PlatformMicrosoft, msLangID_FAEROESE_FAEROE_ISLANDS},
	{"hi", truetype.PlatformMicrosoft, msLangID_HINDI_INDIA},
	{"mt", truetype.PlatformMicrosoft, msLangID_MALTESE_MALTA},
	{"se", truetype.PlatformMicrosoft, msLangID_SAAMI_LAPONIA},

	{"gd", truetype.PlatformMicrosoft, msLangID_SCOTTISH_GAELIC_UNITED_KINGDOM},
	{"ga", truetype.PlatformMicrosoft, msLangID_IRISH_GAELIC_IRELAND},

	{"ms", truetype.PlatformMicrosoft, msLangID_MALAY_MALAYSIA},
	{"ms", truetype.PlatformMicrosoft, msLangID_MALAY_BRUNEI_DARUSSALAM},
	{"kk", truetype.PlatformMicrosoft, msLangID_KAZAKH_KAZAKHSTAN},
	{"sw", truetype.PlatformMicrosoft, msLangID_KISWAHILI_KENYA},
	{"uz", truetype.PlatformMicrosoft, msLangID_UZBEK_UZBEKISTAN_LATIN},
	{"uz", truetype.PlatformMicrosoft, msLangID_UZBEK_UZBEKISTAN_CYRILLIC},
	{"tt", truetype.PlatformMicrosoft, msLangID_TATAR_RUSSIA},
	{"bn", truetype.PlatformMicrosoft, msLangID_BENGALI_INDIA},
	{"pa", truetype.PlatformMicrosoft, msLangID_PUNJABI_INDIA},
	{"gu", truetype.PlatformMicrosoft, msLangID_GUJARATI_INDIA},
	{"or", truetype.PlatformMicrosoft, msLangID_ODIA_INDIA},
	{"ta", truetype.PlatformMicrosoft, msLangID_TAMIL_INDIA},
	{"te", truetype.PlatformMicrosoft, msLangID_TELUGU_INDIA},
	{"kn", truetype.PlatformMicrosoft, msLangID_KANNADA_INDIA},
	{"ml", truetype.PlatformMicrosoft, msLangID_MALAYALAM_INDIA},
	{"as", truetype.PlatformMicrosoft, msLangID_ASSAMESE_INDIA},
	{"mr", truetype.PlatformMicrosoft, msLangID_MARATHI_INDIA},
	{"sa", truetype.PlatformMicrosoft, msLangID_SANSKRIT_INDIA},
	{"kok", truetype.PlatformMicrosoft, msLangID_KONKANI_INDIA},

	/* new as of 2001-01-01 */
	{"ar", truetype.PlatformMicrosoft, msLangID_ARABIC_GENERAL},
	{"zh", truetype.PlatformMicrosoft, msLangID_CHINESE_GENERAL},
	{"en", truetype.PlatformMicrosoft, msLangID_ENGLISH_GENERAL},
	{"fr", truetype.PlatformMicrosoft, msLangID_FRENCH_WEST_INDIES},
	{"fr", truetype.PlatformMicrosoft, msLangID_FRENCH_REUNION},
	{"fr", truetype.PlatformMicrosoft, msLangID_FRENCH_CONGO},

	{"fr", truetype.PlatformMicrosoft, msLangID_FRENCH_SENEGAL},
	{"fr", truetype.PlatformMicrosoft, msLangID_FRENCH_CAMEROON},
	{"fr", truetype.PlatformMicrosoft, msLangID_FRENCH_COTE_D_IVOIRE},
	{"fr", truetype.PlatformMicrosoft, msLangID_FRENCH_MALI},
	{"bs", truetype.PlatformMicrosoft, msLangID_BOSNIAN_BOSNIA_HERZEGOVINA},
	{"ur", truetype.PlatformMicrosoft, msLangID_URDU_INDIA},
	{"tg", truetype.PlatformMicrosoft, msLangID_TAJIK_TAJIKISTAN},
	{"yi", truetype.PlatformMicrosoft, msLangID_YIDDISH_GERMANY},
	{"ky", truetype.PlatformMicrosoft, msLangID_KYRGYZ_KYRGYZSTAN},

	{"tk", truetype.PlatformMicrosoft, msLangID_TURKMEN_TURKMENISTAN},
	{"mn", truetype.PlatformMicrosoft, msLangID_MONGOLIAN_MONGOLIA},

	// the following seems to be inconsistent;   here is the current "official" way:
	{"bo", truetype.PlatformMicrosoft, msLangID_DZONGHKA_BHUTAN},
	/* and here is what is used by Passport SDK */
	{"bo", truetype.PlatformMicrosoft, msLangID_TIBETAN_PRC},
	{"dz", truetype.PlatformMicrosoft, msLangID_DZONGHKA_BHUTAN},
	/* end of inconsistency */

	{"cy", truetype.PlatformMicrosoft, msLangID_WELSH_UNITED_KINGDOM},
	{"km", truetype.PlatformMicrosoft, msLangID_KHMER_CAMBODIA},
	{"lo", truetype.PlatformMicrosoft, msLangID_LAO_LAOS},
	{"my", truetype.PlatformMicrosoft, msLangID_BURMESE_MYANMAR},
	{"gl", truetype.PlatformMicrosoft, msLangID_GALICIAN_GALICIAN},
	{"mni", truetype.PlatformMicrosoft, msLangID_MANIPURI_INDIA},
	{"sd", truetype.PlatformMicrosoft, msLangID_SINDHI_INDIA},
	// the following one is only encountered in Microsoft RTF specification
	{"ks", truetype.PlatformMicrosoft, msLangID_KASHMIRI_PAKISTAN},
	// the following one is not in the Passport list, looks like an omission
	{"ks", truetype.PlatformMicrosoft, msLangID_KASHMIRI_SASIA},
	{"ne", truetype.PlatformMicrosoft, msLangID_NEPALI_NEPAL},
	// "ne",{truetype.PlatformMicrosoft, msLangID_NEPALI_INDIA,},
	{"fy", truetype.PlatformMicrosoft, msLangID_FRISIAN_NETHERLANDS},

	// new as of 2001-03-01 (from Office Xp)
	{"en", truetype.PlatformMicrosoft, msLangID_ENGLISH_HONG_KONG},
	{"en", truetype.PlatformMicrosoft, msLangID_ENGLISH_INDIA},
	{"en", truetype.PlatformMicrosoft, msLangID_ENGLISH_MALAYSIA},
	{"en", truetype.PlatformMicrosoft, msLangID_ENGLISH_SINGAPORE},
	{"syr", truetype.PlatformMicrosoft, msLangID_SYRIAC_SYRIA},
	{"si", truetype.PlatformMicrosoft, msLangID_SINHALA_SRI_LANKA},
	{"chr", truetype.PlatformMicrosoft, msLangID_CHEROKEE_UNITED_STATES},
	{"iu", truetype.PlatformMicrosoft, msLangID_INUKTITUT_CANADA},
	{"am", truetype.PlatformMicrosoft, msLangID_AMHARIC_ETHIOPIA},

	{"ps", truetype.PlatformMicrosoft, msLangID_PASHTO_AFGHANISTAN},
	{"phi", truetype.PlatformMicrosoft, msLangID_FILIPINO_PHILIPPINES},
	{"div", truetype.PlatformMicrosoft, msLangID_DHIVEHI_MALDIVES},

	{"om", truetype.PlatformMicrosoft, msLangID_OROMO_ETHIOPIA},
	{"ti", truetype.PlatformMicrosoft, msLangID_TIGRIGNA_ETHIOPIA},
	{"ti", truetype.PlatformMicrosoft, msLangID_TIGRIGNA_ERYTHREA},

	/* New additions from Windows Xp/Passport SDK 2001-11-10. */

	{"es", truetype.PlatformMicrosoft, msLangID_SPANISH_UNITED_STATES},
	// The following two IDs blatantly violate MS specs by using a sublanguage >,.                                         */
	{"es", truetype.PlatformMicrosoft, msLangID_SPANISH_LATIN_AMERICA},
	{"fr", truetype.PlatformMicrosoft, msLangID_FRENCH_NORTH_AFRICA},

	{"fr", truetype.PlatformMicrosoft, msLangID_FRENCH_MOROCCO},
	{"fr", truetype.PlatformMicrosoft, msLangID_FRENCH_HAITI},
	{"bn", truetype.PlatformMicrosoft, msLangID_BENGALI_BANGLADESH},
	{"ar", truetype.PlatformMicrosoft, msLangID_PUNJABI_ARABIC_PAKISTAN},
	{"mn", truetype.PlatformMicrosoft, msLangID_MONGOLIAN_PRC},
	{"ha", truetype.PlatformMicrosoft, msLangID_HAUSA_NIGERIA},
	{"yo", truetype.PlatformMicrosoft, msLangID_YORUBA_NIGERIA},
	/* language codes from, to, are (still) unknown. */
	{"ibo", truetype.PlatformMicrosoft, msLangID_IGBO_NIGERIA},
	{"kau", truetype.PlatformMicrosoft, msLangID_KANURI_NIGERIA},
	{"gn", truetype.PlatformMicrosoft, msLangID_GUARANI_PARAGUAY},
	{"haw", truetype.PlatformMicrosoft, msLangID_HAWAIIAN_UNITED_STATES},
	{"la", truetype.PlatformMicrosoft, msLangID_LATIN},
	{"so", truetype.PlatformMicrosoft, msLangID_SOMALI_SOMALIA},

	/* Note: Yi does not have a (proper) ISO 639-2 code, since it is mostly */
	/*       not written (but OTOH the peculiar writing system is worth     */
	/*       studying).                                                     */
	//   {  truetype.PlatformMicrosoft,	msLangID_YI_CHINA },

	{"pap", truetype.PlatformMicrosoft, msLangID_PAPIAMENTU_NETHERLANDS_ANTILLES},
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
	fromcode   encoding.Encoding
	PlatformID truetype.PlatformID
	EncodingID truetype.PlatformEncodingID
}{
	{unicode.UTF16(unicode.BigEndian, unicode.IgnoreBOM), truetype.PlatformUnicode, encodingDontCare},
	{charmap.Macintosh, truetype.PlatformMac, truetype.PEMacRoman},
	{japanese.ShiftJIS, truetype.PlatformMac, macIdJapanese},
	{unicode.UTF16(unicode.BigEndian, unicode.IgnoreBOM), truetype.PlatformMicrosoft, truetype.PEMicrosoftSymbolCs},
	{unicode.UTF16(unicode.BigEndian, unicode.IgnoreBOM), truetype.PlatformMicrosoft, truetype.PEMicrosoftUnicodeCs},
	{japanese.ShiftJIS, truetype.PlatformMicrosoft, msIdSjis},
	{simplifiedchinese.HZGB2312, truetype.PlatformMicrosoft, msIdPrc},
	{traditionalchinese.Big5, truetype.PlatformMicrosoft, msIdBig_5},
	{korean.EUCKR, truetype.PlatformMicrosoft, msIdWansung},
	// {truetype.PlatformMicrosoft, msIdJohab, "Johab"}, // Johab is not supported by golang.x/text/encoding
	{unicode.UTF16(unicode.BigEndian, unicode.IgnoreBOM), truetype.PlatformMicrosoft, truetype.PEMicrosoftUcs4},
	{charmap.ISO8859_1, truetype.PlatformIso, isoId_7BitAscii},
	{unicode.UTF16(unicode.BigEndian, unicode.IgnoreBOM), truetype.PlatformIso, isoId_10646},
	{charmap.ISO8859_1, truetype.PlatformIso, isoId_8859_1},
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

var (
	wght = truetype.MustNewTag("wght")
	wdth = truetype.MustNewTag("wdth")
	opsz = truetype.MustNewTag("opsz")
)

func hasHint(face *truetype.Font) bool { return face.HasHint }

// see `loaders`
func getFontFormat(face fonts.Face) string {
	switch face.(type) {
	case *truetype.Font:
		return "TrueType"
	case *bitmap.Font:
		return "PCF"
	case *type1.Font:
		return "Type 1"
	// case *type1c.Font:
	// 	return "CFF"
	default:
		return ""
	}
}

// groups common information betwen the face from a same file
type sharedSets struct {
	cs    Charset
	names []nameMapping
	ls    Langset
	enc   fonts.CmapEncoding
}

// this is the core of the library, which
// loads various information from the font to build the pattern
// If sets is non nil, it is used instead of recomputing it.
// Otherwise it is computed and returned
func queryFace(face fonts.Face, file string, id uint32, sets *sharedSets) (Pattern, *sharedSets) {
	var (
		variableWeight, variableWidth, variableSize, variable bool
		weight, width                                         float32 = -1., -1.
		slant                                                 int32   = -1

		// Support for glyph-variation named-instances.
		instance              *truetype.VarInstance
		varCoords             []float32 // in normalized coordinates
		weightMult, widthMult float32   = 1., 1.

		foundry                                                                       string
		nameCount, nfamily, nfamilyLang, nstyle, nstyleLang, nfullname, nfullnameLang int
		exclusiveLang                                                                 string
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

	if varFont, ok := face.(truetype.FaceVariable); ok {
		variations := varFont.Variations()

		if id>>16 == 0 {
			// Query variable font itself.
			var isValid bool
			for _, axis := range variations.Axis {
				defValue := axis.Default
				minValue := axis.Minimum
				maxValue := axis.Maximum

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
					isValid = true
				}
			}

			// a font is really variable if it has at least one valid axis
			variable = isValid
		} else if index := (id >> 16) - 1; int(index) < len(variations.Instances) {
			// Pull out weight and width from named-instance.

			instance = &variations.Instances[index]

			for i, axis := range variations.Axis {
				value := instance.Coords[i]
				defaultValue := axis.Default
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

			varCoords = varFont.NormalizeVariations(instance.Coords)
			varFont.SetVarCoordinates(varCoords) // used for spacing
		} else { // the non zero instance index is invalid
			return nil, nil
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
		os2 = ttf.OS2
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
	var nameMappings []nameMapping
	if sets == nil {
		nameMappings = make([]nameMapping, len(names))
		for i, p := range names {
			nameMappings[i] = nameMapping{NameEntry: p, idx: uint(i)}
		}

		sort.Slice(nameMappings, func(i, j int) bool { return isLess(nameMappings[i], nameMappings[j]) })
	} else {
		nameMappings = append(nameMappings, sets.names...) // copy
	}

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
						fmt.Printf("\tfound family (NameID: %2d PlatformID: %d EncodingID: %d LanguageID: 0x%04x)\n",
							sname.NameID, sname.PlatformID, sname.EncodingID, sname.LanguageID)
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
						fmt.Printf("\tfound full  (NameID: %2d PlatformID: %d EncodingID: %d LanguageID: 0x%04x)\n",
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
						fmt.Printf("\tfound style (NameID: %2d PlatformID: %d EncodingID: %d LanguageID: 0x%04x)\n",
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
			familylang, res := pat.getAtString(FAMILYLANG, n)
			for ; res == ResultMatch; familylang, res = pat.getAtString(FAMILYLANG, n) {
				if familylang == "en" {
					break
				}
				n++
			}
			if familylang == "" {
				n = 0
			}

			family, res := pat.getAtString(FAMILY, n)
			if res != ResultMatch {
				return nil, nil
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
	for _, style := range pat.GetStrings(STYLE) {
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
	var (
		cs  Charset
		enc fonts.CmapEncoding
	)
	if sets == nil {
		cs, enc = getCharSet(face)
	} else {
		cs, enc = sets.cs.Copy(), sets.enc
	}

	if enc == fonts.EncOther {
		return nil, nil
	}

	// getCharSet() chose the encoding; test it for symbol.
	symbol := enc == fonts.EncSymbol
	pat.AddBool(SYMBOL, symbol)

	spacing := getSpacing(face)

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
			return nil, nil
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
		if sets == nil {
			ls = buildLangSet(cs, exclusiveLang)
		} else {
			ls = sets.ls.Copy()
		}
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

	return pat, &sharedSets{cs, nameMappings, ls, enc}
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

func getSpacing(face fonts.Face) int32 {
	cmap, enc := face.Cmap()
	if enc != fonts.EncUnicode && enc != fonts.EncSymbol {
		return MONO
	}

	advances := make([]int, 0, 3)
	iter := cmap.Iter()
	for iter.Next() && len(advances) < 3 {
		_, glyph := iter.Char()
		advance := int(face.HorizontalAdvance(glyph))
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

	if _, isSil := face.IsGraphite(); isSil {
		complexFeats = []byte("ttable:Silf ")
	}

	gposScripts := face.LayoutTables().GPOS.Scripts
	gsubScripts := face.LayoutTables().GSUB.Scripts
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
