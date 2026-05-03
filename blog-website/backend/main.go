package main

import (
	"blog-website/backend/database"
	"blog-website/backend/handlers"
	"blog-website/backend/middleware"
	"log"
	"net/http"
	"strings"
)

func main() {
	if err := database.Init(); err != nil {
		log.Fatal("Failed to initialize database:", err)
	}
	defer database.Close()

	fs := http.FileServer(http.Dir("backend/static"))
	http.Handle("/css/", fs)
	http.Handle("/js/", fs)
	http.Handle("/admin/", fs)

	// Frontend routes
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path

		switch path {
		case "/":
			http.ServeFile(w, r, "backend/static/index.html")
		case "/login":
			http.ServeFile(w, r, "backend/static/login.html")
		case "/register":
			http.ServeFile(w, r, "backend/static/register.html")
		case "/profile":
			http.ServeFile(w, r, "backend/static/profile.html")
		case "/bookmarks":
			http.ServeFile(w, r, "backend/static/bookmarks.html")
		default:
			// Check if it's a post page
			if strings.HasPrefix(path, "/post/") {
				http.ServeFile(w, r, "backend/static/post.html")
			} else {
				fs.ServeHTTP(w, r)
			}
		}
	})

	// Auth routes
	http.HandleFunc("/api/register", handlers.Register)
	http.HandleFunc("/api/login", handlers.Login)
	http.HandleFunc("/api/logout", handlers.Logout)

	// Posts routes
	http.HandleFunc("/api/posts", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handlers.GetPosts(w, r)
		case http.MethodPost:
			middleware.AuthMiddleware(handlers.CreatePost)(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	http.HandleFunc("/api/posts/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path

		// Check for special endpoints
		if strings.HasSuffix(path, "/related") {
			handlers.GetRelatedPosts(w, r)
			return
		}
		if strings.HasSuffix(path, "/comments") {
			switch r.Method {
			case http.MethodGet:
				handlers.GetPostComments(w, r)
			case http.MethodPost:
				middleware.AuthMiddleware(handlers.CreateComment)(w, r)
			default:
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			}
			return
		}
		if strings.HasSuffix(path, "/like") {
			switch r.Method {
			case http.MethodPost:
				middleware.AuthMiddleware(handlers.LikePost)(w, r)
			case http.MethodDelete:
				middleware.AuthMiddleware(handlers.UnlikePost)(w, r)
			default:
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			}
			return
		}
		if strings.HasSuffix(path, "/bookmark") {
			switch r.Method {
			case http.MethodPost:
				middleware.AuthMiddleware(handlers.BookmarkPost)(w, r)
			case http.MethodDelete:
				middleware.AuthMiddleware(handlers.UnbookmarkPost)(w, r)
			default:
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			}
			return
		}

		// Regular post operations
		switch r.Method {
		case http.MethodGet:
			handlers.GetPost(w, r)
		case http.MethodPut:
			middleware.AuthMiddleware(handlers.UpdatePost)(w, r)
		case http.MethodDelete:
			middleware.AuthMiddleware(handlers.DeletePost)(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Content discovery routes
	http.HandleFunc("/api/categories", handlers.GetCategories)
	http.HandleFunc("/api/categories/", handlers.GetCategoryPosts)
	http.HandleFunc("/api/tags", handlers.GetTags)
	http.HandleFunc("/api/tags/", handlers.GetTagPosts)
	http.HandleFunc("/api/search", handlers.SearchPosts)
	http.HandleFunc("/api/popular", handlers.GetPopularPosts)

	// Comments routes
	http.HandleFunc("/api/comments/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodDelete {
			middleware.AuthMiddleware(handlers.DeleteComment)(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// User routes
	http.HandleFunc("/api/bookmarks", middleware.AuthMiddleware(handlers.GetUserBookmarks))
	http.HandleFunc("/api/users/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path

		if strings.HasSuffix(path, "/profile") {
			switch r.Method {
			case http.MethodGet:
				handlers.GetUserProfile(w, r)
			case http.MethodPut:
				middleware.AuthMiddleware(handlers.UpdateUserProfile)(w, r)
			default:
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			}
			return
		}
		if strings.HasSuffix(path, "/posts") {
			handlers.GetUserPosts(w, r)
			return
		}

		http.Error(w, "Not found", http.StatusNotFound)
	})

	http.HandleFunc("/api/me", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			middleware.AuthMiddleware(handlers.GetCurrentUserProfile)(w, r)
		case http.MethodPut:
			middleware.AuthMiddleware(handlers.UpdateUserProfile)(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Newsletter route
	http.HandleFunc("/api/newsletter", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			handlers.SubscribeNewsletter(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Admin routes
	http.HandleFunc("/api/admin/posts", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			middleware.AdminMiddleware(handlers.GetAllPosts)(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	http.HandleFunc("/admin", func(w http.ResponseWriter, r *http.Request) {
		middleware.AuthMiddleware(handlers.ServeAdminDashboard)(w, r)
	})

	http.HandleFunc("/admin/edit", func(w http.ResponseWriter, r *http.Request) {
		middleware.AuthMiddleware(handlers.ServeEditPost)(w, r)
	})

	log.Println("Server starting on http://localhost:8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal("Server failed to start:", err)
	}
}
