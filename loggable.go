package auditableGorm


type Interface interface {
	Meta() interface{}
	lock()
	isEnabled() bool
	Enable(v bool)
}


type LoggableModel struct {
	Disabled bool `sql:"-" json:"-"`
}

func (LoggableModel) Meta() interface{} { return nil }
func (LoggableModel) lock()             {}
func (l LoggableModel) isEnabled() bool { return !l.Disabled }
func (l LoggableModel) Enable(v bool)   { l.Disabled = !v }

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


func (Audits) TableName() string {
	return "audits"
}

func isLoggable(value interface{}) bool {
	_, ok := value.(Interface)
	return ok
}

func isEnabled(value interface{}) bool {
	v, ok := value.(Interface)
	return ok && v.isEnabled()
}
