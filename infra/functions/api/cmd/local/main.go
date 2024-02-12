package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/haandol/vertical-slice-go-lambda-example/api/internal/constant"
	"github.com/haandol/vertical-slice-go-lambda-example/api/internal/feature/clickstream/handler"
	"github.com/haandol/vertical-slice-go-lambda-example/api/pkg/middleware"
	"github.com/haandol/vertical-slice-go-lambda-example/api/pkg/o11y"
	"github.com/haandol/vertical-slice-go-lambda-example/api/pkg/util/slogger"
	"github.com/joho/godotenv"
)

const (
	isProd bool = false
)

var (
	r *gin.Engine
)

func init() {
	// setup logger
	logger := slogger.Init(isProd)
	logger.Info("initializing...")

	// load .env
	if err := godotenv.Load(); err != nil {
		logger.Error("Error loading .env file")
	}

	// setup o11y
	o11y.InitXray()

	// setup router
	r = gin.Default()
	r.Use(middleware.GinXrayMiddleware("Local"))
	r.Use(middleware.GinSlogWithConfig(logger, &middleware.Config{
		UTC: false,
	}))
	r.Use(cors.New(cors.Config{
		AllowOrigins: []string{"http://localhost:3000"},
		AllowMethods: []string{"POST, GET, PUT, DELETE, OPTIONS"},
		AllowHeaders: []string{"Content-Type", "Authorization", "X-Amzn-Trace-Id", "X-Requested-With"},
		MaxAge:       12 * time.Hour,
	}))
	r.Use(middleware.RecoveryWithSlog(logger, true))

	rg := r.Group("/v1/clickstream")
	rg.POST("/:path", handler.CreateClickEventController)
	rg.GET("/:path", handler.GetClickStreamController)
}

func main() {
	logger := slogger.New()

	httpErr := make(chan error)
	srv := &http.Server{
		Addr:    ":8090",
		Handler: r,
	}
	go func() {
		logger.Info("Starting Server...")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			httpErr <- fmt.Errorf("listenAndServe: %w", err)
		}
	}()

	quit := make(chan os.Signal, 2)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT, os.Interrupt)
	select {
	case err := <-httpErr:
		logger.Error("error occurs", "err", err)
	case <-quit:
		logger.Info("Quit signal received.")
	}

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		defer cancel()
		logger.Info("Closing server...")
		if err := srv.Shutdown(ctx); err != nil {
			logger.Error("error on shuting-down server", "err", err)
		}
	}()
	select {
	case <-quit:
		logger.Info("Received second interrupt signal; quitting without waiting for graceful close")
		os.Exit(1)
	case <-ctx.Done():
		logger.Info("Graceful close complete")
		os.Exit(0)
	case <-time.After(constant.GracefulShutdownTimeout):
		logger.Error("closed by timeout")
		os.Exit(1)
	}
}
