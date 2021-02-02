package common

import (
	"sort"
	"testing"
)

func TestAATFeaturesSorted(t *testing.T) {
	var tags []int
	for _, f := range feature_mappings {
		tags = append(tags, int(f.otFeatureTag))
	}
	if !sort.IntsAreSorted(tags) {
		t.Fatal("expected sorted tags, got %v", tags)
	}
}
