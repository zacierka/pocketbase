package main

import (
	"net/http"

	"github.com/labstack/echo/v5"
)

func GET_HOOK(c echo.Context) error {
	const VERIFY_TOKEN string = "STRAVA"
	mode := c.QueryParam("hub.mode")
	token := c.QueryParam("hub.verify_token")
	challenge := c.QueryParam("hub.challenge")
	if len(mode) == 0 {
		return c.String(http.StatusForbidden, "Invalid mode")
	}
	if len(token) == 0 {
		return c.String(http.StatusForbidden, "Invalid token")
	}
	if mode == "subscribe" && token == VERIFY_TOKEN {
		return c.JSON(http.StatusOK, map[string]string{"hub.challenge": challenge})
	} else {
		if token != VERIFY_TOKEN {
			return c.String(http.StatusForbidden, "Invalid token")
		}
		return c.String(http.StatusOK, "OKAY")
	}
}
