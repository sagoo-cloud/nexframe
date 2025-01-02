package nf

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ServiceWeaver/weaver"
	"github.com/go-openapi/spec"
	"github.com/gorilla/mux"
	"github.com/sagoo-cloud/nexframe/configs"
	"github.com/sagoo-cloud/nexframe/contracts"
	"github.com/sagoo-cloud/nexframe/g"
	"github.com/sagoo-cloud/nexframe/os/file"
	"github.com/sagoo-cloud/nexframe/utils/convert"
	"github.com/sagoo-cloud/nexframe/utils/meta"
	"github.com/sagoo-cloud/nexframe/utils/valid"
	"go.uber.org/zap"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"sync"
	"time"
)

// contextKey 是用于存储自定义值的键的类型
type contextKey string

// APIDefinition 定义API结构
type APIDefinition struct {
	HandlerName  string
	RequestType  reflect.Type
	ResponseType reflect.Type
	Meta         meta.Meta
	Parameters   []spec.Parameter
	Responses    *spec.Responses
}

var (
	// 请求对象池，用于减少GC压力
	requestPool = sync.Pool{
		New: func() interface{} {
			return make([]byte, 32*1024) // 32KB 缓冲区
		},
	}
)

const (
	// 默认的文件上传大小限制：32MB
	defaultMaxMemory = 32 << 20
	// 默认的请求超时时间：30秒
	defaultTimeout = 30 * time.Second
)

// Controller 接口定义控制器的基本结构
type Controller interface {
	// 可以添加通用方法如果需要
}

// APIFramework 核心框架结构
type APIFramework struct {
	config         *configs.ServerConfig
	Host           string
	addr           string
	router         *mux.Router
	definitions    map[string]APIDefinition
	controllers    map[string]Controller
	weaverServices map[string]interface{}
	prefixes       map[string]string
	middlewares    []mux.MiddlewareFunc
	staticDir      string
	wwwRoot        string
	fileSystem     http.FileSystem
	debug          bool
	initialized    bool
	initOnce       sync.Once
	contextValues  map[contextKey]interface{}
	contextMu      sync.RWMutex
	swaggerSpec    *spec.Swagger
	host           string //主域名
	HTTPSCertPath  string
	HTTPSKeyPath   string
	ctx            context.Context
	logger         *log.Logger
}

// NewAPIFramework 创建新的APIFramework实例
func NewAPIFramework() *APIFramework {
	return &APIFramework{
		config:         configs.LoadServerConfig(),
		router:         mux.NewRouter(),
		definitions:    make(map[string]APIDefinition),
		controllers:    make(map[string]Controller),
		weaverServices: make(map[string]interface{}),
		prefixes:       make(map[string]string),
		middlewares:    []mux.MiddlewareFunc{},
		debug:          false,
		initialized:    false,
		initOnce:       sync.Once{},
		contextValues:  make(map[contextKey]interface{}),
		ctx:            context.Background(),
		logger:         log.New(os.Stdout, "", log.LstdFlags),
	}
}

// SetContextValue 设置全局上下文值
func (f *APIFramework) SetContextValue(key string, value interface{}) {
	f.contextMu.Lock()
	defer f.contextMu.Unlock()
	f.contextValues[contextKey(key)] = value
}

// GetContextValue 辅助函数，用于在控制器中获取上下文值
func GetContextValue(ctx context.Context, key string) (interface{}, bool) {
	value := ctx.Value(contextKey(key))
	return value, value != nil
}

// createContextMiddleware 创建注入自定义值的中间件
func (f *APIFramework) createContextMiddleware() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			ctx = context.WithValue(ctx, contracts.CtxKeyForRequest, r)
			f.contextMu.RLock()
			for k, v := range f.contextValues {
				ctx = context.WithValue(ctx, k, v)
			}
			f.contextMu.RUnlock()
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequestFromCtx retrieves and returns the Request object from context.
func RequestFromCtx(ctx context.Context) *http.Request {
	if v := ctx.Value(contracts.CtxKeyForRequest); v != nil {
		return v.(*http.Request)
	}
	return nil
}

// domainCheckMiddleware 获取访问的域名信息的中间件
func (f *APIFramework) domainCheckMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		host := r.Host

		if strings.Contains(host, ":") {
			host = strings.Split(host, ":")[0]
		}

		parts := strings.Split(host, ".")

		var subDomain, secondLevel, topLevel string

		switch len(parts) {
		case 1:
			secondLevel = parts[0]
			topLevel = ""
		case 2:
			secondLevel = parts[0]
			topLevel = parts[1]
		case 3:
			subDomain = parts[0]
			secondLevel = parts[1]
			topLevel = parts[2]
		default:
			subDomain = strings.Join(parts[:len(parts)-2], ".")
			secondLevel = parts[len(parts)-2]
			topLevel = parts[len(parts)-1]
		}

		domainInfo := &contracts.DomainInfo{
			FullDomain:  host,
			SubDomain:   subDomain,
			SecondLevel: secondLevel,
			TopLevel:    topLevel,
		}
		ctx := r.Context()
		ctx = context.WithValue(ctx, contracts.DomainInfoCode, domainInfo)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// EnableDebug 启用调试模式
func (f *APIFramework) EnableDebug() *APIFramework {
	f.debug = true
	return f
}

// WithMiddleware 添加一个或多个中间件
func (f *APIFramework) WithMiddleware(middlewares ...mux.MiddlewareFunc) *APIFramework {
	for _, middleware := range middlewares {
		f.middlewares = append(f.middlewares, middleware)
		f.debugOutput("Added middleware: %T\n", middleware)
	}
	return f
}

// SetStaticDir 设置静态资源目录
func (f *APIFramework) SetStaticDir(dir string) *APIFramework {
	f.staticDir = dir
	return f
}

// SetWebRoot 设置Web根目录
func (f *APIFramework) SetWebRoot(dir string) *APIFramework {
	var realPath string
	if p, err := file.Search(dir); err != nil {
		fmt.Printf(`SetStaticRoot failed: %+v \n`, err)
		realPath = dir
	} else {
		realPath = p
	}

	f.wwwRoot = strings.TrimRight(realPath, file.Separator)
	return f
}

// RegisterController 注册控制器
func (f *APIFramework) RegisterController(prefix string, controllers ...interface{}) error {
	for _, controller := range controllers {
		controllerType := reflect.TypeOf(controller)
		if controllerType.Kind() != reflect.Ptr {
			return fmt.Errorf("controller must be a pointer to struct, got %T", controller)
		}
		controllerType = controllerType.Elem()
		if controllerType.Kind() != reflect.Struct {
			return fmt.Errorf("controller must be a pointer to struct, got %T", controller)
		}

		controllerValue := reflect.ValueOf(controller).Elem()
		controllerName := controllerType.Name()

		// 存储前缀
		f.prefixes[controllerName] = prefix
		// 存储控制器
		f.controllers[controllerName] = controller

		// 注入 APIFramework 实例
		if field := controllerValue.FieldByName("F"); field.IsValid() && field.Type() == reflect.TypeOf(f) {
			field.Set(reflect.ValueOf(f))
		}

		// 尝试调用 Initialize 方法
		if initializer, ok := controller.(interface{ Initialize(*APIFramework) error }); ok {
			if err := initializer.Initialize(f); err != nil {
				return fmt.Errorf("failed to initialize controller %s: %v", controllerName, err)
			}
		}

		// 自动发现和注册 API
		if err := f.discoverAPIs(controllerName, controller); err != nil {
			return fmt.Errorf("failed to discover APIs for controller %s: %v", controllerName, err)
		}

		f.debugOutput("Registered controller: %s with prefix: %s\n", controllerName, prefix)
	}

	return nil
}

// discoverAPIs 自动发现并注册 API
func (f *APIFramework) discoverAPIs(controllerName string, controller interface{}) error {
	f.debugOutput("Discovering APIs for controller: %s\n", controllerName)
	controllerType := reflect.TypeOf(controller)
	for i := 0; i < controllerType.NumMethod(); i++ {
		method := controllerType.Method(i)
		f.debugOutput("Examining method: %s\n", method.Name)
		if method.Type.NumIn() != 3 || method.Type.NumOut() != 2 {
			continue
		}

		reqType := method.Type.In(2)
		respType := method.Type.Out(0)

		// 检查请求类型是否嵌入了 Meta
		if metaField, ok := reqType.Elem().FieldByName("Meta"); ok {
			metaData := extractMeta(metaField.Tag)
			handlerName := fmt.Sprintf("%s.%s", controllerName, method.Name)

			// 使用前缀构建完整路径
			prefix, _ := f.prefixes[controllerName]
			prefixStr := convert.String(prefix)
			fullPath := strings.TrimRight(prefixStr, "/") + "/" + strings.TrimLeft(metaData["path"], "/")

			parameters := f.generateParameters(reqType)
			responses := f.generateResponses(respType)
			f.debugOutput("Generated responses for method %s: %+v\n", method.Name, responses)
			apiDef := APIDefinition{
				HandlerName:  handlerName,
				RequestType:  reqType,
				ResponseType: respType,
				Meta: meta.Meta{
					Path:        fullPath,
					Method:      metaData["method"],
					Summary:     metaData["summary"],
					Description: metaData["description"],
					Tags:        metaData["tags"],
				},
				Parameters: parameters,
				Responses:  responses,
			}

			f.definitions[handlerName] = apiDef
			f.debugOutput("Added API definition for handler: %s", handlerName)
			//f.updateSwaggerSpec(apiDef)
			f.debugOutput("Discovered API: %s %s - %s\n", apiDef.Meta.Method, fullPath, apiDef.Meta.Summary)
		}
	}

	return nil
}

// extractMeta 从字段标签中提取元数据
func extractMeta(tag reflect.StructTag) map[string]string {
	metaData := make(map[string]string)
	for _, key := range []string{"path", "method", "summary", "description", "tags"} {
		if value := tag.Get(key); value != "" {
			metaData[key] = value
		}
	}
	return metaData
}

// injectDependencies 注入依赖（包括框架和 ServiceWeaver 上下文）
func (f *APIFramework) injectDependencies(controller interface{}) {
	controllerValue := reflect.ValueOf(controller).Elem()
	controllerType := controllerValue.Type()

	for i := 0; i < controllerType.NumField(); i++ {
		field := controllerType.Field(i)
		if field.Type == reflect.TypeOf(f) && isExported(field.Name) {
			controllerValue.Field(i).Set(reflect.ValueOf(f))
			f.debugOutput("Injected framework instance into %s\n", controllerType.Name())
		} else if service, err := f.GetWeaverService(field.Name); err == nil {
			controllerValue.Set(reflect.ValueOf(service))
		}
	}
}

// isExported 检查字段是否可导出
func isExported(fieldName string) bool {
	return fieldName[0] >= 'A' && fieldName[0] <= 'Z'
}

// GetController 获取已注册的控制器
func (f *APIFramework) GetController(name string) (interface{}, bool) {
	controller, ok := f.controllers[name]
	return controller, ok
}

// createHandler 创建处理函数
func (f *APIFramework) createHandler(def APIDefinition) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 添加请求超时控制
		ctx, cancel := context.WithTimeout(r.Context(), defaultTimeout)
		defer cancel()
		// 使用对象池获取缓冲区
		buf := requestPool.Get().([]byte)
		defer requestPool.Put(buf)

		// 创建请求对象
		reqValue := reflect.New(def.RequestType.Elem())
		req := reqValue.Interface()

		// 初始化Meta（添加错误处理）
		if err := meta.InitMeta(req); err != nil {
			f.debugOutput("初始化Meta失败", zap.Error(err))
			http.Error(w, "Failed to initialize request metadata", http.StatusInternalServerError)
			return
		}

		// panic恢复
		defer func() {
			if r := recover(); r != nil {
				f.debugOutput("Handler panic",
					zap.Any("panic", r),
					zap.String("handler", def.HandlerName),
					zap.Stack("stack"))
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
		}()

		// 处理请求（其余代码保持不变）
		var err error
		contentType := r.Header.Get("Content-Type")
		if strings.HasPrefix(contentType, "multipart/form-data") {
			err = f.handleMultipartRequest(r, req)
		} else {
			switch r.Method {
			case http.MethodGet:
				err = f.decodeGetRequest(r, req)
			case http.MethodPost, http.MethodPut, http.MethodPatch:
				err = f.decodeJSONRequest(r, req)
			case http.MethodDelete:
				err = f.decodeDeleteRequest(r, req)
			default:
				http.Error(w, "Unsupported method", http.StatusMethodNotAllowed)
				return
			}
		}

		if err != nil {
			f.debugOutput("请求处理失败",
				zap.Error(err),
				zap.String("handler", def.HandlerName))
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// 验证请求
		validator := valid.New()
		if err := validator.Data(req).Run(ctx); err != nil {
			contracts.JsonExit(w, http.StatusBadRequest, "验证失败: "+err.Error())
			return
		}

		// 获取控制器（添加空指针检查）
		controllerName := strings.Split(def.HandlerName, ".")[0]
		controller, ok := f.controllers[controllerName]
		if !ok {
			f.debugOutput("控制器未找到", zap.String("controller", controllerName))
			http.Error(w, "Controller not found", http.StatusInternalServerError)
			return
		}

		// 调用方法（添加方法存在检查）
		methodName := strings.Split(def.HandlerName, ".")[1]
		method := reflect.ValueOf(controller).MethodByName(methodName)
		if !method.IsValid() {
			f.debugOutput("方法未找到",
				zap.String("controller", controllerName),
				zap.String("method", methodName))
			http.Error(w, "Method not found", http.StatusInternalServerError)
			return
		}

		// 执行处理方法
		results := method.Call([]reflect.Value{
			reflect.ValueOf(ctx),
			reqValue,
		})

		// 处理响应（添加结果检查）
		if len(results) < 2 {
			f.debugOutput("方法返回值数量错误",
				zap.String("handler", def.HandlerName))
			http.Error(w, "Invalid handler response", http.StatusInternalServerError)
			return
		}

		if !results[0].IsValid() {
			f.debugOutput("方法返回值无效",
				zap.String("handler", def.HandlerName))
			http.Error(w, "Invalid handler response", http.StatusInternalServerError)
			return
		}

		// 处理错误返回
		if len(results) > 1 && !results[1].IsNil() {
			err := results[1].Interface().(error)
			f.debugOutput("处理请求失败",
				zap.Error(err),
				zap.String("handler", def.HandlerName))
			contracts.JsonExit(w, http.StatusInternalServerError, "内部服务器错误: "+err.Error())
			return
		}

		// 设置自定义头部信息和响应
		if headers, ok := results[0].Interface().(contracts.ResponseWithHeaders); ok {
			// 设置响应头
			for key, value := range headers.Headers {
				w.Header().Set(key, value)
			}

			// 检查是否是文件下载（通过Content-Type判断）
			if ct, exists := headers.Headers["Content-Type"]; exists && ct == "application/force-download" {
				// 文件下载的情况，直接写入数据
				if data, ok := headers.Data.([]byte); ok {
					_, err := w.Write(data)
					if err != nil {
						f.debugOutput("写入文件数据失败", zap.Error(err))
						http.Error(w, "文件下载失败", http.StatusInternalServerError)
					}
					return
				}
			}

			// 非文件下载的普通JSON响应
			contracts.JsonExit(w, 0, "Success", headers.Data)
		} else {
			// 普通响应（没有自定义头部）
			contracts.JsonExit(w, 0, "Success", results[0].Interface())
		}
	}
}

// handleMultipartRequest 处理文件上传请求
func (f *APIFramework) handleMultipartRequest(r *http.Request, dst interface{}) error {
	// 检查请求Content-Type
	contentType := r.Header.Get("Content-Type")
	if !strings.Contains(contentType, "multipart/form-data") {
		return fmt.Errorf("无效的Content-Type: %s, 需要 multipart/form-data", contentType)
	}

	// 使用配置的最大内存限制
	maxMemory := defaultMaxMemory
	if f.config != nil && f.config.MaxUploadSize > 0 {
		maxMemory = f.config.MaxUploadSize
	}

	if err := r.ParseMultipartForm(int64(maxMemory)); err != nil {
		return fmt.Errorf("解析表单失败: %w", err)
	}

	if r.MultipartForm == nil || r.MultipartForm.File == nil {
		return fmt.Errorf("未检测到上传的文件")
	}

	// Debug输出
	if f.debug {
		log.Printf("接收到的表单字段: %v\n", r.MultipartForm.Value)
		log.Printf("接收到的文件字段: %v\n", r.MultipartForm.File)
	}

	// 使用反射获取目标结构体的值
	v := reflect.ValueOf(dst).Elem()
	t := v.Type()

	// 遍历所有字段
	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)

		// 跳过 Meta 字段
		if field.Anonymous && field.Type == reflect.TypeOf(g.Meta{}) {
			continue
		}

		// 获取字段的json标签作为表单字段名
		jsonTag := field.Tag.Get("json")
		if jsonTag == "" {
			jsonTag = field.Name
		}
		jsonTag = strings.Split(jsonTag, ",")[0]

		// Debug输出
		if f.debug {
			log.Printf("处理字段: %s (tag: %s)\n", field.Name, jsonTag)
		}

		// 如果是文件上传字段
		if field.Type == reflect.TypeOf([]meta.FileUploadMeta{}) {
			files := r.MultipartForm.File[jsonTag]
			if len(files) > 0 {
				fileMetaSlice := make([]meta.FileUploadMeta, len(files))
				for i, fileHeader := range files {
					if f.debug {
						log.Printf("处理文件: %s, 大小: %d\n", fileHeader.Filename, fileHeader.Size)
					}

					fileMeta := meta.FileUploadMeta{
						FileName:   fileHeader.Filename,
						Size:       fileHeader.Size,
						FileHeader: fileHeader,
					}

					// 检测文件类型
					if err := fileMeta.DetectContentType(); err != nil {
						return fmt.Errorf("检测文件类型失败: %w", err)
					}

					fileMetaSlice[i] = fileMeta
				}
				v.Field(i).Set(reflect.ValueOf(fileMetaSlice))
			} else if f.debug {
				log.Printf("字段 %s 没有接收到文件\n", jsonTag)
			}
			continue
		}

		// 处理普通表单字段
		if values, ok := r.MultipartForm.Value[jsonTag]; ok && len(values) > 0 {
			if err := setField(v.Field(i), values[0]); err != nil {
				return fmt.Errorf("设置字段 %s 失败: %w", field.Name, err)
			}
			if f.debug {
				log.Printf("设置表单字段 %s = %s\n", jsonTag, values[0])
			}
		} else if f.debug {
			log.Printf("字段 %s 没有接收到值\n", jsonTag)
		}
	}

	return nil
}

// decodeJSONRequest 处理 JSON 请求体，支持复杂参数结构
func (f *APIFramework) decodeJSONRequest(r *http.Request, dst interface{}) error {
	// 检查 Content-Type，如果是 multipart/form-data，则跳过 JSON 解析
	contentType := r.Header.Get("Content-Type")
	if strings.HasPrefix(contentType, "multipart/form-data") {
		return nil
	}

	// 根据 Content-Type 选择不同的处理方式
	if strings.Contains(contentType, "application/x-www-form-urlencoded") {
		// 处理 form-urlencoded 格式的请求
		if err := r.ParseForm(); err != nil {
			return fmt.Errorf("解析表单失败: %w", err)
		}
		return f.decodeStructFromValues(r.Form, reflect.ValueOf(dst).Elem())
	}

	// 处理 JSON 格式的请求
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return fmt.Errorf("读取请求体失败: %v", err)
	}
	defer r.Body.Close()

	f.debugOutput("正在解析 JSON 请求: %T\n", dst)

	// 尝试直接解析到目标结构
	if err := json.Unmarshal(body, dst); err != nil {
		// 如果直接解析失败，尝试使用中间 map 进行解析
		var tempData map[string]interface{}
		if err := json.Unmarshal(body, &tempData); err != nil {
			return fmt.Errorf("JSON解析失败: %v", err)
		}

		// 递归处理复杂结构
		return f.setComplexValue(reflect.ValueOf(dst).Elem(), tempData)
	}

	if f.debug {
		jsonBytes, _ := json.MarshalIndent(dst, "", "  ")
		log.Printf("解析后的请求对象:\n%s", string(jsonBytes))
	}

	return nil
}

// setComplexValue 处理复杂的值设置，支持嵌套结构
func (f *APIFramework) setComplexValue(v reflect.Value, data interface{}) error {
	// 处理空值
	if data == nil {
		return nil
	}

	switch v.Kind() {
	case reflect.Struct:
		// 处理结构体
		m, ok := data.(map[string]interface{})
		if !ok {
			return fmt.Errorf("无法将 %T 转换为结构体", data)
		}
		return f.setStructFields(v, m)

	case reflect.Map:
		// 处理 map
		return f.setMapValue(v, data)

	case reflect.Slice:
		// 处理切片
		return f.setSliceValue(v, data)

	case reflect.Ptr:
		// 处理指针
		if v.IsNil() {
			v.Set(reflect.New(v.Type().Elem()))
		}
		return f.setComplexValue(v.Elem(), data)

	default:
		// 处理基本类型
		return setField(v, data)
	}
}

// setStructFields 处理结构体字段的设置
func (f *APIFramework) setStructFields(v reflect.Value, data map[string]interface{}) error {
	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)

		// 跳过匿名字段
		if field.Anonymous {
			continue
		}

		// 获取字段名
		jsonTag := field.Tag.Get("json")
		if jsonTag == "" {
			jsonTag = field.Name
		}
		jsonTag = strings.Split(jsonTag, ",")[0]

		// 获取字段值
		if value, ok := data[jsonTag]; ok {
			if err := f.setComplexValue(v.Field(i), value); err != nil {
				return fmt.Errorf("设置字段 %s 失败: %w", field.Name, err)
			}
		}
	}
	return nil
}

// setMapValue 处理 map 类型的设置
func (f *APIFramework) setMapValue(v reflect.Value, data interface{}) error {
	if data == nil {
		return nil
	}

	// 确保 map 已初始化
	if v.IsNil() {
		v.Set(reflect.MakeMap(v.Type()))
	}

	// 获取 map 的键值类型
	keyType := v.Type().Key()
	valueType := v.Type().Elem()

	switch d := data.(type) {
	case map[string]interface{}:
		// 处理字符串键的 map
		for key, val := range d {
			// 创建并设置键
			mapKey := reflect.New(keyType).Elem()
			if err := setField(mapKey, key); err != nil {
				return fmt.Errorf("设置map键失败: %w", err)
			}

			// 创建并设置值
			mapVal := reflect.New(valueType).Elem()
			if err := f.setComplexValue(mapVal, val); err != nil {
				return fmt.Errorf("设置map值失败: %w", err)
			}

			v.SetMapIndex(mapKey, mapVal)
		}
	default:
		return fmt.Errorf("不支持的map数据类型: %T", data)
	}

	return nil
}

// setSliceValue 处理切片类型的设置
func (f *APIFramework) setSliceValue(v reflect.Value, data interface{}) error {
	slice, ok := data.([]interface{})
	if !ok {
		// 尝试处理单个值
		newSlice := reflect.MakeSlice(v.Type(), 1, 1)
		if err := f.setComplexValue(newSlice.Index(0), data); err != nil {
			return fmt.Errorf("设置切片元素失败: %w", err)
		}
		v.Set(newSlice)
		return nil
	}

	// 创建新切片
	newSlice := reflect.MakeSlice(v.Type(), len(slice), len(slice))
	for i, item := range slice {
		if err := f.setComplexValue(newSlice.Index(i), item); err != nil {
			return fmt.Errorf("设置切片索引 %d 失败: %w", i, err)
		}
	}
	v.Set(newSlice)
	return nil
}

// decodeDeleteRequest 处理 DELETE 请求
func (f *APIFramework) decodeDeleteRequest(r *http.Request, dst interface{}) error {
	// 首先尝试从 URL 参数解析
	if err := f.decodeGetRequest(r, dst); err != nil {
		return err
	}

	// 如果请求体不为空，也尝试解析 JSON
	if r.ContentLength > 0 {
		if err := json.NewDecoder(r.Body).Decode(dst); err != nil {
			return err
		}
	}

	return nil
}

func (f *APIFramework) decodeGetRequest(r *http.Request, dst interface{}) error {
	values := r.URL.Query()
	return f.decodeStructFromValues(values, reflect.ValueOf(dst).Elem())
}

func (f *APIFramework) decodeStructFromValues(values url.Values, v reflect.Value) error {
	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		fieldValue := v.Field(i)

		// 处理匿名字段
		if field.Anonymous {
			if err := f.handleAnonymousField(values, fieldValue); err != nil {
				return err
			}
			continue
		}

		fieldName, shouldFill := getFieldName(field)
		if !shouldFill {
			continue
		}

		// 处理各种字段类型
		switch field.Type.Kind() {
		case reflect.Struct:
			if err := f.handleStructField(values, fieldValue, fieldName); err != nil {
				return err
			}
		case reflect.Ptr:
			if err := f.handlePtrField(values, fieldValue, fieldName); err != nil {
				return err
			}
		case reflect.Slice:
			if err := f.handleSliceField(values, fieldValue, fieldName, field.Type); err != nil {
				return err
			}
		case reflect.Map:
			if err := f.handleMapField(values, fieldValue, fieldName, field.Type); err != nil {
				return err
			}
		default:
			// 处理基本类型
			if value := values.Get(fieldName); value != "" {
				if err := setField(fieldValue, value); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

// handleAnonymousField 处理匿名字段
func (f *APIFramework) handleAnonymousField(values url.Values, fieldValue reflect.Value) error {
	if fieldValue.Kind() == reflect.Struct {
		return f.decodeStructFromValues(values, fieldValue)
	} else if fieldValue.Kind() == reflect.Ptr && fieldValue.Type().Elem().Kind() == reflect.Struct {
		if fieldValue.IsNil() {
			fieldValue.Set(reflect.New(fieldValue.Type().Elem()))
		}
		return f.decodeStructFromValues(values, fieldValue.Elem())
	}
	return nil
}

// handleStructField 处理结构体字段
func (f *APIFramework) handleStructField(values url.Values, fieldValue reflect.Value, prefix string) error {
	// 创建新的前缀用于嵌套结构
	prefixedValues := make(url.Values)
	for key, vals := range values {
		if strings.HasPrefix(key, prefix+".") || strings.HasPrefix(key, prefix+"[") {
			newKey := strings.TrimPrefix(key, prefix+".")
			newKey = strings.TrimPrefix(newKey, prefix+"[")
			newKey = strings.TrimSuffix(newKey, "]")
			prefixedValues[newKey] = vals
		}
	}
	return f.decodeStructFromValues(prefixedValues, fieldValue)
}

// handlePtrField 处理指针字段
func (f *APIFramework) handlePtrField(values url.Values, fieldValue reflect.Value, fieldName string) error {
	if fieldValue.IsNil() {
		fieldValue.Set(reflect.New(fieldValue.Type().Elem()))
	}
	return f.decodeStructFromValues(values, fieldValue.Elem())
}

// handleSliceField 处理切片字段
func (f *APIFramework) handleSliceField(values url.Values, fieldValue reflect.Value, fieldName string, fieldType reflect.Type) error {
	var sliceValues []string

	// 支持多种数组参数格式
	patterns := []string{
		fieldName,        // 普通格式: key=value&key=value
		fieldName + "[]", // 带方括号格式: key[]=value&key[]=value
		fieldName + "[",  // 带索引格式: key[0]=value&key[1]=value
	}

	// 收集所有匹配的值
	for _, pattern := range patterns {
		if pattern == fieldName || pattern == fieldName+"[]" {
			if vals := values[pattern]; len(vals) > 0 {
				sliceValues = append(sliceValues, vals...)
			}
		} else {
			// 处理带索引的情况
			for key, vals := range values {
				if strings.HasPrefix(key, pattern) && strings.HasSuffix(key, "]") {
					sliceValues = append(sliceValues, vals...)
				}
			}
		}
	}

	if len(sliceValues) > 0 {
		slice := reflect.MakeSlice(fieldType, len(sliceValues), len(sliceValues))
		for i, val := range sliceValues {
			if err := setField(slice.Index(i), val); err != nil {
				return err
			}
		}
		fieldValue.Set(slice)
	}

	return nil
}

// handleMapField 处理 map 字段，支持复杂的嵌套结构
func (f *APIFramework) handleMapField(values url.Values, fieldValue reflect.Value, fieldName string, fieldType reflect.Type) error {
	// 创建新的 map
	mapValue := reflect.MakeMap(fieldType)
	keyType := fieldType.Key()
	elemType := fieldType.Elem()

	// 处理多层嵌套的 map
	prefix := fieldName + "["
	suffix := "]"

	// 遍历所有键值对
	for key, vals := range values {
		if !strings.HasPrefix(key, prefix) || !strings.HasSuffix(key, suffix) {
			continue
		}

		// 提取 map 的键
		mapKey := key[len(prefix):strings.LastIndex(key, suffix)]

		// 处理嵌套的 map 键
		if strings.Contains(mapKey, "[") {
			parts := strings.Split(mapKey, "[")
			for i := range parts {
				parts[i] = strings.TrimSuffix(parts[i], "]")
			}

			// 构建嵌套 map 的值
			currentMap := mapValue
			for i := 0; i < len(parts)-1; i++ {
				keyVal := reflect.New(keyType).Elem()
				if err := setField(keyVal, parts[i]); err != nil {
					return fmt.Errorf("设置嵌套map键失败: %w", err)
				}

				// 如果当前键不存在，创建新的嵌套map
				if !currentMap.MapIndex(keyVal).IsValid() {
					nestedMap := reflect.MakeMap(elemType)
					currentMap.SetMapIndex(keyVal, nestedMap)
				}
				currentMap = currentMap.MapIndex(keyVal)
			}

			// 设置最终的值
			keyVal := reflect.New(keyType).Elem()
			if err := setField(keyVal, parts[len(parts)-1]); err != nil {
				return fmt.Errorf("设置最终map键失败: %w", err)
			}
			val := reflect.New(elemType).Elem()
			if err := setField(val, vals[0]); err != nil {
				return fmt.Errorf("设置map值失败: %w", err)
			}
			currentMap.SetMapIndex(keyVal, val)
		} else {
			// 处理简单的 map 键
			keyVal := reflect.New(keyType).Elem()
			if err := setField(keyVal, mapKey); err != nil {
				return fmt.Errorf("设置简单map键失败: %w", err)
			}
			val := reflect.New(elemType).Elem()
			if err := setField(val, vals[0]); err != nil {
				return fmt.Errorf("设置简单map值失败: %w", err)
			}
			mapValue.SetMapIndex(keyVal, val)
		}
	}

	// 只有在有值的情况下才设置 map
	if mapValue.Len() > 0 {
		fieldValue.Set(mapValue)
	}

	return nil
}

// getFieldName 获取字段的名称
func getFieldName(field reflect.StructField) (string, bool) {
	if tag, ok := field.Tag.Lookup("p"); ok && tag != "" {
		return tag, true
	}

	if tag, ok := field.Tag.Lookup("json"); ok && tag != "" {
		parts := strings.Split(tag, ",")
		if parts[0] != "" {
			return parts[0], true
		}
	}

	// 如果没有标签，使用字段名
	return field.Name, true
}

// GetServer 返回http.Handler接口，用于启动服务
func (f *APIFramework) GetServer() http.Handler {
	f.initOnce.Do(func() {
		f.debugOutput("Initializing framework in GetServer\n")
		f.Init()
	})
	return f.router
}

func (f *APIFramework) SetPort(addr string) {
	f.addr = addr
}
func (f *APIFramework) SetHost(host string) {
	f.host = host
}

func (f *APIFramework) Run(httpServes ...weaver.Listener) (err error) {
	if f.addr == "" {
		f.addr = f.config.Address
	}
	if f.host == "" {
		f.host = f.config.Host
	}

	if len(httpServes) == 0 {
		swaggerUrl := fmt.Sprintf("API Doc: http://localhost%s/swagger/index.html", f.addr)
		log.Printf(swaggerUrl)
		// 创建 HTTP 服务器
		srv := &http.Server{
			Addr:         f.addr,
			Handler:      f.GetServer(),
			ReadTimeout:  f.config.ReadTimeout,
			WriteTimeout: f.config.WriteTimeout,
			IdleTimeout:  f.config.IdleTimeout,
		}

		// 启动 HTTP 服务器
		go func() {
			log.Printf("%s Starting HTTP server on %s", f.config.Name, f.addr)
			if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
				log.Fatalf("HTTP server error: %s,%v", f.config.Name, err)
			}
		}()

		if f.config.HTTPSAddress != "" && f.config.HTTPSCertPath != "" && f.config.HTTPSKeyPath != "" {
			go func() {
				log.Printf("%s Starting HTTPS server on %s", f.config.Name, f.config.HTTPSAddress)
				httpsServer := &http.Server{
					Addr:         f.config.HTTPSAddress,
					Handler:      f.GetServer(),
					ReadTimeout:  f.config.ReadTimeout,
					WriteTimeout: f.config.WriteTimeout,
					IdleTimeout:  f.config.IdleTimeout,
					TLSConfig:    &tls.Config{MinVersion: tls.VersionTLS12},
				}
				if err := httpsServer.ListenAndServeTLS(f.config.HTTPSCertPath, f.config.HTTPSKeyPath); err != nil && !errors.Is(err, http.ErrServerClosed) {
					log.Fatalf("HTTPS server error: %v", err)
				}
			}()
		}
	}

	for _, web := range httpServes {
		addr := strings.Split(web.Addr().String(), ":")
		n := len(addr) - 1
		addPort := ":" + addr[n]
		f.SetPort(addPort)
		swaggerUrl := fmt.Sprintf("API Doc: http://localhost%s/swagger/index.html", addPort)
		log.Printf(swaggerUrl)

		//创建 HTTP 服务器
		srv := &http.Server{
			Handler: f.GetServer(),

			ReadTimeout:  f.config.ReadTimeout,
			WriteTimeout: f.config.WriteTimeout,
			IdleTimeout:  f.config.IdleTimeout,
		}
		//启动 HTTP 服务器
		go func() {
			log.Printf("%s Starting HTTP server on %s", f.config.Name, f.addr)
			if err := srv.Serve(web); err != nil && !errors.Is(err, http.ErrServerClosed) {
				log.Fatalf("HTTP server error: %s,%v", f.config.Name, err)
			}
		}()

		if f.config.HTTPSAddress != "" && f.config.HTTPSCertPath != "" && f.config.HTTPSKeyPath != "" {
			go func() {
				log.Printf("%s Starting HTTPS server on %s", f.config.Name, f.config.HTTPSAddress)
				httpsServer := &http.Server{
					Handler:      f.GetServer(),
					ReadTimeout:  f.config.ReadTimeout,
					WriteTimeout: f.config.WriteTimeout,
					IdleTimeout:  f.config.IdleTimeout,
					TLSConfig:    &tls.Config{MinVersion: tls.VersionTLS12},
				}

				if err := httpsServer.ServeTLS(web, f.config.HTTPSCertPath, f.config.HTTPSKeyPath); err != nil && err != http.ErrServerClosed {
					log.Fatalf("HTTPS server error: %v", err)
				}
			}()
		}
	}

	return nil
}

const dumpTextFormat = ` %s   |    %s     |      %s         `

// PrintAPIRoutes 输出所有注册的API访问地址
func (f *APIFramework) PrintAPIRoutes() {
	fmt.Println("Registered API Routes:")
	fmt.Println("| Method | Path                       | Summary                 \n----------------------------------------------------------------------------")

	var routes []string
	for _, def := range f.definitions {
		route := fmt.Sprintf(dumpTextFormat, def.Meta.Method, def.Meta.Path, def.Meta.Summary)
		routes = append(routes, route)
	}

	// 排序路由以便更容易阅读
	sort.Strings(routes)

	for _, route := range routes {
		fmt.Println(route)
		fmt.Println("----------------------------------------------------------------------------")
	}

	// 生成 Swagger JSON
	f.swaggerSpec = f.generateSwaggerJSON()

}

func (f *APIFramework) generateSwaggerJSON() *spec.Swagger {
	swagger := &spec.Swagger{
		SwaggerProps: spec.SwaggerProps{
			Swagger: "2.0",
			Info: &spec.Info{
				InfoProps: spec.InfoProps{
					Title:       "API Documentation",
					Description: "API documentation generated by the framework",
					Version:     "1.0.0",
				},
			},
			Paths: &spec.Paths{Paths: make(map[string]spec.PathItem)},
		},
	}

	for _, def := range f.definitions {
		path := def.Meta.Path
		method := strings.ToLower(def.Meta.Method)

		operation := &spec.Operation{
			OperationProps: spec.OperationProps{
				Summary:     def.Meta.Summary,
				Description: def.Meta.Description,
				Tags:        strings.Split(def.Meta.Tags, ","),
				Parameters:  def.Parameters,
				Responses:   def.Responses,
			},
		}

		pathItem, ok := swagger.Paths.Paths[path]
		if !ok {
			pathItem = spec.PathItem{}
		}

		switch method {
		case "get":
			pathItem.Get = operation
		case "post":
			pathItem.Post = operation
		case "put":
			pathItem.Put = operation
		case "delete":
			pathItem.Delete = operation
			// 添加其他 HTTP 方法的处理...
		}

		swagger.Paths.Paths[path] = pathItem

	}

	return swagger
}

func (f *APIFramework) debugOutput(format string, v ...any) {
	if f.debug {
		fmt.Printf(format, v...)
	}
}
func (f *APIFramework) Context() context.Context {
	if f.ctx != nil {
		return f.ctx
	}
	return context.Background()
}

func (f *APIFramework) handleFileUpload(w http.ResponseWriter, r *http.Request, config FileConfig) (map[string]*UploadedFile, error) {
	// 检查请求类型
	if !strings.HasPrefix(r.Header.Get("Content-Type"), "multipart/form-data") {
		return nil, fmt.Errorf("invalid content type, expected multipart/form-data")
	}

	// 设置最大内存，超过此大小的文件会被存储到临时文件中
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		return nil, fmt.Errorf("failed to parse multipart form: %w", err)
	}

	var totalSize int64
	files := make(map[string]*UploadedFile)

	// 遍历所有文件字段
	for field, headers := range r.MultipartForm.File {
		fieldConfig, hasConfig := config.Fields[field]

		for _, header := range headers {
			// 验证文件大小
			if header.Size > config.MaxFileSize {
				return nil, fmt.Errorf("file %s exceeds maximum size limit", header.Filename)
			}

			totalSize += header.Size
			if totalSize > config.MaxTotalSize {
				return nil, fmt.Errorf("total upload size exceeds limit")
			}

			// 验证文件类型
			file, err := header.Open()
			if err != nil {
				return nil, fmt.Errorf("failed to open uploaded file: %w", err)
			}
			defer file.Close()

			// 读取文件头以检测实际文件类型
			buff := make([]byte, 512)
			_, err = file.Read(buff)
			if err != nil {
				return nil, fmt.Errorf("failed to read file header: %w", err)
			}
			file.Seek(0, 0) // 重置文件指针

			contentType := http.DetectContentType(buff)
			isAllowedType := false
			allowedTypes := config.AllowedTypes
			if hasConfig {
				allowedTypes = fieldConfig.AllowTypes
			}
			for _, allowed := range allowedTypes {
				if strings.HasPrefix(contentType, allowed) {
					isAllowedType = true
					break
				}
			}
			if !isAllowedType {
				return nil, fmt.Errorf("file type %s is not allowed", contentType)
			}

			files[field] = &UploadedFile{
				Filename:    header.Filename,
				Size:        header.Size,
				ContentType: contentType,
				File:        file,
				FileHeader:  header,
			}
		}
	}

	return files, nil
}

// SaveUploadedFile 保存上传的文件
func (f *APIFramework) SaveUploadedFile(file *UploadedFile, destPath string) error {
	// 创建目标目录
	if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// 创建目标文件
	dst, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer dst.Close()

	// 复制文件内容
	if _, err = io.Copy(dst, file.File); err != nil {
		return fmt.Errorf("failed to save file: %w", err)
	}

	return nil
}
