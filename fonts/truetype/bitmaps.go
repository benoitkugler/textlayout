package truetype

import (
	"github.com/benoitkugler/textlayout/fonts"
)

func (t bitmapTable) availableSizes(avgWidth, upem uint16) []fonts.BitmapSize {
	out := make([]fonts.BitmapSize, 0, len(t))
	for _, size := range t {
		v := size.sizeMetrics(avgWidth, upem)
		/* only use strikes with valid PPEM values */
		if v.XPpem == 0 || v.YPpem == 0 {
			continue
		}
		out = append(out, v)
	}
	return out
}

func (t tableSbix) availableSizes(horizontal *TableHVhea, avgWidth, upem uint16) []fonts.BitmapSize {
	out := make([]fonts.BitmapSize, 0, len(t.strikes))
	for _, size := range t.strikes {
		v := size.sizeMetrics(horizontal, avgWidth, upem)
		/* only use strikes with valid PPEM values */
		if v.XPpem == 0 || v.YPpem == 0 {
			continue
		}
		out = append(out, v)
	}
	return out
}

func inferBitmapWidth(size *fonts.BitmapSize, avgWidth, upem uint16) {
	size.Width = uint16((uint32(avgWidth)*uint32(size.XPpem) + uint32(upem/2)) / uint32(upem))
}

// LoadBitmaps checks for the various bitmaps table and returns
// the first valid
func (f *Font) LoadBitmaps() []fonts.BitmapSize {
	upem := f.Head.UnitsPerEm
	var avgWidth uint16
	os2, _ := f.OS2Table()
	if os2 != nil {
		avgWidth = os2.XAvgCharWidth
	}

	if upem == 0 || os2.Version == 0xFFFF {
		avgWidth = 1
		upem = 1
	}

	// adapted from freetype tt_face_load_sbit
	color, err := f.colorBitmapTable()
	if err == nil {
		return color.availableSizes(avgWidth, upem)
	}

	gray, err := f.grayBitmapTable()
	if err == nil {
		return gray.availableSizes(avgWidth, upem)
	}

	apple, err := f.appleBitmapTable()
	if err == nil {
		return apple.availableSizes(avgWidth, upem)
	}

	sbix, err := f.sbixTable()
	if err == nil {
		hori, _ := f.HheaTable()
		if hori != nil {
			return sbix.availableSizes(hori, avgWidth, upem)
		}
	}

	return nil
}
