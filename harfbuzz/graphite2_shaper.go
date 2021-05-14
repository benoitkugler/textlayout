package harfbuzz

import (
	"strings"

	"github.com/benoitkugler/textlayout/fonts"
	"github.com/benoitkugler/textlayout/fonts/truetype"
	"github.com/benoitkugler/textlayout/graphite"
)

// ported from harfbuzz/src/hb-graphite2.cc
// Copyright © 2011  Martin Hosken
// Copyright © 2011  SIL International
// Copyright © 2011,2012  Google, Inc.  Behdad Esfahbod

var _ shaper = (*shaperGraphite)(nil)

type graphite2Cluster struct {
	baseChar  int
	numChars  int
	baseGlyph int
	numGlyphs int
	cluster   int
	advance   float32
}

// shaperGraphite implements a shaper using Graphite features.
type shaperGraphite graphite.GraphiteFace

func (shaperGraphite) kind() shaperKind { return skGraphite }

func (shaperGraphite) compile(props SegmentProperties, userFeatures []Feature) {
}

// Converts a string into a Tag. Valid tags
// are four characters. Shorter input strings will be
// padded with spaces. Longer input strings will be
// truncated.
// The empty string is mapped to 0.
func tagFromString(str string) truetype.Tag {
	if str == "" {
		return 0
	}
	var chars [4]byte

	if len(str) > 4 {
		str = str[:4]
	}
	copy(chars[:], str)
	for i := len(str); i < 4; i++ {
		chars[i] = ' '
	}

	return newTag(chars[0], chars[1], chars[2], chars[3])
}

// TODO:
func (sh *shaperGraphite) shape(font *Font, buffer *Buffer, features []Feature) {
	// face := font.face
	grface := (*graphite.GraphiteFace)(sh)

	lang := languageToString(buffer.Props.Language)
	var tagLang truetype.Tag
	if lang != "" {
		tagLang = tagFromString(strings.Split(lang, "-")[0])
	}
	feats := grface.FeaturesForLang(tagLang)

	for _, feature := range features {
		if fref := feats.FindFeature(feature.Tag); fref != nil {
			fref.Value = int16(feature.Value)
		}
	}

	//    gr_segment *seg = nil;
	//    const gr_slot *is;
	//    uint ci = 0, ic = 0;
	//    uint curradvx = 0, curradvy = 0;

	chars := make([]rune, len(buffer.Info)) // TODO: can we avoid the allocation here ?
	for i, info := range buffer.Info {
		chars[i] = info.codepoint
	}

	/* TODO ensure_native_direction. */

	scriptTag, _ := otTagsFromScriptAndLanguage(buffer.Props.Script, "")
	tagScript := tagDefaultScript
	if len(scriptTag) != 0 {
		tagScript = scriptTag[len(scriptTag)-1]
	}
	dirMask := 2 | 0
	if buffer.Props.Direction == RightToLeft {
		dirMask = 2 | 1
	}
	seg := grface.Shape(nil, chars, tagScript, feats, dirMask)

	glyphCount := seg.NumGlyphs
	if glyphCount == 0 {
		buffer.Clear()
		return
	}

	// buffer.ensure(glyphCount) // FIXME
	//    scratch = buffer.get_scratch_buffer ();
	//    for ((DIV_CEIL (sizeof (hb_graphite2_cluster_t) * buffer.len, sizeof (*scratch)) +
	// 	   DIV_CEIL (sizeof (hb_codepoint_t) * glyphCount, sizeof (*scratch))) > scratch_size)
	//    {
	// 	 if (unlikely (!buffer.ensure (buffer.allocated * 2)))
	// 	 {
	// 	   if (feats) gr_featureval_destroy (feats);
	// 	   gr_seg_destroy (seg);
	// 	   return false;
	// 	 }
	// 	 scratch = buffer.get_scratch_buffer (&scratch_size);
	//    }

	//  #define ALLOCATE_ARRAY(Type, name, len) \
	//    Type *name = (Type *) scratch; \
	//    do { \
	// 	 uint _consumed = DIV_CEIL ((len) * sizeof (Type), sizeof (*scratch)); \
	// 	 assert (_consumed <= scratch_size); \
	// 	 scratch += _consumed; \
	// 	 scratch_size -= _consumed; \
	//    } while (0)

	//    ALLOCATE_ARRAY (hb_graphite2_cluster_t, clusters, buffer.len);
	//    ALLOCATE_ARRAY (hb_codepoint_t, gids, glyphCount);

	//  #undef ALLOCATE_ARRAY

	var clusters []graphite2Cluster // FIXME
	// memset(clusters, 0, sizeof(clusters[0])*buffer.len) // FIXME
	pg := make([]fonts.GID, 0, glyphCount)

	clusters[0].cluster = buffer.Info[0].Cluster
	upem := font.faceUpem
	xscale := float32(font.XScale / upem)
	yscale := float32(font.YScale / upem)
	yscale *= yscale / xscale
	var curradv float32
	if buffer.Props.Direction.isBackward() {
		curradv = seg.First.Position.X * xscale
		clusters[0].advance = seg.Advance.X*xscale - curradv
	} else {
		clusters[0].advance = 0
	}
	var ci int
	for is, ic := seg.First, 0; is != nil; is, ic = is.Next, ic+1 {
		before := is.Before
		after := is.After
		pg = append(pg, is.GlyphID)
		for clusters[ci].baseChar > before && ci != 0 {
			clusters[ci-1].numChars += clusters[ci].numChars
			clusters[ci-1].numGlyphs += clusters[ci].numGlyphs
			clusters[ci-1].advance += clusters[ci].advance
			ci--
		}

		if is.CanInsertBefore() && clusters[ci].numChars != 0 && before >= clusters[ci].baseChar+clusters[ci].numChars {
			c := &clusters[ci+1]
			c.baseChar = clusters[ci].baseChar + clusters[ci].numChars
			c.cluster = buffer.Info[c.baseChar].Cluster
			c.numChars = before - c.baseChar
			c.baseGlyph = ic
			c.numGlyphs = 0
			if buffer.Props.Direction.isBackward() {
				c.advance = curradv - is.Position.X*xscale
				curradv -= c.advance
			} else {
				c.advance = 0
				clusters[ci].advance += is.Position.X*xscale - curradv
				curradv += clusters[ci].advance
			}
			ci++
		}
		clusters[ci].numGlyphs++

		if clusters[ci].baseChar+clusters[ci].numChars < after+1 {
			clusters[ci].numChars = after + 1 - clusters[ci].baseChar
		}
	}

	if buffer.Props.Direction.isBackward() {
		clusters[ci].advance += curradv
	} else {
		clusters[ci].advance += seg.Advance.X*xscale - curradv
	}
	ci++

	for i := 0; i < ci; i++ {
		for j := 0; j < clusters[i].numGlyphs; j++ {
			info := &buffer.Info[clusters[i].baseGlyph+j]
			info.Glyph = pg[clusters[i].baseGlyph+j]
			info.Cluster = clusters[i].cluster
			info.setInt32(int32(clusters[i].advance)) // all glyphs in the cluster get the same advance
		}
	}
	// buffer.len = glyphCount // FIXME

	/* Positioning. */
	currclus := maxInt
	info := buffer.Info
	pPos := buffer.Pos
	if !buffer.Props.Direction.isBackward() {
		var curradvx, curradvy int32
		for is, index := seg.First, 0; is != nil; index, is = index+1, is.Next {
			pPos := &pPos[index]
			info := &info[index]
			pPos.XOffset = int32(is.Position.X*xscale) - curradvx
			pPos.YOffset = int32(is.Position.Y*yscale) - curradvy
			if info.Cluster != currclus {
				pPos.XAdvance = info.getInt32()
				curradvx += pPos.XAdvance
				currclus = info.Cluster
			} else {
				pPos.XAdvance = 0.
			}

			pPos.YAdvance = int32(is.Advance.Y * yscale)
			curradvy += pPos.YAdvance
		}
	} else {
		curradvx := int32(seg.Advance.X * xscale)
		var curradvy int32
		for is, index := seg.First, 0; is != nil; index, is = index+1, is.Next {
			pPos := &pPos[index]
			info := &info[index]
			if info.Cluster != currclus {
				pPos.XAdvance = info.getInt32()
				curradvx -= pPos.XAdvance
				currclus = info.Cluster
			} else {
				pPos.XAdvance = 0.
			}

			pPos.YAdvance = int32(is.Advance.Y * yscale)
			curradvy -= pPos.YAdvance
			pPos.XOffset = int32(is.Position.X*xscale) - info.getInt32() - curradvx + pPos.XAdvance
			pPos.YOffset = int32(is.Position.Y*yscale) - curradvy
		}
		buffer.reverseClusters()
	}

	buffer.unsafeToBreakAll()
}
