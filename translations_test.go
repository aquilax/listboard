package main

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestTranslations(t *testing.T) {
	Convey("Given TranslationPool", t, func() {
		tp := NewTransPool("")
		Convey("Get gets new language", func() {
			ln := tp.Get("en")
			So(ln, ShouldNotBeNil)
			Convey("Translating works", func() {
				So(ln.Lang("test"), ShouldEqual, "test")
			})
		})
	})
}
