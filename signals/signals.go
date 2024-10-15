package signals

import "context"

// Signal 是表示可以订阅的信号的接口，该信号发出 T 类型的有效载荷。
type Signal[T any] interface {
	// Emit 通知信号的所有订阅者，并传递上下文和有效载荷。
	//
	// 如果上下文有截止日期或可取消属性，监听器必须遵守它。
	// 如果信号是异步的（默认），监听器将在单独的 goroutine 中被调用。
	//
	// 示例：
	// signal := signals.New[int]()
	// signal.AddListener(func(ctx context.Context, payload int) {
	//    // 监听器实现
	//    // ...
	// })
	// signal.Emit(context.Background(), 42)
	Emit(ctx context.Context, payload T) error

	// AddListener 向信号添加一个监听器。
	//
	// 每当信号被发出时，监听器将被调用。它返回添加监听器后的订阅者数量。
	// 它接受一个可选的键，可用于稍后删除监听器或检查监听器是否已添加。
	// 如果具有相同键的监听器已添加到信号中，则返回 -1。
	//
	// 示例：
	// signal := signals.NewSync[int]()
	// count := signal.AddListener(func(ctx context.Context, payload int) {
	//    // 监听器实现
	//    // ...
	// })
	// fmt.Println("添加监听器后的订阅者数量：", count)
	AddListener(handler SignalListener[T], key ...string) int

	// RemoveListener 从信号中移除监听器。
	//
	// 它返回移除监听器后的订阅者数量。
	// 如果未找到监听器，则返回 -1。
	//
	// 示例：
	// signal := signals.NewSync[int]()
	// signal.AddListener(func(ctx context.Context, payload int) {
	//    // 监听器实现
	//    // ...
	// }, "key1")
	// count := signal.RemoveListener("key1")
	// fmt.Println("移除监听器后的订阅者数量：", count)
	RemoveListener(key string) int

	// Reset 通过移除所有订阅者来重置信号，
	// 有效地清除订阅者列表。
	//
	// 当您想停止所有监听器接收更多信号时，可以使用此方法。
	//
	// 示例：
	// signal := signals.New[int]()
	// signal.AddListener(func(ctx context.Context, payload int) {
	//    // 监听器实现
	//    // ...
	// })
	// signal.Reset() // 移除所有监听器
	// fmt.Println("重置后的订阅者数量：", signal.Len())
	Reset()

	// Len 返回订阅信号的监听器数量。
	//
	// 这可用于检查当前有多少监听器在等待信号。
	// 返回值类型为 int。
	//
	// 示例：
	// signal := signals.NewSync[int]()
	// signal.AddListener(func(ctx context.Context, payload int) {
	//    // 监听器实现
	//    // ...
	// })
	// fmt.Println("订阅者数量：", signal.Len())
	Len() int

	// IsEmpty 检查信号是否有任何订阅者。
	//
	// 如果信号没有订阅者，则返回 true，否则返回 false。
	// 这可用于在发出信号之前检查是否有任何监听器。
	//
	// 示例：
	// signal := signals.New[int]()
	// fmt.Println("信号是否为空？", signal.IsEmpty()) // 应打印 true
	// signal.AddListener(func(ctx context.Context, payload int) {
	//    // 监听器实现
	//    // ...
	// })
	// fmt.Println("信号是否为空？", signal.IsEmpty()) // 应打印 false
	IsEmpty() bool
}
