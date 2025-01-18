package handlers

import (
	"forum/models"
	"net/http"
	"strconv"
)

// AddCategoryHandler handles adding a new category
func AddCategoryHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
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

	name := r.FormValue("name")
	if name == "" || len(name) > 30 {
		ErrorHandler(w, r, http.StatusBadRequest, "Category name cannot be empty or longer than 30 symbols")
		return
	}

	// Add category with auto-incremented ID
	if err := models.AddCategory(name); err != nil {
		ErrorHandler(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	http.Redirect(w, r, "/admin/categories", http.StatusSeeOther)
}

// DeleteCategoryHandler handles deleting a category
func DeleteCategoryHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
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

	idStr := r.FormValue("id")
	if idStr == "" {
		ErrorHandler(w, r, http.StatusBadRequest, "Category ID cannot be empty")
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		ErrorHandler(w, r, http.StatusBadRequest, "Invalid category ID format")
		return
	}

	if err := models.DeleteCategory(id); err != nil {
		ErrorHandler(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	http.Redirect(w, r, "/admin/categories", http.StatusSeeOther)
}

// UpdateCategoryHandler handles updating an existing category
func UpdateCategoryHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
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

	idStr := r.FormValue("id")
	if idStr == "" {
		ErrorHandler(w, r, http.StatusBadRequest, "Category ID cannot be empty")
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		ErrorHandler(w, r, http.StatusBadRequest, "Invalid category ID format")
		return
	}

	newName := r.FormValue("name")
	if newName == "" {
		ErrorHandler(w, r, http.StatusBadRequest, "New category name cannot be empty")
		return
	}

	if err := models.UpdateCategory(id, newName); err != nil {
		ErrorHandler(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	http.Redirect(w, r, "/admin/categories", http.StatusSeeOther)
}

func SetControversialHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        ErrorHandler(w, r, http.StatusMethodNotAllowed, "Invalid request method")
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
	
    categoryID, err := strconv.Atoi(r.FormValue("category_id"))
    if err != nil {
        ErrorHandler(w, r, http.StatusBadRequest, "Invalid category ID")
        return
    }

    isControversial := r.FormValue("is_controversial") == "true"

    if err := models.SetCategoryControversialStatus(categoryID, isControversial); err != nil {
        ErrorHandler(w, r, http.StatusInternalServerError, "Failed to update category status")
        return
    }

    http.Redirect(w, r, "/admin/categories", http.StatusSeeOther)
}
