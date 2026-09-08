package database

import (
	"fmt"
	"jobFlow/models"
)

func CreateSkill(skill models.Skill) (int, error) {
	result, err := DB.Exec(`
		INSERT INTO skills (name)
		VALUES (?)
	`, skill.Name)

	if err != nil {
		return 0, fmt.Errorf("creating skill: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("getting skill ID: %w", err)
	}

	return int(id), nil
}

func GetSkillByID(id int) (models.Skill, error) {
	var skill models.Skill

	err := DB.QueryRow(`
		SELECT id, name
		FROM skills
		WHERE id = ?
	`, id).Scan(
		&skill.ID,
		&skill.Name,
	)

	if err != nil {
		return models.Skill{}, fmt.Errorf("getting skill: %w", err)
	}

	return skill, nil
}

func GetSkillByName(name string) (models.Skill, error) {
	var skill models.Skill

	err := DB.QueryRow(`
		SELECT id, name
		FROM skills
		WHERE name = ?
	`, name).Scan(
		&skill.ID,
		&skill.Name,
	)

	if err != nil {
		return models.Skill{}, fmt.Errorf("getting skill by name: %w", err)
	}

	return skill, nil
}

func GetSkills() ([]models.Skill, error) {
	rows, err := DB.Query(`
		SELECT id, name
		FROM skills
		ORDER BY name ASC
	`)

	if err != nil {
		return nil, fmt.Errorf("getting skills: %w", err)
	}
	defer rows.Close()

	var skills []models.Skill

	for rows.Next() {
		var skill models.Skill

		if err := rows.Scan(
			&skill.ID,
			&skill.Name,
		); err != nil {
			return nil, fmt.Errorf("scanning skill: %w", err)
		}

		skills = append(skills, skill)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating skills: %w", err)
	}

	return skills, nil
}
