package models

import (
	"database/sql"
	"html"
	"html/template"
	"strings"
	"time"

	"github.com/gofrs/uuid"
)

type Comment struct {
	ID                 string
	PostID             string
	Content            template.HTML
	CreatedAt          time.Time
	CreatedAtFormatted string
	Likes              int
	Dislikes           int
	Author             string // The username of the comment's author
	UserHasLiked       bool   // Whether the logged-in user has liked this comment
	UserHasDisliked    bool   // Whether the logged-in user has disliked this comment
}

type UserActivity struct {
	CommentsByPost []Post
}

// type UserComment struct {
// 	CommentID                 string
// 	CommentContent            string
// 	CommentCreatedAt          time.Time
// 	CommentCreatedAtFormatted string
// 	PostID                    string
// 	PostContent               string
// 	PostCreatedAt             time.Time
// 	PostCreatedAtFormatted    string
// 	PostLikes                 int
// 	PostDislikes              int
// 	PostAuthor                string
// 	PostCategories            []string
// 	PostImagePath             string
// }

func CreateComment(postID, userID, content string) error {
	commentID, err := uuid.NewV4()
	if err != nil {
		return err
	}

	time := GetLocalTime()
	_, err = db.Exec("INSERT INTO comments (id, post_id, user_id, content, created_at) VALUES (?, ?, ?, ?, ?)",
		commentID.String(), postID, userID, content, time)
	if err != nil {
		return err
	}

	return nil
}

func LikeComment(userID, commentID string) error {
	var interactionID string
	var isLike bool

	err := db.QueryRow("SELECT id, is_like FROM comment_likes WHERE user_id = ? AND comment_id = ?", userID, commentID).Scan(&interactionID, &isLike)
	if err == sql.ErrNoRows {
		// Insert a new like record
		likeID, _ := uuid.NewV4()
		_, err = db.Exec("INSERT INTO comment_likes (id, user_id, comment_id, is_like) VALUES (?, ?, ?, TRUE)", likeID.String(), userID, commentID)
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	} else if isLike {
		_, err = db.Exec("DELETE FROM comment_likes WHERE user_id = ? AND comment_id = ?", userID, commentID)
		return err
	} else {
		// Change the dislike to a like
		_, err = db.Exec("UPDATE comment_likes SET is_like = TRUE WHERE id = ?", interactionID)
		if err != nil {
			return err
		}
	}

	return nil
}

func DislikeComment(userID, commentID string) error {
	var interactionID string
	var isLike bool

	err := db.QueryRow("SELECT id, is_like FROM comment_likes WHERE user_id = ? AND comment_id = ?", userID, commentID).Scan(&interactionID, &isLike)
	if err == sql.ErrNoRows {
		// Insert a new dislike record
		dislikeID, _ := uuid.NewV4()
		_, err = db.Exec("INSERT INTO comment_likes (id, user_id, comment_id, is_like) VALUES (?, ?, ?, FALSE)", dislikeID.String(), userID, commentID)
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	} else if !isLike {
		_, err = db.Exec("DELETE FROM comment_likes WHERE user_id = ? AND comment_id = ?", userID, commentID)
		return err
	} else {
		// Change the like to a dislike
		_, err = db.Exec("UPDATE comment_likes SET is_like = FALSE WHERE id = ?", interactionID)
		if err != nil {
			return err
		}
	}

	return nil
}

func UpdateCommentLikesDislikes(commentID string) error {
	var likeCount, dislikeCount int

	// Count likes
	err := db.QueryRow("SELECT COUNT(*) FROM comment_likes WHERE comment_id = ? AND is_like = TRUE", commentID).Scan(&likeCount)
	if err != nil {
		return err
	}

	// Count dislikes
	err = db.QueryRow("SELECT COUNT(*) FROM comment_likes WHERE comment_id = ? AND is_like = FALSE", commentID).Scan(&dislikeCount)
	if err != nil {
		return err
	}

	// Update the posts table with new like and dislike counts
	_, err = db.Exec("UPDATE comments SET likes = ?, dislikes = ? WHERE id = ?", likeCount, dislikeCount, commentID)
	return err
}

func GetCommentsForPost(postID string) ([]Comment, error) {
	rows, err := db.Query(`
        SELECT comments.id, comments.content, comments.created_at, users.username, comments.likes, comments.dislikes
        FROM comments
        JOIN users ON comments.user_id = users.id
        WHERE comments.post_id = ?
        ORDER BY comments.created_at ASC
    `, postID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var comments []Comment
	for rows.Next() {
		var comment Comment
		var createdAt time.Time
		err := rows.Scan(&comment.ID, &comment.Content, &createdAt, &comment.Author, &comment.Likes, &comment.Dislikes)
		if err != nil {
			return nil, err
		}
		comment.CreatedAtFormatted = createdAt.Format("02.01.2006 15:04")
		comment.Content = template.HTML(strings.ReplaceAll(string(comment.Content), "\n", "<br>"))
		comments = append(comments, comment)
	}

	return comments, nil
}

func GetCommentsByUser(userID string) ([]Post, error) {
	const query = `
	SELECT 
		posts.id AS post_id,
		posts.content AS post_content,
		comments.id AS comment_id,
		comments.content AS comment_content,
		comments.created_at AS comment_created_at,
		comments.likes AS comment_likes,
		comments.dislikes AS comment_dislikes
	FROM comments
	JOIN posts ON comments.post_id = posts.id
	JOIN users ON posts.user_id = users.id
	WHERE comments.user_id = ?
	ORDER BY comments.created_at DESC;
	`

	rows, err := db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Use a map to group comments by post
	groupMap := make(map[string]*Post)

	for rows.Next() {
		var postID, commentID string
		var commentLikes, commentDislikes int
		var commentCreatedAt time.Time
		var postContent, commentContent template.HTML

		if err := rows.Scan(&postID, &postContent, &commentID, &commentContent, &commentCreatedAt, &commentLikes, &commentDislikes); err != nil {
			return nil, err
		}

		commentCreatedAtFormatted := commentCreatedAt.Format("02.01.2006 15:04")
		postContent = template.HTML(strings.ReplaceAll(string(postContent), "\n", "<br>"))
		commentContent = template.HTML(strings.ReplaceAll(string(commentContent), "\n", "<br>"))

		// Check if the post is already in the map
		if _, exists := groupMap[postID]; !exists {
			groupMap[postID] = &Post{
				ID:       postID,
				Content:  postContent,
				Comments: []Comment{},
			}
		}

		// Add the comment to the corresponding post group
		groupMap[postID].Comments = append(groupMap[postID].Comments, Comment{
			ID:                 commentID,
			Content:            commentContent,
			CreatedAtFormatted: commentCreatedAtFormatted,
			Likes:              commentLikes,
			Dislikes:           commentDislikes,
		})
	}

	// Convert the map to a slice
	var groupedComments []Post
	for _, group := range groupMap {
		groupedComments = append(groupedComments, *group)
	}

	return groupedComments, nil
}

func SanitizeInput(input string) string {
	// input = html.UnescapeString(input)
	// input = strings.ReplaceAll(input, "<br>", "")
	// input = strings.ReplaceAll(input, "<BR>", "")
	input = html.EscapeString(input)
	input = strings.TrimSpace(input)
	return input
}

func IsValidContent(content string) bool {
	sanitized := SanitizeInput(content)
	return len(sanitized) > 0
}

func DeleteComment(commentID string) error {
	_, err := db.Exec("DELETE FROM comments WHERE id = ?", commentID)
	return err
}

func GetCommentCount(postID string) (int, error) {
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM comments WHERE post_id = ?", postID).Scan(&count)
	return count, err
}

func UpdateComment(commentID, userID, content string) error {
	const query = `
	UPDATE comments
	SET content = ?, created_at = ?
	WHERE id = ? AND user_id = ?;
	`
	time := GetLocalTime()
	_, err := db.Exec(query, content, time, commentID, userID)
	return err
}

func GetCommentByID(commentID string) (Comment, error) {
	var comment Comment
	err := db.QueryRow(`
		SELECT id, post_id, content, created_at, user_id
		FROM comments
		WHERE id = ?
	`, commentID).Scan(&comment.ID, &comment.PostID, &comment.Content, &comment.CreatedAt, &comment.Author)

	if err != nil {
		return comment, err
	}

	comment.CreatedAtFormatted = comment.CreatedAt.Format("02.01.2006 15:04")
	return comment, nil
}
