package hub

import (
	"fmt"
	"morego/lib/robfig/cron"
)

// https://godoc.org/morego/lib/robfig/cron
func CronTest() {
	c := cron.New()
	c.AddFunc("0 30 * * * *", func() { fmt.Println("Every hour on the half hour") })
	c.AddFunc("@hourly", func() { fmt.Println("Every hour") })
	c.AddFunc("@every 1h30m", func() { fmt.Println("Every hour thirty") })
	c.Start()

	// Funcs are invoked in their own goroutine, asynchronously.

	// Funcs may also be added to a running Cron
	c.AddFunc("@daily", func() { fmt.Println("Every day") })

	// Inspect the cron job entries' next and previous run times.
	//do something

	c.Stop() // Stop the scheduler (does not stop any jobs already running).
	select {}
}