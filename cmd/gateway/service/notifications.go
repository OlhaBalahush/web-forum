package service

import (
	"database/sql"
	"fmt"
	"log"
	"time"
)

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
