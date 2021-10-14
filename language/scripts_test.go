package language

import (
	"os"
	"testing"
	"unicode"
)

func lookupScriptOld(r rune) Script {
	for name, table := range unicode.Scripts {
		if unicode.Is(table, r) {
			return scriptToTag[name]
		}
	}
	return Unknown
}

func loadSample(t testing.TB) []rune {
	b, err := os.ReadFile("tests/Wikipedia-Go.html")
	if err != nil {
		t.Fatal(err)
	}
	return []rune(string(b))
}

func TestFastLookup(t *testing.T) {
	sample := loadSample(t)
	for _, r := range sample {
		g1, g2 := lookupScriptOld(r), LookupScript(r)
		if g1 != g2 {
			t.Fatalf("for rune 0x%x, expected %s, got %s", r, g1, g2)
		}
	}
}

func BenchmarkLookupScript(b *testing.B) {
	sample := loadSample(b)
	b.Run("Map with unicode tables", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			for _, r := range sample {
				_ = lookupScriptOld(r)
			}
		}
	})
	b.Run("Custom script table", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			for _, r := range sample {
				_ = LookupScript(r)
			}
		}
	})
}
