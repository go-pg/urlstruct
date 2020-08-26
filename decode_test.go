package urlstruct_test

import (
	"context"
	"database/sql"
	"encoding"
	"net/url"
	"testing"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/go-pg/urlstruct"
	"github.com/google/uuid"
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

func (f *SubFilter) UnmarshalValues(ctx context.Context, values url.Values) error {
	f.Count++
	return nil
}

//------------------------------------------------------------------------------

type StructMap struct {
	Foo        string
	Bar        string
	UnknownMap map[string][]string `urlstruct:",unknown"`
}

var _ urlstruct.ParamUnmarshaler = (*StructMap)(nil)

func (s *StructMap) UnmarshalParam(ctx context.Context, name string, values []string) error {
	if s.UnknownMap == nil {
		s.UnknownMap = make(map[string][]string)
	}
	s.UnknownMap[name] = values
	return nil
}

//------------------------------------------------------------------------------

type Filter struct {
	unexported string //nolint:unused,structcheck

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

	Uuid []uuid.UUID
}

var _ urlstruct.Unmarshaler = (*Filter)(nil)

func (f *Filter) UnmarshalValues(ctx context.Context, values url.Values) error {
	f.Count++
	return nil
}

var _ = Describe("Decode", func() {
	ctx := context.TODO()

	It("decodes struct from Values", func() {
		f := new(Filter)
		err := urlstruct.Unmarshal(ctx, url.Values{
			"unexported": {"test"},

			"s_map[foo]":   {"foo_value"},
			"s_map[bar]":   {"bar_value"},
			"s_map[hello]": {"world"},

			"field":     {"one"},
			"field_neq": {"two"},
			"field_lt":  {"1"},
			"field_lte": {"2"},
			"field_gt":  {"3"},
			"field_gte": {"4"},

			"multi":     {"one", "two"},
			"multi_neq": {"3", "4"},

			"time":           {"1970-01-01T00:00:00Z"},
			"start_time_gte": {"1970-01-01T00:00:00Z"},

			"null_bool":    {"t"},
			"null_int64":   {"1234"},
			"null_float64": {"1.234"},
			"null_string":  {"string"},

			"map[foo]":   {`bar`},
			"map[hello]": {`world`},
			"map[]":      {"invalid"},
			"map][":      {"invalid"},

			"custom": {"custom"},

			"uuid": {"3fa85f64-5717-4562-b3fc-2c963f66afa6"},
		}, f)
		Expect(err).NotTo(HaveOccurred())

		Expect(f).To(Equal(&Filter{
			SubFilter: SubFilter{Count: 1},
			Sub:       SubFilter{Count: 1},
			Count:     1,

			SMap: StructMap{
				Foo:        "foo_value",
				Bar:        "bar_value",
				UnknownMap: map[string][]string{"hello": {"world"}},
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
			Uuid:   []uuid.UUID{uuid.MustParse("3fa85f64-5717-4562-b3fc-2c963f66afa6")},
		}))
	})

	It("supports names with suffix `[]`", func() {
		f := new(Filter)
		err := urlstruct.Unmarshal(ctx, url.Values{
			"field[]": {"one"},
		}, f)
		Expect(err).NotTo(HaveOccurred())
		Expect(f.Field).To(Equal("one"))
	})

	It("supports names with prefix `:`", func() {
		f := new(Filter)
		err := urlstruct.Unmarshal(ctx, url.Values{
			":field": {"one"},
		}, f)
		Expect(err).NotTo(HaveOccurred())
		Expect(f.Field).To(Equal("one"))
	})

	It("decodes sql.Null*", func() {
		f := new(Filter)
		err := urlstruct.Unmarshal(ctx, url.Values{
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
		err := urlstruct.Unmarshal(ctx, url.Values{}, f)
		Expect(err).NotTo(HaveOccurred())
		Expect(f.Count).To(Equal(1))
		Expect(f.SubFilter.Count).To(Equal(1))
		Expect(f.Sub.Count).To(Equal(1))
	})
})
