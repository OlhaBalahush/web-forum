package handlers

import (
	"bytes"
	"fmt"
	"html/template"
	"main/service"
	"net/http"
	"strconv"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

var templatesMap = map[string]string{
	"index":          "templates/index.html",
	"postpage":       "templates/post.html",
	"profilepage":    "templates/profile.html",
	"createpostpage": "templates/createPost.html",
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

	template, err := template.ParseFiles("templates/index.html")
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

	template, err := template.ParseFiles("templates/index.html")
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

	template, err := template.ParseFiles("templates/index.html")
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

	template, err := template.ParseFiles("templates/index.html")
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

	template, err := template.ParseFiles("templates/index.html")
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

	template, err := template.ParseFiles("templates/index.html")
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
	template, err := template.ParseFiles("templates/error.html")
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
