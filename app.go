// Package nexframe 提供了创建API服务框架的主入口。
package nexframe

import "github.com/sagoo-cloud/nexframe/nf"

// Server 创建并返回一个新的API框架实例。
// name参数目前未使用，保留用于未来扩展。
func Server(name ...interface{}) *nf.APIFramework {
	return nf.NewAPIFramework()
}
