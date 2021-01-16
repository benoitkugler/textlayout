package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"sort"
	"strconv"
	"strings"

	"github.com/benoitkugler/textlayout/pango"
)

func main() {
	content, err := ioutil.ReadFile("rgb.txt")
	if err != nil {
		log.Fatal(err)
	}

	m := make(map[string]pango.AttrColor)
	for _, lineB := range bytes.Split(content, []byte{'\n'}) {
		line := strings.TrimSpace(string(lineB))
		if line == "" {
			continue
		}

		fields := strings.Fields(string(line))
		if len(fields) != 4 {
			log.Fatal("invalid line", string(line))
		}
		r, err := strconv.Atoi(fields[0])
		if err != nil {
			log.Fatal(err)
		}
		g, err := strconv.Atoi(fields[1])
		if err != nil {
			log.Fatal(err)
		}
		b, err := strconv.Atoi(fields[2])
		if err != nil {
			log.Fatal(err)
		}

		k := strings.ToLower(fields[3])
		m[k] = pango.AttrColor{Red: uint16(r * 65535 / 255), Green: uint16(g * 65535 / 255), Blue: uint16(b * 65535 / 255)}
	}

	s := "package pango\n\n// Code generated by color/gen.go - DO NOT EDIT\n\nvar colorEntries = map[string]AttrColor{\n"

	var keys []string
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		s += fmt.Sprintf("%q: {Red: %#v, Green: %#v, Blue: %#v},\n", k, m[k].Red, m[k].Green, m[k].Blue)
	}
	s += "}"

	outFile := "../color_table.go"
	err = ioutil.WriteFile(outFile, []byte(s), os.ModePerm)
	if err != nil {
		log.Fatal(err)
	}
	err = exec.Command("goimports", "-w", outFile).Run()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Done.")
}
