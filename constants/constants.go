package constants

import "time"

const (
	// Session
	SessionCookieName = "session_token"

	// User roles
	RoleIndividual = "individual"
	RoleCompany    = "company"

	// Job statuses
	JobStatusOpen    = "open"
	JobStatusClosed  = "closed"
	JobStatusPending = "pending"

	ApplicationStatusPending  = "pending"
	ApplicationStatusAccepted = "accepted"
	ApplicationStatusRejected = "rejected"

	MaxLoginAttempts      = 5
	LoginAttemptWindow    = 15 * time.Minute
	LoginBlockDuration    = 15 * time.Minute
	MaxProfilePictureSize = 5 << 20
)
