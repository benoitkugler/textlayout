// copied from https://git.maze.io/go/unipdf/src/branch/master/internal/textencoding
package glyphsnames

import (
	"regexp"
	"strconv"
	"strings"
)

// GlyphToRune returns the rune corresponding to glyph `glyph` if there is one.
func GlyphToRune(glyph string) (rune, bool) {
	// We treat glyph "eight.lf" the same as glyph "eight".
	if strings.Contains(string(glyph), ".") {
		groups := rePrefix.FindStringSubmatch(string(glyph))
		if groups != nil {
			glyph = string(groups[1])
		}
	}
	// First lookup the glyph in all the tables.
	if alias, ok := glyphAliases[glyph]; ok {
		glyph = alias
	}
	if r, ok := glyphlistGlyphToRuneMap[glyph]; ok {
		return r, true
	}
	if r, ok := ligatureMap[glyph]; ok {
		return r, true
	}

	// Next try all the glyph naming conventions.
	if groups := reUniEncoding.FindStringSubmatch(string(glyph)); groups != nil {
		n, err := strconv.ParseInt(groups[1], 16, 32)
		if err == nil {
			return rune(n), true
		}
	}

	if groups := reEncoding.FindStringSubmatch(string(glyph)); groups != nil {
		n, err := strconv.Atoi(groups[1])
		if err == nil {
			return rune(n), true
		}
	}

	return 0, false
}

var (
	reEncoding    = regexp.MustCompile(`^[A-Za-z](\d{1,5})$`) // C211
	reUniEncoding = regexp.MustCompile(`^uni([\dA-F]{4})$`)   // uniFB03
	rePrefix      = regexp.MustCompile(`^(\w+)\.\w+$`)        // eight.pnum => eight
)

// ligatureMap are ligatures without corresponding unicode code points. We use the Unicode private
// use area (https://en.wikipedia.org/wiki/Private_Use_Areas) to store them.
// These runes are mapped to strings in RuneToString which uses the reverse mappings in
// ligatureToString.
var ligatureMap = map[string]rune{
	"f_t":   0xe000,
	"f_j":   0xe001,
	"f_b":   0xe002,
	"f_h":   0xe003,
	"f_k":   0xe004,
	"t_t":   0xe005,
	"t_f":   0xe006,
	"f_f_j": 0xe007,
	"f_f_b": 0xe008,
	"f_f_h": 0xe009,
	"f_f_k": 0xe00a,
	"T_h":   0xe00b,
}

var glyphAliases = map[string]string{ // 2462 entries
	"f_f":                            "ff",
	"f_f_i":                          "ffi",
	"f_f_l":                          "ffl",
	"f_i":                            "fi",
	"f_l":                            "fl",
	"ascriptturn":                    "AEmacron",
	"mturndescend":                   "Adblgrave",
	"aturn":                          "Adotmacron",
	"nlftlfthook":                    "Ainvertedbreve",
	"upAlpha":                        "Alpha",
	"Ismallcap":                      "Aringacute",
	"Cbb":                            "BbbC",
	"Cdblstruck":                     "BbbC",
	"Hbb":                            "BbbH",
	"Hdblstruck":                     "BbbH",
	"Nbb":                            "BbbN",
	"Ndblstruck":                     "BbbN",
	"Pbb":                            "BbbP",
	"Pdblstruck":                     "BbbP",
	"Qbb":                            "BbbQ",
	"Qdblstruck":                     "BbbQ",
	"Rbb":                            "BbbR",
	"Rdblstruck":                     "BbbR",
	"Zbb":                            "BbbZ",
	"Zdblstruck":                     "BbbZ",
	"upBeta":                         "Beta",
	"OI":                             "Btopbar",
	"Hmacron":                        "Cacute",
	"Cdot":                           "Cdotaccent",
	"Che":                            "Checyrillic",
	"afii10041":                      "Checyrillic",
	"lcircumflex":                    "Chedescendercyrillic",
	"upChi":                          "Chi",
	"yusbig":                         "Chi",
	"gcursive":                       "DZ",
	"Gbar":                           "DZcaron",
	"Dslash":                         "Dcroat",
	"De":                             "Decyrillic",
	"afii10021":                      "Decyrillic",
	"Khartdes":                       "Deicoptic",
	"increment":                      "Delta",
	"upDelta":                        "Deltagreek",
	"eshlooprev":                     "Dhook",
	"mbfdigamma":                     "Digamma",
	"GeKarev":                        "Digammagreek",
	"upDigamma":                      "Digammagreek",
	"Gsmallcap":                      "Dz",
	"gbar":                           "Dzcaron",
	"Dzhe":                           "Dzhecyrillic",
	"afii10145":                      "Dzhecyrillic",
	"Ecyril":                         "Ecyrillic",
	"afii10053":                      "Ecyrillic",
	"Nsmallcap":                      "Edblgrave",
	"Edot":                           "Edotaccent",
	"OEsmallcap":                     "Einvertedbreve",
	"El":                             "Elcyrillic",
	"afii10029":                      "Elcyrillic",
	"Em":                             "Emcyrillic",
	"afii10030":                      "Emcyrillic",
	"Ng":                             "Eng",
	"kra":                            "Eogonek",
	"upEpsilon":                      "Epsilon",
	"strictequivalence":              "Equiv",
	"Trthook":                        "Ereversed",
	"Ecyrilrev":                      "Ereversedcyrillic",
	"afii10047":                      "Ereversedcyrillic",
	"upEta":                          "Eta",
	"Euler":                          "Eulerconst",
	"euro":                           "Euro",
	"epsilon1revclosed":              "Ezhcaron",
	"Ohook":                          "Feicoptic",
	"Upsilon2":                       "Fhook",
	"Fturn":                          "Finv",
	"FFIsmall":                       "Fsmall",
	"FFLsmall":                       "Fsmall",
	"FFsmall":                        "Fsmall",
	"FIsmall":                        "Fsmall",
	"FLsmall":                        "Fsmall",
	"babygamma":                      "Gacute",
	"upGamma":                        "Gamma",
	"Ustrt":                          "Gangiacoptic",
	"drthook":                        "Gcaron",
	"Gcedilla":                       "Gcommaaccent",
	"Gdot":                           "Gdotaccent",
	"Ge":                             "Gecyrillic",
	"afii10020":                      "Gecyrillic",
	"Geupturn":                       "Gheupturncyrillic",
	"afii10050":                      "Gheupturncyrillic",
	"Game":                           "Gmir",
	"ogoneknosp":                     "Gsmallhook",
	"cturn":                          "Gstroke",
	"whitesquare":                    "H22073",
	"box":                            "H22073",
	"mdlgwhtsquare":                  "H22073",
	"square":                         "H22073",
	"Tertdes":                        "Horicoptic",
	"Inodot":                         "I",
	"yoghhacek":                      "Icaron",
	"Idotaccent":                     "Idot",
	"Ie":                             "Iecyrillic",
	"afii10022":                      "Iecyrillic",
	"Iblackletter":                   "Ifraktur",
	"Ifractur":                       "Ifraktur",
	"Im":                             "Ifraktur",
	"Ii":                             "Iicyrillic",
	"afii10026":                      "Iicyrillic",
	"rturnascend":                    "Iinvertedbreve",
	"Io":                             "Iocyrillic",
	"afii10023":                      "Iocyrillic",
	"upIota":                         "Iota",
	"zbar":                           "Iotaafrican",
	"Yogh":                           "Istroke",
	"upKappa":                        "Kappa",
	"erev":                           "Kcaron",
	"Kcommaaccent":                   "Kcedilla",
	"Kha":                            "Khacyrillic",
	"afii10039":                      "Khacyrillic",
	"Escedilla":                      "Kheicoptic",
	"Yoghrev":                        "Khook",
	"Kje":                            "Kjecyrillic",
	"afii10061":                      "Kjecyrillic",
	"Enrtdes":                        "Koppagreek",
	"upKoppa":                        "Koppagreek",
	"ghacek":                         "LJ",
	"upLambda":                       "Lambda",
	"Lcommaaccent":                   "Lcedilla",
	"gcedilla1":                      "Lcedilla1",
	"Ldot":                           "Ldotaccent",
	"Khacek":                         "Lj",
	"Lje":                            "Ljecyrillic",
	"afii10058":                      "Ljecyrillic",
	"upMu":                           "Mu",
	"tmacron":                        "Ncaron",
	"Ncedilla":                       "Ncommaaccent",
	"tquoteright":                    "Ncommaaccent",
	"arrowdblne":                     "Nearrow",
	"upNu":                           "Nu",
	"arrowdblnw":                     "Nwarrow",
	"Ocyril":                         "Ocyrillic",
	"afii10032":                      "Ocyrillic",
	"Ohungarumlaut":                  "Odblacute",
	"rdescend":                       "Odblgrave",
	"pipe":                           "Ohorn",
	"pipedblbar":                     "Oi",
	"Ohm":                            "Omega",
	"ohm":                            "Omega",
	"upOmega":                        "Omegagreek",
	"mho":                            "Omegainv",
	"ohminverted":                    "Omegainv",
	"upOmicron":                      "Omicron",
	"yat":                            "Omicron",
	"epsilon1rev":                    "Oogonekmacron",
	"YR":                             "Oopen",
	"Ostrokeacute":                   "Oslashacute",
	"lyogh":                          "Oslashacute",
	"Yusbig":                         "Phi",
	"upPhi":                          "Phi",
	"DZhacek":                        "Phook",
	"upPi":                           "Pi",
	"planck":                         "Planckconst",
	"upPsi":                          "Psi",
	"endofproof":                     "QED",
	"eop":                            "QED",
	"Rcommaaccent":                   "Rcedilla",
	"Rsmallcap":                      "Rdblgrave",
	"Rblackletter":                   "Rfraktur",
	"Re":                             "Rfraktur",
	"Rfractur":                       "Rfraktur",
	"upRho":                          "Rho",
	"srthook":                        "Rinvertedbreve",
	"linevertdblnosp":                "Rsmallinverted",
	"Germandbls":                     "S",
	"SS":                             "S",
	"250c":                           "SF010000",
	"253c":                           "SF050000",
	"252c":                           "SF060000",
	"251c":                           "SF080000",
	"255d":                           "SF260000",
	"255c":                           "SF270000",
	"255b":                           "SF280000",
	"255e":                           "SF360000",
	"255f":                           "SF370000",
	"255a":                           "SF380000",
	"256c":                           "SF440000",
	"256b":                           "SF530000",
	"256a":                           "SF540000",
	"EnGe":                           "Sampigreek",
	"upSampi":                        "Sampigreek",
	"bbar":                           "Scaron",
	"circleS":                        "Scircle",
	"trthook":                        "Scommaaccent",
	"arrowdblse":                     "Searrow",
	"Sha":                            "Shacyrillic",
	"afii10042":                      "Shacyrillic",
	"Pehook":                         "Sheicoptic",
	"Ustrtbar":                       "Shimacoptic",
	"upSigma":                        "Sigma",
	"Germandblssmall":                "Ssmall",
	"SSsmall":                        "Ssmall",
	"Kabar":                          "Stigmagreek",
	"upStigma":                       "Stigmagreek",
	"arrowdblsw":                     "Swarrow",
	"upTau":                          "Tau",
	"Kcedilla1":                      "Tcedilla1",
	"Tcedilla":                       "Tcommaaccent",
	"upTheta":                        "Theta",
	"ahacek":                         "Tretroflexhook",
	"Tse":                            "Tsecyrillic",
	"afii10040":                      "Tsecyrillic",
	"Tshe":                           "Tshecyrillic",
	"afii10060":                      "Tshecyrillic",
	"Ucyril":                         "Ucyrillic",
	"afii10037":                      "Ucyrillic",
	"jhookdblbar":                    "Udblgrave",
	"aacutering":                     "Udieresisgrave",
	"Ihacek":                         "Uhorn",
	"Epsilon1":                       "Uhungarumlaut",
	"Udblacute":                      "Uhungarumlaut",
	"fscript":                        "Uogonek",
	"upUpsilon":                      "Upsilon",
	"Upsilonhooksymbol":              "Upsilon1",
	"Zhertdes":                       "Upsilon1",
	"zhertdes":                       "Upsilonacutehooksymbolgreek",
	"Ohacek":                         "Upsilonafrican",
	"Zecedilla":                      "Upsilondieresishooksymbolgreek",
	"Eturn":                          "Uring",
	"Ucyrilbreve":                    "Ushortcyrillic",
	"afii10062":                      "Ushortcyrillic",
	"forceextr":                      "VDash",
	"ohacek":                         "Vhook",
	"Gamma1":                         "Wcircumflex",
	"Yat":                            "Xi",
	"upXi":                           "Xi",
	"Iota1":                          "Ycircumflex",
	"Uhacek":                         "Yhook",
	"Yi":                             "Yicyrillic",
	"afii10056":                      "Yicyrillic",
	"Nhook":                          "Zcaron",
	"Zdot":                           "Zdotaccent",
	"lambdabar":                      "Zdotaccent",
	"upZeta":                         "Zeta",
	"telephoneblack":                 "a4",
	"maltese":                        "a9",
	"maltesecross":                   "a9",
	"pointingindexrightwhite":        "a12",
	"checkmark":                      "a19",
	"bigstar":                        "a35",
	"blackstar":                      "a35",
	"circledstar":                    "a37",
	"varstar":                        "a49",
	"dingasterisk":                   "a56",
	"circlesolid":                    "a71",
	"mdlgblkcircle":                  "a71",
	"bulletaltone":                   "a71",
	"blackcircle":                    "a71",
	"H18533":                         "a71",
	"filledbox":                      "a73",
	"squaresolid":                    "a73",
	"mdlgblksquare":                  "a73",
	"blacksquare":                    "a73",
	"trianglesolid":                  "a76",
	"blackuppointingtriangle":        "a76",
	"bigblacktriangleup":             "a76",
	"triagup":                        "a76",
	"blackdownpointingtriangle":      "a77",
	"triangledownsld":                "a77",
	"triagdn":                        "a77",
	"bigblacktriangledown":           "a77",
	"diamondrhombsolid":              "a78",
	"blackdiamond":                   "a78",
	"mdlgblkdiamond":                 "a78",
	"semicirclelertsld":              "a81",
	"blackrighthalfcircle":           "a81",
	"onecircle":                      "a120",
	"twocircle":                      "a121",
	"threecircle":                    "a122",
	"fourcircle":                     "a123",
	"fivecircle":                     "a124",
	"sixcircle":                      "a125",
	"sevencircle":                    "a126",
	"eightcircle":                    "a127",
	"ninecircle":                     "a128",
	"tencircle":                      "a129",
	"onecircleinversesansserif":      "a150",
	"twocircleinversesansserif":      "a151",
	"threecircleinversesansserif":    "a152",
	"fourcircleinversesansserif":     "a153",
	"fivecircleinversesansserif":     "a154",
	"sixcircleinversesansserif":      "a155",
	"sevencircleinversesansserif":    "a156",
	"eightcircleinversesansserif":    "a157",
	"ninecircleinversesansserif":     "a158",
	"updownarrow":                    "a164",
	"arrowbothv":                     "a164",
	"arrowupdn":                      "a164",
	"draftingarrow":                  "a166",
	"arrowrightheavy":                "a169",
	"Yoghhacek":                      "acaron",
	"acutecmb":                       "acutecomb",
	"arrowanticlockw":                "acwopencirclearrow",
	"upslopeellipsis":                "adots",
	"lrthook":                        "aeacute",
	"lefttoright":                    "afii299",
	"righttoleft":                    "afii300",
	"zerojoin":                       "afii301",
	"Acyril":                         "afii10017",
	"Acyrillic":                      "afii10017",
	"Be":                             "afii10018",
	"Becyrillic":                     "afii10018",
	"Vecyrillic":                     "afii10019",
	"Ve":                             "afii10019",
	"Zhe":                            "afii10024",
	"Zhecyrillic":                    "afii10024",
	"Zecyrillic":                     "afii10025",
	"Ze":                             "afii10025",
	"Iibreve":                        "afii10027",
	"Iishortcyrillic":                "afii10027",
	"Kacyrillic":                     "afii10028",
	"Ka":                             "afii10028",
	"En":                             "afii10031",
	"Encyrillic":                     "afii10031",
	"Pecyril":                        "afii10033",
	"Pecyrillic":                     "afii10033",
	"Ercyrillic":                     "afii10034",
	"Er":                             "afii10034",
	"Es":                             "afii10035",
	"Escyrillic":                     "afii10035",
	"Tecyrillic":                     "afii10036",
	"Te":                             "afii10036",
	"Efcyrillic":                     "afii10038",
	"Ef":                             "afii10038",
	"Shchacyrillic":                  "afii10043",
	"Shcha":                          "afii10043",
	"Hard":                           "afii10044",
	"Hardsigncyrillic":               "afii10044",
	"Yericyrillic":                   "afii10045",
	"Yeri":                           "afii10045",
	"Soft":                           "afii10046",
	"Softsigncyrillic":               "afii10046",
	"Iu":                             "afii10048",
	"IUcyrillic":                     "afii10048",
	"Ia":                             "afii10049",
	"IAcyrillic":                     "afii10049",
	"Dje":                            "afii10051",
	"Djecyrillic":                    "afii10051",
	"Gje":                            "afii10052",
	"Gjecyrillic":                    "afii10052",
	"Dze":                            "afii10054",
	"Dzecyrillic":                    "afii10054",
	"Icyril":                         "afii10055",
	"Icyrillic":                      "afii10055",
	"Je":                             "afii10057",
	"Jecyrillic":                     "afii10057",
	"Nje":                            "afii10059",
	"Njecyrillic":                    "afii10059",
	"acyrillic":                      "afii10065",
	"acyril":                         "afii10065",
	"vecyrillic":                     "afii10067",
	"ve":                             "afii10067",
	"gecyrillic":                     "afii10068",
	"ge":                             "afii10068",
	"decyrillic":                     "afii10069",
	"de":                             "afii10069",
	"io":                             "afii10071",
	"iocyrillic":                     "afii10071",
	"ze":                             "afii10073",
	"zecyrillic":                     "afii10073",
	"iibreve":                        "afii10075",
	"iishortcyrillic":                "afii10075",
	"en":                             "afii10079",
	"encyrillic":                     "afii10079",
	"te":                             "afii10084",
	"tecyrillic":                     "afii10084",
	"ucyrillic":                      "afii10085",
	"ucyril":                         "afii10085",
	"efcyrillic":                     "afii10086",
	"ef":                             "afii10086",
	"kha":                            "afii10087",
	"khacyrillic":                    "afii10087",
	"shacyrillic":                    "afii10090",
	"sha":                            "afii10090",
	"shchacyrillic":                  "afii10091",
	"shcha":                          "afii10091",
	"iu":                             "afii10096",
	"iucyrillic":                     "afii10096",
	"iacyrillic":                     "afii10097",
	"ia":                             "afii10097",
	"dzecyrillic":                    "afii10102",
	"dze":                            "afii10102",
	"icyrillic":                      "afii10103",
	"icyril":                         "afii10103",
	"je":                             "afii10105",
	"jecyrillic":                     "afii10105",
	"njecyrillic":                    "afii10107",
	"nje":                            "afii10107",
	"kjecyrillic":                    "afii10109",
	"kje":                            "afii10109",
	"ushortcyrillic":                 "afii10110",
	"ucyrilbreve":                    "afii10110",
	"Yatcyrillic":                    "afii10146",
	"Fitacyrillic":                   "afii10147",
	"Izhitsacyrillic":                "afii10148",
	"fitacyrillic":                   "afii10195",
	"izhitsacyrillic":                "afii10196",
	"afii10190":                      "afii10196",
	"arabiccomma":                    "afii57388",
	"commaarabic":                    "afii57388",
	"threearabic":                    "afii57395",
	"threehackarabic":                "afii57395",
	"arabicindicdigitthree":          "afii57395",
	"sixhackarabic":                  "afii57398",
	"arabicindicdigitsix":            "afii57398",
	"sixarabic":                      "afii57398",
	"sevenhackarabic":                "afii57399",
	"arabicindicdigitseven":          "afii57399",
	"sevenarabic":                    "afii57399",
	"arabicsemicolon":                "afii57403",
	"semicolonarabic":                "afii57403",
	"questionarabic":                 "afii57407",
	"arabicquestionmark":             "afii57407",
	"alefmaddaabovearabic":           "afii57410",
	"alefwithmaddaabove":             "afii57410",
	"alefhamzaabovearabic":           "afii57411",
	"alefwithhamzaabove":             "afii57411",
	"wawwithhamzaabove":              "afii57412",
	"wawhamzaabovearabic":            "afii57412",
	"teh":                            "afii57418",
	"teharabic":                      "afii57418",
	"hah":                            "afii57421",
	"haharabic":                      "afii57421",
	"khaharabic":                     "afii57422",
	"khah":                           "afii57422",
	"dalarabic":                      "afii57423",
	"dal":                            "afii57423",
	"seenarabic":                     "afii57427",
	"seen":                           "afii57427",
	"sheenarabic":                    "afii57428",
	"sheen":                          "afii57428",
	"sadarabic":                      "afii57429",
	"sad":                            "afii57429",
	"dad":                            "afii57430",
	"dadarabic":                      "afii57430",
	"ainarabic":                      "afii57433",
	"ain":                            "afii57433",
	"feharabic":                      "afii57441",
	"feh":                            "afii57441",
	"qaf":                            "afii57442",
	"qafarabic":                      "afii57442",
	"arabickaf":                      "afii57443",
	"kafarabic":                      "afii57443",
	"lam":                            "afii57444",
	"lamarabic":                      "afii57444",
	"meem":                           "afii57445",
	"meemarabic":                     "afii57445",
	"fathatanarabic":                 "afii57451",
	"fathatan":                       "afii57451",
	"dammatan":                       "afii57452",
	"dammatanarabic":                 "afii57452",
	"dammatanaltonearabic":           "afii57452",
	"kasraarabic":                    "afii57456",
	"kasra":                          "afii57456",
	"jeh":                            "afii57508",
	"jeharabic":                      "afii57508",
	"tteharabic":                     "afii57511",
	"ddalarabic":                     "afii57512",
	"noonghunnaarabic":               "afii57514",
	"arabicae":                       "afii57534",
	"sheqel":                         "afii57636",
	"sheqelhebrew":                   "afii57636",
	"newsheqelsign":                  "afii57636",
	"newsheqel":                      "afii57636",
	"maqaf":                          "afii57645",
	"maqafhebrew":                    "afii57645",
	"gimelhebrew":                    "afii57666",
	"gimel":                          "afii57666",
	"hehebrew":                       "afii57668",
	"he":                             "afii57668",
	"zayin":                          "afii57670",
	"zayinhebrew":                    "afii57670",
	"hethebrew":                      "afii57671",
	"het":                            "afii57671",
	"yodhebrew":                      "afii57673",
	"yod":                            "afii57673",
	"finalkafshevahebrew":            "afii57674",
	"finalkaf":                       "afii57674",
	"finalkafsheva":                  "afii57674",
	"finalkafqamatshebrew":           "afii57674",
	"finalkafhebrew":                 "afii57674",
	"finalkafqamats":                 "afii57674",
	"kaffinal":                       "afii57674",
	"finalnunhebrew":                 "afii57679",
	"nunfinal":                       "afii57679",
	"finalnun":                       "afii57679",
	"pehebrew":                       "afii57684",
	"pe":                             "afii57684",
	"tsadi":                          "afii57686",
	"tsadihebrew":                    "afii57686",
	"shinwithsindot":                 "afii57695",
	"shinsindothebrew":               "afii57695",
	"shinsindot":                     "afii57695",
	"vavvavhebrew":                   "afii57716",
	"vavdbl":                         "afii57716",
	"vavyodhebrew":                   "afii57717",
	"vavyod":                         "afii57717",
	"qamatsquarterhebrew":            "afii57797",
	"qamatsqatanquarterhebrew":       "afii57797",
	"qamats1a":                       "afii57797",
	"qamatshebrew":                   "afii57797",
	"qamatsqatannarrowhebrew":        "afii57797",
	"unitseparator":                  "afii57797",
	"qamatswidehebrew":               "afii57797",
	"qamats":                         "afii57797",
	"qamats27":                       "afii57797",
	"qamatsqatanhebrew":              "afii57797",
	"qamatsqatanwidehebrew":          "afii57797",
	"qamatsde":                       "afii57797",
	"qamats1c":                       "afii57797",
	"qamats29":                       "afii57797",
	"qamatsnarrowhebrew":             "afii57797",
	"qamats10":                       "afii57797",
	"qamats33":                       "afii57797",
	"sheva2e":                        "afii57799",
	"shevaquarterhebrew":             "afii57799",
	"sheva15":                        "afii57799",
	"sheva115":                       "afii57799",
	"shevahebrew":                    "afii57799",
	"sheva":                          "afii57799",
	"endtransblock":                  "afii57799",
	"sheva22":                        "afii57799",
	"shevawidehebrew":                "afii57799",
	"shevanarrowhebrew":              "afii57799",
	"sindothebrew":                   "afii57803",
	"sindot":                         "afii57803",
	"rafehebrew":                     "afii57841",
	"rafe":                           "afii57841",
	"paseq":                          "afii57842",
	"paseqhebrew":                    "afii57842",
	"lscript":                        "afii61289",
	"lsquare":                        "afii61289",
	"liter":                          "afii61289",
	"ell":                            "afii61289",
	"pdf":                            "afii61573",
	"lro":                            "afii61574",
	"rlo":                            "afii61575",
	"zerowidthnonjoiner":             "afii61664",
	"cwm":                            "afii61664",
	"zeronojoin":                     "afii61664",
	"compwordmark":                   "afii61664",
	"arabicfivepointedstar":          "afii63167",
	"asteriskaltonearabic":           "afii63167",
	"asteriskarabic":                 "afii63167",
	"commareversedmod":               "afii64937",
	"numeralgreek":                   "afii64937",
	"ainfinal":                       "ainfinalarabic",
	"aininitial":                     "aininitialarabic",
	"ainmedial":                      "ainmedialarabic",
	"nrthook":                        "ainvertedbreve",
	"afii57664":                      "alef",
	"alefhebrew":                     "alef",
	"afii57415":                      "alefarabic",
	"arabicalef":                     "alefarabic",
	"alefwithmapiq":                  "alefdageshhebrew",
	"aleffinal":                      "aleffinalarabic",
	"alefwithhamzaabovefinal":        "alefhamzaabovefinalarabic",
	"afii57413":                      "alefhamzabelowarabic",
	"alefwithhamzabelow":             "alefhamzabelowarabic",
	"alefwithhamzabelowfinal":        "alefhamzabelowfinalarabic",
	"aleflamed":                      "aleflamedhebrew",
	"alefwithmaddaabovefinal":        "alefmaddaabovefinalarabic",
	"afii57449":                      "alefmaksuraarabic",
	"alefmaksura":                    "alefmaksuraarabic",
	"alefmaksurafinal":               "alefmaksurafinalarabic",
	"yehmedial":                      "alefmaksuramedialarabic",
	"yehmedialarabic":                "alefmaksuramedialarabic",
	"alefwithpatah":                  "alefpatahhebrew",
	"alefwithqamats":                 "alefqamatshebrew",
	"alephmath":                      "aleph",
	"backcong":                       "allequal",
	"upalpha":                        "alpha",
	"c158":                           "amacron",
	"langle":                         "angbracketleft",
	"rangle":                         "angbracketright",
	"afii59770":                      "angkhankhuthai",
	"angbracketleftBig":              "angleleft",
	"angbracketleftBigg":             "angleleft",
	"angbracketleftbig":              "angleleft",
	"angbracketleftbigg":             "angleleft",
	"angbracketrightBig":             "angleright",
	"angbracketrightBigg":            "angleright",
	"angbracketrightbig":             "angleright",
	"angbracketrightbigg":            "angleright",
	"Angstrom":                       "angstrom",
	"acwgapcirclearrow":              "anticlockwise",
	"afii57929":                      "apostrophemod",
	"approachlimit":                  "approaches",
	"doteq":                          "approaches",
	"almostequal":                    "approxequal",
	"approx":                         "approxequal",
	"equaldotleftright":              "approxequalorimage",
	"fallingdotseq":                  "approxequalorimage",
	"tildetrpl":                      "approxident",
	"almostorequal":                  "approxorequal",
	"approxeq":                       "approxorequal",
	"profline":                       "arc",
	"corresponds":                    "arceq",
	"arrowsemanticlockw":             "archleftdown",
	"curvearrowleft":                 "archleftdown",
	"arrowsemclockw":                 "archrightdown",
	"curvearrowright":                "archrightdown",
	"lmidtilde":                      "aringacute",
	"a163":                           "arrowboth",
	"leftrightarrow":                 "arrowboth",
	"downdasharrow":                  "arrowdashdown",
	"leftdasharrow":                  "arrowdashleft",
	"rightdasharrow":                 "arrowdashright",
	"updasharrow":                    "arrowdashup",
	"Leftrightarrow":                 "arrowdblboth",
	"arrowdbllongboth":               "arrowdblboth",
	"dblarrowleft":                   "arrowdblboth",
	"Updownarrow":                    "arrowdblbothv",
	"arrowdbllongbothv":              "arrowdblbothv",
	"Downarrow":                      "arrowdbldown",
	"Leftarrow":                      "arrowdblleft",
	"arrowleftdbl":                   "arrowdblleft",
	"Rightarrow":                     "arrowdblright",
	"dblarrowright":                  "arrowdblright",
	"Uparrow":                        "arrowdblup",
	"downarrow":                      "arrowdown",
	"swarrow":                        "arrowdownleft",
	"searrow":                        "arrowdownright",
	"arrowopendown":                  "arrowdownwhite",
	"downwhitearrow":                 "arrowdownwhite",
	"iotasub":                        "arrowheadrightmod",
	"hookrightarrow":                 "arrowhookleft",
	"hookleftarrow":                  "arrowhookright",
	"leftarrow":                      "arrowleft",
	"leftharpoondown":                "arrowleftbothalf",
	"arrowdblleftnot":                "arrowleftdblstroke",
	"nLeftarrow":                     "arrowleftdblstroke",
	"notdblarrowleft":                "arrowleftdblstroke",
	"arrowparrleftright":             "arrowleftoverright",
	"leftrightarrows":                "arrowleftoverright",
	"arrowopenleft":                  "arrowleftwhite",
	"leftwhitearrow":                 "arrowleftwhite",
	"a161":                           "arrowright",
	"rightarrow":                     "arrowright",
	"rightharpoondown":               "arrowrightbothalf",
	"arrowdblrightnot":               "arrowrightdblstroke",
	"nRightarrow":                    "arrowrightdblstroke",
	"notdblarrowright":               "arrowrightdblstroke",
	"arrowparrrightleft":             "arrowrightoverleft",
	"rightleftarrows":                "arrowrightoverleft",
	"arrowopenright":                 "arrowrightwhite",
	"rightwhitearrow":                "arrowrightwhite",
	"barleftarrow":                   "arrowtableft",
	"rightarrowbar":                  "arrowtabright",
	"leftarrowtail":                  "arrowtailleft",
	"rightarrowtail":                 "arrowtailright",
	"Lleftarrow":                     "arrowtripleleft",
	"Rrightarrow":                    "arrowtripleright",
	"uparrow":                        "arrowup",
	"arrowupdnbse":                   "arrowupdownbase",
	"updownarrowbar":                 "arrowupdownbase",
	"nwarrow":                        "arrowupleft",
	"dblarrowupdown":                 "arrowupleftofdown",
	"updownarrows":                   "arrowupleftofdown",
	"nearrow":                        "arrowupright",
	"arrowopenup":                    "arrowupwhite",
	"upwhitearrow":                   "arrowupwhite",
	"linevert":                       "ascript",
	"macron1":                        "ascriptturned",
	"overscore1":                     "ascriptturned",
	"assertion":                      "assert",
	"ast":                            "asteriskmath",
	"asteriskcentered":               "asteriskmath",
	"approxequalalt":                 "asymptoticallyequal",
	"asymptequal":                    "asymptoticallyequal",
	"simeq":                          "asymptoticallyequal",
	"similarequal":                   "asymptoticallyequal",
	"atsign":                         "at",
	"alternativeayin":                "ayinaltonehebrew",
	"afii57682":                      "ayinhebrew",
	"ayin":                           "ayinhebrew",
	"primedblrev":                    "backdprime",
	"primedblrev1":                   "backdprime",
	"secondrev":                      "backdprime",
	"primetriplerev":                 "backtrprime",
	"primetriplerev1":                "backtrprime",
	"afii59743":                      "bahtthai",
	"vert":                           "bar",
	"verticalbar":                    "bar",
	"tableftright":                   "barleftarrowrightarrowba",
	"home":                           "barovernorthwestarrow",
	"nor":                            "barvee",
	"afii10066":                      "becyrillic",
	"be":                             "becyrillic",
	"afii57416":                      "beharabic",
	"beh":                            "beharabic",
	"behfinal":                       "behfinalarabic",
	"behinitial":                     "behinitialarabic",
	"behmedial":                      "behmedialarabic",
	"behwithmeeminitial":             "behmeeminitialarabic",
	"behwithmeemisolated":            "behmeemisolatedarabic",
	"behwithnoonfinal":               "behnoonfinalarabic",
	"upbeta":                         "beta",
	"Gehook":                         "betasymbolgreek",
	"upvarbeta":                      "betasymbolgreek",
	"betdagesh":                      "betdageshhebrew",
	"betwithdagesh":                  "betdageshhebrew",
	"bethmath":                       "beth",
	"afii57665":                      "bethebrew",
	"bet":                            "bethebrew",
	"betwithrafe":                    "betrafehebrew",
	"acute1":                         "bhook",
	"narylogicalor":                  "bigvee",
	"narylogicaland":                 "bigwedge",
	"ringsubnosp":                    "bilabialclick",
	"circlenwopen":                   "blackcircleulquadwhite",
	"semicircleleftsld":              "blacklefthalfcircle",
	"blackpointerleft":               "blackleftpointingpointer",
	"triaglf":                        "blackleftpointingpointer",
	"blacktriangleleft":              "blackleftpointingtriangle",
	"triangleleftsld1":               "blackleftpointingtriangle",
	"llblacktriangle":                "blacklowerlefttriangle",
	"triangleswsld":                  "blacklowerlefttriangle",
	"lrblacktriangle":                "blacklowerrighttriangle",
	"trianglesesld":                  "blacklowerrighttriangle",
	"filledrect":                     "blackrectangle",
	"hrectangleblack":                "blackrectangle",
	"blackpointerright":              "blackrightpointingpointer",
	"triagrt":                        "blackrightpointingpointer",
	"blacktriangleright":             "blackrightpointingtriangle",
	"trianglerightsld1":              "blackrightpointingtriangle",
	"H18543":                         "blacksmallsquare",
	"smallboxfilled":                 "blacksmallsquare",
	"smblksquare":                    "blacksmallsquare",
	"blacksmiley":                    "blacksmilingface",
	"invsmileface":                   "blacksmilingface",
	"smalltriangleinvsld":            "blacktriangledown",
	"tranglenwsld":                   "blackupperlefttriangle",
	"ulblacktriangle":                "blackupperlefttriangle",
	"trianglenesld":                  "blackupperrighttriangle",
	"urblacktriangle":                "blackupperrighttriangle",
	"blacktriangle":                  "blackuppointingsmalltriangle",
	"smalltrianglesld":               "blackuppointingsmalltriangle",
	"visiblespace":                   "blank",
	"visualspace":                    "blank",
	"blockfull":                      "block",
	"afii59706":                      "bobaimaithai",
	"bottomarc":                      "botsemicircle",
	"squarevertbisect":               "boxbar",
	"braceleftBig":                   "braceleft",
	"braceleftBigg":                  "braceleft",
	"braceleftbig":                   "braceleft",
	"braceleftbigg":                  "braceleft",
	"lbrace":                         "braceleft",
	"bracehtipdownleft":              "braceleftvertical",
	"bracehtipdownright":             "braceleftvertical",
	"bracerightBig":                  "braceright",
	"bracerightBigg":                 "braceright",
	"bracerightbig":                  "braceright",
	"bracerightbigg":                 "braceright",
	"rbrace":                         "braceright",
	"appleopen":                      "bracerightbt",
	"enter":                          "bracerightmid",
	"carriagereturnleft":             "bracerighttp",
	"bracehtipupleft":                "bracerightvertical",
	"bracehtipupright":               "bracerightvertical",
	"bracketleftBig":                 "bracketleft",
	"bracketleftBigg":                "bracketleft",
	"bracketleftbig":                 "bracketleft",
	"bracketleftbigg":                "bracketleft",
	"lbrack":                         "bracketleft",
	"bracketrightBig":                "bracketright",
	"bracketrightBigg":               "bracketright",
	"bracketrightbig":                "bracketright",
	"bracketrightbigg":               "bracketright",
	"rbrack":                         "bracketright",
	"contextmenu":                    "bracketrightbt",
	"power":                          "bracketrighttp",
	"rho1":                           "bridgeinvertedbelowcmb",
	"smblkcircle":                    "bullet",
	"bulletmath":                     "bulletoperator",
	"productdot":                     "bulletoperator",
	"vysmblkcircle":                  "bulletoperator",
	"bullseye1":                      "bullseye",
	"ct":                             "c",
	"overstore":                      "c143",
	"hmacron":                        "cacute",
	"candra":                         "candrabinducmb",
	"whitearrowupfrombar":            "capslock",
	"afii61248":                      "careof",
	"caret":                          "caretinsert",
	"check":                          "caroncmb",
	"carriagerreturn":                "carriagereturn",
	"linevertsub":                    "ccurl",
	"cdotaccent":                     "cdot",
	"Koppa":                          "cedillacmb",
	"ceilingleftBig":                 "ceilingleft",
	"ceilingleftBigg":                "ceilingleft",
	"ceilingleftbig":                 "ceilingleft",
	"ceilingleftbigg":                "ceilingleft",
	"lceil":                          "ceilingleft",
	"ceilingrightBig":                "ceilingright",
	"ceilingrightBigg":               "ceilingright",
	"ceilingrightbig":                "ceilingright",
	"ceilingrightbigg":               "ceilingright",
	"rceil":                          "ceilingright",
	"celsius":                        "centigrade",
	"degreecentigrade":               "centigrade",
	"CL":                             "centreline",
	"afii10089":                      "checyrillic",
	"che":                            "checyrillic",
	"upchi":                          "chi",
	"afii59690":                      "chochangthai",
	"afii59688":                      "chochanthai",
	"afii59689":                      "chochingthai",
	"afii59692":                      "chochoethai",
	"ringequal":                      "circeq",
	"circledast":                     "circleasterisk",
	"circlebottomsld":                "circlebottomhalfblack",
	"enclosecircle":                  "circlecopyrt",
	"circleminus1":                   "circleddash",
	"circledequal":                   "circleequal",
	"circlemultiplydisplay":          "circlemultiply",
	"circlemultiplytext":             "circlemultiply",
	"otimes":                         "circlemultiply",
	"timescircle":                    "circlemultiply",
	"circledot":                      "circleot",
	"circledotdisplay":               "circleot",
	"circledottext":                  "circleot",
	"odot":                           "circleot",
	"circleplusdisplay":              "circleplus",
	"circleplustext":                 "circleplus",
	"oplus":                          "circleplus",
	"pluscircle":                     "circleplus",
	"circledcirc":                    "circlering",
	"circletopsld":                   "circletophalfblack",
	"circlenesld":                    "circleurquadblack",
	"circleverthatch":                "circlevertfill",
	"circlelefthalfblack":            "circlewithlefthalfblack",
	"circleleftsld":                  "circlewithlefthalfblack",
	"circlerighthalfblack":           "circlewithrighthalfblack",
	"circlerightsld":                 "circlewithrighthalfblack",
	"hat":                            "circumflexcmb",
	"hatwide":                        "circumflexcmb",
	"hatwider":                       "circumflexcmb",
	"hatwidest":                      "circumflexcmb",
	"cwgapcirclearrow":               "clockwise",
	"a112":                           "club",
	"clubsuit":                       "club",
	"clubsuitblack":                  "club",
	"varclubsuit":                    "clubsuitwhite",
	"arrowsoutheast":                 "coarmenian",
	"mathcolon":                      "colon",
	"colonequal":                     "coloneq",
	"Colonmonetary":                  "colonmonetary",
	"coloncur":                       "colonmonetary",
	"coloncurrency":                  "colonmonetary",
	"colonsign":                      "colonmonetary",
	"iotadiaeresis":                  "commaabovecmb",
	"ocommatopright":                 "commaaboverightcmb",
	"upsilondiaeresis":               "commareversedabovecmb",
	"oturnedcomma":                   "commaturnedabovecmb",
	"approximatelyequal":             "congruent",
	"cong":                           "congruent",
	"contintegral":                   "contourintegral",
	"contintegraldisplay":            "contourintegral",
	"contintegraltext":               "contourintegral",
	"oint":                           "contourintegral",
	"ACK":                            "controlACK",
	"BEL":                            "controlBEL",
	"BS":                             "controlBS",
	"CAN":                            "controlCAN",
	"CR":                             "controlCR",
	"nonmarkingreturn":               "controlCR",
	"XON":                            "controlDC1",
	"DC1":                            "controlDC1",
	"DC2":                            "controlDC2",
	"XOF":                            "controlDC3",
	"DC3":                            "controlDC3",
	"DC4":                            "controlDC4",
	"DEL":                            "controlDEL",
	"DC0":                            "controlDLE",
	"DLE":                            "controlDLE",
	"EM":                             "controlEM",
	"ENQ":                            "controlENQ",
	"EOT":                            "controlEOT",
	"ESC":                            "controlESC",
	"ETB":                            "controlETB",
	"ETX":                            "controlETX",
	"FF":                             "controlFF",
	"FS":                             "controlFS",
	"IFS":                            "controlFS",
	"GS":                             "controlGS",
	"IGS":                            "controlGS",
	"HT":                             "controlHT",
	"LF":                             "controlLF",
	"NAK":                            "controlNAK",
	".null":                          "controlNULL",
	"NUL":                            "controlNULL",
	"IRS":                            "controlRS",
	"RS":                             "controlRS",
	"SI":                             "controlSI",
	"SO":                             "controlSO",
	"STX":                            "controlSOT",
	"SOH":                            "controlSTX",
	"EOF":                            "controlSUB",
	"SUB":                            "controlSUB",
	"SYN":                            "controlSYN",
	"IUS":                            "controlUS",
	"US":                             "controlUS",
	"VT":                             "controlVT",
	"amalg":                          "coproduct",
	"coprod":                         "coproductdisplay",
	"coproducttext":                  "coproductdisplay",
	"dotdblsubnosp":                  "cstretched",
	"multiplymultiset":               "cupdot",
	"multiset":                       "cupleftarrow",
	"curland":                        "curlyand",
	"curlywedge":                     "curlyand",
	"uprise":                         "curlyand",
	"looparrowleft":                  "curlyleft",
	"curlor":                         "curlyor",
	"curlyvee":                       "curlyor",
	"downfall":                       "curlyor",
	"looparrowright":                 "curlyright",
	"arrowclockw":                    "cwopencirclearrow",
	"dadfinal":                       "dadfinalarabic",
	"dadinitial":                     "dadinitialarabic",
	"dadmedial":                      "dadmedialarabic",
	"afii57807":                      "dagesh",
	"dageshhebrew":                   "dagesh",
	"spaceopenbox":                   "dagesh",
	"ddagger":                        "daggerdbl",
	"daletdageshhebrew":              "daletdagesh",
	"daletwithdagesh":                "daletdagesh",
	"dalethmath":                     "daleth",
	"afii57667":                      "daletqamatshebrew",
	"dalet":                          "daletqamatshebrew",
	"dalethatafpatah":                "daletqamatshebrew",
	"dalethatafpatahhebrew":          "daletqamatshebrew",
	"dalethatafsegol":                "daletqamatshebrew",
	"dalethatafsegolhebrew":          "daletqamatshebrew",
	"dalethebrew":                    "daletqamatshebrew",
	"dalethiriq":                     "daletqamatshebrew",
	"dalethiriqhebrew":               "daletqamatshebrew",
	"daletholam":                     "daletqamatshebrew",
	"daletholamhebrew":               "daletqamatshebrew",
	"daletpatah":                     "daletqamatshebrew",
	"daletpatahhebrew":               "daletqamatshebrew",
	"daletqamats":                    "daletqamatshebrew",
	"daletqubuts":                    "daletqamatshebrew",
	"daletqubutshebrew":              "daletqamatshebrew",
	"daletsegol":                     "daletqamatshebrew",
	"daletsegolhebrew":               "daletqamatshebrew",
	"daletsheva":                     "daletqamatshebrew",
	"daletshevahebrew":               "daletqamatshebrew",
	"dalettsere":                     "daletqamatshebrew",
	"dalettserehebrew":               "daletqamatshebrew",
	"dalfinal":                       "dalfinalarabic",
	"afii57455":                      "dammaarabic",
	"damma":                          "dammaarabic",
	"dammalowarabic":                 "dammaarabic",
	"dammahontatweel":                "dammamedial",
	"dargahebrew":                    "dargalefthebrew",
	"shiftout":                       "dargalefthebrew",
	"excess":                         "dashcolon",
	"dblarrowdown":                   "dblarrowdwn",
	"downdownarrows":                 "dblarrowdwn",
	"twoheadleftarrow":               "dblarrowheadleft",
	"twoheadrightarrow":              "dblarrowheadright",
	"upuparrows":                     "dblarrowup",
	"lBrack":                         "dblbracketleft",
	"rBrack":                         "dblbracketright",
	"doubleintegral":                 "dblintegral",
	"iint":                           "dblintegral",
	"integraldbl":                    "dblintegral",
	"Vert":                           "dblverticalbar",
	"bardbl":                         "dblverticalbar",
	"verticalbardbl":                 "dblverticalbar",
	"vertlinedbl":                    "dblverticalbar",
	"downslopeellipsis":              "ddots",
	"decimalseparatorarabic":         "decimalseparatorpersian",
	"deltaequal":                     "defines",
	"triangleq":                      "defines",
	"kelvin":                         "degreekelvin",
	"devcon4":                        "dehihebrew",
	"khartdes":                       "deicoptic",
	"updelta":                        "delta",
	"macronsubnosp":                  "dezh",
	"gravesub":                       "dhook",
	"a111":                           "diamond",
	"diamondsolid":                   "diamond",
	"vardiamondsuit":                 "diamond",
	"smwhtdiamond":                   "diamondmath",
	"diamondsuit":                    "diamondsuitwhite",
	"ddot":                           "dieresiscmb",
	"dialytikatonos":                 "dieresistonos",
	"bumpeq":                         "difference",
	"c144":                           "divide",
	"div":                            "divide",
	"divideonmultiply":               "dividemultiply",
	"divideontimes":                  "dividemultiply",
	"bar1":                           "divides",
	"mid":                            "divides",
	"vextendsingle":                  "divides",
	"divslash":                       "divisionslash",
	"slashmath":                      "divisionslash",
	"afii10099":                      "djecyrillic",
	"dje":                            "djecyrillic",
	"blockthreeqtrshaded":            "dkshade",
	"shadedark":                      "dkshade",
	"dcroat":                         "dmacron",
	"dslash":                         "dmacron",
	"blocklowhalf":                   "dnblock",
	"afii59694":                      "dochadathai",
	"afii59700":                      "dodekthai",
	"escudo":                         "dollar",
	"mathdollar":                     "dollar",
	"milreis":                        "dollar",
	"iotadiaeresistonos":             "dotaccent",
	"dot":                            "dotaccentcmb",
	"Stigma":                         "dotbelowcomb",
	"dotbelowcmb":                    "dotbelowcomb",
	"breveinvnosp":                   "dotlessjstrokehook",
	"geomproportion":                 "dotsminusdots",
	"proportiongeom":                 "dotsminusdots",
	"circledash":                     "dottedcircle",
	"xbsol":                          "downslope",
	"macronsub":                      "dtail",
	"gamma1":                         "dz",
	"tildesubnosp":                   "dzaltone",
	"Ghacek":                         "dzcaron",
	"underscorenosp":                 "dzcurl",
	"afii10193":                      "dzhecyrillic",
	"dzhe":                           "dzhecyrillic",
	"afii10101":                      "ecyrillic",
	"ecyril":                         "ecyrillic",
	"edot":                           "edotaccent",
	"afii57400":                      "eighthackarabic",
	"arabicindicdigiteight":          "eighthackarabic",
	"eightarabic":                    "eighthackarabic",
	"musicalnotedbl":                 "eighthnotebeamed",
	"twonotes":                       "eighthnotebeamed",
	"eightsub":                       "eightinferior",
	"extendedarabicindicdigiteight":  "eightpersian",
	"afii59768":                      "eightthai",
	"omegaclosed":                    "einvertedbreve",
	"afii10077":                      "elcyrillic",
	"el":                             "elcyrillic",
	"in":                             "element",
	"elipsis":                        "ellipsis",
	"unicodeellipsis":                "ellipsis",
	"vdots":                          "ellipsisvertical",
	"vertellipsis":                   "ellipsisvertical",
	"afii10078":                      "emcyrillic",
	"em":                             "emcyrillic",
	"punctdash":                      "emdash",
	"varnothing":                     "emptyset",
	"rangedash":                      "endash",
	"ng":                             "eng",
	"ringrighthalfcenter":            "eopen",
	"cedillanosp":                    "eopenclosed",
	"ringlefthalfsup":                "eopenreversed",
	"tackdownmid":                    "eopenreversedclosed",
	"tackupmid":                      "eopenreversedhook",
	"upepsilon":                      "epsilon",
	"upvarepsilon":                   "epsilon1",
	"chevertbar":                     "epsilon1",
	"Hcyril":                         "epsiloninv",
	"upbackepsilon":                  "epsiloninv",
	"equalcolon":                     "eqcolon",
	"definequal":                     "eqdef",
	"equalgreater":                   "eqgtr",
	"equalless":                      "eqless",
	"curlyeqsucc":                    "equalorfollows",
	"equalfollows1":                  "equalorfollows",
	"eqslantgtr":                     "equalorgreater",
	"eqslantless":                    "equalorless",
	"curlyeqprec":                    "equalorprecedes",
	"equalprecedes1":                 "equalorprecedes",
	"eqsim":                          "equalorsimilar",
	"minustilde":                     "equalorsimilar",
	"equalinferior":                  "equalsub",
	"equiv":                          "equivalence",
	"asymp":                          "equivasymptotic",
	"afii10082":                      "ercyrillic",
	"er":                             "ercyrillic",
	"acutesub":                       "ereversed",
	"afii10095":                      "ereversedcyrillic",
	"ecyrilrev":                      "ereversedcyrillic",
	"afii10083":                      "escyrillic",
	"es":                             "escyrillic",
	"candrabindunosp":                "esh",
	"apostrophesupnosp":              "eshcurl",
	"commaturnsupnosp":               "eshsquatreversed",
	"upeta":                          "eta",
	"Dbar":                           "eth",
	"Dmacron":                        "eth",
	"matheth":                        "eth",
	"arrowbothvbase":                 "etnahtalefthebrew",
	"etnahtafoukhhebrew":             "etnahtalefthebrew",
	"etnahtafoukhlefthebrew":         "etnahtalefthebrew",
	"etnahtahebrew":                  "etnahtalefthebrew",
	"Exclam":                         "exclamdbl",
	"exists":                         "existential",
	"thereexists":                    "existential",
	"plussubnosp":                    "ezh",
	"jdotlessbar":                    "ezhcaron",
	"minussubnosp":                   "ezhcurl",
	"Udieresishacek":                 "ezhreversed",
	"udieresishacek":                 "ezhtail",
	"degreefahrenheit":               "fahrenheit",
	"degreefarenheit":                "fahrenheit",
	"farenheit":                      "fahrenheit",
	"fathamedial":                    "fathahontatweel",
	"afii57454":                      "fathalowarabic",
	"fatha":                          "fathalowarabic",
	"fathaarabic":                    "fathalowarabic",
	"arrowwaveright":                 "feharmenian",
	"fehfinal":                       "fehfinalarabic",
	"fehinitial":                     "fehinitialarabic",
	"fehmedial":                      "fehmedialarabic",
	"ohook":                          "feicoptic",
	"venus":                          "female",
	"finalkafdagesh":                 "finalkafdageshhebrew",
	"finalkafwithdagesh":             "finalkafdageshhebrew",
	"afii57677":                      "finalmemhebrew",
	"finalmem":                       "finalmemhebrew",
	"memfinal":                       "finalmemhebrew",
	"afii57683":                      "finalpehebrew",
	"finalpe":                        "finalpehebrew",
	"pefinal":                        "finalpehebrew",
	"afii57685":                      "finaltsadi",
	"finaltsadihebrew":               "finaltsadi",
	"tsadifinal":                     "finaltsadi",
	"afii57397":                      "fivearabic",
	"arabicindicdigitfive":           "fivearabic",
	"fivehackarabic":                 "fivearabic",
	"fivesub":                        "fiveinferior",
	"extendedarabicindicdigitfive":   "fivepersian",
	"afii59765":                      "fivethai",
	"floorleftBig":                   "floorleft",
	"floorleftBigg":                  "floorleft",
	"floorleftbig":                   "floorleft",
	"floorleftbigg":                  "floorleft",
	"lfloor":                         "floorleft",
	"floorrightBig":                  "floorright",
	"floorrightBigg":                 "floorright",
	"floorrightbig":                  "floorright",
	"floorrightbigg":                 "floorright",
	"rfloor":                         "floorright",
	"Vcursive":                       "florin",
	"afii59711":                      "fofanthai",
	"afii59709":                      "fofathai",
	"succnapprox":                    "follownotdbleqv",
	"succneqq":                       "follownotslnteql",
	"followsnotequivlnt":             "followornoteqvlnt",
	"succnsim":                       "followornoteqvlnt",
	"notfollowsoreql":                "followsequal",
	"succeq":                         "followsequal",
	"followsequal1":                  "followsorcurly",
	"succcurlyeq":                    "followsorcurly",
	"followsequivlnt":                "followsorequal",
	"succsim":                        "followsorequal",
	"afii59759":                      "fongmanthai",
	"Vdash":                          "forces",
	"force":                          "forces",
	"Vvdash":                         "forcesbar",
	"tacktrpl":                       "forcesbar",
	"pitchfork":                      "fork",
	"afii57396":                      "fourarabic",
	"arabicindicdigitfour":           "fourarabic",
	"fourhackarabic":                 "fourarabic",
	"foursub":                        "fourinferior",
	"extendedarabicindicdigitfour":   "fourpersian",
	"afii59764":                      "fourthai",
	"fracslash":                      "fraction",
	"fraction1":                      "fraction",
	"hturn":                          "gacute",
	"afii57509":                      "gafarabic",
	"gaf":                            "gafarabic",
	"gaffinal":                       "gaffinalarabic",
	"gafinitial":                     "gafinitialarabic",
	"gafmedial":                      "gafmedialarabic",
	"upgamma":                        "gamma",
	"ustrt":                          "gangiacoptic",
	"gcommaaccent":                   "gcedilla",
	"gdotaccent":                     "gdot",
	"Bumpeq":                         "geomequivalent",
	"Doteq":                          "geometricallyequal",
	"equalsdots":                     "geometricallyequal",
	"geomequal":                      "geometricallyequal",
	"endtext":                        "gereshaccenthebrew",
	"geresh":                         "gereshhebrew",
	"endtrans":                       "gereshmuqdamhebrew",
	"enquiry":                        "gershayimaccenthebrew",
	"gershayim":                      "gershayimhebrew",
	"verymuchgreater":                "ggg",
	"afii57434":                      "ghainarabic",
	"ghain":                          "ghainarabic",
	"ghainfinal":                     "ghainfinalarabic",
	"ghaininitial":                   "ghaininitialarabic",
	"ghainmedial":                    "ghainmedialarabic",
	"afii10098":                      "gheupturncyrillic",
	"geupturn":                       "gheupturncyrillic",
	"gimelmath":                      "gimel",
	"gimeldagesh":                    "gimeldageshhebrew",
	"gimelwithdagesh":                "gimeldageshhebrew",
	"afii10100":                      "gjecyrillic",
	"gje":                            "gjecyrillic",
	"hooksubpalatnosp":               "glottalstop",
	"dotsubnosp":                     "glottalstopinverted",
	"hooksubretronosp":               "glottalstopreversed",
	"brevesubnosp":                   "glottalstopstroke",
	"breveinvsubnosp":                "glottalstopstrokereversed",
	"greaternotequivlnt":             "gnsim",
	"nabla":                          "gradient",
	"gravecomb":                      "gravecmb",
	"diaeresistonos":                 "gravelowmod",
	"gtreqqless":                     "greaterdbleqlless",
	"gtrdot":                         "greaterdot",
	"geq":                            "greaterequal",
	"greaterequalless":               "greaterequalorless",
	"greaterlessequal":               "greaterequalorless",
	"gtreqless":                      "greaterequalorless",
	"gnapprox":                       "greaternotdblequal",
	"gneq":                           "greaternotequal",
	"gtrapprox":                      "greaterorapproxeql",
	"greaterequivlnt":                "greaterorequivalent",
	"greaterorsimilar":               "greaterorequivalent",
	"gtrsim":                         "greaterorequivalent",
	"gtrless":                        "greaterorless",
	"gneqq":                          "greaterornotdbleql",
	"greaterornotequal":              "greaterornotdbleql",
	"geqq":                           "greateroverequal",
	"greaterdblequal":                "greateroverequal",
	"notgreaterdblequal":             "greateroverequal",
	"hehaltonearabic":                "haaltonearabic",
	"hahfinal":                       "hahfinalarabic",
	"hahinitial":                     "hahinitialarabic",
	"hahmedial":                      "hahmedialarabic",
	"afii57409":                      "hamzadammaarabic",
	"hamza":                          "hamzadammaarabic",
	"hamzaarabic":                    "hamzadammaarabic",
	"hamzadammatanarabic":            "hamzadammaarabic",
	"hamzafathaarabic":               "hamzadammaarabic",
	"hamzafathatanarabic":            "hamzadammaarabic",
	"hamzalowarabic":                 "hamzadammaarabic",
	"hamzalowkasraarabic":            "hamzadammaarabic",
	"hamzalowkasratanarabic":         "hamzadammaarabic",
	"hamzasukunarabic":               "hamzadammaarabic",
	"afii10092":                      "hardsigncyrillic",
	"hard":                           "hardsigncyrillic",
	"downharpoonleft":                "harpoondownleft",
	"downharpoonright":               "harpoondownright",
	"arrowlefttophalf":               "harpoonleftbarbup",
	"leftharpoonup":                  "harpoonleftbarbup",
	"rightleftharpoons":              "harpoonleftright",
	"arrowrighttophalf":              "harpoonrightbarbup",
	"rightharpoonup":                 "harpoonrightbarbup",
	"leftrightharpoons":              "harpoonrightleft",
	"upharpoonleft":                  "harpoonupleft",
	"upharpoonright":                 "harpoonupright",
	"hatafpatahwidehebrew":           "hatafpatah16",
	"hatafpatahquarterhebrew":        "hatafpatah16",
	"hatafpatahhebrew":               "hatafpatah16",
	"hatafpatahnarrowhebrew":         "hatafpatah16",
	"hatafpatah2f":                   "hatafpatah16",
	"afii57800":                      "hatafpatah16",
	"endmedium":                      "hatafpatah16",
	"hatafpatah23":                   "hatafpatah16",
	"hatafpatah":                     "hatafpatah16",
	"hatafqamatshebrew":              "hatafqamats28",
	"afii57802":                      "hatafqamats28",
	"substitute":                     "hatafqamats28",
	"hatafqamats34":                  "hatafqamats28",
	"hatafqamatswidehebrew":          "hatafqamats28",
	"hatafqamatsnarrowhebrew":        "hatafqamats28",
	"hatafqamatsquarterhebrew":       "hatafqamats28",
	"hatafqamats1b":                  "hatafqamats28",
	"hatafqamats":                    "hatafqamats28",
	"endoffile":                      "hatafqamats28",
	"afii57801":                      "hatafsegolwidehebrew",
	"cancel":                         "hatafsegolwidehebrew",
	"hatafsegol":                     "hatafsegolwidehebrew",
	"hatafsegol17":                   "hatafsegolwidehebrew",
	"hatafsegol24":                   "hatafsegolwidehebrew",
	"hatafsegol30":                   "hatafsegolwidehebrew",
	"hatafsegolhebrew":               "hatafsegolwidehebrew",
	"hatafsegolnarrowhebrew":         "hatafsegolwidehebrew",
	"hatafsegolquarterhebrew":        "hatafsegolwidehebrew",
	"a110":                           "heart",
	"heartsuitblack":                 "heart",
	"varheartsuit":                   "heart",
	"heartsuit":                      "heartsuitwhite",
	"hedagesh":                       "hedageshhebrew",
	"hewithmapiq":                    "hedageshhebrew",
	"afii57470":                      "heharabic",
	"heh":                            "heharabic",
	"hehfinal":                       "hehfinalarabic",
	"hehfinalalttwoarabic":           "hehfinalarabic",
	"hehinitial":                     "hehinitialarabic",
	"hehmedial":                      "hehmedialarabic",
	"rhotichook":                     "henghook",
	"hermitconjmatrix":               "hermitmatrix",
	"tildevertsupnosp":               "hhooksuperior",
	"hiriq":                          "hiriq14",
	"hiriq2d":                        "hiriq14",
	"afii57793":                      "hiriq14",
	"hiriqhebrew":                    "hiriq14",
	"escape":                         "hiriq14",
	"hiriqnarrowhebrew":              "hiriq14",
	"hiriqquarterhebrew":             "hiriq14",
	"hiriq21":                        "hiriq14",
	"hiriqwidehebrew":                "hiriq14",
	"afii59723":                      "hohipthai",
	"afii57806":                      "holamquarterhebrew",
	"holam":                          "holamquarterhebrew",
	"holam19":                        "holamquarterhebrew",
	"holam26":                        "holamquarterhebrew",
	"holam32":                        "holamquarterhebrew",
	"holamhebrew":                    "holamquarterhebrew",
	"holamnarrowhebrew":              "holamquarterhebrew",
	"holamwidehebrew":                "holamquarterhebrew",
	"spaceliteral":                   "holamquarterhebrew",
	"afii59726":                      "honokhukthai",
	"hookabovecomb":                  "hookcmb",
	"ovhook":                         "hookcmb",
	"tertdes":                        "horicoptic",
	"afii00208":                      "horizontalbar",
	"horizbar":                       "horizontalbar",
	"longdash":                       "horizontalbar",
	"quotedash":                      "horizontalbar",
	"rectangle":                      "hrectangle",
	"xsupnosp":                       "hsuperior",
	"SD190100":                       "hturned",
	"Zbar":                           "hv",
	"hyphen-minus":                   "hyphen",
	"hyphenchar":                     "hyphen",
	"hyphenminus":                    "hyphen",
	"hyphen1":                        "hyphentwo",
	"jhacek":                         "icaron",
	"rturn":                          "idblgrave",
	"dquoteright":                    "idieresis",
	"afii10070":                      "iecyrillic",
	"ie":                             "iecyrillic",
	"afii10074":                      "iicyrillic",
	"ii":                             "iicyrillic",
	"integraltrpl":                   "iiint",
	"tripleintegral":                 "iiint",
	"rturnhook":                      "iinvertedbreve",
	"rturnrthook":                    "iinvertedbreve",
	"auxiliaryoff":                   "iluyhebrew",
	"devcon3":                        "iluyhebrew",
	"image":                          "imageof",
	"equaldotrightleft":              "imageorapproximatelyequal",
	"imageorapproxequal":             "imageorapproximatelyequal",
	"risingdotseq":                   "imageorapproximatelyequal",
	"infty":                          "infinity",
	"clwintegral":                    "intclockwise",
	"backslashBig":                   "integerdivide",
	"backslashBigg":                  "integerdivide",
	"backslashbig":                   "integerdivide",
	"backslashbigg":                  "integerdivide",
	"backslashmath":                  "integerdivide",
	"smallsetminus":                  "integerdivide",
	"int":                            "integral",
	"integraldisplay":                "integral",
	"integraltext":                   "integral",
	"intbottom":                      "integralbt",
	"integralbottom":                 "integralbt",
	"integraltop":                    "integraltp",
	"inttop":                         "integraltp",
	"cap":                            "intersection",
	"Cap":                            "intersectiondbl",
	"bigcap":                         "intersectiondisplay",
	"intersectiontext":               "intersectiondisplay",
	"naryintersection":               "intersectiondisplay",
	"sqcap":                          "intersectionsq",
	"bulletinverse":                  "invbullet",
	"inversebullet":                  "invbullet",
	"inversewhitecircle":             "invcircle",
	"whitecircleinverse":             "invcircle",
	"Sinvlazy":                       "invlazys",
	"lazysinv":                       "invlazys",
	"invsemicircledn":                "invwhitelowerhalfcircle",
	"invsemicircleup":                "invwhiteupperhalfcircle",
	"upiota":                         "iota",
	"gammasuper":                     "iotalatin",
	"highcomman":                     "itilde",
	"bridgesubnosp":                  "jcrossedtail",
	"afii57420":                      "jeemarabic",
	"jeem":                           "jeemarabic",
	"jeemfinal":                      "jeemfinalarabic",
	"jeeminitial":                    "jeeminitialarabic",
	"jeemmedial":                     "jeemmedialarabic",
	"jehfinal":                       "jehfinalarabic",
	"overscoredblnosp":               "jsuperior",
	"afii10076":                      "kacyrillic",
	"ka":                             "kacyrillic",
	"afii57675":                      "kaf",
	"kafhebrew":                      "kaf",
	"kafdageshhebrew":                "kafdagesh",
	"kafwithdagesh":                  "kafdagesh",
	"arabickaffinal":                 "kaffinalarabic",
	"kafinitial":                     "kafinitialarabic",
	"kafmedial":                      "kafmedialarabic",
	"kafwithrafe":                    "kafrafehebrew",
	"upkappa":                        "kappa",
	"TeTse":                          "kappasymbolgreek",
	"upvarkappa":                     "kappasymbolgreek",
	"afii57440":                      "kashidaautonosidebearingarabic",
	"kashidaautoarabic":              "kashidaautonosidebearingarabic",
	"tatweel":                        "kashidaautonosidebearingarabic",
	"tatweelarabic":                  "kashidaautonosidebearingarabic",
	"kasrahontatweel":                "kasramedial",
	"afii57453":                      "kasratanarabic",
	"kasratan":                       "kasratanarabic",
	"kcedilla":                       "kcommaaccent",
	"arrowrightnot":                  "keharmenian",
	"homothetic":                     "kernelcontraction",
	"khahfinal":                      "khahfinalarabic",
	"khahinitial":                    "khahinitialarabic",
	"khahmedial":                     "khahmedialarabic",
	"escedilla":                      "kheicoptic",
	"afii59682":                      "khokhaithai",
	"afii59685":                      "khokhonthai",
	"afii59683":                      "khokhuatthai",
	"afii59684":                      "khokhwaithai",
	"afii59771":                      "khomutthai",
	"yoghrev":                        "khook",
	"afii59686":                      "khorakhangthai",
	"afii59681":                      "kokaithai",
	"archdblsubnosp":                 "kturned",
	"afii59749":                      "lakkhangyaothai",
	"lamwithaleffinal":               "lamaleffinalarabic",
	"lamwithalefhamzaabovefinal":     "lamalefhamzaabovefinalarabic",
	"lamwithalefhamzaaboveisolatedd": "lamalefhamzaaboveisolatedarabic",
	"lamwithalefhamzabelowfinal":     "lamalefhamzabelowfinalarabic",
	"lamwithalefhamzabelowisolated":  "lamalefhamzabelowisolatedarabic",
	"lamwithalefisolated":            "lamalefisolatedarabic",
	"lamwithalefmaddaabovefinal":     "lamalefmaddaabovefinalarabic",
	"lamwithalefmaddaaboveisolatedd": "lamalefmaddaaboveisolatedarabic",
	"uplambda":                       "lambda",
	"2bar":                           "lambdastroke",
	"lameddageshhebrew":              "lameddagesh",
	"lamedwithdagesh":                "lameddagesh",
	"afii57676":                      "lamedholamhebrew",
	"lamed":                          "lamedholamhebrew",
	"lamedhebrew":                    "lamedholamhebrew",
	"lamedholam":                     "lamedholamhebrew",
	"lamedholamdagesh":               "lamedholamhebrew",
	"lamedholamdageshhebrew":         "lamedholamhebrew",
	"lamfinal":                       "lamfinalarabic",
	"lamwithhahinitial":              "lamhahinitialarabic",
	"laminitial":                     "laminitialarabic",
	"lammeemjeeminitialarabic":       "laminitialarabic",
	"lammeemkhahinitialarabic":       "laminitialarabic",
	"lamwithjeeminitial":             "lamjeeminitialarabic",
	"lamwithkhahinitial":             "lamkhahinitialarabic",
	"allahisolated":                  "lamlamhehisolatedarabic",
	"lammedial":                      "lammedialarabic",
	"lamwithmeemwithhahinitial":      "lammeemhahinitialarabic",
	"lamwithmeeminitial":             "lammeeminitialarabic",
	"lgwhtcircle":                    "largecircle",
	"yoghtail":                       "lbar",
	"xsuper":                         "lbelt",
	"lcedilla":                       "lcommaaccent",
	"ldot":                           "ldotaccent",
	"droang":                         "leftangleabovecmb",
	"arrowsquiggleleft":              "leftsquigarrow",
	"lesseqqgtr":                     "lessdbleqlgreater",
	"leq":                            "lessequal",
	"lesseqgtr":                      "lessequalorgreater",
	"lessequalgreater":               "lessequalorgreater",
	"lnapprox":                       "lessnotdblequal",
	"lneq":                           "lessnotequal",
	"lessapprox":                     "lessorapproxeql",
	"leqslant":                       "lessorequalslant",
	"notlessorslnteql":               "lessorequalslant",
	"lessequivlnt":                   "lessorequivalent",
	"lessorsimilar":                  "lessorequivalent",
	"lesssim":                        "lessorequivalent",
	"lessgtr":                        "lessorgreater",
	"lessornotdbleql":                "lessornotequal",
	"lneqq":                          "lessornotequal",
	"leqq":                           "lessoverequal",
	"lessdblequal":                   "lessoverequal",
	"notlessdblequal":                "lessoverequal",
	"toneextrahigh":                  "lezh",
	"blocklefthalf":                  "lfblock",
	"glottalrevsuper":                "lhookretroflex",
	"arrowrightdown":                 "linefeed",
	"afii08941":                      "lira",
	"khacek":                         "lj",
	"afii10106":                      "ljecyrillic",
	"lje":                            "ljecyrillic",
	"swquadarc":                      "llarc",
	"verymuchless":                   "lll",
	"ssuper":                         "lmiddletilde",
	"lessnotequivlnt":                "lnsim",
	"afii59724":                      "lochulathai",
	"logicalanddisplay":              "logicaland",
	"logicalandtext":                 "logicaland",
	"wedge":                          "logicaland",
	"neg":                            "logicalnot",
	"logicalordisplay":               "logicalor",
	"logicalortext":                  "logicalor",
	"vee":                            "logicalor",
	"afii59717":                      "lolingthai",
	"Obar":                           "longs",
	"longdbls":                       "longs",
	"longsh":                         "longs",
	"longsi":                         "longs",
	"longsl":                         "longs",
	"slong":                          "longs",
	"slongt":                         "longst",
	"mdlgwhtlozenge":                 "lozenge",
	"sequadarc":                      "lrarc",
	"afii59718":                      "luthai",
	"overscore":                      "macron",
	"underbar":                       "macronbelowcmb",
	"mahapakhlefthebrew":             "mahapakhhebrew",
	"verttab":                        "mahapakhhebrew",
	"afii59755":                      "maichattawathai",
	"afii59752":                      "maiekthai",
	"afii59728":                      "maihanakatthai",
	"afii59751":                      "maitaikhuthai",
	"afii59753":                      "maithothai",
	"afii59754":                      "maitrithai",
	"afii59750":                      "maiyamokthai",
	"male":                           "mars",
	"synch":                          "masoracirclehebrew",
	"measurequal":                    "measeq",
	"rightanglearc":                  "measuredrightangle",
	"meemfinal":                      "meemfinalarabic",
	"meeminitial":                    "meeminitialarabic",
	"meemmedial":                     "meemmedialarabic",
	"meemwithmeeminitial":            "meemmeeminitialarabic",
	"afii57678":                      "mem",
	"memhebrew":                      "mem",
	"memdagesh":                      "memdageshhebrew",
	"memwithdagesh":                  "memdageshhebrew",
	"formfeed":                       "merkhahebrew",
	"merkhalefthebrew":               "merkhahebrew",
	"merkhakefulalefthebrew":         "merkhakefulahebrew",
	"Cblackletter":                   "mfrakC",
	"Cfractur":                       "mfrakC",
	"Cfraktur":                       "mfrakC",
	"Hblackletter":                   "mfrakH",
	"Hfractur":                       "mfrakH",
	"Hfraktur":                       "mfrakH",
	"Zblackletter":                   "mfrakZ",
	"Zfractur":                       "mfrakZ",
	"Zfraktur":                       "mfrakZ",
	"tonelow":                        "mhook",
	"circleminus":                    "minuscircle",
	"ominus":                         "minuscircle",
	"minussub":                       "minusinferior",
	"mp":                             "minusplus",
	"prime":                          "minute",
	"prime1":                         "minute",
	"tonemid":                        "mlonglegturned",
	"truestate":                      "models",
	"afii59713":                      "momathai",
	"Bscript":                        "mscrB",
	"Escript":                        "mscrE",
	"Fscript":                        "mscrF",
	"Hscript":                        "mscrH",
	"Iscript":                        "mscrI",
	"Lscript":                        "mscrL",
	"Mscript":                        "mscrM",
	"Rscript":                        "mscrR",
	"escript":                        "mscre",
	"gscriptmath":                    "mscrg",
	"0script":                        "mscro",
	"oscript":                        "mscro",
	"tonehigh":                       "mturned",
	"mu1":                            "mu",
	"gg":                             "muchgreater",
	"greatermuch":                    "muchgreater",
	"lessmuch":                       "muchless",
	"upmu":                           "mugreek",
	"ltimes":                         "multicloseleft",
	"rtimes":                         "multicloseright",
	"leftthreetimes":                 "multiopenleft",
	"rightthreetimes":                "multiopenright",
	"times":                          "multiply",
	"munahhebrew":                    "munahlefthebrew",
	"eighthnote":                     "musicalnote",
	"flat":                           "musicflatsign",
	"sharp":                          "musicsharpsign",
	"barwedge":                       "nand",
	"notalmostequal":                 "napprox",
	"notequivasymptotic":             "nasymp",
	"hyphennobreak":                  "nbhyphen",
	"Tmacron":                        "ncedilla",
	"ncommaaccent":                   "ncedilla",
	"afii59687":                      "ngonguthai",
	"notgreaterequivlnt":             "ngtrsim",
	"toneextralow":                   "nhookleft",
	"gravenosp":                      "nhookretroflex",
	"afii59757":                      "nikhahitthai",
	"afii57401":                      "ninehackarabic",
	"arabicindicdigitnine":           "ninehackarabic",
	"ninearabic":                     "ninehackarabic",
	"ninesub":                        "nineinferior",
	"extendedarabicindicdigitnine":   "ninepersian",
	"afii59769":                      "ninethai",
	"glottalstopbarinv":              "nlegrightlong",
	"notlessgreater":                 "nlessgtr",
	"notlessequivlnt":                "nlesssim",
	"nbspace":                        "nonbreakingspace",
	"afii59699":                      "nonenthai",
	"afii59705":                      "nonuthai",
	"afii57446":                      "noonarabic",
	"noon":                           "noonarabic",
	"noonfinal":                      "noonfinalarabic",
	"nooninitial":                    "noonhehinitialarabic",
	"nooninitialarabic":              "noonhehinitialarabic",
	"noonwithjeeminitial":            "noonjeeminitialarabic",
	"noonmedial":                     "noonmedialarabic",
	"noonwithmeeminitial":            "noonmeeminitialarabic",
	"noonwithmeemisolated":           "noonmeemisolatedarabic",
	"ncong":                          "notapproxequal",
	"nleftrightarrow":                "notarrowboth",
	"nleftarrow":                     "notarrowleft",
	"nrightarrow":                    "notarrowright",
	"nmid":                           "notbar",
	"notdivides":                     "notbar",
	"nni":                            "notcontains",
	"notowner":                       "notcontains",
	"notsuchthat":                    "notcontains",
	"arrowdbllongbothnot":            "notdblarrowboth",
	"nLeftrightarrow":                "notdblarrowboth",
	"notelementof":                   "notelement",
	"notin":                          "notelement",
	"ne":                             "notequal",
	"nexists":                        "notexistential",
	"nVdash":                         "notforces",
	"notforce":                       "notforces",
	"nVDash":                         "notforcesextra",
	"notforceextr":                   "notforcesextra",
	"ngtr":                           "notgreater",
	"ngeq":                           "notgreaternorequal",
	"notgreaterequal":                "notgreaternorequal",
	"notgreaterequal1":               "notgreaternorequal",
	"ngtrless":                       "notgreaternorless",
	"notgreaterless":                 "notgreaternorless",
	"geqslant":                       "notgreaterorslnteql",
	"greaterorequalslant":            "notgreaterorslnteql",
	"nequiv":                         "notidentical",
	"notequivalence":                 "notidentical",
	"nless":                          "notless",
	"nleq":                           "notlessnorequal",
	"notlessequal":                   "notlessnorequal",
	"notlessequal1":                  "notlessnorequal",
	"notbardbl":                      "notparallel",
	"nparallel":                      "notparallel",
	"notpreceeds":                    "notprecedes",
	"nprec":                          "notprecedes",
	"notsatisfy":                     "notsatisfies",
	"nvDash":                         "notsatisfies",
	"nsim":                           "notsimilar",
	"notpropersubset":                "notsubset",
	"nsubset":                        "notsubset",
	"notreflexsubset":                "notsubseteql",
	"nsubseteq":                      "notsubseteql",
	"notfollows":                     "notsucceeds",
	"nsucc":                          "notsucceeds",
	"notpropersuperset":              "notsuperset",
	"nsupset":                        "notsuperset",
	"notreflexsuperset":              "notsuperseteql",
	"nsupseteq":                      "notsuperseteql",
	"nottriangleleftequal":           "nottriangeqlleft",
	"ntrianglelefteq":                "nottriangeqlleft",
	"nottrianglerightequal":          "nottriangeqlright",
	"ntrianglerighteq":               "nottriangeqlright",
	"ntriangleleft":                  "nottriangleleft",
	"ntriangleright":                 "nottriangleright",
	"notturnstileleft":               "notturnstile",
	"nvdash":                         "notturnstile",
	"preceedsnotequal":               "npreccurlyeq",
	"notasymptequal":                 "nsime",
	"notsubsetsqequal":               "nsqsubseteq",
	"notsupersetsqequal":             "nsqsupseteq",
	"followsnotequal":                "nsucccurlyeq",
	"dbar":                           "ntilde",
	"upnu":                           "nu",
	"octothorpe":                     "numbersign",
	"afii61352":                      "numero",
	"afii57680":                      "nun",
	"nunhebrew":                      "nun",
	"nundageshhebrew":                "nundagesh",
	"nunwithdagesh":                  "nundagesh",
	"afii59725":                      "oangthai",
	"circumflexnosp":                 "obarred",
	"afii10080":                      "ocyrillic",
	"ocyril":                         "ocyrillic",
	"rrthook":                        "odblgrave",
	"arrowwaveleft":                  "oharmenian",
	"pipedbl":                        "ohorn",
	"odblacute":                      "ohungarumlaut",
	"exclam1":                        "oi",
	"volintegral":                    "oiiint",
	"volumeintegral":                 "oiiint",
	"surfaceintegral":                "oiint",
	"surfintegral":                   "oiint",
	"cclwcontintegral":               "ointctrclockwise",
	"rfishhookrev":                   "oinvertedbreve",
	"devcon2":                        "olehebrew",
	"upomega":                        "omega",
	"upvarpi":                        "omega1",
	"pisymbolgreek":                  "omega1",
	"Kartdes":                        "omega1",
	"macronnosp":                     "omegalatinclosed",
	"Gebar":                          "omegatonos",
	"upomicron":                      "omicron",
	"onedotlead":                     "onedotenleader",
	"onedotleader":                   "onedotenleader",
	"afii57393":                      "onehackarabic",
	"arabicindicdigitone":            "onehackarabic",
	"onearabic":                      "onehackarabic",
	"onesub":                         "oneinferior",
	"extendedarabicindicdigitone":    "onepersian",
	"afii59761":                      "onethai",
	"epsilon1revhook":                "oogonekmacron",
	"grave1":                         "oopen",
	"original":                       "origof",
	"rightangle":                     "orthogonal",
	"veebar":                         "orunderscore",
	"xor":                            "orunderscore",
	"mturn":                          "oslashacute",
	"ostrokeacute":                   "oslashacute",
	"overbar":                        "overlinecmb",
	"nHdownarrow":                    "pagedown",
	"nHuparrow":                      "pageup",
	"afii59727":                      "paiyannoithai",
	"bardbl2":                        "parallel",
	"vextenddouble":                  "parallel",
	"filledparallelogram":            "parallelogramblack",
	"parenleftBig":                   "parenleft",
	"parenleftBigg":                  "parenleft",
	"parenleftbig":                   "parenleft",
	"parenleftbigg":                  "parenleft",
	"ornateleftparenthesis":          "parenleftaltonearabic",
	"parenleftsub":                   "parenleftinferior",
	"parenrightBig":                  "parenright",
	"parenrightBigg":                 "parenright",
	"parenrightbig":                  "parenright",
	"parenrightbigg":                 "parenright",
	"ornaterightparenthesis":         "parenrightaltonearabic",
	"parenrightsub":                  "parenrightinferior",
	"help":                           "parenrighttp",
	"partial":                        "partialdiff",
	"null":                           "pashtahebrew",
	"patahquarterhebrew":             "patah11",
	"afii57798":                      "patah11",
	"patah2a":                        "patah11",
	"patahwidehebrew":                "patah11",
	"patahhebrew":                    "patah11",
	"patah1d":                        "patah11",
	"patah":                          "patah11",
	"recordseparator":                "patah11",
	"patahnarrowhebrew":              "patah11",
	"backspace":                      "pazerhebrew",
	"afii10081":                      "pecyrillic",
	"pecyril":                        "pecyrillic",
	"pedageshhebrew":                 "pedagesh",
	"pewithdagesh":                   "pedagesh",
	"finalpewithdagesh":              "pefinaldageshhebrew",
	"afii57506":                      "peharabic",
	"peh":                            "peharabic",
	"pehfinal":                       "pehfinalarabic",
	"pehinitial":                     "pehinitialarabic",
	"pehmedial":                      "pehmedialarabic",
	"pewithrafe":                     "perafehebrew",
	"afii57381":                      "percentarabic",
	"arabicpercentsign":              "percentarabic",
	"cdotp":                          "periodcentered",
	"middot":                         "periodcentered",
	"doublebarwedge":                 "perpcorrespond",
	"bot":                            "perpendicular",
	"Pts":                            "peseta",
	"pesetas":                        "peseta",
	"lcedilla1":                      "peso1",
	"pesoph":                         "peso1",
	"upvarphi":                       "phi",
	"phisymbolgreek":                 "phi1",
	"upphi":                          "phi1",
	"zecedilla":                      "phi1",
	"overscorenosp":                  "philatin",
	"afii59738":                      "phinthuthai",
	"Dzhacek":                        "phook",
	"afii59710":                      "phophanthai",
	"afii59708":                      "phophungthai",
	"afii59712":                      "phosamphaothai",
	"uppi":                           "pi",
	"arrowleftnot":                   "piwrarmenian",
	"planckover2pi1":                 "planckover2pi",
	"hslash":                         "planckover2pi",
	"pi1":                            "plusbelowcmb",
	"plussub":                        "plusinferior",
	"pm":                             "plusminus",
	"afii59707":                      "poplathai",
	"precnapprox":                    "precedenotdbleqv",
	"precneqq":                       "precedenotslnteql",
	"preceedsnotsimilar":             "precedeornoteqvlnt",
	"precnsim":                       "precedeornoteqvlnt",
	"prec":                           "precedes",
	"notprecedesoreql":               "precedesequal",
	"preceq":                         "precedesequal",
	"preccurlyeq":                    "precedesorcurly",
	"precedesequal1":                 "precedesorcurly",
	"precedequivlnt":                 "precedesorequal",
	"precsim":                        "precedesorequal",
	"Rx":                             "prescription",
	"backprime":                      "primereversed",
	"minuterev":                      "primereversed",
	"primerev":                       "primereversed",
	"primerev1":                      "primereversed",
	"primereverse":                   "primereversed",
	"prod":                           "product",
	"productdisplay":                 "product",
	"producttext":                    "product",
	"varbarwedge":                    "projective",
	"subset":                         "propersubset",
	"superset":                       "propersuperset",
	"supset":                         "propersuperset",
	"Colon":                          "proportion",
	"propto":                         "proportional",
	"lowerrank":                      "prurel",
	"uppsi":                          "psi",
	"shiftin":                        "qadmahebrew",
	"qaffinal":                       "qaffinalarabic",
	"qafinitial":                     "qafinitialarabic",
	"qafmedial":                      "qafmedialarabic",
	"acknowledge":                    "qarneyparahebrew",
	"circumflexsubnosp":              "qhook",
	"qofdageshhebrew":                "qofdagesh",
	"qofwithdagesh":                  "qofdagesh",
	"afii57687":                      "qofqubutshebrew",
	"qof":                            "qofqubutshebrew",
	"qofhatafpatah":                  "qofqubutshebrew",
	"qofhatafpatahhebrew":            "qofqubutshebrew",
	"qofhatafsegol":                  "qofqubutshebrew",
	"qofhatafsegolhebrew":            "qofqubutshebrew",
	"qofhebrew":                      "qofqubutshebrew",
	"qofhiriq":                       "qofqubutshebrew",
	"qofhiriqhebrew":                 "qofqubutshebrew",
	"qofholam":                       "qofqubutshebrew",
	"qofholamhebrew":                 "qofqubutshebrew",
	"qofpatah":                       "qofqubutshebrew",
	"qofpatahhebrew":                 "qofqubutshebrew",
	"qofqamats":                      "qofqubutshebrew",
	"qofqamatshebrew":                "qofqubutshebrew",
	"qofqubuts":                      "qofqubutshebrew",
	"qofsegol":                       "qofqubutshebrew",
	"qofsegolhebrew":                 "qofqubutshebrew",
	"qofsheva":                       "qofqubutshebrew",
	"qofshevahebrew":                 "qofqubutshebrew",
	"qoftsere":                       "qofqubutshebrew",
	"qoftserehebrew":                 "qofqubutshebrew",
	"afii57796":                      "qubutswidehebrew",
	"blankb":                         "qubutswidehebrew",
	"qibuts":                         "qubutswidehebrew",
	"qubuts":                         "qubutswidehebrew",
	"qubuts18":                       "qubutswidehebrew",
	"qubuts25":                       "qubutswidehebrew",
	"qubuts31":                       "qubutswidehebrew",
	"qubutshebrew":                   "qubutswidehebrew",
	"qubutsnarrowhebrew":             "qubutswidehebrew",
	"qubutsquarterhebrew":            "qubutswidehebrew",
	"questionequal":                  "questeq",
	"quotesinglleft":                 "quoteleft",
	"quoteleftreversed":              "quotereversed",
	"quotesinglrev":                  "quotereversed",
	"quotesinglright":                "quoteright",
	"napostrophe":                    "quoterightn",
	"radicalBig":                     "radical",
	"radicalBigg":                    "radical",
	"radicalbig":                     "radical",
	"radicalbigg":                    "radical",
	"radicalbt":                      "radical",
	"radicaltp":                      "radical",
	"radicalvertex":                  "radical",
	"sqrt":                           "radical",
	"squareroot":                     "radical",
	"mathratio":                      "ratio",
	"rcommaaccent":                   "rcedilla",
	"Rsmallcapinv":                   "rdblgrave",
	"soundcopyright":                 "recordright",
	"refmark":                        "referencemark",
	"subseteq":                       "reflexsubset",
	"subsetorequal":                  "reflexsubset",
	"supersetorequal":                "reflexsuperset",
	"supseteq":                       "reflexsuperset",
	"circleR":                        "registered",
	"afii57425":                      "reharabic",
	"reh":                            "reharabic",
	"rehyehaleflamarabic":            "reharabic",
	"arrownortheast":                 "reharmenian",
	"rehfinal":                       "rehfinalarabic",
	"reshwithdagesh":                 "reshdageshhebrew",
	"afii57688":                      "reshhiriq",
	"resh":                           "reshhiriq",
	"reshhatafpatah":                 "reshhiriq",
	"reshhatafpatahhebrew":           "reshhiriq",
	"reshhatafsegol":                 "reshhiriq",
	"reshhatafsegolhebrew":           "reshhiriq",
	"reshhebrew":                     "reshhiriq",
	"reshhiriqhebrew":                "reshhiriq",
	"reshholam":                      "reshhiriq",
	"reshholamhebrew":                "reshhiriq",
	"reshpatah":                      "reshhiriq",
	"reshpatahhebrew":                "reshhiriq",
	"reshqamats":                     "reshhiriq",
	"reshqamatshebrew":               "reshhiriq",
	"reshqubuts":                     "reshhiriq",
	"reshqubutshebrew":               "reshhiriq",
	"reshsegol":                      "reshhiriq",
	"reshsegolhebrew":                "reshhiriq",
	"reshsheva":                      "reshhiriq",
	"reshshevahebrew":                "reshhiriq",
	"reshtsere":                      "reshhiriq",
	"reshtserehebrew":                "reshhiriq",
	"backsimeq":                      "revasymptequal",
	"backsim":                        "reversedtilde",
	"revsimilar":                     "reversedtilde",
	"tildereversed":                  "reversedtilde",
	"arrowlongbothnot":               "reviamugrashhebrew",
	"reviahebrew":                    "reviamugrashhebrew",
	"invnot":                         "revlogicalnot",
	"logicalnotreversed":             "revlogicalnot",
	"acutedblnosp":                   "rfishhook",
	"haceknosp":                      "rfishhookreversed",
	"uprho":                          "rho",
	"ringnosp":                       "rhook",
	"dieresisnosp":                   "rhookturned",
	"tetse":                          "rhosymbolgreek",
	"upvarrho":                       "rhosymbolgreek",
	"urcorner":                       "rightanglene",
	"ulcorner":                       "rightanglenw",
	"lrcorner":                       "rightanglese",
	"llcorner":                       "rightanglesw",
	"beta1":                          "righttackbelowcmb",
	"varlrtriangle":                  "righttriangle",
	"ocirc":                          "ringcmb",
	"Upsilon1tonos":                  "ringhalfleftbelowcmb",
	"numeralgreeksub":                "ringhalfright",
	"kappa1":                         "ringhalfrightbelowcmb",
	"eqcirc":                         "ringinequal",
	"hooksupnosp":                    "rlongleg",
	"dotnosp":                        "rlonglegturned",
	"afii59715":                      "roruathai",
	"afii57513":                      "rreharabic",
	"blockrighthalf":                 "rtblock",
	"brevenosp":                      "rturned",
	"acuterightnosp":                 "rturnedsuperior",
	"rturnhooksuper":                 "rturnrthooksuper",
	"rupees":                         "rupee",
	"afii59716":                      "ruthai",
	"sadfinal":                       "sadfinalarabic",
	"sadinitial":                     "sadinitialarabic",
	"sadmedial":                      "sadmedialarabic",
	"afii57681":                      "samekh",
	"samekhhebrew":                   "samekh",
	"samekhdagesh":                   "samekhdageshhebrew",
	"samekhwithdagesh":               "samekhdageshhebrew",
	"afii59730":                      "saraaathai",
	"afii59745":                      "saraaethai",
	"afii59748":                      "saraaimaimalaithai",
	"afii59747":                      "saraaimaimuanthai",
	"afii59731":                      "saraamthai",
	"afii59729":                      "saraathai",
	"afii59744":                      "saraethai",
	"afii59733":                      "saraiithai",
	"afii59732":                      "saraithai",
	"afii59746":                      "saraothai",
	"afii59735":                      "saraueethai",
	"afii59734":                      "sarauethai",
	"afii59736":                      "sarauthai",
	"afii59737":                      "sarauuthai",
	"satisfy":                        "satisfies",
	"vDash":                          "satisfies",
	"length":                         "schwa",
	"afii10846":                      "schwacyrillic",
	"halflength":                     "schwahook",
	"higherrank":                     "scurel",
	"dprime":                         "second",
	"primedbl":                       "second",
	"primedbl1":                      "second",
	"seenfinal":                      "seenfinalarabic",
	"seeninitial":                    "seeninitialarabic",
	"seenmedial":                     "seenmedialarabic",
	"afii57795":                      "segolhebrew",
	"groupseparator":                 "segolhebrew",
	"segol":                          "segolhebrew",
	"segol1f":                        "segolhebrew",
	"segol2c":                        "segolhebrew",
	"segol13":                        "segolhebrew",
	"segolnarrowhebrew":              "segolhebrew",
	"segolquarterhebrew":             "segolhebrew",
	"segolwidehebrew":                "segolhebrew",
	"arrowlongboth":                  "seharmenian",
	"sevensub":                       "seveninferior",
	"extendedarabicindicdigitseven":  "sevenpersian",
	"afii59767":                      "seventhai",
	"afii57457":                      "shaddaarabic",
	"shadda":                         "shaddaarabic",
	"shaddafathatanarabic":           "shaddaarabic",
	"shaddawithdammaisolated":        "shaddadammaarabic",
	"shaddawithdammatanisolated":     "shaddadammatanarabic",
	"shaddawithfathaisolated":        "shaddafathaarabic",
	"shaddamedial":                   "shaddahontatweel",
	"shaddawithkasraisolated":        "shaddakasraarabic",
	"shaddawithkasratanisolated":     "shaddakasratanarabic",
	"shaddawithdammalow":             "shaddawithdammaisolatedlow",
	"shaddawithdammatanlow":          "shaddawithdammatanisolatedlow",
	"shaddawithfathaisolatedlow":     "shaddawithfathalow",
	"shaddawithfathatanisolatedlow":  "shaddawithfathatanlow",
	"shaddawithkasralow":             "shaddawithkasraisolatedlow",
	"shaddawithkasratanlow":          "shaddawithkasratanisolatedlow",
	"blockhalfshaded":                "shade",
	"shademedium":                    "shade",
	"blockqtrshaded":                 "shadelight",
	"ltshade":                        "shadelight",
	"sheenfinal":                     "sheenfinalarabic",
	"sheeninitial":                   "sheeninitialarabic",
	"sheenmedial":                    "sheenmedialarabic",
	"pehook":                         "sheicoptic",
	"Lsh":                            "shiftleft",
	"Rsh":                            "shiftright",
	"ustrtbar":                       "shimacoptic",
	"afii57689":                      "shin",
	"shinhebrew":                     "shin",
	"shindageshhebrew":               "shindagesh",
	"shinwithdagesh":                 "shindagesh",
	"shindageshshindothebrew":        "shindageshshindot",
	"shinwithdageshandshindot":       "shindageshshindot",
	"shindageshsindot":               "shindageshsindothebrew",
	"shinwithdageshandsindot":        "shindageshsindothebrew",
	"afii57804":                      "shindothebrew",
	"shindot":                        "shindothebrew",
	"afii57694":                      "shinshindot",
	"shinshindothebrew":              "shinshindot",
	"shinwithshindot":                "shinshindot",
	"gravedblnosp":                   "shook",
	"upsigma":                        "sigma",
	"upvarsigma":                     "sigma1",
	"sigmafinal":                     "sigma1",
	"Chertdes":                       "sigmalunatesymbolgreek",
	"afii57839":                      "siluqlefthebrew",
	"meteg":                          "siluqlefthebrew",
	"newline":                        "siluqlefthebrew",
	"siluqhebrew":                    "siluqlefthebrew",
	"sim":                            "similar",
	"tildemath":                      "similar",
	"tildeoperator":                  "similar",
	"approxnotequal":                 "simneqq",
	"sine":                           "sinewave",
	"sixsub":                         "sixinferior",
	"extendedarabicindicdigitsix":    "sixpersian",
	"afii59766":                      "sixthai",
	"mathslash":                      "slash",
	"slashBig":                       "slash",
	"slashBigg":                      "slash",
	"slashbig":                       "slash",
	"slashbigg":                      "slash",
	"frown":                          "slurabove",
	"smalltriangleleftsld":           "smallblacktriangleleft",
	"smalltrianglerightsld":          "smallblacktriangleright",
	"elementsmall":                   "smallin",
	"smallelement":                   "smallin",
	"ownersmall":                     "smallni",
	"smallcontains":                  "smallni",
	"slurbelow":                      "smile",
	"whitesmilingface":               "smileface",
	"afii57658":                      "sofpasuqhebrew",
	"sofpasuq":                       "sofpasuqhebrew",
	"sfthyphen":                      "softhyphen",
	"afii10094":                      "softsigncyrillic",
	"soft":                           "softsigncyrillic",
	"dei":                            "soliduslongoverlaycmb",
	"negationslash":                  "soliduslongoverlaycmb",
	"not":                            "soliduslongoverlaycmb",
	"Dei":                            "solidusshortoverlaycmb",
	"afii59721":                      "sorusithai",
	"afii59720":                      "sosalathai",
	"afii59691":                      "sosothai",
	"afii59722":                      "sosuathai",
	"spacehackarabic":                "space",
	"a109":                           "spade",
	"spadesuit":                      "spade",
	"spadesuitblack":                 "spade",
	"varspadesuit":                   "spadesuitwhite",
	"sqimageornotequal":              "sqsubsetneq",
	"sqoriginornotequal":             "sqsupsetneq",
	"sigmalunate":                    "squarebelowcmb",
	"boxcrossdiaghatch":              "squarediagonalcrosshatchfill",
	"squarecrossfill":                "squarediagonalcrosshatchfill",
	"boxdot":                         "squaredot",
	"boxhorizhatch":                  "squarehorizontalfill",
	"squarehfill":                    "squarehorizontalfill",
	"sqsubset":                       "squareimage",
	"squareleftsld":                  "squareleftblack",
	"squaresesld":                    "squarelrblack",
	"boxminus":                       "squareminus",
	"boxtimes":                       "squaremultiply",
	"sqsupset":                       "squareoriginal",
	"boxcrosshatch":                  "squareorthogonalcrosshatchfill",
	"squarehvfill":                   "squareorthogonalcrosshatchfill",
	"boxplus":                        "squareplus",
	"squarerightsld":                 "squarerightblack",
	"squarenwsld":                    "squareulblack",
	"boxleftdiaghatch":               "squareupperlefttolowerrightfill",
	"squarenwsefill":                 "squareupperlefttolowerrightfill",
	"boxrtdiaghatch":                 "squareupperrighttolowerleftfill",
	"squareneswfill":                 "squareupperrighttolowerleftfill",
	"boxverthatch":                   "squareverticalfill",
	"squarevfill":                    "squareverticalfill",
	"blackinwhitesquare":             "squarewhitewithsmallblack",
	"boxnested":                      "squarewhitewithsmallblack",
	"leftrightsquigarrow":            "squiggleleftright",
	"arrowsquiggleright":             "squiggleright",
	"rightsquigarrow":                "squiggleright",
	"boxrounded":                     "squoval",
	"starequal":                      "stareq",
	"Subset":                         "subsetdbl",
	"notsubsetordbleql":              "subsetdblequal",
	"subseteqq":                      "subsetdblequal",
	"notsubsetoreql":                 "subsetnotequal",
	"subsetneq":                      "subsetnotequal",
	"subsetnoteql":                   "subsetnotequal",
	"subsetneqq":                     "subsetornotdbleql",
	"sqsubseteq":                     "subsetsqequal",
	"follows":                        "succeeds",
	"succ":                           "succeeds",
	"contains":                       "suchthat",
	"ni":                             "suchthat",
	"owner":                          "suchthat",
	"afii57458":                      "sukunarabic",
	"sukun":                          "sukunarabic",
	"sukunontatweel":                 "sukunmedial",
	"sum":                            "summation",
	"summationdisplay":               "summation",
	"summationtext":                  "summation",
	"compass":                        "sun",
	"Supset":                         "supersetdbl",
	"notsupersetordbleql":            "supersetdblequal",
	"supseteqq":                      "supersetdblequal",
	"notsupersetoreql":               "supersetnotequal",
	"supersetnoteql":                 "supersetnotequal",
	"supsetneq":                      "supersetnotequal",
	"supsetneqq":                     "supersetornotdbleql",
	"sqsupseteq":                     "supersetsqequal",
	"latticetop":                     "tackdown",
	"top":                            "tackdown",
	"dashv":                          "tackleft",
	"turnstileright":                 "tackleft",
	"afii57431":                      "taharabic",
	"tah":                            "taharabic",
	"tahfinal":                       "tahfinalarabic",
	"tahinitial":                     "tahinitialarabic",
	"tahmedial":                      "tahmedialarabic",
	"fathatanontatweel":              "tatweelwithfathatanabove",
	"uptau":                          "tau",
	"tavdages":                       "tavdagesh",
	"tavdageshhebrew":                "tavdagesh",
	"tavwithdagesh":                  "tavdagesh",
	"afii57690":                      "tavhebrew",
	"tav":                            "tavhebrew",
	"tcaronaltone":                   "tcaron1",
	"barmidshortnosp":                "tccurl",
	"tcommaaccent":                   "tcedilla",
	"kcedilla1":                      "tcedilla1",
	"afii57507":                      "tcheharabic",
	"tcheh":                          "tcheharabic",
	"tchehfinal":                     "tchehfinalarabic",
	"tchehinitial":                   "tchehinitialarabic",
	"tchehmeeminitialarabic":         "tchehinitialarabic",
	"tchehmedial":                    "tchehmedialarabic",
	"tehfinal":                       "tehfinalarabic",
	"tehwithhahinitial":              "tehhahinitialarabic",
	"tehinitial":                     "tehinitialarabic",
	"tehwithjeeminitial":             "tehjeeminitialarabic",
	"afii57417":                      "tehmarbutaarabic",
	"tehmarbuta":                     "tehmarbutaarabic",
	"tehmarbutafinal":                "tehmarbutafinalarabic",
	"tehmedial":                      "tehmedialarabic",
	"tehwithmeeminitial":             "tehmeeminitialarabic",
	"tehwithmeemisolated":            "tehmeemisolatedarabic",
	"tehwithnoonfinal":               "tehnoonfinalarabic",
	"tel":                            "telephone",
	"bell":                           "telishagedolahebrew",
	"datalinkescape":                 "telishaqetanahebrew",
	"devcon0":                        "telishaqetanahebrew",
	"tildemidnosp":                   "tesh",
	"tetdageshhebrew":                "tetdagesh",
	"tetwithdagesh":                  "tetdagesh",
	"afii57672":                      "tethebrew",
	"tet":                            "tethebrew",
	"Lcircumflex":                    "tetsecyrillic",
	"starttext":                      "tevirhebrew",
	"tevirlefthebrew":                "tevirhebrew",
	"afii57424":                      "thalarabic",
	"thal":                           "thalarabic",
	"thalfinal":                      "thalfinalarabic",
	"afii59756":                      "thanthakhatthai",
	"afii57419":                      "theharabic",
	"theh":                           "theharabic",
	"thehfinal":                      "thehfinalarabic",
	"thehinitial":                    "thehinitialarabic",
	"thehmedial":                     "thehmedialarabic",
	"uptheta":                        "theta",
	"gehook":                         "theta1",
	"upvartheta":                     "theta1",
	"thetasymbolgreek":               "theta1",
	"afii59697":                      "thonangmonthothai",
	"Ahacek":                         "thook",
	"afii59698":                      "thophuthaothai",
	"afii59703":                      "thothahanthai",
	"afii59696":                      "thothanthai",
	"afii59704":                      "thothongthai",
	"afii59702":                      "thothungthai",
	"thousandsseparatorpersian":      "thousandsseparatorarabic",
	"threesub":                       "threeinferior",
	"extendedarabicindicdigitthree":  "threepersian",
	"afii59763":                      "threethai",
	"tie":                            "tieconcat",
	"tie1":                           "tieconcat",
	"ilde":                           "tilde",
	"tildewide":                      "tilde",
	"tildewider":                     "tilde",
	"tildewidest":                    "tilde",
	"wideutilde":                     "tildebelowcmb",
	"tildecomb":                      "tildecmb",
	"arrowwaveboth":                  "tipehahebrew",
	"tipehalefthebrew":               "tipehahebrew",
	"arrownorthwest":                 "tiwnarmenian",
	"eturn":                          "tonefive",
	"afii59695":                      "topatakthai",
	"toparc":                         "topsemicircle",
	"afii59701":                      "totaothai",
	"commasuprightnosp":              "tretroflexhook",
	"triangledot":                    "trianglecdot",
	"triangleleftsld":                "triangleleftblack",
	"triangleftequal":                "triangleleftequal",
	"trianglelefteq":                 "triangleleftequal",
	"trianglerightsld":               "trianglerightblack",
	"trianglerighteq":                "trianglerightequal",
	"triangrightequal":               "trianglerightequal",
	"primetripl":                     "trprime",
	"primetripl1":                    "trprime",
	"underscoredblnosp":              "ts",
	"tsadidageshhebrew":              "tsadidagesh",
	"tsadiwithdagesh":                "tsadidagesh",
	"afii10088":                      "tsecyrillic",
	"tse":                            "tsecyrillic",
	"afii57794":                      "tsere12",
	"tserenarrowhebrew":              "tsere12",
	"tserehebrew":                    "tsere12",
	"tsere1e":                        "tsere12",
	"tsere":                          "tsere12",
	"tserewidehebrew":                "tsere12",
	"fileseparator":                  "tsere12",
	"tsere2b":                        "tsere12",
	"tserequarterhebrew":             "tsere12",
	"afii10108":                      "tshecyrillic",
	"tshe":                           "tshecyrillic",
	"commasuprevnosp":                "tturned",
	"iotaturn":                       "turnediota",
	"vdash":                          "turnstileleft",
	"afii57394":                      "twoarabic",
	"arabicindicdigittwo":            "twoarabic",
	"twohackarabic":                  "twoarabic",
	"enleadertwodots":                "twodotleader",
	"twodotenleader":                 "twodotleader",
	"twodotlead":                     "twodotleader",
	"twosub":                         "twoinferior",
	"extendedarabicindicdigittwo":    "twopersian",
	"afii59762":                      "twothai",
	"gravesubnosp":                   "ubar",
	"deltaturn":                      "ubreve",
	"uhungarumlaut":                  "udblacute",
	"eshshortrev":                    "udblgrave",
	"Aacutering":                     "udieresiscaron",
	"ihacek":                         "uhorn",
	"tturn":                          "uinvertedbreve",
	"nwquadarc":                      "ularc",
	"dbllowline":                     "underscoredbl",
	"twolowline":                     "underscoredbl",
	"midhorizellipsis":               "unicodecdots",
	"cup":                            "union",
	"Cup":                            "uniondbl",
	"unionmultidisplay":              "unionmulti",
	"unionmultitext":                 "unionmulti",
	"uplus":                          "unionmulti",
	"sqcup":                          "unionsq",
	"unionsqdisplay":                 "unionsq",
	"unionsqtext":                    "unionsq",
	"bigcup":                         "uniontext",
	"naryunion":                      "uniontext",
	"uniondisplay":                   "uniontext",
	"forall":                         "universal",
	"blockuphalf":                    "upblock",
	"gekarev":                        "updigamma",
	"enrtdes":                        "upkoppa",
	"Kavertbar":                      "upoldKoppa",
	"kavertbar":                      "upoldkoppa",
	"enge":                           "upsampi",
	"upupsilon":                      "upsilon",
	"acutesubnosp":                   "upsilonlatin",
	"xsol":                           "upslope",
	"kabar":                          "upstigma",
	"Upsilon1dieresis":               "uptackbelowcmb",
	"Upsilon1diaeresis":              "uptackbelowcmb",
	"Chevertbar":                     "upvarTheta",
	"nequadarc":                      "urarc",
	"Dbar1":                          "utilde",
	"perspcorrespond":                "vardoublebarwedge",
	"clwcontintegral":                "varointclockwise",
	"triangleright":                  "vartriangleleft",
	"triangleleft":                   "vartriangleright",
	"afii57669":                      "vav",
	"vavhebrew":                      "vav",
	"afii57723":                      "vavdageshhebrew",
	"vavdagesh":                      "vavdageshhebrew",
	"vavdagesh65":                    "vavdageshhebrew",
	"vavwithdagesh":                  "vavdageshhebrew",
	"afii57700":                      "vavholam",
	"vavholamhebrew":                 "vavholam",
	"vavwithholam":                   "vavholam",
	"vec":                            "vector",
	"equiangular":                    "veeeq",
	"afii57505":                      "veharabic",
	"veh":                            "veharabic",
	"vehfinal":                       "vehfinalarabic",
	"vehinitial":                     "vehinitialarabic",
	"vehmedial":                      "vehmedialarabic",
	"Sampi":                          "verticallinebelowcmb",
	"arrowlongbothv":                 "vewarmenian",
	"tackleftsubnosp":                "vhook",
	"vertrectangle":                  "vrectangle",
	"filledvertrect":                 "vrectangleblack",
	"tackrightsubnosp":               "vturned",
	"openbullet1":                    "vysmwhtcircle",
	"ringmath":                       "vysmwhtcircle",
	"afii57448":                      "wawarabic",
	"waw":                            "wawarabic",
	"wawfinal":                       "wawfinalarabic",
	"wawwithhamzaabovefinal":         "wawhamzaabovefinalarabic",
	"estimates":                      "wedgeq",
	"Pscript":                        "weierstrass",
	"wp":                             "weierstrass",
	"openbullet":                     "whitebullet",
	"smwhtcircle":                    "whitebullet",
	"circle":                         "whitecircle",
	"mdlgwhtcircle":                  "whitecircle",
	"diamondrhomb":                   "whitediamond",
	"mdlgwhtdiamond":                 "whitediamond",
	"blackinwhitediamond":            "whitediamondcontainingblacksmalldiamond",
	"diamondrhombnested":             "whitediamondcontainingblacksmalldiamond",
	"smalltriangleinv":               "whitedownpointingsmalltriangle",
	"triangledown":                   "whitedownpointingsmalltriangle",
	"bigtriangledown":                "whitedownpointingtriangle",
	"triangleinv":                    "whitedownpointingtriangle",
	"smalltriangleleft":              "whiteleftpointingsmalltriangle",
	"triangleleft1":                  "whiteleftpointingtriangle",
	"triaglfopen":                    "whitepointerleft",
	"triagrtopen":                    "whitepointerright",
	"smalltriangleright":             "whiterightpointingsmalltriangle",
	"triangleright1":                 "whiterightpointingtriangle",
	"H18551":                         "whitesmallsquare",
	"smallbox":                       "whitesmallsquare",
	"smwhtsquare":                    "whitesmallsquare",
	"bigwhitestar":                   "whitestar",
	"smalltriangle":                  "whiteuppointingsmalltriangle",
	"vartriangle":                    "whiteuppointingsmalltriangle",
	"bigtriangleup":                  "whiteuppointingtriangle",
	"triangle":                       "whiteuppointingtriangle",
	"afii59719":                      "wowaenthai",
	"wr":                             "wreathproduct",
	"diaeresistonosnosp":             "wsuperior",
	"anglesupnosp":                   "wturned",
	"upxi":                           "xi",
	"afii59758":                      "yamakkanthai",
	"afii10194":                      "yatcyrillic",
	"Ibar":                           "ycircumflex",
	"afii57450":                      "yeharabic",
	"yeh":                            "yeharabic",
	"afii57519":                      "yehbarreearabic",
	"yehfinal":                       "yehfinalarabic",
	"afii57414":                      "yehhamzaabovearabic",
	"yehwithhamzaabove":              "yehhamzaabovearabic",
	"yehwithhamzaabovefinal":         "yehhamzaabovefinalarabic",
	"yehwithhamzaaboveinitial":       "yehhamzaaboveinitialarabic",
	"yehwithhamzaabovemedial":        "yehhamzaabovemedialarabic",
	"alefmaksurainitialarabic":       "yehinitialarabic",
	"yehinitial":                     "yehinitialarabic",
	"yehwithmeeminitial":             "yehmeeminitialarabic",
	"yehwithmeemisolated":            "yehmeemisolatedarabic",
	"yehwithnoonfinal":               "yehnoonfinalarabic",
	"Yen":                            "yen",
	"auxiliaryon":                    "yerahbenyomohebrew",
	"devcon1":                        "yerahbenyomohebrew",
	"yerahbenyomolefthebrew":         "yerahbenyomohebrew",
	"afii10093":                      "yericyrillic",
	"yeri":                           "yericyrillic",
	"startofhead":                    "yetivhebrew",
	"uhacek":                         "yhook",
	"afii10104":                      "yicyrillic",
	"yi":                             "yicyrillic",
	"arrowsouthwest":                 "yiwnarmenian",
	"yoddagesh":                      "yoddageshhebrew",
	"yodwithdagesh":                  "yoddageshhebrew",
	"afii57718":                      "yodyodhebrew",
	"yoddbl":                         "yodyodhebrew",
	"afii57705":                      "yodyodpatahhebrew",
	"doubleyodpatah":                 "yodyodpatahhebrew",
	"doubleyodpatahhebrew":           "yodyodpatahhebrew",
	"chertdes":                       "yotgreek",
	"afii59714":                      "yoyakthai",
	"afii59693":                      "yoyingthai",
	"dzhacek":                        "yr",
	"iotasubnosp":                    "ysuperior",
	"hornnosp":                       "yturned",
	"afii57432":                      "zaharabic",
	"zah":                            "zaharabic",
	"zahfinal":                       "zahfinalarabic",
	"zahinitial":                     "zahinitialarabic",
	"zahmedial":                      "zahmedialarabic",
	"afii57426":                      "zainarabic",
	"zain":                           "zainarabic",
	"zainfinal":                      "zainfinalarabic",
	"arrowloopright":                 "zaqefgadolhebrew",
	"arrowloopleft":                  "zaqefqatanhebrew",
	"arrowzigzag":                    "zarqahebrew",
	"zayindagesh":                    "zayindageshhebrew",
	"zayinwithdagesh":                "zayindageshhebrew",
	"nleg":                           "zcaron",
	"tackdownsubnosp":                "zcurl",
	"mcapturn":                       "zdotaccent",
	"zdot":                           "zdotaccent",
	"zerodot":                        "zero",
	"zeroslash":                      "zero",
	"afii57392":                      "zerohackarabic",
	"arabicindicdigitzero":           "zerohackarabic",
	"zeroarabic":                     "zerohackarabic",
	"zerosub":                        "zeroinferior",
	"extendedarabicindicdigitzero":   "zeropersian",
	"afii59760":                      "zerothai",
	"bom":                            "zerowidthjoiner",
	"zerowidthnobreakspace":          "zerowidthjoiner",
	"zerospace":                      "zerowidthspace",
	"upzeta":                         "zeta",
	"afii10072":                      "zhecyrillic",
	"zhe":                            "zhecyrillic",
	"negacknowledge":                 "zinorhebrew",
	"tackupsubnosp":                  "zretroflexhook",
}

var glyphlistGlyphToRuneMap = map[string]rune{ // 6339 entries
	".notdef":                             0xfffd,  // � '\ufffd'
	"250a":                                0x250a,  // ┊ '\u250a'
	"250b":                                0x250b,  // ┋ '\u250b'
	"250d":                                0x250d,  // ┍ '\u250d'
	"250e":                                0x250e,  // ┎ '\u250e'
	"250f":                                0x250f,  // ┏ '\u250f'
	"251a":                                0x251a,  // ┚ '\u251a'
	"251b":                                0x251b,  // ┛ '\u251b'
	"251d":                                0x251d,  // ┝ '\u251d'
	"251e":                                0x251e,  // ┞ '\u251e'
	"251f":                                0x251f,  // ┟ '\u251f'
	"252a":                                0x252a,  // ┪ '\u252a'
	"252b":                                0x252b,  // ┫ '\u252b'
	"252d":                                0x252d,  // ┭ '\u252d'
	"252e":                                0x252e,  // ┮ '\u252e'
	"252f":                                0x252f,  // ┯ '\u252f'
	"253a":                                0x253a,  // ┺ '\u253a'
	"253b":                                0x253b,  // ┻ '\u253b'
	"253d":                                0x253d,  // ┽ '\u253d'
	"253e":                                0x253e,  // ┾ '\u253e'
	"253f":                                0x253f,  // ┿ '\u253f'
	"254a":                                0x254a,  // ╊ '\u254a'
	"254b":                                0x254b,  // ╋ '\u254b'
	"254c":                                0x254c,  // ╌ '\u254c'
	"254d":                                0x254d,  // ╍ '\u254d'
	"254e":                                0x254e,  // ╎ '\u254e'
	"254f":                                0x254f,  // ╏ '\u254f'
	"256d":                                0x256d,  // ╭ '\u256d'
	"256e":                                0x256e,  // ╮ '\u256e'
	"256f":                                0x256f,  // ╯ '\u256f'
	"257a":                                0x257a,  // ╺ '\u257a'
	"257b":                                0x257b,  // ╻ '\u257b'
	"257c":                                0x257c,  // ╼ '\u257c'
	"257d":                                0x257d,  // ╽ '\u257d'
	"257e":                                0x257e,  // ╾ '\u257e'
	"257f":                                0x257f,  // ╿ '\u257f'
	"A":                                   0x0041,  // A 'A'
	"AE":                                  0x00c6,  // Æ '\u00c6'
	"AEacute":                             0x01fc,  // Ǽ '\u01fc'
	"AEmacron":                            0x01e2,  // Ǣ '\u01e2'
	"AEsmall":                             0xf7e6,  //  '\uf7e6'
	"APLboxquestion":                      0x2370,  // ⍰ '\u2370'
	"APLboxupcaret":                       0x2353,  // ⍓ '\u2353'
	"APLnotbackslash":                     0x2340,  // ⍀ '\u2340'
	"APLnotslash":                         0x233f,  // ⌿ '\u233f'
	"Aacute":                              0x00c1,  // Á '\u00c1'
	"Aacutesmall":                         0xf7e1,  //  '\uf7e1'
	"Abreve":                              0x0102,  // Ă '\u0102'
	"Abreveacute":                         0x1eae,  // Ắ '\u1eae'
	"Abrevecyrillic":                      0x04d0,  // Ӑ '\u04d0'
	"Abrevedotbelow":                      0x1eb6,  // Ặ '\u1eb6'
	"Abrevegrave":                         0x1eb0,  // Ằ '\u1eb0'
	"Abrevehookabove":                     0x1eb2,  // Ẳ '\u1eb2'
	"Abrevetilde":                         0x1eb4,  // Ẵ '\u1eb4'
	"Acaron":                              0x01cd,  // Ǎ '\u01cd'
	"Acircle":                             0x24b6,  // Ⓐ '\u24b6'
	"Acircumflex":                         0x00c2,  // Â '\u00c2'
	"Acircumflexacute":                    0x1ea4,  // Ấ '\u1ea4'
	"Acircumflexdotbelow":                 0x1eac,  // Ậ '\u1eac'
	"Acircumflexgrave":                    0x1ea6,  // Ầ '\u1ea6'
	"Acircumflexhookabove":                0x1ea8,  // Ẩ '\u1ea8'
	"Acircumflexsmall":                    0xf7e2,  //  '\uf7e2'
	"Acircumflextilde":                    0x1eaa,  // Ẫ '\u1eaa'
	"Acute":                               0xf6c9,  //  '\uf6c9'
	"Acutesmall":                          0xf7b4,  //  '\uf7b4'
	"Adblgrave":                           0x0200,  // Ȁ '\u0200'
	"Adieresis":                           0x00c4,  // Ä '\u00c4'
	"Adieresiscyrillic":                   0x04d2,  // Ӓ '\u04d2'
	"Adieresismacron":                     0x01de,  // Ǟ '\u01de'
	"Adieresissmall":                      0xf7e4,  //  '\uf7e4'
	"Adotbelow":                           0x1ea0,  // Ạ '\u1ea0'
	"Adotmacron":                          0x01e0,  // Ǡ '\u01e0'
	"Agrave":                              0x00c0,  // À '\u00c0'
	"Agravesmall":                         0xf7e0,  //  '\uf7e0'
	"Ahookabove":                          0x1ea2,  // Ả '\u1ea2'
	"Aiecyrillic":                         0x04d4,  // Ӕ '\u04d4'
	"Ainvertedbreve":                      0x0202,  // Ȃ '\u0202'
	"Alpha":                               0x0391,  // Α '\u0391'
	"Alphatonos":                          0x0386,  // Ά '\u0386'
	"Amacron":                             0x0100,  // Ā '\u0100'
	"Amonospace":                          0xff21,  // Ａ '\uff21'
	"Aogonek":                             0x0104,  // Ą '\u0104'
	"Aring":                               0x00c5,  // Å '\u00c5'
	"Aringacute":                          0x01fa,  // Ǻ '\u01fa'
	"Aringbelow":                          0x1e00,  // Ḁ '\u1e00'
	"Aringsmall":                          0xf7e5,  //  '\uf7e5'
	"Asmall":                              0xf761,  //  '\uf761'
	"Atilde":                              0x00c3,  // Ã '\u00c3'
	"Atildesmall":                         0xf7e3,  //  '\uf7e3'
	"Aybarmenian":                         0x0531,  // Ա '\u0531'
	"B":                                   0x0042,  // B 'B'
	"Barv":                                0x2ae7,  // ⫧ '\u2ae7'
	"BbbA":                                0x1d538, // 𝔸 '\U0001d538'
	"BbbB":                                0x1d539, // 𝔹 '\U0001d539'
	"BbbC":                                0x2102,  // ℂ '\u2102'
	"BbbD":                                0x1d53b, // 𝔻 '\U0001d53b'
	"BbbE":                                0x1d53c, // 𝔼 '\U0001d53c'
	"BbbF":                                0x1d53d, // 𝔽 '\U0001d53d'
	"BbbG":                                0x1d53e, // 𝔾 '\U0001d53e'
	"BbbGamma":                            0x213e,  // ℾ '\u213e'
	"BbbH":                                0x210d,  // ℍ '\u210d'
	"BbbI":                                0x1d540, // 𝕀 '\U0001d540'
	"BbbJ":                                0x1d541, // 𝕁 '\U0001d541'
	"BbbK":                                0x1d542, // 𝕂 '\U0001d542'
	"BbbL":                                0x1d543, // 𝕃 '\U0001d543'
	"BbbM":                                0x1d544, // 𝕄 '\U0001d544'
	"BbbN":                                0x2115,  // ℕ '\u2115'
	"BbbO":                                0x1d546, // 𝕆 '\U0001d546'
	"BbbP":                                0x2119,  // ℙ '\u2119'
	"BbbPi":                               0x213f,  // ℿ '\u213f'
	"BbbQ":                                0x211a,  // ℚ '\u211a'
	"BbbR":                                0x211d,  // ℝ '\u211d'
	"BbbS":                                0x1d54a, // 𝕊 '\U0001d54a'
	"BbbT":                                0x1d54b, // 𝕋 '\U0001d54b'
	"BbbU":                                0x1d54c, // 𝕌 '\U0001d54c'
	"BbbV":                                0x1d54d, // 𝕍 '\U0001d54d'
	"BbbW":                                0x1d54e, // 𝕎 '\U0001d54e'
	"BbbX":                                0x1d54f, // 𝕏 '\U0001d54f'
	"BbbY":                                0x1d550, // 𝕐 '\U0001d550'
	"BbbZ":                                0x2124,  // ℤ '\u2124'
	"Bbba":                                0x1d552, // 𝕒 '\U0001d552'
	"Bbbb":                                0x1d553, // 𝕓 '\U0001d553'
	"Bbbc":                                0x1d554, // 𝕔 '\U0001d554'
	"Bbbd":                                0x1d555, // 𝕕 '\U0001d555'
	"Bbbe":                                0x1d556, // 𝕖 '\U0001d556'
	"Bbbeight":                            0x1d7e0, // 𝟠 '\U0001d7e0'
	"Bbbf":                                0x1d557, // 𝕗 '\U0001d557'
	"Bbbfive":                             0x1d7dd, // 𝟝 '\U0001d7dd'
	"Bbbfour":                             0x1d7dc, // 𝟜 '\U0001d7dc'
	"Bbbg":                                0x1d558, // 𝕘 '\U0001d558'
	"Bbbgamma":                            0x213d,  // ℽ '\u213d'
	"Bbbh":                                0x1d559, // 𝕙 '\U0001d559'
	"Bbbi":                                0x1d55a, // 𝕚 '\U0001d55a'
	"Bbbj":                                0x1d55b, // 𝕛 '\U0001d55b'
	"Bbbk":                                0x1d55c, // 𝕜 '\U0001d55c'
	"Bbbl":                                0x1d55d, // 𝕝 '\U0001d55d'
	"Bbbm":                                0x1d55e, // 𝕞 '\U0001d55e'
	"Bbbn":                                0x1d55f, // 𝕟 '\U0001d55f'
	"Bbbnine":                             0x1d7e1, // 𝟡 '\U0001d7e1'
	"Bbbo":                                0x1d560, // 𝕠 '\U0001d560'
	"Bbbone":                              0x1d7d9, // 𝟙 '\U0001d7d9'
	"Bbbp":                                0x1d561, // 𝕡 '\U0001d561'
	"Bbbpi":                               0x213c,  // ℼ '\u213c'
	"Bbbq":                                0x1d562, // 𝕢 '\U0001d562'
	"Bbbr":                                0x1d563, // 𝕣 '\U0001d563'
	"Bbbs":                                0x1d564, // 𝕤 '\U0001d564'
	"Bbbseven":                            0x1d7df, // 𝟟 '\U0001d7df'
	"Bbbsix":                              0x1d7de, // 𝟞 '\U0001d7de'
	"Bbbsum":                              0x2140,  // ⅀ '\u2140'
	"Bbbt":                                0x1d565, // 𝕥 '\U0001d565'
	"Bbbthree":                            0x1d7db, // 𝟛 '\U0001d7db'
	"Bbbtwo":                              0x1d7da, // 𝟚 '\U0001d7da'
	"Bbbu":                                0x1d566, // 𝕦 '\U0001d566'
	"Bbbv":                                0x1d567, // 𝕧 '\U0001d567'
	"Bbbw":                                0x1d568, // 𝕨 '\U0001d568'
	"Bbbx":                                0x1d569, // 𝕩 '\U0001d569'
	"Bbby":                                0x1d56a, // 𝕪 '\U0001d56a'
	"Bbbz":                                0x1d56b, // 𝕫 '\U0001d56b'
	"Bbbzero":                             0x1d7d8, // 𝟘 '\U0001d7d8'
	"Bcircle":                             0x24b7,  // Ⓑ '\u24b7'
	"Bdotaccent":                          0x1e02,  // Ḃ '\u1e02'
	"Bdotbelow":                           0x1e04,  // Ḅ '\u1e04'
	"Benarmenian":                         0x0532,  // Բ '\u0532'
	"Beta":                                0x0392,  // Β '\u0392'
	"Bhook":                               0x0181,  // Ɓ '\u0181'
	"Blinebelow":                          0x1e06,  // Ḇ '\u1e06'
	"Bmonospace":                          0xff22,  // Ｂ '\uff22'
	"Brevesmall":                          0xf6f4,  //  '\uf6f4'
	"Bsmall":                              0xf762,  //  '\uf762'
	"Bsmallcap":                           0x0229,  // ȩ '\u0229'
	"Btopbar":                             0x0182,  // Ƃ '\u0182'
	"C":                                   0x0043,  // C 'C'
	"Caarmenian":                          0x053e,  // Ծ '\u053e'
	"Cacute":                              0x0106,  // Ć '\u0106'
	"Caron":                               0xf6ca,  //  '\uf6ca'
	"Caronsmall":                          0xf6f5,  //  '\uf6f5'
	"Ccaron":                              0x010c,  // Č '\u010c'
	"Ccedilla":                            0x00c7,  // Ç '\u00c7'
	"Ccedillaacute":                       0x1e08,  // Ḉ '\u1e08'
	"Ccedillasmall":                       0xf7e7,  //  '\uf7e7'
	"Ccircle":                             0x24b8,  // Ⓒ '\u24b8'
	"Ccircumflex":                         0x0108,  // Ĉ '\u0108'
	"Cdotaccent":                          0x010a,  // Ċ '\u010a'
	"Cedillasmall":                        0xf7b8,  //  '\uf7b8'
	"Chaarmenian":                         0x0549,  // Չ '\u0549'
	"Cheabkhasiancyrillic":                0x04bc,  // Ҽ '\u04bc'
	"Checyrillic":                         0x0427,  // Ч '\u0427'
	"Chedescenderabkhasiancyrillic":       0x04be,  // Ҿ '\u04be'
	"Chedescendercyrillic":                0x04b6,  // Ҷ '\u04b6'
	"Chedieresiscyrillic":                 0x04f4,  // Ӵ '\u04f4'
	"Cheharmenian":                        0x0543,  // Ճ '\u0543'
	"Chekhakassiancyrillic":               0x04cb,  // Ӌ '\u04cb'
	"Cheverticalstrokecyrillic":           0x04b8,  // Ҹ '\u04b8'
	"Chi":                                 0x03a7,  // Χ '\u03a7'
	"Chook":                               0x0187,  // Ƈ '\u0187'
	"Circumflexsmall":                     0xf6f6,  //  '\uf6f6'
	"Cmonospace":                          0xff23,  // Ｃ '\uff23'
	"Coarmenian":                          0x0551,  // Ց '\u0551'
	"Coloneq":                             0x2a74,  // ⩴ '\u2a74'
	"Csmall":                              0xf763,  //  '\uf763'
	"D":                                   0x0044,  // D 'D'
	"DDownarrow":                          0x27f1,  // ⟱ '\u27f1'
	"DZ":                                  0x01f1,  // Ǳ '\u01f1'
	"DZcaron":                             0x01c4,  // Ǆ '\u01c4'
	"Daarmenian":                          0x0534,  // Դ '\u0534'
	"Dafrican":                            0x0189,  // Ɖ '\u0189'
	"DashV":                               0x2ae5,  // ⫥ '\u2ae5'
	"DashVDash":                           0x27da,  // ⟚ '\u27da'
	"Dashv":                               0x2ae4,  // ⫤ '\u2ae4'
	"Dcaron":                              0x010e,  // Ď '\u010e'
	"Dcaron1":                             0xf810,  //  '\uf810'
	"Dcedilla":                            0x1e10,  // Ḑ '\u1e10'
	"Dcircle":                             0x24b9,  // Ⓓ '\u24b9'
	"Dcircumflexbelow":                    0x1e12,  // Ḓ '\u1e12'
	"Dcroat":                              0x0110,  // Đ '\u0110'
	"Ddotaccent":                          0x1e0a,  // Ḋ '\u1e0a'
	"Ddotbelow":                           0x1e0c,  // Ḍ '\u1e0c'
	"Ddownarrow":                          0x290b,  // ⤋ '\u290b'
	"Decyrillic":                          0x0414,  // Д '\u0414'
	"Deicoptic":                           0x03ee,  // Ϯ '\u03ee'
	"Delta":                               0x2206,  // ∆ '\u2206'
	"Deltagreek":                          0x0394,  // Δ '\u0394'
	"Dhook":                               0x018a,  // Ɗ '\u018a'
	"Dieresis":                            0xf6cb,  //  '\uf6cb'
	"DieresisAcute":                       0xf6cc,  //  '\uf6cc'
	"DieresisGrave":                       0xf6cd,  //  '\uf6cd'
	"Dieresissmall":                       0xf7a8,  //  '\uf7a8'
	"Digamma":                             0x1d7cb, // 𝟋 '\U0001d7cb'
	"Digammagreek":                        0x03dc,  // Ϝ '\u03dc'
	"Dlinebelow":                          0x1e0e,  // Ḏ '\u1e0e'
	"Dmonospace":                          0xff24,  // Ｄ '\uff24'
	"Dotaccentsmall":                      0xf6f7,  //  '\uf6f7'
	"Dsmall":                              0xf764,  //  '\uf764'
	"Dtopbar":                             0x018b,  // Ƌ '\u018b'
	"Dz":                                  0x01f2,  // ǲ '\u01f2'
	"Dzcaron":                             0x01c5,  // ǅ '\u01c5'
	"Dzeabkhasiancyrillic":                0x04e0,  // Ӡ '\u04e0'
	"Dzhecyrillic":                        0x040f,  // Џ '\u040f'
	"E":                                   0x0045,  // E 'E'
	"Eacute":                              0x00c9,  // É '\u00c9'
	"Eacutesmall":                         0xf7e9,  //  '\uf7e9'
	"Ebreve":                              0x0114,  // Ĕ '\u0114'
	"Ecaron":                              0x011a,  // Ě '\u011a'
	"Ecedillabreve":                       0x1e1c,  // Ḝ '\u1e1c'
	"Echarmenian":                         0x0535,  // Ե '\u0535'
	"Ecircle":                             0x24ba,  // Ⓔ '\u24ba'
	"Ecircumflex":                         0x00ca,  // Ê '\u00ca'
	"Ecircumflexacute":                    0x1ebe,  // Ế '\u1ebe'
	"Ecircumflexbelow":                    0x1e18,  // Ḙ '\u1e18'
	"Ecircumflexdotbelow":                 0x1ec6,  // Ệ '\u1ec6'
	"Ecircumflexgrave":                    0x1ec0,  // Ề '\u1ec0'
	"Ecircumflexhookabove":                0x1ec2,  // Ể '\u1ec2'
	"Ecircumflexsmall":                    0xf7ea,  //  '\uf7ea'
	"Ecircumflextilde":                    0x1ec4,  // Ễ '\u1ec4'
	"Ecyrillic":                           0x0404,  // Є '\u0404'
	"Edblgrave":                           0x0204,  // Ȅ '\u0204'
	"Edieresis":                           0x00cb,  // Ë '\u00cb'
	"Edieresissmall":                      0xf7eb,  //  '\uf7eb'
	"Edotaccent":                          0x0116,  // Ė '\u0116'
	"Edotbelow":                           0x1eb8,  // Ẹ '\u1eb8'
	"Egrave":                              0x00c8,  // È '\u00c8'
	"Egravesmall":                         0xf7e8,  //  '\uf7e8'
	"Eharmenian":                          0x0537,  // Է '\u0537'
	"Ehookabove":                          0x1eba,  // Ẻ '\u1eba'
	"Eightroman":                          0x2167,  // Ⅷ '\u2167'
	"Einvertedbreve":                      0x0206,  // Ȇ '\u0206'
	"Eiotifiedcyrillic":                   0x0464,  // Ѥ '\u0464'
	"Elcyrillic":                          0x041b,  // Л '\u041b'
	"Elevenroman":                         0x216a,  // Ⅺ '\u216a'
	"Emacron":                             0x0112,  // Ē '\u0112'
	"Emacronacute":                        0x1e16,  // Ḗ '\u1e16'
	"Emacrongrave":                        0x1e14,  // Ḕ '\u1e14'
	"Emcyrillic":                          0x041c,  // М '\u041c'
	"Emonospace":                          0xff25,  // Ｅ '\uff25'
	"Endescendercyrillic":                 0x04a2,  // Ң '\u04a2'
	"Eng":                                 0x014a,  // Ŋ '\u014a'
	"Enghecyrillic":                       0x04a4,  // Ҥ '\u04a4'
	"Enhookcyrillic":                      0x04c7,  // Ӈ '\u04c7'
	"Eogonek":                             0x0118,  // Ę '\u0118'
	"Eopen":                               0x0190,  // Ɛ '\u0190'
	"Epsilon":                             0x0395,  // Ε '\u0395'
	"Epsilontonos":                        0x0388,  // Έ '\u0388'
	"Equiv":                               0x2263,  // ≣ '\u2263'
	"Ereversed":                           0x018e,  // Ǝ '\u018e'
	"Ereversedcyrillic":                   0x042d,  // Э '\u042d'
	"Esdescendercyrillic":                 0x04aa,  // Ҫ '\u04aa'
	"Esh":                                 0x01a9,  // Ʃ '\u01a9'
	"Esmall":                              0xf765,  //  '\uf765'
	"Eta":                                 0x0397,  // Η '\u0397'
	"Etarmenian":                          0x0538,  // Ը '\u0538'
	"Etatonos":                            0x0389,  // Ή '\u0389'
	"Eth":                                 0x00d0,  // Ð '\u00d0'
	"Ethsmall":                            0xf7f0,  //  '\uf7f0'
	"Etilde":                              0x1ebc,  // Ẽ '\u1ebc'
	"Etildebelow":                         0x1e1a,  // Ḛ '\u1e1a'
	"Eulerconst":                          0x2107,  // ℇ '\u2107'
	"Euro":                                0x20ac,  // € '\u20ac'
	"Ezh":                                 0x01b7,  // Ʒ '\u01b7'
	"Ezhcaron":                            0x01ee,  // Ǯ '\u01ee'
	"Ezhreversed":                         0x01b8,  // Ƹ '\u01b8'
	"F":                                   0x0046,  // F 'F'
	"Fcircle":                             0x24bb,  // Ⓕ '\u24bb'
	"Fdotaccent":                          0x1e1e,  // Ḟ '\u1e1e'
	"Feharmenian":                         0x0556,  // Ֆ '\u0556'
	"Feicoptic":                           0x03e4,  // Ϥ '\u03e4'
	"Fhook":                               0x0191,  // Ƒ '\u0191'
	"Finv":                                0x2132,  // Ⅎ '\u2132'
	"Fiveroman":                           0x2164,  // Ⅴ '\u2164'
	"Fmonospace":                          0xff26,  // Ｆ '\uff26'
	"Fourroman":                           0x2163,  // Ⅳ '\u2163'
	"Fsmall":                              0xf766,  //  '\uf766'
	"G":                                   0x0047,  // G 'G'
	"GBsquare":                            0x3387,  // ㎇ '\u3387'
	"Gacute":                              0x01f4,  // Ǵ '\u01f4'
	"Gamma":                               0x0393,  // Γ '\u0393'
	"Gammaafrican":                        0x0194,  // Ɣ '\u0194'
	"Gangiacoptic":                        0x03ea,  // Ϫ '\u03ea'
	"Gbreve":                              0x011e,  // Ğ '\u011e'
	"Gcaron":                              0x01e6,  // Ǧ '\u01e6'
	"Gcircle":                             0x24bc,  // Ⓖ '\u24bc'
	"Gcircumflex":                         0x011c,  // Ĝ '\u011c'
	"Gcommaaccent":                        0x0122,  // Ģ '\u0122'
	"Gdotaccent":                          0x0120,  // Ġ '\u0120'
	"Gecyrillic":                          0x0413,  // Г '\u0413'
	"Ghadarmenian":                        0x0542,  // Ղ '\u0542'
	"Ghemiddlehookcyrillic":               0x0494,  // Ҕ '\u0494'
	"Ghestrokecyrillic":                   0x0492,  // Ғ '\u0492'
	"Gheupturncyrillic":                   0x0490,  // Ґ '\u0490'
	"Ghook":                               0x0193,  // Ɠ '\u0193'
	"Gimarmenian":                         0x0533,  // Գ '\u0533'
	"Gmacron":                             0x1e20,  // Ḡ '\u1e20'
	"Gmir":                                0x2141,  // ⅁ '\u2141'
	"Gmonospace":                          0xff27,  // Ｇ '\uff27'
	"Grave":                               0xf6ce,  //  '\uf6ce'
	"Gravesmall":                          0xf760,  //  '\uf760'
	"Gsmall":                              0xf767,  //  '\uf767'
	"Gsmallcaphook":                       0x022b,  // ȫ '\u022b'
	"Gsmallhook":                          0x029b,  // ʛ '\u029b'
	"Gstroke":                             0x01e4,  // Ǥ '\u01e4'
	"Gt":                                  0x2aa2,  // ⪢ '\u2aa2'
	"H":                                   0x0048,  // H 'H'
	"H22073":                              0x25a1,  // □ '\u25a1'
	"HPsquare":                            0x33cb,  // ㏋ '\u33cb'
	"Haabkhasiancyrillic":                 0x04a8,  // Ҩ '\u04a8'
	"Hadescendercyrillic":                 0x04b2,  // Ҳ '\u04b2'
	"Hbar":                                0x0126,  // Ħ '\u0126'
	"Hbrevebelow":                         0x1e2a,  // Ḫ '\u1e2a'
	"Hcedilla":                            0x1e28,  // Ḩ '\u1e28'
	"Hcircle":                             0x24bd,  // Ⓗ '\u24bd'
	"Hcircumflex":                         0x0124,  // Ĥ '\u0124'
	"Hdieresis":                           0x1e26,  // Ḧ '\u1e26'
	"Hdotaccent":                          0x1e22,  // Ḣ '\u1e22'
	"Hdotbelow":                           0x1e24,  // Ḥ '\u1e24'
	"Hermaphrodite":                       0x26a5,  // ⚥ '\u26a5'
	"Hmonospace":                          0xff28,  // Ｈ '\uff28'
	"Hoarmenian":                          0x0540,  // Հ '\u0540'
	"Horicoptic":                          0x03e8,  // Ϩ '\u03e8'
	"Hsmall":                              0xf768,  //  '\uf768'
	"Hsmallcap":                           0x022c,  // Ȭ '\u022c'
	"Hungarumlaut":                        0xf6cf,  //  '\uf6cf'
	"Hungarumlautsmall":                   0xf6f8,  //  '\uf6f8'
	"Hzsquare":                            0x3390,  // ㎐ '\u3390'
	"I":                                   0x0049,  // I 'I'
	"IJ":                                  0x0132,  // Ĳ '\u0132'
	"Iacute":                              0x00cd,  // Í '\u00cd'
	"Iacutesmall":                         0xf7ed,  //  '\uf7ed'
	"Ibreve":                              0x012c,  // Ĭ '\u012c'
	"Icaron":                              0x01cf,  // Ǐ '\u01cf'
	"Icircle":                             0x24be,  // Ⓘ '\u24be'
	"Icircumflex":                         0x00ce,  // Î '\u00ce'
	"Icircumflexsmall":                    0xf7ee,  //  '\uf7ee'
	"Icyril1":                             0x03fc,  // ϼ '\u03fc'
	"Idblgrave":                           0x0208,  // Ȉ '\u0208'
	"Idieresis":                           0x00cf,  // Ï '\u00cf'
	"Idieresisacute":                      0x1e2e,  // Ḯ '\u1e2e'
	"Idieresiscyrillic":                   0x04e4,  // Ӥ '\u04e4'
	"Idieresissmall":                      0xf7ef,  //  '\uf7ef'
	"Idot":                                0x0130,  // İ '\u0130'
	"Idotbelow":                           0x1eca,  // Ị '\u1eca'
	"Iebrevecyrillic":                     0x04d6,  // Ӗ '\u04d6'
	"Iecyrillic":                          0x0415,  // Е '\u0415'
	"Iehook":                              0x03f8,  // ϸ '\u03f8'
	"Iehookogonek":                        0x03fa,  // Ϻ '\u03fa'
	"Ifraktur":                            0x2111,  // ℑ '\u2111'
	"Igrave":                              0x00cc,  // Ì '\u00cc'
	"Igravesmall":                         0xf7ec,  //  '\uf7ec'
	"Ihookabove":                          0x1ec8,  // Ỉ '\u1ec8'
	"Iicyrillic":                          0x0418,  // И '\u0418'
	"Iinvertedbreve":                      0x020a,  // Ȋ '\u020a'
	"Imacron":                             0x012a,  // Ī '\u012a'
	"Imacroncyrillic":                     0x04e2,  // Ӣ '\u04e2'
	"Imonospace":                          0xff29,  // Ｉ '\uff29'
	"Iniarmenian":                         0x053b,  // Ի '\u053b'
	"Iocyrillic":                          0x0401,  // Ё '\u0401'
	"Iogonek":                             0x012e,  // Į '\u012e'
	"Iota":                                0x0399,  // Ι '\u0399'
	"Iotaafrican":                         0x0196,  // Ɩ '\u0196'
	"Iotadiaeresis":                       0x02f3,  // ˳ '\u02f3'
	"Iotadieresis":                        0x03aa,  // Ϊ '\u03aa'
	"Iotatonos":                           0x038a,  // Ί '\u038a'
	"Ismall":                              0xf769,  //  '\uf769'
	"Istroke":                             0x0197,  // Ɨ '\u0197'
	"Itilde":                              0x0128,  // Ĩ '\u0128'
	"Itildebelow":                         0x1e2c,  // Ḭ '\u1e2c'
	"Izhitsadblgravecyrillic":             0x0476,  // Ѷ '\u0476'
	"J":                                   0x004a,  // J 'J'
	"Jaarmenian":                          0x0541,  // Ձ '\u0541'
	"Jcircle":                             0x24bf,  // Ⓙ '\u24bf'
	"Jcircumflex":                         0x0134,  // Ĵ '\u0134'
	"Jheharmenian":                        0x054b,  // Ջ '\u054b'
	"Jmonospace":                          0xff2a,  // Ｊ '\uff2a'
	"Join":                                0x2a1d,  // ⨝ '\u2a1d'
	"Jsmall":                              0xf76a,  //  '\uf76a'
	"K":                                   0x004b,  // K 'K'
	"KBsquare":                            0x3385,  // ㎅ '\u3385'
	"KKsquare":                            0x33cd,  // ㏍ '\u33cd'
	"Kabashkircyrillic":                   0x04a0,  // Ҡ '\u04a0'
	"Kacute":                              0x1e30,  // Ḱ '\u1e30'
	"Kadescendercyrillic":                 0x049a,  // Қ '\u049a'
	"Kahook":                              0x03ff,  // Ͽ '\u03ff'
	"Kahookcyrillic":                      0x04c3,  // Ӄ '\u04c3'
	"Kappa":                               0x039a,  // Κ '\u039a'
	"Kastrokecyrillic":                    0x049e,  // Ҟ '\u049e'
	"Kaverticalstrokecyrillic":            0x049c,  // Ҝ '\u049c'
	"Kcaron":                              0x01e8,  // Ǩ '\u01e8'
	"Kcedilla":                            0x0136,  // Ķ '\u0136'
	"Kcircle":                             0x24c0,  // Ⓚ '\u24c0'
	"Kdotbelow":                           0x1e32,  // Ḳ '\u1e32'
	"Keharmenian":                         0x0554,  // Ք '\u0554'
	"Kenarmenian":                         0x053f,  // Կ '\u053f'
	"Khacyrillic":                         0x0425,  // Х '\u0425'
	"Kheicoptic":                          0x03e6,  // Ϧ '\u03e6'
	"Khook":                               0x0198,  // Ƙ '\u0198'
	"Kjecyrillic":                         0x040c,  // Ќ '\u040c'
	"Klinebelow":                          0x1e34,  // Ḵ '\u1e34'
	"Kmonospace":                          0xff2b,  // Ｋ '\uff2b'
	"Koppacyrillic":                       0x0480,  // Ҁ '\u0480'
	"Koppagreek":                          0x03de,  // Ϟ '\u03de'
	"Ksicyrillic":                         0x046e,  // Ѯ '\u046e'
	"Ksmall":                              0xf76b,  //  '\uf76b'
	"L":                                   0x004c,  // L 'L'
	"LJ":                                  0x01c7,  // Ǉ '\u01c7'
	"LL":                                  0xf6bf,  //  '\uf6bf'
	"LLeftarrow":                          0x2b45,  // ⭅ '\u2b45'
	"Lacute":                              0x0139,  // Ĺ '\u0139'
	"Lambda":                              0x039b,  // Λ '\u039b'
	"Lbrbrak":                             0x27ec,  // ⟬ '\u27ec'
	"Lcaron":                              0x013d,  // Ľ '\u013d'
	"Lcaron1":                             0xf812,  //  '\uf812'
	"Lcedilla":                            0x013b,  // Ļ '\u013b'
	"Lcedilla1":                           0xf81a,  //  '\uf81a'
	"Lcircle":                             0x24c1,  // Ⓛ '\u24c1'
	"Lcircumflexbelow":                    0x1e3c,  // Ḽ '\u1e3c'
	"Ldotaccent":                          0x013f,  // Ŀ '\u013f'
	"Ldotbelow":                           0x1e36,  // Ḷ '\u1e36'
	"Ldotbelowmacron":                     0x1e38,  // Ḹ '\u1e38'
	"Ldsh":                                0x21b2,  // ↲ '\u21b2'
	"Liwnarmenian":                        0x053c,  // Լ '\u053c'
	"Lj":                                  0x01c8,  // ǈ '\u01c8'
	"Ljecyrillic":                         0x0409,  // Љ '\u0409'
	"Llinebelow":                          0x1e3a,  // Ḻ '\u1e3a'
	"Lmonospace":                          0xff2c,  // Ｌ '\uff2c'
	"Longleftarrow":                       0x27f8,  // ⟸ '\u27f8'
	"Longleftrightarrow":                  0x27fa,  // ⟺ '\u27fa'
	"Longmapsfrom":                        0x27fd,  // ⟽ '\u27fd'
	"Longmapsto":                          0x27fe,  // ⟾ '\u27fe'
	"Longrightarrow":                      0x27f9,  // ⟹ '\u27f9'
	"Lparengtr":                           0x2995,  // ⦕ '\u2995'
	"Lslash":                              0x0141,  // Ł '\u0141'
	"Lslashsmall":                         0xf6f9,  //  '\uf6f9'
	"Lsmall":                              0xf76c,  //  '\uf76c'
	"Lsmallcap":                           0x022f,  // ȯ '\u022f'
	"Lt":                                  0x2aa1,  // ⪡ '\u2aa1'
	"Lvzigzag":                            0x29da,  // ⧚ '\u29da'
	"M":                                   0x004d,  // M 'M'
	"MBsquare":                            0x3386,  // ㎆ '\u3386'
	"Macron":                              0xf6d0,  //  '\uf6d0'
	"Macronsmall":                         0xf7af,  //  '\uf7af'
	"Macute":                              0x1e3e,  // Ḿ '\u1e3e'
	"Mapsfrom":                            0x2906,  // ⤆ '\u2906'
	"Mapsto":                              0x2907,  // ⤇ '\u2907'
	"Mcircle":                             0x24c2,  // Ⓜ '\u24c2'
	"Mdotaccent":                          0x1e40,  // Ṁ '\u1e40'
	"Mdotbelow":                           0x1e42,  // Ṃ '\u1e42'
	"Menarmenian":                         0x0544,  // Մ '\u0544'
	"Mmonospace":                          0xff2d,  // Ｍ '\uff2d'
	"Msmall":                              0xf76d,  //  '\uf76d'
	"Mturned":                             0x019c,  // Ɯ '\u019c'
	"Mu":                                  0x039c,  // Μ '\u039c'
	"N":                                   0x004e,  // N 'N'
	"NJ":                                  0x01ca,  // Ǌ '\u01ca'
	"Nacute":                              0x0143,  // Ń '\u0143'
	"Ncaron":                              0x0147,  // Ň '\u0147'
	"Ncedilla1":                           0xf81c,  //  '\uf81c'
	"Ncircle":                             0x24c3,  // Ⓝ '\u24c3'
	"Ncircumflexbelow":                    0x1e4a,  // Ṋ '\u1e4a'
	"Ncommaaccent":                        0x0145,  // Ņ '\u0145'
	"Ndotaccent":                          0x1e44,  // Ṅ '\u1e44'
	"Ndotbelow":                           0x1e46,  // Ṇ '\u1e46'
	"Nearrow":                             0x21d7,  // ⇗ '\u21d7'
	"Nhookleft":                           0x019d,  // Ɲ '\u019d'
	"Nineroman":                           0x2168,  // Ⅸ '\u2168'
	"Nj":                                  0x01cb,  // ǋ '\u01cb'
	"Nlinebelow":                          0x1e48,  // Ṉ '\u1e48'
	"Nmonospace":                          0xff2e,  // Ｎ '\uff2e'
	"Not":                                 0x2aec,  // ⫬ '\u2aec'
	"Nowarmenian":                         0x0546,  // Ն '\u0546'
	"Nsmall":                              0xf76e,  //  '\uf76e'
	"Ntilde":                              0x00d1,  // Ñ '\u00d1'
	"Ntildesmall":                         0xf7f1,  //  '\uf7f1'
	"Nu":                                  0x039d,  // Ν '\u039d'
	"Nwarrow":                             0x21d6,  // ⇖ '\u21d6'
	"O":                                   0x004f,  // O 'O'
	"OE":                                  0x0152,  // Œ '\u0152'
	"OEsmall":                             0xf6fa,  //  '\uf6fa'
	"Oacute":                              0x00d3,  // Ó '\u00d3'
	"Oacutesmall":                         0xf7f3,  //  '\uf7f3'
	"Obarredcyrillic":                     0x04e8,  // Ө '\u04e8'
	"Obarreddieresiscyrillic":             0x04ea,  // Ӫ '\u04ea'
	"Obreve":                              0x014e,  // Ŏ '\u014e'
	"Ocaron":                              0x01d1,  // Ǒ '\u01d1'
	"Ocenteredtilde":                      0x019f,  // Ɵ '\u019f'
	"Ocircle":                             0x24c4,  // Ⓞ '\u24c4'
	"Ocircumflex":                         0x00d4,  // Ô '\u00d4'
	"Ocircumflexacute":                    0x1ed0,  // Ố '\u1ed0'
	"Ocircumflexdotbelow":                 0x1ed8,  // Ộ '\u1ed8'
	"Ocircumflexgrave":                    0x1ed2,  // Ồ '\u1ed2'
	"Ocircumflexhookabove":                0x1ed4,  // Ổ '\u1ed4'
	"Ocircumflexsmall":                    0xf7f4,  //  '\uf7f4'
	"Ocircumflextilde":                    0x1ed6,  // Ỗ '\u1ed6'
	"Ocyrillic":                           0x041e,  // О '\u041e'
	"Odblacute":                           0x0150,  // Ő '\u0150'
	"Odblgrave":                           0x020c,  // Ȍ '\u020c'
	"Odieresis":                           0x00d6,  // Ö '\u00d6'
	"Odieresiscyrillic":                   0x04e6,  // Ӧ '\u04e6'
	"Odieresissmall":                      0xf7f6,  //  '\uf7f6'
	"Odotbelow":                           0x1ecc,  // Ọ '\u1ecc'
	"Ogoneksmall":                         0xf6fb,  //  '\uf6fb'
	"Ograve":                              0x00d2,  // Ò '\u00d2'
	"Ogravesmall":                         0xf7f2,  //  '\uf7f2'
	"Oharmenian":                          0x0555,  // Օ '\u0555'
	"Ohookabove":                          0x1ece,  // Ỏ '\u1ece'
	"Ohorn":                               0x01a0,  // Ơ '\u01a0'
	"Ohornacute":                          0x1eda,  // Ớ '\u1eda'
	"Ohorndotbelow":                       0x1ee2,  // Ợ '\u1ee2'
	"Ohorngrave":                          0x1edc,  // Ờ '\u1edc'
	"Ohornhookabove":                      0x1ede,  // Ở '\u1ede'
	"Ohorntilde":                          0x1ee0,  // Ỡ '\u1ee0'
	"Oi":                                  0x01a2,  // Ƣ '\u01a2'
	"Oinvertedbreve":                      0x020e,  // Ȏ '\u020e'
	"Omacron":                             0x014c,  // Ō '\u014c'
	"Omacronacute":                        0x1e52,  // Ṓ '\u1e52'
	"Omacrongrave":                        0x1e50,  // Ṑ '\u1e50'
	"Omega":                               0x2126,  // Ω '\u2126'
	"Omegacyrillic":                       0x0460,  // Ѡ '\u0460'
	"Omegagreek":                          0x03a9,  // Ω '\u03a9'
	"Omegainv":                            0x2127,  // ℧ '\u2127'
	"Omegaroundcyrillic":                  0x047a,  // Ѻ '\u047a'
	"Omegatitlocyrillic":                  0x047c,  // Ѽ '\u047c'
	"Omegatonos":                          0x038f,  // Ώ '\u038f'
	"Omicron":                             0x039f,  // Ο '\u039f'
	"Omicrontonos":                        0x038c,  // Ό '\u038c'
	"Omonospace":                          0xff2f,  // Ｏ '\uff2f'
	"Oneroman":                            0x2160,  // Ⅰ '\u2160'
	"Oogonek":                             0x01ea,  // Ǫ '\u01ea'
	"Oogonekmacron":                       0x01ec,  // Ǭ '\u01ec'
	"Oopen":                               0x0186,  // Ɔ '\u0186'
	"Oslash":                              0x00d8,  // Ø '\u00d8'
	"Oslashacute":                         0x01fe,  // Ǿ '\u01fe'
	"Oslashsmall":                         0xf7f8,  //  '\uf7f8'
	"Osmall":                              0xf76f,  //  '\uf76f'
	"Otcyrillic":                          0x047e,  // Ѿ '\u047e'
	"Otilde":                              0x00d5,  // Õ '\u00d5'
	"Otildeacute":                         0x1e4c,  // Ṍ '\u1e4c'
	"Otildedieresis":                      0x1e4e,  // Ṏ '\u1e4e'
	"Otildesmall":                         0xf7f5,  //  '\uf7f5'
	"Otimes":                              0x2a37,  // ⨷ '\u2a37'
	"P":                                   0x0050,  // P 'P'
	"Pacute":                              0x1e54,  // Ṕ '\u1e54'
	"Pcircle":                             0x24c5,  // Ⓟ '\u24c5'
	"Pdotaccent":                          0x1e56,  // Ṗ '\u1e56'
	"Peharmenian":                         0x054a,  // Պ '\u054a'
	"Pemiddlehookcyrillic":                0x04a6,  // Ҧ '\u04a6'
	"Phi":                                 0x03a6,  // Φ '\u03a6'
	"Phook":                               0x01a4,  // Ƥ '\u01a4'
	"Pi":                                  0x03a0,  // Π '\u03a0'
	"Piwrarmenian":                        0x0553,  // Փ '\u0553'
	"Planckconst":                         0x210e,  // ℎ '\u210e'
	"Pmonospace":                          0xff30,  // Ｐ '\uff30'
	"Prec":                                0x2abb,  // ⪻ '\u2abb'
	"PropertyLine":                        0x214a,  // ⅊ '\u214a'
	"Psi":                                 0x03a8,  // Ψ '\u03a8'
	"Psicyrillic":                         0x0470,  // Ѱ '\u0470'
	"Psmall":                              0xf770,  //  '\uf770'
	"Q":                                   0x0051,  // Q 'Q'
	"QED":                                 0x220e,  // ∎ '\u220e'
	"Qcircle":                             0x24c6,  // Ⓠ '\u24c6'
	"Qmonospace":                          0xff31,  // Ｑ '\uff31'
	"Qsmall":                              0xf771,  //  '\uf771'
	"Question":                            0x2047,  // ⁇ '\u2047'
	"R":                                   0x0052,  // R 'R'
	"RRightarrow":                         0x2b46,  // ⭆ '\u2b46'
	"Raarmenian":                          0x054c,  // Ռ '\u054c'
	"Racute":                              0x0154,  // Ŕ '\u0154'
	"Rbrbrak":                             0x27ed,  // ⟭ '\u27ed'
	"Rcaron":                              0x0158,  // Ř '\u0158'
	"Rcedilla":                            0x0156,  // Ŗ '\u0156'
	"Rcedilla1":                           0xf81e,  //  '\uf81e'
	"Rcircle":                             0x24c7,  // Ⓡ '\u24c7'
	"Rcircumflex":                         0xf831,  //  '\uf831'
	"Rdblgrave":                           0x0210,  // Ȑ '\u0210'
	"Rdotaccent":                          0x1e58,  // Ṙ '\u1e58'
	"Rdotbelow":                           0x1e5a,  // Ṛ '\u1e5a'
	"Rdotbelowmacron":                     0x1e5c,  // Ṝ '\u1e5c'
	"Rdsh":                                0x21b3,  // ↳ '\u21b3'
	"Reharmenian":                         0x0550,  // Ր '\u0550'
	"Rfraktur":                            0x211c,  // ℜ '\u211c'
	"Rho":                                 0x03a1,  // Ρ '\u03a1'
	"Ringsmall":                           0xf6fc,  //  '\uf6fc'
	"Rinvertedbreve":                      0x0212,  // Ȓ '\u0212'
	"Rlinebelow":                          0x1e5e,  // Ṟ '\u1e5e'
	"Rmonospace":                          0xff32,  // Ｒ '\uff32'
	"Rparenless":                          0x2996,  // ⦖ '\u2996'
	"Rsmall":                              0xf772,  //  '\uf772'
	"Rsmallinverted":                      0x0281,  // ʁ '\u0281'
	"Rsmallinvertedsuperior":              0x02b6,  // ʶ '\u02b6'
	"Rturnsuper":                          0x023f,  // ȿ '\u023f'
	"Rvzigzag":                            0x29db,  // ⧛ '\u29db'
	"S":                                   0x0053,  // S 'S'
	"SD150100":                            0x024f,  // ɏ '\u024f'
	"SF010000":                            0x250c,  // ┌ '\u250c'
	"SF020000":                            0x2514,  // └ '\u2514'
	"SF030000":                            0x2510,  // ┐ '\u2510'
	"SF040000":                            0x2518,  // ┘ '\u2518'
	"SF050000":                            0x253c,  // ┼ '\u253c'
	"SF060000":                            0x252c,  // ┬ '\u252c'
	"SF070000":                            0x2534,  // ┴ '\u2534'
	"SF080000":                            0x251c,  // ├ '\u251c'
	"SF090000":                            0x2524,  // ┤ '\u2524'
	"SF100000":                            0x2500,  // ─ '\u2500'
	"SF110000":                            0x2502,  // │ '\u2502'
	"SF190000":                            0x2561,  // ╡ '\u2561'
	"SF200000":                            0x2562,  // ╢ '\u2562'
	"SF210000":                            0x2556,  // ╖ '\u2556'
	"SF220000":                            0x2555,  // ╕ '\u2555'
	"SF230000":                            0x2563,  // ╣ '\u2563'
	"SF240000":                            0x2551,  // ║ '\u2551'
	"SF250000":                            0x2557,  // ╗ '\u2557'
	"SF260000":                            0x255d,  // ╝ '\u255d'
	"SF270000":                            0x255c,  // ╜ '\u255c'
	"SF280000":                            0x255b,  // ╛ '\u255b'
	"SF360000":                            0x255e,  // ╞ '\u255e'
	"SF370000":                            0x255f,  // ╟ '\u255f'
	"SF380000":                            0x255a,  // ╚ '\u255a'
	"SF390000":                            0x2554,  // ╔ '\u2554'
	"SF400000":                            0x2569,  // ╩ '\u2569'
	"SF410000":                            0x2566,  // ╦ '\u2566'
	"SF420000":                            0x2560,  // ╠ '\u2560'
	"SF430000":                            0x2550,  // ═ '\u2550'
	"SF440000":                            0x256c,  // ╬ '\u256c'
	"SF450000":                            0x2567,  // ╧ '\u2567'
	"SF460000":                            0x2568,  // ╨ '\u2568'
	"SF470000":                            0x2564,  // ╤ '\u2564'
	"SF480000":                            0x2565,  // ╥ '\u2565'
	"SF490000":                            0x2559,  // ╙ '\u2559'
	"SF500000":                            0x2558,  // ╘ '\u2558'
	"SF510000":                            0x2552,  // ╒ '\u2552'
	"SF520000":                            0x2553,  // ╓ '\u2553'
	"SF530000":                            0x256b,  // ╫ '\u256b'
	"SF540000":                            0x256a,  // ╪ '\u256a'
	"Sacute":                              0x015a,  // Ś '\u015a'
	"Sacutedotaccent":                     0x1e64,  // Ṥ '\u1e64'
	"Sampigreek":                          0x03e0,  // Ϡ '\u03e0'
	"Scaron":                              0x0160,  // Š '\u0160'
	"Scarondotaccent":                     0x1e66,  // Ṧ '\u1e66'
	"Scaronsmall":                         0xf6fd,  //  '\uf6fd'
	"Scedilla":                            0x015e,  // Ş '\u015e'
	"Scedilla1":                           0xf816,  //  '\uf816'
	"Schwa":                               0x018f,  // Ə '\u018f'
	"Schwacyrillic":                       0x04d8,  // Ә '\u04d8'
	"Schwadieresiscyrillic":               0x04da,  // Ӛ '\u04da'
	"Scircle":                             0x24c8,  // Ⓢ '\u24c8'
	"Scircumflex":                         0x015c,  // Ŝ '\u015c'
	"Scommaaccent":                        0x0218,  // Ș '\u0218'
	"Sdotaccent":                          0x1e60,  // Ṡ '\u1e60'
	"Sdotbelow":                           0x1e62,  // Ṣ '\u1e62'
	"Sdotbelowdotaccent":                  0x1e68,  // Ṩ '\u1e68'
	"Searrow":                             0x21d8,  // ⇘ '\u21d8'
	"Seharmenian":                         0x054d,  // Ս '\u054d'
	"Sevenroman":                          0x2166,  // Ⅶ '\u2166'
	"Shaarmenian":                         0x0547,  // Շ '\u0547'
	"Shacyrillic":                         0x0428,  // Ш '\u0428'
	"Sheicoptic":                          0x03e2,  // Ϣ '\u03e2'
	"Shhacyrillic":                        0x04ba,  // Һ '\u04ba'
	"Shimacoptic":                         0x03ec,  // Ϭ '\u03ec'
	"Sigma":                               0x03a3,  // Σ '\u03a3'
	"Sixroman":                            0x2165,  // Ⅵ '\u2165'
	"Smonospace":                          0xff33,  // Ｓ '\uff33'
	"Sqcap":                               0x2a4e,  // ⩎ '\u2a4e'
	"Sqcup":                               0x2a4f,  // ⩏ '\u2a4f'
	"Ssmall":                              0xf773,  //  '\uf773'
	"Stigmagreek":                         0x03da,  // Ϛ '\u03da'
	"Succ":                                0x2abc,  // ⪼ '\u2abc'
	"Swarrow":                             0x21d9,  // ⇙ '\u21d9'
	"T":                                   0x0054,  // T 'T'
	"Tau":                                 0x03a4,  // Τ '\u03a4'
	"Tbar":                                0x0166,  // Ŧ '\u0166'
	"Tcaron":                              0x0164,  // Ť '\u0164'
	"Tcaron1":                             0xf814,  //  '\uf814'
	"Tcedilla1":                           0xf818,  //  '\uf818'
	"Tcircle":                             0x24c9,  // Ⓣ '\u24c9'
	"Tcircumflexbelow":                    0x1e70,  // Ṱ '\u1e70'
	"Tcommaaccent":                        0x0162,  // Ţ '\u0162'
	"Tdotaccent":                          0x1e6a,  // Ṫ '\u1e6a'
	"Tdotbelow":                           0x1e6c,  // Ṭ '\u1e6c'
	"Tedescendercyrillic":                 0x04ac,  // Ҭ '\u04ac'
	"Tenroman":                            0x2169,  // Ⅹ '\u2169'
	"Tetsecyrillic":                       0x04b4,  // Ҵ '\u04b4'
	"Theta":                               0x0398,  // Θ '\u0398'
	"Thook":                               0x01ac,  // Ƭ '\u01ac'
	"Thorn":                               0x00de,  // Þ '\u00de'
	"Thornsmall":                          0xf7fe,  //  '\uf7fe'
	"Threeroman":                          0x2162,  // Ⅲ '\u2162'
	"Tildesmall":                          0xf6fe,  //  '\uf6fe'
	"Tiwnarmenian":                        0x054f,  // Տ '\u054f'
	"Tlinebelow":                          0x1e6e,  // Ṯ '\u1e6e'
	"Tmonospace":                          0xff34,  // Ｔ '\uff34'
	"Toarmenian":                          0x0539,  // Թ '\u0539'
	"Tonefive":                            0x01bc,  // Ƽ '\u01bc'
	"Tonesix":                             0x0184,  // Ƅ '\u0184'
	"Tonetwo":                             0x01a7,  // Ƨ '\u01a7'
	"Tretroflexhook":                      0x01ae,  // Ʈ '\u01ae'
	"Tsecyrillic":                         0x0426,  // Ц '\u0426'
	"Tshecyrillic":                        0x040b,  // Ћ '\u040b'
	"Tsmall":                              0xf774,  //  '\uf774'
	"Twelveroman":                         0x216b,  // Ⅻ '\u216b'
	"Tworoman":                            0x2161,  // Ⅱ '\u2161'
	"U":                                   0x0055,  // U 'U'
	"UUparrow":                            0x27f0,  // ⟰ '\u27f0'
	"Uacute":                              0x00da,  // Ú '\u00da'
	"Uacutesmall":                         0xf7fa,  //  '\uf7fa'
	"Ubreve":                              0x016c,  // Ŭ '\u016c'
	"Ucaron":                              0x01d3,  // Ǔ '\u01d3'
	"Ucedilla":                            0xf833,  //  '\uf833'
	"Ucircle":                             0x24ca,  // Ⓤ '\u24ca'
	"Ucircumflex":                         0x00db,  // Û '\u00db'
	"Ucircumflexbelow":                    0x1e76,  // Ṷ '\u1e76'
	"Ucircumflexsmall":                    0xf7fb,  //  '\uf7fb'
	"Ucyrillic":                           0x0423,  // У '\u0423'
	"Udblgrave":                           0x0214,  // Ȕ '\u0214'
	"Udieresis":                           0x00dc,  // Ü '\u00dc'
	"Udieresisacute":                      0x01d7,  // Ǘ '\u01d7'
	"Udieresisbelow":                      0x1e72,  // Ṳ '\u1e72'
	"Udieresiscaron":                      0x01d9,  // Ǚ '\u01d9'
	"Udieresiscyrillic":                   0x04f0,  // Ӱ '\u04f0'
	"Udieresisgrave":                      0x01db,  // Ǜ '\u01db'
	"Udieresismacron":                     0x01d5,  // Ǖ '\u01d5'
	"Udieresissmall":                      0xf7fc,  //  '\uf7fc'
	"Udotbelow":                           0x1ee4,  // Ụ '\u1ee4'
	"Ugrave":                              0x00d9,  // Ù '\u00d9'
	"Ugravesmall":                         0xf7f9,  //  '\uf7f9'
	"Uhookabove":                          0x1ee6,  // Ủ '\u1ee6'
	"Uhorn":                               0x01af,  // Ư '\u01af'
	"Uhornacute":                          0x1ee8,  // Ứ '\u1ee8'
	"Uhorndotbelow":                       0x1ef0,  // Ự '\u1ef0'
	"Uhorngrave":                          0x1eea,  // Ừ '\u1eea'
	"Uhornhookabove":                      0x1eec,  // Ử '\u1eec'
	"Uhorntilde":                          0x1eee,  // Ữ '\u1eee'
	"Uhungarumlaut":                       0x0170,  // Ű '\u0170'
	"Uhungarumlautcyrillic":               0x04f2,  // Ӳ '\u04f2'
	"Uinvertedbreve":                      0x0216,  // Ȗ '\u0216'
	"Ukcyrillic":                          0x0478,  // Ѹ '\u0478'
	"Umacron":                             0x016a,  // Ū '\u016a'
	"Umacroncyrillic":                     0x04ee,  // Ӯ '\u04ee'
	"Umacrondieresis":                     0x1e7a,  // Ṻ '\u1e7a'
	"Umonospace":                          0xff35,  // Ｕ '\uff35'
	"Uogonek":                             0x0172,  // Ų '\u0172'
	"Upsilon":                             0x03a5,  // Υ '\u03a5'
	"Upsilon1":                            0x03d2,  // ϒ '\u03d2'
	"Upsilonacutehooksymbolgreek":         0x03d3,  // ϓ '\u03d3'
	"Upsilonafrican":                      0x01b1,  // Ʊ '\u01b1'
	"Upsilondiaeresis":                    0x02f4,  // ˴ '\u02f4'
	"Upsilondieresis":                     0x03ab,  // Ϋ '\u03ab'
	"Upsilondieresishooksymbolgreek":      0x03d4,  // ϔ '\u03d4'
	"Upsilontonos":                        0x038e,  // Ύ '\u038e'
	"Uring":                               0x016e,  // Ů '\u016e'
	"Ushortcyrillic":                      0x040e,  // Ў '\u040e'
	"Usmall":                              0xf775,  //  '\uf775'
	"Ustraightcyrillic":                   0x04ae,  // Ү '\u04ae'
	"Ustraightstrokecyrillic":             0x04b0,  // Ұ '\u04b0'
	"Utilde":                              0x0168,  // Ũ '\u0168'
	"Utildeacute":                         0x1e78,  // Ṹ '\u1e78'
	"Utildebelow":                         0x1e74,  // Ṵ '\u1e74'
	"Uuparrow":                            0x290a,  // ⤊ '\u290a'
	"V":                                   0x0056,  // V 'V'
	"VDash":                               0x22ab,  // ⊫ '\u22ab'
	"Vbar":                                0x2aeb,  // ⫫ '\u2aeb'
	"Vcircle":                             0x24cb,  // Ⓥ '\u24cb'
	"Vdotbelow":                           0x1e7e,  // Ṿ '\u1e7e'
	"Vee":                                 0x2a54,  // ⩔ '\u2a54'
	"Vewarmenian":                         0x054e,  // Վ '\u054e'
	"Vhook":                               0x01b2,  // Ʋ '\u01b2'
	"Vmonospace":                          0xff36,  // Ｖ '\uff36'
	"Voarmenian":                          0x0548,  // Ո '\u0548'
	"Vsmall":                              0xf776,  //  '\uf776'
	"Vtilde":                              0x1e7c,  // Ṽ '\u1e7c'
	"Vvert":                               0x2980,  // ⦀ '\u2980'
	"W":                                   0x0057,  // W 'W'
	"Wacute":                              0x1e82,  // Ẃ '\u1e82'
	"Wcircle":                             0x24cc,  // Ⓦ '\u24cc'
	"Wcircumflex":                         0x0174,  // Ŵ '\u0174'
	"Wdieresis":                           0x1e84,  // Ẅ '\u1e84'
	"Wdotaccent":                          0x1e86,  // Ẇ '\u1e86'
	"Wdotbelow":                           0x1e88,  // Ẉ '\u1e88'
	"Wedge":                               0x2a53,  // ⩓ '\u2a53'
	"Wgrave":                              0x1e80,  // Ẁ '\u1e80'
	"Wmonospace":                          0xff37,  // Ｗ '\uff37'
	"Wsmall":                              0xf777,  //  '\uf777'
	"X":                                   0x0058,  // X 'X'
	"Xcircle":                             0x24cd,  // Ⓧ '\u24cd'
	"Xdieresis":                           0x1e8c,  // Ẍ '\u1e8c'
	"Xdotaccent":                          0x1e8a,  // Ẋ '\u1e8a'
	"Xeharmenian":                         0x053d,  // Խ '\u053d'
	"Xi":                                  0x039e,  // Ξ '\u039e'
	"Xmonospace":                          0xff38,  // Ｘ '\uff38'
	"Xsmall":                              0xf778,  //  '\uf778'
	"Y":                                   0x0059,  // Y 'Y'
	"Yacute":                              0x00dd,  // Ý '\u00dd'
	"Yacutesmall":                         0xf7fd,  //  '\uf7fd'
	"Ycircle":                             0x24ce,  // Ⓨ '\u24ce'
	"Ycircumflex":                         0x0176,  // Ŷ '\u0176'
	"Ydieresis":                           0x0178,  // Ÿ '\u0178'
	"Ydieresissmall":                      0xf7ff,  //  '\uf7ff'
	"Ydotaccent":                          0x1e8e,  // Ẏ '\u1e8e'
	"Ydotbelow":                           0x1ef4,  // Ỵ '\u1ef4'
	"Yerudieresiscyrillic":                0x04f8,  // Ӹ '\u04f8'
	"Ygrave":                              0x1ef2,  // Ỳ '\u1ef2'
	"Yhook":                               0x01b3,  // Ƴ '\u01b3'
	"Yhookabove":                          0x1ef6,  // Ỷ '\u1ef6'
	"Yiarmenian":                          0x0545,  // Յ '\u0545'
	"Yicyrillic":                          0x0407,  // Ї '\u0407'
	"Yiwnarmenian":                        0x0552,  // Ւ '\u0552'
	"Ymonospace":                          0xff39,  // Ｙ '\uff39'
	"Ysmall":                              0xf779,  //  '\uf779'
	"Ysmallcap":                           0x021f,  // ȟ '\u021f'
	"Ytilde":                              0x1ef8,  // Ỹ '\u1ef8'
	"Yup":                                 0x2144,  // ⅄ '\u2144'
	"Yusbigcyrillic":                      0x046a,  // Ѫ '\u046a'
	"Yusbigiotifiedcyrillic":              0x046c,  // Ѭ '\u046c'
	"Yuslittlecyrillic":                   0x0466,  // Ѧ '\u0466'
	"Yuslittleiotifiedcyrillic":           0x0468,  // Ѩ '\u0468'
	"Z":                                   0x005a,  // Z 'Z'
	"Zaarmenian":                          0x0536,  // Զ '\u0536'
	"Zacute":                              0x0179,  // Ź '\u0179'
	"Zcaron":                              0x017d,  // Ž '\u017d'
	"Zcaronsmall":                         0xf6ff,  //  '\uf6ff'
	"Zcircle":                             0x24cf,  // Ⓩ '\u24cf'
	"Zcircumflex":                         0x1e90,  // Ẑ '\u1e90'
	"Zdotaccent":                          0x017b,  // Ż '\u017b'
	"Zdotbelow":                           0x1e92,  // Ẓ '\u1e92'
	"Zedescendercyrillic":                 0x0498,  // Ҙ '\u0498'
	"Zedieresiscyrillic":                  0x04de,  // Ӟ '\u04de'
	"Zeta":                                0x0396,  // Ζ '\u0396'
	"Zhearmenian":                         0x053a,  // Ժ '\u053a'
	"Zhebreve":                            0x03fd,  // Ͻ '\u03fd'
	"Zhebrevecyrillic":                    0x04c1,  // Ӂ '\u04c1'
	"Zhedescendercyrillic":                0x0496,  // Җ '\u0496'
	"Zhedieresiscyrillic":                 0x04dc,  // Ӝ '\u04dc'
	"Zlinebelow":                          0x1e94,  // Ẕ '\u1e94'
	"Zmonospace":                          0xff3a,  // Ｚ '\uff3a'
	"Zsmall":                              0xf77a,  //  '\uf77a'
	"Zstroke":                             0x01b5,  // Ƶ '\u01b5'
	"a":                                   0x0061,  // a 'a'
	"a1":                                  0x2701,  // ✁ '\u2701'
	"a2":                                  0x2702,  // ✂ '\u2702'
	"a3":                                  0x2704,  // ✄ '\u2704'
	"a4":                                  0x260e,  // ☎ '\u260e'
	"a5":                                  0x2706,  // ✆ '\u2706'
	"a6":                                  0x271d,  // ✝ '\u271d'
	"a7":                                  0x271e,  // ✞ '\u271e'
	"a8":                                  0x271f,  // ✟ '\u271f'
	"a9":                                  0x2720,  // ✠ '\u2720'
	"a10":                                 0x2721,  // ✡ '\u2721'
	"a11":                                 0x261b,  // ☛ '\u261b'
	"a12":                                 0x261e,  // ☞ '\u261e'
	"a13":                                 0x270c,  // ✌ '\u270c'
	"a14":                                 0x270d,  // ✍ '\u270d'
	"a15":                                 0x270e,  // ✎ '\u270e'
	"a16":                                 0x270f,  // ✏ '\u270f'
	"a17":                                 0x2711,  // ✑ '\u2711'
	"a18":                                 0x2712,  // ✒ '\u2712'
	"a19":                                 0x2713,  // ✓ '\u2713'
	"a20":                                 0x2714,  // ✔ '\u2714'
	"a21":                                 0x2715,  // ✕ '\u2715'
	"a22":                                 0x2716,  // ✖ '\u2716'
	"a23":                                 0x2717,  // ✗ '\u2717'
	"a24":                                 0x2718,  // ✘ '\u2718'
	"a25":                                 0x2719,  // ✙ '\u2719'
	"a26":                                 0x271a,  // ✚ '\u271a'
	"a27":                                 0x271b,  // ✛ '\u271b'
	"a28":                                 0x271c,  // ✜ '\u271c'
	"a29":                                 0x2722,  // ✢ '\u2722'
	"a30":                                 0x2723,  // ✣ '\u2723'
	"a31":                                 0x2724,  // ✤ '\u2724'
	"a32":                                 0x2725,  // ✥ '\u2725'
	"a33":                                 0x2726,  // ✦ '\u2726'
	"a34":                                 0x2727,  // ✧ '\u2727'
	"a35":                                 0x2605,  // ★ '\u2605'
	"a36":                                 0x2729,  // ✩ '\u2729'
	"a37":                                 0x272a,  // ✪ '\u272a'
	"a38":                                 0x272b,  // ✫ '\u272b'
	"a39":                                 0x272c,  // ✬ '\u272c'
	"a40":                                 0x272d,  // ✭ '\u272d'
	"a41":                                 0x272e,  // ✮ '\u272e'
	"a42":                                 0x272f,  // ✯ '\u272f'
	"a43":                                 0x2730,  // ✰ '\u2730'
	"a44":                                 0x2731,  // ✱ '\u2731'
	"a45":                                 0x2732,  // ✲ '\u2732'
	"a46":                                 0x2733,  // ✳ '\u2733'
	"a47":                                 0x2734,  // ✴ '\u2734'
	"a48":                                 0x2735,  // ✵ '\u2735'
	"a49":                                 0x2736,  // ✶ '\u2736'
	"a50":                                 0x2737,  // ✷ '\u2737'
	"a51":                                 0x2738,  // ✸ '\u2738'
	"a52":                                 0x2739,  // ✹ '\u2739'
	"a53":                                 0x273a,  // ✺ '\u273a'
	"a54":                                 0x273b,  // ✻ '\u273b'
	"a55":                                 0x273c,  // ✼ '\u273c'
	"a56":                                 0x273d,  // ✽ '\u273d'
	"a57":                                 0x273e,  // ✾ '\u273e'
	"a58":                                 0x273f,  // ✿ '\u273f'
	"a59":                                 0x2740,  // ❀ '\u2740'
	"a60":                                 0x2741,  // ❁ '\u2741'
	"a61":                                 0x2742,  // ❂ '\u2742'
	"a62":                                 0x2743,  // ❃ '\u2743'
	"a63":                                 0x2744,  // ❄ '\u2744'
	"a64":                                 0x2745,  // ❅ '\u2745'
	"a65":                                 0x2746,  // ❆ '\u2746'
	"a66":                                 0x2747,  // ❇ '\u2747'
	"a67":                                 0x2748,  // ❈ '\u2748'
	"a68":                                 0x2749,  // ❉ '\u2749'
	"a69":                                 0x274a,  // ❊ '\u274a'
	"a70":                                 0x274b,  // ❋ '\u274b'
	"a71":                                 0x25cf,  // ● '\u25cf'
	"a72":                                 0x274d,  // ❍ '\u274d'
	"a73":                                 0x25a0,  // ■ '\u25a0'
	"a74":                                 0x274f,  // ❏ '\u274f'
	"a75":                                 0x2751,  // ❑ '\u2751'
	"a76":                                 0x25b2,  // ▲ '\u25b2'
	"a77":                                 0x25bc,  // ▼ '\u25bc'
	"a78":                                 0x25c6,  // ◆ '\u25c6'
	"a79":                                 0x2756,  // ❖ '\u2756'
	"a81":                                 0x25d7,  // ◗ '\u25d7'
	"a82":                                 0x2758,  // ❘ '\u2758'
	"a83":                                 0x2759,  // ❙ '\u2759'
	"a84":                                 0x275a,  // ❚ '\u275a'
	"a85":                                 0xf8de,  //  '\uf8de'
	"a86":                                 0xf8e0,  //  '\uf8e0'
	"a87":                                 0xf8e1,  //  '\uf8e1'
	"a88":                                 0xf8e2,  //  '\uf8e2'
	"a89":                                 0xf8d7,  //  '\uf8d7'
	"a90":                                 0xf8d8,  //  '\uf8d8'
	"a91":                                 0xf8db,  //  '\uf8db'
	"a92":                                 0xf8dc,  //  '\uf8dc'
	"a93":                                 0xf8d9,  //  '\uf8d9'
	"a94":                                 0xf8da,  //  '\uf8da'
	"a95":                                 0xf8e3,  //  '\uf8e3'
	"a96":                                 0xf8e4,  //  '\uf8e4'
	"a97":                                 0x275b,  // ❛ '\u275b'
	"a98":                                 0x275c,  // ❜ '\u275c'
	"a99":                                 0x275d,  // ❝ '\u275d'
	"a100":                                0x275e,  // ❞ '\u275e'
	"a101":                                0x2761,  // ❡ '\u2761'
	"a102":                                0x2762,  // ❢ '\u2762'
	"a103":                                0x2763,  // ❣ '\u2763'
	"a104":                                0x2764,  // ❤ '\u2764'
	"a105":                                0x2710,  // ✐ '\u2710'
	"a106":                                0x2765,  // ❥ '\u2765'
	"a107":                                0x2766,  // ❦ '\u2766'
	"a108":                                0x2767,  // ❧ '\u2767'
	"a117":                                0x2709,  // ✉ '\u2709'
	"a118":                                0x2708,  // ✈ '\u2708'
	"a119":                                0x2707,  // ✇ '\u2707'
	"a120":                                0x2460,  // ① '\u2460'
	"a121":                                0x2461,  // ② '\u2461'
	"a122":                                0x2462,  // ③ '\u2462'
	"a123":                                0x2463,  // ④ '\u2463'
	"a124":                                0x2464,  // ⑤ '\u2464'
	"a125":                                0x2465,  // ⑥ '\u2465'
	"a126":                                0x2466,  // ⑦ '\u2466'
	"a127":                                0x2467,  // ⑧ '\u2467'
	"a128":                                0x2468,  // ⑨ '\u2468'
	"a129":                                0x2469,  // ⑩ '\u2469'
	"a130":                                0x2776,  // ❶ '\u2776'
	"a131":                                0x2777,  // ❷ '\u2777'
	"a132":                                0x2778,  // ❸ '\u2778'
	"a133":                                0x2779,  // ❹ '\u2779'
	"a134":                                0x277a,  // ❺ '\u277a'
	"a135":                                0x277b,  // ❻ '\u277b'
	"a136":                                0x277c,  // ❼ '\u277c'
	"a137":                                0x277d,  // ❽ '\u277d'
	"a138":                                0x277e,  // ❾ '\u277e'
	"a139":                                0x277f,  // ❿ '\u277f'
	"a140":                                0x2780,  // ➀ '\u2780'
	"a141":                                0x2781,  // ➁ '\u2781'
	"a142":                                0x2782,  // ➂ '\u2782'
	"a143":                                0x2783,  // ➃ '\u2783'
	"a144":                                0x2784,  // ➄ '\u2784'
	"a145":                                0x2785,  // ➅ '\u2785'
	"a146":                                0x2786,  // ➆ '\u2786'
	"a147":                                0x2787,  // ➇ '\u2787'
	"a148":                                0x2788,  // ➈ '\u2788'
	"a149":                                0x2789,  // ➉ '\u2789'
	"a150":                                0x278a,  // ➊ '\u278a'
	"a151":                                0x278b,  // ➋ '\u278b'
	"a152":                                0x278c,  // ➌ '\u278c'
	"a153":                                0x278d,  // ➍ '\u278d'
	"a154":                                0x278e,  // ➎ '\u278e'
	"a155":                                0x278f,  // ➏ '\u278f'
	"a156":                                0x2790,  // ➐ '\u2790'
	"a157":                                0x2791,  // ➑ '\u2791'
	"a158":                                0x2792,  // ➒ '\u2792'
	"a159":                                0x2793,  // ➓ '\u2793'
	"a160":                                0x2794,  // ➔ '\u2794'
	"a162":                                0x27a3,  // ➣ '\u27a3'
	"a164":                                0x2195,  // ↕ '\u2195'
	"a165":                                0x2799,  // ➙ '\u2799'
	"a166":                                0x279b,  // ➛ '\u279b'
	"a167":                                0x279c,  // ➜ '\u279c'
	"a168":                                0x279d,  // ➝ '\u279d'
	"a169":                                0x279e,  // ➞ '\u279e'
	"a170":                                0x279f,  // ➟ '\u279f'
	"a171":                                0x27a0,  // ➠ '\u27a0'
	"a172":                                0x27a1,  // ➡ '\u27a1'
	"a173":                                0x27a2,  // ➢ '\u27a2'
	"a174":                                0x27a4,  // ➤ '\u27a4'
	"a175":                                0x27a5,  // ➥ '\u27a5'
	"a176":                                0x27a6,  // ➦ '\u27a6'
	"a177":                                0x27a7,  // ➧ '\u27a7'
	"a178":                                0x27a8,  // ➨ '\u27a8'
	"a179":                                0x27a9,  // ➩ '\u27a9'
	"a180":                                0x27ab,  // ➫ '\u27ab'
	"a181":                                0x27ad,  // ➭ '\u27ad'
	"a182":                                0x27af,  // ➯ '\u27af'
	"a183":                                0x27b2,  // ➲ '\u27b2'
	"a184":                                0x27b3,  // ➳ '\u27b3'
	"a185":                                0x27b5,  // ➵ '\u27b5'
	"a186":                                0x27b8,  // ➸ '\u27b8'
	"a187":                                0x27ba,  // ➺ '\u27ba'
	"a188":                                0x27bb,  // ➻ '\u27bb'
	"a189":                                0x27bc,  // ➼ '\u27bc'
	"a190":                                0x27bd,  // ➽ '\u27bd'
	"a191":                                0x27be,  // ➾ '\u27be'
	"a192":                                0x279a,  // ➚ '\u279a'
	"a193":                                0x27aa,  // ➪ '\u27aa'
	"a194":                                0x27b6,  // ➶ '\u27b6'
	"a195":                                0x27b9,  // ➹ '\u27b9'
	"a196":                                0x2798,  // ➘ '\u2798'
	"a197":                                0x27b4,  // ➴ '\u27b4'
	"a198":                                0x27b7,  // ➷ '\u27b7'
	"a199":                                0x27ac,  // ➬ '\u27ac'
	"a200":                                0x27ae,  // ➮ '\u27ae'
	"a201":                                0x27b1,  // ➱ '\u27b1'
	"a202":                                0x2703,  // ✃ '\u2703'
	"a203":                                0x2750,  // ❐ '\u2750'
	"a204":                                0x2752,  // ❒ '\u2752'
	"a205":                                0xf8dd,  //  '\uf8dd'
	"a206":                                0xf8df,  //  '\uf8df'
	"aabengali":                           0x0986,  // আ '\u0986'
	"aacute":                              0x00e1,  // á '\u00e1'
	"aadeva":                              0x0906,  // आ '\u0906'
	"aagujarati":                          0x0a86,  // આ '\u0a86'
	"aagurmukhi":                          0x0a06,  // ਆ '\u0a06'
	"aamatragurmukhi":                     0x0a3e,  // ਾ '\u0a3e'
	"aarusquare":                          0x3303,  // ㌃ '\u3303'
	"aavowelsignbengali":                  0x09be,  // া '\u09be'
	"aavowelsigndeva":                     0x093e,  // ा '\u093e'
	"aavowelsigngujarati":                 0x0abe,  // ા '\u0abe'
	"abbreviationmarkarmenian":            0x055f,  // ՟ '\u055f'
	"abbreviationsigndeva":                0x0970,  // ॰ '\u0970'
	"abengali":                            0x0985,  // অ '\u0985'
	"abopomofo":                           0x311a,  // ㄚ '\u311a'
	"abreve":                              0x0103,  // ă '\u0103'
	"abreveacute":                         0x1eaf,  // ắ '\u1eaf'
	"abrevecyrillic":                      0x04d1,  // ӑ '\u04d1'
	"abrevedotbelow":                      0x1eb7,  // ặ '\u1eb7'
	"abrevegrave":                         0x1eb1,  // ằ '\u1eb1'
	"abrevehookabove":                     0x1eb3,  // ẳ '\u1eb3'
	"abrevetilde":                         0x1eb5,  // ẵ '\u1eb5'
	"acaron":                              0x01ce,  // ǎ '\u01ce'
	"accountof":                           0x2100,  // ℀ '\u2100'
	"accurrent":                           0x23e6,  // ⏦ '\u23e6'
	"acidfree":                            0x267e,  // ♾ '\u267e'
	"acircle":                             0x24d0,  // ⓐ '\u24d0'
	"acircumflex":                         0x00e2,  // â '\u00e2'
	"acircumflexacute":                    0x1ea5,  // ấ '\u1ea5'
	"acircumflexdotbelow":                 0x1ead,  // ậ '\u1ead'
	"acircumflexgrave":                    0x1ea7,  // ầ '\u1ea7'
	"acircumflexhookabove":                0x1ea9,  // ẩ '\u1ea9'
	"acircumflextilde":                    0x1eab,  // ẫ '\u1eab'
	"acute":                               0x00b4,  // ´ '\u00b4'
	"acutebelowcmb":                       0x0317,  // ̗ '\u0317'
	"acutecomb":                           0x0301,  // ́ '\u0301'
	"acutedeva":                           0x0954,  // ॔ '\u0954'
	"acutelowmod":                         0x02cf,  // ˏ '\u02cf'
	"acutenosp":                           0x0274,  // ɴ '\u0274'
	"acutetonecmb":                        0x0341,  // ́ '\u0341'
	"acwcirclearrow":                      0x2940,  // ⥀ '\u2940'
	"acwleftarcarrow":                     0x2939,  // ⤹ '\u2939'
	"acwopencirclearrow":                  0x21ba,  // ↺ '\u21ba'
	"acwoverarcarrow":                     0x293a,  // ⤺ '\u293a'
	"acwunderarcarrow":                    0x293b,  // ⤻ '\u293b'
	"adblgrave":                           0x0201,  // ȁ '\u0201'
	"addakgurmukhi":                       0x0a71,  // ੱ '\u0a71'
	"addresssubject":                      0x2101,  // ℁ '\u2101'
	"adeva":                               0x0905,  // अ '\u0905'
	"adieresis":                           0x00e4,  // ä '\u00e4'
	"adieresiscyrillic":                   0x04d3,  // ӓ '\u04d3'
	"adieresismacron":                     0x01df,  // ǟ '\u01df'
	"adotbelow":                           0x1ea1,  // ạ '\u1ea1'
	"adotmacron":                          0x01e1,  // ǡ '\u01e1'
	"adots":                               0x22f0,  // ⋰ '\u22f0'
	"ae":                                  0x00e6,  // æ '\u00e6'
	"aeacute":                             0x01fd,  // ǽ '\u01fd'
	"aekorean":                            0x3150,  // ㅐ '\u3150'
	"aemacron":                            0x01e3,  // ǣ '\u01e3'
	"afii299":                             0x200e,  //  '\u200e'
	"afii300":                             0x200f,  //  '\u200f'
	"afii301":                             0x200d,  //  '\u200d'
	"afii10017":                           0x0410,  // А '\u0410'
	"afii10018":                           0x0411,  // Б '\u0411'
	"afii10019":                           0x0412,  // В '\u0412'
	"afii10024":                           0x0416,  // Ж '\u0416'
	"afii10025":                           0x0417,  // З '\u0417'
	"afii10027":                           0x0419,  // Й '\u0419'
	"afii10028":                           0x041a,  // К '\u041a'
	"afii10031":                           0x041d,  // Н '\u041d'
	"afii10033":                           0x041f,  // П '\u041f'
	"afii10034":                           0x0420,  // Р '\u0420'
	"afii10035":                           0x0421,  // С '\u0421'
	"afii10036":                           0x0422,  // Т '\u0422'
	"afii10038":                           0x0424,  // Ф '\u0424'
	"afii10043":                           0x0429,  // Щ '\u0429'
	"afii10044":                           0x042a,  // Ъ '\u042a'
	"afii10045":                           0x042b,  // Ы '\u042b'
	"afii10046":                           0x042c,  // Ь '\u042c'
	"afii10048":                           0x042e,  // Ю '\u042e'
	"afii10049":                           0x042f,  // Я '\u042f'
	"afii10051":                           0x0402,  // Ђ '\u0402'
	"afii10052":                           0x0403,  // Ѓ '\u0403'
	"afii10054":                           0x0405,  // Ѕ '\u0405'
	"afii10055":                           0x0406,  // І '\u0406'
	"afii10057":                           0x0408,  // Ј '\u0408'
	"afii10059":                           0x040a,  // Њ '\u040a'
	"afii10063":                           0xf6c4,  //  '\uf6c4'
	"afii10064":                           0xf6c5,  //  '\uf6c5'
	"afii10065":                           0x0430,  // а '\u0430'
	"afii10067":                           0x0432,  // в '\u0432'
	"afii10068":                           0x0433,  // г '\u0433'
	"afii10069":                           0x0434,  // д '\u0434'
	"afii10071":                           0x0451,  // ё '\u0451'
	"afii10073":                           0x0437,  // з '\u0437'
	"afii10075":                           0x0439,  // й '\u0439'
	"afii10079":                           0x043d,  // н '\u043d'
	"afii10084":                           0x0442,  // т '\u0442'
	"afii10085":                           0x0443,  // у '\u0443'
	"afii10086":                           0x0444,  // ф '\u0444'
	"afii10087":                           0x0445,  // х '\u0445'
	"afii10090":                           0x0448,  // ш '\u0448'
	"afii10091":                           0x0449,  // щ '\u0449'
	"afii10096":                           0x044e,  // ю '\u044e'
	"afii10097":                           0x044f,  // я '\u044f'
	"afii10102":                           0x0455,  // ѕ '\u0455'
	"afii10103":                           0x0456,  // і '\u0456'
	"afii10105":                           0x0458,  // ј '\u0458'
	"afii10107":                           0x045a,  // њ '\u045a'
	"afii10109":                           0x045c,  // ќ '\u045c'
	"afii10110":                           0x045e,  // ў '\u045e'
	"afii10146":                           0x0462,  // Ѣ '\u0462'
	"afii10147":                           0x0472,  // Ѳ '\u0472'
	"afii10148":                           0x0474,  // Ѵ '\u0474'
	"afii10192":                           0xf6c6,  //  '\uf6c6'
	"afii10195":                           0x0473,  // ѳ '\u0473'
	"afii10196":                           0x0475,  // ѵ '\u0475'
	"afii10831":                           0xf6c7,  //  '\uf6c7'
	"afii10832":                           0xf6c8,  //  '\uf6c8'
	"afii57388":                           0x060c,  // ، '\u060c'
	"afii57395":                           0x0663,  // ٣ '\u0663'
	"afii57398":                           0x0666,  // ٦ '\u0666'
	"afii57399":                           0x0667,  // ٧ '\u0667'
	"afii57403":                           0x061b,  // ؛ '\u061b'
	"afii57407":                           0x061f,  // ؟ '\u061f'
	"afii57410":                           0x0622,  // آ '\u0622'
	"afii57411":                           0x0623,  // أ '\u0623'
	"afii57412":                           0x0624,  // ؤ '\u0624'
	"afii57418":                           0x062a,  // ت '\u062a'
	"afii57421":                           0x062d,  // ح '\u062d'
	"afii57422":                           0x062e,  // خ '\u062e'
	"afii57423":                           0x062f,  // د '\u062f'
	"afii57427":                           0x0633,  // س '\u0633'
	"afii57428":                           0x0634,  // ش '\u0634'
	"afii57429":                           0x0635,  // ص '\u0635'
	"afii57430":                           0x0636,  // ض '\u0636'
	"afii57433":                           0x0639,  // ع '\u0639'
	"afii57441":                           0x0641,  // ف '\u0641'
	"afii57442":                           0x0642,  // ق '\u0642'
	"afii57443":                           0x0643,  // ك '\u0643'
	"afii57444":                           0x0644,  // ل '\u0644'
	"afii57445":                           0x0645,  // م '\u0645'
	"afii57451":                           0x064b,  // ً '\u064b'
	"afii57452":                           0x064c,  // ٌ '\u064c'
	"afii57456":                           0x0650,  // ِ '\u0650'
	"afii57508":                           0x0698,  // ژ '\u0698'
	"afii57511":                           0x0679,  // ٹ '\u0679'
	"afii57512":                           0x0688,  // ڈ '\u0688'
	"afii57514":                           0x06ba,  // ں '\u06ba'
	"afii57534":                           0x06d5,  // ە '\u06d5'
	"afii57636":                           0x20aa,  // ₪ '\u20aa'
	"afii57645":                           0x05be,  // ־ '\u05be'
	"afii57666":                           0x05d2,  // ג '\u05d2'
	"afii57668":                           0x05d4,  // ה '\u05d4'
	"afii57670":                           0x05d6,  // ז '\u05d6'
	"afii57671":                           0x05d7,  // ח '\u05d7'
	"afii57673":                           0x05d9,  // י '\u05d9'
	"afii57674":                           0x05da,  // ך '\u05da'
	"afii57679":                           0x05df,  // ן '\u05df'
	"afii57684":                           0x05e4,  // פ '\u05e4'
	"afii57686":                           0x05e6,  // צ '\u05e6'
	"afii57695":                           0xfb2b,  // שׂ '\ufb2b'
	"afii57716":                           0x05f0,  // װ '\u05f0'
	"afii57717":                           0x05f1,  // ױ '\u05f1'
	"afii57797":                           0x05b8,  // ָ '\u05b8'
	"afii57799":                           0x05b0,  // ְ '\u05b0'
	"afii57803":                           0x05c2,  // ׂ '\u05c2'
	"afii57841":                           0x05bf,  // ֿ '\u05bf'
	"afii57842":                           0x05c0,  // ׀ '\u05c0'
	"afii61289":                           0x2113,  // ℓ '\u2113'
	"afii61573":                           0x202c,  //  '\u202c'
	"afii61574":                           0x202d,  //  '\u202d'
	"afii61575":                           0x202e,  //  '\u202e'
	"afii61664":                           0x200c,  //  '\u200c'
	"afii63167":                           0x066d,  // ٭ '\u066d'
	"afii64937":                           0x02bd,  // ʽ '\u02bd'
	"agrave":                              0x00e0,  // à '\u00e0'
	"agujarati":                           0x0a85,  // અ '\u0a85'
	"agurmukhi":                           0x0a05,  // ਅ '\u0a05'
	"ahiragana":                           0x3042,  // あ '\u3042'
	"ahookabove":                          0x1ea3,  // ả '\u1ea3'
	"aibengali":                           0x0990,  // ঐ '\u0990'
	"aibopomofo":                          0x311e,  // ㄞ '\u311e'
	"aideva":                              0x0910,  // ऐ '\u0910'
	"aiecyrillic":                         0x04d5,  // ӕ '\u04d5'
	"aigujarati":                          0x0a90,  // ઐ '\u0a90'
	"aigurmukhi":                          0x0a10,  // ਐ '\u0a10'
	"aimatragurmukhi":                     0x0a48,  // ੈ '\u0a48'
	"ainfinalarabic":                      0xfeca,  // ﻊ '\ufeca'
	"aininitialarabic":                    0xfecb,  // ﻋ '\ufecb'
	"ainisolated":                         0xfec9,  // ﻉ '\ufec9'
	"ainmedialarabic":                     0xfecc,  // ﻌ '\ufecc'
	"ainvertedbreve":                      0x0203,  // ȃ '\u0203'
	"aivowelsignbengali":                  0x09c8,  // ৈ '\u09c8'
	"aivowelsigndeva":                     0x0948,  // ै '\u0948'
	"aivowelsigngujarati":                 0x0ac8,  // ૈ '\u0ac8'
	"akatakana":                           0x30a2,  // ア '\u30a2'
	"akatakanahalfwidth":                  0xff71,  // ｱ '\uff71'
	"akorean":                             0x314f,  // ㅏ '\u314f'
	"alef":                                0x05d0,  // א '\u05d0'
	"alefarabic":                          0x0627,  // ا '\u0627'
	"alefdageshhebrew":                    0xfb30,  // אּ '\ufb30'
	"aleffinalarabic":                     0xfe8e,  // ﺎ '\ufe8e'
	"alefhamzaabovefinalarabic":           0xfe84,  // ﺄ '\ufe84'
	"alefhamzabelowarabic":                0x0625,  // إ '\u0625'
	"alefhamzabelowfinalarabic":           0xfe88,  // ﺈ '\ufe88'
	"alefisolated":                        0xfe8d,  // ﺍ '\ufe8d'
	"aleflamedhebrew":                     0xfb4f,  // ﭏ '\ufb4f'
	"alefmaddaabovefinalarabic":           0xfe82,  // ﺂ '\ufe82'
	"alefmaksuraarabic":                   0x0649,  // ى '\u0649'
	"alefmaksurafinalarabic":              0xfef0,  // ﻰ '\ufef0'
	"alefmaksuraisolated":                 0xfeef,  // ﻯ '\ufeef'
	"alefmaksuramedialarabic":             0xfef4,  // ﻴ '\ufef4'
	"alefpatahhebrew":                     0xfb2e,  // אַ '\ufb2e'
	"alefqamatshebrew":                    0xfb2f,  // אָ '\ufb2f'
	"alefwasla":                           0x0671,  // ٱ '\u0671'
	"alefwaslafinal":                      0xfb51,  // ﭑ '\ufb51'
	"alefwaslaisolated":                   0xfb50,  // ﭐ '\ufb50'
	"alefwithfathatanfinal":               0xfd3c,  // ﴼ '\ufd3c'
	"alefwithfathatanisolated":            0xfd3d,  // ﴽ '\ufd3d'
	"alefwithhamzaaboveisolated":          0xfe83,  // ﺃ '\ufe83'
	"alefwithhamzabelowisolated":          0xfe87,  // ﺇ '\ufe87'
	"alefwithmaddaaboveisolated":          0xfe81,  // ﺁ '\ufe81'
	"aleph":                               0x2135,  // ℵ '\u2135'
	"allequal":                            0x224c,  // ≌ '\u224c'
	"alpha":                               0x03b1,  // α '\u03b1'
	"alphatonos":                          0x03ac,  // ά '\u03ac'
	"altselector":                         0xd802,  //  '\ufffd'
	"amacron":                             0x0101,  // ā '\u0101'
	"amonospace":                          0xff41,  // ａ '\uff41'
	"ampersand":                           0x0026,  // & '&'
	"ampersandmonospace":                  0xff06,  // ＆ '\uff06'
	"ampersandsmall":                      0xf726,  //  '\uf726'
	"amsquare":                            0x33c2,  // ㏂ '\u33c2'
	"anbopomofo":                          0x3122,  // ㄢ '\u3122'
	"angbopomofo":                         0x3124,  // ㄤ '\u3124'
	"angbracketleft":                      0x27e8,  // ⟨ '\u27e8'
	"angbracketright":                     0x27e9,  // ⟩ '\u27e9'
	"angdnr":                              0x299f,  // ⦟ '\u299f'
	"angkhankhuthai":                      0x0e5a,  // ๚ '\u0e5a'
	"angle":                               0x2220,  // ∠ '\u2220'
	"anglebracketleft":                    0x3008,  // 〈 '\u3008'
	"anglebracketleftvertical":            0xfe3f,  // ︿ '\ufe3f'
	"anglebracketright":                   0x3009,  // 〉 '\u3009'
	"anglebracketrightvertical":           0xfe40,  // ﹀ '\ufe40'
	"angleleft":                           0x2329,  // 〈 '\u2329'
	"angleright":                          0x232a,  // 〉 '\u232a'
	"angles":                              0x299e,  // ⦞ '\u299e'
	"angleubar":                           0x29a4,  // ⦤ '\u29a4'
	"angstrom":                            0x212b,  // Å '\u212b'
	"annuity":                             0x20e7,  // ⃧ '\u20e7'
	"anoteleia":                           0x0387,  // · '\u0387'
	"anticlockwise":                       0x27f2,  // ⟲ '\u27f2'
	"anudattadeva":                        0x0952,  // ॒ '\u0952'
	"anusvarabengali":                     0x0982,  // ং '\u0982'
	"anusvaradeva":                        0x0902,  // ं '\u0902'
	"anusvaragujarati":                    0x0a82,  // ં '\u0a82'
	"aogonek":                             0x0105,  // ą '\u0105'
	"apaatosquare":                        0x3300,  // ㌀ '\u3300'
	"aparen":                              0x249c,  // ⒜ '\u249c'
	"apostrophe":                          0x0245,  // Ʌ '\u0245'
	"apostrophearmenian":                  0x055a,  // ՚ '\u055a'
	"apostrophemod":                       0x02bc,  // ʼ '\u02bc'
	"apostropherev":                       0x0246,  // Ɇ '\u0246'
	"apple":                               0xf8ff,  //  '\uf8ff'
	"approaches":                          0x2250,  // ≐ '\u2250'
	"approxeqq":                           0x2a70,  // ⩰ '\u2a70'
	"approxequal":                         0x2248,  // ≈ '\u2248'
	"approxequalorimage":                  0x2252,  // ≒ '\u2252'
	"approxident":                         0x224b,  // ≋ '\u224b'
	"approxorequal":                       0x224a,  // ≊ '\u224a'
	"araeaekorean":                        0x318e,  // ㆎ '\u318e'
	"araeakorean":                         0x318d,  // ㆍ '\u318d'
	"arc":                                 0x2312,  // ⌒ '\u2312'
	"arceq":                               0x2258,  // ≘ '\u2258'
	"archleftdown":                        0x21b6,  // ↶ '\u21b6'
	"archrightdown":                       0x21b7,  // ↷ '\u21b7'
	"arighthalfring":                      0x1e9a,  // ẚ '\u1e9a'
	"aring":                               0x00e5,  // å '\u00e5'
	"aringacute":                          0x01fb,  // ǻ '\u01fb'
	"aringbelow":                          0x1e01,  // ḁ '\u1e01'
	"arrowbardown":                        0x0590,  //  '\u0590'
	"arrowbarleft":                        0x058d,  // ֍ '\u058d'
	"arrowbarright":                       0x058f,  // ֏ '\u058f'
	"arrowbarup":                          0x058e,  // ֎ '\u058e'
	"arrowboth":                           0x2194,  // ↔ '\u2194'
	"arrowdashdown":                       0x21e3,  // ⇣ '\u21e3'
	"arrowdashleft":                       0x21e0,  // ⇠ '\u21e0'
	"arrowdashright":                      0x21e2,  // ⇢ '\u21e2'
	"arrowdashup":                         0x21e1,  // ⇡ '\u21e1'
	"arrowdblboth":                        0x21d4,  // ⇔ '\u21d4'
	"arrowdblbothv":                       0x21d5,  // ⇕ '\u21d5'
	"arrowdbldown":                        0x21d3,  // ⇓ '\u21d3'
	"arrowdblleft":                        0x21d0,  // ⇐ '\u21d0'
	"arrowdblright":                       0x21d2,  // ⇒ '\u21d2'
	"arrowdblup":                          0x21d1,  // ⇑ '\u21d1'
	"arrowdown":                           0x2193,  // ↓ '\u2193'
	"arrowdownleft":                       0x2199,  // ↙ '\u2199'
	"arrowdownright":                      0x2198,  // ↘ '\u2198'
	"arrowdownwhite":                      0x21e9,  // ⇩ '\u21e9'
	"arrowheaddownmod":                    0x02c5,  // ˅ '\u02c5'
	"arrowheadleftmod":                    0x02c2,  // ˂ '\u02c2'
	"arrowheadrightmod":                   0x02c3,  // ˃ '\u02c3'
	"arrowheadupmod":                      0x02c4,  // ˄ '\u02c4'
	"arrowhookleft":                       0x21aa,  // ↪ '\u21aa'
	"arrowhookright":                      0x21a9,  // ↩ '\u21a9'
	"arrowhorizex":                        0xf8e7,  //  '\uf8e7'
	"arrowleft":                           0x2190,  // ← '\u2190'
	"arrowleftbothalf":                    0x21bd,  // ↽ '\u21bd'
	"arrowleftdblstroke":                  0x21cd,  // ⇍ '\u21cd'
	"arrowleftoverright":                  0x21c6,  // ⇆ '\u21c6'
	"arrowleftwhite":                      0x21e6,  // ⇦ '\u21e6'
	"arrowright":                          0x2192,  // → '\u2192'
	"arrowrightbothalf":                   0x21c1,  // ⇁ '\u21c1'
	"arrowrightdblstroke":                 0x21cf,  // ⇏ '\u21cf'
	"arrowrightoverleft":                  0x21c4,  // ⇄ '\u21c4'
	"arrowrightwhite":                     0x21e8,  // ⇨ '\u21e8'
	"arrowtableft":                        0x21e4,  // ⇤ '\u21e4'
	"arrowtabright":                       0x21e5,  // ⇥ '\u21e5'
	"arrowtailleft":                       0x21a2,  // ↢ '\u21a2'
	"arrowtailright":                      0x21a3,  // ↣ '\u21a3'
	"arrowtripleleft":                     0x21da,  // ⇚ '\u21da'
	"arrowtripleright":                    0x21db,  // ⇛ '\u21db'
	"arrowup":                             0x2191,  // ↑ '\u2191'
	"arrowupdownbase":                     0x21a8,  // ↨ '\u21a8'
	"arrowupleft":                         0x2196,  // ↖ '\u2196'
	"arrowupleftofdown":                   0x21c5,  // ⇅ '\u21c5'
	"arrowupright":                        0x2197,  // ↗ '\u2197'
	"arrowupwhite":                        0x21e7,  // ⇧ '\u21e7'
	"arrowvertex":                         0xf8e6,  //  '\uf8e6'
	"ascendercompwordmark":                0xd80a,  //  '\ufffd'
	"asciicircum":                         0x005e,  // ^ '^'
	"asciicircummonospace":                0xff3e,  // ＾ '\uff3e'
	"asciitilde":                          0x007e,  // ~ '~'
	"asciitildemonospace":                 0xff5e,  // ～ '\uff5e'
	"ascript":                             0x0251,  // ɑ '\u0251'
	"ascriptturned":                       0x0252,  // ɒ '\u0252'
	"asmallhiragana":                      0x3041,  // ぁ '\u3041'
	"asmallkatakana":                      0x30a1,  // ァ '\u30a1'
	"asmallkatakanahalfwidth":             0xff67,  // ｧ '\uff67'
	"assert":                              0x22a6,  // ⊦ '\u22a6'
	"asteq":                               0x2a6e,  // ⩮ '\u2a6e'
	"asteraccent":                         0x20f0,  // ⃰ '\u20f0'
	"asterisk":                            0x002a,  // * '*'
	"asteriskmath":                        0x2217,  // ∗ '\u2217'
	"asteriskmonospace":                   0xff0a,  // ＊ '\uff0a'
	"asterisksmall":                       0xfe61,  // ﹡ '\ufe61'
	"asterism":                            0x2042,  // ⁂ '\u2042'
	"astrosun":                            0x2609,  // ☉ '\u2609'
	"asuperior":                           0xf6e9,  //  '\uf6e9'
	"asymptoticallyequal":                 0x2243,  // ≃ '\u2243'
	"at":                                  0x0040,  // @ '@'
	"atilde":                              0x00e3,  // ã '\u00e3'
	"atmonospace":                         0xff20,  // ＠ '\uff20'
	"atsmall":                             0xfe6b,  // ﹫ '\ufe6b'
	"aturned":                             0x0250,  // ɐ '\u0250'
	"aubengali":                           0x0994,  // ঔ '\u0994'
	"aubopomofo":                          0x3120,  // ㄠ '\u3120'
	"audeva":                              0x0914,  // औ '\u0914'
	"augujarati":                          0x0a94,  // ઔ '\u0a94'
	"augurmukhi":                          0x0a14,  // ਔ '\u0a14'
	"aulengthmarkbengali":                 0x09d7,  // ৗ '\u09d7'
	"aumatragurmukhi":                     0x0a4c,  // ੌ '\u0a4c'
	"auvowelsignbengali":                  0x09cc,  // ৌ '\u09cc'
	"auvowelsigndeva":                     0x094c,  // ौ '\u094c'
	"auvowelsigngujarati":                 0x0acc,  // ૌ '\u0acc'
	"avagrahadeva":                        0x093d,  // ऽ '\u093d'
	"awint":                               0x2a11,  // ⨑ '\u2a11'
	"aybarmenian":                         0x0561,  // ա '\u0561'
	"ayinaltonehebrew":                    0xfb20,  // ﬠ '\ufb20'
	"ayinhebrew":                          0x05e2,  // ע '\u05e2'
	"b":                                   0x0062,  // b 'b'
	"bNot":                                0x2aed,  // ⫭ '\u2aed'
	"babengali":                           0x09ac,  // ব '\u09ac'
	"backdprime":                          0x2036,  // ‶ '\u2036'
	"backed":                              0x024c,  // Ɍ '\u024c'
	"backslash":                           0x005c,  // \\ '\\'
	"backslashmonospace":                  0xff3c,  // ＼ '\uff3c'
	"backtrprime":                         0x2037,  // ‷ '\u2037'
	"badeva":                              0x092c,  // ब '\u092c'
	"bagmember":                           0x22ff,  // ⋿ '\u22ff'
	"bagujarati":                          0x0aac,  // બ '\u0aac'
	"bagurmukhi":                          0x0a2c,  // ਬ '\u0a2c'
	"bahiragana":                          0x3070,  // ば '\u3070'
	"bahtthai":                            0x0e3f,  // ฿ '\u0e3f'
	"bakatakana":                          0x30d0,  // バ '\u30d0'
	"bar":                                 0x007c,  // | '|'
	"barV":                                0x2aea,  // ⫪ '\u2aea'
	"barcap":                              0x2a43,  // ⩃ '\u2a43'
	"barcup":                              0x2a42,  // ⩂ '\u2a42'
	"bardownharpoonleft":                  0x2961,  // ⥡ '\u2961'
	"bardownharpoonright":                 0x295d,  // ⥝ '\u295d'
	"barleftarrowrightarrowba":            0x21b9,  // ↹ '\u21b9'
	"barleftharpoondown":                  0x2956,  // ⥖ '\u2956'
	"barleftharpoonup":                    0x2952,  // ⥒ '\u2952'
	"barmidlongnosp":                      0x02a9,  // ʩ '\u02a9'
	"barmonospace":                        0xff5c,  // ｜ '\uff5c'
	"barovernorthwestarrow":               0x21b8,  // ↸ '\u21b8'
	"barrightarrowdiamond":                0x2920,  // ⤠ '\u2920'
	"barrightharpoondown":                 0x295f,  // ⥟ '\u295f'
	"barrightharpoonup":                   0x295b,  // ⥛ '\u295b'
	"baruparrow":                          0x2912,  // ⤒ '\u2912'
	"barupharpoonleft":                    0x2958,  // ⥘ '\u2958'
	"barupharpoonright":                   0x2954,  // ⥔ '\u2954'
	"barvee":                              0x22bd,  // ⊽ '\u22bd'
	"bbopomofo":                           0x3105,  // ㄅ '\u3105'
	"bbrktbrk":                            0x23b6,  // ⎶ '\u23b6'
	"bcircle":                             0x24d1,  // ⓑ '\u24d1'
	"bdotaccent":                          0x1e03,  // ḃ '\u1e03'
	"bdotbelow":                           0x1e05,  // ḅ '\u1e05'
	"bdtriplevdash":                       0x2506,  // ┆ '\u2506'
	"beamedsixteenthnotes":                0x266c,  // ♬ '\u266c'
	"because":                             0x2235,  // ∵ '\u2235'
	"becyrillic":                          0x0431,  // б '\u0431'
	"beharabic":                           0x0628,  // ب '\u0628'
	"behfinalarabic":                      0xfe90,  // ﺐ '\ufe90'
	"behinitialarabic":                    0xfe91,  // ﺑ '\ufe91'
	"behiragana":                          0x3079,  // べ '\u3079'
	"behisolated":                         0xfe8f,  // ﺏ '\ufe8f'
	"behmedialarabic":                     0xfe92,  // ﺒ '\ufe92'
	"behmeeminitialarabic":                0xfc9f,  // ﲟ '\ufc9f'
	"behmeemisolatedarabic":               0xfc08,  // ﰈ '\ufc08'
	"behnoonfinalarabic":                  0xfc6d,  // ﱭ '\ufc6d'
	"behwithalefmaksurafinal":             0xfc6e,  // ﱮ '\ufc6e'
	"behwithalefmaksuraisolated":          0xfc09,  // ﰉ '\ufc09'
	"behwithhahinitial":                   0xfc9d,  // ﲝ '\ufc9d'
	"behwithhehinitial":                   0xe812,  //  '\ue812'
	"behwithjeeminitial":                  0xfc9c,  // ﲜ '\ufc9c'
	"behwithkhahinitial":                  0xfc9e,  // ﲞ '\ufc9e'
	"behwithrehfinal":                     0xfc6a,  // ﱪ '\ufc6a'
	"behwithyehfinal":                     0xfc6f,  // ﱯ '\ufc6f'
	"behwithyehisolated":                  0xfc0a,  // ﰊ '\ufc0a'
	"bekatakana":                          0x30d9,  // ベ '\u30d9'
	"benarmenian":                         0x0562,  // բ '\u0562'
	"benzenr":                             0x23e3,  // ⏣ '\u23e3'
	"beta":                                0x03b2,  // β '\u03b2'
	"betasymbolgreek":                     0x03d0,  // ϐ '\u03d0'
	"betdageshhebrew":                     0xfb31,  // בּ '\ufb31'
	"beth":                                0x2136,  // ℶ '\u2136'
	"bethebrew":                           0x05d1,  // ב '\u05d1'
	"betrafehebrew":                       0xfb4c,  // בֿ '\ufb4c'
	"between":                             0x226c,  // ≬ '\u226c'
	"bhabengali":                          0x09ad,  // ভ '\u09ad'
	"bhadeva":                             0x092d,  // भ '\u092d'
	"bhagujarati":                         0x0aad,  // ભ '\u0aad'
	"bhagurmukhi":                         0x0a2d,  // ਭ '\u0a2d'
	"bhook":                               0x0253,  // ɓ '\u0253'
	"bigbot":                              0x27d8,  // ⟘ '\u27d8'
	"bigcupdot":                           0x2a03,  // ⨃ '\u2a03'
	"biginterleave":                       0x2afc,  // ⫼ '\u2afc'
	"bigodot":                             0x2a00,  // ⨀ '\u2a00'
	"bigoplus":                            0x2a01,  // ⨁ '\u2a01'
	"bigotimes":                           0x2a02,  // ⨂ '\u2a02'
	"bigslopedvee":                        0x2a57,  // ⩗ '\u2a57'
	"bigslopedwedge":                      0x2a58,  // ⩘ '\u2a58'
	"bigsqcap":                            0x2a05,  // ⨅ '\u2a05'
	"bigsqcup":                            0x2a06,  // ⨆ '\u2a06'
	"bigtalloblong":                       0x2aff,  // ⫿ '\u2aff'
	"bigtimes":                            0x2a09,  // ⨉ '\u2a09'
	"bigtop":                              0x27d9,  // ⟙ '\u27d9'
	"bigtriangleleft":                     0x2a1e,  // ⨞ '\u2a1e'
	"biguplus":                            0x2a04,  // ⨄ '\u2a04'
	"bigvee":                              0x22c1,  // ⋁ '\u22c1'
	"bigwedge":                            0x22c0,  // ⋀ '\u22c0'
	"bihiragana":                          0x3073,  // び '\u3073'
	"bikatakana":                          0x30d3,  // ビ '\u30d3'
	"bilabialclick":                       0x0298,  // ʘ '\u0298'
	"bindigurmukhi":                       0x0a02,  // ਂ '\u0a02'
	"birusquare":                          0x3331,  // ㌱ '\u3331'
	"blackcircledownarrow":                0x29ed,  // ⧭ '\u29ed'
	"blackcircledrightdot":                0x2688,  // ⚈ '\u2688'
	"blackcircledtwodots":                 0x2689,  // ⚉ '\u2689'
	"blackcircleulquadwhite":              0x25d5,  // ◕ '\u25d5'
	"blackdiamonddownarrow":               0x29ea,  // ⧪ '\u29ea'
	"blackhourglass":                      0x29d7,  // ⧗ '\u29d7'
	"blacklefthalfcircle":                 0x25d6,  // ◖ '\u25d6'
	"blackleftpointingpointer":            0x25c4,  // ◄ '\u25c4'
	"blackleftpointingtriangle":           0x25c0,  // ◀ '\u25c0'
	"blacklenticularbracketleft":          0x3010,  // 【 '\u3010'
	"blacklenticularbracketleftvertical":  0xfe3b,  // ︻ '\ufe3b'
	"blacklenticularbracketright":         0x3011,  // 】 '\u3011'
	"blacklenticularbracketrightvertical": 0xfe3c,  // ︼ '\ufe3c'
	"blacklowerlefttriangle":              0x25e3,  // ◣ '\u25e3'
	"blacklowerrighttriangle":             0x25e2,  // ◢ '\u25e2'
	"blackrectangle":                      0x25ac,  // ▬ '\u25ac'
	"blackrightpointingpointer":           0x25ba,  // ► '\u25ba'
	"blackrightpointingtriangle":          0x25b6,  // ▶ '\u25b6'
	"blacksmallsquare":                    0x25aa,  // ▪ '\u25aa'
	"blacksmilingface":                    0x263b,  // ☻ '\u263b'
	"blacktriangledown":                   0x25be,  // ▾ '\u25be'
	"blackupperlefttriangle":              0x25e4,  // ◤ '\u25e4'
	"blackupperrighttriangle":             0x25e5,  // ◥ '\u25e5'
	"blackuppointingsmalltriangle":        0x25b4,  // ▴ '\u25b4'
	"blank":                               0x2423,  // ␣ '\u2423'
	"blinebelow":                          0x1e07,  // ḇ '\u1e07'
	"blkhorzoval":                         0x2b2c,  // ⬬ '\u2b2c'
	"blkvertoval":                         0x2b2e,  // ⬮ '\u2b2e'
	"block":                               0x2588,  // █ '\u2588'
	"bmonospace":                          0xff42,  // ｂ '\uff42'
	"bobaimaithai":                        0x0e1a,  // บ '\u0e1a'
	"bohiragana":                          0x307c,  // ぼ '\u307c'
	"bokatakana":                          0x30dc,  // ボ '\u30dc'
	"botsemicircle":                       0x25e1,  // ◡ '\u25e1'
	"bowtie":                              0x22c8,  // ⋈ '\u22c8'
	"boxast":                              0x29c6,  // ⧆ '\u29c6'
	"boxbar":                              0x25eb,  // ◫ '\u25eb'
	"boxbox":                              0x29c8,  // ⧈ '\u29c8'
	"boxbslash":                           0x29c5,  // ⧅ '\u29c5'
	"boxcircle":                           0x29c7,  // ⧇ '\u29c7'
	"boxdiag":                             0x29c4,  // ⧄ '\u29c4'
	"boxonbox":                            0x29c9,  // ⧉ '\u29c9'
	"bparen":                              0x249d,  // ⒝ '\u249d'
	"bqsquare":                            0x33c3,  // ㏃ '\u33c3'
	"braceex":                             0xf8f4,  //  '\uf8f4'
	"braceleft":                           0x007b,  // { '{'
	"braceleftbt":                         0xf8f3,  //  '\uf8f3'
	"braceleftmid":                        0xf8f2,  //  '\uf8f2'
	"braceleftmonospace":                  0xff5b,  // ｛ '\uff5b'
	"braceleftsmall":                      0xfe5b,  // ﹛ '\ufe5b'
	"bracelefttp":                         0xf8f1,  //  '\uf8f1'
	"braceleftvertical":                   0xfe37,  // ︷ '\ufe37'
	"braceright":                          0x007d,  // } '}'
	"bracerightbt":                        0xf8fe,  //  '\uf8fe'
	"bracerightmid":                       0xf8fd,  //  '\uf8fd'
	"bracerightmonospace":                 0xff5d,  // ｝ '\uff5d'
	"bracerightsmall":                     0xfe5c,  // ﹜ '\ufe5c'
	"bracerighttp":                        0xf8fc,  //  '\uf8fc'
	"bracerightvertical":                  0xfe38,  // ︸ '\ufe38'
	"bracketleft":                         0x005b,  // [ '['
	"bracketleftbt":                       0xf8f0,  //  '\uf8f0'
	"bracketleftex":                       0xf8ef,  //  '\uf8ef'
	"bracketleftmonospace":                0xff3b,  // ［ '\uff3b'
	"bracketleftquill":                    0x2045,  // ⁅ '\u2045'
	"bracketlefttp":                       0xf8ee,  //  '\uf8ee'
	"bracketright":                        0x005d,  // ] ']'
	"bracketrightbt":                      0xf8fb,  //  '\uf8fb'
	"bracketrightex":                      0xf8fa,  //  '\uf8fa'
	"bracketrightmonospace":               0xff3d,  // ］ '\uff3d'
	"bracketrightquill":                   0x2046,  // ⁆ '\u2046'
	"bracketrighttp":                      0xf8f9,  //  '\uf8f9'
	"breve":                               0x02d8,  // ˘ '\u02d8'
	"breve1":                              0xf006,  //  '\uf006'
	"brevebelowcmb":                       0x032e,  // ̮ '\u032e'
	"brevecmb":                            0x0306,  // ̆ '\u0306'
	"breveinvertedbelowcmb":               0x032f,  // ̯ '\u032f'
	"breveinvertedcmb":                    0x0311,  // ̑ '\u0311'
	"breveinverteddoublecmb":              0x0361,  // ͡ '\u0361'
	"bridgebelowcmb":                      0x032a,  // ̪ '\u032a'
	"bridgeinvertedbelowcmb":              0x033a,  // ̺ '\u033a'
	"bridgeinvsubnosp":                    0x02ad,  // ʭ '\u02ad'
	"brokenbar":                           0x00a6,  // ¦ '\u00a6'
	"bsimilarleftarrow":                   0x2b41,  // ⭁ '\u2b41'
	"bsimilarrightarrow":                  0x2b47,  // ⭇ '\u2b47'
	"bsolhsub":                            0x27c8,  // ⟈ '\u27c8'
	"bstroke":                             0x0180,  // ƀ '\u0180'
	"bsuperior":                           0xf6ea,  //  '\uf6ea'
	"btimes":                              0x2a32,  // ⨲ '\u2a32'
	"btopbar":                             0x0183,  // ƃ '\u0183'
	"buhiragana":                          0x3076,  // ぶ '\u3076'
	"bukatakana":                          0x30d6,  // ブ '\u30d6'
	"bullet":                              0x2022,  // • '\u2022'
	"bulletoperator":                      0x2219,  // ∙ '\u2219'
	"bullseye":                            0x25ce,  // ◎ '\u25ce'
	"bumpeqq":                             0x2aae,  // ⪮ '\u2aae'
	"c":                                   0x0063,  // c 'c'
	"c128":                                0x0080,  //  '\u0080'
	"c129":                                0x0081,  //  '\u0081'
	"c141":                                0x008d,  //  '\u008d'
	"c142":                                0x008e,  //  '\u008e'
	"c143":                                0x008f,  //  '\u008f'
	"caarmenian":                          0x056e,  // ծ '\u056e'
	"cabengali":                           0x099a,  // চ '\u099a'
	"cacute":                              0x0107,  // ć '\u0107'
	"cadauna":                             0x2106,  // ℆ '\u2106'
	"cadeva":                              0x091a,  // च '\u091a'
	"cagujarati":                          0x0a9a,  // ચ '\u0a9a'
	"cagurmukhi":                          0x0a1a,  // ਚ '\u0a1a'
	"calsquare":                           0x3388,  // ㎈ '\u3388'
	"candrabindubengali":                  0x0981,  // ঁ '\u0981'
	"candrabinducmb":                      0x0310,  // ̐ '\u0310'
	"candrabindudeva":                     0x0901,  // ँ '\u0901'
	"candrabindugujarati":                 0x0a81,  // ઁ '\u0a81'
	"capbarcup":                           0x2a49,  // ⩉ '\u2a49'
	"capdot":                              0x2a40,  // ⩀ '\u2a40'
	"capitalcompwordmark":                 0xd809,  //  '\ufffd'
	"capovercup":                          0x2a47,  // ⩇ '\u2a47'
	"capslock":                            0x21ea,  // ⇪ '\u21ea'
	"capwedge":                            0x2a44,  // ⩄ '\u2a44'
	"careof":                              0x2105,  // ℅ '\u2105'
	"caretinsert":                         0x2038,  // ‸ '\u2038'
	"caron":                               0x02c7,  // ˇ '\u02c7'
	"caron1":                              0xf00a,  //  '\uf00a'
	"caronbelowcmb":                       0x032c,  // ̬ '\u032c'
	"caroncmb":                            0x030c,  // ̌ '\u030c'
	"carriagereturn":                      0x21b5,  // ↵ '\u21b5'
	"cbopomofo":                           0x3118,  // ㄘ '\u3118'
	"ccaron":                              0x010d,  // č '\u010d'
	"ccedilla":                            0x00e7,  // ç '\u00e7'
	"ccedillaacute":                       0x1e09,  // ḉ '\u1e09'
	"ccircle":                             0x24d2,  // ⓒ '\u24d2'
	"ccircumflex":                         0x0109,  // ĉ '\u0109'
	"ccurl":                               0x0255,  // ɕ '\u0255'
	"ccwundercurvearrow":                  0x293f,  // ⤿ '\u293f'
	"cdot":                                0x010b,  // ċ '\u010b'
	"cdsquare":                            0x33c5,  // ㏅ '\u33c5'
	"cedilla":                             0x00b8,  // ¸ '\u00b8'
	"cedilla1":                            0xf008,  //  '\uf008'
	"cedilla2":                            0xf00d,  //  '\uf00d'
	"cedillacmb":                          0x0327,  // ̧ '\u0327'
	"ceilingleft":                         0x2308,  // ⌈ '\u2308'
	"ceilingright":                        0x2309,  // ⌉ '\u2309'
	"cent":                                0x00a2,  // ¢ '\u00a2'
	"centigrade":                          0x2103,  // ℃ '\u2103'
	"centinferior":                        0xf6df,  //  '\uf6df'
	"centmonospace":                       0xffe0,  // ￠ '\uffe0'
	"centoldstyle":                        0xf7a2,  //  '\uf7a2'
	"centreline":                          0x2104,  // ℄ '\u2104'
	"centsuperior":                        0xf6e0,  //  '\uf6e0'
	"chaarmenian":                         0x0579,  // չ '\u0579'
	"chabengali":                          0x099b,  // ছ '\u099b'
	"chadeva":                             0x091b,  // छ '\u091b'
	"chagujarati":                         0x0a9b,  // છ '\u0a9b'
	"chagurmukhi":                         0x0a1b,  // ਛ '\u0a1b'
	"chbopomofo":                          0x3114,  // ㄔ '\u3114'
	"cheabkhasiancyrillic":                0x04bd,  // ҽ '\u04bd'
	"checyrillic":                         0x0447,  // ч '\u0447'
	"chedescenderabkhasiancyrillic":       0x04bf,  // ҿ '\u04bf'
	"chedescendercyrillic":                0x04b7,  // ҷ '\u04b7'
	"chedieresiscyrillic":                 0x04f5,  // ӵ '\u04f5'
	"cheharmenian":                        0x0573,  // ճ '\u0573'
	"chekhakassiancyrillic":               0x04cc,  // ӌ '\u04cc'
	"cheverticalstrokecyrillic":           0x04b9,  // ҹ '\u04b9'
	"chi":                                 0x03c7,  // χ '\u03c7'
	"chieuchacirclekorean":                0x3277,  // ㉷ '\u3277'
	"chieuchaparenkorean":                 0x3217,  // ㈗ '\u3217'
	"chieuchcirclekorean":                 0x3269,  // ㉩ '\u3269'
	"chieuchkorean":                       0x314a,  // ㅊ '\u314a'
	"chieuchparenkorean":                  0x3209,  // ㈉ '\u3209'
	"chochangthai":                        0x0e0a,  // ช '\u0e0a'
	"chochanthai":                         0x0e08,  // จ '\u0e08'
	"chochingthai":                        0x0e09,  // ฉ '\u0e09'
	"chochoethai":                         0x0e0c,  // ฌ '\u0e0c'
	"chook":                               0x0188,  // ƈ '\u0188'
	"cieucacirclekorean":                  0x3276,  // ㉶ '\u3276'
	"cieucaparenkorean":                   0x3216,  // ㈖ '\u3216'
	"cieuccirclekorean":                   0x3268,  // ㉨ '\u3268'
	"cieuckorean":                         0x3148,  // ㅈ '\u3148'
	"cieucparenkorean":                    0x3208,  // ㈈ '\u3208'
	"cieucuparenkorean":                   0x321c,  // ㈜ '\u321c'
	"cirE":                                0x29c3,  // ⧃ '\u29c3'
	"cirbot":                              0x27df,  // ⟟ '\u27df'
	"circeq":                              0x2257,  // ≗ '\u2257'
	"circleasterisk":                      0x229b,  // ⊛ '\u229b'
	"circlebottomhalfblack":               0x25d2,  // ◒ '\u25d2'
	"circlecopyrt":                        0x20dd,  // ⃝ '\u20dd'
	"circledbullet":                       0x29bf,  // ⦿ '\u29bf'
	"circleddash":                         0x229d,  // ⊝ '\u229d'
	"circledivide":                        0x2298,  // ⊘ '\u2298'
	"circledownarrow":                     0x29ec,  // ⧬ '\u29ec'
	"circledparallel":                     0x29b7,  // ⦷ '\u29b7'
	"circledrightdot":                     0x2686,  // ⚆ '\u2686'
	"circledtwodots":                      0x2687,  // ⚇ '\u2687'
	"circledvert":                         0x29b6,  // ⦶ '\u29b6'
	"circledwhitebullet":                  0x29be,  // ⦾ '\u29be'
	"circleequal":                         0x229c,  // ⊜ '\u229c'
	"circlehbar":                          0x29b5,  // ⦵ '\u29b5'
	"circlellquad":                        0x25f5,  // ◵ '\u25f5'
	"circlelrquad":                        0x25f6,  // ◶ '\u25f6'
	"circlemultiply":                      0x2297,  // ⊗ '\u2297'
	"circleonleftarrow":                   0x2b30,  // ⬰ '\u2b30'
	"circleonrightarrow":                  0x21f4,  // ⇴ '\u21f4'
	"circleot":                            0x2299,  // ⊙ '\u2299'
	"circleplus":                          0x2295,  // ⊕ '\u2295'
	"circlepostalmark":                    0x3036,  // 〶 '\u3036'
	"circlering":                          0x229a,  // ⊚ '\u229a'
	"circletophalfblack":                  0x25d3,  // ◓ '\u25d3'
	"circleulquad":                        0x25f4,  // ◴ '\u25f4'
	"circleurquad":                        0x25f7,  // ◷ '\u25f7'
	"circleurquadblack":                   0x25d4,  // ◔ '\u25d4'
	"circlevertfill":                      0x25cd,  // ◍ '\u25cd'
	"circlewithlefthalfblack":             0x25d0,  // ◐ '\u25d0'
	"circlewithrighthalfblack":            0x25d1,  // ◑ '\u25d1'
	"circumflex":                          0x02c6,  // ˆ '\u02c6'
	"circumflex1":                         0xf003,  //  '\uf003'
	"circumflexbelowcmb":                  0x032d,  // ̭ '\u032d'
	"circumflexcmb":                       0x0302,  // ̂ '\u0302'
	"cirfnint":                            0x2a10,  // ⨐ '\u2a10'
	"cirmid":                              0x2aef,  // ⫯ '\u2aef'
	"cirscir":                             0x29c2,  // ⧂ '\u29c2'
	"clear":                               0x2327,  // ⌧ '\u2327'
	"clickalveolar":                       0x01c2,  // ǂ '\u01c2'
	"clickdental":                         0x01c0,  // ǀ '\u01c0'
	"clicklateral":                        0x01c1,  // ǁ '\u01c1'
	"clickretroflex":                      0x01c3,  // ǃ '\u01c3'
	"clockwise":                           0x27f3,  // ⟳ '\u27f3'
	"closedvarcap":                        0x2a4d,  // ⩍ '\u2a4d'
	"closedvarcup":                        0x2a4c,  // ⩌ '\u2a4c'
	"closedvarcupsmashprod":               0x2a50,  // ⩐ '\u2a50'
	"closure":                             0x2050,  // ⁐ '\u2050'
	"club":                                0x2663,  // ♣ '\u2663'
	"clubsuitwhite":                       0x2667,  // ♧ '\u2667'
	"cmcubedsquare":                       0x33a4,  // ㎤ '\u33a4'
	"cmonospace":                          0xff43,  // ｃ '\uff43'
	"cmsquaredsquare":                     0x33a0,  // ㎠ '\u33a0'
	"coarmenian":                          0x0581,  // ց '\u0581'
	"colon":                               0x003a,  // : ':'
	"coloneq":                             0x2254,  // ≔ '\u2254'
	"colonmonetary":                       0x20a1,  // ₡ '\u20a1'
	"colonmonospace":                      0xff1a,  // ： '\uff1a'
	"colonsmall":                          0xfe55,  // ﹕ '\ufe55'
	"colontriangularhalfmod":              0x02d1,  // ˑ '\u02d1'
	"colontriangularmod":                  0x02d0,  // ː '\u02d0'
	"comma":                               0x002c,  // , ','
	"commaabovecmb":                       0x0313,  // ̓ '\u0313'
	"commaaboverightcmb":                  0x0315,  // ̕ '\u0315'
	"commaaccent":                         0xf6c3,  //  '\uf6c3'
	"commaarmenian":                       0x055d,  // ՝ '\u055d'
	"commainferior":                       0xf6e1,  //  '\uf6e1'
	"commaminus":                          0x2a29,  // ⨩ '\u2a29'
	"commamonospace":                      0xff0c,  // ， '\uff0c'
	"commareversedabovecmb":               0x0314,  // ̔ '\u0314'
	"commasmall":                          0xfe50,  // ﹐ '\ufe50'
	"commasubnosp":                        0x0299,  // ʙ '\u0299'
	"commasuperior":                       0xf6e2,  //  '\uf6e2'
	"commaturnedabovecmb":                 0x0312,  // ̒ '\u0312'
	"commaturnedmod":                      0x02bb,  // ʻ '\u02bb'
	"complement":                          0x2201,  // ∁ '\u2201'
	"concavediamond":                      0x27e1,  // ⟡ '\u27e1'
	"concavediamondtickleft":              0x27e2,  // ⟢ '\u27e2'
	"concavediamondtickright":             0x27e3,  // ⟣ '\u27e3'
	"congdot":                             0x2a6d,  // ⩭ '\u2a6d'
	"congruent":                           0x2245,  // ≅ '\u2245'
	"conictaper":                          0x2332,  // ⌲ '\u2332'
	"conjquant":                           0x2a07,  // ⨇ '\u2a07'
	"contourintegral":                     0x222e,  // ∮ '\u222e'
	"control":                             0x2303,  // ⌃ '\u2303'
	"controlACK":                          0x0006,  //  '\x06'
	"controlBEL":                          0x0007,  //  '\a'
	"controlBS":                           0x0008,  //  '\b'
	"controlCAN":                          0x0018,  //  '\x18'
	"controlCR":                           0x000d,  //  '\r'
	"controlDC1":                          0x0011,  //  '\x11'
	"controlDC2":                          0x0012,  //  '\x12'
	"controlDC3":                          0x0013,  //  '\x13'
	"controlDC4":                          0x0014,  //  '\x14'
	"controlDEL":                          0x007f,  //  '\u007f'
	"controlDLE":                          0x0010,  //  '\x10'
	"controlEM":                           0x0019,  //  '\x19'
	"controlENQ":                          0x0005,  //  '\x05'
	"controlEOT":                          0x0004,  //  '\x04'
	"controlESC":                          0x001b,  //  '\x1b'
	"controlETB":                          0x0017,  //  '\x17'
	"controlETX":                          0x0003,  //  '\x03'
	"controlFF":                           0x000c,  //  '\f'
	"controlFS":                           0x001c,  //  '\x1c'
	"controlGS":                           0x001d,  //  '\x1d'
	"controlHT":                           0x0009,  //  '\t'
	"controlLF":                           0x000a,  //  '\n'
	"controlNAK":                          0x0015,  //  '\x15'
	"controlNULL":                         0x0000,  //  '\x00'
	"controlRS":                           0x001e,  //  '\x1e'
	"controlSI":                           0x000f,  //  '\x0f'
	"controlSO":                           0x000e,  //  '\x0e'
	"controlSOT":                          0x0002,  //  '\x02'
	"controlSTX":                          0x0001,  //  '\x01'
	"controlSUB":                          0x001a,  //  '\x1a'
	"controlSYN":                          0x0016,  //  '\x16'
	"controlUS":                           0x001f,  //  '\x1f'
	"controlVT":                           0x000b,  //  '\v'
	"coproduct":                           0x2a3f,  // ⨿ '\u2a3f'
	"coproductdisplay":                    0x2210,  // ∐ '\u2210'
	"copyright":                           0x00a9,  // © '\u00a9'
	"copyrightsans":                       0xf8e9,  //  '\uf8e9'
	"copyrightserif":                      0xf6d9,  //  '\uf6d9'
	"cornerbracketleft":                   0x300c,  // 「 '\u300c'
	"cornerbracketlefthalfwidth":          0xff62,  // ｢ '\uff62'
	"cornerbracketleftvertical":           0xfe41,  // ﹁ '\ufe41'
	"cornerbracketright":                  0x300d,  // 」 '\u300d'
	"cornerbracketrighthalfwidth":         0xff63,  // ｣ '\uff63'
	"cornerbracketrightvertical":          0xfe42,  // ﹂ '\ufe42'
	"corporationsquare":                   0x337f,  // ㍿ '\u337f'
	"cosquare":                            0x33c7,  // ㏇ '\u33c7'
	"coverkgsquare":                       0x33c6,  // ㏆ '\u33c6'
	"cparen":                              0x249e,  // ⒞ '\u249e'
	"cruzeiro":                            0x20a2,  // ₢ '\u20a2'
	"cstretch":                            0x0227,  // ȧ '\u0227'
	"cstretched":                          0x0297,  // ʗ '\u0297'
	"csub":                                0x2acf,  // ⫏ '\u2acf'
	"csube":                               0x2ad1,  // ⫑ '\u2ad1'
	"csup":                                0x2ad0,  // ⫐ '\u2ad0'
	"csupe":                               0x2ad2,  // ⫒ '\u2ad2'
	"cuberoot":                            0x221b,  // ∛ '\u221b'
	"cupbarcap":                           0x2a48,  // ⩈ '\u2a48'
	"cupdot":                              0x228d,  // ⊍ '\u228d'
	"cupleftarrow":                        0x228c,  // ⊌ '\u228c'
	"cupovercap":                          0x2a46,  // ⩆ '\u2a46'
	"cupvee":                              0x2a45,  // ⩅ '\u2a45'
	"curlyand":                            0x22cf,  // ⋏ '\u22cf'
	"curlyleft":                           0x21ab,  // ↫ '\u21ab'
	"curlyor":                             0x22ce,  // ⋎ '\u22ce'
	"curlyright":                          0x21ac,  // ↬ '\u21ac'
	"currency":                            0x00a4,  // ¤ '\u00a4'
	"curvearrowleftplus":                  0x293d,  // ⤽ '\u293d'
	"curvearrowrightminus":                0x293c,  // ⤼ '\u293c'
	"cwcirclearrow":                       0x2941,  // ⥁ '\u2941'
	"cwopencirclearrow":                   0x21bb,  // ↻ '\u21bb'
	"cwrightarcarrow":                     0x2938,  // ⤸ '\u2938'
	"cwundercurvearrow":                   0x293e,  // ⤾ '\u293e'
	"cyrBreve":                            0xf6d1,  //  '\uf6d1'
	"cyrFlex":                             0xf6d2,  //  '\uf6d2'
	"cyrbreve":                            0xf6d4,  //  '\uf6d4'
	"cyrflex":                             0xf6d5,  //  '\uf6d5'
	"d":                                   0x0064,  // d 'd'
	"daarmenian":                          0x0564,  // դ '\u0564'
	"dabengali":                           0x09a6,  // দ '\u09a6'
	"dadeva":                              0x0926,  // द '\u0926'
	"dadfinalarabic":                      0xfebe,  // ﺾ '\ufebe'
	"dadinitialarabic":                    0xfebf,  // ﺿ '\ufebf'
	"dadisolated":                         0xfebd,  // ﺽ '\ufebd'
	"dadmedialarabic":                     0xfec0,  // ﻀ '\ufec0'
	"dagesh":                              0x05bc,  // ּ '\u05bc'
	"dagger":                              0x2020,  // † '\u2020'
	"daggerdbl":                           0x2021,  // ‡ '\u2021'
	"dagujarati":                          0x0aa6,  // દ '\u0aa6'
	"dagurmukhi":                          0x0a26,  // ਦ '\u0a26'
	"dahiragana":                          0x3060,  // だ '\u3060'
	"dakatakana":                          0x30c0,  // ダ '\u30c0'
	"daletdagesh":                         0xfb33,  // דּ '\ufb33'
	"daleth":                              0x2138,  // ℸ '\u2138'
	"daletqamatshebrew":                   0x05d3,  // ד '\u05d3'
	"dalfinalarabic":                      0xfeaa,  // ﺪ '\ufeaa'
	"dalisolated":                         0xfea9,  // ﺩ '\ufea9'
	"dammaarabic":                         0x064f,  // ُ '\u064f'
	"dammaisolated":                       0xfe78,  // ﹸ '\ufe78'
	"dammalow":                            0xe821,  //  '\ue821'
	"dammamedial":                         0xfe79,  // ﹹ '\ufe79'
	"dammaonhamza":                        0xe835,  //  '\ue835'
	"dammatanisolated":                    0xfe72,  // ﹲ '\ufe72'
	"dammatanlow":                         0xe824,  //  '\ue824'
	"dammatanonhamza":                     0xe836,  //  '\ue836'
	"danda":                               0x0964,  // । '\u0964'
	"danger":                              0x2621,  // ☡ '\u2621'
	"dargalefthebrew":                     0x05a7,  // ֧ '\u05a7'
	"dashV":                               0x2ae3,  // ⫣ '\u2ae3'
	"dashVdash":                           0x27db,  // ⟛ '\u27db'
	"dashcolon":                           0x2239,  // ∹ '\u2239'
	"dashleftharpoondown":                 0x296b,  // ⥫ '\u296b'
	"dashrightharpoondown":                0x296d,  // ⥭ '\u296d'
	"dasiapneumatacyrilliccmb":            0x0485,  // ҅ '\u0485'
	"dbkarow":                             0x290f,  // ⤏ '\u290f'
	"dblGrave":                            0xf6d3,  //  '\uf6d3'
	"dblanglebracketleft":                 0x300a,  // 《 '\u300a'
	"dblanglebracketleftvertical":         0xfe3d,  // ︽ '\ufe3d'
	"dblanglebracketright":                0x300b,  // 》 '\u300b'
	"dblanglebracketrightvertical":        0xfe3e,  // ︾ '\ufe3e'
	"dblarchinvertedbelowcmb":             0x032b,  // ̫ '\u032b'
	"dblarrowdwn":                         0x21ca,  // ⇊ '\u21ca'
	"dblarrowheaddown":                    0x058a,  // ֊ '\u058a'
	"dblarrowheadleft":                    0x219e,  // ↞ '\u219e'
	"dblarrowheadright":                   0x21a0,  // ↠ '\u21a0'
	"dblarrowheadup":                      0x0588,  //  '\u0588'
	"dblarrowup":                          0x21c8,  // ⇈ '\u21c8'
	"dblbracketleft":                      0x27e6,  // ⟦ '\u27e6'
	"dblbracketright":                     0x27e7,  // ⟧ '\u27e7'
	"dbldanda":                            0x0965,  // ॥ '\u0965'
	"dblgrave":                            0xf6d6,  //  '\uf6d6'
	"dblgravecmb":                         0x030f,  // ̏ '\u030f'
	"dblintegral":                         0x222c,  // ∬ '\u222c'
	"dbllowlinecmb":                       0x0333,  // ̳ '\u0333'
	"dbloverlinecmb":                      0x033f,  // ̿ '\u033f'
	"dblprimemod":                         0x02ba,  // ʺ '\u02ba'
	"dblverticalbar":                      0x2016,  // ‖ '\u2016'
	"dblverticallineabovecmb":             0x030e,  // ̎ '\u030e'
	"dbopomofo":                           0x3109,  // ㄉ '\u3109'
	"dbsquare":                            0x33c8,  // ㏈ '\u33c8'
	"dcaron":                              0x010f,  // ď '\u010f'
	"dcaron1":                             0xf811,  //  '\uf811'
	"dcedilla":                            0x1e11,  // ḑ '\u1e11'
	"dcircle":                             0x24d3,  // ⓓ '\u24d3'
	"dcircumflexbelow":                    0x1e13,  // ḓ '\u1e13'
	"ddabengali":                          0x09a1,  // ড '\u09a1'
	"ddadeva":                             0x0921,  // ड '\u0921'
	"ddagujarati":                         0x0aa1,  // ડ '\u0aa1'
	"ddagurmukhi":                         0x0a21,  // ਡ '\u0a21'
	"ddalfinalarabic":                     0xfb89,  // ﮉ '\ufb89'
	"ddddot":                              0x20dc,  // ⃜ '\u20dc'
	"dddhadeva":                           0x095c,  // ड़ '\u095c'
	"dddot":                               0x20db,  // ⃛ '\u20db'
	"ddhabengali":                         0x09a2,  // ঢ '\u09a2'
	"ddhadeva":                            0x0922,  // ढ '\u0922'
	"ddhagujarati":                        0x0aa2,  // ઢ '\u0aa2'
	"ddhagurmukhi":                        0x0a22,  // ਢ '\u0a22'
	"ddotaccent":                          0x1e0b,  // ḋ '\u1e0b'
	"ddotbelow":                           0x1e0d,  // ḍ '\u1e0d'
	"ddots":                               0x22f1,  // ⋱ '\u22f1'
	"ddotseq":                             0x2a77,  // ⩷ '\u2a77'
	"decimalseparatorpersian":             0x066b,  // ٫ '\u066b'
	"defines":                             0x225c,  // ≜ '\u225c'
	"degree":                              0x00b0,  // ° '\u00b0'
	"degreekelvin":                        0x212a,  // K '\u212a'
	"dehihebrew":                          0x05ad,  // ֭ '\u05ad'
	"dehiragana":                          0x3067,  // で '\u3067'
	"deicoptic":                           0x03ef,  // ϯ '\u03ef'
	"dekatakana":                          0x30c7,  // デ '\u30c7'
	"delete":                              0x05ba,  // ֺ '\u05ba'
	"deleteleft":                          0x232b,  // ⌫ '\u232b'
	"deleteright":                         0x2326,  // ⌦ '\u2326'
	"delta":                               0x03b4,  // δ '\u03b4'
	"deltaturned":                         0x018d,  // ƍ '\u018d'
	"denominatorminusonenumeratorbengali": 0x09f8,  // ৸ '\u09f8'
	"dezh":                                0x02a4,  // ʤ '\u02a4'
	"dhabengali":                          0x09a7,  // ধ '\u09a7'
	"dhadeva":                             0x0927,  // ध '\u0927'
	"dhagujarati":                         0x0aa7,  // ધ '\u0aa7'
	"dhagurmukhi":                         0x0a27,  // ਧ '\u0a27'
	"dhook":                               0x0257,  // ɗ '\u0257'
	"diaeresis":                           0x0088,  //  '\u0088'
	"dialytikatonoscmb":                   0x0344,  // ̈́ '\u0344'
	"diameter":                            0x2300,  // ⌀ '\u2300'
	"diamond":                             0x2666,  // ♦ '\u2666'
	"diamondbotblack":                     0x2b19,  // ⬙ '\u2b19'
	"diamondcdot":                         0x27d0,  // ⟐ '\u27d0'
	"diamondleftarrow":                    0x291d,  // ⤝ '\u291d'
	"diamondleftarrowbar":                 0x291f,  // ⤟ '\u291f'
	"diamondleftblack":                    0x2b16,  // ⬖ '\u2b16'
	"diamondmath":                         0x22c4,  // ⋄ '\u22c4'
	"diamondrightblack":                   0x2b17,  // ⬗ '\u2b17'
	"diamondsuitwhite":                    0x2662,  // ♢ '\u2662'
	"diamondtopblack":                     0x2b18,  // ⬘ '\u2b18'
	"dicei":                               0x2680,  // ⚀ '\u2680'
	"diceii":                              0x2681,  // ⚁ '\u2681'
	"diceiii":                             0x2682,  // ⚂ '\u2682'
	"diceiv":                              0x2683,  // ⚃ '\u2683'
	"dicev":                               0x2684,  // ⚄ '\u2684'
	"dicevi":                              0x2685,  // ⚅ '\u2685'
	"dieresis":                            0x00a8,  // ¨ '\u00a8'
	"dieresis1":                           0xf005,  //  '\uf005'
	"dieresisacute":                       0xf6d7,  //  '\uf6d7'
	"dieresisbelowcmb":                    0x0324,  // ̤ '\u0324'
	"dieresiscmb":                         0x0308,  // ̈ '\u0308'
	"dieresisgrave":                       0xf6d8,  //  '\uf6d8'
	"dieresistonos":                       0x0385,  // ΅ '\u0385'
	"difference":                          0x224f,  // ≏ '\u224f'
	"dihiragana":                          0x3062,  // ぢ '\u3062'
	"dikatakana":                          0x30c2,  // ヂ '\u30c2'
	"disin":                               0x22f2,  // ⋲ '\u22f2'
	"disjquant":                           0x2a08,  // ⨈ '\u2a08'
	"dittomark":                           0x3003,  // 〃 '\u3003'
	"divide":                              0x00f7,  // ÷ '\u00f7'
	"dividemultiply":                      0x22c7,  // ⋇ '\u22c7'
	"divides":                             0x2223,  // ∣ '\u2223'
	"divisionslash":                       0x2215,  // ∕ '\u2215'
	"djecyrillic":                         0x0452,  // ђ '\u0452'
	"dkshade":                             0x2593,  // ▓ '\u2593'
	"dkshade1":                            0xf823,  //  '\uf823'
	"dlinebelow":                          0x1e0f,  // ḏ '\u1e0f'
	"dlsquare":                            0x3397,  // ㎗ '\u3397'
	"dmacron":                             0x0111,  // đ '\u0111'
	"dmonospace":                          0xff44,  // ｄ '\uff44'
	"dnblock":                             0x2584,  // ▄ '\u2584'
	"dneightblock":                        0x2581,  // ▁ '\u2581'
	"dnfiveeighthblock":                   0x2585,  // ▅ '\u2585'
	"dnquarterblock":                      0x2582,  // ▂ '\u2582'
	"dnseveneighthblock":                  0x2587,  // ▇ '\u2587'
	"dnthreeeighthblock":                  0x2583,  // ▃ '\u2583'
	"dnthreequarterblock":                 0x2586,  // ▆ '\u2586'
	"dochadathai":                         0x0e0e,  // ฎ '\u0e0e'
	"dodekthai":                           0x0e14,  // ด '\u0e14'
	"dohiragana":                          0x3069,  // ど '\u3069'
	"dokatakana":                          0x30c9,  // ド '\u30c9'
	"dollar":                              0x0024,  // $ '$'
	"dollarinferior":                      0xf6e3,  //  '\uf6e3'
	"dollarmonospace":                     0xff04,  // ＄ '\uff04'
	"dollaroldstyle":                      0xf724,  //  '\uf724'
	"dollarsmall":                         0xfe69,  // ﹩ '\ufe69'
	"dollarsuperior":                      0xf6e4,  //  '\uf6e4'
	"dong":                                0x20ab,  // ₫ '\u20ab'
	"dorusquare":                          0x3326,  // ㌦ '\u3326'
	"dotaccent":                           0x02d9,  // ˙ '\u02d9'
	"dotaccentcmb":                        0x0307,  // ̇ '\u0307'
	"dotbelowcomb":                        0x0323,  // ̣ '\u0323'
	"dotcircle1":                          0xf820,  //  '\uf820'
	"dotequiv":                            0x2a67,  // ⩧ '\u2a67'
	"dotkatakana":                         0x30fb,  // ・ '\u30fb'
	"dotlessi":                            0x0131,  // ı '\u0131'
	"dotlessj":                            0xf6be,  //  '\uf6be'
	"dotlessjstrokehook":                  0x0284,  // ʄ '\u0284'
	"dotmath":                             0x22c5,  // ⋅ '\u22c5'
	"dotminus":                            0x2238,  // ∸ '\u2238'
	"dotplus":                             0x2214,  // ∔ '\u2214'
	"dotsim":                              0x2a6a,  // ⩪ '\u2a6a'
	"dotsminusdots":                       0x223a,  // ∺ '\u223a'
	"dottedcircle":                        0x25cc,  // ◌ '\u25cc'
	"dottedsquare":                        0x2b1a,  // ⬚ '\u2b1a'
	"dottimes":                            0x2a30,  // ⨰ '\u2a30'
	"doublebarvee":                        0x2a62,  // ⩢ '\u2a62'
	"doubleplus":                          0x29fa,  // ⧺ '\u29fa'
	"downarrowbar":                        0x2913,  // ⤓ '\u2913'
	"downarrowbarred":                     0x2908,  // ⤈ '\u2908'
	"downfishtail":                        0x297f,  // ⥿ '\u297f'
	"downharpoonleftbar":                  0x2959,  // ⥙ '\u2959'
	"downharpoonrightbar":                 0x2955,  // ⥕ '\u2955'
	"downharpoonsleftright":               0x2965,  // ⥥ '\u2965'
	"downrightcurvedarrow":                0x2935,  // ⤵ '\u2935'
	"downslope":                           0x29f9,  // ⧹ '\u29f9'
	"downtackbelowcmb":                    0x031e,  // ̞ '\u031e'
	"downtackmod":                         0x02d5,  // ˕ '\u02d5'
	"downtriangleleftblack":               0x29e8,  // ⧨ '\u29e8'
	"downtrianglerightblack":              0x29e9,  // ⧩ '\u29e9'
	"downuparrows":                        0x21f5,  // ⇵ '\u21f5'
	"downupharpoonsleftright":             0x296f,  // ⥯ '\u296f'
	"downzigzagarrow":                     0x21af,  // ↯ '\u21af'
	"dparen":                              0x249f,  // ⒟ '\u249f'
	"drbkarow":                            0x2910,  // ⤐ '\u2910'
	"dsol":                                0x29f6,  // ⧶ '\u29f6'
	"dsub":                                0x2a64,  // ⩤ '\u2a64'
	"dsuperior":                           0xf6eb,  //  '\uf6eb'
	"dtail":                               0x0256,  // ɖ '\u0256'
	"dtopbar":                             0x018c,  // ƌ '\u018c'
	"dualmap":                             0x29df,  // ⧟ '\u29df'
	"duhiragana":                          0x3065,  // づ '\u3065'
	"dukatakana":                          0x30c5,  // ヅ '\u30c5'
	"dyogh":                               0x0234,  // ȴ '\u0234'
	"dz":                                  0x01f3,  // ǳ '\u01f3'
	"dzaltone":                            0x02a3,  // ʣ '\u02a3'
	"dzcaron":                             0x01c6,  // ǆ '\u01c6'
	"dzcurl":                              0x02a5,  // ʥ '\u02a5'
	"dzeabkhasiancyrillic":                0x04e1,  // ӡ '\u04e1'
	"dzhecyrillic":                        0x045f,  // џ '\u045f'
	"e":                                   0x0065,  // e 'e'
	"eacute":                              0x00e9,  // é '\u00e9'
	"earth":                               0x2641,  // ♁ '\u2641'
	"ebengali":                            0x098f,  // এ '\u098f'
	"ebopomofo":                           0x311c,  // ㄜ '\u311c'
	"ebreve":                              0x0115,  // ĕ '\u0115'
	"ecandradeva":                         0x090d,  // ऍ '\u090d'
	"ecandragujarati":                     0x0a8d,  // ઍ '\u0a8d'
	"ecandravowelsigndeva":                0x0945,  // ॅ '\u0945'
	"ecandravowelsigngujarati":            0x0ac5,  // ૅ '\u0ac5'
	"ecaron":                              0x011b,  // ě '\u011b'
	"ecedillabreve":                       0x1e1d,  // ḝ '\u1e1d'
	"echarmenian":                         0x0565,  // ե '\u0565'
	"echyiwnarmenian":                     0x0587,  // և '\u0587'
	"ecircle":                             0x24d4,  // ⓔ '\u24d4'
	"ecircumflex":                         0x00ea,  // ê '\u00ea'
	"ecircumflexacute":                    0x1ebf,  // ế '\u1ebf'
	"ecircumflexbelow":                    0x1e19,  // ḙ '\u1e19'
	"ecircumflexdotbelow":                 0x1ec7,  // ệ '\u1ec7'
	"ecircumflexgrave":                    0x1ec1,  // ề '\u1ec1'
	"ecircumflexhookabove":                0x1ec3,  // ể '\u1ec3'
	"ecircumflextilde":                    0x1ec5,  // ễ '\u1ec5'
	"ecyrillic":                           0x0454,  // є '\u0454'
	"edblgrave":                           0x0205,  // ȅ '\u0205'
	"edeva":                               0x090f,  // ए '\u090f'
	"edieresis":                           0x00eb,  // ë '\u00eb'
	"edotaccent":                          0x0117,  // ė '\u0117'
	"edotbelow":                           0x1eb9,  // ẹ '\u1eb9'
	"eegurmukhi":                          0x0a0f,  // ਏ '\u0a0f'
	"eematragurmukhi":                     0x0a47,  // ੇ '\u0a47'
	"egrave":                              0x00e8,  // è '\u00e8'
	"egsdot":                              0x2a98,  // ⪘ '\u2a98'
	"egujarati":                           0x0a8f,  // એ '\u0a8f'
	"eharmenian":                          0x0567,  // է '\u0567'
	"ehbopomofo":                          0x311d,  // ㄝ '\u311d'
	"ehiragana":                           0x3048,  // え '\u3048'
	"ehookabove":                          0x1ebb,  // ẻ '\u1ebb'
	"eibopomofo":                          0x311f,  // ㄟ '\u311f'
	"eight":                               0x0038,  // 8 '8'
	"eightbengali":                        0x09ee,  // ৮ '\u09ee'
	"eightdeva":                           0x096e,  // ८ '\u096e'
	"eighteencircle":                      0x2471,  // ⑱ '\u2471'
	"eighteenparen":                       0x2485,  // ⒅ '\u2485'
	"eighteenperiod":                      0x2499,  // ⒙ '\u2499'
	"eightgujarati":                       0x0aee,  // ૮ '\u0aee'
	"eightgurmukhi":                       0x0a6e,  // ੮ '\u0a6e'
	"eighthackarabic":                     0x0668,  // ٨ '\u0668'
	"eighthangzhou":                       0x3028,  // 〨 '\u3028'
	"eighthnotebeamed":                    0x266b,  // ♫ '\u266b'
	"eightideographicparen":               0x3227,  // ㈧ '\u3227'
	"eightinferior":                       0x2088,  // ₈ '\u2088'
	"eightmonospace":                      0xff18,  // ８ '\uff18'
	"eightoldstyle":                       0xf738,  //  '\uf738'
	"eightparen":                          0x247b,  // ⑻ '\u247b'
	"eightperiod":                         0x248f,  // ⒏ '\u248f'
	"eightpersian":                        0x06f8,  // ۸ '\u06f8'
	"eightroman":                          0x2177,  // ⅷ '\u2177'
	"eightsuperior":                       0x2078,  // ⁸ '\u2078'
	"eightthai":                           0x0e58,  // ๘ '\u0e58'
	"einvertedbreve":                      0x0207,  // ȇ '\u0207'
	"eiotifiedcyrillic":                   0x0465,  // ѥ '\u0465'
	"ekatakana":                           0x30a8,  // エ '\u30a8'
	"ekatakanahalfwidth":                  0xff74,  // ｴ '\uff74'
	"ekonkargurmukhi":                     0x0a74,  // ੴ '\u0a74'
	"ekorean":                             0x3154,  // ㅔ '\u3154'
	"elcyrillic":                          0x043b,  // л '\u043b'
	"element":                             0x2208,  // ∈ '\u2208'
	"elevencircle":                        0x246a,  // ⑪ '\u246a'
	"elevenparen":                         0x247e,  // ⑾ '\u247e'
	"elevenperiod":                        0x2492,  // ⒒ '\u2492'
	"elevenroman":                         0x217a,  // ⅺ '\u217a'
	"elinters":                            0x23e7,  // ⏧ '\u23e7'
	"ellipsis":                            0x2026,  // … '\u2026'
	"ellipsisvertical":                    0x22ee,  // ⋮ '\u22ee'
	"elsdot":                              0x2a97,  // ⪗ '\u2a97'
	"emacron":                             0x0113,  // ē '\u0113'
	"emacronacute":                        0x1e17,  // ḗ '\u1e17'
	"emacrongrave":                        0x1e15,  // ḕ '\u1e15'
	"emcyrillic":                          0x043c,  // м '\u043c'
	"emdash":                              0x2014,  // — '\u2014'
	"emdashvertical":                      0xfe31,  // ︱ '\ufe31'
	"emonospace":                          0xff45,  // ｅ '\uff45'
	"emphasismarkarmenian":                0x055b,  // ՛ '\u055b'
	"emptyset":                            0x2205,  // ∅ '\u2205'
	"emptysetoarr":                        0x29b3,  // ⦳ '\u29b3'
	"emptysetoarrl":                       0x29b4,  // ⦴ '\u29b4'
	"emptysetobar":                        0x29b1,  // ⦱ '\u29b1'
	"emptysetocirc":                       0x29b2,  // ⦲ '\u29b2'
	"emptyslot":                           0xd801,  //  '\ufffd'
	"emquad":                              0x2001,  //  '\u2001'
	"emspace":                             0x2003,  //  '\u2003'
	"enbopomofo":                          0x3123,  // ㄣ '\u3123'
	"enclosediamond":                      0x20df,  // ⃟ '\u20df'
	"enclosesquare":                       0x20de,  // ⃞ '\u20de'
	"enclosetriangle":                     0x20e4,  // ⃤ '\u20e4'
	"endash":                              0x2013,  // – '\u2013'
	"endashvertical":                      0xfe32,  // ︲ '\ufe32'
	"endescendercyrillic":                 0x04a3,  // ң '\u04a3'
	"eng":                                 0x014b,  // ŋ '\u014b'
	"engbopomofo":                         0x3125,  // ㄥ '\u3125'
	"enghecyrillic":                       0x04a5,  // ҥ '\u04a5'
	"enhookcyrillic":                      0x04c8,  // ӈ '\u04c8'
	"enquad":                              0x2000,  //  '\u2000'
	"enspace":                             0x2002,  //  '\u2002'
	"eogonek":                             0x0119,  // ę '\u0119'
	"eokorean":                            0x3153,  // ㅓ '\u3153'
	"eopen":                               0x025b,  // ɛ '\u025b'
	"eopenclosed":                         0x029a,  // ʚ '\u029a'
	"eopenreversed":                       0x025c,  // ɜ '\u025c'
	"eopenreversedclosed":                 0x025e,  // ɞ '\u025e'
	"eopenreversedhook":                   0x025d,  // ɝ '\u025d'
	"eparen":                              0x24a0,  // ⒠ '\u24a0'
	"eparsl":                              0x29e3,  // ⧣ '\u29e3'
	"epsilon":                             0x03b5,  // ε '\u03b5'
	"epsilon1":                            0x03f5,  // ϵ '\u03f5'
	"epsilonclosed":                       0x022a,  // Ȫ '\u022a'
	"epsiloninv":                          0x03f6,  // ϶ '\u03f6'
	"epsilontonos":                        0x03ad,  // έ '\u03ad'
	"eqcolon":                             0x2255,  // ≕ '\u2255'
	"eqdef":                               0x225d,  // ≝ '\u225d'
	"eqdot":                               0x2a66,  // ⩦ '\u2a66'
	"eqeq":                                0x2a75,  // ⩵ '\u2a75'
	"eqeqeq":                              0x2a76,  // ⩶ '\u2a76'
	"eqgtr":                               0x22dd,  // ⋝ '\u22dd'
	"eqless":                              0x22dc,  // ⋜ '\u22dc'
	"eqqgtr":                              0x2a9a,  // ⪚ '\u2a9a'
	"eqqless":                             0x2a99,  // ⪙ '\u2a99'
	"eqqplus":                             0x2a71,  // ⩱ '\u2a71'
	"eqqsim":                              0x2a73,  // ⩳ '\u2a73'
	"eqqslantgtr":                         0x2a9c,  // ⪜ '\u2a9c'
	"eqqslantless":                        0x2a9b,  // ⪛ '\u2a9b'
	"equal":                               0x003d,  // = '='
	"equalleftarrow":                      0x2b40,  // ⭀ '\u2b40'
	"equalmonospace":                      0xff1d,  // ＝ '\uff1d'
	"equalorfollows":                      0x22df,  // ⋟ '\u22df'
	"equalorgreater":                      0x2a96,  // ⪖ '\u2a96'
	"equalorless":                         0x2a95,  // ⪕ '\u2a95'
	"equalorprecedes":                     0x22de,  // ⋞ '\u22de'
	"equalorsimilar":                      0x2242,  // ≂ '\u2242'
	"equalparallel":                       0x22d5,  // ⋕ '\u22d5'
	"equalrightarrow":                     0x2971,  // ⥱ '\u2971'
	"equalsmall":                          0xfe66,  // ﹦ '\ufe66'
	"equalsub":                            0x208c,  // ₌ '\u208c'
	"equalsuperior":                       0x207c,  // ⁼ '\u207c'
	"equivDD":                             0x2a78,  // ⩸ '\u2a78'
	"equivVert":                           0x2a68,  // ⩨ '\u2a68'
	"equivVvert":                          0x2a69,  // ⩩ '\u2a69'
	"equivalence":                         0x2261,  // ≡ '\u2261'
	"equivasymptotic":                     0x224d,  // ≍ '\u224d'
	"eqvparsl":                            0x29e5,  // ⧥ '\u29e5'
	"erbopomofo":                          0x3126,  // ㄦ '\u3126'
	"ercyrillic":                          0x0440,  // р '\u0440'
	"ereversed":                           0x0258,  // ɘ '\u0258'
	"ereversedcyrillic":                   0x044d,  // э '\u044d'
	"errbarblackcircle":                   0x29f3,  // ⧳ '\u29f3'
	"errbarblackdiamond":                  0x29f1,  // ⧱ '\u29f1'
	"errbarblacksquare":                   0x29ef,  // ⧯ '\u29ef'
	"errbarcircle":                        0x29f2,  // ⧲ '\u29f2'
	"errbardiamond":                       0x29f0,  // ⧰ '\u29f0'
	"errbarsquare":                        0x29ee,  // ⧮ '\u29ee'
	"escyrillic":                          0x0441,  // с '\u0441'
	"esdescendercyrillic":                 0x04ab,  // ҫ '\u04ab'
	"esh":                                 0x0283,  // ʃ '\u0283'
	"eshcurl":                             0x0286,  // ʆ '\u0286'
	"eshortdeva":                          0x090e,  // ऎ '\u090e'
	"eshortvowelsigndeva":                 0x0946,  // ॆ '\u0946'
	"eshreversedloop":                     0x01aa,  // ƪ '\u01aa'
	"eshsquatreversed":                    0x0285,  // ʅ '\u0285'
	"esmallhiragana":                      0x3047,  // ぇ '\u3047'
	"esmallkatakana":                      0x30a7,  // ェ '\u30a7'
	"esmallkatakanahalfwidth":             0xff6a,  // ｪ '\uff6a'
	"estimated":                           0x212e,  // ℮ '\u212e'
	"esuperior":                           0xf6ec,  //  '\uf6ec'
	"eta":                                 0x03b7,  // η '\u03b7'
	"etarmenian":                          0x0568,  // ը '\u0568'
	"etatonos":                            0x03ae,  // ή '\u03ae'
	"eth":                                 0x00f0,  // ð '\u00f0'
	"etilde":                              0x1ebd,  // ẽ '\u1ebd'
	"etildebelow":                         0x1e1b,  // ḛ '\u1e1b'
	"etnahtalefthebrew":                   0x0591,  // ֑ '\u0591'
	"eturned":                             0x01dd,  // ǝ '\u01dd'
	"eukorean":                            0x3161,  // ㅡ '\u3161'
	"eurocurrency":                        0x20a0,  // ₠ '\u20a0'
	"evowelsignbengali":                   0x09c7,  // ে '\u09c7'
	"evowelsigndeva":                      0x0947,  // े '\u0947'
	"evowelsigngujarati":                  0x0ac7,  // ે '\u0ac7'
	"exclam":                              0x0021,  // ! '!'
	"exclamarmenian":                      0x055c,  // ՜ '\u055c'
	"exclamdbl":                           0x203c,  // ‼ '\u203c'
	"exclamdown":                          0x00a1,  // ¡ '\u00a1'
	"exclamdownsmall":                     0xf7a1,  //  '\uf7a1'
	"exclammonospace":                     0xff01,  // ！ '\uff01'
	"exclamsmall":                         0xf721,  //  '\uf721'
	"existential":                         0x2203,  // ∃ '\u2203'
	"ezh":                                 0x0292,  // ʒ '\u0292'
	"ezhcaron":                            0x01ef,  // ǯ '\u01ef'
	"ezhcurl":                             0x0293,  // ʓ '\u0293'
	"ezhreversed":                         0x01b9,  // ƹ '\u01b9'
	"ezhtail":                             0x01ba,  // ƺ '\u01ba'
	"f":                                   0x0066,  // f 'f'
	"f70e":                                0xf70e,  //  '\uf70e'
	"f70a":                                0xf70a,  //  '\uf70a'
	"f70c":                                0xf70c,  //  '\uf70c'
	"f70d":                                0xf70d,  //  '\uf70d'
	"f70b":                                0xf70b,  //  '\uf70b'
	"f70f":                                0xf70f,  //  '\uf70f'
	"f71c":                                0xf71c,  //  '\uf71c'
	"f71a":                                0xf71a,  //  '\uf71a'
	"f71d":                                0xf71d,  //  '\uf71d'
	"f700":                                0xf700,  //  '\uf700'
	"f701":                                0xf701,  //  '\uf701'
	"f702":                                0xf702,  //  '\uf702'
	"f703":                                0xf703,  //  '\uf703'
	"f704":                                0xf704,  //  '\uf704'
	"f705":                                0xf705,  //  '\uf705'
	"f706":                                0xf706,  //  '\uf706'
	"f707":                                0xf707,  //  '\uf707'
	"f708":                                0xf708,  //  '\uf708'
	"f709":                                0xf709,  //  '\uf709'
	"f710":                                0xf710,  //  '\uf710'
	"f711":                                0xf711,  //  '\uf711'
	"f712":                                0xf712,  //  '\uf712'
	"f713":                                0xf713,  //  '\uf713'
	"f714":                                0xf714,  //  '\uf714'
	"f715":                                0xf715,  //  '\uf715'
	"f716":                                0xf716,  //  '\uf716'
	"f717":                                0xf717,  //  '\uf717'
	"f718":                                0xf718,  //  '\uf718'
	"f719":                                0xf719,  //  '\uf719'
	"fadeva":                              0x095e,  // फ़ '\u095e'
	"fagurmukhi":                          0x0a5e,  // ਫ਼ '\u0a5e'
	"fahrenheit":                          0x2109,  // ℉ '\u2109'
	"farsiyeh":                            0x06cc,  // ی '\u06cc'
	"farsiyehfinal":                       0xfbfd,  // ﯽ '\ufbfd'
	"farsiyehisolated":                    0xfbfc,  // ﯼ '\ufbfc'
	"fathahontatweel":                     0xfe77,  // ﹷ '\ufe77'
	"fathaisolated":                       0xfe76,  // ﹶ '\ufe76'
	"fathalow":                            0xe820,  //  '\ue820'
	"fathalowarabic":                      0x064e,  // َ '\u064e'
	"fathaonhamza":                        0xe832,  //  '\ue832'
	"fathatanisolated":                    0xfe70,  // ﹰ '\ufe70'
	"fathatanlow":                         0xe823,  //  '\ue823'
	"fathatanonhamza":                     0xe833,  //  '\ue833'
	"fbopomofo":                           0x3108,  // ㄈ '\u3108'
	"fbowtie":                             0x29d3,  // ⧓ '\u29d3'
	"fcircle":                             0x24d5,  // ⓕ '\u24d5'
	"fcmp":                                0x2a3e,  // ⨾ '\u2a3e'
	"fdiagovnearrow":                      0x292f,  // ⤯ '\u292f'
	"fdiagovrdiag":                        0x292c,  // ⤬ '\u292c'
	"fdotaccent":                          0x1e1f,  // ḟ '\u1e1f'
	"feharmenian":                         0x0586,  // ֆ '\u0586'
	"fehfinalarabic":                      0xfed2,  // ﻒ '\ufed2'
	"fehinitialarabic":                    0xfed3,  // ﻓ '\ufed3'
	"fehisolated":                         0xfed1,  // ﻑ '\ufed1'
	"fehmedialarabic":                     0xfed4,  // ﻔ '\ufed4'
	"fehwithalefmaksuraisolated":          0xfc31,  // ﰱ '\ufc31'
	"fehwithyehisolated":                  0xfc32,  // ﰲ '\ufc32'
	"feicoptic":                           0x03e5,  // ϥ '\u03e5'
	"female":                              0x2640,  // ♀ '\u2640'
	"ff":                                  0xfb00,  // ﬀ '\ufb00'
	"ffi":                                 0xfb03,  // ﬃ '\ufb03'
	"ffl":                                 0xfb04,  // ﬄ '\ufb04'
	"fi":                                  0xfb01,  // ﬁ '\ufb01'
	"fifteencircle":                       0x246e,  // ⑮ '\u246e'
	"fifteenparen":                        0x2482,  // ⒂ '\u2482'
	"fifteenperiod":                       0x2496,  // ⒖ '\u2496'
	"figuredash":                          0x2012,  // ‒ '\u2012'
	"figurespace":                         0x2007,  //  '\u2007'
	"finalkafdageshhebrew":                0xfb3a,  // ךּ '\ufb3a'
	"finalkafwithqamats":                  0xe803,  //  '\ue803'
	"finalkafwithsheva":                   0xe802,  //  '\ue802'
	"finalmemhebrew":                      0x05dd,  // ם '\u05dd'
	"finalpehebrew":                       0x05e3,  // ף '\u05e3'
	"finaltsadi":                          0x05e5,  // ץ '\u05e5'
	"fint":                                0x2a0f,  // ⨏ '\u2a0f'
	"firsttonechinese":                    0x02c9,  // ˉ '\u02c9'
	"fisheye":                             0x25c9,  // ◉ '\u25c9'
	"five":                                0x0035,  // 5 '5'
	"fivearabic":                          0x0665,  // ٥ '\u0665'
	"fivebengali":                         0x09eb,  // ৫ '\u09eb'
	"fivedeva":                            0x096b,  // ५ '\u096b'
	"fiveeighths":                         0x215d,  // ⅝ '\u215d'
	"fivegujarati":                        0x0aeb,  // ૫ '\u0aeb'
	"fivegurmukhi":                        0x0a6b,  // ੫ '\u0a6b'
	"fivehangzhou":                        0x3025,  // 〥 '\u3025'
	"fiveideographicparen":                0x3224,  // ㈤ '\u3224'
	"fiveinferior":                        0x2085,  // ₅ '\u2085'
	"fivemonospace":                       0xff15,  // ５ '\uff15'
	"fiveoldstyle":                        0xf735,  //  '\uf735'
	"fiveparen":                           0x2478,  // ⑸ '\u2478'
	"fiveperiod":                          0x248c,  // ⒌ '\u248c'
	"fivepersian":                         0x06f5,  // ۵ '\u06f5'
	"fiveroman":                           0x2174,  // ⅴ '\u2174'
	"fivesixth":                           0x215a,  // ⅚ '\u215a'
	"fivesuperior":                        0x2075,  // ⁵ '\u2075'
	"fivethai":                            0x0e55,  // ๕ '\u0e55'
	"fl":                                  0xfb02,  // ﬂ '\ufb02'
	"floorleft":                           0x230a,  // ⌊ '\u230a'
	"floorright":                          0x230b,  // ⌋ '\u230b'
	"florin":                              0x0192,  // ƒ '\u0192'
	"fltns":                               0x23e5,  // ⏥ '\u23e5'
	"fmonospace":                          0xff46,  // ｆ '\uff46'
	"fmsquare":                            0x3399,  // ㎙ '\u3399'
	"fofanthai":                           0x0e1f,  // ฟ '\u0e1f'
	"fofathai":                            0x0e1d,  // ฝ '\u0e1d'
	"follownotdbleqv":                     0x2aba,  // ⪺ '\u2aba'
	"follownotslnteql":                    0x2ab6,  // ⪶ '\u2ab6'
	"followornoteqvlnt":                   0x22e9,  // ⋩ '\u22e9'
	"followsequal":                        0x2ab0,  // ⪰ '\u2ab0'
	"followsorcurly":                      0x227d,  // ≽ '\u227d'
	"followsorequal":                      0x227f,  // ≿ '\u227f'
	"fongmanthai":                         0x0e4f,  // ๏ '\u0e4f'
	"forces":                              0x22a9,  // ⊩ '\u22a9'
	"forcesbar":                           0x22aa,  // ⊪ '\u22aa'
	"fork":                                0x22d4,  // ⋔ '\u22d4'
	"forks":                               0x2adc,  // ⫝̸ '\u2adc'
	"forksnot":                            0x2add,  // ⫝ '\u2add'
	"forkv":                               0x2ad9,  // ⫙ '\u2ad9'
	"four":                                0x0034,  // 4 '4'
	"fourarabic":                          0x0664,  // ٤ '\u0664'
	"fourbengali":                         0x09ea,  // ৪ '\u09ea'
	"fourdeva":                            0x096a,  // ४ '\u096a'
	"fourfifths":                          0x2158,  // ⅘ '\u2158'
	"fourgujarati":                        0x0aea,  // ૪ '\u0aea'
	"fourgurmukhi":                        0x0a6a,  // ੪ '\u0a6a'
	"fourhangzhou":                        0x3024,  // 〤 '\u3024'
	"fourideographicparen":                0x3223,  // ㈣ '\u3223'
	"fourinferior":                        0x2084,  // ₄ '\u2084'
	"fourmonospace":                       0xff14,  // ４ '\uff14'
	"fournumeratorbengali":                0x09f7,  // ৷ '\u09f7'
	"fouroldstyle":                        0xf734,  //  '\uf734'
	"fourparen":                           0x2477,  // ⑷ '\u2477'
	"fourperemspace":                      0x2005,  //  '\u2005'
	"fourperiod":                          0x248b,  // ⒋ '\u248b'
	"fourpersian":                         0x06f4,  // ۴ '\u06f4'
	"fourroman":                           0x2173,  // ⅳ '\u2173'
	"foursuperior":                        0x2074,  // ⁴ '\u2074'
	"fourteencircle":                      0x246d,  // ⑭ '\u246d'
	"fourteenparen":                       0x2481,  // ⒁ '\u2481'
	"fourteenperiod":                      0x2495,  // ⒕ '\u2495'
	"fourthai":                            0x0e54,  // ๔ '\u0e54'
	"fourthroot":                          0x221c,  // ∜ '\u221c'
	"fourthtonechinese":                   0x02cb,  // ˋ '\u02cb'
	"fourvdots":                           0x2999,  // ⦙ '\u2999'
	"fparen":                              0x24a1,  // ⒡ '\u24a1'
	"fraction":                            0x2044,  // ⁄ '\u2044'
	"franc":                               0x20a3,  // ₣ '\u20a3'
	"fronted":                             0x024b,  // ɋ '\u024b'
	"fullouterjoin":                       0x27d7,  // ⟗ '\u27d7'
	"g":                                   0x0067,  // g 'g'
	"gabengali":                           0x0997,  // গ '\u0997'
	"gacute":                              0x01f5,  // ǵ '\u01f5'
	"gadeva":                              0x0917,  // ग '\u0917'
	"gafarabic":                           0x06af,  // گ '\u06af'
	"gaffinalarabic":                      0xfb93,  // ﮓ '\ufb93'
	"gafinitialarabic":                    0xfb94,  // ﮔ '\ufb94'
	"gafisolated":                         0xfb92,  // ﮒ '\ufb92'
	"gafmedialarabic":                     0xfb95,  // ﮕ '\ufb95'
	"gagujarati":                          0x0a97,  // ગ '\u0a97'
	"gagurmukhi":                          0x0a17,  // ਗ '\u0a17'
	"gahiragana":                          0x304c,  // が '\u304c'
	"gakatakana":                          0x30ac,  // ガ '\u30ac'
	"gamma":                               0x03b3,  // γ '\u03b3'
	"gammalatinsmall":                     0x0263,  // ɣ '\u0263'
	"gammasuperior":                       0x02e0,  // ˠ '\u02e0'
	"gangiacoptic":                        0x03eb,  // ϫ '\u03eb'
	"gbopomofo":                           0x310d,  // ㄍ '\u310d'
	"gbreve":                              0x011f,  // ğ '\u011f'
	"gcaron":                              0x01e7,  // ǧ '\u01e7'
	"gcedilla":                            0x0123,  // ģ '\u0123'
	"gcircle":                             0x24d6,  // ⓖ '\u24d6'
	"gcircumflex":                         0x011d,  // ĝ '\u011d'
	"gdot":                                0x0121,  // ġ '\u0121'
	"gebar":                               0x03cf,  // Ϗ '\u03cf'
	"gehiragana":                          0x3052,  // げ '\u3052'
	"gekatakana":                          0x30b2,  // ゲ '\u30b2'
	"geomequivalent":                      0x224e,  // ≎ '\u224e'
	"geometricallyequal":                  0x2251,  // ≑ '\u2251'
	"geqqslant":                           0x2afa,  // ⫺ '\u2afa'
	"gereshaccenthebrew":                  0x059c,  // ֜ '\u059c'
	"gereshhebrew":                        0x05f3,  // ׳ '\u05f3'
	"gereshmuqdamhebrew":                  0x059d,  // ֝ '\u059d'
	"germandbls":                          0x00df,  // ß '\u00df'
	"gershayimaccenthebrew":               0x059e,  // ֞ '\u059e'
	"gershayimhebrew":                     0x05f4,  // ״ '\u05f4'
	"gescc":                               0x2aa9,  // ⪩ '\u2aa9'
	"gesdot":                              0x2a80,  // ⪀ '\u2a80'
	"gesdoto":                             0x2a82,  // ⪂ '\u2a82'
	"gesdotol":                            0x2a84,  // ⪄ '\u2a84'
	"gesles":                              0x2a94,  // ⪔ '\u2a94'
	"getamark":                            0x3013,  // 〓 '\u3013'
	"ggg":                                 0x22d9,  // ⋙ '\u22d9'
	"gggnest":                             0x2af8,  // ⫸ '\u2af8'
	"ghabengali":                          0x0998,  // ঘ '\u0998'
	"ghadarmenian":                        0x0572,  // ղ '\u0572'
	"ghadeva":                             0x0918,  // घ '\u0918'
	"ghagujarati":                         0x0a98,  // ઘ '\u0a98'
	"ghagurmukhi":                         0x0a18,  // ਘ '\u0a18'
	"ghainarabic":                         0x063a,  // غ '\u063a'
	"ghainfinalarabic":                    0xfece,  // ﻎ '\ufece'
	"ghaininitialarabic":                  0xfecf,  // ﻏ '\ufecf'
	"ghainisolated":                       0xfecd,  // ﻍ '\ufecd'
	"ghainmedialarabic":                   0xfed0,  // ﻐ '\ufed0'
	"ghemiddlehookcyrillic":               0x0495,  // ҕ '\u0495'
	"ghestrokecyrillic":                   0x0493,  // ғ '\u0493'
	"gheupturncyrillic":                   0x0491,  // ґ '\u0491'
	"ghhadeva":                            0x095a,  // ग़ '\u095a'
	"ghhagurmukhi":                        0x0a5a,  // ਗ਼ '\u0a5a'
	"ghook":                               0x0260,  // ɠ '\u0260'
	"ghzsquare":                           0x3393,  // ㎓ '\u3393'
	"gihiragana":                          0x304e,  // ぎ '\u304e'
	"gikatakana":                          0x30ae,  // ギ '\u30ae'
	"gimarmenian":                         0x0563,  // գ '\u0563'
	"gimel":                               0x2137,  // ℷ '\u2137'
	"gimeldageshhebrew":                   0xfb32,  // גּ '\ufb32'
	"gjecyrillic":                         0x0453,  // ѓ '\u0453'
	"glE":                                 0x2a92,  // ⪒ '\u2a92'
	"gla":                                 0x2aa5,  // ⪥ '\u2aa5'
	"gleichstark":                         0x29e6,  // ⧦ '\u29e6'
	"glj":                                 0x2aa4,  // ⪤ '\u2aa4'
	"glottal":                             0x0249,  // ɉ '\u0249'
	"glottalinvertedstroke":               0x01be,  // ƾ '\u01be'
	"glottalrev":                          0x024a,  // Ɋ '\u024a'
	"glottalstop":                         0x0294,  // ʔ '\u0294'
	"glottalstopbar":                      0x0231,  // ȱ '\u0231'
	"glottalstopbarrev":                   0x0232,  // Ȳ '\u0232'
	"glottalstopinv":                      0x0226,  // Ȧ '\u0226'
	"glottalstopinverted":                 0x0296,  // ʖ '\u0296'
	"glottalstopmod":                      0x02c0,  // ˀ '\u02c0'
	"glottalstopreversed":                 0x0295,  // ʕ '\u0295'
	"glottalstopreversedmod":              0x02c1,  // ˁ '\u02c1'
	"glottalstopreversedsuperior":         0x02e4,  // ˤ '\u02e4'
	"glottalstoprevinv":                   0x0225,  // ȥ '\u0225'
	"glottalstopstroke":                   0x02a1,  // ʡ '\u02a1'
	"glottalstopstrokereversed":           0x02a2,  // ʢ '\u02a2'
	"gmacron":                             0x1e21,  // ḡ '\u1e21'
	"gmonospace":                          0xff47,  // ｇ '\uff47'
	"gnsim":                               0x22e7,  // ⋧ '\u22e7'
	"gohiragana":                          0x3054,  // ご '\u3054'
	"gokatakana":                          0x30b4,  // ゴ '\u30b4'
	"gparen":                              0x24a2,  // ⒢ '\u24a2'
	"gpasquare":                           0x33ac,  // ㎬ '\u33ac'
	"gradient":                            0x2207,  // ∇ '\u2207'
	"grave":                               0x0060,  // ` '`'
	"gravebelowcmb":                       0x0316,  // ̖ '\u0316'
	"gravecmb":                            0x0300,  // ̀ '\u0300'
	"gravedeva":                           0x0953,  // ॓ '\u0953'
	"graveleftnosp":                       0x02b3,  // ʳ '\u02b3'
	"gravelowmod":                         0x02ce,  // ˎ '\u02ce'
	"gravemonospace":                      0xff40,  // ｀ '\uff40'
	"gravetonecmb":                        0x0340,  // ̀ '\u0340'
	"greater":                             0x003e,  // > '>'
	"greaterdbleqlless":                   0x2a8c,  // ⪌ '\u2a8c'
	"greaterdot":                          0x22d7,  // ⋗ '\u22d7'
	"greaterequal":                        0x2265,  // ≥ '\u2265'
	"greaterequalorless":                  0x22db,  // ⋛ '\u22db'
	"greatermonospace":                    0xff1e,  // ＞ '\uff1e'
	"greaternotdblequal":                  0x2a8a,  // ⪊ '\u2a8a'
	"greaternotequal":                     0x2a88,  // ⪈ '\u2a88'
	"greaterorapproxeql":                  0x2a86,  // ⪆ '\u2a86'
	"greaterorequivalent":                 0x2273,  // ≳ '\u2273'
	"greaterorless":                       0x2277,  // ≷ '\u2277'
	"greaterornotdbleql":                  0x2269,  // ≩ '\u2269'
	"greateroverequal":                    0x2267,  // ≧ '\u2267'
	"greatersmall":                        0xfe65,  // ﹥ '\ufe65'
	"gscript":                             0x0261,  // ɡ '\u0261'
	"gsime":                               0x2a8e,  // ⪎ '\u2a8e'
	"gsiml":                               0x2a90,  // ⪐ '\u2a90'
	"gstroke":                             0x01e5,  // ǥ '\u01e5'
	"gtcc":                                0x2aa7,  // ⪧ '\u2aa7'
	"gtcir":                               0x2a7a,  // ⩺ '\u2a7a'
	"gtlpar":                              0x29a0,  // ⦠ '\u29a0'
	"gtquest":                             0x2a7c,  // ⩼ '\u2a7c'
	"gtrarr":                              0x2978,  // ⥸ '\u2978'
	"guhiragana":                          0x3050,  // ぐ '\u3050'
	"guillemotleft":                       0x00ab,  // « '\u00ab'
	"guillemotright":                      0x00bb,  // » '\u00bb'
	"guilsinglleft":                       0x2039,  // ‹ '\u2039'
	"guilsinglright":                      0x203a,  // › '\u203a'
	"gukatakana":                          0x30b0,  // グ '\u30b0'
	"guramusquare":                        0x3318,  // ㌘ '\u3318'
	"gysquare":                            0x33c9,  // ㏉ '\u33c9'
	"h":                                   0x0068,  // h 'h'
	"haabkhasiancyrillic":                 0x04a9,  // ҩ '\u04a9'
	"haaltonearabic":                      0x06c1,  // ہ '\u06c1'
	"habengali":                           0x09b9,  // হ '\u09b9'
	"haceksubnosp":                        0x029f,  // ʟ '\u029f'
	"hadescendercyrillic":                 0x04b3,  // ҳ '\u04b3'
	"hadeva":                              0x0939,  // ह '\u0939'
	"hagujarati":                          0x0ab9,  // હ '\u0ab9'
	"hagurmukhi":                          0x0a39,  // ਹ '\u0a39'
	"hahfinalarabic":                      0xfea2,  // ﺢ '\ufea2'
	"hahinitialarabic":                    0xfea3,  // ﺣ '\ufea3'
	"hahiragana":                          0x306f,  // は '\u306f'
	"hahisolated":                         0xfea1,  // ﺡ '\ufea1'
	"hahmedialarabic":                     0xfea4,  // ﺤ '\ufea4'
	"hahwithmeeminitial":                  0xfcaa,  // ﲪ '\ufcaa'
	"hairspace":                           0x200a,  //  '\u200a'
	"haitusquare":                         0x332a,  // ㌪ '\u332a'
	"hakatakana":                          0x30cf,  // ハ '\u30cf'
	"hakatakanahalfwidth":                 0xff8a,  // ﾊ '\uff8a'
	"halantgurmukhi":                      0x0a4d,  // ੍ '\u0a4d'
	"hamzadammaarabic":                    0x0621,  // ء '\u0621'
	"hamzaisolated":                       0xfe80,  // ﺀ '\ufe80'
	"hangulfiller":                        0x3164,  // ㅤ '\u3164'
	"hardsigncyrillic":                    0x044a,  // ъ '\u044a'
	"harpoondownleft":                     0x21c3,  // ⇃ '\u21c3'
	"harpoondownright":                    0x21c2,  // ⇂ '\u21c2'
	"harpoonleftbarbup":                   0x21bc,  // ↼ '\u21bc'
	"harpoonleftright":                    0x21cc,  // ⇌ '\u21cc'
	"harpoonrightbarbup":                  0x21c0,  // ⇀ '\u21c0'
	"harpoonrightleft":                    0x21cb,  // ⇋ '\u21cb'
	"harpoonupleft":                       0x21bf,  // ↿ '\u21bf'
	"harpoonupright":                      0x21be,  // ↾ '\u21be'
	"harrowextender":                      0x23af,  // ⎯ '\u23af'
	"hasquare":                            0x33ca,  // ㏊ '\u33ca'
	"hatafpatah16":                        0x05b2,  // ֲ '\u05b2'
	"hatafqamats28":                       0x05b3,  // ֳ '\u05b3'
	"hatafsegolwidehebrew":                0x05b1,  // ֱ '\u05b1'
	"hatapprox":                           0x2a6f,  // ⩯ '\u2a6f'
	"hbar":                                0x0127,  // ħ '\u0127'
	"hbopomofo":                           0x310f,  // ㄏ '\u310f'
	"hbrevebelow":                         0x1e2b,  // ḫ '\u1e2b'
	"hcedilla":                            0x1e29,  // ḩ '\u1e29'
	"hcircle":                             0x24d7,  // ⓗ '\u24d7'
	"hcircumflex":                         0x0125,  // ĥ '\u0125'
	"hcyril":                              0x03f7,  // Ϸ '\u03f7'
	"hdieresis":                           0x1e27,  // ḧ '\u1e27'
	"hdotaccent":                          0x1e23,  // ḣ '\u1e23'
	"hdotbelow":                           0x1e25,  // ḥ '\u1e25'
	"heart":                               0x2665,  // ♥ '\u2665'
	"heartsuitwhite":                      0x2661,  // ♡ '\u2661'
	"hedageshhebrew":                      0xfb34,  // הּ '\ufb34'
	"heharabic":                           0x0647,  // ه '\u0647'
	"hehfinalaltonearabic":                0xfba7,  // ﮧ '\ufba7'
	"hehfinalarabic":                      0xfeea,  // ﻪ '\ufeea'
	"hehhamzaabovefinalarabic":            0xfba5,  // ﮥ '\ufba5'
	"hehhamzaaboveisolatedarabic":         0xfba4,  // ﮤ '\ufba4'
	"hehinitialaltonearabic":              0xfba8,  // ﮨ '\ufba8'
	"hehinitialarabic":                    0xfeeb,  // ﻫ '\ufeeb'
	"hehiragana":                          0x3078,  // へ '\u3078'
	"hehisolated":                         0xfee9,  // ﻩ '\ufee9'
	"hehmedialaltonearabic":               0xfba9,  // ﮩ '\ufba9'
	"hehmedialarabic":                     0xfeec,  // ﻬ '\ufeec'
	"hehwithmeeminitial":                  0xfcd8,  // ﳘ '\ufcd8'
	"heiseierasquare":                     0x337b,  // ㍻ '\u337b'
	"hekatakana":                          0x30d8,  // ヘ '\u30d8'
	"hekatakanahalfwidth":                 0xff8d,  // ﾍ '\uff8d'
	"hekutaarusquare":                     0x3336,  // ㌶ '\u3336'
	"henghook":                            0x0267,  // ɧ '\u0267'
	"hermitmatrix":                        0x22b9,  // ⊹ '\u22b9'
	"herutusquare":                        0x3339,  // ㌹ '\u3339'
	"hexagon":                             0x2394,  // ⎔ '\u2394'
	"hexagonblack":                        0x2b23,  // ⬣ '\u2b23'
	"hhook":                               0x0266,  // ɦ '\u0266'
	"hhooksuper":                          0x023a,  // Ⱥ '\u023a'
	"hhooksuperior":                       0x02b1,  // ʱ '\u02b1'
	"hieuhacirclekorean":                  0x327b,  // ㉻ '\u327b'
	"hieuhaparenkorean":                   0x321b,  // ㈛ '\u321b'
	"hieuhcirclekorean":                   0x326d,  // ㉭ '\u326d'
	"hieuhkorean":                         0x314e,  // ㅎ '\u314e'
	"hieuhparenkorean":                    0x320d,  // ㈍ '\u320d'
	"highhamza":                           0x0674,  // ٴ '\u0674'
	"hihiragana":                          0x3072,  // ひ '\u3072'
	"hikatakana":                          0x30d2,  // ヒ '\u30d2'
	"hikatakanahalfwidth":                 0xff8b,  // ﾋ '\uff8b'
	"hiriq14":                             0x05b4,  // ִ '\u05b4'
	"hknearrow":                           0x2924,  // ⤤ '\u2924'
	"hknwarrow":                           0x2923,  // ⤣ '\u2923'
	"hksearow":                            0x2925,  // ⤥ '\u2925'
	"hkswarow":                            0x2926,  // ⤦ '\u2926'
	"hlinebelow":                          0x1e96,  // ẖ '\u1e96'
	"hmonospace":                          0xff48,  // ｈ '\uff48'
	"hoarmenian":                          0x0570,  // հ '\u0570'
	"hohipthai":                           0x0e2b,  // ห '\u0e2b'
	"hohiragana":                          0x307b,  // ほ '\u307b'
	"hokatakana":                          0x30db,  // ホ '\u30db'
	"hokatakanahalfwidth":                 0xff8e,  // ﾎ '\uff8e'
	"holamquarterhebrew":                  0x05b9,  // ֹ '\u05b9'
	"honokhukthai":                        0x0e2e,  // ฮ '\u0e2e'
	"hookcmb":                             0x0309,  // ̉ '\u0309'
	"hookpalatalizedbelowcmb":             0x0321,  // ̡ '\u0321'
	"hookretroflexbelowcmb":               0x0322,  // ̢ '\u0322'
	"hoonsquare":                          0x3342,  // ㍂ '\u3342'
	"horicoptic":                          0x03e9,  // ϩ '\u03e9'
	"horizontalbar":                       0x2015,  // ― '\u2015'
	"horiztab":                            0x05a2,  // ֢ '\u05a2'
	"horncmb":                             0x031b,  // ̛ '\u031b'
	"hotsprings":                          0x2668,  // ♨ '\u2668'
	"hourglass":                           0x29d6,  // ⧖ '\u29d6'
	"house":                               0x2302,  // ⌂ '\u2302'
	"hparen":                              0x24a3,  // ⒣ '\u24a3'
	"hrectangle":                          0x25ad,  // ▭ '\u25ad'
	"hsuper":                              0x0239,  // ȹ '\u0239'
	"hsuperior":                           0x02b0,  // ʰ '\u02b0'
	"hturned":                             0x0265,  // ɥ '\u0265'
	"huhiragana":                          0x3075,  // ふ '\u3075'
	"huiitosquare":                        0x3333,  // ㌳ '\u3333'
	"hukatakana":                          0x30d5,  // フ '\u30d5'
	"hukatakanahalfwidth":                 0xff8c,  // ﾌ '\uff8c'
	"hungarumlaut":                        0x02dd,  // ˝ '\u02dd'
	"hungarumlaut1":                       0xf009,  //  '\uf009'
	"hungarumlautcmb":                     0x030b,  // ̋ '\u030b'
	"hv":                                  0x0195,  // ƕ '\u0195'
	"hyphen":                              0x002d,  // - '-'
	"hyphenbullet":                        0x2043,  // ⁃ '\u2043'
	"hyphendot":                           0x2027,  // ‧ '\u2027'
	"hypheninferior":                      0xf6e5,  //  '\uf6e5'
	"hyphenmonospace":                     0xff0d,  // － '\uff0d'
	"hyphensmall":                         0xfe63,  // ﹣ '\ufe63'
	"hyphensuperior":                      0xf6e6,  //  '\uf6e6'
	"hyphentwo":                           0x2010,  // ‐ '\u2010'
	"hzigzag":                             0x3030,  // 〰 '\u3030'
	"i":                                   0x0069,  // i 'i'
	"iacute":                              0x00ed,  // í '\u00ed'
	"ibar":                                0x01f8,  // Ǹ '\u01f8'
	"ibengali":                            0x0987,  // ই '\u0987'
	"ibopomofo":                           0x3127,  // ㄧ '\u3127'
	"ibreve":                              0x012d,  // ĭ '\u012d'
	"icaron":                              0x01d0,  // ǐ '\u01d0'
	"icircle":                             0x24d8,  // ⓘ '\u24d8'
	"icircumflex":                         0x00ee,  // î '\u00ee'
	"idblgrave":                           0x0209,  // ȉ '\u0209'
	"ideographearthcircle":                0x328f,  // ㊏ '\u328f'
	"ideographfirecircle":                 0x328b,  // ㊋ '\u328b'
	"ideographicallianceparen":            0x323f,  // ㈿ '\u323f'
	"ideographiccallparen":                0x323a,  // ㈺ '\u323a'
	"ideographiccentrecircle":             0x32a5,  // ㊥ '\u32a5'
	"ideographicclose":                    0x3006,  // 〆 '\u3006'
	"ideographiccomma":                    0x3001,  // 、 '\u3001'
	"ideographiccommaleft":                0xff64,  // ､ '\uff64'
	"ideographiccongratulationparen":      0x3237,  // ㈷ '\u3237'
	"ideographiccorrectcircle":            0x32a3,  // ㊣ '\u32a3'
	"ideographicearthparen":               0x322f,  // ㈯ '\u322f'
	"ideographicenterpriseparen":          0x323d,  // ㈽ '\u323d'
	"ideographicexcellentcircle":          0x329d,  // ㊝ '\u329d'
	"ideographicfestivalparen":            0x3240,  // ㉀ '\u3240'
	"ideographicfinancialcircle":          0x3296,  // ㊖ '\u3296'
	"ideographicfinancialparen":           0x3236,  // ㈶ '\u3236'
	"ideographicfireparen":                0x322b,  // ㈫ '\u322b'
	"ideographichaveparen":                0x3232,  // ㈲ '\u3232'
	"ideographichighcircle":               0x32a4,  // ㊤ '\u32a4'
	"ideographiciterationmark":            0x3005,  // 々 '\u3005'
	"ideographiclaborcircle":              0x3298,  // ㊘ '\u3298'
	"ideographiclaborparen":               0x3238,  // ㈸ '\u3238'
	"ideographicleftcircle":               0x32a7,  // ㊧ '\u32a7'
	"ideographiclowcircle":                0x32a6,  // ㊦ '\u32a6'
	"ideographicmedicinecircle":           0x32a9,  // ㊩ '\u32a9'
	"ideographicmetalparen":               0x322e,  // ㈮ '\u322e'
	"ideographicmoonparen":                0x322a,  // ㈪ '\u322a'
	"ideographicnameparen":                0x3234,  // ㈴ '\u3234'
	"ideographicperiod":                   0x3002,  // 。 '\u3002'
	"ideographicprintcircle":              0x329e,  // ㊞ '\u329e'
	"ideographicreachparen":               0x3243,  // ㉃ '\u3243'
	"ideographicrepresentparen":           0x3239,  // ㈹ '\u3239'
	"ideographicresourceparen":            0x323e,  // ㈾ '\u323e'
	"ideographicrightcircle":              0x32a8,  // ㊨ '\u32a8'
	"ideographicsecretcircle":             0x3299,  // ㊙ '\u3299'
	"ideographicselfparen":                0x3242,  // ㉂ '\u3242'
	"ideographicsocietyparen":             0x3233,  // ㈳ '\u3233'
	"ideographicspace":                    0x3000,  //  '\u3000'
	"ideographicspecialparen":             0x3235,  // ㈵ '\u3235'
	"ideographicstockparen":               0x3231,  // ㈱ '\u3231'
	"ideographicstudyparen":               0x323b,  // ㈻ '\u323b'
	"ideographicsunparen":                 0x3230,  // ㈰ '\u3230'
	"ideographicsuperviseparen":           0x323c,  // ㈼ '\u323c'
	"ideographicwaterparen":               0x322c,  // ㈬ '\u322c'
	"ideographicwoodparen":                0x322d,  // ㈭ '\u322d'
	"ideographiczero":                     0x3007,  // 〇 '\u3007'
	"ideographmetalcircle":                0x328e,  // ㊎ '\u328e'
	"ideographmooncircle":                 0x328a,  // ㊊ '\u328a'
	"ideographnamecircle":                 0x3294,  // ㊔ '\u3294'
	"ideographsuncircle":                  0x3290,  // ㊐ '\u3290'
	"ideographwatercircle":                0x328c,  // ㊌ '\u328c'
	"ideographwoodcircle":                 0x328d,  // ㊍ '\u328d'
	"ideva":                               0x0907,  // इ '\u0907'
	"idieresis":                           0x00ef,  // ï '\u00ef'
	"idieresisacute":                      0x1e2f,  // ḯ '\u1e2f'
	"idieresiscyrillic":                   0x04e5,  // ӥ '\u04e5'
	"idotbelow":                           0x1ecb,  // ị '\u1ecb'
	"iebrevecyrillic":                     0x04d7,  // ӗ '\u04d7'
	"iecyrillic":                          0x0435,  // е '\u0435'
	"iehook":                              0x03f9,  // Ϲ '\u03f9'
	"iehookogonek":                        0x03fb,  // ϻ '\u03fb'
	"ieungacirclekorean":                  0x3275,  // ㉵ '\u3275'
	"ieungaparenkorean":                   0x3215,  // ㈕ '\u3215'
	"ieungcirclekorean":                   0x3267,  // ㉧ '\u3267'
	"ieungkorean":                         0x3147,  // ㅇ '\u3147'
	"ieungparenkorean":                    0x3207,  // ㈇ '\u3207'
	"igrave":                              0x00ec,  // ì '\u00ec'
	"igujarati":                           0x0a87,  // ઇ '\u0a87'
	"igurmukhi":                           0x0a07,  // ਇ '\u0a07'
	"ihiragana":                           0x3044,  // い '\u3044'
	"ihookabove":                          0x1ec9,  // ỉ '\u1ec9'
	"iibengali":                           0x0988,  // ঈ '\u0988'
	"iicyrillic":                          0x0438,  // и '\u0438'
	"iideva":                              0x0908,  // ई '\u0908'
	"iigujarati":                          0x0a88,  // ઈ '\u0a88'
	"iigurmukhi":                          0x0a08,  // ਈ '\u0a08'
	"iiiint":                              0x2a0c,  // ⨌ '\u2a0c'
	"iiint":                               0x222d,  // ∭ '\u222d'
	"iimatragurmukhi":                     0x0a40,  // ੀ '\u0a40'
	"iinfin":                              0x29dc,  // ⧜ '\u29dc'
	"iinvertedbreve":                      0x020b,  // ȋ '\u020b'
	"iivowelsignbengali":                  0x09c0,  // ী '\u09c0'
	"iivowelsigndeva":                     0x0940,  // ी '\u0940'
	"iivowelsigngujarati":                 0x0ac0,  // ી '\u0ac0'
	"ij":                                  0x0133,  // ĳ '\u0133'
	"ikatakana":                           0x30a4,  // イ '\u30a4'
	"ikatakanahalfwidth":                  0xff72,  // ｲ '\uff72'
	"ikorean":                             0x3163,  // ㅣ '\u3163'
	"iluyhebrew":                          0x05ac,  // ֬ '\u05ac'
	"imacron":                             0x012b,  // ī '\u012b'
	"imacroncyrillic":                     0x04e3,  // ӣ '\u04e3'
	"imageof":                             0x22b7,  // ⊷ '\u22b7'
	"imageorapproximatelyequal":           0x2253,  // ≓ '\u2253'
	"imath":                               0x1d6a4, // 𝚤 '\U0001d6a4'
	"imatragurmukhi":                      0x0a3f,  // ਿ '\u0a3f'
	"imonospace":                          0xff49,  // ｉ '\uff49'
	"infinity":                            0x221e,  // ∞ '\u221e'
	"iniarmenian":                         0x056b,  // ի '\u056b'
	"intBar":                              0x2a0e,  // ⨎ '\u2a0e'
	"intbar":                              0x2a0d,  // ⨍ '\u2a0d'
	"intcap":                              0x2a19,  // ⨙ '\u2a19'
	"intclockwise":                        0x2231,  // ∱ '\u2231'
	"intcup":                              0x2a1a,  // ⨚ '\u2a1a'
	"integerdivide":                       0x2216,  // ∖ '\u2216'
	"integral":                            0x222b,  // ∫ '\u222b'
	"integralbt":                          0x2321,  // ⌡ '\u2321'
	"integralex":                          0xf8f5,  //  '\uf8f5'
	"integraltp":                          0x2320,  // ⌠ '\u2320'
	"intercal":                            0x22ba,  // ⊺ '\u22ba'
	"interleave":                          0x2af4,  // ⫴ '\u2af4'
	"interrobang":                         0x203d,  // ‽ '\u203d'
	"interrobangdown":                     0x2e18,  // ⸘ '\u2e18'
	"intersection":                        0x2229,  // ∩ '\u2229'
	"intersectiondbl":                     0x22d2,  // ⋒ '\u22d2'
	"intersectiondisplay":                 0x22c2,  // ⋂ '\u22c2'
	"intersectionsq":                      0x2293,  // ⊓ '\u2293'
	"intextender":                         0x23ae,  // ⎮ '\u23ae'
	"intisquare":                          0x3305,  // ㌅ '\u3305'
	"intlarhk":                            0x2a17,  // ⨗ '\u2a17'
	"intprod":                             0x2a3c,  // ⨼ '\u2a3c'
	"intprodr":                            0x2a3d,  // ⨽ '\u2a3d'
	"intx":                                0x2a18,  // ⨘ '\u2a18'
	"invbullet":                           0x25d8,  // ◘ '\u25d8'
	"invcircle":                           0x25d9,  // ◙ '\u25d9'
	"invlazys":                            0x223e,  // ∾ '\u223e'
	"invwhitelowerhalfcircle":             0x25db,  // ◛ '\u25db'
	"invwhiteupperhalfcircle":             0x25da,  // ◚ '\u25da'
	"iogonek":                             0x012f,  // į '\u012f'
	"iota":                                0x03b9,  // ι '\u03b9'
	"iota1":                               0x01f9,  // ǹ '\u01f9'
	"iotadieresis":                        0x03ca,  // ϊ '\u03ca'
	"iotadieresistonos":                   0x0390,  // ΐ '\u0390'
	"iotalatin":                           0x0269,  // ɩ '\u0269'
	"iotatonos":                           0x03af,  // ί '\u03af'
	"iparen":                              0x24a4,  // ⒤ '\u24a4'
	"irigurmukhi":                         0x0a72,  // ੲ '\u0a72'
	"isinE":                               0x22f9,  // ⋹ '\u22f9'
	"isindot":                             0x22f5,  // ⋵ '\u22f5'
	"isinobar":                            0x22f7,  // ⋷ '\u22f7'
	"isins":                               0x22f4,  // ⋴ '\u22f4'
	"isinvb":                              0x22f8,  // ⋸ '\u22f8'
	"ismallhiragana":                      0x3043,  // ぃ '\u3043'
	"ismallkatakana":                      0x30a3,  // ィ '\u30a3'
	"ismallkatakanahalfwidth":             0xff68,  // ｨ '\uff68'
	"issharbengali":                       0x09fa,  // ৺ '\u09fa'
	"istroke":                             0x0268,  // ɨ '\u0268'
	"isuperior":                           0xf6ed,  //  '\uf6ed'
	"iterationhiragana":                   0x309d,  // ゝ '\u309d'
	"iterationkatakana":                   0x30fd,  // ヽ '\u30fd'
	"itilde":                              0x0129,  // ĩ '\u0129'
	"itildebelow":                         0x1e2d,  // ḭ '\u1e2d'
	"iubopomofo":                          0x3129,  // ㄩ '\u3129'
	"ivowelsignbengali":                   0x09bf,  // ি '\u09bf'
	"ivowelsigndeva":                      0x093f,  // ि '\u093f'
	"ivowelsigngujarati":                  0x0abf,  // િ '\u0abf'
	"izhitsadblgravecyrillic":             0x0477,  // ѷ '\u0477'
	"j":                                   0x006a,  // j 'j'
	"jaarmenian":                          0x0571,  // ձ '\u0571'
	"jabengali":                           0x099c,  // জ '\u099c'
	"jadeva":                              0x091c,  // ज '\u091c'
	"jagujarati":                          0x0a9c,  // જ '\u0a9c'
	"jagurmukhi":                          0x0a1c,  // ਜ '\u0a1c'
	"jbopomofo":                           0x3110,  // ㄐ '\u3110'
	"jcaron":                              0x01f0,  // ǰ '\u01f0'
	"jcircle":                             0x24d9,  // ⓙ '\u24d9'
	"jcircumflex":                         0x0135,  // ĵ '\u0135'
	"jcrossedtail":                        0x029d,  // ʝ '\u029d'
	"jcrosstail":                          0x022d,  // ȭ '\u022d'
	"jdotlessstroke":                      0x025f,  // ɟ '\u025f'
	"jeemarabic":                          0x062c,  // ج '\u062c'
	"jeemfinalarabic":                     0xfe9e,  // ﺞ '\ufe9e'
	"jeeminitialarabic":                   0xfe9f,  // ﺟ '\ufe9f'
	"jeemisolated":                        0xfe9d,  // ﺝ '\ufe9d'
	"jeemmedialarabic":                    0xfea0,  // ﺠ '\ufea0'
	"jeemwithmeeminitial":                 0xfca8,  // ﲨ '\ufca8'
	"jehfinalarabic":                      0xfb8b,  // ﮋ '\ufb8b'
	"jehisolated":                         0xfb8a,  // ﮊ '\ufb8a'
	"jhabengali":                          0x099d,  // ঝ '\u099d'
	"jhadeva":                             0x091d,  // झ '\u091d'
	"jhagujarati":                         0x0a9d,  // ઝ '\u0a9d'
	"jhagurmukhi":                         0x0a1d,  // ਝ '\u0a1d'
	"jheharmenian":                        0x057b,  // ջ '\u057b'
	"jis":                                 0x3004,  // 〄 '\u3004'
	"jmath":                               0x1d6a5, // 𝚥 '\U0001d6a5'
	"jmonospace":                          0xff4a,  // ｊ '\uff4a'
	"jparen":                              0x24a5,  // ⒥ '\u24a5'
	"jsuper":                              0x023b,  // Ȼ '\u023b'
	"jsuperior":                           0x02b2,  // ʲ '\u02b2'
	"k":                                   0x006b,  // k 'k'
	"kabashkircyrillic":                   0x04a1,  // ҡ '\u04a1'
	"kabengali":                           0x0995,  // ক '\u0995'
	"kacute":                              0x1e31,  // ḱ '\u1e31'
	"kacyrillic":                          0x043a,  // к '\u043a'
	"kadescendercyrillic":                 0x049b,  // қ '\u049b'
	"kadeva":                              0x0915,  // क '\u0915'
	"kaf":                                 0x05db,  // כ '\u05db'
	"kafdagesh":                           0xfb3b,  // כּ '\ufb3b'
	"kaffinalarabic":                      0xfeda,  // ﻚ '\ufeda'
	"kafinitialarabic":                    0xfedb,  // ﻛ '\ufedb'
	"kafisolated":                         0xfed9,  // ﻙ '\ufed9'
	"kafmedialarabic":                     0xfedc,  // ﻜ '\ufedc'
	"kafrafehebrew":                       0xfb4d,  // כֿ '\ufb4d'
	"kagujarati":                          0x0a95,  // ક '\u0a95'
	"kagurmukhi":                          0x0a15,  // ਕ '\u0a15'
	"kahiragana":                          0x304b,  // か '\u304b'
	"kahook":                              0x0400,  // Ѐ '\u0400'
	"kahookcyrillic":                      0x04c4,  // ӄ '\u04c4'
	"kakatakana":                          0x30ab,  // カ '\u30ab'
	"kakatakanahalfwidth":                 0xff76,  // ｶ '\uff76'
	"kappa":                               0x03ba,  // κ '\u03ba'
	"kappasymbolgreek":                    0x03f0,  // ϰ '\u03f0'
	"kapyeounmieumkorean":                 0x3171,  // ㅱ '\u3171'
	"kapyeounphieuphkorean":               0x3184,  // ㆄ '\u3184'
	"kapyeounpieupkorean":                 0x3178,  // ㅸ '\u3178'
	"kapyeounssangpieupkorean":            0x3179,  // ㅹ '\u3179'
	"karoriisquare":                       0x330d,  // ㌍ '\u330d'
	"kartdes":                             0x03d7,  // ϗ '\u03d7'
	"kashidaautonosidebearingarabic":      0x0640,  // ـ '\u0640'
	"kasmallkatakana":                     0x30f5,  // ヵ '\u30f5'
	"kasquare":                            0x3384,  // ㎄ '\u3384'
	"kasraisolated":                       0xfe7a,  // ﹺ '\ufe7a'
	"kasralow":                            0xe826,  //  '\ue826'
	"kasramedial":                         0xfe7b,  // ﹻ '\ufe7b'
	"kasratanarabic":                      0x064d,  // ٍ '\u064d'
	"kasratanisolated":                    0xfe74,  // ﹴ '\ufe74'
	"kasratanlow":                         0xe827,  //  '\ue827'
	"kastrokecyrillic":                    0x049f,  // ҟ '\u049f'
	"katahiraprolongmarkhalfwidth":        0xff70,  // ｰ '\uff70'
	"kaverticalstrokecyrillic":            0x049d,  // ҝ '\u049d'
	"kbopomofo":                           0x310e,  // ㄎ '\u310e'
	"kcalsquare":                          0x3389,  // ㎉ '\u3389'
	"kcaron":                              0x01e9,  // ǩ '\u01e9'
	"kcircle":                             0x24da,  // ⓚ '\u24da'
	"kcommaaccent":                        0x0137,  // ķ '\u0137'
	"kdotbelow":                           0x1e33,  // ḳ '\u1e33'
	"keharmenian":                         0x0584,  // ք '\u0584'
	"keheh":                               0x06a9,  // ک '\u06a9'
	"kehehfinal":                          0xfb8f,  // ﮏ '\ufb8f'
	"kehehinitial":                        0xfb90,  // ﮐ '\ufb90'
	"kehehisolated":                       0xfb8e,  // ﮎ '\ufb8e'
	"kehehmedial":                         0xfb91,  // ﮑ '\ufb91'
	"kehiragana":                          0x3051,  // け '\u3051'
	"kekatakana":                          0x30b1,  // ケ '\u30b1'
	"kekatakanahalfwidth":                 0xff79,  // ｹ '\uff79'
	"kenarmenian":                         0x056f,  // կ '\u056f'
	"kernelcontraction":                   0x223b,  // ∻ '\u223b'
	"kesmallkatakana":                     0x30f6,  // ヶ '\u30f6'
	"kgreenlandic":                        0x0138,  // ĸ '\u0138'
	"khabengali":                          0x0996,  // খ '\u0996'
	"khadeva":                             0x0916,  // ख '\u0916'
	"khagujarati":                         0x0a96,  // ખ '\u0a96'
	"khagurmukhi":                         0x0a16,  // ਖ '\u0a16'
	"khahfinalarabic":                     0xfea6,  // ﺦ '\ufea6'
	"khahinitialarabic":                   0xfea7,  // ﺧ '\ufea7'
	"khahisolated":                        0xfea5,  // ﺥ '\ufea5'
	"khahmedialarabic":                    0xfea8,  // ﺨ '\ufea8'
	"khahwithmeeminitial":                 0xfcac,  // ﲬ '\ufcac'
	"kheicoptic":                          0x03e7,  // ϧ '\u03e7'
	"khhadeva":                            0x0959,  // ख़ '\u0959'
	"khhagurmukhi":                        0x0a59,  // ਖ਼ '\u0a59'
	"khieukhacirclekorean":                0x3278,  // ㉸ '\u3278'
	"khieukhaparenkorean":                 0x3218,  // ㈘ '\u3218'
	"khieukhcirclekorean":                 0x326a,  // ㉪ '\u326a'
	"khieukhkorean":                       0x314b,  // ㅋ '\u314b'
	"khieukhparenkorean":                  0x320a,  // ㈊ '\u320a'
	"khokhaithai":                         0x0e02,  // ข '\u0e02'
	"khokhonthai":                         0x0e05,  // ฅ '\u0e05'
	"khokhuatthai":                        0x0e03,  // ฃ '\u0e03'
	"khokhwaithai":                        0x0e04,  // ค '\u0e04'
	"khomutthai":                          0x0e5b,  // ๛ '\u0e5b'
	"khook":                               0x0199,  // ƙ '\u0199'
	"khorakhangthai":                      0x0e06,  // ฆ '\u0e06'
	"khzsquare":                           0x3391,  // ㎑ '\u3391'
	"kihiragana":                          0x304d,  // き '\u304d'
	"kikatakana":                          0x30ad,  // キ '\u30ad'
	"kikatakanahalfwidth":                 0xff77,  // ｷ '\uff77'
	"kiroguramusquare":                    0x3315,  // ㌕ '\u3315'
	"kiromeetorusquare":                   0x3316,  // ㌖ '\u3316'
	"kirosquare":                          0x3314,  // ㌔ '\u3314'
	"kiyeokacirclekorean":                 0x326e,  // ㉮ '\u326e'
	"kiyeokaparenkorean":                  0x320e,  // ㈎ '\u320e'
	"kiyeokcirclekorean":                  0x3260,  // ㉠ '\u3260'
	"kiyeokkorean":                        0x3131,  // ㄱ '\u3131'
	"kiyeokparenkorean":                   0x3200,  // ㈀ '\u3200'
	"kiyeoksioskorean":                    0x3133,  // ㄳ '\u3133'
	"klinebelow":                          0x1e35,  // ḵ '\u1e35'
	"klsquare":                            0x3398,  // ㎘ '\u3398'
	"kmcubedsquare":                       0x33a6,  // ㎦ '\u33a6'
	"kmonospace":                          0xff4b,  // ｋ '\uff4b'
	"kmsquaredsquare":                     0x33a2,  // ㎢ '\u33a2'
	"kohiragana":                          0x3053,  // こ '\u3053'
	"kohmsquare":                          0x33c0,  // ㏀ '\u33c0'
	"kokaithai":                           0x0e01,  // ก '\u0e01'
	"kokatakana":                          0x30b3,  // コ '\u30b3'
	"kokatakanahalfwidth":                 0xff7a,  // ｺ '\uff7a'
	"kooposquare":                         0x331e,  // ㌞ '\u331e'
	"koppacyrillic":                       0x0481,  // ҁ '\u0481'
	"koreanstandardsymbol":                0x327f,  // ㉿ '\u327f'
	"koroniscmb":                          0x0343,  // ̓ '\u0343'
	"kparen":                              0x24a6,  // ⒦ '\u24a6'
	"kpasquare":                           0x33aa,  // ㎪ '\u33aa'
	"ksicyrillic":                         0x046f,  // ѯ '\u046f'
	"ktsquare":                            0x33cf,  // ㏏ '\u33cf'
	"kturn":                               0x022e,  // Ȯ '\u022e'
	"kturned":                             0x029e,  // ʞ '\u029e'
	"kuhiragana":                          0x304f,  // く '\u304f'
	"kukatakana":                          0x30af,  // ク '\u30af'
	"kukatakanahalfwidth":                 0xff78,  // ｸ '\uff78'
	"kvsquare":                            0x33b8,  // ㎸ '\u33b8'
	"kwsquare":                            0x33be,  // ㎾ '\u33be'
	"l":                                   0x006c,  // l 'l'
	"lAngle":                              0x27ea,  // ⟪ '\u27ea'
	"lBrace":                              0x2983,  // ⦃ '\u2983'
	"lParen":                              0x2985,  // ⦅ '\u2985'
	"labengali":                           0x09b2,  // ল '\u09b2'
	"lacute":                              0x013a,  // ĺ '\u013a'
	"ladeva":                              0x0932,  // ल '\u0932'
	"lagujarati":                          0x0ab2,  // લ '\u0ab2'
	"lagurmukhi":                          0x0a32,  // ਲ '\u0a32'
	"lakkhangyaothai":                     0x0e45,  // ๅ '\u0e45'
	"lamaleffinalarabic":                  0xfefc,  // ﻼ '\ufefc'
	"lamalefhamzaabovefinalarabic":        0xfef8,  // ﻸ '\ufef8'
	"lamalefhamzaaboveisolatedarabic":     0xfef7,  // ﻷ '\ufef7'
	"lamalefhamzabelowfinalarabic":        0xfefa,  // ﻺ '\ufefa'
	"lamalefhamzabelowisolatedarabic":     0xfef9,  // ﻹ '\ufef9'
	"lamalefisolatedarabic":               0xfefb,  // ﻻ '\ufefb'
	"lamalefmaddaabovefinalarabic":        0xfef6,  // ﻶ '\ufef6'
	"lamalefmaddaaboveisolatedarabic":     0xfef5,  // ﻵ '\ufef5'
	"lambda":                              0x03bb,  // λ '\u03bb'
	"lambdastroke":                        0x019b,  // ƛ '\u019b'
	"lameddagesh":                         0xfb3c,  // לּ '\ufb3c'
	"lamedholamhebrew":                    0x05dc,  // ל '\u05dc'
	"lamedwithdageshandholam":             0xe805,  //  '\ue805'
	"lamedwithholam":                      0xe804,  //  '\ue804'
	"lamfinalarabic":                      0xfede,  // ﻞ '\ufede'
	"lamhahinitialarabic":                 0xfcca,  // ﳊ '\ufcca'
	"laminitialarabic":                    0xfedf,  // ﻟ '\ufedf'
	"lamisolated":                         0xfedd,  // ﻝ '\ufedd'
	"lamjeeminitialarabic":                0xfcc9,  // ﳉ '\ufcc9'
	"lamkhahinitialarabic":                0xfccb,  // ﳋ '\ufccb'
	"lamlamhehisolatedarabic":             0xfdf2,  // ﷲ '\ufdf2'
	"lammedialarabic":                     0xfee0,  // ﻠ '\ufee0'
	"lammeemhahinitialarabic":             0xfd88,  // ﶈ '\ufd88'
	"lammeeminitialarabic":                0xfccc,  // ﳌ '\ufccc'
	"lamwithalefmaksuraisolated":          0xfc43,  // ﱃ '\ufc43'
	"lamwithhahisolated":                  0xfc40,  // ﱀ '\ufc40'
	"lamwithhehinitial":                   0xfccd,  // ﳍ '\ufccd'
	"lamwithjeemisolated":                 0xfc3f,  // ﰿ '\ufc3f'
	"lamwithkhahisolated":                 0xfc41,  // ﱁ '\ufc41'
	"lamwithmeemisolated":                 0xfc42,  // ﱂ '\ufc42'
	"lamwithmeemwithjeeminitial":          0xe811,  //  '\ue811'
	"lamwithyehisolated":                  0xfc44,  // ﱄ '\ufc44'
	"langledot":                           0x2991,  // ⦑ '\u2991'
	"laplac":                              0x29e0,  // ⧠ '\u29e0'
	"largecircle":                         0x25ef,  // ◯ '\u25ef'
	"lat":                                 0x2aab,  // ⪫ '\u2aab'
	"late":                                0x2aad,  // ⪭ '\u2aad'
	"lbag":                                0x27c5,  // ⟅ '\u27c5'
	"lbar":                                0x019a,  // ƚ '\u019a'
	"lbbar":                               0x2114,  // ℔ '\u2114'
	"lbelt":                               0x026c,  // ɬ '\u026c'
	"lblkbrbrak":                          0x2997,  // ⦗ '\u2997'
	"lbopomofo":                           0x310c,  // ㄌ '\u310c'
	"lbracelend":                          0x23a9,  // ⎩ '\u23a9'
	"lbracemid":                           0x23a8,  // ⎨ '\u23a8'
	"lbraceuend":                          0x23a7,  // ⎧ '\u23a7'
	"lbrackextender":                      0x23a2,  // ⎢ '\u23a2'
	"lbracklend":                          0x23a3,  // ⎣ '\u23a3'
	"lbracklltick":                        0x298f,  // ⦏ '\u298f'
	"lbrackubar":                          0x298b,  // ⦋ '\u298b'
	"lbrackuend":                          0x23a1,  // ⎡ '\u23a1'
	"lbrackultick":                        0x298d,  // ⦍ '\u298d'
	"lbrbrak":                             0x2772,  // ❲ '\u2772'
	"lcaron":                              0x013e,  // ľ '\u013e'
	"lcaron1":                             0xf813,  //  '\uf813'
	"lcircle":                             0x24db,  // ⓛ '\u24db'
	"lcircumflexbelow":                    0x1e3d,  // ḽ '\u1e3d'
	"lcommaaccent":                        0x013c,  // ļ '\u013c'
	"lcurvyangle":                         0x29fc,  // ⧼ '\u29fc'
	"ldotaccent":                          0x0140,  // ŀ '\u0140'
	"ldotbelow":                           0x1e37,  // ḷ '\u1e37'
	"ldotbelowmacron":                     0x1e39,  // ḹ '\u1e39'
	"leftangleabovecmb":                   0x031a,  // ̚ '\u031a'
	"leftarrowapprox":                     0x2b4a,  // ⭊ '\u2b4a'
	"leftarrowbackapprox":                 0x2b42,  // ⭂ '\u2b42'
	"leftarrowbsimilar":                   0x2b4b,  // ⭋ '\u2b4b'
	"leftarrowless":                       0x2977,  // ⥷ '\u2977'
	"leftarrowonoplus":                    0x2b32,  // ⬲ '\u2b32'
	"leftarrowplus":                       0x2946,  // ⥆ '\u2946'
	"leftarrowshortrightarrow":            0x2943,  // ⥃ '\u2943'
	"leftarrowsimilar":                    0x2973,  // ⥳ '\u2973'
	"leftarrowsubset":                     0x297a,  // ⥺ '\u297a'
	"leftarrowtriangle":                   0x21fd,  // ⇽ '\u21fd'
	"leftarrowx":                          0x2b3e,  // ⬾ '\u2b3e'
	"leftbkarrow":                         0x290c,  // ⤌ '\u290c'
	"leftcurvedarrow":                     0x2b3f,  // ⬿ '\u2b3f'
	"leftdbkarrow":                        0x290e,  // ⤎ '\u290e'
	"leftdbltail":                         0x291b,  // ⤛ '\u291b'
	"leftdotarrow":                        0x2b38,  // ⬸ '\u2b38'
	"leftdowncurvedarrow":                 0x2936,  // ⤶ '\u2936'
	"leftfishtail":                        0x297c,  // ⥼ '\u297c'
	"leftharpoonaccent":                   0x20d0,  // ⃐ '\u20d0'
	"leftharpoondownbar":                  0x295e,  // ⥞ '\u295e'
	"leftharpoonsupdown":                  0x2962,  // ⥢ '\u2962'
	"leftharpoonupbar":                    0x295a,  // ⥚ '\u295a'
	"leftharpoonupdash":                   0x296a,  // ⥪ '\u296a'
	"leftleftarrows":                      0x21c7,  // ⇇ '\u21c7'
	"leftmoon":                            0x263e,  // ☾ '\u263e'
	"leftouterjoin":                       0x27d5,  // ⟕ '\u27d5'
	"leftrightarrowcircle":                0x2948,  // ⥈ '\u2948'
	"leftrightarrowtriangle":              0x21ff,  // ⇿ '\u21ff'
	"leftrightharpoondowndown":            0x2950,  // ⥐ '\u2950'
	"leftrightharpoondownup":              0x294b,  // ⥋ '\u294b'
	"leftrightharpoonsdown":               0x2967,  // ⥧ '\u2967'
	"leftrightharpoonsup":                 0x2966,  // ⥦ '\u2966'
	"leftrightharpoonupdown":              0x294a,  // ⥊ '\u294a'
	"leftrightharpoonupup":                0x294e,  // ⥎ '\u294e'
	"leftsquigarrow":                      0x21dc,  // ⇜ '\u21dc'
	"lefttackbelowcmb":                    0x0318,  // ̘ '\u0318'
	"lefttail":                            0x2919,  // ⤙ '\u2919'
	"leftthreearrows":                     0x2b31,  // ⬱ '\u2b31'
	"leftwavearrow":                       0x219c,  // ↜ '\u219c'
	"leqqslant":                           0x2af9,  // ⫹ '\u2af9'
	"lescc":                               0x2aa8,  // ⪨ '\u2aa8'
	"lesdot":                              0x2a7f,  // ⩿ '\u2a7f'
	"lesdoto":                             0x2a81,  // ⪁ '\u2a81'
	"lesdotor":                            0x2a83,  // ⪃ '\u2a83'
	"lesges":                              0x2a93,  // ⪓ '\u2a93'
	"less":                                0x003c,  // < '<'
	"lessdbleqlgreater":                   0x2a8b,  // ⪋ '\u2a8b'
	"lessdot":                             0x22d6,  // ⋖ '\u22d6'
	"lessequal":                           0x2264,  // ≤ '\u2264'
	"lessequalorgreater":                  0x22da,  // ⋚ '\u22da'
	"lessmonospace":                       0xff1c,  // ＜ '\uff1c'
	"lessnotdblequal":                     0x2a89,  // ⪉ '\u2a89'
	"lessnotequal":                        0x2a87,  // ⪇ '\u2a87'
	"lessorapproxeql":                     0x2a85,  // ⪅ '\u2a85'
	"lessorequalslant":                    0x2a7d,  // ⩽ '\u2a7d'
	"lessorequivalent":                    0x2272,  // ≲ '\u2272'
	"lessorgreater":                       0x2276,  // ≶ '\u2276'
	"lessornotequal":                      0x2268,  // ≨ '\u2268'
	"lessoverequal":                       0x2266,  // ≦ '\u2266'
	"lesssmall":                           0xfe64,  // ﹤ '\ufe64'
	"lezh":                                0x026e,  // ɮ '\u026e'
	"lfblock":                             0x258c,  // ▌ '\u258c'
	"lfbowtie":                            0x29d1,  // ⧑ '\u29d1'
	"lfeighthblock":                       0x258f,  // ▏ '\u258f'
	"lffiveeighthblock":                   0x258b,  // ▋ '\u258b'
	"lfquarterblock":                      0x258e,  // ▎ '\u258e'
	"lfseveneighthblock":                  0x2589,  // ▉ '\u2589'
	"lfthreeeighthblock":                  0x258d,  // ▍ '\u258d'
	"lfthreequarterblock":                 0x258a,  // ▊ '\u258a'
	"lftimes":                             0x29d4,  // ⧔ '\u29d4'
	"lgE":                                 0x2a91,  // ⪑ '\u2a91'
	"lgblkcircle":                         0x2b24,  // ⬤ '\u2b24'
	"lgblksquare":                         0x2b1b,  // ⬛ '\u2b1b'
	"lgwhtsquare":                         0x2b1c,  // ⬜ '\u2b1c'
	"lhookretroflex":                      0x026d,  // ɭ '\u026d'
	"linefeed":                            0x21b4,  // ↴ '\u21b4'
	"lineseparator":                       0x2028,  //  '\u2028'
	"linevertnosp":                        0x0280,  // ʀ '\u0280'
	"linevertsubnosp":                     0x029c,  // ʜ '\u029c'
	"lira":                                0x20a4,  // ₤ '\u20a4'
	"liwnarmenian":                        0x056c,  // լ '\u056c'
	"lj":                                  0x01c9,  // ǉ '\u01c9'
	"ljecyrillic":                         0x0459,  // љ '\u0459'
	"ll":                                  0xf6c0,  //  '\uf6c0'
	"lladeva":                             0x0933,  // ळ '\u0933'
	"llagujarati":                         0x0ab3,  // ળ '\u0ab3'
	"llangle":                             0x2989,  // ⦉ '\u2989'
	"llarc":                               0x25df,  // ◟ '\u25df'
	"llinebelow":                          0x1e3b,  // ḻ '\u1e3b'
	"lll":                                 0x22d8,  // ⋘ '\u22d8'
	"llladeva":                            0x0934,  // ऴ '\u0934'
	"lllnest":                             0x2af7,  // ⫷ '\u2af7'
	"llparenthesis":                       0x2987,  // ⦇ '\u2987'
	"lltriangle":                          0x25fa,  // ◺ '\u25fa'
	"llvocalicbengali":                    0x09e1,  // ৡ '\u09e1'
	"llvocalicdeva":                       0x0961,  // ॡ '\u0961'
	"llvocalicvowelsignbengali":           0x09e3,  // ৣ '\u09e3'
	"llvocalicvowelsigndeva":              0x0963,  // ॣ '\u0963'
	"lmiddletilde":                        0x026b,  // ɫ '\u026b'
	"lmonospace":                          0xff4c,  // ｌ '\uff4c'
	"lmoustache":                          0x23b0,  // ⎰ '\u23b0'
	"lmsquare":                            0x33d0,  // ㏐ '\u33d0'
	"lnsim":                               0x22e6,  // ⋦ '\u22e6'
	"lochulathai":                         0x0e2c,  // ฬ '\u0e2c'
	"logicaland":                          0x2227,  // ∧ '\u2227'
	"logicalnot":                          0x00ac,  // ¬ '\u00ac'
	"logicalor":                           0x2228,  // ∨ '\u2228'
	"logonek":                             0xf830,  //  '\uf830'
	"lolingthai":                          0x0e25,  // ล '\u0e25'
	"longdashv":                           0x27de,  // ⟞ '\u27de'
	"longdivision":                        0x27cc,  // ⟌ '\u27cc'
	"longleftarrow":                       0x27f5,  // ⟵ '\u27f5'
	"longleftrightarrow":                  0x27f7,  // ⟷ '\u27f7'
	"longleftsquigarrow":                  0x2b33,  // ⬳ '\u2b33'
	"longmapsfrom":                        0x27fb,  // ⟻ '\u27fb'
	"longmapsto":                          0x27fc,  // ⟼ '\u27fc'
	"longrightarrow":                      0x27f6,  // ⟶ '\u27f6'
	"longrightsquigarrow":                 0x27ff,  // ⟿ '\u27ff'
	"longs":                               0x017f,  // ſ '\u017f'
	"longst":                              0xfb05,  // ﬅ '\ufb05'
	"lowered":                             0x024e,  // Ɏ '\u024e'
	"lowint":                              0x2a1c,  // ⨜ '\u2a1c'
	"lowlinecenterline":                   0xfe4e,  // ﹎ '\ufe4e'
	"lowlinecmb":                          0x0332,  // ̲ '\u0332'
	"lowlinedashed":                       0xfe4d,  // ﹍ '\ufe4d'
	"lozenge":                             0x25ca,  // ◊ '\u25ca'
	"lozengeminus":                        0x27e0,  // ⟠ '\u27e0'
	"lparen":                              0x24a7,  // ⒧ '\u24a7'
	"lparenextender":                      0x239c,  // ⎜ '\u239c'
	"lparenlend":                          0x239d,  // ⎝ '\u239d'
	"lparenless":                          0x2993,  // ⦓ '\u2993'
	"lparenuend":                          0x239b,  // ⎛ '\u239b'
	"lrarc":                               0x25de,  // ◞ '\u25de'
	"lre":                                 0x202a,  //  '\u202a'
	"lrtriangle":                          0x25ff,  // ◿ '\u25ff'
	"lrtriangleeq":                        0x29e1,  // ⧡ '\u29e1'
	"lsime":                               0x2a8d,  // ⪍ '\u2a8d'
	"lsimg":                               0x2a8f,  // ⪏ '\u2a8f'
	"lslash":                              0x0142,  // ł '\u0142'
	"lsqhook":                             0x2acd,  // ⫍ '\u2acd'
	"lsuper":                              0x026a,  // ɪ '\u026a'
	"lsuperior":                           0xf6ee,  //  '\uf6ee'
	"ltcc":                                0x2aa6,  // ⪦ '\u2aa6'
	"ltcir":                               0x2a79,  // ⩹ '\u2a79'
	"ltlarr":                              0x2976,  // ⥶ '\u2976'
	"ltquest":                             0x2a7b,  // ⩻ '\u2a7b'
	"ltrivb":                              0x29cf,  // ⧏ '\u29cf'
	"ltshade1":                            0xf821,  //  '\uf821'
	"luthai":                              0x0e26,  // ฦ '\u0e26'
	"lvboxline":                           0x23b8,  // ⎸ '\u23b8'
	"lvocalicbengali":                     0x098c,  // ঌ '\u098c'
	"lvocalicdeva":                        0x090c,  // ऌ '\u090c'
	"lvocalicvowelsignbengali":            0x09e2,  // ৢ '\u09e2'
	"lvocalicvowelsigndeva":               0x0962,  // ॢ '\u0962'
	"lvzigzag":                            0x29d8,  // ⧘ '\u29d8'
	"lxsquare":                            0x33d3,  // ㏓ '\u33d3'
	"m":                                   0x006d,  // m 'm'
	"mabengali":                           0x09ae,  // ম '\u09ae'
	"macron":                              0x00af,  // ¯ '\u00af'
	"macronbelowcmb":                      0x0331,  // ̱ '\u0331'
	"macroncmb":                           0x0304,  // ̄ '\u0304'
	"macronlowmod":                        0x02cd,  // ˍ '\u02cd'
	"macronmonospace":                     0xffe3,  // ￣ '\uffe3'
	"macute":                              0x1e3f,  // ḿ '\u1e3f'
	"madeva":                              0x092e,  // म '\u092e'
	"magujarati":                          0x0aae,  // મ '\u0aae'
	"magurmukhi":                          0x0a2e,  // ਮ '\u0a2e'
	"mahapakhhebrew":                      0x05a4,  // ֤ '\u05a4'
	"mahiragana":                          0x307e,  // ま '\u307e'
	"maichattawalowleftthai":              0xf895,  //  '\uf895'
	"maichattawalowrightthai":             0xf894,  //  '\uf894'
	"maichattawathai":                     0x0e4b,  // ๋ '\u0e4b'
	"maichattawaupperleftthai":            0xf893,  //  '\uf893'
	"maieklowleftthai":                    0xf88c,  //  '\uf88c'
	"maieklowrightthai":                   0xf88b,  //  '\uf88b'
	"maiekthai":                           0x0e48,  // ่ '\u0e48'
	"maiekupperleftthai":                  0xf88a,  //  '\uf88a'
	"maihanakatleftthai":                  0xf884,  //  '\uf884'
	"maihanakatthai":                      0x0e31,  // ั '\u0e31'
	"maitaikhuleftthai":                   0xf889,  //  '\uf889'
	"maitaikhuthai":                       0x0e47,  // ็ '\u0e47'
	"maitholowleftthai":                   0xf88f,  //  '\uf88f'
	"maitholowrightthai":                  0xf88e,  //  '\uf88e'
	"maithothai":                          0x0e49,  // ้ '\u0e49'
	"maithoupperleftthai":                 0xf88d,  //  '\uf88d'
	"maitrilowleftthai":                   0xf892,  //  '\uf892'
	"maitrilowrightthai":                  0xf891,  //  '\uf891'
	"maitrithai":                          0x0e4a,  // ๊ '\u0e4a'
	"maitriupperleftthai":                 0xf890,  //  '\uf890'
	"maiyamokthai":                        0x0e46,  // ๆ '\u0e46'
	"makatakana":                          0x30de,  // マ '\u30de'
	"makatakanahalfwidth":                 0xff8f,  // ﾏ '\uff8f'
	"mansyonsquare":                       0x3347,  // ㍇ '\u3347'
	"mapsdown":                            0x21a7,  // ↧ '\u21a7'
	"mapsfrom":                            0x21a4,  // ↤ '\u21a4'
	"mapsto":                              0x21a6,  // ↦ '\u21a6'
	"mapsup":                              0x21a5,  // ↥ '\u21a5'
	"mars":                                0x2642,  // ♂ '\u2642'
	"masoracirclehebrew":                  0x05af,  // ֯ '\u05af'
	"masquare":                            0x3383,  // ㎃ '\u3383'
	"mbfA":                                0x1d400, // 𝐀 '\U0001d400'
	"mbfAlpha":                            0x1d6a8, // 𝚨 '\U0001d6a8'
	"mbfB":                                0x1d401, // 𝐁 '\U0001d401'
	"mbfBeta":                             0x1d6a9, // 𝚩 '\U0001d6a9'
	"mbfC":                                0x1d402, // 𝐂 '\U0001d402'
	"mbfChi":                              0x1d6be, // 𝚾 '\U0001d6be'
	"mbfD":                                0x1d403, // 𝐃 '\U0001d403'
	"mbfDelta":                            0x1d6ab, // 𝚫 '\U0001d6ab'
	"mbfDigamma":                          0x1d7ca, // 𝟊 '\U0001d7ca'
	"mbfE":                                0x1d404, // 𝐄 '\U0001d404'
	"mbfEpsilon":                          0x1d6ac, // 𝚬 '\U0001d6ac'
	"mbfEta":                              0x1d6ae, // 𝚮 '\U0001d6ae'
	"mbfF":                                0x1d405, // 𝐅 '\U0001d405'
	"mbfG":                                0x1d406, // 𝐆 '\U0001d406'
	"mbfGamma":                            0x1d6aa, // 𝚪 '\U0001d6aa'
	"mbfH":                                0x1d407, // 𝐇 '\U0001d407'
	"mbfI":                                0x1d408, // 𝐈 '\U0001d408'
	"mbfIota":                             0x1d6b0, // 𝚰 '\U0001d6b0'
	"mbfJ":                                0x1d409, // 𝐉 '\U0001d409'
	"mbfK":                                0x1d40a, // 𝐊 '\U0001d40a'
	"mbfKappa":                            0x1d6b1, // 𝚱 '\U0001d6b1'
	"mbfL":                                0x1d40b, // 𝐋 '\U0001d40b'
	"mbfLambda":                           0x1d6b2, // 𝚲 '\U0001d6b2'
	"mbfM":                                0x1d40c, // 𝐌 '\U0001d40c'
	"mbfMu":                               0x1d6b3, // 𝚳 '\U0001d6b3'
	"mbfN":                                0x1d40d, // 𝐍 '\U0001d40d'
	"mbfNu":                               0x1d6b4, // 𝚴 '\U0001d6b4'
	"mbfO":                                0x1d40e, // 𝐎 '\U0001d40e'
	"mbfOmega":                            0x1d6c0, // 𝛀 '\U0001d6c0'
	"mbfOmicron":                          0x1d6b6, // 𝚶 '\U0001d6b6'
	"mbfP":                                0x1d40f, // 𝐏 '\U0001d40f'
	"mbfPhi":                              0x1d6bd, // 𝚽 '\U0001d6bd'
	"mbfPi":                               0x1d6b7, // 𝚷 '\U0001d6b7'
	"mbfPsi":                              0x1d6bf, // 𝚿 '\U0001d6bf'
	"mbfQ":                                0x1d410, // 𝐐 '\U0001d410'
	"mbfR":                                0x1d411, // 𝐑 '\U0001d411'
	"mbfRho":                              0x1d6b8, // 𝚸 '\U0001d6b8'
	"mbfS":                                0x1d412, // 𝐒 '\U0001d412'
	"mbfSigma":                            0x1d6ba, // 𝚺 '\U0001d6ba'
	"mbfT":                                0x1d413, // 𝐓 '\U0001d413'
	"mbfTau":                              0x1d6bb, // 𝚻 '\U0001d6bb'
	"mbfTheta":                            0x1d6af, // 𝚯 '\U0001d6af'
	"mbfU":                                0x1d414, // 𝐔 '\U0001d414'
	"mbfUpsilon":                          0x1d6bc, // 𝚼 '\U0001d6bc'
	"mbfV":                                0x1d415, // 𝐕 '\U0001d415'
	"mbfW":                                0x1d416, // 𝐖 '\U0001d416'
	"mbfX":                                0x1d417, // 𝐗 '\U0001d417'
	"mbfXi":                               0x1d6b5, // 𝚵 '\U0001d6b5'
	"mbfY":                                0x1d418, // 𝐘 '\U0001d418'
	"mbfZ":                                0x1d419, // 𝐙 '\U0001d419'
	"mbfZeta":                             0x1d6ad, // 𝚭 '\U0001d6ad'
	"mbfa":                                0x1d41a, // 𝐚 '\U0001d41a'
	"mbfalpha":                            0x1d6c2, // 𝛂 '\U0001d6c2'
	"mbfb":                                0x1d41b, // 𝐛 '\U0001d41b'
	"mbfbeta":                             0x1d6c3, // 𝛃 '\U0001d6c3'
	"mbfc":                                0x1d41c, // 𝐜 '\U0001d41c'
	"mbfchi":                              0x1d6d8, // 𝛘 '\U0001d6d8'
	"mbfd":                                0x1d41d, // 𝐝 '\U0001d41d'
	"mbfdelta":                            0x1d6c5, // 𝛅 '\U0001d6c5'
	"mbfe":                                0x1d41e, // 𝐞 '\U0001d41e'
	"mbfepsilon":                          0x1d6c6, // 𝛆 '\U0001d6c6'
	"mbfeta":                              0x1d6c8, // 𝛈 '\U0001d6c8'
	"mbff":                                0x1d41f, // 𝐟 '\U0001d41f'
	"mbffrakA":                            0x1d56c, // 𝕬 '\U0001d56c'
	"mbffrakB":                            0x1d56d, // 𝕭 '\U0001d56d'
	"mbffrakC":                            0x1d56e, // 𝕮 '\U0001d56e'
	"mbffrakD":                            0x1d56f, // 𝕯 '\U0001d56f'
	"mbffrakE":                            0x1d570, // 𝕰 '\U0001d570'
	"mbffrakF":                            0x1d571, // 𝕱 '\U0001d571'
	"mbffrakG":                            0x1d572, // 𝕲 '\U0001d572'
	"mbffrakH":                            0x1d573, // 𝕳 '\U0001d573'
	"mbffrakI":                            0x1d574, // 𝕴 '\U0001d574'
	"mbffrakJ":                            0x1d575, // 𝕵 '\U0001d575'
	"mbffrakK":                            0x1d576, // 𝕶 '\U0001d576'
	"mbffrakL":                            0x1d577, // 𝕷 '\U0001d577'
	"mbffrakM":                            0x1d578, // 𝕸 '\U0001d578'
	"mbffrakN":                            0x1d579, // 𝕹 '\U0001d579'
	"mbffrakO":                            0x1d57a, // 𝕺 '\U0001d57a'
	"mbffrakP":                            0x1d57b, // 𝕻 '\U0001d57b'
	"mbffrakQ":                            0x1d57c, // 𝕼 '\U0001d57c'
	"mbffrakR":                            0x1d57d, // 𝕽 '\U0001d57d'
	"mbffrakS":                            0x1d57e, // 𝕾 '\U0001d57e'
	"mbffrakT":                            0x1d57f, // 𝕿 '\U0001d57f'
	"mbffrakU":                            0x1d580, // 𝖀 '\U0001d580'
	"mbffrakV":                            0x1d581, // 𝖁 '\U0001d581'
	"mbffrakW":                            0x1d582, // 𝖂 '\U0001d582'
	"mbffrakX":                            0x1d583, // 𝖃 '\U0001d583'
	"mbffrakY":                            0x1d584, // 𝖄 '\U0001d584'
	"mbffrakZ":                            0x1d585, // 𝖅 '\U0001d585'
	"mbffraka":                            0x1d586, // 𝖆 '\U0001d586'
	"mbffrakb":                            0x1d587, // 𝖇 '\U0001d587'
	"mbffrakc":                            0x1d588, // 𝖈 '\U0001d588'
	"mbffrakd":                            0x1d589, // 𝖉 '\U0001d589'
	"mbffrake":                            0x1d58a, // 𝖊 '\U0001d58a'
	"mbffrakf":                            0x1d58b, // 𝖋 '\U0001d58b'
	"mbffrakg":                            0x1d58c, // 𝖌 '\U0001d58c'
	"mbffrakh":                            0x1d58d, // 𝖍 '\U0001d58d'
	"mbffraki":                            0x1d58e, // 𝖎 '\U0001d58e'
	"mbffrakj":                            0x1d58f, // 𝖏 '\U0001d58f'
	"mbffrakk":                            0x1d590, // 𝖐 '\U0001d590'
	"mbffrakl":                            0x1d591, // 𝖑 '\U0001d591'
	"mbffrakm":                            0x1d592, // 𝖒 '\U0001d592'
	"mbffrakn":                            0x1d593, // 𝖓 '\U0001d593'
	"mbffrako":                            0x1d594, // 𝖔 '\U0001d594'
	"mbffrakp":                            0x1d595, // 𝖕 '\U0001d595'
	"mbffrakq":                            0x1d596, // 𝖖 '\U0001d596'
	"mbffrakr":                            0x1d597, // 𝖗 '\U0001d597'
	"mbffraks":                            0x1d598, // 𝖘 '\U0001d598'
	"mbffrakt":                            0x1d599, // 𝖙 '\U0001d599'
	"mbffraku":                            0x1d59a, // 𝖚 '\U0001d59a'
	"mbffrakv":                            0x1d59b, // 𝖛 '\U0001d59b'
	"mbffrakw":                            0x1d59c, // 𝖜 '\U0001d59c'
	"mbffrakx":                            0x1d59d, // 𝖝 '\U0001d59d'
	"mbffraky":                            0x1d59e, // 𝖞 '\U0001d59e'
	"mbffrakz":                            0x1d59f, // 𝖟 '\U0001d59f'
	"mbfg":                                0x1d420, // 𝐠 '\U0001d420'
	"mbfgamma":                            0x1d6c4, // 𝛄 '\U0001d6c4'
	"mbfh":                                0x1d421, // 𝐡 '\U0001d421'
	"mbfi":                                0x1d422, // 𝐢 '\U0001d422'
	"mbfiota":                             0x1d6ca, // 𝛊 '\U0001d6ca'
	"mbfitA":                              0x1d468, // 𝑨 '\U0001d468'
	"mbfitAlpha":                          0x1d71c, // 𝜜 '\U0001d71c'
	"mbfitB":                              0x1d469, // 𝑩 '\U0001d469'
	"mbfitBeta":                           0x1d71d, // 𝜝 '\U0001d71d'
	"mbfitC":                              0x1d46a, // 𝑪 '\U0001d46a'
	"mbfitChi":                            0x1d732, // 𝜲 '\U0001d732'
	"mbfitD":                              0x1d46b, // 𝑫 '\U0001d46b'
	"mbfitDelta":                          0x1d71f, // 𝜟 '\U0001d71f'
	"mbfitE":                              0x1d46c, // 𝑬 '\U0001d46c'
	"mbfitEpsilon":                        0x1d720, // 𝜠 '\U0001d720'
	"mbfitEta":                            0x1d722, // 𝜢 '\U0001d722'
	"mbfitF":                              0x1d46d, // 𝑭 '\U0001d46d'
	"mbfitG":                              0x1d46e, // 𝑮 '\U0001d46e'
	"mbfitGamma":                          0x1d71e, // 𝜞 '\U0001d71e'
	"mbfitH":                              0x1d46f, // 𝑯 '\U0001d46f'
	"mbfitI":                              0x1d470, // 𝑰 '\U0001d470'
	"mbfitIota":                           0x1d724, // 𝜤 '\U0001d724'
	"mbfitJ":                              0x1d471, // 𝑱 '\U0001d471'
	"mbfitK":                              0x1d472, // 𝑲 '\U0001d472'
	"mbfitKappa":                          0x1d725, // 𝜥 '\U0001d725'
	"mbfitL":                              0x1d473, // 𝑳 '\U0001d473'
	"mbfitLambda":                         0x1d726, // 𝜦 '\U0001d726'
	"mbfitM":                              0x1d474, // 𝑴 '\U0001d474'
	"mbfitMu":                             0x1d727, // 𝜧 '\U0001d727'
	"mbfitN":                              0x1d475, // 𝑵 '\U0001d475'
	"mbfitNu":                             0x1d728, // 𝜨 '\U0001d728'
	"mbfitO":                              0x1d476, // 𝑶 '\U0001d476'
	"mbfitOmega":                          0x1d734, // 𝜴 '\U0001d734'
	"mbfitOmicron":                        0x1d72a, // 𝜪 '\U0001d72a'
	"mbfitP":                              0x1d477, // 𝑷 '\U0001d477'
	"mbfitPhi":                            0x1d731, // 𝜱 '\U0001d731'
	"mbfitPi":                             0x1d72b, // 𝜫 '\U0001d72b'
	"mbfitPsi":                            0x1d733, // 𝜳 '\U0001d733'
	"mbfitQ":                              0x1d478, // 𝑸 '\U0001d478'
	"mbfitR":                              0x1d479, // 𝑹 '\U0001d479'
	"mbfitRho":                            0x1d72c, // 𝜬 '\U0001d72c'
	"mbfitS":                              0x1d47a, // 𝑺 '\U0001d47a'
	"mbfitSigma":                          0x1d72e, // 𝜮 '\U0001d72e'
	"mbfitT":                              0x1d47b, // 𝑻 '\U0001d47b'
	"mbfitTau":                            0x1d72f, // 𝜯 '\U0001d72f'
	"mbfitTheta":                          0x1d723, // 𝜣 '\U0001d723'
	"mbfitU":                              0x1d47c, // 𝑼 '\U0001d47c'
	"mbfitUpsilon":                        0x1d730, // 𝜰 '\U0001d730'
	"mbfitV":                              0x1d47d, // 𝑽 '\U0001d47d'
	"mbfitW":                              0x1d47e, // 𝑾 '\U0001d47e'
	"mbfitX":                              0x1d47f, // 𝑿 '\U0001d47f'
	"mbfitXi":                             0x1d729, // 𝜩 '\U0001d729'
	"mbfitY":                              0x1d480, // 𝒀 '\U0001d480'
	"mbfitZ":                              0x1d481, // 𝒁 '\U0001d481'
	"mbfitZeta":                           0x1d721, // 𝜡 '\U0001d721'
	"mbfita":                              0x1d482, // 𝒂 '\U0001d482'
	"mbfitalpha":                          0x1d736, // 𝜶 '\U0001d736'
	"mbfitb":                              0x1d483, // 𝒃 '\U0001d483'
	"mbfitbeta":                           0x1d737, // 𝜷 '\U0001d737'
	"mbfitc":                              0x1d484, // 𝒄 '\U0001d484'
	"mbfitchi":                            0x1d74c, // 𝝌 '\U0001d74c'
	"mbfitd":                              0x1d485, // 𝒅 '\U0001d485'
	"mbfitdelta":                          0x1d739, // 𝜹 '\U0001d739'
	"mbfite":                              0x1d486, // 𝒆 '\U0001d486'
	"mbfitepsilon":                        0x1d73a, // 𝜺 '\U0001d73a'
	"mbfiteta":                            0x1d73c, // 𝜼 '\U0001d73c'
	"mbfitf":                              0x1d487, // 𝒇 '\U0001d487'
	"mbfitg":                              0x1d488, // 𝒈 '\U0001d488'
	"mbfitgamma":                          0x1d738, // 𝜸 '\U0001d738'
	"mbfith":                              0x1d489, // 𝒉 '\U0001d489'
	"mbfiti":                              0x1d48a, // 𝒊 '\U0001d48a'
	"mbfitiota":                           0x1d73e, // 𝜾 '\U0001d73e'
	"mbfitj":                              0x1d48b, // 𝒋 '\U0001d48b'
	"mbfitk":                              0x1d48c, // 𝒌 '\U0001d48c'
	"mbfitkappa":                          0x1d73f, // 𝜿 '\U0001d73f'
	"mbfitl":                              0x1d48d, // 𝒍 '\U0001d48d'
	"mbfitlambda":                         0x1d740, // 𝝀 '\U0001d740'
	"mbfitm":                              0x1d48e, // 𝒎 '\U0001d48e'
	"mbfitmu":                             0x1d741, // 𝝁 '\U0001d741'
	"mbfitn":                              0x1d48f, // 𝒏 '\U0001d48f'
	"mbfitnabla":                          0x1d735, // 𝜵 '\U0001d735'
	"mbfitnu":                             0x1d742, // 𝝂 '\U0001d742'
	"mbfito":                              0x1d490, // 𝒐 '\U0001d490'
	"mbfitomega":                          0x1d74e, // 𝝎 '\U0001d74e'
	"mbfitomicron":                        0x1d744, // 𝝄 '\U0001d744'
	"mbfitp":                              0x1d491, // 𝒑 '\U0001d491'
	"mbfitpartial":                        0x1d74f, // 𝝏 '\U0001d74f'
	"mbfitphi":                            0x1d74b, // 𝝋 '\U0001d74b'
	"mbfitpi":                             0x1d745, // 𝝅 '\U0001d745'
	"mbfitpsi":                            0x1d74d, // 𝝍 '\U0001d74d'
	"mbfitq":                              0x1d492, // 𝒒 '\U0001d492'
	"mbfitr":                              0x1d493, // 𝒓 '\U0001d493'
	"mbfitrho":                            0x1d746, // 𝝆 '\U0001d746'
	"mbfits":                              0x1d494, // 𝒔 '\U0001d494'
	"mbfitsansA":                          0x1d63c, // 𝘼 '\U0001d63c'
	"mbfitsansAlpha":                      0x1d790, // 𝞐 '\U0001d790'
	"mbfitsansB":                          0x1d63d, // 𝘽 '\U0001d63d'
	"mbfitsansBeta":                       0x1d791, // 𝞑 '\U0001d791'
	"mbfitsansC":                          0x1d63e, // 𝘾 '\U0001d63e'
	"mbfitsansChi":                        0x1d7a6, // 𝞦 '\U0001d7a6'
	"mbfitsansD":                          0x1d63f, // 𝘿 '\U0001d63f'
	"mbfitsansDelta":                      0x1d793, // 𝞓 '\U0001d793'
	"mbfitsansE":                          0x1d640, // 𝙀 '\U0001d640'
	"mbfitsansEpsilon":                    0x1d794, // 𝞔 '\U0001d794'
	"mbfitsansEta":                        0x1d796, // 𝞖 '\U0001d796'
	"mbfitsansF":                          0x1d641, // 𝙁 '\U0001d641'
	"mbfitsansG":                          0x1d642, // 𝙂 '\U0001d642'
	"mbfitsansGamma":                      0x1d792, // 𝞒 '\U0001d792'
	"mbfitsansH":                          0x1d643, // 𝙃 '\U0001d643'
	"mbfitsansI":                          0x1d644, // 𝙄 '\U0001d644'
	"mbfitsansIota":                       0x1d798, // 𝞘 '\U0001d798'
	"mbfitsansJ":                          0x1d645, // 𝙅 '\U0001d645'
	"mbfitsansK":                          0x1d646, // 𝙆 '\U0001d646'
	"mbfitsansKappa":                      0x1d799, // 𝞙 '\U0001d799'
	"mbfitsansL":                          0x1d647, // 𝙇 '\U0001d647'
	"mbfitsansLambda":                     0x1d79a, // 𝞚 '\U0001d79a'
	"mbfitsansM":                          0x1d648, // 𝙈 '\U0001d648'
	"mbfitsansMu":                         0x1d79b, // 𝞛 '\U0001d79b'
	"mbfitsansN":                          0x1d649, // 𝙉 '\U0001d649'
	"mbfitsansNu":                         0x1d79c, // 𝞜 '\U0001d79c'
	"mbfitsansO":                          0x1d64a, // 𝙊 '\U0001d64a'
	"mbfitsansOmega":                      0x1d7a8, // 𝞨 '\U0001d7a8'
	"mbfitsansOmicron":                    0x1d79e, // 𝞞 '\U0001d79e'
	"mbfitsansP":                          0x1d64b, // 𝙋 '\U0001d64b'
	"mbfitsansPhi":                        0x1d7a5, // 𝞥 '\U0001d7a5'
	"mbfitsansPi":                         0x1d79f, // 𝞟 '\U0001d79f'
	"mbfitsansPsi":                        0x1d7a7, // 𝞧 '\U0001d7a7'
	"mbfitsansQ":                          0x1d64c, // 𝙌 '\U0001d64c'
	"mbfitsansR":                          0x1d64d, // 𝙍 '\U0001d64d'
	"mbfitsansRho":                        0x1d7a0, // 𝞠 '\U0001d7a0'
	"mbfitsansS":                          0x1d64e, // 𝙎 '\U0001d64e'
	"mbfitsansSigma":                      0x1d7a2, // 𝞢 '\U0001d7a2'
	"mbfitsansT":                          0x1d64f, // 𝙏 '\U0001d64f'
	"mbfitsansTau":                        0x1d7a3, // 𝞣 '\U0001d7a3'
	"mbfitsansTheta":                      0x1d797, // 𝞗 '\U0001d797'
	"mbfitsansU":                          0x1d650, // 𝙐 '\U0001d650'
	"mbfitsansUpsilon":                    0x1d7a4, // 𝞤 '\U0001d7a4'
	"mbfitsansV":                          0x1d651, // 𝙑 '\U0001d651'
	"mbfitsansW":                          0x1d652, // 𝙒 '\U0001d652'
	"mbfitsansX":                          0x1d653, // 𝙓 '\U0001d653'
	"mbfitsansXi":                         0x1d79d, // 𝞝 '\U0001d79d'
	"mbfitsansY":                          0x1d654, // 𝙔 '\U0001d654'
	"mbfitsansZ":                          0x1d655, // 𝙕 '\U0001d655'
	"mbfitsansZeta":                       0x1d795, // 𝞕 '\U0001d795'
	"mbfitsansa":                          0x1d656, // 𝙖 '\U0001d656'
	"mbfitsansalpha":                      0x1d7aa, // 𝞪 '\U0001d7aa'
	"mbfitsansb":                          0x1d657, // 𝙗 '\U0001d657'
	"mbfitsansbeta":                       0x1d7ab, // 𝞫 '\U0001d7ab'
	"mbfitsansc":                          0x1d658, // 𝙘 '\U0001d658'
	"mbfitsanschi":                        0x1d7c0, // 𝟀 '\U0001d7c0'
	"mbfitsansd":                          0x1d659, // 𝙙 '\U0001d659'
	"mbfitsansdelta":                      0x1d7ad, // 𝞭 '\U0001d7ad'
	"mbfitsanse":                          0x1d65a, // 𝙚 '\U0001d65a'
	"mbfitsansepsilon":                    0x1d7ae, // 𝞮 '\U0001d7ae'
	"mbfitsanseta":                        0x1d7b0, // 𝞰 '\U0001d7b0'
	"mbfitsansf":                          0x1d65b, // 𝙛 '\U0001d65b'
	"mbfitsansg":                          0x1d65c, // 𝙜 '\U0001d65c'
	"mbfitsansgamma":                      0x1d7ac, // 𝞬 '\U0001d7ac'
	"mbfitsansh":                          0x1d65d, // 𝙝 '\U0001d65d'
	"mbfitsansi":                          0x1d65e, // 𝙞 '\U0001d65e'
	"mbfitsansiota":                       0x1d7b2, // 𝞲 '\U0001d7b2'
	"mbfitsansj":                          0x1d65f, // 𝙟 '\U0001d65f'
	"mbfitsansk":                          0x1d660, // 𝙠 '\U0001d660'
	"mbfitsanskappa":                      0x1d7b3, // 𝞳 '\U0001d7b3'
	"mbfitsansl":                          0x1d661, // 𝙡 '\U0001d661'
	"mbfitsanslambda":                     0x1d7b4, // 𝞴 '\U0001d7b4'
	"mbfitsansm":                          0x1d662, // 𝙢 '\U0001d662'
	"mbfitsansmu":                         0x1d7b5, // 𝞵 '\U0001d7b5'
	"mbfitsansn":                          0x1d663, // 𝙣 '\U0001d663'
	"mbfitsansnabla":                      0x1d7a9, // 𝞩 '\U0001d7a9'
	"mbfitsansnu":                         0x1d7b6, // 𝞶 '\U0001d7b6'
	"mbfitsanso":                          0x1d664, // 𝙤 '\U0001d664'
	"mbfitsansomega":                      0x1d7c2, // 𝟂 '\U0001d7c2'
	"mbfitsansomicron":                    0x1d7b8, // 𝞸 '\U0001d7b8'
	"mbfitsansp":                          0x1d665, // 𝙥 '\U0001d665'
	"mbfitsanspartial":                    0x1d7c3, // 𝟃 '\U0001d7c3'
	"mbfitsansphi":                        0x1d7bf, // 𝞿 '\U0001d7bf'
	"mbfitsanspi":                         0x1d7b9, // 𝞹 '\U0001d7b9'
	"mbfitsanspsi":                        0x1d7c1, // 𝟁 '\U0001d7c1'
	"mbfitsansq":                          0x1d666, // 𝙦 '\U0001d666'
	"mbfitsansr":                          0x1d667, // 𝙧 '\U0001d667'
	"mbfitsansrho":                        0x1d7ba, // 𝞺 '\U0001d7ba'
	"mbfitsanss":                          0x1d668, // 𝙨 '\U0001d668'
	"mbfitsanssigma":                      0x1d7bc, // 𝞼 '\U0001d7bc'
	"mbfitsanst":                          0x1d669, // 𝙩 '\U0001d669'
	"mbfitsanstau":                        0x1d7bd, // 𝞽 '\U0001d7bd'
	"mbfitsanstheta":                      0x1d7b1, // 𝞱 '\U0001d7b1'
	"mbfitsansu":                          0x1d66a, // 𝙪 '\U0001d66a'
	"mbfitsansupsilon":                    0x1d7be, // 𝞾 '\U0001d7be'
	"mbfitsansv":                          0x1d66b, // 𝙫 '\U0001d66b'
	"mbfitsansvarTheta":                   0x1d7a1, // 𝞡 '\U0001d7a1'
	"mbfitsansvarepsilon":                 0x1d7c4, // 𝟄 '\U0001d7c4'
	"mbfitsansvarkappa":                   0x1d7c6, // 𝟆 '\U0001d7c6'
	"mbfitsansvarphi":                     0x1d7c7, // 𝟇 '\U0001d7c7'
	"mbfitsansvarpi":                      0x1d7c9, // 𝟉 '\U0001d7c9'
	"mbfitsansvarrho":                     0x1d7c8, // 𝟈 '\U0001d7c8'
	"mbfitsansvarsigma":                   0x1d7bb, // 𝞻 '\U0001d7bb'
	"mbfitsansvartheta":                   0x1d7c5, // 𝟅 '\U0001d7c5'
	"mbfitsansw":                          0x1d66c, // 𝙬 '\U0001d66c'
	"mbfitsansx":                          0x1d66d, // 𝙭 '\U0001d66d'
	"mbfitsansxi":                         0x1d7b7, // 𝞷 '\U0001d7b7'
	"mbfitsansy":                          0x1d66e, // 𝙮 '\U0001d66e'
	"mbfitsansz":                          0x1d66f, // 𝙯 '\U0001d66f'
	"mbfitsanszeta":                       0x1d7af, // 𝞯 '\U0001d7af'
	"mbfitsigma":                          0x1d748, // 𝝈 '\U0001d748'
	"mbfitt":                              0x1d495, // 𝒕 '\U0001d495'
	"mbfittau":                            0x1d749, // 𝝉 '\U0001d749'
	"mbfittheta":                          0x1d73d, // 𝜽 '\U0001d73d'
	"mbfitu":                              0x1d496, // 𝒖 '\U0001d496'
	"mbfitupsilon":                        0x1d74a, // 𝝊 '\U0001d74a'
	"mbfitv":                              0x1d497, // 𝒗 '\U0001d497'
	"mbfitvarTheta":                       0x1d72d, // 𝜭 '\U0001d72d'
	"mbfitvarepsilon":                     0x1d750, // 𝝐 '\U0001d750'
	"mbfitvarkappa":                       0x1d752, // 𝝒 '\U0001d752'
	"mbfitvarphi":                         0x1d753, // 𝝓 '\U0001d753'
	"mbfitvarpi":                          0x1d755, // 𝝕 '\U0001d755'
	"mbfitvarrho":                         0x1d754, // 𝝔 '\U0001d754'
	"mbfitvarsigma":                       0x1d747, // 𝝇 '\U0001d747'
	"mbfitvartheta":                       0x1d751, // 𝝑 '\U0001d751'
	"mbfitw":                              0x1d498, // 𝒘 '\U0001d498'
	"mbfitx":                              0x1d499, // 𝒙 '\U0001d499'
	"mbfitxi":                             0x1d743, // 𝝃 '\U0001d743'
	"mbfity":                              0x1d49a, // 𝒚 '\U0001d49a'
	"mbfitz":                              0x1d49b, // 𝒛 '\U0001d49b'
	"mbfitzeta":                           0x1d73b, // 𝜻 '\U0001d73b'
	"mbfj":                                0x1d423, // 𝐣 '\U0001d423'
	"mbfk":                                0x1d424, // 𝐤 '\U0001d424'
	"mbfkappa":                            0x1d6cb, // 𝛋 '\U0001d6cb'
	"mbfl":                                0x1d425, // 𝐥 '\U0001d425'
	"mbflambda":                           0x1d6cc, // 𝛌 '\U0001d6cc'
	"mbfm":                                0x1d426, // 𝐦 '\U0001d426'
	"mbfmu":                               0x1d6cd, // 𝛍 '\U0001d6cd'
	"mbfn":                                0x1d427, // 𝐧 '\U0001d427'
	"mbfnabla":                            0x1d6c1, // 𝛁 '\U0001d6c1'
	"mbfnu":                               0x1d6ce, // 𝛎 '\U0001d6ce'
	"mbfo":                                0x1d428, // 𝐨 '\U0001d428'
	"mbfomega":                            0x1d6da, // 𝛚 '\U0001d6da'
	"mbfomicron":                          0x1d6d0, // 𝛐 '\U0001d6d0'
	"mbfp":                                0x1d429, // 𝐩 '\U0001d429'
	"mbfpartial":                          0x1d6db, // 𝛛 '\U0001d6db'
	"mbfphi":                              0x1d6df, // 𝛟 '\U0001d6df'
	"mbfpi":                               0x1d6d1, // 𝛑 '\U0001d6d1'
	"mbfpsi":                              0x1d6d9, // 𝛙 '\U0001d6d9'
	"mbfq":                                0x1d42a, // 𝐪 '\U0001d42a'
	"mbfr":                                0x1d42b, // 𝐫 '\U0001d42b'
	"mbfrho":                              0x1d6d2, // 𝛒 '\U0001d6d2'
	"mbfs":                                0x1d42c, // 𝐬 '\U0001d42c'
	"mbfsansA":                            0x1d5d4, // 𝗔 '\U0001d5d4'
	"mbfsansAlpha":                        0x1d756, // 𝝖 '\U0001d756'
	"mbfsansB":                            0x1d5d5, // 𝗕 '\U0001d5d5'
	"mbfsansBeta":                         0x1d757, // 𝝗 '\U0001d757'
	"mbfsansC":                            0x1d5d6, // 𝗖 '\U0001d5d6'
	"mbfsansChi":                          0x1d76c, // 𝝬 '\U0001d76c'
	"mbfsansD":                            0x1d5d7, // 𝗗 '\U0001d5d7'
	"mbfsansDelta":                        0x1d759, // 𝝙 '\U0001d759'
	"mbfsansE":                            0x1d5d8, // 𝗘 '\U0001d5d8'
	"mbfsansEpsilon":                      0x1d75a, // 𝝚 '\U0001d75a'
	"mbfsansEta":                          0x1d75c, // 𝝜 '\U0001d75c'
	"mbfsansF":                            0x1d5d9, // 𝗙 '\U0001d5d9'
	"mbfsansG":                            0x1d5da, // 𝗚 '\U0001d5da'
	"mbfsansGamma":                        0x1d758, // 𝝘 '\U0001d758'
	"mbfsansH":                            0x1d5db, // 𝗛 '\U0001d5db'
	"mbfsansI":                            0x1d5dc, // 𝗜 '\U0001d5dc'
	"mbfsansIota":                         0x1d75e, // 𝝞 '\U0001d75e'
	"mbfsansJ":                            0x1d5dd, // 𝗝 '\U0001d5dd'
	"mbfsansK":                            0x1d5de, // 𝗞 '\U0001d5de'
	"mbfsansKappa":                        0x1d75f, // 𝝟 '\U0001d75f'
	"mbfsansL":                            0x1d5df, // 𝗟 '\U0001d5df'
	"mbfsansLambda":                       0x1d760, // 𝝠 '\U0001d760'
	"mbfsansM":                            0x1d5e0, // 𝗠 '\U0001d5e0'
	"mbfsansMu":                           0x1d761, // 𝝡 '\U0001d761'
	"mbfsansN":                            0x1d5e1, // 𝗡 '\U0001d5e1'
	"mbfsansNu":                           0x1d762, // 𝝢 '\U0001d762'
	"mbfsansO":                            0x1d5e2, // 𝗢 '\U0001d5e2'
	"mbfsansOmega":                        0x1d76e, // 𝝮 '\U0001d76e'
	"mbfsansOmicron":                      0x1d764, // 𝝤 '\U0001d764'
	"mbfsansP":                            0x1d5e3, // 𝗣 '\U0001d5e3'
	"mbfsansPhi":                          0x1d76b, // 𝝫 '\U0001d76b'
	"mbfsansPi":                           0x1d765, // 𝝥 '\U0001d765'
	"mbfsansPsi":                          0x1d76d, // 𝝭 '\U0001d76d'
	"mbfsansQ":                            0x1d5e4, // 𝗤 '\U0001d5e4'
	"mbfsansR":                            0x1d5e5, // 𝗥 '\U0001d5e5'
	"mbfsansRho":                          0x1d766, // 𝝦 '\U0001d766'
	"mbfsansS":                            0x1d5e6, // 𝗦 '\U0001d5e6'
	"mbfsansSigma":                        0x1d768, // 𝝨 '\U0001d768'
	"mbfsansT":                            0x1d5e7, // 𝗧 '\U0001d5e7'
	"mbfsansTau":                          0x1d769, // 𝝩 '\U0001d769'
	"mbfsansTheta":                        0x1d75d, // 𝝝 '\U0001d75d'
	"mbfsansU":                            0x1d5e8, // 𝗨 '\U0001d5e8'
	"mbfsansUpsilon":                      0x1d76a, // 𝝪 '\U0001d76a'
	"mbfsansV":                            0x1d5e9, // 𝗩 '\U0001d5e9'
	"mbfsansW":                            0x1d5ea, // 𝗪 '\U0001d5ea'
	"mbfsansX":                            0x1d5eb, // 𝗫 '\U0001d5eb'
	"mbfsansXi":                           0x1d763, // 𝝣 '\U0001d763'
	"mbfsansY":                            0x1d5ec, // 𝗬 '\U0001d5ec'
	"mbfsansZ":                            0x1d5ed, // 𝗭 '\U0001d5ed'
	"mbfsansZeta":                         0x1d75b, // 𝝛 '\U0001d75b'
	"mbfsansa":                            0x1d5ee, // 𝗮 '\U0001d5ee'
	"mbfsansalpha":                        0x1d770, // 𝝰 '\U0001d770'
	"mbfsansb":                            0x1d5ef, // 𝗯 '\U0001d5ef'
	"mbfsansbeta":                         0x1d771, // 𝝱 '\U0001d771'
	"mbfsansc":                            0x1d5f0, // 𝗰 '\U0001d5f0'
	"mbfsanschi":                          0x1d786, // 𝞆 '\U0001d786'
	"mbfsansd":                            0x1d5f1, // 𝗱 '\U0001d5f1'
	"mbfsansdelta":                        0x1d773, // 𝝳 '\U0001d773'
	"mbfsanse":                            0x1d5f2, // 𝗲 '\U0001d5f2'
	"mbfsanseight":                        0x1d7f4, // 𝟴 '\U0001d7f4'
	"mbfsansepsilon":                      0x1d774, // 𝝴 '\U0001d774'
	"mbfsanseta":                          0x1d776, // 𝝶 '\U0001d776'
	"mbfsansf":                            0x1d5f3, // 𝗳 '\U0001d5f3'
	"mbfsansfive":                         0x1d7f1, // 𝟱 '\U0001d7f1'
	"mbfsansfour":                         0x1d7f0, // 𝟰 '\U0001d7f0'
	"mbfsansg":                            0x1d5f4, // 𝗴 '\U0001d5f4'
	"mbfsansgamma":                        0x1d772, // 𝝲 '\U0001d772'
	"mbfsansh":                            0x1d5f5, // 𝗵 '\U0001d5f5'
	"mbfsansi":                            0x1d5f6, // 𝗶 '\U0001d5f6'
	"mbfsansiota":                         0x1d778, // 𝝸 '\U0001d778'
	"mbfsansj":                            0x1d5f7, // 𝗷 '\U0001d5f7'
	"mbfsansk":                            0x1d5f8, // 𝗸 '\U0001d5f8'
	"mbfsanskappa":                        0x1d779, // 𝝹 '\U0001d779'
	"mbfsansl":                            0x1d5f9, // 𝗹 '\U0001d5f9'
	"mbfsanslambda":                       0x1d77a, // 𝝺 '\U0001d77a'
	"mbfsansm":                            0x1d5fa, // 𝗺 '\U0001d5fa'
	"mbfsansmu":                           0x1d77b, // 𝝻 '\U0001d77b'
	"mbfsansn":                            0x1d5fb, // 𝗻 '\U0001d5fb'
	"mbfsansnabla":                        0x1d76f, // 𝝯 '\U0001d76f'
	"mbfsansnine":                         0x1d7f5, // 𝟵 '\U0001d7f5'
	"mbfsansnu":                           0x1d77c, // 𝝼 '\U0001d77c'
	"mbfsanso":                            0x1d5fc, // 𝗼 '\U0001d5fc'
	"mbfsansomega":                        0x1d788, // 𝞈 '\U0001d788'
	"mbfsansomicron":                      0x1d77e, // 𝝾 '\U0001d77e'
	"mbfsansone":                          0x1d7ed, // 𝟭 '\U0001d7ed'
	"mbfsansp":                            0x1d5fd, // 𝗽 '\U0001d5fd'
	"mbfsanspartial":                      0x1d789, // 𝞉 '\U0001d789'
	"mbfsansphi":                          0x1d785, // 𝞅 '\U0001d785'
	"mbfsanspi":                           0x1d77f, // 𝝿 '\U0001d77f'
	"mbfsanspsi":                          0x1d787, // 𝞇 '\U0001d787'
	"mbfsansq":                            0x1d5fe, // 𝗾 '\U0001d5fe'
	"mbfsansr":                            0x1d5ff, // 𝗿 '\U0001d5ff'
	"mbfsansrho":                          0x1d780, // 𝞀 '\U0001d780'
	"mbfsanss":                            0x1d600, // 𝘀 '\U0001d600'
	"mbfsansseven":                        0x1d7f3, // 𝟳 '\U0001d7f3'
	"mbfsanssigma":                        0x1d782, // 𝞂 '\U0001d782'
	"mbfsanssix":                          0x1d7f2, // 𝟲 '\U0001d7f2'
	"mbfsanst":                            0x1d601, // 𝘁 '\U0001d601'
	"mbfsanstau":                          0x1d783, // 𝞃 '\U0001d783'
	"mbfsanstheta":                        0x1d777, // 𝝷 '\U0001d777'
	"mbfsansthree":                        0x1d7ef, // 𝟯 '\U0001d7ef'
	"mbfsanstwo":                          0x1d7ee, // 𝟮 '\U0001d7ee'
	"mbfsansu":                            0x1d602, // 𝘂 '\U0001d602'
	"mbfsansupsilon":                      0x1d784, // 𝞄 '\U0001d784'
	"mbfsansv":                            0x1d603, // 𝘃 '\U0001d603'
	"mbfsansvarTheta":                     0x1d767, // 𝝧 '\U0001d767'
	"mbfsansvarepsilon":                   0x1d78a, // 𝞊 '\U0001d78a'
	"mbfsansvarkappa":                     0x1d78c, // 𝞌 '\U0001d78c'
	"mbfsansvarphi":                       0x1d78d, // 𝞍 '\U0001d78d'
	"mbfsansvarpi":                        0x1d78f, // 𝞏 '\U0001d78f'
	"mbfsansvarrho":                       0x1d78e, // 𝞎 '\U0001d78e'
	"mbfsansvarsigma":                     0x1d781, // 𝞁 '\U0001d781'
	"mbfsansvartheta":                     0x1d78b, // 𝞋 '\U0001d78b'
	"mbfsansw":                            0x1d604, // 𝘄 '\U0001d604'
	"mbfsansx":                            0x1d605, // 𝘅 '\U0001d605'
	"mbfsansxi":                           0x1d77d, // 𝝽 '\U0001d77d'
	"mbfsansy":                            0x1d606, // 𝘆 '\U0001d606'
	"mbfsansz":                            0x1d607, // 𝘇 '\U0001d607'
	"mbfsanszero":                         0x1d7ec, // 𝟬 '\U0001d7ec'
	"mbfsanszeta":                         0x1d775, // 𝝵 '\U0001d775'
	"mbfscrA":                             0x1d4d0, // 𝓐 '\U0001d4d0'
	"mbfscrB":                             0x1d4d1, // 𝓑 '\U0001d4d1'
	"mbfscrC":                             0x1d4d2, // 𝓒 '\U0001d4d2'
	"mbfscrD":                             0x1d4d3, // 𝓓 '\U0001d4d3'
	"mbfscrE":                             0x1d4d4, // 𝓔 '\U0001d4d4'
	"mbfscrF":                             0x1d4d5, // 𝓕 '\U0001d4d5'
	"mbfscrG":                             0x1d4d6, // 𝓖 '\U0001d4d6'
	"mbfscrH":                             0x1d4d7, // 𝓗 '\U0001d4d7'
	"mbfscrI":                             0x1d4d8, // 𝓘 '\U0001d4d8'
	"mbfscrJ":                             0x1d4d9, // 𝓙 '\U0001d4d9'
	"mbfscrK":                             0x1d4da, // 𝓚 '\U0001d4da'
	"mbfscrL":                             0x1d4db, // 𝓛 '\U0001d4db'
	"mbfscrM":                             0x1d4dc, // 𝓜 '\U0001d4dc'
	"mbfscrN":                             0x1d4dd, // 𝓝 '\U0001d4dd'
	"mbfscrO":                             0x1d4de, // 𝓞 '\U0001d4de'
	"mbfscrP":                             0x1d4df, // 𝓟 '\U0001d4df'
	"mbfscrQ":                             0x1d4e0, // 𝓠 '\U0001d4e0'
	"mbfscrR":                             0x1d4e1, // 𝓡 '\U0001d4e1'
	"mbfscrS":                             0x1d4e2, // 𝓢 '\U0001d4e2'
	"mbfscrT":                             0x1d4e3, // 𝓣 '\U0001d4e3'
	"mbfscrU":                             0x1d4e4, // 𝓤 '\U0001d4e4'
	"mbfscrV":                             0x1d4e5, // 𝓥 '\U0001d4e5'
	"mbfscrW":                             0x1d4e6, // 𝓦 '\U0001d4e6'
	"mbfscrX":                             0x1d4e7, // 𝓧 '\U0001d4e7'
	"mbfscrY":                             0x1d4e8, // 𝓨 '\U0001d4e8'
	"mbfscrZ":                             0x1d4e9, // 𝓩 '\U0001d4e9'
	"mbfscra":                             0x1d4ea, // 𝓪 '\U0001d4ea'
	"mbfscrb":                             0x1d4eb, // 𝓫 '\U0001d4eb'
	"mbfscrc":                             0x1d4ec, // 𝓬 '\U0001d4ec'
	"mbfscrd":                             0x1d4ed, // 𝓭 '\U0001d4ed'
	"mbfscre":                             0x1d4ee, // 𝓮 '\U0001d4ee'
	"mbfscrf":                             0x1d4ef, // 𝓯 '\U0001d4ef'
	"mbfscrg":                             0x1d4f0, // 𝓰 '\U0001d4f0'
	"mbfscrh":                             0x1d4f1, // 𝓱 '\U0001d4f1'
	"mbfscri":                             0x1d4f2, // 𝓲 '\U0001d4f2'
	"mbfscrj":                             0x1d4f3, // 𝓳 '\U0001d4f3'
	"mbfscrk":                             0x1d4f4, // 𝓴 '\U0001d4f4'
	"mbfscrl":                             0x1d4f5, // 𝓵 '\U0001d4f5'
	"mbfscrm":                             0x1d4f6, // 𝓶 '\U0001d4f6'
	"mbfscrn":                             0x1d4f7, // 𝓷 '\U0001d4f7'
	"mbfscro":                             0x1d4f8, // 𝓸 '\U0001d4f8'
	"mbfscrp":                             0x1d4f9, // 𝓹 '\U0001d4f9'
	"mbfscrq":                             0x1d4fa, // 𝓺 '\U0001d4fa'
	"mbfscrr":                             0x1d4fb, // 𝓻 '\U0001d4fb'
	"mbfscrs":                             0x1d4fc, // 𝓼 '\U0001d4fc'
	"mbfscrt":                             0x1d4fd, // 𝓽 '\U0001d4fd'
	"mbfscru":                             0x1d4fe, // 𝓾 '\U0001d4fe'
	"mbfscrv":                             0x1d4ff, // 𝓿 '\U0001d4ff'
	"mbfscrw":                             0x1d500, // 𝔀 '\U0001d500'
	"mbfscrx":                             0x1d501, // 𝔁 '\U0001d501'
	"mbfscry":                             0x1d502, // 𝔂 '\U0001d502'
	"mbfscrz":                             0x1d503, // 𝔃 '\U0001d503'
	"mbfsigma":                            0x1d6d4, // 𝛔 '\U0001d6d4'
	"mbft":                                0x1d42d, // 𝐭 '\U0001d42d'
	"mbftau":                              0x1d6d5, // 𝛕 '\U0001d6d5'
	"mbftheta":                            0x1d6c9, // 𝛉 '\U0001d6c9'
	"mbfu":                                0x1d42e, // 𝐮 '\U0001d42e'
	"mbfupsilon":                          0x1d6d6, // 𝛖 '\U0001d6d6'
	"mbfv":                                0x1d42f, // 𝐯 '\U0001d42f'
	"mbfvarTheta":                         0x1d6b9, // 𝚹 '\U0001d6b9'
	"mbfvarepsilon":                       0x1d6dc, // 𝛜 '\U0001d6dc'
	"mbfvarkappa":                         0x1d6de, // 𝛞 '\U0001d6de'
	"mbfvarphi":                           0x1d6d7, // 𝛗 '\U0001d6d7'
	"mbfvarpi":                            0x1d6e1, // 𝛡 '\U0001d6e1'
	"mbfvarrho":                           0x1d6e0, // 𝛠 '\U0001d6e0'
	"mbfvarsigma":                         0x1d6d3, // 𝛓 '\U0001d6d3'
	"mbfvartheta":                         0x1d6dd, // 𝛝 '\U0001d6dd'
	"mbfw":                                0x1d430, // 𝐰 '\U0001d430'
	"mbfx":                                0x1d431, // 𝐱 '\U0001d431'
	"mbfxi":                               0x1d6cf, // 𝛏 '\U0001d6cf'
	"mbfy":                                0x1d432, // 𝐲 '\U0001d432'
	"mbfz":                                0x1d433, // 𝐳 '\U0001d433'
	"mbfzeta":                             0x1d6c7, // 𝛇 '\U0001d6c7'
	"mbopomofo":                           0x3107,  // ㄇ '\u3107'
	"mbsquare":                            0x33d4,  // ㏔ '\u33d4'
	"mcircle":                             0x24dc,  // ⓜ '\u24dc'
	"mcubedsquare":                        0x33a5,  // ㎥ '\u33a5'
	"mdblkcircle":                         0x26ab,  // ⚫ '\u26ab'
	"mdblkdiamond":                        0x2b25,  // ⬥ '\u2b25'
	"mdblklozenge":                        0x2b27,  // ⬧ '\u2b27'
	"mdblksquare":                         0x25fc,  // ◼ '\u25fc'
	"mdlgblklozenge":                      0x29eb,  // ⧫ '\u29eb'
	"mdotaccent":                          0x1e41,  // ṁ '\u1e41'
	"mdotbelow":                           0x1e43,  // ṃ '\u1e43'
	"mdsmblkcircle":                       0x2981,  // ⦁ '\u2981'
	"mdsmblksquare":                       0x25fe,  // ◾ '\u25fe'
	"mdsmwhtcircle":                       0x26ac,  // ⚬ '\u26ac'
	"mdsmwhtsquare":                       0x25fd,  // ◽ '\u25fd'
	"mdwhtcircle":                         0x26aa,  // ⚪ '\u26aa'
	"mdwhtdiamond":                        0x2b26,  // ⬦ '\u2b26'
	"mdwhtlozenge":                        0x2b28,  // ⬨ '\u2b28'
	"mdwhtsquare":                         0x25fb,  // ◻ '\u25fb'
	"measangledltosw":                     0x29af,  // ⦯ '\u29af'
	"measangledrtose":                     0x29ae,  // ⦮ '\u29ae'
	"measangleldtosw":                     0x29ab,  // ⦫ '\u29ab'
	"measanglelutonw":                     0x29a9,  // ⦩ '\u29a9'
	"measanglerdtose":                     0x29aa,  // ⦪ '\u29aa'
	"measanglerutone":                     0x29a8,  // ⦨ '\u29a8'
	"measangleultonw":                     0x29ad,  // ⦭ '\u29ad'
	"measangleurtone":                     0x29ac,  // ⦬ '\u29ac'
	"measeq":                              0x225e,  // ≞ '\u225e'
	"measuredangle":                       0x2221,  // ∡ '\u2221'
	"measuredangleleft":                   0x299b,  // ⦛ '\u299b'
	"measuredrightangle":                  0x22be,  // ⊾ '\u22be'
	"medblackstar":                        0x2b51,  // ⭑ '\u2b51'
	"medwhitestar":                        0x2b50,  // ⭐ '\u2b50'
	"meemfinalarabic":                     0xfee2,  // ﻢ '\ufee2'
	"meeminitialarabic":                   0xfee3,  // ﻣ '\ufee3'
	"meemisolated":                        0xfee1,  // ﻡ '\ufee1'
	"meemmedialarabic":                    0xfee4,  // ﻤ '\ufee4'
	"meemmeeminitialarabic":               0xfcd1,  // ﳑ '\ufcd1'
	"meemmeemisolatedarabic":              0xfc48,  // ﱈ '\ufc48'
	"meemwithhahinitial":                  0xfccf,  // ﳏ '\ufccf'
	"meemwithjeeminitial":                 0xfcce,  // ﳎ '\ufcce'
	"meemwithkhahinitial":                 0xfcd0,  // ﳐ '\ufcd0'
	"meetorusquare":                       0x334d,  // ㍍ '\u334d'
	"mehiragana":                          0x3081,  // め '\u3081'
	"meizierasquare":                      0x337e,  // ㍾ '\u337e'
	"mekatakana":                          0x30e1,  // メ '\u30e1'
	"mekatakanahalfwidth":                 0xff92,  // ﾒ '\uff92'
	"mem":                                 0x05de,  // מ '\u05de'
	"memdageshhebrew":                     0xfb3e,  // מּ '\ufb3e'
	"menarmenian":                         0x0574,  // մ '\u0574'
	"merkhahebrew":                        0x05a5,  // ֥ '\u05a5'
	"merkhakefulahebrew":                  0x05a6,  // ֦ '\u05a6'
	"mfrakA":                              0x1d504, // 𝔄 '\U0001d504'
	"mfrakB":                              0x1d505, // 𝔅 '\U0001d505'
	"mfrakC":                              0x212d,  // ℭ '\u212d'
	"mfrakD":                              0x1d507, // 𝔇 '\U0001d507'
	"mfrakE":                              0x1d508, // 𝔈 '\U0001d508'
	"mfrakF":                              0x1d509, // 𝔉 '\U0001d509'
	"mfrakG":                              0x1d50a, // 𝔊 '\U0001d50a'
	"mfrakH":                              0x210c,  // ℌ '\u210c'
	"mfrakJ":                              0x1d50d, // 𝔍 '\U0001d50d'
	"mfrakK":                              0x1d50e, // 𝔎 '\U0001d50e'
	"mfrakL":                              0x1d50f, // 𝔏 '\U0001d50f'
	"mfrakM":                              0x1d510, // 𝔐 '\U0001d510'
	"mfrakN":                              0x1d511, // 𝔑 '\U0001d511'
	"mfrakO":                              0x1d512, // 𝔒 '\U0001d512'
	"mfrakP":                              0x1d513, // 𝔓 '\U0001d513'
	"mfrakQ":                              0x1d514, // 𝔔 '\U0001d514'
	"mfrakS":                              0x1d516, // 𝔖 '\U0001d516'
	"mfrakT":                              0x1d517, // 𝔗 '\U0001d517'
	"mfrakU":                              0x1d518, // 𝔘 '\U0001d518'
	"mfrakV":                              0x1d519, // 𝔙 '\U0001d519'
	"mfrakW":                              0x1d51a, // 𝔚 '\U0001d51a'
	"mfrakX":                              0x1d51b, // 𝔛 '\U0001d51b'
	"mfrakY":                              0x1d51c, // 𝔜 '\U0001d51c'
	"mfrakZ":                              0x2128,  // ℨ '\u2128'
	"mfraka":                              0x1d51e, // 𝔞 '\U0001d51e'
	"mfrakb":                              0x1d51f, // 𝔟 '\U0001d51f'
	"mfrakc":                              0x1d520, // 𝔠 '\U0001d520'
	"mfrakd":                              0x1d521, // 𝔡 '\U0001d521'
	"mfrake":                              0x1d522, // 𝔢 '\U0001d522'
	"mfrakf":                              0x1d523, // 𝔣 '\U0001d523'
	"mfrakg":                              0x1d524, // 𝔤 '\U0001d524'
	"mfrakh":                              0x1d525, // 𝔥 '\U0001d525'
	"mfraki":                              0x1d526, // 𝔦 '\U0001d526'
	"mfrakj":                              0x1d527, // 𝔧 '\U0001d527'
	"mfrakk":                              0x1d528, // 𝔨 '\U0001d528'
	"mfrakl":                              0x1d529, // 𝔩 '\U0001d529'
	"mfrakm":                              0x1d52a, // 𝔪 '\U0001d52a'
	"mfrakn":                              0x1d52b, // 𝔫 '\U0001d52b'
	"mfrako":                              0x1d52c, // 𝔬 '\U0001d52c'
	"mfrakp":                              0x1d52d, // 𝔭 '\U0001d52d'
	"mfrakq":                              0x1d52e, // 𝔮 '\U0001d52e'
	"mfrakr":                              0x1d52f, // 𝔯 '\U0001d52f'
	"mfraks":                              0x1d530, // 𝔰 '\U0001d530'
	"mfrakt":                              0x1d531, // 𝔱 '\U0001d531'
	"mfraku":                              0x1d532, // 𝔲 '\U0001d532'
	"mfrakv":                              0x1d533, // 𝔳 '\U0001d533'
	"mfrakw":                              0x1d534, // 𝔴 '\U0001d534'
	"mfrakx":                              0x1d535, // 𝔵 '\U0001d535'
	"mfraky":                              0x1d536, // 𝔶 '\U0001d536'
	"mfrakz":                              0x1d537, // 𝔷 '\U0001d537'
	"mhook":                               0x0271,  // ɱ '\u0271'
	"mhzsquare":                           0x3392,  // ㎒ '\u3392'
	"micro":                               0x0095,  //  '\u0095'
	"midbarvee":                           0x2a5d,  // ⩝ '\u2a5d'
	"midbarwedge":                         0x2a5c,  // ⩜ '\u2a5c'
	"midcir":                              0x2af0,  // ⫰ '\u2af0'
	"middledotkatakanahalfwidth":          0xff65,  // ･ '\uff65'
	"mieumacirclekorean":                  0x3272,  // ㉲ '\u3272'
	"mieumaparenkorean":                   0x3212,  // ㈒ '\u3212'
	"mieumcirclekorean":                   0x3264,  // ㉤ '\u3264'
	"mieumkorean":                         0x3141,  // ㅁ '\u3141'
	"mieumpansioskorean":                  0x3170,  // ㅰ '\u3170'
	"mieumparenkorean":                    0x3204,  // ㈄ '\u3204'
	"mieumpieupkorean":                    0x316e,  // ㅮ '\u316e'
	"mieumsioskorean":                     0x316f,  // ㅯ '\u316f'
	"mihiragana":                          0x307f,  // み '\u307f'
	"mikatakana":                          0x30df,  // ミ '\u30df'
	"mikatakanahalfwidth":                 0xff90,  // ﾐ '\uff90'
	"mill":                                0x20a5,  // ₥ '\u20a5'
	"minus":                               0x2212,  // − '\u2212'
	"minusbelowcmb":                       0x0320,  // ̠ '\u0320'
	"minuscircle":                         0x2296,  // ⊖ '\u2296'
	"minusdot":                            0x2a2a,  // ⨪ '\u2a2a'
	"minusfdots":                          0x2a2b,  // ⨫ '\u2a2b'
	"minusinferior":                       0x208b,  // ₋ '\u208b'
	"minusmod":                            0x02d7,  // ˗ '\u02d7'
	"minusplus":                           0x2213,  // ∓ '\u2213'
	"minusrdots":                          0x2a2c,  // ⨬ '\u2a2c'
	"minussuperior":                       0x207b,  // ⁻ '\u207b'
	"minute":                              0x2032,  // ′ '\u2032'
	"miribaarusquare":                     0x334a,  // ㍊ '\u334a'
	"mirisquare":                          0x3349,  // ㍉ '\u3349'
	"mitA":                                0x1d434, // 𝐴 '\U0001d434'
	"mitAlpha":                            0x1d6e2, // 𝛢 '\U0001d6e2'
	"mitB":                                0x1d435, // 𝐵 '\U0001d435'
	"mitBbbD":                             0x2145,  // ⅅ '\u2145'
	"mitBbbd":                             0x2146,  // ⅆ '\u2146'
	"mitBbbe":                             0x2147,  // ⅇ '\u2147'
	"mitBbbi":                             0x2148,  // ⅈ '\u2148'
	"mitBbbj":                             0x2149,  // ⅉ '\u2149'
	"mitBeta":                             0x1d6e3, // 𝛣 '\U0001d6e3'
	"mitC":                                0x1d436, // 𝐶 '\U0001d436'
	"mitChi":                              0x1d6f8, // 𝛸 '\U0001d6f8'
	"mitD":                                0x1d437, // 𝐷 '\U0001d437'
	"mitDelta":                            0x1d6e5, // 𝛥 '\U0001d6e5'
	"mitE":                                0x1d438, // 𝐸 '\U0001d438'
	"mitEpsilon":                          0x1d6e6, // 𝛦 '\U0001d6e6'
	"mitEta":                              0x1d6e8, // 𝛨 '\U0001d6e8'
	"mitF":                                0x1d439, // 𝐹 '\U0001d439'
	"mitG":                                0x1d43a, // 𝐺 '\U0001d43a'
	"mitGamma":                            0x1d6e4, // 𝛤 '\U0001d6e4'
	"mitH":                                0x1d43b, // 𝐻 '\U0001d43b'
	"mitI":                                0x1d43c, // 𝐼 '\U0001d43c'
	"mitIota":                             0x1d6ea, // 𝛪 '\U0001d6ea'
	"mitJ":                                0x1d43d, // 𝐽 '\U0001d43d'
	"mitK":                                0x1d43e, // 𝐾 '\U0001d43e'
	"mitKappa":                            0x1d6eb, // 𝛫 '\U0001d6eb'
	"mitL":                                0x1d43f, // 𝐿 '\U0001d43f'
	"mitLambda":                           0x1d6ec, // 𝛬 '\U0001d6ec'
	"mitM":                                0x1d440, // 𝑀 '\U0001d440'
	"mitMu":                               0x1d6ed, // 𝛭 '\U0001d6ed'
	"mitN":                                0x1d441, // 𝑁 '\U0001d441'
	"mitNu":                               0x1d6ee, // 𝛮 '\U0001d6ee'
	"mitO":                                0x1d442, // 𝑂 '\U0001d442'
	"mitOmega":                            0x1d6fa, // 𝛺 '\U0001d6fa'
	"mitOmicron":                          0x1d6f0, // 𝛰 '\U0001d6f0'
	"mitP":                                0x1d443, // 𝑃 '\U0001d443'
	"mitPhi":                              0x1d6f7, // 𝛷 '\U0001d6f7'
	"mitPi":                               0x1d6f1, // 𝛱 '\U0001d6f1'
	"mitPsi":                              0x1d6f9, // 𝛹 '\U0001d6f9'
	"mitQ":                                0x1d444, // 𝑄 '\U0001d444'
	"mitR":                                0x1d445, // 𝑅 '\U0001d445'
	"mitRho":                              0x1d6f2, // 𝛲 '\U0001d6f2'
	"mitS":                                0x1d446, // 𝑆 '\U0001d446'
	"mitSigma":                            0x1d6f4, // 𝛴 '\U0001d6f4'
	"mitT":                                0x1d447, // 𝑇 '\U0001d447'
	"mitTau":                              0x1d6f5, // 𝛵 '\U0001d6f5'
	"mitTheta":                            0x1d6e9, // 𝛩 '\U0001d6e9'
	"mitU":                                0x1d448, // 𝑈 '\U0001d448'
	"mitUpsilon":                          0x1d6f6, // 𝛶 '\U0001d6f6'
	"mitV":                                0x1d449, // 𝑉 '\U0001d449'
	"mitW":                                0x1d44a, // 𝑊 '\U0001d44a'
	"mitX":                                0x1d44b, // 𝑋 '\U0001d44b'
	"mitXi":                               0x1d6ef, // 𝛯 '\U0001d6ef'
	"mitY":                                0x1d44c, // 𝑌 '\U0001d44c'
	"mitZ":                                0x1d44d, // 𝑍 '\U0001d44d'
	"mitZeta":                             0x1d6e7, // 𝛧 '\U0001d6e7'
	"mita":                                0x1d44e, // 𝑎 '\U0001d44e'
	"mitalpha":                            0x1d6fc, // 𝛼 '\U0001d6fc'
	"mitb":                                0x1d44f, // 𝑏 '\U0001d44f'
	"mitbeta":                             0x1d6fd, // 𝛽 '\U0001d6fd'
	"mitc":                                0x1d450, // 𝑐 '\U0001d450'
	"mitchi":                              0x1d712, // 𝜒 '\U0001d712'
	"mitd":                                0x1d451, // 𝑑 '\U0001d451'
	"mitdelta":                            0x1d6ff, // 𝛿 '\U0001d6ff'
	"mite":                                0x1d452, // 𝑒 '\U0001d452'
	"mitepsilon":                          0x1d700, // 𝜀 '\U0001d700'
	"miteta":                              0x1d702, // 𝜂 '\U0001d702'
	"mitf":                                0x1d453, // 𝑓 '\U0001d453'
	"mitg":                                0x1d454, // 𝑔 '\U0001d454'
	"mitgamma":                            0x1d6fe, // 𝛾 '\U0001d6fe'
	"miti":                                0x1d456, // 𝑖 '\U0001d456'
	"mitiota":                             0x1d704, // 𝜄 '\U0001d704'
	"mitj":                                0x1d457, // 𝑗 '\U0001d457'
	"mitk":                                0x1d458, // 𝑘 '\U0001d458'
	"mitkappa":                            0x1d705, // 𝜅 '\U0001d705'
	"mitl":                                0x1d459, // 𝑙 '\U0001d459'
	"mitlambda":                           0x1d706, // 𝜆 '\U0001d706'
	"mitm":                                0x1d45a, // 𝑚 '\U0001d45a'
	"mitmu":                               0x1d707, // 𝜇 '\U0001d707'
	"mitn":                                0x1d45b, // 𝑛 '\U0001d45b'
	"mitnabla":                            0x1d6fb, // 𝛻 '\U0001d6fb'
	"mitnu":                               0x1d708, // 𝜈 '\U0001d708'
	"mito":                                0x1d45c, // 𝑜 '\U0001d45c'
	"mitomega":                            0x1d714, // 𝜔 '\U0001d714'
	"mitomicron":                          0x1d70a, // 𝜊 '\U0001d70a'
	"mitp":                                0x1d45d, // 𝑝 '\U0001d45d'
	"mitpartial":                          0x1d715, // 𝜕 '\U0001d715'
	"mitphi":                              0x1d711, // 𝜑 '\U0001d711'
	"mitpi":                               0x1d70b, // 𝜋 '\U0001d70b'
	"mitpsi":                              0x1d713, // 𝜓 '\U0001d713'
	"mitq":                                0x1d45e, // 𝑞 '\U0001d45e'
	"mitr":                                0x1d45f, // 𝑟 '\U0001d45f'
	"mitrho":                              0x1d70c, // 𝜌 '\U0001d70c'
	"mits":                                0x1d460, // 𝑠 '\U0001d460'
	"mitsansA":                            0x1d608, // 𝘈 '\U0001d608'
	"mitsansB":                            0x1d609, // 𝘉 '\U0001d609'
	"mitsansC":                            0x1d60a, // 𝘊 '\U0001d60a'
	"mitsansD":                            0x1d60b, // 𝘋 '\U0001d60b'
	"mitsansE":                            0x1d60c, // 𝘌 '\U0001d60c'
	"mitsansF":                            0x1d60d, // 𝘍 '\U0001d60d'
	"mitsansG":                            0x1d60e, // 𝘎 '\U0001d60e'
	"mitsansH":                            0x1d60f, // 𝘏 '\U0001d60f'
	"mitsansI":                            0x1d610, // 𝘐 '\U0001d610'
	"mitsansJ":                            0x1d611, // 𝘑 '\U0001d611'
	"mitsansK":                            0x1d612, // 𝘒 '\U0001d612'
	"mitsansL":                            0x1d613, // 𝘓 '\U0001d613'
	"mitsansM":                            0x1d614, // 𝘔 '\U0001d614'
	"mitsansN":                            0x1d615, // 𝘕 '\U0001d615'
	"mitsansO":                            0x1d616, // 𝘖 '\U0001d616'
	"mitsansP":                            0x1d617, // 𝘗 '\U0001d617'
	"mitsansQ":                            0x1d618, // 𝘘 '\U0001d618'
	"mitsansR":                            0x1d619, // 𝘙 '\U0001d619'
	"mitsansS":                            0x1d61a, // 𝘚 '\U0001d61a'
	"mitsansT":                            0x1d61b, // 𝘛 '\U0001d61b'
	"mitsansU":                            0x1d61c, // 𝘜 '\U0001d61c'
	"mitsansV":                            0x1d61d, // 𝘝 '\U0001d61d'
	"mitsansW":                            0x1d61e, // 𝘞 '\U0001d61e'
	"mitsansX":                            0x1d61f, // 𝘟 '\U0001d61f'
	"mitsansY":                            0x1d620, // 𝘠 '\U0001d620'
	"mitsansZ":                            0x1d621, // 𝘡 '\U0001d621'
	"mitsansa":                            0x1d622, // 𝘢 '\U0001d622'
	"mitsansb":                            0x1d623, // 𝘣 '\U0001d623'
	"mitsansc":                            0x1d624, // 𝘤 '\U0001d624'
	"mitsansd":                            0x1d625, // 𝘥 '\U0001d625'
	"mitsanse":                            0x1d626, // 𝘦 '\U0001d626'
	"mitsansf":                            0x1d627, // 𝘧 '\U0001d627'
	"mitsansg":                            0x1d628, // 𝘨 '\U0001d628'
	"mitsansh":                            0x1d629, // 𝘩 '\U0001d629'
	"mitsansi":                            0x1d62a, // 𝘪 '\U0001d62a'
	"mitsansj":                            0x1d62b, // 𝘫 '\U0001d62b'
	"mitsansk":                            0x1d62c, // 𝘬 '\U0001d62c'
	"mitsansl":                            0x1d62d, // 𝘭 '\U0001d62d'
	"mitsansm":                            0x1d62e, // 𝘮 '\U0001d62e'
	"mitsansn":                            0x1d62f, // 𝘯 '\U0001d62f'
	"mitsanso":                            0x1d630, // 𝘰 '\U0001d630'
	"mitsansp":                            0x1d631, // 𝘱 '\U0001d631'
	"mitsansq":                            0x1d632, // 𝘲 '\U0001d632'
	"mitsansr":                            0x1d633, // 𝘳 '\U0001d633'
	"mitsanss":                            0x1d634, // 𝘴 '\U0001d634'
	"mitsanst":                            0x1d635, // 𝘵 '\U0001d635'
	"mitsansu":                            0x1d636, // 𝘶 '\U0001d636'
	"mitsansv":                            0x1d637, // 𝘷 '\U0001d637'
	"mitsansw":                            0x1d638, // 𝘸 '\U0001d638'
	"mitsansx":                            0x1d639, // 𝘹 '\U0001d639'
	"mitsansy":                            0x1d63a, // 𝘺 '\U0001d63a'
	"mitsansz":                            0x1d63b, // 𝘻 '\U0001d63b'
	"mitsigma":                            0x1d70e, // 𝜎 '\U0001d70e'
	"mitt":                                0x1d461, // 𝑡 '\U0001d461'
	"mittau":                              0x1d70f, // 𝜏 '\U0001d70f'
	"mittheta":                            0x1d703, // 𝜃 '\U0001d703'
	"mitu":                                0x1d462, // 𝑢 '\U0001d462'
	"mitupsilon":                          0x1d710, // 𝜐 '\U0001d710'
	"mitv":                                0x1d463, // 𝑣 '\U0001d463'
	"mitvarTheta":                         0x1d6f3, // 𝛳 '\U0001d6f3'
	"mitvarepsilon":                       0x1d716, // 𝜖 '\U0001d716'
	"mitvarkappa":                         0x1d718, // 𝜘 '\U0001d718'
	"mitvarphi":                           0x1d719, // 𝜙 '\U0001d719'
	"mitvarpi":                            0x1d71b, // 𝜛 '\U0001d71b'
	"mitvarrho":                           0x1d71a, // 𝜚 '\U0001d71a'
	"mitvarsigma":                         0x1d70d, // 𝜍 '\U0001d70d'
	"mitvartheta":                         0x1d717, // 𝜗 '\U0001d717'
	"mitw":                                0x1d464, // 𝑤 '\U0001d464'
	"mitx":                                0x1d465, // 𝑥 '\U0001d465'
	"mitxi":                               0x1d709, // 𝜉 '\U0001d709'
	"mity":                                0x1d466, // 𝑦 '\U0001d466'
	"mitz":                                0x1d467, // 𝑧 '\U0001d467'
	"mitzeta":                             0x1d701, // 𝜁 '\U0001d701'
	"mlcp":                                0x2adb,  // ⫛ '\u2adb'
	"mlonglegturned":                      0x0270,  // ɰ '\u0270'
	"mlsquare":                            0x3396,  // ㎖ '\u3396'
	"mmcubedsquare":                       0x33a3,  // ㎣ '\u33a3'
	"mmonospace":                          0xff4d,  // ｍ '\uff4d'
	"mmsquaredsquare":                     0x339f,  // ㎟ '\u339f'
	"models":                              0x22a7,  // ⊧ '\u22a7'
	"modtwosum":                           0x2a0a,  // ⨊ '\u2a0a'
	"mohiragana":                          0x3082,  // も '\u3082'
	"mohmsquare":                          0x33c1,  // ㏁ '\u33c1'
	"mokatakana":                          0x30e2,  // モ '\u30e2'
	"mokatakanahalfwidth":                 0xff93,  // ﾓ '\uff93'
	"molsquare":                           0x33d6,  // ㏖ '\u33d6'
	"momathai":                            0x0e21,  // ม '\u0e21'
	"moverssquare":                        0x33a7,  // ㎧ '\u33a7'
	"moverssquaredsquare":                 0x33a8,  // ㎨ '\u33a8'
	"mparen":                              0x24a8,  // ⒨ '\u24a8'
	"mpasquare":                           0x33ab,  // ㎫ '\u33ab'
	"msansA":                              0x1d5a0, // 𝖠 '\U0001d5a0'
	"msansB":                              0x1d5a1, // 𝖡 '\U0001d5a1'
	"msansC":                              0x1d5a2, // 𝖢 '\U0001d5a2'
	"msansD":                              0x1d5a3, // 𝖣 '\U0001d5a3'
	"msansE":                              0x1d5a4, // 𝖤 '\U0001d5a4'
	"msansF":                              0x1d5a5, // 𝖥 '\U0001d5a5'
	"msansG":                              0x1d5a6, // 𝖦 '\U0001d5a6'
	"msansH":                              0x1d5a7, // 𝖧 '\U0001d5a7'
	"msansI":                              0x1d5a8, // 𝖨 '\U0001d5a8'
	"msansJ":                              0x1d5a9, // 𝖩 '\U0001d5a9'
	"msansK":                              0x1d5aa, // 𝖪 '\U0001d5aa'
	"msansL":                              0x1d5ab, // 𝖫 '\U0001d5ab'
	"msansM":                              0x1d5ac, // 𝖬 '\U0001d5ac'
	"msansN":                              0x1d5ad, // 𝖭 '\U0001d5ad'
	"msansO":                              0x1d5ae, // 𝖮 '\U0001d5ae'
	"msansP":                              0x1d5af, // 𝖯 '\U0001d5af'
	"msansQ":                              0x1d5b0, // 𝖰 '\U0001d5b0'
	"msansR":                              0x1d5b1, // 𝖱 '\U0001d5b1'
	"msansS":                              0x1d5b2, // 𝖲 '\U0001d5b2'
	"msansT":                              0x1d5b3, // 𝖳 '\U0001d5b3'
	"msansU":                              0x1d5b4, // 𝖴 '\U0001d5b4'
	"msansV":                              0x1d5b5, // 𝖵 '\U0001d5b5'
	"msansW":                              0x1d5b6, // 𝖶 '\U0001d5b6'
	"msansX":                              0x1d5b7, // 𝖷 '\U0001d5b7'
	"msansY":                              0x1d5b8, // 𝖸 '\U0001d5b8'
	"msansZ":                              0x1d5b9, // 𝖹 '\U0001d5b9'
	"msansa":                              0x1d5ba, // 𝖺 '\U0001d5ba'
	"msansb":                              0x1d5bb, // 𝖻 '\U0001d5bb'
	"msansc":                              0x1d5bc, // 𝖼 '\U0001d5bc'
	"msansd":                              0x1d5bd, // 𝖽 '\U0001d5bd'
	"msanse":                              0x1d5be, // 𝖾 '\U0001d5be'
	"msanseight":                          0x1d7ea, // 𝟪 '\U0001d7ea'
	"msansf":                              0x1d5bf, // 𝖿 '\U0001d5bf'
	"msansfive":                           0x1d7e7, // 𝟧 '\U0001d7e7'
	"msansfour":                           0x1d7e6, // 𝟦 '\U0001d7e6'
	"msansg":                              0x1d5c0, // 𝗀 '\U0001d5c0'
	"msansh":                              0x1d5c1, // 𝗁 '\U0001d5c1'
	"msansi":                              0x1d5c2, // 𝗂 '\U0001d5c2'
	"msansj":                              0x1d5c3, // 𝗃 '\U0001d5c3'
	"msansk":                              0x1d5c4, // 𝗄 '\U0001d5c4'
	"msansl":                              0x1d5c5, // 𝗅 '\U0001d5c5'
	"msansm":                              0x1d5c6, // 𝗆 '\U0001d5c6'
	"msansn":                              0x1d5c7, // 𝗇 '\U0001d5c7'
	"msansnine":                           0x1d7eb, // 𝟫 '\U0001d7eb'
	"msanso":                              0x1d5c8, // 𝗈 '\U0001d5c8'
	"msansone":                            0x1d7e3, // 𝟣 '\U0001d7e3'
	"msansp":                              0x1d5c9, // 𝗉 '\U0001d5c9'
	"msansq":                              0x1d5ca, // 𝗊 '\U0001d5ca'
	"msansr":                              0x1d5cb, // 𝗋 '\U0001d5cb'
	"msanss":                              0x1d5cc, // 𝗌 '\U0001d5cc'
	"msansseven":                          0x1d7e9, // 𝟩 '\U0001d7e9'
	"msanssix":                            0x1d7e8, // 𝟨 '\U0001d7e8'
	"msanst":                              0x1d5cd, // 𝗍 '\U0001d5cd'
	"msansthree":                          0x1d7e5, // 𝟥 '\U0001d7e5'
	"msanstwo":                            0x1d7e4, // 𝟤 '\U0001d7e4'
	"msansu":                              0x1d5ce, // 𝗎 '\U0001d5ce'
	"msansv":                              0x1d5cf, // 𝗏 '\U0001d5cf'
	"msansw":                              0x1d5d0, // 𝗐 '\U0001d5d0'
	"msansx":                              0x1d5d1, // 𝗑 '\U0001d5d1'
	"msansy":                              0x1d5d2, // 𝗒 '\U0001d5d2'
	"msansz":                              0x1d5d3, // 𝗓 '\U0001d5d3'
	"msanszero":                           0x1d7e2, // 𝟢 '\U0001d7e2'
	"mscrA":                               0x1d49c, // 𝒜 '\U0001d49c'
	"mscrB":                               0x212c,  // ℬ '\u212c'
	"mscrC":                               0x1d49e, // 𝒞 '\U0001d49e'
	"mscrD":                               0x1d49f, // 𝒟 '\U0001d49f'
	"mscrE":                               0x2130,  // ℰ '\u2130'
	"mscrF":                               0x2131,  // ℱ '\u2131'
	"mscrG":                               0x1d4a2, // 𝒢 '\U0001d4a2'
	"mscrH":                               0x210b,  // ℋ '\u210b'
	"mscrI":                               0x2110,  // ℐ '\u2110'
	"mscrJ":                               0x1d4a5, // 𝒥 '\U0001d4a5'
	"mscrK":                               0x1d4a6, // 𝒦 '\U0001d4a6'
	"mscrL":                               0x2112,  // ℒ '\u2112'
	"mscrM":                               0x2133,  // ℳ '\u2133'
	"mscrN":                               0x1d4a9, // 𝒩 '\U0001d4a9'
	"mscrO":                               0x1d4aa, // 𝒪 '\U0001d4aa'
	"mscrP":                               0x1d4ab, // 𝒫 '\U0001d4ab'
	"mscrQ":                               0x1d4ac, // 𝒬 '\U0001d4ac'
	"mscrR":                               0x211b,  // ℛ '\u211b'
	"mscrS":                               0x1d4ae, // 𝒮 '\U0001d4ae'
	"mscrT":                               0x1d4af, // 𝒯 '\U0001d4af'
	"mscrU":                               0x1d4b0, // 𝒰 '\U0001d4b0'
	"mscrV":                               0x1d4b1, // 𝒱 '\U0001d4b1'
	"mscrW":                               0x1d4b2, // 𝒲 '\U0001d4b2'
	"mscrX":                               0x1d4b3, // 𝒳 '\U0001d4b3'
	"mscrY":                               0x1d4b4, // 𝒴 '\U0001d4b4'
	"mscrZ":                               0x1d4b5, // 𝒵 '\U0001d4b5'
	"mscra":                               0x1d4b6, // 𝒶 '\U0001d4b6'
	"mscrb":                               0x1d4b7, // 𝒷 '\U0001d4b7'
	"mscrc":                               0x1d4b8, // 𝒸 '\U0001d4b8'
	"mscrd":                               0x1d4b9, // 𝒹 '\U0001d4b9'
	"mscre":                               0x212f,  // ℯ '\u212f'
	"mscrf":                               0x1d4bb, // 𝒻 '\U0001d4bb'
	"mscrg":                               0x210a,  // ℊ '\u210a'
	"mscrh":                               0x1d4bd, // 𝒽 '\U0001d4bd'
	"mscri":                               0x1d4be, // 𝒾 '\U0001d4be'
	"mscrj":                               0x1d4bf, // 𝒿 '\U0001d4bf'
	"mscrk":                               0x1d4c0, // 𝓀 '\U0001d4c0'
	"mscrl":                               0x1d4c1, // 𝓁 '\U0001d4c1'
	"mscrm":                               0x1d4c2, // 𝓂 '\U0001d4c2'
	"mscrn":                               0x1d4c3, // 𝓃 '\U0001d4c3'
	"mscro":                               0x2134,  // ℴ '\u2134'
	"mscrp":                               0x1d4c5, // 𝓅 '\U0001d4c5'
	"mscrq":                               0x1d4c6, // 𝓆 '\U0001d4c6'
	"mscrr":                               0x1d4c7, // 𝓇 '\U0001d4c7'
	"mscrs":                               0x1d4c8, // 𝓈 '\U0001d4c8'
	"mscrt":                               0x1d4c9, // 𝓉 '\U0001d4c9'
	"mscru":                               0x1d4ca, // 𝓊 '\U0001d4ca'
	"mscrv":                               0x1d4cb, // 𝓋 '\U0001d4cb'
	"mscrw":                               0x1d4cc, // 𝓌 '\U0001d4cc'
	"mscrx":                               0x1d4cd, // 𝓍 '\U0001d4cd'
	"mscry":                               0x1d4ce, // 𝓎 '\U0001d4ce'
	"mscrz":                               0x1d4cf, // 𝓏 '\U0001d4cf'
	"mssquare":                            0x33b3,  // ㎳ '\u33b3'
	"msuperior":                           0xf6ef,  //  '\uf6ef'
	"mttA":                                0x1d670, // 𝙰 '\U0001d670'
	"mttB":                                0x1d671, // 𝙱 '\U0001d671'
	"mttC":                                0x1d672, // 𝙲 '\U0001d672'
	"mttD":                                0x1d673, // 𝙳 '\U0001d673'
	"mttE":                                0x1d674, // 𝙴 '\U0001d674'
	"mttF":                                0x1d675, // 𝙵 '\U0001d675'
	"mttG":                                0x1d676, // 𝙶 '\U0001d676'
	"mttH":                                0x1d677, // 𝙷 '\U0001d677'
	"mttI":                                0x1d678, // 𝙸 '\U0001d678'
	"mttJ":                                0x1d679, // 𝙹 '\U0001d679'
	"mttK":                                0x1d67a, // 𝙺 '\U0001d67a'
	"mttL":                                0x1d67b, // 𝙻 '\U0001d67b'
	"mttM":                                0x1d67c, // 𝙼 '\U0001d67c'
	"mttN":                                0x1d67d, // 𝙽 '\U0001d67d'
	"mttO":                                0x1d67e, // 𝙾 '\U0001d67e'
	"mttP":                                0x1d67f, // 𝙿 '\U0001d67f'
	"mttQ":                                0x1d680, // 𝚀 '\U0001d680'
	"mttR":                                0x1d681, // 𝚁 '\U0001d681'
	"mttS":                                0x1d682, // 𝚂 '\U0001d682'
	"mttT":                                0x1d683, // 𝚃 '\U0001d683'
	"mttU":                                0x1d684, // 𝚄 '\U0001d684'
	"mttV":                                0x1d685, // 𝚅 '\U0001d685'
	"mttW":                                0x1d686, // 𝚆 '\U0001d686'
	"mttX":                                0x1d687, // 𝚇 '\U0001d687'
	"mttY":                                0x1d688, // 𝚈 '\U0001d688'
	"mttZ":                                0x1d689, // 𝚉 '\U0001d689'
	"mtta":                                0x1d68a, // 𝚊 '\U0001d68a'
	"mttb":                                0x1d68b, // 𝚋 '\U0001d68b'
	"mttc":                                0x1d68c, // 𝚌 '\U0001d68c'
	"mttd":                                0x1d68d, // 𝚍 '\U0001d68d'
	"mtte":                                0x1d68e, // 𝚎 '\U0001d68e'
	"mtteight":                            0x1d7fe, // 𝟾 '\U0001d7fe'
	"mttf":                                0x1d68f, // 𝚏 '\U0001d68f'
	"mttfive":                             0x1d7fb, // 𝟻 '\U0001d7fb'
	"mttfour":                             0x1d7fa, // 𝟺 '\U0001d7fa'
	"mttg":                                0x1d690, // 𝚐 '\U0001d690'
	"mtth":                                0x1d691, // 𝚑 '\U0001d691'
	"mtti":                                0x1d692, // 𝚒 '\U0001d692'
	"mttj":                                0x1d693, // 𝚓 '\U0001d693'
	"mttk":                                0x1d694, // 𝚔 '\U0001d694'
	"mttl":                                0x1d695, // 𝚕 '\U0001d695'
	"mttm":                                0x1d696, // 𝚖 '\U0001d696'
	"mttn":                                0x1d697, // 𝚗 '\U0001d697'
	"mttnine":                             0x1d7ff, // 𝟿 '\U0001d7ff'
	"mtto":                                0x1d698, // 𝚘 '\U0001d698'
	"mttone":                              0x1d7f7, // 𝟷 '\U0001d7f7'
	"mttp":                                0x1d699, // 𝚙 '\U0001d699'
	"mttq":                                0x1d69a, // 𝚚 '\U0001d69a'
	"mttr":                                0x1d69b, // 𝚛 '\U0001d69b'
	"mtts":                                0x1d69c, // 𝚜 '\U0001d69c'
	"mttseven":                            0x1d7fd, // 𝟽 '\U0001d7fd'
	"mttsix":                              0x1d7fc, // 𝟼 '\U0001d7fc'
	"mttt":                                0x1d69d, // 𝚝 '\U0001d69d'
	"mttthree":                            0x1d7f9, // 𝟹 '\U0001d7f9'
	"mtttwo":                              0x1d7f8, // 𝟸 '\U0001d7f8'
	"mttu":                                0x1d69e, // 𝚞 '\U0001d69e'
	"mttv":                                0x1d69f, // 𝚟 '\U0001d69f'
	"mttw":                                0x1d6a0, // 𝚠 '\U0001d6a0'
	"mttx":                                0x1d6a1, // 𝚡 '\U0001d6a1'
	"mtty":                                0x1d6a2, // 𝚢 '\U0001d6a2'
	"mttz":                                0x1d6a3, // 𝚣 '\U0001d6a3'
	"mttzero":                             0x1d7f6, // 𝟶 '\U0001d7f6'
	"mturned":                             0x026f,  // ɯ '\u026f'
	"mu":                                  0x00b5,  // µ '\u00b5'
	"muasquare":                           0x3382,  // ㎂ '\u3382'
	"muchgreater":                         0x226b,  // ≫ '\u226b'
	"muchless":                            0x226a,  // ≪ '\u226a'
	"mufsquare":                           0x338c,  // ㎌ '\u338c'
	"mugreek":                             0x03bc,  // μ '\u03bc'
	"mugsquare":                           0x338d,  // ㎍ '\u338d'
	"muhiragana":                          0x3080,  // む '\u3080'
	"mukatakana":                          0x30e0,  // ム '\u30e0'
	"mukatakanahalfwidth":                 0xff91,  // ﾑ '\uff91'
	"mulsquare":                           0x3395,  // ㎕ '\u3395'
	"multicloseleft":                      0x22c9,  // ⋉ '\u22c9'
	"multicloseright":                     0x22ca,  // ⋊ '\u22ca'
	"multimap":                            0x22b8,  // ⊸ '\u22b8'
	"multimapinv":                         0x27dc,  // ⟜ '\u27dc'
	"multiopenleft":                       0x22cb,  // ⋋ '\u22cb'
	"multiopenright":                      0x22cc,  // ⋌ '\u22cc'
	"multiply":                            0x00d7,  // × '\u00d7'
	"mumsquare":                           0x339b,  // ㎛ '\u339b'
	"munahlefthebrew":                     0x05a3,  // ֣ '\u05a3'
	"musicalnote":                         0x266a,  // ♪ '\u266a'
	"musicflatsign":                       0x266d,  // ♭ '\u266d'
	"musicsharpsign":                      0x266f,  // ♯ '\u266f'
	"mussquare":                           0x33b2,  // ㎲ '\u33b2'
	"muvsquare":                           0x33b6,  // ㎶ '\u33b6'
	"muwsquare":                           0x33bc,  // ㎼ '\u33bc'
	"mvmegasquare":                        0x33b9,  // ㎹ '\u33b9'
	"mvsquare":                            0x33b7,  // ㎷ '\u33b7'
	"mwmegasquare":                        0x33bf,  // ㎿ '\u33bf'
	"mwsquare":                            0x33bd,  // ㎽ '\u33bd'
	"n":                                   0x006e,  // n 'n'
	"nVleftarrow":                         0x21fa,  // ⇺ '\u21fa'
	"nVleftarrowtail":                     0x2b3a,  // ⬺ '\u2b3a'
	"nVleftrightarrow":                    0x21fc,  // ⇼ '\u21fc'
	"nVrightarrow":                        0x21fb,  // ⇻ '\u21fb'
	"nVrightarrowtail":                    0x2915,  // ⤕ '\u2915'
	"nVtwoheadleftarrow":                  0x2b35,  // ⬵ '\u2b35'
	"nVtwoheadleftarrowtail":              0x2b3d,  // ⬽ '\u2b3d'
	"nVtwoheadrightarrow":                 0x2901,  // ⤁ '\u2901'
	"nVtwoheadrightarrowtail":             0x2918,  // ⤘ '\u2918'
	"nabengali":                           0x09a8,  // ন '\u09a8'
	"nacute":                              0x0144,  // ń '\u0144'
	"nadeva":                              0x0928,  // न '\u0928'
	"nagujarati":                          0x0aa8,  // ન '\u0aa8'
	"nagurmukhi":                          0x0a28,  // ਨ '\u0a28'
	"nahiragana":                          0x306a,  // な '\u306a'
	"naira":                               0x20a6,  // ₦ '\u20a6'
	"nakatakana":                          0x30ca,  // ナ '\u30ca'
	"nakatakanahalfwidth":                 0xff85,  // ﾅ '\uff85'
	"nand":                                0x22bc,  // ⊼ '\u22bc'
	"napprox":                             0x2249,  // ≉ '\u2249'
	"nasquare":                            0x3381,  // ㎁ '\u3381'
	"nasymp":                              0x226d,  // ≭ '\u226d'
	"natural":                             0x266e,  // ♮ '\u266e'
	"nbhyphen":                            0x2011,  // ‑ '\u2011'
	"nbopomofo":                           0x310b,  // ㄋ '\u310b'
	"ncaron":                              0x0148,  // ň '\u0148'
	"ncedilla":                            0x0146,  // ņ '\u0146'
	"ncedilla1":                           0xf81d,  //  '\uf81d'
	"ncircle":                             0x24dd,  // ⓝ '\u24dd'
	"ncircumflexbelow":                    0x1e4b,  // ṋ '\u1e4b'
	"ndotaccent":                          0x1e45,  // ṅ '\u1e45'
	"ndotbelow":                           0x1e47,  // ṇ '\u1e47'
	"nehiragana":                          0x306d,  // ね '\u306d'
	"nekatakana":                          0x30cd,  // ネ '\u30cd'
	"nekatakanahalfwidth":                 0xff88,  // ﾈ '\uff88'
	"neovnwarrow":                         0x2931,  // ⤱ '\u2931'
	"neovsearrow":                         0x292e,  // ⤮ '\u292e'
	"neswarrow":                           0x2922,  // ⤢ '\u2922'
	"neuter":                              0x26b2,  // ⚲ '\u26b2'
	"nfsquare":                            0x338b,  // ㎋ '\u338b'
	"ngabengali":                          0x0999,  // ঙ '\u0999'
	"ngadeva":                             0x0919,  // ङ '\u0919'
	"ngagujarati":                         0x0a99,  // ઙ '\u0a99'
	"ngagurmukhi":                         0x0a19,  // ਙ '\u0a19'
	"ngonguthai":                          0x0e07,  // ง '\u0e07'
	"ngtrsim":                             0x2275,  // ≵ '\u2275'
	"nhVvert":                             0x2af5,  // ⫵ '\u2af5'
	"nhiragana":                           0x3093,  // ん '\u3093'
	"nhookleft":                           0x0272,  // ɲ '\u0272'
	"nhookretroflex":                      0x0273,  // ɳ '\u0273'
	"nhpar":                               0x2af2,  // ⫲ '\u2af2'
	"nieunacirclekorean":                  0x326f,  // ㉯ '\u326f'
	"nieunaparenkorean":                   0x320f,  // ㈏ '\u320f'
	"nieuncieuckorean":                    0x3135,  // ㄵ '\u3135'
	"nieuncirclekorean":                   0x3261,  // ㉡ '\u3261'
	"nieunhieuhkorean":                    0x3136,  // ㄶ '\u3136'
	"nieunkorean":                         0x3134,  // ㄴ '\u3134'
	"nieunpansioskorean":                  0x3168,  // ㅨ '\u3168'
	"nieunparenkorean":                    0x3201,  // ㈁ '\u3201'
	"nieunsioskorean":                     0x3167,  // ㅧ '\u3167'
	"nieuntikeutkorean":                   0x3166,  // ㅦ '\u3166'
	"nihiragana":                          0x306b,  // に '\u306b'
	"nikatakana":                          0x30cb,  // ニ '\u30cb'
	"nikatakanahalfwidth":                 0xff86,  // ﾆ '\uff86'
	"nikhahitleftthai":                    0xf899,  //  '\uf899'
	"nikhahitthai":                        0x0e4d,  // ํ '\u0e4d'
	"nine":                                0x0039,  // 9 '9'
	"ninebengali":                         0x09ef,  // ৯ '\u09ef'
	"ninedeva":                            0x096f,  // ९ '\u096f'
	"ninegujarati":                        0x0aef,  // ૯ '\u0aef'
	"ninegurmukhi":                        0x0a6f,  // ੯ '\u0a6f'
	"ninehackarabic":                      0x0669,  // ٩ '\u0669'
	"ninehangzhou":                        0x3029,  // 〩 '\u3029'
	"nineideographicparen":                0x3228,  // ㈨ '\u3228'
	"nineinferior":                        0x2089,  // ₉ '\u2089'
	"ninemonospace":                       0xff19,  // ９ '\uff19'
	"nineoldstyle":                        0xf739,  //  '\uf739'
	"nineparen":                           0x247c,  // ⑼ '\u247c'
	"nineperiod":                          0x2490,  // ⒐ '\u2490'
	"ninepersian":                         0x06f9,  // ۹ '\u06f9'
	"nineroman":                           0x2178,  // ⅸ '\u2178'
	"ninesuperior":                        0x2079,  // ⁹ '\u2079'
	"nineteencircle":                      0x2472,  // ⑲ '\u2472'
	"nineteenparen":                       0x2486,  // ⒆ '\u2486'
	"nineteenperiod":                      0x249a,  // ⒚ '\u249a'
	"ninethai":                            0x0e59,  // ๙ '\u0e59'
	"niobar":                              0x22fe,  // ⋾ '\u22fe'
	"nis":                                 0x22fc,  // ⋼ '\u22fc'
	"nisd":                                0x22fa,  // ⋺ '\u22fa'
	"nj":                                  0x01cc,  // ǌ '\u01cc'
	"nkatakana":                           0x30f3,  // ン '\u30f3'
	"nkatakanahalfwidth":                  0xff9d,  // ﾝ '\uff9d'
	"nlegrightlong":                       0x019e,  // ƞ '\u019e'
	"nlessgtr":                            0x2278,  // ≸ '\u2278'
	"nlesssim":                            0x2274,  // ≴ '\u2274'
	"nlinebelow":                          0x1e49,  // ṉ '\u1e49'
	"nmonospace":                          0xff4e,  // ｎ '\uff4e'
	"nmsquare":                            0x339a,  // ㎚ '\u339a'
	"nnabengali":                          0x09a3,  // ণ '\u09a3'
	"nnadeva":                             0x0923,  // ण '\u0923'
	"nnagujarati":                         0x0aa3,  // ણ '\u0aa3'
	"nnagurmukhi":                         0x0a23,  // ਣ '\u0a23'
	"nnnadeva":                            0x0929,  // ऩ '\u0929'
	"nohiragana":                          0x306e,  // の '\u306e'
	"nokatakana":                          0x30ce,  // ノ '\u30ce'
	"nokatakanahalfwidth":                 0xff89,  // ﾉ '\uff89'
	"nonbreakingspace":                    0x00a0,  //  '\u00a0'
	"nonenthai":                           0x0e13,  // ณ '\u0e13'
	"nonuthai":                            0x0e19,  // น '\u0e19'
	"noonarabic":                          0x0646,  // ن '\u0646'
	"noonfinalarabic":                     0xfee6,  // ﻦ '\ufee6'
	"noonghunnafinalarabic":               0xfb9f,  // ﮟ '\ufb9f'
	"noonhehinitialarabic":                0xfee7,  // ﻧ '\ufee7'
	"noonisolated":                        0xfee5,  // ﻥ '\ufee5'
	"noonjeeminitialarabic":               0xfcd2,  // ﳒ '\ufcd2'
	"noonjeemisolatedarabic":              0xfc4b,  // ﱋ '\ufc4b'
	"noonmedialarabic":                    0xfee8,  // ﻨ '\ufee8'
	"noonmeeminitialarabic":               0xfcd5,  // ﳕ '\ufcd5'
	"noonmeemisolatedarabic":              0xfc4e,  // ﱎ '\ufc4e'
	"noonnoonfinalarabic":                 0xfc8d,  // ﲍ '\ufc8d'
	"noonwithalefmaksurafinal":            0xfc8e,  // ﲎ '\ufc8e'
	"noonwithalefmaksuraisolated":         0xfc4f,  // ﱏ '\ufc4f'
	"noonwithhahinitial":                  0xfcd3,  // ﳓ '\ufcd3'
	"noonwithhehinitial":                  0xe815,  //  '\ue815'
	"noonwithkhahinitial":                 0xfcd4,  // ﳔ '\ufcd4'
	"noonwithyehfinal":                    0xfc8f,  // ﲏ '\ufc8f'
	"noonwithyehisolated":                 0xfc50,  // ﱐ '\ufc50'
	"noonwithzainfinal":                   0xfc70,  // ﱰ '\ufc70'
	"notapproxequal":                      0x2247,  // ≇ '\u2247'
	"notarrowboth":                        0x21ae,  // ↮ '\u21ae'
	"notarrowleft":                        0x219a,  // ↚ '\u219a'
	"notarrowright":                       0x219b,  // ↛ '\u219b'
	"notbar":                              0x2224,  // ∤ '\u2224'
	"notcontains":                         0x220c,  // ∌ '\u220c'
	"notdblarrowboth":                     0x21ce,  // ⇎ '\u21ce'
	"notelement":                          0x2209,  // ∉ '\u2209'
	"notequal":                            0x2260,  // ≠ '\u2260'
	"notexistential":                      0x2204,  // ∄ '\u2204'
	"notforces":                           0x22ae,  // ⊮ '\u22ae'
	"notforcesextra":                      0x22af,  // ⊯ '\u22af'
	"notgreater":                          0x226f,  // ≯ '\u226f'
	"notgreaternorequal":                  0x2271,  // ≱ '\u2271'
	"notgreaternorless":                   0x2279,  // ≹ '\u2279'
	"notgreaterorslnteql":                 0x2a7e,  // ⩾ '\u2a7e'
	"notidentical":                        0x2262,  // ≢ '\u2262'
	"notless":                             0x226e,  // ≮ '\u226e'
	"notlessnorequal":                     0x2270,  // ≰ '\u2270'
	"notparallel":                         0x2226,  // ∦ '\u2226'
	"notprecedes":                         0x2280,  // ⊀ '\u2280'
	"notsatisfies":                        0x22ad,  // ⊭ '\u22ad'
	"notsimilar":                          0x2241,  // ≁ '\u2241'
	"notsubset":                           0x2284,  // ⊄ '\u2284'
	"notsubseteql":                        0x2288,  // ⊈ '\u2288'
	"notsucceeds":                         0x2281,  // ⊁ '\u2281'
	"notsuperset":                         0x2285,  // ⊅ '\u2285'
	"notsuperseteql":                      0x2289,  // ⊉ '\u2289'
	"nottriangeqlleft":                    0x22ec,  // ⋬ '\u22ec'
	"nottriangeqlright":                   0x22ed,  // ⋭ '\u22ed'
	"nottriangleleft":                     0x22ea,  // ⋪ '\u22ea'
	"nottriangleright":                    0x22eb,  // ⋫ '\u22eb'
	"notturnstile":                        0x22ac,  // ⊬ '\u22ac'
	"nowarmenian":                         0x0576,  // ն '\u0576'
	"nparen":                              0x24a9,  // ⒩ '\u24a9'
	"npolint":                             0x2a14,  // ⨔ '\u2a14'
	"npreccurlyeq":                        0x22e0,  // ⋠ '\u22e0'
	"nsime":                               0x2244,  // ≄ '\u2244'
	"nsqsubseteq":                         0x22e2,  // ⋢ '\u22e2'
	"nsqsupseteq":                         0x22e3,  // ⋣ '\u22e3'
	"nssquare":                            0x33b1,  // ㎱ '\u33b1'
	"nsucccurlyeq":                        0x22e1,  // ⋡ '\u22e1'
	"nsuperior":                           0x207f,  // ⁿ '\u207f'
	"ntilde":                              0x00f1,  // ñ '\u00f1'
	"nu":                                  0x03bd,  // ν '\u03bd'
	"nuhiragana":                          0x306c,  // ぬ '\u306c'
	"nukatakana":                          0x30cc,  // ヌ '\u30cc'
	"nukatakanahalfwidth":                 0xff87,  // ﾇ '\uff87'
	"nuktabengali":                        0x09bc,  // ় '\u09bc'
	"nuktadeva":                           0x093c,  // ़ '\u093c'
	"nuktagujarati":                       0x0abc,  // ઼ '\u0abc'
	"nuktagurmukhi":                       0x0a3c,  // ਼ '\u0a3c'
	"numbersign":                          0x0023,  // # '#'
	"numbersignmonospace":                 0xff03,  // ＃ '\uff03'
	"numbersignsmall":                     0xfe5f,  // ﹟ '\ufe5f'
	"numeralsigngreek":                    0x0374,  // ʹ '\u0374'
	"numeralsignlowergreek":               0x0375,  // ͵ '\u0375'
	"numero":                              0x2116,  // № '\u2116'
	"nun":                                 0x05e0,  // נ '\u05e0'
	"nundagesh":                           0xfb40,  // נּ '\ufb40'
	"nvLeftarrow":                         0x2902,  // ⤂ '\u2902'
	"nvLeftrightarrow":                    0x2904,  // ⤄ '\u2904'
	"nvRightarrow":                        0x2903,  // ⤃ '\u2903'
	"nvinfty":                             0x29de,  // ⧞ '\u29de'
	"nvleftarrow":                         0x21f7,  // ⇷ '\u21f7'
	"nvleftarrowtail":                     0x2b39,  // ⬹ '\u2b39'
	"nvleftrightarrow":                    0x21f9,  // ⇹ '\u21f9'
	"nvrightarrow":                        0x21f8,  // ⇸ '\u21f8'
	"nvrightarrowtail":                    0x2914,  // ⤔ '\u2914'
	"nvsquare":                            0x33b5,  // ㎵ '\u33b5'
	"nvtwoheadleftarrow":                  0x2b34,  // ⬴ '\u2b34'
	"nvtwoheadleftarrowtail":              0x2b3c,  // ⬼ '\u2b3c'
	"nvtwoheadrightarrow":                 0x2900,  // ⤀ '\u2900'
	"nvtwoheadrightarrowtail":             0x2917,  // ⤗ '\u2917'
	"nwovnearrow":                         0x2932,  // ⤲ '\u2932'
	"nwsearrow":                           0x2921,  // ⤡ '\u2921'
	"nwsquare":                            0x33bb,  // ㎻ '\u33bb'
	"nyabengali":                          0x099e,  // ঞ '\u099e'
	"nyadeva":                             0x091e,  // ञ '\u091e'
	"nyagujarati":                         0x0a9e,  // ઞ '\u0a9e'
	"nyagurmukhi":                         0x0a1e,  // ਞ '\u0a1e'
	"o":                                   0x006f,  // o 'o'
	"oacute":                              0x00f3,  // ó '\u00f3'
	"oangthai":                            0x0e2d,  // อ '\u0e2d'
	"obar":                                0x233d,  // ⌽ '\u233d'
	"obarred":                             0x0275,  // ɵ '\u0275'
	"obarredcyrillic":                     0x04e9,  // ө '\u04e9'
	"obarreddieresiscyrillic":             0x04eb,  // ӫ '\u04eb'
	"obengali":                            0x0993,  // ও '\u0993'
	"obopomofo":                           0x311b,  // ㄛ '\u311b'
	"obot":                                0x29ba,  // ⦺ '\u29ba'
	"obrbrak":                             0x23e0,  // ⏠ '\u23e0'
	"obreve":                              0x014f,  // ŏ '\u014f'
	"obslash":                             0x29b8,  // ⦸ '\u29b8'
	"ocandradeva":                         0x0911,  // ऑ '\u0911'
	"ocandragujarati":                     0x0a91,  // ઑ '\u0a91'
	"ocandravowelsigndeva":                0x0949,  // ॉ '\u0949'
	"ocandravowelsigngujarati":            0x0ac9,  // ૉ '\u0ac9'
	"ocaron":                              0x01d2,  // ǒ '\u01d2'
	"ocircle":                             0x24de,  // ⓞ '\u24de'
	"ocircumflex":                         0x00f4,  // ô '\u00f4'
	"ocircumflexacute":                    0x1ed1,  // ố '\u1ed1'
	"ocircumflexdotbelow":                 0x1ed9,  // ộ '\u1ed9'
	"ocircumflexgrave":                    0x1ed3,  // ồ '\u1ed3'
	"ocircumflexhookabove":                0x1ed5,  // ổ '\u1ed5'
	"ocircumflextilde":                    0x1ed7,  // ỗ '\u1ed7'
	"ocyrillic":                           0x043e,  // о '\u043e'
	"odblgrave":                           0x020d,  // ȍ '\u020d'
	"odeva":                               0x0913,  // ओ '\u0913'
	"odieresis":                           0x00f6,  // ö '\u00f6'
	"odieresiscyrillic":                   0x04e7,  // ӧ '\u04e7'
	"odiv":                                0x2a38,  // ⨸ '\u2a38'
	"odotbelow":                           0x1ecd,  // ọ '\u1ecd'
	"odotslashdot":                        0x29bc,  // ⦼ '\u29bc'
	"oe":                                  0x0153,  // œ '\u0153'
	"oekorean":                            0x315a,  // ㅚ '\u315a'
	"ogonek":                              0x02db,  // ˛ '\u02db'
	"ogonekcmb":                           0x0328,  // ̨ '\u0328'
	"ograve":                              0x00f2,  // ò '\u00f2'
	"ogreaterthan":                        0x29c1,  // ⧁ '\u29c1'
	"ogujarati":                           0x0a93,  // ઓ '\u0a93'
	"oharmenian":                          0x0585,  // օ '\u0585'
	"ohiragana":                           0x304a,  // お '\u304a'
	"ohookabove":                          0x1ecf,  // ỏ '\u1ecf'
	"ohorn":                               0x01a1,  // ơ '\u01a1'
	"ohornacute":                          0x1edb,  // ớ '\u1edb'
	"ohorndotbelow":                       0x1ee3,  // ợ '\u1ee3'
	"ohorngrave":                          0x1edd,  // ờ '\u1edd'
	"ohornhookabove":                      0x1edf,  // ở '\u1edf'
	"ohorntilde":                          0x1ee1,  // ỡ '\u1ee1'
	"ohungarumlaut":                       0x0151,  // ő '\u0151'
	"oi":                                  0x01a3,  // ƣ '\u01a3'
	"oiiint":                              0x2230,  // ∰ '\u2230'
	"oiint":                               0x222f,  // ∯ '\u222f'
	"ointctrclockwise":                    0x2233,  // ∳ '\u2233'
	"oinvertedbreve":                      0x020f,  // ȏ '\u020f'
	"okatakana":                           0x30aa,  // オ '\u30aa'
	"okatakanahalfwidth":                  0xff75,  // ｵ '\uff75'
	"okorean":                             0x3157,  // ㅗ '\u3157'
	"olcross":                             0x29bb,  // ⦻ '\u29bb'
	"olehebrew":                           0x05ab,  // ֫ '\u05ab'
	"olessthan":                           0x29c0,  // ⧀ '\u29c0'
	"omacron":                             0x014d,  // ō '\u014d'
	"omacronacute":                        0x1e53,  // ṓ '\u1e53'
	"omacrongrave":                        0x1e51,  // ṑ '\u1e51'
	"omdeva":                              0x0950,  // ॐ '\u0950'
	"omega":                               0x03c9,  // ω '\u03c9'
	"omega1":                              0x03d6,  // ϖ '\u03d6'
	"omegacyrillic":                       0x0461,  // ѡ '\u0461'
	"omegalatinclosed":                    0x0277,  // ɷ '\u0277'
	"omegaroundcyrillic":                  0x047b,  // ѻ '\u047b'
	"omegatitlocyrillic":                  0x047d,  // ѽ '\u047d'
	"omegatonos":                          0x03ce,  // ώ '\u03ce'
	"omgujarati":                          0x0ad0,  // ૐ '\u0ad0'
	"omicron":                             0x03bf,  // ο '\u03bf'
	"omicrontonos":                        0x03cc,  // ό '\u03cc'
	"omonospace":                          0xff4f,  // ｏ '\uff4f'
	"one":                                 0x0031,  // 1 '1'
	"onebengali":                          0x09e7,  // ১ '\u09e7'
	"onedeva":                             0x0967,  // १ '\u0967'
	"onedotenleader":                      0x2024,  // ․ '\u2024'
	"oneeighth":                           0x215b,  // ⅛ '\u215b'
	"onefifth":                            0x2155,  // ⅕ '\u2155'
	"onefitted":                           0xf6dc,  //  '\uf6dc'
	"onegujarati":                         0x0ae7,  // ૧ '\u0ae7'
	"onegurmukhi":                         0x0a67,  // ੧ '\u0a67'
	"onehackarabic":                       0x0661,  // ١ '\u0661'
	"onehalf":                             0x00bd,  // ½ '\u00bd'
	"onehangzhou":                         0x3021,  // 〡 '\u3021'
	"oneideographicparen":                 0x3220,  // ㈠ '\u3220'
	"oneinferior":                         0x2081,  // ₁ '\u2081'
	"onemonospace":                        0xff11,  // １ '\uff11'
	"onenumeratorbengali":                 0x09f4,  // ৴ '\u09f4'
	"oneoldstyle":                         0xf731,  //  '\uf731'
	"oneparen":                            0x2474,  // ⑴ '\u2474'
	"oneperiod":                           0x2488,  // ⒈ '\u2488'
	"onepersian":                          0x06f1,  // ۱ '\u06f1'
	"onequarter":                          0x00bc,  // ¼ '\u00bc'
	"oneroman":                            0x2170,  // ⅰ '\u2170'
	"onesixth":                            0x2159,  // ⅙ '\u2159'
	"onesuperior":                         0x00b9,  // ¹ '\u00b9'
	"onethai":                             0x0e51,  // ๑ '\u0e51'
	"onethird":                            0x2153,  // ⅓ '\u2153'
	"oogonek":                             0x01eb,  // ǫ '\u01eb'
	"oogonekmacron":                       0x01ed,  // ǭ '\u01ed'
	"oogurmukhi":                          0x0a13,  // ਓ '\u0a13'
	"oomatragurmukhi":                     0x0a4b,  // ੋ '\u0a4b'
	"oopen":                               0x0254,  // ɔ '\u0254'
	"oparen":                              0x24aa,  // ⒪ '\u24aa'
	"operp":                               0x29b9,  // ⦹ '\u29b9'
	"opluslhrim":                          0x2a2d,  // ⨭ '\u2a2d'
	"oplusrhrim":                          0x2a2e,  // ⨮ '\u2a2e'
	"option":                              0x2325,  // ⌥ '\u2325'
	"ordfeminine":                         0x00aa,  // ª '\u00aa'
	"ordmasculine":                        0x00ba,  // º '\u00ba'
	"origof":                              0x22b6,  // ⊶ '\u22b6'
	"orthogonal":                          0x221f,  // ∟ '\u221f'
	"orunderscore":                        0x22bb,  // ⊻ '\u22bb'
	"oshortdeva":                          0x0912,  // ऒ '\u0912'
	"oshortvowelsigndeva":                 0x094a,  // ॊ '\u094a'
	"oslash":                              0x00f8,  // ø '\u00f8'
	"oslashacute":                         0x01ff,  // ǿ '\u01ff'
	"osmallhiragana":                      0x3049,  // ぉ '\u3049'
	"osmallkatakana":                      0x30a9,  // ォ '\u30a9'
	"osmallkatakanahalfwidth":             0xff6b,  // ｫ '\uff6b'
	"osuperior":                           0xf6f0,  //  '\uf6f0'
	"otcyrillic":                          0x047f,  // ѿ '\u047f'
	"otilde":                              0x00f5,  // õ '\u00f5'
	"otildeacute":                         0x1e4d,  // ṍ '\u1e4d'
	"otildedieresis":                      0x1e4f,  // ṏ '\u1e4f'
	"otimeshat":                           0x2a36,  // ⨶ '\u2a36'
	"otimeslhrim":                         0x2a34,  // ⨴ '\u2a34'
	"otimesrhrim":                         0x2a35,  // ⨵ '\u2a35'
	"oubopomofo":                          0x3121,  // ㄡ '\u3121'
	"ounce":                               0x2125,  // ℥ '\u2125'
	"overbrace":                           0x23de,  // ⏞ '\u23de'
	"overbracket":                         0x23b4,  // ⎴ '\u23b4'
	"overleftarrow":                       0x20d6,  // ⃖ '\u20d6'
	"overleftrightarrow":                  0x20e1,  // ⃡ '\u20e1'
	"overline":                            0x203e,  // ‾ '\u203e'
	"overlinecenterline":                  0xfe4a,  // ﹊ '\ufe4a'
	"overlinecmb":                         0x0305,  // ̅ '\u0305'
	"overlinedashed":                      0xfe49,  // ﹉ '\ufe49'
	"overlinedblwavy":                     0xfe4c,  // ﹌ '\ufe4c'
	"overlinewavy":                        0xfe4b,  // ﹋ '\ufe4b'
	"overparen":                           0x23dc,  // ⏜ '\u23dc'
	"ovowelsignbengali":                   0x09cb,  // ো '\u09cb'
	"ovowelsigndeva":                      0x094b,  // ो '\u094b'
	"ovowelsigngujarati":                  0x0acb,  // ો '\u0acb'
	"p":                                   0x0070,  // p 'p'
	"paampssquare":                        0x3380,  // ㎀ '\u3380'
	"paasentosquare":                      0x332b,  // ㌫ '\u332b'
	"pabengali":                           0x09aa,  // প '\u09aa'
	"pacute":                              0x1e55,  // ṕ '\u1e55'
	"padeva":                              0x092a,  // प '\u092a'
	"pagedown":                            0x21df,  // ⇟ '\u21df'
	"pageup":                              0x21de,  // ⇞ '\u21de'
	"pagujarati":                          0x0aaa,  // પ '\u0aaa'
	"pagurmukhi":                          0x0a2a,  // ਪ '\u0a2a'
	"pahiragana":                          0x3071,  // ぱ '\u3071'
	"paiyannoithai":                       0x0e2f,  // ฯ '\u0e2f'
	"pakatakana":                          0x30d1,  // パ '\u30d1'
	"palatalizationcyrilliccmb":           0x0484,  // ҄ '\u0484'
	"palochkacyrillic":                    0x04c0,  // Ӏ '\u04c0'
	"pansioskorean":                       0x317f,  // ㅿ '\u317f'
	"paragraph":                           0x00b6,  // ¶ '\u00b6'
	"paragraphseparator":                  0x2029,  //  '\u2029'
	"parallel":                            0x2225,  // ∥ '\u2225'
	"parallelogram":                       0x25b1,  // ▱ '\u25b1'
	"parallelogramblack":                  0x25b0,  // ▰ '\u25b0'
	"parenleft":                           0x0028,  // ( '('
	"parenleftaltonearabic":               0xfd3e,  // ﴾ '\ufd3e'
	"parenleftbt":                         0xf8ed,  //  '\uf8ed'
	"parenleftex":                         0xf8ec,  //  '\uf8ec'
	"parenleftinferior":                   0x208d,  // ₍ '\u208d'
	"parenleftmonospace":                  0xff08,  // （ '\uff08'
	"parenleftsmall":                      0xfe59,  // ﹙ '\ufe59'
	"parenleftsuperior":                   0x207d,  // ⁽ '\u207d'
	"parenlefttp":                         0xf8eb,  //  '\uf8eb'
	"parenleftvertical":                   0xfe35,  // ︵ '\ufe35'
	"parenright":                          0x0029,  // ) ')'
	"parenrightaltonearabic":              0xfd3f,  // ﴿ '\ufd3f'
	"parenrightbt":                        0xf8f8,  //  '\uf8f8'
	"parenrightex":                        0xf8f7,  //  '\uf8f7'
	"parenrightinferior":                  0x208e,  // ₎ '\u208e'
	"parenrightmonospace":                 0xff09,  // ） '\uff09'
	"parenrightsmall":                     0xfe5a,  // ﹚ '\ufe5a'
	"parenrightsuperior":                  0x207e,  // ⁾ '\u207e'
	"parenrighttp":                        0xf8f6,  //  '\uf8f6'
	"parenrightvertical":                  0xfe36,  // ︶ '\ufe36'
	"parsim":                              0x2af3,  // ⫳ '\u2af3'
	"partialdiff":                         0x2202,  // ∂ '\u2202'
	"partialmeetcontraction":              0x2aa3,  // ⪣ '\u2aa3'
	"pashtahebrew":                        0x0599,  // ֙ '\u0599'
	"pasquare":                            0x33a9,  // ㎩ '\u33a9'
	"patah11":                             0x05b7,  // ַ '\u05b7'
	"pazerhebrew":                         0x05a1,  // ֡ '\u05a1'
	"pbopomofo":                           0x3106,  // ㄆ '\u3106'
	"pcircle":                             0x24df,  // ⓟ '\u24df'
	"pdotaccent":                          0x1e57,  // ṗ '\u1e57'
	"pecyrillic":                          0x043f,  // п '\u043f'
	"pedagesh":                            0xfb44,  // פּ '\ufb44'
	"peezisquare":                         0x333b,  // ㌻ '\u333b'
	"pefinaldageshhebrew":                 0xfb43,  // ףּ '\ufb43'
	"peharabic":                           0x067e,  // پ '\u067e'
	"peharmenian":                         0x057a,  // պ '\u057a'
	"pehfinalarabic":                      0xfb57,  // ﭗ '\ufb57'
	"pehinitialarabic":                    0xfb58,  // ﭘ '\ufb58'
	"pehiragana":                          0x307a,  // ぺ '\u307a'
	"pehisolated":                         0xfb56,  // ﭖ '\ufb56'
	"pehmedialarabic":                     0xfb59,  // ﭙ '\ufb59'
	"pehwithhehinitial":                   0xe813,  //  '\ue813'
	"pekatakana":                          0x30da,  // ペ '\u30da'
	"pemiddlehookcyrillic":                0x04a7,  // ҧ '\u04a7'
	"pentagon":                            0x2b20,  // ⬠ '\u2b20'
	"pentagonblack":                       0x2b1f,  // ⬟ '\u2b1f'
	"perafehebrew":                        0xfb4e,  // פֿ '\ufb4e'
	"percent":                             0x0025,  // % '%'
	"percentarabic":                       0x066a,  // ٪ '\u066a'
	"percentmonospace":                    0xff05,  // ％ '\uff05'
	"percentsmall":                        0xfe6a,  // ﹪ '\ufe6a'
	"period":                              0x002e,  // . '.'
	"periodarmenian":                      0x0589,  // ։ '\u0589'
	"periodcentered":                      0x00b7,  // · '\u00b7'
	"periodcentered.0":                    0x0097,  //  '\u0097'
	"periodhalfwidth":                     0xff61,  // ｡ '\uff61'
	"periodinferior":                      0xf6e7,  //  '\uf6e7'
	"periodmonospace":                     0xff0e,  // ． '\uff0e'
	"periodsmall":                         0xfe52,  // ﹒ '\ufe52'
	"periodsuperior":                      0xf6e8,  //  '\uf6e8'
	"perispomenigreekcmb":                 0x0342,  // ͂ '\u0342'
	"perp":                                0x27c2,  // ⟂ '\u27c2'
	"perpcorrespond":                      0x2a5e,  // ⩞ '\u2a5e'
	"perpendicular":                       0x22a5,  // ⊥ '\u22a5'
	"perps":                               0x2ae1,  // ⫡ '\u2ae1'
	"pertenthousand":                      0x2031,  // ‱ '\u2031'
	"perthousand":                         0x2030,  // ‰ '\u2030'
	"peseta":                              0x20a7,  // ₧ '\u20a7'
	"peso1":                               0xf81b,  //  '\uf81b'
	"pfsquare":                            0x338a,  // ㎊ '\u338a'
	"phabengali":                          0x09ab,  // ফ '\u09ab'
	"phadeva":                             0x092b,  // फ '\u092b'
	"phagujarati":                         0x0aab,  // ફ '\u0aab'
	"phagurmukhi":                         0x0a2b,  // ਫ '\u0a2b'
	"phi":                                 0x03c6,  // φ '\u03c6'
	"phi1":                                0x03d5,  // ϕ '\u03d5'
	"phieuphacirclekorean":                0x327a,  // ㉺ '\u327a'
	"phieuphaparenkorean":                 0x321a,  // ㈚ '\u321a'
	"phieuphcirclekorean":                 0x326c,  // ㉬ '\u326c'
	"phieuphkorean":                       0x314d,  // ㅍ '\u314d'
	"phieuphparenkorean":                  0x320c,  // ㈌ '\u320c'
	"philatin":                            0x0278,  // ɸ '\u0278'
	"phinthuthai":                         0x0e3a,  // ฺ '\u0e3a'
	"phook":                               0x01a5,  // ƥ '\u01a5'
	"phophanthai":                         0x0e1e,  // พ '\u0e1e'
	"phophungthai":                        0x0e1c,  // ผ '\u0e1c'
	"phosamphaothai":                      0x0e20,  // ภ '\u0e20'
	"pi":                                  0x03c0,  // π '\u03c0'
	"pieupacirclekorean":                  0x3273,  // ㉳ '\u3273'
	"pieupaparenkorean":                   0x3213,  // ㈓ '\u3213'
	"pieupcieuckorean":                    0x3176,  // ㅶ '\u3176'
	"pieupcirclekorean":                   0x3265,  // ㉥ '\u3265'
	"pieupkiyeokkorean":                   0x3172,  // ㅲ '\u3172'
	"pieupkorean":                         0x3142,  // ㅂ '\u3142'
	"pieupparenkorean":                    0x3205,  // ㈅ '\u3205'
	"pieupsioskiyeokkorean":               0x3174,  // ㅴ '\u3174'
	"pieupsioskorean":                     0x3144,  // ㅄ '\u3144'
	"pieupsiostikeutkorean":               0x3175,  // ㅵ '\u3175'
	"pieupthieuthkorean":                  0x3177,  // ㅷ '\u3177'
	"pieuptikeutkorean":                   0x3173,  // ㅳ '\u3173'
	"pihiragana":                          0x3074,  // ぴ '\u3074'
	"pikatakana":                          0x30d4,  // ピ '\u30d4'
	"piwrarmenian":                        0x0583,  // փ '\u0583'
	"planckover2pi":                       0x210f,  // ℏ '\u210f'
	"plus":                                0x002b,  // + '+'
	"plusbelowcmb":                        0x031f,  // ̟ '\u031f'
	"plusdot":                             0x2a25,  // ⨥ '\u2a25'
	"pluseqq":                             0x2a72,  // ⩲ '\u2a72'
	"plushat":                             0x2a23,  // ⨣ '\u2a23'
	"plusinferior":                        0x208a,  // ₊ '\u208a'
	"plusminus":                           0x00b1,  // ± '\u00b1'
	"plusmod":                             0x02d6,  // ˖ '\u02d6'
	"plusmonospace":                       0xff0b,  // ＋ '\uff0b'
	"plussim":                             0x2a26,  // ⨦ '\u2a26'
	"plussmall":                           0xfe62,  // ﹢ '\ufe62'
	"plussubtwo":                          0x2a27,  // ⨧ '\u2a27'
	"plussuperior":                        0x207a,  // ⁺ '\u207a'
	"plustrif":                            0x2a28,  // ⨨ '\u2a28'
	"pmonospace":                          0xff50,  // ｐ '\uff50'
	"pmsquare":                            0x33d8,  // ㏘ '\u33d8'
	"pohiragana":                          0x307d,  // ぽ '\u307d'
	"pointingindexdownwhite":              0x261f,  // ☟ '\u261f'
	"pointingindexleftwhite":              0x261c,  // ☜ '\u261c'
	"pointingindexupwhite":                0x261d,  // ☝ '\u261d'
	"pointint":                            0x2a15,  // ⨕ '\u2a15'
	"pokatakana":                          0x30dd,  // ポ '\u30dd'
	"poplathai":                           0x0e1b,  // ป '\u0e1b'
	"postalmark":                          0x3012,  // 〒 '\u3012'
	"postalmarkface":                      0x3020,  // 〠 '\u3020'
	"pparen":                              0x24ab,  // ⒫ '\u24ab'
	"precapprox":                          0x2ab7,  // ⪷ '\u2ab7'
	"precedenotdbleqv":                    0x2ab9,  // ⪹ '\u2ab9'
	"precedenotslnteql":                   0x2ab5,  // ⪵ '\u2ab5'
	"precedeornoteqvlnt":                  0x22e8,  // ⋨ '\u22e8'
	"precedes":                            0x227a,  // ≺ '\u227a'
	"precedesequal":                       0x2aaf,  // ⪯ '\u2aaf'
	"precedesorcurly":                     0x227c,  // ≼ '\u227c'
	"precedesorequal":                     0x227e,  // ≾ '\u227e'
	"preceqq":                             0x2ab3,  // ⪳ '\u2ab3'
	"precneq":                             0x2ab1,  // ⪱ '\u2ab1'
	"prescription":                        0x211e,  // ℞ '\u211e'
	"primedblmod":                         0x0243,  // Ƀ '\u0243'
	"primemod":                            0x02b9,  // ʹ '\u02b9'
	"primereversed":                       0x2035,  // ‵ '\u2035'
	"product":                             0x220f,  // ∏ '\u220f'
	"profsurf":                            0x2313,  // ⌓ '\u2313'
	"projective":                          0x2305,  // ⌅ '\u2305'
	"prolongedkana":                       0x30fc,  // ー '\u30fc'
	"propellor":                           0x2318,  // ⌘ '\u2318'
	"propersubset":                        0x2282,  // ⊂ '\u2282'
	"propersuperset":                      0x2283,  // ⊃ '\u2283'
	"proportion":                          0x2237,  // ∷ '\u2237'
	"proportional":                        0x221d,  // ∝ '\u221d'
	"prurel":                              0x22b0,  // ⊰ '\u22b0'
	"psi":                                 0x03c8,  // ψ '\u03c8'
	"psicyrillic":                         0x0471,  // ѱ '\u0471'
	"psilipneumatacyrilliccmb":            0x0486,  // ҆ '\u0486'
	"pssquare":                            0x33b0,  // ㎰ '\u33b0'
	"puhiragana":                          0x3077,  // ぷ '\u3077'
	"pukatakana":                          0x30d7,  // プ '\u30d7'
	"pullback":                            0x27d3,  // ⟓ '\u27d3'
	"punctuationspace":                    0x2008,  //  '\u2008'
	"pushout":                             0x27d4,  // ⟔ '\u27d4'
	"pvsquare":                            0x33b4,  // ㎴ '\u33b4'
	"pwsquare":                            0x33ba,  // ㎺ '\u33ba'
	"q":                                   0x0071,  // q 'q'
	"qadeva":                              0x0958,  // क़ '\u0958'
	"qadmahebrew":                         0x05a8,  // ֨ '\u05a8'
	"qaffinalarabic":                      0xfed6,  // ﻖ '\ufed6'
	"qafinitialarabic":                    0xfed7,  // ﻗ '\ufed7'
	"qafisolated":                         0xfed5,  // ﻕ '\ufed5'
	"qafmedialarabic":                     0xfed8,  // ﻘ '\ufed8'
	"qarneyparahebrew":                    0x059f,  // ֟ '\u059f'
	"qbopomofo":                           0x3111,  // ㄑ '\u3111'
	"qcircle":                             0x24e0,  // ⓠ '\u24e0'
	"qhook":                               0x02a0,  // ʠ '\u02a0'
	"qmonospace":                          0xff51,  // ｑ '\uff51'
	"qofdagesh":                           0xfb47,  // קּ '\ufb47'
	"qofqubutshebrew":                     0x05e7,  // ק '\u05e7'
	"qparen":                              0x24ac,  // ⒬ '\u24ac'
	"qprime":                              0x2057,  // ⁗ '\u2057'
	"quarternote":                         0x2669,  // ♩ '\u2669'
	"qubutswidehebrew":                    0x05bb,  // ֻ '\u05bb'
	"questeq":                             0x225f,  // ≟ '\u225f'
	"question":                            0x003f,  // ? '?'
	"questionarmenian":                    0x055e,  // ՞ '\u055e'
	"questiondown":                        0x00bf,  // ¿ '\u00bf'
	"questiondownsmall":                   0xf7bf,  //  '\uf7bf'
	"questiongreek":                       0x037e,  // ; '\u037e'
	"questionmonospace":                   0xff1f,  // ？ '\uff1f'
	"questionsmall":                       0xf73f,  //  '\uf73f'
	"quotedbl":                            0x0022,  // " '"'
	"quotedblbase":                        0x201e,  // „ '\u201e'
	"quotedblleft":                        0x201c,  // “ '\u201c'
	"quotedblmonospace":                   0xff02,  // ＂ '\uff02'
	"quotedblprime":                       0x301e,  // 〞 '\u301e'
	"quotedblprimereversed":               0x301d,  // 〝 '\u301d'
	"quotedblrev":                         0x201f,  // ‟ '\u201f'
	"quotedblright":                       0x201d,  // ” '\u201d'
	"quoteleft":                           0x2018,  // ‘ '\u2018'
	"quoteleftmod":                        0x0244,  // Ʉ '\u0244'
	"quotereversed":                       0x201b,  // ‛ '\u201b'
	"quoteright":                          0x2019,  // ’ '\u2019'
	"quoterightn":                         0x0149,  // ŉ '\u0149'
	"quotesinglbase":                      0x201a,  // ‚ '\u201a'
	"quotesingle":                         0x0027,  // \' '\''
	"quotesinglemonospace":                0xff07,  // ＇ '\uff07'
	"r":                                   0x0072,  // r 'r'
	"rAngle":                              0x27eb,  // ⟫ '\u27eb'
	"rBrace":                              0x2984,  // ⦄ '\u2984'
	"rParen":                              0x2986,  // ⦆ '\u2986'
	"raarmenian":                          0x057c,  // ռ '\u057c'
	"rabengali":                           0x09b0,  // র '\u09b0'
	"racute":                              0x0155,  // ŕ '\u0155'
	"radeva":                              0x0930,  // र '\u0930'
	"radical":                             0x221a,  // √ '\u221a'
	"radicalex":                           0xf8e5,  //  '\uf8e5'
	"radoverssquare":                      0x33ae,  // ㎮ '\u33ae'
	"radoverssquaredsquare":               0x33af,  // ㎯ '\u33af'
	"radsquare":                           0x33ad,  // ㎭ '\u33ad'
	"ragujarati":                          0x0ab0,  // ર '\u0ab0'
	"ragurmukhi":                          0x0a30,  // ਰ '\u0a30'
	"rahiragana":                          0x3089,  // ら '\u3089'
	"raised":                              0x024d,  // ɍ '\u024d'
	"rakatakana":                          0x30e9,  // ラ '\u30e9'
	"rakatakanahalfwidth":                 0xff97,  // ﾗ '\uff97'
	"ralowerdiagonalbengali":              0x09f1,  // ৱ '\u09f1'
	"ramiddlediagonalbengali":             0x09f0,  // ৰ '\u09f0'
	"ramshorn":                            0x0264,  // ɤ '\u0264'
	"rangledot":                           0x2992,  // ⦒ '\u2992'
	"rangledownzigzagarrow":               0x237c,  // ⍼ '\u237c'
	"ratio":                               0x2236,  // ∶ '\u2236'
	"rayaleflam":                          0xe816,  //  '\ue816'
	"rbag":                                0x27c6,  // ⟆ '\u27c6'
	"rblkbrbrak":                          0x2998,  // ⦘ '\u2998'
	"rbopomofo":                           0x3116,  // ㄖ '\u3116'
	"rbracelend":                          0x23ad,  // ⎭ '\u23ad'
	"rbracemid":                           0x23ac,  // ⎬ '\u23ac'
	"rbraceuend":                          0x23ab,  // ⎫ '\u23ab'
	"rbrackextender":                      0x23a5,  // ⎥ '\u23a5'
	"rbracklend":                          0x23a6,  // ⎦ '\u23a6'
	"rbracklrtick":                        0x298e,  // ⦎ '\u298e'
	"rbrackubar":                          0x298c,  // ⦌ '\u298c'
	"rbrackuend":                          0x23a4,  // ⎤ '\u23a4'
	"rbrackurtick":                        0x2990,  // ⦐ '\u2990'
	"rbrbrak":                             0x2773,  // ❳ '\u2773'
	"rcaron":                              0x0159,  // ř '\u0159'
	"rcedilla":                            0x0157,  // ŗ '\u0157'
	"rcedilla1":                           0xf81f,  //  '\uf81f'
	"rcircle":                             0x24e1,  // ⓡ '\u24e1'
	"rcircumflex":                         0xf832,  //  '\uf832'
	"rcurvyangle":                         0x29fd,  // ⧽ '\u29fd'
	"rdblgrave":                           0x0211,  // ȑ '\u0211'
	"rdiagovfdiag":                        0x292b,  // ⤫ '\u292b'
	"rdiagovsearrow":                      0x2930,  // ⤰ '\u2930'
	"rdotaccent":                          0x1e59,  // ṙ '\u1e59'
	"rdotbelow":                           0x1e5b,  // ṛ '\u1e5b'
	"rdotbelowmacron":                     0x1e5d,  // ṝ '\u1e5d'
	"recordright":                         0x2117,  // ℗ '\u2117'
	"referencemark":                       0x203b,  // ※ '\u203b'
	"reflexsubset":                        0x2286,  // ⊆ '\u2286'
	"reflexsuperset":                      0x2287,  // ⊇ '\u2287'
	"registered":                          0x00ae,  // ® '\u00ae'
	"registersans":                        0xf8e8,  //  '\uf8e8'
	"registerserif":                       0xf6da,  //  '\uf6da'
	"reharabic":                           0x0631,  // ر '\u0631'
	"reharmenian":                         0x0580,  // ր '\u0580'
	"rehfinalarabic":                      0xfeae,  // ﺮ '\ufeae'
	"rehiragana":                          0x308c,  // れ '\u308c'
	"rehisolated":                         0xfead,  // ﺭ '\ufead'
	"rekatakana":                          0x30ec,  // レ '\u30ec'
	"rekatakanahalfwidth":                 0xff9a,  // ﾚ '\uff9a'
	"reshdageshhebrew":                    0xfb48,  // רּ '\ufb48'
	"reshhiriq":                           0x05e8,  // ר '\u05e8'
	"response":                            0x211f,  // ℟ '\u211f'
	"revangle":                            0x29a3,  // ⦣ '\u29a3'
	"revangleubar":                        0x29a5,  // ⦥ '\u29a5'
	"revasymptequal":                      0x22cd,  // ⋍ '\u22cd'
	"revemptyset":                         0x29b0,  // ⦰ '\u29b0'
	"reversedtilde":                       0x223d,  // ∽ '\u223d'
	"reviamugrashhebrew":                  0x0597,  // ֗ '\u0597'
	"revlogicalnot":                       0x2310,  // ⌐ '\u2310'
	"revnmid":                             0x2aee,  // ⫮ '\u2aee'
	"rfbowtie":                            0x29d2,  // ⧒ '\u29d2'
	"rfishhook":                           0x027e,  // ɾ '\u027e'
	"rfishhookreversed":                   0x027f,  // ɿ '\u027f'
	"rftimes":                             0x29d5,  // ⧕ '\u29d5'
	"rhabengali":                          0x09dd,  // ঢ় '\u09dd'
	"rhadeva":                             0x095d,  // ढ़ '\u095d'
	"rho":                                 0x03c1,  // ρ '\u03c1'
	"rhook":                               0x027d,  // ɽ '\u027d'
	"rhookturned":                         0x027b,  // ɻ '\u027b'
	"rhookturnedsuperior":                 0x02b5,  // ʵ '\u02b5'
	"rhosymbolgreek":                      0x03f1,  // ϱ '\u03f1'
	"rhotichookmod":                       0x02de,  // ˞ '\u02de'
	"rieulacirclekorean":                  0x3271,  // ㉱ '\u3271'
	"rieulaparenkorean":                   0x3211,  // ㈑ '\u3211'
	"rieulcirclekorean":                   0x3263,  // ㉣ '\u3263'
	"rieulhieuhkorean":                    0x3140,  // ㅀ '\u3140'
	"rieulkiyeokkorean":                   0x313a,  // ㄺ '\u313a'
	"rieulkiyeoksioskorean":               0x3169,  // ㅩ '\u3169'
	"rieulkorean":                         0x3139,  // ㄹ '\u3139'
	"rieulmieumkorean":                    0x313b,  // ㄻ '\u313b'
	"rieulpansioskorean":                  0x316c,  // ㅬ '\u316c'
	"rieulparenkorean":                    0x3203,  // ㈃ '\u3203'
	"rieulphieuphkorean":                  0x313f,  // ㄿ '\u313f'
	"rieulpieupkorean":                    0x313c,  // ㄼ '\u313c'
	"rieulpieupsioskorean":                0x316b,  // ㅫ '\u316b'
	"rieulsioskorean":                     0x313d,  // ㄽ '\u313d'
	"rieulthieuthkorean":                  0x313e,  // ㄾ '\u313e'
	"rieultikeutkorean":                   0x316a,  // ㅪ '\u316a'
	"rieulyeorinhieuhkorean":              0x316d,  // ㅭ '\u316d'
	"rightanglemdot":                      0x299d,  // ⦝ '\u299d'
	"rightanglene":                        0x231d,  // ⌝ '\u231d'
	"rightanglenw":                        0x231c,  // ⌜ '\u231c'
	"rightanglese":                        0x231f,  // ⌟ '\u231f'
	"rightanglesqr":                       0x299c,  // ⦜ '\u299c'
	"rightanglesw":                        0x231e,  // ⌞ '\u231e'
	"rightarrowapprox":                    0x2975,  // ⥵ '\u2975'
	"rightarrowbackapprox":                0x2b48,  // ⭈ '\u2b48'
	"rightarrowbsimilar":                  0x2b4c,  // ⭌ '\u2b4c'
	"rightarrowdiamond":                   0x291e,  // ⤞ '\u291e'
	"rightarrowgtr":                       0x2b43,  // ⭃ '\u2b43'
	"rightarrowonoplus":                   0x27f4,  // ⟴ '\u27f4'
	"rightarrowplus":                      0x2945,  // ⥅ '\u2945'
	"rightarrowshortleftarrow":            0x2942,  // ⥂ '\u2942'
	"rightarrowsimilar":                   0x2974,  // ⥴ '\u2974'
	"rightarrowsupset":                    0x2b44,  // ⭄ '\u2b44'
	"rightarrowtriangle":                  0x21fe,  // ⇾ '\u21fe'
	"rightarrowx":                         0x2947,  // ⥇ '\u2947'
	"rightbkarrow":                        0x290d,  // ⤍ '\u290d'
	"rightcurvedarrow":                    0x2933,  // ⤳ '\u2933'
	"rightdbltail":                        0x291c,  // ⤜ '\u291c'
	"rightdotarrow":                       0x2911,  // ⤑ '\u2911'
	"rightdowncurvedarrow":                0x2937,  // ⤷ '\u2937'
	"rightfishtail":                       0x297d,  // ⥽ '\u297d'
	"rightharpoonaccent":                  0x20d1,  // ⃑ '\u20d1'
	"rightharpoondownbar":                 0x2957,  // ⥗ '\u2957'
	"rightharpoonsupdown":                 0x2964,  // ⥤ '\u2964'
	"rightharpoonupbar":                   0x2953,  // ⥓ '\u2953'
	"rightharpoonupdash":                  0x296c,  // ⥬ '\u296c'
	"rightimply":                          0x2970,  // ⥰ '\u2970'
	"rightleftharpoonsdown":               0x2969,  // ⥩ '\u2969'
	"rightleftharpoonsup":                 0x2968,  // ⥨ '\u2968'
	"rightmoon":                           0x263d,  // ☽ '\u263d'
	"rightouterjoin":                      0x27d6,  // ⟖ '\u27d6'
	"rightpentagon":                       0x2b54,  // ⭔ '\u2b54'
	"rightpentagonblack":                  0x2b53,  // ⭓ '\u2b53'
	"rightrightarrows":                    0x21c9,  // ⇉ '\u21c9'
	"righttackbelowcmb":                   0x0319,  // ̙ '\u0319'
	"righttail":                           0x291a,  // ⤚ '\u291a'
	"rightthreearrows":                    0x21f6,  // ⇶ '\u21f6'
	"righttriangle":                       0x22bf,  // ⊿ '\u22bf'
	"rightwavearrow":                      0x219d,  // ↝ '\u219d'
	"rihiragana":                          0x308a,  // り '\u308a'
	"rikatakana":                          0x30ea,  // リ '\u30ea'
	"rikatakanahalfwidth":                 0xff98,  // ﾘ '\uff98'
	"ring":                                0x02da,  // ˚ '\u02da'
	"ring1":                               0xf007,  //  '\uf007'
	"ringbelowcmb":                        0x0325,  // ̥ '\u0325'
	"ringcmb":                             0x030a,  // ̊ '\u030a'
	"ringfitted":                          0xd80d,  //  '\ufffd'
	"ringhalfleft":                        0x02bf,  // ʿ '\u02bf'
	"ringhalfleftarmenian":                0x0559,  // ՙ '\u0559'
	"ringhalfleftbelowcmb":                0x031c,  // ̜ '\u031c'
	"ringhalfleftcentered":                0x02d3,  // ˓ '\u02d3'
	"ringhalfright":                       0x02be,  // ʾ '\u02be'
	"ringhalfrightbelowcmb":               0x0339,  // ̹ '\u0339'
	"ringhalfrightcentered":               0x02d2,  // ˒ '\u02d2'
	"ringinequal":                         0x2256,  // ≖ '\u2256'
	"ringlefthalfsubnosp":                 0x028f,  // ʏ '\u028f'
	"ringlefthalfsuper":                   0x0248,  // Ɉ '\u0248'
	"ringplus":                            0x2a22,  // ⨢ '\u2a22'
	"ringrighthalfsubnosp":                0x02ac,  // ʬ '\u02ac'
	"ringrighthalfsuper":                  0x0247,  // ɇ '\u0247'
	"rinvertedbreve":                      0x0213,  // ȓ '\u0213'
	"rittorusquare":                       0x3351,  // ㍑ '\u3351'
	"rle":                                 0x202b,  //  '\u202b'
	"rlinebelow":                          0x1e5f,  // ṟ '\u1e5f'
	"rlongleg":                            0x027c,  // ɼ '\u027c'
	"rlonglegturned":                      0x027a,  // ɺ '\u027a'
	"rmonospace":                          0xff52,  // ｒ '\uff52'
	"rmoustache":                          0x23b1,  // ⎱ '\u23b1'
	"rohiragana":                          0x308d,  // ろ '\u308d'
	"rokatakana":                          0x30ed,  // ロ '\u30ed'
	"rokatakanahalfwidth":                 0xff9b,  // ﾛ '\uff9b'
	"roruathai":                           0x0e23,  // ร '\u0e23'
	"rparen":                              0x24ad,  // ⒭ '\u24ad'
	"rparenextender":                      0x239f,  // ⎟ '\u239f'
	"rparengtr":                           0x2994,  // ⦔ '\u2994'
	"rparenlend":                          0x23a0,  // ⎠ '\u23a0'
	"rparenuend":                          0x239e,  // ⎞ '\u239e'
	"rppolint":                            0x2a12,  // ⨒ '\u2a12'
	"rrabengali":                          0x09dc,  // ড় '\u09dc'
	"rradeva":                             0x0931,  // ऱ '\u0931'
	"rragurmukhi":                         0x0a5c,  // ੜ '\u0a5c'
	"rrangle":                             0x298a,  // ⦊ '\u298a'
	"rreharabic":                          0x0691,  // ڑ '\u0691'
	"rrehfinalarabic":                     0xfb8d,  // ﮍ '\ufb8d'
	"rrparenthesis":                       0x2988,  // ⦈ '\u2988'
	"rrvocalicbengali":                    0x09e0,  // ৠ '\u09e0'
	"rrvocalicdeva":                       0x0960,  // ॠ '\u0960'
	"rrvocalicgujarati":                   0x0ae0,  // ૠ '\u0ae0'
	"rrvocalicvowelsignbengali":           0x09c4,  // ৄ '\u09c4'
	"rrvocalicvowelsigndeva":              0x0944,  // ॄ '\u0944'
	"rrvocalicvowelsigngujarati":          0x0ac4,  // ૄ '\u0ac4'
	"rsolbar":                             0x29f7,  // ⧷ '\u29f7'
	"rsqhook":                             0x2ace,  // ⫎ '\u2ace'
	"rsub":                                0x2a65,  // ⩥ '\u2a65'
	"rsuper":                              0x023c,  // ȼ '\u023c'
	"rsuperior":                           0xf6f1,  //  '\uf6f1'
	"rtblock":                             0x2590,  // ▐ '\u2590'
	"rteighthblock":                       0x2595,  // ▕ '\u2595'
	"rtriltri":                            0x29ce,  // ⧎ '\u29ce'
	"rturned":                             0x0279,  // ɹ '\u0279'
	"rturnedsuperior":                     0x02b4,  // ʴ '\u02b4'
	"rturnrthooksuper":                    0x023e,  // Ⱦ '\u023e'
	"rturnsuper":                          0x023d,  // Ƚ '\u023d'
	"ruhiragana":                          0x308b,  // る '\u308b'
	"rukatakana":                          0x30eb,  // ル '\u30eb'
	"rukatakanahalfwidth":                 0xff99,  // ﾙ '\uff99'
	"ruledelayed":                         0x29f4,  // ⧴ '\u29f4'
	"rupee":                               0x20a8,  // ₨ '\u20a8'
	"rupeemarkbengali":                    0x09f2,  // ৲ '\u09f2'
	"rupeesignbengali":                    0x09f3,  // ৳ '\u09f3'
	"rupiah":                              0xf6dd,  //  '\uf6dd'
	"ruthai":                              0x0e24,  // ฤ '\u0e24'
	"rvboxline":                           0x23b9,  // ⎹ '\u23b9'
	"rvocalicbengali":                     0x098b,  // ঋ '\u098b'
	"rvocalicdeva":                        0x090b,  // ऋ '\u090b'
	"rvocalicgujarati":                    0x0a8b,  // ઋ '\u0a8b'
	"rvocalicvowelsignbengali":            0x09c3,  // ৃ '\u09c3'
	"rvocalicvowelsigndeva":               0x0943,  // ृ '\u0943'
	"rvocalicvowelsigngujarati":           0x0ac3,  // ૃ '\u0ac3'
	"rvzigzag":                            0x29d9,  // ⧙ '\u29d9'
	"s":                                   0x0073,  // s 's'
	"sabengali":                           0x09b8,  // স '\u09b8'
	"sacute":                              0x015b,  // ś '\u015b'
	"sacutedotaccent":                     0x1e65,  // ṥ '\u1e65'
	"sadeva":                              0x0938,  // स '\u0938'
	"sadfinalarabic":                      0xfeba,  // ﺺ '\ufeba'
	"sadinitialarabic":                    0xfebb,  // ﺻ '\ufebb'
	"sadisolated":                         0xfeb9,  // ﺹ '\ufeb9'
	"sadmedialarabic":                     0xfebc,  // ﺼ '\ufebc'
	"sagujarati":                          0x0ab8,  // સ '\u0ab8'
	"sagurmukhi":                          0x0a38,  // ਸ '\u0a38'
	"sahiragana":                          0x3055,  // さ '\u3055'
	"sakatakana":                          0x30b5,  // サ '\u30b5'
	"sakatakanahalfwidth":                 0xff7b,  // ｻ '\uff7b'
	"sallallahoualayhewasallamarabic":     0xfdfa,  // ﷺ '\ufdfa'
	"samekh":                              0x05e1,  // ס '\u05e1'
	"samekhdageshhebrew":                  0xfb41,  // סּ '\ufb41'
	"sansLmirrored":                       0x2143,  // ⅃ '\u2143'
	"sansLturned":                         0x2142,  // ⅂ '\u2142'
	"saraaathai":                          0x0e32,  // า '\u0e32'
	"saraaethai":                          0x0e41,  // แ '\u0e41'
	"saraaimaimalaithai":                  0x0e44,  // ไ '\u0e44'
	"saraaimaimuanthai":                   0x0e43,  // ใ '\u0e43'
	"saraamthai":                          0x0e33,  // ำ '\u0e33'
	"saraathai":                           0x0e30,  // ะ '\u0e30'
	"saraethai":                           0x0e40,  // เ '\u0e40'
	"saraiileftthai":                      0xf886,  //  '\uf886'
	"saraiithai":                          0x0e35,  // ี '\u0e35'
	"saraileftthai":                       0xf885,  //  '\uf885'
	"saraithai":                           0x0e34,  // ิ '\u0e34'
	"saraothai":                           0x0e42,  // โ '\u0e42'
	"saraueeleftthai":                     0xf888,  //  '\uf888'
	"saraueethai":                         0x0e37,  // ื '\u0e37'
	"saraueleftthai":                      0xf887,  //  '\uf887'
	"sarauethai":                          0x0e36,  // ึ '\u0e36'
	"sarauthai":                           0x0e38,  // ุ '\u0e38'
	"sarauuthai":                          0x0e39,  // ู '\u0e39'
	"satisfies":                           0x22a8,  // ⊨ '\u22a8'
	"sbopomofo":                           0x3119,  // ㄙ '\u3119'
	"scaron":                              0x0161,  // š '\u0161'
	"scarondotaccent":                     0x1e67,  // ṧ '\u1e67'
	"scedilla":                            0x015f,  // ş '\u015f'
	"scedilla1":                           0xf817,  //  '\uf817'
	"schwa":                               0x0259,  // ə '\u0259'
	"schwacyrillic":                       0x04d9,  // ә '\u04d9'
	"schwadieresiscyrillic":               0x04db,  // ӛ '\u04db'
	"schwahook":                           0x025a,  // ɚ '\u025a'
	"scircle":                             0x24e2,  // ⓢ '\u24e2'
	"scircumflex":                         0x015d,  // ŝ '\u015d'
	"scommaaccent":                        0x0219,  // ș '\u0219'
	"scpolint":                            0x2a13,  // ⨓ '\u2a13'
	"scruple":                             0x2108,  // ℈ '\u2108'
	"scurel":                              0x22b1,  // ⊱ '\u22b1'
	"sdotaccent":                          0x1e61,  // ṡ '\u1e61'
	"sdotbelow":                           0x1e63,  // ṣ '\u1e63'
	"sdotbelowdotaccent":                  0x1e69,  // ṩ '\u1e69'
	"seagullbelowcmb":                     0x033c,  // ̼ '\u033c'
	"seagullsubnosp":                      0x02af,  // ʯ '\u02af'
	"second":                              0x2033,  // ″ '\u2033'
	"secondtonechinese":                   0x02ca,  // ˊ '\u02ca'
	"section":                             0x00a7,  // § '\u00a7'
	"seenfinalarabic":                     0xfeb2,  // ﺲ '\ufeb2'
	"seeninitialarabic":                   0xfeb3,  // ﺳ '\ufeb3'
	"seenisolated":                        0xfeb1,  // ﺱ '\ufeb1'
	"seenmedialarabic":                    0xfeb4,  // ﺴ '\ufeb4'
	"seenwithmeeminitial":                 0xfcb0,  // ﲰ '\ufcb0'
	"segolhebrew":                         0x05b6,  // ֶ '\u05b6'
	"segoltahebrew":                       0x0592,  // ֒ '\u0592'
	"seharmenian":                         0x057d,  // ս '\u057d'
	"sehiragana":                          0x305b,  // せ '\u305b'
	"sekatakana":                          0x30bb,  // セ '\u30bb'
	"sekatakanahalfwidth":                 0xff7e,  // ｾ '\uff7e'
	"semicolon":                           0x003b,  // ; ';'
	"semicolonmonospace":                  0xff1b,  // ； '\uff1b'
	"semicolonsmall":                      0xfe54,  // ﹔ '\ufe54'
	"semivoicedmarkkana":                  0x309c,  // ゜ '\u309c'
	"semivoicedmarkkanahalfwidth":         0xff9f,  // ﾟ '\uff9f'
	"sentisquare":                         0x3322,  // ㌢ '\u3322'
	"sentosquare":                         0x3323,  // ㌣ '\u3323'
	"seovnearrow":                         0x292d,  // ⤭ '\u292d'
	"servicemark":                         0x2120,  // ℠ '\u2120'
	"setminus":                            0x29f5,  // ⧵ '\u29f5'
	"seven":                               0x0037,  // 7 '7'
	"sevenbengali":                        0x09ed,  // ৭ '\u09ed'
	"sevendeva":                           0x096d,  // ७ '\u096d'
	"seveneighths":                        0x215e,  // ⅞ '\u215e'
	"sevengujarati":                       0x0aed,  // ૭ '\u0aed'
	"sevengurmukhi":                       0x0a6d,  // ੭ '\u0a6d'
	"sevenhangzhou":                       0x3027,  // 〧 '\u3027'
	"sevenideographicparen":               0x3226,  // ㈦ '\u3226'
	"seveninferior":                       0x2087,  // ₇ '\u2087'
	"sevenmonospace":                      0xff17,  // ７ '\uff17'
	"sevenoldstyle":                       0xf737,  //  '\uf737'
	"sevenparen":                          0x247a,  // ⑺ '\u247a'
	"sevenperiod":                         0x248e,  // ⒎ '\u248e'
	"sevenpersian":                        0x06f7,  // ۷ '\u06f7'
	"sevenroman":                          0x2176,  // ⅶ '\u2176'
	"sevensuperior":                       0x2077,  // ⁷ '\u2077'
	"seventeencircle":                     0x2470,  // ⑰ '\u2470'
	"seventeenparen":                      0x2484,  // ⒄ '\u2484'
	"seventeenperiod":                     0x2498,  // ⒘ '\u2498'
	"seventhai":                           0x0e57,  // ๗ '\u0e57'
	"shaarmenian":                         0x0577,  // շ '\u0577'
	"shabengali":                          0x09b6,  // শ '\u09b6'
	"shaddaarabic":                        0x0651,  // ّ '\u0651'
	"shaddadammaarabic":                   0xfc61,  // ﱡ '\ufc61'
	"shaddadammatanarabic":                0xfc5e,  // ﱞ '\ufc5e'
	"shaddafathaarabic":                   0xfc60,  // ﱠ '\ufc60'
	"shaddahontatweel":                    0xfe7d,  // ﹽ '\ufe7d'
	"shaddaisolated":                      0xfe7c,  // ﹼ '\ufe7c'
	"shaddakasraarabic":                   0xfc62,  // ﱢ '\ufc62'
	"shaddakasratanarabic":                0xfc5f,  // ﱟ '\ufc5f'
	"shaddalow":                           0xe825,  //  '\ue825'
	"shaddawithdammaisolatedlow":          0xe829,  //  '\ue829'
	"shaddawithdammamedial":               0xfcf3,  // ﳳ '\ufcf3'
	"shaddawithdammatanisolatedlow":       0xe82b,  //  '\ue82b'
	"shaddawithfathalow":                  0xe828,  //  '\ue828'
	"shaddawithfathamedial":               0xfcf2,  // ﳲ '\ufcf2'
	"shaddawithfathatanisolated":          0xe818,  //  '\ue818'
	"shaddawithfathatanlow":               0xe82a,  //  '\ue82a'
	"shaddawithkasraisolatedlow":          0xe82c,  //  '\ue82c'
	"shaddawithkasramedial":               0xfcf4,  // ﳴ '\ufcf4'
	"shaddawithkasratanisolatedlow":       0xe82d,  //  '\ue82d'
	"shade":                               0x2592,  // ▒ '\u2592'
	"shade1":                              0xf822,  //  '\uf822'
	"shadelight":                          0x2591,  // ░ '\u2591'
	"shadeva":                             0x0936,  // श '\u0936'
	"shagujarati":                         0x0ab6,  // શ '\u0ab6'
	"shagurmukhi":                         0x0a36,  // ਸ਼ '\u0a36'
	"shalshelethebrew":                    0x0593,  // ֓ '\u0593'
	"shbopomofo":                          0x3115,  // ㄕ '\u3115'
	"sheenfinalarabic":                    0xfeb6,  // ﺶ '\ufeb6'
	"sheeninitialarabic":                  0xfeb7,  // ﺷ '\ufeb7'
	"sheenisolated":                       0xfeb5,  // ﺵ '\ufeb5'
	"sheenmedialarabic":                   0xfeb8,  // ﺸ '\ufeb8'
	"sheenwithmeeminitial":                0xfd30,  // ﴰ '\ufd30'
	"sheicoptic":                          0x03e3,  // ϣ '\u03e3'
	"shhacyrillic":                        0x04bb,  // һ '\u04bb'
	"shiftleft":                           0x21b0,  // ↰ '\u21b0'
	"shiftright":                          0x21b1,  // ↱ '\u21b1'
	"shimacoptic":                         0x03ed,  // ϭ '\u03ed'
	"shin":                                0x05e9,  // ש '\u05e9'
	"shindagesh":                          0xfb49,  // שּ '\ufb49'
	"shindageshshindot":                   0xfb2c,  // שּׁ '\ufb2c'
	"shindageshsindothebrew":              0xfb2d,  // שּׂ '\ufb2d'
	"shindothebrew":                       0x05c1,  // ׁ '\u05c1'
	"shinshindot":                         0xfb2a,  // שׁ '\ufb2a'
	"shook":                               0x0282,  // ʂ '\u0282'
	"shortdowntack":                       0x2adf,  // ⫟ '\u2adf'
	"shortlefttack":                       0x2ade,  // ⫞ '\u2ade'
	"shortrightarrowleftarrow":            0x2944,  // ⥄ '\u2944'
	"shortuptack":                         0x2ae0,  // ⫠ '\u2ae0'
	"shuffle":                             0x29e2,  // ⧢ '\u29e2'
	"sigma":                               0x03c3,  // σ '\u03c3'
	"sigma1":                              0x03c2,  // ς '\u03c2'
	"sigmalunatesymbolgreek":              0x03f2,  // ϲ '\u03f2'
	"sihiragana":                          0x3057,  // し '\u3057'
	"sikatakana":                          0x30b7,  // シ '\u30b7'
	"sikatakanahalfwidth":                 0xff7c,  // ｼ '\uff7c'
	"siluqlefthebrew":                     0x05bd,  // ֽ '\u05bd'
	"simgE":                               0x2aa0,  // ⪠ '\u2aa0'
	"simgtr":                              0x2a9e,  // ⪞ '\u2a9e'
	"similar":                             0x223c,  // ∼ '\u223c'
	"similarleftarrow":                    0x2b49,  // ⭉ '\u2b49'
	"similarrightarrow":                   0x2972,  // ⥲ '\u2972'
	"simlE":                               0x2a9f,  // ⪟ '\u2a9f'
	"simless":                             0x2a9d,  // ⪝ '\u2a9d'
	"simminussim":                         0x2a6c,  // ⩬ '\u2a6c'
	"simneqq":                             0x2246,  // ≆ '\u2246'
	"simplus":                             0x2a24,  // ⨤ '\u2a24'
	"simrdots":                            0x2a6b,  // ⩫ '\u2a6b'
	"sinewave":                            0x223f,  // ∿ '\u223f'
	"siosacirclekorean":                   0x3274,  // ㉴ '\u3274'
	"siosaparenkorean":                    0x3214,  // ㈔ '\u3214'
	"sioscieuckorean":                     0x317e,  // ㅾ '\u317e'
	"sioscirclekorean":                    0x3266,  // ㉦ '\u3266'
	"sioskiyeokkorean":                    0x317a,  // ㅺ '\u317a'
	"sioskorean":                          0x3145,  // ㅅ '\u3145'
	"siosnieunkorean":                     0x317b,  // ㅻ '\u317b'
	"siosparenkorean":                     0x3206,  // ㈆ '\u3206'
	"siospieupkorean":                     0x317d,  // ㅽ '\u317d'
	"siostikeutkorean":                    0x317c,  // ㅼ '\u317c'
	"six":                                 0x0036,  // 6 '6'
	"sixbengali":                          0x09ec,  // ৬ '\u09ec'
	"sixdeva":                             0x096c,  // ६ '\u096c'
	"sixgujarati":                         0x0aec,  // ૬ '\u0aec'
	"sixgurmukhi":                         0x0a6c,  // ੬ '\u0a6c'
	"sixhangzhou":                         0x3026,  // 〦 '\u3026'
	"sixideographicparen":                 0x3225,  // ㈥ '\u3225'
	"sixinferior":                         0x2086,  // ₆ '\u2086'
	"sixmonospace":                        0xff16,  // ６ '\uff16'
	"sixoldstyle":                         0xf736,  //  '\uf736'
	"sixparen":                            0x2479,  // ⑹ '\u2479'
	"sixperemspace":                       0x2006,  //  '\u2006'
	"sixperiod":                           0x248d,  // ⒍ '\u248d'
	"sixpersian":                          0x06f6,  // ۶ '\u06f6'
	"sixroman":                            0x2175,  // ⅵ '\u2175'
	"sixsuperior":                         0x2076,  // ⁶ '\u2076'
	"sixteencircle":                       0x246f,  // ⑯ '\u246f'
	"sixteencurrencydenominatorbengali":   0x09f9,  // ৹ '\u09f9'
	"sixteenparen":                        0x2483,  // ⒃ '\u2483'
	"sixteenperiod":                       0x2497,  // ⒗ '\u2497'
	"sixthai":                             0x0e56,  // ๖ '\u0e56'
	"slash":                               0x002f,  // / '/'
	"slashlongnosp":                       0x02ab,  // ʫ '\u02ab'
	"slashmonospace":                      0xff0f,  // ／ '\uff0f'
	"slashshortnosp":                      0x02aa,  // ʪ '\u02aa'
	"slongdotaccent":                      0x1e9b,  // ẛ '\u1e9b'
	"slurabove":                           0x2322,  // ⌢ '\u2322'
	"smallblacktriangleleft":              0x25c2,  // ◂ '\u25c2'
	"smallblacktriangleright":             0x25b8,  // ▸ '\u25b8'
	"smallhighmadda":                      0x06e4,  // ۤ '\u06e4'
	"smallin":                             0x220a,  // ∊ '\u220a'
	"smallni":                             0x220d,  // ∍ '\u220d'
	"smashtimes":                          0x2a33,  // ⨳ '\u2a33'
	"smblkdiamond":                        0x2b29,  // ⬩ '\u2b29'
	"smblklozenge":                        0x2b2a,  // ⬪ '\u2b2a'
	"smeparsl":                            0x29e4,  // ⧤ '\u29e4'
	"smile":                               0x2323,  // ⌣ '\u2323'
	"smileface":                           0x263a,  // ☺ '\u263a'
	"smonospace":                          0xff53,  // ｓ '\uff53'
	"smt":                                 0x2aaa,  // ⪪ '\u2aaa'
	"smte":                                0x2aac,  // ⪬ '\u2aac'
	"smwhitestar":                         0x2b52,  // ⭒ '\u2b52'
	"smwhtlozenge":                        0x2b2b,  // ⬫ '\u2b2b'
	"sofpasuqhebrew":                      0x05c3,  // ׃ '\u05c3'
	"softhyphen":                          0x00ad,  //  '\u00ad'
	"softsigncyrillic":                    0x044c,  // ь '\u044c'
	"sohiragana":                          0x305d,  // そ '\u305d'
	"sokatakana":                          0x30bd,  // ソ '\u30bd'
	"sokatakanahalfwidth":                 0xff7f,  // ｿ '\uff7f'
	"soliduslongoverlaycmb":               0x0338,  // ̸ '\u0338'
	"solidusshortoverlaycmb":              0x0337,  // ̷ '\u0337'
	"sorusithai":                          0x0e29,  // ษ '\u0e29'
	"sosalathai":                          0x0e28,  // ศ '\u0e28'
	"sosothai":                            0x0e0b,  // ซ '\u0e0b'
	"sosuathai":                           0x0e2a,  // ส '\u0e2a'
	"space":                               0x0020,  //   ' '
	"spade":                               0x2660,  // ♠ '\u2660'
	"spadesuitwhite":                      0x2664,  // ♤ '\u2664'
	"sparen":                              0x24ae,  // ⒮ '\u24ae'
	"sphericalangle":                      0x2222,  // ∢ '\u2222'
	"sphericalangleup":                    0x29a1,  // ⦡ '\u29a1'
	"sqint":                               0x2a16,  // ⨖ '\u2a16'
	"sqlozenge":                           0x2311,  // ⌑ '\u2311'
	"sqrtbottom":                          0x23b7,  // ⎷ '\u23b7'
	"sqsubsetneq":                         0x22e4,  // ⋤ '\u22e4'
	"sqsupsetneq":                         0x22e5,  // ⋥ '\u22e5'
	"squarebelowcmb":                      0x033b,  // ̻ '\u033b'
	"squarebotblack":                      0x2b13,  // ⬓ '\u2b13'
	"squarecc":                            0x33c4,  // ㏄ '\u33c4'
	"squarecm":                            0x339d,  // ㎝ '\u339d'
	"squarediagonalcrosshatchfill":        0x25a9,  // ▩ '\u25a9'
	"squaredot":                           0x22a1,  // ⊡ '\u22a1'
	"squarehorizontalfill":                0x25a4,  // ▤ '\u25a4'
	"squareimage":                         0x228f,  // ⊏ '\u228f'
	"squarekg":                            0x338f,  // ㎏ '\u338f'
	"squarekm":                            0x339e,  // ㎞ '\u339e'
	"squarekmcapital":                     0x33ce,  // ㏎ '\u33ce'
	"squareleftblack":                     0x25e7,  // ◧ '\u25e7'
	"squarellblack":                       0x2b15,  // ⬕ '\u2b15'
	"squarellquad":                        0x25f1,  // ◱ '\u25f1'
	"squareln":                            0x33d1,  // ㏑ '\u33d1'
	"squarelog":                           0x33d2,  // ㏒ '\u33d2'
	"squarelrblack":                       0x25ea,  // ◪ '\u25ea'
	"squarelrquad":                        0x25f2,  // ◲ '\u25f2'
	"squaremg":                            0x338e,  // ㎎ '\u338e'
	"squaremil":                           0x33d5,  // ㏕ '\u33d5'
	"squareminus":                         0x229f,  // ⊟ '\u229f'
	"squaremm":                            0x339c,  // ㎜ '\u339c'
	"squaremsquared":                      0x33a1,  // ㎡ '\u33a1'
	"squaremultiply":                      0x22a0,  // ⊠ '\u22a0'
	"squareoriginal":                      0x2290,  // ⊐ '\u2290'
	"squareorthogonalcrosshatchfill":      0x25a6,  // ▦ '\u25a6'
	"squareplus":                          0x229e,  // ⊞ '\u229e'
	"squarerightblack":                    0x25e8,  // ◨ '\u25e8'
	"squaresubnosp":                       0x02ae,  // ʮ '\u02ae'
	"squaretopblack":                      0x2b12,  // ⬒ '\u2b12'
	"squareulblack":                       0x25e9,  // ◩ '\u25e9'
	"squareulquad":                        0x25f0,  // ◰ '\u25f0'
	"squareupperlefttolowerrightfill":     0x25a7,  // ▧ '\u25a7'
	"squareupperrighttolowerleftfill":     0x25a8,  // ▨ '\u25a8'
	"squareurblack":                       0x2b14,  // ⬔ '\u2b14'
	"squareurquad":                        0x25f3,  // ◳ '\u25f3'
	"squareverticalfill":                  0x25a5,  // ▥ '\u25a5'
	"squarewhitewithsmallblack":           0x25a3,  // ▣ '\u25a3'
	"squiggleleftright":                   0x21ad,  // ↭ '\u21ad'
	"squiggleright":                       0x21dd,  // ⇝ '\u21dd'
	"squoval":                             0x25a2,  // ▢ '\u25a2'
	"srsquare":                            0x33db,  // ㏛ '\u33db'
	"ssabengali":                          0x09b7,  // ষ '\u09b7'
	"ssadeva":                             0x0937,  // ष '\u0937'
	"ssagujarati":                         0x0ab7,  // ષ '\u0ab7'
	"ssangcieuckorean":                    0x3149,  // ㅉ '\u3149'
	"ssanghieuhkorean":                    0x3185,  // ㆅ '\u3185'
	"ssangieungkorean":                    0x3180,  // ㆀ '\u3180'
	"ssangkiyeokkorean":                   0x3132,  // ㄲ '\u3132'
	"ssangnieunkorean":                    0x3165,  // ㅥ '\u3165'
	"ssangpieupkorean":                    0x3143,  // ㅃ '\u3143'
	"ssangsioskorean":                     0x3146,  // ㅆ '\u3146'
	"ssangtikeutkorean":                   0x3138,  // ㄸ '\u3138'
	"sslash":                              0x2afd,  // ⫽ '\u2afd'
	"ssuperior":                           0xf6f2,  //  '\uf6f2'
	"st":                                  0xfb06,  // ﬆ '\ufb06'
	"star":                                0x22c6,  // ⋆ '\u22c6'
	"stareq":                              0x225b,  // ≛ '\u225b'
	"sterling":                            0x00a3,  // £ '\u00a3'
	"sterlingmonospace":                   0xffe1,  // ￡ '\uffe1'
	"strns":                               0x23e4,  // ⏤ '\u23e4'
	"strokelongoverlaycmb":                0x0336,  // ̶ '\u0336'
	"strokeshortoverlaycmb":               0x0335,  // ̵ '\u0335'
	"subedot":                             0x2ac3,  // ⫃ '\u2ac3'
	"submult":                             0x2ac1,  // ⫁ '\u2ac1'
	"subrarr":                             0x2979,  // ⥹ '\u2979'
	"subsetapprox":                        0x2ac9,  // ⫉ '\u2ac9'
	"subsetcirc":                          0x27c3,  // ⟃ '\u27c3'
	"subsetdbl":                           0x22d0,  // ⋐ '\u22d0'
	"subsetdblequal":                      0x2ac5,  // ⫅ '\u2ac5'
	"subsetdot":                           0x2abd,  // ⪽ '\u2abd'
	"subsetnotequal":                      0x228a,  // ⊊ '\u228a'
	"subsetornotdbleql":                   0x2acb,  // ⫋ '\u2acb'
	"subsetplus":                          0x2abf,  // ⪿ '\u2abf'
	"subsetsqequal":                       0x2291,  // ⊑ '\u2291'
	"subsim":                              0x2ac7,  // ⫇ '\u2ac7'
	"subsub":                              0x2ad5,  // ⫕ '\u2ad5'
	"subsup":                              0x2ad3,  // ⫓ '\u2ad3'
	"succapprox":                          0x2ab8,  // ⪸ '\u2ab8'
	"succeeds":                            0x227b,  // ≻ '\u227b'
	"succeqq":                             0x2ab4,  // ⪴ '\u2ab4'
	"succneq":                             0x2ab2,  // ⪲ '\u2ab2'
	"suchthat":                            0x220b,  // ∋ '\u220b'
	"suhiragana":                          0x3059,  // す '\u3059'
	"sukatakana":                          0x30b9,  // ス '\u30b9'
	"sukatakanahalfwidth":                 0xff7d,  // ｽ '\uff7d'
	"sukunarabic":                         0x0652,  // ْ '\u0652'
	"sukunisolated":                       0xfe7e,  // ﹾ '\ufe7e'
	"sukunlow":                            0xe822,  //  '\ue822'
	"sukunmedial":                         0xfe7f,  // ﹿ '\ufe7f'
	"sukunonhamza":                        0xe834,  //  '\ue834'
	"sumbottom":                           0x23b3,  // ⎳ '\u23b3'
	"sumint":                              0x2a0b,  // ⨋ '\u2a0b'
	"summation":                           0x2211,  // ∑ '\u2211'
	"sumtop":                              0x23b2,  // ⎲ '\u23b2'
	"sun":                                 0x263c,  // ☼ '\u263c'
	"supdsub":                             0x2ad8,  // ⫘ '\u2ad8'
	"supedot":                             0x2ac4,  // ⫄ '\u2ac4'
	"superscriptalef":                     0x0670,  // ٰ '\u0670'
	"supersetdbl":                         0x22d1,  // ⋑ '\u22d1'
	"supersetdblequal":                    0x2ac6,  // ⫆ '\u2ac6'
	"supersetnotequal":                    0x228b,  // ⊋ '\u228b'
	"supersetornotdbleql":                 0x2acc,  // ⫌ '\u2acc'
	"supersetsqequal":                     0x2292,  // ⊒ '\u2292'
	"suphsol":                             0x27c9,  // ⟉ '\u27c9'
	"suphsub":                             0x2ad7,  // ⫗ '\u2ad7'
	"suplarr":                             0x297b,  // ⥻ '\u297b'
	"supmult":                             0x2ac2,  // ⫂ '\u2ac2'
	"supsetapprox":                        0x2aca,  // ⫊ '\u2aca'
	"supsetcirc":                          0x27c4,  // ⟄ '\u27c4'
	"supsetdot":                           0x2abe,  // ⪾ '\u2abe'
	"supsetplus":                          0x2ac0,  // ⫀ '\u2ac0'
	"supsim":                              0x2ac8,  // ⫈ '\u2ac8'
	"supsub":                              0x2ad4,  // ⫔ '\u2ad4'
	"supsup":                              0x2ad6,  // ⫖ '\u2ad6'
	"svsquare":                            0x33dc,  // ㏜ '\u33dc'
	"syouwaerasquare":                     0x337c,  // ㍼ '\u337c'
	"t":                                   0x0074,  // t 't'
	"tabengali":                           0x09a4,  // ত '\u09a4'
	"tackdown":                            0x22a4,  // ⊤ '\u22a4'
	"tackleft":                            0x22a3,  // ⊣ '\u22a3'
	"tadeva":                              0x0924,  // त '\u0924'
	"tagujarati":                          0x0aa4,  // ત '\u0aa4'
	"tagurmukhi":                          0x0a24,  // ਤ '\u0a24'
	"taharabic":                           0x0637,  // ط '\u0637'
	"tahfinalarabic":                      0xfec2,  // ﻂ '\ufec2'
	"tahinitialarabic":                    0xfec3,  // ﻃ '\ufec3'
	"tahiragana":                          0x305f,  // た '\u305f'
	"tahisolated":                         0xfec1,  // ﻁ '\ufec1'
	"tahmedialarabic":                     0xfec4,  // ﻄ '\ufec4'
	"taisyouerasquare":                    0x337d,  // ㍽ '\u337d'
	"takatakana":                          0x30bf,  // タ '\u30bf'
	"takatakanahalfwidth":                 0xff80,  // ﾀ '\uff80'
	"talloblong":                          0x2afe,  // ⫾ '\u2afe'
	"tatweelwithfathatanabove":            0xfe71,  // ﹱ '\ufe71'
	"tau":                                 0x03c4,  // τ '\u03c4'
	"tavdagesh":                           0xfb4a,  // תּ '\ufb4a'
	"tavhebrew":                           0x05ea,  // ת '\u05ea'
	"tbar":                                0x0167,  // ŧ '\u0167'
	"tbopomofo":                           0x310a,  // ㄊ '\u310a'
	"tcaron":                              0x0165,  // ť '\u0165'
	"tcaron1":                             0xf815,  //  '\uf815'
	"tccurl":                              0x02a8,  // ʨ '\u02a8'
	"tcedilla":                            0x0163,  // ţ '\u0163'
	"tcedilla1":                           0xf819,  //  '\uf819'
	"tcheharabic":                         0x0686,  // چ '\u0686'
	"tchehfinalarabic":                    0xfb7b,  // ﭻ '\ufb7b'
	"tchehinitialarabic":                  0xfb7c,  // ﭼ '\ufb7c'
	"tchehisolated":                       0xfb7a,  // ﭺ '\ufb7a'
	"tchehmedialarabic":                   0xfb7d,  // ﭽ '\ufb7d'
	"tcircle":                             0x24e3,  // ⓣ '\u24e3'
	"tcircumflexbelow":                    0x1e71,  // ṱ '\u1e71'
	"tdieresis":                           0x1e97,  // ẗ '\u1e97'
	"tdotaccent":                          0x1e6b,  // ṫ '\u1e6b'
	"tdotbelow":                           0x1e6d,  // ṭ '\u1e6d'
	"tedescendercyrillic":                 0x04ad,  // ҭ '\u04ad'
	"tehfinalarabic":                      0xfe96,  // ﺖ '\ufe96'
	"tehhahinitialarabic":                 0xfca2,  // ﲢ '\ufca2'
	"tehhahisolatedarabic":                0xfc0c,  // ﰌ '\ufc0c'
	"tehinitialarabic":                    0xfe97,  // ﺗ '\ufe97'
	"tehiragana":                          0x3066,  // て '\u3066'
	"tehisolated":                         0xfe95,  // ﺕ '\ufe95'
	"tehjeeminitialarabic":                0xfca1,  // ﲡ '\ufca1'
	"tehjeemisolatedarabic":               0xfc0b,  // ﰋ '\ufc0b'
	"tehmarbutaarabic":                    0x0629,  // ة '\u0629'
	"tehmarbutafinalarabic":               0xfe94,  // ﺔ '\ufe94'
	"tehmarbutaisolated":                  0xfe93,  // ﺓ '\ufe93'
	"tehmedialarabic":                     0xfe98,  // ﺘ '\ufe98'
	"tehmeeminitialarabic":                0xfca4,  // ﲤ '\ufca4'
	"tehmeemisolatedarabic":               0xfc0e,  // ﰎ '\ufc0e'
	"tehnoonfinalarabic":                  0xfc73,  // ﱳ '\ufc73'
	"tehwithalefmaksurafinal":             0xfc74,  // ﱴ '\ufc74'
	"tehwithhehinitial":                   0xe814,  //  '\ue814'
	"tehwithkhahinitial":                  0xfca3,  // ﲣ '\ufca3'
	"tehwithyehfinal":                     0xfc75,  // ﱵ '\ufc75'
	"tehwithyehisolated":                  0xfc10,  // ﰐ '\ufc10'
	"tekatakana":                          0x30c6,  // テ '\u30c6'
	"tekatakanahalfwidth":                 0xff83,  // ﾃ '\uff83'
	"telephone":                           0x2121,  // ℡ '\u2121'
	"telishagedolahebrew":                 0x05a0,  // ֠ '\u05a0'
	"telishaqetanahebrew":                 0x05a9,  // ֩ '\u05a9'
	"tenideographicparen":                 0x3229,  // ㈩ '\u3229'
	"tenparen":                            0x247d,  // ⑽ '\u247d'
	"tenperiod":                           0x2491,  // ⒑ '\u2491'
	"tenroman":                            0x2179,  // ⅹ '\u2179'
	"tesh":                                0x02a7,  // ʧ '\u02a7'
	"tetdagesh":                           0xfb38,  // טּ '\ufb38'
	"tethebrew":                           0x05d8,  // ט '\u05d8'
	"tetsecyrillic":                       0x04b5,  // ҵ '\u04b5'
	"tevirhebrew":                         0x059b,  // ֛ '\u059b'
	"thabengali":                          0x09a5,  // থ '\u09a5'
	"thadeva":                             0x0925,  // थ '\u0925'
	"thagujarati":                         0x0aa5,  // થ '\u0aa5'
	"thagurmukhi":                         0x0a25,  // ਥ '\u0a25'
	"thalarabic":                          0x0630,  // ذ '\u0630'
	"thalfinalarabic":                     0xfeac,  // ﺬ '\ufeac'
	"thalisolated":                        0xfeab,  // ﺫ '\ufeab'
	"thanthakhatlowleftthai":              0xf898,  //  '\uf898'
	"thanthakhatlowrightthai":             0xf897,  //  '\uf897'
	"thanthakhatthai":                     0x0e4c,  // ์ '\u0e4c'
	"thanthakhatupperleftthai":            0xf896,  //  '\uf896'
	"theharabic":                          0x062b,  // ث '\u062b'
	"thehfinalarabic":                     0xfe9a,  // ﺚ '\ufe9a'
	"thehinitialarabic":                   0xfe9b,  // ﺛ '\ufe9b'
	"thehisolated":                        0xfe99,  // ﺙ '\ufe99'
	"thehmedialarabic":                    0xfe9c,  // ﺜ '\ufe9c'
	"thehwithmeeminitial":                 0xfca6,  // ﲦ '\ufca6'
	"thehwithmeemisolated":                0xfc12,  // ﰒ '\ufc12'
	"therefore":                           0x2234,  // ∴ '\u2234'
	"thermod":                             0x29e7,  // ⧧ '\u29e7'
	"theta":                               0x03b8,  // θ '\u03b8'
	"theta1":                              0x03d1,  // ϑ '\u03d1'
	"thieuthacirclekorean":                0x3279,  // ㉹ '\u3279'
	"thieuthaparenkorean":                 0x3219,  // ㈙ '\u3219'
	"thieuthcirclekorean":                 0x326b,  // ㉫ '\u326b'
	"thieuthkorean":                       0x314c,  // ㅌ '\u314c'
	"thieuthparenkorean":                  0x320b,  // ㈋ '\u320b'
	"thinspace":                           0x2009,  //  '\u2009'
	"thirteencircle":                      0x246c,  // ⑬ '\u246c'
	"thirteenparen":                       0x2480,  // ⒀ '\u2480'
	"thirteenperiod":                      0x2494,  // ⒔ '\u2494'
	"thonangmonthothai":                   0x0e11,  // ฑ '\u0e11'
	"thook":                               0x01ad,  // ƭ '\u01ad'
	"thophuthaothai":                      0x0e12,  // ฒ '\u0e12'
	"thorn":                               0x00fe,  // þ '\u00fe'
	"thothahanthai":                       0x0e17,  // ท '\u0e17'
	"thothanthai":                         0x0e10,  // ฐ '\u0e10'
	"thothongthai":                        0x0e18,  // ธ '\u0e18'
	"thothungthai":                        0x0e16,  // ถ '\u0e16'
	"thousandcyrillic":                    0x0482,  // ҂ '\u0482'
	"thousandsseparatorarabic":            0x066c,  // ٬ '\u066c'
	"three":                               0x0033,  // 3 '3'
	"threebengali":                        0x09e9,  // ৩ '\u09e9'
	"threedangle":                         0x27c0,  // ⟀ '\u27c0'
	"threedeva":                           0x0969,  // ३ '\u0969'
	"threedotcolon":                       0x2af6,  // ⫶ '\u2af6'
	"threeeighths":                        0x215c,  // ⅜ '\u215c'
	"threefifths":                         0x2157,  // ⅗ '\u2157'
	"threegujarati":                       0x0ae9,  // ૩ '\u0ae9'
	"threegurmukhi":                       0x0a69,  // ੩ '\u0a69'
	"threehangzhou":                       0x3023,  // 〣 '\u3023'
	"threeideographicparen":               0x3222,  // ㈢ '\u3222'
	"threeinferior":                       0x2083,  // ₃ '\u2083'
	"threemonospace":                      0xff13,  // ３ '\uff13'
	"threenumeratorbengali":               0x09f6,  // ৶ '\u09f6'
	"threeoldstyle":                       0xf733,  //  '\uf733'
	"threeparen":                          0x2476,  // ⑶ '\u2476'
	"threeperemspace":                     0x2004,  //  '\u2004'
	"threeperiod":                         0x248a,  // ⒊ '\u248a'
	"threepersian":                        0x06f3,  // ۳ '\u06f3'
	"threequarters":                       0x00be,  // ¾ '\u00be'
	"threequartersemdash":                 0xf6de,  //  '\uf6de'
	"threeroman":                          0x2172,  // ⅲ '\u2172'
	"threesuperior":                       0x00b3,  // ³ '\u00b3'
	"threethai":                           0x0e53,  // ๓ '\u0e53'
	"threeunderdot":                       0x20e8,  // ⃨ '\u20e8'
	"thzsquare":                           0x3394,  // ㎔ '\u3394'
	"tieconcat":                           0x2040,  // ⁀ '\u2040'
	"tieinfty":                            0x29dd,  // ⧝ '\u29dd'
	"tihiragana":                          0x3061,  // ち '\u3061'
	"tikatakana":                          0x30c1,  // チ '\u30c1'
	"tikatakanahalfwidth":                 0xff81,  // ﾁ '\uff81'
	"tikeutacirclekorean":                 0x3270,  // ㉰ '\u3270'
	"tikeutaparenkorean":                  0x3210,  // ㈐ '\u3210'
	"tikeutcirclekorean":                  0x3262,  // ㉢ '\u3262'
	"tikeutkorean":                        0x3137,  // ㄷ '\u3137'
	"tikeutparenkorean":                   0x3202,  // ㈂ '\u3202'
	"tilde":                               0x02dc,  // ˜ '\u02dc'
	"tilde1":                              0xf004,  //  '\uf004'
	"tildebelowcmb":                       0x0330,  // ̰ '\u0330'
	"tildecmb":                            0x0303,  // ̃ '\u0303'
	"tildedoublecmb":                      0x0360,  // ͠ '\u0360'
	"tildenosp":                           0x0276,  // ɶ '\u0276'
	"tildeoverlaycmb":                     0x0334,  // ̴ '\u0334'
	"tildeverticalcmb":                    0x033e,  // ̾ '\u033e'
	"timesbar":                            0x2a31,  // ⨱ '\u2a31'
	"tipehahebrew":                        0x0596,  // ֖ '\u0596'
	"tippigurmukhi":                       0x0a70,  // ੰ '\u0a70'
	"titlocyrilliccmb":                    0x0483,  // ҃ '\u0483'
	"tiwnarmenian":                        0x057f,  // տ '\u057f'
	"tlinebelow":                          0x1e6f,  // ṯ '\u1e6f'
	"tminus":                              0x29ff,  // ⧿ '\u29ff'
	"tmonospace":                          0xff54,  // ｔ '\uff54'
	"toarmenian":                          0x0569,  // թ '\u0569'
	"toea":                                0x2928,  // ⤨ '\u2928'
	"tohiragana":                          0x3068,  // と '\u3068'
	"tokatakana":                          0x30c8,  // ト '\u30c8'
	"tokatakanahalfwidth":                 0xff84,  // ﾄ '\uff84'
	"tona":                                0x2927,  // ⤧ '\u2927'
	"tonebarextrahighmod":                 0x02e5,  // ˥ '\u02e5'
	"tonebarextralowmod":                  0x02e9,  // ˩ '\u02e9'
	"tonebarhighmod":                      0x02e6,  // ˦ '\u02e6'
	"tonebarlowmod":                       0x02e8,  // ˨ '\u02e8'
	"tonebarmidmod":                       0x02e7,  // ˧ '\u02e7'
	"tonefive":                            0x01bd,  // ƽ '\u01bd'
	"tonesix":                             0x0185,  // ƅ '\u0185'
	"tonetwo":                             0x01a8,  // ƨ '\u01a8'
	"tonos":                               0x0384,  // ΄ '\u0384'
	"tonsquare":                           0x3327,  // ㌧ '\u3327'
	"topatakthai":                         0x0e0f,  // ฏ '\u0e0f'
	"topbot":                              0x2336,  // ⌶ '\u2336'
	"topcir":                              0x2af1,  // ⫱ '\u2af1'
	"topfork":                             0x2ada,  // ⫚ '\u2ada'
	"topsemicircle":                       0x25e0,  // ◠ '\u25e0'
	"tortoiseshellbracketleft":            0x3014,  // 〔 '\u3014'
	"tortoiseshellbracketleftsmall":       0xfe5d,  // ﹝ '\ufe5d'
	"tortoiseshellbracketleftvertical":    0xfe39,  // ︹ '\ufe39'
	"tortoiseshellbracketright":           0x3015,  // 〕 '\u3015'
	"tortoiseshellbracketrightsmall":      0xfe5e,  // ﹞ '\ufe5e'
	"tortoiseshellbracketrightvertical":   0xfe3a,  // ︺ '\ufe3a'
	"tosa":                                0x2929,  // ⤩ '\u2929'
	"totaothai":                           0x0e15,  // ต '\u0e15'
	"towa":                                0x292a,  // ⤪ '\u292a'
	"tpalatalhook":                        0x01ab,  // ƫ '\u01ab'
	"tparen":                              0x24af,  // ⒯ '\u24af'
	"tplus":                               0x29fe,  // ⧾ '\u29fe'
	"trademark":                           0x2122,  // ™ '\u2122'
	"trademarksans":                       0xf8ea,  //  '\uf8ea'
	"trademarkserif":                      0xf6db,  //  '\uf6db'
	"trapezium":                           0x23e2,  // ⏢ '\u23e2'
	"tretroflexhook":                      0x0288,  // ʈ '\u0288'
	"trianglebullet":                      0x2023,  // ‣ '\u2023'
	"trianglecdot":                        0x25ec,  // ◬ '\u25ec'
	"triangleleftblack":                   0x25ed,  // ◭ '\u25ed'
	"triangleleftequal":                   0x22b4,  // ⊴ '\u22b4'
	"triangleminus":                       0x2a3a,  // ⨺ '\u2a3a'
	"triangleodot":                        0x29ca,  // ⧊ '\u29ca'
	"triangleplus":                        0x2a39,  // ⨹ '\u2a39'
	"trianglerightblack":                  0x25ee,  // ◮ '\u25ee'
	"trianglerightequal":                  0x22b5,  // ⊵ '\u22b5'
	"triangles":                           0x29cc,  // ⧌ '\u29cc'
	"triangleserifs":                      0x29cd,  // ⧍ '\u29cd'
	"triangletimes":                       0x2a3b,  // ⨻ '\u2a3b'
	"triangleubar":                        0x29cb,  // ⧋ '\u29cb'
	"tripleplus":                          0x29fb,  // ⧻ '\u29fb'
	"trprime":                             0x2034,  // ‴ '\u2034'
	"trslash":                             0x2afb,  // ⫻ '\u2afb'
	"ts":                                  0x02a6,  // ʦ '\u02a6'
	"tsadidagesh":                         0xfb46,  // צּ '\ufb46'
	"tsecyrillic":                         0x0446,  // ц '\u0446'
	"tsere12":                             0x05b5,  // ֵ '\u05b5'
	"tshecyrillic":                        0x045b,  // ћ '\u045b'
	"tsuperior":                           0xf6f3,  //  '\uf6f3'
	"ttabengali":                          0x099f,  // ট '\u099f'
	"ttadeva":                             0x091f,  // ट '\u091f'
	"ttagujarati":                         0x0a9f,  // ટ '\u0a9f'
	"ttagurmukhi":                         0x0a1f,  // ਟ '\u0a1f'
	"ttehfinalarabic":                     0xfb67,  // ﭧ '\ufb67'
	"ttehinitialarabic":                   0xfb68,  // ﭨ '\ufb68'
	"ttehmedialarabic":                    0xfb69,  // ﭩ '\ufb69'
	"tthabengali":                         0x09a0,  // ঠ '\u09a0'
	"tthadeva":                            0x0920,  // ठ '\u0920'
	"tthagujarati":                        0x0aa0,  // ઠ '\u0aa0'
	"tthagurmukhi":                        0x0a20,  // ਠ '\u0a20'
	"tturned":                             0x0287,  // ʇ '\u0287'
	"tuhiragana":                          0x3064,  // つ '\u3064'
	"tukatakana":                          0x30c4,  // ツ '\u30c4'
	"tukatakanahalfwidth":                 0xff82,  // ﾂ '\uff82'
	"turnangle":                           0x29a2,  // ⦢ '\u29a2'
	"turnediota":                          0x2129,  // ℩ '\u2129'
	"turnednot":                           0x2319,  // ⌙ '\u2319'
	"turnstileleft":                       0x22a2,  // ⊢ '\u22a2'
	"tusmallhiragana":                     0x3063,  // っ '\u3063'
	"tusmallkatakana":                     0x30c3,  // ッ '\u30c3'
	"tusmallkatakanahalfwidth":            0xff6f,  // ｯ '\uff6f'
	"twelvecircle":                        0x246b,  // ⑫ '\u246b'
	"twelveparen":                         0x247f,  // ⑿ '\u247f'
	"twelveperiod":                        0x2493,  // ⒓ '\u2493'
	"twelveroman":                         0x217b,  // ⅻ '\u217b'
	"twelveudash":                         0xd80c,  //  '\ufffd'
	"twentycircle":                        0x2473,  // ⑳ '\u2473'
	"twentyhangzhou":                      0x5344,  // 卄 '\u5344'
	"twentyparen":                         0x2487,  // ⒇ '\u2487'
	"twentyperiod":                        0x249b,  // ⒛ '\u249b'
	"two":                                 0x0032,  // 2 '2'
	"twoarabic":                           0x0662,  // ٢ '\u0662'
	"twobengali":                          0x09e8,  // ২ '\u09e8'
	"twocaps":                             0x2a4b,  // ⩋ '\u2a4b'
	"twocups":                             0x2a4a,  // ⩊ '\u2a4a'
	"twodeva":                             0x0968,  // २ '\u0968'
	"twodotleader":                        0x2025,  // ‥ '\u2025'
	"twodotleadervertical":                0xfe30,  // ︰ '\ufe30'
	"twofifths":                           0x2156,  // ⅖ '\u2156'
	"twogujarati":                         0x0ae8,  // ૨ '\u0ae8'
	"twogurmukhi":                         0x0a68,  // ੨ '\u0a68'
	"twohangzhou":                         0x3022,  // 〢 '\u3022'
	"twoheaddownarrow":                    0x21a1,  // ↡ '\u21a1'
	"twoheadleftarrowtail":                0x2b3b,  // ⬻ '\u2b3b'
	"twoheadleftdbkarrow":                 0x2b37,  // ⬷ '\u2b37'
	"twoheadmapsfrom":                     0x2b36,  // ⬶ '\u2b36'
	"twoheadmapsto":                       0x2905,  // ⤅ '\u2905'
	"twoheadrightarrowtail":               0x2916,  // ⤖ '\u2916'
	"twoheaduparrow":                      0x219f,  // ↟ '\u219f'
	"twoheaduparrowcircle":                0x2949,  // ⥉ '\u2949'
	"twoideographicparen":                 0x3221,  // ㈡ '\u3221'
	"twoinferior":                         0x2082,  // ₂ '\u2082'
	"twomonospace":                        0xff12,  // ２ '\uff12'
	"twonumeratorbengali":                 0x09f5,  // ৵ '\u09f5'
	"twooldstyle":                         0xf732,  //  '\uf732'
	"twoparen":                            0x2475,  // ⑵ '\u2475'
	"twoperiod":                           0x2489,  // ⒉ '\u2489'
	"twopersian":                          0x06f2,  // ۲ '\u06f2'
	"tworoman":                            0x2171,  // ⅱ '\u2171'
	"twostroke":                           0x01bb,  // ƻ '\u01bb'
	"twosuperior":                         0x00b2,  // ² '\u00b2'
	"twothai":                             0x0e52,  // ๒ '\u0e52'
	"twothirds":                           0x2154,  // ⅔ '\u2154'
	"typecolon":                           0x2982,  // ⦂ '\u2982'
	"u":                                   0x0075,  // u 'u'
	"u2643":                               0x2643,  // ♃ '\u2643'
	"uacute":                              0x00fa,  // ú '\u00fa'
	"ubar":                                0x0289,  // ʉ '\u0289'
	"ubengali":                            0x0989,  // উ '\u0989'
	"ubopomofo":                           0x3128,  // ㄨ '\u3128'
	"ubrbrak":                             0x23e1,  // ⏡ '\u23e1'
	"ubreve":                              0x016d,  // ŭ '\u016d'
	"ucaron":                              0x01d4,  // ǔ '\u01d4'
	"ucedilla":                            0xf834,  //  '\uf834'
	"ucircle":                             0x24e4,  // ⓤ '\u24e4'
	"ucircumflex":                         0x00fb,  // û '\u00fb'
	"ucircumflexbelow":                    0x1e77,  // ṷ '\u1e77'
	"udattadeva":                          0x0951,  // ॑ '\u0951'
	"udblacute":                           0x0171,  // ű '\u0171'
	"udblgrave":                           0x0215,  // ȕ '\u0215'
	"udeva":                               0x0909,  // उ '\u0909'
	"udieresis":                           0x00fc,  // ü '\u00fc'
	"udieresisacute":                      0x01d8,  // ǘ '\u01d8'
	"udieresisbelow":                      0x1e73,  // ṳ '\u1e73'
	"udieresiscaron":                      0x01da,  // ǚ '\u01da'
	"udieresiscyrillic":                   0x04f1,  // ӱ '\u04f1'
	"udieresisgrave":                      0x01dc,  // ǜ '\u01dc'
	"udieresismacron":                     0x01d6,  // ǖ '\u01d6'
	"udotbelow":                           0x1ee5,  // ụ '\u1ee5'
	"ugrave":                              0x00f9,  // ù '\u00f9'
	"ugujarati":                           0x0a89,  // ઉ '\u0a89'
	"ugurmukhi":                           0x0a09,  // ਉ '\u0a09'
	"uhiragana":                           0x3046,  // う '\u3046'
	"uhookabove":                          0x1ee7,  // ủ '\u1ee7'
	"uhorn":                               0x01b0,  // ư '\u01b0'
	"uhornacute":                          0x1ee9,  // ứ '\u1ee9'
	"uhorndotbelow":                       0x1ef1,  // ự '\u1ef1'
	"uhorngrave":                          0x1eeb,  // ừ '\u1eeb'
	"uhornhookabove":                      0x1eed,  // ử '\u1eed'
	"uhorntilde":                          0x1eef,  // ữ '\u1eef'
	"uhungarumlautcyrillic":               0x04f3,  // ӳ '\u04f3'
	"uinvertedbreve":                      0x0217,  // ȗ '\u0217'
	"ukatakana":                           0x30a6,  // ウ '\u30a6'
	"ukatakanahalfwidth":                  0xff73,  // ｳ '\uff73'
	"ukcyrillic":                          0x0479,  // ѹ '\u0479'
	"ukorean":                             0x315c,  // ㅜ '\u315c'
	"ularc":                               0x25dc,  // ◜ '\u25dc'
	"ultriangle":                          0x25f8,  // ◸ '\u25f8'
	"umacron":                             0x016b,  // ū '\u016b'
	"umacroncyrillic":                     0x04ef,  // ӯ '\u04ef'
	"umacrondieresis":                     0x1e7b,  // ṻ '\u1e7b'
	"umatragurmukhi":                      0x0a41,  // ੁ '\u0a41'
	"uminus":                              0x2a41,  // ⩁ '\u2a41'
	"umonospace":                          0xff55,  // ｕ '\uff55'
	"underbrace":                          0x23df,  // ⏟ '\u23df'
	"underbracket":                        0x23b5,  // ⎵ '\u23b5'
	"underleftarrow":                      0x20ee,  // ⃮ '\u20ee'
	"underleftharpoondown":                0x20ed,  // ⃭ '\u20ed'
	"underparen":                          0x23dd,  // ⏝ '\u23dd'
	"underrightarrow":                     0x20ef,  // ⃯ '\u20ef'
	"underrightharpoondown":               0x20ec,  // ⃬ '\u20ec'
	"underscore":                          0x005f,  // _ '_'
	"underscoredbl":                       0x2017,  // ‗ '\u2017'
	"underscoremonospace":                 0xff3f,  // ＿ '\uff3f'
	"underscorevertical":                  0xfe33,  // ︳ '\ufe33'
	"underscorewavy":                      0xfe4f,  // ﹏ '\ufe4f'
	"undertie":                            0x203f,  // ‿ '\u203f'
	"unicodecdots":                        0x22ef,  // ⋯ '\u22ef'
	"union":                               0x222a,  // ∪ '\u222a'
	"uniondbl":                            0x22d3,  // ⋓ '\u22d3'
	"unionmulti":                          0x228e,  // ⊎ '\u228e'
	"unionsq":                             0x2294,  // ⊔ '\u2294'
	"uniontext":                           0x22c3,  // ⋃ '\u22c3'
	"universal":                           0x2200,  // ∀ '\u2200'
	"uogonek":                             0x0173,  // ų '\u0173'
	"upand":                               0x214b,  // ⅋ '\u214b'
	"uparen":                              0x24b0,  // ⒰ '\u24b0'
	"uparrowbarred":                       0x2909,  // ⤉ '\u2909'
	"uparrowoncircle":                     0x29bd,  // ⦽ '\u29bd'
	"upblock":                             0x2580,  // ▀ '\u2580'
	"updigamma":                           0x03dd,  // ϝ '\u03dd'
	"updownharpoonleftleft":               0x2951,  // ⥑ '\u2951'
	"updownharpoonleftright":              0x294d,  // ⥍ '\u294d'
	"updownharpoonrightleft":              0x294c,  // ⥌ '\u294c'
	"updownharpoonrightright":             0x294f,  // ⥏ '\u294f'
	"updownharpoonsleftright":             0x296e,  // ⥮ '\u296e'
	"upeighthblock":                       0x2594,  // ▔ '\u2594'
	"upfishtail":                          0x297e,  // ⥾ '\u297e'
	"upharpoonleftbar":                    0x2960,  // ⥠ '\u2960'
	"upharpoonrightbar":                   0x295c,  // ⥜ '\u295c'
	"upharpoonsleftright":                 0x2963,  // ⥣ '\u2963'
	"upin":                                0x27d2,  // ⟒ '\u27d2'
	"upint":                               0x2a1b,  // ⨛ '\u2a1b'
	"upkoppa":                             0x03df,  // ϟ '\u03df'
	"upoldKoppa":                          0x03d8,  // Ϙ '\u03d8'
	"upoldkoppa":                          0x03d9,  // ϙ '\u03d9'
	"upperdothebrew":                      0x05c4,  // ׄ '\u05c4'
	"uprightcurvearrow":                   0x2934,  // ⤴ '\u2934'
	"upsampi":                             0x03e1,  // ϡ '\u03e1'
	"upsilon":                             0x03c5,  // υ '\u03c5'
	"upsilondiaeresistonos":               0x02f9,  // ˹ '\u02f9'
	"upsilondieresis":                     0x03cb,  // ϋ '\u03cb'
	"upsilondieresistonos":                0x03b0,  // ΰ '\u03b0'
	"upsilonlatin":                        0x028a,  // ʊ '\u028a'
	"upsilontonos":                        0x03cd,  // ύ '\u03cd'
	"upslope":                             0x29f8,  // ⧸ '\u29f8'
	"upstigma":                            0x03db,  // ϛ '\u03db'
	"uptackbelowcmb":                      0x031d,  // ̝ '\u031d'
	"uptackmod":                           0x02d4,  // ˔ '\u02d4'
	"upvarTheta":                          0x03f4,  // ϴ '\u03f4'
	"uragurmukhi":                         0x0a73,  // ੳ '\u0a73'
	"urarc":                               0x25dd,  // ◝ '\u25dd'
	"uring":                               0x016f,  // ů '\u016f'
	"urtriangle":                          0x25f9,  // ◹ '\u25f9'
	"usmallhiragana":                      0x3045,  // ぅ '\u3045'
	"usmallkatakana":                      0x30a5,  // ゥ '\u30a5'
	"usmallkatakanahalfwidth":             0xff69,  // ｩ '\uff69'
	"ustraightcyrillic":                   0x04af,  // ү '\u04af'
	"ustraightstrokecyrillic":             0x04b1,  // ұ '\u04b1'
	"utilde":                              0x0169,  // ũ '\u0169'
	"utildeacute":                         0x1e79,  // ṹ '\u1e79'
	"utildebelow":                         0x1e75,  // ṵ '\u1e75'
	"uubengali":                           0x098a,  // ঊ '\u098a'
	"uudeva":                              0x090a,  // ऊ '\u090a'
	"uugujarati":                          0x0a8a,  // ઊ '\u0a8a'
	"uugurmukhi":                          0x0a0a,  // ਊ '\u0a0a'
	"uumatragurmukhi":                     0x0a42,  // ੂ '\u0a42'
	"uuvowelsignbengali":                  0x09c2,  // ূ '\u09c2'
	"uuvowelsigndeva":                     0x0942,  // ू '\u0942'
	"uuvowelsigngujarati":                 0x0ac2,  // ૂ '\u0ac2'
	"uvowelsignbengali":                   0x09c1,  // ু '\u09c1'
	"uvowelsigndeva":                      0x0941,  // ु '\u0941'
	"uvowelsigngujarati":                  0x0ac1,  // ુ '\u0ac1'
	"v":                                   0x0076,  // v 'v'
	"vBar":                                0x2ae8,  // ⫨ '\u2ae8'
	"vBarv":                               0x2ae9,  // ⫩ '\u2ae9'
	"vDdash":                              0x2ae2,  // ⫢ '\u2ae2'
	"vadeva":                              0x0935,  // व '\u0935'
	"vagujarati":                          0x0ab5,  // વ '\u0ab5'
	"vagurmukhi":                          0x0a35,  // ਵ '\u0a35'
	"vakatakana":                          0x30f7,  // ヷ '\u30f7'
	"varVdash":                            0x2ae6,  // ⫦ '\u2ae6'
	"varcarriagereturn":                   0x23ce,  // ⏎ '\u23ce'
	"vardoublebarwedge":                   0x2306,  // ⌆ '\u2306'
	"varhexagon":                          0x2b21,  // ⬡ '\u2b21'
	"varhexagonblack":                     0x2b22,  // ⬢ '\u2b22'
	"varhexagonlrbonds":                   0x232c,  // ⌬ '\u232c'
	"varika":                              0xfb1e,  // ﬞ '\ufb1e'
	"varisinobar":                         0x22f6,  // ⋶ '\u22f6'
	"varisins":                            0x22f3,  // ⋳ '\u22f3'
	"varniobar":                           0x22fd,  // ⋽ '\u22fd'
	"varnis":                              0x22fb,  // ⋻ '\u22fb'
	"varointclockwise":                    0x2232,  // ∲ '\u2232'
	"vartriangleleft":                     0x22b2,  // ⊲ '\u22b2'
	"vartriangleright":                    0x22b3,  // ⊳ '\u22b3'
	"varveebar":                           0x2a61,  // ⩡ '\u2a61'
	"vav":                                 0x05d5,  // ו '\u05d5'
	"vavdageshhebrew":                     0xfb35,  // וּ '\ufb35'
	"vavholam":                            0xfb4b,  // וֹ '\ufb4b'
	"vbraceextender":                      0x23aa,  // ⎪ '\u23aa'
	"vbrtri":                              0x29d0,  // ⧐ '\u29d0'
	"vcircle":                             0x24e5,  // ⓥ '\u24e5'
	"vdotbelow":                           0x1e7f,  // ṿ '\u1e7f'
	"vectimes":                            0x2a2f,  // ⨯ '\u2a2f'
	"vector":                              0x20d7,  // ⃗ '\u20d7'
	"veedot":                              0x27c7,  // ⟇ '\u27c7'
	"veedoublebar":                        0x2a63,  // ⩣ '\u2a63'
	"veeeq":                               0x225a,  // ≚ '\u225a'
	"veemidvert":                          0x2a5b,  // ⩛ '\u2a5b'
	"veeodot":                             0x2a52,  // ⩒ '\u2a52'
	"veeonvee":                            0x2a56,  // ⩖ '\u2a56'
	"veeonwedge":                          0x2a59,  // ⩙ '\u2a59'
	"veharabic":                           0x06a4,  // ڤ '\u06a4'
	"vehfinalarabic":                      0xfb6b,  // ﭫ '\ufb6b'
	"vehinitialarabic":                    0xfb6c,  // ﭬ '\ufb6c'
	"vehisolated":                         0xfb6a,  // ﭪ '\ufb6a'
	"vehmedialarabic":                     0xfb6d,  // ﭭ '\ufb6d'
	"vekatakana":                          0x30f9,  // ヹ '\u30f9'
	"versicle":                            0x2123,  // ℣ '\u2123'
	"verticallineabovecmb":                0x030d,  // ̍ '\u030d'
	"verticallinebelowcmb":                0x0329,  // ̩ '\u0329'
	"verticallinelowmod":                  0x02cc,  // ˌ '\u02cc'
	"verticallinemod":                     0x02c8,  // ˈ '\u02c8'
	"vertoverlay":                         0x20d2,  // ⃒ '\u20d2'
	"vewarmenian":                         0x057e,  // վ '\u057e'
	"vhook":                               0x028b,  // ʋ '\u028b'
	"viewdata":                            0x2317,  // ⌗ '\u2317'
	"vikatakana":                          0x30f8,  // ヸ '\u30f8'
	"viramabengali":                       0x09cd,  // ্ '\u09cd'
	"viramadeva":                          0x094d,  // ् '\u094d'
	"viramagujarati":                      0x0acd,  // ્ '\u0acd'
	"visargabengali":                      0x0983,  // ঃ '\u0983'
	"visargadeva":                         0x0903,  // ः '\u0903'
	"visargagujarati":                     0x0a83,  // ઃ '\u0a83'
	"vlongdash":                           0x27dd,  // ⟝ '\u27dd'
	"vmonospace":                          0xff56,  // ｖ '\uff56'
	"voarmenian":                          0x0578,  // ո '\u0578'
	"voicediterationhiragana":             0x309e,  // ゞ '\u309e'
	"voicediterationkatakana":             0x30fe,  // ヾ '\u30fe'
	"voicedmarkkana":                      0x309b,  // ゛ '\u309b'
	"voicedmarkkanahalfwidth":             0xff9e,  // ﾞ '\uff9e'
	"vokatakana":                          0x30fa,  // ヺ '\u30fa'
	"vparen":                              0x24b1,  // ⒱ '\u24b1'
	"vrectangle":                          0x25af,  // ▯ '\u25af'
	"vrectangleblack":                     0x25ae,  // ▮ '\u25ae'
	"vscript":                             0x021b,  // ț '\u021b'
	"vtilde":                              0x1e7d,  // ṽ '\u1e7d'
	"vturn":                               0x021c,  // Ȝ '\u021c'
	"vturned":                             0x028c,  // ʌ '\u028c'
	"vuhiragana":                          0x3094,  // ゔ '\u3094'
	"vukatakana":                          0x30f4,  // ヴ '\u30f4'
	"vysmblksquare":                       0x2b1d,  // ⬝ '\u2b1d'
	"vysmwhtcircle":                       0x2218,  // ∘ '\u2218'
	"vysmwhtsquare":                       0x2b1e,  // ⬞ '\u2b1e'
	"vzigzag":                             0x299a,  // ⦚ '\u299a'
	"w":                                   0x0077,  // w 'w'
	"wacute":                              0x1e83,  // ẃ '\u1e83'
	"waekorean":                           0x3159,  // ㅙ '\u3159'
	"wahiragana":                          0x308f,  // わ '\u308f'
	"wakatakana":                          0x30ef,  // ワ '\u30ef'
	"wakatakanahalfwidth":                 0xff9c,  // ﾜ '\uff9c'
	"wakorean":                            0x3158,  // ㅘ '\u3158'
	"wasmallhiragana":                     0x308e,  // ゎ '\u308e'
	"wasmallkatakana":                     0x30ee,  // ヮ '\u30ee'
	"wattosquare":                         0x3357,  // ㍗ '\u3357'
	"wavedash":                            0x301c,  // 〜 '\u301c'
	"wavyunderscorevertical":              0xfe34,  // ︴ '\ufe34'
	"wawarabic":                           0x0648,  // و '\u0648'
	"wawfinalarabic":                      0xfeee,  // ﻮ '\ufeee'
	"wawhamzaabovefinalarabic":            0xfe86,  // ﺆ '\ufe86'
	"wawisolated":                         0xfeed,  // ﻭ '\ufeed'
	"wawwithhamzaaboveisolated":           0xfe85,  // ﺅ '\ufe85'
	"wbsquare":                            0x33dd,  // ㏝ '\u33dd'
	"wcircle":                             0x24e6,  // ⓦ '\u24e6'
	"wcircumflex":                         0x0175,  // ŵ '\u0175'
	"wdieresis":                           0x1e85,  // ẅ '\u1e85'
	"wdotaccent":                          0x1e87,  // ẇ '\u1e87'
	"wdotbelow":                           0x1e89,  // ẉ '\u1e89'
	"wedgebar":                            0x2a5f,  // ⩟ '\u2a5f'
	"wedgedot":                            0x27d1,  // ⟑ '\u27d1'
	"wedgedoublebar":                      0x2a60,  // ⩠ '\u2a60'
	"wedgemidvert":                        0x2a5a,  // ⩚ '\u2a5a'
	"wedgeodot":                           0x2a51,  // ⩑ '\u2a51'
	"wedgeonwedge":                        0x2a55,  // ⩕ '\u2a55'
	"wedgeq":                              0x2259,  // ≙ '\u2259'
	"wehiragana":                          0x3091,  // ゑ '\u3091'
	"weierstrass":                         0x2118,  // ℘ '\u2118'
	"wekatakana":                          0x30f1,  // ヱ '\u30f1'
	"wekorean":                            0x315e,  // ㅞ '\u315e'
	"weokorean":                           0x315d,  // ㅝ '\u315d'
	"wgrave":                              0x1e81,  // ẁ '\u1e81'
	"whitebullet":                         0x25e6,  // ◦ '\u25e6'
	"whitecircle":                         0x25cb,  // ○ '\u25cb'
	"whitecornerbracketleft":              0x300e,  // 『 '\u300e'
	"whitecornerbracketleftvertical":      0xfe43,  // ﹃ '\ufe43'
	"whitecornerbracketright":             0x300f,  // 』 '\u300f'
	"whitecornerbracketrightvertical":     0xfe44,  // ﹄ '\ufe44'
	"whitediamond":                        0x25c7,  // ◇ '\u25c7'
	"whitediamondcontainingblacksmalldiamond": 0x25c8, // ◈ '\u25c8'
	"whitedownpointingsmalltriangle":          0x25bf, // ▿ '\u25bf'
	"whitedownpointingtriangle":               0x25bd, // ▽ '\u25bd'
	"whiteinwhitetriangle":                    0x27c1, // ⟁ '\u27c1'
	"whiteleftpointingsmalltriangle":          0x25c3, // ◃ '\u25c3'
	"whiteleftpointingtriangle":               0x25c1, // ◁ '\u25c1'
	"whitelenticularbracketleft":              0x3016, // 〖 '\u3016'
	"whitelenticularbracketright":             0x3017, // 〗 '\u3017'
	"whitepointerleft":                        0x25c5, // ◅ '\u25c5'
	"whitepointerright":                       0x25bb, // ▻ '\u25bb'
	"whiterightpointingsmalltriangle":         0x25b9, // ▹ '\u25b9'
	"whiterightpointingtriangle":              0x25b7, // ▷ '\u25b7'
	"whitesmallsquare":                        0x25ab, // ▫ '\u25ab'
	"whitesquaretickleft":                     0x27e4, // ⟤ '\u27e4'
	"whitesquaretickright":                    0x27e5, // ⟥ '\u27e5'
	"whitestar":                               0x2606, // ☆ '\u2606'
	"whitetelephone":                          0x260f, // ☏ '\u260f'
	"whitetortoiseshellbracketleft":           0x3018, // 〘 '\u3018'
	"whitetortoiseshellbracketright":          0x3019, // 〙 '\u3019'
	"whiteuppointingsmalltriangle":            0x25b5, // ▵ '\u25b5'
	"whiteuppointingtriangle":                 0x25b3, // △ '\u25b3'
	"whthorzoval":                             0x2b2d, // ⬭ '\u2b2d'
	"whtvertoval":                             0x2b2f, // ⬯ '\u2b2f'
	"wideangledown":                           0x29a6, // ⦦ '\u29a6'
	"wideangleup":                             0x29a7, // ⦧ '\u29a7'
	"widebridgeabove":                         0x20e9, // ⃩ '\u20e9'
	"wihiragana":                              0x3090, // ゐ '\u3090'
	"wikatakana":                              0x30f0, // ヰ '\u30f0'
	"wikorean":                                0x315f, // ㅟ '\u315f'
	"wmonospace":                              0xff57, // ｗ '\uff57'
	"wohiragana":                              0x3092, // を '\u3092'
	"wokatakana":                              0x30f2, // ヲ '\u30f2'
	"wokatakanahalfwidth":                     0xff66, // ｦ '\uff66'
	"won":                                     0x20a9, // ₩ '\u20a9'
	"wonmonospace":                            0xffe6, // ￦ '\uffe6'
	"wowaenthai":                              0x0e27, // ว '\u0e27'
	"wparen":                                  0x24b2, // ⒲ '\u24b2'
	"wreathproduct":                           0x2240, // ≀ '\u2240'
	"wring":                                   0x1e98, // ẘ '\u1e98'
	"wsuper":                                  0x0240, // ɀ '\u0240'
	"wsuperior":                               0x02b7, // ʷ '\u02b7'
	"wturn":                                   0x021d, // ȝ '\u021d'
	"wturned":                                 0x028d, // ʍ '\u028d'
	"wynn":                                    0x01bf, // ƿ '\u01bf'
	"x":                                       0x0078, // x 'x'
	"xabovecmb":                               0x033d, // ̽ '\u033d'
	"xbopomofo":                               0x3112, // ㄒ '\u3112'
	"xcircle":                                 0x24e7, // ⓧ '\u24e7'
	"xdieresis":                               0x1e8d, // ẍ '\u1e8d'
	"xdotaccent":                              0x1e8b, // ẋ '\u1e8b'
	"xeharmenian":                             0x056d, // խ '\u056d'
	"xi":                                      0x03be, // ξ '\u03be'
	"xmonospace":                              0xff58, // ｘ '\uff58'
	"xparen":                                  0x24b3, // ⒳ '\u24b3'
	"xsuperior":                               0x02e3, // ˣ '\u02e3'
	"y":                                       0x0079, // y 'y'
	"yaadosquare":                             0x334e, // ㍎ '\u334e'
	"yabengali":                               0x09af, // য '\u09af'
	"yacute":                                  0x00fd, // ý '\u00fd'
	"yadeva":                                  0x092f, // य '\u092f'
	"yaekorean":                               0x3152, // ㅒ '\u3152'
	"yagujarati":                              0x0aaf, // ય '\u0aaf'
	"yagurmukhi":                              0x0a2f, // ਯ '\u0a2f'
	"yahiragana":                              0x3084, // や '\u3084'
	"yakatakana":                              0x30e4, // ヤ '\u30e4'
	"yakatakanahalfwidth":                     0xff94, // ﾔ '\uff94'
	"yakorean":                                0x3151, // ㅑ '\u3151'
	"yamakkanthai":                            0x0e4e, // ๎ '\u0e4e'
	"yasmallhiragana":                         0x3083, // ゃ '\u3083'
	"yasmallkatakana":                         0x30e3, // ャ '\u30e3'
	"yasmallkatakanahalfwidth":                0xff6c, // ｬ '\uff6c'
	"yatcyrillic":                             0x0463, // ѣ '\u0463'
	"ycircle":                                 0x24e8, // ⓨ '\u24e8'
	"ycircumflex":                             0x0177, // ŷ '\u0177'
	"ydieresis":                               0x00ff, // ÿ '\u00ff'
	"ydotaccent":                              0x1e8f, // ẏ '\u1e8f'
	"ydotbelow":                               0x1ef5, // ỵ '\u1ef5'
	"yeharabic":                               0x064a, // ي '\u064a'
	"yehbarreearabic":                         0x06d2, // ے '\u06d2'
	"yehbarreefinalarabic":                    0xfbaf, // ﮯ '\ufbaf'
	"yehfinalarabic":                          0xfef2, // ﻲ '\ufef2'
	"yehhamzaabovearabic":                     0x0626, // ئ '\u0626'
	"yehhamzaabovefinalarabic":                0xfe8a, // ﺊ '\ufe8a'
	"yehhamzaaboveinitialarabic":              0xfe8b, // ﺋ '\ufe8b'
	"yehhamzaabovemedialarabic":               0xfe8c, // ﺌ '\ufe8c'
	"yehinitialarabic":                        0xfef3, // ﻳ '\ufef3'
	"yehisolated":                             0xfef1, // ﻱ '\ufef1'
	"yehmeeminitialarabic":                    0xfcdd, // ﳝ '\ufcdd'
	"yehmeemisolatedarabic":                   0xfc58, // ﱘ '\ufc58'
	"yehnoonfinalarabic":                      0xfc94, // ﲔ '\ufc94'
	"yehthreedotsbelowarabic":                 0x06d1, // ۑ '\u06d1'
	"yehwithalefmaksurafinal":                 0xfc95, // ﲕ '\ufc95'
	"yehwithalefmaksuraisolated":              0xfc59, // ﱙ '\ufc59'
	"yehwithhahinitial":                       0xfcdb, // ﳛ '\ufcdb'
	"yehwithhamzaaboveisolated":               0xfe89, // ﺉ '\ufe89'
	"yehwithjeeminitial":                      0xfcda, // ﳚ '\ufcda'
	"yehwithkhahinitial":                      0xfcdc, // ﳜ '\ufcdc'
	"yehwithrehfinal":                         0xfc91, // ﲑ '\ufc91'
	"yekorean":                                0x3156, // ㅖ '\u3156'
	"yen":                                     0x00a5, // ¥ '\u00a5'
	"yenmonospace":                            0xffe5, // ￥ '\uffe5'
	"yeokorean":                               0x3155, // ㅕ '\u3155'
	"yeorinhieuhkorean":                       0x3186, // ㆆ '\u3186'
	"yerahbenyomohebrew":                      0x05aa, // ֪ '\u05aa'
	"yericyrillic":                            0x044b, // ы '\u044b'
	"yerudieresiscyrillic":                    0x04f9, // ӹ '\u04f9'
	"yesieungkorean":                          0x3181, // ㆁ '\u3181'
	"yesieungpansioskorean":                   0x3183, // ㆃ '\u3183'
	"yesieungsioskorean":                      0x3182, // ㆂ '\u3182'
	"yetivhebrew":                             0x059a, // ֚ '\u059a'
	"ygrave":                                  0x1ef3, // ỳ '\u1ef3'
	"yhook":                                   0x01b4, // ƴ '\u01b4'
	"yhookabove":                              0x1ef7, // ỷ '\u1ef7'
	"yiarmenian":                              0x0575, // յ '\u0575'
	"yicyrillic":                              0x0457, // ї '\u0457'
	"yikorean":                                0x3162, // ㅢ '\u3162'
	"yinyang":                                 0x262f, // ☯ '\u262f'
	"yiwnarmenian":                            0x0582, // ւ '\u0582'
	"ymonospace":                              0xff59, // ｙ '\uff59'
	"yoddageshhebrew":                         0xfb39, // יּ '\ufb39'
	"yodyodhebrew":                            0x05f2, // ײ '\u05f2'
	"yodyodpatahhebrew":                       0xfb1f, // ײַ '\ufb1f'
	"yogh":                                    0x0222, // Ȣ '\u0222'
	"yoghcurl":                                0x0223, // ȣ '\u0223'
	"yohiragana":                              0x3088, // よ '\u3088'
	"yoikorean":                               0x3189, // ㆉ '\u3189'
	"yokatakana":                              0x30e8, // ヨ '\u30e8'
	"yokatakanahalfwidth":                     0xff96, // ﾖ '\uff96'
	"yokorean":                                0x315b, // ㅛ '\u315b'
	"yosmallhiragana":                         0x3087, // ょ '\u3087'
	"yosmallkatakana":                         0x30e7, // ョ '\u30e7'
	"yosmallkatakanahalfwidth":                0xff6e, // ｮ '\uff6e'
	"yotgreek":                                0x03f3, // ϳ '\u03f3'
	"yoyaekorean":                             0x3188, // ㆈ '\u3188'
	"yoyakorean":                              0x3187, // ㆇ '\u3187'
	"yoyakthai":                               0x0e22, // ย '\u0e22'
	"yoyingthai":                              0x0e0d, // ญ '\u0e0d'
	"yparen":                                  0x24b4, // ⒴ '\u24b4'
	"ypogegrammeni":                           0x037a, // ͺ '\u037a'
	"ypogegrammenigreekcmb":                   0x0345, // ͅ '\u0345'
	"yr":                                      0x01a6, // Ʀ '\u01a6'
	"yring":                                   0x1e99, // ẙ '\u1e99'
	"ysuper":                                  0x0241, // Ɂ '\u0241'
	"ysuperior":                               0x02b8, // ʸ '\u02b8'
	"ytilde":                                  0x1ef9, // ỹ '\u1ef9'
	"yturn":                                   0x021e, // Ȟ '\u021e'
	"yturned":                                 0x028e, // ʎ '\u028e'
	"yuhiragana":                              0x3086, // ゆ '\u3086'
	"yuikorean":                               0x318c, // ㆌ '\u318c'
	"yukatakana":                              0x30e6, // ユ '\u30e6'
	"yukatakanahalfwidth":                     0xff95, // ﾕ '\uff95'
	"yukorean":                                0x3160, // ㅠ '\u3160'
	"yusbigcyrillic":                          0x046b, // ѫ '\u046b'
	"yusbigiotifiedcyrillic":                  0x046d, // ѭ '\u046d'
	"yuslittlecyrillic":                       0x0467, // ѧ '\u0467'
	"yuslittleiotifiedcyrillic":               0x0469, // ѩ '\u0469'
	"yusmallhiragana":                         0x3085, // ゅ '\u3085'
	"yusmallkatakana":                         0x30e5, // ュ '\u30e5'
	"yusmallkatakanahalfwidth":                0xff6d, // ｭ '\uff6d'
	"yuyekorean":                              0x318b, // ㆋ '\u318b'
	"yuyeokorean":                             0x318a, // ㆊ '\u318a'
	"yyabengali":                              0x09df, // য় '\u09df'
	"yyadeva":                                 0x095f, // य़ '\u095f'
	"z":                                       0x007a, // z 'z'
	"zaarmenian":                              0x0566, // զ '\u0566'
	"zacute":                                  0x017a, // ź '\u017a'
	"zadeva":                                  0x095b, // ज़ '\u095b'
	"zagurmukhi":                              0x0a5b, // ਜ਼ '\u0a5b'
	"zaharabic":                               0x0638, // ظ '\u0638'
	"zahfinalarabic":                          0xfec6, // ﻆ '\ufec6'
	"zahinitialarabic":                        0xfec7, // ﻇ '\ufec7'
	"zahiragana":                              0x3056, // ざ '\u3056'
	"zahisolated":                             0xfec5, // ﻅ '\ufec5'
	"zahmedialarabic":                         0xfec8, // ﻈ '\ufec8'
	"zainarabic":                              0x0632, // ز '\u0632'
	"zainfinalarabic":                         0xfeb0, // ﺰ '\ufeb0'
	"zainisolated":                            0xfeaf, // ﺯ '\ufeaf'
	"zakatakana":                              0x30b6, // ザ '\u30b6'
	"zaqefgadolhebrew":                        0x0595, // ֕ '\u0595'
	"zaqefqatanhebrew":                        0x0594, // ֔ '\u0594'
	"zarqahebrew":                             0x0598, // ֘ '\u0598'
	"zayindageshhebrew":                       0xfb36, // זּ '\ufb36'
	"zbopomofo":                               0x3117, // ㄗ '\u3117'
	"zcaron":                                  0x017e, // ž '\u017e'
	"zcircle":                                 0x24e9, // ⓩ '\u24e9'
	"zcircumflex":                             0x1e91, // ẑ '\u1e91'
	"zcmp":                                    0x2a1f, // ⨟ '\u2a1f'
	"zcurl":                                   0x0291, // ʑ '\u0291'
	"zdotaccent":                              0x017c, // ż '\u017c'
	"zdotbelow":                               0x1e93, // ẓ '\u1e93'
	"zedescendercyrillic":                     0x0499, // ҙ '\u0499'
	"zedieresiscyrillic":                      0x04df, // ӟ '\u04df'
	"zehiragana":                              0x305c, // ぜ '\u305c'
	"zekatakana":                              0x30bc, // ゼ '\u30bc'
	"zero":                                    0x0030, // 0 '0'
	"zerobengali":                             0x09e6, // ০ '\u09e6'
	"zerodeva":                                0x0966, // ० '\u0966'
	"zerogujarati":                            0x0ae6, // ૦ '\u0ae6'
	"zerogurmukhi":                            0x0a66, // ੦ '\u0a66'
	"zerohackarabic":                          0x0660, // ٠ '\u0660'
	"zeroinferior":                            0x2080, // ₀ '\u2080'
	"zeromonospace":                           0xff10, // ０ '\uff10'
	"zerooldstyle":                            0xf730, //  '\uf730'
	"zeropersian":                             0x06f0, // ۰ '\u06f0'
	"zerosuperior":                            0x2070, // ⁰ '\u2070'
	"zerothai":                                0x0e50, // ๐ '\u0e50'
	"zerowidthjoiner":                         0xfeff, //  '\ufeff'
	"zerowidthspace":                          0x200b, //  '\u200b'
	"zeta":                                    0x03b6, // ζ '\u03b6'
	"zhbopomofo":                              0x3113, // ㄓ '\u3113'
	"zhearmenian":                             0x056a, // ժ '\u056a'
	"zhebreve":                                0x03fe, // Ͼ '\u03fe'
	"zhebrevecyrillic":                        0x04c2, // ӂ '\u04c2'
	"zhecyrillic":                             0x0436, // ж '\u0436'
	"zhedescendercyrillic":                    0x0497, // җ '\u0497'
	"zhedieresiscyrillic":                     0x04dd, // ӝ '\u04dd'
	"zihiragana":                              0x3058, // じ '\u3058'
	"zikatakana":                              0x30b8, // ジ '\u30b8'
	"zinorhebrew":                             0x05ae, // ֮ '\u05ae'
	"zlinebelow":                              0x1e95, // ẕ '\u1e95'
	"zmonospace":                              0xff5a, // ｚ '\uff5a'
	"zohiragana":                              0x305e, // ぞ '\u305e'
	"zokatakana":                              0x30be, // ゾ '\u30be'
	"zparen":                                  0x24b5, // ⒵ '\u24b5'
	"zpipe":                                   0x2a20, // ⨠ '\u2a20'
	"zproject":                                0x2a21, // ⨡ '\u2a21'
	"zretroflexhook":                          0x0290, // ʐ '\u0290'
	"zrthook":                                 0x0220, // Ƞ '\u0220'
	"zstroke":                                 0x01b6, // ƶ '\u01b6'
	"zuhiragana":                              0x305a, // ず '\u305a'
	"zukatakana":                              0x30ba, // ズ '\u30ba'
}
