package handlers

import (
	"html/template"
	"log"
	"net/http"
	"strconv"

	"forum/models"
)

// MainPageHandler - Displays the main page with posts and user information if logged in
func MainPageHandler(w http.ResponseWriter, r *http.Request) {
	var username string
	loggedIn := false
	var userID string

	// Check if the user is logged in
	cookie, err := r.Cookie("session_token")
	if err == nil {
		sessionToken := cookie.Value

		// Get the username of the logged-in user
		userID, username, err = models.GetIDBySessionToken(sessionToken)
		if err == nil {
			loggedIn = true // User is logged in
		}
	}

	role, _ := models.GetUserRole(userID)
	isAdminOrModerator := role == "Administrator" || role == "Moderator"
	isAdmin := role == "Administrator"

	// Get filters from query parameters
	categoryID, err := strconv.Atoi(r.URL.Query().Get("category"))
	if err != nil {
		categoryID = 0
	}

	// Retrieve all posts
	posts, err := models.GetFilteredPosts(loggedIn, userID, categoryID)
	if err != nil {
		ErrorHandler(w, r, http.StatusInternalServerError, "Error fetching posts")
		return
	}

	// Retrieve all categories
	categories, err := models.GetAllCategories()
	if err != nil {
		ErrorHandler(w, r, http.StatusInternalServerError, "Error fetching categories")
		return
	}

	pending_posts, err := models.GetPendingPosts()
	if err != nil {
		ErrorHandler(w, r, http.StatusInternalServerError, "Error fetching pending posts")
		return
	}

	applications, err := models.GetAllModeratorApplications()
	if err != nil {
		ErrorHandler(w, r, http.StatusInternalServerError, "Failed to fetch moderator applications")
		return
	}

	notificationCount, err := models.GetNotificationCount(userID)
	if err != nil {
		ErrorHandler(w, r, http.StatusInternalServerError, "Failed to fetch notifications count")
		return
	}

	tmpl, err := template.ParseFiles("templates/index.html")
	if err != nil {
		ErrorHandler(w, r, http.StatusInternalServerError, "Error loading template")
		return
	}
	if r.URL.Path != "/" {
		ErrorHandler(w, r, http.StatusNotFound, http.StatusText(http.StatusNotFound))
		return
	}

	data := struct {
		Posts      []models.Post
		Categories []models.Category
		LoggedIn   bool
		Username   string
		SelectedCategory   int
		IsAdminOrModerator bool
		IsAdmin            bool
		ReportCount        int
		ApplicationCount   int
		NotificationCount int
	}{
		Posts:      posts,
		Categories: categories,
		LoggedIn:   loggedIn,
		Username:   username,
		SelectedCategory:   categoryID,
		IsAdminOrModerator: isAdminOrModerator,
		IsAdmin:            isAdmin,
		ReportCount:        len(pending_posts),
		ApplicationCount:   len(applications),
		NotificationCount: notificationCount,
	}

	err = tmpl.Execute(w, data)
	if err != nil {
		log.Println("Error executing template:", err)
	}
}
