package database

import (
	"fmt"
	"jobFlow/models"
)

func AddSkillToUser(userID, skillID int) error {
	_, err := DB.Exec(`
		INSERT INTO user_skills (user_id, skill_id)
		VALUES (?, ?)
	`, userID, skillID)

	if err != nil {
		return fmt.Errorf("adding skill to user: %w", err)
	}

	return nil
}

func RemoveSkillFromUser(userID, skillID int) error {
	result, err := DB.Exec(`
		DELETE FROM user_skills
		WHERE user_id = ? AND skill_id = ?
	`, userID, skillID)

	if err != nil {
		return fmt.Errorf("removing skill from user: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("checking removed skill: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("skill not assigned to user")
	}

	return nil
}

func GetSkillsByUserID(userID int) ([]models.Skill, error) {
	rows, err := DB.Query(`
		SELECT
			s.id,
			s.name
		FROM skills s
		JOIN user_skills us ON us.skill_id = s.id
		WHERE us.user_id = ?
		ORDER BY s.name ASC
	`)

	if err != nil {
		return nil, fmt.Errorf("getting user skills: %w", err)
	}
	defer rows.Close()

	var skills []models.Skill

	for rows.Next() {
		var skill models.Skill

		if err := rows.Scan(
			&skill.ID,
			&skill.Name,
		); err != nil {
			return nil, fmt.Errorf("scanning user skill: %w", err)
		}

		skills = append(skills, skill)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating user skills: %w", err)
	}

	return skills, nil
}
