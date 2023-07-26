package service

import (
	"database/sql"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/OlhaBalahush/web-forum/functions/helpers"
)

var (
	ErrBadFileExtension = errors.New("only .png, .jpeg, .svg and .gif extensions are allowed")
	ErrBadFileSize      = errors.New("filesize must be less than 20MB")
)

func createTablePosts(db *sql.DB) {
	posts_table := `CREATE TABLE IF NOT EXISTS posts (
		id 			INTEGER PRIMARY KEY AUTOINCREMENT,
        title       TEXT NOT NULL,
        text        TEXT NOT NULL,
        user_id 	INTEGER REFERENCES users (id),   
        created_at  DATE DEFAULT CURRENT_TIMESTAMP NOT NULL,
		image_path  TEXT NOT NULL
	);`

	_, err := db.Exec(posts_table)

	if err != nil {
		log.Fatal("create tables posts ", err)
	}
}

func CreatePost(post Post) (int, error) {
	db := GetDBAddr()
	createTablePosts(db)

	query := `INSERT INTO posts (title, text, user_id, image_path) VALUES (?, ?, ?, ?)`
	res, err := db.Exec(query, post.Title, post.Text, post.Creator.Id, post.ImagePath)
	if err != nil {
		return 0, err
	}

	postID, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	AddCategories(int(postID), post.Categories)

	return int(postID), nil
}

func GetAllPostsAndCategories(currentUser User) (*[]Post, *[]string, error) {
	db := GetDBAddr()
	createTablePosts(db)

	var posts []Post
	var categories []string

	rows, err := db.Query(`SELECT * FROM posts
	ORDER by created_at DESC`)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	for rows.Next() {
		var (
			id        int
			title     string
			text      string
			userID    int
			createdAt time.Time
			imagePath string
		)

		err = rows.Scan(&id, &title, &text, &userID, &createdAt, &imagePath)
		if err != nil {
			log.Fatal(err)
		}

		if _, err := os.Stat(imagePath); err != nil {
			imagePath = ""
		}

		time := createdAt.Format("January 02, 2006")
		post := Post{
			Id:         id,
			Title:      title,
			Text:       text,
			Categories: categories,
			CreatedAt:  time,
			ImagePath:  imagePath,
		}

		creator, _ := GetUserById(userID)
		postCategories, _ := GetCategoriesByPostID(id)
		likes, dislikes, _ := GetAllLikesDislikesByPostID(post.Id)
		mark := IsLikePostIdUserId(currentUser.Id, id)
		switch mark {
		case 1:
			post.LikedByUser = true
			post.DislikedByUser = false
		case -1:
			post.LikedByUser = false
			post.DislikedByUser = true
		}

		post.Creator = *creator
		post.Likes, post.Dislikes = likes, dislikes
		post.Categories = *postCategories

		posts = append(posts, post)

		for _, category := range *postCategories {
			if !helpers.IsAny(categories, category) {
				categories = append(categories, category)
			}
		}

		fuck, _ := GetCategoriesByPostID(-1)
		for _, category := range *fuck {
			if !helpers.IsAny(categories, category) {
				categories = append(categories, category)
			}
		}

	}

	return &posts, &categories, nil
}

func GetPostsByUserId(userID int) (*[]Post, error) {
	db := GetDBAddr()
	createTablePosts(db)

	var posts []Post
	rows, err := db.Query("SELECT * FROM posts WHERE user_id = ?", userID)
	if err != nil {
		// TODO: sth else for this err
		log.Fatal(err)
	}
	defer rows.Close()

	for rows.Next() {
		var (
			id        int
			title     string
			text      string
			userId    int
			createdAt time.Time
			imagePath string
		)

		err = rows.Scan(&id, &title, &text, &userId, &createdAt, &imagePath)
		if err != nil {
			// TODO: sth else for this err
			log.Fatal(err)
		}

		time := createdAt.Format("January 02, 2006")
		postCategories, _ := GetCategoriesByPostID(id)
		if _, err := os.Stat(imagePath); err != nil {
			imagePath = ""
		}

		post := Post{
			Id:               id,
			Title:            title,
			Text:             text,
			Categories:       *postCategories,
			CreatedAt:        time,
			ImagePath:        imagePath,
			CurrentUsersPost: userID == userId,
		}

		likes, dislikes, _ := GetAllLikesDislikesByPostID(post.Id)
		creator, _ := GetUserById(userID)
		post.Likes, post.Dislikes = likes, dislikes
		post.Creator = *creator

		posts = append(posts, post)
	}

	return &posts, nil
}

func GetPostById(userID, postID int) (*Post, error) {
	db := GetDBAddr()
	createTablePosts(db)

	var post Post
	var creatorID int
	var time time.Time
	row := db.QueryRow("SELECT * FROM posts WHERE id = ?", postID)
	err := row.Scan(&post.Id, &post.Title, &post.Text, &creatorID, &time, &post.ImagePath)
	if err != nil {
		return nil, err
	}
	if _, err := os.Stat(post.ImagePath); err != nil {
		post.ImagePath = ""
	}

	creator, _ := GetUserById(creatorID)
	comments, _ := GetCommentsByPostId(userID, post.Id)
	likes, dislikes, _ := GetAllLikesDislikesByPostID(post.Id)
	categories, _ := GetCategoriesByPostID(postID)

	post.Creator = *creator
	post.Comments = *comments
	post.Categories = *categories
	post.Likes, post.Dislikes = likes, dislikes
	post.CreatedAt = time.Format("January 02, 2006")
	post.CurrentUsersPost = creatorID == userID
	return &post, nil
}

func SortPostsByLikes(order string) (*[]Post, error) {
	db := GetDBAddr()
	createTablePosts(db)

	posts := make([]Post, 0)
	rows, err := db.Query(`
		SELECT p.id AS post_id, p.title, p.text, p.user_id, p.created_at, COUNT(CASE WHEN r.mark = 1 THEN r.id ELSE NULL END) as like_count, COUNT(CASE WHEN r.mark = -1 THEN r.id ELSE NULL END) as dislike_count
		FROM posts p
		LEFT JOIN relations_likes r ON p.id = r.post_id
		GROUP BY p.id
		ORDER BY like_count
		` + order)

	if err != nil {
		return &posts, err
	}

	defer rows.Close()

	for rows.Next() {
		var post Post
		var userID int
		var time time.Time
		err := rows.Scan(&post.Id, &post.Title, &post.Text, &userID, &time, &post.Likes, &post.Dislikes)
		post.CreatedAt = time.Format("January 02, 2006")

		if err != nil {
			return &posts, err
		}

		creator, _ := GetUserById(userID)
		categories, _ := GetCategoriesByPostID(post.Id)
		post.Creator = *creator
		post.Categories = *categories

		posts = append(posts, post)
	}

	if err := rows.Err(); err != nil {
		panic(err)
	}

	return &posts, nil
}

func GetUserImage(r *http.Request, savePath string) (string, error) {
	file, header, err := r.FormFile("attachment")
	if err != nil {
		return "", err
	}
	allowedFilesize := 20000000
	allowedExtensions := map[string]bool{
		".png":  true,
		".jpeg": true,
		".jpg":  true,
		".svg":  true,
		".gif":  true,
	}
	extension := filepath.Ext(header.Filename)
	if !allowedExtensions[extension] {
		return "", ErrBadFileExtension
	}
	if header.Size > int64(allowedFilesize) {
		return "", ErrBadFileSize
	}
	defer file.Close()
	newFilename := "upload-*" + extension
	tempFile, err := ioutil.TempFile(savePath, newFilename)
	if err != nil {
		return "", err
	}
	defer tempFile.Close()
	fileBytes, err := ioutil.ReadAll(file)
	if err != nil {
		return "", err
	}
	tempFile.Write(fileBytes)
	return tempFile.Name(), nil
}

func DeletePost(postID int) error {
	db := GetDBAddr()
	createTablePosts(db)

	query := `DELETE FROM posts WHERE id = ?;`
	_, err := db.Exec(query, postID)

	if err != nil {
		return err
	}
	return nil
}

func EditPost(post Post, userID int) error {
	db := GetDBAddr()
	createTablePosts(db)

	query := "UPDATE posts SET title=?, text=? WHERE id=? AND user_id=?"
	_, err := db.Exec(query, post.Title, post.Text, post.Id, userID)
	if err != nil {
		return err
	}

	return nil

}
