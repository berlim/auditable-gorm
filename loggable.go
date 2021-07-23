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
	ID              int64  `gorm:"auto:id"`
	Auditable_id    int64  `gorm:"column:auditable_id"`
	Auditable_type  string `gorm:"column:auditable_type"`
	User_id         int64  `gorm:"column:user_id"`
	User_name       string `gorm:"column:username"`
	Action          string `gorm:"column:action"`
	Audited_changes string `gorm:"column:audited_changes"`
	Version         int64  `gorm:"column:version"`
	Remote_address  string `gorm:"column:remote_address"`
	Request_uuid    string `gorm:"column:request_uuid"`
}

const AUDIT_DATA_CTX_KEY = "AUDIT_PLUGIN_DATA"

type AuditData struct {
	UUID    string
	Address string
}

func (Audits) TableName() string {
	return "audits"
}
