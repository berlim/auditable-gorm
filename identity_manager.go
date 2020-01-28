package auditableGorm

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"reflect"
)

type identityMap map[string]interface{}

// identityManager is used as cache.
type identityManager struct {
	m identityMap
}

func newIdentityManager() *identityManager {
	return &identityManager{
		m: make(identityMap),
	}
}

func (im *identityManager) save(value, pk interface{}) {
	d := reflect.TypeOf(value)

	v := reflect.ValueOf(d)
	i := reflect.Indirect(v)

	t := i.Type()

	name := t.Name()
	qtd := t.NumField()

	fmt.Printf("Type: %v Qtd: %v", name, qtd)

	newValue := value

	key := genIdentityKey(t, pk)
	im.m[key] = newValue

	value = nil
}

func (im identityManager) get(value, pk interface{}) interface{} {
	d := reflect.TypeOf(value)
	v := reflect.ValueOf(d)
	i := reflect.Indirect(v)
	t := i.Type()

	key := genIdentityKey(t, pk)
	m, ok := im.m[key]
	if !ok {
		return nil
	}


	return m
}

func genIdentityKey(t reflect.Type, pk interface{}) string {
	key := fmt.Sprintf("%v_%s", pk, t.Name())
	b := md5.Sum([]byte(key))

	return hex.EncodeToString(b[:])
}
