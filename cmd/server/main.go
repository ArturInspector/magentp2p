package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/dechat/exchange-service/internal/adapters"
	"github.com/dechat/exchange-service/internal/api"
	"github.com/dechat/exchange-service/internal/config"
	"github.com/dechat/exchange-service/internal/models"
	"github.com/dechat/exchange-service/internal/services"
	"github.com/dechat/exchange-service/internal/storage"
	"github.com/gorilla/mux"
)

func main() {
	// Load config
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Connect to database
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.DBName,
		cfg.Database.SSLMode,
	)

	db, err := storage.New(dsn)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Initialize blockchain adapters
	chainAdapters := make(map[models.Chain]adapters.BlockchainAdapter)
	for chainName, chainCfg := range cfg.Chains {
		chain := models.Chain(chainName)
		adapter, err := adapters.NewEVMAdapter(chainCfg.RPCURL, chainCfg.ChainID)
		if err != nil {
			log.Printf("Warning: failed to initialize adapter for %s: %v", chainName, err)
			continue
		}
		chainAdapters[chain] = adapter
		log.Printf("Initialized adapter for chain: %s", chainName)
	}

	// Initialize services
	walletService := services.NewWalletService(db, chainAdapters)
	withdrawalService := services.NewWithdrawalService(db, chainAdapters)
	depositMonitor := services.NewDepositMonitor(db, chainAdapters)

	// Start deposit monitor
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := depositMonitor.Start(ctx); err != nil {
		log.Fatalf("Failed to start deposit monitor: %v", err)
	}

	// Start withdrawal processor (runs periodically)
	go func() {
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				for chain := range chainAdapters {
					if err := withdrawalService.ProcessPendingWithdrawals(ctx, chain); err != nil {
						log.Printf("Error processing withdrawals for %s: %v", chain, err)
					}
				}
			}
		}
	}()

	// Setup API routes
	handlers := api.NewHandlers(walletService, withdrawalService)
	router := mux.NewRouter()

	router.HandleFunc("/health", handlers.HealthCheck).Methods("GET")
	router.HandleFunc("/api/v1/deposit/address", handlers.GenerateDepositAddress).Methods("POST")
	router.HandleFunc("/api/v1/balance/{chain}", handlers.GetBalance).Methods("GET")
	router.HandleFunc("/api/v1/withdrawal", handlers.CreateWithdrawal).Methods("POST")

	// Start HTTP server
	addr := fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port)
	srv := &http.Server{
		Addr:    addr,
		Handler: router,
	}

	// Graceful shutdown
	go func() {
		log.Printf("Server starting on %s", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")
	cancel()

	ctxShutdown, cancelShutdown := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelShutdown()

	if err := srv.Shutdown(ctxShutdown); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}
