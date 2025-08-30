package models

import (
	"errors"
	"net/mail"
	"strings"
	"time"

	"github.com/pranav244872/lenslocked.com/config"
	"github.com/pranav244872/lenslocked.com/hash"
	"github.com/pranav244872/lenslocked.com/rand"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

///////////////////////////////////////////////////////////////////////////////
// Errors
///////////////////////////////////////////////////////////////////////////////

var (
	ErrorNotFound          = errors.New("models: resource not found")
	ErrorInvalidId         = errors.New("models: ID provided was invalid")
	ErrorIncorrectPassword = errors.New("models: incorrect password provided")
	ErrorEmailTaken        = errors.New("models: email address is already in use")
)

///////////////////////////////////////////////////////////////////////////////
// User Model
///////////////////////////////////////////////////////////////////////////////

type User struct {
	ID           int64 `gorm:"primaryKey;autoIncrement"`
	Name         string
	Email        string `gorm:"not null;uniqueIndex"`
	Password     string `gorm:"-"`
	PasswordHash string `gorm:"not null"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
	DeletedAt    gorm.DeletedAt `gorm:"index"`
	Remember     string         `gorm:"-"`
	RememberHash string         `gorm:"not null;uniqueIndex"`
}

///////////////////////////////////////////////////////////////////////////////
// Database layer interface
///////////////////////////////////////////////////////////////////////////////

type UserDB interface {
	// Create
	Create(user *User) error

	// Read
	ByID(id int64) (*User, error)
	ByEmail(email string) (*User, error)
	ByRememberToken(token string) (*User, error)

	// Update
	Update(user *User) error

	// Delete
	Delete(id int64) error

	// Lifecycle Methods
	Close() error
	AutoMigrate() error
	DestructiveReset() error
}

///////////////////////////////////////////////////////////////////////////////
// Service Layer
///////////////////////////////////////////////////////////////////////////////

type UserService struct {
	DB UserDB
}

func NewUserService(connectionInfo string) (*UserService, error) {

	// Create shared tools first
	hmac := hash.NewHMAC(config.HMACKey)

	// Create db layer implementation
	ug, err := newUserGorm(connectionInfo)
	if err != nil {
		return nil, err
	}

	// create validation layer
	uv := newUserValidator(ug, hmac)

	// Create service layer
	us := &UserService{
		DB: uv,
	}

	return us, nil
}

// Authenticate
func (us *UserService) Authenticate(email, password string) (*User, error) {
	foundUser, err := us.DB.ByEmail(email)
	if err != nil {
		// Includes ErrorNotFound
		return nil, err
	}

	err = bcrypt.CompareHashAndPassword(
		[]byte(foundUser.PasswordHash),
		[]byte(password+config.PassPepper),
	)

	switch err {
	case nil:
		return foundUser, nil
	case bcrypt.ErrMismatchedHashAndPassword:
		return nil, ErrorIncorrectPassword
	default:
		return nil, err
	}
}

///////////////////////////////////////////////////////////////////////////////
// Validation Layer
///////////////////////////////////////////////////////////////////////////////

type userValidator struct {
	UserDB
	hmac hash.HMAC
}

func newUserValidator(nextLayer UserDB, hmac hash.HMAC) *userValidator {
	return &userValidator{
		UserDB: nextLayer,
		hmac:   hmac,
	}
}

// Create
func (uv *userValidator) Create(user *User) error {
	err := runUserValFns(user,
		uv.passwordRequired,
		uv.passwordLength,
		uv.hashPassword,
		uv.passwordHashRequired,
		uv.setRemember,
		uv.rememberMinBytes,
		uv.hashRemember,
		uv.rememberHashRequired,
		uv.normalizeEmail,
		uv.requireEmail,
		uv.emailFormat,
		uv.emailIsAvail,
	)
	if err != nil {
		return err
	}
	return uv.UserDB.Create(user)
}

// Read By RememberToken
func (uv *userValidator) ByRememberToken(token string) (*User, error) {
	rememberHash := uv.hmac.Hash(token)
	return uv.UserDB.ByRememberToken(rememberHash)
}

func (uv *userValidator) Update(user *User) error {
	err := runUserValFns(user,
		uv.passwordLength,
		uv.hashPassword, // Will be skipped if password is ""
		uv.passwordHashRequired,
		uv.hashRemember, // Will be skipped if remember is ""
		uv.rememberHashRequired,
		uv.normalizeEmail,
		uv.requireEmail,
		uv.emailFormat,
		uv.emailIsAvail,
	)
	if err != nil {
		return err
	}
	return uv.UserDB.Update(user)
}

// Delete
func (uv *userValidator) Delete(id int64) error {
	var user User
	user.ID = id
	err := runUserValFns(&user, uv.idGreaterThan(0))
	if err != nil {
		return err
	}
	return uv.UserDB.Delete(id)
}

// --- Validation Helpers ---

type userValFn func(*User) error

func runUserValFns(user *User, fns ...userValFn) error {
	for _, fn := range fns {
		if err := fn(user); err != nil {
			return err
		}
	}
	return nil
}

func (uv *userValidator) idGreaterThan(n int64) userValFn {
	return func(user *User) error {
		if user.ID <= n {
			return ErrorInvalidId
		}
		return nil
	}
}

func (uv *userValidator) passwordRequired(user *User) error {
	if user.Password == "" {
		return errors.New("password is required")
	}
	return nil
}

func (uv *userValidator) passwordLength(user *User) error {
	if user.Password == "" {
		return nil
	}
	if len(user.Password) < 8 {
		return errors.New("password must be at least 8 characters long")
	}
	return nil
}

func (uv *userValidator) hashPassword(user *User) error {
	if user.Password == "" {
		return nil
	}
	pwBytes := []byte(user.Password + config.PassPepper)
	hashedBytes, err := bcrypt.GenerateFromPassword(pwBytes, bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	user.PasswordHash = string(hashedBytes)
	user.Password = ""
	return nil
}

func (uv *userValidator) passwordHashRequired(user *User) error {
	if user.PasswordHash == "" {
		return errors.New("password hashing failed")
	}
	return nil
}

func (uv *userValidator) setRemember(user *User) error {
	if user.Remember != "" {
		return nil
	}
	token, err := rand.RememberToken()
	if err != nil {
		return err
	}
	user.Remember = token
	return nil
}

func (uv *userValidator) rememberMinBytes(user *User) error {
	l, err := rand.NBytes(user.Remember)

	if err != nil {
		return err
	}

	if l != 32 {
		return errors.New("remember token length is not 32")
	}
	return nil
}

func (uv *userValidator) hashRemember(user *User) error {
	if user.Remember == "" {
		return nil
	}
	user.RememberHash = uv.hmac.Hash(user.Remember)
	return nil
}

func (uv *userValidator) rememberHashRequired(user *User) error {
	if user.RememberHash == "" {
		return errors.New("remember hashing failed")
	}
	return nil
}

func (uv *userValidator) normalizeEmail(user *User) error {
	user.Email = strings.ToLower(user.Email)
	user.Email = strings.TrimSpace(user.Email)
	return nil
}

func (uv *userValidator) requireEmail(user *User) error {
	if user.Email == "" {
		return errors.New("email is required")
	}
	return nil
}

func (uv *userValidator) emailFormat(user *User) error {
	if user.Email == "" {
		return nil
	}
	_, err := mail.ParseAddress(user.Email)
	if err != nil {
		return errors.New("email is not a valid format")
	}
	return nil
}

func (uv *userValidator) emailIsAvail(user *User) error {
	existing, err := uv.ByEmail(user.Email)
	if err == ErrorNotFound {
		return nil
	}
	if err != nil {
		return err
	}
	if user.ID != existing.ID {
		return ErrorEmailTaken
	}
	return nil
}

///////////////////////////////////////////////////////////////////////////////
// Database Layer
///////////////////////////////////////////////////////////////////////////////

// this is implementation of UserDB interface
type userGorm struct {
	db *gorm.DB
}

// constructor which returns an instance of userGorm
func newUserGorm(connectionInfo string) (*userGorm, error) {
	// initialize db connection
	db, err := gorm.Open(postgres.Open(connectionInfo), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})

	if err != nil {
		return nil, err
	}

	return &userGorm{
		db: db,
	}, nil
}

// --- CRUD operations ---

// Create user
func (ug *userGorm) Create(user *User) error {
	// create db entry
	return ug.db.Create(user).Error
}

// Retrieve by id
func (ug *userGorm) ByID(id int64) (*User, error) {
	var user User
	db := ug.db.Where("id = ?", id)
	err := first(db, &user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// Retrieve by email
func (ug *userGorm) ByEmail(email string) (*User, error) {
	var user User

	db := ug.db.Where("email = ?", email)
	err := first(db, &user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// Retrive by given remember token
func (ug *userGorm) ByRememberToken(rememberHash string) (*User, error) {
	var user User

	// Look for that hash in our database
	err := ug.db.Where("remember_hash = ?", rememberHash).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// Update
func (ug *userGorm) Update(user *User) error {
	return ug.db.Save(user).Error
}

// Delete
func (ug *userGorm) Delete(id int64) error {
	user := User{ID: id}
	result := ug.db.Delete(&user)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return ErrorNotFound
	}

	return nil
}

// closes the database connection.
func (ug *userGorm) Close() error {
	sqlDB, err := ug.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// runs the GORM auto-migration.
func (ug *userGorm) AutoMigrate() error {
	return ug.db.AutoMigrate(&User{})
}

// drops and rebuilds the user table.
func (ug *userGorm) DestructiveReset() error {
	if err := ug.db.Migrator().DropTable(&User{}); err != nil {
		return err
	}
	return ug.AutoMigrate()
}

///////////////////////////////////////////////////////////////////////////////
// Helper functions
///////////////////////////////////////////////////////////////////////////////

func first(db *gorm.DB, dst any) error {
	err := db.First(dst).Error
	if err == gorm.ErrRecordNotFound {
		return ErrorNotFound
	}
	return err
}

func all(db *gorm.DB, dst any) error {
	return db.Find(dst).Error
}
