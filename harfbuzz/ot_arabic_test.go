package harfbuzz

import "testing"

func TestNumArabicLookup(t *testing.T) {
	if len(arabicFallbackFeatures) > arabicFallbackMaxLookups {
		t.Error()
	}

	// static_assert (sizeof (arabicWin1256GsubLookups.manifestData) ==  ARABIC_FALLBACK_MAX_LOOKUPS * sizeof (ManifestLookup), "");
}
