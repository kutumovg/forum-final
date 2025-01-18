package handlers

// comments page
import (
	"forum/models"
	"net/http"
)

func CreateCommentHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		ErrorHandler(w, r, http.StatusMethodNotAllowed, http.StatusText(http.StatusMethodNotAllowed))
		return
	}

	// Check if the user is logged in
	cookie, err := r.Cookie("session_token")
	if err != nil || cookie.Value == "" {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	userID, _, err := models.GetIDBySessionToken(cookie.Value)
	if err != nil {
		ErrorHandler(w, r, http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized))
		return
	}
	postID := r.FormValue("post_id")
	content := r.FormValue("content")

	content = models.SanitizeInput(content)
	if !models.IsValidContent(content) {
		ErrorHandler(w, r, http.StatusBadRequest, "Content is required to create a comment")
		return
	}

	err = models.CreateComment(postID, userID, content)
	if err != nil {
		ErrorHandler(w, r, http.StatusInternalServerError, "Error creating comment")
		return
	}

	http.Redirect(w, r, "/post?id="+postID, http.StatusSeeOther)
}

func LikeCommentHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		ErrorHandler(w, r, http.StatusMethodNotAllowed, http.StatusText(http.StatusMethodNotAllowed))
		return
	}

	// Check if the user is logged in
	cookie, err := r.Cookie("session_token")
	if err != nil || cookie.Value == "" {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	userID, _, err := models.GetIDBySessionToken(cookie.Value)
	if err != nil {
		ErrorHandler(w, r, http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized))
		return
	}
	commentID := r.FormValue("comment_id")
	postID := r.FormValue("post_id")

	err = models.LikeComment(userID, commentID)
	if err != nil {
		// if err.Error() == "you have already liked this comment" {
		// 	http.Redirect(w, r, "/post?id="+postID+"&notification=already_liked", http.StatusSeeOther)
		// 	return
		// }

		http.Error(w, "Error liking comment", http.StatusInternalServerError)
		return
	}

	// Update the post's like and dislike counts
	err = models.UpdateCommentLikesDislikes(commentID)
	if err != nil {
		http.Error(w, "Error updating like count: "+err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/post?id="+postID, http.StatusSeeOther)
}

func DislikeCommentHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		ErrorHandler(w, r, http.StatusMethodNotAllowed, http.StatusText(http.StatusMethodNotAllowed))
		return
	}

	// Check if the user is logged in
	cookie, err := r.Cookie("session_token")
	if err != nil || cookie.Value == "" {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	userID, _, err := models.GetIDBySessionToken(cookie.Value)
	if err != nil {
		ErrorHandler(w, r, http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized))
		return
	}
	commentID := r.FormValue("comment_id")
	postID := r.FormValue("post_id")

	err = models.DislikeComment(userID, commentID)
	if err != nil {
		// if err.Error() == "you have already disliked this comment" {
		// 	http.Redirect(w, r, "/post?id="+postID+"&notification=already_disliked", http.StatusSeeOther)
		// 	return
		// }

		http.Error(w, "Error disliking comment", http.StatusInternalServerError)
		return
	}

	// Update the post's like and dislike counts
	err = models.UpdateCommentLikesDislikes(commentID)
	if err != nil {
		ErrorHandler(w, r, http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
		return
	}

	http.Redirect(w, r, "/post?id="+postID, http.StatusSeeOther)
}

func DeleteCommentHandler(w http.ResponseWriter, r *http.Request) {
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
	if err != nil || (role != "Administrator" && role != "Moderator") {
		ErrorHandler(w, r, http.StatusUnauthorized, "Not enough privilege")
		return
	}

	commentID := r.FormValue("comment_id")
	if commentID == "" {
		ErrorHandler(w, r, http.StatusBadRequest, "Missing comment ID")
		return
	}

	if err := models.DeleteComment(commentID); err != nil {
		ErrorHandler(w, r, http.StatusInternalServerError, "Failed to delete comment")
		return
	}

	http.Redirect(w, r, r.Header.Get("Referer"), http.StatusSeeOther)
}