package main

import (
	"log"
	"net/http"

	"forum/handlers"
	"forum/models"
)

func main() {
	db, err := initDB()
	if err != nil {
		log.Fatal(err)
	}

	models.SetDB(db)

	// Routes
	http.HandleFunc("/", handlers.MainPageHandler)
	http.HandleFunc("/register", handlers.RegisterHandler)
	http.HandleFunc("/login", handlers.LoginHandler)
	http.HandleFunc("/logout", handlers.LogoutHandler)
	http.HandleFunc("/create_post", handlers.CreatePostHandler)
	http.HandleFunc("/new_post", handlers.CreatePostPageHandler)
	http.HandleFunc("/post", handlers.PostPageHandler)
	http.HandleFunc("/like", handlers.LikeHandler)
	http.HandleFunc("/dislike", handlers.DislikeHandler)
	http.HandleFunc("/create_comment", handlers.CreateCommentHandler)
	http.HandleFunc("/like_comment", handlers.LikeCommentHandler)
	http.HandleFunc("/dislike_comment", handlers.DislikeCommentHandler)
	http.HandleFunc("/my_posts", handlers.MyPostsHandler)
	http.HandleFunc("/liked_posts", handlers.LikedPostsHandler)
	http.HandleFunc("/users", handlers.UsersPageHandler)
	http.HandleFunc("/promote", handlers.PromoteToModeratorHandler)
	http.HandleFunc("/demote", handlers.DemoteToUserHandler)
	http.HandleFunc("/admin/categories", handlers.AdminCategoriesHandler)
	http.HandleFunc("/admin/categories/add", handlers.AddCategoryHandler)
	http.HandleFunc("/admin/categories/delete", handlers.DeleteCategoryHandler)
	http.HandleFunc("/admin/categories/update", handlers.UpdateCategoryHandler)
	http.HandleFunc("/admin/categories/controversial", handlers.SetControversialHandler)
	http.HandleFunc("/admin/pending", handlers.GetPendingPostsHandler)
	http.HandleFunc("/admin/approve", handlers.ApprovePostHandler)
	http.HandleFunc("/admin/delete_post", handlers.DeletePostHandler)
	http.HandleFunc("/admin/delete_comment", handlers.DeleteCommentHandler)
	http.HandleFunc("/admin/unapprove", handlers.MarkPostForModerationHandler)
	http.HandleFunc("/moderator", handlers.RenderApplyModeratorPageHandler)
	http.HandleFunc("/moderator/applications", handlers.RenderModeratorApplicationsPageHandler)
	http.HandleFunc("/moderator/approve", handlers.ApproveModeratorApplicationHandler)
	http.HandleFunc("/moderator/reject", handlers.RejectModeratorApplicationHandler)
	http.HandleFunc("/apply_moderator", handlers.ApplyModeratorHandler)
	http.HandleFunc("/auth/google/login", handlers.GoogleLoginHandler)
	http.HandleFunc("/auth/google/callback", handlers.GoogleCallbackHandler)
	http.HandleFunc("/auth/github/login", handlers.GitHubLoginHandler)
	http.HandleFunc("/auth/github/callback", handlers.GitHubCallbackHandler)
	http.Handle("/ui/", http.StripPrefix("/ui/", http.FileServer(http.Dir("./ui"))))
	http.Handle("/uploads/", http.StripPrefix("/uploads/", http.FileServer(http.Dir("uploads"))))

	log.Println("Server started on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
