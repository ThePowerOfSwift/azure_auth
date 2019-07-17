package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"io/ioutil"
	"net/http"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/microsoft"
)

// wrapper to oauth2.Token. I can not implement interface TokenInterface with non local value
type oToken struct {
	*oauth2.Token
	TemporaryToken string
}

func (t oToken) publicToken() string {
	return ""
}
func (t oToken) getTemporaryToken() string {
	return t.TemporaryToken
}
func (t oToken) getAccessToken() string {
	return t.AccessToken
}
func (t oToken) getRefreshToken() string {
	return t.RefreshToken
}

var (
	xOauth2Config = oauth2.Config{
		ClientID:     ClientIdConst,
		ClientSecret: ClientSecretConst,
		RedirectURL:  RedirectOauthUrlConst,
		Endpoint:     microsoft.AzureADEndpoint(TenantConst),
		Scopes:       OuathScopes,
	}
)

func oauthUrlHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, AuthUrl)
}

// Auth handler which will redirect to AAD
func oauthHandler(w http.ResponseWriter, r *http.Request) {
	state := randToken(48)
	authorizationURL := xOauth2Config.AuthCodeURL(state)
	http.Redirect(w, r, authorizationURL, 301)
}

// process the redirection from AAD
func aadAuthHandler(w http.ResponseWriter, r *http.Request) {
	authorizationCode := r.URL.Query().Get("code")

	ck, err := r.Cookie("state")
	if err == nil && (r.URL.Query().Get("state") != ck.Value) {
		fmt.Fprintf(w, "Error: State is not the same")
	}
	oAuthToken, err := xOauth2Config.Exchange(context.Background(), authorizationCode)
	if err != nil {
		panic(err)
	}

	meResponse := getMeRequest(oAuthToken.AccessToken)
	defer meResponse.Body.Close()

	var loggedUserInfo LoggedUserInfo
	meBytes, err := ioutil.ReadAll(meResponse.Body)
	handleError(err)

	err = json.Unmarshal(meBytes, &loggedUserInfo)
	handleError(err)

	token := oToken{oAuthToken, fmt.Sprint(uuid.New())}

	dataToSave := NewUserDataToSave(token, loggedUserInfo)

	StoreItem(db, dataToSave)

	tempTokenURL := generateTempTokenUrl(token.TemporaryToken)
	http.Redirect(w, r, tempTokenURL, 301)
}

func authWithTempTokenHandler(w http.ResponseWriter, r *http.Request) {
	keys := r.URL.Query()

	temporaryToken := keys.Get("temporary_token")

	validateParams(temporaryToken)

	clientPublicToken, err := ChangeTemporaryTokenToPublic(db, temporaryToken)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	fmt.Fprintf(w, clientPublicToken)
}
