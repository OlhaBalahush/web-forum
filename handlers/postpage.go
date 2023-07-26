package handlers

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"strconv"
	"strings"

	"01.kood.tech/git/Olya/forum-moderation.git/service"
)

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

	template, err := template.ParseFiles("templates/post.html")
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

	template, err := template.ParseFiles("templates/createPost.html")
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
	template, err := template.ParseFiles("templates/createPost.html")
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
	_, err := template.ParseFiles("templates/createPost.html")
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
	_, err := template.ParseFiles("templates/createPost.html")
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
