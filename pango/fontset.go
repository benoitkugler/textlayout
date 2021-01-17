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
// The concretes types implementing this interface MUST be valid map keys.
type Fontset interface {
	// Returns the language of the fontset
	GetLanguage() Language

	// Iterates through all the fonts in a fontset, calling `fn` for each one.
	// If `fn` returns `true`, that stops the iteration.
	Foreach(fn FontsetForeachFunc)

	// // Returns the font in the fontset that contains the best glyph for the Unicode character `wc`.
	// GetFont(wc rune) Font
}

// Returns `true` stops the iteration
type FontsetForeachFunc = func(font Font) bool

func getFontCache(fontset Fontset) *FontCache {
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

// uses loadSet and select the first font
func loadFont(fontmap FontMap, context *Context, description *FontDescription) Font {
	var language Language
	if context != nil {
		language = context.GetLanguage()
	}

	fontset := fontmap.LoadFontset(context, description, language)
	if fontset == nil {
		return nil
	}

	var outFont Font
	fontset.Foreach(func(font Font) bool { // select the first font and stops
		outFont = font
		return true
	})

	return outFont
}

// FontMap represents the set of fonts available for a
// particular rendering system. It is the top-level object
// of all font access. From a FontMap, a font set is loaded.
//
// The concretes types implementing this interface MUST be valid map keys.
type FontMap interface {
	// Load a set of fonts in the fontmap that can be used to render a font matching `desc`.
	// Returns nil if no font matched.
	LoadFontset(context *Context, desc *FontDescription, language Language) Fontset

	// const char     *shape_engine_type; the type of rendering-system-dependent engines that can handle fonts of this fonts loaded with this fontmap.

	// Returns the current serial number of the fontmap.  The serial number is
	// initialized to an small number larger than zero when a new fontmap
	// is created and is increased whenever the fontmap is changed. It may
	// wrap, but will never have the value 0. Since it can wrap, never compare
	// it with "less than", always use "not equals".
	//
	// The fontmap can only be changed using backend-specific API, like changing
	// fontmap resolution.
	GetSerial() uint

	// // Gets the FontFace to which `font` belongs.
	// // The concrete type of `font` is guarenteed to be consitent.
	// GetFace(font Font) FontFace

	// // Forces a change in the context, which will cause any Context
	// // using this fontmap to change.
	// //
	// // This function is only useful when implementing a new backend
	// // for Pango, something applications won't do. Backends should
	// // call this function if they have attached extra data to the context
	// // and such data is Changed.
	// Changed()

	// // List all available families. This method is mainly
	// // used in debugging and testing.
	// ListFamilies() []FontFamily

	// // Gets a font family by name. This method is mainly
	// // used in debugging and testing.
	// GetFamily(name string) FontFamily

	// // Loads the font in the fontmap that is the closest match for `desc`.
	// // Returns nil if no font matched.
	// LoadFont(context *Context, desc *FontDescription) Font
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
