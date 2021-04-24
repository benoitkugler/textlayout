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
)

type (
	GID = fonts.GID
	Tag = uint32
)
