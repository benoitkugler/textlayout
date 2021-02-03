package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path"
)

// download the database files and save them locally

const (
	urlXML       = "https://unicode.org/Public/12.0.0/ucdxml/ucd.nounihan.grouped.zip"
	urlUnicode   = "https://unicode.org/Public/12.0.0/ucd/UnicodeData.txt"
	urlEmoji     = "https://www.unicode.org/Public/emoji/12.0/emoji-data.txt"
	urlLineBreak = "https://www.unicode.org/Public/12.0.0/ucd/LineBreak.txt"
	urlMirroring = "https://www.unicode.org/Public/12.0.0/ucd/BidiMirroring.txt"
	urlArabic    = "https://unicode.org/Public/12.0.0/ucd/ArabicShaping.txt"
	// urlBlocks    = "https://unicode.org/Public/12.0.0/ucd/Blocks.txt"
	// urlIndic     = "https://unicode.org/Public/UCD/latest/ucd/IndicSyllabicCategory.txt"
)

// TODO: à compléter

func fetchData(url string) {
	fmt.Println("Downloading", url, "...")
	resp, err := http.Get(url)
	check(err)

	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	check(err)

	filename := path.Base(url)
	err = ioutil.WriteFile(filename, data, os.ModePerm)
	check(err)
}
