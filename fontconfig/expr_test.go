package fontconfig

import "testing"

func TestTypeEquality(t *testing.T) {
	// we use interface implemented by empty struct to
	// define "types"
	var (
		a typeMeta = typeString{}
		b typeMeta = typeString{}
		c typeMeta = typeFloat{}
	)
	if a != b {
		t.Error("expected equal")
	}
	if a == c {
		t.Error("expected different")
	}
}
