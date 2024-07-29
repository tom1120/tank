package test

import (
	"testing"

	"github.com/robfig/cron"
)

func TestSecords(t *testing.T) {
	cronJob := cron.New()
	cronJob.AddFunc("* * * * * *", func() {
		t.Log("cron job run")

	})
	cronJob.Start()
	defer cronJob.Stop()
	select {}
}

func TestCronJob(t *testing.T) {
	cronJob := cron.New()
	cronJob.AddJob("* * * * * *", cron.FuncJob(func() {
		t.Log("cron job run")
	}))
	cronJob.Start()
	defer cronJob.Stop()
	select {}
}
