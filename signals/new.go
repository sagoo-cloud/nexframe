package signals

// NewSync 创建一个新的信号，可用于同步发射和监听事件。
//
// 示例:
//
//	signal := signals.NewSync[int]()
//	signal.AddListener(func(ctx context.Context, payload int) {
//	    // 监听器实现
//	    // ...
//	})
//	signal.Emit(context.Background(), 42)
func NewSync[T any]() Signal[T] {
	s := &SyncSignal[T]{
		BaseSignal: BaseSignal[T]{
			subscribers:    make([]keyedListener[T], 0),
			subscribersMap: make(map[string]SignalListener[T]),
		},
	}
	return s
}

// New 创建一个新的信号，可用于异步发射和监听事件。
//
// 示例:
//
//	signal := signals.New[int]()
//	signal.AddListener(func(ctx context.Context, payload int) {
//	    // 监听器实现
//	    // ...
//	})
//	signal.Emit(context.Background(), 42)
func New[T any]() Signal[T] {
	s := &AsyncSignal[T]{
		BaseSignal: BaseSignal[T]{
			subscribers:    make([]keyedListener[T], 0),
			subscribersMap: make(map[string]SignalListener[T]),
		},
	}
	return s
}
