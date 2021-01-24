package fontconfig

import (
	"bytes"
	"fmt"
	"testing"
)

// ported from fontconfig/test/test-bz1744377.c: 2000 Keith Packard
func TestOldSyntax(t *testing.T) {
	doc := []byte(`
	<fontconfig>
  		<include ignore_missing="yes">blahblahblah</include>
	</fontconfig>
	`)

	cfg := NewConfig()
	err := cfg.LoadFromMemory(bytes.NewReader(doc))
	if err == nil {
		t.Error("expected error on old syntax")
	}
}

func TestParseConfs(t *testing.T) {
	const dir = "confs"
	cfg := NewConfig()
	err := cfg.LoadFromDir(dir)
	if err != nil {
		t.Errorf("failed to load directory %s:%s", dir, err)
	}
	fmt.Println(len(cfg.acceptPatterns), len(cfg.rejectPatterns), len(cfg.subst))
}
