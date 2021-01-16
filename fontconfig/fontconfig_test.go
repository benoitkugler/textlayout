package fontconfig

import (
	"math"
	"testing"
)

func TestWeightFromOT(t *testing.T) {
	if w := int(FcWeightFromOpenTypeDouble(float64(math.MaxInt32))); w != FC_WEIGHT_EXTRABLACK {
		t.Errorf("expected ExtraBlack, got %d", w)
	}
}
