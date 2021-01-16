package pango

import "sync"

var (
	fontsetCaches     = map[Fontset]*FontCache{}
	fontsetCachesLock sync.Mutex

	// TODO only one map per context ?
	// it only for warnings anyway, probably not a big deal...
	fontmapScriptWarnings = map[struct {
		ft     FontMap
		script Script
	}]bool{}
	fontmapScriptWarningsLock sync.Mutex

	fontShapeFailWarnings     = map[Font]bool{}
	fontShapeFailWarningsLock sync.Mutex
)

// Fontset represents a set of Font to use when rendering text.
// It is the result of resolving a FontDescription against a particular Context.
// The concretes types implementing this interface shouls be pointers, since
// they will be used as map keys: they MUST at least be comparable types.
type Fontset interface {
	// Returns the font in the fontset that contains the best glyph for the Unicode character `wc`.
	get_font(wc rune) Font
	// Get overall metric information for the fonts in the fontset.
	get_metrics() FontMetrics
	// Returns the language of the fontset
	get_language() Language
	// Iterates through all the fonts in a fontset, calling `fn` for each one.
	// If `fn` returns `true`, that stops the iteration.
	foreach(fn FontsetForeachFunc)
}

// Returns `true` stops the iteration
type FontsetForeachFunc = func(font Font) bool

func get_font_cache(fontset Fontset) *FontCache {
	fontsetCachesLock.Lock()
	defer fontsetCachesLock.Unlock()

	cache := fontsetCaches[fontset]
	if cache != nil {
		return cache
	}
	cache = NewFontCache()
	fontsetCaches[fontset] = cache
	return cache
}

// FontMap represents the set of fonts available for a
// particular rendering system.
// The concretes types implementing this interface shouls be pointers, since
// they will be used as map keys: they MUST at least be comparable types.
type FontMap interface {
	// Loads the font in the fontmap that is the closest match for `desc`.
	// Returns nil if no font matched.
	load_font(context *Context, desc *FontDescription) Font
	// List all available families
	list_families() []*FontFamily
	// Load a set of fonts in the fontmap that can be used to render a font matching `desc`.
	// Returns nil if no font matched.
	load_fontset(context *Context, desc *FontDescription, language Language) Fontset

	// const char     *shape_engine_type; the type of rendering-system-dependent engines that can handle fonts of this fonts loaded with this fontmap.

	// Returns the current serial number of the fontmap.  The serial number is
	// initialized to an small number larger than zero when a new fontmap
	// is created and is increased whenever the fontmap is changed. It may
	// wrap, but will never have the value 0. Since it can wrap, never compare
	// it with "less than", always use "not equals".
	//
	// The fontmap can only be changed using backend-specific API, like changing
	// fontmap resolution.
	get_serial() uint

	// Forces a change in the context, which will cause any Context
	// using this fontmap to change.
	//
	// This function is only useful when implementing a new backend
	// for Pango, something applications won't do. Backends should
	// call this function if they have attached extra data to the context
	// and such data is changed.
	changed()

	// Gets a font family by name.
	get_family(name string) *FontFamily

	// Gets the FontFace to which `font` belongs.
	get_face(font Font) *FontFace
}

// return true if not already warned, and keep track of the anwser
func shouldWarn(fontmap FontMap, script Script) bool {
	fontmapScriptWarningsLock.Lock()
	defer fontmapScriptWarningsLock.Unlock()

	key := struct {
		ft     FontMap
		script Script
	}{
		ft:     fontmap,
		script: script,
	}
	if fontmapScriptWarnings[key] {
		return false
	}
	fontmapScriptWarnings[key] = true
	return true
}
