package pango

import (
	"fmt"
	"log"
	"unicode"
)

// itemizeWithBaseDir is like `Itemize`, but the base direction to use when
// computing bidirectional levels is specified explicitly rather than gotten from `context`.
//
// `cachedIter` should be an iterator over `attrs` currently positioned at a
// range before or containing `startIndex`; `cachedIter` will be advanced to
// the range covering the position just after `startIndex` + `length`.
// (i.e. if itemizing in a loop, just keep passing in the same `cachedIter`).
func (context *Context) itemizeWithBaseDir(baseDir Direction, text []rune,
	startIndex, length int,
	attrs AttrList, cachedIter *attrIterator) *ItemList {

	if context == nil || len(text) == 0 || length == 0 {
		return nil
	}
	items := context.itemizeWithFont(text[:startIndex+length], startIndex, nil, baseDir, attrs, cachedIter)
	return context.postProcessItems(text, nil, items)
}

// We don't want space characters to affect font selection; in general,
// it's always wrong to select a font just to render a space.
// We assume that all fonts have the ASCII space, and for other space
// characters if they don't, HarfBuzz will compatibility-decompose them
// to ASCII space...
// See bugs #355987 and #701652.
//
// We don't want to change fonts just for variation selectors.
// See bug #781123.
//
// Finally, don't change fonts for line or paragraph separators.
//
// Note that we want spaces to use the 'better' font, comparing
// the font that is used before and after the space. This is handled
// in addCharacter().
func considerAsSpace(wc rune) bool {
	isIn := unicode.In(wc, unicode.Cc, unicode.Cf, unicode.Cs, unicode.Zl, unicode.Zp)
	return isIn || (unicode.Is(unicode.Zs, wc) && wc != '\u1680' /* OGHAM SPACE MARK */) ||
		(wc >= '\ufe00' && wc <= '\ufe0f') || (wc >= '\U000e0100' && wc <= '\U000e01ef')
}

func (state *itemizeState) processRun() {
	lastWasForcedBreak := false

	state.updateForNewRun()

	if debugMode { //  We should never get an empty run
		assert(state.runEnd > state.runStart, fmt.Sprintf("processRun: %d <= %d", state.runEnd, state.runStart))
	}

	for pos, wc := range state.text[state.runStart:state.runEnd] {
		isForcedBreak := (wc == '\t' || wc == lineSeparator)
		var (
			font    fontElement
			isSpace bool
		)
		if considerAsSpace(wc) {
			font.font = nil
			font.position = 0xFFFF
			isSpace = true
		} else {
			font, _ = state.getFont(wc)
			isSpace = false
		}

		state.addCharacter(font.font, font.position, isForcedBreak || lastWasForcedBreak, pos+state.runStart, isSpace)

		lastWasForcedBreak = isForcedBreak
	}

	/* Finish the final item from the current segment */
	state.item.Length = state.runEnd - state.item.Offset
	if state.item.Analysis.Font == nil {
		font, ok := state.getFont(' ')
		if !ok {
			// only warn once per fontmap/script pair
			if shouldWarn(state.context.fontMap, state.script) {
				log.Printf("failed to choose a font for script %s: expect ugly output", state.script)
			}
		}
		state.fillFont(font.font)
	}
	state.item = nil
}
