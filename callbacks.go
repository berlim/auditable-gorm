package auditableGorm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"reflect"
	"strings"

	"gorm.io/gorm"
)

const (
	ACTION_CREATE = "create"
	ACTION_UPDATE = "update"
	ACTION_DELETE = "delete"
)

// Hook for after_create.
func (p *Plugin) addCreated(db *gorm.DB) {
	saveAudit(db, p.db, ACTION_CREATE, auditProps)
}

// Hook for after_delete.
func (p *Plugin) addDeleted(db *gorm.DB) {
	saveAudit(db, p.db, ACTION_DELETE, auditProps)
}

// Hook for after_update.
func (p *Plugin) addUpdated(db *gorm.DB) {
	saveAudit(db, p.db, ACTION_UPDATE, func(db *gorm.DB, id int64) bytes.Buffer {
		buff := bytes.Buffer{}

		original := map[string]interface{}{}
		// using db instead of p.db will generate "database lock" error
		p.db.Table(db.Statement.Table).Where("id = ?", id).Find(&original)

		if dest, err := getModelAsMap(db.Statement.Model); err == nil {
			for destK, destV := range dest {
				destK = strings.ToLower(destK)
				if originalV, ok := original[destK]; ok && originalV != destV {
					strDestVal := fmt.Sprintf("%v", destV)
					strOriginalVal := fmt.Sprintf("%v", originalV)
					if isZero(originalV) {
						buff.WriteString(
							fmt.Sprintf("\n%s:\n- %s", destK, strDestVal))
					} else {
						if strDestVal != strOriginalVal {
							buff.WriteString(
								fmt.Sprintf("\n%s:\n- %s\n- %s", destK, strOriginalVal, strDestVal))
						}
					}
				}
			}
		}

		return buff
	})
}

func isZero(val interface{}) bool {
	v := reflect.ValueOf(val)
	return !v.IsValid() || reflect.DeepEqual(v.Interface(), reflect.Zero(v.Type()).Interface())
}

func getModelAsMap(model interface{}) (out map[string]interface{}, err error) {
	b, err := json.Marshal(model)
	if err != nil {
		return
	}
	json.Unmarshal(b, &out)
	return
}

func saveAudit(db, pluginDb *gorm.DB, action string, fnChanges func(db *gorm.DB, id int64) bytes.Buffer) {
	auditTableName := getTableName()
	if db.Statement.Schema.Name == "Audits" {
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
			Audited_changes: fmt.Sprintf("---%s", buff.String())}
		// insert using pluginDb because using just db will attempt to inser
		// on users table. I don't know why
		if err := pluginDb.Table(auditTableName).Create(&audit).Error; err != nil {
			log.Printf("audits insert error - %v", err)
		}
	}
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
			buff.WriteString(fmt.Sprintf("\n%s: %v", field.Name, fieldValue))
		}
	}
	return
}
