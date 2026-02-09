package main

import (
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/go-chi/chi/v5"

	"logger/internal/httpapi"
	"logger/internal/sink"
)

func main() {
	addr := normalizeAddr(os.Getenv("PORT"))
	logDir := os.Getenv("LOG_DIR")
	if logDir == "" {
		logDir = "./logs"
	}

	fileSink, err := sink.NewFileSink(logDir)
	if err != nil {
		log.Fatalf("failed to initialise file sink: %v", err)
	}
	defer func() {
		if err := fileSink.Close(); err != nil {
			log.Printf("error closing file sink: %v", err)
		}
	}()

	r := chi.NewRouter()
	handler := httpapi.NewLoggerHandler(fileSink)

	r.Post("/logs", handler.PostLog)

	log.Printf("logging service listening on %s, writing to %s", addr, logDir)
	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

func normalizeAddr(port string) string {
	port = strings.TrimSpace(port)
	if port == "" {
		return ":8080"
	}
	if strings.HasPrefix(port, ":") {
		return port
	}
	return ":" + port
}

