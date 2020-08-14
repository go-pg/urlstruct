package urlstruct

import (
	"reflect"

	"github.com/vmihailenco/tagparser"
)

type Field struct {
	Type  reflect.Type
	Name  string
	Index []int
	Tag   *tagparser.Tag

	noDecode  bool
	scanValue scannerFunc
}

func (f *Field) init() {
	_, f.noDecode = f.Tag.Options["nodecode"]

	if f.Type.Kind() == reflect.Slice {
		f.scanValue = sliceScanner(f.Type)
	} else {
		f.scanValue = scanner(f.Type)
	}
}

func (f *Field) Value(strct reflect.Value) reflect.Value {
	return strct.FieldByIndex(f.Index)
}
