package main

import (
	"fmt"
	"main/handlers"
	mw "main/middleware"
	"net/http"
	"path/filepath"
)

func routes() http.Handler {
	fileServer := http.FileServer(http.Dir("./static")) //fileServer := http.FileServer(neuteredFileSystem{http.Dir("./static")})
	mux := http.NewServeMux()
	mux.Handle("/static/", http.StripPrefix("/static", fileServer))
	mux.HandleFunc("/", handlers.RenderMainPage)
	mux.HandleFunc("/sign-up", handlers.SignUp)
	mux.HandleFunc("/log-in", handlers.LogIn)
	mux.HandleFunc("/google/login/", handlers.GoogleLogin)
	mux.HandleFunc("/callback", handlers.GoogleCallback)
	mux.HandleFunc("/github/login/", handlers.GithubLogin)
	mux.HandleFunc("/login/github/callback", handlers.GithubCallback)
	mux.HandleFunc("/log-out", handlers.LogOut)
	mux.HandleFunc("/profilepage/", handlers.RenderProfilePage)
	mux.HandleFunc("/update-profile", handlers.UpdateProfile)
	mux.HandleFunc("/update-user-img", handlers.UpdateProfilePicture)
	mux.HandleFunc("/request-moderator", handlers.SendRequestModerator)
	mux.HandleFunc("/de-promote-user", handlers.DePromoteUser)
	// TODO
	mux.HandleFunc("/add-category", handlers.AddCategory)
	mux.HandleFunc("/delete-category", handlers.DeleteCategory)

	mux.HandleFunc("/response/", handlers.AddResponse)

	mux.HandleFunc("/createpostpage/", handlers.RenderCreatePostPage)
	mux.HandleFunc("/create-post", handlers.CreatePost)
	mux.HandleFunc("/postpage/", handlers.RenderPostPage)
	mux.HandleFunc("/postpage/edit-delete-report/", handlers.EditOrDeletePost)
	mux.HandleFunc("/like-dislike", handlers.LikeDislike)
	mux.HandleFunc("/comment/edit-delete/", handlers.EditOrDeleteComment)
	mux.HandleFunc("/comment", handlers.AddComment)
	mux.HandleFunc("/search", handlers.HandlerSearch)
	mux.HandleFunc("/filter/", handlers.HandlerFilterCategory)
	mux.HandleFunc("/filter-date", handlers.HandlerUpdateFilters)
	mux.HandleFunc("/sort-up", handlers.SortLikesUp)
	mux.HandleFunc("/sort-down", handlers.SortLikesDown)
	mux.HandleFunc("/create-post-categories", handlers.CreatePostCategories)
	mux.HandleFunc("/delete-post-category", handlers.DeletePostCategory)
	mux.HandleFunc("/notification/", handlers.Notifications)
	fmt.Println("Server started on the http://localhost:8080/")
	fmt.Println("Press Ctrl+C to stop the server")
	return mw.RecoverPanic(mw.Limit(mw.SecureHeaders(mux)))
}

type neuteredFileSystem struct {
	fs http.FileSystem
}

// creates a safe filesystem, disabling browsing through the static folder in the browser
func (nfs neuteredFileSystem) Open(path string) (http.File, error) {
	f, err := nfs.fs.Open(path)
	if err != nil {
		return nil, err
	}
	s, _ := f.Stat()
	if s.IsDir() {
		index := filepath.Join(path, "index.html")
		if _, err := nfs.fs.Open(index); err != nil {
			closeErr := f.Close()
			if closeErr != nil {
				return nil, closeErr
			}
			return nil, err
		}
	}
	return f, nil
}
