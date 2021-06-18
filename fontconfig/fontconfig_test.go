package fontconfig

import (
	"fmt"
	"math"
	"sync"
	"testing"
)

func TestWeightFromOT(t *testing.T) {
	if w := int(WeightFromOT(float32(math.MaxInt32))); w != WEIGHT_EXTRABLACK {
		t.Errorf("expected ExtraBlack, got %d", w)
	}
}

func TestDefaultFontDirs(t *testing.T) {
	dirs, err := DefaultFontDirs()
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(dirs)
}

func TestConcurrentDefaultLang(t *testing.T) {
	var w sync.WaitGroup
	w.Add(2)
	go func() {
		fmt.Println(getDefaultLangs())
		w.Done()
	}()
	go func() {
		fmt.Println(getDefaultLangs())
		w.Done()
	}()
	w.Wait()
}
