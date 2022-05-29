package main

import (
	"crypto/md5"
	"encoding/hex"
	"log"
	"strconv"
	"time"

	"github.com/aquilax/tripcode"
	"github.com/gosimple/slug"
	"github.com/microcosm-cc/bluemonday"
	"github.com/russross/blackfriday/v2"
)

func hfTime(t time.Time) string {
	return t.Format("01.02.2006 15:04")
}

func getLanguageSlug(lang string) func(s string) string {
	return func(s string) string {
		return slug.MakeLang(s, lang)
	}
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
	unsafe := blackfriday.Run([]byte(t))
	return string(bluemonday.UGCPolicy().SanitizeBytes(unsafe))
}

func hfGravatar(tripCode string) string {
	if tripCode == "" {
		return "http://www.gravatar.com/avatar/00000000000000000000000000000000?d=retro"
	}
	hash := md5.Sum([]byte(tripCode))
	return "http://www.gravatar.com/avatar/" + hex.EncodeToString(hash[:]) + "?d=retro"
}

func getPageNumber(pageStr string) int {
	page := 1
	var err error
	if len(pageStr) != 0 {
		page, err = strconv.Atoi(pageStr)
		if err != nil {
			log.Printf("%s is not a valid page number", pageStr)
			page = 1
		}
	}
	return page - 1
}
