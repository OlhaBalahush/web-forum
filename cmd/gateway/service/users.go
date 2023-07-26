package service

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/OlhaBalahush/web-forum/cmd/gateway/helpers"
)

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
