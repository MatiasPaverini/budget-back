package main

import (
	"context"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/mpaverini/budget-back/internal/account"
	"github.com/mpaverini/budget-back/internal/config"
	"github.com/mpaverini/budget-back/internal/db"
	"github.com/mpaverini/budget-back/internal/httpapi"
	"github.com/mpaverini/budget-back/internal/indicator"
	"github.com/mpaverini/budget-back/internal/platform/devauth"
	firebaseauth "github.com/mpaverini/budget-back/internal/platform/firebase"
	"github.com/mpaverini/budget-back/internal/platform/postgres"
	"github.com/mpaverini/budget-back/internal/recurringcharge"
	"github.com/mpaverini/budget-back/internal/transaction"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	pool, err := postgres.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("postgres: %v", err)
	}
	defer pool.Close()

	var authMiddleware func(http.Handler) http.Handler
	switch cfg.AuthMode {
	case config.AuthModeDev:
		authMiddleware = devauth.NewMiddleware(devauth.DefaultUserID).Wrap
	default:
		verifier, err := firebaseauth.NewVerifier(ctx, cfg.FirebaseCredentialsFile)
		if err != nil {
			log.Fatalf("firebase: %v", err)
		}
		authMiddleware = verifier.Middleware
	}

	queries := db.New(pool)

	router := httpapi.NewRouter(httpapi.Dependencies{
		Auth:             authMiddleware,
		Accounts:         account.NewService(queries),
		Transactions:     transaction.NewService(queries),
		Indicators:       indicator.NewService(queries),
		RecurringCharges: recurringcharge.NewService(queries),
		IPCSeriesID:      cfg.IPCSeriesID,
	})

	server := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Printf("shutdown: %v", err)
		}
	}()

	log.Printf("listening on :%s", cfg.Port)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server: %v", err)
	}
}
