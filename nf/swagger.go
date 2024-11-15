package nf

import (
	"encoding/json"
	"github.com/go-openapi/spec"
	"github.com/sagoo-cloud/nexframe/g"
	"net/http"
	"reflect"
	"strings"
	"time"
)

// generateParameters 生成 Swagger 参数定义
func (f *APIFramework) generateParameters(reqType reflect.Type) []spec.Parameter {
	var params []spec.Parameter
	processedTypes := make(map[reflect.Type]bool)

	var generateParams func(t reflect.Type, prefix string)
	generateParams = func(t reflect.Type, prefix string) {
		if processedTypes[t] {
			return // 避免循环引用
		}
		processedTypes[t] = true

		t = deref(t)
		for i := 0; i < t.NumField(); i++ {
			field := t.Field(i)

			// 跳过 g.Meta 字段
			if field.Anonymous && field.Type == reflect.TypeOf(g.Meta{}) {
				continue
			}

			jsonTag := field.Tag.Get("json")
			if jsonTag == "" {
				jsonTag = strings.ToLower(field.Name)
			}
			jsonTag = strings.Split(jsonTag, ",")[0] // 处理 json tag 中的选项

			paramName := prefix + jsonTag

			if field.Anonymous || (field.Type.Kind() == reflect.Struct && field.Type != reflect.TypeOf(time.Time{})) {
				// 处理嵌入字段和嵌套结构
				generateParams(field.Type, prefix)
			} else {
				param := spec.Parameter{
					ParamProps: spec.ParamProps{
						Name:        paramName,
						In:          "query",
						Description: field.Tag.Get("description"),
						Required:    strings.Contains(field.Tag.Get("v"), "required"),
					},
					SimpleSchema: spec.SimpleSchema{
						Type:   f.getSwaggerType(field.Type),
						Format: f.getSwaggerFormat(field.Type),
					},
				}

				// 处理指针类型
				if field.Type.Kind() == reflect.Ptr {
					param.SimpleSchema.Type = f.getSwaggerType(field.Type.Elem())
					param.VendorExtensible.AddExtension("x-nullable", true)
				}

				// 处理数组类型
				if field.Type.Kind() == reflect.Slice || field.Type.Kind() == reflect.Array {
					param.Type = "array"
					param.Items = &spec.Items{
						SimpleSchema: spec.SimpleSchema{
							Type: f.getSwaggerType(field.Type.Elem()),
						},
					}
				}

				params = append(params, param)
			}
		}
	}

	generateParams(reqType, "")
	return params
}

func (f *APIFramework) getSwaggerType(t reflect.Type) string {
	t = deref(t)
	switch t.Kind() {
	case reflect.Bool:
		return "boolean"
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return "integer"
	case reflect.Float32, reflect.Float64:
		return "number"
	case reflect.String:
		return "string"
	case reflect.Slice, reflect.Array:
		return "array"
	case reflect.Struct:
		if t == reflect.TypeOf(time.Time{}) {
			return "string"
		}
		return "object"
	default:
		return "string"
	}
}

func (f *APIFramework) getSwaggerFormat(t reflect.Type) string {
	t = deref(t)
	switch t.Kind() {
	case reflect.Int64, reflect.Uint64:
		return "int64"
	case reflect.Int32, reflect.Uint32:
		return "int32"
	case reflect.Float32:
		return "float"
	case reflect.Float64:
		return "double"
	default:
		if t == reflect.TypeOf(time.Time{}) {
			return "date-time"
		}
		return ""
	}
}

func deref(t reflect.Type) reflect.Type {
	if t.Kind() == reflect.Ptr {
		return t.Elem()
	}
	return t
}

// generateResponses 生成 Swagger 响应定义
func (f *APIFramework) generateResponses(respType reflect.Type) *spec.Responses {
	f.debugOutput("Generating responses for type: %v", respType)
	schema := f.generateDetailedResponseSchema(respType, 0, make(map[reflect.Type]bool))
	f.debugOutput("Generated schema for responses: %+v", schema)
	return &spec.Responses{
		ResponsesProps: spec.ResponsesProps{
			StatusCodeResponses: map[int]spec.Response{
				200: {
					ResponseProps: spec.ResponseProps{
						Description: "Successful response",
						Schema:      schema,
					},
				},
			},
		},
	}
}

const maxDepth = 10

func (f *APIFramework) generateDetailedResponseSchema(fieldType reflect.Type, depth int, visited map[reflect.Type]bool) *spec.Schema {
	f.debugOutput("Generating detailed response schema for type: %v, depth: %d", fieldType, depth)
	if depth > maxDepth {
		f.debugOutput("Max depth reached for type: %v", fieldType)
		return &spec.Schema{
			SchemaProps: spec.SchemaProps{
				Type:        []string{"object"},
				Description: "Max depth reached",
			},
		}
	}

	fieldType = deref(fieldType)

	if visited[fieldType] {
		f.debugOutput("Circular reference detected for type: %v", fieldType)
		return &spec.Schema{
			SchemaProps: spec.SchemaProps{
				Type:        []string{"object"},
				Description: "Circular reference",
			},
		}
	}
	visited[fieldType] = true
	defer delete(visited, fieldType)

	var schema *spec.Schema
	switch fieldType.Kind() {
	case reflect.Struct:
		if fieldType == reflect.TypeOf(time.Time{}) {
			schema = &spec.Schema{
				SchemaProps: spec.SchemaProps{
					Type:   []string{"string"},
					Format: "date-time",
				},
			}
		} else {
			schema = f.generateSchemaForStruct(fieldType, depth+1, visited)
		}
	case reflect.Slice, reflect.Array:
		itemSchema := f.generateDetailedResponseSchema(fieldType.Elem(), depth+1, visited)
		schema = &spec.Schema{
			SchemaProps: spec.SchemaProps{
				Type:  []string{"array"},
				Items: &spec.SchemaOrArray{Schema: itemSchema},
			},
		}
	case reflect.Map:
		valueSchema := f.generateDetailedResponseSchema(fieldType.Elem(), depth+1, visited)
		schema = &spec.Schema{
			SchemaProps: spec.SchemaProps{
				Type:                 []string{"object"},
				AdditionalProperties: &spec.SchemaOrBool{Schema: valueSchema},
			},
		}
	default:
		schema = &spec.Schema{
			SchemaProps: spec.SchemaProps{
				Type:   []string{f.getSwaggerType(fieldType)},
				Format: f.getSwaggerFormat(fieldType),
			},
		}
	}
	f.debugOutput("Generated schema for type %v: %+v", fieldType, schema)
	return schema
}

func (f *APIFramework) generateSchemaForStruct(t reflect.Type, depth int, visited map[reflect.Type]bool) *spec.Schema {
	f.debugOutput("Generating schema for struct: %v, depth: %d", t, depth)
	properties := make(map[string]spec.Schema)
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		f.debugOutput("Processing field: %s", field.Name)
		if field.PkgPath != "" && !field.Anonymous { // 忽略未导出字段
			f.debugOutput("Skipping unexported field: %s", field.Name)
			continue
		}

		jsonTag := field.Tag.Get("json")
		if jsonTag == "-" {
			f.debugOutput("Skipping field %s due to json:\"-\" tag", field.Name)
			continue
		}

		propertyName := getPropertyName(field)
		f.debugOutput("Property name for field %s: %s", field.Name, propertyName)
		fieldSchema := f.generateDetailedResponseSchema(field.Type, depth+1, visited)
		if fieldSchema != nil {
			newSchema := *fieldSchema
			if description := field.Tag.Get("description"); description != "" {
				newSchema.Description = description
				f.debugOutput("Added description for field %s: %s", field.Name, description)
			}
			properties[propertyName] = newSchema
		} else {
			f.debugOutput("Warning: nil schema generated for field %s", field.Name)
		}
	}

	schema := &spec.Schema{
		SchemaProps: spec.SchemaProps{
			Type:       []string{"object"},
			Properties: properties,
		},
	}
	f.debugOutput("Generated schema for struct %v: %+v", t, schema)
	return schema
}

func getPropertyName(field reflect.StructField) string {
	jsonTag := field.Tag.Get("json")
	if jsonTag != "" {
		parts := strings.Split(jsonTag, ",")
		if parts[0] != "" {
			return parts[0]
		}
	}
	return field.Name
}

// addSwaggerPath 添加路径到 Swagger 规范
func (f *APIFramework) addSwaggerPath(def APIDefinition) {
	path := f.swaggerSpec.Paths.Paths[def.Meta.Path]
	operation := &spec.Operation{
		OperationProps: spec.OperationProps{
			Summary:     def.Meta.Summary,
			Description: def.Meta.Description,
			Tags:        strings.Split(def.Meta.Tags, ","),
			Parameters:  f.getSwaggerParams(def.RequestType),
			Responses:   f.getSwaggerResponses(def.ResponseType),
		},
	}

	switch strings.ToUpper(def.Meta.Method) {
	case "GET":
		path.Get = operation
	case "POST":
		path.Post = operation
	case "PUT":
		path.Put = operation
	case "DELETE":
		path.Delete = operation
	}

	f.swaggerSpec.Paths.Paths[def.Meta.Path] = path
}

// getSwaggerParams 从请求类型生成 Swagger 参数
func (f *APIFramework) getSwaggerParams(reqType reflect.Type) []spec.Parameter {
	var params []spec.Parameter
	for i := 0; i < reqType.Elem().NumField(); i++ {
		field := reqType.Elem().Field(i)
		if field.Anonymous {
			continue
		}
		param := spec.Parameter{
			ParamProps: spec.ParamProps{
				Name:        field.Tag.Get("json"),
				In:          "query",
				Description: field.Tag.Get("description"),
				Required:    strings.Contains(field.Tag.Get("v"), "required"),
			},
			SimpleSchema: spec.SimpleSchema{
				Type: field.Type.String(),
			},
		}
		params = append(params, param)
	}
	return params
}

// getSwaggerResponses 从响应类型生成 Swagger 响应
func (f *APIFramework) getSwaggerResponses(respType reflect.Type) *spec.Responses {
	return &spec.Responses{
		ResponsesProps: spec.ResponsesProps{
			StatusCodeResponses: map[int]spec.Response{
				200: {
					ResponseProps: spec.ResponseProps{
						Description: "Successful response",
						Schema: &spec.Schema{
							SchemaProps: spec.SchemaProps{
								Type: []string{"object"},
							},
						},
					},
				},
			},
		},
	}
}

// serveSwaggerSpec 提供 Swagger 规范 JSON
func (f *APIFramework) serveSwaggerSpec(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(f.swaggerSpec)
}
