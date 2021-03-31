package harfbuzz

import (
	"fmt"
	"io"
	"strings"

	"github.com/benoitkugler/textlayout/fonts"
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
	showGlyphNames bool
	showPositions  bool
	showAdvances   bool
	showClusters   bool
	showText       bool
	showUnicode    bool
	showLineNum    bool
	showExtents    bool
	showFlags      bool
	trace          bool
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

type outputBufferT struct {
	out         io.Writer
	format      formatOptionsT
	formatFlags int
	lineNo      int
	font        *Font
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
	//  options.get_file_handle ();
	//  gs = g_string_new (nullptr);
	//  lineNo = 0;
	//  font = hb_font_reference (fontOpts.get_font ());
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
	if !out.format.showGlyphNames {
		flags |= HB_BUFFER_SERIALIZE_FLAG_NO_GLYPH_NAMES
	}
	if !out.format.showClusters {
		flags |= HB_BUFFER_SERIALIZE_FLAG_NO_CLUSTERS
	}
	if !out.format.showPositions {
		flags |= HB_BUFFER_SERIALIZE_FLAG_NO_POSITIONS
	}
	if !out.format.showAdvances {
		flags |= HB_BUFFER_SERIALIZE_FLAG_NO_ADVANCES
	}
	if out.format.showExtents {
		flags |= HB_BUFFER_SERIALIZE_FLAG_GLYPH_EXTENTS
	}
	if out.format.showFlags {
		flags |= HB_BUFFER_SERIALIZE_FLAG_GLYPH_FLAGS
	}
	out.formatFlags = flags
	if out.format.trace {
		// hb_buffer_set_message_func(buffer, message_func, this, nullptr)
	}
}

func (out *outputBufferT) newLine() { out.lineNo++ }

func (out *outputBufferT) consumeText(buffer *Buffer, text string) {
	s := out.format.serializeBufferOfText(buffer, out.lineNo, text, font)
	fmt.Fprintf(out.out, "%s", s)
}

func (out *outputBufferT) outputError(message string) {
	var gs strings.Builder
	out.format.serializeLineNo(out.lineNo, &gs)
	fmt.Fprintf(&gs, "error: %s", message)
	gs.WriteByte('\n')
	fmt.Fprintf(out.out, "%s", gs.String())
}

//    void consume_glyphs (buffer *Buffer,
// 				const char   *text,
// 				unsigned int  text_len,
// 				hb_bool_t     utf8Clusters)
//    {
// 	 g_string_set_size (gs, 0);
// 	 format.serialize_buffer_of_glyphs (buffer, lineNo, text, text_len, font,
// 						outputFormat, formatFlags, gs);
// 	 fprintf (options.fp, "%s", gs.str);
//    }

//    void finish (buffer *Buffer, const fontOptionsT *fontOpts)
//    {
// 	 hb_buffer_set_message_func (buffer, nullptr, nullptr, nullptr);
// 	 hb_font_destroy (font);
// 	 g_string_free (gs, true);
// 	 gs = nullptr;
// 	 font = nullptr;
//    }
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

type fontOptionsT struct { // char *font_file;
	// mutable hb_blob_t *blob;
	// int face_index;
	// hb_variation_t *variations;
	// unsigned int num_variations;
	defaultFontSize int
	// int x_ppem;
	// int y_ppem;
	// double ptem;
	subpixelBits int
	fontSizeX    float64
	fontSizeY    float64
	// char *font_funcs;
	// int ft_load_flags;
	// mutable hb_font_t *font;
}

func newFontOptions(defaultFontSize, subpixelBits int) fontOptionsT {
	return fontOptionsT{
		defaultFontSize: defaultFontSize,
		subpixelBits:    subpixelBits,
		fontSizeX:       float64(defaultFontSize),
		fontSizeY:       float64(defaultFontSize),
	}
}

type textOptionsT struct {
	lines                 []string
	textBefore, textAfter string

	// int text_len;
	// char *text;
	// char *text_file;

	// private:
	// FILE *fp;
	// GString *gs;
	// char *line;
	// unsigned int line_len;
}

// template <typename consumer_t, int defaultFontSize, int subpixelBits>
type mainFontTextT struct {
	// option_parser_t options; "[FONT-FILE] [TEXT]"
	fontOpts fontOptionsT
	input    textOptionsT
	consumer shapeConsumerT
}

func newFormatOptionsT() formatOptionsT {
	return formatOptionsT{
		showGlyphNames: true,
		showPositions:  true,
		showAdvances:   true,
		showClusters:   true,
		showText:       false,
		showUnicode:    false,
		showLineNum:    false,
		showExtents:    false,
		showFlags:      false,
		trace:          false,
	}
}

type shapeOptionsT struct {
	/* Buffer properties */
	props SegmentProperties

	/* Buffer flags */
	bot                       bool
	eot                       bool
	preserveDefaultIgnorables bool
	removeDefaultIgnorables   bool

	features []Feature
	//  char **shapers;
	//   utf8Clusters bool
	invisibleGlyph          fonts.GlyphIndex
	clusterLevel            ClusterLevel
	normalizeGlyphs, verify bool
	numIterations           int
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

	diff := hb_buffer_diff(reconstruction, buffer, (hb_codepoint_t)-1, 0)
	if diff {
		/* Return the reconstructed result instead so it can be inspected. */
		buffer.Info = nil
		buffer.Pos = nil
		appendBuffer(buffer, reconstruction, 0, len(reconstruction.Info))

		return fmt.Errorf("safe-to-break test failed: %s", diff)
	}

	return nil
}

type shapeConsumerT struct { // bool failed;
	shaper shapeOptionsT
	output outputBufferT
	// hb_font_t *font;
	buffer *Buffer
}

//   shapeConsumerT (option_parser_t *parser)
// 		  : failed (false),
// 		    shaper (parser),
// 		    output (parser),
// 		    font (nullptr),
// 		    buffer (nullptr) {}

func (sh *shapeConsumerT) init(buffer *Buffer, fontOpts fontOptionsT) {
	// font = hb_font_reference (fontOpts.get_font ());
	// failed = false;
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
			failed = true
			sh.output.outputError(err.Error())
			if hb_buffer_get_content_type(buffer) == HB_BUFFER_CONTENT_TYPE_GLYPHS {
				break
			} else {
				return
			}
		}
	}

	sh.output.consume_glyphs(buffer, text, text_len, shaper.utf8Clusters)
}

//   void finish (const font_options_t *fontOpts)
//   {
//     output.finish (buffer, fontOpts);
//     hb_font_destroy (font);
//     font = nullptr;
//     hb_buffer_destroy (buffer);
//     buffer = nullptr;
//   }

func newMainFontTextT(defaultFontSize, subpixelBits int) mainFontTextT {
	return mainFontTextT{
		fontOpts: newFontOptions(defaultFontSize, subpixelBits),
		// input(&options),
		// consumer(&options),
	}
}

func (mft mainFontTextT) main() {
	// options.parse (&argc, &argv);

	// argc--, argv++;
	// if (argc && !fontOpts.font_file) fontOpts.font_file = locale_to_utf8 (argv[0]), argc--, argv++;
	// if (argc && !input.text && !input.text_file) input.text = locale_to_utf8 (argv[0]), argc--, argv++;
	// if (argc)
	//   fail (true, "Too many arguments on the command line");
	// if (!fontOpts.font_file)
	//   options.usage ();
	// if (!input.text && !input.text_file)
	//   input.text_file = g_strdup ("-");

	buffer := NewBuffer()
	mft.consumer.init(buffer, mft.fontOpts)

	for _, line := range mft.input.lines {
		mft.consumer.consumeLine(line, mft.input.textBefore, mft.input.textAfter)
	}

	consumer.finish(&fontOpts)

	return
}

func shape(input []string) error {
	var driver mainFontTextT
	driver.main(input)
}
