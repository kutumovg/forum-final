package main

import (
	"log"
	"net/http"

	"golang.org/x/time/rate"

	"forum/handlers"
	"forum/models"
)

func RateLimiter(next http.Handler) http.Handler {
	limiter := rate.NewLimiter(1, 3) // 1 request per second with a burst of 3
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !limiter.Allow() {
			handlers.ErrorHandler(w, r, http.StatusTooManyRequests, http.StatusText(http.StatusTooManyRequests))
			return
		}
		next.ServeHTTP(w, r)
	})
}

func redirectToHTTPS(w http.ResponseWriter, r *http.Request) {
	// Build the HTTPS URL
	httpsURL := "https://" + r.Host + r.URL.String()
	http.Redirect(w, r, httpsURL, http.StatusMovedPermanently)
}

func main() {
	db, err := initDB()
	if err != nil {
		log.Fatal(err)
	}

	models.SetDB(db)

	// Routes
	wrappedMux := http.NewServeMux()

	wrappedMux.Handle("/ui/", http.StripPrefix("/ui/", http.FileServer(http.Dir("./ui"))))
	wrappedMux.Handle("/uploads/", http.StripPrefix("/uploads/", http.FileServer(http.Dir("uploads"))))

	rateLimitedMux := http.NewServeMux()
	rateLimitedMux.HandleFunc("/", handlers.MainPageHandler)
	rateLimitedMux.HandleFunc("/register", handlers.RegisterHandler)
	rateLimitedMux.HandleFunc("/login", handlers.LoginHandler)
	rateLimitedMux.HandleFunc("/logout", handlers.LogoutHandler)
	rateLimitedMux.HandleFunc("/create_post", handlers.CreatePostHandler)
	rateLimitedMux.HandleFunc("/new_post", handlers.CreatePostPageHandler)
	rateLimitedMux.HandleFunc("/post", handlers.PostPageHandler)
	rateLimitedMux.HandleFunc("/like", handlers.LikeHandler)
	rateLimitedMux.HandleFunc("/dislike", handlers.DislikeHandler)
	rateLimitedMux.HandleFunc("/create_comment", handlers.CreateCommentHandler)
	rateLimitedMux.HandleFunc("/like_comment", handlers.LikeCommentHandler)
	rateLimitedMux.HandleFunc("/dislike_comment", handlers.DislikeCommentHandler)
	rateLimitedMux.HandleFunc("/my_posts", handlers.MyPostsHandler)
	rateLimitedMux.HandleFunc("/liked_posts", handlers.LikedPostsHandler)
	rateLimitedMux.HandleFunc("/users", handlers.UsersPageHandler)
	rateLimitedMux.HandleFunc("/promote", handlers.PromoteToModeratorHandler)
	rateLimitedMux.HandleFunc("/demote", handlers.DemoteToUserHandler)
	rateLimitedMux.HandleFunc("/admin/categories", handlers.AdminCategoriesHandler)
	rateLimitedMux.HandleFunc("/admin/categories/add", handlers.AddCategoryHandler)
	rateLimitedMux.HandleFunc("/admin/categories/delete", handlers.DeleteCategoryHandler)
	rateLimitedMux.HandleFunc("/admin/categories/update", handlers.UpdateCategoryHandler)
	rateLimitedMux.HandleFunc("/admin/categories/controversial", handlers.SetControversialHandler)
	rateLimitedMux.HandleFunc("/admin/pending", handlers.GetPendingPostsHandler)
	rateLimitedMux.HandleFunc("/admin/approve", handlers.ApprovePostHandler)
	rateLimitedMux.HandleFunc("/admin/delete_post", handlers.DeletePostHandler)
	rateLimitedMux.HandleFunc("/admin/delete_comment", handlers.DeleteCommentHandler)
	rateLimitedMux.HandleFunc("/admin/unapprove", handlers.MarkPostForModerationHandler)
	rateLimitedMux.HandleFunc("/moderator", handlers.RenderApplyModeratorPageHandler)
	rateLimitedMux.HandleFunc("/moderator/applications", handlers.RenderModeratorApplicationsPageHandler)
	rateLimitedMux.HandleFunc("/moderator/approve", handlers.ApproveModeratorApplicationHandler)
	rateLimitedMux.HandleFunc("/moderator/reject", handlers.RejectModeratorApplicationHandler)
	rateLimitedMux.HandleFunc("/apply_moderator", handlers.ApplyModeratorHandler)
	rateLimitedMux.HandleFunc("/auth/google/login", handlers.GoogleLoginHandler)
	rateLimitedMux.HandleFunc("/auth/google/callback", handlers.GoogleCallbackHandler)
	rateLimitedMux.HandleFunc("/auth/github/login", handlers.GitHubLoginHandler)
	rateLimitedMux.HandleFunc("/auth/github/callback", handlers.GitHubCallbackHandler)
	rateLimitedMux.HandleFunc("/notifications", handlers.ShowNotificationsHandler)
	rateLimitedMux.HandleFunc("/my_comments", handlers.MyCommentsPageHandler)
	rateLimitedMux.HandleFunc("/edit-post", handlers.EditPostHandler)
	rateLimitedMux.HandleFunc("/edit-comment", handlers.EditCommentHandler)

	wrappedMux.Handle("/", RateLimiter(rateLimitedMux))

	go func() {
		log.Println("Starting HTTPS server on https://localhost:443...")
		err := http.ListenAndServeTLS(
			":443",
			"/certs/fullchain1.pem",
			"/certs/privkey1.pem",
			wrappedMux,
		)
		if err != nil {
			log.Fatalf("Failed to start HTTPS server: %v", err)
		}
	}()

	log.Println("Starting HTTP redirect server on http://localhost:80...")
	err = http.ListenAndServe(":80", http.HandlerFunc(redirectToHTTPS))
	if err != nil {
		log.Fatalf("Failed to start HTTP redirect server: %v", err)
	}
	// log.Println("Server started on http://localhost:8080")
	// log.Fatal(http.ListenAndServe(":8080", wrappedMux))
}
