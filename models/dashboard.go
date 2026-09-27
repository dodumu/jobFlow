package models

type DashboardPageData struct {
	User          User
	CurrentUserID int
	Posts         []FeedPost
}
