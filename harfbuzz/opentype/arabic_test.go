package opentype

import "testing"

func TestNumArabicLookup(t *testing.T) {
	if len(arabicFallbackFeatures) > ARABIC_FALLBACK_MAX_LOOKUPS {
		t.Error()
	}

	// static_assert (sizeof (arabicWin1256GsubLookups.manifestData) ==  ARABIC_FALLBACK_MAX_LOOKUPS * sizeof (ManifestLookup), "");
}
