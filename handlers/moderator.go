package handlers

import (
	"html/template"
	"net/http"

	"forum/models"
)

func ApplyModeratorHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		ErrorHandler(w, r, http.StatusMethodNotAllowed, http.StatusText(http.StatusMethodNotAllowed))
		return
	}

	cookie, err := r.Cookie("session_token")
	if err != nil || cookie.Value == "" {
		ErrorHandler(w, r, http.StatusUnauthorized, "Unauthorized")
		return
	}

	userID, _, err := models.GetIDBySessionToken(cookie.Value)
	if err != nil {
		ErrorHandler(w, r, http.StatusUnauthorized, "Unauthorized")
		return
	}

	role, err := models.GetUserRole(userID)
	if err != nil || role != "User" {
		ErrorHandler(w, r, http.StatusUnauthorized, "You must be a regular user to apply for moderator")
		return
	}

	// Check if the user has already applied
	alreadyApplied, err := models.HasAppliedForModerator(userID)
	if err != nil {
		ErrorHandler(w, r, http.StatusInternalServerError, "Failed to check application status")
		return
	}

	if alreadyApplied {
		err := models.CancelModerator(userID)
		if err != nil {
			ErrorHandler(w, r, http.StatusInternalServerError, "Failed to submit application")
			return
		}
	} else {
		err := models.ApplyForModerator(userID)
		if err != nil {
			ErrorHandler(w, r, http.StatusInternalServerError, "Failed to submit application")
			return
		}
	}

	http.Redirect(w, r, "/moderator", http.StatusSeeOther)
}

func RenderApplyModeratorPageHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		ErrorHandler(w, r, http.StatusMethodNotAllowed, http.StatusText(http.StatusMethodNotAllowed))
		return
	}

	loggedIn := false

	cookie, err := r.Cookie("session_token")
	if err != nil || cookie.Value == "" {
		ErrorHandler(w, r, http.StatusUnauthorized, "Unauthorized")
		return
	}

	userID, _, err := models.GetIDBySessionToken(cookie.Value)
	if err != nil {
		ErrorHandler(w, r, http.StatusUnauthorized, "Unauthorized")
		return
	} else {
		loggedIn = true
	}

	role, err := models.GetUserRole(userID)
	if err != nil || role != "User" {
		ErrorHandler(w, r, http.StatusUnauthorized, "You must be a regular user to apply for moderator")
		return
	}

	// Check if the user has already applied
	applied, err := models.HasAppliedForModerator(userID)
	if err != nil {
		ErrorHandler(w, r, http.StatusInternalServerError, "Failed to check application status")
		return
	}

	var applicationStatus string
	if applied {
		applicationStatus = "Pending"
	} else {
		applicationStatus = "Not Applied"
	}

	tmpl, err := template.ParseFiles("templates/moderator.html")
	if err != nil {
		ErrorHandler(w, r, http.StatusInternalServerError, "Failed to load template")
		return
	}

	err = tmpl.Execute(w, struct {
		ApplicationSubmitted bool
		ApplicationStatus    string
		LoggedIn             bool
	}{
		ApplicationSubmitted: applied,
		ApplicationStatus:    applicationStatus,
		LoggedIn:             loggedIn,
	})
	if err != nil {
		ErrorHandler(w, r, http.StatusInternalServerError, "Error rendering template")
	}
}

func ApproveModeratorApplicationHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		ErrorHandler(w, r, http.StatusMethodNotAllowed, http.StatusText(http.StatusMethodNotAllowed))
		return
	}

	userID := r.FormValue("user_id")
	if userID == "" {
		ErrorHandler(w, r, http.StatusBadRequest, "Invalid user ID")
		return
	}

	role, err := models.GetUserRole(userID)
	if err != nil || role == "Administrator" || role == "Moderator" {
		ErrorHandler(w, r, http.StatusUnauthorized, "Invalid user role")
		return
	}

	if err := models.PromoteUser(userID); err != nil {
		ErrorHandler(w, r, http.StatusInternalServerError, "Failed to approve application")
		return
	}

	http.Redirect(w, r, "/users", http.StatusSeeOther)
}

// RejectModeratorApplicationHandler handles rejecting a moderator application
func RejectModeratorApplicationHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		ErrorHandler(w, r, http.StatusMethodNotAllowed, http.StatusText(http.StatusMethodNotAllowed))
		return
	}

	userID := r.FormValue("user_id")
	if userID == "" {
		ErrorHandler(w, r, http.StatusBadRequest, "Invalid user ID")
		return
	}

	role, err := models.GetUserRole(userID)
	if err != nil || role == "Administrator" || role == "Moderator" {
		ErrorHandler(w, r, http.StatusUnauthorized, "Invalid user role")
		return
	}

	if err := models.CancelModerator(userID); err != nil {
		ErrorHandler(w, r, http.StatusInternalServerError, "Failed to reject application")
		return
	}

	http.Redirect(w, r, "/users", http.StatusSeeOther)
}

func RenderModeratorApplicationsPageHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		ErrorHandler(w, r, http.StatusMethodNotAllowed, http.StatusText(http.StatusMethodNotAllowed))
		return
	}

	// Fetch all moderator applications
	applications, err := models.GetAllModeratorApplications()
	if err != nil {
		ErrorHandler(w, r, http.StatusInternalServerError, "Failed to fetch moderator applications")
		return
	}

	tmpl, err := template.ParseFiles("templates/applications.html")
	if err != nil {
		ErrorHandler(w, r, http.StatusInternalServerError, "Failed to load template")
		return
	}

	err = tmpl.Execute(w, struct {
		Applications []models.ModeratorApplication
	}{
		Applications: applications,
	})
	if err != nil {
		ErrorHandler(w, r, http.StatusInternalServerError, "Error rendering template")
	}
}
