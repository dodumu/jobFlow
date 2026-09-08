package models

type FeedPost struct {
	Post

	AuthorFirstName string
	AuthorLastName  string
	AuthorUsername  string

	LikeCount    int
	CommentCount int
	ShareCount   int

	HasLiked bool
}
