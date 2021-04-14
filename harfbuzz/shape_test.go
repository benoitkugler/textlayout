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
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/benoitkugler/textlayout/fonts"
	"github.com/benoitkugler/textlayout/fonts/truetype"
	"github.com/benoitkugler/textlayout/language"
)

// ported from harfbuzz/util/hb-shape.cc, main-font-text.hh Copyright © 2010, 2011,2012  Google, Inc. Behdad Esfahbod

const (
	serializeNoClusters = 1 << iota
	serializeFlagNoPositions
	serializeFlagNoGlyphNames
	serializeFlagGlyphExtents
	serializeFlagGlyphFlags
	serializeFlagNoAdvances

	serializeDefault = 0x00000000
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
	var x, y Position
	for i, glyph := range buffer.Info {
		if flags&serializeFlagNoGlyphNames != 0 {
			fmt.Fprintf(gs, "%d", glyph.Glyph)
		} else {
			gs.WriteString(font.glyphToString(glyph.Glyph))
		}

		if flags&serializeNoClusters == 0 {
			fmt.Fprintf(gs, "=%d", glyph.Cluster)
		}
		pos := buffer.Pos[i]

		if flags&serializeFlagNoPositions == 0 {
			if x+pos.XOffset != 0 || y+pos.YOffset != 0 {
				fmt.Fprintf(gs, "@%d,%d", x+pos.XOffset, y+pos.YOffset)
			}
			if flags&serializeFlagNoAdvances == 0 {
				fmt.Fprintf(gs, "+%d", pos.XAdvance)
				if pos.YAdvance != 0 {
					fmt.Fprintf(gs, ",%d", pos.YAdvance)
				}
			}
		}

		if (flags & serializeFlagGlyphExtents) != 0 {
			extents, _ := font.getGlyphExtents(glyph.Glyph)
			fmt.Fprintf(gs, "<%d,%d,%d,%d>", extents.XBearing, extents.YBearing, extents.Width, extents.Height)
		}

		if i != len(buffer.Info)-1 {
			gs.WriteByte('|')
		}

		if flags&serializeFlagNoAdvances != 0 {
			x += pos.XAdvance
			y += pos.YAdvance
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
		fm.serialize(buffer, font, serializeDefault, &out)
		out.WriteByte('\n')
	}
	return out.String()
}

func (fm formatOptionsT) serializeBufferOfGlyphs(buffer *Buffer, lineNo int, font *Font,
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
// 			 formatFlags (serializeFlagDefault) {}

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
	flags := serializeDefault
	if out.format.hideGlyphNames {
		flags |= serializeFlagNoGlyphNames
	}
	if out.format.hideClusters {
		flags |= serializeNoClusters
	}
	if out.format.hidePositions {
		flags |= serializeFlagNoPositions
	}
	if out.format.hideAdvances {
		flags |= serializeFlagNoAdvances
	}
	if out.format.showExtents {
		flags |= serializeFlagGlyphExtents
	}
	if out.format.showFlags {
		flags |= serializeFlagGlyphFlags
	}
	out.formatFlags = flags
}

func (out *outputBufferT) newLine() { out.lineNo++ }

func (out *outputBufferT) consumeText(buffer *Buffer, text []rune) {
	s := out.format.serializeBufferOfText(buffer, out.lineNo, string(text), out.font)
	fmt.Fprintf(out.out, "%s", s)
}

func (out *outputBufferT) serializeShapeOutput(buffer *Buffer) string {
	return out.format.serializeBufferOfGlyphs(buffer, out.lineNo, out.font,
		out.formatFlags)
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
	font *Font // cached value of getFont()

	fontFile   string
	variations []Variation
	fontIndex  int // index of the font in the file

	defaultFontSize      int
	subpixelBits         int
	fontSizeX, fontSizeY float64
	ptem                 float64
	yPpem, xPpem         uint16
}

const fontSizeUpem = 0x7FFFFFFF

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

	f, err := os.Open(fo.fontFile)
	check(err)

	fonts, err := truetype.Loader.Load(f)
	check(err)

	if fo.fontIndex >= len(fonts) {
		// harfbuzz seems to be OK with an invalid font
		// in pratice, it seems useless to do shaping without
		// font, so we dont support it, meaning we skip this test
		check(fmt.Errorf("invalid font index %d for length %d", fo.fontIndex, len(fonts)))
	}
	face := fonts[fo.fontIndex]

	/* Create the face */

	fo.font = NewFont(face.LoadMetrics())

	if fo.fontSizeX == fontSizeUpem {
		fo.fontSizeX = float64(fo.font.faceUpem)
	}
	if fo.fontSizeY == fontSizeUpem {
		fo.fontSizeY = float64(fo.font.faceUpem)
	}

	fo.font.XPpem, fo.font.YPpem = fo.xPpem, fo.yPpem
	fo.font.Ptem = float32(fo.ptem)

	scaleX := scalbnf(fo.fontSizeX, fo.subpixelBits)
	scaleY := scalbnf(fo.fontSizeY, fo.subpixelBits)
	fo.font.XScale, fo.font.YScale = scaleX, scaleY

	fo.font.setVariations(fo.variations)

	return fo.font
}

func scalbnf(x float64, exp int) int32 {
	return int32(x * (math.Pow(2, float64(exp))))
}

// see variationsUsage
func (opts *fontOptionsT) parseVariations(s string) error {
	// remove possible quote
	s = strings.Trim(s, `"`)

	variations := strings.Split(s, ",")
	opts.variations = make([]Variation, len(variations))

	var err error
	for i, feature := range variations {
		opts.variations[i], err = ParseVariation(feature)
		if err != nil {
			return err
		}
	}
	return nil
}

type textOptionsT struct {
	textBefore, textAfter string
	text                  []rune
}

func (opts *textOptionsT) parseUnicodes(s string) error {
	runes := strings.Split(s, ",")
	opts.text = make([]rune, len(runes))
	for i, r := range runes {
		if _, err := fmt.Sscanf(r, "U+%x", &opts.text[i]); err == nil {
			continue
		}
		if _, err := fmt.Sscanf(r, "0x%x", &opts.text[i]); err == nil {
			continue
		}
		return fmt.Errorf("invalid unicode rune : %s", r)
	}
	return nil
}

type mainFontTextT struct {
	consumer shapeConsumerT
	input    textOptionsT
	fontOpts fontOptionsT
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

func (so *shapeOptionsT) populateBuffer(buffer *Buffer, text []rune, textBefore, textAfter string) {
	clearBufferContents(buffer)

	if textBefore != "" {
		t := []rune(textBefore)
		buffer.AddRunes(t, len(t), 0)
	}
	buffer.AddRunes(text, 0, len(text))
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

		if debugMode >= 1 {
			fmt.Println()
			fmt.Printf("VERIFY SAFE TO BREAK : start %d end %d text start %d end %d\n", start, end, textStart, textEnd)
			fmt.Println()
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

func (opts *shapeOptionsT) parseDirection(s string) error {
	switch toLower(s[0]) {
	case 'l':
		opts.props.Direction = LeftToRight
	case 'r':
		opts.props.Direction = RightToLeft
	case 't':
		opts.props.Direction = TopToBottom
	case 'b':
		opts.props.Direction = BottomToTop
	default:
		return fmt.Errorf("invalid direction %s", s)
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

// returns the serialized shaped output
func (sh *shapeConsumerT) consumeLine(text []rune, textBefore, textAfter string) (string, error) {
	sh.output.newLine()
	for n := sh.shaper.numIterations; n != 0; n-- {
		sh.shaper.populateBuffer(sh.buffer, text, textBefore, textAfter)
		if n == 1 {
			sh.output.consumeText(sh.buffer, text)
		}
		if err := sh.shaper.shape(sh.font, sh.buffer); err != nil {
			return "", err
		}
	}

	return sh.output.serializeShapeOutput(sh.buffer), nil
}

func (mft *mainFontTextT) main(out io.Writer) (string, error) {
	mft.consumer.output.out = out

	buffer := NewBuffer()
	mft.consumer.init(buffer, mft.fontOpts)

	return mft.consumer.consumeLine(mft.input.text, mft.input.textBefore, mft.input.textAfter)
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

func (opts *shapeOptionsT) parseFeatures(s string) error {
	// remove possible quote
	s = strings.Trim(s, `"`)

	features := strings.Split(s, ",")
	opts.features = make([]Feature, len(features))

	var err error
	for i, feature := range features {
		opts.features[i], err = parseFeature(feature)
		if err != nil {
			return err
		}
	}
	return nil
}

func (opts *fontOptionsT) parseFontSize(arg string) error {
	if arg == "upem" {
		opts.fontSizeY = fontSizeUpem
		opts.fontSizeX = fontSizeUpem
		return nil
	}
	n, err := fmt.Sscanf(arg, "%f %f", &opts.fontSizeX, &opts.fontSizeY)
	if err != io.EOF {
		return fmt.Errorf("font-size argument should be one or two space-separated numbers")
	}
	if n == 1 {
		opts.fontSizeY = opts.fontSizeX
	}
	return nil
}

func (opts *fontOptionsT) parseFontPpem(arg string) error {
	n, err := fmt.Sscanf(arg, "%d %d", &opts.xPpem, &opts.yPpem)
	if err != io.EOF {
		return fmt.Errorf("font-ppem argument should be one or two space-separated integers")
	}
	if n == 1 {
		opts.yPpem = opts.xPpem
	}
	return nil
}

const variationsUsage = `Comma-separated list of font variations

    Variations are set globally. The format for specifying variation settings
    follows.  All valid CSS font-variation-settings values other than 'normal'
    and 'inherited' are also accepted, although not documented below.

    The format is a tag, optionally followed by an equals sign, followed by a
    number. For example:

      "wght=500"
      "slnt=-7.5";
`

// parse the options, written in command line format
func parseOptions(options string) mainFontTextT {
	flags := flag.NewFlagSet("options", flag.ContinueOnError)

	var fmtOpts formatOptionsT
	flags.BoolVar(&fmtOpts.hideClusters, "no-clusters", false, "Do not output cluster indices")
	flags.BoolVar(&fmtOpts.hideGlyphNames, "no-glyph-names", false, "Output glyph indices instead of names")
	flags.BoolVar(&fmtOpts.hidePositions, "no-positions", false, "Do not output glyph positions")
	flags.BoolVar(&fmtOpts.hideAdvances, "no-advances", false, "Do not output glyph advances")
	flags.BoolVar(&fmtOpts.showExtents, "show-extents", false, "Output glyph extents")
	flags.BoolVar(&fmtOpts.showFlags, "show-flags", false, "Output glyph flags")

	ned := flags.Bool("ned", false, "No Extra Data; Do not output clusters or advances")

	var shapeOpts shapeOptionsT
	flags.Func("features", featuresUsage, shapeOpts.parseFeatures)
	flags.String("list-shapers", "", "(ignored)")
	flags.String("shaper", "", "(ignored)")
	flags.String("shapers", "", "(ignored)")
	flags.Func("direction", "Set text direction (default: auto)", shapeOpts.parseDirection)
	flags.Func("language", "Set text language (default: $LANG)", func(s string) error {
		shapeOpts.props.Language = NewLanguage(s)
		return nil
	})
	flags.Func("script", "Set text script, as an ISO-15924 tag (default: auto)", func(s string) error {
		var err error
		shapeOpts.props.Script, err = language.ParseScript(s)
		return err
	})
	flags.BoolVar(&shapeOpts.removeDefaultIgnorables, "remove-default-ignorables", false, "Remove Default-Ignorable characters")
	flags.BoolVar(&shapeOpts.preserveDefaultIgnorables, "preserve-default-ignorables", false, "Preserve Default-Ignorable characters")
	flags.Func("cluster-level", "Cluster merging level (0/1/2, default: 0)", func(s string) error {
		l, err := strconv.Atoi(s)
		if err != nil {
			return fmt.Errorf("invalid cluster-level option: %s", err)
		}
		if l < 0 || l > 2 {
			return fmt.Errorf("invalid cluster-level option : %d", l)
		}
		shapeOpts.clusterLevel = ClusterLevel(l)
		return nil
	})
	flags.IntVar(&shapeOpts.numIterations, "num-iterations", 1, "Run shaper N times (default: 1)")
	// {"bot",		0, 0, G_OPTION_ARG_NONE,	&this->bot,			"Treat text as beginning-of-paragraph",	nullptr},
	// {"eot",		0, 0, G_OPTION_ARG_NONE,	&this->eot,			"Treat text as end-of-paragraph",	nullptr},
	// {"invisible-glyph",	0, 0, G_OPTION_ARG_INT,		&this->invisible_glyph,		"Glyph value to replace Default-Ignorables with",	nullptr},
	// {"utf8-clusters",	0, 0, G_OPTION_ARG_NONE,	&this->utf8_clusters,		"Use UTF8 byte indices, not char indices",	nullptr},
	// {"normalize-glyphs",0, 0, G_OPTION_ARG_NONE,	&this->normalize_glyphs,	"Rearrange glyph clusters in nominal order",	nullptr},
	// {"verify",		0, 0, G_OPTION_ARG_NONE,	&this->verify,			"Perform sanity checks on shaping results",	nullptr},

	fontOpts := newFontOptions(fontSizeUpem, 0)

	flags.StringVar(&fontOpts.fontFile, "font-file", "", "Set font file-name")
	flags.IntVar(&fontOpts.fontIndex, "face-index", 0, "Set face index (default: 0)")
	flags.Func("font-size", "Font size", fontOpts.parseFontSize)
	flags.Func("font-ppem", "Set x,y pixels per EM (default: 0; disabled)", fontOpts.parseFontPpem)
	flags.Float64Var(&fontOpts.ptem, "font-ptem", 0, "Set font point-size (default: 0; disabled)")
	flags.Func("variations", variationsUsage, fontOpts.parseVariations)
	flags.String("font-funcs", "", "(ignored)")
	flags.String("ft-load-flags", "", "(ignored)")

	err := flags.Parse(strings.Split(options, " "))
	check(err)

	if *ned {
		fmtOpts.hideClusters = true
		fmtOpts.hideAdvances = true
	}
	return mainFontTextT{
		fontOpts: fontOpts,
		consumer: shapeConsumerT{
			shaper: shapeOpts,
			output: outputBufferT{
				format: fmtOpts,
			},
		},
	}
}

// harfbuzz seems to be OK with an invalid font
// in pratice, it seems useless to do shaping without
// font, so we dont support it, meaning we skip this test
func (fontOpts fontOptionsT) skipInvalidFontIndex() bool {
	f, err := os.Open(fontOpts.fontFile)
	check(err)

	fonts, err := truetype.Loader.Load(f)
	check(err)

	if fontOpts.fontIndex >= len(fonts) {
		fmt.Printf("skipping invalid font index %d in font %s\n", fontOpts.fontIndex, fontOpts.fontFile)
		return true
	}
	return false
}

// parses and run one test given as line in .tests files
func runOneShapingTest(t *testing.T, dir, line string) {
	chunks := strings.Split(line, ":")
	if len(chunks) != 4 {
		check(fmt.Errorf("invalid test file: line %s", line))
	}
	fontFileHash, options, unicodes, glyphsExpected := chunks[0], chunks[1], chunks[2], chunks[3]

	splitHash := strings.Split(fontFileHash, "@")
	fontFile := filepath.Join(dir, splitHash[0])
	if len(splitHash) >= 2 {
		ff, err := ioutil.ReadFile(fontFile)
		check(err)

		hash := sha1.Sum(ff)
		trimmedHash := strings.TrimSpace(string(hash[:]))
		if exp := splitHash[1]; trimmedHash != exp {
			check(fmt.Errorf("invalid font file hash: expected %s, got %s", exp, trimmedHash))
		}
	}

	verify := glyphsExpected != "*"

	var text textOptionsT
	text.parseUnicodes(unicodes)

	driver := parseOptions(options)
	driver.consumer.shaper.verify = false
	driver.input = text
	driver.fontOpts.fontFile = fontFile

	if driver.fontOpts.skipInvalidFontIndex() {
		return
	}

	// actual does the shaping
	output, err := driver.main(os.Stdout)
	if err != nil {
		t.Fatal("line ", line, ":", err)
	}

	if got := strings.TrimSpace(output); verify && glyphsExpected != got {
		t.Fatalf("for dir %s and line\n%s\n, expected :\n%s\n got \n%s", dir, line, glyphsExpected, got)
	}
}

// opens and parses a test file containing
// the font file, the unicode input and the expected result
func processHarfbuzzTestFile(t *testing.T, dir, filename string) {
	f, err := ioutil.ReadFile(filename)
	check(err)

	for _, line := range strings.Split(string(f), "\n") {
		if strings.HasPrefix(line, "#") || strings.TrimSpace(line) == "" { // skip comments
			continue
		}

		runOneShapingTest(t, dir, line)
	}
}

func dirFiles(t *testing.T, dir string) []string {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	var filenames []string
	for _, fi := range files {
		filenames = append(filenames, filepath.Join(dir, fi.Name()))
	}
	return filenames
}

func TestShapeExpected(t *testing.T) {
	for _, file := range dirFiles(t, "testdata/data/aots/tests") {
		processHarfbuzzTestFile(t, "testdata/data/aots/tests", file)
	}
	for _, file := range dirFiles(t, "testdata/data/in-house/tests") {
		if file == "testdata/data/in-house/tests/macos.tests" {
			// this requires fonts from the system
			continue
		}
		processHarfbuzzTestFile(t, "testdata/data/in-house/tests", file)
	}
	for _, file := range dirFiles(t, "testdata/data/text-rendering-tests/tests") {
		processHarfbuzzTestFile(t, "testdata/data/text-rendering-tests/tests", file)
	}
}

func TestDebug(t *testing.T) {
	runOneShapingTest(t, "testdata/data/in-house/tests",
		`../fonts/TTC.ttc:--face-index=0 --font-funcs=ot:U+2026,U+0020,U+002E:[ellipsis=0+723|space=1+250|period=2+241]`)
}
