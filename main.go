package main

import (
	"database/sql"
	"log"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

func main() {
	app := pocketbase.New()

	// fires for every auth collection
	// app.OnRecordAuthWithOAuth2Request().BindFunc(func(e *core.RecordAuthWithOAuth2RequestEvent) error {
	// 	// e.App
	// 	// e.Collection
	// 	// e.ProviderName
	// 	// e.ProviderClient
	// 	// e.Record (could be nil)
	// 	// e.OAuth2User
	// 	// e.CreateData
	// 	// e.IsNewRecord
	// 	// and all RequestEvent fields...

	// 	return e.Next()
	// })

	// fires only for "users" and "managers" auth collections
	app.OnRecordAuthWithOAuth2Request("users").BindFunc(func(e *core.RecordAuthWithOAuth2RequestEvent) error {
		// fmt.Println("new users oauth request")

		// e.App.Logger().Debug("new users oauth request")
		e.App.Logger().Debug("new users oauth request")
		// e.Collection
		e.App.Logger().Debug("provider name", e.ProviderName)
		e.App.Logger().Debug("record", e.Record)
		e.App.Logger().Debug("oauth2 user", e.OAuth2User)
		e.App.Logger().Debug("create data", e.CreateData)
		e.App.Logger().Debug("is new record", e.IsNewRecord)
		// e.ProviderName
		// e.ProviderClient
		// e.Record (could be nil)
		// e.OAuth2User
		// e.CreateData
		// e.IsNewRecord
		// and all RequestEvent fields...

		user, err := app.FindFirstRecordByData("users", "email", e.OAuth2User.RawUser["email"])

		if err != nil {
			if err == sql.ErrNoRows {
				e.App.Logger().Warn("user not found")
			}
		} else {
			e.App.Logger().Debug("user found", user)
		}
		return e.Next()
	})

	if err := app.Start(); err != nil {
		log.Fatal(err)
	}
}
