package language

import (
	"fmt"
	"reflect"
	"testing"
)

func TestLanguage(t *testing.T) {
	fmt.Println(DefaultLanguage())
}

func TestSimpleInheritance(t *testing.T) {
	l := NewLanguage("en_US_someVariant")
	if sh := l.SimpleInheritance(); !reflect.DeepEqual(sh, []Language{l, "en-us", "en"}) {
		t.Fatalf("unexpected inheritance %v", sh)
	}

	l = NewLanguage("fr")
	if sh := l.SimpleInheritance(); !reflect.DeepEqual(sh, []Language{l}) {
		t.Fatalf("unexpected inheritance %v", sh)
	}
}
