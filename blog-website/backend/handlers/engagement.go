package handlers

import (
	"blog-website/backend/database"
	"blog-website/backend/models"
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func GetPostComments(w http.ResponseWriter, r *http.Request) {
	postIDStr := r.URL.Path[len("/api/posts/"):]
	if strings.Contains(postIDStr, "/comments") {
		postIDStr = strings.TrimSuffix(postIDStr, "/comments")
	}

	postID, err := strconv.Atoi(postIDStr)
	if err != nil {
		http.Error(w, "Invalid post ID", http.StatusBadRequest)
		return
	}

	db := database.GetDB()
	rows, err := db.Query(`
		SELECT c.id, c.post_id, c.user_id, u.username, c.parent_id, c.content, c.created_at, c.updated_at
		FROM comments c
		JOIN users u ON c.user_id = u.id
		WHERE c.post_id = ? AND c.parent_id IS NULL
		ORDER BY c.created_at DESC
	`, postID)

	if err != nil {
		http.Error(w, "Failed to fetch comments", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var comments []models.Comment
	for rows.Next() {
		var comment models.Comment
		var parentID sql.NullInt64
		if err := rows.Scan(&comment.ID, &comment.PostID, &comment.UserID, &comment.Username,
			&parentID, &comment.Content, &comment.CreatedAt, &comment.UpdatedAt); err != nil {
			continue
		}
		if parentID.Valid {
			pid := int(parentID.Int64)
			comment.ParentID = &pid
		}

		// Fetch replies
		comment.Replies = getCommentReplies(db, comment.ID)
		comments = append(comments, comment)
	}

	json.NewEncoder(w).Encode(comments)
}

func getCommentReplies(db *sql.DB, parentID int) []models.Comment {
	rows, err := db.Query(`
		SELECT c.id, c.post_id, c.user_id, u.username, c.parent_id, c.content, c.created_at, c.updated_at
		FROM comments c
		JOIN users u ON c.user_id = u.id
		WHERE c.parent_id = ?
		ORDER BY c.created_at ASC
	`, parentID)

	if err != nil {
		return []models.Comment{}
	}
	defer rows.Close()

	var replies []models.Comment
	for rows.Next() {
		var reply models.Comment
		var replyParentID sql.NullInt64
		if err := rows.Scan(&reply.ID, &reply.PostID, &reply.UserID, &reply.Username,
			&replyParentID, &reply.Content, &reply.CreatedAt, &reply.UpdatedAt); err != nil {
			continue
		}
		if replyParentID.Valid {
			pid := int(replyParentID.Int64)
			reply.ParentID = &pid
		}
		replies = append(replies, reply)
	}

	return replies
}

func CreateComment(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	postIDStr := r.URL.Path[len("/api/posts/"):]
	if strings.Contains(postIDStr, "/comments") {
		postIDStr = strings.TrimSuffix(postIDStr, "/comments")
	}

	postID, err := strconv.Atoi(postIDStr)
	if err != nil {
		http.Error(w, "Invalid post ID", http.StatusBadRequest)
		return
	}

	var req models.CreateCommentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	db := database.GetDB()
	var commentID int
	err = db.QueryRow(
		"INSERT INTO comments (post_id, user_id, parent_id, content) VALUES (?, ?, ?, ?) RETURNING id",
		postID, userID, req.ParentID, req.Content,
	).Scan(&commentID)

	if err != nil {
		http.Error(w, "Failed to create comment", http.StatusInternalServerError)
		return
	}

	// Update comment count
	db.Exec("UPDATE posts SET comment_count = comment_count + 1 WHERE id = ?", postID)

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":   "Comment created successfully",
		"comment_id": commentID,
	})
}

func DeleteComment(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	commentIDStr := r.URL.Path[len("/api/comments/"):]
	commentID, err := strconv.Atoi(commentIDStr)
	if err != nil {
		http.Error(w, "Invalid comment ID", http.StatusBadRequest)
		return
	}

	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	db := database.GetDB()
	var postID int
	err = db.QueryRow("SELECT post_id FROM comments WHERE id = ?", commentID).Scan(&postID)
	if err != nil {
		http.Error(w, "Comment not found", http.StatusNotFound)
		return
	}

	// Check if user owns the comment or is admin
	var authorID int
	err = db.QueryRow("SELECT user_id FROM comments WHERE id = ?", commentID).Scan(&authorID)
	if err != nil {
		http.Error(w, "Comment not found", http.StatusNotFound)
		return
	}

	userIDInt, _ := strconv.Atoi(userID)
	if authorID != userIDInt {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	_, err = db.Exec("DELETE FROM comments WHERE id = ?", commentID)
	if err != nil {
		http.Error(w, "Failed to delete comment", http.StatusInternalServerError)
		return
	}

	// Update comment count
	db.Exec("UPDATE posts SET comment_count = comment_count - 1 WHERE id = ?", postID)

	json.NewEncoder(w).Encode(map[string]string{
		"message": "Comment deleted successfully",
	})
}

func LikePost(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	postIDStr := r.URL.Path[len("/api/posts/"):]
	if strings.Contains(postIDStr, "/like") {
		postIDStr = strings.TrimSuffix(postIDStr, "/like")
	}

	postID, err := strconv.Atoi(postIDStr)
	if err != nil {
		http.Error(w, "Invalid post ID", http.StatusBadRequest)
		return
	}

	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	db := database.GetDB()
	_, err = db.Exec(
		"INSERT OR IGNORE INTO likes (user_id, post_id) VALUES (?, ?)",
		userID, postID,
	)

	if err != nil {
		http.Error(w, "Failed to like post", http.StatusInternalServerError)
		return
	}

	// Update like count
	db.Exec("UPDATE posts SET like_count = like_count + 1 WHERE id = ?", postID)

	json.NewEncoder(w).Encode(map[string]string{
		"message": "Post liked successfully",
	})
}

func UnlikePost(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	postIDStr := r.URL.Path[len("/api/posts/"):]
	if strings.Contains(postIDStr, "/like") {
		postIDStr = strings.TrimSuffix(postIDStr, "/like")
	}

	postID, err := strconv.Atoi(postIDStr)
	if err != nil {
		http.Error(w, "Invalid post ID", http.StatusBadRequest)
		return
	}

	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	db := database.GetDB()
	_, err = db.Exec(
		"DELETE FROM likes WHERE user_id = ? AND post_id = ?",
		userID, postID,
	)

	if err != nil {
		http.Error(w, "Failed to unlike post", http.StatusInternalServerError)
		return
	}

	// Update like count
	db.Exec("UPDATE posts SET like_count = like_count - 1 WHERE id = ?", postID)

	json.NewEncoder(w).Encode(map[string]string{
		"message": "Post unliked successfully",
	})
}

func BookmarkPost(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	postIDStr := r.URL.Path[len("/api/posts/"):]
	if strings.Contains(postIDStr, "/bookmark") {
		postIDStr = strings.TrimSuffix(postIDStr, "/bookmark")
	}

	postID, err := strconv.Atoi(postIDStr)
	if err != nil {
		http.Error(w, "Invalid post ID", http.StatusBadRequest)
		return
	}

	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	db := database.GetDB()
	_, err = db.Exec(
		"INSERT OR IGNORE INTO bookmarks (user_id, post_id) VALUES (?, ?)",
		userID, postID,
	)

	if err != nil {
		http.Error(w, "Failed to bookmark post", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{
		"message": "Post bookmarked successfully",
	})
}

func UnbookmarkPost(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	postIDStr := r.URL.Path[len("/api/posts/"):]
	if strings.Contains(postIDStr, "/bookmark") {
		postIDStr = strings.TrimSuffix(postIDStr, "/bookmark")
	}

	postID, err := strconv.Atoi(postIDStr)
	if err != nil {
		http.Error(w, "Invalid post ID", http.StatusBadRequest)
		return
	}

	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	db := database.GetDB()
	_, err = db.Exec(
		"DELETE FROM bookmarks WHERE user_id = ? AND post_id = ?",
		userID, postID,
	)

	if err != nil {
		http.Error(w, "Failed to unbookmark post", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{
		"message": "Post unbookmarked successfully",
	})
}

func GetUserBookmarks(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	db := database.GetDB()
	rows, err := db.Query(`
		SELECT p.id, p.title, p.content, p.excerpt, p.author_id, u.username, p.category_id, c.name,
		       p.cover_image_url, p.reading_time, p.view_count, p.published, p.created_at, p.updated_at
		FROM posts p
		JOIN users u ON p.author_id = u.id
		LEFT JOIN categories c ON p.category_id = c.id
		JOIN bookmarks b ON p.id = b.post_id
		WHERE b.user_id = ?
		ORDER BY b.created_at DESC
	`, userID)

	if err != nil {
		http.Error(w, "Failed to fetch bookmarks", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var posts []models.Post
	for rows.Next() {
		var post models.Post
		var categoryName sql.NullString
		if err := rows.Scan(&post.ID, &post.Title, &post.Content, &post.Excerpt, &post.AuthorID, &post.Author,
			&post.CategoryID, &categoryName, &post.CoverImageURL, &post.ReadingTime, &post.ViewCount,
			&post.Published, &post.CreatedAt, &post.UpdatedAt); err != nil {
			continue
		}
		if categoryName.Valid {
			post.Category = categoryName.String
		}
		posts = append(posts, post)
	}

	json.NewEncoder(w).Encode(posts)
}

func SubscribeNewsletter(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Email string `json:"email"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	db := database.GetDB()
	_, err := db.Exec(
		"INSERT OR REPLACE INTO newsletter_subscribers (email, is_active, subscribed_at, unsubscribed_at) VALUES (?, 1, ?, NULL)",
		req.Email, time.Now(),
	)

	if err != nil {
		http.Error(w, "Failed to subscribe", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{
		"message": "Successfully subscribed to newsletter",
	})
}
