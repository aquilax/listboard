package main

import (
	"math"
	"net/url"
	"strconv"
)

type Page struct {
	Num int
	URL string
}

type Pages []Page

type PaginationConfig struct {
	ipp   int
	page  int
	total int
	url   string
	param string
}

func Pagination(pc PaginationConfig) Pages {
	pCount := int(math.Ceil(float64(pc.total) / float64(pc.ipp)))
	if pc.total <= pc.ipp {
		return make(Pages, 0)
	}
	pages := make(Pages, pCount)
	// Normalize first page
	if pc.page == 0 {
		pc.page = 1
	}
	pUrl, _ := url.Parse(pc.url)
	val := pUrl.Query()
	// Number of pages

	for i := 1; i <= pCount; i++ {
		// Don't set the url for the current page
		tURL := ""
		if i != pc.page {
			val.Set(pc.param, strconv.Itoa(i))
			pUrl.RawQuery = val.Encode()
			tURL = pUrl.String()
		}
		pages[i-1] = Page{i, tURL}
	}
	return pages
}
