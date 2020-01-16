package urlstruct_test

import (
	"reflect"

	"github.com/go-pg/urlstruct"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

type MyStruct struct {
	tableName struct{} `urlstruct:"myname"`
}

var _ = Describe("DescribeStruct", func() {
	It("handles tableName", func() {
		info := urlstruct.DescribeStruct(reflect.TypeOf((*MyStruct)(nil)))
		Expect(info.TableName).To(Equal("myname"))
	})
})
