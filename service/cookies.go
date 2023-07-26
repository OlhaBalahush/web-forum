package service

import (
	"database/sql"
	"errors"
	"log"
	"main/helpers"
	"net/http"
	"time"
)

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
