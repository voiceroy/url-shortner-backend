package middleware

import (
	"context"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

const (
	BurstLimit               = 4
	NormalLimit              = 1
	RateLimiterResetInterval = 5 * time.Minute
)

var limiters = sync.Map{}

type Client struct {
	Limiter  *rate.Limiter
	LastSeen time.Time
}

func StartRateLimiterCleanup(wg *sync.WaitGroup, ctx context.Context) {
	wg.Go(func() {
		ticker := time.NewTicker(RateLimiterResetInterval)

		for {
			select {
			case <-ticker.C:
				{
					limiters.Range(func(key, value any) bool {
						client := key.(*Client)
						if time.Since(client.LastSeen) > 5*time.Minute {
							limiters.Delete(key)
						}

						return true
					})
				}
			case <-ctx.Done():
				{
					log.Println("Stopping Rate Limiter")
					return
				}
			}
		}
	})
}

func RateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		clientIP := c.ClientIP()

		newClient := &Client{
			rate.NewLimiter(NormalLimit, BurstLimit),
			time.Now(),
		}

		v, loaded := limiters.LoadOrStore(clientIP, newClient)
		client := v.(*Client)
		if loaded {
			client.LastSeen = time.Now()
		}

		if client.Limiter.Allow() {
			c.Next()
		} else {
			c.AbortWithStatus(http.StatusTooManyRequests)
		}
	}
}
