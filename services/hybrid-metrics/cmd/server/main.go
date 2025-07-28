package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time" 

	"hybrid-metrics/internal/api"
	"hybrid-metrics/internal/collector"
	"hybrid-metrics/internal/config"
	"hybrid-metrics/internal/fusion"
)

func main() {
	log.Println("🚀 Starting Hybrid Metrics Server...")

	// Load configuration
	cfg := config.Load()
	log.Printf("📊 Configuration loaded: GraphQL URL: %s, Port: %s", cfg.GraphQLURL, cfg.Port)

	// Initialize components
	complexityCollector := collector.NewComplexityCollector(cfg.GraphQLURL)
	systemCollector := collector.NewSystemCollector()
	fusionAlgorithm := fusion.NewFusionAlgorithm(cfg.FusionConfig)

	// Start collectors
	log.Println("📈 Starting metric collectors...")
	
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start complexity collection
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()
		
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				complexity, err := complexityCollector.Collect()
				if err != nil {
					log.Printf("⚠️ Error collecting complexity: %v", err)
					continue
				}
				
				log.Printf("🔍 Collected complexity: %.2f", complexity)
				fusionAlgorithm.AddComplexityData(complexity, time.Now())
			}
		}
	}()

	// Start system metrics collection
	go func() {
		ticker := time.NewTicker(3 * time.Second)
		defer ticker.Stop()
		
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				systemMetrics, err := systemCollector.Collect()
				if err != nil {
					log.Printf("⚠️ Error collecting system metrics: %v", err)
					continue
				}
				
				log.Printf("💻 System metrics - CPU: %.2f%%, Memory: %.2f%%", 
					systemMetrics.CPUUsage, systemMetrics.MemoryUsage)
				fusionAlgorithm.AddSystemData(systemMetrics, time.Now())
			}
		}
	}()

	// Start fusion algorithm processing
	go func() {
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()
		
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				decision := fusionAlgorithm.ProcessMetrics()
				if decision.ShouldScale {
					log.Printf("🎯 SCALING DECISION: %s (Factor: %.2f, Reason: %s)", 
						decision.Action, decision.ScalingFactor, decision.Reason)
				}
			}
		}
	}()

	// Start HTTP API server
	apiServer := api.NewServer(fusionAlgorithm)
	httpServer := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: apiServer.Router(),
	}

	// Start server in goroutine
	go func() {
		log.Printf("🌐 HTTP API server starting on port %s", cfg.Port)
		log.Printf("📊 Metrics available at: http://localhost:%s/metrics", cfg.Port)
		log.Printf("🔍 Fusion stats at: http://localhost:%s/fusion-stats", cfg.Port)
		log.Printf("📈 Health check at: http://localhost:%s/health", cfg.Port)
		
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("❌ Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("🛑 Shutting down Hybrid Metrics Server...")

	// Graceful shutdown
	cancel()
	
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		log.Printf("❌ Server shutdown error: %v", err)
	}

	log.Println("✅ Hybrid Metrics Server stopped gracefully")
}