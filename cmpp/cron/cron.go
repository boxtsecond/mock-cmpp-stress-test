package cron

import (
	"github.com/jasonlvhit/gocron"

)
func Start()  {
	UpdateAccountCache()

	go func() {
		s := gocron.NewScheduler()
		s.Every(5).Minutes().Do(UpdateAccountCache)
		<-s.Start()
	}()
}