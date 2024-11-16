// Package g 提供全局工具函数。
package g

import (
	"context"
	"github.com/sagoo-cloud/nexframe/utils"
)

// Try 实现了类似try语句的功能，使用内部panic...recover机制。
// 如果执行过程中发生任何异常，返回对应的error；否则返回nil。
// 参数:
//   - ctx: 上下文对象，可用于传递请求作用域的数据
//   - try: 要执行的函数，可能会panic的代码块
func Try(ctx context.Context, try func(ctx context.Context)) (err error) {
	return utils.Try(ctx, try)
}

// TryCatch 实现了类似try...catch语句的功能，使用内部panic...recover机制。
// 当发生异常时，自动调用catch函数并将异常作为error传入。
// 注意：如果catch函数内部也发生panic，当前goroutine将会panic。
//
// 参数:
//   - ctx: 上下文对象，可用于传递请求作用域的数据
//   - try: 要执行的函数，可能会panic的代码块
//   - catch: 异常处理函数，接收上下文和异常作为参数
func TryCatch(ctx context.Context, try func(ctx context.Context), catch func(ctx context.Context, exception error)) {
	utils.TryCatch(ctx, try, catch)
}
