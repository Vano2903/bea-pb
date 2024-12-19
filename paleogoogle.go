package main

import (
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/pocketbase/pocketbase/tools/auth"
	"golang.org/x/oauth2"
)

// PaleoGoogle is the unique name of the Google provider.
const NamePaleoGoogle string = "paleogoogle"

// Google allows authentication via Google OAuth2.
type PaleoGoogle struct {
	auth.Google
}

// NewPaleoGoogleProvider creates new PaleoGoogle provider instance with some defaults.
func NewPaleoGoogleProvider() *PaleoGoogle {

	p := &PaleoGoogle{
		*auth.NewGoogleProvider(),
	}

	p.SetDisplayName("PaleoGoogle")
	return p
	// return &Google{auth.BaseProvider{
	// 	ctx:         context.Background(),
	// 	displayName: "Google",
	// 	pkce:        true,
	// 	scopes: []string{
	// 		"https://www.googleapis.com/auth/userinfo.profile",
	// 		"https://www.googleapis.com/auth/userinfo.email",
	// 	},
	// 	authURL:     "https://accounts.google.com/o/oauth2/auth",
	// 	tokenURL:    "https://accounts.google.com/o/oauth2/token",
	// 	userInfoURL: "https://www.googleapis.com/oauth2/v1/userinfo",
	// }}
}

func (p *PaleoGoogle) FetchRawUserInfo(token *oauth2.Token) ([]byte, error) {

	u, err := url.Parse(p.AuthURL())
	if err != nil {
		return nil, err
	}
	v := u.Query()
	v.Add("hd", "itispaleocapa.it")

	u.RawQuery = v.Encode()
	req, err := http.NewRequestWithContext(p.BaseProvider.Context(), "GET", u.String(), nil)
	if err != nil {
		return nil, err
	}

	// return p.sendRawUserInfoRequest(req, token)
	client := p.Client(token)

	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	result, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	// http.Client.Get doesn't treat non 2xx responses as error
	if res.StatusCode >= 400 {
		return nil, fmt.Errorf(
			"failed to fetch OAuth2 user profile via %s (%d):\n%s",
			u.String(),
			res.StatusCode,
			string(result),
		)
	}

	return result, nil
}

// FetchAuthUser returns an AuthUser instance based the Google's user api.
// func (p *PaleoGoogle) FetchAuthUser(token *oauth2.Token) (*auth.AuthUser, error) {
// 	data, err := p.FetchRawUserInfo(token)
// 	if err != nil {
// 		return nil, err
// 	}

// 	rawUser := map[string]any{}
// 	if err := json.Unmarshal(data, &rawUser); err != nil {
// 		return nil, err
// 	}

// 	extracted := struct {
// 		Id            string `json:"id"`
// 		Name          string `json:"name"`
// 		Email         string `json:"email"`
// 		Picture       string `json:"picture"`
// 		VerifiedEmail bool   `json:"verified_email"`
// 	}{}
// 	if err := json.Unmarshal(data, &extracted); err != nil {
// 		return nil, err
// 	}

// 	user := &AuthUser{
// 		Id:           extracted.Id,
// 		Name:         extracted.Name,
// 		AvatarURL:    extracted.Picture,
// 		RawUser:      rawUser,
// 		AccessToken:  token.AccessToken,
// 		RefreshToken: token.RefreshToken,
// 	}

// 	user.Expiry, _ = types.ParseDateTime(token.Expiry)

// 	if extracted.VerifiedEmail {
// 		user.Email = extracted.Email
// 	}

// 	return user, nil
// }
