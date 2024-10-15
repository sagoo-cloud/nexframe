package signals

import "context"

// SignalListener 是一个类型定义，用于定义作为信号监听器的函数。
// 这个函数接受两个参数：
//  1. 一个 `context.Context` 类型的上下文。这通常用于超时和取消信号，
//     并且可以在 API 边界和进程之间携带请求作用域的值。
//  2. 一个泛型类型 `T` 的有效载荷。这可以是任何类型，表示监听器函数
//     将要处理的数据或信号。
//
// 该函数不返回任何值。
type SignalListener[T any] func(context.Context, T)
