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
	HideNavbar     bool
}

type EditProfilePageData struct {
	Profile    UserProfile
	Skills     []Skill
	Error      string
	HideNavbar bool
}
