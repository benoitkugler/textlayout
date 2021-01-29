package fribidi

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"strings"
	"testing"

	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/encoding/unicode"
)

const maxStrLen = 65000

//    /* Break help string into little ones, to assure ISO C89 conformance */
//    printf ("Usage: " appname " [OPTION]... [FILE]...\n"
// 	   "A command line interface for the " FRIBIDI_NAME " library.\n"
// 	   "Convert a logical string to visual.\n"
// 	   "\n"
// 	   "  -h, --help            Display this information and exit\n"
// 	   "  -V, --version         Display version information and exit\n"
// 	   "  -v, --verbose         Verbose mode, same as --basedir --ltov --vtol\n"
// 	   "                        --levels\n");
//    printf ("  -d, --debug           Output debug information\n"
// 	   "  -t, --test            Test " FRIBIDI_NAME
// 	   ", same as --clean --nobreak\n"
// 	   "                        --showinput --reordernsm --width %d\n",
// 	   default_text_width);
//    printf ("  -c, --charset CS      Specify character set, default is %s\n"
// 	   "      --charsetdesc CS  Show descriptions for character set CS and exit\n"
// 	   "      --caprtl          Old style: set character set to CapRTL\n",
// 	   char_set);
//    printf ("      --showinput       Output the input string too\n"
// 	   "      --nopad           Do not right justify RTL lines\n"
// 	   "      --nobreak         Do not break long lines\n"
// 	   "  -w, --width W         Screen width for padding, default is %d, but if\n"
// 	   "                        environment variable COLUMNS is defined, its value\n"
// 	   "                        will be used, --width overrides both of them.\n",
// 	   default_text_width);
//    printf
// 	 ("  -B, --bol BOL         Output string BOL before the visual string\n"
// 	  "  -E, --eol EOL         Output string EOL after the visual string\n"
// 	  "      --rtl             Force base direction to RTL\n"
// 	  "      --ltr             Force base direction to LTR\n"
// 	  "      --wrtl            Set base direction to RTL if no strong character found\n");
//    printf
// 	 ("      --wltr            Set base direction to LTR if no strong character found\n"
// 	  "                        (default)\n"
// 	  "      --nomirror        Turn mirroring off, to do it later\n"
// 	  "      --reordernsm      Reorder NSM sequences to follow their base character\n"
// 	  "      --clean           Remove explicit format codes in visual string\n"
// 	  "                        output, currently does not affect other outputs\n"
// 	  "      --basedir         Output Base Direction\n");
//    printf ("      --ltov            Output Logical to Visual position map\n"
// 	   "      --vtol            Output Visual to Logical position map\n"
// 	   "      --levels          Output Embedding Levels\n"
// 	   "      --novisual        Do not output the visual string, to be used with\n"
// 	   "                        --basedir, --ltov, --vtol, --levels\n");
//    printf ("  All string indexes are zero based\n" "\n" "Output:\n"
// 	   "  For each line of input, output something like this:\n"
// 	   "    [input-str` => '][BOL][[padding space]visual-str][EOL]\n"
// 	   "    [\\n base-dir][\\n ltov-map][\\n vtol-map][\\n levels]\n");

type charsetI interface {
	decode(input []byte) []rune
	encode(input []rune) []byte
}

type stdCharset struct {
	encoding.Encoding
}

func (s stdCharset) decode(input []byte) []rune {
	u, err := s.Encoding.NewDecoder().Bytes(input)
	if err != nil {
		log.Fatal(err)
	}
	return []rune(string(u))
}

func (s stdCharset) encode(input []rune) []byte {
	u, err := s.Encoding.NewEncoder().Bytes([]byte(string(input)))
	if err != nil {
		log.Fatal(err)
	}
	return u
}

func parseCharset(filename string) charsetI {
	switch enc := strings.Split(filename, "_")[1]; enc {
	case "UTF-8":
		return stdCharset{unicode.UTF8}
	case "ISO8859-8":
		return stdCharset{charmap.ISO8859_8}
	case "CapRTL":
		return capRTLCharset{}
	default:
		panic("unsupported encoding " + enc)
	}
}

func printDetails(nlFound string, fileOut io.Writer, base ParType, logToVis []int, out Visual) string {
	fmt.Fprintf(fileOut, "%s", nlFound)
	if dirToLevel(base) != 0 {
		fmt.Fprintf(fileOut, "Base direction: %s", "R")
	} else {
		fmt.Fprintf(fileOut, "Base direction: %s", "L")
	}
	nlFound = "\n"
	fmt.Fprintf(fileOut, "%s", nlFound)
	for _, ltov := range logToVis {
		fmt.Fprintf(fileOut, "%d ", ltov)
	}
	nlFound = "\n"
	fmt.Fprintf(fileOut, "%s", nlFound)
	for _, vtol := range out.VisualToLogical {
		fmt.Fprintf(fileOut, "%d ", vtol)
	}
	nlFound = "\n"
	fmt.Fprintf(fileOut, "%s", nlFound)
	for _, level := range out.EmbeddingLevels {
		fmt.Fprintf(fileOut, "%d ", level)
	}
	return "\n"
}

// go fmt uses length in runes to apply padding, not bytes,
// so we need to emulate the C behavior
func bytesPadding(v []byte, widthInBytes int) []byte {
	if len(v) >= widthInBytes { // no padding needed
		return nil
	}
	padding := bytes.Repeat([]byte{' '}, widthInBytes-len(v)) // convert to unicode points
	return padding
}

func processFile(filename string, fileOut io.Writer) error {
	const textWidth = 80

	doClean, doReorderNsm, doMirror := true, true, true
	showInput := true
	doBreak := false

	doPad := true
	showVisual := true
	showDetails := false
	// char_set := "UTF-8"
	// bol_text := nil
	// eol_text := nil
	var inputBaseDirection ParType = ON

	// s = getenv("COLUMNS")
	// if s {
	// 	i = atoi(s)
	// 	if i > 0 {
	// 		text_width = i
	// 	}
	// }

	charset := parseCharset(filename)

	flags := DefaultFlags.adjust(ShapeMirroring, doMirror)
	flags = flags.adjust(ReorderNSM, doReorderNsm)

	paddingWidth := textWidth
	if showInput {
		paddingWidth = (textWidth - 10) / 2
	}
	breakWidth := 3 * maxStrLen
	if doBreak {
		breakWidth = paddingWidth
	}

	b, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}

	/* Read and process input one line at a time */
	for _, line := range bytes.Split(b, []byte{'\n'}) {
		if len(line) == 0 {
			continue
		}
		logical := charset.decode(line)

		var nlFound = ""

		/* Create a bidi string. */
		base := inputBaseDirection

		out, _ := LogicalToVisual(flags, logical, &base)
		logToVis := out.LogicalToVisual()

		if showInput {
			fmt.Fprintf(fileOut, "%s%s => ", line, bytesPadding(line, paddingWidth))
		}

		/* Remove explicit marks, if asked for. */
		if doClean {
			out.Str = removeBidiMarks(out.Str, logToVis, out.VisualToLogical, out.EmbeddingLevels)
		}
		if showVisual {
			fmt.Fprintf(fileOut, "%s", nlFound)
			// if bol_text {
			// 	fmt.Fprintf(fileOut, "%s", bol_text)
			// }

			/* Convert it to input charset and print. */
			var st int
			for idx := 0; idx < len(out.Str); {
				var inlen int

				wid := breakWidth
				st = idx
				if _, isCapRTL := charset.(capRTLCharset); !isCapRTL {
					for wid > 0 && idx < len(out.Str) {
						if GetBidiType(out.Str[idx]).isExplicitOrIsolateOrBnOrNsm() {
							wid -= 0
						} else {
							wid -= 1
						}
						idx++
					}
				} else {
					for wid > 0 && idx < len(out.Str) {
						wid--
						idx++
					}
				}
				if wid < 0 && idx-st > 1 {
					idx--
				}
				inlen = idx - st

				outBytes := charset.encode(out.Str[st : inlen+st])
				var w int
				if base.IsRtl() && doPad {
					w = paddingWidth + len(outBytes) - (breakWidth - wid)
				}
				fmt.Fprintf(fileOut, "%s%s", bytesPadding(outBytes, w), outBytes)

				if idx < len(out.Str) {
					fmt.Fprintln(fileOut)
				}
			}
			// if eol_text {
			// 	fmt.Fprintf(fileOut, "%s", eol_text)
			// }

			nlFound = "\n"
		}
		if showDetails {
			nlFound = printDetails(nlFound, fileOut, base, logToVis, out)
		}
		if nlFound != "" {
			fmt.Fprintln(fileOut)
		}
	}
	return nil
}

func TestShape(t *testing.T) {
	for _, file := range []string{
		"test/test_CapRTL_explicit",
		"test/test_CapRTL_explicit",
		"test/test_CapRTL_implicit",
		"test/test_CapRTL_isolate",
		"test/test_ISO8859-8_hebrew",
		"test/test_UTF-8_persian",
		"test/test_UTF-8_reordernsm",
	} {
		var out bytes.Buffer
		err := processFile(file+".input", &out)
		if err != nil {
			t.Fatal("error in test file", file, err)
		}

		ref, err := ioutil.ReadFile(file + ".reference")
		if err != nil {
			t.Fatal(err)
		}
		if string(ref) != out.String() {
			t.Errorf("file %s: expected\n%s\ngot\n%s", file, ref, out.String())
		}
	}
}

func TestTypes(t *testing.T) {
	cs := [...]CharType{LTR, RTL, EN, ON, WLTR, WRTL, PDF, LRI, RLI, FSI, BS, NSM, AL, AN, CS, ET, PDI, LRO, RLO, RLE, LRE, WS, ES, BN}
	csStrings := [...]string{"LTR", "RTL", "EN", "ON", "WLTR", "WRTL", "PDF", "LRI", "RLI", "FSI", "BS", "NSM", "AL", "AN", "CS", "ET", "PDI", "LRO", "RLO", "RLE", "LRE", "WS", "ES", "BN"}
	for i, c := range cs {
		if c.String() != csStrings[i] {
			t.Error()
		}
	}
}
