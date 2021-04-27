// Package graphite Graphite implements a "smart font" system developed
// specifically to handle the complexities of lesser-known languages of the world.
package graphite

import (
	"fmt"

	"github.com/benoitkugler/textlayout/fonts"
	"github.com/benoitkugler/textlayout/fonts/truetype"
)

const debugMode = 0

// graphite
var (
	tagSilf = truetype.MustNewTag("Silf")
	tagSill = truetype.MustNewTag("Sill")
	tagFeat = truetype.MustNewTag("Feat")
	tagGloc = truetype.MustNewTag("Gloc")
	tagGlat = truetype.MustNewTag("Glat")
)

type (
	GID = fonts.GID
	Tag = uint32
)

type position struct {
	x, y float32
}

type glyph struct {
	attrs   attributeSet
	advance int16 // horizontal
}

type graphiteFace struct {
	cmap      truetype.Cmap
	htmx      truetype.TableHVmtx
	attrs     tableGlat
	silf      TableSilf
	sill      TableSill
	feat      TableFeat
	numGlyphs GID
}

// getGlyph is safe to call with invalid gid
// it returns nil for pseudo glyph
func (f *graphiteFace) getGlyph(gid GID) *glyph {
	if gid < f.numGlyphs {
		return &glyph{advance: f.htmx[gid].Advance, attrs: f.attrs[gid]}
	}
	return nil
}

func (f *graphiteFace) getGlyphAttr(gid GID, attr uint16) int16 {
	if glyph := f.getGlyph(gid); glyph != nil {
		return glyph.attrs.get(attr)
	}
	return 0
}

func (f *graphiteFace) runGraphite(seg *segment, silf *silfSubtable) {
	if debugMode >= 1 {
		fmt.Printf("RUN graphite: segment %v, passes %v", seg, silf.passes)
	}

	if seg.dir&3 == 3 && silf.IBidi == 0xFF {
		seg.doMirror(silf.AttrMirroring)
	}
	res := silf.runGraphite(seg, 0, silf.positionPass(), true)
	if res {
		seg.associateChars(0, seg.charInfoCount())
		if silf.flags() & 0x20 {
			res &= seg.initCollisions()
		}
		if res {
			res &= silf.runGraphite(seg, silf.positionPass(), silf.numPasses(), false)
		}
	}
}
