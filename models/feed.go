package models

type FeedComment struct {
	Comment

	AuthorFirstName string
	AuthorLastName  string
	AuthorUsername  string
}

type FeedPost struct {
	Post

	// Author identity.
	AuthorType string

	// Individual author.
	AuthorFirstName string
	AuthorLastName  string
	AuthorUsername  string

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
	OriginalAuthorFirstName string
	OriginalAuthorLastName  string
	OriginalAuthorUsername  string
	OriginalCompanyName     string
	OriginalCompanyLogo     string
	OriginalContent         string
	OriginalType            string
}
