package main

import (
	"testing"
	"time"
)

func TestNewSpamGuard(t *testing.T) {
	t.Run("spamguard blocks too frequent posts", func(t *testing.T) {
		var canPost bool
		sg := NewSpamGuard("1s")
		if canPost = sg.CanPost("test"); !canPost {
			t.Errorf("Expected to be allowed to make first post")
		}
		if canPost = sg.CanPost("test"); canPost {
			t.Errorf("Expected to be disallowed to make first post")
		}
		d, _ := time.ParseDuration("1h")
		sg.clean(time.Now().Add(d))
		if canPost = sg.CanPost("test"); !canPost {
			t.Errorf("Expected to be allowed to make third post after time has passed")
		}
	})
}
