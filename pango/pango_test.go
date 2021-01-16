package pango

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"testing"
)

func assertFalse(t *testing.T, b bool, message string) {
	if b {
		t.Error(message + ": expected false, got true")
	}
}
func assertTrue(t *testing.T, b bool, message string) {
	if !b {
		t.Error(message + ": expected true, got false")
	}
}

func assertEquals(t *testing.T, a, b interface{}) {
	if a != b {
		t.Errorf("expected same values, got %v and %v", a, b)
	}
}

// return an error if there is a difference or if the content of the
// file could not be read
func diffWithFile(text string, filename string) error {
	b, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}

	if !bytes.Equal([]byte(text), b) {
		return fmt.Errorf("expected\n%s\n got \n%s", b, text)
	}
	return nil
}
