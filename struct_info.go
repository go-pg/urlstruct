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
	TableName    string
	Fields       []*Field
	unmarshalers []*unmarshalerField
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

			if reflect.PtrTo(sfType).Implements(unmarshalerType) {
				var idx []int
				idx = append(idx, baseIndex...)
				idx = append(idx, sf.Index...)
				meta.unmarshalers = append(meta.unmarshalers, &unmarshalerField{
					Index: idx,
				})
			} else {
				addFields(meta, sfType, sf.Index)
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
			var idx []int
			idx = append(idx, baseIndex...)
			idx = append(idx, f.Index...)
			f.Index = idx
		}
		meta.Fields = append(meta.Fields, f)
	}
}
