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

var (
	db = InitDB()

	timeout = time.Duration(5 * time.Second)
	client  = http.Client{
		Timeout: timeout,
	}

	OuathScopes       = []string{"offline_access", "openid"}
	ClientIdConst     = getenv("CLIENT_ID")
	TenantConst       = getenv("TENANT")
	ClientSecretConst = getenv("CLIENT_SECRET")
	BaseUrl           = getenv("BASE_URL")
	DbUser            = getenv("DB_USER")
	DbPassword        = getenv("DB_PASSWORD")
	DbPort            = getenv("DB_PORT")
	DbName            = getenv("DB_NAME")
	DbHost            = getenv("DB_HOST")

	authority = Authority{"login.microsoftonline.com", os.Getenv("TENANT")}
)

func getenv(name string) string {
	_, err := os.Stat(".env")
	if err == nil {
		godotenv.Load()
	}
	v := os.Getenv(name)
	if v == "" {
		panic("Missing required environment variable " + name)
	}
	return v
}

func getMeHandler(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("Authorization")

	user := FindUserByPubToken(token)
	if (user == User{}) {
		http.Error(w, "Record not found", http.StatusNotFound)
		return
	}
	meResponse := getMeRequest(user.AccessToken)
	defer meResponse.Body.Close()

	if meResponse.StatusCode != 200 {
		if !retryWithRefresh(&user) {
			panic("Something went wrong, retry to sign in please")
		}
	}

	meBytes, err := ioutil.ReadAll(meResponse.Body)
	if err != nil {
		panic(fmt.Errorf("ERROR: %s", err))
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(meBytes)
}

func retryWithRefresh(user *User) bool {
	params := url.Values{}

	params.Add("grant_type", "refresh_token")
	params.Add("refresh_token", user.RefreshToken)
	params.Add("client_id", ClientIdConst)
	params.Add("client_secret", ClientSecretConst)
	params.Add("resource", "https://graph.microsoft.com")

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

	RefreshToken(user, refreshTokenResponse)

	return true
}

func getPhotoHandler(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("Authorization")

	user := FindUserByPubToken(token)
	if (user == User{}) {
		http.Error(w, "Record not found", http.StatusNotFound)
		return
	}

	tokenStr := fmt.Sprint("Bearer ", user.AccessToken)

	picRequest, err := http.NewRequest("GET", "https://graph.microsoft.com/v1.0/me/photo/$value", nil)
	picRequest.Header.Set("Authorization", tokenStr)
	picResponse, err := client.Do(picRequest)
	if picResponse.StatusCode != 200 {
		fmt.Println("Something went wrong, accessing user picture")
		return
	}
	if err != nil {
		fmt.Errorf("ERROR: %s", err)
		return
	}
	pictureBinary, _ := ioutil.ReadAll(picResponse.Body)
	if pictureBinary == nil {
		w.Write([]byte{})
	}

	w.Write(pictureBinary)
}

func main() {
	defer db.Close()
	m := martini.Classic()

	m.Get("/get_me", getMeHandler)
	m.Get("/get_user_photo", getPhotoHandler)
	m.Post("/auth_with_temporary_token", authWithTempTokenHandler)
	m.Get("/auth", oauthHandler)
	m.Get("/auth_url", oauthUrlHandler)
	m.Get("/auth/azureactivedirectory/callback", aadAuthHandler)
	m.Run()
}
