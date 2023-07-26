package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"main/service"
	"net/http"

	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

var (
	//oAuth2.0 Configurations
	googleOauthConfig = &oauth2.Config{
		ClientID:     "536872213393-28a91f51gi5ionh3q9iqp0v5qbhf4gk9.apps.googleusercontent.com",
		ClientSecret: "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",
		RedirectURL:  "http://localhost:8080/callback",
		Scopes:       []string{"https://www.googleapis.com/auth/userinfo.email", "https://www.googleapis.com/auth/userinfo.profile"},
		Endpoint:     google.Endpoint,
	}
	randomState = "random"
)

func GoogleLogin(w http.ResponseWriter, r *http.Request) {
	url := googleOauthConfig.AuthCodeURL(randomState)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func GoogleCallback(w http.ResponseWriter, r *http.Request) {
	ServerResp.Err = errResp{}
	page := getLastPage(r)
	// fmt.Println("I am here")

	if r.FormValue("state") != randomState {
		fmt.Println("state is not valid")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	//state
	state := r.URL.Query()["state"][0]
	if state != "random" {
		fmt.Println("state don't mutch")
		fmt.Fprint(w, "state don't mutch")
		return
	}

	//code
	code := r.URL.Query()["code"][0]
	//exchange code for token
	token, err := googleOauthConfig.Exchange(context.Background(), code)
	if err != nil {
		fmt.Println("Code-Token Exchange Failed")
		fmt.Fprintln(w, "Code-Token Exchange Failed")
	}

	//use google api to get user info
	resp, err := http.Get("https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + token.AccessToken)
	if err != nil {
		fmt.Println("User Data fetch Failed")
		fmt.Fprintln(w, "User Data fetch Failed")
	}
	defer resp.Body.Close()

	//receive the userinfo in json and decode
	var userData struct {
		Email string `json:"email"`
		Name  string `json:"name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&userData); err != nil {
		log.Fatal(err)
	}

	user, err := service.GetUserByEmail(userData.Email)
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
			Name:  userData.Name,
			Email: userData.Email,
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
			// TODO NORMAL ERROR HANDLER
			fmt.Println(err)
		}
		http.SetCookie(w, cookie)
		ServerResp.CurrentUser = newUser
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}
