package main

import (
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

	if err := app.Start(); err != nil {
		log.Fatal(err)
	}
}
