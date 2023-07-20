package auditableGorm

import (
	"gorm.io/gorm"
)

type Plugin struct {
	db     *gorm.DB
	opts   options
	dbName string
}

func Register(db *gorm.DB, dbName string, opts ...Option) (Plugin, error) {
	o := options{}
	for _, option := range opts {
		option(&o)
	}
	p := Plugin{db: db, opts: o, dbName: dbName}
	db.Callback().Create().After("gorm:create").Register("audit:after_create", p.addCreated)
	db.Callback().Delete().Before("gorm:delete").Register("audit:after_delete", p.addDeleted)
	db.Callback().Update().Before("gorm:update").Register("audit:before_update", p.addUpdated)
	db.Callback().Update().After("gorm:query").Register("audit:after_query", p.addQuery)
	return p, nil
}
