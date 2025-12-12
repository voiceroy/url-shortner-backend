package main

import (
	"context"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"go-shorten/config"
	"go-shorten/internal/middleware"
	"go-shorten/internal/repository"
	"go-shorten/internal/router"
	"go-shorten/internal/store"

	"github.com/gin-gonic/gin"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Llongfile)

	gin.SetMode(config.GinMode)

	store.ConnectToDB()

	middleware.StartRateLimiterCleanup()
	repository.StartOldURLsCleanup()
	store.StartCacheCleanup()

	srv := &http.Server{
		Addr:         ":8080",
		Handler:      router.SetupRouter(),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	done := make(chan bool)
	go func(done chan bool) {
		ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
		defer stop()

		<-ctx.Done()
		log.Println("Shutting down server...")
		stop()

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := srv.Shutdown(ctx); err != nil {
			log.Fatal("Server forced to shutdown:", err)
		}

		log.Println("Server exiting")
		done <- true
	}(done)

	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("listen: %s\n", err)
	}

	<-done
}
