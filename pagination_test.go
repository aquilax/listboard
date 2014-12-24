package main

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestPagination(t *testing.T) {
	Convey("Given Pagination", t, func() {
		Convey("Returns empty list if total < ipp", func() {
			pl := Pagination(&PagConfig{
				ipp: 10,
				page: 0,
				total: 3,
			});
			So(len(*pl), ShouldEqual, 0)
		})
		Convey("Returns correct number of pages", func() {
			pl := Pagination(&PagConfig{
				ipp: 10,
				page: 0,
				total: 13,
				url: "http://example.com",
				param: "page",
			});
			So(len(*pl), ShouldEqual, 2)
			So((*pl)[0].Num, ShouldEqual, 1)
			So((*pl)[0].URL, ShouldEqual, "")
			So((*pl)[1].Num, ShouldEqual, 2)
			So((*pl)[1].URL, ShouldEqual, "http://example.com?page=2")
		})
	})
}
