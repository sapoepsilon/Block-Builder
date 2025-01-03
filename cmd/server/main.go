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

	"docker-management-system/internal/api/handlers"
	"docker-management-system/internal/docker"
	gorillaHandlers "github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	httpSwagger "github.com/swaggo/http-swagger"
)

// HealthCheckResponse is the response structure for health check
type HealthCheckResponse struct {
	Status string `json:"status"`
}

// loggingMiddleware logs HTTP request details
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		log.Printf("Started %s %s", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
		log.Printf("Completed %s %s in %v", r.Method, r.URL.Path, time.Since(start))
	})
}

// main function
func main() {
	// Initialize router with logging middleware
	router := mux.NewRouter()
	router.Use(loggingMiddleware)
	
	// Add CORS middleware
	corsMiddleware := gorillaHandlers.CORS(
		gorillaHandlers.AllowedOrigins([]string{"*"}),
		gorillaHandlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}),
		gorillaHandlers.AllowedHeaders([]string{"Content-Type", "Authorization", "X-Requested-With"}),
		gorillaHandlers.AllowCredentials(),
	)
	
	// Apply CORS middleware to all routes
	handler := corsMiddleware(router)

	// Initialize Docker client
	dockerClient, err := docker.NewClient("unix:///var/run/docker.sock", "", false, "")
	if err != nil {
		log.Fatalf("Failed to create Docker client: %v", err)
	}

	// Initialize container handler
	containerHandler := handlers.NewContainerHandler(dockerClient)

	// Register routes
	router.HandleFunc("/health", healthCheckHandler).Methods("GET", "OPTIONS")

	// Container routes with explicit OPTIONS handling
	apiRouter := router.PathPrefix("/api/v1").Subrouter()
	apiRouter.HandleFunc("/containers", containerHandler.ListContainers).Methods("GET", "OPTIONS")
	apiRouter.HandleFunc("/containers/{id}", containerHandler.GetContainer).Methods("GET", "OPTIONS")
	apiRouter.HandleFunc("/containers/{id}/logs", containerHandler.GetContainerLogs).Methods("GET", "OPTIONS")
	apiRouter.HandleFunc("/containers/{id}", containerHandler.DeleteContainer).Methods("DELETE", "OPTIONS")

	// Legacy routes without /api/v1 prefix for backward compatibility
	router.HandleFunc("/containers", containerHandler.ListContainers).Methods("GET", "OPTIONS")
	router.HandleFunc("/containers/{id}", containerHandler.GetContainer).Methods("GET", "OPTIONS")
	router.HandleFunc("/containers/{id}/logs", containerHandler.GetContainerLogs).Methods("GET", "OPTIONS")
	router.HandleFunc("/containers/{id}", containerHandler.DeleteContainer).Methods("DELETE", "OPTIONS")

	// Serve Swagger files
	router.PathPrefix("/swagger/").Handler(http.StripPrefix("/swagger/", http.FileServer(http.Dir("docs"))))

	// Swagger UI
	router.PathPrefix("/swagger-ui/").Handler(httpSwagger.Handler(
		httpSwagger.URL("/swagger/swagger.json"),
		httpSwagger.DeepLinking(true),
		httpSwagger.DocExpansion("none"),
		httpSwagger.DomID("swagger-ui"),
	))

	// Create a new HTTP server with timeouts
	srv := &http.Server{
		Handler:      handler,  // Use the wrapped handler with CORS
		Addr:         ":8080",
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Channel to listen for interrupt signals
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Start the server in a goroutine
	go func() {
		log.Printf("Starting server on %s...", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	<-quit
	log.Println("Shutting down server...")
	
	// Create a deadline for shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Attempt graceful shutdown
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("Server shutdown error: %v", err)
		log.Fatal("Server forced to shutdown")
	}

	log.Println("Server gracefully stopped")
}

// healthCheckHandler handles the health check requests
func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	response := HealthCheckResponse{Status: "UP"}
	w.Header().Set("Content-Type", "application/json")
	
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding health check response: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}
