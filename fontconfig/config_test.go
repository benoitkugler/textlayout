package fontconfig

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"testing"

	"github.com/benoitkugler/textlayout/fonts"
)

// ported from fontconfig/test/test-conf.c Copyright © 2000 Keith Packard,  2018 Akira TAGOH

const testFontDir = "/usr/share/fonts"

// const testFontDir = "/System/Library/fonts"

func init() {
	// when developping, comment this line to speed up several test
	// setupCacheFile()
}

func setupCacheFile() {
	fmt.Println("Setting cache file up...")

	c := NewConfig()
	if err := c.LoadFromDir("confs"); err != nil {
		log.Fatal(err)
	}

	fs, err := c.ScanFontDirectories(testFontDir)
	if err != nil {
		log.Fatal("setting up cache file for tests", err)
	}

	f, err := os.Create("test/cache.fc")
	if err != nil {
		log.Fatal("setting up cache file for tests", err)
	}
	defer f.Close()

	err = fs.Serialize(f)
	if err != nil {
		log.Fatal("setting up cache file for tests", err)
	}

	fmt.Println("Test cache file written.")
}

func cachedFS() Fontset {
	f, err := os.Open("test/cache.fc")
	if err != nil {
		log.Fatal("opening cache file for tests", err)
	}
	defer f.Close()

	out, err := LoadFontset(f)
	if err != nil {
		log.Fatal("opening cache file for tests", err)
	}
	return out
}

func ExampleConfig() {
	c := NewConfig()
	if err := c.LoadFromDir("confs"); err != nil {
		log.Fatal(err)
	}
	fontDirs, err := DefaultFontDirs()
	if err != nil {
		log.Fatal(err)
	}
	_, err = c.ScanFontDirectories(fontDirs...)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("success")
	// Output: success
}

func TestGetFonts(t *testing.T) {
	fs := cachedFS()
	fmt.Println("fonts from cache:", len(fs))
	for _, p := range fs {
		if _, ok := p.GetString(FILE); !ok {
			t.Error("file not present")
		}
		if _, ok := p.GetInt(INDEX); !ok {
			t.Error("index not present")
		}
		if p.Format() == "" {
			t.Error("missing format")
		}

		cs, _ := p.GetCharset(CHARSET)
		if cs.Len() == 0 {
			t.Error("empty charset")
		}

		faceID := p.FaceID()
		if faceID == (fonts.FaceID{}) {
			t.Error("empty face")
		}
	}
}

func buildPattern(config *Config, jsonObject map[string]interface{}) (Pattern, error) {
	pat := NewPattern()

	for key, val := range jsonObject {
		o := config.getRegisterObjectType(key)
		var v Value
		switch val := val.(type) {
		case bool:
			v = False
			if val {
				v = True
			}
		case float32:
			panic("float32")
		case float64:
			v = Float(val)
		case int:
			v = Int(val)
		case string:
			switch o.typeInfo.(type) {
			case typeRange, typeFloat, typeInt:
				c := nameGetConstant(val)
				if c == nil {
					return nil, fmt.Errorf("value %s for key %s is not a known constant", val, key)
				}
				if c.object.String() != key {
					return nil, fmt.Errorf("value %s is a constant of different object (expected %s, got %s)",
						val, c.object, key)
				}
				v = Int(c.value)
			default:
				if val == "DontCare" {
					v = DontCare
				} else {
					v = String(val)
				}
			}
		case nil:
			continue
		default:
			return nil, fmt.Errorf("unexpected object to build a pattern: (%s %T)", key, val)
		}
		pat.Add(o.object, v, true)
	}
	return pat, nil
}

func buildFs(config *Config, fonts []map[string]interface{}) (Fontset, error) {
	var fs Fontset
	for _, m := range fonts {
		out, err := buildPattern(config, m)
		if err != nil {
			return nil, err
		}
		fs = append(fs, out)
	}
	return fs, nil
}

func runTest(data Fontset, config *Config, tests testData) error {
	for i, obj := range tests.Tests {
		method := obj.Method
		query, err := buildPattern(config, obj.Query)
		if err != nil {
			return err
		}
		result, err := buildPattern(config, obj.Result)
		if err != nil {
			return err
		}
		resultFs, err := buildFs(config, obj.ResultFs)
		if err != nil {
			return err
		}

		if method == "match" {
			config.Substitute(query, nil, MatchQuery)
			query.SubstituteDefault()
			match := data.Match(query, config)
			if match == nil {
				return errors.New("no match")
			}
			for obj, vals := range result {
				for x, vr := range *vals {
					vm, _ := match.GetAt(obj, x)
					if !valueEqual(vm, vr.Value) {
						return fmt.Errorf("test %d: expected %v, got %v", i, vr.Value, vm)
					}
				}
			}
		} else if method == "list" {
			fs := data.List(query)
			if len(fs) != len(resultFs) {
				return fmt.Errorf("unexpected number of results: expected %d, got %d", len(resultFs), len(fs))
			}
			for i, expFont := range resultFs {
				gotFont := fs[i]
				for obj, vals := range expFont {
					for x, vr := range *vals {
						vm, _ := gotFont.GetAt(obj, x)
						if !valueEqual(vm, vr.Value) {
							return fmt.Errorf("expected %v, got %v", vr.Value, vm)
						}
					}
				}
			}
		} else {
			return fmt.Errorf("unknown testing method: %s", method)
		}
	}
	return nil
}

type oneTest struct {
	Method   string                   `json:"method,omitempty"`
	Query    map[string]interface{}   `json:"query,omitempty"`
	Result   map[string]interface{}   `json:"result,omitempty"`
	ResultFs []map[string]interface{} `json:"result_fs,omitempty"`
}

type testData struct {
	Fonts []map[string]interface{} `json:"fonts,omitempty"`
	Tests []oneTest                `json:"tests,omitempty"`
}

func runScenario(config *Config, file string) error {
	b, err := ioutil.ReadFile(file)
	if err != nil {
		return fmt.Errorf("unable to read test file: %s", err)
	}
	var tmp testData
	err = json.Unmarshal(b, &tmp)
	if err != nil {
		return fmt.Errorf("unable to read test file: %s", err)
	}

	fs, err := buildFs(config, tmp.Fonts)
	if err != nil {
		return fmt.Errorf("invalid fonts: %s", err)
	}
	err = runTest(fs, config, tmp)
	return err
}

func TestConfigScenario(t *testing.T) {
	for _, name := range []string{
		"45-generic",
		"60-generic",
		"90-synthetic",
	} {
		configFile := "confs/" + name + ".conf"
		testScenario := "test/test-" + name + ".json"

		config := NewConfig()

		f, err := os.Open(configFile)
		if err != nil {
			t.Fatalf("failed to load config: %s", err)
		}
		err = config.LoadFromMemory(f)
		if err != nil {
			t.Fatalf("failed to load config: %s", err)
		}
		f.Close()

		err = runScenario(config, testScenario)
		if err != nil {
			t.Fatalf("test %s: %s", name, err)
		}
	}
}

func TestMatch(t *testing.T) {
	fs := cachedFS()
	pattern := BuildPattern(PatternElement{Object: LANG, Value: NewLangset("fr")})
	fmt.Println(len(fs.List(pattern, FILE, INDEX, LANG)))

	noto := BuildPattern(PatternElement{Object: FAMILY, Value: String("Noto Color Emoji")})
	fmt.Println(len(fs.List(noto, FAMILY, PIXEL_SIZE)))
}

func TestFontTypes(t *testing.T) {
	fs := cachedFS()
	counts := map[FontFormat]int{}
	for _, font := range fs {
		counts[font.Format()]++
	}
	fmt.Println(counts)
}

func TestCustomConfig(t *testing.T) {
	fontFilename := "dummy"
	fontFamily := "arial"
	fontconfigStyle := "roman"
	fontconfigWeight := "regular"
	fontconfigStretch := "normal"
	featuresSttring := ""
	xml := fmt.Sprintf(`<?xml version="1.0"?>
			<!DOCTYPE fontconfig SYSTEM "fonts.dtd">
			<fontconfig>
			  <match target="scan">
				<test name="file" compare="eq">
				  <string>%s</string>
				</test>
				<edit name="family" mode="assign_replace">
				  <string>%s</string>
				</edit>
				<edit name="slant" mode="assign_replace">
				  <const>%s</const>
				</edit>
				<edit name="weight" mode="assign_replace">
				  <const>%s</const>
				</edit>
				<edit name="width" mode="assign_replace">
				  <const>%s</const>
				</edit>
			  </match>
			  <match target="font">
				<test name="file" compare="eq">
				  <string>%s</string>
				</test>
				<edit name="fontfeatures" mode="assign_replace">%s</edit>
			  </match>
			</fontconfig>`, fontFilename, fontFamily, fontconfigStyle,
		fontconfigWeight, fontconfigStretch, fontFilename, featuresSttring)

	config := Standard.Copy()
	err := config.LoadFromMemory(bytes.NewReader([]byte(xml)))
	if err != nil {
		t.Fatalf("Failed to load fontconfig config: %s", err)
	}
}
