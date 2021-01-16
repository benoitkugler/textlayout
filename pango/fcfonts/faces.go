package fcfonts

import (
	"sort"
	"strings"

	"github.com/benoitkugler/textlayout/fontconfig"
	"github.com/benoitkugler/textlayout/pango"
)

var (
	_ pango.FontFace   = (*PangoFcFace)(nil)
	_ pango.FontFamily = (*PangoFcFamily)(nil)
)

type PangoFcFace struct {
	family  *PangoFcFamily
	style   string
	pattern fontconfig.Pattern

	fake    bool
	regular bool

	// parent_instance FontFace
}

func newFace(family *PangoFcFamily, style string, pattern fontconfig.Pattern, fake, regular bool) *PangoFcFace {
	var face PangoFcFace
	face.style = style
	face.pattern = pattern
	face.family = family
	face.fake = fake
	face.regular = regular
	return &face
}

func (face *PangoFcFace) Describe() pango.FontDescription {
	family := face.family

	if family == nil {
		return pango.NewFontDescription()
	}

	if face.fake {
		switch face.style {
		case "Regular":
			return family.makeAliasDescription(false, false)
		case "Bold":
			return family.makeAliasDescription(true, false)
		case "Italic":
			return family.makeAliasDescription(false, true)
		default: /* Bold Italic */
			return family.makeAliasDescription(true, true)
		}
	}
	return fdFromPattern(face.pattern, false)
}

// fdFromPattern creates a `FontDescription` that matches the specified
// 'fontconfig' pattern as closely as possible. Many possible Fontconfig
// pattern values don't make sense in the context of `FontDescription`, so will be ignored.
// If `includeSize` is true, the description will include the size from
// the pattern; otherwise the resulting description will be unsized.
// (only `FC_SIZE` is examined, not `FC_PIXEL_SIZE`)
func fdFromPattern(pattern fontconfig.Pattern, includeSize bool) pango.FontDescription {
	desc := pango.NewFontDescription()

	s, _ := pattern.FcPatternObjectGetString(fontconfig.FC_FAMILY, 0)
	desc.Setfamily(s)

	style := pango.STYLE_NORMAL
	if slant, ok := pattern.GetInt(fontconfig.FC_SLANT); ok {
		switch slant {
		case fontconfig.SLANT_ROMAN:
			style = pango.STYLE_NORMAL
		case fontconfig.SLANT_ITALIC:
			style = pango.STYLE_ITALIC
		case fontconfig.SLANT_OBLIQUE:
			style = pango.STYLE_OBLIQUE
		default:
			style = pango.STYLE_NORMAL
		}
	}
	desc.Setstyle(style)

	weight := pango.PANGO_WEIGHT_NORMAL
	if ws := pattern.GetFloats(fontconfig.FC_WEIGHT); len(ws) != 0 {
		weight = pango.Weight(fontconfig.FcWeightToOpenTypeDouble(ws[0]))
	}
	desc.Setweight(weight)

	stretch := pango.STRETCH_NORMAL
	if w, ok := pattern.GetInt(fontconfig.FC_WIDTH); ok {
		switch w {
		case fontconfig.WIDTH_NORMAL:
			stretch = pango.STRETCH_NORMAL
		case fontconfig.WIDTH_ULTRACONDENSED:
			stretch = pango.STRETCH_ULTRA_CONDENSED
		case fontconfig.WIDTH_EXTRACONDENSED:
			stretch = pango.STRETCH_EXTRA_CONDENSED
		case fontconfig.WIDTH_CONDENSED:
			stretch = pango.STRETCH_CONDENSED
		case fontconfig.WIDTH_SEMICONDENSED:
			stretch = pango.STRETCH_SEMI_CONDENSED
		case fontconfig.WIDTH_SEMIEXPANDED:
			stretch = pango.STRETCH_SEMI_EXPANDED
		case fontconfig.WIDTH_EXPANDED:
			stretch = pango.STRETCH_EXPANDED
		case fontconfig.WIDTH_EXTRAEXPANDED:
			stretch = pango.STRETCH_EXTRA_EXPANDED
		case fontconfig.WIDTH_ULTRAEXPANDED:
			stretch = pango.STRETCH_ULTRA_EXPANDED
		default:
			stretch = pango.STRETCH_NORMAL
		}
	}
	desc.Setstretch(stretch)

	desc.Setvariant(pango.PANGO_VARIANT_NORMAL)

	if size, ok := pattern.GetFloat(fontconfig.FC_SIZE); includeSize && ok {
		desc.SetSize(int(size * float64(pango.PangoScale)))
	}

	// gravity is a bit different.  we don't want to set it if it was not set on the pattern
	if s, res := pattern.FcPatternObjectGetString(fcGravity, 0); res == fontconfig.FcResultMatch {
		gravity, _ := pango.GravityMap.FromString(s)
		desc.Setgravity(pango.Gravity(gravity))
	}

	if s, _ := pattern.FcPatternObjectGetString(fcFontVariations, 0); includeSize && s != "" {
		desc.Setvariations(s)
	}

	return desc
}

func (face *PangoFcFace) ListSizes() []int {
	if face.family == nil || face.family.fontmap == nil {
		return nil
	}

	pattern := fontconfig.NewFcPattern()
	pattern.Add(fontconfig.FC_FAMILY, fontconfig.String(face.family.family_name), true)
	pattern.Add(fontconfig.FC_STYLE, fontconfig.String(face.style), true)

	fontset := fontconfig.List(nil, pattern, fontconfig.FC_PIXEL_SIZE)

	var (
		dpi = -1.
		out []int
	)
	for _, font := range fontset {
		for _, size := range font.GetFloats(fontconfig.FC_PIXEL_SIZE) {
			if dpi < 0 {
				dpi = face.family.fontmap.GetResolution(nil)
			}
			sizeI := int(float64(pango.PangoScale) * size * 72.0 / dpi)
			out = append(out, sizeI)
		}
	}

	sort.Ints(out)

	return out
}

func (face *PangoFcFace) GetFaceName() string         { return face.style }
func (face *PangoFcFace) IsSynthesized() bool         { return face.fake }
func (face *PangoFcFace) GetFamily() pango.FontFamily { return face.family }

type PangoFcFamily struct {
	// parent_instance FontFamily

	fontmap     *PangoFcFontMap
	family_name string

	patterns fontconfig.FcFontSet
	// nil means not initialized
	faces []*PangoFcFace

	spacing  int /* FC_SPACING */
	variable bool
}

func (family *PangoFcFamily) makeAliasDescription(bold, italic bool) pango.FontDescription {
	out := pango.NewFontDescription()

	out.Setfamily(family.family_name)

	out.Setstyle(pango.STYLE_NORMAL)
	if italic {
		out.Setstyle(pango.STYLE_ITALIC)
	}
	out.Setweight(pango.PANGO_WEIGHT_NORMAL)
	if bold {
		out.Setweight(pango.PANGO_WEIGHT_BOLD)
	}

	return out
}

func compareFace(f1, f2 *PangoFcFace) bool {
	w1, ok := f1.pattern.GetInt(fontconfig.FC_WEIGHT)
	if !ok {
		w1 = fontconfig.WEIGHT_MEDIUM
	}

	s1, ok := f1.pattern.GetInt(fontconfig.FC_SLANT)
	if !ok {
		s1 = fontconfig.SLANT_ROMAN
	}

	w2, ok := f2.pattern.GetInt(fontconfig.FC_WEIGHT)
	if !ok {
		w2 = fontconfig.WEIGHT_MEDIUM
	}

	s2, ok := f2.pattern.GetInt(fontconfig.FC_SLANT)
	if !ok {
		s2 = fontconfig.SLANT_ROMAN
	}

	if s1 != s2 {
		// roman < italic < oblique
		return s1 < s2
	}

	return w1 < w2 // from light to heavy
}

func is_alias_family(familyName string) bool {
	if familyName == "" {
		return false
	}
	switch familyName[0] {
	case 'c', 'C':
		return strings.EqualFold(familyName, "cursive")
	case 'f', 'F':
		return strings.EqualFold(familyName, "fantasy")
	case 'm', 'M':
		return strings.EqualFold(familyName, "monospace")
	case 's', 'S':
		return strings.EqualFold(familyName, "sans") ||
			strings.EqualFold(familyName, "serif") ||
			strings.EqualFold(familyName, "system-ui")
	}

	return false
}

func (family *PangoFcFamily) ensureFaces() {
	if family.faces != nil { // already initialized
		return
	}

	if is_alias_family(family.family_name) || family.fontmap.Priv.Closed {
		family.faces = []*PangoFcFace{
			newFace(family, "Regular", nil, true, true),
			newFace(family, "Bold", nil, true, false),
			newFace(family, "Italic", nil, true, false),
			newFace(family, "Bold Italic", nil, true, false),
		}
		return
	}

	const (
		REGULAR = iota
		ITALIC
		BOLD
		BOLD_ITALIC
	)

	var hasFace [4]bool // Regular, Italic, Bold, Bold Italic

	fontset := family.patterns

	// at most we have 3 additional artifical faces
	faces := make([]*PangoFcFace, 0, len(fontset)+3)

	regularWeight := 0

	for _, font := range fontset {
		weight := fontconfig.WEIGHT_MEDIUM
		if i, ok := font.GetInt(fontconfig.FC_WEIGHT); ok {
			weight = i
		}

		slant := fontconfig.SLANT_ROMAN
		if i, ok := font.GetInt(fontconfig.FC_SLANT); ok {
			slant = i
		}

		variable, res := font.FcPatternObjectGetBool(fontconfig.FC_VARIABLE, 0)
		if res != fontconfig.FcResultMatch {
			variable = 0
		}
		if variable != 0 /* skip the variable face */ {
			continue
		}

		var isRegular bool

		fontStyle, _ := font.FcPatternObjectGetString(fontconfig.FC_STYLE, 0)
		if fontStyle == "Regular" {
			regularWeight = fontconfig.WEIGHT_MEDIUM
			isRegular = true
		}

		var style string
		if weight <= fontconfig.WEIGHT_MEDIUM {
			if slant == fontconfig.SLANT_ROMAN {
				hasFace[REGULAR] = true
				style = "Regular"
				if weight > regularWeight {
					regularWeight = weight
					isRegular = true
				}
			} else {
				hasFace[ITALIC] = true
				style = "Italic"
			}
		} else {
			if slant == fontconfig.SLANT_ROMAN {
				hasFace[BOLD] = true
				style = "Bold"
			} else {
				hasFace[BOLD_ITALIC] = true
				style = "Bold Italic"
			}
		}

		if fontStyle == "" {
			fontStyle = style
		}
		faces = append(faces, newFace(family, fontStyle, font, false, isRegular))
	}

	if hasFace[REGULAR] {
		if !hasFace[ITALIC] {
			faces = append(faces, newFace(family, "Italic", nil, true, false))
		}
		if !hasFace[BOLD] {
			faces = append(faces, newFace(family, "Bold", nil, true, false))
		}
	}
	if (hasFace[REGULAR] || hasFace[ITALIC] || hasFace[BOLD]) && !hasFace[BOLD_ITALIC] {
		faces = append(faces, newFace(family, "Bold Italic", nil, true, false))
	}

	sort.Slice(faces, func(i, j int) bool { return compareFace(faces[i], faces[j]) })

	family.faces = faces // now != nil
}

func (family *PangoFcFamily) ListFaces() []pango.FontFace {

	family.ensureFaces()

	out := make([]pango.FontFace, len(family.faces)) // shallow copy
	for i, f := range family.faces {
		out[i] = f
	}
	return out
}

func (family *PangoFcFamily) GetFace(name string) pango.FontFace {

	family.ensureFaces()

	for _, face := range family.faces {
		if name == face.GetFaceName() || (name == "" && face.regular) {
			return face
		}
	}

	return nil
}

func (family *PangoFcFamily) GetName() string { return family.family_name }

func (family *PangoFcFamily) IsMonospace() bool {
	return family.spacing == fontconfig.FC_MONO ||
		family.spacing == fontconfig.FC_DUAL ||
		family.spacing == fontconfig.FC_CHARCELL
}

func (family *PangoFcFamily) IsVariable() bool { return family.variable }
