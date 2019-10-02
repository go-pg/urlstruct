package urlstruct

import (
	"net/url"
	"reflect"
	"strings"

	"github.com/vmihailenco/tagparser"
)

var unmarshalerType = reflect.TypeOf((*Unmarshaler)(nil)).Elem()

type Unmarshaler interface {
	UnmarshalValues(url.Values) error
}

type unmarshalerField struct {
	Index []int
}

type StructInfo struct {
	TableName string
	Fields    []*Field

	isUnmarshaler bool
	unmarshalers  []*unmarshalerField
}

func newStructInfo(typ reflect.Type) *StructInfo {
	meta := &StructInfo{
		Fields: make([]*Field, 0, typ.NumField()),
	}
	addFields(meta, typ, nil)
	return meta
}

func (s *StructInfo) decode(strct reflect.Value, name string, values []string) error {
	name = strings.TrimPrefix(name, ":")
	name = strings.TrimSuffix(name, "[]")

	field := s.Field(name)
	if field == nil || field.noDecode {
		return nil
	}
	return field.scanValue(field.Value(strct), values)
}

func (s *StructInfo) Field(name string) *Field {
	col, op := splitColumnOperator(name, "__")
	for _, f := range s.Fields {
		if f.Column == col && f.Op == op {
			return f
		}
	}
	return nil
}

func addFields(meta *StructInfo, typ reflect.Type, baseIndex []int) {
	if baseIndex != nil {
		baseIndex = baseIndex[:len(baseIndex):len(baseIndex)]
	}

	meta.isUnmarshaler = isUnmarshaler(typ)

	for i := 0; i < typ.NumField(); i++ {
		sf := typ.Field(i)
		if sf.Anonymous {
			tag := sf.Tag.Get("urlstruct")
			if tag == "-" {
				continue
			}

			sfType := sf.Type
			if sfType.Kind() == reflect.Ptr {
				sfType = sfType.Elem()
			}
			if sfType.Kind() != reflect.Struct {
				continue
			}

			addFields(meta, sfType, sf.Index)

			if isUnmarshaler(reflect.PtrTo(sfType)) {
				meta.unmarshalers = append(meta.unmarshalers, &unmarshalerField{
					Index: append(baseIndex, sf.Index...),
				})
			}

			continue
		}

		if sf.Name == "tableName" {
			tag := tagparser.Parse(sf.Tag.Get("urlstruct"))
			name, _ := tagparser.Unquote(tag.Name)
			meta.TableName = name
			continue
		}

		f := newField(meta, sf)
		if f == nil {
			continue
		}
		if len(baseIndex) > 0 {
			f.Index = append(baseIndex, f.Index...)
		}
		meta.Fields = append(meta.Fields, f)
	}
}

var (
	urlValuesType = reflect.TypeOf((*url.Values)(nil)).Elem()
	errorType     = reflect.TypeOf((*error)(nil)).Elem()
)

func isUnmarshaler(typ reflect.Type) bool {
	for i := 0; i < typ.NumMethod(); i++ {
		meth := typ.Method(i)
		if meth.Name == "UnmarshalValues" &&
			meth.Type.NumIn() == 2 &&
			meth.Type.NumOut() == 1 &&
			meth.Type.In(1) == urlValuesType &&
			meth.Type.Out(0) == errorType {
			return true
		}
	}
	return false
}
