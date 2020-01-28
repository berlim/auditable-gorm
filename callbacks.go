package auditableGorm

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/jinzhu/gorm"
)

var im = newIdentityManager()

const (
	actionCreate = "create"
	actionUpdate = "update"
	actionDelete = "delete"
)

type UpdateDiff map[string]interface{}

// Hook for after_query.
func (p *Plugin) trackEntity(scope *gorm.Scope) {
	if !isLoggable(scope.Value) || !isEnabled(scope.Value) {
		return
	}

	v := reflect.Indirect(reflect.ValueOf(scope.Value))

	pkName := scope.PrimaryField().Name
	if v.Kind() == reflect.Slice {
		for i := 0; i < v.Len(); i++ {
			sv := reflect.Indirect(v.Index(i))
			el := sv.Interface()
			if !isLoggable(el) {
				continue
			}

			im.save(el, sv.FieldByName(pkName))
		}
		return
	}

	m := v.Interface()
	if !isLoggable(m) {
		return
	}

	im.save(v.Interface(), scope.PrimaryKeyValue())
}

// Hook for after_create.
func (p *Plugin) addCreated(scope *gorm.Scope) {

	v := reflect.Indirect(reflect.ValueOf(scope.Value))
	m := v.Interface()

	loggable := isLoggable(m)
	enable := isEnabled(m)
	if loggable && enable {
		_ = addRecord(scope, actionCreate)
	}
}

// Hook for after_update.
func (p *Plugin) addUpdated(scope *gorm.Scope) {
	loggable := isLoggable(scope.Value)
	enable := isEnabled(scope.Value)
	if (!loggable || !enable) {
		return
	}

	_ = addUpdateRecord(scope, p.opts)
}

// Hook for after_delete.
func (p *Plugin) addDeleted(scope *gorm.Scope) {
	if isLoggable(scope.Value) && isEnabled(scope.Value) {
		_ = addRecord(scope, actionDelete)
	}
}

func addUpdateRecord(scope *gorm.Scope, opts options) error {
	cl, err := newChangeLog(scope, actionUpdate)
	if err != nil {
		return err
	}


		diff := computeUpdateDiff(scope)

		if diff != nil {
			formatedDiff := FormatDiff(diff)

			cl.Audited_changes = formatedDiff

			scope.DB().Create(cl)
		}




	return nil
}

func newChangeLog(scope *gorm.Scope, action string) (*Audits, error) {
	var newVersion int64

	rawObject, err := json.Marshal(scope.Value)
	if err != nil {
		return nil, err
	}

	newVersion = 1
	auditable_id := scope.PrimaryKeyValue().(int64)
	auditable_type := scope.GetModelStruct().ModelType.Name()

	if (action == "update"){
		var lastVersion Audits
		scope.DB().Table("audits").Select("version").Where("auditable_id = ? and auditable_type = ?", auditable_id, auditable_type).Last(&lastVersion)
		newVersion = lastVersion.Version + 1
	}



	return &Audits{
		Action:     		action,
		Auditable_id:  		auditable_id,
		Auditable_type: 	auditable_type,
		Audited_changes:    string(rawObject),
		Version:			newVersion,
	}, nil
}

// Writes new change log row to db.
func addRecord(scope *gorm.Scope, action string) error {
	cl, err := newChangeLog(scope, action)
	if err != nil {
		return nil
	}

	return scope.DB().Create(cl).Error
}

func computeUpdateDiff(scope *gorm.Scope) UpdateDiff {
	old := im.get(scope.Value, scope.PrimaryKeyValue())
	if old == nil {
		return nil
	}

	ov := reflect.Indirect(reflect.ValueOf(old))
	nv := reflect.Indirect(reflect.ValueOf(scope.Value))
	names := getLoggableFieldNames(old)

	diff := make(UpdateDiff)

	havingChanges := false
	for _, name := range names {
		ofv := ov.FieldByName(name).Interface()
		nfv := nv.FieldByName(name).Interface()
		if ofv != nfv {
			diff[name] = fmt.Sprintf("old: %v | new: %v", ofv, nfv)
			havingChanges = true
		}
	}

	if !havingChanges {
		return nil
	}

	return diff
}
