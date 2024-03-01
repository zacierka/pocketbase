package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/forms"
)

func storeAuthData(app *pocketbase.PocketBase, tokenData *TokenExchange) error {
	// check if user exists
	record, err := app.Dao().FindFirstRecordByData("users", "stravaId", string(tokenData.athlete.Id))
	if err != nil {
		return fmt.Errorf("Unable to fetch User") // is this a new user or error fetching
	}
	// if user exists - update?
	form := forms.NewRecordUpsert(app, record)

	// or form.LoadRequest(r, "")
	form.LoadData(map[string]any{ // id, blah, blah, blah
		"title":          "Lorem ipsum",
		"active":         true,
		"someOtherField": 123,
	})

	// validate and submit (internally it calls app.Dao().SaveRecord(record) in a transaction)
	if err := form.Submit(); err != nil {
		return err
	}

	// if user is new - store

	// store user here

	return nil

}

func validateExchangeParams(code string, scope string) error {
	var err error = fmt.Errorf("INVALID")

	if code == "" || scope != "" {
		return err
	}

	return nil
}

func exchangeToken(code string, scope string, tokenData *TokenExchange) error {

	if validateExchangeParams(code, scope) != nil {
		return fmt.Errorf("Invalid Parameters")
	}
	// handle code and scope
	resp, err := http.PostForm("https://www.strava.com/api/v3/oauth/token",
		url.Values{"client_id": {stravaClientId}, "client_secret": {stravaClientSecret}, "code": {code}, "grant_type": {"authorization_code"}})
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(body, &tokenData); err != nil {
		return err
	}

	fmt.Printf("%+v\n", &tokenData)

	return nil
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
