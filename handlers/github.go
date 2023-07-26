package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"01.kood.tech/git/Olya/forum-moderation.git/service"

	"golang.org/x/oauth2"
)

type UserInfo struct {
	AvatarURL string `json:"avatar_url"`
	Name      string `json:"login"`
	Email     string `json:"email"`
}

type EmailInfo struct {
	Email      string `json:"email"`
	Primary    bool   `json:"primary"`
	Verified   bool   `json:"verified"`
	Visibility string `json:"visibility"`
}

// Define the OAuth2 configuration
var oauthConf = &oauth2.Config{
	ClientID:     "f2e71e45fcfbc297f361",
	ClientSecret: "eb46900d691da450a039eab2b877c234214df565",
	RedirectURL:  "http://localhost:8080/login/github/callback",
	Scopes:       []string{"user:email"},
	Endpoint: oauth2.Endpoint{
		AuthURL:  "https://github.com/login/oauth/authorize",
		TokenURL: "https://github.com/login/oauth/access_token",
	},
}

func GithubLogin(w http.ResponseWriter, r *http.Request) {
	// Redirect user to GitHub authorization URL
	url := oauthConf.AuthCodeURL("state", oauth2.AccessTypeOnline)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func GithubCallback(w http.ResponseWriter, r *http.Request) {
	ServerResp.Err = errResp{}
	page := getLastPage(r)

	// Exchange authorization code for access token
	code := r.URL.Query()["code"][0]
	token := GetGithubAccessToken(code)

	name, email, err := GetGithubData(token)

	if err != nil {
		// Handle error
		http.Error(w, "Failed to get user information", http.StatusInternalServerError)
		return
	}

	user, err := service.GetUserByEmail(email)
	if err == nil {
		cookie, err := service.AddCookie(user.Id)
		if err != nil {
			// TODO NORMAL ERROR HANDLER
			fmt.Println(err)
		}

		http.SetCookie(w, cookie)
		ServerResp.CurrentUser = *user
	} else {
		newUser := service.User{
			Name:  name,
			Email: email,
		}
		id, err := service.CreateUser(newUser)
		if err != nil {
			// error if username or email already in use
			ServerResp.Err.ErrorCode = 1
			ServerResp.Err.Message = err.Error()
			renderAuthAttempt(w, http.StatusUnauthorized, page)
			return
		}

		cookie, err := service.AddCookie(id)
		if err != nil {
			// TODO NORMAL ERROR HANDLE
			fmt.Println(err)
		}

		http.SetCookie(w, cookie)
		ServerResp.CurrentUser = newUser
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func GetGithubAccessToken(code string) string {
	requestBodyMap := map[string]string{
		"client_id":     oauthConf.ClientID,
		"client_secret": oauthConf.ClientSecret,
		"code":          code,
	}
	requestJSON, _ := json.Marshal(requestBodyMap)

	req, reqerr := http.NewRequest("POST", "https://github.com/login/oauth/access_token", bytes.NewBuffer(requestJSON))
	if reqerr != nil {
		log.Panic("Request creation failed")
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, resperr := http.DefaultClient.Do(req)
	if resperr != nil {
		log.Panic("Request failed")
	}

	respbody, _ := ioutil.ReadAll(resp.Body)

	type githubAccessTokenResponse struct {
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
		Scope       string `json:"scope"`
	}

	var ghresp githubAccessTokenResponse
	json.Unmarshal(respbody, &ghresp)

	return ghresp.AccessToken
}

func GetGithubData(accessToken string) (string, string, error) {
	req, reqerr := http.NewRequest("GET", "https://api.github.com/user", nil)
	if reqerr != nil {
		log.Panic("API Request creation failed")
	}

	authorizationHeaderValue := fmt.Sprintf("token %s", accessToken)
	req.Header.Set("Authorization", authorizationHeaderValue)

	resp, resperr := http.DefaultClient.Do(req)
	if resperr != nil {
		log.Panic("Request failed")
	}

	respbody, _ := ioutil.ReadAll(resp.Body)

	req, reqerr = http.NewRequest("GET", "https://api.github.com/user/emails", nil)
	if reqerr != nil {
		log.Panic("API Request creation failed")
	}

	req.Header.Set("Authorization", authorizationHeaderValue)

	resp, resperr = http.DefaultClient.Do(req)
	if resperr != nil {
		log.Panic("Request failed")
	}

	emailResp, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", "", err
	}

	var ghUser UserInfo
	var ghEmail []EmailInfo

	json.Unmarshal(respbody, &ghUser)
	json.Unmarshal(emailResp, &ghEmail)

	return ghUser.Name, ghEmail[0].Email, nil
}
