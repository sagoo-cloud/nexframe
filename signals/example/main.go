package main

import (
	"os"
	"time"

	"github.com/sagoo-cloud/nexframe/signals/example/example"
)

func main() {

	// 如果第一个参数是 "-sync"，则运行同步示例
	if len(os.Args) > 1 && os.Args[1] == "-sync" {
		example.RunSync()
	} else {
		example.RunAsync()
	}

	// 等待一秒钟，让信号能够被处理
	time.Sleep(time.Second)
}
