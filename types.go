package main

import strava "github.com/strava/go.strava"

/* Strava Types */

// Strava Data callback struct
type StravaData struct {
	ObjectType string                 `json:"object_type"`     // "athlete" or "activity"
	ObjectId   int                    `json:"object_id"`       // id of athlete or activity
	AspectType string                 `json:"aspect_type"`     // "create" "update" "delete"
	OwnerId    int                    `json:"owner_id"`        // ID of the athlete who owns the event
	EventTime  int                    `json:"event_time"`      // epoch
	SubID      int                    `json:"subscription_id"` // subscription id
	X          map[string]interface{} `json:"-"`
}

/* OAUTH Types */

// Refresh token
type RefreshToken struct {
	TokenType    string `json:"Bearer"`        // bearer
	AccessToken  string `json:"access_token"`  // access token
	ExpiresAt    string `json:"expires_at"`    // expires at
	ExpiresIn    string `json:"expires_in"`    // expires in
	RefreshToken string `json:"refresh_token"` // refresh token
}

type TokenExchange struct {
	AuthData RefreshToken
	athlete  strava.AthleteDetailed
}
