package fontconfig

// Standard exposes the parsed configuration
// described in the 'confs' folder.
var Standard = &Config{
	subst: []ruleSet{{name: "confs/09-autohint-if-no-hinting.conf", description: "Enable autohinter if font doesn't have any hinting", domain: "", subst: [matchKindEnd][]directive{nil, {{
		tests: []ruleTest{{expr: &expression{u: Bool(0), op: 5}, kind: 1, qual: 0, object: 51, op: 22}},
		edits: []ruleEdit{{expr: expression{u: Bool(1), op: 5}, binding: 0, object: 19, op: 15}},
	}}, nil}},
		{name: "confs/10-autohint.conf", description: "Enable autohinter", domain: "", subst: [matchKindEnd][]directive{{{
			tests: nil,
			edits: []ruleEdit{{expr: expression{u: Bool(1), op: 5}, binding: 0, object: 19, op: 15}},
		}}, nil, nil}},
		{name: "confs/10-hinting-full.conf", description: "Set hintfull to hintstyle", domain: "", subst: [matchKindEnd][]directive{{{
			tests: nil,
			edits: []ruleEdit{{expr: expression{u: String("hintfull"), op: 10}, binding: 0, object: 16, op: 15}},
		}}, nil, nil}},
		{name: "confs/10-hinting-medium.conf", description: "Set hintmedium to hintstyle", domain: "", subst: [matchKindEnd][]directive{{{
			tests: nil,
			edits: []ruleEdit{{expr: expression{u: String("hintmedium"), op: 10}, binding: 0, object: 16, op: 15}},
		}}, nil, nil}},
		{name: "confs/10-hinting-none.conf", description: "Set hintnone to hintstyle", domain: "", subst: [matchKindEnd][]directive{{{
			tests: nil,
			edits: []ruleEdit{{expr: expression{u: String("hintnone"), op: 10}, binding: 0, object: 16, op: 15}},
		}}, nil, nil}},
		{name: "confs/10-hinting-slight.conf", description: "Set hintslight to hintstyle", domain: "", subst: [matchKindEnd][]directive{{{
			tests: nil,
			edits: []ruleEdit{{expr: expression{u: String("hintslight"), op: 10}, binding: 0, object: 16, op: 15}},
		}}, nil, nil}},
		{name: "confs/10-no-sub-pixel.conf", description: "Disable sub-pixel rendering", domain: "", subst: [matchKindEnd][]directive{{{
			tests: nil,
			edits: []ruleEdit{{expr: expression{u: String("none"), op: 10}, binding: 0, object: 27, op: 15}},
		}}, nil, nil}},
		{name: "confs/10-scale-bitmap-fonts.conf", description: "Bitmap scaling", domain: "", subst: [matchKindEnd][]directive{nil, {{
			tests: []ruleTest{{expr: &expression{u: Bool(0), op: 5}, kind: 1, qual: 0, object: 24, op: 22}},
			edits: []ruleEdit{{expr: expression{u: exprTree{&expression{u: exprName{object: 12, kind: 0}, op: 9}, &expression{u: exprName{object: 12, kind: 1}, op: 9}}, op: 34}, binding: 0, object: 74, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: Bool(0), op: 5}, kind: 1, qual: 0, object: 24, op: 22},
				{expr: &expression{u: Bool(0), op: 5}, kind: 1, qual: 0, object: 25, op: 22},
				{expr: &expression{u: Bool(1), op: 5}, kind: 1, qual: 0, object: 17, op: 22}},
			edits: []ruleEdit{{expr: expression{u: exprTree{&expression{u: exprTree{&expression{u: exprName{object: 74, kind: -1}, op: 9}, &expression{u: Float(1.2), op: 1}}, op: 27}, &expression{u: exprTree{&expression{u: exprName{object: 74, kind: -1}, op: 9}, &expression{u: Float(0.8), op: 1}}, op: 29}}, op: 21}, binding: 0, object: 75, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: Bool(1), op: 5}, kind: 1, qual: 0, object: 0, op: 22}},
			edits: []ruleEdit{{expr: expression{u: Float(1), op: 1}, binding: 0, object: 74, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: Bool(0), op: 5}, kind: 1, qual: 0, object: 24, op: 22},
				{expr: &expression{u: Float(1), op: 1}, kind: 1, qual: 0, object: 0, op: 23}},
			edits: []ruleEdit{{expr: expression{u: exprTree{&expression{u: exprName{object: 32, kind: -1}, op: 9}, &expression{u: exprMatrix{xx: &expression{u: exprName{object: 74, kind: -1}, op: 9}, xy: &expression{u: Float(0), op: 1}, yx: &expression{u: Float(0), op: 1}, yy: &expression{u: exprName{object: 74, kind: -1}, op: 9}}, op: 3}}, op: 33}, binding: 0, object: 32, op: 11},
				{expr: expression{u: exprTree{&expression{u: exprName{object: 10, kind: -1}, op: 9}, &expression{u: exprName{object: 74, kind: -1}, op: 9}}, op: 34}, binding: 0, object: 10, op: 11}},
		}}, nil}},
		{name: "confs/10-sub-pixel-bgr.conf", description: "Enable sub-pixel rendering with the BGR stripes layout", domain: "", subst: [matchKindEnd][]directive{{{
			tests: nil,
			edits: []ruleEdit{{expr: expression{u: String("bgr"), op: 10}, binding: 0, object: 27, op: 15}},
		}}, nil, nil}},
		{name: "confs/10-sub-pixel-rgb.conf", description: "Enable sub-pixel rendering with the RGB stripes layout", domain: "", subst: [matchKindEnd][]directive{{{
			tests: nil,
			edits: []ruleEdit{{expr: expression{u: String("rgb"), op: 10}, binding: 0, object: 27, op: 15}},
		}}, nil, nil}},
		{name: "confs/10-sub-pixel-vbgr.conf", description: "Enable sub-pixel rendering with the vertical BGR stripes layout", domain: "", subst: [matchKindEnd][]directive{{{
			tests: nil,
			edits: []ruleEdit{{expr: expression{u: String("vbgr"), op: 10}, binding: 0, object: 27, op: 15}},
		}}, nil, nil}},
		{name: "confs/10-sub-pixel-vrgb.conf", description: "Enable sub-pixel rendering with the vertical RGB stripes layout", domain: "", subst: [matchKindEnd][]directive{{{
			tests: nil,
			edits: []ruleEdit{{expr: expression{u: String("vrgb"), op: 10}, binding: 0, object: 27, op: 15}},
		}}, nil, nil}},
		{name: "confs/10-unhinted.conf", description: "Disable hinting", domain: "", subst: [matchKindEnd][]directive{{{
			tests: nil,
			edits: []ruleEdit{{expr: expression{u: Bool(0), op: 5}, binding: 0, object: 17, op: 15}},
		}}, nil, nil}},
		{name: "confs/11-lcdfilter-default.conf", description: "Use lcddefault as default for LCD filter", domain: "", subst: [matchKindEnd][]directive{{{
			tests: nil,
			edits: []ruleEdit{{expr: expression{u: String("lcddefault"), op: 10}, binding: 0, object: 41, op: 15}},
		}}, nil, nil}},
		{name: "confs/11-lcdfilter-legacy.conf", description: "Use lcdlegacy as default for LCD filter", domain: "", subst: [matchKindEnd][]directive{{{
			tests: nil,
			edits: []ruleEdit{{expr: expression{u: String("lcdlegacy"), op: 10}, binding: 0, object: 41, op: 15}},
		}}, nil, nil}},
		{name: "confs/11-lcdfilter-light.conf", description: "Use lcdlight as default for LCD filter", domain: "", subst: [matchKindEnd][]directive{{{
			tests: nil,
			edits: []ruleEdit{{expr: expression{u: String("lcdlight"), op: 10}, binding: 0, object: 41, op: 15}},
		}}, nil, nil}},
		{name: "confs/20-unhint-small-vera.conf", description: "Disable hinting for Bitstream Vera fonts when the size is less than 8ppem", domain: "", subst: [matchKindEnd][]directive{nil, {{
			tests: []ruleTest{{expr: &expression{u: String("Bitstream Vera Sans"), op: 2}, kind: 1, qual: 0, object: 1, op: 65558},
				{expr: &expression{u: Float(7.5), op: 1}, kind: 1, qual: 0, object: 12, op: 27}},
			edits: []ruleEdit{{expr: expression{u: Bool(0), op: 5}, binding: 0, object: 17, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("Bitstream Vera Serif"), op: 2}, kind: 1, qual: 0, object: 1, op: 65558},
				{expr: &expression{u: Float(7.5), op: 1}, kind: 1, qual: 0, object: 12, op: 27}},
			edits: []ruleEdit{{expr: expression{u: Bool(0), op: 5}, binding: 0, object: 17, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("Bitstream Vera Sans Mono"), op: 2}, kind: 1, qual: 0, object: 1, op: 65558},
				{expr: &expression{u: Float(7.5), op: 1}, kind: 1, qual: 0, object: 12, op: 27}},
			edits: []ruleEdit{{expr: expression{u: Bool(0), op: 5}, binding: 0, object: 17, op: 11}},
		}}, nil}},
		{name: "confs/25-unhint-nonlatin.conf", description: "Disable hinting for CJK fonts", domain: "", subst: [matchKindEnd][]directive{nil, {{
			tests: []ruleTest{{expr: &expression{u: String("Kochi Mincho"), op: 2}, kind: 1, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: Bool(0), op: 5}, binding: 0, object: 17, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("Kochi Gothic"), op: 2}, kind: 1, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: Bool(0), op: 5}, binding: 0, object: 17, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("Sazanami Mincho"), op: 2}, kind: 1, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: Bool(0), op: 5}, binding: 0, object: 17, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("Sazanami Gothic"), op: 2}, kind: 1, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: Bool(0), op: 5}, binding: 0, object: 17, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("Baekmuk Batang"), op: 2}, kind: 1, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: Bool(0), op: 5}, binding: 0, object: 17, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("Baekmuk Dotum"), op: 2}, kind: 1, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: Bool(0), op: 5}, binding: 0, object: 17, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("Baekmuk Gulim"), op: 2}, kind: 1, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: Bool(0), op: 5}, binding: 0, object: 17, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("Baekmuk Headline"), op: 2}, kind: 1, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: Bool(0), op: 5}, binding: 0, object: 17, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("AR PL Mingti2L Big5"), op: 2}, kind: 1, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: Bool(0), op: 5}, binding: 0, object: 17, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("AR PL ShanHeiSun Uni"), op: 2}, kind: 1, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: Bool(0), op: 5}, binding: 0, object: 17, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("AR PL KaitiM Big5"), op: 2}, kind: 1, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: Bool(0), op: 5}, binding: 0, object: 17, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("AR PL ZenKai Uni"), op: 2}, kind: 1, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: Bool(0), op: 5}, binding: 0, object: 17, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("AR PL SungtiL GB"), op: 2}, kind: 1, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: Bool(0), op: 5}, binding: 0, object: 17, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("AR PL KaitiM GB"), op: 2}, kind: 1, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: Bool(0), op: 5}, binding: 0, object: 17, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("ZYSong18030"), op: 2}, kind: 1, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: Bool(0), op: 5}, binding: 0, object: 17, op: 11}},
		}}, nil}},
		{name: "confs/30-metric-aliases.conf", description: "Set substitutions for similar/metric-compatible families", domain: "", subst: [matchKindEnd][]directive{{{
			tests: []ruleTest{{expr: &expression{u: String("Nimbus Sans L"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("Helvetica"), op: 2}, binding: 2, object: 1, op: 16}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("Nimbus Sans"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("Helvetica"), op: 2}, binding: 2, object: 1, op: 16}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("TeX Gyre Heros"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("Helvetica"), op: 2}, binding: 2, object: 1, op: 16}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("Nimbus Sans Narrow"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("Helvetica Narrow"), op: 2}, binding: 2, object: 1, op: 16}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("TeX Gyre Heros Cn"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("Helvetica Narrow"), op: 2}, binding: 2, object: 1, op: 16}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("Nimbus Roman No9 L"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("Times"), op: 2}, binding: 2, object: 1, op: 16}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("Nimbus Roman"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("Times"), op: 2}, binding: 2, object: 1, op: 16}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("TeX Gyre Termes"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("Times"), op: 2}, binding: 2, object: 1, op: 16}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("Nimbus Mono L"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("Courier"), op: 2}, binding: 2, object: 1, op: 16}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("Nimbus Mono"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("Courier"), op: 2}, binding: 2, object: 1, op: 16}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("Nimbus Mono PS"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("Courier"), op: 2}, binding: 2, object: 1, op: 16}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("TeX Gyre Cursor"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("Courier"), op: 2}, binding: 2, object: 1, op: 16}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("Avant Garde"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("ITC Avant Garde Gothic"), op: 2}, binding: 2, object: 1, op: 16}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("URW Gothic L"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("ITC Avant Garde Gothic"), op: 2}, binding: 2, object: 1, op: 16}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("URW Gothic"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("ITC Avant Garde Gothic"), op: 2}, binding: 2, object: 1, op: 16}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("TeX Gyre Adventor"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("ITC Avant Garde Gothic"), op: 2}, binding: 2, object: 1, op: 16}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("Bookman"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("ITC Bookman"), op: 2}, binding: 2, object: 1, op: 16}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("URW Bookman L"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("ITC Bookman"), op: 2}, binding: 2, object: 1, op: 16}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("Bookman URW"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("ITC Bookman"), op: 2}, binding: 2, object: 1, op: 16}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("URW Bookman"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("ITC Bookman"), op: 2}, binding: 2, object: 1, op: 16}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("TeX Gyre Bonum"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("ITC Bookman"), op: 2}, binding: 2, object: 1, op: 16}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("Bookman Old Style"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("ITC Bookman"), op: 2}, binding: 2, object: 1, op: 16}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("Zapf Chancery"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("ITC Zapf Chancery"), op: 2}, binding: 2, object: 1, op: 16}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("URW Chancery L"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("ITC Zapf Chancery"), op: 2}, binding: 2, object: 1, op: 16}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("Chancery URW"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("ITC Zapf Chancery"), op: 2}, binding: 2, object: 1, op: 16}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("Z003"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("ITC Zapf Chancery"), op: 2}, binding: 2, object: 1, op: 16}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("TeX Gyre Chorus"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("ITC Zapf Chancery"), op: 2}, binding: 2, object: 1, op: 16}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("URW Palladio L"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("Palatino"), op: 2}, binding: 2, object: 1, op: 16}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("Palladio URW"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("Palatino"), op: 2}, binding: 2, object: 1, op: 16}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("P052"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("Palatino"), op: 2}, binding: 2, object: 1, op: 16}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("TeX Gyre Pagella"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("Palatino"), op: 2}, binding: 2, object: 1, op: 16}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("Palatino Linotype"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("Palatino"), op: 2}, binding: 2, object: 1, op: 16}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("Century Schoolbook L"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("New Century Schoolbook"), op: 2}, binding: 2, object: 1, op: 16}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("Century SchoolBook URW"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("New Century Schoolbook"), op: 2}, binding: 2, object: 1, op: 16}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("C059"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("New Century Schoolbook"), op: 2}, binding: 2, object: 1, op: 16}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("TeX Gyre Schola"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("New Century Schoolbook"), op: 2}, binding: 2, object: 1, op: 16}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("Century Schoolbook"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("New Century Schoolbook"), op: 2}, binding: 2, object: 1, op: 16}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("Arimo"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("Arial"), op: 2}, binding: 2, object: 1, op: 16}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("Liberation Sans"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("Arial"), op: 2}, binding: 2, object: 1, op: 16}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("Liberation Sans Narrow"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("Arial Narrow"), op: 2}, binding: 2, object: 1, op: 16}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("Albany"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("Arial"), op: 2}, binding: 2, object: 1, op: 16}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("Albany AMT"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("Arial"), op: 2}, binding: 2, object: 1, op: 16}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("Tinos"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("Times New Roman"), op: 2}, binding: 2, object: 1, op: 16}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("Liberation Serif"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("Times New Roman"), op: 2}, binding: 2, object: 1, op: 16}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("Thorndale"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("Times New Roman"), op: 2}, binding: 2, object: 1, op: 16}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("Thorndale AMT"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("Times New Roman"), op: 2}, binding: 2, object: 1, op: 16}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("Cousine"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("Courier New"), op: 2}, binding: 2, object: 1, op: 16}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("Liberation Mono"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("Courier New"), op: 2}, binding: 2, object: 1, op: 16}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("Cumberland"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("Courier New"), op: 2}, binding: 2, object: 1, op: 16}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("Cumberland AMT"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("Courier New"), op: 2}, binding: 2, object: 1, op: 16}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("Gelasio"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("Georgia"), op: 2}, binding: 2, object: 1, op: 16}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("Caladea"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("Cambria"), op: 2}, binding: 2, object: 1, op: 16}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("Carlito"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("Calibri"), op: 2}, binding: 2, object: 1, op: 16}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("SymbolNeu"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("Symbol"), op: 2}, binding: 2, object: 1, op: 16}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("Helvetica"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("Arial"), op: 2}, binding: 0, object: 1, op: 16}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("Helvetica Narrow"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("Arial Narrow"), op: 2}, binding: 0, object: 1, op: 16}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("Times"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("Times New Roman"), op: 2}, binding: 0, object: 1, op: 16}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("Courier"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("Courier New"), op: 2}, binding: 0, object: 1, op: 16}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("Arial"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("Helvetica"), op: 2}, binding: 0, object: 1, op: 16}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("Arial Narrow"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("Helvetica Narrow"), op: 2}, binding: 0, object: 1, op: 16}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("Times New Roman"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("Times"), op: 2}, binding: 0, object: 1, op: 16}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("Courier New"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("Courier"), op: 2}, binding: 0, object: 1, op: 16}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("Helvetica"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("TeX Gyre Heros"), op: 2}, binding: 2, object: 1, op: 15}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("Helvetica Narrow"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("TeX Gyre Heros Cn"), op: 2}, binding: 2, object: 1, op: 15}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("Times"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("TeX Gyre Termes"), op: 2}, binding: 2, object: 1, op: 15}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("Courier"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("TeX Gyre Cursor"), op: 2}, binding: 2, object: 1, op: 15}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("Courier Std"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("Courier"), op: 2}, binding: 2, object: 1, op: 15}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("ITC Avant Garde Gothic"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("TeX Gyre Adventor"), op: 2}, binding: 2, object: 1, op: 15}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("ITC Bookman"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: exprTree{&expression{u: String("Bookman Old Style"), op: 2}, &expression{u: String("TeX Gyre Bonum"), op: 2}}, op: 36}, binding: 2, object: 1, op: 15}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("ITC Zapf Chancery"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("TeX Gyre Chorus"), op: 2}, binding: 2, object: 1, op: 15}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("Palatino"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: exprTree{&expression{u: String("Palatino Linotype"), op: 2}, &expression{u: String("TeX Gyre Pagella"), op: 2}}, op: 36}, binding: 2, object: 1, op: 15}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("New Century Schoolbook"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: exprTree{&expression{u: String("Century Schoolbook"), op: 2}, &expression{u: String("TeX Gyre Schola"), op: 2}}, op: 36}, binding: 2, object: 1, op: 15}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("Arial"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: exprTree{&expression{u: String("Arimo"), op: 2}, &expression{u: exprTree{&expression{u: String("Liberation Sans"), op: 2}, &expression{u: exprTree{&expression{u: String("Albany"), op: 2}, &expression{u: String("Albany AMT"), op: 2}}, op: 36}}, op: 36}}, op: 36}, binding: 2, object: 1, op: 15}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("Arial Narrow"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("Liberation Sans Narrow"), op: 2}, binding: 2, object: 1, op: 15}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("Times New Roman"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: exprTree{&expression{u: String("Tinos"), op: 2}, &expression{u: exprTree{&expression{u: String("Liberation Serif"), op: 2}, &expression{u: exprTree{&expression{u: String("Thorndale"), op: 2}, &expression{u: String("Thorndale AMT"), op: 2}}, op: 36}}, op: 36}}, op: 36}, binding: 2, object: 1, op: 15}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("Courier New"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: exprTree{&expression{u: String("Cousine"), op: 2}, &expression{u: exprTree{&expression{u: String("Liberation Mono"), op: 2}, &expression{u: exprTree{&expression{u: String("Cumberland"), op: 2}, &expression{u: String("Cumberland AMT"), op: 2}}, op: 36}}, op: 36}}, op: 36}, binding: 2, object: 1, op: 15}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("Georgia"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("Gelasio"), op: 2}, binding: 2, object: 1, op: 15}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("Cambria"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("Caladea"), op: 2}, binding: 2, object: 1, op: 15}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("Calibri"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("Carlito"), op: 2}, binding: 2, object: 1, op: 15}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("Symbol"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("SymbolNeu"), op: 2}, binding: 2, object: 1, op: 15}},
		}}, nil, nil}},
		{name: "confs/35-lang-normalize.conf", description: "", domain: "", subst: [matchKindEnd][]directive{{{
			tests: []ruleTest{{expr: &expression{u: String("aa"), op: 2}, kind: 0, qual: 0, object: 34, op: 24}},
			edits: []ruleEdit{{expr: expression{u: String("aa"), op: 2}, binding: 2, object: 34, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("ab"), op: 2}, kind: 0, qual: 0, object: 34, op: 24}},
			edits: []ruleEdit{{expr: expression{u: String("ab"), op: 2}, binding: 2, object: 34, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("af"), op: 2}, kind: 0, qual: 0, object: 34, op: 24}},
			edits: []ruleEdit{{expr: expression{u: String("af"), op: 2}, binding: 2, object: 34, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("ak"), op: 2}, kind: 0, qual: 0, object: 34, op: 24}},
			edits: []ruleEdit{{expr: expression{u: String("ak"), op: 2}, binding: 2, object: 34, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("am"), op: 2}, kind: 0, qual: 0, object: 34, op: 24}},
			edits: []ruleEdit{{expr: expression{u: String("am"), op: 2}, binding: 2, object: 34, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("an"), op: 2}, kind: 0, qual: 0, object: 34, op: 24}},
			edits: []ruleEdit{{expr: expression{u: String("an"), op: 2}, binding: 2, object: 34, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("ar"), op: 2}, kind: 0, qual: 0, object: 34, op: 24}},
			edits: []ruleEdit{{expr: expression{u: String("ar"), op: 2}, binding: 2, object: 34, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("as"), op: 2}, kind: 0, qual: 0, object: 34, op: 24}},
			edits: []ruleEdit{{expr: expression{u: String("as"), op: 2}, binding: 2, object: 34, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("ast"), op: 2}, kind: 0, qual: 0, object: 34, op: 24}},
			edits: []ruleEdit{{expr: expression{u: String("ast"), op: 2}, binding: 2, object: 34, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("av"), op: 2}, kind: 0, qual: 0, object: 34, op: 24}},
			edits: []ruleEdit{{expr: expression{u: String("av"), op: 2}, binding: 2, object: 34, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("ay"), op: 2}, kind: 0, qual: 0, object: 34, op: 24}},
			edits: []ruleEdit{{expr: expression{u: String("ay"), op: 2}, binding: 2, object: 34, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("ba"), op: 2}, kind: 0, qual: 0, object: 34, op: 24}},
			edits: []ruleEdit{{expr: expression{u: String("ba"), op: 2}, binding: 2, object: 34, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("be"), op: 2}, kind: 0, qual: 0, object: 34, op: 24}},
			edits: []ruleEdit{{expr: expression{u: String("be"), op: 2}, binding: 2, object: 34, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("bg"), op: 2}, kind: 0, qual: 0, object: 34, op: 24}},
			edits: []ruleEdit{{expr: expression{u: String("bg"), op: 2}, binding: 2, object: 34, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("bh"), op: 2}, kind: 0, qual: 0, object: 34, op: 24}},
			edits: []ruleEdit{{expr: expression{u: String("bh"), op: 2}, binding: 2, object: 34, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("bho"), op: 2}, kind: 0, qual: 0, object: 34, op: 24}},
			edits: []ruleEdit{{expr: expression{u: String("bho"), op: 2}, binding: 2, object: 34, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("bi"), op: 2}, kind: 0, qual: 0, object: 34, op: 24}},
			edits: []ruleEdit{{expr: expression{u: String("bi"), op: 2}, binding: 2, object: 34, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("bin"), op: 2}, kind: 0, qual: 0, object: 34, op: 24}},
			edits: []ruleEdit{{expr: expression{u: String("bin"), op: 2}, binding: 2, object: 34, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("bm"), op: 2}, kind: 0, qual: 0, object: 34, op: 24}},
			edits: []ruleEdit{{expr: expression{u: String("bm"), op: 2}, binding: 2, object: 34, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("bn"), op: 2}, kind: 0, qual: 0, object: 34, op: 24}},
			edits: []ruleEdit{{expr: expression{u: String("bn"), op: 2}, binding: 2, object: 34, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("bo"), op: 2}, kind: 0, qual: 0, object: 34, op: 24}},
			edits: []ruleEdit{{expr: expression{u: String("bo"), op: 2}, binding: 2, object: 34, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("br"), op: 2}, kind: 0, qual: 0, object: 34, op: 24}},
			edits: []ruleEdit{{expr: expression{u: String("br"), op: 2}, binding: 2, object: 34, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("brx"), op: 2}, kind: 0, qual: 0, object: 34, op: 24}},
			edits: []ruleEdit{{expr: expression{u: String("brx"), op: 2}, binding: 2, object: 34, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("bs"), op: 2}, kind: 0, qual: 0, object: 34, op: 24}},
			edits: []ruleEdit{{expr: expression{u: String("bs"), op: 2}, binding: 2, object: 34, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("bua"), op: 2}, kind: 0, qual: 0, object: 34, op: 24}},
			edits: []ruleEdit{{expr: expression{u: String("bua"), op: 2}, binding: 2, object: 34, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("byn"), op: 2}, kind: 0, qual: 0, object: 34, op: 24}},
			edits: []ruleEdit{{expr: expression{u: String("byn"), op: 2}, binding: 2, object: 34, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("ca"), op: 2}, kind: 0, qual: 0, object: 34, op: 24}},
			edits: []ruleEdit{{expr: expression{u: String("ca"), op: 2}, binding: 2, object: 34, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("ce"), op: 2}, kind: 0, qual: 0, object: 34, op: 24}},
			edits: []ruleEdit{{expr: expression{u: String("ce"), op: 2}, binding: 2, object: 34, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("ch"), op: 2}, kind: 0, qual: 0, object: 34, op: 24}},
			edits: []ruleEdit{{expr: expression{u: String("ch"), op: 2}, binding: 2, object: 34, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("chm"), op: 2}, kind: 0, qual: 0, object: 34, op: 24}},
			edits: []ruleEdit{{expr: expression{u: String("chm"), op: 2}, binding: 2, object: 34, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("chr"), op: 2}, kind: 0, qual: 0, object: 34, op: 24}},
			edits: []ruleEdit{{expr: expression{u: String("chr"), op: 2}, binding: 2, object: 34, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("co"), op: 2}, kind: 0, qual: 0, object: 34, op: 24}},
			edits: []ruleEdit{{expr: expression{u: String("co"), op: 2}, binding: 2, object: 34, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("crh"), op: 2}, kind: 0, qual: 0, object: 34, op: 24}},
			edits: []ruleEdit{{expr: expression{u: String("crh"), op: 2}, binding: 2, object: 34, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("cs"), op: 2}, kind: 0, qual: 0, object: 34, op: 24}},
			edits: []ruleEdit{{expr: expression{u: String("cs"), op: 2}, binding: 2, object: 34, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("csb"), op: 2}, kind: 0, qual: 0, object: 34, op: 24}},
			edits: []ruleEdit{{expr: expression{u: String("csb"), op: 2}, binding: 2, object: 34, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("cu"), op: 2}, kind: 0, qual: 0, object: 34, op: 24}},
			edits: []ruleEdit{{expr: expression{u: String("cu"), op: 2}, binding: 2, object: 34, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("cv"), op: 2}, kind: 0, qual: 0, object: 34, op: 24}},
			edits: []ruleEdit{{expr: expression{u: String("cv"), op: 2}, binding: 2, object: 34, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("cy"), op: 2}, kind: 0, qual: 0, object: 34, op: 24}},
			edits: []ruleEdit{{expr: expression{u: String("cy"), op: 2}, binding: 2, object: 34, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("da"), op: 2}, kind: 0, qual: 0, object: 34, op: 24}},
			edits: []ruleEdit{{expr: expression{u: String("da"), op: 2}, binding: 2, object: 34, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("de"), op: 2}, kind: 0, qual: 0, object: 34, op: 24}},
			edits: []ruleEdit{{expr: expression{u: String("de"), op: 2}, binding: 2, object: 34, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("doi"), op: 2}, kind: 0, qual: 0, object: 34, op: 24}},
			edits: []ruleEdit{{expr: expression{u: String("doi"), op: 2}, binding: 2, object: 34, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("dv"), op: 2}, kind: 0, qual: 0, object: 34, op: 24}},
			edits: []ruleEdit{{expr: expression{u: String("dv"), op: 2}, binding: 2, object: 34, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("dz"), op: 2}, kind: 0, qual: 0, object: 34, op: 24}},
			edits: []ruleEdit{{expr: expression{u: String("dz"), op: 2}, binding: 2, object: 34, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("ee"), op: 2}, kind: 0, qual: 0, object: 34, op: 24}},
			edits: []ruleEdit{{expr: expression{u: String("ee"), op: 2}, binding: 2, object: 34, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("el"), op: 2}, kind: 0, qual: 0, object: 34, op: 24}},
			edits: []ruleEdit{{expr: expression{u: String("el"), op: 2}, binding: 2, object: 34, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("en"), op: 2}, kind: 0, qual: 0, object: 34, op: 24}},
			edits: []ruleEdit{{expr: expression{u: String("en"), op: 2}, binding: 2, object: 34, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("eo"), op: 2}, kind: 0, qual: 0, object: 34, op: 24}},
			edits: []ruleEdit{{expr: expression{u: String("eo"), op: 2}, binding: 2, object: 34, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("es"), op: 2}, kind: 0, qual: 0, object: 34, op: 24}},
			edits: []ruleEdit{{expr: expression{u: String("es"), op: 2}, binding: 2, object: 34, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("et"), op: 2}, kind: 0, qual: 0, object: 34, op: 24}},
			edits: []ruleEdit{{expr: expression{u: String("et"), op: 2}, binding: 2, object: 34, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("eu"), op: 2}, kind: 0, qual: 0, object: 34, op: 24}},
			edits: []ruleEdit{{expr: expression{u: String("eu"), op: 2}, binding: 2, object: 34, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("fa"), op: 2}, kind: 0, qual: 0, object: 34, op: 24}},
			edits: []ruleEdit{{expr: expression{u: String("fa"), op: 2}, binding: 2, object: 34, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("fat"), op: 2}, kind: 0, qual: 0, object: 34, op: 24}},
			edits: []ruleEdit{{expr: expression{u: String("fat"), op: 2}, binding: 2, object: 34, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("ff"), op: 2}, kind: 0, qual: 0, object: 34, op: 24}},
			edits: []ruleEdit{{expr: expression{u: String("ff"), op: 2}, binding: 2, object: 34, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("fi"), op: 2}, kind: 0, qual: 0, object: 34, op: 24}},
			edits: []ruleEdit{{expr: expression{u: String("fi"), op: 2}, binding: 2, object: 34, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("fil"), op: 2}, kind: 0, qual: 0, object: 34, op: 24}},
			edits: []ruleEdit{{expr: expression{u: String("fil"), op: 2}, binding: 2, object: 34, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("fj"), op: 2}, kind: 0, qual: 0, object: 34, op: 24}},
			edits: []ruleEdit{{expr: expression{u: String("fj"), op: 2}, binding: 2, object: 34, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("fo"), op: 2}, kind: 0, qual: 0, object: 34, op: 24}},
			edits: []ruleEdit{{expr: expression{u: String("fo"), op: 2}, binding: 2, object: 34, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("fr"), op: 2}, kind: 0, qual: 0, object: 34, op: 24}},
			edits: []ruleEdit{{expr: expression{u: String("fr"), op: 2}, binding: 2, object: 34, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("fur"), op: 2}, kind: 0, qual: 0, object: 34, op: 24}},
			edits: []ruleEdit{{expr: expression{u: String("fur"), op: 2}, binding: 2, object: 34, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("fy"), op: 2}, kind: 0, qual: 0, object: 34, op: 24}},
			edits: []ruleEdit{{expr: expression{u: String("fy"), op: 2}, binding: 2, object: 34, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("ga"), op: 2}, kind: 0, qual: 0, object: 34, op: 24}},
			edits: []ruleEdit{{expr: expression{u: String("ga"), op: 2}, binding: 2, object: 34, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("gd"), op: 2}, kind: 0, qual: 0, object: 34, op: 24}},
			edits: []ruleEdit{{expr: expression{u: String("gd"), op: 2}, binding: 2, object: 34, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("gez"), op: 2}, kind: 0, qual: 0, object: 34, op: 24}},
			edits: []ruleEdit{{expr: expression{u: String("gez"), op: 2}, binding: 2, object: 34, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("gl"), op: 2}, kind: 0, qual: 0, object: 34, op: 24}},
			edits: []ruleEdit{{expr: expression{u: String("gl"), op: 2}, binding: 2, object: 34, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("gn"), op: 2}, kind: 0, qual: 0, object: 34, op: 24}},
			edits: []ruleEdit{{expr: expression{u: String("gn"), op: 2}, binding: 2, object: 34, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("gu"), op: 2}, kind: 0, qual: 0, object: 34, op: 24}},
			edits: []ruleEdit{{expr: expression{u: String("gu"), op: 2}, binding: 2, object: 34, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("gv"), op: 2}, kind: 0, qual: 0, object: 34, op: 24}},
			edits: []ruleEdit{{expr: expression{u: String("gv"), op: 2}, binding: 2, object: 34, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("ha"), op: 2}, kind: 0, qual: 0, object: 34, op: 24}},
			edits: []ruleEdit{{expr: expression{u: String("ha"), op: 2}, binding: 2, object: 34, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("haw"), op: 2}, kind: 0, qual: 0, object: 34, op: 24}},
			edits: []ruleEdit{{expr: expression{u: String("haw"), op: 2}, binding: 2, object: 34, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("he"), op: 2}, kind: 0, qual: 0, object: 34, op: 24}},
			edits: []ruleEdit{{expr: expression{u: String("he"), op: 2}, binding: 2, object: 34, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("hi"), op: 2}, kind: 0, qual: 0, object: 34, op: 24}},
			edits: []ruleEdit{{expr: expression{u: String("hi"), op: 2}, binding: 2, object: 34, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("hne"), op: 2}, kind: 0, qual: 0, object: 34, op: 24}},
			edits: []ruleEdit{{expr: expression{u: String("hne"), op: 2}, binding: 2, object: 34, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("ho"), op: 2}, kind: 0, qual: 0, object: 34, op: 24}},
			edits: []ruleEdit{{expr: expression{u: String("ho"), op: 2}, binding: 2, object: 34, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("hr"), op: 2}, kind: 0, qual: 0, object: 34, op: 24}},
			edits: []ruleEdit{{expr: expression{u: String("hr"), op: 2}, binding: 2, object: 34, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("hsb"), op: 2}, kind: 0, qual: 0, object: 34, op: 24}},
			edits: []ruleEdit{{expr: expression{u: String("hsb"), op: 2}, binding: 2, object: 34, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("ht"), op: 2}, kind: 0, qual: 0, object: 34, op: 24}},
			edits: []ruleEdit{{expr: expression{u: String("ht"), op: 2}, binding: 2, object: 34, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("hu"), op: 2}, kind: 0, qual: 0, object: 34, op: 24}},
			edits: []ruleEdit{{expr: expression{u: String("hu"), op: 2}, binding: 2, object: 34, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("hy"), op: 2}, kind: 0, qual: 0, object: 34, op: 24}},
			edits: []ruleEdit{{expr: expression{u: String("hy"), op: 2}, binding: 2, object: 34, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("hz"), op: 2}, kind: 0, qual: 0, object: 34, op: 24}},
			edits: []ruleEdit{{expr: expression{u: String("hz"), op: 2}, binding: 2, object: 34, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("ia"), op: 2}, kind: 0, qual: 0, object: 34, op: 24}},
			edits: []ruleEdit{{expr: expression{u: String("ia"), op: 2}, binding: 2, object: 34, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("id"), op: 2}, kind: 0, qual: 0, object: 34, op: 24}},
			edits: []ruleEdit{{expr: expression{u: String("id"), op: 2}, binding: 2, object: 34, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("ie"), op: 2}, kind: 0, qual: 0, object: 34, op: 24}},
			edits: []ruleEdit{{expr: expression{u: String("ie"), op: 2}, binding: 2, object: 34, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("ig"), op: 2}, kind: 0, qual: 0, object: 34, op: 24}},
			edits: []ruleEdit{{expr: expression{u: String("ig"), op: 2}, binding: 2, object: 34, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("ii"), op: 2}, kind: 0, qual: 0, object: 34, op: 24}},
			edits: []ruleEdit{{expr: expression{u: String("ii"), op: 2}, binding: 2, object: 34, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("ik"), op: 2}, kind: 0, qual: 0, object: 34, op: 24}},
			edits: []ruleEdit{{expr: expression{u: String("ik"), op: 2}, binding: 2, object: 34, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("io"), op: 2}, kind: 0, qual: 0, object: 34, op: 24}},
			edits: []ruleEdit{{expr: expression{u: String("io"), op: 2}, binding: 2, object: 34, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("is"), op: 2}, kind: 0, qual: 0, object: 34, op: 24}},
			edits: []ruleEdit{{expr: expression{u: String("is"), op: 2}, binding: 2, object: 34, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("it"), op: 2}, kind: 0, qual: 0, object: 34, op: 24}},
			edits: []ruleEdit{{expr: expression{u: String("it"), op: 2}, binding: 2, object: 34, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("iu"), op: 2}, kind: 0, qual: 0, object: 34, op: 24}},
			edits: []ruleEdit{{expr: expression{u: String("iu"), op: 2}, binding: 2, object: 34, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("ja"), op: 2}, kind: 0, qual: 0, object: 34, op: 24}},
			edits: []ruleEdit{{expr: expression{u: String("ja"), op: 2}, binding: 2, object: 34, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("jv"), op: 2}, kind: 0, qual: 0, object: 34, op: 24}},
			edits: []ruleEdit{{expr: expression{u: String("jv"), op: 2}, binding: 2, object: 34, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("ka"), op: 2}, kind: 0, qual: 0, object: 34, op: 24}},
			edits: []ruleEdit{{expr: expression{u: String("ka"), op: 2}, binding: 2, object: 34, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("kaa"), op: 2}, kind: 0, qual: 0, object: 34, op: 24}},
			edits: []ruleEdit{{expr: expression{u: String("kaa"), op: 2}, binding: 2, object: 34, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("kab"), op: 2}, kind: 0, qual: 0, object: 34, op: 24}},
			edits: []ruleEdit{{expr: expression{u: String("kab"), op: 2}, binding: 2, object: 34, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("ki"), op: 2}, kind: 0, qual: 0, object: 34, op: 24}},
			edits: []ruleEdit{{expr: expression{u: String("ki"), op: 2}, binding: 2, object: 34, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("kj"), op: 2}, kind: 0, qual: 0, object: 34, op: 24}},
			edits: []ruleEdit{{expr: expression{u: String("kj"), op: 2}, binding: 2, object: 34, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("kk"), op: 2}, kind: 0, qual: 0, object: 34, op: 24}},
			edits: []ruleEdit{{expr: expression{u: String("kk"), op: 2}, binding: 2, object: 34, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("kl"), op: 2}, kind: 0, qual: 0, object: 34, op: 24}},
			edits: []ruleEdit{{expr: expression{u: String("kl"), op: 2}, binding: 2, object: 34, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("km"), op: 2}, kind: 0, qual: 0, object: 34, op: 24}},
			edits: []ruleEdit{{expr: expression{u: String("km"), op: 2}, binding: 2, object: 34, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("kn"), op: 2}, kind: 0, qual: 0, object: 34, op: 24}},
			edits: []ruleEdit{{expr: expression{u: String("kn"), op: 2}, binding: 2, object: 34, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("ko"), op: 2}, kind: 0, qual: 0, object: 34, op: 24}},
			edits: []ruleEdit{{expr: expression{u: String("ko"), op: 2}, binding: 2, object: 34, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("kok"), op: 2}, kind: 0, qual: 0, object: 34, op: 24}},
			edits: []ruleEdit{{expr: expression{u: String("kok"), op: 2}, binding: 2, object: 34, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("kr"), op: 2}, kind: 0, qual: 0, object: 34, op: 24}},
			edits: []ruleEdit{{expr: expression{u: String("kr"), op: 2}, binding: 2, object: 34, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("ks"), op: 2}, kind: 0, qual: 0, object: 34, op: 24}},
			edits: []ruleEdit{{expr: expression{u: String("ks"), op: 2}, binding: 2, object: 34, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("kum"), op: 2}, kind: 0, qual: 0, object: 34, op: 24}},
			edits: []ruleEdit{{expr: expression{u: String("kum"), op: 2}, binding: 2, object: 34, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("kv"), op: 2}, kind: 0, qual: 0, object: 34, op: 24}},
			edits: []ruleEdit{{expr: expression{u: String("kv"), op: 2}, binding: 2, object: 34, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("kw"), op: 2}, kind: 0, qual: 0, object: 34, op: 24}},
			edits: []ruleEdit{{expr: expression{u: String("kw"), op: 2}, binding: 2, object: 34, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("kwm"), op: 2}, kind: 0, qual: 0, object: 34, op: 24}},
			edits: []ruleEdit{{expr: expression{u: String("kwm"), op: 2}, binding: 2, object: 34, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("ky"), op: 2}, kind: 0, qual: 0, object: 34, op: 24}},
			edits: []ruleEdit{{expr: expression{u: String("ky"), op: 2}, binding: 2, object: 34, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("la"), op: 2}, kind: 0, qual: 0, object: 34, op: 24}},
			edits: []ruleEdit{{expr: expression{u: String("la"), op: 2}, binding: 2, object: 34, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("lah"), op: 2}, kind: 0, qual: 0, object: 34, op: 24}},
			edits: []ruleEdit{{expr: expression{u: String("lah"), op: 2}, binding: 2, object: 34, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("lb"), op: 2}, kind: 0, qual: 0, object: 34, op: 24}},
			edits: []ruleEdit{{expr: expression{u: String("lb"), op: 2}, binding: 2, object: 34, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("lez"), op: 2}, kind: 0, qual: 0, object: 34, op: 24}},
			edits: []ruleEdit{{expr: expression{u: String("lez"), op: 2}, binding: 2, object: 34, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("lg"), op: 2}, kind: 0, qual: 0, object: 34, op: 24}},
			edits: []ruleEdit{{expr: expression{u: String("lg"), op: 2}, binding: 2, object: 34, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("li"), op: 2}, kind: 0, qual: 0, object: 34, op: 24}},
			edits: []ruleEdit{{expr: expression{u: String("li"), op: 2}, binding: 2, object: 34, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("ln"), op: 2}, kind: 0, qual: 0, object: 34, op: 24}},
			edits: []ruleEdit{{expr: expression{u: String("ln"), op: 2}, binding: 2, object: 34, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("lo"), op: 2}, kind: 0, qual: 0, object: 34, op: 24}},
			edits: []ruleEdit{{expr: expression{u: String("lo"), op: 2}, binding: 2, object: 34, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("lt"), op: 2}, kind: 0, qual: 0, object: 34, op: 24}},
			edits: []ruleEdit{{expr: expression{u: String("lt"), op: 2}, binding: 2, object: 34, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("lv"), op: 2}, kind: 0, qual: 0, object: 34, op: 24}},
			edits: []ruleEdit{{expr: expression{u: String("lv"), op: 2}, binding: 2, object: 34, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("mai"), op: 2}, kind: 0, qual: 0, object: 34, op: 24}},
			edits: []ruleEdit{{expr: expression{u: String("mai"), op: 2}, binding: 2, object: 34, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("mg"), op: 2}, kind: 0, qual: 0, object: 34, op: 24}},
			edits: []ruleEdit{{expr: expression{u: String("mg"), op: 2}, binding: 2, object: 34, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("mh"), op: 2}, kind: 0, qual: 0, object: 34, op: 24}},
			edits: []ruleEdit{{expr: expression{u: String("mh"), op: 2}, binding: 2, object: 34, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("mi"), op: 2}, kind: 0, qual: 0, object: 34, op: 24}},
			edits: []ruleEdit{{expr: expression{u: String("mi"), op: 2}, binding: 2, object: 34, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("mk"), op: 2}, kind: 0, qual: 0, object: 34, op: 24}},
			edits: []ruleEdit{{expr: expression{u: String("mk"), op: 2}, binding: 2, object: 34, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("ml"), op: 2}, kind: 0, qual: 0, object: 34, op: 24}},
			edits: []ruleEdit{{expr: expression{u: String("ml"), op: 2}, binding: 2, object: 34, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("mni"), op: 2}, kind: 0, qual: 0, object: 34, op: 24}},
			edits: []ruleEdit{{expr: expression{u: String("mni"), op: 2}, binding: 2, object: 34, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("mo"), op: 2}, kind: 0, qual: 0, object: 34, op: 24}},
			edits: []ruleEdit{{expr: expression{u: String("mo"), op: 2}, binding: 2, object: 34, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("mr"), op: 2}, kind: 0, qual: 0, object: 34, op: 24}},
			edits: []ruleEdit{{expr: expression{u: String("mr"), op: 2}, binding: 2, object: 34, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("ms"), op: 2}, kind: 0, qual: 0, object: 34, op: 24}},
			edits: []ruleEdit{{expr: expression{u: String("ms"), op: 2}, binding: 2, object: 34, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("mt"), op: 2}, kind: 0, qual: 0, object: 34, op: 24}},
			edits: []ruleEdit{{expr: expression{u: String("mt"), op: 2}, binding: 2, object: 34, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("my"), op: 2}, kind: 0, qual: 0, object: 34, op: 24}},
			edits: []ruleEdit{{expr: expression{u: String("my"), op: 2}, binding: 2, object: 34, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("na"), op: 2}, kind: 0, qual: 0, object: 34, op: 24}},
			edits: []ruleEdit{{expr: expression{u: String("na"), op: 2}, binding: 2, object: 34, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("nb"), op: 2}, kind: 0, qual: 0, object: 34, op: 24}},
			edits: []ruleEdit{{expr: expression{u: String("nb"), op: 2}, binding: 2, object: 34, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("nds"), op: 2}, kind: 0, qual: 0, object: 34, op: 24}},
			edits: []ruleEdit{{expr: expression{u: String("nds"), op: 2}, binding: 2, object: 34, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("ne"), op: 2}, kind: 0, qual: 0, object: 34, op: 24}},
			edits: []ruleEdit{{expr: expression{u: String("ne"), op: 2}, binding: 2, object: 34, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("ng"), op: 2}, kind: 0, qual: 0, object: 34, op: 24}},
			edits: []ruleEdit{{expr: expression{u: String("ng"), op: 2}, binding: 2, object: 34, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("nl"), op: 2}, kind: 0, qual: 0, object: 34, op: 24}},
			edits: []ruleEdit{{expr: expression{u: String("nl"), op: 2}, binding: 2, object: 34, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("nn"), op: 2}, kind: 0, qual: 0, object: 34, op: 24}},
			edits: []ruleEdit{{expr: expression{u: String("nn"), op: 2}, binding: 2, object: 34, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("no"), op: 2}, kind: 0, qual: 0, object: 34, op: 24}},
			edits: []ruleEdit{{expr: expression{u: String("no"), op: 2}, binding: 2, object: 34, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("nqo"), op: 2}, kind: 0, qual: 0, object: 34, op: 24}},
			edits: []ruleEdit{{expr: expression{u: String("nqo"), op: 2}, binding: 2, object: 34, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("nr"), op: 2}, kind: 0, qual: 0, object: 34, op: 24}},
			edits: []ruleEdit{{expr: expression{u: String("nr"), op: 2}, binding: 2, object: 34, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("nso"), op: 2}, kind: 0, qual: 0, object: 34, op: 24}},
			edits: []ruleEdit{{expr: expression{u: String("nso"), op: 2}, binding: 2, object: 34, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("nv"), op: 2}, kind: 0, qual: 0, object: 34, op: 24}},
			edits: []ruleEdit{{expr: expression{u: String("nv"), op: 2}, binding: 2, object: 34, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("ny"), op: 2}, kind: 0, qual: 0, object: 34, op: 24}},
			edits: []ruleEdit{{expr: expression{u: String("ny"), op: 2}, binding: 2, object: 34, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("oc"), op: 2}, kind: 0, qual: 0, object: 34, op: 24}},
			edits: []ruleEdit{{expr: expression{u: String("oc"), op: 2}, binding: 2, object: 34, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("om"), op: 2}, kind: 0, qual: 0, object: 34, op: 24}},
			edits: []ruleEdit{{expr: expression{u: String("om"), op: 2}, binding: 2, object: 34, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("or"), op: 2}, kind: 0, qual: 0, object: 34, op: 24}},
			edits: []ruleEdit{{expr: expression{u: String("or"), op: 2}, binding: 2, object: 34, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("os"), op: 2}, kind: 0, qual: 0, object: 34, op: 24}},
			edits: []ruleEdit{{expr: expression{u: String("os"), op: 2}, binding: 2, object: 34, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("ota"), op: 2}, kind: 0, qual: 0, object: 34, op: 24}},
			edits: []ruleEdit{{expr: expression{u: String("ota"), op: 2}, binding: 2, object: 34, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("pa"), op: 2}, kind: 0, qual: 0, object: 34, op: 24}},
			edits: []ruleEdit{{expr: expression{u: String("pa"), op: 2}, binding: 2, object: 34, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("pl"), op: 2}, kind: 0, qual: 0, object: 34, op: 24}},
			edits: []ruleEdit{{expr: expression{u: String("pl"), op: 2}, binding: 2, object: 34, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("pt"), op: 2}, kind: 0, qual: 0, object: 34, op: 24}},
			edits: []ruleEdit{{expr: expression{u: String("pt"), op: 2}, binding: 2, object: 34, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("qu"), op: 2}, kind: 0, qual: 0, object: 34, op: 24}},
			edits: []ruleEdit{{expr: expression{u: String("qu"), op: 2}, binding: 2, object: 34, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("quz"), op: 2}, kind: 0, qual: 0, object: 34, op: 24}},
			edits: []ruleEdit{{expr: expression{u: String("quz"), op: 2}, binding: 2, object: 34, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("rm"), op: 2}, kind: 0, qual: 0, object: 34, op: 24}},
			edits: []ruleEdit{{expr: expression{u: String("rm"), op: 2}, binding: 2, object: 34, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("rn"), op: 2}, kind: 0, qual: 0, object: 34, op: 24}},
			edits: []ruleEdit{{expr: expression{u: String("rn"), op: 2}, binding: 2, object: 34, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("ro"), op: 2}, kind: 0, qual: 0, object: 34, op: 24}},
			edits: []ruleEdit{{expr: expression{u: String("ro"), op: 2}, binding: 2, object: 34, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("ru"), op: 2}, kind: 0, qual: 0, object: 34, op: 24}},
			edits: []ruleEdit{{expr: expression{u: String("ru"), op: 2}, binding: 2, object: 34, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("rw"), op: 2}, kind: 0, qual: 0, object: 34, op: 24}},
			edits: []ruleEdit{{expr: expression{u: String("rw"), op: 2}, binding: 2, object: 34, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("sa"), op: 2}, kind: 0, qual: 0, object: 34, op: 24}},
			edits: []ruleEdit{{expr: expression{u: String("sa"), op: 2}, binding: 2, object: 34, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("sah"), op: 2}, kind: 0, qual: 0, object: 34, op: 24}},
			edits: []ruleEdit{{expr: expression{u: String("sah"), op: 2}, binding: 2, object: 34, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("sat"), op: 2}, kind: 0, qual: 0, object: 34, op: 24}},
			edits: []ruleEdit{{expr: expression{u: String("sat"), op: 2}, binding: 2, object: 34, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("sc"), op: 2}, kind: 0, qual: 0, object: 34, op: 24}},
			edits: []ruleEdit{{expr: expression{u: String("sc"), op: 2}, binding: 2, object: 34, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("sco"), op: 2}, kind: 0, qual: 0, object: 34, op: 24}},
			edits: []ruleEdit{{expr: expression{u: String("sco"), op: 2}, binding: 2, object: 34, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("sd"), op: 2}, kind: 0, qual: 0, object: 34, op: 24}},
			edits: []ruleEdit{{expr: expression{u: String("sd"), op: 2}, binding: 2, object: 34, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("se"), op: 2}, kind: 0, qual: 0, object: 34, op: 24}},
			edits: []ruleEdit{{expr: expression{u: String("se"), op: 2}, binding: 2, object: 34, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("sel"), op: 2}, kind: 0, qual: 0, object: 34, op: 24}},
			edits: []ruleEdit{{expr: expression{u: String("sel"), op: 2}, binding: 2, object: 34, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("sg"), op: 2}, kind: 0, qual: 0, object: 34, op: 24}},
			edits: []ruleEdit{{expr: expression{u: String("sg"), op: 2}, binding: 2, object: 34, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("sh"), op: 2}, kind: 0, qual: 0, object: 34, op: 24}},
			edits: []ruleEdit{{expr: expression{u: String("sh"), op: 2}, binding: 2, object: 34, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("shs"), op: 2}, kind: 0, qual: 0, object: 34, op: 24}},
			edits: []ruleEdit{{expr: expression{u: String("shs"), op: 2}, binding: 2, object: 34, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("si"), op: 2}, kind: 0, qual: 0, object: 34, op: 24}},
			edits: []ruleEdit{{expr: expression{u: String("si"), op: 2}, binding: 2, object: 34, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("sid"), op: 2}, kind: 0, qual: 0, object: 34, op: 24}},
			edits: []ruleEdit{{expr: expression{u: String("sid"), op: 2}, binding: 2, object: 34, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("sk"), op: 2}, kind: 0, qual: 0, object: 34, op: 24}},
			edits: []ruleEdit{{expr: expression{u: String("sk"), op: 2}, binding: 2, object: 34, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("sl"), op: 2}, kind: 0, qual: 0, object: 34, op: 24}},
			edits: []ruleEdit{{expr: expression{u: String("sl"), op: 2}, binding: 2, object: 34, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("sm"), op: 2}, kind: 0, qual: 0, object: 34, op: 24}},
			edits: []ruleEdit{{expr: expression{u: String("sm"), op: 2}, binding: 2, object: 34, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("sma"), op: 2}, kind: 0, qual: 0, object: 34, op: 24}},
			edits: []ruleEdit{{expr: expression{u: String("sma"), op: 2}, binding: 2, object: 34, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("smj"), op: 2}, kind: 0, qual: 0, object: 34, op: 24}},
			edits: []ruleEdit{{expr: expression{u: String("smj"), op: 2}, binding: 2, object: 34, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("smn"), op: 2}, kind: 0, qual: 0, object: 34, op: 24}},
			edits: []ruleEdit{{expr: expression{u: String("smn"), op: 2}, binding: 2, object: 34, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("sms"), op: 2}, kind: 0, qual: 0, object: 34, op: 24}},
			edits: []ruleEdit{{expr: expression{u: String("sms"), op: 2}, binding: 2, object: 34, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("sn"), op: 2}, kind: 0, qual: 0, object: 34, op: 24}},
			edits: []ruleEdit{{expr: expression{u: String("sn"), op: 2}, binding: 2, object: 34, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("so"), op: 2}, kind: 0, qual: 0, object: 34, op: 24}},
			edits: []ruleEdit{{expr: expression{u: String("so"), op: 2}, binding: 2, object: 34, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("sq"), op: 2}, kind: 0, qual: 0, object: 34, op: 24}},
			edits: []ruleEdit{{expr: expression{u: String("sq"), op: 2}, binding: 2, object: 34, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("sr"), op: 2}, kind: 0, qual: 0, object: 34, op: 24}},
			edits: []ruleEdit{{expr: expression{u: String("sr"), op: 2}, binding: 2, object: 34, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("ss"), op: 2}, kind: 0, qual: 0, object: 34, op: 24}},
			edits: []ruleEdit{{expr: expression{u: String("ss"), op: 2}, binding: 2, object: 34, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("st"), op: 2}, kind: 0, qual: 0, object: 34, op: 24}},
			edits: []ruleEdit{{expr: expression{u: String("st"), op: 2}, binding: 2, object: 34, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("su"), op: 2}, kind: 0, qual: 0, object: 34, op: 24}},
			edits: []ruleEdit{{expr: expression{u: String("su"), op: 2}, binding: 2, object: 34, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("sv"), op: 2}, kind: 0, qual: 0, object: 34, op: 24}},
			edits: []ruleEdit{{expr: expression{u: String("sv"), op: 2}, binding: 2, object: 34, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("sw"), op: 2}, kind: 0, qual: 0, object: 34, op: 24}},
			edits: []ruleEdit{{expr: expression{u: String("sw"), op: 2}, binding: 2, object: 34, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("syr"), op: 2}, kind: 0, qual: 0, object: 34, op: 24}},
			edits: []ruleEdit{{expr: expression{u: String("syr"), op: 2}, binding: 2, object: 34, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("ta"), op: 2}, kind: 0, qual: 0, object: 34, op: 24}},
			edits: []ruleEdit{{expr: expression{u: String("ta"), op: 2}, binding: 2, object: 34, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("te"), op: 2}, kind: 0, qual: 0, object: 34, op: 24}},
			edits: []ruleEdit{{expr: expression{u: String("te"), op: 2}, binding: 2, object: 34, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("tg"), op: 2}, kind: 0, qual: 0, object: 34, op: 24}},
			edits: []ruleEdit{{expr: expression{u: String("tg"), op: 2}, binding: 2, object: 34, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("th"), op: 2}, kind: 0, qual: 0, object: 34, op: 24}},
			edits: []ruleEdit{{expr: expression{u: String("th"), op: 2}, binding: 2, object: 34, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("tig"), op: 2}, kind: 0, qual: 0, object: 34, op: 24}},
			edits: []ruleEdit{{expr: expression{u: String("tig"), op: 2}, binding: 2, object: 34, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("tk"), op: 2}, kind: 0, qual: 0, object: 34, op: 24}},
			edits: []ruleEdit{{expr: expression{u: String("tk"), op: 2}, binding: 2, object: 34, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("tl"), op: 2}, kind: 0, qual: 0, object: 34, op: 24}},
			edits: []ruleEdit{{expr: expression{u: String("tl"), op: 2}, binding: 2, object: 34, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("tn"), op: 2}, kind: 0, qual: 0, object: 34, op: 24}},
			edits: []ruleEdit{{expr: expression{u: String("tn"), op: 2}, binding: 2, object: 34, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("to"), op: 2}, kind: 0, qual: 0, object: 34, op: 24}},
			edits: []ruleEdit{{expr: expression{u: String("to"), op: 2}, binding: 2, object: 34, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("tr"), op: 2}, kind: 0, qual: 0, object: 34, op: 24}},
			edits: []ruleEdit{{expr: expression{u: String("tr"), op: 2}, binding: 2, object: 34, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("ts"), op: 2}, kind: 0, qual: 0, object: 34, op: 24}},
			edits: []ruleEdit{{expr: expression{u: String("ts"), op: 2}, binding: 2, object: 34, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("tt"), op: 2}, kind: 0, qual: 0, object: 34, op: 24}},
			edits: []ruleEdit{{expr: expression{u: String("tt"), op: 2}, binding: 2, object: 34, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("tw"), op: 2}, kind: 0, qual: 0, object: 34, op: 24}},
			edits: []ruleEdit{{expr: expression{u: String("tw"), op: 2}, binding: 2, object: 34, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("ty"), op: 2}, kind: 0, qual: 0, object: 34, op: 24}},
			edits: []ruleEdit{{expr: expression{u: String("ty"), op: 2}, binding: 2, object: 34, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("tyv"), op: 2}, kind: 0, qual: 0, object: 34, op: 24}},
			edits: []ruleEdit{{expr: expression{u: String("tyv"), op: 2}, binding: 2, object: 34, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("ug"), op: 2}, kind: 0, qual: 0, object: 34, op: 24}},
			edits: []ruleEdit{{expr: expression{u: String("ug"), op: 2}, binding: 2, object: 34, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("uk"), op: 2}, kind: 0, qual: 0, object: 34, op: 24}},
			edits: []ruleEdit{{expr: expression{u: String("uk"), op: 2}, binding: 2, object: 34, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("ur"), op: 2}, kind: 0, qual: 0, object: 34, op: 24}},
			edits: []ruleEdit{{expr: expression{u: String("ur"), op: 2}, binding: 2, object: 34, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("uz"), op: 2}, kind: 0, qual: 0, object: 34, op: 24}},
			edits: []ruleEdit{{expr: expression{u: String("uz"), op: 2}, binding: 2, object: 34, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("ve"), op: 2}, kind: 0, qual: 0, object: 34, op: 24}},
			edits: []ruleEdit{{expr: expression{u: String("ve"), op: 2}, binding: 2, object: 34, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("vi"), op: 2}, kind: 0, qual: 0, object: 34, op: 24}},
			edits: []ruleEdit{{expr: expression{u: String("vi"), op: 2}, binding: 2, object: 34, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("vo"), op: 2}, kind: 0, qual: 0, object: 34, op: 24}},
			edits: []ruleEdit{{expr: expression{u: String("vo"), op: 2}, binding: 2, object: 34, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("vot"), op: 2}, kind: 0, qual: 0, object: 34, op: 24}},
			edits: []ruleEdit{{expr: expression{u: String("vot"), op: 2}, binding: 2, object: 34, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("wa"), op: 2}, kind: 0, qual: 0, object: 34, op: 24}},
			edits: []ruleEdit{{expr: expression{u: String("wa"), op: 2}, binding: 2, object: 34, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("wal"), op: 2}, kind: 0, qual: 0, object: 34, op: 24}},
			edits: []ruleEdit{{expr: expression{u: String("wal"), op: 2}, binding: 2, object: 34, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("wen"), op: 2}, kind: 0, qual: 0, object: 34, op: 24}},
			edits: []ruleEdit{{expr: expression{u: String("wen"), op: 2}, binding: 2, object: 34, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("wo"), op: 2}, kind: 0, qual: 0, object: 34, op: 24}},
			edits: []ruleEdit{{expr: expression{u: String("wo"), op: 2}, binding: 2, object: 34, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("xh"), op: 2}, kind: 0, qual: 0, object: 34, op: 24}},
			edits: []ruleEdit{{expr: expression{u: String("xh"), op: 2}, binding: 2, object: 34, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("yap"), op: 2}, kind: 0, qual: 0, object: 34, op: 24}},
			edits: []ruleEdit{{expr: expression{u: String("yap"), op: 2}, binding: 2, object: 34, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("yi"), op: 2}, kind: 0, qual: 0, object: 34, op: 24}},
			edits: []ruleEdit{{expr: expression{u: String("yi"), op: 2}, binding: 2, object: 34, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("yo"), op: 2}, kind: 0, qual: 0, object: 34, op: 24}},
			edits: []ruleEdit{{expr: expression{u: String("yo"), op: 2}, binding: 2, object: 34, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("za"), op: 2}, kind: 0, qual: 0, object: 34, op: 24}},
			edits: []ruleEdit{{expr: expression{u: String("za"), op: 2}, binding: 2, object: 34, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("zu"), op: 2}, kind: 0, qual: 0, object: 34, op: 24}},
			edits: []ruleEdit{{expr: expression{u: String("zu"), op: 2}, binding: 2, object: 34, op: 11}},
		}}, nil, nil}},
		{name: "confs/40-nonlatin.conf", description: "Set substitutions for non-Latin fonts", domain: "", subst: [matchKindEnd][]directive{{{
			tests: []ruleTest{{expr: &expression{u: String("Nazli"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("serif"), op: 2}, binding: 0, object: 1, op: 16}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("Lotoos"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("serif"), op: 2}, binding: 0, object: 1, op: 16}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("Mitra"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("serif"), op: 2}, binding: 0, object: 1, op: 16}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("Ferdosi"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("serif"), op: 2}, binding: 0, object: 1, op: 16}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("Badr"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("serif"), op: 2}, binding: 0, object: 1, op: 16}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("Zar"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("serif"), op: 2}, binding: 0, object: 1, op: 16}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("Titr"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("serif"), op: 2}, binding: 0, object: 1, op: 16}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("Jadid"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("serif"), op: 2}, binding: 0, object: 1, op: 16}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("Kochi Mincho"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("serif"), op: 2}, binding: 0, object: 1, op: 16}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("AR PL SungtiL GB"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("serif"), op: 2}, binding: 0, object: 1, op: 16}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("AR PL Mingti2L Big5"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("serif"), op: 2}, binding: 0, object: 1, op: 16}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("ＭＳ 明朝"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("serif"), op: 2}, binding: 0, object: 1, op: 16}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("NanumMyeongjo"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("serif"), op: 2}, binding: 0, object: 1, op: 16}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("UnBatang"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("serif"), op: 2}, binding: 0, object: 1, op: 16}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("Baekmuk Batang"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("serif"), op: 2}, binding: 0, object: 1, op: 16}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("MgOpen Canonica"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("serif"), op: 2}, binding: 0, object: 1, op: 16}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("Sazanami Mincho"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("serif"), op: 2}, binding: 0, object: 1, op: 16}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("AR PL ZenKai Uni"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("serif"), op: 2}, binding: 0, object: 1, op: 16}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("ZYSong18030"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("serif"), op: 2}, binding: 0, object: 1, op: 16}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("FreeSerif"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("serif"), op: 2}, binding: 0, object: 1, op: 16}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("SimSun"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("serif"), op: 2}, binding: 0, object: 1, op: 16}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("Arshia"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("sans-serif"), op: 2}, binding: 0, object: 1, op: 16}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("Elham"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("sans-serif"), op: 2}, binding: 0, object: 1, op: 16}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("Farnaz"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("sans-serif"), op: 2}, binding: 0, object: 1, op: 16}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("Nasim"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("sans-serif"), op: 2}, binding: 0, object: 1, op: 16}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("Sina"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("sans-serif"), op: 2}, binding: 0, object: 1, op: 16}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("Roya"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("sans-serif"), op: 2}, binding: 0, object: 1, op: 16}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("Koodak"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("sans-serif"), op: 2}, binding: 0, object: 1, op: 16}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("Terafik"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("sans-serif"), op: 2}, binding: 0, object: 1, op: 16}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("Kochi Gothic"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("sans-serif"), op: 2}, binding: 0, object: 1, op: 16}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("AR PL KaitiM GB"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("sans-serif"), op: 2}, binding: 0, object: 1, op: 16}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("AR PL KaitiM Big5"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("sans-serif"), op: 2}, binding: 0, object: 1, op: 16}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("ＭＳ ゴシック"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("sans-serif"), op: 2}, binding: 0, object: 1, op: 16}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("NanumGothic"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("sans-serif"), op: 2}, binding: 0, object: 1, op: 16}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("UnDotum"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("sans-serif"), op: 2}, binding: 0, object: 1, op: 16}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("Baekmuk Dotum"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("sans-serif"), op: 2}, binding: 0, object: 1, op: 16}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("MgOpen Modata"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("sans-serif"), op: 2}, binding: 0, object: 1, op: 16}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("Sazanami Gothic"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("sans-serif"), op: 2}, binding: 0, object: 1, op: 16}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("AR PL ShanHeiSun Uni"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("sans-serif"), op: 2}, binding: 0, object: 1, op: 16}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("ZYSong18030"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("sans-serif"), op: 2}, binding: 0, object: 1, op: 16}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("FreeSans"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("sans-serif"), op: 2}, binding: 0, object: 1, op: 16}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("NSimSun"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("monospace"), op: 2}, binding: 0, object: 1, op: 16}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("ZYSong18030"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("monospace"), op: 2}, binding: 0, object: 1, op: 16}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("NanumGothicCoding"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("monospace"), op: 2}, binding: 0, object: 1, op: 16}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("FreeMono"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("monospace"), op: 2}, binding: 0, object: 1, op: 16}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("Homa"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("fantasy"), op: 2}, binding: 0, object: 1, op: 16}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("Kamran"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("fantasy"), op: 2}, binding: 0, object: 1, op: 16}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("Fantezi"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("fantasy"), op: 2}, binding: 0, object: 1, op: 16}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("Tabassom"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("fantasy"), op: 2}, binding: 0, object: 1, op: 16}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("IranNastaliq"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("cursive"), op: 2}, binding: 0, object: 1, op: 16}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("Nafees Nastaleeq"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("cursive"), op: 2}, binding: 0, object: 1, op: 16}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("Noto Sans Arabic UI"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("system-ui"), op: 2}, binding: 0, object: 1, op: 16}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("Noto Sans Bengali UI"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("system-ui"), op: 2}, binding: 0, object: 1, op: 16}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("Noto Sans Devanagari UI"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("system-ui"), op: 2}, binding: 0, object: 1, op: 16}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("Noto Sans Gujarati UI"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("system-ui"), op: 2}, binding: 0, object: 1, op: 16}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("Noto Sans Gurmukhi UI"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("system-ui"), op: 2}, binding: 0, object: 1, op: 16}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("Noto Sans Kannada UI"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("system-ui"), op: 2}, binding: 0, object: 1, op: 16}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("Noto Sans Khmer UI"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("system-ui"), op: 2}, binding: 0, object: 1, op: 16}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("Noto Sans Lao UI"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("system-ui"), op: 2}, binding: 0, object: 1, op: 16}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("Noto Sans Malayalam UI"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("system-ui"), op: 2}, binding: 0, object: 1, op: 16}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("Noto Sans Myanmar UI"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("system-ui"), op: 2}, binding: 0, object: 1, op: 16}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("Noto Sans Oriya UI"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("system-ui"), op: 2}, binding: 0, object: 1, op: 16}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("Noto Sans Sinhala UI"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("system-ui"), op: 2}, binding: 0, object: 1, op: 16}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("Noto Sans Tamil UI"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("system-ui"), op: 2}, binding: 0, object: 1, op: 16}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("Noto Sans Telugu UI"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("system-ui"), op: 2}, binding: 0, object: 1, op: 16}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("Noto Sans Thai UI"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("system-ui"), op: 2}, binding: 0, object: 1, op: 16}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("Leelawadee UI"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("system-ui"), op: 2}, binding: 0, object: 1, op: 16}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("Nirmala UI"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("system-ui"), op: 2}, binding: 0, object: 1, op: 16}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("Yu Gothic UI"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("system-ui"), op: 2}, binding: 0, object: 1, op: 16}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("Meiryo UI"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("system-ui"), op: 2}, binding: 0, object: 1, op: 16}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("MS UI Gothic"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("system-ui"), op: 2}, binding: 0, object: 1, op: 16}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("Khmer UI"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("system-ui"), op: 2}, binding: 0, object: 1, op: 16}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("Lao UI"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("system-ui"), op: 2}, binding: 0, object: 1, op: 16}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("Microsoft JhengHei UI"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("system-ui"), op: 2}, binding: 0, object: 1, op: 16}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("Microsoft YaHei UI"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("system-ui"), op: 2}, binding: 0, object: 1, op: 16}},
		}}, nil, nil}},
		{name: "confs/45-generic.conf", description: "Set substitutions for emoji/math fonts", domain: "", subst: [matchKindEnd][]directive{{{
			tests: []ruleTest{{expr: &expression{u: String("Noto Color Emoji"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("emoji"), op: 2}, binding: 2, object: 1, op: 16}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("Apple Color Emoji"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("emoji"), op: 2}, binding: 2, object: 1, op: 16}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("Segoe UI Emoji"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("emoji"), op: 2}, binding: 2, object: 1, op: 16}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("Twitter Color Emoji"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("emoji"), op: 2}, binding: 2, object: 1, op: 16}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("EmojiOne Mozilla"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("emoji"), op: 2}, binding: 2, object: 1, op: 16}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("Emoji Two"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("emoji"), op: 2}, binding: 2, object: 1, op: 16}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("JoyPixels"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("emoji"), op: 2}, binding: 2, object: 1, op: 16}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("Emoji One"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("emoji"), op: 2}, binding: 2, object: 1, op: 16}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("Noto Emoji"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("emoji"), op: 2}, binding: 2, object: 1, op: 16}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("Android Emoji"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("emoji"), op: 2}, binding: 2, object: 1, op: 16}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("emoji"), op: 2}, kind: 0, qual: 0, object: 1, op: 22}},
			edits: []ruleEdit{{expr: expression{u: String("und-zsye"), op: 2}, binding: 0, object: 34, op: 14}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("und-zsye"), op: 2}, kind: 0, qual: 0, object: 34, op: 22},
				{expr: &expression{u: String("emoji"), op: 2}, kind: 0, qual: 1, object: 1, op: 23}},
			edits: []ruleEdit{{expr: expression{u: String("emoji"), op: 2}, binding: 1, object: 1, op: 15}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("XITS Math"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("math"), op: 2}, binding: 2, object: 1, op: 16}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("STIX Two Math"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("math"), op: 2}, binding: 2, object: 1, op: 16}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("Cambria Math"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("math"), op: 2}, binding: 2, object: 1, op: 16}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("Latin Modern Math"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("math"), op: 2}, binding: 2, object: 1, op: 16}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("Minion Math"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("math"), op: 2}, binding: 2, object: 1, op: 16}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("Lucida Math"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("math"), op: 2}, binding: 2, object: 1, op: 16}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("Asana Math"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("math"), op: 2}, binding: 2, object: 1, op: 16}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("math"), op: 2}, kind: 0, qual: 0, object: 1, op: 22}},
			edits: []ruleEdit{{expr: expression{u: String("und-zmth"), op: 2}, binding: 0, object: 34, op: 14}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("und-zmth"), op: 2}, kind: 0, qual: 0, object: 34, op: 22},
				{expr: &expression{u: String("math"), op: 2}, kind: 0, qual: 1, object: 1, op: 23}},
			edits: []ruleEdit{{expr: expression{u: String("math"), op: 2}, binding: 1, object: 1, op: 15}},
		}}, nil, nil}},
		{name: "confs/45-latin.conf", description: "Set substitutions for Latin fonts", domain: "", subst: [matchKindEnd][]directive{{{
			tests: []ruleTest{{expr: &expression{u: String("Bitstream Vera Serif"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("serif"), op: 2}, binding: 0, object: 1, op: 16}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("Cambria"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("serif"), op: 2}, binding: 0, object: 1, op: 16}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("Constantia"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("serif"), op: 2}, binding: 0, object: 1, op: 16}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("DejaVu Serif"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("serif"), op: 2}, binding: 0, object: 1, op: 16}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("Elephant"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("serif"), op: 2}, binding: 0, object: 1, op: 16}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("Garamond"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("serif"), op: 2}, binding: 0, object: 1, op: 16}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("Georgia"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("serif"), op: 2}, binding: 0, object: 1, op: 16}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("Liberation Serif"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("serif"), op: 2}, binding: 0, object: 1, op: 16}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("Luxi Serif"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("serif"), op: 2}, binding: 0, object: 1, op: 16}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("MS Serif"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("serif"), op: 2}, binding: 0, object: 1, op: 16}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("Nimbus Roman No9 L"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("serif"), op: 2}, binding: 0, object: 1, op: 16}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("Nimbus Roman"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("serif"), op: 2}, binding: 0, object: 1, op: 16}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("Palatino Linotype"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("serif"), op: 2}, binding: 0, object: 1, op: 16}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("Thorndale AMT"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("serif"), op: 2}, binding: 0, object: 1, op: 16}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("Thorndale"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("serif"), op: 2}, binding: 0, object: 1, op: 16}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("Times New Roman"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("serif"), op: 2}, binding: 0, object: 1, op: 16}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("Times"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("serif"), op: 2}, binding: 0, object: 1, op: 16}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("Albany AMT"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("sans-serif"), op: 2}, binding: 0, object: 1, op: 16}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("Albany"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("sans-serif"), op: 2}, binding: 0, object: 1, op: 16}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("Arial Unicode MS"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("sans-serif"), op: 2}, binding: 0, object: 1, op: 16}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("Arial"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("sans-serif"), op: 2}, binding: 0, object: 1, op: 16}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("Bitstream Vera Sans"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("sans-serif"), op: 2}, binding: 0, object: 1, op: 16}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("Britannic"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("sans-serif"), op: 2}, binding: 0, object: 1, op: 16}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("Calibri"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("sans-serif"), op: 2}, binding: 0, object: 1, op: 16}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("Candara"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("sans-serif"), op: 2}, binding: 0, object: 1, op: 16}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("Century Gothic"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("sans-serif"), op: 2}, binding: 0, object: 1, op: 16}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("Corbel"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("sans-serif"), op: 2}, binding: 0, object: 1, op: 16}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("DejaVu Sans"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("sans-serif"), op: 2}, binding: 0, object: 1, op: 16}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("Helvetica"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("sans-serif"), op: 2}, binding: 0, object: 1, op: 16}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("Haettenschweiler"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("sans-serif"), op: 2}, binding: 0, object: 1, op: 16}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("Liberation Sans"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("sans-serif"), op: 2}, binding: 0, object: 1, op: 16}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("MS Sans Serif"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("sans-serif"), op: 2}, binding: 0, object: 1, op: 16}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("Nimbus Sans L"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("sans-serif"), op: 2}, binding: 0, object: 1, op: 16}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("Nimbus Sans"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("sans-serif"), op: 2}, binding: 0, object: 1, op: 16}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("Luxi Sans"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("sans-serif"), op: 2}, binding: 0, object: 1, op: 16}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("Tahoma"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("sans-serif"), op: 2}, binding: 0, object: 1, op: 16}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("Trebuchet MS"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("sans-serif"), op: 2}, binding: 0, object: 1, op: 16}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("Twentieth Century"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("sans-serif"), op: 2}, binding: 0, object: 1, op: 16}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("Verdana"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("sans-serif"), op: 2}, binding: 0, object: 1, op: 16}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("Andale Mono"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("monospace"), op: 2}, binding: 0, object: 1, op: 16}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("Bitstream Vera Sans Mono"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("monospace"), op: 2}, binding: 0, object: 1, op: 16}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("Consolas"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("monospace"), op: 2}, binding: 0, object: 1, op: 16}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("Courier New"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("monospace"), op: 2}, binding: 0, object: 1, op: 16}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("Courier"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("monospace"), op: 2}, binding: 0, object: 1, op: 16}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("Courier Std"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("monospace"), op: 2}, binding: 0, object: 1, op: 16}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("Cumberland AMT"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("monospace"), op: 2}, binding: 0, object: 1, op: 16}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("Cumberland"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("monospace"), op: 2}, binding: 0, object: 1, op: 16}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("DejaVu Sans Mono"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("monospace"), op: 2}, binding: 0, object: 1, op: 16}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("Fixedsys"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("monospace"), op: 2}, binding: 0, object: 1, op: 16}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("Inconsolata"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("monospace"), op: 2}, binding: 0, object: 1, op: 16}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("Liberation Mono"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("monospace"), op: 2}, binding: 0, object: 1, op: 16}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("Luxi Mono"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("monospace"), op: 2}, binding: 0, object: 1, op: 16}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("Nimbus Mono L"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("monospace"), op: 2}, binding: 0, object: 1, op: 16}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("Nimbus Mono"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("monospace"), op: 2}, binding: 0, object: 1, op: 16}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("Nimbus Mono PS"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("monospace"), op: 2}, binding: 0, object: 1, op: 16}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("Terminal"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("monospace"), op: 2}, binding: 0, object: 1, op: 16}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("Bauhaus Std"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("fantasy"), op: 2}, binding: 0, object: 1, op: 16}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("Cooper Std"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("fantasy"), op: 2}, binding: 0, object: 1, op: 16}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("Copperplate Gothic Std"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("fantasy"), op: 2}, binding: 0, object: 1, op: 16}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("Impact"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("fantasy"), op: 2}, binding: 0, object: 1, op: 16}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("Comic Sans MS"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("cursive"), op: 2}, binding: 0, object: 1, op: 16}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("ITC Zapf Chancery Std"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("cursive"), op: 2}, binding: 0, object: 1, op: 16}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("Zapfino"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("cursive"), op: 2}, binding: 0, object: 1, op: 16}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("Cantarell"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("system-ui"), op: 2}, binding: 0, object: 1, op: 16}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("Noto Sans UI"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("system-ui"), op: 2}, binding: 0, object: 1, op: 16}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("Segoe UI"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("system-ui"), op: 2}, binding: 0, object: 1, op: 16}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("Segoe UI Historic"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("system-ui"), op: 2}, binding: 0, object: 1, op: 16}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("Segoe UI Symbol"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("system-ui"), op: 2}, binding: 0, object: 1, op: 16}},
		}}, nil, nil}},
		{name: "confs/49-sansserif.conf", description: "Add sans-serif to the family when no generic name", domain: "", subst: [matchKindEnd][]directive{{{
			tests: []ruleTest{{expr: &expression{u: String("sans-serif"), op: 2}, kind: 0, qual: 1, object: 1, op: 23},
				{expr: &expression{u: String("serif"), op: 2}, kind: 0, qual: 1, object: 1, op: 23},
				{expr: &expression{u: String("monospace"), op: 2}, kind: 0, qual: 1, object: 1, op: 23}},
			edits: []ruleEdit{{expr: expression{u: String("sans-serif"), op: 2}, binding: 0, object: 1, op: 16}},
		}}, nil, nil}},
		{name: "confs/60-generic.conf", description: "Set preferable fonts for emoji/math fonts", domain: "", subst: [matchKindEnd][]directive{{{
			tests: []ruleTest{{expr: &expression{u: String("und-zsye"), op: 2}, kind: 0, qual: 0, object: 34, op: 22},
				{expr: &expression{u: Bool(1), op: 5}, kind: 0, qual: 1, object: 47, op: 23},
				{expr: &expression{u: Bool(0), op: 5}, kind: 0, qual: 1, object: 47, op: 23}},
			edits: []ruleEdit{{expr: expression{u: Bool(1), op: 5}, binding: 0, object: 47, op: 15}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("emoji"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: exprTree{&expression{u: String("Noto Color Emoji"), op: 2}, &expression{u: exprTree{&expression{u: String("Apple Color Emoji"), op: 2}, &expression{u: exprTree{&expression{u: String("Segoe UI Emoji"), op: 2}, &expression{u: exprTree{&expression{u: String("Twitter Color Emoji"), op: 2}, &expression{u: exprTree{&expression{u: String("EmojiOne Mozilla"), op: 2}, &expression{u: exprTree{&expression{u: String("Emoji Two"), op: 2}, &expression{u: exprTree{&expression{u: String("JoyPixels"), op: 2}, &expression{u: exprTree{&expression{u: String("Emoji One"), op: 2}, &expression{u: exprTree{&expression{u: String("Noto Emoji"), op: 2}, &expression{u: String("Android Emoji"), op: 2}}, op: 36}}, op: 36}}, op: 36}}, op: 36}}, op: 36}}, op: 36}}, op: 36}}, op: 36}}, op: 36}, binding: 2, object: 1, op: 14}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("math"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: exprTree{&expression{u: String("XITS Math"), op: 2}, &expression{u: exprTree{&expression{u: String("STIX Two Math"), op: 2}, &expression{u: exprTree{&expression{u: String("Cambria Math"), op: 2}, &expression{u: exprTree{&expression{u: String("Latin Modern Math"), op: 2}, &expression{u: exprTree{&expression{u: String("Minion Math"), op: 2}, &expression{u: exprTree{&expression{u: String("Lucida Math"), op: 2}, &expression{u: String("Asana Math"), op: 2}}, op: 36}}, op: 36}}, op: 36}}, op: 36}}, op: 36}}, op: 36}, binding: 2, object: 1, op: 14}},
		}}, nil, nil}},
		{name: "confs/60-latin.conf", description: "Set preferable fonts for Latin", domain: "", subst: [matchKindEnd][]directive{{{
			tests: []ruleTest{{expr: &expression{u: String("serif"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: exprTree{&expression{u: String("DejaVu Serif"), op: 2}, &expression{u: exprTree{&expression{u: String("Times New Roman"), op: 2}, &expression{u: exprTree{&expression{u: String("Thorndale AMT"), op: 2}, &expression{u: exprTree{&expression{u: String("Luxi Serif"), op: 2}, &expression{u: exprTree{&expression{u: String("Nimbus Roman No9 L"), op: 2}, &expression{u: exprTree{&expression{u: String("Nimbus Roman"), op: 2}, &expression{u: String("Times"), op: 2}}, op: 36}}, op: 36}}, op: 36}}, op: 36}}, op: 36}}, op: 36}, binding: 0, object: 1, op: 14}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("sans-serif"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: exprTree{&expression{u: String("DejaVu Sans"), op: 2}, &expression{u: exprTree{&expression{u: String("Verdana"), op: 2}, &expression{u: exprTree{&expression{u: String("Arial"), op: 2}, &expression{u: exprTree{&expression{u: String("Albany AMT"), op: 2}, &expression{u: exprTree{&expression{u: String("Luxi Sans"), op: 2}, &expression{u: exprTree{&expression{u: String("Nimbus Sans L"), op: 2}, &expression{u: exprTree{&expression{u: String("Nimbus Sans"), op: 2}, &expression{u: exprTree{&expression{u: String("Helvetica"), op: 2}, &expression{u: exprTree{&expression{u: String("Lucida Sans Unicode"), op: 2}, &expression{u: exprTree{&expression{u: String("BPG Glaho International"), op: 2}, &expression{u: String("Tahoma"), op: 2}}, op: 36}}, op: 36}}, op: 36}}, op: 36}}, op: 36}}, op: 36}}, op: 36}}, op: 36}}, op: 36}}, op: 36}, binding: 0, object: 1, op: 14}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("monospace"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: exprTree{&expression{u: String("DejaVu Sans Mono"), op: 2}, &expression{u: exprTree{&expression{u: String("Inconsolata"), op: 2}, &expression{u: exprTree{&expression{u: String("Andale Mono"), op: 2}, &expression{u: exprTree{&expression{u: String("Courier New"), op: 2}, &expression{u: exprTree{&expression{u: String("Cumberland AMT"), op: 2}, &expression{u: exprTree{&expression{u: String("Luxi Mono"), op: 2}, &expression{u: exprTree{&expression{u: String("Nimbus Mono L"), op: 2}, &expression{u: exprTree{&expression{u: String("Nimbus Mono"), op: 2}, &expression{u: exprTree{&expression{u: String("Nimbus Mono PS"), op: 2}, &expression{u: String("Courier"), op: 2}}, op: 36}}, op: 36}}, op: 36}}, op: 36}}, op: 36}}, op: 36}}, op: 36}}, op: 36}}, op: 36}, binding: 0, object: 1, op: 14}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("fantasy"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: exprTree{&expression{u: String("Impact"), op: 2}, &expression{u: exprTree{&expression{u: String("Copperplate Gothic Std"), op: 2}, &expression{u: exprTree{&expression{u: String("Cooper Std"), op: 2}, &expression{u: String("Bauhaus Std"), op: 2}}, op: 36}}, op: 36}}, op: 36}, binding: 0, object: 1, op: 14}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("cursive"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: exprTree{&expression{u: String("ITC Zapf Chancery Std"), op: 2}, &expression{u: exprTree{&expression{u: String("Zapfino"), op: 2}, &expression{u: String("Comic Sans MS"), op: 2}}, op: 36}}, op: 36}, binding: 0, object: 1, op: 14}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("system-ui"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: exprTree{&expression{u: String("Cantarell"), op: 2}, &expression{u: exprTree{&expression{u: String("Noto Sans UI"), op: 2}, &expression{u: exprTree{&expression{u: String("Segoe UI"), op: 2}, &expression{u: exprTree{&expression{u: String("Segoe UI Historic"), op: 2}, &expression{u: String("Segoe UI Symbol"), op: 2}}, op: 36}}, op: 36}}, op: 36}}, op: 36}, binding: 0, object: 1, op: 14}},
		}}, nil, nil}},
		{name: "confs/65-fonts-persian.conf", description: "", domain: "", subst: [matchKindEnd][]directive{{{
			tests: []ruleTest{{expr: &expression{u: String("Nesf"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("Nesf2"), op: 2}, binding: 2, object: 1, op: 15}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("Nesf2"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("Persian_sansserif_default"), op: 2}, binding: 2, object: 1, op: 15}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("Nazanin"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("Nazli"), op: 2}, binding: 2, object: 1, op: 15}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("Lotus"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("Lotoos"), op: 2}, binding: 2, object: 1, op: 15}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("Yaqut"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("Yaghoot"), op: 2}, binding: 2, object: 1, op: 15}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("Yaghut"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("Yaghoot"), op: 2}, binding: 2, object: 1, op: 15}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("Traffic"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("Terafik"), op: 2}, binding: 2, object: 1, op: 15}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("Ferdowsi"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("Ferdosi"), op: 2}, binding: 2, object: 1, op: 15}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("Fantezy"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("Fantezi"), op: 2}, binding: 2, object: 1, op: 15}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("Jadid"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("Persian_title"), op: 2}, binding: 2, object: 1, op: 15}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("Titr"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("Persian_title"), op: 2}, binding: 2, object: 1, op: 15}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("Kamran"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: exprTree{&expression{u: String("Persian_fantasy"), op: 2}, &expression{u: String("Homa"), op: 2}}, op: 36}, binding: 2, object: 1, op: 15}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("Homa"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: exprTree{&expression{u: String("Persian_fantasy"), op: 2}, &expression{u: String("Kamran"), op: 2}}, op: 36}, binding: 2, object: 1, op: 15}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("Fantezi"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("Persian_fantasy"), op: 2}, binding: 2, object: 1, op: 15}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("Tabassom"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("Persian_fantasy"), op: 2}, binding: 2, object: 1, op: 15}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("Arshia"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("Persian_square"), op: 2}, binding: 2, object: 1, op: 15}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("Nasim"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("Persian_square"), op: 2}, binding: 2, object: 1, op: 15}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("Elham"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: exprTree{&expression{u: String("Persian_square"), op: 2}, &expression{u: String("Farnaz"), op: 2}}, op: 36}, binding: 2, object: 1, op: 15}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("Farnaz"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: exprTree{&expression{u: String("Persian_square"), op: 2}, &expression{u: String("Elham"), op: 2}}, op: 36}, binding: 2, object: 1, op: 15}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("Sina"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("Persian_square"), op: 2}, binding: 2, object: 1, op: 15}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("Persian_title"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: exprTree{&expression{u: String("Titr"), op: 2}, &expression{u: exprTree{&expression{u: String("Jadid"), op: 2}, &expression{u: String("Persian_serif"), op: 2}}, op: 36}}, op: 36}, binding: 2, object: 1, op: 15}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("Persian_fantasy"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: exprTree{&expression{u: String("Homa"), op: 2}, &expression{u: exprTree{&expression{u: String("Kamran"), op: 2}, &expression{u: exprTree{&expression{u: String("Fantezi"), op: 2}, &expression{u: exprTree{&expression{u: String("Tabassom"), op: 2}, &expression{u: String("Persian_square"), op: 2}}, op: 36}}, op: 36}}, op: 36}}, op: 36}, binding: 2, object: 1, op: 15}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("Persian_square"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: exprTree{&expression{u: String("Arshia"), op: 2}, &expression{u: exprTree{&expression{u: String("Elham"), op: 2}, &expression{u: exprTree{&expression{u: String("Farnaz"), op: 2}, &expression{u: exprTree{&expression{u: String("Nasim"), op: 2}, &expression{u: exprTree{&expression{u: String("Sina"), op: 2}, &expression{u: String("Persian_serif"), op: 2}}, op: 36}}, op: 36}}, op: 36}}, op: 36}}, op: 36}, binding: 2, object: 1, op: 15}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("serif"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: exprTree{&expression{u: String("Nazli"), op: 2}, &expression{u: exprTree{&expression{u: String("Lotoos"), op: 2}, &expression{u: exprTree{&expression{u: String("Mitra"), op: 2}, &expression{u: exprTree{&expression{u: String("Ferdosi"), op: 2}, &expression{u: exprTree{&expression{u: String("Badr"), op: 2}, &expression{u: String("Zar"), op: 2}}, op: 36}}, op: 36}}, op: 36}}, op: 36}}, op: 36}, binding: 0, object: 1, op: 15}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("sans-serif"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: exprTree{&expression{u: String("Roya"), op: 2}, &expression{u: exprTree{&expression{u: String("Koodak"), op: 2}, &expression{u: String("Terafik"), op: 2}}, op: 36}}, op: 36}, binding: 0, object: 1, op: 15}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("monospace"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("Terafik"), op: 2}, binding: 0, object: 1, op: 15}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("fantasy"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: exprTree{&expression{u: String("Homa"), op: 2}, &expression{u: exprTree{&expression{u: String("Kamran"), op: 2}, &expression{u: exprTree{&expression{u: String("Fantezi"), op: 2}, &expression{u: String("Tabassom"), op: 2}}, op: 36}}, op: 36}}, op: 36}, binding: 0, object: 1, op: 15}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("cursive"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: exprTree{&expression{u: String("IranNastaliq"), op: 2}, &expression{u: String("Nafees Nastaleeq"), op: 2}}, op: 36}, binding: 0, object: 1, op: 15}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("serif"), op: 2}, kind: 0, qual: 0, object: 1, op: 22},
				{expr: &expression{u: Int(200), op: 0}, kind: 0, qual: 0, object: 8, op: 30},
				{expr: &expression{u: Float(24), op: 1}, kind: 0, qual: 0, object: 10, op: 30}},
			edits: []ruleEdit{{expr: expression{u: String("Titr"), op: 2}, binding: 0, object: 1, op: 14}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("sans-serif"), op: 2}, kind: 0, qual: 0, object: 1, op: 22},
				{expr: &expression{u: Int(200), op: 0}, kind: 0, qual: 0, object: 8, op: 30},
				{expr: &expression{u: Float(24), op: 1}, kind: 0, qual: 0, object: 10, op: 30}},
			edits: []ruleEdit{{expr: expression{u: String("Titr"), op: 2}, binding: 0, object: 1, op: 14}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("Persian_sansserif_default"), op: 2}, kind: 0, qual: 0, object: 1, op: 22},
				{expr: &expression{u: Int(200), op: 0}, kind: 0, qual: 0, object: 8, op: 30},
				{expr: &expression{u: Float(24), op: 1}, kind: 0, qual: 0, object: 10, op: 30}},
			edits: []ruleEdit{{expr: expression{u: String("Titr"), op: 2}, binding: 2, object: 1, op: 14}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("Persian_sansserif_default"), op: 2}, kind: 0, qual: 0, object: 1, op: 22}},
			edits: []ruleEdit{{expr: expression{u: String("Roya"), op: 2}, binding: 2, object: 1, op: 11}},
		}}, {{
			tests: []ruleTest{{expr: &expression{u: String("TURNED-OFF"), op: 2}, kind: 1, qual: 0, object: 14, op: 22},
				{expr: &expression{u: String("farsiweb"), op: 2}, kind: 1, qual: 0, object: 14, op: 22},
				{expr: &expression{u: String("roman"), op: 10}, kind: 1, qual: 0, object: 7, op: 22},
				{expr: &expression{u: String("roman"), op: 10}, kind: 0, qual: 0, object: 7, op: 23}},
			edits: []ruleEdit{{expr: expression{u: exprTree{&expression{u: exprName{object: 32, kind: -1}, op: 9}, &expression{u: exprMatrix{xx: &expression{u: Float(1), op: 1}, xy: &expression{u: Float(-0.2), op: 1}, yx: &expression{u: Float(0), op: 1}, yy: &expression{u: Float(1), op: 1}}, op: 3}}, op: 33}, binding: 0, object: 32, op: 11},
				{expr: expression{u: String("oblique"), op: 10}, binding: 0, object: 7, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("farsiweb"), op: 2}, kind: 1, qual: 0, object: 14, op: 22}},
			edits: []ruleEdit{{expr: expression{u: Bool(0), op: 5}, binding: 0, object: 19, op: 11},
				{expr: expression{u: Bool(0), op: 5}, binding: 0, object: 17, op: 11},
				{expr: expression{u: Bool(0), op: 5}, binding: 0, object: 39, op: 11}},
		}}, {{
			tests: []ruleTest{{expr: &expression{u: String("Elham"), op: 2}, kind: 2, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("farsiweb"), op: 2}, binding: 0, object: 14, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("Homa"), op: 2}, kind: 2, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("farsiweb"), op: 2}, binding: 0, object: 14, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("Koodak"), op: 2}, kind: 2, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("farsiweb"), op: 2}, binding: 0, object: 14, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("Nazli"), op: 2}, kind: 2, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("farsiweb"), op: 2}, binding: 0, object: 14, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("Roya"), op: 2}, kind: 2, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("farsiweb"), op: 2}, binding: 0, object: 14, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("Terafik"), op: 2}, kind: 2, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("farsiweb"), op: 2}, binding: 0, object: 14, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("Titr"), op: 2}, kind: 2, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("farsiweb"), op: 2}, binding: 0, object: 14, op: 11}},
		}}}},
		{name: "confs/65-khmer.conf", description: "", domain: "", subst: [matchKindEnd][]directive{{{
			tests: []ruleTest{{expr: &expression{u: String("serif"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("Khmer OS\""), op: 2}, binding: 0, object: 1, op: 14}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("sans-serif"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("Khmer OS\""), op: 2}, binding: 0, object: 1, op: 14}},
		}}, nil, nil}},
		{name: "confs/65-nonlatin.conf", description: "Set preferable fonts for non-Latin", domain: "", subst: [matchKindEnd][]directive{{{
			tests: []ruleTest{{expr: &expression{u: String("serif"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: exprTree{&expression{u: String("Artsounk"), op: 2}, &expression{u: exprTree{&expression{u: String("BPG UTF8 M"), op: 2}, &expression{u: exprTree{&expression{u: String("Kinnari"), op: 2}, &expression{u: exprTree{&expression{u: String("Norasi"), op: 2}, &expression{u: exprTree{&expression{u: String("Frank Ruehl"), op: 2}, &expression{u: exprTree{&expression{u: String("Dror"), op: 2}, &expression{u: exprTree{&expression{u: String("JG LaoTimes"), op: 2}, &expression{u: exprTree{&expression{u: String("Saysettha Unicode"), op: 2}, &expression{u: exprTree{&expression{u: String("Pigiarniq"), op: 2}, &expression{u: exprTree{&expression{u: String("B Davat"), op: 2}, &expression{u: exprTree{&expression{u: String("B Compset"), op: 2}, &expression{u: exprTree{&expression{u: String("Kacst-Qr"), op: 2}, &expression{u: exprTree{&expression{u: String("Urdu Nastaliq Unicode"), op: 2}, &expression{u: exprTree{&expression{u: String("Raghindi"), op: 2}, &expression{u: exprTree{&expression{u: String("Mukti Narrow"), op: 2}, &expression{u: exprTree{&expression{u: String("malayalam"), op: 2}, &expression{u: exprTree{&expression{u: String("Sampige"), op: 2}, &expression{u: exprTree{&expression{u: String("padmaa"), op: 2}, &expression{u: exprTree{&expression{u: String("Hapax Berbère"), op: 2}, &expression{u: exprTree{&expression{u: String("MS Mincho"), op: 2}, &expression{u: exprTree{&expression{u: String("SimSun"), op: 2}, &expression{u: exprTree{&expression{u: String("PMingLiu"), op: 2}, &expression{u: exprTree{&expression{u: String("WenQuanYi Zen Hei"), op: 2}, &expression{u: exprTree{&expression{u: String("WenQuanYi Bitmap Song"), op: 2}, &expression{u: exprTree{&expression{u: String("AR PL ShanHeiSun Uni"), op: 2}, &expression{u: exprTree{&expression{u: String("AR PL New Sung"), op: 2}, &expression{u: exprTree{&expression{u: String("ZYSong18030"), op: 2}, &expression{u: exprTree{&expression{u: String("HanyiSong"), op: 2}, &expression{u: exprTree{&expression{u: String("MgOpen Canonica"), op: 2}, &expression{u: exprTree{&expression{u: String("Sazanami Mincho"), op: 2}, &expression{u: exprTree{&expression{u: String("IPAMonaMincho"), op: 2}, &expression{u: exprTree{&expression{u: String("IPAMincho"), op: 2}, &expression{u: exprTree{&expression{u: String("Kochi Mincho"), op: 2}, &expression{u: exprTree{&expression{u: String("AR PL SungtiL GB"), op: 2}, &expression{u: exprTree{&expression{u: String("AR PL Mingti2L Big5"), op: 2}, &expression{u: exprTree{&expression{u: String("AR PL Zenkai Uni"), op: 2}, &expression{u: exprTree{&expression{u: String("ＭＳ 明朝"), op: 2}, &expression{u: exprTree{&expression{u: String("ZYSong18030"), op: 2}, &expression{u: exprTree{&expression{u: String("NanumMyeongjo"), op: 2}, &expression{u: exprTree{&expression{u: String("UnBatang"), op: 2}, &expression{u: exprTree{&expression{u: String("Baekmuk Batang"), op: 2}, &expression{u: exprTree{&expression{u: String("KacstQura"), op: 2}, &expression{u: exprTree{&expression{u: String("Frank Ruehl CLM"), op: 2}, &expression{u: exprTree{&expression{u: String("Lohit Bengali"), op: 2}, &expression{u: exprTree{&expression{u: String("Lohit Gujarati"), op: 2}, &expression{u: exprTree{&expression{u: String("Lohit Hindi"), op: 2}, &expression{u: exprTree{&expression{u: String("Lohit Marathi"), op: 2}, &expression{u: exprTree{&expression{u: String("Lohit Maithili"), op: 2}, &expression{u: exprTree{&expression{u: String("Lohit Kashmiri"), op: 2}, &expression{u: exprTree{&expression{u: String("Lohit Konkani"), op: 2}, &expression{u: exprTree{&expression{u: String("Lohit Nepali"), op: 2}, &expression{u: exprTree{&expression{u: String("Lohit Sindhi"), op: 2}, &expression{u: exprTree{&expression{u: String("Lohit Punjabi"), op: 2}, &expression{u: exprTree{&expression{u: String("Lohit Tamil"), op: 2}, &expression{u: exprTree{&expression{u: String("Rachana"), op: 2}, &expression{u: exprTree{&expression{u: String("Lohit Malayalam"), op: 2}, &expression{u: exprTree{&expression{u: String("Lohit Kannada"), op: 2}, &expression{u: exprTree{&expression{u: String("Lohit Telugu"), op: 2}, &expression{u: exprTree{&expression{u: String("Lohit Oriya"), op: 2}, &expression{u: String("LKLUG"), op: 2}}, op: 36}}, op: 36}}, op: 36}}, op: 36}}, op: 36}}, op: 36}}, op: 36}}, op: 36}}, op: 36}}, op: 36}}, op: 36}}, op: 36}}, op: 36}}, op: 36}}, op: 36}}, op: 36}}, op: 36}}, op: 36}}, op: 36}}, op: 36}}, op: 36}}, op: 36}}, op: 36}}, op: 36}}, op: 36}}, op: 36}}, op: 36}}, op: 36}}, op: 36}}, op: 36}}, op: 36}}, op: 36}}, op: 36}}, op: 36}}, op: 36}}, op: 36}}, op: 36}}, op: 36}}, op: 36}}, op: 36}}, op: 36}}, op: 36}}, op: 36}}, op: 36}}, op: 36}}, op: 36}}, op: 36}}, op: 36}}, op: 36}}, op: 36}}, op: 36}}, op: 36}}, op: 36}}, op: 36}}, op: 36}}, op: 36}}, op: 36}}, op: 36}}, op: 36}, binding: 0, object: 1, op: 14}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("sans-serif"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: exprTree{&expression{u: String("Nachlieli"), op: 2}, &expression{u: exprTree{&expression{u: String("Lucida Sans Unicode"), op: 2}, &expression{u: exprTree{&expression{u: String("Yudit Unicode"), op: 2}, &expression{u: exprTree{&expression{u: String("Kerkis"), op: 2}, &expression{u: exprTree{&expression{u: String("ArmNet Helvetica"), op: 2}, &expression{u: exprTree{&expression{u: String("Artsounk"), op: 2}, &expression{u: exprTree{&expression{u: String("BPG UTF8 M"), op: 2}, &expression{u: exprTree{&expression{u: String("Waree"), op: 2}, &expression{u: exprTree{&expression{u: String("Loma"), op: 2}, &expression{u: exprTree{&expression{u: String("Garuda"), op: 2}, &expression{u: exprTree{&expression{u: String("Umpush"), op: 2}, &expression{u: exprTree{&expression{u: String("Saysettha Unicode"), op: 2}, &expression{u: exprTree{&expression{u: String("JG Lao Old Arial"), op: 2}, &expression{u: exprTree{&expression{u: String("GF Zemen Unicode"), op: 2}, &expression{u: exprTree{&expression{u: String("Pigiarniq"), op: 2}, &expression{u: exprTree{&expression{u: String("B Davat"), op: 2}, &expression{u: exprTree{&expression{u: String("B Compset"), op: 2}, &expression{u: exprTree{&expression{u: String("Kacst-Qr"), op: 2}, &expression{u: exprTree{&expression{u: String("Urdu Nastaliq Unicode"), op: 2}, &expression{u: exprTree{&expression{u: String("Raghindi"), op: 2}, &expression{u: exprTree{&expression{u: String("Mukti Narrow"), op: 2}, &expression{u: exprTree{&expression{u: String("malayalam"), op: 2}, &expression{u: exprTree{&expression{u: String("Sampige"), op: 2}, &expression{u: exprTree{&expression{u: String("padmaa"), op: 2}, &expression{u: exprTree{&expression{u: String("Hapax Berbère"), op: 2}, &expression{u: exprTree{&expression{u: String("MS Gothic"), op: 2}, &expression{u: exprTree{&expression{u: String("UmePlus P Gothic"), op: 2}, &expression{u: exprTree{&expression{u: String("Microsoft YaHei"), op: 2}, &expression{u: exprTree{&expression{u: String("Microsoft JhengHei"), op: 2}, &expression{u: exprTree{&expression{u: String("WenQuanYi Zen Hei"), op: 2}, &expression{u: exprTree{&expression{u: String("WenQuanYi Bitmap Song"), op: 2}, &expression{u: exprTree{&expression{u: String("AR PL ShanHeiSun Uni"), op: 2}, &expression{u: exprTree{&expression{u: String("AR PL New Sung"), op: 2}, &expression{u: exprTree{&expression{u: String("MgOpen Modata"), op: 2}, &expression{u: exprTree{&expression{u: String("VL Gothic"), op: 2}, &expression{u: exprTree{&expression{u: String("IPAMonaGothic"), op: 2}, &expression{u: exprTree{&expression{u: String("IPAGothic"), op: 2}, &expression{u: exprTree{&expression{u: String("Sazanami Gothic"), op: 2}, &expression{u: exprTree{&expression{u: String("Kochi Gothic"), op: 2}, &expression{u: exprTree{&expression{u: String("AR PL KaitiM GB"), op: 2}, &expression{u: exprTree{&expression{u: String("AR PL KaitiM Big5"), op: 2}, &expression{u: exprTree{&expression{u: String("AR PL ShanHeiSun Uni"), op: 2}, &expression{u: exprTree{&expression{u: String("AR PL SungtiL GB"), op: 2}, &expression{u: exprTree{&expression{u: String("AR PL Mingti2L Big5"), op: 2}, &expression{u: exprTree{&expression{u: String("ＭＳ ゴシック"), op: 2}, &expression{u: exprTree{&expression{u: String("ZYSong18030"), op: 2}, &expression{u: exprTree{&expression{u: String("TSCu_Paranar"), op: 2}, &expression{u: exprTree{&expression{u: String("NanumGothic"), op: 2}, &expression{u: exprTree{&expression{u: String("UnDotum"), op: 2}, &expression{u: exprTree{&expression{u: String("Baekmuk Dotum"), op: 2}, &expression{u: exprTree{&expression{u: String("Baekmuk Gulim"), op: 2}, &expression{u: exprTree{&expression{u: String("KacstQura"), op: 2}, &expression{u: exprTree{&expression{u: String("Lohit Bengali"), op: 2}, &expression{u: exprTree{&expression{u: String("Lohit Gujarati"), op: 2}, &expression{u: exprTree{&expression{u: String("Lohit Hindi"), op: 2}, &expression{u: exprTree{&expression{u: String("Lohit Marathi"), op: 2}, &expression{u: exprTree{&expression{u: String("Lohit Maithili"), op: 2}, &expression{u: exprTree{&expression{u: String("Lohit Kashmiri"), op: 2}, &expression{u: exprTree{&expression{u: String("Lohit Konkani"), op: 2}, &expression{u: exprTree{&expression{u: String("Lohit Nepali"), op: 2}, &expression{u: exprTree{&expression{u: String("Lohit Sindhi"), op: 2}, &expression{u: exprTree{&expression{u: String("Lohit Punjabi"), op: 2}, &expression{u: exprTree{&expression{u: String("Lohit Tamil"), op: 2}, &expression{u: exprTree{&expression{u: String("Meera"), op: 2}, &expression{u: exprTree{&expression{u: String("Lohit Malayalam"), op: 2}, &expression{u: exprTree{&expression{u: String("Lohit Kannada"), op: 2}, &expression{u: exprTree{&expression{u: String("Lohit Telugu"), op: 2}, &expression{u: exprTree{&expression{u: String("Lohit Oriya"), op: 2}, &expression{u: String("LKLUG"), op: 2}}, op: 36}}, op: 36}}, op: 36}}, op: 36}}, op: 36}}, op: 36}}, op: 36}}, op: 36}}, op: 36}}, op: 36}}, op: 36}}, op: 36}}, op: 36}}, op: 36}}, op: 36}}, op: 36}}, op: 36}}, op: 36}}, op: 36}}, op: 36}}, op: 36}}, op: 36}}, op: 36}}, op: 36}}, op: 36}}, op: 36}}, op: 36}}, op: 36}}, op: 36}}, op: 36}}, op: 36}}, op: 36}}, op: 36}}, op: 36}}, op: 36}}, op: 36}}, op: 36}}, op: 36}}, op: 36}}, op: 36}}, op: 36}}, op: 36}}, op: 36}}, op: 36}}, op: 36}}, op: 36}}, op: 36}}, op: 36}}, op: 36}}, op: 36}}, op: 36}}, op: 36}}, op: 36}}, op: 36}}, op: 36}}, op: 36}}, op: 36}}, op: 36}}, op: 36}}, op: 36}}, op: 36}}, op: 36}}, op: 36}}, op: 36}}, op: 36}}, op: 36}}, op: 36}}, op: 36}, binding: 0, object: 1, op: 14}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("monospace"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: exprTree{&expression{u: String("Miriam Mono"), op: 2}, &expression{u: exprTree{&expression{u: String("VL Gothic"), op: 2}, &expression{u: exprTree{&expression{u: String("IPAMonaGothic"), op: 2}, &expression{u: exprTree{&expression{u: String("IPAGothic"), op: 2}, &expression{u: exprTree{&expression{u: String("Sazanami Gothic"), op: 2}, &expression{u: exprTree{&expression{u: String("Kochi Gothic"), op: 2}, &expression{u: exprTree{&expression{u: String("AR PL KaitiM GB"), op: 2}, &expression{u: exprTree{&expression{u: String("MS Gothic"), op: 2}, &expression{u: exprTree{&expression{u: String("UmePlus Gothic"), op: 2}, &expression{u: exprTree{&expression{u: String("NSimSun"), op: 2}, &expression{u: exprTree{&expression{u: String("MingLiu"), op: 2}, &expression{u: exprTree{&expression{u: String("AR PL ShanHeiSun Uni"), op: 2}, &expression{u: exprTree{&expression{u: String("AR PL New Sung Mono"), op: 2}, &expression{u: exprTree{&expression{u: String("HanyiSong"), op: 2}, &expression{u: exprTree{&expression{u: String("AR PL SungtiL GB"), op: 2}, &expression{u: exprTree{&expression{u: String("AR PL Mingti2L Big5"), op: 2}, &expression{u: exprTree{&expression{u: String("ZYSong18030"), op: 2}, &expression{u: exprTree{&expression{u: String("NanumGothicCoding"), op: 2}, &expression{u: exprTree{&expression{u: String("NanumGothic"), op: 2}, &expression{u: exprTree{&expression{u: String("UnDotum"), op: 2}, &expression{u: exprTree{&expression{u: String("Baekmuk Dotum"), op: 2}, &expression{u: exprTree{&expression{u: String("Baekmuk Gulim"), op: 2}, &expression{u: exprTree{&expression{u: String("TlwgTypo"), op: 2}, &expression{u: exprTree{&expression{u: String("TlwgTypist"), op: 2}, &expression{u: exprTree{&expression{u: String("TlwgTypewriter"), op: 2}, &expression{u: exprTree{&expression{u: String("TlwgMono"), op: 2}, &expression{u: exprTree{&expression{u: String("Hasida"), op: 2}, &expression{u: exprTree{&expression{u: String("GF Zemen Unicode"), op: 2}, &expression{u: exprTree{&expression{u: String("Hapax Berbère"), op: 2}, &expression{u: exprTree{&expression{u: String("Lohit Bengali"), op: 2}, &expression{u: exprTree{&expression{u: String("Lohit Gujarati"), op: 2}, &expression{u: exprTree{&expression{u: String("Lohit Hindi"), op: 2}, &expression{u: exprTree{&expression{u: String("Lohit Marathi"), op: 2}, &expression{u: exprTree{&expression{u: String("Lohit Maithili"), op: 2}, &expression{u: exprTree{&expression{u: String("Lohit Kashmiri"), op: 2}, &expression{u: exprTree{&expression{u: String("Lohit Konkani"), op: 2}, &expression{u: exprTree{&expression{u: String("Lohit Nepali"), op: 2}, &expression{u: exprTree{&expression{u: String("Lohit Sindhi"), op: 2}, &expression{u: exprTree{&expression{u: String("Lohit Punjabi"), op: 2}, &expression{u: exprTree{&expression{u: String("Lohit Tamil"), op: 2}, &expression{u: exprTree{&expression{u: String("Meera"), op: 2}, &expression{u: exprTree{&expression{u: String("Lohit Malayalam"), op: 2}, &expression{u: exprTree{&expression{u: String("Lohit Kannada"), op: 2}, &expression{u: exprTree{&expression{u: String("Lohit Telugu"), op: 2}, &expression{u: exprTree{&expression{u: String("Lohit Oriya"), op: 2}, &expression{u: String("LKLUG"), op: 2}}, op: 36}}, op: 36}}, op: 36}}, op: 36}}, op: 36}}, op: 36}}, op: 36}}, op: 36}}, op: 36}}, op: 36}}, op: 36}}, op: 36}}, op: 36}}, op: 36}}, op: 36}}, op: 36}}, op: 36}}, op: 36}}, op: 36}}, op: 36}}, op: 36}}, op: 36}}, op: 36}}, op: 36}}, op: 36}}, op: 36}}, op: 36}}, op: 36}}, op: 36}}, op: 36}}, op: 36}}, op: 36}}, op: 36}}, op: 36}}, op: 36}}, op: 36}}, op: 36}}, op: 36}}, op: 36}}, op: 36}}, op: 36}}, op: 36}}, op: 36}}, op: 36}}, op: 36}, binding: 0, object: 1, op: 14}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("system-ui"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: exprTree{&expression{u: String("Noto Sans Arabic UI"), op: 2}, &expression{u: exprTree{&expression{u: String("Noto Sans Bengali UI"), op: 2}, &expression{u: exprTree{&expression{u: String("Noto Sans Devanagari UI"), op: 2}, &expression{u: exprTree{&expression{u: String("Noto Sans Gujarati UI"), op: 2}, &expression{u: exprTree{&expression{u: String("Noto Sans Gurmukhi UI"), op: 2}, &expression{u: exprTree{&expression{u: String("Noto Sans Kannada UI"), op: 2}, &expression{u: exprTree{&expression{u: String("Noto Sans Khmer UI"), op: 2}, &expression{u: exprTree{&expression{u: String("Noto Sans Lao UI"), op: 2}, &expression{u: exprTree{&expression{u: String("Noto Sans Malayalam UI"), op: 2}, &expression{u: exprTree{&expression{u: String("Noto Sans Myanmar UI"), op: 2}, &expression{u: exprTree{&expression{u: String("Noto Sans Oriya UI"), op: 2}, &expression{u: exprTree{&expression{u: String("Noto Sans Sinhala UI"), op: 2}, &expression{u: exprTree{&expression{u: String("Noto Sans Tamil UI"), op: 2}, &expression{u: exprTree{&expression{u: String("Noto Sans Telugu UI"), op: 2}, &expression{u: exprTree{&expression{u: String("Noto Sans Thai UI"), op: 2}, &expression{u: exprTree{&expression{u: String("Leelawadee UI"), op: 2}, &expression{u: exprTree{&expression{u: String("Nirmala UI"), op: 2}, &expression{u: exprTree{&expression{u: String("Yu Gothic UI"), op: 2}, &expression{u: exprTree{&expression{u: String("Meiryo UI"), op: 2}, &expression{u: exprTree{&expression{u: String("MS UI Gothic"), op: 2}, &expression{u: exprTree{&expression{u: String("Khmer UI"), op: 2}, &expression{u: exprTree{&expression{u: String("Lao UI"), op: 2}, &expression{u: exprTree{&expression{u: String("Microsoft YaHei UI"), op: 2}, &expression{u: String("Microsoft JhengHei UI"), op: 2}}, op: 36}}, op: 36}}, op: 36}}, op: 36}}, op: 36}}, op: 36}}, op: 36}}, op: 36}}, op: 36}}, op: 36}}, op: 36}}, op: 36}}, op: 36}}, op: 36}}, op: 36}}, op: 36}}, op: 36}}, op: 36}}, op: 36}}, op: 36}}, op: 36}}, op: 36}}, op: 36}, binding: 0, object: 1, op: 14}},
		}}, nil, nil}},
		{name: "confs/69-unifont.conf", description: "", domain: "", subst: [matchKindEnd][]directive{{{
			tests: []ruleTest{{expr: &expression{u: String("serif"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: exprTree{&expression{u: String("FreeSerif"), op: 2}, &expression{u: exprTree{&expression{u: String("Code2000"), op: 2}, &expression{u: String("Code2001"), op: 2}}, op: 36}}, op: 36}, binding: 0, object: 1, op: 14}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("sans-serif"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: exprTree{&expression{u: String("FreeSans"), op: 2}, &expression{u: exprTree{&expression{u: String("Arial Unicode MS"), op: 2}, &expression{u: exprTree{&expression{u: String("Arial Unicode"), op: 2}, &expression{u: exprTree{&expression{u: String("Code2000"), op: 2}, &expression{u: String("Code2001"), op: 2}}, op: 36}}, op: 36}}, op: 36}}, op: 36}, binding: 0, object: 1, op: 14}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("monospace"), op: 2}, kind: 0, qual: 0, object: 1, op: 65558}},
			edits: []ruleEdit{{expr: expression{u: String("FreeMono"), op: 2}, binding: 0, object: 1, op: 14}},
		}}, nil, nil}},
		{name: "confs/70-no-bitmaps.conf", description: "Reject bitmap fonts", domain: "", subst: [matchKindEnd][]directive{nil, nil, nil}},
		{name: "confs/70-yes-bitmaps.conf", description: "Accept bitmap fonts", domain: "", subst: [matchKindEnd][]directive{nil, nil, nil}},
		{name: "confs/80-delicious.conf", description: "", domain: "", subst: [matchKindEnd][]directive{nil, nil, {{
			tests: []ruleTest{{expr: &expression{u: String("Delicious"), op: 2}, kind: 2, qual: 0, object: 1, op: 65558},
				{expr: &expression{u: String("Heavy"), op: 2}, kind: 2, qual: 0, object: 3, op: 22}},
			edits: []ruleEdit{{expr: expression{u: String("heavy"), op: 10}, binding: 0, object: 8, op: 11}},
		}}}},
		{name: "confs/90-synthetic.conf", description: "", domain: "", subst: [matchKindEnd][]directive{nil, {{
			tests: []ruleTest{{expr: &expression{u: String("roman"), op: 10}, kind: 1, qual: 0, object: 7, op: 22},
				{expr: &expression{u: String("roman"), op: 10}, kind: 0, qual: 0, object: 7, op: 23}},
			edits: []ruleEdit{{expr: expression{u: exprTree{&expression{u: exprName{object: 32, kind: -1}, op: 9}, &expression{u: exprMatrix{xx: &expression{u: Float(1), op: 1}, xy: &expression{u: Float(0.2), op: 1}, yx: &expression{u: Float(0), op: 1}, yy: &expression{u: Float(1), op: 1}}, op: 3}}, op: 33}, binding: 0, object: 32, op: 11},
				{expr: expression{u: String("oblique"), op: 10}, binding: 0, object: 7, op: 11},
				{expr: expression{u: Bool(0), op: 5}, binding: 0, object: 39, op: 11}},
		}, {
			tests: []ruleTest{{expr: &expression{u: String("medium"), op: 10}, kind: 1, qual: 0, object: 8, op: 28},
				{expr: &expression{u: String("bold"), op: 10}, kind: 0, qual: 0, object: 8, op: 30}},
			edits: []ruleEdit{{expr: expression{u: Bool(1), op: 5}, binding: 0, object: 38, op: 11},
				{expr: expression{u: String("bold"), op: 10}, binding: 0, object: 8, op: 11}},
		}}, nil}}},
	customObjects:  map[string]Object{"pixelsizefixupfactor": 0x4a, "scalingnotneeded": 0x4b},
	acceptGlobs:    map[string]bool{},
	rejectGlobs:    map[string]bool{},
	acceptPatterns: Fontset{Pattern{25: &valueList{valueElt{Value: Bool(0), Binding: 1}}}},
	rejectPatterns: Fontset{Pattern{25: &valueList{valueElt{Value: Bool(0), Binding: 1}}}},
	maxObjects:     22,
}