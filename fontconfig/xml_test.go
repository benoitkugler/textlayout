package fontconfig

import (
	"io/ioutil"
	"path/filepath"
	"strings"
	"testing"
)

// ported from fontconfig/test/test-bz1744377.c: 2000 Keith Packard
func TestParse(t *testing.T) {
	doc := []byte(`
	<fontconfig>
  		<include ignore_missing="yes">blahblahblah</include>
	</fontconfig>
	`)
	doc2 := []byte(`
	<fontconfig>
  		<include ignore_missing="no">blahblahblah</include>
	</fontconfig>
	`)
	cfg := NewFcConfig()

	if err := cfg.ParseAndLoadFromMemory(doc); err != nil {
		t.Errorf("expected no error since 'ignore_missing' is true, got %s", err)
	}
	if err := cfg.ParseAndLoadFromMemory(doc2); err == nil {
		t.Error("expected error on invalid include")
	}
}

func TestParseConfs(t *testing.T) {
	const dir = "test/confs"
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	for _, file := range files {
		if !strings.HasSuffix(file.Name(), ".conf") {
			continue
		}
		b, err := ioutil.ReadFile(filepath.Join(dir, file.Name()))
		if err != nil {
			t.Fatal(err)
		}

		cfg := NewFcConfig()

		if err := cfg.parseAndLoadFromMemory(file.Name(), b, true); err != nil {
			t.Errorf("file %s: %s", file.Name(), err)
		}
	}
}
