package services

import (
	"golang.org/x/crypto/bcrypt"
	"todolist/database"
	"todolist/models"
)

func CreateUser(user *models.User) (map[string]interface{}, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	user.Password = string(hashedPassword)

	// Insert user data into the database
	query := "INSERT INTO users (username, password) VALUES (:1, :2)"
	_, err = database.DB.Exec(query, user.Username, user.Password)
	if err != nil {
		return nil, err
	}

	query = "SELECT id FROM users WHERE username = :1 AND ROWNUM = 1 ORDER BY id DESC"
	err = database.DB.QueryRow(query, user.Username).Scan(&user.ID)
	if err != nil {
		return nil, err
	}

	// Prepare the response data
	data := map[string]interface{}{
		"id":       user.ID,
		"username": user.Username,
	}
	return data, nil
}
