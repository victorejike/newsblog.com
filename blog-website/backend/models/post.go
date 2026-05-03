package models

import "time"

type Post struct {
	ID             int       `json:"id"`
	Title          string    `json:"title"`
	Content        string    `json:"content"`
	Excerpt        string    `json:"excerpt,omitempty"`
	AuthorID       int       `json:"author_id"`
	Author         string    `json:"author,omitempty"`
	CategoryID     int       `json:"category_id,omitempty"`
	Category       string    `json:"category,omitempty"`
	CoverImageURL  string    `json:"cover_image_url,omitempty"`
	ReadingTime    int       `json:"reading_time,omitempty"`
	ViewCount      int       `json:"view_count"`
	LikeCount      int       `json:"like_count,omitempty"`
	CommentCount   int       `json:"comment_count,omitempty"`
	Tags           []string  `json:"tags,omitempty"`
	Published      bool      `json:"published"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type CreatePostRequest struct {
	Title         string   `json:"title"`
	Content       string   `json:"content"`
	Excerpt       string   `json:"excerpt"`
	CategoryID    int      `json:"category_id"`
	CoverImageURL string   `json:"cover_image_url"`
	Tags          []string `json:"tags"`
	Published     bool     `json:"published"`
}

type UpdatePostRequest struct {
	Title         string   `json:"title"`
	Content       string   `json:"content"`
	Excerpt       string   `json:"excerpt"`
	CategoryID    int      `json:"category_id"`
	CoverImageURL string   `json:"cover_image_url"`
	Tags          []string `json:"tags"`
	Published     bool     `json:"published"`
}

type Category struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Slug        string    `json:"slug"`
	Description string    `json:"description,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

type Tag struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Slug      string    `json:"slug"`
	CreatedAt time.Time `json:"created_at"`
}

type Comment struct {
	ID        int       `json:"id"`
	PostID    int       `json:"post_id"`
	UserID    int       `json:"user_id"`
	Username  string    `json:"username"`
	ParentID  *int      `json:"parent_id,omitempty"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Replies   []Comment `json:"replies,omitempty"`
}

type CreateCommentRequest struct {
	Content  string `json:"content"`
	ParentID *int   `json:"parent_id,omitempty"`
}

type Like struct {
	ID        int       `json:"id"`
	UserID    int       `json:"user_id"`
	PostID    int       `json:"post_id"`
	CreatedAt time.Time `json:"created_at"`
}

type Bookmark struct {
	ID        int       `json:"id"`
	UserID    int       `json:"user_id"`
	PostID    int       `json:"post_id"`
	CreatedAt time.Time `json:"created_at"`
}

type NewsletterSubscriber struct {
	ID            int       `json:"id"`
	Email         string    `json:"email"`
	IsActive      bool      `json:"is_active"`
	SubscribedAt  time.Time `json:"subscribed_at"`
	UnsubscribedAt *time.Time `json:"unsubscribed_at,omitempty"`
}
