package truetype

import (
	"testing"
)

func TestFloat1616(t *testing.T) {
	for _, u := range []uint32{
		Float1616ToUint(-12345.456),
		2*(1<<16) | 24563,
		24563,
		1 << 1,
		1 << 17,
		1 << 18,
	} {
		f := Float1616FromUint(u)
		if got := Float1616ToUint(f); u != got {
			t.Fatalf("invalid conversion for %b: %d %d", u, u, got)
		}
	}
}
