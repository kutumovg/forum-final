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

func CreateComment(postID, userID, content string) error {
	commentID, err := uuid.NewV4()
	if err != nil {
		return err
	}

	_, err = db.Exec("INSERT INTO comments (id, post_id, user_id, content, created_at) VALUES (?, ?, ?, ?, ?)",
		commentID.String(), postID, userID, content, time.Now())
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