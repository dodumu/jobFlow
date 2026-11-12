package handlers

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"jobFlow/constants"
)

// reviewFixture creates two companies, an applicant, a job owned by "acme"
// and one pending application to that job.
type reviewFixture struct {
	acmeID        int // company that owns the job
	rivalID       int // a different company
	applicantID   int
	otherPersonID int // an individual who did not apply
	applicationID int
}

func newReviewFixture(t *testing.T) reviewFixture {
	t.Helper()
	setupTestDB(t)

	f := reviewFixture{
		acmeID:        registerUser(t, "acme", constants.RoleCompany),
		rivalID:       registerUser(t, "rival", constants.RoleCompany),
		applicantID:   registerUser(t, "applicant", constants.RoleIndividual),
		otherPersonID: registerUser(t, "bystander", constants.RoleIndividual),
	}

	jobID := createJob(t, f.acmeID, constants.JobStatusOpen, time.Now().UTC().AddDate(0, 0, 7))
	f.applicationID = createApplication(t, jobID, f.applicantID)

	return f
}

func TestReviewApplicationPermissions(t *testing.T) {
	tests := []struct {
		name       string
		handler    http.HandlerFunc
		actor      func(f reviewFixture) int
		wantCode   int
		wantStatus string // application status afterwards
	}{
		{"owner accepts", AcceptApplicationHandler, func(f reviewFixture) int { return f.acmeID }, http.StatusSeeOther, constants.ApplicationStatusAccepted},
		{"owner rejects", RejectApplicationHandler, func(f reviewFixture) int { return f.acmeID }, http.StatusSeeOther, constants.ApplicationStatusRejected},
		{"other company cannot accept", AcceptApplicationHandler, func(f reviewFixture) int { return f.rivalID }, http.StatusForbidden, constants.ApplicationStatusPending},
		{"other company cannot reject", RejectApplicationHandler, func(f reviewFixture) int { return f.rivalID }, http.StatusForbidden, constants.ApplicationStatusPending},
		{"applicant cannot accept own application", AcceptApplicationHandler, func(f reviewFixture) int { return f.applicantID }, http.StatusForbidden, constants.ApplicationStatusPending},
		{"not logged in", AcceptApplicationHandler, func(f reviewFixture) int { return 0 }, http.StatusUnauthorized, constants.ApplicationStatusPending},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			f := newReviewFixture(t)

			target := fmt.Sprintf("/applications/%d/accept", f.applicationID)
			rec := requestWithID(tc.handler, http.MethodPost, target, tc.actor(f), f.applicationID)

			if rec.Code != tc.wantCode {
				t.Errorf("status = %d, want %d", rec.Code, tc.wantCode)
			}
			if got := applicationStatus(t, f.applicationID); got != tc.wantStatus {
				t.Errorf("application status = %q, want %q", got, tc.wantStatus)
			}
		})
	}
}

func TestApplicationCannotBeReviewedTwice(t *testing.T) {
	f := newReviewFixture(t)
	target := fmt.Sprintf("/applications/%d/accept", f.applicationID)

	first := requestWithID(AcceptApplicationHandler, http.MethodPost, target, f.acmeID, f.applicationID)
	if first.Code != http.StatusSeeOther {
		t.Fatalf("first review: status = %d, want %d", first.Code, http.StatusSeeOther)
	}

	second := requestWithID(RejectApplicationHandler, http.MethodPost, target, f.acmeID, f.applicationID)
	if second.Code != http.StatusBadRequest {
		t.Errorf("second review: status = %d, want %d", second.Code, http.StatusBadRequest)
	}
	if got := applicationStatus(t, f.applicationID); got != constants.ApplicationStatusAccepted {
		t.Errorf("status = %q, want the first decision to stand (%q)", got, constants.ApplicationStatusAccepted)
	}
}

func TestViewApplicationAccess(t *testing.T) {
	tests := []struct {
		name     string
		handler  http.HandlerFunc
		target   string
		actor    func(f reviewFixture) int
		wantCode int
	}{
		{"company owner views applicant", ViewApplicationHandler, "/applications/%d", func(f reviewFixture) int { return f.acmeID }, http.StatusOK},
		{"other company is refused", ViewApplicationHandler, "/applications/%d", func(f reviewFixture) int { return f.rivalID }, http.StatusForbidden},
		{"individual cannot use company view", ViewApplicationHandler, "/applications/%d", func(f reviewFixture) int { return f.applicantID }, http.StatusForbidden},
		{"applicant views own application", ViewMyApplicationHandler, "/my-applications/%d", func(f reviewFixture) int { return f.applicantID }, http.StatusOK},
		{"other individual is refused", ViewMyApplicationHandler, "/my-applications/%d", func(f reviewFixture) int { return f.otherPersonID }, http.StatusForbidden},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			f := newReviewFixture(t)

			target := fmt.Sprintf(tc.target, f.applicationID)
			rec := requestWithID(tc.handler, http.MethodGet, target, tc.actor(f), f.applicationID)

			if rec.Code != tc.wantCode {
				t.Errorf("status = %d, want %d", rec.Code, tc.wantCode)
			}
		})
	}
}
