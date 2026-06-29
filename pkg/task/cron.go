package task

import (
	"sync"

	"github.com/robfig/cron/v3"
)

var (
	c   *cron.Cron
	once sync.Once
)

func Init() {
	once.Do(func() {
		c = cron.New()
	})
}

func AddJob(spec string, cmd func()) (cron.EntryID, error) {
	Init()
	return c.AddFunc(spec, cmd)
}

func Start() {
	Init()
	c.Start()
}

func Stop() {
	if c != nil {
		c.Stop()
	}
}
