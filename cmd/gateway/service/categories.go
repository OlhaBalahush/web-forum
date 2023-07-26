package service

import (
	"database/sql"
	"fmt"
	"log"
)

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
