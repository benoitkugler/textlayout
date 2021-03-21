package unicodedata

import "testing"

func TestUnicodeNormalization(t *testing.T) {
	assertCompose := func(a, b rune, okExp bool, abExp rune) {
		ab, ok := Compose(a, b)
		if ok != okExp || ab != abExp {
			t.Errorf("expected %d, %v got %d, %v", abExp, okExp, ab, ok)
		}
	}

	/* Not composable */
	assertCompose(0x0041, 0x0042, false, 0)
	assertCompose(0x0041, 0, false, 0)
	assertCompose(0x0066, 0x0069, false, 0)

	/* Singletons should not compose */
	assertCompose(0x212B, 0, false, 0)
	assertCompose(0x00C5, 0, false, 0)
	assertCompose(0x2126, 0, false, 0)
	assertCompose(0x03A9, 0, false, 0)

	/* Non-starter pairs should not compose */
	assertCompose(0x0308, 0x0301, false, 0) /* !0x0344 */
	assertCompose(0x0F71, 0x0F72, false, 0) /* !0x0F73 */

	/* Pairs */
	assertCompose(0x0041, 0x030A, true, 0x00C5)
	assertCompose(0x006F, 0x0302, true, 0x00F4)
	assertCompose(0x1E63, 0x0307, true, 0x1E69)
	assertCompose(0x0073, 0x0323, true, 0x1E63)
	assertCompose(0x0064, 0x0307, true, 0x1E0B)
	assertCompose(0x0064, 0x0323, true, 0x1E0D)

	/* Hangul */
	assertCompose(0xD4CC, 0x11B6, true, 0xD4DB)
	assertCompose(0x1111, 0x1171, true, 0xD4CC)
	assertCompose(0xCE20, 0x11B8, true, 0xCE31)
	assertCompose(0x110E, 0x1173, true, 0xCE20)

	assertCompose(0xAC00, 0x11A7, false, 0)
	assertCompose(0xAC00, 0x11A8, true, 0xAC01)
	assertCompose(0xAC01, 0x11A8, false, 0)

	assertDecompose := func(ab rune, expOk bool, expA, expB rune) {
		a, b, ok := Decompose(ab)
		if ok != expOk || a != expA || b != expB {
			t.Errorf("decompose: expected 0x%x, 0x%x, %v got 0x%x, 0x%x, %v", expA, expB, expOk, a, b, ok)
		}
	}

	/* Not decomposable */
	assertDecompose(0x0041, false, 0x0041, 0)
	assertDecompose(0xFB01, false, 0xFB01, 0)
	assertDecompose(0x1F1EF, false, 0x1F1EF, 0)

	/* Singletons */
	assertDecompose(0x212B, true, 0x00C5, 0)
	assertDecompose(0x2126, true, 0x03A9, 0)

	/* Non-starter pairs decompose, but not compose */
	assertDecompose(0x0344, true, 0x0308, 0x0301)
	assertDecompose(0x0F73, true, 0x0F71, 0x0F72)

	/* Pairs */
	assertDecompose(0x00C5, true, 0x0041, 0x030A)
	assertDecompose(0x00F4, true, 0x006F, 0x0302)
	assertDecompose(0x1E69, true, 0x1E63, 0x0307)
	assertDecompose(0x1E63, true, 0x0073, 0x0323)
	assertDecompose(0x1E0B, true, 0x0064, 0x0307)
	assertDecompose(0x1E0D, true, 0x0064, 0x0323)

	/* Hangul */
	assertDecompose(0xD4DB, true, 0xD4CC, 0x11B6)
	assertDecompose(0xD4CC, true, 0x1111, 0x1171)
	assertDecompose(0xCE31, true, 0xCE20, 0x11B8)
	assertDecompose(0xCE20, true, 0x110E, 0x1173)
}
