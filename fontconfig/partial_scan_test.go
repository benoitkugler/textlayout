package fontconfig

import (
	"fmt"
	"testing"
	"time"
)

func TestPartialScan(t *testing.T) {
	ti := time.Now()
	out, err := PartialScanFontDirectories(testFontDir)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("%d fonts partially scanned in %s (average: %s)\n", len(out), time.Since(ti), time.Since(ti)/time.Duration(len(out)))
}

func BenchmarkPartialScan(b *testing.B) {
	for i := 0; i < b.N; i++ {
		PartialScanFontDirectories(testFontDir)
	}
}
