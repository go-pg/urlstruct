package urlstruct

import (
	"net/url"
	"reflect"
	"strings"
)

type structDecoder struct {
	v             reflect.Value
	sinfo         *StructInfo
	unknownFields *fieldMap
}

func (d *structDecoder) Decode(values url.Values) error {
	var maps map[string][]string
	for name, values := range values {
		if name, key, ok := mapKey(name); ok {
			if maps == nil {
				maps = make(map[string][]string)
			}
			maps[name] = append(maps[name], key, values[0])
			continue
		}

		if err := d.DecodeField(name, values); err != nil {
			return err
		}
	}

	for name, values := range maps {
		if err := d.DecodeField(name, values); err != nil {
			return nil
		}
	}

	for _, f := range d.sinfo.unmarshalers {
		fv := d.v.FieldByIndex(f.Index)
		if fv.Kind() == reflect.Struct {
			fv = fv.Addr()
		} else if fv.IsNil() {
			fv.Set(reflect.New(fv.Type().Elem()))
		}

		u := fv.Interface().(Unmarshaler)
		if err := u.UnmarshalValues(values); err != nil {
			return err
		}
	}

	if d.sinfo.isUnmarshaler {
		return d.v.Addr().Interface().(Unmarshaler).UnmarshalValues(values)
	}

	return nil
}

func (d *structDecoder) DecodeField(name string, values []string) error {
	name = strings.TrimPrefix(name, ":")
	name = strings.TrimSuffix(name, "[]")

	if field := d.sinfo.Field(name); field != nil && !field.noDecode {
		return field.scanValue(field.Value(d.v), values)
	}

	if d.sinfo.unknownFieldsIndex == nil {
		return nil
	}

	if d.unknownFields == nil {
		d.unknownFields = newFieldMap(d.v.FieldByIndex(d.sinfo.unknownFieldsIndex))
	}
	return d.unknownFields.Decode(name, values)
}
