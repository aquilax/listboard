package main

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestListboard(t *testing.T) {
	Convey("Given new Listboard", t, func() {
		lb := NewListboard()
		Convey("Listboards is not nil", func() {
			So(lb, ShouldNotBeNil)
		})
	})
}
