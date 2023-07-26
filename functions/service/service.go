package service

import (
	"database/sql"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/OlhaBalahush/web-forum/functions/db"
	"github.com/OlhaBalahush/web-forum/functions/helpers"
)

type User struct {
	Id                  int
	Name                string
	Email               string
	Password            string
	About               string
	Role                int // 1 - user / 2 - moderator / 3 - admin
	Posts               []Post
	ImagePath           string
	CreatedAt           string
	CommentEdit         string
	NewNotifications    []Notification
	ReadedNotifications []Notification
	Request             int // 0 - no request 1 - request was sent
}

type Post struct {
	Id               int
	Title            string
	Text             string
	Categories       []string
	Creator          User
	Comments         []Comment
	Likes            int
	Dislikes         int
	LikedByUser      bool
	DislikedByUser   bool
	CreatedAt        string
	ImagePath        string
	CurrentUsersPost bool
	Edit             bool
}

type Comment struct {
	Id                  int
	Text                string
	Creator             User
	Post                Post
	Likes               int
	Dislikes            int
	LikedByUser         bool
	DislikedByUser      bool
	CreatedAt           string
	CurrentUsersComment bool
}

type RelationLike struct {
	Id     int
	UserId int
	PostId int
	Mark   int
}

type Notification struct {
	Id        int
	Reciver   int
	Action    string
	WhoDid    User
	PostID    int
	PostTitle string
	Message   string
	Seen      bool
	CreatedAt string
}

func GetDBAddr() *sql.DB {
	dbConn, err := db.OpenDatabase()
	if err != nil {
		// TODO: render error page?
		log.Fatalf("could not initialize database connection: %s", err)
	}

	return dbConn.GetDB()
}

func createTableUsers(db *sql.DB) {
	posts_table := `CREATE TABLE IF NOT EXISTS posts (
		id 				INTEGER PRIMARY KEY AUTOINCREMENT,
        user_name		TEXT NOT NULL,
        email TEXT		NOT NULL UNIQUE,
        password		TEXT NOT NULL,
        about			TEXT,        
		image_path		TEXT NOT NULL,
        role            INTEGER NOT NULL,
        created_at		DATE DEFAULT CURRENT_TIMESTAMP NOT NULL
	);`

	_, err := db.Exec(posts_table)
	if err != nil {
		log.Fatal(err)
	}
}

func CreateUser(user User) (int, error) {
	db := GetDBAddr()
	createTableUsers(db)

	nameInUse, _ := GetUserByName(user.Name)
	emailInUse, _ := GetUserByEmail(user.Email)
	if nameInUse.Name == user.Name {
		return -1, errors.New("username is already in use")
	}
	if emailInUse.Email == user.Email {
		return -1, errors.New("email is already in use")
	}

	query := "INSERT INTO users (user_name, email, password, about, image_path, role) VALUES (?, ?, ?, ?, ?, ?)"
	hashedPassword, err := helpers.GetPasswordHash(user.Password)
	if err != nil {
		return -1, errors.New("failed creating account")
	}

	imagePath := ""
	u, err := db.Exec(query, user.Name, user.Email, hashedPassword, user.About, &imagePath, 1)
	if err != nil {
		return -1, errors.New("failed creating account")
	}

	id, err := u.LastInsertId()
	if err != nil {
		return -1, errors.New("failed creating account")
	}
	return int(id), nil
}

// TODO ?
func UpdateUser(user User) error {
	db := GetDBAddr()
	createTableUsers(db)

	var err error
	query := `UPDATE users SET about = ? WHERE id = ?`
	_, err = db.Exec(query, user.About, user.Id)

	if err != nil {
		return err
	}

	return nil
}

func UpdateUserPicture(user User) error {
	db := GetDBAddr()
	createTableUsers(db)

	var err error
	query := `UPDATE users SET image_path = ? WHERE id = ?`
	_, err = db.Exec(query, user.ImagePath, user.Id)

	if err != nil {
		return err
	}

	return nil
}

func UpdateUserRole(user User) error {
	db := GetDBAddr()
	createTableUsers(db)

	var err error
	query := `UPDATE users SET role = ? WHERE id = ?`
	_, err = db.Exec(query, user.Role, user.Id)

	if err != nil {
		return err
	}

	return nil
}

func Login(user User) (*User, error) {
	db := GetDBAddr()
	createTableUsers(db)

	u, _ := GetUserByEmail(user.Email)
	if u.Email != user.Email {
		return &User{}, errors.New("no user with such email")
	}

	err := helpers.CheckPassword(user.Password, u.Password)
	if err != nil {
		return &User{}, errors.New("wrong password")
	}

	u.Password = ""
	return u, nil
}

func GetUserById(userID int) (*User, error) {
	db := GetDBAddr()
	createTableUsers(db)

	u := User{}

	query := "SELECT id, user_name, image_path, role FROM users WHERE id = ?"
	err := db.QueryRow(query, userID).Scan(&u.Id, &u.Name, &u.ImagePath, &u.Role)
	if err != nil {
		return &User{}, err
	}
	if err != nil {
		return &User{}, err
	}

	return &u, nil
}

func GetUserByName(username string) (*User, error) {
	db := GetDBAddr()
	createTableUsers(db)

	u := User{}

	query := "SELECT id, user_name, email, about, image_path, role, created_at FROM users WHERE user_name = ?"
	err := db.QueryRow(query, username).Scan(&u.Id, &u.Name, &u.Email, &u.About, &u.ImagePath, &u.Role, &u.CreatedAt)
	if err != nil {
		return &User{}, err
	}
	if _, err := os.Stat(u.ImagePath); err != nil {
		u.ImagePath = ""
	}

	posts, _ := GetPostsByUserId(u.Id)
	u.Posts = *posts

	return &u, nil
}

func GetUserByEmail(email string) (*User, error) {
	db := GetDBAddr()
	createTableUsers(db)

	u := User{}

	query := "SELECT id, user_name, email, password FROM users WHERE email = ?"
	err := db.QueryRow(query, email).Scan(&u.Id, &u.Name, &u.Email, &u.Password)
	if err != nil {
		return &User{}, err
	}
	return &u, nil
}

func getAdmin() (*User, error) {
	db := GetDBAddr()
	createTableUsers(db)

	u := User{}

	query := "SELECT id, user_name, email, about, image_path, role, created_at FROM users WHERE role = ?"
	err := db.QueryRow(query, 3).Scan(&u.Id, &u.Name, &u.Email, &u.About, &u.ImagePath, &u.Role, &u.CreatedAt)
	if err != nil {
		return &User{}, err
	}
	if _, err := os.Stat(u.ImagePath); err != nil {
		u.ImagePath = ""
	}

	posts, _ := GetPostsByUserId(u.Id)
	u.Posts = *posts

	fmt.Println("Admin is: ", u.Name)

	return &u, nil
}

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

func createTableCategories(db *sql.DB) {
	table := `CREATE TABLE IF NOT EXISTS relations_categories (
        post_id 	INTEGER REFERENCES posts (id),
        category 	TEXT NOT NULL
);`
	_, err := db.Exec(table)
	if err != nil {
		log.Fatal(err)
	}
}

func AddCategories(postID int, categories []string) error {
	db := GetDBAddr()
	createTableCategories(db)

	for _, category := range categories {
		query := `INSERT OR IGNORE INTO relations_categories (post_id, category) VALUES (?, ?)`
		_, err := db.Exec(query, postID, category)
		if err != nil {
			return err
		}
	}

	return nil
}

func AddCategory(category string) error {
	db := GetDBAddr()
	createTableCategories(db)

	query := `INSERT OR IGNORE INTO relations_categories (post_id, category) VALUES (?, ?)`
	//TODO
	_, err := db.Exec(query, -1, category)
	if err != nil {
		fmt.Println("Am I here?")
		return err
	}
	fmt.Println(query)

	return nil
}

func DeleteCategories(postID int) error {
	db := GetDBAddr()
	createTableCategories(db)
	query := `DELETE FROM relations_categories WHERE post_id = ?`
	_, err := db.Exec(query, postID)
	if err != nil {
		return err
	}
	return nil
}

func DeleteCategory(category string) error {
	db := GetDBAddr()
	createTableCategories(db)
	query := `DELETE FROM relations_categories WHERE category = ?`
	_, err := db.Exec(query, category)
	if err != nil {
		return err
	}
	return nil
}

func GetCategoriesByPostID(postID int) (*[]string, error) {
	db := GetDBAddr()
	createTableCategories(db)

	var categories []string
	query := "SELECT category FROM relations_categories WHERE post_id = ?"
	rows, err := db.Query(query, postID)
	if err != nil {
		return &categories, err
	}
	defer rows.Close()

	for rows.Next() {
		var category string
		err = rows.Scan(&category)
		if err != nil {
			return &categories, err
		}
		categories = append(categories, category)
	}

	err = rows.Err()
	if err != nil {
		return &categories, err
	}
	return &categories, nil
}

func createTableComments(db *sql.DB) {
	commentsTable := `CREATE TABLE IF NOT EXISTS comments (
        id 			INTEGER PRIMARY KEY AUTOINCREMENT,
        text 	    TEXT NOT NULL,
        user_id 	INTEGER REFERENCES users (id),
        post_id 	INTEGER REFERENCES posts (id),
        created_at 	DATE DEFAULT CURRENT_TIMESTAMP NOT NULL
		);`
	query, err := db.Prepare(commentsTable)
	if err != nil {
		log.Fatal(err)
	}
	query.Exec()
}

func GetCommentsByPostId(currentUserID, postID int) (*[]Comment, error) {
	db := GetDBAddr()
	createTableComments(db)

	var comments []Comment

	query := "SELECT id, text, user_id, post_id, created_at FROM comments WHERE post_id = ?"
	rows, err := db.Query(query, postID)
	if err != nil {
		return &comments, err
	}
	defer rows.Close()

	for rows.Next() {
		var comment Comment
		var userID, postID int
		var time time.Time
		err = rows.Scan(&comment.Id, &comment.Text, &userID, &postID, &time)
		if err != nil {
			return &comments, err
		}

		creator, _ := GetUserById(userID)
		comment.Creator = *creator

		likes, dislikes, _ := GetAllLikesDislikesByCommentID(comment.Id)
		comment.Likes = likes
		comment.Dislikes = dislikes
		comment.CreatedAt = time.Format("2006-01-02 15:04:05")

		if currentUserID == userID {
			comment.CurrentUsersComment = true
		}

		comments = append(comments, comment)
	}

	err = rows.Err()
	if err != nil {
		return &comments, err
	}

	return &comments, nil
}

func GetCommentByPostId(postID, commentID int) string {
	db := GetDBAddr()
	var comment string
	query := "SELECT text FROM comments WHERE id = ? AND post_id = ?"
	res := db.QueryRow(query, commentID, postID)
	res.Scan(&comment)
	return comment
}

func AddComment(postID, userID int, comment string) (int, error) {
	db := GetDBAddr()
	createTableComments(db)

	stmt := "INSERT INTO comments(text, user_id, post_id) VALUES (?, ?, ?)"

	res, err := db.Exec(stmt, comment, userID, postID)
	if err != nil {
		return -1, err
	}
	err = AddNotification(userID, postID, 2)
	if err != nil {
		return -1, err
	}
	commentID, _ := res.LastInsertId()
	return int(commentID), nil
}

func EditComment(postID, userID int, text, newText string) error {
	db := GetDBAddr()
	createTableComments(db)

	stmt := "UPDATE comments SET text = ? WHERE text = ? AND user_id = ? AND post_id = ?"
	_, err := db.Exec(stmt, newText, text, userID, postID)
	if err != nil {
		return err
	}
	return nil
}

func DeleteComment(postID, commentID int) error {
	db := GetDBAddr()
	createTableComments(db)
	stmt := "DELETE FROM comments WHERE post_id = ? AND id = ?"

	_, err := db.Exec(stmt, postID, commentID)
	if err != nil {
		return err
	}
	return nil
}

func GetAllCommentsByUserId(userID int) ([]Comment, error) {
	db := GetDBAddr()
	createTableComments(db)

	var comments []Comment

	query := "SELECT id, text, user_id, post_id, created_at FROM comments WHERE user_id = ?"
	rows, err := db.Query(query, userID)
	if err != nil {
		return comments, err
	}
	defer rows.Close()

	for rows.Next() {
		var comment Comment
		var userID, postID int
		var time time.Time
		err = rows.Scan(&comment.Id, &comment.Text, &userID, &postID, &time)
		if err != nil {
			return comments, err
		}

		creator, _ := GetUserById(userID)
		comment.Creator = *creator

		likes, dislikes, _ := GetAllLikesDislikesByCommentID(comment.Id)
		comment.Likes = likes
		comment.Dislikes = dislikes
		comment.CreatedAt = time.Format("2006-01-02 15:04:05")

		post, err := GetPostById(userID, postID)
		if err != nil {
			comment.Post = Post{}
		} else {
			comment.Post = *post
		}

		comments = append(comments, comment)
	}

	err = rows.Err()
	if err != nil {
		return comments, err
	}

	return comments, nil
}

func createTableCookies(db *sql.DB) {
	cookies_table := `CREATE TABLE IF NOT EXISTS cookies (
		user_id         INTEGER REFERENCES users (id),
        name            TEXT NOT NULL,
        value           TEXT NOT NULL UNIQUE,
        expires         DATETIME NOT NULL
    );`

	_, err := db.Exec(cookies_table)
	if err != nil {
		// change the err
		log.Fatal(err)
	}

}

func AddCookie(userID int) (*http.Cookie, error) {
	cookie := helpers.CreateCookie()

	db := GetDBAddr()
	createTableCookies(db)

	stmt, err := db.Prepare("INSERT INTO cookies (user_id, name, value, expires) VALUES(?,?,?,?)")
	if err != nil {
		return &http.Cookie{}, err
	}
	defer stmt.Close()

	_, err = stmt.Exec(userID, &cookie.Name, &cookie.Value, &cookie.Expires)
	if err != nil {
		return &http.Cookie{}, err
	}

	return cookie, nil
}

func CheckCookie(cookie *http.Cookie) (User, error) {
	db := GetDBAddr()
	createTableCookies(db)

	var userID int
	formDB := &http.Cookie{}
	row := db.QueryRow("SELECT * FROM cookies WHERE value = ?", &cookie.Value)
	err := row.Scan(&userID, &formDB.Name, &formDB.Value, &formDB.Expires)
	if err != nil {
		return User{}, err
	}

	if formDB.Value != cookie.Value || time.Now().After(formDB.Expires) {
		return User{}, errors.New("not valid cookie")
	}

	user, err := GetUserById(userID)
	if err != nil {
		return User{}, err
	}

	return *user, nil
}

func DeleteCookie(userID int) error {
	db := GetDBAddr()
	createTableCookies(db)

	stmt, err := db.Prepare("DELETE FROM cookies WHERE user_id=?")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(userID)
	return err
}

func createTableRelationLikes(db *sql.DB) {
	table := `CREATE TABLE IF NOT EXISTS relations_likes (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        user_id INTEGER REFERENCES users (id),
        post_id INTEGER REFERENCES posts (id),
        mark INTEGER NOT NULL
    );`
	_, err := db.Exec(table)
	if err != nil {
		log.Fatal(err)
	}
}

func AddLikeDislike(userID, postID, mark int) error {
	db := GetDBAddr()
	createTableRelationLikes(db)

	query := `INSERT INTO relations_likes (user_id, post_id, mark) VALUES (?, ?, ?)`

	_, err := db.Exec(query, userID, postID, mark)
	if err != nil {
		return err
	}
	err = AddNotification(userID, postID, mark)
	if err != nil {
		return err
	}

	return nil
}

func UpdateLikeDislike(userID, postID, mark int) error {
	db := GetDBAddr()
	createTableRelationLikes(db)

	var action string

	oldMark := IsLikePostIdUserId(userID, postID)
	switch oldMark {
	case 1:
		action = "liked"
	case -1:
		action = "disliked"
	}

	var oldNotificationId int
	query := `SELECT id FROM notifications_requests WHERE requestor_id = ? AND post_id = ? AND action = ?`
	err := db.QueryRow(query, userID, postID, action).Scan(&oldNotificationId)
	if err != nil {
		fmt.Println(err)
	}

	query = `UPDATE relations_likes SET mark = ? WHERE user_id = ? AND post_id = ?`
	_, err = db.Exec(query, mark, userID, postID)
	if err != nil {
		return err
	}

	DeleteNotification(oldNotificationId)
	AddNotification(userID, postID, mark)

	return nil
}

func DeleteLikeDislike(userID, postID, mark int) error {
	db := GetDBAddr()
	createTableRelationLikes(db)

	query := `DELETE FROM relations_likes WHERE user_id = ? AND post_id = ?`
	_, err := db.Exec(query, userID, postID)
	if err != nil {
		return err
	}

	err = DeleteNotificationByPostIdUserId(userID, postID)
	if err != nil {
		return err
	}
	return nil
}

func IsLikePostIdUserId(userID, postID int) int {
	db := GetDBAddr()
	createTableRelationLikes(db)

	var mark int

	query := "SELECT mark FROM relations_likes WHERE post_id = ? AND user_id = ?"
	err := db.QueryRow(query, postID, userID).Scan(&mark)
	if err != nil {
		return 0
	}

	return mark
}

func GetAllLikesDislikesByPostID(postID int) (int, int, error) {
	db := GetDBAddr()
	createTableRelationLikes(db)

	var likes, dislikes int

	query := "SELECT * FROM relations_likes WHERE post_id = ?"

	rows, err := db.Query(query, postID)

	if err != nil {
		return 0, 0, err
	}
	defer rows.Close()

	for rows.Next() {
		var relationMarks RelationLike
		err = rows.Scan(&relationMarks.Id, &relationMarks.UserId, &relationMarks.PostId, &relationMarks.Mark)
		if err != nil {
			return 0, 0, err
		}
		if relationMarks.Mark > 0 {
			likes++
		} else {
			dislikes++
		}
	}

	err = rows.Err()
	if err != nil {
		return 0, 0, err
	}

	return likes, dislikes, nil
}

func GetAllPostsLikedDislikedByUserId(userID int) ([]Post, error) {
	db := GetDBAddr()
	createTableRelationLikes(db)

	var posts []Post

	query := "SELECT * FROM relations_likes WHERE user_id = ?"

	rows, err := db.Query(query, userID)

	if err != nil {
		return posts, err
	}
	defer rows.Close()

	for rows.Next() {
		var relationMarks RelationLike
		err = rows.Scan(&relationMarks.Id, &relationMarks.UserId, &relationMarks.PostId, &relationMarks.Mark)
		if err == nil {
			post, err := GetPostById(userID, relationMarks.PostId)
			if err == nil {
				posts = append(posts, *post)
			}
		}

	}

	err = rows.Err()
	if err != nil {
		return posts, err
	}

	return posts, nil
}

func createTableRelationLikesComments(db *sql.DB) {
	table := `CREATE TABLE IF NOT EXISTS relations_likes_comments (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        user_id INTEGER REFERENCES users (id),
        comment_id INTEGER REFERENCES comments (id),
        mark INTEGER NOT NULL
    );`
	_, err := db.Exec(table)
	if err != nil {
		log.Fatal(err)
	}
}

func AddLikeDislikeComment(userID, commentID, mark int) error {
	db := GetDBAddr()
	createTableRelationLikesComments(db)

	query := `INSERT INTO relations_likes_comments (user_id, comment_id, mark) VALUES (?, ?, ?)`
	_, err := db.Exec(query, userID, commentID, mark)
	if err != nil {
		return err
	}

	return nil
}

func UpdateLikeDislikeComments(userID, commentID, mark int) error {
	db := GetDBAddr()
	createTableRelationLikesComments(db)

	query := `UPDATE relations_likes_comments SET mark = ? WHERE user_id = ? AND comment_id = ?`
	_, err := db.Exec(query, mark, userID, commentID)
	if err != nil {
		return err
	}

	return nil
}

func DeleteLikeDislikeComment(userID, commentID int) error {
	db := GetDBAddr()
	createTableRelationLikesComments(db)

	query := `DELETE FROM relations_likes_comments WHERE user_id = ? AND comment_id = ?`
	_, err := db.Exec(query, userID, commentID)

	if err != nil {
		return err
	}

	return nil
}

func IsLikeCommentIdUserId(userID, commentID int) int {
	db := GetDBAddr()
	createTableRelationLikesComments(db)

	var mark int

	query := "SELECT mark FROM relations_likes_comments WHERE comment_id = ? AND user_id = ?"
	err := db.QueryRow(query, commentID, userID).Scan(&mark)
	if err != nil {
		return 0
	}

	return mark
}

func GetAllLikesDislikesByCommentID(commentID int) (int, int, error) {
	db := GetDBAddr()
	createTableRelationLikesComments(db)

	var likes, dislikes int

	query := "SELECT mark FROM relations_likes_comments WHERE comment_id = ?"
	rows, err := db.Query(query, commentID)
	if err != nil {
		return 0, 0, err
	}
	defer rows.Close()

	for rows.Next() {
		var mark int
		err = rows.Scan(&mark)
		if err != nil {
			return 0, 0, err
		}
		if mark > 0 {
			likes++
		} else {
			dislikes++
		}

	}

	err = rows.Err()
	if err != nil {
		return 0, 0, err
	}

	return likes, dislikes, nil
}

func createTableNotifications(db *sql.DB) {
	table := `CREATE TABLE IF NOT EXISTS notifications_requests (
			id 				INTEGER PRIMARY KEY AUTOINCREMENT,
			reciver_id      INTEGER REFERENCES users (id),
			requestor_id    INTEGER REFERENCES users (id),
			action          TEXT NOT NULL,
			post_id         INTEGER REFERENCES posts (id),
			message         TEXT,
			seen            INTEGER DEFAULT 0,
			created_at 		DATE DEFAULT CURRENT_TIMESTAMP NOT NULL
		);`
	_, err := db.Exec(table)
	if err != nil {
		log.Fatal(err)
	}
}

func AddNotification(userID, postID, mark int) error {
	db, action, reciverID, err := getActionAndReceiver(mark, postID)

	query := `INSERT INTO notifications_requests (reciver_id, requestor_id, action, post_id, message) VALUES (?, ?, ?, ? , ?)`
	if userID == reciverID {
		return nil
	}
	_, err = db.Exec(query, reciverID, userID, action, postID, "")
	if err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}

func AddRequest(userID, postID int, action, message string) error {
	db := GetDBAddr()
	admin, _ := getAdmin()

	query := `INSERT INTO notifications_requests (reciver_id, requestor_id, action, post_id, message) VALUES (?, ?, ?, ?, ?)`
	_, err := db.Exec(query, admin.Id, userID, action, postID, message)
	if err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}

func AddResponse(userID, postID int, action, message string) error {
	db := GetDBAddr()
	admin, _ := getAdmin()
	query := `INSERT INTO notifications_requests (reciver_id, requestor_id, action, post_id, message) VALUES (?, ?, ?, ?, ?)`
	_, err := db.Exec(query, userID, admin.Id, action, postID, message)
	if err != nil {
		fmt.Println(err)
		return err
	}
	fmt.Println("I'm here")
	return nil
}

func IsRequestAlreadyExisted(userID, postID int, action string) error {
	db := GetDBAddr()
	admin, _ := getAdmin()
	query := `SELECT id FROM notifications_requests WHERE reciver_id = ? AND requestor_id = ? AND action = ? AND post_id = ?`
	res, err := db.Exec(query, admin.Id, userID, action, postID)
	fmt.Println("res", res)
	if err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}

func UpdateNotifications(userID, postID, mark int) error {
	db, action, reciverID, _ := getActionAndReceiver(mark, postID)

	query := `UPDATE notifications_requests SET action = ? WHERE requestor_id = ? AND post_id = ? AND reciver_id = ?`
	_, err := db.Exec(query, action, userID, postID, reciverID)
	if err != nil {
		return err
	}

	return nil
}

func DeleteNotification(id int) error {
	db := GetDBAddr()
	createTableNotifications(db)

	query := `DELETE FROM notifications_requests WHERE id = ?`
	_, err := db.Exec(query, id)

	if err != nil {
		return err
	}

	return nil
}

func DeleteNotificationByPostIdUserId(userID, postID int) error {
	db := GetDBAddr()
	createTableNotifications(db)

	query := `DELETE FROM notifications_requests WHERE requestor_id = ? AND post_id = ?`
	_, err := db.Exec(query, userID, postID)

	if err != nil {
		return err
	}

	return nil
}

func getActionAndReceiver(mark, postID int) (*sql.DB, string, int, error) {
	db := GetDBAddr()
	createTableNotifications(db)
	post, err := GetPostById(0, postID)
	if err != nil {
		return nil, "", 0, err
	}
	recieverID := post.Creator.Id
	var action string
	switch mark {
	case 1:
		action = "liked"
	case -1:
		action = "disliked"
	case 2:
		action = "commented"
	// case 3:
	// 	action = "liked the comment"
	default:
		action = ""
	}
	return db, action, recieverID, nil
}

func GetNotifications(userID int) ([]Notification, error) {
	db := GetDBAddr()
	createTableNotifications(db)
	query := `SELECT * FROM notifications_requests WHERE reciver_id = ?`
	res, err := db.Query(query, userID)
	notifications := []Notification{}
	var notificationsToDelete []int
	if err != nil {
		return notifications, nil
	}

	for res.Next() {
		var notification Notification
		var requestorId int
		var postID int
		var message string

		err = res.Scan(&notification.Id, &notification.Reciver, &requestorId, &notification.Action, &postID, &message, &notification.Seen, &notification.CreatedAt)
		if err != nil {
			fmt.Println(err)
			return notifications, nil
		}

		user, err := GetUserById(requestorId)
		if err != nil {
			notification.WhoDid.Id = -1
			notification.WhoDid.Name = "unknown user"
		} else {
			notification.WhoDid = *user
		}

		notification.Message = message

		if postID != -1 {
			post, err := GetPostById(userID, postID)
			if err != nil {
				notificationsToDelete = append(notificationsToDelete, notification.Id)
			} else {
				notification.PostID = post.Id
				notification.PostTitle = post.Title
				notifications = append(notifications, notification)
			}
		} else {
			notification.PostID = -1
			notification.PostTitle = ""
			notifications = append(notifications, notification)
		}

	}

	for _, id := range notificationsToDelete {
		DeleteNotification(id)
	}

	return notifications, nil
}

func GetUserRequest(userID int) (int, error) {
	db := GetDBAddr()
	createTableNotifications(db)

	var id *int

	query := `SELECT id FROM notifications_requests WHERE action = ? AND requestor_id = ?`
	// res, err := db.Query(query, userID, "request")
	err := db.QueryRow(query, "request", userID).Scan(&id)
	if err != nil {
		return -1, nil
	}

	return *id, nil
}

func SetNotificationSeen(notificationID int) error {
	db := GetDBAddr()
	createTableNotifications(db)
	stmt := "UPDATE notifications_requests SET seen = 1 WHERE id = ?"

	_, err := db.Exec(stmt, notificationID)
	if err != nil {
		return err
	}

	return nil
}

func GetNotificationId(id int) (*Notification, error) {
	db := GetDBAddr()
	createTablePosts(db)

	var request Notification
	var requestorId int
	var seen int
	var time time.Time
	row := db.QueryRow("SELECT * FROM notifications_requests WHERE id = ?", id)
	err := row.Scan(&request.Id, &request.Reciver, &requestorId, &request.Action, &request.PostID, &request.Message, &seen, &time)
	if err != nil {
		return nil, err
	}

	requestor, _ := GetUserById(requestorId)

	request.WhoDid = *requestor
	if seen == 0 {
		request.Seen = false
	} else {
		request.Seen = true
	}
	request.CreatedAt = time.Format("January 02, 2006")
	return &request, nil
}
