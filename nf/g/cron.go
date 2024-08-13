package g

import (
	"github.com/robfig/cron/v3"
)

var cronEntity *cron.Cron

func init() {
	cronEntity = cron.New()

}
func Cron() *cron.Cron {
	return cronEntity
}
