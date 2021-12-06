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

func TestLanguage_IsDerivedFrom(t *testing.T) {
	type args struct {
		root Language
	}
	tests := []struct {
		name string
		l    Language
		args args
		want bool
	}{
		{
			name: "",
			l:    "fr-FR",
			args: args{"fr"},
			want: true,
		},
		{
			name: "",
			l:    "ca",
			args: args{"cat"},
			want: false,
		},
		{
			name: "",
			l:    "ca",
			args: args{"ca"},
			want: true,
		},
	}
	for _, tt := range tests {
		if got := tt.l.IsDerivedFrom(tt.args.root); got != tt.want {
			t.Errorf("Language.IsDerivedFrom() = %v, want %v", got, tt.want)
		}
	}
}
