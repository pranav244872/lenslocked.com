package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/pranav244872/lenslocked.com/config"
	"github.com/pranav244872/lenslocked.com/controllers"
	"github.com/pranav244872/lenslocked.com/models"
)

///////////////////////////////////////////////////////////////////////////////
// Helper Functions
///////////////////////////////////////////////////////////////////////////////

// must panics on any non-nil error
func must(err error) {
	if err != nil {
		panic(err)
	}
}

///////////////////////////////////////////////////////////////////////////////
// Main
///////////////////////////////////////////////////////////////////////////////

func main() {
	// Load environment variables from .env
	config.LoadEnv()

	// Connect to database
	dsn := config.GetDSN()
	userService, err := models.NewUserService(dsn)
	must(err)
	defer func() {
		log.Println("Closing database connection...")
		if err := userService.Close(); err != nil {
			log.Printf("Error closing DB: %v", err)
		}
	}()

	log.Println("Database connected")
	// Auto-migrate schema
	must(userService.DestructiveReset())

	// Controllers
	usersC := controllers.NewUsers(userService)

	// Router
	r := mux.NewRouter()

	// User routes
	r.HandleFunc("/api/signup", usersC.Create).Methods("POST")
	r.HandleFunc("/api/login", usersC.Login).Methods("POST")

	// CORS configuration
	allowedOrigins := handlers.AllowedOrigins([]string{"http://localhost:5173"})
	allowedMethods := handlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE", "OPTIONS"})
	allowedHeaders := handlers.AllowedHeaders([]string{"X-Requested-With", "Content-Type", "Authorization"})

	// Graceful shutdown
	go func() {
		log.Println("Server starting on :3000")
		if err := http.ListenAndServe(":3000", handlers.CORS(allowedOrigins, allowedMethods, allowedHeaders)(r)); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Could not start server: %v", err)
		}
	}()

	// Listen for Ctrl+C or SIGTERM
	waitForShutdown()
}

///////////////////////////////////////////////////////////////////////////////
// Graceful Shutdown Helper
///////////////////////////////////////////////////////////////////////////////

func waitForShutdown() {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	<-quit
	log.Println("Shutting down server gracefully...")
}
