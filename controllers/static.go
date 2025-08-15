package controllers

import "github.com/pranav244872/lenslocked.com/views"

// Static struct to serve static pages
type Static struct {
	Home *views.View
	Contact *views.View
	FAQ *views.View
}

// Constructor for Static struct, returns an instance of Static struct
func NewStatic() *Static {
	return &Static{
		Home: views.NewView("main", "static/home"),
		Contact: views.NewView("main", "static/contact"),
		FAQ: views.NewView("main", "static/faq"),
	}
}
