package args

import (
	"flag"
	"testing"
)

var (
	Name   string
	Mode   string
	Server string
	Config string
	Cmd    string
	Args   string
)

func init() {

	flag.StringVar(&Name, "name", "app", "服务名称")
	flag.StringVar(&Mode, "mode", "dev", "开发模式")

	flag.StringVar(&Server, "servers", "http,event", "需要启动的服务器")
	flag.StringVar(&Config, "configs", "env,consul", "顺序环境配置")

	flag.StringVar(&Cmd, "cmd", "cmd", "cli命令")
	flag.StringVar(&Args, "args", "{}", "json参数")

	var _ = func() bool {
		testing.Init()
		return true
	}()
	if !flag.Parsed() {
		flag.Parse()
	}
}
