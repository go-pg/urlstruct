package urlstruct

import (
	"fmt"
	"net/url"
	"reflect"
	"strings"
	"sync"
)

var globalMap structInfoMap

func DescribeStruct(typ reflect.Type) *StructInfo {
	return globalMap.DescribeStruct(typ)
}

func Unmarshal(values url.Values, strct interface{}) error {
	return globalMap.Unmarshal(values, strct)
}

type structInfoMap struct {
	m sync.Map
}

func (m *structInfoMap) DescribeStruct(typ reflect.Type) *StructInfo {
	if typ.Kind() != reflect.Struct {
		panic(fmt.Errorf("got %s, wanted %s", typ.Kind(), reflect.Struct))
	}

	if v, ok := m.m.Load(typ); ok {
		return v.(*StructInfo)
	}

	sinfo := newStructInfo(typ)
	if v, loaded := m.m.LoadOrStore(typ, sinfo); loaded {
		return v.(*StructInfo)
	}
	return sinfo
}

// Unmarshal unmarshals url values into the struct.
func (m *structInfoMap) Unmarshal(values url.Values, strct interface{}) error {
	v := reflect.Indirect(reflect.ValueOf(strct))
	d := &structDecoder{
		v:     v,
		sinfo: m.DescribeStruct(v.Type()),
	}
	return d.Decode(values)
}

func mapKey(s string) (name string, key string, ok bool) {
	ind := strings.IndexByte(s, '[')
	if ind == -1 || s[len(s)-1] != ']' {
		return "", "", false
	}
	key = s[ind+1 : len(s)-1]
	if key == "" {
		return "", "", false
	}
	name = s[:ind]
	return name, key, true
}
