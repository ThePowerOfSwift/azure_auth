package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/go-martini/martini"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

type OauthResponse struct {
	TokenType    string `json:"token_type"`
	Scope        string `json:"scope"`
	ExpiresIn    string `json:"expires_in"`
	ExtExpiresIn string `json:"ext_expires_in"`
	ExpiresOn    string `json:"expires_on"`
	NotBefore    string `json:"not_before"`
	Resource     string `json:"resource"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	IDToken      string `json:"id_token"`
}

// implementation of TokenInterface interface
func (o OauthResponse) getTemporaryToken() string {
	return ""
}
func (o OauthResponse) publicToken() string {
	return fmt.Sprint(uuid.New())
}
func (o OauthResponse) getAccessToken() string {
	return o.AccessToken
}
func (o OauthResponse) getRefreshToken() string {
	return o.RefreshToken
}

type refreshTokenResponse struct {
	TokenType    string `json:"token_type"`
	Scope        string `json:"scope"`
	ExpiresIn    string `json:"expires_in"`
	ExtExpiresIn string `json:"ext_expires_in"`
	ExpiresOn    string `json:"expires_on"`
	NotBefore    string `json:"not_before"`
	Resource     string `json:"resource"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type LoggedUserInfo struct {
	OdataContext      string        `json:"@odata.context"`
	BusinessPhones    []interface{} `json:"businessPhones"`
	DisplayName       string        `json:"displayName"`
	GivenName         string        `json:"givenName"`
	JobTitle          string        `json:"jobTitle"`
	Mail              string        `json:"mail"`
	MobilePhone       interface{}   `json:"mobilePhone"`
	OfficeLocation    string        `json:"officeLocation"`
	PreferredLanguage interface{}   `json:"preferredLanguage"`
	Surname           string        `json:"surname"`
	UserPrincipalName string        `json:"userPrincipalName"`
	ID                string        `json:"id"`
}

var (
	db = InitDB("./authData")
	myEnv, envErr         = godotenv.Read()

	timeout = time.Duration(5 * time.Second)
	client  = http.Client{
		Timeout: timeout,
	}

	OuathScopes           = []string{"offline_access", "openid"}
	ClientIdConst         = getenv("CLIENT_ID")
	TenantConst           = getenv("TENANT")
	ResourcePathConst     = getenv("RESOURCE_PATH")
	WorldWideAuthority    = getenv("WORLD_WIDE_AUTHORITY")
	ClientSecretConst     = getenv("CLIENT_SECRET")
	BaseUrl				  = getenv("BASE_URL")

	authority             = Authority{WorldWideAuthority, os.Getenv("TENANT")}
)

func getenv(name string) string{
	if envErr != nil {
		panic("Error while reading local variables")
	}
	v := myEnv[name]
	if v == "" {
		panic("Missing required environment variable " + name)
	}
	return v
}

func init() {
	CreateTable(db)
}

func getMeRequest(token string) *http.Response {
	meRequest, err := http.NewRequest("GET", "https://graph.microsoft.com/v1.0/me", nil)
	if err != nil {
		panic(fmt.Errorf("ERROR: %s", err))
	}

	tokenStr := fmt.Sprint("Bearer ", token)
	meRequest.Header.Set("Authorization", tokenStr)

	meResponse, err := client.Do(meRequest)
	if meResponse.StatusCode != 200 {
		panic("Something went wrong, retry to sign in please")
	}

	if err != nil {
		panic(fmt.Errorf("ERROR: %s", err))
	}

	return meResponse
}

func getMeHandler(w http.ResponseWriter, r *http.Request) {
	keys := r.URL.Query()
	token := keys.Get("token")
	validateParams(token)

	it, err := FindRecord(db, token)
	if err != nil {
		http.Error(w, err.Error(), 404)
		return
	}

	meResponse := getMeRequest(it[0].AccessToken)
	defer meResponse.Body.Close()

	if meResponse.StatusCode != 200 {
		if !retryWithRefresh(it[0].ClientPublicToken, it[0].RefreshToken) {
			panic("Something went wrong, retry to sign in please")
		}
	}

	meBytes, err := ioutil.ReadAll(meResponse.Body)
	if err != nil {
		panic(fmt.Errorf("ERROR: %s", err))
	}

	var loggedUserInfo LoggedUserInfo
	err = json.Unmarshal(meBytes, &loggedUserInfo)
	if err != nil {
		panic(fmt.Errorf("ERROR: %s", err))
	}

	js, err := json.Marshal(loggedUserInfo)
	if err != nil {
		panic(fmt.Errorf("ERROR: %s", err))
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

func retryWithRefresh(clientToken, refreshToken string) bool {
	params := url.Values{}

	params.Add("grant_type", "refresh_token")
	params.Add("refresh_token", refreshToken)
	params.Add("client_id", ClientIdConst)
	params.Add("client_secret", ClientSecretConst)
	params.Add("resource", ResourcePathConst)

	urlBytes := []byte(strings.TrimSpace(params.Encode()))

	request, err := http.NewRequest("POST", fmt.Sprint(authority), bytes.NewReader(urlBytes))
	if err != nil {
		panic(fmt.Errorf("ERROR: %s", err))
	}

	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Set("client-request-id", fmt.Sprint(uuid.Must(uuid.NewRandom())))
	request.Header.Set("client-return-client-request-id", "true")

	response, _ := client.Do(request)
	if response.StatusCode != 200 {
		panic("Something went wrong, retry to sign in please")
	}

	defer response.Body.Close()

	var refreshTokenResponse refreshTokenResponse
	meBytes, err := ioutil.ReadAll(response.Body)
	if err != nil {
		panic(fmt.Errorf("ERROR: %s", err))
	}

	err = json.Unmarshal(meBytes, &refreshTokenResponse)
	if err != nil {
		panic(fmt.Errorf("ERROR: %s", err))
	}

	err = UpdateItem(db, clientToken, refreshTokenResponse.AccessToken, refreshTokenResponse.RefreshToken)
	if err != nil {
		panic(fmt.Errorf("ERROR: %s", err))
	}
	return true
}
func main() {
	defer db.Close()
	m := martini.Classic()

	m.Get("/get_me", getMeHandler)
	m.Post("/auth_with_temporary_token", authWithTempTokenHandler)
	m.Get("/auth", oauthHandler)
	m.Get("/auth_url", oauthUrlHandler)
	m.Get("/auth/azureactivedirectory/callback", aadAuthHandler)
	m.Run()
}
