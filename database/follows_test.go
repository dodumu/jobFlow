package database_test

import (
	"path/filepath"
	"testing"

	"jobFlow/database"
	"jobFlow/models"
)

func createFollowTestUser(t *testing.T, username string) int {
	t.Helper()

	result, err := database.DB.Exec(`
		INSERT INTO users (
			first_name,
			last_name,
			username,
			password_hash,
			email,
			date_of_birth,
			role
		)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`,
		"Test",
		"User",
		username,
		"fake-test-hash",
		username+"@example.com",
		"2000-01-01",
		"individual",
	)

	if err != nil {
		t.Fatalf("failed to create test user: %v", err)
	}

	userID, err := result.LastInsertId()
	if err != nil {
		t.Fatalf("failed to get test user ID: %v", err)
	}

	return int(userID)
}
func TestCreateFollow(t *testing.T) {
	// Initialize a temporary database.
	dbPath := filepath.Join(t.TempDir(), "test.db")

	if err := database.InitDB(dbPath); err != nil {
		t.Fatalf("failed to initialize test database: %v", err)
	}

	// Close the database after the test.
	t.Cleanup(func() {
		if database.DB != nil {
			database.DB.Close()
			database.DB = nil
		}
	})

	// Create two test users.
	followerID := createFollowTestUser(t, "follower")
	followingID := createFollowTestUser(t, "following")

	// Create the follow relationship.
	follow := models.Follow{
		FollowerID:  followerID,
		FollowingID: followingID,
	}

	err := database.CreateFollow(follow)
	if err != nil {
		t.Fatalf("failed to create follow: %v", err)
	}

	// Verify the relationship exists.
	isFollowing, err := database.IsFollowing(followerID, followingID)
	if err != nil {
		t.Fatalf("failed to check follow relationship: %v", err)
	}

	if !isFollowing {
		t.Fatal("expected follower to be following the target user")
	}
}

func TestCreateDuplicateFollow(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")

	if err := database.InitDB(dbPath); err != nil {
		t.Fatalf("failed to initialize test databse: %v", err)
	}

	t.Cleanup(func() {
		if database.DB != nil {
			database.DB.Close()
			database.DB = nil
		}
	})

	followerID := createFollowTestUser(t, "follower")
	followingID := createFollowTestUser(t, "following")

	follow := models.Follow{
		FollowerID:  followerID,
		FollowingID: followingID,
	}

	if err := database.CreateFollow(follow); err != nil {
		t.Fatalf("first follow should succeed: %v", err)
	}

	err := database.CreateFollow(follow)

	if err == nil {
		t.Fatalf("expected duplicate follow to fail, but it succeeded")
	}
}

func TestFollowSelf(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")

	if err := database.InitDB(dbPath); err != nil {
		t.Fatalf("failed to initialize test database: %v", err)
	}

	t.Cleanup(func() {
		if database.DB != nil {
			database.DB.Close()
			database.DB = nil
		}
	})

	userID := createFollowTestUser(t, "selfuser")

	follow := models.Follow{
		FollowerID:  userID,
		FollowingID: userID,
	}

	err := database.CreateFollow(follow)
	if err == nil {
		t.Fatal("expected self-follow to fail, but it succeeded")
	}
}

func TestDeleteFollow(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")

	if err := database.InitDB(dbPath); err != nil {
		t.Fatalf("failed to initialize test database: %v", err)
	}

	t.Cleanup(func() {
		if database.DB != nil {
			database.DB.Close()
			database.DB = nil
		}
	})

	followerID := createFollowTestUser(t, "follower")
	followingID := createFollowTestUser(t, "following")

	follow := models.Follow{
		FollowerID:  followerID,
		FollowingID: followingID,
	}

	// Create the relationship first.
	if err := database.CreateFollow(follow); err != nil {
		t.Fatalf("failed to create follow: %v", err)
	}

	// Delete the relationship.
	if err := database.DeleteFollow(followerID, followingID); err != nil {
		t.Fatalf("failed to delete follow: %v", err)
	}

	// Confirm the relationship no longer exists.
	isFollowing, err := database.IsFollowing(followerID, followingID)
	if err != nil {
		t.Fatalf("failed to check follow relationship: %v", err)
	}

	if isFollowing {
		t.Fatal("expected follow relationship to be deleted")
	}
}

func TestFollowCounts(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")

	if err := database.InitDB(dbPath); err != nil {
		t.Fatalf("failed to initialize test database: %v", err)
	}

	t.Cleanup(func() {
		if database.DB != nil {
			database.DB.Close()
			database.DB = nil
		}
	})

	user1 := createFollowTestUser(t, "userone")
	user2 := createFollowTestUser(t, "usertwo")

	// Check initial counts.
	followers, err := database.GetFollowerCount(user2)
	if err != nil {
		t.Fatalf("failed to get follower count: %v", err)
	}

	if followers != 0 {
		t.Fatalf("expected 0 followers, got %d", followers)
	}

	following, err := database.GetFollowingCount(user1)
	if err != nil {
		t.Fatalf("failed to get following count: %v", err)
	}

	if following != 0 {
		t.Fatalf("expected 0 following, got %d", following)
	}

	// Create follow relationship.
	if err := database.CreateFollow(models.Follow{
		FollowerID:  user1,
		FollowingID: user2,
	}); err != nil {
		t.Fatalf("failed to create follow: %v", err)
	}

	// Verify updated counts.
	followers, err = database.GetFollowerCount(user2)
	if err != nil {
		t.Fatalf("failed to get updated follower count: %v", err)
	}

	if followers != 1 {
		t.Fatalf("expected 1 follower, got %d", followers)
	}

	following, err = database.GetFollowingCount(user1)
	if err != nil {
		t.Fatalf("failed to get updated following count: %v", err)
	}

	if following != 1 {
		t.Fatalf("expected 1 following, got %d", following)
	}

	// Unfollow and verify counts return to zero.
	if err := database.DeleteFollow(user1, user2); err != nil {
		t.Fatalf("failed to delete follow: %v", err)
	}

	followers, err = database.GetFollowerCount(user2)
	if err != nil {
		t.Fatalf("failed to get follower count after unfollow: %v", err)
	}

	if followers != 0 {
		t.Fatalf("expected 0 followers after unfollow, got %d", followers)
	}

	following, err = database.GetFollowingCount(user1)
	if err != nil {
		t.Fatalf("failed to get following count after unfollow: %v", err)
	}

	if following != 0 {
		t.Fatalf("expected 0 following after unfollow, got %d", following)
	}
}

func TestGetFollowersAndFollowing(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")

	if err := database.InitDB(dbPath); err != nil {
		t.Fatalf("failed to initialize test database: %v", err)
	}

	t.Cleanup(func() {
		if database.DB != nil {
			database.DB.Close()
			database.DB = nil
		}
	})

	// Create three users.
	user1 := createFollowTestUser(t, "userone")
	user2 := createFollowTestUser(t, "usertwo")
	user3 := createFollowTestUser(t, "userthree")

	// user1 and user3 follow user2.
	follows := []models.Follow{
		{FollowerID: user1, FollowingID: user2},
		{FollowerID: user3, FollowingID: user2},
	}

	for _, follow := range follows {
		if err := database.CreateFollow(follow); err != nil {
			t.Fatalf("failed to create follow: %v", err)
		}
	}

	// Get users following user2.
	followers, err := database.GetFollowers(user2)
	if err != nil {
		t.Fatalf("failed to get followers: %v", err)
	}

	if len(followers) != 2 {
		t.Fatalf("expected 2 followers, got %d", len(followers))
	}

	// Get accounts user1 is following.
	following, err := database.GetFollowing(user1)
	if err != nil {
		t.Fatalf("failed to get following: %v", err)
	}

	if len(following) != 1 {
		t.Fatalf("expected user1 to follow 1 account, got %d", len(following))
	}

	if following[0].ID != user2 {
		t.Fatalf("expected user1 to follow user2 (ID %d), got ID %d",
			user2, following[0].ID)
	}
}
