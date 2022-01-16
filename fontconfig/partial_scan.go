package fontconfig

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/benoitkugler/textlayout/fonts"
	"github.com/benoitkugler/textlayout/fonts/bitmap"
	"github.com/benoitkugler/textlayout/fonts/truetype"
	"github.com/benoitkugler/textlayout/fonts/type1"
)

func scanFontFile(file fonts.Resource) ([]fonts.FontDescriptor, FontFormat) {
	out, err := truetype.ScanFont(file)
	if err == nil {
		return out, TrueType
	}
	out, err = type1.ScanFont(file)
	if err == nil {
		return out, Type1
	}
	out, err = bitmap.ScanFont(file)
	if err == nil {
		return out, PCF
	}
	return nil, ""
}

// reject several extensions which are for sure not supported font files
func validFontFile(name string) bool {
	// ignore hidden file
	if name == "" || name[0] == '.' {
		return false
	}
	if strings.HasSuffix(name, ".enc.gz") || // encodings
		strings.HasSuffix(name, ".afm") || // metrics (ascii)
		strings.HasSuffix(name, ".pfm") || // metrics (binary)
		strings.HasSuffix(name, ".dir") || // summary
		strings.HasSuffix(name, ".scale") ||
		strings.HasSuffix(name, ".alias") {
		return false
	}
	return true
}

// descriptions are appended to `dst`, which is returned
func partialReadDir(dir string, seen map[string]bool, dst []fonts.FaceDescription) ([]fonts.FaceDescription, error) {
	walkFn := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("invalid font location: %s", err)
		}

		if info.IsDir() { // keep going
			if seen[path] {
				return filepath.SkipDir
			}
			seen[path] = true
			return nil
		}

		if info.Mode()&os.ModeSymlink != 0 {
			path, err = filepath.EvalSymlinks(path)
			if err != nil {
				return err
			}
		}
		if !validFontFile(info.Name()) {
			return nil
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}

		fds, _ := scanFontFile(file)

		file.Close()

		for _, fd := range fds {
			dst = append(dst, fonts.FaceDescription{Family: fd.Family()})
		}
		return nil
	}

	err := filepath.Walk(dir, walkFn)

	return dst, err
}

// PartialScanFontDirectories walk through the given directories
// and scan each font file to extract a summary.
// An error is returned if the directory traversal fails, not for invalid font files,
// which are simply ignored.
func PartialScanFontDirectories(dirs ...string) ([]fonts.FaceDescription, error) {
	seen := make(map[string]bool) // keep track of visited dirs to avoid double includes
	var (
		out []fonts.FaceDescription
		err error
	)
	for _, dir := range dirs {
		out, err = partialReadDir(dir, seen, out)
		if err != nil {
			return nil, err
		}
	}
	return out, nil
}
