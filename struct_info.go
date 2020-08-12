package urlstruct

import (
	"context"
	"net/url"
	"reflect"

	"github.com/codemodus/kace"
	"github.com/vmihailenco/tagparser"
)

type Unmarshaler interface {
	UnmarshalValues(ctx context.Context, values url.Values) error
}

type ParamUnmarshaler interface {
	UnmarshalParam(ctx context.Context, name string, values []string) error
}

//------------------------------------------------------------------------------

type StructInfo struct {
	fields   []*Field
	fieldMap map[string]*Field

	structs map[string][]int

	isUnmarshaler      bool
	isParamUnmarshaler bool
	unmarshalerIndexes [][]int
}

func newStructInfo(typ reflect.Type) *StructInfo {
	sinfo := &StructInfo{
		fields:   make([]*Field, 0, typ.NumField()),
		fieldMap: make(map[string]*Field),

		isUnmarshaler:      isUnmarshaler(reflect.PtrTo(typ)),
		isParamUnmarshaler: isParamUnmarshaler(reflect.PtrTo(typ)),
	}
	addFields(sinfo, typ, nil)
	return sinfo
}

func (s *StructInfo) Field(name string) *Field {
	return s.fieldMap[name]
}

func addFields(sinfo *StructInfo, typ reflect.Type, baseIndex []int) {
	for i := 0; i < typ.NumField(); i++ {
		sf := typ.Field(i)
		if sf.PkgPath != "" && !sf.Anonymous {
			continue
		}

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

			if isUnmarshaler(reflect.PtrTo(sfType)) {
				index := joinIndex(baseIndex, sf.Index)
				sinfo.unmarshalerIndexes = append(sinfo.unmarshalerIndexes, index)
			}

			addFields(sinfo, sfType, joinIndex(baseIndex, sf.Index))
		} else {
			addField(sinfo, sf, baseIndex)
		}
	}
}

func addField(sinfo *StructInfo, sf reflect.StructField, baseIndex []int) {
	tag := tagparser.Parse(sf.Tag.Get("urlstruct"))
	if tag.Name == "-" {
		return
	}

	name := tag.Name
	if name == "" {
		name = sf.Name
	}
	index := joinIndex(baseIndex, sf.Index)

	if sf.Type.Kind() == reflect.Struct {
		if sinfo.structs == nil {
			sinfo.structs = make(map[string][]int)
		}
		sinfo.structs[kace.Snake(name)] = append(baseIndex, sf.Index...)
	}

	if isUnmarshaler(reflect.PtrTo(sf.Type)) {
		sinfo.unmarshalerIndexes = append(sinfo.unmarshalerIndexes, index)
	}

	f := &Field{
		Type:  sf.Type,
		Name:  kace.Snake(name),
		Index: index,
		Tag:   tag,
	}
	f.init()

	if f.scanValue != nil {
		sinfo.fields = append(sinfo.fields, f)
		sinfo.fieldMap[f.Name] = f
	}
}

func joinIndex(base, idx []int) []int {
	if len(base) == 0 {
		return idx
	}
	r := make([]int, 0, len(base)+len(idx))
	r = append(r, base...)
	r = append(r, idx...)
	return r
}

//------------------------------------------------------------------------------

var (
	contextType   = reflect.TypeOf((*context.Context)(nil)).Elem()
	urlValuesType = reflect.TypeOf((*url.Values)(nil)).Elem()
	errorType     = reflect.TypeOf((*error)(nil)).Elem()
)

func isUnmarshaler(typ reflect.Type) bool {
	for i := 0; i < typ.NumMethod(); i++ {
		meth := typ.Method(i)
		if meth.Name == "UnmarshalValues" &&
			meth.Type.NumIn() == 3 &&
			meth.Type.NumOut() == 1 &&
			meth.Type.In(1) == contextType &&
			meth.Type.In(2) == urlValuesType &&
			meth.Type.Out(0) == errorType {
			return true
		}
	}
	return false
}

var (
	stringType      = reflect.TypeOf("")
	stringSliceType = reflect.TypeOf((*[]string)(nil)).Elem()
)

func isParamUnmarshaler(typ reflect.Type) bool {
	for i := 0; i < typ.NumMethod(); i++ {
		meth := typ.Method(i)
		if meth.Name == "UnmarshalParam" &&
			meth.Type.NumIn() == 4 &&
			meth.Type.NumOut() == 1 &&
			meth.Type.In(1) == contextType &&
			meth.Type.In(2) == stringType &&
			meth.Type.In(3) == stringSliceType &&
			meth.Type.Out(0) == errorType {
			return true
		}
	}
	return false
}
