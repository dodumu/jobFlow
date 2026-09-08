package models

type ProfilePageData struct {
	User         User
	Profile      UserProfile
	Skills       []Skill
	Experiences  []Experience
	Education    []Education
	Company      *Company
}