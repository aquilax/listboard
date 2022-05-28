package main

import (
	"crypto/md5"
	"encoding/hex"
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

func getTripCode(s string) string {
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
	return len(t) > 0
}

func renderText(t string) string {
	extensions := 0
	extensions |= blackfriday.HTML_USE_XHTML
	extensions |= blackfriday.HTML_USE_SMARTYPANTS
	extensions |= blackfriday.HTML_SMARTYPANTS_FRACTIONS
	extensions |= blackfriday.HTML_SMARTYPANTS_LATEX_DASHES
	extensions |= blackfriday.EXTENSION_AUTOLINK
	extensions |= blackfriday.EXTENSION_HARD_LINE_BREAK
	extensions |= blackfriday.EXTENSION_NO_INTRA_EMPHASIS
	extensions |= blackfriday.EXTENSION_TABLES
	extensions |= blackfriday.EXTENSION_FENCED_CODE
	extensions |= blackfriday.EXTENSION_STRIKETHROUGH
	extensions |= blackfriday.EXTENSION_SPACE_HEADERS
	extensions |= blackfriday.EXTENSION_HEADER_IDS

	htmlFlags := 0 |
		blackfriday.HTML_USE_XHTML |
		blackfriday.HTML_USE_SMARTYPANTS |
		blackfriday.HTML_SMARTYPANTS_FRACTIONS

	renderer := blackfriday.HtmlRenderer(htmlFlags, "", "")
	unsafe := blackfriday.Markdown([]byte(t), renderer, extensions)
	return string(bluemonday.UGCPolicy().SanitizeBytes(unsafe))
}

func hfGravatar(tripcode string) string {
	if tripcode == "" {
		return "http://www.gravatar.com/avatar/00000000000000000000000000000000?d=retro"
	}
	hash := md5.Sum([]byte(tripcode))
	return "http://www.gravatar.com/avatar/" + hex.EncodeToString(hash[:]) + "?d=retro"
}
