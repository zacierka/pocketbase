package main

import (
	"fmt"
	"sync"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/webhook"
	strava "github.com/strava/go.strava"
)

func send(wg *sync.WaitGroup, client webhook.Client, activity *strava.ActivityDetailed) {
	defer wg.Done()

	if _, err := client.CreateMessage(discord.NewWebhookMessageCreateBuilder().
		SetContentf("test %s", activity.Name).
		Build(),
	); err != nil {
		fmt.Println("error")
	}
}
