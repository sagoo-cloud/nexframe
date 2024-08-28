package convert

import (
	"context"
	"encoding/json"
	"github.com/sagoo-cloud/nexframe/encoding/gbinary"
	"github.com/sagoo-cloud/nexframe/os/zlog/intlog"
	"github.com/sagoo-cloud/nexframe/utils/reflection"
	"math"
	"reflect"
)

// Bytes converts `any` to []byte.
func Bytes(any interface{}) []byte {
	if any == nil {
		return nil
	}
	switch value := any.(type) {
	case string:
		buf := make([]byte, len(value))
		copy(buf, value)
		return buf
	case []byte:
		return value

	default:
		if f, ok := value.(iBytes); ok {
			return f.Bytes()
		}
		originValueAndKind := reflection.OriginValueAndKind(any)
		switch originValueAndKind.OriginKind {
		case reflect.Map:
			bytes, err := json.Marshal(any)
			if err != nil {
				intlog.Errorf(context.TODO(), `%+v`, err)
			}
			return bytes

		case reflect.Array, reflect.Slice:
			var (
				ok    = true
				bytes = make([]byte, originValueAndKind.OriginValue.Len())
			)
			for i := range bytes {
				int32Value := Int32(originValueAndKind.OriginValue.Index(i).Interface())
				if int32Value < 0 || int32Value > math.MaxUint8 {
					ok = false
					break
				}
				bytes[i] = byte(int32Value)
			}
			if ok {
				return bytes
			}
		default:
		}
		return gbinary.Encode(any)
	}
}
