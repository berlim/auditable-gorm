package auditableGorm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
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
		p.db.Table(mountUpdateTableName(db)).Where("id = ?", id).Find(&original)

		if dest, err := getModelAsMap(db.Statement.Model); err == nil {
			for destK, destV := range dest {
				destK = strings.ToLower(destK)
				if checkIgnoreKey(destK) {
					continue
				}
				if originalV, ok := original[destK]; ok {
					var destS, originalS string
					switch destV.(type) {
					case float32, float64:
						switch originalV.(type) {
						case float32, float64:
							destS = fmt.Sprintf("%v", destV)
							originalS = fmt.Sprintf("%v", originalV)
						default:
							destS = fmt.Sprintf("%.f", destV)
							originalS = fmt.Sprintf("%v", originalV)
						}
					default:
						destS = fmt.Sprintf("%v", destV)
						originalS = fmt.Sprintf("%v", originalV)
					}
					if originalS != destS {
						buff.WriteString(
							fmt.Sprintf("\n%s:\n- %v\n- %v", destK, originalS, destS))
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

func saveAudit(cbDB, db *gorm.DB, action string, fnChanges func(db *gorm.DB, id int64) bytes.Buffer) {
	if cbDB.Statement.Schema.Name == "Audits" || checkAuditName(cbDB) {
		return
	}
	var id int64
	idValue, isZero := cbDB.Statement.Schema.PrioritizedPrimaryField.ValueOf(cbDB.Statement.ReflectValue)
	if !isZero {
		id = idValue.(int64)
	}
	buff := fnChanges(cbDB, id)
	if buff.Len() > 0 {
		auditData := getAuditData(cbDB)
		audit := Audits{
			Auditable_id:    id,
			Action:          action,
			Auditable_type:  cbDB.Statement.Schema.Name,
			Version:         int64(1),
			Request_uuid:    auditData.UUID,
			Remote_address:  auditData.Address,
			Audited_changes: fmt.Sprintf("---%s", buff.String()),
			Created_at:      time.Now()}
		if err := db.Table(getAuditTableName()).Create(&audit).Error; err != nil {
			log.Printf("audits insert error - %v", err)
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
		if checkIgnoreKey(field.DBName) {
			continue
		}
		fieldValue, isZero := field.ValueOf(db.Statement.ReflectValue)
		if !isZero {
			buff.WriteString(fmt.Sprintf("\n%s: %v", field.DBName, fieldValue))
		}
	}
	return
}

var keysToIgnore = map[string]bool{
	"id":         true,
	"created_at": true,
	"updated_at": true,
	"password":   true,
}

func checkIgnoreKey(key string) bool {
	_, ok := keysToIgnore[key]
	return ok
}
