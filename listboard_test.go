package main

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestListboard(t *testing.T) {
	Convey("Given new Listboard", t, func() {
		lb := NewListboard()
		Convey("Listboards is not nil", func() {
			So(lb, ShouldNotBeNil)
		})
	})
}
