package service

import (
	"database/sql"
	"log"
)

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
