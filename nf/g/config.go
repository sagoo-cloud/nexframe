package g

import "github.com/sagoo-cloud/nexframe/configs"

func Cfg() *configs.ConfigEntity {
	return configs.GetInstance()
}
