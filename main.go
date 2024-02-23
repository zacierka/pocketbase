package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/labstack/echo/v5"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/models"
)

type PostObject struct {
	UserId  string `json:"userId"`
	Content string `json:"content"`
}

func main() {
	app := pocketbase.New()

	// serves static files from the provided public dir (if exists)
	app.OnBeforeServe().Add(func(e *core.ServeEvent) error {
		e.Router.GET("/*", apis.StaticDirectoryHandler(os.DirFS("./pb_public"), false))
		return nil
	})

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

		return nil
	})

	/* Strava Subscription Handler */
	app.OnBeforeServe().Add(func(e *core.ServeEvent) error {
		e.Router.GET("/strava/webhook", func(c echo.Context) error {

			const VERIFY_TOKEN string = "STRAVA"
			// Parses the query params
			mode := c.PathParam("hub.mode")
			token := c.PathParam("hub.verify_token:")
			challenge := c.PathParam("hub.challenge")
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
				// Responds with '403 Forbidden' if verify tokens do not match
				return c.String(http.StatusForbidden, "Invalid token")
			}
		})
		return nil
	})

	/* Strava Data Callback  */
	app.OnBeforeServe().Add(func(e *core.ServeEvent) error {
		e.Router.POST("/oauth/strava", func(c echo.Context) error {
			/* Do some funky unmarshalling here and see what we get so we can store it */
			fmt.Println("Got Data back! But what is it")
			return c.String(http.StatusOK, "EVENT_RECEIVED")
		})

		return nil
	})

	if err := app.Start(); err != nil {
		log.Fatal(err)
	}
}
