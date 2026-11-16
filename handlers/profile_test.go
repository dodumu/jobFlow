package handlers

import (
	"fmt"
	"net/http"
	"strings"
	"testing"

	"jobFlow/constants"
)

func TestOwnProfileShowsEditButtonNotFollow(t *testing.T) {
	setupTestDB(t)
	userID := registerUser(t, "ada", constants.RoleIndividual)

	rec := requestAs(ProfileHandler, http.MethodGet, "/profile", userID, nil)
	body := rec.Body.String()

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if !strings.Contains(body, "Edit Profile") {
		t.Error("your own profile should show the Edit Profile button")
	}
	if strings.Contains(body, fmt.Sprintf(`action="/users/%d/follow"`, userID)) {
		t.Error("your own profile must not offer to follow yourself")
	}
}

func TestOtherProfileShowsFollowNotEdit(t *testing.T) {
	setupTestDB(t)
	viewerID := registerUser(t, "ada", constants.RoleIndividual)
	targetID := registerUser(t, "grace", constants.RoleIndividual)

	target := fmt.Sprintf("/users/%d", targetID)
	rec := requestWithID(PublicProfileHandler, http.MethodGet, target, viewerID, targetID)
	body := rec.Body.String()

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if !strings.Contains(body, fmt.Sprintf(`action="/users/%d/follow"`, targetID)) {
		t.Error("someone else's profile should offer a Follow button")
	}
	if strings.Contains(body, "Edit Profile") {
		t.Error("someone else's profile must not show Edit Profile")
	}
}

func TestOwnProfileViaPublicURLShowsEdit(t *testing.T) {
	setupTestDB(t)
	userID := registerUser(t, "ada", constants.RoleIndividual)

	target := fmt.Sprintf("/users/%d", userID)
	rec := requestWithID(PublicProfileHandler, http.MethodGet, target, userID, userID)

	if !strings.Contains(rec.Body.String(), "Edit Profile") {
		t.Error("viewing your own public URL should still show Edit Profile")
	}
}