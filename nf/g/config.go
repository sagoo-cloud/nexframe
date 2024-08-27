package g

import "github.com/sagoo-cloud/nexframe/nf/configs"

func Cfg() *configs.ConfigEntity {
	return configs.GetInstance()
}
