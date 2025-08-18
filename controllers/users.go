package controllers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/pranav244872/lenslocked.com/models"
	"github.com/pranav244872/lenslocked.com/views"
)

// users controller
type Users struct {
	NewView     *views.View
	LoginView   *views.View
	UserService *models.UserService
}

// constructor for users struct
func NewUsers(us *models.UserService) *Users {
	return &Users{
		NewView:     views.NewView("main", "users/new"),
		LoginView:   views.NewView("main", "users/login"),
		UserService: us,
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
	Name     string `schema:"name"`
	Email    string `schema:"email"`
	Dob      string `schema:"dob"`
	Password string `schema:"password"`
}

// Create is used to process the signup form when a user tries to create a new account
// POST /signup
func (u *Users) Create(w http.ResponseWriter, r *http.Request) {
	var form SignupForm
	if err := parseForm(r, &form); err != nil {
		panic(err)
	}

	dob, err := time.Parse("2006-01-02", form.Dob)
	if err != nil {
		http.Error(w, "Invalid date format", http.StatusBadRequest)
		return
	}

	user := models.User{
		Name:     form.Name,
		Email:    form.Email,
		Password: form.Password,
		DOB:      dob,
	}

	if err := u.UserService.Create(&user); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Fprintln(w, "User is", user)
}

// the data we will get through :POST at /login
type LoginForm struct {
	Email    string `schema:"email"`
	Password string `schema:"password"`
}

func (u *Users) Login(w http.ResponseWriter, r *http.Request) {
	form := LoginForm{}
	if err := parseForm(r, &form); err != nil {
		panic(err)
	}
	user, err := u.UserService.Authenticate(form.Email, form.Password)
	switch err {
	case models.ErrorNotFound:
		fmt.Fprintln(w, "Invalid email address.")
	case models.ErrorInvalidPassword:
		fmt.Fprintln(w, "Invalid password provided")
	case nil:
		fmt.Fprintln(w, user)
	default:
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
