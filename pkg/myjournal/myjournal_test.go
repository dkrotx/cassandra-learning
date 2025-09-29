package myjournal

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/dkrotx/cassandra-learning/pkg/entities"
	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/cassandra"
	"go.uber.org/config"
	"go.uber.org/zap"
)

var testUserUUID = entities.UserID(uuid.MustParse("10000000-1000-f000-f000-000000000000"))

// JournalTestSuite holds the database connection and container for all tests
type JournalTestSuite struct {
	suite.Suite
	db        *JournalDB
	container testcontainers.Container
	ctx       context.Context
	cancel    context.CancelFunc
}

// SetupSuite runs once before all tests
func (suite *JournalTestSuite) SetupSuite() {
	// Skip if testcontainers is not available
	if testing.Short() {
		suite.T().Skip("Skipping integration test")
	}

	// Create context with longer timeout for container startup
	suite.ctx, suite.cancel = context.WithTimeout(context.Background(), 3*time.Minute)
	suite.T().Logf("Container startup timeout: 3 minutes")

	// Start Cassandra container using the dedicated cassandra module
	suite.T().Logf("Starting Cassandra container...")
	cassandraContainer, err := cassandra.Run(suite.ctx, "cassandra:4.1",
		cassandra.WithInitScripts("../../schema/myjournal.cql"),
	)
	suite.Require().NoError(err)
	suite.container = cassandraContainer

	suite.T().Logf("Cassandra container started successfully")

	// Get connection details
	host, err := cassandraContainer.ConnectionHost(suite.ctx)
	suite.Require().NoError(err)

	// Log connection details for debugging
	suite.T().Logf("Connecting to Cassandra at %s", host)

	// Create config for test database
	configYAML := fmt.Sprintf(`
cassandra:
  hosts: ["%s"]
  keyspace: "myjournal"
  log_queries: false
`, host)

	configProvider, err := config.NewYAML(config.Source(strings.NewReader(configYAML)))
	suite.Require().NoError(err)

	logger, err := zap.NewDevelopment()
	suite.Require().NoError(err)

	// Create database connection
	db, err := NewDB(configProvider, logger)
	suite.Require().NoError(err)
	suite.db = db

	suite.T().Logf("Connected to myjournal keyspace successfully")
}

// TearDownSuite runs once after all tests
func (suite *JournalTestSuite) TearDownSuite() {
	if suite.db != nil && suite.db.session != nil {
		suite.db.session.Close()
	}
	if suite.container != nil {
		if err := suite.container.Terminate(suite.ctx); err != nil {
			suite.T().Logf("Failed to terminate container: %v", err)
		}
	}
	if suite.cancel != nil {
		suite.cancel()
	}
}

// SetupTest runs before each test (for cleanup)
func (suite *JournalTestSuite) SetupTest() {
	// Clean up any existing data before each test
	suite.cleanupTestData()
}

// cleanupTestData removes all test data to ensure test isolation
func (suite *JournalTestSuite) cleanupTestData() {
	// Clear posts_by_user table
	suite.db.session.Query(`TRUNCATE posts_by_user`).Exec()

	// Clear user_by_postid table
	suite.db.session.Query(`TRUNCATE user_by_postid`).Exec()
}

func (suite *JournalTestSuite) TestCreatePost() {
	err := suite.db.CreatePost(context.Background(),
		testUserUUID,
		&entities.PostData{
			Title: "test title",
			Body:  "test body",
			Tags:  []string{"Skanderborg", "school"},
		},
	)

	suite.Require().NoError(err)
}

func (suite *JournalTestSuite) TestReadPostsByUser() {
	posts, err := suite.db.ReadPostsByUser(context.Background(), testUserUUID)
	suite.Require().NoError(err)
	for _, post := range posts {
		fmt.Printf("%+v\n", post)
	}
}

func (suite *JournalTestSuite) TestDeletePost() {
	// First create a post
	err := suite.db.CreatePost(context.Background(),
		testUserUUID,
		&entities.PostData{
			Title: "test delete title",
			Body:  "test delete body",
			Tags:  []string{"test", "delete"},
		},
	)
	suite.Require().NoError(err)

	// Read posts to get the post ID
	posts, err := suite.db.ReadPostsByUser(context.Background(), testUserUUID)
	suite.Require().NoError(err)
	suite.Require().NotEmpty(posts)

	// Find the post we just created
	var postToDelete *entities.Post
	for _, post := range posts {
		if post.Title == "test delete title" {
			postToDelete = post
			break
		}
	}
	suite.Require().NotNil(postToDelete, "Should find the post we just created")

	// Delete the post
	err = suite.db.DeletePost(context.Background(), postToDelete.PostID)
	suite.Require().NoError(err)

	// Verify post is deleted
	postsAfter, err := suite.db.ReadPostsByUser(context.Background(), testUserUUID)
	suite.Require().NoError(err)

	// Check that our specific post is gone
	for _, post := range postsAfter {
		suite.Require().NotEqual("test delete title", post.Title, "Deleted post should not be found")
	}
}

func (suite *JournalTestSuite) TestDeletePostByUser() {
	// First create a post
	err := suite.db.CreatePost(context.Background(),
		testUserUUID,
		&entities.PostData{
			Title: "test delete by user title",
			Body:  "test delete by user body",
			Tags:  []string{"test", "delete", "user"},
		},
	)
	suite.Require().NoError(err)

	// Read posts to get the post ID
	posts, err := suite.db.ReadPostsByUser(context.Background(), testUserUUID)
	suite.Require().NoError(err)
	suite.Require().NotEmpty(posts)

	// Find the post we just created
	var postToDelete *entities.Post
	for _, post := range posts {
		if post.Title == "test delete by user title" {
			postToDelete = post
			break
		}
	}
	suite.Require().NotNil(postToDelete, "Should find the post we just created")

	// Delete the post with user validation
	err = suite.db.DeletePostByUser(context.Background(), postToDelete.PostID, testUserUUID)
	suite.Require().NoError(err)

	// Verify post is deleted
	postsAfter, err := suite.db.ReadPostsByUser(context.Background(), testUserUUID)
	suite.Require().NoError(err)

	// Check that our specific post is gone
	for _, post := range postsAfter {
		suite.Require().NotEqual("test delete by user title", post.Title, "Deleted post should not be found")
	}
}

func (suite *JournalTestSuite) TestAddTags() {
	// First create a post
	err := suite.db.CreatePost(context.Background(),
		testUserUUID,
		&entities.PostData{
			Title: "test add tags title",
			Body:  "test add tags body",
			Tags:  []string{"original"},
		},
	)
	suite.Require().NoError(err)

	// Read posts to get the post ID
	posts, err := suite.db.ReadPostsByUser(context.Background(), testUserUUID)
	suite.Require().NoError(err)
	suite.Require().NotEmpty(posts)

	// Find the post we just created
	var postToUpdate *entities.Post
	for _, post := range posts {
		if post.Title == "test add tags title" {
			postToUpdate = post
			break
		}
	}
	suite.Require().NotNil(postToUpdate, "Should find the post we just created")

	// Add tags
	newTags := []string{"newtag1", "newtag2"}
	err = suite.db.AddTags(context.Background(), postToUpdate.PostID, testUserUUID, newTags)
	suite.Require().NoError(err)

	// Verify tags were added by reading the post again
	postsAfter, err := suite.db.ReadPostsByUser(context.Background(), testUserUUID)
	suite.Require().NoError(err)

	var updatedPost *entities.Post
	for _, post := range postsAfter {
		if post.Title == "test add tags title" {
			updatedPost = post
			break
		}
	}
	suite.Require().NotNil(updatedPost, "Should find the updated post")

	// Check that new tags are present
	suite.Require().Contains(updatedPost.Tags, "original", "Original tag should still be present")
	suite.Require().Contains(updatedPost.Tags, "newtag1", "New tag 1 should be present")
	suite.Require().Contains(updatedPost.Tags, "newtag2", "New tag 2 should be present")
}

func (suite *JournalTestSuite) TestRemoveTags() {
	// First create a post with multiple tags
	err := suite.db.CreatePost(context.Background(),
		testUserUUID,
		&entities.PostData{
			Title: "test remove tags title",
			Body:  "test remove tags body",
			Tags:  []string{"keep1", "remove1", "keep2", "remove2"},
		},
	)
	suite.Require().NoError(err)

	// Read posts to get the post ID
	posts, err := suite.db.ReadPostsByUser(context.Background(), testUserUUID)
	suite.Require().NoError(err)
	suite.Require().NotEmpty(posts)

	// Find the post we just created
	var postToUpdate *entities.Post
	for _, post := range posts {
		if post.Title == "test remove tags title" {
			postToUpdate = post
			break
		}
	}
	suite.Require().NotNil(postToUpdate, "Should find the post we just created")

	// Remove some tags
	tagsToRemove := []string{"remove1", "remove2"}
	err = suite.db.RemoveTags(context.Background(), postToUpdate.PostID, testUserUUID, tagsToRemove)
	suite.Require().NoError(err)

	// Verify tags were removed by reading the post again
	postsAfter, err := suite.db.ReadPostsByUser(context.Background(), testUserUUID)
	suite.Require().NoError(err)

	var updatedPost *entities.Post
	for _, post := range postsAfter {
		if post.Title == "test remove tags title" {
			updatedPost = post
			break
		}
	}
	suite.Require().NotNil(updatedPost, "Should find the updated post")

	// Check that removed tags are gone and kept tags remain
	suite.Require().Contains(updatedPost.Tags, "keep1", "Keep tag 1 should still be present")
	suite.Require().Contains(updatedPost.Tags, "keep2", "Keep tag 2 should still be present")
	suite.Require().NotContains(updatedPost.Tags, "remove1", "Remove tag 1 should be gone")
	suite.Require().NotContains(updatedPost.Tags, "remove2", "Remove tag 2 should be gone")
}

// TestJournalSuite runs all tests using the test suite
func TestJournalSuite(t *testing.T) {
	suite.Run(t, new(JournalTestSuite))
}
