package models

type FeedComment struct {
	Comment

	AuthorFirstName string
	AuthorLastName  string
	AuthorUsername  string
}

type FeedPost struct {
	Post

	// Person who appears at the top of the feed item.
	AuthorFirstName string
	AuthorLastName  string
	AuthorUsername  string

	// Engagement information.
	LikeCount    int
	CommentCount int
	ShareCount   int
	HasLiked     bool

	// Original post information.
	// These are populated only when this post is a shared post.
	OriginalAuthorFirstName string
	OriginalAuthorLastName  string
	OriginalAuthorUsername  string
	OriginalContent         string
	OriginalType            string
}
