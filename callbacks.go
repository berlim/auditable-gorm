package auditableGorm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	ACTION_CREATE = "create"
	ACTION_UPDATE = "update"
	ACTION_DELETE = "delete"
)

// Hook for after_create.
func (p *Plugin) addCreated(db *gorm.DB) {
	saveAudit(db, p.db, ACTION_CREATE)
}

// Hook for after_delete.
func (p *Plugin) addDeleted(db *gorm.DB) {
	saveAudit(db, p.db, ACTION_DELETE)
}

// Hook for after_update.
func (p *Plugin) addUpdated(db *gorm.DB) {
	if db.Statement.Schema.Table == "audits" {
		return
	}
	// var buff bytes.Buffer
	var id int64
	idValue, isZero := db.Statement.Schema.PrioritizedPrimaryField.ValueOf(db.Statement.ReflectValue)
	if !isZero {
		id = idValue.(int64)
	}
	fmt.Println(id)

	original := map[string]interface{}{}
	db.Find(&original, id)

	if dest, err := getModelAsMap(db.Statement.Model); err == nil {
		for destK, destV := range dest {
			// TODO: check if value is diff using reflect
			if originalV, ok := original[strings.ToLower(destK)]; ok && originalV != destV {
				fmt.Println(destV)
			}
		}
		fmt.Println(original)
		fmt.Println(dest)
		fmt.Println("out")
	}
}

func getModelAsMap(model interface{}) (out map[string]interface{}, err error) {
	b, err := json.Marshal(model)
	if err != nil {
		return
	}
	json.Unmarshal(b, &out)
	return
}

func saveAudit(db, pluginDb *gorm.DB, action string) {
	if db.Statement.Schema.Table == "audits" {
		return
	}
	var buff bytes.Buffer
	var id int64
	idValue, isZero := db.Statement.Schema.PrioritizedPrimaryField.ValueOf(db.Statement.ReflectValue)
	if !isZero {
		id = idValue.(int64)
	}
	for _, field := range db.Statement.Schema.Fields {
		fieldValue, isZero := field.ValueOf(db.Statement.ReflectValue)
		if !isZero {
			buff.WriteString(fmt.Sprintf("\n%s: %v", field.Name, fieldValue))
		}
	}
	// TODO: verificar se o model implementa a interface AuditableModel
	if buff.Len() > 0 {
		uuid, _ := uuid.NewUUID()

		audit := Audits{
			Auditable_id:    id,
			Action:          action,
			Auditable_type:  db.Statement.Schema.Name,
			Version:         int64(1),
			Request_uuid:    uuid.String(),
			Audited_changes: fmt.Sprintf("---%s", buff.String())}

		pluginDb.Table("audits").Create(&audit)
	}
}

// func newChangeLog(scope *gorm.Scope, action string) (*Audits, error) {
// 	var newVersion int64

// 	rawObject, err := json.Marshal(scope.Value)
// 	if err != nil {
// 		return nil, err
// 	}

// 	newVersion = 1
// 	auditable_id := scope.PrimaryKeyValue().(int64)
// 	auditable_type := scope.GetModelStruct().ModelType.Name()

// 	if action == "update" {
// 		//var lastVersion Audits
// 		//scope.DB().Table("paymentx.audits").Select("version").Where("auditable_id = ? and auditable_type = ?", auditable_id, auditable_type).Last(&lastVersion)
// 		newVersion = 1
// 	}

// 	return &Audits{
// 		Action:          action,
// 		Auditable_id:    auditable_id,
// 		Auditable_type:  auditable_type,
// 		Audited_changes: string(rawObject),
// 		Version:         newVersion,
// 		Remote_address:  remoteAddres,
// 		Request_uuid:    requestUUID,
// 	}, nil
// }

// Writes new change log row to db.
// func addRecord(scope *gorm.Scope, action string) error {
// 	cl, err := newChangeLog(scope, action)
// 	if err != nil {
// 		return nil
// 	}

// 	return scope.DB().Table("paymentx.audits").Create(cl).Error
// }

// func computeUpdateDiff(scope *gorm.Scope) UpdateDiff {
// 	old := im.get(scope.Value, scope.PrimaryKeyValue())
// 	if old == nil {
// 		return nil
// 	}

// 	ov := reflect.Indirect(reflect.ValueOf(old))
// 	nv := reflect.Indirect(reflect.ValueOf(scope.Value))
// 	names := getLoggableFieldNames(old)

// 	diff := make(UpdateDiff)

// 	havingChanges := false
// 	for _, name := range names {
// 		ofv := ov.FieldByName(name).Interface()
// 		nfv := nv.FieldByName(name).Interface()
// 		if ofv != nfv {
// 			diff[name] = fmt.Sprintf("old: %v | new: %v", ofv, nfv)
// 			havingChanges = true
// 		}
// 	}

// 	if !havingChanges {
// 		return nil
// 	}

// 	return diff
// }
