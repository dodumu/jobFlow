package handlers

import (
	"database/sql"
	"errors"
	"fmt"
	"jobFlow/database"
	"jobFlow/models"
	"jobFlow/utils"
	"log"
	"net/http"
	"strconv"
)

func FollowHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w, r)
		return
	}

	targetID, err := strconv.Atoi(r.PathValue("id"))

	if err != nil || targetID <= 0 {
		RenderError(w, r, http.StatusBadRequest, "Invalid Profile", "The requested profile ID is invalid")
		return
	}

	viewerID, ok := utils.GetUserID(r)
	if !ok {
		unauthorized(w, r, "You must be logged in to followe someone")
		return
	}

	if viewerID == targetID {
		RenderError(w, r, http.StatusBadRequest, "Invalid Action", "You cannot follow yourself.")
		return
	}

	_, err = database.GetUserByID(targetID)
	if errors.Is(err, sql.ErrNoRows) {
		notFound(w, r, "Profile Not Found", "The requested profile does not exitst")
		return
	}

	if err != nil {
		log.Printf("FollowHandler GetUserByID error := %v", err)
		serverError(w, r, "We couldn't find the requested profile.")
		return
	}

	isFollowing, err := database.IsFollowing(viewerID, targetID)
	if err != nil {
		log.Printf("FollowHandler IsFollowing error: %v", err)
		serverError(w, r, "We couldn't check the follow relationship.")
		return
	}
	if !isFollowing {
		follow := models.Follow{
			FollowerID:  viewerID,
			FollowingID: targetID,
		}

		if err := database.CreateFollow(follow); err != nil {
			log.Printf("FollowHandler CreateFollow error: %v", err)
			serverError(w, r, "We couldn't follow this profile")
			return
		}

		// Create a notification for the user being followed.
		_, err := database.CreateNotification(
			int64(targetID),
			sql.NullInt64{
				Int64: int64(viewerID),
				Valid: true,
			},
			"follow",
			"Someone started following you.",
			fmt.Sprintf("/users/%d", viewerID),
		)

		if err != nil {
			log.Printf("FollowHandler CreateNotification error: %v", err)
		}
	}

	http.Redirect(w, r, fmt.Sprintf("/users/%d", targetID), http.StatusSeeOther)
}

func UnfollowHandler(w http.ResponseWriter, r *http.Request) {
	// Only allow POST requests.
	if r.Method != http.MethodPost {
		methodNotAllowed(w, r)
		return
	}

	// Get the ID of the profile being unfollowed.
	targetID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || targetID <= 0 {
		RenderError(w, r, http.StatusBadRequest,
			"Invalid Profile",
			"The requested profile ID is invalid.")
		return
	}

	// Get the logged-in user's ID.
	viewerID, ok := utils.GetUserID(r)
	if !ok {
		unauthorized(w, r, "You must be logged in to unfollow someone.")
		return
	}

	// Prevent users from unfollowing themselves.
	if viewerID == targetID {
		RenderError(w, r, http.StatusBadRequest,
			"Invalid Action",
			"You cannot unfollow yourself.")
		return
	}

	// Confirm that the target user exists.
	_, err = database.GetUserByID(targetID)
	if errors.Is(err, sql.ErrNoRows) {
		notFound(w, r, "Profile Not Found", "The requested profile does not exist.")
		return
	}

	if err != nil {
		log.Printf("UnfollowHandler GetUserByID error: %v", err)
		serverError(w, r, "We couldn't find the requested profile.")
		return
	}

	// Check whether the viewer follows this user.
	isFollowing, err := database.IsFollowing(viewerID, targetID)
	if err != nil {
		log.Printf("UnfollowHandler IsFollowing error: %v", err)
		serverError(w, r, "We couldn't check the follow relationship.")
		return
	}

	// Only delete an existing relationship.
	if isFollowing {
		if err := database.DeleteFollow(viewerID, targetID); err != nil {
			log.Printf("UnfollowHandler DeleteFollow error: %v", err)
			serverError(w, r, "We couldn't unfollow this profile.")
			return
		}
	}

	// Return to the profile.
	http.Redirect(w, r, fmt.Sprintf("/users/%d", targetID),
		http.StatusSeeOther)
}

func FollowersHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w, r)
		return
	}

	userID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || userID <= 0 {
		RenderError(w, r, http.StatusBadRequest, "Bad Request", "Invalid user ID")
		return
	}

	user, err := database.GetUserByID(userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			notFound(w, r, "Not Found", "User could not be found")
			return
		}

		log.Printf("FollowersHandler GetUserByID error: %v", err)
		serverError(w, r, "Could not retrieve user")
		return
	}

	followers, err := database.GetFollowers(userID)
	if err != nil {
		log.Printf("FollowersHandler GetFollowers error: %v", err)
		serverError(w, r, "Could not retrieve followers")
		return
	}

	followerData := models.FollowListPageData{
		ProfileUser: user,
		Users:       followers,
		Title:       "Followers",
	}

	err = utils.RenderTemplate(w, r, "followers.html", followerData)

	if err != nil {
		serverError(w, r, "Could not render followers page")
		return
	}
}

func FollowingHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w, r)
		return
	}

	userID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || userID <= 0 {
		RenderError(w, r, http.StatusBadRequest, "Bad Request", "Invalid user ID")
		return
	}

	user, err := database.GetUserByID(userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			notFound(w, r, "Not Found", "User could not be found")
			return
		}

		log.Printf("FollowingHandler GetUserByID error: %v", err)
		serverError(w, r, "Could not retrieve user")
		return
	}

	following, err := database.GetFollowing(userID)
	if err != nil {
		log.Printf("FollowingHandler GetFollowing error: %v", err)
		serverError(w, r, "Could not retrieve following list")
		return
	}

	followingData := models.FollowListPageData{
		ProfileUser: user,
		Users:       following,
		Title:       "Following",
	}

	err = utils.RenderTemplate(w, r, "following.html", followingData)
	if err != nil {
		serverError(w, r, "Could not render following page")
		return
	}
}
