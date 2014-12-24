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

type Pages []*Page

type PagConfig struct {
	ipp   int
	page  int
	total int
	url   string
	param string
}

func Pagination(pc *PagConfig) *Pages {
	if pc.total <= pc.ipp {
		return nil
	}
	var pages Pages
	// Normalize first page
	if pc.page == 0 {
		pc.page = 1
	}
	pUrl, _ := url.Parse(pc.url)
	val := pUrl.Query()
	// Number of pages
	pCount := int(math.Ceil(float64(pc.total) / float64(pc.ipp)))
	for i := 1; i <= pCount; i++ {
		// Don't set the url for the current page
		tURL := ""
		if i != pc.page {
			val.Set(pc.param, strconv.Itoa(i))
			pUrl.RawQuery = val.Encode()
			tURL = pUrl.String()
		}
		pages = append(pages, &Page{i, tURL})
	}
	return &pages
}
