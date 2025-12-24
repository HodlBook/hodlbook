package database

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDatabase_NewAndClose(t *testing.T) {
	db, err := New(WithPath("file::memory:?cache=shared"))
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
