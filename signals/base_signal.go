package signals

// keyedListener 表示监听器和用于标识的可选键的组合。
type keyedListener[T any] struct {
	key      string
	listener SignalListener[T]
}

// BaseSignal 提供 Signal 接口的基本实现。
// 它旨在用作底层信号机制的抽象基础。
//
// 示例:
//
//	type MyDerivedSignal[T any] struct {
//		BaseSignal[T]
//		// MyDerivedSignal 特有的其他字段或方法
//	}
//
//	func (s *MyDerivedSignal[T]) Emit(ctx context.Context, payload T) {
//		// 发射信号的自定义实现
//	}
type BaseSignal[T any] struct {
	subscribers    []keyedListener[T]
	subscribersMap map[string]SignalListener[T]
}

// AddListener 向信号添加监听器。每当信号被发射时，监听器将被调用。
// 它返回添加监听器后的订阅者数量。它接受一个可选的键，
// 该键可以用于以后移除监听器或检查监听器是否已经被添加。
// 如果具有相同键的监听器已经添加到信号中，则返回 -1。
//
// 示例:
//
//	signal := signals.New[int]()
//	count := signal.AddListener(func(ctx context.Context, payload int) {
//		// 监听器实现
//		// ...
//	}, "key1")
//	fmt.Println("添加监听器后的订阅者数量:", count)
func (s *BaseSignal[T]) AddListener(listener SignalListener[T], key ...string) int {
	if listener == nil {
		return -1 // 返回 -1 表示添加失败
	}

	if len(key) > 0 {
		if _, ok := s.subscribersMap[key[0]]; ok {
			return -1
		}
		s.subscribersMap[key[0]] = listener
		s.subscribers = append(s.subscribers, keyedListener[T]{
			key:      key[0],
			listener: listener,
		})
	} else {
		s.subscribers = append(s.subscribers, keyedListener[T]{
			listener: listener,
		})
	}

	return len(s.subscribers)
}

// RemoveListener 从信号中移除监听器。
// 它返回移除监听器后的订阅者数量。如果未找到监听器，则返回 -1。
//
// 示例:
//
//	signal := signals.New[int]()
//	signal.AddListener(func(ctx context.Context, payload int) {
//		// 监听器实现
//		// ...
//	}, "key1")
//	count := signal.RemoveListener("key1")
//	fmt.Println("移除监听器后的订阅者数量:", count)
func (s *BaseSignal[T]) RemoveListener(key string) int {
	if _, ok := s.subscribersMap[key]; ok {
		delete(s.subscribersMap, key)

		for i, sub := range s.subscribers {
			if sub.key == key {
				s.subscribers = append(s.subscribers[:i], s.subscribers[i+1:]...)
				break
			}
		}
		return len(s.subscribers)
	}

	return -1
}

// Reset 通过从信号中移除所有订阅者来重置信号，
// 有效地清除订阅者列表。
// 当你想停止所有监听器接收进一步的信号时，可以使用此方法。
//
// 示例:
//
//	signal := signals.New[int]()
//	signal.AddListener(func(ctx context.Context, payload int) {
//		// 监听器实现
//		// ...
//	})
//	signal.Reset() // 移除所有监听器
//	fmt.Println("重置后的订阅者数量:", signal.Len())
func (s *BaseSignal[T]) Reset() {
	s.subscribers = s.subscribers[:0] // 清空切片但保留容量
	clear(s.subscribersMap)           // 清空 map 但保留底层数据结构
}

// Len 返回订阅信号的监听器数量。
// 这可以用来检查当前有多少监听器在等待信号。
// 返回值类型为 int。
//
// 示例:
//
//	signal := signals.New[int]()
//	signal.AddListener(func(ctx context.Context, payload int) {
//		// 监听器实现
//		// ...
//	})
//	fmt.Println("订阅者数量:", signal.Len())
func (s *BaseSignal[T]) Len() int {
	return len(s.subscribers)
}

// IsEmpty 检查信号是否有任何订阅者。
// 如果信号没有订阅者，则返回 true，否则返回 false。
// 这可以用来在发射信号之前检查是否有任何监听器。
//
// 示例:
//
//	signal := signals.New[int]()
//	fmt.Println("信号是否为空?", signal.IsEmpty()) // 应打印 true
//	signal.AddListener(func(ctx context.Context, payload int) {
//		// 监听器实现
//		// ...
//	})
//	fmt.Println("信号是否为空?", signal.IsEmpty()) // 应打印 false
func (s *BaseSignal[T]) IsEmpty() bool {
	return len(s.subscribers) == 0
}

// Emit 在 BaseSignal 中未实现，如果被调用会返回一个错误。
// 它应该由派生类型实现。
//
// 示例:
//
//	type MyDerivedSignal[T any] struct {
//		BaseSignal[T]
//		// MyDerivedSignal 特有的其他字段或方法
//	}
//
//	func (s *MyDerivedSignal[T]) Emit(ctx context.Context, payload T) error {
//		// 发射信号的自定义实现
//	}
