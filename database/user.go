package database

import (
	"jobFlow/models"
)

func CreateUser(user models.User) (int, error) {
	query := `INSERT INTO users (username, password_hash, name, email, date_of_birth, role)  VALUES(?, ?, ?, ?, ?, ?)`
	res, err := DB.Exec(query, user.Username, user.PasswordHash, user.Name, user.Email, user.DOB, user.Role)
	if err != nil {
		return 0, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	return int(id), nil
}

func GetUserByID(id int) (models.User, error) {
	query := `SELECT id, username, password_hash, name, email, date_of_birth, role, created_at  FROM users  WHERE id = ?`
	var user models.User
	err := DB.QueryRow(query, id).Scan(&user.ID,
		&user.Username,
		&user.PasswordHash,
		&user.Name,
		&user.Email,
		&user.DOB,
		&user.Role,
		&user.CreatedAt)
	if err != nil {
		return models.User{}, err
	}
	return user, nil
}

func GetUserByEmail(email string) (models.User, error) {
	query := `SELECT id, username, password_hash, name, email, date_of_birth, role, created_at  FROM users  WHERE email = ?`
	var user models.User
	err := DB.QueryRow(query, email).Scan(&user.ID,
		&user.Username,
		&user.PasswordHash,
		&user.Name,
		&user.Email,
		&user.DOB,
		&user.Role,
		&user.CreatedAt)
	if err != nil {
		return models.User{}, err
	}
	return user, nil
}

func GetUserByUsername(username string) (models.User, error) {
	query := `SELECT id, username, password_hash, name, email, date_of_birth, role, created_at  FROM users  WHERE username = ?`
	var user models.User
	err := DB.QueryRow(query, username).Scan(&user.ID,
		&user.Username,
		&user.PasswordHash,
		&user.Name,
		&user.Email,
		&user.DOB,
		&user.Role,
		&user.CreatedAt)
	if err != nil {
		return models.User{}, err
	}
	return user, nil
}
