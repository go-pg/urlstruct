package urlstruct

import (
	"context"
	"net/url"
	"reflect"
)

// Unmarshal unmarshals the URL query values into the struct.
func Unmarshal(ctx context.Context, values url.Values, strct interface{}) error {
	d := newStructDecoder(reflect.ValueOf(strct))
	return d.Decode(ctx, values)
}
