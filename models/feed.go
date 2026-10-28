package models

type FeedComment struct {
	Comment

	AuthorFirstName      string
	AuthorLastName       string
	AuthorUsername       string
	AuthorProfilePicture string
}

type FeedPost struct {
	Post

	// Author identity.
	AuthorType string
	AuthorID   int
	CompanyID  int

	// Individual author.
	AuthorFirstName      string
	AuthorLastName       string
	AuthorUsername       string
	AuthorProfilePicture string
	AuthorCompanyID      int

	// Company author.
	CompanyName string
	CompanyLogo string

	// Engagement information.
	LikeCount    int
	CommentCount int
	ShareCount   int
	HasLiked     bool

	//Authorization
	CanDelete bool

	// Original post information.
	OriginalAuthorType      string
	OriginalAuthorID        int
	OriginalCompanyID       int
	OriginalAuthorFirstName string
	OriginalAuthorLastName  string
	OriginalAuthorUsername  string
	OriginalCompanyName     string
	OriginalCompanyLogo     string
	OriginalContent         string
	OriginalType            string
}

type HomePageData struct {
	User          User
	CurrentUserID int
	Posts         []FeedPost
	HideNavbar    bool
}
