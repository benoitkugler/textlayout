package truetype

import (
	"fmt"
	"os"
	"sort"
	"testing"
)

func TestBinarySearch(t *testing.T) {
	filename := "testdata/Raleway-v4020-Regular.otf"
	file, err := os.Open(filename)
	if err != nil {
		t.Fatalf("Failed to open %q: %s\n", filename, err)
	}

	font, err := Parse(file)
	if err != nil {
		t.Fatalf("Parse(%q) err = %q, want nil", filename, err)
	}

	pos, err := font.GposTable()
	if err != nil {
		t.Fatal(err)
	}
	sub, err := font.GsubTable()
	if err != nil {
		t.Fatal(err)
	}

	for _, table := range []*TableLayout{pos, sub} {
		var tags []int
		for _, s := range table.Scripts {
			tags = append(tags, int(s.Tag))
		}
		if !sort.IntsAreSorted(tags) {
			t.Fatalf("tag not sorted: %v", tags)
		}
		for i, s := range table.Scripts {
			ptr := table.FindScript(s.Tag)
			if ptr != i {
				t.Errorf("binary search failed for script tag %s", s.Tag)
			}
		}

		s := table.FindScript(Tag(0)) // invalid
		if s != -1 {
			t.Errorf("binary search should have failed")
		}

		// now check the languages

		for _, script := range table.Scripts {
			var tags []int
			for _, s := range script.Languages {
				tags = append(tags, int(s.Tag))
			}
			if !sort.IntsAreSorted(tags) {
				t.Fatalf("tag not sorted: %v", tags)
			}
			for i, l := range script.Languages {
				ptr := script.FindLanguage(l.Tag)
				if ptr != i {
					t.Errorf("binary search failed for language tag %s", l.Tag)
				}
			}

			s := script.FindLanguage(Tag(0)) // invalid
			if s != -1 {
				t.Errorf("binary search should have failed")
			}
		}
	}
}

func TestFeatureVariations(t *testing.T) {
	filename := "testdata/Commissioner-VF.ttf"
	file, err := os.Open(filename)
	if err != nil {
		t.Fatalf("Failed to open %q: %s\n", filename, err)
	}

	font, err := Parse(file)
	if err != nil {
		t.Fatalf("Parse(%q) err = %q, want nil", filename, err)
	}

	gsub, err := font.GsubTable()
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(gsub.FeatureVariations)

	gdef, err := font.GDefTable()
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(gdef.Class)
}
