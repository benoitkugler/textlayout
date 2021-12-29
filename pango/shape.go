package pango

import (
	"log"
	"strings"
	"sync"
	"unicode"

	"github.com/benoitkugler/textlayout/fonts/truetype"
	"github.com/benoitkugler/textlayout/harfbuzz"
)

// These are the default ignorables that we render as hexboxes
// with nicks if PANGO_SHOW_IGNORABLES is used.
// The cairo hexbox drawing code assumes
// that these nicks are 1-6 ASCII chars
var ignorables = []struct {
	nick string
	ch   rune
}{
	{"SHY", 0x00ad},   /* SOFT HYPHEN */
	{"CGJ", 0x034f},   /* COMBINING GRAPHEME JOINER */
	{"ALM", 0x061c},   /* ARABIC LETTER MARK */
	{"ZWS", 0x200b},   /* ZERO WIDTH SPACE */
	{"ZWNJ", 0x200c},  /* ZERO WIDTH NON-JOINER */
	{"ZWJ", 0x200d},   /* ZERO WIDTH JOINER */
	{"LRM", 0x200e},   /* LEFT-TO-RIGHT MARK */
	{"RLM", 0x200f},   /* RIGHT-TO-LEFT MARK */
	{"LS", 0x2028},    /* LINE SEPARATOR */
	{"PS", 0x2029},    /* PARAGRAPH SEPARATOR */
	{"LRE", 0x202a},   /* LEFT-TO-RIGHT EMBEDDING */
	{"RLE", 0x202b},   /* RIGHT-TO-LEFT EMBEDDING */
	{"PDF", 0x202c},   /* POP DIRECTIONAL FORMATTING */
	{"LRO", 0x202d},   /* LEFT-TO-RIGHT OVERRIDE */
	{"RLO", 0x202e},   /* RIGHT-TO-LEFT OVERRIDE */
	{"WJ", 0x2060},    /* WORD JOINER */
	{"FA", 0x2061},    /* FUNCTION APPLICATION */
	{"IT", 0x2062},    /* INVISIBLE TIMES */
	{"IS", 0x2063},    /* INVISIBLE SEPARATOR */
	{"LRI", 0x2066},   /* LEFT-TO-RIGHT ISOLATE */
	{"RLI", 0x2067},   /* RIGHT-TO-LEFT ISOLATE */
	{"FSI", 0x2068},   /* FIRST STRONG ISOLATE */
	{"PDI", 0x2069},   /* POP DIRECTIONAL ISOLATE */
	{"ZWNBS", 0xfeff}, /* ZERO WIDTH NO-BREAK SPACE */
}

func getIgnorable(ch rune) string {
	for _, ign := range ignorables {
		if ch < ign.ch {
			return ""
		}

		if ch == ign.ch {
			return ign.nick
		}
	}
	return ""
}

func (analysis *Analysis) findShowFlags() ShowFlags {
	var flags ShowFlags

	for _, attr := range analysis.ExtraAttrs {
		if attr.Kind == ATTR_SHOW {
			flags |= ShowFlags(attr.Data.(AttrInt))
		}
	}

	return flags
}

func (analysis *Analysis) findTextTransform() TextTransform {
	transform := TEXT_TRANSFORM_NONE
	for _, attr := range analysis.ExtraAttrs {
		if attr.Kind == ATTR_TEXT_TRANSFORM {
			transform = TextTransform(attr.Data.(AttrInt))
		}
	}

	return transform
}

func (analysis *Analysis) collectFeatures() (out []harfbuzz.Feature) {
	out = analysis.Font.GetFeatures()

	for _, attr := range analysis.ExtraAttrs {
		if attr.Kind == ATTR_FONT_FEATURES {
			feats := strings.Split(string(attr.Data.(AttrString)), ",")
			for _, feat := range feats {
				ft, err := harfbuzz.ParseFeature(feat)
				if err != nil {
					continue
				}
				ft.Start = attr.StartIndex
				ft.End = attr.EndIndex
				out = append(out, ft)
			}
		}
	}

	/* Turn off ligatures when letterspacing */
	tags := [...]truetype.Tag{
		truetype.MustNewTag("liga"),
		truetype.MustNewTag("clig"),
		truetype.MustNewTag("dlig"),
	}
	for _, attr := range analysis.ExtraAttrs {
		if attr.Kind == ATTR_LETTER_SPACING {
			for _, tag := range tags {
				out = append(out, harfbuzz.Feature{
					Tag:   tag,
					Value: 0,
					Start: attr.StartIndex,
					End:   attr.EndIndex,
				})
			}
		}
	}
	return out
}

var (
	cachedBuffer     = harfbuzz.NewBuffer()
	cachedBufferLock sync.Mutex
)

// shape a subpart of the paragraph (starting at `itemOffset`, with length `itemLength`)
// and write the results into `glyphs`
func (glyphs *GlyphString) harfbuzzShape(analysis *Analysis, paragraphText []rune,
	itemOffset, itemLength int, logAttrs []CharAttr, numChars int) {

	showFlags := analysis.findShowFlags()
	transform := analysis.findTextTransform()
	font := analysis.Font
	hbFont := font.GetHarfbuzzFont()

	dir := harfbuzz.LeftToRight
	if analysis.Gravity.IsVertical() {
		dir = harfbuzz.TopToBottom
	}
	if analysis.Level%2 != 0 {
		dir = dir.Reverse()
	}
	if analysis.Gravity.IsImproper() {
		dir = dir.Reverse()
	}

	flags := harfbuzz.Bot | harfbuzz.Eot

	if showFlags&SHOW_IGNORABLES != 0 {
		flags |= harfbuzz.PreserveDefaultIgnorables
	}

	/* setup buffer */
	cachedBufferLock.Lock()
	cachedBuffer.Clear()
	defer cachedBufferLock.Unlock()

	cachedBuffer.Props.Direction = dir
	cachedBuffer.Props.Script = analysis.Script
	cachedBuffer.Props.Language = analysis.Language
	cachedBuffer.ClusterLevel = harfbuzz.MonotoneCharacters
	cachedBuffer.Flags = flags

	hyphenIndex := itemOffset + itemLength - 1
	if analysis.Flags&AFNeedHyphen != 0 {
		if logAttrs[numChars].IsBreakRemovesPreceding() {
			itemLength -= 1
		}
	}

	// Add pre-context
	cachedBuffer.AddRunes(paragraphText, itemOffset, 0)

	if transform == TEXT_TRANSFORM_NONE {
		cachedBuffer.AddRunes(paragraphText, itemOffset, itemLength)
	} else {
		for i, ch := range paragraphText[itemOffset : itemOffset+itemLength] {
			switch transform {
			case TEXT_TRANSFORM_LOWERCASE:
				ch = unicode.ToLower(ch)
			case TEXT_TRANSFORM_UPPERCASE:
				ch = unicode.ToUpper(ch)
			case TEXT_TRANSFORM_CAPITALIZE:
				if logAttrs[i].IsWordStart() {
					ch = unicode.ToTitle(ch)
				}
			}
			cachedBuffer.AddRune(ch, i)
		}
	}

	// Add post-context
	cachedBuffer.AddRunes(paragraphText, itemOffset+itemLength, 0)

	if analysis.Flags&AFNeedHyphen != 0 {
		/* Insert either a Unicode or ASCII hyphen. We may
		* want to look for script-specific hyphens here.  */

		if _, ok := hbFont.Face().NominalGlyph(0x2010); ok {
			cachedBuffer.AddRune(0x2010, hyphenIndex)
		} else if _, ok := hbFont.Face().NominalGlyph('-'); ok {
			cachedBuffer.AddRune('-', hyphenIndex)
		}
	}

	features := analysis.collectFeatures()

	cachedBuffer.Shape(hbFont, features)

	if analysis.Gravity.IsImproper() {
		cachedBuffer.Reverse()
	}

	/* buffer output */
	glyphInfos := cachedBuffer.Info
	glyphs.setSize(len(glyphInfos))
	infos := glyphs.Glyphs
	lastCluster := -1
	for i, inf := range glyphInfos {
		infos[i].Glyph = Glyph(inf.Glyph)
		glyphs.LogClusters[i] = inf.Cluster - itemOffset
		infos[i].attr.isClusterStart = glyphs.LogClusters[i] != lastCluster
		lastCluster = glyphs.LogClusters[i]
	}

	positions := cachedBuffer.Pos
	if analysis.Gravity.IsVertical() {
		for i, pos := range positions {
			/* 90 degrees rotation counter-clockwise. */
			infos[i].Geometry.Width = Unit(-pos.YAdvance)
			infos[i].Geometry.XOffset = Unit(-pos.YOffset)
			infos[i].Geometry.YOffset = Unit(-pos.XOffset)
		}
	} else /* horizontal */ {
		for i, pos := range positions {
			infos[i].Geometry.Width = Unit(pos.XAdvance)
			infos[i].Geometry.XOffset = Unit(pos.XOffset)
			infos[i].Geometry.YOffset = Unit(-pos.YOffset)
		}
	}
}

// Shape converts the characters in `item` into glyphs.
//
// This is similar to [func@Pango.shape_with_flags], except it takes a
// `PangoItem` instead of separate `item_text` and @analysis arguments.
// It takes `logAttrs`, which may be used in implementing text
// transforms.
func (item *Item) Shape(paragraphText []rune, logAttrs []CharAttr, glyphs *GlyphString, flags shapeFlags) {
	glyphs.shapeInternal(paragraphText, item.Offset, item.Length, &item.Analysis, logAttrs, item.Length, flags)
}

// shapeInternal is similar to shapeRange(), except it also takes
// flags that can influence the shaping process.
func (glyphs *GlyphString) shapeInternal(paragraphText []rune, itemOffset, itemLength int, analysis *Analysis,
	logAttrs []CharAttr, numChars int, flags shapeFlags) {

	itemText := paragraphText[itemOffset : itemOffset+itemLength]

	if analysis.Font != nil {
		glyphs.harfbuzzShape(analysis, paragraphText, itemOffset, itemLength, logAttrs, numChars)

		if len(glyphs.Glyphs) == 0 {
			if debugMode {
				// If a font has been correctly chosen, but no glyphs are output,
				// there's probably something wrong with the font.
				log.Printf("shaping failure, expect ugly output. font='%s', text='%s' : %v",
					analysis.Font.Describe(false), string(itemText), itemText)
			}
		}
	}

	if len(glyphs.Glyphs) == 0 {
		glyphs.fallbackShape(itemText, analysis)
		if len(glyphs.Glyphs) == 0 {
			return
		}
	}

	// make sure last_cluster is invalid
	lastCluster := glyphs.LogClusters[0] - 1
	for i, lo := range glyphs.LogClusters {
		// Set glyphs[i].attr.is_cluster_start based on logClusters[]
		if lo != lastCluster {
			glyphs.Glyphs[i].attr.isClusterStart = true
			lastCluster = lo
		} else {
			glyphs.Glyphs[i].attr.isClusterStart = false
		}

		// Shift glyph if width is negative, and negate width.
		// This is useful for rotated font matrices and shouldn't harm in normal cases.
		if glyphs.Glyphs[i].Geometry.Width < 0 {
			glyphs.Glyphs[i].Geometry.Width = -glyphs.Glyphs[i].Geometry.Width
			glyphs.Glyphs[i].Geometry.XOffset += glyphs.Glyphs[i].Geometry.Width
		}
	}

	// Make sure glyphstring direction conforms to analysis.level
	if lc := glyphs.LogClusters; (analysis.Level&1) != 0 && lc[0] < lc[len(lc)-1] {
		log.Println("pango: expected RTL run but got LTR. Fixing.")

		// *Fix* it so we don't crash later
		glyphs.reverse()
	}

	if flags&shapeROUND_POSITIONS != 0 {
		for i := range glyphs.Glyphs {
			glyphs.Glyphs[i].Geometry.Width = glyphs.Glyphs[i].Geometry.Width.Round()
			glyphs.Glyphs[i].Geometry.XOffset = glyphs.Glyphs[i].Geometry.XOffset.Round()
			glyphs.Glyphs[i].Geometry.YOffset = glyphs.Glyphs[i].Geometry.YOffset.Round()
		}
	}
}
