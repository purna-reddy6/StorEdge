package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"go.uber.org/zap"

	"github.com/storedge/storedge/services/financing-enwrs/internal/handler"
	"github.com/storedge/storedge/services/financing-enwrs/internal/repository"
	"github.com/storedge/storedge/services/financing-enwrs/internal/service"
)

func main() {
	_ = godotenv.Load()

	logger, _ := zap.NewProduction()
	if os.Getenv("APP_ENV") == "development" {
		logger, _ = zap.NewDevelopment()
	}
	defer logger.Sync()

	dbURL := getenv("DATABASE_URL", "postgres://storedge:storedge_dev@localhost:5432/storedge?sslmode=disable")
	nerlURL := getenv("NERL_API_URL", "")
	nerlKey := getenv("NERL_API_KEY", "")
	originationFee := 0.015 // 1.5% from blueprint
	port := getenv("PORT", "8082")

	db, err := repository.NewPostgres(dbURL)
	if err != nil {
		logger.Fatal("postgres connect failed", zap.Error(err))
	}
	defer db.Close()

	enwrsRepo := repository.NewENWRsRepository(db)
	loanRepo := repository.NewLoanRepository(db)

	nerlClient := service.NewRepositoryClient(nerlURL, nerlKey, logger)
	enwrsSvc := service.NewENWRsService(enwrsRepo, loanRepo, nerlClient, originationFee, logger)

	router := handler.NewRouter(enwrsSvc, logger)
	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	go func() {
		logger.Info("financing-enwrs service started", zap.String("port", port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("server failed", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	srv.Shutdown(ctx)
	logger.Info("financing-enwrs service stopped")
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
