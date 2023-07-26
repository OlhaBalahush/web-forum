package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/OlhaBalahush/web-forum/functions/helpers"
	"github.com/OlhaBalahush/web-forum/functions/service"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
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

var templatesMap = map[string]string{
	"index":          "../public/index.html",
	"postpage":       "../public/templates/post.html",
	"profilepage":    "../public/templates/profile.html",
	"createpostpage": "../public/templates/createPost.html",
}

type resp struct {
	CurrentUser             service.User
	Categories              []string
	Posts                   []service.Post
	LikedDislikedByCurrUser []service.Post
	CommentsByCurrUser      []service.Comment
	Post                    service.Post
	PostCash                service.Post
	User                    service.User
	Err                     errResp
}

type errResp struct {
	ErrorCode int
	Message   string
}

var ServerResp resp

func RenderMainPage(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		RenderErrorPage(w, http.StatusNotFound)
		return
	}

	if r.Method != "GET" {
		RenderErrorPage(w, http.StatusMethodNotAllowed)
		return
	}

	ServerResp.Err = errResp{}
	ServerResp.CurrentUser = getCurrentUser(r)
	ServerResp.CurrentUser.NewNotifications = []service.Notification{}
	ServerResp.CurrentUser.ReadedNotifications = []service.Notification{}

	allUserNotifications, _ := service.GetNotifications(ServerResp.CurrentUser.Id)
	for _, n := range allUserNotifications {
		if n.Seen {
			ServerResp.CurrentUser.ReadedNotifications = append(ServerResp.CurrentUser.ReadedNotifications, n)
		} else {
			ServerResp.CurrentUser.NewNotifications = append(ServerResp.CurrentUser.NewNotifications, n)
		}
	}

	posts, categories, err := service.GetAllPostsAndCategories(ServerResp.CurrentUser)

	if err != nil {
		RenderErrorPage(w, http.StatusInternalServerError)
		return
	}

	ServerResp.Posts, ServerResp.Categories = *posts, *categories
	ServerResp.PostCash = service.Post{}

	template, err := template.ParseFiles("../public/index.html")
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

func HandlerUpdateFilters(w http.ResponseWriter, r *http.Request) {
	from, err := strconv.Atoi(r.FormValue("fromInput"))
	if err != nil {
		RenderErrorPage(w, http.StatusBadRequest)
		return
	}
	to, err := strconv.Atoi(r.FormValue("toInput"))
	if err != nil {
		RenderErrorPage(w, http.StatusBadRequest)
		return
	}

	var filteredPosts []service.Post
	for _, post := range ServerResp.Posts {
		layout := "January 02, 2006"
		createdAt, err := time.Parse(layout, post.CreatedAt)
		if err != nil {
			RenderErrorPage(w, http.StatusInternalServerError)
			return
		}
		fromTime := time.Date(from, time.January, 1, 0, 0, 0, 0, time.UTC)
		toTime := time.Date(to, time.December, 31, 23, 59, 59, 0, time.UTC)
		if createdAt.After(fromTime) && createdAt.Before(toTime) {
			filteredPosts = append(filteredPosts, post)
		}
	}

	ServerResp.Posts = filteredPosts

	template, err := template.ParseFiles("../public/index.html")
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

func HandlerFilterCategory(w http.ResponseWriter, r *http.Request) {
	pathSplitted := strings.Split(r.URL.Path, "/")
	category := pathSplitted[2]
	if pathSplitted[1] != "filter" {
		RenderErrorPage(w, http.StatusNotFound)
		return
	}

	filteredPosts := filterPostsByCategory(category)

	ServerResp.Posts = filteredPosts

	template, err := template.ParseFiles("../public/index.html")
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

func HandlerSearch(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/search" {
		RenderErrorPage(w, http.StatusNotFound)
		return
	}

	if r.Method != "POST" {
		RenderErrorPage(w, http.StatusMethodNotAllowed)
		return
	}

	searchtext := r.FormValue("searchBar")
	var matchingPosts []service.Post
	for _, post := range ServerResp.Posts {
		if strings.Contains(strings.ToLower(post.Title), strings.ToLower(searchtext)) ||
			strings.Contains(strings.ToLower(post.Text), strings.ToLower(searchtext)) {
			matchingPosts = append(matchingPosts, post)
		} else {
			for _, category := range post.Categories {
				if strings.Contains(strings.ToLower(category), strings.ToLower(searchtext)) {
					matchingPosts = append(matchingPosts, post)
					break
				}
			}
		}
	}

	ServerResp.Posts = matchingPosts

	template, err := template.ParseFiles("../public/index.html")
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

func SortLikesUp(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/sort-up" {
		RenderErrorPage(w, http.StatusNotFound)
		return
	}

	if r.Method != "GET" {
		RenderErrorPage(w, http.StatusMethodNotAllowed)
		return
	}

	posts, err := service.SortPostsByLikes(`DESC`)
	if err != nil {
		RenderErrorPage(w, http.StatusInternalServerError)
		return
	}

	// filter posts from the slice of posts above
	filteredPosts := []service.Post{}
	for _, post := range *posts {
		for _, filteredPost := range ServerResp.Posts {
			if post.Id == filteredPost.Id {
				filteredPosts = append(filteredPosts, post)
			}
		}
	}

	ServerResp.Posts = filteredPosts

	template, err := template.ParseFiles("../public/index.html")
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

func SortLikesDown(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/sort-down" {
		RenderErrorPage(w, http.StatusNotFound)
		return
	}

	if r.Method != "GET" {
		RenderErrorPage(w, http.StatusMethodNotAllowed)
		return
	}

	posts, err := service.SortPostsByLikes(`ASC`)
	if err != nil {
		RenderErrorPage(w, http.StatusInternalServerError)
		return
	}

	// filter posts from the slice of posts above
	filteredPosts := []service.Post{}
	for _, post := range *posts {
		for _, filteredPost := range ServerResp.Posts {
			if post.Id == filteredPost.Id {
				filteredPosts = append(filteredPosts, post)
			}
		}
	}

	ServerResp.Posts = filteredPosts

	template, err := template.ParseFiles("../public/index.html")
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

func RenderErrorPage(w http.ResponseWriter, statusCode int) {
	template, err := template.ParseFiles("../public/templates/error.html")
	if err != nil {
		http.Error(w, fmt.Sprint("Error parsing:", err), statusCode)
		return
	}

	res := errResp{
		ErrorCode: statusCode,
		Message:   http.StatusText(statusCode),
	}

	w.WriteHeader(statusCode)
	template.Execute(w, res)
}

func filterPostsByCategory(category string) []service.Post {
	filteredPosts := make([]service.Post, 0)

	posts, _, err := service.GetAllPostsAndCategories(ServerResp.CurrentUser)
	if err != nil {
		// render error page?
		return filteredPosts
	}

	for _, post := range *posts {
		if strings.Contains(strings.Join(post.Categories, ","), category) {
			filteredPosts = append(filteredPosts, post)
		}
	}

	return filteredPosts
}

func AddCategory(w http.ResponseWriter, r *http.Request) {

	// fmt.Println(ServerResp.CurrentUser.Role)

	if ServerResp.CurrentUser.Role != 3 {
		RenderErrorPage(w, 403)
		return
	}

	category := r.FormValue("add-category-input")
	fmt.Println(category)

	service.AddCategory(category)

	requestId, _ := service.GetUserRequest(ServerResp.User.Id)
	if requestId != -1 {
		err := service.DeleteNotification(requestId)
		if err != nil {
			fmt.Println("error")
		}
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func DeleteCategory(w http.ResponseWriter, r *http.Request) {

	if ServerResp.CurrentUser.Role != 3 {
		RenderErrorPage(w, 403)
		return
	}

	category := r.FormValue("category")
	service.DeleteCategory(category)

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// TODO

func AddResponse(w http.ResponseWriter, r *http.Request) {
	// take id of notification
	response := r.FormValue("response")

	pathSplitted := strings.Split(r.URL.Path, "/")
	notId, _ := strconv.Atoi(pathSplitted[2])
	// get notification
	notification, _ := service.GetNotificationId(notId)
	// create response
	service.AddResponse(notification.WhoDid.Id, notification.PostID, "response", response)
	// set notification as readed
	service.SetNotificationSeen(notId)

	postId := strconv.Itoa(notification.PostID)
	link := "/postpage/" + postId
	http.Redirect(w, r, link, http.StatusSeeOther)
}

func RenderPostPage(w http.ResponseWriter, r *http.Request) {
	ServerResp.Err = errResp{}
	ServerResp.PostCash = service.Post{}
	pathSplitted := strings.Split(r.URL.Path, "/")
	postID, err := strconv.Atoi(pathSplitted[2])
	if pathSplitted[1] != "postpage" || err != nil {
		RenderErrorPage(w, http.StatusNotFound)
		return
	}

	currentComment := ServerResp.CurrentUser.CommentEdit
	ServerResp.CurrentUser = getCurrentUser(r)

	ServerResp.CurrentUser.NewNotifications = []service.Notification{}
	ServerResp.CurrentUser.ReadedNotifications = []service.Notification{}

	allUserNotifications, _ := service.GetNotifications(ServerResp.CurrentUser.Id)
	for _, n := range allUserNotifications {
		if n.Seen {
			ServerResp.CurrentUser.ReadedNotifications = append(ServerResp.CurrentUser.ReadedNotifications, n)
		} else {
			ServerResp.CurrentUser.NewNotifications = append(ServerResp.CurrentUser.NewNotifications, n)
		}
	}

	ServerResp.CurrentUser.CommentEdit = currentComment

	post, err := service.GetPostById(ServerResp.CurrentUser.Id, postID)
	if err != nil {
		RenderErrorPage(w, http.StatusNotFound)
		return
	}
	ServerResp.Post = *post

	mark := service.IsLikePostIdUserId(ServerResp.CurrentUser.Id, ServerResp.Post.Id)
	switch mark {
	case 1:
		ServerResp.Post.LikedByUser = true
		ServerResp.Post.DislikedByUser = false
	case -1:
		ServerResp.Post.LikedByUser = false
		ServerResp.Post.DislikedByUser = true
	}

	for index, c := range ServerResp.Post.Comments {
		commentMark := service.IsLikeCommentIdUserId(ServerResp.CurrentUser.Id, c.Id)
		switch commentMark {
		case 1:
			c.LikedByUser = true
			c.DislikedByUser = false
		case -1:
			c.LikedByUser = false
			c.DislikedByUser = true
		}
		ServerResp.Post.Comments[index] = c
	}

	template, err := template.ParseFiles("../public/templates/post.html")
	if err != nil {
		fmt.Println("err2")
		fmt.Println(err)
		RenderErrorPage(w, http.StatusInternalServerError)
		return
	}

	var temp bytes.Buffer
	err = template.Execute(&temp, ServerResp)
	if err != nil {
		fmt.Println("err1")
		fmt.Println(err)
		RenderErrorPage(w, http.StatusInternalServerError)
		return
	}
	w.Write(temp.Bytes())
}

func RenderCreatePostPage(w http.ResponseWriter, r *http.Request) {
	ServerResp.Err = errResp{}
	ServerResp.CurrentUser = getCurrentUser(r)

	if ServerResp.CurrentUser.Name == "" {
		RenderErrorPage(w, http.StatusNotFound)
		return
	}

	template, err := template.ParseFiles("../public/templates/createPost.html")
	if err != nil {
		RenderErrorPage(w, http.StatusInternalServerError)
		return
	}

	if len(ServerResp.Categories) == 0 {
		_, categories, _ := service.GetAllPostsAndCategories(ServerResp.CurrentUser)
		ServerResp.Categories = *categories
	}

	var temp bytes.Buffer
	err = template.Execute(&temp, ServerResp)
	if err != nil {
		RenderErrorPage(w, http.StatusInternalServerError)
		return
	}

	w.Write(temp.Bytes())
}

func CreatePost(w http.ResponseWriter, r *http.Request) {
	ServerResp.Err = errResp{}
	template, err := template.ParseFiles("../public/templates/createPost.html")
	if err != nil {
		fmt.Println("FUCK109", err)

		RenderErrorPage(w, http.StatusInternalServerError)
		return
	}

	if r.Method != "POST" {
		fmt.Println("FUCK19", err)

		RenderErrorPage(w, http.StatusMethodNotAllowed)
		return
	}

	r.ParseMultipartForm(20 << 20)

	title := r.FormValue("title-create-post")

	category := ServerResp.PostCash.Categories
	text := r.FormValue("text-create-post")
	newPost := service.Post{
		Title:      title,
		Text:       text,
		Categories: category,
		Creator:    ServerResp.CurrentUser,
	}

	file, _, err := r.FormFile("attachment")
	if err == nil {
		if _, err := os.Stat("static/uploads"); os.IsNotExist(err) {
			err := os.Mkdir("static/uploads", 0755)
			if err != nil {
				// TODO normal error
				ServerResp.Err.ErrorCode = 5
				ServerResp.Err.Message = errors.New("image attaching is temporarily unavailable").Error()
				template.Execute(w, ServerResp)
				return
			}
		}

		filepath, err := service.GetUserImage(r, "static/uploads")

		if err == service.ErrBadFileExtension || err == service.ErrBadFileSize {
			// TODO need to add a proper errorcode
			ServerResp.Err.ErrorCode = 5
			ServerResp.Err.Message = err.Error()
			template.Execute(w, ServerResp)
			return
		} else if err != nil {
			RenderErrorPage(w, http.StatusInternalServerError)
			return
		}

		if filepath != "" {
			newPost.ImagePath = filepath
		}
		defer file.Close()
	}

	if title == "" || len(category) == 0 || text == "" {
		ServerResp.Err.ErrorCode = 4
		ServerResp.Err.Message = "you can not create empty post"
		template.Execute(w, ServerResp)
		return
	}

	newPost.Id, err = service.CreatePost(newPost)
	if err != nil {
		ServerResp.Err.ErrorCode = 4
		ServerResp.Err.Message = "failed to create post"
		template.Execute(w, ServerResp)
		return
	}

	ServerResp.PostCash = service.Post{}

	link := "/postpage/" + strconv.Itoa(newPost.Id)
	http.Redirect(w, r, link, http.StatusSeeOther)
}

// add category to bufferPost
func CreatePostCategories(w http.ResponseWriter, r *http.Request) {
	ServerResp.Err = errResp{}
	_, err := template.ParseFiles("../public/templates/createPost.html")
	link := "/createpostpage/" + ServerResp.CurrentUser.Name
	if err != nil {
		RenderErrorPage(w, http.StatusInternalServerError)
		return
	}

	if r.Method != "POST" {
		RenderErrorPage(w, http.StatusMethodNotAllowed)
		return
	}

	r.ParseMultipartForm(20 << 20)

	category := r.FormValue("category-create-post")

	if category == "" {
		http.Redirect(w, r, link, http.StatusSeeOther)
		return
	}
	//??????????????????????????????????????????????? TODO
	var categories []string
	categories = append(categories, ServerResp.PostCash.Categories...)
	categories = append(categories, category)

	newPost := service.Post{}

	if ServerResp.PostCash.Id != -1 {
		newPost = service.Post{
			Id:         -1,
			Categories: categories,
			Creator:    ServerResp.CurrentUser,
		}

		if ServerResp.PostCash.Edit {
			newPost.Id = ServerResp.PostCash.Id
			newPost.Edit = true
		}
		ServerResp.PostCash = newPost
	} else {

		if len(ServerResp.PostCash.Categories) >= 5 {
			// TODO: write error
			ServerResp.Err.ErrorCode = 5
			ServerResp.Err.Message = "you can't add more than 5 categories"
			//

			renderAuthAttempt(w, http.StatusBadRequest, strings.Split(link, "/")[1])
			// http.Redirect(w, r, link, http.StatusSeeOther)
			return
		}

		isAlreadyExist := false
		for _, c := range ServerResp.PostCash.Categories {
			if c == category {
				isAlreadyExist = true
				break
			}
		}
		if !isAlreadyExist {
			ServerResp.PostCash.Categories = append(ServerResp.PostCash.Categories, category)
		}
	}

	title := r.FormValue("title-create-post")
	text := r.FormValue("text-create-post")

	ServerResp.PostCash.Title = title
	ServerResp.PostCash.Text = text

	http.Redirect(w, r, link, http.StatusSeeOther)
}

// delete category to bufferPost
func DeletePostCategory(w http.ResponseWriter, r *http.Request) {
	ServerResp.Err = errResp{}
	_, err := template.ParseFiles("../public/templates/createPost.html")
	link := "/createpostpage/" + ServerResp.CurrentUser.Name
	if err != nil {
		RenderErrorPage(w, http.StatusInternalServerError)
		return
	}

	// CHANGE ERROR
	if r.Method != "POST" {
		RenderErrorPage(w, http.StatusMethodNotAllowed)
		return
	}

	r.ParseMultipartForm(20 << 20)

	categoryToDelete := r.FormValue("category-to-delete")

	var newCategories []string
	for _, c := range ServerResp.PostCash.Categories {
		if c != categoryToDelete {
			newCategories = append(newCategories, c)
		}
	}

	title := r.FormValue("title-create-post")
	text := r.FormValue("text-create-post")

	ServerResp.PostCash.Title = title
	ServerResp.PostCash.Text = text

	ServerResp.PostCash.Categories = newCategories
	http.Redirect(w, r, link, http.StatusSeeOther)
}

func LikeDislike(w http.ResponseWriter, r *http.Request) {
	if ServerResp.CurrentUser.Id < 1 {
		ServerResp.Err.ErrorCode = 3
		ServerResp.Err.Message = "only registered users can like posts"

		renderAuthAttempt(w, http.StatusUnauthorized, "postpage")
		return
	}

	if r.Method != "POST" {
		RenderErrorPage(w, http.StatusMethodNotAllowed)
		return
	}

	err := r.ParseForm()
	if err != nil {
		RenderErrorPage(w, http.StatusBadRequest)
		return
	}

	// what user wants to do
	likedislike := r.FormValue("like-dislike")
	if likedislike == "" {
		likedislike = r.FormValue("like-dislike-comment")
		likeDislikeComment(w, r, likedislike)
		return
	}

	// what do we ave in DB for this user and post
	mark := service.IsLikePostIdUserId(ServerResp.CurrentUser.Id, ServerResp.Post.Id)

	switch mark {
	case 0:
		// case person has no likes nor dislikes on the post
		if likedislike == "like" {
			service.AddLikeDislike(ServerResp.CurrentUser.Id, ServerResp.Post.Id, 1)
		} else if likedislike == "dislike" {
			service.AddLikeDislike(ServerResp.CurrentUser.Id, ServerResp.Post.Id, -1)
		}
	case 1:
		// case person already has like on the post
		if likedislike == "dislike" {
			service.UpdateLikeDislike(ServerResp.CurrentUser.Id, ServerResp.Post.Id, -1)
		} else if likedislike == "like" {
			service.DeleteLikeDislike(ServerResp.CurrentUser.Id, ServerResp.Post.Id, 1)
		}
	case -1:
		// case person already has dislike on the post
		if likedislike == "like" {
			service.UpdateLikeDislike(ServerResp.CurrentUser.Id, ServerResp.Post.Id, 1)
		} else if likedislike == "dislike" {
			service.DeleteLikeDislike(ServerResp.CurrentUser.Id, ServerResp.Post.Id, -1)
		}
	}

	link := "/postpage/" + strconv.Itoa(ServerResp.Post.Id)

	http.Redirect(w, r, link, http.StatusSeeOther)
}

func likeDislikeComment(w http.ResponseWriter, r *http.Request, likedislike string) {
	// what do we have in DB for this user and post
	likedislikeSplitted := strings.Split(likedislike, "-")
	likedislike = likedislikeSplitted[0]
	commentId, _ := strconv.Atoi(likedislikeSplitted[1])
	mark := service.IsLikeCommentIdUserId(ServerResp.CurrentUser.Id, commentId)

	switch mark {
	case 0:
		// case person has no likes nor dislikes on the post
		if likedislike == "like" {
			service.AddLikeDislikeComment(ServerResp.CurrentUser.Id, commentId, 1)
		} else if likedislike == "dislike" {
			service.AddLikeDislikeComment(ServerResp.CurrentUser.Id, commentId, -1)
		}
	case 1:
		// case person already has like on the post
		if likedislike == "dislike" {
			service.UpdateLikeDislikeComments(ServerResp.CurrentUser.Id, commentId, -1)
		} else if likedislike == "like" {
			service.DeleteLikeDislikeComment(ServerResp.CurrentUser.Id, commentId)
		}
	case -1:
		// case person already has dislike on the post
		if likedislike == "like" {
			service.UpdateLikeDislikeComments(ServerResp.CurrentUser.Id, commentId, 1)
		} else if likedislike == "dislike" {
			service.DeleteLikeDislikeComment(ServerResp.CurrentUser.Id, commentId)
		}
	}

	link := "/postpage/" + strconv.Itoa(ServerResp.Post.Id) + "/#" + strconv.Itoa(commentId)

	http.Redirect(w, r, link, http.StatusSeeOther)
}

func AddComment(w http.ResponseWriter, r *http.Request) {
	ServerResp.Err = errResp{}
	if r.URL.Path != "/comment" {
		RenderErrorPage(w, http.StatusNotFound)
		return
	}
	if r.Method != "POST" {
		RenderErrorPage(w, http.StatusMethodNotAllowed)
		return
	}
	comment := r.FormValue("username-input")
	if ServerResp.CurrentUser.CommentEdit != "" {
		oldComment := ServerResp.CurrentUser.CommentEdit
		service.EditComment(ServerResp.Post.Id, ServerResp.CurrentUser.Id, oldComment, comment)
		link := "/postpage/" + strconv.Itoa(ServerResp.Post.Id)
		ServerResp.CurrentUser.CommentEdit = ""
		http.Redirect(w, r, link, http.StatusSeeOther)
		return
	}

	if comment != "" {
		commentID, err := service.AddComment(ServerResp.Post.Id, ServerResp.CurrentUser.Id, comment)
		if err != nil {
			RenderErrorPage(w, http.StatusInternalServerError)
			return
		}
		link := "/postpage/" + strconv.Itoa(ServerResp.Post.Id) + "/#" + strconv.Itoa(commentID)
		http.Redirect(w, r, link, http.StatusSeeOther)
	}
}

func EditOrDeleteComment(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	editDelete := r.FormValue("edit-delete")
	choices := strings.Split(editDelete, "-")
	action, commentID := choices[0], choices[1]
	comId, err := strconv.Atoi(commentID)
	if err != nil {
		fmt.Println(err)
		return
	}
	postId := ServerResp.Post.Id
	link := "/postpage/" + strconv.Itoa(postId) + "/#" + commentID

	switch action {
	case "delete":
		service.DeleteComment(postId, comId)
		// service.DeleteNotificationByPostIdUserId(ServerResp.CurrentUser.Id, postId)
	case "edit":
		comment := service.GetCommentByPostId(postId, comId)
		ServerResp.CurrentUser.CommentEdit = comment
		http.Redirect(w, r, link, http.StatusSeeOther)
	}
	http.Redirect(w, r, link, http.StatusSeeOther)
}

func Notifications(w http.ResponseWriter, r *http.Request) {
	url := r.URL.Path
	tokens := strings.Split(url, "/")
	notificationID, _ := strconv.Atoi(tokens[len(tokens)-1])
	notification, _ := service.GetNotificationId(notificationID)
	postID := tokens[len(tokens)-2]
	// userID := ServerResp.CurrentUser.Id
	if notification.Action != "request" {
		service.SetNotificationSeen(notificationID)
	}
	link := "/postpage/" + postID
	http.Redirect(w, r, link, http.StatusSeeOther)
}

func EditOrDeletePost(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	editDelete := r.FormValue("edit-delete-report")
	choices := strings.Split(editDelete, "-")
	action, postID := choices[0], choices[1]
	link := "/postpage/" + postID
	pID, _ := strconv.Atoi(postID)
	switch action {
	case "delete":
		_ = service.DeletePost(pID)
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	case "edit":
		post, err := service.GetPostById(ServerResp.CurrentUser.Id, pID)
		if err != nil {
			fmt.Println(err)
		}
		post.Edit = true
		ServerResp.PostCash = *post
		link := "/createpostpage/" + ServerResp.CurrentUser.Name
		http.Redirect(w, r, link, http.StatusSeeOther)
		return
	case "report":
		//TODO
		service.AddRequest(ServerResp.CurrentUser.Id, pID, "request", "illegal post")
		http.Redirect(w, r, link, http.StatusSeeOther)
		return
	}
	http.Redirect(w, r, link, http.StatusSeeOther)
}

func RenderProfilePage(w http.ResponseWriter, r *http.Request) {
	ServerResp.Err = errResp{}
	pathSplitted := strings.Split(r.URL.Path, "/")
	username := pathSplitted[2]
	if pathSplitted[1] != "profilepage" {
		RenderErrorPage(w, http.StatusNotFound)
		return
	}

	ServerResp.CurrentUser = getCurrentUser(r)

	ServerResp.CurrentUser.NewNotifications = []service.Notification{}
	ServerResp.CurrentUser.ReadedNotifications = []service.Notification{}

	allUserNotifications, _ := service.GetNotifications(ServerResp.CurrentUser.Id)
	for _, n := range allUserNotifications {
		if n.Seen {
			ServerResp.CurrentUser.ReadedNotifications = append(ServerResp.CurrentUser.ReadedNotifications, n)
		} else {
			ServerResp.CurrentUser.NewNotifications = append(ServerResp.CurrentUser.NewNotifications, n)
		}
	}

	ServerResp.LikedDislikedByCurrUser, _ = service.GetAllPostsLikedDislikedByUserId(ServerResp.CurrentUser.Id)
	ServerResp.CommentsByCurrUser, _ = service.GetAllCommentsByUserId(ServerResp.CurrentUser.Id)

	user, err := service.GetUserByName(username)
	if err != nil {
		RenderErrorPage(w, http.StatusNotFound)
		return
	}

	if err != nil {
		fmt.Println("user.Request, err")
		return
	}

	ServerResp.User = *user
	//TODO
	requestId, _ := service.GetUserRequest(ServerResp.User.Id)
	if requestId != -1 {
		ServerResp.User.Request = 1
	} else {
		ServerResp.User.Request = 0
	}

	template, err := template.ParseFiles("../public/templates/profile.html")
	if err != nil {
		RenderErrorPage(w, http.StatusInternalServerError)
		return
	}

	// if cookie exist, change it
	http.SetCookie(w, &http.Cookie{
		Name:    "page",
		Value:   strings.Split(r.URL.Path, "/")[1],
		Expires: time.Now().Add(time.Hour * 168),
	})

	var temp bytes.Buffer
	err = template.Execute(&temp, ServerResp)
	if err != nil {
		RenderErrorPage(w, http.StatusInternalServerError)
		return
	}

	w.Write(temp.Bytes())
}

func UpdateProfile(w http.ResponseWriter, r *http.Request) {
	ServerResp.Err = errResp{}
	template, err := template.ParseFiles("../public/templates/profile.html")
	if err != nil {
		RenderErrorPage(w, http.StatusInternalServerError)
		return
	}
	if r.Method != "POST" {
		RenderErrorPage(w, http.StatusMethodNotAllowed)
		return
	}

	r.ParseMultipartForm(20 << 20)
	filepath, err := service.GetUserImage(r, "static/userImages/profilePictures")
	if err == service.ErrBadFileExtension || err == service.ErrBadFileSize {
		// need to add a proper errorcode
		ServerResp.Err.ErrorCode = 5
		ServerResp.Err.Message = err.Error()
		template.Execute(w, ServerResp)
		return
	}
	about := r.FormValue("create-about-input")
	ServerResp.CurrentUser.ImagePath = filepath
	ServerResp.CurrentUser.About = about
	err = service.UpdateUser(ServerResp.CurrentUser)
	if err != nil {
		// change error handler
		fmt.Println(err)
	}

	link := "/profilepage/" + ServerResp.CurrentUser.Name

	http.Redirect(w, r, link, http.StatusSeeOther)
}

func UpdateProfilePicture(w http.ResponseWriter, r *http.Request) {
	ServerResp.Err = errResp{}
	template, err := template.ParseFiles("../public/templates/profile.html")
	if err != nil {
		RenderErrorPage(w, http.StatusInternalServerError)
		return
	}
	if r.Method != "POST" {
		RenderErrorPage(w, http.StatusMethodNotAllowed)
		return
	}

	r.ParseMultipartForm(20 << 20)

	file, _, err := r.FormFile("attachment")
	if err == nil {
		if _, err := os.Stat("static/uploads"); os.IsNotExist(err) {
			err := os.Mkdir("static/uploads", 0755)
			if err != nil {
				// TODO normal error
				ServerResp.Err.ErrorCode = 5
				ServerResp.Err.Message = errors.New("image attaching is temporarily unavailable").Error()
				template.Execute(w, ServerResp)
				return
			}
		}
		if _, err := os.Stat("static/uploads/avatars"); os.IsNotExist(err) {
			err := os.Mkdir("static/uploads/avatars", 0755)
			if err != nil {
				// TODO normal error
				ServerResp.Err.ErrorCode = 5
				ServerResp.Err.Message = errors.New("image attaching is temporarily unavailable").Error()
				template.Execute(w, ServerResp)
				return
			}
		}

		filepath, err := service.GetUserImage(r, "static/uploads/avatars")
		if err == service.ErrBadFileExtension || err == service.ErrBadFileSize {
			// TODO need to add a proper errorcode
			ServerResp.Err.ErrorCode = 5
			ServerResp.Err.Message = err.Error()
			template.Execute(w, ServerResp)
			return
		} else if err != nil {
			fmt.Println("here2", err)

			RenderErrorPage(w, http.StatusInternalServerError)
			return
		}

		if filepath != "" {
			ServerResp.CurrentUser.ImagePath = filepath
		}
	}
	defer file.Close()

	err = service.UpdateUserPicture(ServerResp.CurrentUser)
	if err != nil {
		// TODO change error handler
		fmt.Println("err updating user", err)
	}

	link := "/profilepage/" + ServerResp.CurrentUser.Name

	http.Redirect(w, r, link, http.StatusSeeOther)
}

// TODO
func SendRequestModerator(w http.ResponseWriter, r *http.Request) {
	link := "/profilepage/" + ServerResp.CurrentUser.Name
	// TODO
	// check if the notification is already exist

	requestId, _ := service.GetUserRequest(ServerResp.CurrentUser.Id)
	if requestId != -1 {
		err := service.DeleteNotification(requestId)
		if err != nil {
			fmt.Println("error")
		}
		http.Redirect(w, r, link, http.StatusSeeOther)
		return
	}

	// create notification
	err := service.AddRequest(ServerResp.CurrentUser.Id, -1, "request", "for moderator")
	if err != nil {
		// TODO normal error
		fmt.Println("error")
		// template.Execute(w, ServerResp)
		return
	}

	http.Redirect(w, r, link, http.StatusSeeOther)
}

func DePromoteUser(w http.ResponseWriter, r *http.Request) {
	// change role of user which is the page

	if ServerResp.User.Role == 1 {
		ServerResp.User.Role = 2
	} else {
		ServerResp.User.Role = 1
	}

	service.UpdateUserRole(ServerResp.User)

	// delete request (if it is existed)
	requestId, _ := service.GetUserRequest(ServerResp.User.Id)
	if requestId != -1 {
		err := service.DeleteNotification(requestId)
		if err != nil {
			fmt.Println("error")
		}
	}

	link := "/profilepage/" + ServerResp.User.Name

	http.Redirect(w, r, link, http.StatusSeeOther)
}

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
