package auditableGorm

import (
	"github.com/google/uuid"
	"github.com/jinzhu/gorm"
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
	callback.Query().After("gorm:after_query").Register("loggable:query", p.trackEntity)
	callback.Create().After("gorm:after_create").Register("loggable:create", p.addCreated)
	callback.Update().After("gorm:after_update").Register("loggable:update", p.addUpdated)
	callback.Delete().After("gorm:after_delete").Register("loggable:delete", p.addDeleted)
	return p, nil
}

func (p *Plugin) SetRemoteAdress(ip string) error{
	uuid, err := uuid.NewUUID()
	if err != nil{
		return err
	}

	requestUUID = uuid.String()
	remoteAddres = ip
	
	return err
}