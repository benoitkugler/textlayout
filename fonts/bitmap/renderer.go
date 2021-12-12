package bitmap

import (
	"github.com/benoitkugler/textlayout/fonts"
)

var _ fonts.FaceRenderer = (*Font)(nil)

func (f *Font) GlyphData(gid fonts.GID, xPpem, yPpem uint16) fonts.GlyphData {
	if int(gid) > len(f.bitmap.offsets) {
		return nil
	}

	start := f.bitmap.offsets[gid]
	end := len(f.bitmap.data)
	if v := int(gid + 1); v != len(f.bitmap.offsets) {
		end = int(f.bitmap.offsets[v])
	}

	met := f.metrics[gid]
	width := int(met.rightSideBearing - met.leftSideBearing)
	height := int(met.characterAscent + met.characterDescent)

	// TODO: expose the padding options

	out := fonts.GlyphBitmap{
		Data:   f.bitmap.data[start:end],
		Format: fonts.BlackAndWhite,
		Width:  width,
		Height: height,
	}

	return out
}
