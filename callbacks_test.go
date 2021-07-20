package auditableGorm

import (
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type User struct {
	ID    int64 `gorm:"auto:id"`
	Name  string
	Age   int
	Email string
}

func (u User) GetRequestUUID() string {
	return "uuidexample"
}

func (u User) GetRequestIP() string {
	return "127.0.0.1"
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

	audits := Audits{}
	db.First(&audits)

	assert.Equal(t, user.ID, audits.Auditable_id)
	assert.Equal(t, ACTION_CREATE, audits.Action)
	assert.Equal(t, "User", audits.Auditable_type)
	assert.Equal(t, int64(1), audits.Version)
	assert.Equal(t, "---\nID: 1\nName: Janderson\nAge: 28\nEmail: example@email.com", audits.Audited_changes)
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
