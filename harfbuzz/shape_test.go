package harfbuzz

import (
	"crypto/sha1"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math"
	"os"
	"strings"
	"testing"

	"github.com/benoitkugler/textlayout/fonts"
	"github.com/benoitkugler/textlayout/fonts/truetype"
)

// ported from harfbuzz/util/hb-shape.cc, main-font-text.hh Copyright © 2010, 2011,2012  Google, Inc. Behdad Esfahbod

const (
	HB_BUFFER_SERIALIZE_FLAG_DEFAULT        = 0x00000000
	HB_BUFFER_SERIALIZE_FLAG_NO_CLUSTERS    = 0x00000001
	HB_BUFFER_SERIALIZE_FLAG_NO_POSITIONS   = 0x00000002
	HB_BUFFER_SERIALIZE_FLAG_NO_GLYPH_NAMES = 0x00000004
	HB_BUFFER_SERIALIZE_FLAG_GLYPH_EXTENTS  = 0x00000008
	HB_BUFFER_SERIALIZE_FLAG_GLYPH_FLAGS    = 0x00000010
	HB_BUFFER_SERIALIZE_FLAG_NO_ADVANCES    = 0x00000020
)

type formatOptionsT struct {
	hideGlyphNames bool
	hidePositions  bool
	hideAdvances   bool
	hideClusters   bool
	showText       bool
	showUnicode    bool
	showLineNum    bool
	showExtents    bool
	showFlags      bool
	// trace          bool
}

func (fm formatOptionsT) serializeLineNo(lineNo int, gs *strings.Builder) {
	if fm.showLineNum {
		fmt.Fprintf(gs, "%d: ", lineNo)
	}
}

func (fm formatOptionsT) serialize(buffer *Buffer, font *Font, flags int, gs *strings.Builder) {
	gs.WriteByte('[')
	for i, glyph := range buffer.Info {
		// TODO: names
		fmt.Fprintf(gs, "%d=%d", glyph.Glyph, glyph.Cluster)
		pos := buffer.Pos[i]
		if pos.XOffset != 0 || pos.YOffset != 0 {
			fmt.Fprintf(gs, "@%d,%d", pos.XOffset, pos.YOffset)
		}
		fmt.Fprintf(gs, "+%d", pos.XAdvance)
		if pos.YAdvance != 0 {
			fmt.Fprintf(gs, ",%d", pos.YAdvance)
		}

		// if (flags & HB_BUFFER_SERIALIZE_FLAG_GLYPH_EXTENTS)
		// {
		//   hb_glyph_extents_t extents;
		//   hb_font_get_glyph_extents(font, info[i].codepoint, &extents);
		//   p += hb_max (0, snprintf (p, ARRAY_LENGTH (b) - (p - b), "<%d,%d,%d,%d>", extents.x_bearing, extents.y_bearing, extents.width, extents.height));
		// }
		if i != len(buffer.Info)-1 {
			gs.WriteByte('|')
		}
	}
	gs.WriteByte(']')
}

func (fm formatOptionsT) serializeBufferOfText(buffer *Buffer, lineNo int, text string, font *Font) string {
	var out strings.Builder
	if fm.showText {
		fm.serializeLineNo(lineNo, &out)
		out.WriteByte('(')
		out.WriteString(text)
		out.WriteByte(')')
		out.WriteByte('\n')
	}

	if fm.showUnicode {
		fm.serializeLineNo(lineNo, &out)
		fm.serialize(buffer, font, HB_BUFFER_SERIALIZE_FLAG_DEFAULT, &out)
		out.WriteByte('\n')
	}
	return out.String()
}

func (fm formatOptionsT) serializeBufferOfGlyphs(buffer *Buffer, lineNo int, text string, font *Font,
	flags int) string {
	var out strings.Builder
	fm.serializeLineNo(lineNo, &out)
	fm.serialize(buffer, font, flags, &out)
	out.WriteByte('\n')
	return out.String()
}

type outputBufferT struct {
	out         io.Writer
	font        *Font
	formatFlags int
	lineNo      int
	format      formatOptionsT
}

//    outputBufferT (option_parser_t *parser)
// 		   : options (parser, hb_buffer_serialize_list_formats ()),
// 			 format (parser),
// 			 gs (nullptr),
// 			 lineNo (0),
// 			 font (nullptr),
// 			 outputFormat (HB_BUFFER_SERIALIZE_FORMAT_INVALID),
// 			 formatFlags (HB_BUFFER_SERIALIZE_FLAG_DEFAULT) {}

func (out *outputBufferT) init(buffer *Buffer, fontOpts fontOptionsT) {
	out.font = fontOpts.getFont()
	out.lineNo = 0
	//  options.get_file_handle ();
	//  gs = g_string_new (nullptr);
	//  lineNo = 0;
	//  if (!options.outputFormat)
	//    outputFormat = HB_BUFFER_SERIALIZE_FORMAT_TEXT;
	//  else
	//    outputFormat = hb_buffer_serialize_format_from_string (options.outputFormat, -1);
	//  /* An empty "outputFormat" parameter basically skips output generating.
	//   * Useful for benchmarking. */
	//  if ((!options.outputFormat || *options.outputFormat) &&
	//  !hb_buffer_serialize_format_to_string (outputFormat))
	//  {
	//    if (options.explicit_output_format)
	//  fail (false, "Unknown output format `%s'; supported formats are: %s",
	// 	   options.outputFormat,
	// 	   g_strjoinv ("/", const_cast<char**> (options.supported_formats)));
	//    else
	//  /* Just default to TEXT if not explicitly requested and the
	//   * file extension is not recognized. */
	//  outputFormat = HB_BUFFER_SERIALIZE_FORMAT_TEXT;
	//  }
	flags := HB_BUFFER_SERIALIZE_FLAG_DEFAULT
	if out.format.hideGlyphNames {
		flags |= HB_BUFFER_SERIALIZE_FLAG_NO_GLYPH_NAMES
	}
	if out.format.hideClusters {
		flags |= HB_BUFFER_SERIALIZE_FLAG_NO_CLUSTERS
	}
	if out.format.hidePositions {
		flags |= HB_BUFFER_SERIALIZE_FLAG_NO_POSITIONS
	}
	if out.format.hideAdvances {
		flags |= HB_BUFFER_SERIALIZE_FLAG_NO_ADVANCES
	}
	if out.format.showExtents {
		flags |= HB_BUFFER_SERIALIZE_FLAG_GLYPH_EXTENTS
	}
	if out.format.showFlags {
		flags |= HB_BUFFER_SERIALIZE_FLAG_GLYPH_FLAGS
	}
	out.formatFlags = flags
}

func (out *outputBufferT) newLine() { out.lineNo++ }

func (out *outputBufferT) consumeText(buffer *Buffer, text string) {
	s := out.format.serializeBufferOfText(buffer, out.lineNo, text, out.font)
	fmt.Fprintf(out.out, "%s", s)
}

func (out *outputBufferT) outputError(message string) {
	var gs strings.Builder
	out.format.serializeLineNo(out.lineNo, &gs)
	fmt.Fprintf(&gs, "error: %s", message)
	gs.WriteByte('\n')
	fmt.Fprintf(out.out, "%s", gs.String())
}

func (out *outputBufferT) consumeGlyphs(buffer *Buffer, text string) {
	s := out.format.serializeBufferOfGlyphs(buffer, out.lineNo, text, out.font,
		out.formatFlags)
	fmt.Fprintf(out.out, "%s", s)
}

//    static hb_bool_t
//    message_func (buffer *Buffer,
// 		 hb_font_t *font,
// 		 const char *message,
// 		 void *user_data)
//    {
// 	 outputBufferT *that = (outputBufferT *) user_data;
// 	 that.trace (buffer, font, message);
// 	 return true;
//    }
//    void
//    trace (buffer *Buffer,
// 	  hb_font_t *font,
// 	  const char *message)
//    {
// 	 g_string_set_size (gs, 0);
// 	 format.serialize_line_no (lineNo, gs);
// 	 g_string_append_printf (gs, "trace: %s	buffer: ", message);
// 	 format.serialize (buffer, font, outputFormat, formatFlags, gs);
// 	 g_string_append_c (gs, '\n');
// 	 fprintf (options.fp, "%s", gs.str);
//    }
//    protected:
//    output_options_t options;
//    formatOptionsT format;
//    GString *gs;
//    unsigned int lineNo;
//    hb_font_t *font;
//    hb_buffer_serialize_format_t outputFormat;
//    hb_buffer_serialize_flags_t formatFlags;

// static char *
// locale_to_utf8 (char *s)
// {
//   char *t;
//   GError *error = nullptr;

//   t = g_locale_to_utf8 (s, -1, nullptr, nullptr, &error);
//   if (!t)
//   {
//      fail (true, "Failed converting text to UTF-8");
//   }

//   return t;
// }

type fontOptionsT struct {
	font                 *Font // cached value of getFont()
	fontFile             string
	variations           []truetype.Variation
	defaultFontSize      int
	subpixelBits         int
	fontSizeX, fontSizeY float64
	ptem                 float32
	yPpem, xPpem         uint16
}

const fontSizeUpem float64 = 0x7FFFFFFF

func newFontOptions(defaultFontSize, subpixelBits int) fontOptionsT {
	return fontOptionsT{
		defaultFontSize: defaultFontSize,
		subpixelBits:    subpixelBits,
		fontSizeX:       float64(defaultFontSize),
		fontSizeY:       float64(defaultFontSize),
	}
}

func (fo *fontOptionsT) getFont() *Font {
	if fo.font != nil {
		return fo.font
	}

	/* Create the blob */
	if fo.fontFile == "" {
		log.Fatal("no font file specified")
	}

	face := openFontFile(fo.fontFile)

	/* Create the face */

	fo.font = NewFont(face.LoadMetrics())

	if fo.fontSizeX == fontSizeUpem {
		fo.fontSizeX = float64(fo.font.faceUpem)
	}
	if fo.fontSizeY == fontSizeUpem {
		fo.fontSizeY = float64(fo.font.faceUpem)
	}

	fo.font.XPpem, fo.font.YPpem = fo.xPpem, fo.yPpem
	fo.font.Ptem = fo.ptem

	scaleX := scalbnf(fo.fontSizeX, fo.subpixelBits)
	scaleY := scalbnf(fo.fontSizeY, fo.subpixelBits)
	fo.font.XScale, fo.font.YScale = scaleX, scaleY

	fo.font.setVariations(fo.variations)

	return fo.font
}

func scalbnf(x float64, exp int) int32 {
	return int32(x * (math.Pow(2, float64(exp))))
}

type textOptionsT struct {
	textBefore, textAfter string
	lines                 []string
}

// template <typename consumer_t, int defaultFontSize, int subpixelBits>
type mainFontTextT struct {
	// option_parser_t options; "[FONT-FILE] [TEXT]"
	consumer shapeConsumerT
	input    textOptionsT
	fontOpts fontOptionsT
}

func newFormatOptionsT() formatOptionsT {
	return formatOptionsT{
		showText:    false,
		showUnicode: false,
		showLineNum: false,
		showExtents: false,
		showFlags:   false,
	}
}

type shapeOptionsT struct {
	props                     SegmentProperties
	features                  []Feature
	numIterations             int
	invisibleGlyph            fonts.GlyphIndex
	clusterLevel              ClusterLevel
	bot                       bool
	eot                       bool
	preserveDefaultIgnorables bool
	normalizeGlyphs           bool
	verify                    bool
	removeDefaultIgnorables   bool
}

func (so *shapeOptionsT) setupBuffer(buffer *Buffer) {
	buffer.Props = so.props
	var flags Flags
	if so.bot {
		flags |= Bot
	}
	if so.eot {
		flags |= Eot
	}
	if so.preserveDefaultIgnorables {
		flags |= PreserveDefaultIgnorables
	}
	if so.removeDefaultIgnorables {
		flags |= RemoveDefaultIgnorables
	}
	buffer.Flags = flags
	buffer.Invisible = so.invisibleGlyph
	buffer.ClusterLevel = so.clusterLevel
	buffer.guessSegmentProperties()
}

func copyBufferProperties(dst, src *Buffer) {
	dst.Props = src.Props
	dst.Flags = src.Flags
	dst.ClusterLevel = src.ClusterLevel
}

func clearBufferContents(buffer *Buffer) {
	repl := buffer.Replacement
	*buffer = Buffer{Replacement: repl}
}

func appendBuffer(dst, src *Buffer, start, end int) {
	dst.Info = append(dst.Info, src.Info[start:end]...)
	dst.Pos = append(dst.Pos, src.Pos[start:end]...)
}

func (so *shapeOptionsT) populateBuffer(buffer *Buffer, text, textBefore, textAfter string) {
	clearBufferContents(buffer)

	if textBefore != "" {
		t := []rune(textBefore)
		buffer.AddRunes(t, len(t), 0)
	}
	t := []rune(text)
	buffer.AddRunes(t, 0, len(t))
	if textAfter != "" {
		t := []rune(textAfter)
		buffer.AddRunes(t, 0, 0)
	}

	so.setupBuffer(buffer)
}

func (so *shapeOptionsT) shape(font *Font, buffer *Buffer) error {
	var textBuffer *Buffer
	if so.verify {
		textBuffer = NewBuffer()
		appendBuffer(textBuffer, buffer, 0, len(buffer.Info))
	}
	buffer.Shape(font, so.features)
	// if !hb_shape_full(font, buffer, features, num_features, shapers) {
	// 	if error {
	// 		*error = "all shapers failed."
	// 	}
	// 	return false
	// }

	if so.normalizeGlyphs {
		buffer.normalizeGlyphs()
	}

	if so.verify {
		if err := so.verifyBuffer(buffer, textBuffer, font); err != nil {
			return err
		}
	}

	return nil
}

func (so *shapeOptionsT) verifyBuffer(buffer, textBuffer *Buffer, font *Font) error {
	if err := so.verifyBufferMonotone(buffer); err != nil {
		return err
	}
	if err := so.verifyBufferSafeToBreak(buffer, textBuffer, font); err != nil {
		return err
	}
	return nil
}

/* Check that clusters are monotone. */
func (so *shapeOptionsT) verifyBufferMonotone(buffer *Buffer) error {
	if so.clusterLevel == MonotoneGraphemes || so.clusterLevel == MonotoneCharacters {
		isForward := buffer.Props.Direction.isForward()

		info := buffer.Info

		for i := 1; i < len(info); i++ {
			if info[i-1].Cluster != info[i].Cluster && (info[i-1].Cluster < info[i].Cluster) != isForward {
				return fmt.Errorf("cluster at index %d is not monotone", i)
			}
		}
	}

	return nil
}

func (so *shapeOptionsT) verifyBufferSafeToBreak(buffer, textBuffer *Buffer, font *Font) error {
	if so.clusterLevel != MonotoneGraphemes && so.clusterLevel != MonotoneCharacters {
		/* Cannot perform this check without monotone clusters.
		 * Then again, unsafe-to-break flag is much harder to use without
		 * monotone clusters. */
		return nil
	}

	/* Check that breaking up shaping at safe-to-break is indeed safe. */

	fragment, reconstruction := NewBuffer(), NewBuffer()
	copyBufferProperties(reconstruction, buffer)

	info := buffer.Info
	text := textBuffer.Info

	/* Chop text and shape fragments. */
	forward := buffer.Props.Direction.isForward()
	start := 0
	textStart := len(textBuffer.Info)
	if forward {
		textStart = 0
	}
	textEnd := textStart
	for end := 1; end < len(buffer.Info)+1; end++ {
		offset := 1
		if forward {
			offset = 0
		}
		if end < len(buffer.Info) && (info[end].Cluster == info[end-1].Cluster ||
			info[end-offset].mask&GlyphFlagUnsafeToBreak != 0) {
			continue
		}

		/* Shape segment corresponding to glyphs start..end. */
		if end == len(buffer.Info) {
			if forward {
				textEnd = len(textBuffer.Info)
			} else {
				textStart = 0
			}
		} else {
			if forward {
				cluster := info[end].Cluster
				for textEnd < len(textBuffer.Info) && text[textEnd].Cluster < cluster {
					textEnd++
				}
			} else {
				cluster := info[end-1].Cluster
				for textStart != 0 && text[textStart-1].Cluster >= cluster {
					textStart--
				}
			}
		}
		if !(textStart < textEnd) {
			return fmt.Errorf("unexpected %d >= %d", textStart, textEnd)
		}

		if debugMode {
			fmt.Printf("start %d end %d text start %d end %d\n", start, end, textStart, textEnd)
		}

		clearBufferContents(fragment)
		copyBufferProperties(fragment, buffer)

		/* TODO: Add pre/post context text. */
		flags := fragment.Flags
		if 0 < textStart {
			flags = (flags & ^Bot)
		}
		if textEnd < len(textBuffer.Info) {
			flags = (flags & ^Eot)
		}
		fragment.Flags = flags

		appendBuffer(fragment, textBuffer, textStart, textEnd)
		fragment.Shape(font, so.features)
		// if !hb_shape_full(font, fragment, features, num_features, shapers) {
		// 	if error {
		// 		*error = "all shapers failed while shaping fragment."
		// 	}
		// 	return false
		// }
		appendBuffer(reconstruction, fragment, 0, len(fragment.Info))

		start = end
		if forward {
			textStart = textEnd
		} else {
			textEnd = textStart
		}
	}

	diff := bufferDiff(reconstruction, buffer, ^fonts.GlyphIndex(0), 0)
	if diff != HB_BUFFER_DIFF_FLAG_EQUAL {
		/* Return the reconstructed result instead so it can be inspected. */
		buffer.Info = nil
		buffer.Pos = nil
		appendBuffer(buffer, reconstruction, 0, len(reconstruction.Info))

		return fmt.Errorf("safe-to-break test failed: %d", diff)
	}

	return nil
}

type shapeConsumerT struct {
	font   *Font
	buffer *Buffer
	output outputBufferT
	shaper shapeOptionsT
}

func (sh *shapeConsumerT) init(buffer *Buffer, fontOpts fontOptionsT) {
	sh.font = fontOpts.getFont()
	sh.buffer = buffer
	sh.output.init(buffer, fontOpts)
}

func (sh *shapeConsumerT) consumeLine(text, textBefore, textAfter string) {
	sh.output.newLine()

	for n := sh.shaper.numIterations; n != 0; n-- {
		sh.shaper.populateBuffer(sh.buffer, text, textBefore, textAfter)
		if n == 1 {
			sh.output.consumeText(sh.buffer, text)
		}
		if err := sh.shaper.shape(sh.font, sh.buffer); err != nil {
			sh.output.outputError(err.Error())
			break
		}
	}

	sh.output.consumeGlyphs(sh.buffer, text)
}

func newMainFontTextT(defaultFontSize, subpixelBits int) mainFontTextT {
	return mainFontTextT{
		fontOpts: newFontOptions(defaultFontSize, subpixelBits),
		// input(&options),
		// consumer(&options),
	}
}

func (mft *mainFontTextT) main(out io.Writer, fontFile, text string) {
	mft.fontOpts.fontFile = fontFile
	mft.input.lines = strings.Split(text, "\n")
	mft.consumer.output.out = out

	buffer := NewBuffer()
	mft.consumer.init(buffer, mft.fontOpts)

	for _, line := range mft.input.lines {
		mft.consumer.consumeLine(line, mft.input.textBefore, mft.input.textAfter)
	}
}

func shape(fontFile, text string) {
	var driver mainFontTextT
	driver.main(os.Stdout, fontFile, text)
}

const featuresUsage = `Comma-separated list of font features

    Features can be enabled or disabled, either globally or limited to
    specific character ranges.  The format for specifying feature settings
    follows.  All valid CSS font-feature-settings values other than 'normal'
    and the global values are also accepted, though not documented below.
    CSS string escapes are not supported.

    The range indices refer to the positions between Unicode characters,
    unless the --utf8-clusters is provided, in which case range indices
    refer to UTF-8 byte indices. The position before the first character
    is always 0.

    The format is Python-esque.  Here is how it all works:

      Syntax:       Value:    Start:    End:

    Setting value:
      "kern"        1         0         ∞         // Turn feature on
      "+kern"       1         0         ∞         // Turn feature on
      "-kern"       0         0         ∞         // Turn feature off
      "kern=0"      0         0         ∞         // Turn feature off
      "kern=1"      1         0         ∞         // Turn feature on
      "aalt=2"      2         0         ∞         // Choose 2nd alternate

    Setting index:
      "kern[]"      1         0         ∞         // Turn feature on
      "kern[:]"     1         0         ∞         // Turn feature on
      "kern[5:]"    1         5         ∞         // Turn feature on, partial
      "kern[:5]"    1         0         5         // Turn feature on, partial
      "kern[3:5]"   1         3         5         // Turn feature on, range
      "kern[3]"     1         3         3+1       // Turn feature on, single char

    Mixing it all:

      "aalt[3:5]=2" 2         3         5         // Turn 2nd alternate on for range
`


func (opts *shapeOptionsT) parseFeatures (s string) error {
	if s == "" {
		opts.features = nil 
		return nil
	}

  features := strings.Split(s, ",")
  opts.features = make([]Feature, len(features))
 
  p = s;
  shape_opts->num_features = 0;
  while (p && *p) {
    char *end = strchr (p, ',');
    if (hb_feature_from_string (p, end ? end - p : -1, &shape_opts->features[shape_opts->num_features]))
      shape_opts->num_features++;
    p = end ? end + 1 : nullptr;
  }

  return true;
}

// parse the options, written in command line format
func parseOptions(options string) formatOptionsT {
	flags := flag.NewFlagSet("options", flag.ContinueOnError)

	var fmtOpts formatOptionsT
	flags.BoolVar(&fmtOpts.hideClusters, "no-clusters", false, "Do not output cluster indices")
	flags.BoolVar(&fmtOpts.hideGlyphNames, "no-glyph-names", false, "Output glyph indices instead of names")
	flags.BoolVar(&fmtOpts.hidePositions, "no-positions", false, "Do not output glyph positions")
	flags.BoolVar(&fmtOpts.hideAdvances, "no-advances", false, "Do not output glyph advances")

	var shapeOpts shapeOptionsT

	featuresUsage
	flags.Func("features", featuresUsage, )
	// var fontOpts fontOptionsT

	err := flags.Parse(strings.Split(options, " "))
	check(err)

	return fmtOpts
}


// opens and parses a test file containing
// the font file, the unicode input and the expected result
func readHarfbuzzTestFile(filename string) {
	f, err := ioutil.ReadFile(filename)
	check(err)

	for _, line := range strings.Split(string(f), "\n") {
		if strings.HasPrefix(line, "#") || strings.TrimSpace(line) == "" { // skip comments
			continue
		}

		chunks := strings.Split(line, ":")
		if len(chunks) != 4 {
			check(fmt.Errorf("invalid test file: line %s", line))
		}
		fontFile, options, unicodes, glyphsExpected := chunks[0], chunks[1], chunks[2], chunks[3]

		withHash := strings.Split(fontFile, "@")
		if len(withHash) >= 2 {
			ff, err := ioutil.ReadFile(fontFile)
			check(err)

			hash := sha1.Sum(ff)
			trimmedHash := strings.TrimSpace(string(hash[:]))
			if exp := withHash[1]; trimmedHash != exp {
				check(fmt.Errorf("invalid font file hash: expected %s, got %s", exp, trimmedHash))
			}
		}

		verify := glyphsExpected != "*"

		fmt.Println("options :", options)
		fmt.Println(unicodes, verify)

		fmt.Println(parseOptions(options))
	}
}

func TestShapeExpected(t *testing.T) {
	readHarfbuzzTestFile("testdata/data/aots/tests/classdef1_empty.tests")
}
