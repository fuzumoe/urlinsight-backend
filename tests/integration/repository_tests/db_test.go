package repository_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/fuzumoe/urlinsight-backend/internal/repository"
	"github.com/fuzumoe/urlinsight-backend/tests/utils"
)

// TestNewDB_Integration tests the NewDB function with a real MySQL database connection.
func TestNewDB_Integration(t *testing.T) {
	// fallback DSN if env var isn't set.
	testDatabase := os.Getenv("TEST_DATABASE")
	dsn := ""
	if testDatabase == "" {
		dsn = "urlinsight_user:secret@tcp(localhost:3309)/urlinsight_user?parseTime=true"
	} else {

		dsn = "urlinsight_user:secret@tcp(localhost:3309)/" + testDatabase + "?parseTime=true"
	}

	if err := utils.InitTestSuite(); err != nil {
		println("Failed to setup test suite:", err.Error())
		os.Exit(1)
	}

	t.Run("NewDB", func(t *testing.T) {

		db, err := repository.NewDB(dsn)
		require.NoError(t, err, "NewDB should not return an error")
		require.NotNil(t, db, "db should not be nil")

		sqlDB, err := db.DB()
		require.NoError(t, err, "Should be able to retrieve sql.DB")
		require.NotNil(t, sqlDB, "sql.DB should not be nil")

		err = sqlDB.Ping()
		require.NoError(t, err, "Should be able to ping DB")

		stats := sqlDB.Stats()
		assert.Greater(t, stats.OpenConnections, 0, "Should have at least one open connection")
	})

	utils.CleanTestData(t)

}
