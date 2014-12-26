package main

import (
	"time"

	"github.com/aquilax/tripcode"
	"github.com/gosimple/slug"
	"github.com/microcosm-cc/bluemonday"
	"github.com/russross/blackfriday"
)

func hfTime(t time.Time) string {
	return t.Format("01.02.2006 15:04")
}

func hfSlug(s string) string {
	return slug.Make(s) + ".html"
}

func hfMod(n int, mod int) int {
	return n % mod
}

func getTripcode(s string) string {
	return tripcode.Tripcode(s)
}

func getVote(t string) int {
	if t == "y" {
		return 1
	}
	if t == "n" {
		return -1
	}
	return 0
}

func inHoneypot(t string) bool {
	if len(t) > 0 {
		return true
	}
	return false
}

func renderText(t string) string {
	extensions := 0
	extensions |= blackfriday.EXTENSION_HARD_LINE_BREAK
	extensions |= blackfriday.HTML_USE_XHTML
	extensions |= blackfriday.HTML_USE_SMARTYPANTS
	extensions |= blackfriday.HTML_SMARTYPANTS_FRACTIONS
	extensions |= blackfriday.HTML_SMARTYPANTS_LATEX_DASHES
	htmlFlags := 0 |
		blackfriday.HTML_USE_XHTML |
		blackfriday.HTML_USE_SMARTYPANTS |
		blackfriday.HTML_SMARTYPANTS_FRACTIONS

	renderer := blackfriday.HtmlRenderer(htmlFlags, "", "")
	unsafe := blackfriday.Markdown([]byte(t), renderer, extensions)
	return string(bluemonday.UGCPolicy().SanitizeBytes(unsafe))
}
