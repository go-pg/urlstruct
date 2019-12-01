package urlstruct_test

import (
	"database/sql"
	"encoding"
	"net/url"
	"testing"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/go-pg/urlstruct"
)

func TestGinkgo(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "urlstruct")
}

//------------------------------------------------------------------------------

type CustomField struct {
	S string
}

var _ encoding.TextUnmarshaler = (*CustomField)(nil)

func (f *CustomField) UnmarshalText(text []byte) error {
	f.S = string(text)
	return nil
}

//------------------------------------------------------------------------------

type SubFilter struct {
	Count int
}

var _ urlstruct.Unmarshaler = (*SubFilter)(nil)

func (f *SubFilter) UnmarshalValues(values url.Values) error {
	f.Count++
	return nil
}

//------------------------------------------------------------------------------

type StructMap struct {
	Foo   string
	Bar   string
	Extra map[string][]string `urlstruct:",unknown"`
}

//------------------------------------------------------------------------------

type Filter struct {
	SubFilter
	Sub   SubFilter
	SMap  StructMap
	Count int

	Field    string
	FieldNEQ string
	FieldLT  int8
	FieldLTE int16
	FieldGT  int32
	FieldGTE int64

	Multi    []string
	MultiNEQ []int

	Time         time.Time
	StartTimeGTE time.Time

	NullBool    sql.NullBool
	NullInt64   sql.NullInt64
	NullFloat64 sql.NullFloat64
	NullString  sql.NullString

	Map    map[string]string
	Custom CustomField

	Omit []byte `pg:"-"`

	Unknown map[string][]string `urlstruct:",unknown"`
}

var _ urlstruct.Unmarshaler = (*Filter)(nil)

func (f *Filter) UnmarshalValues(values url.Values) error {
	f.Count++
	return nil
}

var _ = Describe("Decode", func() {
	It("decodes struct from Values", func() {
		f := new(Filter)
		err := urlstruct.Unmarshal(url.Values{
			"s_map[foo]":   {"foo_value"},
			"s_map[bar]":   {"bar_value"},
			"s_map[hello]": {"world"},

			"field":      {"one"},
			"field__neq": {"two"},
			"field__lt":  {"1"},
			"field__lte": {"2"},
			"field__gt":  {"3"},
			"field__gte": {"4"},

			"multi":      {"one", "two"},
			"multi__neq": {"3", "4"},

			"time":            {"1970-01-01T00:00:00Z"},
			"start_time__gte": {"1970-01-01T00:00:00Z"},

			"null_bool":    {"t"},
			"null_int64":   {"1234"},
			"null_float64": {"1.234"},
			"null_string":  {"string"},

			"map[foo]":   {`bar`},
			"map[hello]": {`world`},
			"map[]":      {"invalid"},
			"map][":      {"invalid"},

			"custom": {"custom"},
		}, f)
		Expect(err).NotTo(HaveOccurred())

		Expect(f).To(Equal(&Filter{
			SubFilter: SubFilter{Count: 1},
			Sub:       SubFilter{Count: 1},
			Count:     1,

			SMap: StructMap{
				Foo:   "foo_value",
				Bar:   "bar_value",
				Extra: map[string][]string{"hello": {"world"}},
			},

			Field:    "one",
			FieldNEQ: "two",
			FieldLT:  1,
			FieldLTE: 2,
			FieldGT:  3,
			FieldGTE: 4,

			Multi:    []string{"one", "two"},
			MultiNEQ: []int{3, 4},

			Time:         time.Unix(0, 0).UTC(),
			StartTimeGTE: time.Unix(0, 0).UTC(),

			NullBool:    sql.NullBool{Bool: true, Valid: true},
			NullInt64:   sql.NullInt64{Int64: 1234, Valid: true},
			NullFloat64: sql.NullFloat64{Float64: 1.234, Valid: true},
			NullString:  sql.NullString{String: "string", Valid: true},

			Map:    map[string]string{"foo": "bar", "hello": "world"},
			Custom: CustomField{S: "custom"},
			Omit:   nil,

			Unknown: map[string][]string{"map][": []string{"invalid"}},
		}))
	})

	It("supports names with suffix `[]`", func() {
		f := new(Filter)
		err := urlstruct.Unmarshal(url.Values{
			"field[]": {"one"},
		}, f)
		Expect(err).NotTo(HaveOccurred())
		Expect(f.Field).To(Equal("one"))
	})

	It("supports names with prefix `:`", func() {
		f := new(Filter)
		err := urlstruct.Unmarshal(url.Values{
			":field": {"one"},
		}, f)
		Expect(err).NotTo(HaveOccurred())
		Expect(f.Field).To(Equal("one"))
	})

	It("decodes sql.Null*", func() {
		f := new(Filter)
		err := urlstruct.Unmarshal(url.Values{
			"null_bool":    {""},
			"null_int64":   {""},
			"null_float64": {""},
			"null_string":  {""},
		}, f)
		Expect(err).NotTo(HaveOccurred())

		Expect(f.NullBool.Valid).To(BeTrue())
		Expect(f.NullBool.Bool).To(BeZero())

		Expect(f.NullInt64.Valid).To(BeTrue())
		Expect(f.NullInt64.Int64).To(BeZero())

		Expect(f.NullFloat64.Valid).To(BeTrue())
		Expect(f.NullFloat64.Float64).To(BeZero())

		Expect(f.NullString.Valid).To(BeTrue())
		Expect(f.NullString.String).To(BeZero())
	})

	It("calls UnmarshalValues", func() {
		f := new(Filter)
		err := urlstruct.Unmarshal(url.Values{}, f)
		Expect(err).NotTo(HaveOccurred())
		Expect(f.Count).To(Equal(1))
		Expect(f.SubFilter.Count).To(Equal(1))
		Expect(f.Sub.Count).To(Equal(1))
	})
})
