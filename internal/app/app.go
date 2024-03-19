package app

import (
	"context"
	"errors"
	"fmt"
	"github.com/igurianova/logistic_optimizer/internal/config"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// Run Метод запускающий приложение
func Run() error {
	cfg := config.Read()

	mux := http.NewServeMux()
	routeConfig := config.RouteConfig{}

	for basePath, handler := range routeConfig.Routes() {
		mux.Handle(basePath, handler)
	}

	srv := &http.Server{
		Addr:           "localhost:" + cfg.HTTPAddr,
		Handler:        mux,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	stopped := make(chan struct{})
	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
		<-sigint
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := srv.Shutdown(ctx); err != nil {
			log.Printf("HTTP Server Shutdown Error: %v", err)
		}
		close(stopped)
	}()

	fmt.Println("Logistic Optimizer App started!")
	log.Printf("Starting HTTP server on %s", cfg.HTTPAddr)

	if err := srv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("HTTP server ListenAndServe Error: %v", err)
	}

	<-stopped
	fmt.Println("Logistic Optimizer App stopped :(")
	return nil
}
