package main

import (
	"sync"
	"time"
)

type SpamGuard struct {
	duration string
	posts    map[string]time.Time
	mutex    *sync.Mutex
}

func NewSpamGuard(duration string) *SpamGuard {
	return &SpamGuard{
		duration: duration,
		posts:    make(map[string]time.Time),
		mutex:    &sync.Mutex{},
	}
}

func (sg *SpamGuard) CanPost(id string) bool {
	result := true
	now := time.Now()
	sg.mutex.Lock()
	expires, found := sg.posts[id]
	if found {
		if expires.After(now) {
			// Blocked
			result = false
		}
	} else {
		// Add to block
		d, err := time.ParseDuration(sg.duration)
		if err != nil {
			panic(err)
		}
		sg.posts[id] = now.Add(d)
	}
	sg.clean(now)
	sg.mutex.Unlock()
	return result
}

func (sg *SpamGuard) clean(now time.Time) {
	for key, expires := range sg.posts {
		if expires.Before(now) {
			delete(sg.posts, key)
		}
	}
}
