package main

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

func NewLanguage(lang string) *Language {
	return &Language{
		found: false,
		tr:    make(Translations),
	}
}

func (tp *TransPool) Get(lang string) *Language {
	l, ok := tp.languages[lang]
	if !ok {
		tp.languages[lang] = NewLanguage(lang)
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
