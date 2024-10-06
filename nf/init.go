package nf

import (
	"github.com/sagoo-cloud/nexframe/nf/swagger"
	"github.com/sagoo-cloud/nexframe/utils/meta"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"strings"
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

	if f.config.StatsVizEnabled {
		f.EnableStatsviz()
	}

	// 遍历定义并设置路由
	for _, def := range f.definitions {
		testReq := reflect.New(def.RequestType.Elem()).Interface()
		if err := meta.InitMeta(testReq); err != nil {
			log.Printf("Warning: Failed to initialize Meta for %T: %v", testReq, err)
		}

		handler := f.createHandler(def)
		f.router.HandleFunc(def.Meta.Path, handler).Methods(def.Meta.Method)

		if f.debug {
			log.Printf("Registered route: %s %s", def.Meta.Method, def.Meta.Path)
		}
	}

	// 应用中间件
	if f.debug {
		log.Printf("Applying %d middlewares\n", len(f.middlewares))
	}

	f.UseErrorHandlingMiddleware()
	f.router.Use(f.createContextMiddleware())
	f.router.Use(f.domainCheckMiddleware)

	for i, mw := range f.middlewares {
		if f.debug {
			log.Printf("Applying middleware %d: %T\n", i, mw)
		}
		f.router.Use(mw)
	}

	// Swagger 相关设置
	swaggerPath := "/swagger/"
	if f.config.SwaggerPath != "" {
		swaggerPath = f.config.SwaggerPath
	}
	if f.config.OpenApiPath != "" {
		f.router.HandleFunc(f.config.OpenApiPath, f.serveSwaggerSpec)
	} else {
		f.router.HandleFunc(swaggerPath+"doc.json", f.serveSwaggerSpec)
	}

	swaggerHandler := swagger.Handler(
		swagger.TemplateContent(f.config.SwaggerUITemplate),
	)
	f.router.PathPrefix(swaggerPath).Handler(swaggerHandler)

	// 设置静态文件服务
	if f.config.FileServerEnabled && f.wwwRoot != "" {
		fileServer := f.NewStaticHandler(http.Dir(f.wwwRoot), "")
		f.router.PathPrefix("/").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 如果路径以 swaggerPath 开头，不处理
			if strings.HasPrefix(r.URL.Path, swaggerPath) {
				http.NotFound(w, r)
				return
			}

			path := filepath.Join(f.wwwRoot, r.URL.Path)
			_, err := os.Stat(path)

			if os.IsNotExist(err) {
				if r.URL.Path != "/" {
					http.NotFound(w, r)
					return
				}
				path = filepath.Join(f.wwwRoot, "index.html")
				if _, err := os.Stat(path); os.IsNotExist(err) {
					http.NotFound(w, r)
					return
				}
			} else if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			fileServer.ServeHTTP(w, r)
		})

		if f.debug {
			log.Printf("Static file server enabled for root: %s", f.wwwRoot)
		}
	}

	f.initialized = true
	if f.debug {
		log.Println("Framework initialization completed")
	}
}
