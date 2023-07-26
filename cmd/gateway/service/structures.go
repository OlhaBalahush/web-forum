package service

import (
	"database/sql"
	"log"

	"github.com/OlhaBalahush/web-forum/cmd/gateway/db"
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
