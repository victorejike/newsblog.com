package handlers

import (
	"net/http"
)

func ServeAdminDashboard(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "backend/static/admin/dashboard.html")
}

func ServeEditPost(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "backend/static/admin/edit-post.html")
}
