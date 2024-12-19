package main

import (
	"database/sql"
	"log"
	"strconv"
	"time"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/auth"
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

	// get collection
	l.Info("providers in collection", e.Collection.Name, e.Collection.OAuth2.Providers)

	l.Info("googleOauthHandler",
		"providerName", e.ProviderName,
		"record", e.Record,
		"OAuth2User", e.OAuth2User)

	return e.Next()
}

func wrapFactory[T auth.Provider](factory func() T) auth.ProviderFactoryFunc {
	return func() auth.Provider {
		return factory()
	}
}

// Google allows authentication via Google OAuth2.
// type Google struct {
// 	auth.BaseProvider
// }

// // NewGoogleProvider creates new Google provider instance with some defaults.
// func NewGoogleProvider() *Google {

// 	return &Google{auth.BaseProvider{
// 		ctx:         context.Background(),
// 		displayName: "Google",
// 		pkce:        true,
// 		scopes: []string{
// 			"https://www.googleapis.com/auth/userinfo.profile",
// 			"https://www.googleapis.com/auth/userinfo.email",
// 		},
// 		authURL:     "https://accounts.google.com/o/oauth2/auth",
// 		tokenURL:    "https://accounts.google.com/o/oauth2/token",
// 		userInfoURL: "https://www.googleapis.com/oauth2/v1/userinfo",
// 	}}
// }

func init() {
	auth.Providers[auth.NameGoogle] = wrapFactory(NewPaleoGoogleProvider)

}

func main() {
	app := pocketbase.New()

	app.OnBootstrap().BindFunc(func(e *core.BootstrapEvent) error {
		if err := e.Next(); err != nil {
			return err
		}
		// get users collection
		collection, err := e.App.FindCollectionByNameOrId("users")
		if err != nil {
			log.Fatal(err)
		}

		// get oauth2
		// paleoGoogle := NewPaleoGoogleProvider()
		// auth.NameGoogle
		// auth.Providers[auth.NameGoogle] = func() auth.Provider {
		// 	return NewPaleoGoogleProvider()
		// }

		// g, ex := collection.OAuth2.GetProviderConfig(auth.NameGoogle)
		// if !ex {
		// 	e.App.Logger().Error("set google and then restart")
		// 	return nil
		// }
		// collection.OAuth2.Providers = append(collection.OAuth2.Providers, core.OAuth2ProviderConfig{
		// 	Name:         NamePaleoGoogle,
		// 	DisplayName:  NamePaleoGoogle,
		// 	ClientId:     g.ClientId,
		// 	ClientSecret: g.ClientSecret,
		// 	AuthURL:      g.AuthURL,
		// 	TokenURL:     g.TokenURL,
		// 	UserInfoURL:  g.UserInfoURL,
		// 	Extra:        g.Extra,
		// })

		// search for google oauth
		// for i, p := range collection.OAuth2.Providers {
		// 	if p.Name == auth.NameGoogle {
		// 		u, err := url.Parse(p.AuthURL)
		// 		if err != nil {
		// 			return err
		// 		}
		// 		v := u.Query()
		// 		v.Add("hd", "itispaleocapa.it")

		// 		u.RawQuery = v.Encode()
		// 		collection.OAuth2.Providers[i].AuthURL = u.String()
		// 		e.App.Logger().Info("google oauth should be updated")
		// 	}
		// }

		// paleo, exists := collection.OAuth2.GetProviderConfig(NamePaleoGoogle)
		// if !exists {
		// 	log.Fatal("something is wrong setting paleogoogle")
		// }
		e.App.Logger().Info("providers of collection", collection.OAuth2.Providers)
		// e.App.Logger().Info("paleo provider", paleo)
		// e.App.Logger().Info("auth provider validation, paleogoogle", paleo.Validate())
		return nil
	})

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
