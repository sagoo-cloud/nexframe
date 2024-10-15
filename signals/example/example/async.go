package example

import (
	"context"
	"fmt"
)

func RunAsync() {

	// 为 RecordCreatedAsync 信号添加一个监听器
	RecordCreated.AddListener(func(ctx context.Context, record Record) {
		fmt.Println("记录已创建:", record)
	}, "key1") // <- 键是可选的，用于之后移除监听器

	// 为 RecordUpdatedAsync 信号添加一个监听器
	RecordUpdated.AddListener(func(ctx context.Context, record Record) {
		fmt.Println("记录已更新:", record)
	})

	// 为 RecordDeleted 信号添加一个监听器
	RecordDeleted.AddListener(func(ctx context.Context, record Record) {
		fmt.Println("记录已删除:", record)
	})

	ctx := context.Background()

	// 发射 RecordCreatedAsync 信号
	RecordCreated.Emit(ctx, Record{ID: 3, Name: "Record C"})

	// 发射 RecordUpdated 信号
	RecordUpdated.Emit(ctx, Record{ID: 2, Name: "Record B"})

	// 发射 RecordDeleted 信号
	RecordDeleted.Emit(ctx, Record{ID: 1, Name: "Record A"})
}
