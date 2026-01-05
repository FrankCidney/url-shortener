package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type response struct {
	Message string `json:"message"`
	Status string `json:"status,omitempty"`
	StatusCode int `json:"status_code,omitempty"`
}

func HandleHome(w http.ResponseWriter, r *http.Request) {
	var res = response{
		Message: "hehehehehelloooooo",
	}

	resJson, err := json.Marshal(res)
	if err != nil {
		log.Printf("marshaling error: %v", err)
	}
	log.Println("resJson", string(resJson))

	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(resJson)
	if err != nil {
		log.Printf("error writing body (HandleHome): %v", err)
	}
}

func Start(ctx context.Context) error {
	mux := http.NewServeMux()

	// handle routes
	mux.HandleFunc("/", HandleHome)

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