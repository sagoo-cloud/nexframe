package nf

import (
	"fmt"
	"github.com/sagoo-cloud/nexframe/nf/swagger"
	"github.com/sagoo-cloud/nexframe/utils/meta"
	"log"
	"net/http"
	"os"
	"reflect"
)

// Init 初始化框架，设置路由和处理函数
func (f *APIFramework) Init() {
	if f.initialized {
		if f.debug {
			log.Println("Framework already initialized")
		}
		return
	}

	// 设置主域名
	if f.host != "" {
		f.router.Host(f.host)
	}

	// 遍历定义并设置路由
	for _, def := range f.definitions {
		// 创建一个测试实例并尝试初始化 Meta
		testReq := reflect.New(def.RequestType.Elem()).Interface()
		if err := meta.InitMeta(testReq); err != nil {
			log.Printf("Warning: Failed to initialize Meta for %T: %v", testReq, err)
			// 这里可以选择继续初始化，或者在遇到错误时中断
			// 如果选择中断，可以使用 return 语句
			// return
		}

		handler := f.createHandler(def)
		f.router.HandleFunc(def.Meta.Path, handler).Methods(def.Meta.Method)
		f.addSwaggerPath(def)

		if f.debug {
			log.Printf("Registered route: %s %s", def.Meta.Method, def.Meta.Path)
		}
	}

	// 应用中间件
	if f.debug {
		log.Printf("Applying %d middlewares\n", len(f.middlewares))
	}

	// 应用上下文中间件
	f.router.Use(f.createContextMiddleware())

	for i, mw := range f.middlewares {
		if f.debug {
			log.Printf("Applying middleware %d: %T\n", i, mw)
		}
		f.router.Use(mw)
	}

	if f.debug {
		// 添加一个测试路由来验证中间件
		f.router.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintln(w, "Test route")
		})
		log.Println("Added test route: /test")
	}

	// 添加 Swagger UI 路由
	f.router.HandleFunc("/swagger/doc.json", f.serveSwaggerSpec)
	// Swagger UI 路由
	f.router.PathPrefix("/swagger/").Handler(swagger.WrapHandler)

	// 设置静态资源路由
	if f.wwwRoot != "" && f.fileSystem != nil {
		f.router.PathPrefix("/assets/").Handler(f.NewStaticHandler(f.fileSystem, "assets"))
	}

	// 确保静态目录存在
	if f.staticDir != "" {
		if err := os.MkdirAll(f.staticDir, 0755); err != nil {
			log.Printf("创建%s目录失败:%+v\n", f.staticDir, err)
		}
		// 设置 webroot 根目录
		f.router.PathPrefix("/").Handler(http.StripPrefix("/", http.FileServer(http.Dir(f.staticDir))))
	}

	f.initialized = true
	if f.debug {
		log.Println("Framework initialization completed")
	}
}
