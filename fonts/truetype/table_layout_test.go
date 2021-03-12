package truetype

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
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

// func TestFindSub(t *testing.T) {
// 	dirs := []string{
// 		"/home/benoit/Téléchargements/harfbuzz/test/shaping/data/aots/fonts",
// 		"/home/benoit/Téléchargements/harfbuzz/test/shaping/data/in-house/fonts",
// 		"/home/benoit/Téléchargements/harfbuzz/test/shaping/data/text-rendering-tests/fonts",
// 		"/home/benoit/Téléchargements/harfbuzz/test/shaping/data/text-rendering-tests/fonts",
// 		"/home/benoit/go/src/github.com/benoitkugler/textlayout/fonts/truetype/testdata",
// 		"/usr/share/fonts/opentype",
// 		"/usr/share/fonts/truetype",
// 	}

// 	var readDir func(dir string)
// 	readDir = func(dir string) {
// 		files, err := ioutil.ReadDir(dir)
// 		if err != nil {
// 			t.Fatal(err)
// 		}

// 		for _, fi := range files {
// 			path := filepath.Join(dir, fi.Name())
// 			if fi.IsDir() {
// 				readDir(path)
// 				continue
// 			}

// 			file, err := os.Open(path)
// 			if err != nil {
// 				t.Fatalf("Failed to open %q: %s\n", fi.Name(), err)
// 			}

// 			fonts, err := Loader.Load(file)
// 			if err != nil {
// 				fmt.Println("\t\tskipping", fi.Name(), ":", err)
// 				continue
// 			}
// 			for _, font := range fonts {
// 				font := font.(*Font)

// 				if s := font.tables[tagSbix]; s != nil {
// 					fmt.Println("found sbix:", path)
// 				}

// 			}
// 		}
// 	}

// 	for _, dir := range dirs {
// 		readDir(dir)
// 	}
// }

func extractCFF(table []byte, name string) error {
	const dir = "/home/benoit/go/src/github.com/benoitkugler/textlayout/fonts/type1C/test/ttf"
	return ioutil.WriteFile(filepath.Join(dir, name), table, os.ModePerm)
}

func dirFiles(t *testing.T, dir string) []string {
	files, err := ioutil.ReadDir(dir)
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
	filename := "testdata/Commissioner-VF.ttf"
	file, err := os.Open(filename)
	if err != nil {
		t.Fatalf("Failed to open %q: %s\n", filename, err)
	}

	font, err := Parse(file)
	if err != nil {
		t.Fatalf("Parse(%q) err = %q, want nil", filename, err)
	}

	gsub, err := font.GSUBTable()
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(gsub.FeatureVariations)

	gdef, err := font.GDEFTable()
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(gdef.Class)
}
