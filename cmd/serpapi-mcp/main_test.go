package main

import (
	"os"
	"testing"
)

// TestEnvOrSetEnv verifies that envOr returns the environment value when set.
func TestEnvOrSetEnv(t *testing.T) {
	const key = "TEST_GSD_ENV_OR"
	os.Setenv(key, "fromenv")
	defer os.Unsetenv(key)

	if got := envOr(key, "default"); got != "fromenv" {
		t.Errorf("envOr(%q, \"default\") = %q, want \"fromenv\"", key, got)
	}
}

// TestEnvOrUnsetEnv verifies that envOr returns the fallback when the variable is not set.
func TestEnvOrUnsetEnv(t *testing.T) {
	const key = "TEST_GSD_ENV_OR_NONE"
	os.Unsetenv(key)

	if got := envOr(key, "fallback"); got != "fallback" {
		t.Errorf("envOr(%q, \"fallback\") = %q, want \"fallback\"", key, got)
	}
}

// TestEnvOrEmptyEnv verifies that envOr treats an empty string as unset and returns fallback.
func TestEnvOrEmptyEnv(t *testing.T) {
	const key = "TEST_GSD_ENV_OR_EMPTY"
	os.Setenv(key, "")
	defer os.Unsetenv(key)

	if got := envOr(key, "fallback"); got != "fallback" {
		t.Errorf("envOr(%q, \"fallback\") = %q, want \"fallback\" (empty treated as unset)", key, got)
	}
}

// TestEnvIntOrValid verifies that envIntOr returns the integer value when valid.
func TestEnvIntOrValid(t *testing.T) {
	const key = "TEST_GSD_ENV_INT"
	os.Setenv(key, "9090")
	defer os.Unsetenv(key)

	if got := envIntOr(key, 8000); got != 9090 {
		t.Errorf("envIntOr(%q, 8000) = %d, want 9090", key, got)
	}
}

// TestEnvIntOrInvalid verifies that envIntOr returns fallback for non-integer values.
func TestEnvIntOrInvalid(t *testing.T) {
	const key = "TEST_GSD_ENV_INT_BAD"
	os.Setenv(key, "notanumber")
	defer os.Unsetenv(key)

	if got := envIntOr(key, 8000); got != 8000 {
		t.Errorf("envIntOr(%q, 8000) = %d, want 8000 (fallback for invalid int)", key, got)
	}
}

// TestEnvIntOrUnset verifies that envIntOr returns fallback when variable is not set.
func TestEnvIntOrUnset(t *testing.T) {
	const key = "TEST_GSD_ENV_INT_NONE_XYZ"
	os.Unsetenv(key)

	if got := envIntOr(key, 3000); got != 3000 {
		t.Errorf("envIntOr(%q, 3000) = %d, want 3000", key, got)
	}
}

// TestEnvBoolOrTruthy verifies that envBoolOr accepts "1", "true", "yes" (case-insensitive).
func TestEnvBoolOrTruthy(t *testing.T) {
	const key = "TEST_GSD_ENV_BOOL_TRUTHY"
	truthyValues := []string{"1", "true", "True", "TRUE", "yes", "Yes", "YES"}

	for _, val := range truthyValues {
		t.Run(val, func(t *testing.T) {
			os.Setenv(key, val)
			defer os.Unsetenv(key)

			if got := envBoolOr(key, false); !got {
				t.Errorf("envBoolOr(%q, false) with value %q = false, want true", key, val)
			}
		})
	}
}

// TestEnvBoolOrFalsy verifies that "0", "false", "no" return false (not the fallback).
// The envBoolOr implementation returns false for any value that isn't "1"/"true"/"yes".
func TestEnvBoolOrFalsy(t *testing.T) {
	const key = "TEST_GSD_ENV_BOOL_FALSY"
	falsyValues := []string{"0", "false", "False", "no", "No"}

	for _, val := range falsyValues {
		t.Run(val, func(t *testing.T) {
			os.Setenv(key, val)
			defer os.Unsetenv(key)

			// envBoolOr returns false for non-truthy values, regardless of fallback
			if got := envBoolOr(key, true); got {
				t.Errorf("envBoolOr(%q, true) with value %q = true, want false", key, val)
			}
		})
	}
}

// TestEnvBoolOrUnset verifies that envBoolOr returns the fallback when variable is not set.
func TestEnvBoolOrUnset(t *testing.T) {
	const key = "TEST_GSD_ENV_BOOL_NONE_XYZ"
	os.Unsetenv(key)

	if got := envBoolOr(key, true); !got {
		t.Errorf("envBoolOr(%s, true) = false, want true (fallback)", key)
	}

	if got := envBoolOr(key, false); got {
		t.Errorf("envBoolOr(%s, false) = true, want false (fallback)", key)
	}
}
