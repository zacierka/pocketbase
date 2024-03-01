package main

import (
	"flag"
	"log"
	"net/http"
	"os"

	"github.com/labstack/echo/v5"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
)

var (
	stravaClientId     string
	stravaClientSecret string
)

func main() {
	flag.StringVar(&stravaClientId, "c", "", "Strava client id")
	flag.StringVar(&stravaClientSecret, "s", "", "Strava client secret")
	flag.Parse()

	app := pocketbase.New()

	app.OnBeforeServe().Add(func(e *core.ServeEvent) error {

		e.Router.GET("/*", apis.StaticDirectoryHandler(os.DirFS("./pb_public"), true))

		e.Router.GET("/exchange_token", func(c echo.Context) error {
			code := c.QueryParam("code")
			scope := c.QueryParam("scope")

			denied := c.QueryParam("error")
			if denied != "" {
				return c.String(http.StatusForbidden, "access_denied")
			}

			token := new(TokenExchange)
			if error := exchangeToken(code, scope, token); error != nil {
				return c.String(http.StatusForbidden, error.Error())
			}

			// store token
			go storeAuthData(app, token)

			return c.String(http.StatusOK, "Success")
		})

		return nil
	})

	if err := app.Start(); err != nil {
		log.Fatal(err)
	}
}
