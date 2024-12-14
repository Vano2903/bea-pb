package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"time"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/auth"
	"github.com/pocketbase/pocketbase/tools/security"
	"github.com/pocketbase/pocketbase/tools/types"
	"golang.org/x/oauth2"
)

func init() {
	// auth.Providers[NamePaleoid] = wrapFactory(NewGithubProvider)
}

const NamePaleoid string = "paleoid"

var _ auth.Provider = (*Paleoid)(nil)

type Paleoid struct {
	auth.BaseProvider
}

func NewPaleoidProvider() *Paleoid {
	p := new(Paleoid)

	p.SetDisplayName("PaleoID")
	p.SetAuthURL("https://id.paleo.bg.it/oauth/authorize")
	p.SetTokenURL("https://id.paleo.bg.it/oauth/token")
	p.SetUserInfoURL("https://id.paleo.bg.it/api/v2/user")
	return p
}

func (p Paleoid) FetchAuthUser(token *oauth2.Token) (*auth.AuthUser, error) {
	data, err := p.FetchRawUserInfo(token)
	if err != nil {
		return nil, err
	}

	rawUser := map[string]any{}
	if err := json.Unmarshal(data, &rawUser); err != nil {
		return nil, err
	}

	extracted := struct {
		Name      string `json:"nome"`
		LastName  string `json:"cognome"`
		Email     string `json:"email"`
		StudentID string `json:"matricola"`
		// AvatarURL string `json:"avatar_url"`
		// info struct {
		// 	Classe string `json:"classe"`
		// } `json:"info_studente"`
	}{}
	if err := json.Unmarshal(data, &extracted); err != nil {
		return nil, err
	}

	user := &auth.AuthUser{
		Id:          extracted.StudentID,
		Name:        extracted.Name,
		Username:    extracted.LastName,
		Email:       extracted.Email,
		RawUser:     rawUser,
		AccessToken: token.AccessToken,
	}

	user.Expiry, _ = types.ParseDateTime(time.Now().Add(time.Duration(token.ExpiresIn)))

	return user, nil
}

func wrapProvider() auth.ProviderFactoryFunc {
	return func() auth.Provider {
		return NewPaleoidProvider()
	}
}

func init() {
	auth.Providers[NamePaleoid] = wrapProvider()
}

func main() {
	app := pocketbase.New()
	// create a new OAuth2 provider
	// paleoid := NewPaleoidProvider()
	// app.Logger().Debug("provider", "name", paleoid.DisplayName())

	// auth.Providers[NamePaleoid] = wrapProvider()

	app.OnBootstrap().BindFunc(func(e *core.BootstrapEvent) error {
		if err := e.Next(); err != nil {
			return err
		}

		// register the provider
		users, err := e.App.FindCollectionByNameOrId("users")
		if err != nil {
			log.Fatal(err)
		}

		users.OAuth2.Enabled = true
		users.OAuth2.Providers = append(users.OAuth2.Providers, core.OAuth2ProviderConfig{
			Name: NamePaleoid,
		})
		return e.Next()
	})

	// users.OAuth2.Providers = append(users.OAuth2.Providers, core.OAuth2ProviderConfig{
	// 	Name: NamePaleoid,
	// })
	// .AddOAuth2Provider(paleoid)
	// auth.NewGithubProvider()NewPaleoidProvider
	// 	paleoid, err := auth.NewProviderByName("oidc")
	// 	if err != nil {
	// 		log.Fatal(err)
	// 	}

	// 	paleoid.FetchAuthUser = func(token *oauth2.Token) (user *auth.AuthUser, err error) {
	// 		data, err := paleoid.FetchRawUserInfo(token)
	// 		if err != nil {
	// 			return nil, err
	// 		}

	// 		rawUser := map[string]interface{}{}
	// 		if err := json.Unmarshal(data, &rawUser); err != nil {
	// 			return nil, err
	// 		}

	// 		emailAny, ok := e.OAuth2User.RawUser["email"]
	// 		if !ok {
	// 			e.App.Logger().Error("email not found")
	// 			return e.Next()
	// 		}
	// 		email := emailAny.(string)

	// 		info := e.OAuth2User.RawUser["info_studente"].(map[string]interface{})

	// 		collection, err := app.FindCollectionByNameOrId("users")
	// 		if err != nil {
	// 			return err
	// 		}

	// 		user, err := app.FindAuthRecordByEmail(collection, email)

	// 		// e.IsNewRecord = false
	// 		if err != nil {
	// 			if err == sql.ErrNoRows {
	// 				e.App.Logger().Warn("user not found")
	// 				user = core.NewRecord(collection)
	// 				user.SetEmail(email)
	// 				user.SetVerified(true)
	// 				user.SetPassword(security.RandomString(16))
	// 				user.Set("studentid", e.OAuth2User.RawUser["matricola"])
	// 				user.Set("name", e.OAuth2User.RawUser["nome"])
	// 				user.Set("surname", e.OAuth2User.RawUser["cognome"])
	// 				user.Set("class", info["classe"])

	// 				user.Set("roles", "studente")

	// 		return &auth.AuthUser{
	// 			RawUser: rawUser,
	// 		}, nil
	// 	}
	// 	// paleoid.SetDisplayName("PaleoID")

	app.OnRecordAuthWithOAuth2Request("users").UnbindAll()
	app.OnRecordAuthWithOAuth2Request("users").BindFunc(func(e *core.RecordAuthWithOAuth2RequestEvent) error {
		for _, provider := range e.Collection.OAuth2.Providers {
			if e.ProviderName == provider.Name {
				e.App.Logger().Debug("provider found", "name", e.ProviderName)
				p, err := provider.InitProvider()
				if err != nil {
					return err
				}
				e.App.Logger().Debug("provider initialized", "name", e.ProviderName, "displayName", p.DisplayName())

				e.ProviderClient = p
				break
			}
		}
		p, err := auth.NewProviderByName(e.ProviderName)
		if err != nil {
			e.App.Logger().Error("provider not found", "name", e.ProviderName)
			return err
		}
		e.App.Logger().Debug("provider p", p.DisplayName())

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
