package service

import (
	"database/sql"
	"log"
	"time"
)

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
