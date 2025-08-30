package controllers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/pranav244872/lenslocked.com/models"
)

///////////////////////////////////////////////////////////////////////////////
// Users Controller
///////////////////////////////////////////////////////////////////////////////

// Users controller struct to handle user-related routes
type Users struct {
	UserService models.UserService
}

// Constructor for Users controller
func NewUsers(us models.UserService) *Users {
	return &Users{
		UserService: us,
	}
}

///////////////////////////////////////////////////////////////////////////////
// User Signup
///////////////////////////////////////////////////////////////////////////////

// SignupForm defines the expected JSON structure for a new user signup
type SignupForm struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Create handles the user signup process and it automatically logs the user in
// It expects a JSON payload and returns a JSON response
func (u *Users) Create(w http.ResponseWriter, r *http.Request) {
	var form SignupForm
	if err := parseJSON(r, &form); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	user := models.User{
		Name:     form.Name,
		Email:    form.Email,
		Password: form.Password,
	}

	if err := u.UserService.DB.Create(&user); err != nil {
		// Check for the specific error for a duplicate email.
		if errors.Is(err, models.ErrorEmailTaken) {
			http.Error(w, err.Error(), http.StatusConflict) // 409 Conflict
			return
		}

		http.Error(w, err.Error(), http.StatusBadRequest) // 400 Bad Request
		return
	}

	// automatically sign the user in after they create an account
	if err := u.signIn(w, &user); err != nil {
		http.Error(w, "Something went wrong during sign-in", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "User created and logged in successfully!"})
}

///////////////////////////////////////////////////////////////////////////////
// User Login
///////////////////////////////////////////////////////////////////////////////

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
		// 3. Relay the Result
		switch err {
		case models.ErrorNotFound, models.ErrorIncorrectPassword:
			http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		default:
			http.Error(w, "Something went wrong", http.StatusInternalServerError)
		}
		return
	}

	if err := u.signIn(w, user); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Respond with user data (or token, in a real app)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Login successful!"})
}

// CookieTest acts as a protected endpoint to verify a user's session.
func (u *Users) CookieTest(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("remember_token")
	if err != nil {
		http.Error(w, "Authentication required", http.StatusUnauthorized)
		return
	}

	user, err := u.UserService.DB.ByRememberToken(cookie.Value)
	if err != nil {
		http.Error(w, "Invalid session token", http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

///////////////////////////////////////////////////////////////////////////////
// Helper functions
///////////////////////////////////////////////////////////////////////////////

// signIn helper function sign in users via cookies
func (u *Users) signIn(w http.ResponseWriter, user *models.User) error {
	if err := u.UserService.DB.Update(user); err != nil {
		return err
	}

	cookie := http.Cookie{
		Name:     "remember_token",
		Value:    user.Remember,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteNoneMode,
	}

	http.SetCookie(w, &cookie)
	return nil
}
