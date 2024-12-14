package main

import (
	"database/sql"
	"log"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/auth"
	"github.com/pocketbase/pocketbase/tools/security"
)

func main() {
	app := pocketbase.New()

	app.OnRecordAuthWithOAuth2Request("users").BindFunc(func(e *core.RecordAuthWithOAuth2RequestEvent) error {
		for _, provider := range e.Collection.OAuth2.Providers {
			if e.ProviderName == provider.Name {
				e.App.Logger().Debug("provider found", "name", e.ProviderName)
				p, err := provider.InitProvider()
				if err != nil {
					return err
				}
				e.App.Logger().Debug("provider initialized", "name", e.ProviderName)

				e.ProviderClient = p
				break
			}
		}

		p, err := auth.NewProviderByName(e.ProviderName)
		if err != nil {
			e.App.Logger().Error("provider not found", "name", e.ProviderName)
			return err
		}
		e.App.Logger().Debug("provider p", p.AuthURL())

		e.App.Logger().Debug("provider ", "name", e.ProviderName, "client", e.ProviderClient)

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

		collection, err := app.FindCollectionByNameOrId("users")
		if err != nil {
			return err
		}

		user, err := app.FindAuthRecordByEmail(collection, email)

		// e.IsNewRecord = false
		if err != nil {
			if err == sql.ErrNoRows {
				e.App.Logger().Warn("user not found")
				user = core.NewRecord(collection)
				user.SetEmail(email)
				user.SetVerified(true)
				user.SetPassword(security.RandomString(16))
				user.Set("studentid", e.OAuth2User.RawUser["matricola"])
				user.Set("name", e.OAuth2User.RawUser["nome"])
				user.Set("surname", e.OAuth2User.RawUser["cognome"])
				user.Set("class", info["classe"])

				user.Set("roles", "studente")
			} else {
				return err
			}
		} else {
			if user.Verified() {
				e.Record = user
				return e.Next()
			}

			user.SetVerified(true)
			user.Set("studentid", e.OAuth2User.RawUser["matricola"])
		}

		if err := app.Save(user); err != nil {
			return err
		}

		e.Record = user
		return e.Next()
	})

	if err := app.Start(); err != nil {
		log.Fatal(err)
	}
}
