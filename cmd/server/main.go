package main

import (
	"context"
	// "database/sql"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"shortener/internal/db"
	"shortener/internal/shorten"
)

// type response struct {
// 	Message string `json:"message"`
// 	Status string `json:"status,omitempty"`
// 	StatusCode int `json:"status_code,omitempty"`
// }

// func HandleHome(w http.ResponseWriter, r *http.Request) {
// 	var res = response{
// 		Message: "hehehehehelloooooo",
// 	}

// 	resJson, err := json.Marshal(res)
// 	if err != nil {
// 		log.Printf("marshaling error: %v", err)
// 	}
// 	log.Println("resJson", string(resJson))

// 	w.Header().Set("Content-Type", "application/json")
// 	_, err = w.Write(resJson)
// 	if err != nil {
// 		log.Printf("error writing body (HandleHome): %v", err)
// 	}
// }

func Start(ctx context.Context) error {
	// 1. Create infra / dependencies
	// dsn := "postgress://tester:password@localhost:5432/testdb?sslmode=disable"
	
	// db, err := sql.Open("postgres", dsn)
	// if err != nil {
	// 	log.Fatalf("db error: %v", err)
	// }

	cfg := db.Config{
		Host: "localhost",
		Port: 5432,
		User: "tester",
		Password: "password",
		DBName: "testdb",
		SSLMode: "disable",
	}

	sqlDB, err := db.Connect(cfg)
	if err != nil {
		log.Fatalf("db error: %v", err)
	}

	// store := shorten.NewMemStore()
	store := shorten.NewPGStore(sqlDB)
	generator := shorten.NewBase62Generator()
	shortener := shorten.NewShortener(store, generator)

	// 2. Create mux
	mux := http.NewServeMux()

	// 3. Register routes
	shorten.RegisterRoutes(mux, shortener)

	// 4. Create and start server
	server := http.Server{
		Addr: ":8080",
		Handler: mux,
	}

	go func() {
		log.Println("server starting on :8080")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen error: %v", err)
		}
	}()

	// Block until root context is cancelled
	<-ctx.Done()

	log.Println("shutdown signal received")

	// Graceful shutdown
	shutDownCtx, cancel := context.WithTimeout(ctx, 5 * time.Second)
	defer cancel()
	return server.Shutdown(shutDownCtx)
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := Start(ctx); err != nil {
		log.Fatalf("server stopped with error: %v", err)
	}
}