package truetype

import "testing"

func TestParsedTag(t *testing.T) {
	tag := MustNewTag("head")
	if tag != 0x68656164 {
		t.Errorf("head != %v", tag)
	}
}

func TestNewTag(t *testing.T) {
	tag := Tag(0x74727565)

	if tag != 0x74727565 {
		t.Errorf("true != %v", tag)
	}
}

func TestTagEquality(t *testing.T) {
	t1 := Tag(0x74727565)
	t2 := Tag(0x74727565)

	if t1 != t2 {
		t.Errorf("equality failed for number")
	}

	if MustNewTag("true") != t1 {
		t.Errorf("equality failed %v %v", MustNewTag("true"), t1)
	}
}
