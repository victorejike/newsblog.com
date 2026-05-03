package handlers

import (
	"blog-website/backend/database"
	"blog-website/backend/models"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

func GetUserProfile(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.URL.Path[len("/api/users/"):]
	if strings.Contains(userIDStr, "/profile") {
		userIDStr = strings.TrimSuffix(userIDStr, "/profile")
	}

	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	db := database.GetDB()
	var user models.User
	var bio, avatarURL, websiteURL, twitterHandle sql.NullString
	err = db.QueryRow(`
		SELECT id, username, email, bio, avatar_url, website_url, twitter_handle, is_admin, created_at
		FROM users WHERE id = ?
	`, userID).Scan(&user.ID, &user.Username, &user.Email, &bio, &avatarURL,
		&websiteURL, &twitterHandle, &user.IsAdmin, &user.CreatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "User not found", http.StatusNotFound)
		} else {
			http.Error(w, "Database error", http.StatusInternalServerError)
		}
		return
	}

	if bio.Valid {
		user.Bio = bio.String
	}
	if avatarURL.Valid {
		user.AvatarURL = avatarURL.String
	}
	if websiteURL.Valid {
		user.WebsiteURL = websiteURL.String
	}
	if twitterHandle.Valid {
		user.TwitterHandle = twitterHandle.String
	}

	json.NewEncoder(w).Encode(user)
}

func UpdateUserProfile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userIDStr := r.Header.Get("X-User-ID")
	if userIDStr == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	var req models.UpdateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	db := database.GetDB()
	_, err = db.Exec(`
		UPDATE users SET bio = ?, avatar_url = ?, website_url = ?, twitter_handle = ?
		WHERE id = ?
	`, req.Bio, req.AvatarURL, req.WebsiteURL, req.TwitterHandle, userID)

	if err != nil {
		http.Error(w, "Failed to update profile", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{
		"message": "Profile updated successfully",
	})
}

func GetUserPosts(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.URL.Path[len("/api/users/"):]
	if strings.Contains(userIDStr, "/posts") {
		userIDStr = strings.TrimSuffix(userIDStr, "/posts")
	}

	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	db := database.GetDB()
	rows, err := db.Query(`
		SELECT p.id, p.title, p.content, p.excerpt, p.author_id, u.username, p.category_id, c.name,
		       p.cover_image_url, p.reading_time, p.view_count, p.published, p.created_at, p.updated_at
		FROM posts p
		JOIN users u ON p.author_id = u.id
		LEFT JOIN categories c ON p.category_id = c.id
		WHERE p.author_id = ? AND p.published = 1
		ORDER BY p.created_at DESC
	`, userID)

	if err != nil {
		http.Error(w, "Failed to fetch user posts", http.StatusInternalServerError)
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

func GetCurrentUserProfile(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.Header.Get("X-User-ID")
	if userIDStr == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	db := database.GetDB()
	var user models.User
	var bio, avatarURL, websiteURL, twitterHandle sql.NullString
	err = db.QueryRow(`
		SELECT id, username, email, bio, avatar_url, website_url, twitter_handle, is_admin, created_at
		FROM users WHERE id = ?
	`, userID).Scan(&user.ID, &user.Username, &user.Email, &bio, &avatarURL,
		&websiteURL, &twitterHandle, &user.IsAdmin, &user.CreatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "User not found", http.StatusNotFound)
		} else {
			fmt.Printf("Database error: %v\n", err)
			http.Error(w, "Database error", http.StatusInternalServerError)
		}
		return
	}

	if bio.Valid {
		user.Bio = bio.String
	}
	if avatarURL.Valid {
		user.AvatarURL = avatarURL.String
	}
	if websiteURL.Valid {
		user.WebsiteURL = websiteURL.String
	}
	if twitterHandle.Valid {
		user.TwitterHandle = twitterHandle.String
	}

	json.NewEncoder(w).Encode(user)
}
