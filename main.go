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
		if err := userService.DB.Close(); err != nil {
			log.Printf("Error closing DB: %v", err)
		}
	}()

	log.Println("Database connected")
	// Auto-migrate schema
	must(userService.DB.AutoMigrate())

	// Controllers
	usersC := controllers.NewUsers(userService)

	// Router
	r := mux.NewRouter()

	// User routes
	r.HandleFunc("/api/signup", usersC.Create).Methods("POST")
	r.HandleFunc("/api/login", usersC.Login).Methods("POST")
	r.HandleFunc("/api/cookietest", usersC.CookieTest).Methods("GET")

	// CORS configuration
	allowedOrigins := handlers.AllowedOrigins([]string{config.ClientOrigin})
	allowedMethods := handlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE", "OPTIONS"})
	allowedHeaders := handlers.AllowedHeaders([]string{"X-Requested-With", "Content-Type", "Authorization"})
	allowedCredentials := handlers.AllowCredentials()

	// Graceful shutdown
	go func() {
		addr := config.ServerHost + ":" + config.ServerPort
		log.Println("Server starting on https://" + addr)

		handler := handlers.CORS(
			allowedOrigins, allowedMethods, allowedHeaders, allowedCredentials,
		)(r)

		if err := http.ListenAndServeTLS(":"+config.ServerPort, config.CertFile, config.KeyFile, handler); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Could not start HTTPS server on https://%s: %v", addr, err)
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
