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
	db.Callback().Create().After("gorm:create").Register("audit:after_create", p.addCreated)
	db.Callback().Delete().Before("gorm:delete").Register("audit:after_delete", p.addDeleted)
	db.Callback().Update().Before("gorm:update").Register("audit:before_update", p.addUpdated)
	return p, nil
}
