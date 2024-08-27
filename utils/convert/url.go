package convert

import (
	"encoding/json"
	"net/url"
	"reflect"
	"strconv"
)

func FormEncode(params map[string]interface{}) url.Values {
	data := url.Values{}
	for k, param := range params {
		paramsType := reflect.TypeOf(param)
		switch paramsType.String() {
		case "string":
			data.Set(k, param.(string))
		case "int":
			data.Set(k, strconv.Itoa(param.(int)))
		default:
			str, _ := json.Marshal(param)
			data.Set(k, string(str))

		}
	}
	return data
}
