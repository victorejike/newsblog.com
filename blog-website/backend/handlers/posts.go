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

func GetPosts(w http.ResponseWriter, r *http.Request) {
	db := database.GetDB()
	rows, err := db.Query(`
		SELECT p.id, p.title, p.content, p.excerpt, p.author_id, u.username, p.category_id, c.name,
		       p.cover_image_url, p.reading_time, p.view_count, p.published, p.created_at, p.updated_at
		FROM posts p
		JOIN users u ON p.author_id = u.id
		LEFT JOIN categories c ON p.category_id = c.id
		WHERE p.published = 1
		ORDER BY p.created_at DESC
	`)

	if err != nil {
		http.Error(w, "Failed to fetch posts", http.StatusInternalServerError)
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
		// Get tags for this post
		post.Tags = getPostTags(db, post.ID)
		posts = append(posts, post)
	}

	json.NewEncoder(w).Encode(posts)
}

func getPostTags(db *sql.DB, postID int) []string {
	rows, err := db.Query(`
		SELECT t.name FROM tags t
		JOIN post_tags pt ON t.id = pt.tag_id
		WHERE pt.post_id = ?
	`, postID)
	if err != nil {
		return []string{}
	}
	defer rows.Close()

	var tags []string
	for rows.Next() {
		var tagName string
		if err := rows.Scan(&tagName); err != nil {
			continue
		}
		tags = append(tags, tagName)
	}
	return tags
}

func GetPost(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Path[len("/api/posts/"):]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid post ID", http.StatusBadRequest)
		return
	}

	db := database.GetDB()
	var post models.Post
	var categoryName sql.NullString
	err = db.QueryRow(`
		SELECT p.id, p.title, p.content, p.excerpt, p.author_id, u.username, p.category_id, c.name,
		       p.cover_image_url, p.reading_time, p.view_count, p.published, p.created_at, p.updated_at
		FROM posts p
		JOIN users u ON p.author_id = u.id
		LEFT JOIN categories c ON p.category_id = c.id
		WHERE p.id = ? AND p.published = 1
	`, id).Scan(&post.ID, &post.Title, &post.Content, &post.Excerpt, &post.AuthorID, &post.Author,
		&post.CategoryID, &categoryName, &post.CoverImageURL, &post.ReadingTime, &post.ViewCount,
		&post.Published, &post.CreatedAt, &post.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Post not found", http.StatusNotFound)
		} else {
			http.Error(w, "Database error", http.StatusInternalServerError)
		}
		return
	}

	if categoryName.Valid {
		post.Category = categoryName.String
	}

	// Get tags for this post
	post.Tags = getPostTags(db, post.ID)

	// Increment view count
	db.Exec("UPDATE posts SET view_count = view_count + 1 WHERE id = ?", id)

	json.NewEncoder(w).Encode(post)
}

func CreatePost(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req models.CreatePostRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Calculate reading time (average 200 words per minute)
	readingTime := calculateReadingTime(req.Content)

	db := database.GetDB()
	result, err := db.Exec(
		"INSERT INTO posts (title, content, excerpt, author_id, category_id, cover_image_url, reading_time, published) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
		req.Title, req.Content, req.Excerpt, userID, req.CategoryID, req.CoverImageURL, readingTime, req.Published,
	)

	if err != nil {
		http.Error(w, "Failed to create post", http.StatusInternalServerError)
		return
	}

	id, _ := result.LastInsertId()

	// Add tags
	if len(req.Tags) > 0 {
		for _, tagName := range req.Tags {
			tagID := getOrCreateTag(db, tagName)
			db.Exec("INSERT OR IGNORE INTO post_tags (post_id, tag_id) VALUES (?, ?)", id, tagID)
		}
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Post created successfully",
		"id":      id,
	})
}

func calculateReadingTime(content string) int {
	words := strings.Fields(content)
	return len(words) / 200
}

func getOrCreateTag(db *sql.DB, tagName string) int {
	slug := strings.ToLower(strings.ReplaceAll(tagName, " ", "-"))
	var tagID int
	err := db.QueryRow("SELECT id FROM tags WHERE name = ?", tagName).Scan(&tagID)
	if err == sql.ErrNoRows {
		result, _ := db.Exec("INSERT INTO tags (name, slug) VALUES (?, ?)", tagName, slug)
		id, _ := result.LastInsertId()
	tagID = int(id)
	}
	return tagID
}

func UpdatePost(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	idStr := r.URL.Path[len("/api/posts/"):]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid post ID", http.StatusBadRequest)
		return
	}

	var req models.UpdatePostRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Calculate reading time
	readingTime := calculateReadingTime(req.Content)

	db := database.GetDB()
	_, err = db.Exec(
		"UPDATE posts SET title = ?, content = ?, excerpt = ?, category_id = ?, cover_image_url = ?, reading_time = ?, published = ?, updated_at = ? WHERE id = ?",
		req.Title, req.Content, req.Excerpt, req.CategoryID, req.CoverImageURL, readingTime, req.Published, time.Now(), id,
	)

	if err != nil {
		http.Error(w, "Failed to update post", http.StatusInternalServerError)
		return
	}

	// Update tags
	db.Exec("DELETE FROM post_tags WHERE post_id = ?", id)
	if len(req.Tags) > 0 {
		for _, tagName := range req.Tags {
			tagID := getOrCreateTag(db, tagName)
			db.Exec("INSERT OR IGNORE INTO post_tags (post_id, tag_id) VALUES (?, ?)", id, tagID)
		}
	}

	json.NewEncoder(w).Encode(map[string]string{
		"message": "Post updated successfully",
	})
}

func DeletePost(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	idStr := r.URL.Path[len("/api/posts/"):]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid post ID", http.StatusBadRequest)
		return
	}

	db := database.GetDB()
	_, err = db.Exec("DELETE FROM posts WHERE id = ?", id)

	if err != nil {
		http.Error(w, "Failed to delete post", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{
		"message": "Post deleted successfully",
	})
}

func GetAllPosts(w http.ResponseWriter, r *http.Request) {
	db := database.GetDB()
	rows, err := db.Query(`
		SELECT p.id, p.title, p.content, p.excerpt, p.author_id, u.username, p.category_id, c.name,
		       p.cover_image_url, p.reading_time, p.view_count, p.published, p.created_at, p.updated_at
		FROM posts p
		JOIN users u ON p.author_id = u.id
		LEFT JOIN categories c ON p.category_id = c.id
		ORDER BY p.created_at DESC
	`)

	if err != nil {
		http.Error(w, "Failed to fetch posts", http.StatusInternalServerError)
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
		post.Tags = getPostTags(db, post.ID)
		posts = append(posts, post)
	}

	json.NewEncoder(w).Encode(posts)
}
