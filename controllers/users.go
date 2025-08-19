package controllers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/pranav244872/lenslocked.com/models"
)

// users controller
type Users struct {
	UserService *models.UserService
}

// constructor for users struct
func NewUsers(us *models.UserService) *Users {
	return &Users{
		UserService: us,
	}
}

// SignupForm defines the expected JSON structure for a new user signup
type SignupForm struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Dob      string `json:"dob"`
	Password string `json:"password"`
}

// Create handles the user signup process
// It expects a JSON payload and returns a JSON response
func (u *Users) Create(w http.ResponseWriter, r *http.Request) {
	var form SignupForm
	if err := parseJSON(r, &form); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
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

	// Send a sucess response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "User Created successfully!"})
}

// LoginForm defines the expected JSON structure for user login
type LoginForm struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Login handles the user authentication process
func (u *Users) Login(w http.ResponseWriter, r *http.Request) {
	form := LoginForm{}
	if err := parseJSON(r, &form); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	user, err := u.UserService.Authenticate(form.Email, form.Password)
	if err != nil {
		switch err {
		case models.ErrorNotFound:
			http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		default:
			http.Error(w, "Something went wrong", http.StatusInternalServerError)
		}
		return
	}

	// On successful authentication, we would typically generate and return a JWT
	// For now, we'll just return the user's data
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}
