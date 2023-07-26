package handlers

import (
	"bytes"
	"fmt"
	"html/template"
	"main/helpers"
	"main/service"
	"net/http"
	"regexp"
	"strings"
)

func SignUp(w http.ResponseWriter, r *http.Request) {
	ServerResp.Err = errResp{}
	page := getLastPage(r)

	r.ParseForm()
	username := r.FormValue("username-sign-up")
	if !helpers.IsPrintable(username) {
		ServerResp.Err.ErrorCode = 1
		ServerResp.Err.Message = "only printable symbols allowed in username"
		renderAuthAttempt(w, http.StatusUnauthorized, page)
		return
	} else if !helpers.IsNameLenOk(username) {
		ServerResp.Err.ErrorCode = 1
		ServerResp.Err.Message = "username should be more than 3 symbols and less than 10"
		renderAuthAttempt(w, http.StatusUnauthorized, page)
		return
	}

	email := r.FormValue("email-address-sign-up")

	if !regexp.MustCompile(`[a-z0-9.\-_]+@[a-z0-9]+\.[a-z0-9]+`).Match([]byte(email)) {
		ServerResp.Err.ErrorCode = 1
		ServerResp.Err.Message = "invalid email"
		renderAuthAttempt(w, http.StatusUnauthorized, page)
		return
	}

	password := r.FormValue("password-sign-up")

	newUser := service.User{
		Name:     username,
		Email:    email,
		Password: password,
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
		fmt.Println(err)
	}
	http.SetCookie(w, cookie)

	ServerResp.CurrentUser = newUser

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func LogIn(w http.ResponseWriter, r *http.Request) {
	ServerResp.Err = errResp{}

	page := getLastPage(r)

	r.ParseForm()
	email := r.FormValue("email-address-log-in")
	password := r.FormValue("password-log-in")

	maybeUser := service.User{
		Email:    email,
		Password: password,
	}

	u, err := service.Login(maybeUser)
	if err != nil {
		ServerResp.Err.ErrorCode = 2
		ServerResp.Err.Message = err.Error()
		renderAuthAttempt(w, http.StatusUnauthorized, page)
		return
	}

	cookie, err := service.AddCookie(u.Id)
	if err != nil {
		fmt.Println(err)
	}
	http.SetCookie(w, cookie)

	ServerResp.CurrentUser = *u
	renderAuthAttempt(w, http.StatusOK, page)
}

func LogOut(w http.ResponseWriter, r *http.Request) {
	err := service.DeleteCookie(ServerResp.CurrentUser.Id)
	if err != nil {
		fmt.Println(err)
	}
	http.SetCookie(w, &http.Cookie{
		Name:   "potato_batat_bulba",
		Value:  "0",
		MaxAge: -1,
	})
	ServerResp.CurrentUser = service.User{}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func getLastPage(r *http.Request) string {
	page := "index"
	refererPage := r.Header.Get("Referer")
	for key := range templatesMap {
		if strings.Contains(refererPage, key) {
			page = key
		}
	}

	return page
}

func getCurrentUser(r *http.Request) service.User {
	cookie, err := r.Cookie("potato_batat_bulba")
	if err == nil {
		user, _ := service.CheckCookie(cookie)
		return user
	} else {
		return service.User{}
	}
}

func renderAuthAttempt(w http.ResponseWriter, status int, page string) {
	templatePath, ok := templatesMap[page]
	if !ok {
		//the template does not exist
		RenderErrorPage(w, http.StatusInternalServerError)
		return
	}
	template, err := template.ParseFiles(templatePath)
	if err != nil {
		RenderErrorPage(w, http.StatusInternalServerError)
		return
	}

	var temp bytes.Buffer
	err = template.Execute(&temp, ServerResp)
	if err != nil {
		RenderErrorPage(w, http.StatusInternalServerError)
		return
	}

	w.Write(temp.Bytes())
}
