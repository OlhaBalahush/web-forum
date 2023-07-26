package service

import (
	"database/sql"
	"fmt"
	"log"
)

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
