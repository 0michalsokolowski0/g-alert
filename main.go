// Application which greets you.
package main

import (
	"context"
	"fmt"
	"os/exec"
	"time"

	"github.com/go-co-op/gocron"
)

func main() {
	// TODO: move it to env var to make it configurable
	scheduler := gocron.NewScheduler(time.Local)

	// TODO: cron expression should be passed as env var
	_, err := scheduler.Cron("*/1 * * * *").Do(func() {
		output, err := exec.CommandContext(context.Background(), "googler", "--json", "--np", "--tld pl", "--exact 'search_phrase'").Output()
		if err != nil {
			panic(err)
		}
		o := string(output)
		fmt.Println(o)
	})
	if err != nil {
		panic(err)
	}

	scheduler.StartBlocking()
}
