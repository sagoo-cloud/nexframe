# NexFrame

NexFrame是一款基于 Go 语言免费开源的，快速、简单的企业级应用开发框架，旨在为开发者提供高效、可靠、可扩展的应用开发框架。

使用请参考【[开发手册](https://sagoo.gitbook.io/nexframe)】

## 主要特性

* **微服务架构：** 通过 **serviceweaver** 提供强大的微服务能力，支持服务注册、发现、负载均衡等功能，帮助开发者构建灵活、可伸缩的分布式系统。
* **ORM 支持：** 集成 **GORM** ORM 框架，简化数据库操作，提供丰富的功能，如结构映射、查询构建、关联关系等，提升开发效率。
* **高性能路由：** 使用 **gorilla/mux** 构建高性能的 HTTP 路由系统，支持 URL 参数、正则表达式匹配等功能，灵活配置路由规则。
* **API 自动绑定：** 借鉴 **GoFrame** 的设计理念，实现 API 定义与控制器自动绑定，简化 API 开发流程，提高开发效率。
* **OpenAPI 文档自动生成：** 支持自动生成 OpenAPI 规范文档，方便 API 的调试和集成。
* **模块化设计：** 采用模块化设计，各模块之间耦合度低，易于扩展和维护。
* **丰富的中间件：** 提供丰富的中间件，如日志记录、错误处理、权限验证等，方便开发者定制化开发。
* **强大的工具集：** 提供一系列开发工具，如代码生成、测试框架等，提升开发效率。

## **优势**

* **高性能：** 基于 Go 语言的高并发特性，以及 **serviceweaver**、**gorilla/mux** 等高性能库，保证了框架的高性能。
* **易用性：** 提供简洁易用的 API，降低开发门槛，提高开发效率。
* **可扩展性：** 采用模块化设计，易于扩展和定制。
* **可靠性：** 经过大量测试和验证，保证了框架的稳定性和可靠性。

## **适用场景**

* **微服务架构：** 适用于构建大型、复杂的微服务系统。
* **企业级应用：** 适用于开发企业级的后台管理系统、业务系统等。

## web服务

```go
package main

import (
	"fmt"
	"github.com/sagoo-cloud/nexframe"
	"net/http"
)

func main() {
	server := nexframe.Server()
	// 注册控制器
	err := server.BindHandlerFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println(w, "Hello, world!")
	})
	if err != nil {
		return
	}
	server.SetPort(":8080")
	server.Run()
}

```

