package auditableGorm

import (
	"log"
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type User struct {
	Id    int64 `gorm:"auto:id"`
	Name  string
	Age   int
	Email string
	Desc  string
}

// func (u User) GetRequestUUID() string {
// 	return "uuidexample"
// }

// func (u User) GetRequestIP() string {
// 	return "127.0.0.1"
// }

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

	assert.Equal(t, user.Id, audits.Auditable_id)
	assert.Equal(t, ACTION_CREATE, audits.Action)
	assert.Equal(t, "User", audits.Auditable_type)
	assert.Equal(t, int64(1), audits.Version)
	assert.Equal(t, "---\nId: 1\nName: Janderson\nAge: 28\nEmail: example@email.com", audits.Audited_changes)
}

func TestAddDelete(t *testing.T) {
	cleanDB(t)
	db := connectDB(t)

	user := getUser()

	if err := db.Create(&user).Error; err != nil {
		t.Fatal(err)
	}

	if err := db.Delete(&user).Error; err != nil {
		t.Fatal(err)
	}

	audits := Audits{}
	db.Last(&audits)

	assert.Equal(t, user.Id, audits.Auditable_id)
	assert.Equal(t, ACTION_DELETE, audits.Action)
	assert.Equal(t, "User", audits.Auditable_type)
	assert.Equal(t, int64(1), audits.Version)
	assert.Equal(t, "---\nId: 1\nName: Janderson\nAge: 28\nEmail: example@email.com", audits.Audited_changes)
}

func TestAddUpdate(t *testing.T) {
	cleanDB(t)
	db := connectDB(t)

	user := getUser()

	if err := db.Create(&user).Error; err != nil {
		t.Fatal(err)
	}

	user.Name = "Janderson Updated"
	user.Email = "updated@email.com"
	user.Desc = "example"
	if err := db.Save(&user).Error; err != nil {
		t.Fatal(err)
	}

	audits := Audits{}
	db.Last(&audits)

	assert.Equal(t, user.Id, audits.Auditable_id)
	assert.Equal(t, ACTION_UPDATE, audits.Action)
	assert.Equal(t, "User", audits.Auditable_type)
	assert.Equal(t, int64(1), audits.Version)
	assert.Equal(t, "---\nname:\n- Janderson\n- Janderson Updated\nemail:\n- example@email.com\n- updated@email.com\ndesc:\n- example", audits.Audited_changes)
}

func getUser() User {
	return User{
		Name:  "Janderson",
		Age:   28,
		Email: "example@email.com"}
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
	db, err := gorm.Open(sqlite.Open(DB_PATH), &gorm.Config{
		Logger: logger.New(
			log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
			logger.Config{
				// SlowThreshold: time.Second, // Slow SQL threshold
				LogLevel: logger.Silent, // Log level
				Colorful: false,         // Disable color
			},
		),
	})
	if err != nil {
		t.Fatal(err)
	}

	sqlDb, err := db.DB()
	if err != nil {
		panic("failed to connect database")
	}
	sqlDb.SetMaxIdleConns(10)
	sqlDb.SetMaxOpenConns(10)

	db.AutoMigrate(&User{})
	db.AutoMigrate(&Audits{})

	Register(db)
	return db
}
