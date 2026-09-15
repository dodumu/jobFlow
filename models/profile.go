package models

type ProfilePageData struct {
	User             User
	Profile          UserProfile
	Company          *Company
	Skills           []Skill
	Experiences      []Experience
	Education        []Education
	Preferences      UserPreference
	EmploymentTypes  []string
	WorkArrangements []string
	Posts            []FeedPost

	FollowerCount  int
	FollowingCount int
}
