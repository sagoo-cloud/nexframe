package nf

import (
	"fmt"
	"github.com/arl/statsviz"
	"log"
	"net/http"
	"runtime"
	"time"
)

func (f *APIFramework) EnableStatsviz() {

	// 开启性能分析
	runtime.SetMutexProfileFraction(1) // (非必需)开启对锁调用的跟踪
	runtime.SetBlockProfileRate(1)     // (非必需)开启对阻塞操作的跟踪
	// 将Go程序运行时的各种内部数据进行可视化的展示，如可以展示：堆、对象、协程、GC等信息
	err := statsviz.RegisterDefault()
	if err == nil {
		log.Printf(fmt.Sprintf("System performance analysis browser to http://localhost%s/debug/statsviz/", f.config.StatsVizPort))

		go func() {
			defer func() {
				if err := recover(); err != nil {
					log.Printf("panic: %+v", err)
				}
			}()

			s := &http.Server{
				Addr:         f.config.StatsVizPort,
				ReadTimeout:  30 * time.Second,
				WriteTimeout: 30 * time.Second,
				IdleTimeout:  30 * time.Second,
			}
			if err := s.ListenAndServe(); err != nil {
				fmt.Println(err)
			}
		}()
	}
}
