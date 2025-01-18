package handlers

import (
	"html/template"
	"net/http"

	"forum/models"
)

func PromoteToModeratorHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	cookie, err := r.Cookie("session_token")
	if err != nil || cookie.Value == "" {
		ErrorHandler(w, r, http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized))
		return
	}

	adminID, _, err := models.GetIDBySessionToken(cookie.Value)
	if err != nil {
		ErrorHandler(w, r, http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized))
		return
	}

	role, err := models.GetUserRole(adminID)
	if err != nil || role != "Administrator" {
		http.Error(w, "Unauthorized", http.StatusForbidden)
		return
	}

	userID := r.FormValue("user_id")
	if err := models.PromoteUser(userID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/users", http.StatusSeeOther)
}

func DemoteToUserHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	cookie, err := r.Cookie("session_token")
	if err != nil || cookie.Value == "" {
		ErrorHandler(w, r, http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized))
		return
	}

	adminID, _, err := models.GetIDBySessionToken(cookie.Value)
	if err != nil {
		ErrorHandler(w, r, http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized))
		return
	}

	role, err := models.GetUserRole(adminID)
	if err != nil || role != "Administrator" {
		http.Error(w, "Unauthorized", http.StatusForbidden)
		return
	}

	userID := r.FormValue("user_id")
	if err := models.DemoteUser(userID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/users", http.StatusSeeOther)
}

func UsersPageHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		ErrorHandler(w, r, http.StatusMethodNotAllowed, http.StatusText(http.StatusMethodNotAllowed))
		return
	}

	cookie, err := r.Cookie("session_token")
	if err != nil || cookie.Value == "" {
		ErrorHandler(w, r, http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized))
		return
	}

	adminID, _, err := models.GetIDBySessionToken(cookie.Value)
	if err != nil {
		ErrorHandler(w, r, http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized))
		return
	}

	role, err := models.GetUserRole(adminID)
	if err != nil || role != "Administrator" {
		ErrorHandler(w, r, http.StatusUnauthorized, "Not enough privilege")
		return
	}

	moderators, err := models.GetUsersByRole("Moderator")
	if err != nil {
		http.Error(w, "Failed to fetch moderators", http.StatusInternalServerError)
		return
	}

	users, err := models.GetUsersByRole("User")
	if err != nil {
		http.Error(w, "Failed to fetch users", http.StatusInternalServerError)
		return
	}

	applications, err := models.GetAllModeratorApplications()
	if err != nil {
		ErrorHandler(w, r, http.StatusInternalServerError, "Failed to fetch moderator applications")
		return
	}

	tmpl, err := template.ParseFiles("templates/users.html")
	if err != nil {
		ErrorHandler(w, r, http.StatusInternalServerError, "Error loading template")
		return
	}

	data := struct {
		Moderators   []models.User
		Users        []models.User
		Applications []models.ModeratorApplication
	}{
		Moderators:   moderators,
		Users:        users,
		Applications: applications,
	}

	tmpl.Execute(w, data)
}

func AdminCategoriesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		ErrorHandler(w, r, http.StatusMethodNotAllowed, http.StatusText(http.StatusMethodNotAllowed))
		return
	}

	cookie, err := r.Cookie("session_token")
	if err != nil || cookie.Value == "" {
		ErrorHandler(w, r, http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized))
		return
	}

	adminID, _, err := models.GetIDBySessionToken(cookie.Value)
	if err != nil {
		ErrorHandler(w, r, http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized))
		return
	}

	role, err := models.GetUserRole(adminID)
	if err != nil || role != "Administrator" {
		ErrorHandler(w, r, http.StatusUnauthorized, "Not enough privilege")
		return
	}

	categories, err := models.GetAllCategories()
	if err != nil {
		http.Error(w, "Unable to fetch categories", http.StatusInternalServerError)
		return
	}

	tmpl, err := template.ParseFiles("templates/categories.html")
	if err != nil {
		http.Error(w, "Template not found", http.StatusInternalServerError)
		return
	}

	tmpl.Execute(w, struct {
		Categories []models.Category
	}{
		Categories: categories,
	})
}
