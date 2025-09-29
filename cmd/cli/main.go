package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/dkrotx/cassandra-learning/pkg/entities"
	"github.com/dkrotx/cassandra-learning/pkg/myjournal"
	"github.com/google/uuid"
	"github.com/urfave/cli/v3"
	"go.uber.org/config"
	"go.uber.org/zap"
)

func main() {
	// Initialize database
	db, err := initDB()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize database: %v\n", err)
		os.Exit(1)
	}

	// Run the CLI
	runCLI(db)
}

func runCLI(db *myjournal.JournalDB) {
	cliApp := &cli.Command{
		Name:  "myjournal-cli",
		Usage: "A CLI tool for managing journal posts",
		Commands: []*cli.Command{
			{
				Name:    "create",
				Aliases: []string{"c"},
				Usage:   "Create a new post",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "user-id",
						Aliases:  []string{"u"},
						Usage:    "User ID (UUID format)",
						Required: true,
					},
					&cli.StringFlag{
						Name:     "title",
						Aliases:  []string{"t"},
						Usage:    "Post title",
						Required: true,
					},
					&cli.StringFlag{
						Name:     "body",
						Aliases:  []string{"b"},
						Usage:    "Post body",
						Required: true,
					},
					&cli.StringSliceFlag{
						Name:    "tags",
						Aliases: []string{"tag"},
						Usage:   "Post tags (can be specified multiple times)",
					},
				},
				Action: func(ctx context.Context, c *cli.Command) error {
					return createPost(ctx, c, db)
				},
			},
			{
				Name:    "read",
				Aliases: []string{"r"},
				Usage:   "Read posts by user",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "user-id",
						Aliases:  []string{"u"},
						Usage:    "User ID (UUID format)",
						Required: true,
					},
				},
				Action: func(ctx context.Context, c *cli.Command) error {
					return readPosts(ctx, c, db)
				},
			},
			{
				Name:    "delete",
				Aliases: []string{"d"},
				Usage:   "Delete a post",
				Commands: []*cli.Command{
					{
						Name:    "by-id",
						Aliases: []string{"id"},
						Usage:   "Delete post by post ID only",
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:     "post-id",
								Aliases:  []string{"p"},
								Usage:    "Post ID (UUID format)",
								Required: true,
							},
						},
						Action: func(ctx context.Context, c *cli.Command) error {
							return deletePost(ctx, c, db)
						},
					},
					{
						Name:    "by-user",
						Aliases: []string{"user"},
						Usage:   "Delete post by post ID and user ID",
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:     "post-id",
								Aliases:  []string{"p"},
								Usage:    "Post ID (UUID format)",
								Required: true,
							},
							&cli.StringFlag{
								Name:     "user-id",
								Aliases:  []string{"u"},
								Usage:    "User ID (UUID format)",
								Required: true,
							},
						},
						Action: func(ctx context.Context, c *cli.Command) error {
							return deletePostByUser(ctx, c, db)
						},
					},
				},
			},
			{
				Name:    "tags",
				Aliases: []string{"t"},
				Usage:   "Manage post tags",
				Commands: []*cli.Command{
					{
						Name:    "add",
						Aliases: []string{"a"},
						Usage:   "Add tags to a post",
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:     "post-id",
								Aliases:  []string{"p"},
								Usage:    "Post ID (UUID format)",
								Required: true,
							},
							&cli.StringFlag{
								Name:     "user-id",
								Aliases:  []string{"u"},
								Usage:    "User ID (UUID format)",
								Required: true,
							},
							&cli.StringSliceFlag{
								Name:     "tags",
								Aliases:  []string{"tag"},
								Usage:    "Tags to add (can be specified multiple times)",
								Required: true,
							},
						},
						Action: func(ctx context.Context, c *cli.Command) error {
							return addTags(ctx, c, db)
						},
					},
					{
						Name:    "remove",
						Aliases: []string{"r"},
						Usage:   "Remove tags from a post",
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:     "post-id",
								Aliases:  []string{"p"},
								Usage:    "Post ID (UUID format)",
								Required: true,
							},
							&cli.StringFlag{
								Name:     "user-id",
								Aliases:  []string{"u"},
								Usage:    "User ID (UUID format)",
								Required: true,
							},
							&cli.StringSliceFlag{
								Name:     "tags",
								Aliases:  []string{"tag"},
								Usage:    "Tags to remove (can be specified multiple times)",
								Required: true,
							},
						},
						Action: func(ctx context.Context, c *cli.Command) error {
							return removeTags(ctx, c, db)
						},
					},
				},
			},
		},
	}

	if err := cliApp.Run(context.Background(), os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func createPost(ctx context.Context, c *cli.Command, db *myjournal.JournalDB) error {
	// Parse user ID
	userIDStr := c.String("user-id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return fmt.Errorf("invalid user ID: %v", err)
	}

	// Get post data
	title := c.String("title")
	body := c.String("body")
	tags := c.StringSlice("tags")

	// Create post
	postData := &entities.PostData{
		Title: title,
		Body:  body,
		Tags:  tags,
	}

	err = db.CreatePost(ctx, entities.UserID(userID), postData)
	if err != nil {
		return fmt.Errorf("failed to create post: %v", err)
	}

	fmt.Printf("Successfully created post: %s\n", title)
	return nil
}

func readPosts(ctx context.Context, c *cli.Command, db *myjournal.JournalDB) error {
	// Parse user ID
	userIDStr := c.String("user-id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return fmt.Errorf("invalid user ID: %v", err)
	}

	// Read posts
	posts, err := db.ReadPostsByUser(ctx, entities.UserID(userID))
	if err != nil {
		return fmt.Errorf("failed to read posts: %v", err)
	}

	// Display posts
	if len(posts) == 0 {
		fmt.Println("No posts found for this user.")
		return nil
	}

	fmt.Printf("Found %d posts for user %s:\n\n", len(posts), userIDStr)
	for i, post := range posts {
		fmt.Printf("Post %d:\n", i+1)
		fmt.Printf("  ID: %s\n", post.PostID)
		fmt.Printf("  Title: %s\n", post.Title)
		fmt.Printf("  Body: %s\n", post.Body)
		fmt.Printf("  Created At: %s\n", post.CreatedAt.Format("2006-01-02 15:04:05"))
		fmt.Printf("  Tags: %s\n", strings.Join(post.Tags, ", "))
		fmt.Println()
	}

	return nil
}

func deletePost(ctx context.Context, c *cli.Command, db *myjournal.JournalDB) error {
	// Parse post ID
	postIDStr := c.String("post-id")
	postID, err := uuid.Parse(postIDStr)
	if err != nil {
		return fmt.Errorf("invalid post ID: %v", err)
	}

	// Delete post
	err = db.DeletePost(ctx, entities.PostID(postID))
	if err != nil {
		return fmt.Errorf("failed to delete post: %v", err)
	}

	fmt.Printf("Successfully deleted post: %s\n", postIDStr)
	return nil
}

func deletePostByUser(ctx context.Context, c *cli.Command, db *myjournal.JournalDB) error {
	// Parse post ID
	postIDStr := c.String("post-id")
	postID, err := uuid.Parse(postIDStr)
	if err != nil {
		return fmt.Errorf("invalid post ID: %v", err)
	}

	// Parse user ID
	userIDStr := c.String("user-id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return fmt.Errorf("invalid user ID: %v", err)
	}

	// Delete post
	err = db.DeletePostByUser(ctx, entities.PostID(postID), entities.UserID(userID))
	if err != nil {
		return fmt.Errorf("failed to delete post: %v", err)
	}

	fmt.Printf("Successfully deleted post %s for user %s\n", postIDStr, userIDStr)
	return nil
}

func addTags(ctx context.Context, c *cli.Command, db *myjournal.JournalDB) error {
	// Parse post ID
	postIDStr := c.String("post-id")
	postID, err := uuid.Parse(postIDStr)
	if err != nil {
		return fmt.Errorf("invalid post ID: %v", err)
	}

	// Parse user ID
	userIDStr := c.String("user-id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return fmt.Errorf("invalid user ID: %v", err)
	}

	// Get tags
	tags := c.StringSlice("tags")

	// Add tags
	err = db.AddTags(ctx, entities.PostID(postID), entities.UserID(userID), tags)
	if err != nil {
		return fmt.Errorf("failed to add tags: %v", err)
	}

	fmt.Printf("Successfully added tags %v to post %s\n", tags, postIDStr)
	return nil
}

func removeTags(ctx context.Context, c *cli.Command, db *myjournal.JournalDB) error {
	// Parse post ID
	postIDStr := c.String("post-id")
	postID, err := uuid.Parse(postIDStr)
	if err != nil {
		return fmt.Errorf("invalid post ID: %v", err)
	}

	// Parse user ID
	userIDStr := c.String("user-id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return fmt.Errorf("invalid user ID: %v", err)
	}

	// Get tags
	tags := c.StringSlice("tags")

	// Remove tags
	err = db.RemoveTags(ctx, entities.PostID(postID), entities.UserID(userID), tags)
	if err != nil {
		return fmt.Errorf("failed to remove tags: %v", err)
	}

	fmt.Printf("Successfully removed tags %v from post %s\n", tags, postIDStr)
	return nil
}

func initDB() (*myjournal.JournalDB, error) {
	// Load config from config/config.yaml
	configProvider, err := config.NewYAML(config.File("config/config.yaml"))
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %v", err)
	}

	// Create logger
	logger, err := zap.NewDevelopment()
	if err != nil {
		return nil, fmt.Errorf("failed to create logger: %v", err)
	}

	// Create database connection
	db, err := myjournal.NewDB(configProvider, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create database connection: %v", err)
	}

	return db, nil
}
