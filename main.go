package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/labstack/echo/v5"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/models"
)

type PostObject struct {
	UserId  string `json:"userId"`
	Content string `json:"content"`
}

type StravaData struct {
	ObjectType string                 `json:"object_type"`     // "athlete" or "activity"
	ObjectId   int                    `json:"object_id"`       // id of athlete or activity
	AspectType string                 `json:"aspect_type"`     // "create" "update" "delete"
	OwnerId    int                    `json:"owner_id"`        // ID of the athlete who owns the event
	EventTime  int                    `json:"event_time"`      // epoch
	SubID      int                    `json:"subscription_id"` // subscription id
	X          map[string]interface{} `json:"-"`
}

func main() {
	app := pocketbase.New()

	// api/blog/post
	// accept posts from rest endpoint
	app.OnBeforeServe().Add(func(e *core.ServeEvent) error {
		e.Router.POST("api/blog/post", func(c echo.Context) error {
			const collectionName string = "posts"
			var post PostObject
			if err := c.Bind(&post); err != nil {
				return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid JSON"})
			}

			// possible validation filters here

			userrecord, err := app.Dao().FindFirstRecordByData("users", "discordId", post.UserId)
			if err != nil {
				app.Logger().Warn("Could not find existing record with discordId: " + post.UserId)
				return err
			}

			collection, err := app.Dao().FindCollectionByNameOrId(collectionName)
			if err != nil {
				app.Logger().Warn("Could not find collection " + collectionName)
				return err
			}

			record := models.NewRecord(collection)
			record.Set("user", userrecord.Id)
			record.Set("content", post.Content)

			if err := app.Dao().SaveRecord(record); err != nil {
				return err
			}
			app.Logger().Info("Created posts record for user " + userrecord.Id)

			return c.JSON(http.StatusOK, map[string]interface{}{"message": "Created New Record for " + userrecord.GetString("discordId")})
		} /* optional middlewares */)

		e.Router.GET("/strava/webhook", func(c echo.Context) error {

			const VERIFY_TOKEN string = "STRAVA"
			// Parses the query params
			mode := c.QueryParam("hub.mode")
			token := c.QueryParam("hub.verify_token")
			challenge := c.QueryParam("hub.challenge")
			fmt.Printf("%s mode, %s token, %s challenge", mode, token, challenge)
			// Checks if a token and mode is in the query string of the request
			if len(mode) == 0 {
				return c.String(http.StatusForbidden, "Invalid mode")
			}
			if len(token) == 0 {
				return c.String(http.StatusForbidden, "Invalid token")
			}
			// Verifies that the mode and token sent are valid
			if mode == "subscribe" && token == VERIFY_TOKEN {
				// Responds with the challenge token from the request
				fmt.Println("WEBHOOK_VERIFIED")
				return c.JSON(http.StatusOK, map[string]string{"hub.challenge": challenge})
			} else {
				fmt.Println(mode)
				if token != VERIFY_TOKEN {
					return c.String(http.StatusForbidden, "Invalid token")
				}
				return c.String(http.StatusOK, "OKAY")
			}
		})

		e.Router.POST("/strava/webhook", func(c echo.Context) error {
			/* Do some funky unmarshalling here and see what we get so we can store it */
			fmt.Println("Got Data back! But what is it")
			var stravaData StravaData
			if err := c.Bind(&stravaData); err != nil {
				return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid JSON"})
			}

			fmt.Printf("%+v", stravaData)
			if stravaData.ObjectType == "activity" && stravaData.AspectType == "create" {
				// new post to strava
				// findout who this is from and then fetch some stats for this item entry
				// https://www.strava.com/api/v3/activities/{id}?include_all_efforts

			}

			return c.String(http.StatusOK, "EVENT_RECEIVED")
		})

		return nil
	})

	if err := app.Start(); err != nil {
		log.Fatal(err)
	}
}
