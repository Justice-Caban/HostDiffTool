package main

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"strings"

	"github.com/improbable-eng/grpc-web/go/grpcweb"
	"google.golang.org/grpc"

	"github.com/justicecaban/host-diff-tool/backend/internal/data"
	"github.com/justicecaban/host-diff-tool/backend/internal/server"
	"github.com/justicecaban/host-diff-tool/proto"
)

const dbPath = "./data/snapshots.db"

func main() {
	// Create data directory if it doesn't exist
	if err := os.MkdirAll("./data", 0755); err != nil {
		log.Fatalf("failed to create data directory: %v", err)
	}

	// Initialize database
	db, err := data.NewDB(dbPath)
	if err != nil {
		log.Fatalf("failed to initialize database: %v", err)
	}
	defer db.Close()

	// Create a new gRPC server
	grpcServer := grpc.NewServer()

	// Register the HostService
	hostServiceServer := server.NewServer(db)
	proto.RegisterHostServiceServer(grpcServer, hostServiceServer)

	// Start native gRPC server on port 9090 in a goroutine
	go func() {
		lis, err := net.Listen("tcp", ":9090")
		if err != nil {
			log.Fatalf("failed to listen on port 9090: %v", err)
		}
		log.Println("Starting native gRPC server on :9090")
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("failed to serve gRPC: %v", err)
		}
	}()

	// Create a gRPC-Web wrapper for browser clients
	wrappedGrpc := grpcweb.WrapServer(grpcServer)

	// Create a new HTTP server multiplexer
	mux := http.NewServeMux()

	// Register the gRPC-Web wrapper with logging middleware
	mux.Handle("/", loggingMiddleware(wrappedGrpc))

	// Create a new HTTP server for gRPC-Web on port 8080
	httpServer := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	log.Println("Starting gRPC-Web HTTP server on :8080")
	if err := httpServer.ListenAndServe(); err != nil {
		log.Fatalf("failed to serve HTTP: %v", err)
	}
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Incoming Request: %s %s", r.Method, r.URL.Path)
		log.Printf("Headers: %v", r.Header)

		// Log body for POST requests
		if r.Method == http.MethodPost {
			body, err := io.ReadAll(r.Body)
			if err != nil {
				log.Printf("Error reading request body: %v", err)
			} else {
				log.Printf("Raw Body: %s", body)

				contentType := r.Header.Get("Content-Type")
				if strings.Contains(contentType, "application/json") {
					var jsonBody interface{}
					if err := json.Unmarshal(body, &jsonBody); err != nil {
						log.Printf("Error unmarshaling JSON body: %v", err)
					} else {
						log.Printf("JSON Body: %+v", jsonBody)
					}
				} else if strings.Contains(contentType, "application/grpc-web+proto") {
					// For protobuf, we'll just log the raw body for now, as decoding it here requires specific message types.
					log.Printf("Protobuf Body (raw): %x", body)
				}
			}
			// Restore the body for the next handler
			r.Body = io.NopCloser(bytes.NewBuffer(body))
		}

		next.ServeHTTP(w, r)
	})
}
