package handlers

import (
	"log"
	"os"
	"testing"

	"jobFlow/utils"
)

// TestMain runs once before all tests in this package.
func TestMain(m *testing.M) {
	// Tests run inside handlers/, but the templates are found relative to
	// the project root, so step up one folder first.
	if err := os.Chdir(".."); err != nil {
		log.Fatalf("changing to project root: %v", err)
	}

	if err := utils.LoadTemplates(); err != nil {
		log.Fatalf("loading templates: %v", err)
	}

	os.Exit(m.Run())
}