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
		// e.App.Logger().Debug("new users oauth request")
		// // e.Collection
		// e.App.Logger().Debug("provider name", e.ProviderName)
		// e.App.Logger().Debug("record", e.Record)
		// e.App.Logger().Debug("oauth2 user", e.OAuth2User)
		// e.App.Logger().Debug("create data", e.CreateData)
		// e.App.Logger().Debug("is new record", e.IsNewRecord)
		// e.ProviderName
		// e.ProviderClient
		// e.Record (could be nil)
		// e.OAuth2User
		// e.CreateData
		// e.IsNewRecord
		// and all RequestEvent fields...

		emailAny, ok := e.OAuth2User.RawUser["email"]
		if !ok {
			e.App.Logger().Error("email not found")
			return e.Next()
		}
		email := emailAny.(string)

		info := e.OAuth2User.RawUser["info_studente"].(map[string]interface{})
		// classe := gjson.Get(info, "classe")
		// e.App.Logger().Debug("info utente", info)
		// e.App.Logger().Debug("info utente classe", info["classe"])

		collection, err := app.FindCollectionByNameOrId("users")
		if err != nil {
			return err
		}

		user, err := app.FindAuthRecordByEmail(collection, email)
		if user.Verified() {
			// e.App.Logger().Debug("user already verified", user)
			return e.Next()
		}

		e.IsNewRecord = false
		if err != nil {
			if err == sql.ErrNoRows {
				e.App.Logger().Warn("user not found")
				user = core.NewRecord(collection)
				user.SetEmail(email)
				user.SetVerified(true)
				user.Set("studentid", e.OAuth2User.RawUser["matricola"])
				user.Set("name", e.OAuth2User.RawUser["nome"])
				user.Set("surname", e.OAuth2User.RawUser["cognome"])
				user.Set("class", info["classe"])

				user.Set("roles", "studente")

				e.App.Logger().Debug("new user", user)
				err := app.Save(user)
				if err != nil {
					return err
				}
			}
		} else {
			// e.App.Logger().Debug("user found", user)
			user.SetVerified(true)
			user.Set("studentid", e.OAuth2User.RawUser["matricola"])
			// e.App.Logger().Debug("user updated", user)

			err := app.Save(user)
			if err != nil {
				return err
			}
		}
		return e.Next()
	})

	if err := app.Start(); err != nil {
		log.Fatal(err)
	}
}
