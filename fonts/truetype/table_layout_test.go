package truetype

import (
	"bytes"
	"fmt"
	"path/filepath"
	"sort"
	"testing"

	testdata "github.com/benoitkugler/textlayout-testdata/truetype"
)

func TestBinarySearch(t *testing.T) {
	filename := "Raleway-v4020-Regular.otf"
	file, err := testdata.Files.ReadFile(filename)
	if err != nil {
		t.Fatalf("Failed to open %q: %s\n", filename, err)
	}

	font, err := NewFontParser(bytes.NewReader(file))
	if err != nil {
		t.Fatalf("Parse(%q) err = %q, want nil", filename, err)
	}

	pos, err := font.GPOSTable()
	if err != nil {
		t.Fatal(err)
	}
	sub, err := font.GSUBTable()
	if err != nil {
		t.Fatal(err)
	}

	for _, table := range []TableLayout{pos.TableLayout, sub.TableLayout} {
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

func dirFiles(t *testing.T, dir string) []string {
	files, err := testdata.Files.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	var filenames []string
	for _, fi := range files {
		filenames = append(filenames, filepath.Join(dir, fi.Name()))
	}
	return filenames
}

func TestFeatureVariations(t *testing.T) {
	filename := "Commissioner-VF.ttf"
	file, err := testdata.Files.ReadFile(filename)
	if err != nil {
		t.Fatalf("Failed to open %q: %s\n", filename, err)
	}

	font, err := NewFontParser(bytes.NewReader(file))
	if err != nil {
		t.Fatalf("Parse(%q) err = %q, want nil", filename, err)
	}

	gsub, err := font.GSUBTable()
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(gsub.FeatureVariations)

	names, err := font.tryAndLoadNameTable()
	if err != nil {
		t.Fatal(err)
	}

	fvar, err := font.tryAndLoadFvarTable(names)
	if err != nil {
		t.Fatal(err)
	}

	gdef, err := font.GDEFTable(len(fvar.Axis))
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(gdef.Class)
}
