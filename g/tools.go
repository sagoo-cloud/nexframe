package g

import "github.com/sagoo-cloud/nexframe/utils/dump"

// Dump 打印变量的详细信息。
func Dump(values ...interface{}) {
	dump.Dump(values...)
}
