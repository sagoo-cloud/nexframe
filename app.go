// NexFrame是一款基于 Go 语言免费开源的，快速、简单的企业级应用开发框架，旨在为开发者提供高效、可靠、可扩展的应用开发框架。
// 通过集成ServiceWeaver服务能力，实现单体开发，微服务运行的效果。
package nexframe

import "github.com/sagoo-cloud/nexframe/nf"

// Server 创建并返回一个新的API框架实例。
// name参数目前未使用，保留用于未来扩展。
func Server(name ...interface{}) *nf.APIFramework {
	return nf.NewAPIFramework()
}
