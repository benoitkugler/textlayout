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

type Position struct {
	x, y float32
}

func (p Position) add(other Position) Position {
	return Position{p.x + other.x, p.y + other.y}
}

// returns p - other
func (p Position) sub(other Position) Position {
	return Position{p.x - other.x, p.y - other.y}
}

func (p Position) scale(s float32) Position {
	return Position{p.x * s, p.y * s}
}

type rect struct {
	bl, tr Position
}

func (r rect) scale(s float32) rect {
	return rect{r.bl.scale(s), r.tr.scale(s)}
}

func (r rect) addPosition(pos Position) rect {
	return rect{r.bl.add(pos), r.tr.add(pos)}
}

func (r rect) widen(other rect) rect {
	out := r
	if r.bl.x > other.bl.x {
		out.bl.x = other.bl.x
	}
	if r.bl.y > other.bl.y {
		out.bl.y = other.bl.y
	}
	if r.tr.x < other.tr.x {
		out.tr.x = other.tr.x
	}
	if r.tr.y < other.tr.y {
		out.tr.y = other.tr.y
	}
	return out
}

const (
	kgmetLsb = iota
	kgmetRsb
	kgmetBbTop
	kgmetBbBottom
	kgmetBbLeft
	kgmetBbRight
	kgmetBbHeight
	kgmetBbWidth
	kgmetAdvWidth
	kgmetAdvHeight
	kgmetAscent
	kgmetDescent
)

type glyph struct {
	attrs   attributeSet
	advance struct{ x, y int16 }
	bbox    rect
}

func (g glyph) getMetric(metric uint8) int32 {
	switch metric {
	case kgmetLsb:
		return int32(g.bbox.bl.x)
	case kgmetRsb:
		return int32(g.advance.x) - int32(g.bbox.tr.x)
	case kgmetBbTop:
		return int32(g.bbox.tr.y)
	case kgmetBbBottom:
		return int32(g.bbox.bl.y)
	case kgmetBbLeft:
		return int32(g.bbox.bl.x)
	case kgmetBbRight:
		return int32(g.bbox.tr.x)
	case kgmetBbHeight:
		return int32(g.bbox.tr.y - g.bbox.bl.y)
	case kgmetBbWidth:
		return int32(g.bbox.tr.x - g.bbox.bl.x)
	case kgmetAdvWidth:
		return int32(g.advance.x)
	case kgmetAdvHeight:
		return int32(g.advance.y)
	default:
		return 0
	}
}

type graphiteFont struct {
	scale    float32 // scales from design units to ppm
	isHinted bool
}

type graphiteFace struct {
	cmap          truetype.Cmap
	htmx          truetype.TableHVmtx
	attrs         tableGlat
	silf          TableSilf
	sill          TableSill
	feat          TableFeat
	glyphs        truetype.TableGlyf
	numAttributes uint16 //  number of glyph attributes per glyph

	ascent, descent int32
}

func ParseFont(f fonts.Resource) (graphiteFace, error) {
	var out graphiteFace

	font, err := truetype.Parse(f)
	if err != nil {
		return out, err
	}

	out.cmap, _ = font.Cmap.BestEncoding()

	out.htmx, err = font.HtmxTable()
	if err != nil {
		return out, err
	}
	out.glyphs, err = font.GlyfTable()
	if err != nil {
		return out, err
	}

	ta, err := font.GetRawTable(tagSill)
	if err != nil {
		return out, err
	}
	out.sill, err = parseTableSill(ta)
	if err != nil {
		return out, err
	}

	ta, err = font.GetRawTable(tagFeat)
	if err != nil {
		return out, err
	}
	out.feat, err = parseTableFeat(ta)
	if err != nil {
		return out, err
	}

	ta, err = font.GetRawTable(tagGloc)
	if err != nil {
		return out, err
	}
	locations, numAttributes, err := parseTableGloc(ta, int(font.NumGlyphs))
	if err != nil {
		return out, err
	}

	out.numAttributes = numAttributes
	ta, err = font.GetRawTable(tagGlat)
	if err != nil {
		return out, err
	}
	out.attrs, err = parseTableGlat(ta, locations)
	if err != nil {
		return out, err
	}

	ta, err = font.GetRawTable(tagSilf)
	if err != nil {
		return out, err
	}
	out.silf, err = parseTableSilf(ta, numAttributes, uint16(len(out.feat)))
	if err != nil {
		return out, err
	}

	return out, nil
}

// getGlyph is safe to call with invalid gid
// it returns nil for pseudo glyph
func (f *graphiteFace) getGlyph(gid GID) *glyph {
	// TODO: preprocess these data when loading the face
	if int(gid) < len(f.glyphs) {
		data := f.glyphs[gid]

		return &glyph{
			advance: struct {
				x int16
				y int16
			}{x: f.htmx[gid].Advance}, attrs: f.attrs[gid],
			bbox: rect{
				bl: Position{float32(data.Xmin), float32(data.Ymin)},
				tr: Position{float32(data.Xmax), float32(data.Ymax)},
			},
		}
	}
	return nil
}

func (f *graphiteFace) getGlyphAttr(gid GID, attr uint16) int16 {
	if glyph := f.getGlyph(gid); glyph != nil {
		return glyph.attrs.get(attr)
	}
	return 0
}

func (f *graphiteFace) getGlyphMetric(gid GID, metric uint8) int32 {
	switch metric {
	case kgmetAscent:
		return f.ascent
	case kgmetDescent:
		return f.descent
	}
	if glyph := f.getGlyph(gid); glyph != nil {
		return glyph.getMetric(metric)
	}
	return 0
}

func (f *graphiteFace) runGraphite(seg *Segment, silf *silfSubtable) {
	if debugMode >= 1 {
		fmt.Printf("RUN graphite: segment %v, passes %v", seg, silf.passes)
	}

	// 	if seg.dir&3 == 3 && silf.IBidi == 0xFF {
	// 		seg.doMirror(silf.AttrMirroring)
	// 	}
	// 	res := silf.runGraphite(seg, 0, silf.positionPass(), true)
	// 	if res {
	// 		seg.associateChars(0, seg.charInfoCount())
	// 		if silf.flags() & 0x20 {
	// 			res &= seg.initCollisions()
	// 		}
	// 		if res {
	// 			res &= silf.runGraphite(seg, silf.positionPass(), silf.numPasses(), false)
	// 		}
	// 	}
}
