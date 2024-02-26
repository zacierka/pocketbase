package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"time"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/webhook"
	"github.com/labstack/echo/v5"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/models"
	strava "github.com/strava/go.strava"
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

type RefreshToken struct {
	TokenType    string `json:"Bearer"`        // bearer
	AccessToken  string `json:"access_token"`  // access token
	ExpiresAt    string `json:"expires_at"`    // expires at
	ExpiresIn    string `json:"expires_in"`    // expires in
	RefreshToken string `json:"refresh_token"` // refresh token
}

var (
	stravaClientId     string
	stravaClientSecret string
)

func main() {
	flag.StringVar(&stravaClientId, "c", "", "Strava client id")
	flag.StringVar(&stravaClientSecret, "s", "", "Strava client secret")
	flag.Parse()

	app := pocketbase.New()

	app.OnRecordAfterAuthWithOAuth2Request().Add(func(e *core.RecordAuthWithOAuth2Event) error {
		collection, err := app.Dao().FindCollectionByNameOrId("users_strava")
		if err != nil {
			return err
		}

		record := models.NewRecord(collection)

		user, err := app.Dao().FindFirstRecordByData("users", "username", e.Record.Username())
		if err != nil {
			return err
		}
		if user != nil {
			record.Set("user", e.Record.Id)
			record.Set("stravaId", e.OAuth2User.RawUser["id"])
			record.Set("accessToken", e.OAuth2User.AccessToken)
			record.Set("refreshToken", e.OAuth2User.RefreshToken)
			record.Set("expiry", e.OAuth2User.Expiry)
			record.Set("refreshToken", e.OAuth2User.RefreshToken)
			record.Set("rawUser", e.OAuth2User.RawUser)
			if err := app.Dao().SaveRecord(record); err != nil {
				return err
			}
		}

		return nil
	})

	app.OnBeforeServe().Add(func(e *core.ServeEvent) error {

		e.Router.GET("/test", func(c echo.Context) error {
			user, err := app.Dao().FindFirstRecordByData("users_strava", "stravaId", "")
			if err != nil {
				return err
			}
			refreshToken(0, user.GetString("refreshToken"))

			return nil

		})

		e.Router.GET("/strava/webhook", func(c echo.Context) error {
			const VERIFY_TOKEN string = "STRAVA"
			mode := c.QueryParam("hub.mode")
			token := c.QueryParam("hub.verify_token")
			challenge := c.QueryParam("hub.challenge")
			fmt.Printf("%s mode, %s token, %s challenge", mode, token, challenge)
			if len(mode) == 0 {
				return c.String(http.StatusForbidden, "Invalid mode")
			}
			if len(token) == 0 {
				return c.String(http.StatusForbidden, "Invalid token")
			}
			if mode == "subscribe" && token == VERIFY_TOKEN {
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
			var stravaData StravaData
			if err := c.Bind(&stravaData); err != nil {
				return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid JSON"})
			}
			if stravaData.ObjectType == "activity" && stravaData.AspectType == "create" {
				user, err := app.Dao().FindFirstRecordByData("users_strava", "stravaId", stravaData.OwnerId)
				if err != nil {
					return err
				}

				// is token expired?
				var expiry time.Time = user.GetDateTime("expiry").Time()
				if expiry.Compare(time.Now()) == -1 {
					refreshToken(stravaData.OwnerId, user.GetString("refreshToken"))
				}

				var accessToken string = user.GetString("accessToken")

				client := strava.NewClient(accessToken)
				activity, err := strava.NewActivitiesService(client).Get(int64(stravaData.ObjectId)).Do()
				if err != nil {
					fmt.Println(err)
				}

				dwebhook, err := webhook.NewWithURL("")
				if err != nil {
					fmt.Print(err)
				}
				defer dwebhook.Close(context.TODO())

				var wg sync.WaitGroup
				wg.Add(1)
				go send(&wg, dwebhook, activity)
				wg.Wait()
			}

			return c.String(http.StatusOK, "EVENT_RECEIVED")
		})

		return nil
	})

	if err := app.Start(); err != nil {
		log.Fatal(err)
	}
}

func send(wg *sync.WaitGroup, client webhook.Client, activity *strava.ActivityDetailed) {
	defer wg.Done()

	if _, err := client.CreateMessage(discord.NewWebhookMessageCreateBuilder().
		SetContentf("test %s", activity.Name).
		Build(),
	); err != nil {
		fmt.Println("error")
	}
}

func refreshToken(stravaId int, refreshToken string) error {
	// post to /oauth/token 6hr exp

	id := strconv.Itoa(stravaId)
	fmt.Printf("%s %s \n", id, refreshToken)
	resp, err := http.PostForm("https://www.strava.com/api/v3/oauth/token",
		url.Values{"client_id": {stravaClientId}, "client_secret": {stravaClientSecret}, "grant_type": {"refresh_token"}, "refresh_token": {refreshToken}})
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	refreshData := RefreshToken{}

	if err := json.Unmarshal(body, &refreshData); err != nil {
		return err
	}

	fmt.Printf("%+v\n", refreshData)

	// update data for user [accessToken, refreshToken, Expires<At|In>]

	return nil
}
