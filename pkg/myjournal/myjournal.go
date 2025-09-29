package myjournal

import (
	"context"
	"fmt"

	"github.com/dkrotx/cassandra-learning/pkg/entities"
	"github.com/gocql/gocql"
	"github.com/google/uuid"
)

func (db *JournalDB) CreatePost(ctx context.Context, userID entities.UserID, post *entities.PostData) error {
	batch := db.session.NewBatch(gocql.LoggedBatch)

	postID := gocql.TimeUUID()

	batch.Query(`INSERT INTO posts_by_user (user_id, post_id, post, tags) VALUES (?, ?, {title: ?, body: ?}, ?)`,
		userID.String(),
		postID,
		post.Title,
		post.Body,
		post.Tags,
	)

	batch.Query(`INSERT INTO user_by_postid (post_id, user_id) VALUES (?, ?)`,
		postID,
		userID.String(),
	)

	if err := db.session.ExecuteBatch(batch); err != nil {
		return fmt.Errorf("failed to create post: %v", err)
	}
	return nil
}

func (db *JournalDB) ReadPostsByUser(ctx context.Context, userID entities.UserID) ([]*entities.Post, error) {
	query := `SELECT post_id, post.title, post.body, tags FROM posts_by_user WHERE user_id = ? LIMIT ?`
	var posts []*entities.Post

	iter := db.session.Query(query, userID.String(), 3).WithContext(ctx).Iter()

	var postID gocql.UUID
	var tags []string
	var title string
	var body string

	for iter.Scan(&postID, &title, &body, &tags) {
		post := &entities.Post{
			PostID:    uuid.UUID(postID),
			CreatedAt: postID.Time(),
			PostData: entities.PostData{
				Title: title,
				Body:  body,
				Tags:  tags,
			},
		}

		posts = append(posts, post)
	}

	// Check for any errors that occurred during iteration
	if err := iter.Close(); err != nil {
		return nil, fmt.Errorf("failed to read posts: %v", err)
	}

	return posts, nil
}

func (db *JournalDB) DeletePost(ctx context.Context, postID entities.PostID) error {
	// First, get the user_id for this post
	var userID string
	query := `SELECT user_id FROM user_by_postid WHERE post_id = ?`
	if err := db.session.Query(query, gocql.UUID(postID)).WithContext(ctx).Scan(&userID); err != nil {
		return fmt.Errorf("failed to find user for post: %v", err)
	}

	// Delete from both tables
	batch := db.session.NewBatch(gocql.LoggedBatch)

	batch.Query(`DELETE FROM posts_by_user WHERE user_id = ? AND post_id = ?`,
		userID,
		gocql.UUID(postID),
	)

	batch.Query(`DELETE FROM user_by_postid WHERE post_id = ?`,
		gocql.UUID(postID),
	)

	if err := db.session.ExecuteBatch(batch); err != nil {
		return fmt.Errorf("failed to delete post: %v", err)
	}

	return nil
}

func (db *JournalDB) DeletePostByUser(ctx context.Context, postID entities.PostID, userID entities.UserID) error {
	// Verify the post belongs to the user
	var foundUserID string
	query := `SELECT user_id FROM user_by_postid WHERE post_id = ?`
	if err := db.session.Query(query, gocql.UUID(postID)).WithContext(ctx).Scan(&foundUserID); err != nil {
		return fmt.Errorf("failed to find post: %v", err)
	}

	if foundUserID != userID.String() {
		return fmt.Errorf("post does not belong to the specified user")
	}

	// Delete from both tables
	batch := db.session.NewBatch(gocql.LoggedBatch)

	batch.Query(`DELETE FROM posts_by_user WHERE user_id = ? AND post_id = ?`,
		userID.String(),
		gocql.UUID(postID),
	)

	batch.Query(`DELETE FROM user_by_postid WHERE post_id = ?`,
		gocql.UUID(postID),
	)

	if err := db.session.ExecuteBatch(batch); err != nil {
		return fmt.Errorf("failed to delete post: %v", err)
	}

	return nil
}

func (db *JournalDB) AddTags(ctx context.Context, postID entities.PostID, userID entities.UserID, tags []string) error {
	// Add tags to the post
	query := `UPDATE posts_by_user SET tags = tags + ? WHERE user_id = ? AND post_id = ?`
	if err := db.session.Query(query, tags, userID.String(), gocql.UUID(postID)).WithContext(ctx).Exec(); err != nil {
		return fmt.Errorf("failed to add tags: %v", err)
	}

	return nil
}

func (db *JournalDB) RemoveTags(ctx context.Context, postID entities.PostID, userID entities.UserID, tags []string) error {
	// Remove tags from the post
	query := `UPDATE posts_by_user SET tags = tags - ? WHERE user_id = ? AND post_id = ?`
	if err := db.session.Query(query, tags, userID.String(), gocql.UUID(postID)).WithContext(ctx).Exec(); err != nil {
		return fmt.Errorf("failed to remove tags: %v", err)
	}

	return nil
}
