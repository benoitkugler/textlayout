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
	version          = "13.0.0"
	urlXML           = "https://unicode.org/Public/" + version + "/ucdxml/ucd.nounihan.grouped.zip"
	urlUnicode       = "https://unicode.org/Public/" + version + "/ucd/UnicodeData.txt"
	urlEmoji         = "https://unicode.org/Public/" + version + "/ucd/emoji/emoji-data.txt"
	urlEmojiTest     = "https://unicode.org/Public/emoji/13.1/emoji-test.txt"
	urlMirroring     = "https://unicode.org/Public/" + version + "/ucd/BidiMirroring.txt"
	urlArabic        = "https://unicode.org/Public/" + version + "/ucd/ArabicShaping.txt"
	urlScripts       = "https://unicode.org/Public/" + version + "/ucd/Scripts.txt"
	urlIndic1        = "https://unicode.org/Public/" + version + "/ucd/IndicSyllabicCategory.txt"
	urlIndic2        = "https://unicode.org/Public/" + version + "/ucd/IndicPositionalCategory.txt"
	urlBlocks        = "https://unicode.org/Public/" + version + "/ucd/Blocks.txt"
	urlLineBreak     = "https://unicode.org/Public/" + version + "/ucd/LineBreak.txt"
	urlSentenceBreak = "https://unicode.org/Public/" + version + "/ucd/auxiliary/SentenceBreakProperty.txt"
	urlDerivedCore   = "https://unicode.org/Public/" + version + "/ucd/DerivedCoreProperties.txt"
)

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
