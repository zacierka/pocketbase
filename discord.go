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
	"github.com/pocketbase/pocketbase/daos"
	strava "github.com/strava/go.strava"
)

func getDiscordWebhookURL() string {
	channel := os.Getenv("DISCORD_WH_CHANNEL")
	id := os.Getenv("DISCORD_WH_ID")
	return fmt.Sprintf("https://discord.com/api/webhooks/%s/%s", channel, id)

}

func sendDiscordActivity(dao *daos.Dao, activity *strava.ActivityDetailed) {
	dwebhook, err := webhook.NewWithURL(getDiscordWebhookURL())
	if err != nil {
		fmt.Print(err)
	}
	defer dwebhook.Close(context.TODO())

	var wg sync.WaitGroup
	wg.Add(1)
	go send(&wg, dwebhook, activity, dao)
	wg.Wait()
}

func send(wg *sync.WaitGroup, client webhook.Client, activity *strava.ActivityDetailed, dao *daos.Dao) {
	defer wg.Done()

	username := resolveDiscordNameFromStravaID(dao, activity.Athlete.Id)
	dist := metersToMiles(activity.Distance)
	duration, _ := time.ParseDuration(fmt.Sprintf("%ds", activity.MovingTime))
	pace := calculatePace(activity.MovingTime, dist)

	embed := discord.NewEmbedBuilder().
		SetTitle(fmt.Sprintf("%s logged a %s", username, activity.Type.String())).
		//SetDescription(activity.Description).
		SetURLf("https://www.strava.com/activities/%s", strconv.FormatInt(activity.Id, 10)).
		SetColor(0x37ff00).
		SetThumbnail("https://pbs.twimg.com/profile_images/900411562250256384/ALkwa0jf_200x200.jpg").
		SetFields(
			discord.EmbedField{
				Name:   "Distance",
				Value:  fmt.Sprintf("%.2f", dist),
				Inline: boolPointer(true),
			},
			discord.EmbedField{
				Name:   "Time",
				Value:  duration.String(),
				Inline: boolPointer(true),
			},
			discord.EmbedField{
				Name:   "Pace",
				Value:  fmt.Sprintf("%s/mi", pace.String()),
				Inline: boolPointer(true),
			},
		).
		SetTimestamp(time.Now()).
		SetFooter("Strava Notifier", "https://pbs.twimg.com/profile_images/900411562250256384/ALkwa0jf_400x400.jpg").
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

func resolveDiscordNameFromStravaID(dao *daos.Dao, stravaId int64) string {

	stravaUser, error := dao.FindFirstRecordByData("users_strava", "stravaId", strconv.FormatInt(stravaId, 10))
	if error != nil {
		return "NA"
	}

	if errs := dao.ExpandRecord(stravaUser, []string{"user"}, nil); len(errs) > 0 {
		return "NA"
	}

	discordId := stravaUser.ExpandedOne("user").GetString("discordId")
	username, error := dao.FindFirstRecordByData("discord", "discordId", discordId)
	if error != nil {
		return "NA"
	}
	return username.GetString("name")
}
