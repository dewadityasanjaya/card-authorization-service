package service_test

import (
	"os"
	"testing"

	"github.com/dewadityasanjaya/card-authorization-service/pkg/logger"
)

// TestMain runs before all tests in this package
func TestMain(m *testing.M) {
	// Initialize logger so it's not nil during tests
	logger.Init("development")
	defer logger.Sync()

	// Run all tests
	os.Exit(m.Run())
}
