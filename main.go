package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

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

// notFound is a 404 handler
func notFound(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
	fmt.Fprint(w, "<h1>Oops! This page does not exist 🛸</h1>")
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
	staticC := controllers.NewStatic()
	usersC := controllers.NewUsers(userService)

	// Router
	r := mux.NewRouter()

	// Static assets
	r.PathPrefix("/static/").Handler(
		http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))),
	)

	// Static pages
	r.Handle("/", staticC.Home).Methods("GET")
	r.Handle("/contact", staticC.Contact).Methods("GET")
	r.Handle("/faq", staticC.FAQ).Methods("GET")

	// User routes
	r.HandleFunc("/signup", usersC.New).Methods("GET")
	r.HandleFunc("/signup", usersC.Create).Methods("POST")

	// 404
	r.NotFoundHandler = http.HandlerFunc(notFound)

	// Graceful shutdown
	go func() {
		log.Println("Server starting on :3000")
		if err := http.ListenAndServe(":3000", r); err != nil && err != http.ErrServerClosed {
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
