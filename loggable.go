package auditableGorm

import (
	"fmt"
)

// Interface is used to get metadata from your models.
type Interface interface {
	// Meta should return structure, that can be converted to json.
	Meta() interface{}
	// lock makes available only embedding structures.
	lock()
	// check if callback enabled
	isEnabled() bool
	// enable/disable loggable
	Enable(v bool)
}

// LoggableModel is a root structure, which implement Interface.
// Embed LoggableModel to your model so that Plugin starts tracking changes.
type LoggableModel struct {
	Disabled bool `sql:"-" json:"-"`
}

func (LoggableModel) Meta() interface{} { return nil }
func (LoggableModel) lock()             {}
func (l LoggableModel) isEnabled() bool { return !l.Disabled }
func (l LoggableModel) Enable(v bool)   { l.Disabled = !v }

// ChangeLog is a main entity, which used to log changes.
// Commonly, ChangeLog is stored in 'change_logs' table.

type Audits struct {
	Id					int64  `gorm:"auto:id"`
	Auditable_id		int64  `gorm:"column:auditable_id"`
	Auditable_type		string  `gorm:"column:auditable_type"`
	User_id				int64  `gorm:"column:user_id"`
	User_name			string  `gorm:"column:username"`
	Action				string  `gorm:"column:action"`
	Audited_changes		string  `gorm:"column:audited_changes"`
	Version				int64  `gorm:"column:version"`
	Remote_address		string  `gorm:"column:remote_address"`
}





func interfaceToString(v interface{}) string {
	switch val := v.(type) {
	case string:
		return val
	default:
		return fmt.Sprint(v)
	}
}

func isLoggable(value interface{}) bool {
	_, ok := value.(Interface)
	return ok
}

func isEnabled(value interface{}) bool {
	v, ok := value.(Interface)
	return ok && v.isEnabled()
}
