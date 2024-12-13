package main

import (
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
		e.App.Logger().Debug("new users oauth request", e)
		// e.Collection
		// e.ProviderName
		// e.ProviderClient
		// e.Record (could be nil)
		// e.OAuth2User
		// e.CreateData
		// e.IsNewRecord
		// and all RequestEvent fields...

		return e.Next()
	})

	if err := app.Start(); err != nil {
		log.Fatal(err)
	}
}
