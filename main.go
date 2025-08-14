package main

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/pranav244872/lenslocked.com/views"
)

var homeView *views.View
var contactView *views.View
var faqView *views.View

//////////////////////////////////////////////////////////////////////////
// Helper functions
//////////////////////////////////////////////////////////////////////////

// A helper function that panics on any error
func must(err error) {
	if err != nil {
		panic(err)
	}
}

//////////////////////////////////////////////////////////////////////////
// Handlers
//////////////////////////////////////////////////////////////////////////

// home page handler
func home(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	must(homeView.Render(w, nil))
}

// contact page handler
func contact(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	must(contactView.Render(w, nil))
}

// faq page handler
func faq(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	must(faqView.Render(w, nil))
}

// 404 page
func notFound(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
	fmt.Fprint(w, "<h1>Oops! This page does not exist 🛸</h1>")
}

//////////////////////////////////////////////////////////////////////////
// Main routing
//////////////////////////////////////////////////////////////////////////

func main() {
	homeView = views.NewView("main", "views/home.gohtml")
	contactView = views.NewView("main", "views/contact.gohtml")
	faqView = views.NewView("main", "views/faq.gohtml")

	r := mux.NewRouter()

	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))
	r.HandleFunc("/", home)
	r.HandleFunc("/contact", contact)
	r.HandleFunc("/faq", faq)

	r.NotFoundHandler = http.HandlerFunc(notFound)

	http.ListenAndServe(":3000", r)
}
