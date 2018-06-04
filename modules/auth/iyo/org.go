package iyo

import (
	"encoding/json"
	"fmt"
	"net/http"
	"crypto/tls"
	"code.gitea.io/gitea/modules/cache"
	"time"
	"sync"
)

const iyoOrgURL = "https://itsyou.online/api/organizations"
//const iyoOrgURL = "https://dev.itsyou.online:8443/api/organizations"

type Organization struct {
	GlobalId 	string 			`json:"globalid"`
	Children 	[]Organization	`json:"children"`
}

// Get all organization children that the user is member of
func (p *Provider) GetUserOrganizations(userName string) ([]string, error){
	accessToken, err := p.getOrgAccessToken()
	if err != nil {
		return nil, err
	}

	childOrganizations , err := p.getOrganizations(accessToken)
	if err != nil {
		return nil, err
	}

	// Check for user membership using go routine
	userOrganizations := make([]string, 0)
	var m sync.RWMutex
	wg := sync.WaitGroup{}
	wg.Add(len(childOrganizations))
	addIfMemberOfOrg := func(username, globalId, accessToken string ){
		if userIsMemberOfOrg(userName, globalId, accessToken){
			m.Lock()
			userOrganizations = append(userOrganizations, globalId)
			m.Unlock()
		}
		wg.Done()
	}
	for _, globalId := range childOrganizations {
		go addIfMemberOfOrg(userName, globalId, accessToken)
	}
	wg.Wait()

	return userOrganizations, nil
}

// Get all organization children that the user is member of
func (p *Provider) IsChildOrganization(org string) (bool) {
	if org == p.config.ClientID {
		return true
	}

	accessToken, err := p.getOrgAccessToken()
	if err != nil {
		return false
	}

	childOrganizations, err := p.getOrganizations(accessToken)
	if err != nil {
		return false
	}
	for _, globalId := range childOrganizations {
		if org == globalId {
			return true
		}
	}
	return false
}

func (p *Provider) getOrganizations(accessToken string) ([]string, error){
	endpoint := fmt.Sprintf("/%v/tree", p.config.ClientID)

	hc := &http.Client{}
	/*hc.Transport = &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}*/
	req, _ := http.NewRequest("GET", iyoOrgURL + endpoint, nil)
	req.Header.Set("Authorization", "token " + accessToken)
	resp, err := hc.Do(req)
	if err != nil || resp == nil {
		return nil, err
	}
	response := Organization{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return nil, err
	}
	children := recursiveGetChildren(response, []string {})
	return children, nil
}

func recursiveGetChildren(org Organization, children []string) ([]string) {
	children = append(children, org.GlobalId)
	for _, org := range org.Children {
		children = recursiveGetChildren(org, children)
	}
	return children
}

// makes an api call to itsyou.online to verify if the user has access to the organization
func userIsMemberOfOrg(username string, globalId string, accessToken string) bool {
	endpoint := fmt.Sprintf("/%v/users/ismember/%v", globalId, username)

	hc := &http.Client{}
	/*hc.Transport = &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}*/
	req, _ := http.NewRequest("GET", iyoOrgURL + endpoint, nil)
	req.Header.Set("Authorization", "token " + accessToken)
	resp, err := hc.Do(req)
	if err != nil || resp == nil {
		return false
	}

	response := &struct {
		IsMember bool
	}{}

	err = json.NewDecoder(resp.Body).Decode(response)
	if err != nil {
		return false
	}
	return response.IsMember
}

// Get Access token for Organization to get all its children organizations
func (p *Provider) getOrgAccessToken() (string, error){
	// try to get access token from cache if it is still valid for more than 5 minutes
	expiry, _ := cache.Get("access_token_expiry")
	if expiry, ok := expiry.(int64); ok {
		if expiry > (time.Now().Unix() + 5 * 60) {
			token, _ := cache.Get("access_token")
			if token, ok := token.(string); ok {
				return token, nil
			}
		}
	}
	hc := &http.Client{}
	req, err := http.NewRequest("POST", tokenURL, nil)
	if err != nil {
		return "", err
	}

	q := req.URL.Query()
	q.Add("client_id", p.config.ClientID)
	q.Add("client_secret", p.config.ClientSecret)
	q.Add("grant_type", "client_credentials")
	req.URL.RawQuery = q.Encode()
	hc.Transport = &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	resp, err := hc.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	response := &struct {
		AccessToken string      `json:"access_token"`
		TokenType   string      `json:"token_type"`
		Scope       string      `json:"scope"`
		ExpiresIn   int64       `json:"expires_in"`
		Info        interface{} `json:"info"`
	}{}

	err = json.NewDecoder(resp.Body).Decode(response)
	if err != nil {
		return "", err
	}
	cache.Put("access_token", response.AccessToken)
	cache.Put("access_token_expiry", time.Now().Unix() + response.ExpiresIn)
	return response.AccessToken, nil
}
