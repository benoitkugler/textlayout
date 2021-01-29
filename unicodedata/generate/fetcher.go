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
	urlIndic     = "https://unicode.org/Public/UCD/latest/ucd/IndicSyllabicCategory.txt"
	urlEmoji     = "https://www.unicode.org/Public/emoji/12.0/emoji-data.txt"
	urlLineBreak = "https://www.unicode.org/Public/12.0.0/ucd/LineBreak.txt"
)

// TODO: à compléter

func fetchEmojis() {
	fmt.Println("Downloading", urlEmoji, "...")
	fetchData(urlEmoji)
}

func fetchData(url string) {
	resp, err := http.Get(url)
	check(err)

	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	check(err)

	filename := path.Base(url)
	err = ioutil.WriteFile(filename, data, os.ModePerm)
	check(err)
}
