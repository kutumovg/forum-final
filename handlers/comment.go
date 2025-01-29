package handlers

// comments page
import (
	"forum/models"
	"html/template"
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

	authorID, err := models.GetAuthorIDByPost(postID)
	if err != nil {
		ErrorHandler(w, r, http.StatusInternalServerError, "Error creating comment")
		return
	}

	err = models.AddNotification(userID, authorID, &postID, nil, "commented")
	if err != nil {
		ErrorHandler(w, r, http.StatusInternalServerError, "Error creating notification")

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

	authorID, err := models.GetAuthorIDByComment(commentID)
	if err != nil {
		ErrorHandler(w, r, http.StatusInternalServerError, "Error creating comment")
		return
	}

	err = models.AddNotification(userID, authorID, nil, &commentID, "liked")
	if err != nil {
		ErrorHandler(w, r, http.StatusInternalServerError, "Error creating notification")
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

	authorID, err := models.GetAuthorIDByComment(commentID)
	if err != nil {
		ErrorHandler(w, r, http.StatusInternalServerError, "Error creating comment")
		return
	}

	err = models.AddNotification(userID, authorID, nil, &commentID, "disliked")
	if err != nil {
		ErrorHandler(w, r, http.StatusInternalServerError, "Error creating notification")
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

	// adminID, _, err := models.GetIDBySessionToken(cookie.Value)
	// if err != nil {
	// 	ErrorHandler(w, r, http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized))
	// 	return
	// }

	// role, err := models.GetUserRole(adminID)
	// if err != nil || (role != "Administrator" && role != "Moderator") {
	// 	ErrorHandler(w, r, http.StatusUnauthorized, "Not enough privilege")
	// 	return
	// }

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

func MyCommentsPageHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		ErrorHandler(w, r, http.StatusMethodNotAllowed, http.StatusText(http.StatusMethodNotAllowed))
		return
	}

	var loggedIn bool
	var username string
	var userID string
	cookie, err := r.Cookie("session_token")
	if err != nil || cookie.Value == "" {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	userID, username, err = models.GetIDBySessionToken(cookie.Value)
	if err == nil {
		loggedIn = true
	}

	// Fetch comments for the post
	groupedComments, err := models.GetCommentsByUser(userID)
	if err != nil {
		ErrorHandler(w, r, http.StatusInternalServerError, "Error fetching comments")
		return
	}

	// Load the comments.html template
	tmpl, err := template.ParseFiles("templates/my_comments.html")
	if err != nil {
		http.Error(w, "Error loading template", http.StatusInternalServerError)
		return
	}

	data := struct {
		Posts    []models.Post
		LoggedIn bool
		Username string
	}{
		Posts:    groupedComments,
		LoggedIn: loggedIn,
		Username: username,
	}

	tmpl.Execute(w, data)
}

func EditCommentHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		commentID := r.URL.Query().Get("id")
		if commentID == "" {
			ErrorHandler(w, r, http.StatusBadRequest, "Missing comment ID")
			return
		}

		loggedIn := false
		var userID string

		cookie, err := r.Cookie("session_token")
		if err == nil {
			sessionToken := cookie.Value

			// Get the username of the logged-in user
			userID, _, err = models.GetIDBySessionToken(sessionToken)
			if err == nil {
				loggedIn = true // User is logged in
			} else {
				http.Redirect(w, r, "/login", http.StatusSeeOther)
				return
			}
		}

		comment, err := models.GetCommentByID(commentID)
		if err != nil {
			ErrorHandler(w, r, http.StatusNotFound, "Comment not found")
			return
		}

		if comment.Author != userID {
			ErrorHandler(w, r, http.StatusForbidden, "You are not authorized to edit this comment")
			return
		}

		tmpl, err := template.ParseFiles("templates/edit_comment.html")
		if err != nil {
			http.Error(w, "Error loading template", http.StatusInternalServerError)
			return
		}

		err = tmpl.Execute(w, struct {
			Comment  models.Comment
			LoggedIn bool
		}{
			Comment:  comment,
			LoggedIn: loggedIn,
		})
		if err != nil {
			ErrorHandler(w, r, http.StatusInternalServerError, "Error rendering template")
		}
	} else if r.Method == http.MethodPost {
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

		commentID := r.FormValue("comment_id")
		content := r.FormValue("content")

		if commentID == "" || content == "" {
			ErrorHandler(w, r, http.StatusBadRequest, "Missing comment ID or content")
			return
		}

		err = models.UpdateComment(commentID, userID, content)
		if err != nil {
			ErrorHandler(w, r, http.StatusInternalServerError, "Failed to update comment")
			return
		}

		postID, _ := models.GetPostIDByCommentID(commentID)
		http.Redirect(w, r, "/post?id="+postID, http.StatusSeeOther)
	} else {
		ErrorHandler(w, r, http.StatusMethodNotAllowed, http.StatusText(http.StatusMethodNotAllowed))
	}
}
