package fontconfig

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"testing"
	"time"
)

// ported from fontconfig/test/test-conf.c Copyright © 2000 Keith Packard,  2018 Akira TAGOH

const testFontDir = "/usr/share/fonts"

func init() {
	// in order to speed up several test
	// we use a cache file: uncomment once
	// setupCacheFile()
}

func setupCacheFile() {
	var c Config
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
}

func cachedFS() Fontset {
	f, err := os.Open("test/cache.fc")
	if err != nil {
		log.Fatal("opening cache file for tests", err)
	}
	out, err := LoadFontset(f)
	if err != nil {
		log.Fatal("opening cache file for tests", err)
	}
	return out
}

func TestDefault(t *testing.T) {
	c := NewConfig()
	if err := c.LoadFromDir("confs"); err != nil {
		t.Fatal(err)
	}

	fontDirs, err := DefaultFontDirs()
	if err != nil {
		t.Fatal(err)
	}
	ti := time.Now()
	fs, err := c.ScanFontDirectories(fontDirs...)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(time.Since(ti))
	fmt.Println("patterns collected: ", len(fs))
}

func TestGetFonts(t *testing.T) {
	fs := cachedFS()
	for _, p := range fs {
		_, ok := p.GetString(FILE)
		if !ok {
			t.Error("file not present")
		}
	}
}

func buildPattern(jsonObject map[string]interface{}) (Pattern, error) {
	pat := NewPattern()

	for key, val := range jsonObject {
		o := getRegisterObjectType(key)
		var v Value
		switch val := val.(type) {
		case bool:
			v = FcFalse
			if val {
				v = FcTrue
			}
		case float64:
			v = Float(val)
		case int:
			v = Int(val)
		case string:
			switch o.typeInfo.(type) {
			case typeRange, typeFloat, typeInteger:
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
					v = FcDontCare
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

func buildFs(fonts []map[string]interface{}) (Fontset, error) {
	var fs Fontset
	for _, m := range fonts {
		out, err := buildPattern(m)
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
		query, err := buildPattern(obj.Query)
		if err != nil {
			return err
		}
		result, err := buildPattern(obj.Result)
		if err != nil {
			return err
		}
		resultFs, err := buildFs(obj.ResultFs)
		if err != nil {
			return err
		}

		if method == "match" {
			config.SubstituteWithPat(query, nil, MatchQuery)
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

	fs, err := buildFs(tmp.Fonts)
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

func BenchmarkLoad(b *testing.B) {
	var (
		c    Config
		seen = make(strSet)
	)
	for i := 0; i < b.N; i++ {
		_, err := c.readDir(testFontDir, seen)
		if err != nil {
			b.Fatal(err)
		}
	}
}
