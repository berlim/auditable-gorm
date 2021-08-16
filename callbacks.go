package auditableGorm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"reflect"
	"strings"
	"time"

	"gorm.io/gorm"
)

const (
	ACTION_CREATE = "create"
	ACTION_UPDATE = "update"
	ACTION_DELETE = "delete"
)

// Hook for after_create.
func (p *Plugin) addCreated(db *gorm.DB) {
	saveAudit(db, ACTION_CREATE, auditProps)
}

// Hook for after_delete.
func (p *Plugin) addDeleted(db *gorm.DB) {
	saveAudit(db, ACTION_DELETE, auditProps)
}

// Hook for after_update.
func (p *Plugin) addUpdated(db *gorm.DB) {
	saveAudit(db, ACTION_UPDATE, func(db *gorm.DB, id int64) bytes.Buffer {
		buff := bytes.Buffer{}

		original := map[string]interface{}{}
		// using db instead of p.db will generate "database lock" error
		db.Table(mountUpdateTableName(db)).Where("id = ?", id).Find(&original)

		if dest, err := getModelAsMap(db.Statement.Model); err == nil {
			for destK, destV := range dest {
				destK = strings.ToLower(destK)
				if originalV, ok := original[destK]; ok && originalV != destV {
					if !reflect.DeepEqual(destV, destK) {
						buff.WriteString(
							fmt.Sprintf("\n%s:\n- %v\n- %v", destK, originalV, destV))
					}
				}
			}
		}

		return buff
	})
}

func getModelAsMap(model interface{}) (out map[string]interface{}, err error) {
	b, err := json.Marshal(model)
	if err != nil {
		return
	}
	json.Unmarshal(b, &out)
	return
}

func saveAudit(db *gorm.DB, action string, fnChanges func(db *gorm.DB, id int64) bytes.Buffer) {
	if db.Statement.Schema.Name == "Audits" || checkAuditName(db) {
		return
	}
	var id int64
	idValue, isZero := db.Statement.Schema.PrioritizedPrimaryField.ValueOf(db.Statement.ReflectValue)
	if !isZero {
		id = idValue.(int64)
	}
	buff := fnChanges(db, id)
	if buff.Len() > 0 {
		auditData := getAuditData(db)
		audit := Audits{
			Auditable_id:    id,
			Action:          action,
			Auditable_type:  db.Statement.Schema.Name,
			Version:         int64(1),
			Request_uuid:    auditData.UUID,
			Remote_address:  auditData.Address,
			Audited_changes: fmt.Sprintf("---%s", buff.String()),
			Created_at:      time.Now()}
		db.Transaction(func(tx *gorm.DB) error {
			return tx.Model(&Audits{}).Create(&audit).Error
		})

		auditErr := db.Exec(fmt.Sprintf(`
			INSERT INTO
				%s
				(
					auditable_id,
					"action",
					auditable_type,
					"version",
					request_uuid,
					remote_address,
					audited_changes,
					created_at)
			VALUES
				(%v, %q, %q, %v, %q, %q, "%s", %q);
		`, getAuditTableName(),
			audit.Auditable_id,
			audit.Action,
			audit.Auditable_type,
			audit.Version,
			audit.Request_uuid,
			audit.Remote_address,
			audit.Audited_changes,
			audit.Created_at.String())).Error
		if auditErr != nil {
			log.Printf("audits insert error - %v", auditErr)
		}
	}
}

func checkAuditName(db *gorm.DB) bool {
	auditTable := getAuditTableName()
	if names := strings.Split(auditTable, "."); len(names) == 2 {
		auditTable = names[1]
	}
	return auditTable == db.Statement.Table
}

func mountUpdateTableName(db *gorm.DB) string {
	return db.Config.NamingStrategy.TableName(db.Statement.Table)
}

func getAuditData(db *gorm.DB) AuditData {
	data, ok := db.Statement.Context.Value(AUDIT_DATA_CTX_KEY).(AuditData)
	if ok {
		return data
	}
	return AuditData{}
}

func auditProps(db *gorm.DB, id int64) (buff bytes.Buffer) {
	for _, field := range db.Statement.Schema.Fields {
		fieldValue, isZero := field.ValueOf(db.Statement.ReflectValue)
		if !isZero {
			buff.WriteString(fmt.Sprintf("\n%s: %v", field.DBName, fieldValue))
		}
	}
	return
}
