package main

import (
	"fmt"
	"math"
	"time"

	strava "github.com/strava/go.strava"
)

type StravaData struct {
	ObjectType string                 `json:"object_type"`     // "athlete" or "activity"
	ObjectId   int                    `json:"object_id"`       // id of athlete or activity
	AspectType string                 `json:"aspect_type"`     // "create" "update" "delete"
	OwnerId    int                    `json:"owner_id"`        // ID of the athlete who owns the event
	EventTime  int                    `json:"event_time"`      // epoch
	SubID      int                    `json:"subscription_id"` // subscription id
	X          map[string]interface{} `json:"-"`
}

type RefreshToken struct {
	TokenType    string `json:"Bearer"`        // bearer
	AccessToken  string `json:"access_token"`  // access token
	ExpiresAt    string `json:"expires_at"`    // expires at
	ExpiresIn    string `json:"expires_in"`    // expires in
	RefreshToken string `json:"refresh_token"` // refresh token
}

func isExpired(expiry time.Time) bool {
	var ret bool = false
	if time.Now().After(expiry) {
		ret = true
	}
	return ret
}

func getStravaAthleteActivity(token string, id int64) *strava.ActivityDetailed {
	client := strava.NewClient(token)
	activity, err := strava.NewActivitiesService(client).Get(int64(id)).Do()
	if err != nil {
		fmt.Println("bad fetch")
		fmt.Println(err)
		return nil
	}
	return activity

}

func calculatePace(seconds int, distance float64) time.Duration {
	raw_pace := float64(seconds) / distance // pace/mi in seconds
	min := int(raw_pace / 60.00)            // seconds to mins
	rem := int(math.Mod(raw_pace, 1) * 60)  // remainder to seconds

	var pace, error = time.ParseDuration(fmt.Sprintf("%dm%ds", min, rem))
	if error != nil {
		pace = time.Duration(0)
	}

	return pace

}
