package store

import (
	"sync"
	"time"
)

var cache = sync.Map{}

func StartCacheCleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	go func() {
		for range ticker.C {
			cache.Clear()
		}
	}()
}

func AddToCache(key, value string) {
	cache.Store(key, value)
}

func GetFromCache(key string) (string, bool) {
	if value, found := cache.Load(key); !found {
		return "", false
	} else {
		return value.(string), true
	}
}
