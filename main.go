package main

import (
	"database/sql"
	"log"
	"strconv"
	"time"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/security"
	"github.com/pocketbase/pocketbase/tools/types"
)

func paleoidOauthHandler(app *pocketbase.PocketBase, e *core.RecordAuthWithOAuth2RequestEvent) error {
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

	matricola := strconv.Itoa(int(e.OAuth2User.RawUser["matricola"].(float64)))

	if err != nil {
		if err != sql.ErrNoRows {
			return err
		}

		e.App.Logger().Warn("user not found")
		user = core.NewRecord(collection)

		user.SetEmail(email)
		user.SetVerified(true)
		user.SetPassword(security.RandomString(16))
		user.Set("studentid", matricola)
		user.Set("name", e.OAuth2User.RawUser["nome"])
		user.Set("surname", e.OAuth2User.RawUser["cognome"])

		e.OAuth2User.Id = matricola + "-" + security.RandomString(10)
		e.OAuth2User.Email = email
		e.OAuth2User.Name = e.OAuth2User.RawUser["nome"].(string)
		e.OAuth2User.Username = e.OAuth2User.RawUser["cognome"].(string)
		e.OAuth2User.Expiry, _ = types.ParseDateTime(time.Now().Add(time.Hour))

		user.Set("class", info["classe"])
		user.Set("roles", "studente")
		// e.Record = user
	} else {
		if user.Verified() {
			e.Record = user
			e.OAuth2User.Id = matricola + "-" + security.RandomString(10)
			return apis.RecordAuthResponse(e.RequestEvent, user, core.MFAMethodOAuth2, e.OAuth2User)
		}
		e.OAuth2User.Id = matricola + "-" + security.RandomString(10)
		e.OAuth2User.Email = email
		e.OAuth2User.Name = user.Get("name").(string)
		e.OAuth2User.Username = user.Get("surname").(string)
		e.OAuth2User.Expiry, _ = types.ParseDateTime(time.Now().Add(time.Hour))

		user.SetVerified(true)
		user.Set("studentid", e.OAuth2User.RawUser["matricola"])

		if err := app.Save(user); err != nil {
			return err
		}

		return apis.RecordAuthResponse(e.RequestEvent, user, core.MFAMethodOAuth2, e.OAuth2User)
	}

	if err := app.Save(user); err != nil {
		return err
	}
	e.Record = user
	return e.Next()
}

func googleOauthHandler(app *pocketbase.PocketBase, e *core.RecordAuthWithOAuth2RequestEvent) error {
	l := app.Logger()
	l.Info("googleOauthHandler",
		"providerName", e.ProviderName,
		"record", e.Record,
		"OAuth2User", e.OAuth2User)

	return e.Next()
}
func main() {
	app := pocketbase.New()

	app.OnRecordAuthWithOAuth2Request("users").BindFunc(func(e *core.RecordAuthWithOAuth2RequestEvent) error {
		// return paleoidOauthHandler(app, e)
		return googleOauthHandler(app, e)
	})

	// app.OnMailerSend().BindFunc(func(e *core.MailerEvent) error {
	// 	l := app.Logger()
	// 	l.Info("MailerSend",
	// 		"from", e.Message.From,
	// 		"to", e.Message.To,
	// 		"subject", e.Message.Subject,
	// 		"body", e.Message.Text)
	// 	return nil
	// })

	if err := app.Start(); err != nil {
		log.Fatal(err)
	}
}
