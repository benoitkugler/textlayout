// Generate the correspondance between ISO 15924 scripts tag
// and Unicode table name as used by the unicode package.
package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"sort"
	"strings"
)

func parse() map[uint32]string {
	b, err := ioutil.ReadFile("iso15924.txt")
	if err != nil {
		log.Fatal(err)
	}
	m := map[uint32]string{}
	for _, line := range bytes.Split(b, []byte{'\n'}) {
		chunks := strings.Split(string(line), ";")
		if bytes.HasPrefix(line, []byte{'#'}) || len(chunks) < 6 {
			continue // comment or empty line
		}
		code := chunks[0]
		if len(code) != 4 {
			log.Fatal("invalid code ", code)
		}

		if code == "Geok" {
			continue // special case: duplicate tag
		}

		pva := chunks[4]
		if pva == "" {
			continue
		}
		tag := binary.BigEndian.Uint32([]byte(strings.ToLower(code)))
		m[tag] = pva
	}
	return m
}

func main() {
	m := parse()

	out := "../scripts_table.go"
	f, err := os.Create(out)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Fprintln(f, `package language
	
	// Code generated by gen/scripts.go DO NOT EDIT.

	const (
		`)

	var sortedKeys []uint32
	for k := range m {
		sortedKeys = append(sortedKeys, k)
	}
	sort.Slice(sortedKeys, func(i, j int) bool { return sortedKeys[i] < sortedKeys[j] })

	for _, k := range sortedKeys {
		v := m[k]
		fmt.Fprintf(f, "%s = Script(0x%08x)\n", v, k)
	}
	fmt.Fprintln(f, ")")

	fmt.Fprintln(f, "var scriptToTag = map[string]Script{")
	for _, k := range sortedKeys {
		v := m[k]
		fmt.Fprintf(f, "%q : %s,\n", v, v)
	}
	fmt.Fprintln(f, "}")

	if err := f.Close(); err != nil {
		log.Fatal(err)
	}
	exec.Command("goimports", "-w", out).Run()
}
