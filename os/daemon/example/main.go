package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/sagoo-cloud/nexframe/os/daemon"
)

func main() {
	// 解析命令行参数
	svcFlag := flag.String("service", "", "控制系统服务（start, stop, restart, install, uninstall）")
	flag.Parse()

	// 创建服务管理器
	sm, err := daemon.NewServiceManager(
		"MyService",
		"My Service Display Name",
		"This is my service description",
	)
	if err != nil {
		log.Fatalf("创建服务管理器失败: %v", err)
	}

	// 如果指定了服务控制命令，执行相应操作
	if *svcFlag != "" {
		err := sm.Control(*svcFlag)
		if err != nil {
			log.Fatalf("控制服务失败: %v", err)
		}
		return
	}

	// 尝试作为守护进程运行
	if err := daemon.Daemonize(); err != nil {
		log.Fatalf("守护进程化失败: %v", err)
	}

	// 运行服务
	if daemon.IsDaemon() {
		if err := sm.Run(); err != nil {
			log.Fatalf("运行服务失败: %v", err)
		}
	} else {
		// 这里是你的主程序逻辑
		fmt.Println("服务开始运行...")
		for {
			fmt.Println("服务正在运行...")
			time.Sleep(10 * time.Second)
		}
	}
}
