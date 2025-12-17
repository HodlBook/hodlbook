package database

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDatabase_NewAndClose(t *testing.T) {
	dbPath := filepath.Join(os.TempDir(), "test_hodlbook.db")
	defer os.Remove(dbPath)

	db, err := New(WithPath(dbPath))
	require.NoError(t, err)
	require.NotNil(t, db.Get())
	require.NoError(t, db.Close())
}

func TestDatabase_DefaultPath(t *testing.T) {
	db, err := New(WithPath(""))
	require.NoError(t, err)
	require.NotNil(t, db.Get())
	require.NoError(t, db.Close())
}
