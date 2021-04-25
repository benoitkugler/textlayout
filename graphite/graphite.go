// Package graphite Graphite implements a "smart font" system developed
// specifically to handle the complexities of lesser-known languages of the world.
package graphite

import (
	"github.com/benoitkugler/textlayout/fonts"
	"github.com/benoitkugler/textlayout/fonts/truetype"
)

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

type graphiteFace struct {
	attrs     tableGlat
	silf      TableSilf
	sill      TableSill
	feat      TableFeat
	numGlyphs uint16
}
