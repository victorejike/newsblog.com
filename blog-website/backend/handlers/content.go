package handlers

import (
	"blog-website/backend/database"
	"blog-website/backend/models"
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
)

func GetCategories(w http.ResponseWriter, r *http.Request) {
	db := database.GetDB()
	rows, err := db.Query("SELECT id, name, slug, description, created_at FROM categories ORDER BY name")
	if err != nil {
		http.Error(w, "Failed to fetch categories", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var categories []models.Category
	for rows.Next() {
		var cat models.Category
		var description sql.NullString
		if err := rows.Scan(&cat.ID, &cat.Name, &cat.Slug, &description, &cat.CreatedAt); err != nil {
			continue
		}
		if description.Valid {
			cat.Description = description.String
		}
		categories = append(categories, cat)
	}

	json.NewEncoder(w).Encode(categories)
}

func GetCategoryPosts(w http.ResponseWriter, r *http.Request) {
	categorySlug := r.URL.Path[len("/api/categories/"):]
	if strings.Contains(categorySlug, "/posts") {
		categorySlug = strings.TrimSuffix(categorySlug, "/posts")
	}

	db := database.GetDB()
	rows, err := db.Query(`
		SELECT p.id, p.title, p.content, p.excerpt, p.author_id, u.username, p.category_id, c.name,
		       p.cover_image_url, p.reading_time, p.view_count, p.published, p.created_at, p.updated_at
		FROM posts p
		JOIN users u ON p.author_id = u.id
		LEFT JOIN categories c ON p.category_id = c.id
		WHERE c.slug = ? AND p.published = 1
		ORDER BY p.created_at DESC
	`, categorySlug)

	if err != nil {
		http.Error(w, "Failed to fetch category posts", http.StatusInternalServerError)
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

func GetTags(w http.ResponseWriter, r *http.Request) {
	db := database.GetDB()
	rows, err := db.Query("SELECT id, name, slug, created_at FROM tags ORDER BY name")
	if err != nil {
		http.Error(w, "Failed to fetch tags", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var tags []models.Tag
	for rows.Next() {
		var tag models.Tag
		if err := rows.Scan(&tag.ID, &tag.Name, &tag.Slug, &tag.CreatedAt); err != nil {
			continue
		}
		tags = append(tags, tag)
	}

	json.NewEncoder(w).Encode(tags)
}

func GetTagPosts(w http.ResponseWriter, r *http.Request) {
	tagSlug := r.URL.Path[len("/api/tags/"):]
	if strings.Contains(tagSlug, "/posts") {
		tagSlug = strings.TrimSuffix(tagSlug, "/posts")
	}

	db := database.GetDB()
	rows, err := db.Query(`
		SELECT DISTINCT p.id, p.title, p.content, p.excerpt, p.author_id, u.username, p.category_id, c.name,
		       p.cover_image_url, p.reading_time, p.view_count, p.published, p.created_at, p.updated_at
		FROM posts p
		JOIN users u ON p.author_id = u.id
		LEFT JOIN categories c ON p.category_id = c.id
		JOIN post_tags pt ON p.id = pt.post_id
		JOIN tags t ON pt.tag_id = t.id
		WHERE t.slug = ? AND p.published = 1
		ORDER BY p.created_at DESC
	`, tagSlug)

	if err != nil {
		http.Error(w, "Failed to fetch tag posts", http.StatusInternalServerError)
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

func SearchPosts(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		http.Error(w, "Search query required", http.StatusBadRequest)
		return
	}

	searchTerm := "%" + query + "%"

	db := database.GetDB()
	rows, err := db.Query(`
		SELECT p.id, p.title, p.content, p.excerpt, p.author_id, u.username, p.category_id, c.name,
		       p.cover_image_url, p.reading_time, p.view_count, p.published, p.created_at, p.updated_at
		FROM posts p
		JOIN users u ON p.author_id = u.id
		LEFT JOIN categories c ON p.category_id = c.id
		WHERE p.published = 1 AND (p.title LIKE ? OR p.content LIKE ? OR p.excerpt LIKE ?)
		ORDER BY p.created_at DESC
	`, searchTerm, searchTerm, searchTerm)

	if err != nil {
		http.Error(w, "Failed to search posts", http.StatusInternalServerError)
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

func GetPopularPosts(w http.ResponseWriter, r *http.Request) {
	db := database.GetDB()
	rows, err := db.Query(`
		SELECT p.id, p.title, p.content, p.excerpt, p.author_id, u.username, p.category_id, c.name,
		       p.cover_image_url, p.reading_time, p.view_count, p.published, p.created_at, p.updated_at
		FROM posts p
		JOIN users u ON p.author_id = u.id
		LEFT JOIN categories c ON p.category_id = c.id
		WHERE p.published = 1
		ORDER BY p.view_count DESC
		LIMIT 10
	`)

	if err != nil {
		http.Error(w, "Failed to fetch popular posts", http.StatusInternalServerError)
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

func GetRelatedPosts(w http.ResponseWriter, r *http.Request) {
	postIDStr := r.URL.Path[len("/api/posts/"):]
	if strings.Contains(postIDStr, "/related") {
		postIDStr = strings.TrimSuffix(postIDStr, "/related")
	}

	postID, err := strconv.Atoi(postIDStr)
	if err != nil {
		http.Error(w, "Invalid post ID", http.StatusBadRequest)
		return
	}

	db := database.GetDB()
	var categoryID sql.NullInt64
	err = db.QueryRow("SELECT category_id FROM posts WHERE id = ?", postID).Scan(&categoryID)
	if err != nil {
		http.Error(w, "Failed to fetch post", http.StatusInternalServerError)
		return
	}

	var rows *sql.Rows
	if categoryID.Valid {
		rows, err = db.Query(`
			SELECT p.id, p.title, p.content, p.excerpt, p.author_id, u.username, p.category_id, c.name,
			       p.cover_image_url, p.reading_time, p.view_count, p.published, p.created_at, p.updated_at
			FROM posts p
			JOIN users u ON p.author_id = u.id
			LEFT JOIN categories c ON p.category_id = c.id
			WHERE p.published = 1 AND p.id != ? AND p.category_id = ?
			ORDER BY p.created_at DESC
			LIMIT 5
		`, postID, categoryID.Int64)
	} else {
		rows, err = db.Query(`
			SELECT p.id, p.title, p.content, p.excerpt, p.author_id, u.username, p.category_id, c.name,
			       p.cover_image_url, p.reading_time, p.view_count, p.published, p.created_at, p.updated_at
			FROM posts p
			JOIN users u ON p.author_id = u.id
			LEFT JOIN categories c ON p.category_id = c.id
			WHERE p.published = 1 AND p.id != ?
			ORDER BY p.created_at DESC
			LIMIT 5
		`, postID)
	}

	if err != nil {
		http.Error(w, "Failed to fetch related posts", http.StatusInternalServerError)
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
