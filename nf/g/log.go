package g

import (
	"github.com/sagoo-cloud/nexframe/os/zlog"
)

var logger zlog.Logger

func init() {
	logger = zlog.GetLogger()
}
func Log(name ...string) zlog.Logger {
	return logger
}
