package utils

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetEnv_Existing(t *testing.T) {
	os.Setenv("FOO_BAR", "qux")
	val := GetEnv("FOO_BAR", "baz")
	require.Equal(t, "qux", val)
	os.Unsetenv("FOO_BAR")
}
func TestGetEnv_Default(t *testing.T) {
	os.Unsetenv("FOO_BAR")
	val := GetEnv("FOO_BAR", "baz")
	require.Equal(t, "baz", val)
}
