package main

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/knadh/listmonk/internal/subimporter"
	"github.com/knadh/listmonk/models"
	"github.com/labstack/echo"
)

// handleBounceWebhook renders the HTML preview of a template.
func handleBounceWebhook(c echo.Context) error {
	var (
		app     = c.Get("app").(*App)
		service = c.Param("id")

		b models.Bounce
	)

	switch service {
	// Native postback.
	case "":
		if err := c.Bind(&b); err != nil {
			return err
		}

		if err := validateBounceFields(b, app); err != nil {
			return err
		}

		b.Email = strings.ToLower(b.Email)

		if len(b.Meta) == 0 {
			b.Meta = json.RawMessage("{}")
		}

		if b.CreatedAt.Year() == 0 {
			b.CreatedAt = time.Now()
		}

	// Amazon SES.
	case "ses":

	// SendGrid.
	case "sendgrid":
	default:
		return echo.NewHTTPError(http.StatusBadRequest, app.i18n.Ts("bounce.unknownService"))
	}

	// Record the bounce.
	if err := app.bounce.Record(b); err != nil {
		app.log.Printf("error recording bounce: %v", err)
	}

	return c.JSON(http.StatusOK, okResp{true})
}

func validateBounceFields(b models.Bounce, app *App) error {
	if b.Email == "" && b.SubscriberUUID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, app.i18n.T("globals.messages.invalidData"))
	}

	if b.Email != "" && !subimporter.IsEmail(b.Email) {
		return echo.NewHTTPError(http.StatusBadRequest, app.i18n.T("globals.messages.invalidEmail"))
	}

	if b.SubscriberUUID != "" && !reUUID.MatchString(b.SubscriberUUID) {
		return echo.NewHTTPError(http.StatusBadRequest, app.i18n.T("globals.messages.invalidUUID"))
	}

	return nil
}
