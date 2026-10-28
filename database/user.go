package database

import (
	"fmt"
	"jobFlow/constants"
	"jobFlow/models"
)

func createUser(exec DBExecutor, user models.User) (int, error) {
	query := `
		INSERT INTO users (
			username,
			password_hash,
			first_name,
			last_name,
			email,
			date_of_birth,
			role
		)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`

	res, err := exec.Exec(
		query,
		user.Username,
		user.PasswordHash,
		user.FirstName,
		user.LastName,
		user.Email,
		user.DOB,
		user.Role,
	)
	if err != nil {
		return 0, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	return int(id), nil
}

//	func CreateUser(user models.User) (int, error) {
//		return createUser(DB, user)
//	}
func GetUserByID(id int) (models.User, error) {
	query := `SELECT id, first_name, last_name, username, password_hash, email, date_of_birth, role, created_at
	FROM users
	WHERE id = ?`
	var user models.User
	err := DB.QueryRow(query, id).Scan(&user.ID,
		&user.FirstName,
		&user.LastName,
		&user.Username,
		&user.PasswordHash,
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
	query := `SELECT id, username, password_hash, first_name, last_name, email, date_of_birth, role, created_at  FROM users  WHERE email = ?`
	var user models.User
	err := DB.QueryRow(query, email).Scan(&user.ID,
		&user.Username,
		&user.PasswordHash,
		&user.FirstName,
		&user.LastName,
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
	query := `SELECT id, username, password_hash, first_name, last_name, email, date_of_birth, role, created_at  FROM users  WHERE username = ?`
	var user models.User
	err := DB.QueryRow(query, username).Scan(&user.ID,
		&user.Username,
		&user.PasswordHash,
		&user.FirstName,
		&user.LastName,
		&user.Email,
		&user.DOB,
		&user.Role,
		&user.CreatedAt)
	if err != nil {
		return models.User{}, err
	}
	return user, nil
}

func GetAllIndividualUsers() ([]models.User, error) {
	query := `SELECT id, username, password_hash, first_name, last_name, email, date_of_birth, role, created_at FROM users WHERE role = ?`

	var allUsers []models.User

	res, err := DB.Query(query, constants.RoleIndividual)

	if err != nil {
		return nil, err
	}
	defer res.Close()
	for res.Next() {
		var user models.User

		err := res.Scan(&user.ID,
			&user.Username,
			&user.PasswordHash,
			&user.FirstName,
			&user.LastName,
			&user.Email,
			&user.DOB,
			&user.Role,
			&user.CreatedAt)

		if err != nil {
			return nil, err
		}

		allUsers = append(allUsers, user)

	}
	if err := res.Err(); err != nil {
		return nil, fmt.Errorf("iterating users: %w", err)
	}
	return allUsers, nil
}
