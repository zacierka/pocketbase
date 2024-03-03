package main

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/webhook"
	strava "github.com/strava/go.strava"
)

func getDiscordWebhookURL() string {
	channel := os.Getenv("DISCORD_WH_CHANNEL")
	id := os.Getenv("DISCORD_WH_ID")
	return fmt.Sprintf("https://discord.com/api/webhooks/%s/%s", channel, id)

}

func sendDiscordActivity(activity *strava.ActivityDetailed) {
	dwebhook, err := webhook.NewWithURL(getDiscordWebhookURL())
	if err != nil {
		fmt.Print(err)
	}
	defer dwebhook.Close(context.TODO())

	var wg sync.WaitGroup
	wg.Add(1)
	go send(&wg, dwebhook, activity)
	wg.Wait()
}

func send(wg *sync.WaitGroup, client webhook.Client, activity *strava.ActivityDetailed) {
	defer wg.Done()

	embed := discord.NewEmbedBuilder().
		SetTitle(activity.Name).
		SetDescription(activity.Description).
		SetURLf("https://www.strava.com/activities/%s", strconv.FormatInt(activity.Id, 10)).
		SetColor(0x37ff00).
		SetFields(
			discord.EmbedField{
				Name:   "Distance",
				Value:  strconv.FormatFloat(activity.Distance, 'f', -1, 64),
				Inline: boolPointer(true),
			},
			discord.EmbedField{
				Name:   "Time",
				Value:  strconv.Itoa(activity.ElapsedTime),
				Inline: boolPointer(true),
			},
			discord.EmbedField{
				Name:   "Pace",
				Value:  "UNK /mi",
				Inline: boolPointer(true),
			},
		).
		SetTimestamp(time.Now()).
		SetFooter("Strava Notifier", "").
		Build()

	if _, err := client.CreateMessage(discord.NewWebhookMessageCreateBuilder().
		SetEmbeds(embed).
		Build(),
	); err != nil {
		fmt.Println(err)
	}
}

func boolPointer(b bool) *bool {
	return &b
}
