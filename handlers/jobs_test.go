package handlers

import (
	"fmt"
	"jobFlow/constants"
	"jobFlow/database"
	"net/http"
	"net/url"
	"testing"
	"time"
)

func TestApplyJobHandler(t *testing.T) {
	now := time.Now().UTC()
	today := now.Truncate(24 * time.Hour)

	tests := []struct {
		name        string
		jobStatus   string
		deadline    time.Time
		actor       string // "individual", "company" or "nobody"
		coverLetter string
		wantStatus  int
		wantApplied bool
	}{
		{"open job, future deadline", constants.JobStatusOpen, now.AddDate(0, 0, 7), "individual", "I am a good fit.", http.StatusSeeOther, true},
		{"deadline is today, still open", constants.JobStatusOpen, today, "individual", "I am a good fit.", http.StatusSeeOther, true},
		{"deadline passed", constants.JobStatusOpen, now.AddDate(0, 0, -3), "individual", "I am a good fit.", http.StatusBadRequest, false},
		{"job closed by company", constants.JobStatusClosed, now.AddDate(0, 0, 7), "individual", "I am a good fit.", http.StatusBadRequest, false},
		{"company account cannot apply", constants.JobStatusOpen, now.AddDate(0, 0, 7), "company", "I am a good fit.", http.StatusForbidden, false},
		{"not logged in", constants.JobStatusOpen, now.AddDate(0, 0, 7), "nobody", "I am a good fit.", http.StatusUnauthorized, false},
		{"empty cover letter", constants.JobStatusOpen, now.AddDate(0, 0, 7), "individual", "   ", http.StatusBadRequest, false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			setupTestDB(t)

			companyID := registerUser(t, "acme", constants.RoleCompany)
			applicantID := registerUser(t, "applicant", constants.RoleIndividual)
			jobID := createJob(t, companyID, tc.jobStatus, tc.deadline)

			actorID := 0
			switch tc.actor {
			case "individual":
				actorID = applicantID
			case "company":
				actorID = companyID
			}

			target := fmt.Sprintf("/jobs/%d/apply", jobID)
			rec := requestAs(ApplyJobHandler, http.MethodPost, target, actorID, url.Values{
				"cover_letter": {tc.coverLetter},
			})

			if rec.Code != tc.wantStatus {
				t.Errorf("status = %d, want %d", rec.Code, tc.wantStatus)
			}

			applied, err := database.HasUserApplied(jobID, applicantID)
			if err != nil {
				t.Fatalf("checking application: %v", err)
			}
			if applied != tc.wantApplied {
				t.Errorf("application stored = %v, want %v", applied, tc.wantApplied)
			}
		})
	}
}

func TestApplyJobTwiceIsRejected(t *testing.T) {
	setupTestDB(t)

	companyID := registerUser(t, "acme", constants.RoleCompany)
	applicantID := registerUser(t, "applicant", constants.RoleIndividual)
	jobID := createJob(t, companyID, constants.JobStatusOpen, time.Now().UTC().AddDate(0, 0, 7))

	target := fmt.Sprintf("/jobs/%d/apply", jobID)
	form := url.Values{"cover_letter": {"First attempt."}}

	first := requestAs(ApplyJobHandler, http.MethodPost, target, applicantID, form)
	if first.Code != http.StatusSeeOther {
		t.Fatalf("first application: status = %d, want %d", first.Code, http.StatusSeeOther)
	}

	second := requestAs(ApplyJobHandler, http.MethodPost, target, applicantID, form)
	if second.Code != http.StatusBadRequest {
		t.Errorf("second application: status = %d, want %d", second.Code, http.StatusBadRequest)
	}
}
