package pango

import (
	"fmt"
	"strings"
	"sync"

	"github.com/benoitkugler/textlayout/fonts/truetype"
	"github.com/benoitkugler/textlayout/harfbuzz"
)

/* The cairo hexbox drawing code assumes
 * that these nicks are 1-6 ASCII chars */
var ignorables = []struct {
	nick string
	ch   rune
}{
	{"SHY", 0x00ad},   /* SOFT HYPHEN */
	{"CGJ", 0x034f},   /* COMBINING GRAPHEME JOINER */
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
	{"ZWNBS", 0xfeff}, /* ZERO WIDTH NO-BREAK SPACE */
}

func getIgnorable(ch rune) string {
	for _, ign := range ignorables {
		if ch == ign.ch {
			return ign.nick
		}
	}
	return ""
}

func (analysis *Analysis) findShowFlags() ShowFlags {
	var flags ShowFlags

	for _, attr := range analysis.extra_attrs {
		if attr.Type == ATTR_SHOW {
			flags |= ShowFlags(attr.Data.(AttrInt))
		}
	}

	return flags
}

func (analysis *Analysis) applyExtraAttributes() (out []harfbuzz.Feature) {
	for _, attr := range analysis.extra_attrs {
		if attr.Type == ATTR_FONT_FEATURES {
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
	for _, attr := range analysis.extra_attrs {
		if attr.Type == ATTR_LETTER_SPACING {
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
func (glyphs *GlyphString) pango_hb_shape(font Font, analysis *Analysis, paragraphText []rune,
	itemOffset, itemLength int) {

	showFlags := analysis.findShowFlags()
	hb_font := font.GetHBFont()
	fmt.Println(font.Describe(false).String())
	features := font.GetFeatures()
	features = append(features, analysis.applyExtraAttributes()...)

	dir := harfbuzz.LeftToRight
	if analysis.gravity.IsVertical() {
		dir = harfbuzz.TopToBottom
	}
	if analysis.level%2 != 0 {
		dir = dir.Reverse()
	}
	if analysis.gravity.IsImproper() {
		dir = dir.Reverse()
	}

	flags := harfbuzz.Bot | harfbuzz.Eot

	if showFlags&PANGO_SHOW_IGNORABLES != 0 {
		flags |= harfbuzz.PreserveDefaultIgnorables
	}

	/* setup buffer */
	cachedBufferLock.Lock()
	cachedBuffer.Clear()
	defer cachedBufferLock.Unlock()

	cachedBuffer.Props.Direction = dir
	cachedBuffer.Props.Script = analysis.script
	cachedBuffer.Props.Language = analysis.language
	cachedBuffer.ClusterLevel = harfbuzz.MonotoneCharacters
	cachedBuffer.Flags = flags
	cachedBuffer.Invisible = 0xFFFF // TODO: check that

	cachedBuffer.AddRunes(paragraphText, itemOffset, itemLength)
	if analysis.flags&PANGO_ANALYSIS_FLAG_NEED_HYPHEN != 0 {
		/* Insert either a Unicode or ASCII hyphen. We may
		 * want to look for script-specific hyphens here.  */

		if _, ok := hb_font.Face().NominalGlyph(0x2010); ok {
			cachedBuffer.AddRune(0x2010, itemOffset+itemLength-1)
		} else if _, ok := hb_font.Face().NominalGlyph('-'); ok {
			cachedBuffer.AddRune('-', itemOffset+itemLength-1)
		}
	}

	cachedBuffer.Shape(hb_font, features)

	if analysis.gravity.IsImproper() {
		cachedBuffer.Reverse()
	}

	/* buffer output */
	glyphInfos := cachedBuffer.Info
	glyphs.setSize(len(glyphInfos))
	infos := glyphs.Glyphs
	lastCluster := -1
	for i, inf := range glyphInfos {
		infos[i].glyph = Glyph(inf.Glyph)
		glyphs.logClusters[i] = inf.Cluster - itemOffset
		infos[i].attr.isClusterStart = glyphs.logClusters[i] != lastCluster
		lastCluster = glyphs.logClusters[i]
	}

	positions := cachedBuffer.Pos
	if analysis.gravity.IsVertical() {
		for i, pos := range positions {
			/* 90 degrees rotation counter-clockwise. */
			infos[i].Geometry.Width = GlyphUnit(pos.YAdvance)
			infos[i].Geometry.xOffset = GlyphUnit(pos.YOffset)
			infos[i].Geometry.yOffset = GlyphUnit(-pos.XOffset)
		}
	} else /* horizontal */ {
		for i, pos := range positions {
			infos[i].Geometry.Width = GlyphUnit(pos.XAdvance)
			infos[i].Geometry.xOffset = GlyphUnit(pos.XOffset)
			infos[i].Geometry.yOffset = GlyphUnit(-pos.YOffset)
		}
	}
}
