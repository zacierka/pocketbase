package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	strava "github.com/strava/go.strava"
)

func FetchAthleteActivityByID(stravaService strava.Client, objectID int, ownerID int) {
	fmt.Printf("Getting activity for %d", ownerID)

	url := "https://www.strava.com/api/v3/activities/{id}" + fmt.Sprintf("%s", objectID)

	authHeader := "Bearer " + k.GetStravaAccessToken()

	// Build request; include authHeader
	req, err := http.NewRequest("GET", url, nil)
	req.Header.Add("Authorization", authHeader)
	if err != nil {
		log.Fatal(err)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	activityList := ActivityList{}

	err = json.NewDecoder(resp.Body).Decode(&activityList)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Got " + fmt.Sprint(len(activityList)) + " activities for " + k.FullName())

	return activityList
}
