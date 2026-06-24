package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"go.uber.org/zap"

	"github.com/storedge/storedge/services/search-match/internal/handler"
	"github.com/storedge/storedge/services/search-match/internal/repository"
	"github.com/storedge/storedge/services/search-match/internal/service"
)

func main() {
	_ = godotenv.Load()

	logger, _ := zap.NewProduction()
	if os.Getenv("APP_ENV") == "development" {
		logger, _ = zap.NewDevelopment()
	}
	defer logger.Sync()

	dbURL := mustEnv("DATABASE_URL", "postgres://storedge:storedge_dev@localhost:5432/storedge?sslmode=disable")
	redisURL := mustEnv("REDIS_URL", "redis://localhost:6379/0")
	jwtSecret := mustEnv("JWT_SECRET", "dev_secret_change_in_production")
	aiEngineURL := mustEnv("AI_ENGINE_URL", "http://localhost:8084")
	httpPort := mustEnv("HTTP_PORT", "8080")
	grpcPort := mustEnv("GRPC_PORT", "50051")

	db, err := repository.NewPostgres(dbURL)
	if err != nil {
		logger.Fatal("failed to connect to postgres", zap.Error(err))
	}
	defer db.Close()
	logger.Info("connected to PostgreSQL")

	rdb, err := repository.NewRedis(redisURL)
	if err != nil {
		logger.Fatal("failed to connect to redis", zap.Error(err))
	}
	defer rdb.Close()
	logger.Info("connected to Redis")

	warehouseRepo := repository.NewWarehouseRepository(db)
	bookingRepo := repository.NewBookingRepository(db)
	pricingCache := repository.NewPricingCache(rdb)

	matchingSvc := service.NewMatchingService(warehouseRepo, pricingCache, aiEngineURL, logger)
	bookingSvc := service.NewBookingService(bookingRepo, warehouseRepo, pricingCache, logger)
	authSvc := service.NewAuthService(db, jwtSecret, logger)

	// HTTP REST server
	httpRouter := handler.NewRouter(matchingSvc, bookingSvc, authSvc, jwtSecret, logger)
	httpServer := &http.Server{
		Addr:         ":" + httpPort,
		Handler:      httpRouter,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// gRPC server
	grpcServer := handler.NewGRPCServer(matchingSvc, bookingSvc, logger)
	grpcListener, err := net.Listen("tcp", ":"+grpcPort)
	if err != nil {
		logger.Fatal("failed to listen on gRPC port", zap.Error(err))
	}

	go func() {
		logger.Info("starting HTTP server", zap.String("port", httpPort))
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("HTTP server failed", zap.Error(err))
		}
	}()

	go func() {
		logger.Info("starting gRPC server", zap.String("port", grpcPort))
		if err := grpcServer.Serve(grpcListener); err != nil {
			logger.Fatal("gRPC server failed", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("shutting down servers...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(ctx); err != nil {
		logger.Error("HTTP server forced shutdown", zap.Error(err))
	}
	grpcServer.GracefulStop()
	logger.Info("servers stopped")
}

func mustEnv(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	if defaultVal == "" {
		panic(fmt.Sprintf("required env var %s is not set", key))
	}
	return defaultVal
}
