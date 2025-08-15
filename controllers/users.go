package controllers

import (
	"fmt"
	"net/http"

	"github.com/pranav244872/lenslocked.com/views"
)

// users controller
type Users struct {
	NewView *views.View
}

// constructor for users struct
func NewUsers() *Users {
	return &Users{
		NewView: views.NewView("main", "users/new"),
	}
}

// New is used to render the form where a user can create a new user account
// GET /signup
// this is the handler
func (u *Users) New(w http.ResponseWriter, r *http.Request) {
	if err := u.NewView.Render(w, nil); err != nil {
		panic(err)
	}
}

// the data we will get through POST /signup
type SignupForm struct {
	Email    string `schema:"email"`
	Password string `schema:"password"`
}

// Create is used to process the signup form when a user tries to create a new account
// POST /signup
func (u *Users) Create(w http.ResponseWriter, r *http.Request) {
	var form SignupForm
	if err := parseForm(r, &form); err != nil {
		panic(err)
	}
	fmt.Fprintln(w, "Email is: ", form.Email)
	fmt.Fprintln(w, "Password is: ", form.Password)
}
