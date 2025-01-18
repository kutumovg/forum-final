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
	wrappedMux.HandleFunc("/", handlers.MainPageHandler)
	wrappedMux.HandleFunc("/register", handlers.RegisterHandler)
	wrappedMux.HandleFunc("/login", handlers.LoginHandler)
	wrappedMux.HandleFunc("/logout", handlers.LogoutHandler)
	wrappedMux.HandleFunc("/create_post", handlers.CreatePostHandler)
	wrappedMux.HandleFunc("/new_post", handlers.CreatePostPageHandler)
	wrappedMux.HandleFunc("/post", handlers.PostPageHandler)
	wrappedMux.HandleFunc("/like", handlers.LikeHandler)
	wrappedMux.HandleFunc("/dislike", handlers.DislikeHandler)
	wrappedMux.HandleFunc("/create_comment", handlers.CreateCommentHandler)
	wrappedMux.HandleFunc("/like_comment", handlers.LikeCommentHandler)
	wrappedMux.HandleFunc("/dislike_comment", handlers.DislikeCommentHandler)
	wrappedMux.HandleFunc("/my_posts", handlers.MyPostsHandler)
	wrappedMux.HandleFunc("/liked_posts", handlers.LikedPostsHandler)
	wrappedMux.HandleFunc("/users", handlers.UsersPageHandler)
	wrappedMux.HandleFunc("/promote", handlers.PromoteToModeratorHandler)
	wrappedMux.HandleFunc("/demote", handlers.DemoteToUserHandler)
	wrappedMux.HandleFunc("/admin/categories", handlers.AdminCategoriesHandler)
	wrappedMux.HandleFunc("/admin/categories/add", handlers.AddCategoryHandler)
	wrappedMux.HandleFunc("/admin/categories/delete", handlers.DeleteCategoryHandler)
	wrappedMux.HandleFunc("/admin/categories/update", handlers.UpdateCategoryHandler)
	wrappedMux.HandleFunc("/admin/categories/controversial", handlers.SetControversialHandler)
	wrappedMux.HandleFunc("/admin/pending", handlers.GetPendingPostsHandler)
	wrappedMux.HandleFunc("/admin/approve", handlers.ApprovePostHandler)
	wrappedMux.HandleFunc("/admin/delete_post", handlers.DeletePostHandler)
	wrappedMux.HandleFunc("/admin/delete_comment", handlers.DeleteCommentHandler)
	wrappedMux.HandleFunc("/admin/unapprove", handlers.MarkPostForModerationHandler)
	wrappedMux.HandleFunc("/moderator", handlers.RenderApplyModeratorPageHandler)
	wrappedMux.HandleFunc("/moderator/applications", handlers.RenderModeratorApplicationsPageHandler)
	wrappedMux.HandleFunc("/moderator/approve", handlers.ApproveModeratorApplicationHandler)
	wrappedMux.HandleFunc("/moderator/reject", handlers.RejectModeratorApplicationHandler)
	wrappedMux.HandleFunc("/apply_moderator", handlers.ApplyModeratorHandler)
	wrappedMux.HandleFunc("/auth/google/login", handlers.GoogleLoginHandler)
	wrappedMux.HandleFunc("/auth/google/callback", handlers.GoogleCallbackHandler)
	wrappedMux.HandleFunc("/auth/github/login", handlers.GitHubLoginHandler)
	wrappedMux.HandleFunc("/auth/github/callback", handlers.GitHubCallbackHandler)
	
	http.Handle("/ui/", http.StripPrefix("/ui/", http.FileServer(http.Dir("./ui"))))
	http.Handle("/uploads/", http.StripPrefix("/uploads/", http.FileServer(http.Dir("uploads"))))

	secureMux := RateLimiter(wrappedMux)

	// // HTTPS server
	// log.Println("Starting server on https://localhost:8080")
	// err = http.ListenAndServeTLS(
	// 	":8080",
	// 	"/etc/letsencrypt/live/forum.qarjy.kz/fullchain.pem", // Path to SSL certificate
	// 	"/etc/letsencrypt/live/forum.qarjy.kz/privkey.pem",   // Path to SSL private key
	// 	secureMux,
	// )
	// if err != nil {
	// 	log.Fatal(err)
	// }

	go func() {
		log.Println("Starting HTTPS server on https://localhost:443...")
		err := http.ListenAndServeTLS(
			":443",
			"/certs/fullchain1.pem",
			"/certs/privkey1.pem",
			secureMux,
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
}
