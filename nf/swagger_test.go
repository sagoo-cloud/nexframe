package nf

import (
	"github.com/go-openapi/spec"
	"github.com/sagoo-cloud/nexframe/utils/meta"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

// PageInput 是一个通用的分页结构
type PageInput struct {
	Page     int `json:"page" description:"页码"`
	PageSize int `json:"pageSize" description:"每页数量"`
}

// ListReq 是一个包含嵌套 PageInput 的请求结构
type ListReq struct {
	meta.Meta `path:"/list" method:"GET" summary:"获取站点列表" tags:"站点管理"`
	Name      string `json:"name" description:"站点名称"`
	Status    *int32 `json:"status" description:"发布状态: 0=未发布，1=已发布"`
	PageInput
}

func TestGenerateParameters(t *testing.T) {
	f := NewAPIFramework()

	reqType := reflect.TypeOf(&ListReq{})

	params := f.generateParameters(reqType)

	// 验证生成的参数
	assert.Len(t, params, 4, "应该生成 4 个参数")

	// 创建一个 map 来存储参数，便于后续断言
	paramMap := make(map[string]spec.Parameter)
	for _, param := range params {
		paramMap[param.Name] = param
	}

	// 验证 name 参数
	assert.Contains(t, paramMap, "name")
	assert.Equal(t, "string", paramMap["name"].Type)
	assert.Equal(t, "站点名称", paramMap["name"].Description)

	// 验证 status 参数
	assert.Contains(t, paramMap, "status")
	assert.Equal(t, "integer", paramMap["status"].Type)
	assert.Equal(t, "发布状态: 0=未发布，1=已发布", paramMap["status"].Description)
	assert.True(t, paramMap["status"].Extensions["x-nullable"].(bool))

	// 验证 page 参数
	assert.Contains(t, paramMap, "page")
	assert.Equal(t, "integer", paramMap["page"].Type)
	assert.Equal(t, "页码", paramMap["page"].Description)

	// 验证 pageSize 参数
	assert.Contains(t, paramMap, "pageSize")
	assert.Equal(t, "integer", paramMap["pageSize"].Type)
	assert.Equal(t, "每页数量", paramMap["pageSize"].Description)
}
