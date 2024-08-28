package nexframe

import "github.com/sagoo-cloud/nexframe/nf"

func Server(name ...interface{}) *nf.APIFramework {
	return nf.NewAPIFramework()
}
