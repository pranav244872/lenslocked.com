package models

import (
	"errors"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

///////////////////////////////////////////////////////////////////////////////
// Errors
///////////////////////////////////////////////////////////////////////////////

var (
	ErrorNotFound  = errors.New("models: resource not found")
	ErrorInvalidId = errors.New("models: ID provided was invalid")
)

///////////////////////////////////////////////////////////////////////////////
// User Model
///////////////////////////////////////////////////////////////////////////////

type User struct {
	ID        int64 `gorm:"primaryKey;autoIncrement"`
	Name      string
	Email     string    `gorm:"not null;uniqueIndex"`
	DOB       time.Time `gorm:"not null;check:dob <= CURRENT_DATE - INTERVAL '18 years' AND dob >= CURRENT_DATE - INTERVAL '150 years'"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

// User Service, through which we will interact with the users table
type UserService struct {
	db *gorm.DB
}

// constructor which returns an instance of UserService
func NewUserService(connectionInfo string) (*UserService, error) {
	db, err := gorm.Open(postgres.Open(connectionInfo), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, err
	}

	return &UserService{
		db: db,
	}, nil
}

///////////////////////////////////////////////////////////////////////////////
// Lifecycle Methods
/////////////////////////////////////////////////////////////////////////////////

// Method to close the UserService db connection
func (us *UserService) Close() error {
	sqlDB, err := us.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// Create table
func (us *UserService) AutoMigrate() error {
	return us.db.AutoMigrate(&User{})
}

// DestructiveReset drops the user table and rebuilds it
func (us *UserService) DestructiveReset() error {
	err := us.db.Migrator().DropTable(&User{})
	if err != nil {
		return err
	}
	return us.AutoMigrate()
}

///////////////////////////////////////////////////////////////////////////////
// CRUD Methods
///////////////////////////////////////////////////////////////////////////////

// Create user
func (us *UserService) Create(user *User) error {
	return us.db.Create(user).Error
}

// Retrieve by id
func (us *UserService) ByID(id int64) (*User, error) {
	var user User
	db := us.db.Where("id = ?", id)
	err := first(db, &user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// Retrieve by email
func (us *UserService) ByEmail(email string) (*User, error) {
	var user User

	db := us.db.Where("email = ?", email)
	err := first(db, &user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// Get all users with the specific age
func (us *UserService) ByAge(age uint8) ([]User, error) {
	var users []User

	now := time.Now()

	// Calculate date range for people who are exactly `age` years old
	// Born between (today - age - 1 year + 1 day) and (today - age)
	latestDOB := now.AddDate(-int(age), 0, 0)
	earliestDOB := now.AddDate(-int(age)-1, 0, 1)

	db := us.db.Where("dob >= ? AND dob <= ?", earliestDOB, latestDOB)
	err := all(db, &users)
	if err != nil {
		return nil, err
	}
	return users, nil
}

// Get all users within an age range
func (us *UserService) ByAgeRange(minAge uint8, maxAge uint8) ([]User, error) {
	var users []User

	now := time.Now()

	// Calculate date range for people who are exactly `age` years old
	// Born between (today - age - 1 year + 1 day) and (today - age)
	latestDOB := now.AddDate(-int(minAge), 0, 0)
	earliestDOB := now.AddDate(-int(maxAge)-1, 0, 1)

	db := us.db.Where("dob >= ? AND dob <= ?", earliestDOB, latestDOB)
	err := all(db, &users)
	if err != nil {
		return nil, err
	}
	return users, nil
}

// Update
func (us *UserService) Update(user *User) error {
	return us.db.Save(user).Error
}

// Delete
func (us *UserService) Delete(id int64) error {
	if id == 0 {
		return ErrorInvalidId
	}

	user := User{ID: id}
	result := us.db.Delete(&user)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return ErrorNotFound
	}

	return nil
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
