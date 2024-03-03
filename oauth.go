package main

import (
	"database/sql"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"os"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/daos"
	"github.com/pocketbase/pocketbase/forms"
	"github.com/pocketbase/pocketbase/models"
)

func handleAuthRequest(app *pocketbase.PocketBase, e *core.RecordAuthWithOAuth2Event) error {
	_, err := app.Dao().FindFirstRecordByData("users_strava", "stravaId", e.OAuth2User.RawUser["id"])
	if err != nil {
		if err == sql.ErrNoRows { // new user
			// issue here when you re-register. Calls again for some reason. Unique constraint set for workaround
			// maybe the field isnt a string. maybe its fetching a float for id
			res := handleNewUser(app.Dao(), e)
			if res != nil {
				return err
			}
			return nil
		}
	}
	// existing user? do i care?
	//refreshToken(app, user.GetInt("stravaId"), user.GetString("refreshToken"))
	return nil
}

func handleNewUser(dao *daos.Dao, e *core.RecordAuthWithOAuth2Event) error { // create new extry in users_strava
	collection, err := dao.FindCollectionByNameOrId("users_strava")
	if err != nil {
		return err
	}

	record := models.NewRecord(collection)
	record.Set("user", e.Record.Id)
	record.Set("stravaId", e.OAuth2User.RawUser["id"])
	record.Set("accessToken", e.OAuth2User.AccessToken)
	record.Set("refreshToken", e.OAuth2User.RefreshToken)
	record.Set("expiry", e.OAuth2User.Expiry)
	record.Set("refreshToken", e.OAuth2User.RefreshToken)
	record.Set("rawUser", e.OAuth2User.RawUser)

	if err := dao.SaveRecord(record); err != nil {
		return err
	}
	return nil
}

// func handleExistingUser(record *models.Record, e *core.RecordAuthWithOAuth2Event, dao *daos.Dao) error {
// 	// set individual fields
// 	// or bulk load with record.Load(map[string]any{...})
// 	record.Set("user", e.Record.Id)
// 	record.Set("stravaId", e.OAuth2User.RawUser["id"])
// 	record.Set("accessToken", e.OAuth2User.AccessToken)
// 	record.Set("refreshToken", e.OAuth2User.RefreshToken)
// 	record.Set("expiry", e.OAuth2User.Expiry)
// 	record.Set("refreshToken", e.OAuth2User.RefreshToken)
// 	record.Set("rawUser", e.OAuth2User.RawUser)

// 	if err := dao.SaveRecord(record); err != nil {
// 		return err
// 	}
// 	return nil
// }

func refreshToken(app *pocketbase.PocketBase, stravaId int, refreshToken string) error {
	resp, err := http.PostForm("https://www.strava.com/api/v3/oauth/token",
		url.Values{"client_id": {os.Getenv("STRAVA_CLIENT_ID")}, "client_secret": {os.Getenv("STRAVA_CLIENT_SECRET")}, "grant_type": {"refresh_token"}, "refresh_token": {refreshToken}})
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

	// update data for user [accessToken, refreshToken, Expires<At|In>]
	user, err := app.Dao().FindFirstRecordByData("users_strava", "stravaId", stravaId)
	if err != nil {
		return err
	}
	form := forms.NewRecordUpsert(app, user)

	form.LoadData(map[string]any{
		"accessToken":  refreshData.AccessToken,
		"refreshToken": refreshData.RefreshToken,
		"expiry":       refreshData.ExpiresAt,
	})

	if err := form.Submit(); err != nil {
		return err
	}

	return nil
}
