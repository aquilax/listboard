package main

import (
	"github.com/gosimple/slug"
	"time"
)

func hfLang(t string) string {
	return t
}

func hfTime(t time.Time) string {
	return t.Format("01.02.2006 15.04")
}

func hfSlug(s string) string {
	return slug.Make(s) + ".html"
}

func hfAnchor(link, label string) string {
	return label
}

func hfAnchorTr(link, label string) string {
	return label
}
