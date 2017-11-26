package iyo

import (
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/markbates/goth"
)

// Session stores data during the auth process with itsyou.online.
type Session struct {
	AuthURL     	string
	AccessToken 	string
	ExpiresAt   	time.Time
	UserName    	string
	Organizations 	[]string
}

// GetAuthURL will return the URL set by calling the `BeginAuth` function on the itsyou.online provider.
func (s Session) GetAuthURL() (string, error) {
	if s.AuthURL == "" {
		return "", errors.New(goth.NoAuthUrlErrorMessage)
	}
	return s.AuthURL, nil
}

// Authorize the session with itsyou.online and return the access token to be stored for future use.
func (s *Session) Authorize(provider goth.Provider, params goth.Params) (string, error) {
	p := provider.(*Provider)
	token, err := p.config.Exchange(goth.ContextForClient(p.Client()), params.Get("code"), params.Get("state"))
	if err != nil {
		return "", err
	}

	if !token.Valid() {
		return "", errors.New("Invalid token received from provider")
	}
	info := token.Extra("info").(map[string]interface{})
	s.AccessToken = token.AccessToken
	s.ExpiresAt = token.Expiry
	s.UserName = info["username"].(string)
	// Get All children organizations of the organization whose client_secret was used
	// and the user is member of it
	userOrganizations, err := p.GetUserOrganizations(s.UserName)
	if err == nil {
		s.Organizations = userOrganizations
	}
	return token.AccessToken, err
}

// Marshal the session into a string
func (s Session) Marshal() string {
	b, _ := json.Marshal(s)
	return string(b)
}

func (s Session) String() string {
	return s.Marshal()
}

// UnmarshalSession will unmarshal a JSON string into a session.
func (p *Provider) UnmarshalSession(data string) (goth.Session, error) {
	sess := &Session{}
	err := json.NewDecoder(strings.NewReader(data)).Decode(sess)
	return sess, err
}
