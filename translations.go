package main

import (
	"encoding/json"
	"log"
	"os"
)

type Translations map[string]string

type Language struct {
	found bool
	tr    Translations
}

type TransPool struct {
	basePath  string
	languages map[string]*Language
}

func NewTransPool(basePath string) *TransPool {
	return &TransPool{
		basePath:  basePath,
		languages: make(map[string]*Language),
	}
}

func NewLanguage(basePath, lang string) *Language {
	var t Translations
	found := true
	fileName := basePath + lang + ".json"
	file, err := os.Open(fileName)
	if err != nil {
		found = false
		log.Printf("Language %s not found", fileName)
	}
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&t); err != nil {
		log.Printf("Error loading language %s, %s", fileName, err)
		found = false
		t = make(Translations)
	}
	return &Language{
		found: found,
		tr:    t,
	}
}

func (tp *TransPool) Get(lang string) *Language {
	var l *Language
	var ok bool
	l, ok = tp.languages[lang]
	if !ok {
		l = NewLanguage(tp.basePath, lang)
		tp.languages[lang] = l
	}
	return l
}

func (l *Language) Lang(text string) string {
	if !l.found {
		// Language was not found, return the string
		return text
	}
	res, ok := l.tr[text]
	if !ok {
		// Key was not found
		return text
	}
	// Return translated string
	return res
}
