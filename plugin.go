package auditableGorm

import (
	"gorm.io/gorm"
)

type Plugin struct {
	db   *gorm.DB
	opts options
}

func Register(db *gorm.DB, opts ...Option) (Plugin, error) {
	o := options{}
	for _, option := range opts {
		option(&o)
	}
	p := Plugin{db: db, opts: o}
	callback := db.Callback()
	callback.Create().After("gorm:create").Register("loggable:create", p.addCreated)
	callback.Update().After("gorm:update").Register("loggable:update", p.addUpdated)
	callback.Delete().After("gorm:delete").Register("loggable:delete", p.addDeleted)
	return p, nil
}
