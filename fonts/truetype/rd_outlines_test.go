package truetype

import (
	"reflect"
	"testing"

	testdata "github.com/benoitkugler/textlayout-testdata/truetype"
	"github.com/benoitkugler/textlayout/fonts"
	"golang.org/x/image/font/sfnt"
	fx "golang.org/x/image/math/fixed"
)

func pt(x, y float32) fonts.SegmentPoint {
	return fonts.SegmentPoint{X: x, Y: y}
}

func moveTo(xa, ya float32) fonts.Segment {
	return fonts.Segment{
		Op:   fonts.SegmentOpMoveTo,
		Args: [3]fonts.SegmentPoint{pt(xa, ya)},
	}
}

func lineTo(xa, ya float32) fonts.Segment {
	return fonts.Segment{
		Op:   fonts.SegmentOpLineTo,
		Args: [3]fonts.SegmentPoint{pt(xa, ya)},
	}
}

func quadTo(xa, ya, xb, yb float32) fonts.Segment {
	return fonts.Segment{
		Op:   fonts.SegmentOpQuadTo,
		Args: [3]fonts.SegmentPoint{pt(xa, ya), pt(xb, yb)},
	}
}

func cubeTo(xa, ya, xb, yb, xc, yc float32) fonts.Segment {
	return fonts.Segment{
		Op:   fonts.SegmentOpCubeTo,
		Args: [3]fonts.SegmentPoint{pt(xa, ya), pt(xb, yb), pt(xc, yc)},
	}
}

func translate(dx, dy float32, s fonts.Segment) fonts.Segment {
	args := s.ArgsSlice()
	for i := range args {
		args[i].X += dx
		args[i].Y += dy
	}
	return s
}

func transform_(txx, txy, tyx, tyy uint16, dx, dy float32, s fonts.Segment) fonts.Segment {
	args := s.ArgsSlice()
	for i := range args {
		args[i] = tform(txx, txy, tyx, tyy, dx, dy, args[i])
	}
	return s
}

// transformArgs applies an affine transformation to args. The t?? arguments
// are 2.14 fixed point values.
func tform(txx, txy, tyx, tyy uint16, dx, dy float32, p fonts.SegmentPoint) fonts.SegmentPoint {
	const half = 1 << 13
	return fonts.SegmentPoint{
		X: dx +
			p.X*fixed214ToFloat(txx) +
			p.Y*fixed214ToFloat(tyx),
		Y: dy +
			p.X*fixed214ToFloat(txy) +
			p.Y*fixed214ToFloat(tyy),
	}
}

// adapted from sfnt/sfnt_test.go
func TestGlyfSegments1(t *testing.T) {
	f := loadFont(t, "segments.ttf")

	// expecteds' vectors correspond 1-to-1 to what's in the glyfTest.sfd file,
	// although FontForge's SFD format stores quadratic Bézier curves as cubics
	// with duplicated off-curve points. quadTo(bx, by, cx, cy) is stored as
	// "bx by bx by cx cy".
	//
	// The .notdef, .null and nonmarkingreturn glyphs aren't explicitly in the
	// SFD file, but for some unknown reason, FontForge generates them in the
	// TrueType file.
	expecteds := [][]fonts.Segment{{
		// .notdef
		// - contour #0
		moveTo(68, 0),
		lineTo(68, 1365),
		lineTo(612, 1365),
		lineTo(612, 0),
		lineTo(68, 0),
		// - contour #1
		moveTo(136, 68),
		lineTo(544, 68),
		lineTo(544, 1297),
		lineTo(136, 1297),
		lineTo(136, 68),
	}, {
		// .null
		// Empty glyph.
	}, {
		// nonmarkingreturn
		// Empty glyph.
	}, {
		// zero
		// - contour #0
		moveTo(614, 1434),
		quadTo(369, 1434, 369, 614),
		quadTo(369, 471, 435, 338),
		quadTo(502, 205, 614, 205),
		quadTo(860, 205, 860, 1024),
		quadTo(860, 1167, 793, 1300),
		quadTo(727, 1434, 614, 1434),
		// - contour #1
		moveTo(614, 1638),
		quadTo(1024, 1638, 1024, 819),
		quadTo(1024, 0, 614, 0),
		quadTo(205, 0, 205, 819),
		quadTo(205, 1638, 614, 1638),
	}, {
		// one
		// - contour #0
		moveTo(205, 0),
		lineTo(205, 1638),
		lineTo(614, 1638),
		lineTo(614, 0),
		lineTo(205, 0),
	}, {
		// five
		// - contour #0
		moveTo(0, 0),
		lineTo(0, 100),
		lineTo(400, 100),
		lineTo(400, 0),
		lineTo(0, 0),
	}, {
		// six
		// - contour #0
		moveTo(0, 0),
		lineTo(0, 100),
		lineTo(400, 100),
		lineTo(400, 0),
		lineTo(0, 0),
		// - contour #1
		translate(111, 234, moveTo(205, 0)),
		translate(111, 234, lineTo(205, 1638)),
		translate(111, 234, lineTo(614, 1638)),
		translate(111, 234, lineTo(614, 0)),
		translate(111, 234, lineTo(205, 0)),
	}, {
		// seven
		// - contour #0
		moveTo(0, 0),
		lineTo(0, 100),
		lineTo(400, 100),
		lineTo(400, 0),
		lineTo(0, 0),
		// - contour #1
		transform_(1<<13, 0, 0, 1<<13, 56, 117, moveTo(205, 0)),
		transform_(1<<13, 0, 0, 1<<13, 56, 117, lineTo(205, 1638)),
		transform_(1<<13, 0, 0, 1<<13, 56, 117, lineTo(614, 1638)),
		transform_(1<<13, 0, 0, 1<<13, 56, 117, lineTo(614, 0)),
		transform_(1<<13, 0, 0, 1<<13, 56, 117, lineTo(205, 0)),
	}, {
		// eight
		// - contour #0
		moveTo(0, 0),
		lineTo(0, 100),
		lineTo(400, 100),
		lineTo(400, 0),
		lineTo(0, 0),
		// - contour #1
		transform_(3<<13, 0, 0, 1<<13, 56, 117, moveTo(205, 0)),
		transform_(3<<13, 0, 0, 1<<13, 56, 117, lineTo(205, 1638)),
		transform_(3<<13, 0, 0, 1<<13, 56, 117, lineTo(614, 1638)),
		transform_(3<<13, 0, 0, 1<<13, 56, 117, lineTo(614, 0)),
		transform_(3<<13, 0, 0, 1<<13, 56, 117, lineTo(205, 0)),
	}, {
		// nine
		// - contour #0
		moveTo(0, 0),
		lineTo(0, 100),
		lineTo(400, 100),
		lineTo(400, 0),
		lineTo(0, 0),
		// - contour #1
		transform_(22381, 8192, 5996, 14188, 237, 258, moveTo(205, 0)),
		transform_(22381, 8192, 5996, 14188, 237, 258, lineTo(205, 1638)),
		transform_(22381, 8192, 5996, 14188, 237, 258, lineTo(614, 1638)),
		transform_(22381, 8192, 5996, 14188, 237, 258, lineTo(614, 0)),
		transform_(22381, 8192, 5996, 14188, 237, 258, lineTo(205, 0)),
	}}

	if len(f.Glyf) != len(expecteds) {
		t.Fatalf("number of glyphs: expected %d, got %d", len(expecteds), len(f.Glyf))
	}

	for i, expected := range expecteds {
		var points []contourPoint
		f.getPointsForGlyph(fonts.GID(i), 0, &points)
		got := buildSegments(points[:len(points)-phantomCount])
		if len(expected) == 0 {
			expected = nil
		}
		if !reflect.DeepEqual(expected, got) {
			t.Fatalf("GID %d, expected %v, got %v", i, expected, got)
		}
	}
}

func TestGlyfSegments2(t *testing.T) {
	// copied from fontforge .sdf saved file
	expecteds := [...][]fonts.Segment{
		{
			// .notdef
			moveTo(100, 0),
			lineTo(100, 1456),
			lineTo(808, 1456),
			lineTo(808, 0),
			lineTo(100, 0),
			moveTo(194, 1402),
			lineTo(452, 796),
			lineTo(709, 1402),
			lineTo(194, 1402),
			moveTo(480, 728),
			lineTo(754, 84),
			lineTo(754, 1372),
			lineTo(480, 728),
			moveTo(154, 1360),
			lineTo(154, 96),
			lineTo(422, 728),
			lineTo(154, 1360),
			moveTo(194, 54),
			lineTo(709, 54),
			lineTo(452, 660),
			lineTo(194, 54),
		},
		{},
		{},
		{},
		{},
		{},
		{
			moveTo(576, 1456),
			lineTo(369, 448),
			lineTo(133, 448),
			lineTo(276, 1456),
			lineTo(576, 1456),
			moveTo(40, 131),
			quadTo(38, 199, 83, 244.5),
			quadTo(128, 290, 195, 291),
			quadTo(261, 292, 307, 250),
			quadTo(353, 208, 355, 142),
			quadTo(357, 74, 312, 29),
			quadTo(267, -16, 200, -18),
			quadTo(135, -19, 88, 23),
			quadTo(41, 65, 40, 131),
		},
		{
			moveTo(697, 1383),
			lineTo(598, 987),
			lineTo(438, 987),
			lineTo(524, 1537),
			lineTo(721, 1537),
			lineTo(697, 1383),
			moveTo(381, 1383),
			lineTo(283, 987),
			lineTo(121, 987),
			lineTo(208, 1537),
			lineTo(406, 1537),
			lineTo(381, 1383),
		},

		{
			moveTo(469, 0),
			lineTo(611, 410),
			lineTo(431, 410),
			lineTo(290, 0),
			lineTo(104, 0),
			lineTo(246, 410),
			lineTo(28, 410),
			lineTo(59, 583),
			lineTo(305, 583),
			lineTo(403, 867),
			lineTo(180, 867),
			lineTo(211, 1040),
			lineTo(462, 1040),
			lineTo(606, 1456),
			lineTo(790, 1456),
			lineTo(647, 1040),
			lineTo(826, 1040),
			lineTo(970, 1456),
			lineTo(1156, 1456),
			lineTo(1013, 1040),
			lineTo(1222, 1040),
			lineTo(1192, 867),
			lineTo(953, 867),
			lineTo(855, 583),
			lineTo(1070, 583),
			lineTo(1039, 410),
			lineTo(796, 410),
			lineTo(655, 0),
			lineTo(469, 0),
			moveTo(490, 583),
			lineTo(669, 583),
			lineTo(768, 867),
			lineTo(588, 867),
			lineTo(490, 583),
		},
		{
			moveTo(1013, 393),
			quadTo(1002, 205, 875.5, 103.5),
			quadTo(749, 2, 573, -16),
			lineTo(534, -215),
			lineTo(378, -215),
			lineTo(417, -15),
			quadTo(229, 14, 142.5, 145),
			quadTo(56, 276, 61, 458),
			lineTo(343, 457),
			quadTo(339, 365, 374, 289),
			quadTo(409, 213, 516, 212),
			quadTo(600, 211, 661, 259),
			quadTo(722, 307, 734, 391),
			quadTo(747, 478, 700, 525.5),
			quadTo(653, 573, 580, 606),
			quadTo(427, 675, 315, 775),
			quadTo(203, 875, 215, 1062),
			quadTo(226, 1248, 352.5, 1351.5),
			quadTo(479, 1455, 654, 1473),
			lineTo(695, 1688),
			lineTo(851, 1688),
			lineTo(810, 1468),
			quadTo(981, 1432, 1050.5, 1299.5),
			quadTo(1120, 1167, 1116, 1005),
			lineTo(833, 1006),
			quadTo(837, 1083, 814, 1163),
			quadTo(791, 1243, 693, 1245),
			quadTo(610, 1247, 558, 1195),
			quadTo(506, 1143, 495, 1064),
			quadTo(484, 983, 526.5, 937.5),
			quadTo(569, 892, 647, 855),
			quadTo(813, 776, 918.5, 675),
			quadTo(1024, 574, 1013, 393),
		},
	}
	font := loadFont(t, "Roboto-BoldItalic.ttf")

	for i, expected := range expecteds {
		var points []contourPoint
		font.getPointsForGlyph(fonts.GID(i), 0, &points)
		got := buildSegments(points[:len(points)-phantomCount])
		if len(expected) == 0 {
			expected = nil
		}
		if !reflect.DeepEqual(expected, got) {
			t.Fatalf("GID %d, expected %v, got %v", i, expected, got)
		}
	}
}

func TestGlyfSegments3(t *testing.T) {
	for _, filename := range [...]string{
		"Roboto-BoldItalic.ttf",
		"Commissioner-VF.ttf",
		"FreeSerif.ttf",
	} {
		font := loadFont(t, filename)

		b, _ := testdata.Files.ReadFile(filename)
		fontS, err := sfnt.Parse(b)
		if err != nil {
			t.Fatal(filename, err)
		}

		for i := 0; i < font.NumGlyphs; i++ {
			var points []contourPoint
			font.getPointsForGlyph(fonts.GID(i), 0, &points)
			got := buildSegments(points[:len(points)-phantomCount])

			expected, _ := fontS.LoadGlyph(nil, sfnt.GlyphIndex(i), fx.Int26_6(fontS.UnitsPerEm()), nil)

			// since sfnt sometimes has rounding errors,
			// we only check for the structure of the segments
			if len(expected) != len(got) {
				t.Fatalf("GID %d", i)
			}
			for j, g := range got {
				if g.Op != fonts.SegmentOp(expected[j].Op) {
					t.Fatalf("GID %d: %d != %d", i, g.Op, expected[j].Op)
				}
			}
		}
	}
}

func TestCFFSegments(t *testing.T) {
	// wants' vectors correspond 1-to-1 to what's in the CFFTest.sfd file
	expecteds := [][]fonts.Segment{{
		// .notdef
		// - contour #0
		moveTo(50, 0),
		lineTo(450, 0),
		lineTo(450, 533),
		lineTo(50, 533),
		lineTo(50, 0),
		// - contour #1
		moveTo(100, 50),
		lineTo(100, 483),
		lineTo(400, 483),
		lineTo(400, 50),
		lineTo(100, 50),
	}, {
		// zero
		// - contour #0
		moveTo(300, 700),
		cubeTo(380, 700, 420, 580, 420, 500),
		cubeTo(420, 350, 390, 100, 300, 100),
		cubeTo(220, 100, 180, 220, 180, 300),
		cubeTo(180, 450, 210, 700, 300, 700),
		// - contour #1
		moveTo(300, 800),
		cubeTo(200, 800, 100, 580, 100, 400),
		cubeTo(100, 220, 200, 0, 300, 0),
		cubeTo(400, 0, 500, 220, 500, 400),
		cubeTo(500, 580, 400, 800, 300, 800),
	}, {
		// one
		// - contour #0
		moveTo(100, 0),
		lineTo(300, 0),
		lineTo(300, 800),
		lineTo(100, 800),
		lineTo(100, 0),
	}, {
		// Q
		// - contour #0
		moveTo(657, 237),
		lineTo(289, 387),
		lineTo(519, 615),
		lineTo(657, 237),
		// - contour #1
		moveTo(792, 169),
		cubeTo(867, 263, 926, 502, 791, 665),
		cubeTo(645, 840, 380, 831, 228, 673),
		cubeTo(71, 509, 110, 231, 242, 93),
		cubeTo(369, -39, 641, 18, 722, 93),
		lineTo(802, 3),
		lineTo(864, 83),
		lineTo(792, 169),
	}, {
		// uni4E2D
		// - contour #0
		moveTo(141, 520),
		lineTo(137, 356),
		lineTo(245, 400),
		lineTo(331, 26),
		lineTo(355, 414),
		lineTo(463, 434),
		lineTo(453, 620),
		lineTo(341, 592),
		lineTo(331, 758),
		lineTo(243, 752),
		lineTo(235, 562),
		lineTo(141, 520),
	}}

	font := loadFont(t, "CFFTest.otf")

	for i, expected := range expecteds {
		got, err := font.glyphDataFromCFF1(fonts.GID(i))
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(expected, got.Segments) {
			t.Fatalf("GID %d, expected\n%v, got\n%v", i, expected, got)
		}
	}
}
