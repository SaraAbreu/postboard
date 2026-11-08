package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/SaraAbreu/postboard/api/internal/handlers"
	"github.com/SaraAbreu/postboard/api/internal/store"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	if err := run(logger); err != nil {
		logger.Error("la API se detuvo con error", "err", err)
		os.Exit(1)
	}
}

func run(logger *slog.Logger) error {
	port := getenv("PORT", "8080")
	dbURL := os.Getenv("DATABASE_URL")

	// Si no hay DATABASE_URL, usamos memoria (útil para probar rápido sin Docker).
	var postStore store.PostStore
	if dbURL == "" {
		logger.Warn("DATABASE_URL vacío: usando almacenamiento en memoria")
		postStore = store.NewMemoryStore()
	} else {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		pool, err := pgxpool.New(ctx, dbURL)
		if err != nil {
			return err
		}
		defer pool.Close()
		if err := pool.Ping(ctx); err != nil {
			return err
		}
		logger.Info("conectado a PostgreSQL")
		postStore = store.NewPostgresStore(pool)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, _ *http.Request) {
		w.Write([]byte("ok"))
	})
	handlers.NewPostHandler(postStore, logger).Routes(mux)

	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      withCORS(withLogging(logger, mux)),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown: al recibir Ctrl+C o SIGTERM, terminamos las peticiones en curso.
	errCh := make(chan error, 1)
	go func() {
		logger.Info("API escuchando", "port", port)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-errCh:
		return err
	case <-stop:
		logger.Info("apagando servidor...")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return srv.Shutdown(ctx)
	}
}

func withLogging(logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		logger.Info("request", "method", r.Method, "path", r.URL.Path, "duration", time.Since(start).String())
	})
}

func withCORS(next http.Handler) http.Handler {
	origin := getenv("CORS_ORIGIN", "http://localhost:5173") // puerto por defecto de Vite
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
