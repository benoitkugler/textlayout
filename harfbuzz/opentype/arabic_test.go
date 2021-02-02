package opentype

import "testing"

func TestNumArabicLookup(t *testing.T) {
	if len(arabicFallbackFeatures) > ARABIC_FALLBACK_MAX_LOOKUPS {
		t.Error()
	}

	// static_assert (sizeof (arabic_win1256_gsub_lookups.manifestData) ==  ARABIC_FALLBACK_MAX_LOOKUPS * sizeof (ManifestLookup), "");

}
