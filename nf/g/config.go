package g

import "github.com/sagoo-cloud/nexframe/nf/config"

func Cfg() *config.ConfigEntity {
	return config.GetInstance()
}
