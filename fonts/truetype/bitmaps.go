package truetype

type sizeMetrics struct {
	xPpem uint16 /* horizontal pixels per EM               */
	yPpem uint16 /* vertical pixels per EM                 */

	height uint16 /* text height       */
}

func (t bitmapTable) availableSizes() []sizeMetrics {
	out := make([]sizeMetrics, len(t))
	for i, size := range t {
		out[i] = size.sizeMetrics()
	}
	return out
}

func (t tableSbix) availableSizes(horizontal *TableHVhea, upem uint16) []sizeMetrics {
	out := make([]sizeMetrics, len(t.strikes))
	for i, size := range t.strikes {
		out[i] = size.sizeMetrics(horizontal, upem)
	}
	return out
}

// check for the various bitmaps table and returns
// the first valid
func (f *Font) loadBitmaps() []sizeMetrics {
	// adapted from freetype tt_face_load_sbit
	color, err := f.colorBitmapTable()
	if err == nil {
		return color.availableSizes()
	}

	gray, err := f.grayBitmapTable()
	if err == nil {
		return gray.availableSizes()
	}

	apple, err := f.appleBitmapTable()
	if err == nil {
		return apple.availableSizes()
	}

	sbix, err := f.sbixTable()
	if err == nil {
		hori, _ := f.HheaTable()
		if hori != nil {
			return sbix.availableSizes(hori, f.Head.UnitsPerEm)
		}
	}

	return nil
}
