package auditableGorm

import (
	"os"
	"path"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type User struct {
	ID    int64 `gorm:"auto:id"`
	Name  string
	Age   int
	Email string
}

var DB_PATH = path.Join("db", "test.db")

func TestAddCreated(t *testing.T) {
	cleanDB(t)
	db := connectDB(t)

	user := User{
		Name:  "Janderson",
		Age:   28,
		Email: "example@email.com"}

	if err := db.Create(&user).Error; err != nil {
		t.Fatal(err)
	}
}

// cleanDB to always run tests on a fresh db
func cleanDB(t *testing.T) {
	t.Helper()
	if _, err := os.Stat(DB_PATH); err == nil {
		if err := os.Remove(DB_PATH); err != nil {
			t.Fatal(err)
		}
	}
}

func connectDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(DB_PATH), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	Register(db)
	db.AutoMigrate(&User{})
	db.AutoMigrate(&Audits{})
	return db
}
