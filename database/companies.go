package database

import (
	"fmt"
	"jobFlow/models"
)

func CreateCompany(company models.Company) (int, error) {
	query := `INSERT INTO companies (user_id, company_name, description, website, location, logo) VALUES(?, ?, ?, ?, ?, ?)`
	res, err := DB.Exec(query, company.UserID, company.CompanyName, company.Description, company.Website, company.Location, company.Logo)
	if err != nil {
		return 0, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	return int(id), nil
}

func GetCompanyByID(id int) (models.Company, error) {
	query := `SELECT id, user_id, company_name, description, website, location, logo, created_at  FROM companies  WHERE id = ?`
	var company models.Company
	err := DB.QueryRow(query, id).Scan(&company.ID,
		&company.UserID,
		&company.CompanyName,
		&company.Description,
		&company.Website,
		&company.Location,
		&company.Logo,
		&company.CreatedAt)
	if err != nil {
		return models.Company{}, err
	}
	return company, nil
}

func GetCompanyByUserID(userID int) (models.Company, error) {
	query := `SELECT id, user_id, company_name, description, website, location, logo, created_at  FROM companies  WHERE user_id = ?`
	var company models.Company
	err := DB.QueryRow(query, userID).Scan(&company.ID,
		&company.UserID,
		&company.CompanyName,
		&company.Description,
		&company.Website,
		&company.Location,
		&company.Logo,
		&company.CreatedAt)
	if err != nil {
		return models.Company{}, err
	}
	return company, nil
}

func GetCompanyByName(name string) (models.Company, error) {
	query := `SELECT id, user_id, company_name, description, website, location, logo, created_at  FROM companies  WHERE company_name = ?`
	var company models.Company
	err := DB.QueryRow(query, name).Scan(&company.ID,
		&company.UserID,
		&company.CompanyName,
		&company.Description,
		&company.Website,
		&company.Location,
		&company.Logo,
		&company.CreatedAt)
	if err != nil {
		return models.Company{}, err
	}
	return company, nil
}

func UpdateCompany(id int, company models.Company) error {
	query := `UPDATE companies SET company_name = ?, description = ?, website = ?,  location = ?, logo = ?  WHERE id = ?`

	res, err := DB.Exec(query, company.CompanyName, company.Description, company.Website, company.Location, company.Logo, id)

	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("company with id %d not found", id)
	}
	return nil
}
