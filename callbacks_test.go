package auditableGorm

import (
	"context"
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type User struct {
	Id     int64 `gorm:"auto:id"`
	Name   string
	Age    int
	Email  string
	Desc   string
	Amount float64
}

var DB_PATH = path.Join("db", "test.db")

const (
	IP   = "127.0.0.1"
	UUID = "xxx-xxx"
)

func TestAddCreated(t *testing.T) {
	t.Run("with context", func(t *testing.T) {
		db := connectDB(t, true)
		cleanDB(t, db)

		user := User{
			Name:   "Janderson",
			Age:    28,
			Email:  "example@email.com",
			Amount: 122.99}

		if err := db.Create(&user).Error; err != nil {
			t.Fatal(err)
		}

		audits := Audits{}
		db.First(&audits)

		userMap, _ := getModelAsMap(user)
		originalMap, err := getOriginal(user.Id, "audit_test", "users")

		assert.Nil(t, err)
		assert.Equal(t, originalMap, userMap)

		assert.Equal(t, user.Id, audits.Auditable_id)
		assert.Equal(t, ACTION_CREATE, audits.Action)
		assert.Equal(t, "User", audits.Auditable_type)
		assert.Equal(t, int64(1), audits.Version)
		assert.Equal(t, IP, audits.Remote_address)
		assert.Equal(t, UUID, audits.Request_uuid)
		assert.Equal(t, "---\nName: Janderson\nAge: 28\nEmail: example@email.com\nAmount: 122.99", audits.Audited_changes)
		cleanAuditFiles()
	})

	t.Run("without context", func(t *testing.T) {
		db := connectDB(t, false)
		cleanDB(t, db)

		user := User{
			Name:   "Janderson",
			Age:    28,
			Email:  "example@email.com",
			Amount: 122.99}

		if err := db.Create(&user).Error; err != nil {
			t.Fatal(err)
		}

		audits := Audits{}
		db.First(&audits)

		userMap, _ := getModelAsMap(user)
		originalMap, err := getOriginal(user.Id, "audit_test", "users")

		assert.Nil(t, err)
		assert.Equal(t, originalMap, userMap)

		assert.Equal(t, user.Id, audits.Auditable_id)
		assert.Equal(t, ACTION_CREATE, audits.Action)
		assert.Equal(t, "User", audits.Auditable_type)
		assert.Equal(t, int64(1), audits.Version)
		assert.Equal(t, "", audits.Remote_address)
		assert.Equal(t, "", audits.Request_uuid)
		assert.Equal(t, "---\nName: Janderson\nAge: 28\nEmail: example@email.com\nAmount: 122.99", audits.Audited_changes)
		cleanAuditFiles()
	})
}

func TestAddDelete(t *testing.T) {
	t.Run("with context", func(t *testing.T) {
		db := connectDB(t, true)
		cleanDB(t, db)

		user := getUser()

		if err := db.Create(&user).Error; err != nil {
			t.Fatal(err)
		}

		if err := db.Delete(&user).Error; err != nil {
			t.Fatal(err)
		}

		audits := Audits{}
		db.Last(&audits)

		_, err := getOriginal(user.Id, "audit_test", "users")

		fmt.Println(err.Error())
		assert.NotNil(t, err)

		assert.Equal(t, user.Id, audits.Auditable_id)
		assert.Equal(t, ACTION_DELETE, audits.Action)
		assert.Equal(t, "User", audits.Auditable_type)
		assert.Equal(t, int64(1), audits.Version)
		assert.Equal(t, IP, audits.Remote_address)
		assert.Equal(t, UUID, audits.Request_uuid)
		assert.Equal(t, "---\nName: Janderson\nAge: 28\nEmail: example@email.com\nAmount: 122.99", audits.Audited_changes)
		cleanAuditFiles()
	})

	t.Run("without context", func(t *testing.T) {
		db := connectDB(t, false)
		cleanDB(t, db)

		user := getUser()

		if err := db.Create(&user).Error; err != nil {
			t.Fatal(err)
		}

		if err := db.Delete(&user).Error; err != nil {
			t.Fatal(err)
		}

		audits := Audits{}
		db.Last(&audits)

		_, err := getOriginal(user.Id, "audit_test", "users")

		assert.NotNil(t, err)

		assert.Equal(t, user.Id, audits.Auditable_id)
		assert.Equal(t, ACTION_DELETE, audits.Action)
		assert.Equal(t, "User", audits.Auditable_type)
		assert.Equal(t, int64(1), audits.Version)
		assert.Equal(t, "", audits.Remote_address)
		assert.Equal(t, "", audits.Request_uuid)
		assert.Equal(t, "---\nName: Janderson\nAge: 28\nEmail: example@email.com\nAmount: 122.99", audits.Audited_changes)
		cleanAuditFiles()
	})
}

func TestAddUpdate(t *testing.T) {
	t.Run("with context", func(t *testing.T) {
		db := connectDB(t, true)
		cleanDB(t, db)

		user := getUser()

		if err := db.Create(&user).Error; err != nil {
			t.Fatal(err)
		}

		user.Name = "Janderson Updated"
		user.Email = "updated@email.com"
		user.Desc = "example"
		user.Age = 12
		user.Amount = 133.12
		if err := db.Save(&user).Error; err != nil {
			t.Fatal(err)
		}

		audits := Audits{}
		db.Last(&audits)

		userMap, _ := getModelAsMap(user)
		originalMap, err := getOriginal(user.Id, "audit_test", "users")

		assert.Nil(t, err)
		assert.Equal(t, originalMap, userMap)

		assert.Equal(t, user.Id, audits.Auditable_id)
		assert.Equal(t, ACTION_UPDATE, audits.Action)
		assert.Equal(t, "User", audits.Auditable_type)
		assert.Equal(t, int64(1), audits.Version)
		assert.Equal(t, IP, audits.Remote_address)
		assert.Equal(t, UUID, audits.Request_uuid)
		assert.Contains(t, audits.Audited_changes, "\nName:\n- Janderson\n- Janderson Updated")
		assert.Contains(t, audits.Audited_changes, "\nEmail:\n- example@email.com\n- updated@email.com")
		assert.Contains(t, audits.Audited_changes, "\nAge:\n- 28\n- 12")
		assert.Contains(t, audits.Audited_changes, "\nAmount:\n- 122.99\n- 133.12")
		assert.Contains(t, audits.Audited_changes, "\nDesc:\n- \n- example")
		cleanAuditFiles()
	})

	t.Run("without context", func(t *testing.T) {
		db := connectDB(t, false)
		cleanDB(t, db)

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

		userMap, _ := getModelAsMap(user)
		originalMap, err := getOriginal(user.Id, "audit_test", "users")

		assert.Nil(t, err)
		assert.Equal(t, originalMap, userMap)

		assert.Equal(t, user.Id, audits.Auditable_id)
		assert.Equal(t, ACTION_UPDATE, audits.Action)
		assert.Equal(t, "User", audits.Auditable_type)
		assert.Equal(t, int64(1), audits.Version)
		assert.Equal(t, "", audits.Remote_address)
		assert.Equal(t, "", audits.Request_uuid)
		assert.Contains(t, audits.Audited_changes, "\nName:\n- Janderson\n- Janderson Updated")
		assert.Contains(t, audits.Audited_changes, "\nEmail:\n- example@email.com\n- updated@email.com")
		assert.Contains(t, audits.Audited_changes, "\nDesc:\n- \n- example")
		cleanAuditFiles()
	})
}

func TestAddQuery(t *testing.T) {
	t.Run("with context", func(t *testing.T) {
		db := connectDB(t, true)
		cleanDB(t, db)

		user := getUser()

		if err := db.Create(&user).Error; err != nil {
			t.Fatal(err)
		}

		db.Find(&user, "id = ?", user.Id)
		userMap, _ := getModelAsMap(user)
		originalMap, err := getOriginal(user.Id, "audit_test", "users")

		assert.Nil(t, err)
		assert.Equal(t, originalMap, userMap)
		cleanAuditFiles()
	})

	t.Run("without context", func(t *testing.T) {
		db := connectDB(t, false)
		cleanDB(t, db)

		user := getUser()

		if err := db.Create(&user).Error; err != nil {
			t.Fatal(err)
		}

		db.Find(&user, "id = ?", user.Id)
		userMap, _ := getModelAsMap(user)
		originalMap, err := getOriginal(user.Id, "audit_test", "users")

		assert.Nil(t, err)
		assert.Equal(t, originalMap, userMap)
		cleanAuditFiles()
	})
}

func getUser() User {
	return User{
		Name:   "Janderson",
		Age:    28,
		Email:  "example@email.com",
		Amount: 122.99}
}

// cleanDB to always run tests on a fresh db
func cleanAuditFiles() {
	d, _ := os.Open("tmp/")
	defer d.Close()
	files, _ := d.Readdir(-1)

	for _, file := range files {
		if file.Mode().IsRegular() {
			if filepath.Ext(file.Name()) == ".audit" {
				err := os.Remove(fmt.Sprintf("tmp/%s", file.Name()))
				fmt.Println(err)
			}
		}
	}
}
func cleanDB(t *testing.T, db *gorm.DB) {
	_ = db.Exec("TRUNCATE TABLE users").Error
	_ = db.Exec("TRUNCATE TABLE audits").Error
	_ = db.Exec("SELECT SETVAL('users_id_seq', COALESCE(MAX(id), 1) ) FROM users;").Error
	_ = db.Exec("SELECT SETVAL('audits_id_seq', COALESCE(MAX(id), 1) ) FROM audits;").Error

}

func connectDB(t *testing.T, withCtx bool) *gorm.DB {
	t.Helper()
	dsn := "host=localhost user=doadmin password=0 dbname=audit_test port=5432 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.New(
			log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
			logger.Config{
				// SlowThreshold: time.Second, // Slow SQL threshold
				LogLevel: logger.Silent, // Log level
				Colorful: true,          // Disable color
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

	if withCtx {
		auditData := AuditData{UUID: UUID, Address: IP}
		ctx := context.WithValue(context.Background(), AUDIT_DATA_CTX_KEY, auditData)
		db = db.WithContext(ctx)
	}

	Register(db, "audit_test")

	return db
}
