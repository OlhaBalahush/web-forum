package handlers

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"main/service"
	"net/http"
	"os"
	"strings"
	"time"
)

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

	template, err := template.ParseFiles("templates/profile.html")
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
	template, err := template.ParseFiles("templates/profile.html")
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
	template, err := template.ParseFiles("templates/profile.html")
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
