package database

import (
	"fmt"
	"jobFlow/models"
	"strings"
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

func AddSkillsByName(userID int, skillName string) error {
	skillName = strings.TrimSpace(skillName)
	if skillName == "" {
		return fmt.Errorf("Skill cannot be empty")
	}
	query := `INSERT INTO skills (name)
			  VALUES (?)
			  ON CONFLICT(name) DO NOTHING
	`

	_, err := DB.Exec(query, skillName)
	if err != nil {
		return fmt.Errorf("creating skill: %w", err)
	}
	var skillID int

	err = DB.QueryRow(`SELECT id FROM skills
		WHERE name = ?
		`, skillName).Scan(&skillID)

	if err != nil {
		return fmt.Errorf("getting skill ID: %w", err)
	}

	err = AddSkillToUser(userID, skillID)
	if err != nil {
		return fmt.Errorf("assinging skill to user: %w", err)
	}
	return nil
}

func UpdateUserSkills(userID int, newSkills []string, removeSkillIDs []int) error {
	tx, err := DB.Begin()
	if err != nil {
		return fmt.Errorf("starting skill update transaction: %w", err)
	}
	defer tx.Rollback()

	// 1. Remove selected skills from this user only.
	for _, skillID := range removeSkillIDs {
		_, err := tx.Exec(`
            DELETE FROM user_skills
            WHERE user_id = ? AND skill_id = ?
        `, userID, skillID)
		if err != nil {
			return fmt.Errorf("removing skill %d: %w", skillID, err)
		}
	}

	// 2. Add each new skill, reusing existing skill records.
	for _, name := range newSkills {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}

		_, err := tx.Exec(`
            INSERT INTO skills (name)
            VALUES (?)
            ON CONFLICT(name) DO NOTHING
        `, name)
		if err != nil {
			return fmt.Errorf("creating skill %q: %w", name, err)
		}

		var skillID int
		err = tx.QueryRow(`
            SELECT id FROM skills WHERE name = ?
        `, name).Scan(&skillID)
		if err != nil {
			return fmt.Errorf("getting skill ID for %q: %w", name, err)
		}

		_, err = tx.Exec(`
            INSERT INTO user_skills (user_id, skill_id)
            VALUES (?, ?)
            ON CONFLICT(user_id, skill_id) DO NOTHING
        `, userID, skillID)
		if err != nil {
			return fmt.Errorf("assigning skill %q: %w", name, err)
		}
	}

	// 3. Commit all skill changes together.
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("saving skill changes: %w", err)
	}

	return nil
}
