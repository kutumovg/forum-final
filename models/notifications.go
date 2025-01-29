package models

import (
	"database/sql"
	"time"
)

// Notification represents a user notification
type Notification struct {
	ID                 int
	Username           string
	AuthorID           string
	PostID             *string
	CommentID          *string
	Action             string
	CreatedAt          time.Time
	CreatedAtFormatted string
	PostContent        *string
	CommentContent     *string
	PostURL            *string
	IsRead             bool
}

// AddNotification creates a new notification in the database
func AddNotification(userID string, authorID string, postID, commentID *string, action string) error {
	_, err := db.Exec("INSERT INTO notifications (user_id, author_id, post_id, comment_id, action, created_at) VALUES (?, ?, ?, ?, ?, ?)", userID, authorID, postID, commentID, action, time.Now())
	return err
}

// GetNotifications retrieves notifications for a specific user
func GetNotifications(authorID string) ([]Notification, error) {
	const query = `
	SELECT n.id, u.username AS username, n.author_id, n.post_id, n.comment_id, n.action, n.created_at, n.is_read, p.content AS post_content, c.content AS comment_content
    FROM notifications n
	LEFT JOIN users u ON n.user_id = u.id
    LEFT JOIN posts p ON n.post_id = p.id
    LEFT JOIN comments c ON n.comment_id = c.id
    WHERE n.author_id = ?
    ORDER BY n.created_at DESC
	`
	rows, err := db.Query(query, authorID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notifications []Notification
	for rows.Next() {
		var n Notification
		var createdAt time.Time
		var isRead bool
		var postID, commentID, postContent, commentContent sql.NullString
		if err := rows.Scan(&n.ID, &n.Username, &n.AuthorID, &postID, &commentID, &n.Action, &createdAt, &isRead, &postContent, &commentContent); err != nil {
			return nil, err
		}
		n.CreatedAtFormatted = createdAt.Format("02.01.2006 15:04")
		n.IsRead = isRead
		if postID.Valid {
			n.PostID = &postID.String
			if postContent.Valid {
				truncated := truncateContent(postContent.String)
				n.PostContent = &truncated
			} else {
				continue
			}
			postURL := "/post?id=" + postID.String
			n.PostURL = &postURL
		}
		if commentID.Valid {
			n.CommentID = &commentID.String
			if commentContent.Valid {
				truncated := truncateContent(commentContent.String)
				n.CommentContent = &truncated
			} else {
				continue
			}
			postID, err := GetPostIDByCommentID(*n.CommentID)
			if err != nil {
				return nil, err
			}
			if postID != "" {
				postURL := "/post?id=" + postID
				n.PostURL = &postURL
			}
		}
		notifications = append(notifications, n)
	}

	if err := markNotificationsAsRead(authorID); err != nil {
		return nil, err
	}
	return notifications, nil
}

func truncateContent(content string) string {
	if len(content) > 80 {
		return content[:80] + "..."
	}
	return content
}

func markNotificationsAsRead(authorID string) error {
	const query = `
    UPDATE notifications
    SET is_read = TRUE
    WHERE author_id = ? AND is_read = FALSE
    `
	_, err := db.Exec(query, authorID)
	return err
}

func GetNotificationCount(authorID string) (int, error) {
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM notifications WHERE author_id = ? AND is_read = FALSE", authorID).Scan(&count)
	return count, err
}
