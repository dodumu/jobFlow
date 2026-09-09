package models

type FeedComment struct {
	Comment

	AuthorFirstName string
	AuthorLastName  string
	AuthorUsername  string
}

type FeedPost struct {
	Post

	AuthorFirstName string
	AuthorLastName  string
	AuthorUsername  string

	LikeCount    int
	CommentCount int
	ShareCount   int

	HasLiked bool

	Comments []FeedComment
}
