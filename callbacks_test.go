package auditableGorm

import (
	"context"
	"log"
	"os"
	"path"
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

		assert.Equal(t, user.Id, audits.Auditable_id)
		assert.Equal(t, ACTION_CREATE, audits.Action)
		assert.Equal(t, "User", audits.Auditable_type)
		assert.Equal(t, int64(1), audits.Version)
		assert.Equal(t, IP, audits.Remote_address)
		assert.Equal(t, UUID, audits.Request_uuid)
		assert.Equal(t, "---\nname: Janderson\nage: 28\nemail: example@email.com\namount: 122.99", audits.Audited_changes)
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

		assert.Equal(t, user.Id, audits.Auditable_id)
		assert.Equal(t, ACTION_CREATE, audits.Action)
		assert.Equal(t, "User", audits.Auditable_type)
		assert.Equal(t, int64(1), audits.Version)
		assert.Equal(t, "", audits.Remote_address)
		assert.Equal(t, "", audits.Request_uuid)
		assert.Equal(t, "---\nname: Janderson\nage: 28\nemail: example@email.com\namount: 122.99", audits.Audited_changes)
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

		assert.Equal(t, user.Id, audits.Auditable_id)
		assert.Equal(t, ACTION_DELETE, audits.Action)
		assert.Equal(t, "User", audits.Auditable_type)
		assert.Equal(t, int64(1), audits.Version)
		assert.Equal(t, IP, audits.Remote_address)
		assert.Equal(t, UUID, audits.Request_uuid)
		assert.Equal(t, "---\nname: Janderson\nage: 28\nemail: example@email.com\namount: 122.99", audits.Audited_changes)
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

		assert.Equal(t, user.Id, audits.Auditable_id)
		assert.Equal(t, ACTION_DELETE, audits.Action)
		assert.Equal(t, "User", audits.Auditable_type)
		assert.Equal(t, int64(1), audits.Version)
		assert.Equal(t, "", audits.Remote_address)
		assert.Equal(t, "", audits.Request_uuid)
		assert.Equal(t, "---\nname: Janderson\nage: 28\nemail: example@email.com\namount: 122.99", audits.Audited_changes)
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

		assert.Equal(t, user.Id, audits.Auditable_id)
		assert.Equal(t, ACTION_UPDATE, audits.Action)
		assert.Equal(t, "User", audits.Auditable_type)
		assert.Equal(t, int64(1), audits.Version)
		assert.Equal(t, IP, audits.Remote_address)
		assert.Equal(t, UUID, audits.Request_uuid)
		assert.Contains(t, audits.Audited_changes, "\nname:\n- Janderson\n- Janderson Updated")
		assert.Contains(t, audits.Audited_changes, "\nemail:\n- example@email.com\n- updated@email.com")
		assert.Contains(t, audits.Audited_changes, "\nage:\n- 28\n- 12")
		assert.Contains(t, audits.Audited_changes, "\namount:\n- 122.99\n- 133.12")
		assert.Contains(t, audits.Audited_changes, "\ndesc:\n- \n- example")
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

		assert.Equal(t, user.Id, audits.Auditable_id)
		assert.Equal(t, ACTION_UPDATE, audits.Action)
		assert.Equal(t, "User", audits.Auditable_type)
		assert.Equal(t, int64(1), audits.Version)
		assert.Equal(t, "", audits.Remote_address)
		assert.Equal(t, "", audits.Request_uuid)
		assert.Contains(t, audits.Audited_changes, "\nname:\n- Janderson\n- Janderson Updated")
		assert.Contains(t, audits.Audited_changes, "\nemail:\n- example@email.com\n- updated@email.com")
		assert.Contains(t, audits.Audited_changes, "\ndesc:\n- \n- example")
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
func cleanDB(t *testing.T, db *gorm.DB) {
	db.Where("id > ?", 0).Delete(&User{})
	db.Where("id > ?", 0).Delete(&Audits{})
}

func connectDB(t *testing.T, withCtx bool) *gorm.DB {
	t.Helper()
	dsn := "host=localhost user=postgres password=root dbname=audit_test port=5432 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
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

	if withCtx {
		auditData := AuditData{UUID: UUID, Address: IP}
		ctx := context.WithValue(context.Background(), AUDIT_DATA_CTX_KEY, auditData)
		db = db.WithContext(ctx)
	}

	Register(db)

	return db
}
