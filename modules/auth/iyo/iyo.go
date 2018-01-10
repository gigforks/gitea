// Package iyo implements the OAuth2 protocol for authenticating users through itsyou.online.
package iyo

import (
	"bytes"
	"errors"
	"io"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"encoding/json"
	"github.com/markbates/goth"
	"golang.org/x/oauth2"
	"crypto/tls"
)

const (
	authURL         string = "https://itsyou.online/v1/oauth/authorize"
	//authURL         string = "https://dev.itsyou.online:8443/v1/oauth/authorize"
	tokenURL        string = "https://itsyou.online/v1/oauth/access_token"
	//tokenURL        string = "https://dev.itsyou.online:8443/v1/oauth/access_token"
	endpointProfile string = "https://itsyou.online/api/users/%s/info"
	//endpointProfile string = "https://dev.itsyou.online:8443/api/users/%s/info"
)

// New creates a new IYO provider, and sets up important connection details.
// You should always call `iyo.New` to get a new Provider. Never try to create
// one manually.
func New(clientKey, secret, callbackURL string, scopes ...string) *Provider {
	p := &Provider{
		ClientKey:           clientKey,
		Secret:              secret,
		CallbackURL:         callbackURL,
		providerName:        "itsyou.online",
	}
	p.config = newConfig(p, scopes)
	return p
}

// Provider is the implementation of `goth.Provider` for accessing itsyou.online.
type Provider struct {
	ClientKey    string
	Secret       string
	CallbackURL  string
	HTTPClient   *http.Client
	config       *oauth2.Config
	providerName string
}

// Name is the name used to retrieve this provider later.
func (p *Provider) Name() string {
	return p.providerName
}

// SetName is to update the name of the provider (needed in case of multiple providers of 1 type)
func (p *Provider) SetName(name string) {
	p.providerName = name
}

func (p *Provider) Client() *http.Client {
	client := goth.HTTPClientWithFallBack(p.HTTPClient)

	//TODO Remove https skip
	client.Transport = &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	return client
}

// Debug is a no-op for the iyo package.
func (p *Provider) Debug(debug bool) {}

// BeginAuth asks itsyou.online for an authentication end-point.
func (p *Provider) BeginAuth(state string) (goth.Session, error) {
	authCodeURL := p.config.AuthCodeURL(state)
	session := &Session{
		AuthURL: authCodeURL,
	}
	return session, nil
}

// FetchUser will go to itsyou.online and access basic information about the user.
func (p *Provider) FetchUser(session goth.Session) (goth.User, error) {
	sess := session.(*Session)
	user := goth.User{
		Name:	     sess.UserName,
		AccessToken: sess.AccessToken,
		Provider:    p.Name(),
		ExpiresAt:   sess.ExpiresAt,
	}

	if user.AccessToken == "" {
		// data is not yet retrieved since accessToken is still empty
		return user, fmt.Errorf("%s cannot get user information without accessToken", p.providerName)
	}
	profileUrl := fmt.Sprintf(endpointProfile, user.Name)
	response, err := p.Client().Get(profileUrl + "?access_token=" + url.QueryEscape(sess.AccessToken))
	if err != nil {
		return user, err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return user, fmt.Errorf("%s responded with a %d trying to fetch user information", p.providerName, response.StatusCode)
	}

	bits, err := ioutil.ReadAll(response.Body)
	if err != nil {

		return user, err
	}
	err = json.NewDecoder(bytes.NewReader(bits)).Decode(&user.RawData)
	if err != nil {
		return user, err
	}

	err = userFromReader(bytes.NewReader(bits), &user)
	return user, err
}

func userFromReader(reader io.Reader, user *goth.User) error {
	type EmailEntry struct {
		EmailAddress string `json:"emailaddress"`
		Label        string `json:"label"`
	}

	u := struct {
		UserName 	string		`json:"username"`
		EmailAddresses 	[]EmailEntry 	`json:"emailaddresses"`
	}{}
	err := json.NewDecoder(reader).Decode(&u)
	if err != nil {
		return err
	}
	user.UserID = u.UserName
	if len(u.EmailAddresses) > 0 {
		user.Email = u.EmailAddresses[0].EmailAddress
	}
	return err
}

func newConfig(provider *Provider, scopes []string) *oauth2.Config {
	c := &oauth2.Config{
		ClientID:     provider.ClientKey,
		ClientSecret: provider.Secret,
		RedirectURL:  provider.CallbackURL,
		Endpoint: oauth2.Endpoint{
			AuthURL:  authURL,
			TokenURL: tokenURL,
		},
		Scopes: scopes,
	}
	return c
}

//RefreshToken refresh token is not provided for now by iyo
func (p *Provider) RefreshToken(refreshToken string) (*oauth2.Token, error) {
	return nil, errors.New("Refresh token is not provided by itsyou.online")
}

//RefreshTokenAvailable refresh token is not provided for now by iyo
func (p *Provider) RefreshTokenAvailable() bool {
	return false
}
