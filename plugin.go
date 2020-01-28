package auditableGorm

import (
	"github.com/jinzhu/gorm"
)

// Plugin is a hook for gorm.
type Plugin struct {
	db   *gorm.DB
	opts options
}

// Register initializes Plugin for provided gorm.DB.
// There is also available some options, that should be passed there.
// Options cannot be set after initialization.
func Register(db *gorm.DB, opts ...Option) (Plugin, error) {
	err := db.AutoMigrate(&Audits{}).Error
	if err != nil {
		return Plugin{}, err
	}
	o := options{}
	for _, option := range opts {
		option(&o)
	}
	p := Plugin{db: db, opts: o}
	callback := db.Callback()
	callback.Query().After("gorm:after_query").Register("loggable:query", p.trackEntity)
	callback.Create().After("gorm:after_create").Register("loggable:create", p.addCreated)
	callback.Update().After("gorm:after_update").Register("loggable:update", p.addUpdated)
	callback.Delete().After("gorm:after_delete").Register("loggable:delete", p.addDeleted)
	return p, nil
}