package pango

import (
	"github.com/benoitkugler/textlayout/language"
)

/**
 * SECTION:scripts
 * @short_description:Identifying writing systems and languages
 * @title:Scripts and Languages
 *
 * The functions in this section are used to identify the writing
 * system, or <firstterm>script</firstterm> of individual characters
 * and of ranges within a larger text string.
 */

type Script = language.Script

const parenStackDepth = 128

type parenStackEntry struct {
	pairIndex  int
	scriptCode Script
}

// scriptIter is used to break a string of
// Unicode text into runs by Unicode script.
//
// The range to which the iterator currently points,
// is the set of indexes `i` where `script_start` <= i < `script_end`.
// That is, it doesn't include the character stored at `script_end`.
type scriptIter struct {
	text []rune

	textEnd     int // limit of a run
	scriptStart int // index into text
	scriptEnd   int // index into text
	scriptCode  Script

	parenStack [parenStackDepth]parenStackEntry
	parenSp    int
}

// newScriptIter creates a new script iterator,initialized to point at the first range in the text.
// If the string is empty, it will point at an empty range.
func newScriptIter(text []rune) *scriptIter {
	var out scriptIter
	out.reset(text, 0, len(text))
	return &out
}

func (iter *scriptIter) reset(text []rune, textStart, length int) {
	iter.text = text
	iter.textEnd = textStart + length
	iter.scriptStart, iter.scriptEnd = textStart, textStart
	iter.scriptCode = language.Common
	iter.parenSp = -1
	iter.next()
}

var paired_chars = [...]rune{
	0x0028, 0x0029, /* ascii paired punctuation */
	0x003c, 0x003e,
	0x005b, 0x005d,
	0x007b, 0x007d,
	0x00ab, 0x00bb, /* guillemets */
	0x0f3a, 0x0f3b, /* tibetan */
	0x0f3c, 0x0f3d,
	0x169b, 0x169c, /* ogham */
	0x2018, 0x2019, /* general punctuation */
	0x201c, 0x201d,
	0x2039, 0x203a,
	0x2045, 0x2046,
	0x207d, 0x207e,
	0x208d, 0x208e,
	0x27e6, 0x27e7, /* math */
	0x27e8, 0x27e9,
	0x27ea, 0x27eb,
	0x27ec, 0x27ed,
	0x27ee, 0x27ef,
	0x2983, 0x2984,
	0x2985, 0x2986,
	0x2987, 0x2988,
	0x2989, 0x298a,
	0x298b, 0x298c,
	0x298d, 0x298e,
	0x298f, 0x2990,
	0x2991, 0x2992,
	0x2993, 0x2994,
	0x2995, 0x2996,
	0x2997, 0x2998,
	0x29fc, 0x29fd,
	0x2e02, 0x2e03,
	0x2e04, 0x2e05,
	0x2e09, 0x2e0a,
	0x2e0c, 0x2e0d,
	0x2e1c, 0x2e1d,
	0x2e20, 0x2e21,
	0x2e22, 0x2e23,
	0x2e24, 0x2e25,
	0x2e26, 0x2e27,
	0x2e28, 0x2e29,
	0x3008, 0x3009, /* chinese paired punctuation */
	0x300a, 0x300b,
	0x300c, 0x300d,
	0x300e, 0x300f,
	0x3010, 0x3011,
	0x3014, 0x3015,
	0x3016, 0x3017,
	0x3018, 0x3019,
	0x301a, 0x301b,
	0xfe59, 0xfe5a,
	0xfe5b, 0xfe5c,
	0xfe5d, 0xfe5e,
	0xff08, 0xff09,
	0xff3b, 0xff3d,
	0xff5b, 0xff5d,
	0xff5f, 0xff60,
	0xff62, 0xff63,
}

func get_pair_index(ch rune) int {
	lower := 0
	upper := len(paired_chars) - 1

	for lower <= upper {
		mid := (lower + upper) / 2

		if ch < paired_chars[mid] {
			upper = mid - 1
		} else if ch > paired_chars[mid] {
			lower = mid + 1
		} else {
			return mid
		}
	}

	return -1
}

// next advances to the next range. If `iter`
// is already at the end, it is left unchanged and `false`
// is returned.
func (iter *scriptIter) next() bool {
	if iter.scriptEnd >= iter.textEnd {
		return false
	}

	start_sp := iter.parenSp
	iter.scriptCode = language.Common
	iter.scriptStart = iter.scriptEnd

	for ; iter.scriptEnd < iter.textEnd; iter.scriptEnd++ {
		ch := iter.text[iter.scriptEnd]

		var pair_index int

		sc := language.LookupScript(ch)
		if sc != language.Common {
			pair_index = -1
		} else {
			pair_index = get_pair_index(ch)
		}

		/*
		* Paired character handling:
		*
		* if it's an open character, push it onto the stack.
		* if it's a close character, find the matching open on the
		* stack, and use that script code. Any non-matching open
		* characters above it on the stack will be poped.
		 */
		if pair_index >= 0 {
			if pair_index&1 == 0 { // is open ?
				/*
				* If the paren stack is full, empty it. This
				* means that deeply nested paired punctuation
				* characters will be ignored, but that's an unusual
				* case, and it's better to ignore them than to
				* write off the end of the stack...
				 */
				iter.parenSp++
				if iter.parenSp >= parenStackDepth {
					iter.parenSp = 0
				}

				iter.parenStack[iter.parenSp].pairIndex = pair_index
				iter.parenStack[iter.parenSp].scriptCode = iter.scriptCode
			} else if iter.parenSp >= 0 {
				pi := pair_index & ^1

				for iter.parenSp >= 0 && iter.parenStack[iter.parenSp].pairIndex != pi {
					iter.parenSp--
				}

				if iter.parenSp < start_sp {
					start_sp = iter.parenSp
				}

				if iter.parenSp >= 0 {
					sc = iter.parenStack[iter.parenSp].scriptCode
				}
			}
		}

		if iter.scriptCode.IsSameScript(sc) {
			if !iter.scriptCode.IsRealScript() && sc.IsRealScript() {
				iter.scriptCode = sc

				/*
				* now that we have a final script code, fix any open
				* characters we pushed before we knew the script code.
				 */
				for start_sp < iter.parenSp {
					start_sp++
					iter.parenStack[start_sp].scriptCode = iter.scriptCode
				}
			}

			/*
			* if this character is a close paired character,
			* pop it from the stack
			 */
			if pair_index >= 0 && pair_index&1 != 0 && iter.parenSp >= 0 {
				iter.parenSp--

				if iter.parenSp < start_sp {
					start_sp = iter.parenSp
				}
			}
		} else {
			/* Different script, we're done */
			break
		}
	}

	return true
}
