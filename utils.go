package main

import (
	"github.com/aquilax/tripcode"
	"github.com/gosimple/slug"
	"github.com/microcosm-cc/bluemonday"
	"github.com/russross/blackfriday"
	"time"
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
	unsafe := blackfriday.MarkdownCommon([]byte(t))
	return string(bluemonday.UGCPolicy().SanitizeBytes(unsafe))
}
