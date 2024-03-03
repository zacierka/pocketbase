package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/labstack/echo/v5"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
)

func main() {
	app := pocketbase.New()

	app.OnRecordAfterAuthWithOAuth2Request().Add(func(e *core.RecordAuthWithOAuth2Event) error {
		err := handleAuthRequest(app, e)
		if err != nil {
			return err
		}
		return nil
	})

	app.OnBeforeServe().Add(func(e *core.ServeEvent) error {

		e.Router.GET("/*", apis.StaticDirectoryHandler(os.DirFS("./pb_public"), true))

		e.Router.GET("/strava/webhook", GET_HOOK)

		e.Router.POST("/strava/webhook", func(c echo.Context) error {
			var stravaData StravaData
			if err := c.Bind(&stravaData); err != nil {
				log.Println("POST[/strava/webhook] Failed bind for StravaData")
				return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid JSON"})
			}
			if stravaData.ObjectType == "activity" && stravaData.AspectType == "create" {
				user, err := app.Dao().FindFirstRecordByData("users_strava", "stravaId", stravaData.OwnerId)
				if err != nil {
					log.Printf("POST[/strava/webhook] Failed to Fetch users_strava for id: %d\n", stravaData.OwnerId)
					return c.String(http.StatusOK, "EVENT_RECEIVED")
				}

				var expiry time.Time = user.GetDateTime("expiry").Time()
				if isExpired(expiry) {
					refreshToken(app, stravaData.OwnerId, user.GetString("refreshToken"))
				}

				var accessToken string = user.GetString("accessToken")
				activity := getStravaAthleteActivity(accessToken, int64(stravaData.ObjectId))
				//fmt.Println(resolveDiscordNameFromStravaID(app.Dao(), strconv.Itoa(stravaData.OwnerId)))
				if activity == nil {
					log.Printf("POST[/strava/webhook] Failed to populate Activity for id: %d\n", stravaData.OwnerId)
					return c.String(http.StatusOK, "EVENT_RECEIVED")
				}

				sendDiscordActivity(activity)
			}

			return c.String(http.StatusOK, "EVENT_RECEIVED")
		})

		return nil
	})

	if err := app.Start(); err != nil {
		log.Fatal(err)
	}
}

/*
func resolveDiscordNameFromStravaID(dao *daos.Dao, stravaId string) string {
	stravaUser, error := dao.FindFirstRecordByData("users_strava", "stravaId", stravaId)
	if error != nil {
		return ""
	}

	if errs := dao.ExpandRecord(stravaUser, []string{"user"}, nil); len(errs) > 0 {
		return ""
	}

	return stravaUser.ExpandedOne("user").GetString("discordId")
}
*/
