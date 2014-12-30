package main

import (
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func TestSpamGuard(t *testing.T) {
	Convey("Given SpamGuard", t, func() {
		Convey("Blocking works", func() {
			sg := NewSpamGuard("1s")
			So(sg.CanPost("test"), ShouldBeTrue)
			So(sg.CanPost("test"), ShouldBeFalse)
			d, _ := time.ParseDuration("1h")
			sg.clean(time.Now().Add(d))
			So(sg.CanPost("test"), ShouldBeTrue)
		})
	})
}
