package handlers

import (
	"database/sql"
	"html/template"
	"net/http"

	"forum/models"
)

// Handler for creating a post
func CreatePostHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		ErrorHandler(w, r, http.StatusMethodNotAllowed, http.StatusText(http.StatusMethodNotAllowed))
		return
	}

	cookie, err := r.Cookie("session_token")
	if err != nil || cookie.Value == "" {
		ErrorHandler(w, r, http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized))
		return
	}

	userID, _, err := models.GetIDBySessionToken(cookie.Value)
	if err != nil {
		ErrorHandler(w, r, http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized))
		return
	}

	content := r.FormValue("content")
	categories := r.Form["categories"]

	content = models.SanitizeInput(content)
	if !models.IsValidContent(content) || len(categories) == 0 {
		ErrorHandler(w, r, http.StatusBadRequest, "Content and at least one category are required to create a post")
		return
	}

	var imagePath string
	if file, header, err := r.FormFile("image"); err == nil {
		defer file.Close()

		// Validate the image
		if err := validateImage(file, header); err != nil {
			ErrorHandler(w, r, http.StatusBadRequest, err.Error())
			return
		}

		// Save the image
		imagePath, err = saveImage(file, header)
		if err != nil {
			ErrorHandler(w, r, http.StatusInternalServerError, err.Error())
			return
		}
	}

	requiresModeration := false
	for _, categoryID := range categories {
		isControversial, err := models.IsCategoryControversial(categoryID)
		if err != nil {
			ErrorHandler(w, r, http.StatusInternalServerError, "Error checking category")
			return
		}
		if isControversial {
			requiresModeration = true
			break
		}
	}

	postID, err := models.CreatePost(userID, content, imagePath, !requiresModeration)
	if err != nil {
		ErrorHandler(w, r, http.StatusInternalServerError, "Error creating post")
		return
	}

	for _, categoryID := range categories {
		err = models.AddCategoryToPost(postID, categoryID)
		if err != nil {
			ErrorHandler(w, r, http.StatusInternalServerError, "Error associating category")
			return
		}
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// Handler for liking a post
func LikeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		ErrorHandler(w, r, http.StatusMethodNotAllowed, http.StatusText(http.StatusMethodNotAllowed))
		return
	}

	// Check if the user is logged in
	cookie, err := r.Cookie("session_token")
	if err != nil || cookie.Value == "" {
		ErrorHandler(w, r, http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized))
		return
	}

	userID, _, err := models.GetIDBySessionToken(cookie.Value)
	if err != nil {
		ErrorHandler(w, r, http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized))
		return
	}
	postID := r.FormValue("post_id")

	// Like the post
	err = models.LikePost(userID, postID)
	if err != nil {
		// if err.Error() == "you have already liked this post" {
		// 	// Redirect back to the main page with a notification
		// 	http.Redirect(w, r, "/?notification=already_liked", http.StatusSeeOther)
		// 	return
		// }

		http.Error(w, "Error liking post: "+err.Error(), http.StatusInternalServerError)
		ErrorHandler(w, r, http.StatusInternalServerError, "Error liking post")
		return
	}

	// Update the post's like and dislike counts
	err = models.UpdatePostLikesDislikes(postID)
	if err != nil {
		ErrorHandler(w, r, http.StatusInternalServerError, "Error updating like count")
		return
	}

	// Redirect back to the main page
	// http.Redirect(w, r, "/", http.StatusSeeOther)
	referer := r.Header.Get("Referer")
	http.Redirect(w, r, referer, http.StatusSeeOther)
	// if strings.Contains(referer, "/post") {
	// 	// If the referer is a post page, redirect back to that page
	// 	http.Redirect(w, r, referer, http.StatusSeeOther)
	// } else {
	// 	// Otherwise, redirect to the homepage
	// 	http.Redirect(w, r, "/", http.StatusSeeOther)
	// }
}

// Handler for disliking a post
func DislikeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		ErrorHandler(w, r, http.StatusMethodNotAllowed, http.StatusText(http.StatusMethodNotAllowed))
		return
	}

	// Check if the user is logged in
	cookie, err := r.Cookie("session_token")
	if err != nil || cookie.Value == "" {
		ErrorHandler(w, r, http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized))
		return
	}

	userID, _, err := models.GetIDBySessionToken(cookie.Value)
	if err != nil {
		ErrorHandler(w, r, http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized))
		return
	}
	postID := r.FormValue("post_id")

	// Dislike the post
	err = models.DislikePost(userID, postID)
	if err != nil {
		// if err.Error() == "you have already disliked this post" {
		// 	// Redirect back to the main page with a notification
		// 	http.Redirect(w, r, "/?notification=already_disliked", http.StatusSeeOther)
		// 	return
		// }
		ErrorHandler(w, r, http.StatusInternalServerError, "Error disliking post")
		return
	}

	// Update the post's like and dislike counts
	err = models.UpdatePostLikesDislikes(postID)
	if err != nil {
		ErrorHandler(w, r, http.StatusInternalServerError, "Error updating dislike count")
		return
	}

	// Redirect back to the main page
	// http.Redirect(w, r, "/", http.StatusSeeOther)
	referer := r.Header.Get("Referer")
	http.Redirect(w, r, referer, http.StatusSeeOther)
}

func PostPageHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		ErrorHandler(w, r, http.StatusMethodNotAllowed, http.StatusText(http.StatusMethodNotAllowed))
		return
	}

	var loggedIn bool
	var username string
	var userID string
	cookie, err := r.Cookie("session_token")
	if err == nil {
		userID, username, err = models.GetIDBySessionToken(cookie.Value)
		if err == nil {
			loggedIn = true
		}
	}

	role, _ := models.GetUserRole(userID)
	isAdminOrModerator := role == "Administrator" || role == "Moderator"

	// Get the post ID from the query string
	postID := r.URL.Query().Get("id")
	if postID == "" {
		ErrorHandler(w, r, http.StatusBadRequest, "Missing post ID")
		return
	}

	// Fetch the post by ID
	post, err := models.GetPostByID(postID)
	if err != nil {
		if err == sql.ErrNoRows {
			ErrorHandler(w, r, http.StatusNotFound, "Post not found")
			return
		}
		ErrorHandler(w, r, http.StatusInternalServerError, "Error fetching post")
		return
	}

	// Fetch comments for the post
	comments, err := models.GetCommentsForPost(postID)
	if err != nil {
		ErrorHandler(w, r, http.StatusInternalServerError, "Error fetching comments")
		return
	}

	// notification := r.URL.Query().Get("notification")

	// Load the comments.html template
	tmpl, err := template.ParseFiles("templates/comments.html")
	if err != nil {
		http.Error(w, "Error loading template", http.StatusInternalServerError)
		return
	}

	data := struct {
		Post     models.Post
		Comments []models.Comment
		LoggedIn bool
		Username string
		// Notification string
		IsAdminOrModerator bool
	}{
		Post:     post,
		Comments: comments,
		LoggedIn: loggedIn,
		Username: username,
		// Notification: notification,
		IsAdminOrModerator: isAdminOrModerator,
	}

	tmpl.Execute(w, data)
}

func MyPostsHandler(w http.ResponseWriter, r *http.Request) {
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

	// Fetch posts created by the logged-in user
	posts, err := models.GetPostsByUser(userID)
	if err != nil {
		ErrorHandler(w, r, http.StatusInternalServerError, "Error fetching posts")
		return
	}

	// Render the posts page with "My Posts"
	tmpl, err := template.ParseFiles("templates/posts.html")
	if err != nil {
		http.Error(w, "Error loading template", http.StatusInternalServerError)
		return
	}

	data := struct {
		Posts      []models.Post
		LoggedIn   bool
		IsApproved bool
	}{
		Posts:      posts,
		LoggedIn:   true,
		IsApproved: true,
	}

	tmpl.Execute(w, data)
}

func LikedPostsHandler(w http.ResponseWriter, r *http.Request) {
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

	// Fetch posts liked by the logged-in user
	posts, err := models.GetLikedPostsByUser(userID)
	if err != nil {
		ErrorHandler(w, r, http.StatusInternalServerError, "Error fetching liked posts")
		return
	}

	// Render the posts page with "Liked Posts"
	tmpl, err := template.ParseFiles("templates/posts.html")
	if err != nil {
		http.Error(w, "Error loading template", http.StatusInternalServerError)
		return
	}

	data := struct {
		Posts      []models.Post
		LoggedIn   bool
		IsApproved bool
	}{
		Posts:      posts,
		LoggedIn:   true,
		IsApproved: true,
	}

	tmpl.Execute(w, data)
}

func GetPendingPostsHandler(w http.ResponseWriter, r *http.Request) {
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

	posts, err := models.GetPendingPosts()
	if err != nil {
		ErrorHandler(w, r, http.StatusInternalServerError, "Error fetching pending posts")
		return
	}

	tmpl, err := template.ParseFiles("templates/posts.html")
	if err != nil {
		ErrorHandler(w, r, http.StatusInternalServerError, "Error loading template")
		return
	}

	data := struct {
		Posts      []models.Post
		LoggedIn   bool
		IsApproved bool
	}{
		Posts:      posts,
		LoggedIn:   true,
		IsApproved: false,
	}

	tmpl.Execute(w, data)
}

func ApprovePostHandler(w http.ResponseWriter, r *http.Request) {
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

	postID := r.FormValue("post_id")
	if postID == "" {
		ErrorHandler(w, r, http.StatusBadRequest, "Missing post ID")
		return
	}

	if err := models.ApprovePost(postID); err != nil {
		ErrorHandler(w, r, http.StatusInternalServerError, "Failed to approve post")
		return
	}

	http.Redirect(w, r, "/admin/pending", http.StatusSeeOther)
}

func DeletePostHandler(w http.ResponseWriter, r *http.Request) {
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

	postID := r.FormValue("post_id")
	if postID == "" {
		ErrorHandler(w, r, http.StatusBadRequest, "Missing post ID")
		return
	}

	if err := models.DeletePost(postID); err != nil {
		ErrorHandler(w, r, http.StatusInternalServerError, "Failed to delete post")
		return
	}

	http.Redirect(w, r, r.Header.Get("Referer"), http.StatusSeeOther)
}

func MarkPostForModerationHandler(w http.ResponseWriter, r *http.Request) {
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

	postID := r.FormValue("post_id")
	if postID == "" {
		ErrorHandler(w, r, http.StatusBadRequest, "Missing post ID")
		return
	}

	if err := models.MarkPostForModeration(postID); err != nil {
		ErrorHandler(w, r, http.StatusInternalServerError, "Failed to mark post as unapproved")
		return
	}

	http.Redirect(w, r, r.Header.Get("Referer"), http.StatusSeeOther)
}

func CreatePostPageHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		ErrorHandler(w, r, http.StatusMethodNotAllowed, http.StatusText(http.StatusMethodNotAllowed))
		return
	}

	loggedIn := false

	cookie, err := r.Cookie("session_token")
	if err == nil {
		sessionToken := cookie.Value

		// Get the username of the logged-in user
		_, _, err = models.GetIDBySessionToken(sessionToken)
		if err == nil {
			loggedIn = true // User is logged in
		} else {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
	}

	categories, err := models.GetAllCategories()
	if err != nil {
		ErrorHandler(w, r, http.StatusInternalServerError, "Failed to load categories")
		return
	}

	tmpl, err := template.ParseFiles("templates/new_post.html")
	if err != nil {
		ErrorHandler(w, r, http.StatusInternalServerError, "Failed to load template")
		return
	}

	err = tmpl.Execute(w, struct {
		Categories []models.Category
		LoggedIn   bool
	}{
		Categories: categories,
		LoggedIn:   loggedIn,
	})
	if err != nil {
		ErrorHandler(w, r, http.StatusInternalServerError, "Error rendering template")
	}
}
