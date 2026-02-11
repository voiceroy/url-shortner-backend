package store

import (
	"context"
	"log"
	"sync"
	"time"
)

var cache = sync.Map{}

func StartCacheCleanup(wg *sync.WaitGroup, ctx context.Context) {
	wg.Go(func() {
		ticker := time.NewTicker(5 * time.Minute)

		for {
			select {
			case <-ticker.C:
				{
					cache.Clear()
				}
			case <-ctx.Done():
				{
					log.Println("Stopping Cache Cleanup")
					return
				}
			}
		}
	})
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
