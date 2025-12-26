package provider

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/gonzolino/gotado/v2"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"golang.org/x/oauth2"
)

func TestToTypesStr(t *testing.T) {
	cases := []struct {
		s        string
		expected types.String
	}{
		{
			s:        "test",
			expected: types.StringValue("test"),
		},
	}

	for _, c := range cases {
		actual := toTypesString(&c.s)
		if !actual.Equal(c.expected) {
			t.Fatalf("Expected: %#v, got: %#v", c.expected, actual)
		}
	}
}

func TestBoolToPower(t *testing.T) {
	if boolToPower(true) != gotado.PowerOn {
		t.Fatalf("Expected: %s, got: %s", gotado.PowerOn, boolToPower(true))
	}

	if boolToPower(false) != gotado.PowerOff {
		t.Fatalf("Expected: %s, got: %s", gotado.PowerOff, boolToPower(false))
	}
}

func TestUpdateToken(t *testing.T) {
	// Create a temporary directory for test files
	tmpDir := t.TempDir()

	t.Run("successfully writes token to file", func(t *testing.T) {
		tokenPath := filepath.Join(tmpDir, "test_token.json")
		testToken := &oauth2.Token{
			AccessToken:  "test-access-token",
			TokenType:    "Bearer",
			RefreshToken: "test-refresh-token",
			Expiry:       time.Now().Add(24 * time.Hour),
		}

		err := updateToken(testToken, tokenPath)
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		// Verify file exists and has correct permissions
		fileInfo, err := os.Stat(tokenPath)
		if err != nil {
			t.Fatalf("Expected file to exist, got error: %v", err)
		}

		// Check file permissions (0600 on Unix-like systems)
		// 0177 is a mask to check group/other permissions are not set (only owner read/write allowed)
		if fileInfo.Mode().Perm()&0177 != 0 {
			t.Errorf("File permissions too permissive: %v", fileInfo.Mode().Perm())
		}

		// Verify token content
		content, err := os.ReadFile(tokenPath)
		if err != nil {
			t.Fatalf("Failed to read token file: %v", err)
		}

		var readToken oauth2.Token
		if err := json.Unmarshal(content, &readToken); err != nil {
			t.Fatalf("Failed to unmarshal token: %v", err)
		}

		if readToken.AccessToken != testToken.AccessToken {
			t.Errorf("Expected access token %s, got %s", testToken.AccessToken, readToken.AccessToken)
		}
		if readToken.RefreshToken != testToken.RefreshToken {
			t.Errorf("Expected refresh token %s, got %s", testToken.RefreshToken, readToken.RefreshToken)
		}
	})

	t.Run("overwrites existing file", func(t *testing.T) {
		tokenPath := filepath.Join(tmpDir, "overwrite_token.json")

		// Write first token
		firstToken := &oauth2.Token{AccessToken: "first-token"}
		err := updateToken(firstToken, tokenPath)
		if err != nil {
			t.Fatalf("Failed to write first token: %v", err)
		}

		// Write second token
		secondToken := &oauth2.Token{AccessToken: "second-token"}
		err = updateToken(secondToken, tokenPath)
		if err != nil {
			t.Fatalf("Failed to write second token: %v", err)
		}

		// Verify only second token is present
		content, err := os.ReadFile(tokenPath)
		if err != nil {
			t.Fatalf("Failed to read token file: %v", err)
		}

		var readToken oauth2.Token
		if err := json.Unmarshal(content, &readToken); err != nil {
			t.Fatalf("Failed to unmarshal token: %v", err)
		}

		if readToken.AccessToken != "second-token" {
			t.Errorf("Expected access token 'second-token', got %s", readToken.AccessToken)
		}
	})

	t.Run("fails with invalid path", func(t *testing.T) {
		// Use a path that cannot be created (subdirectory doesn't exist)
		invalidPath := filepath.Join(tmpDir, "nonexistent", "subdir", "token.json")
		testToken := &oauth2.Token{AccessToken: "test-token"}

		err := updateToken(testToken, invalidPath)
		if err == nil {
			t.Error("Expected error for invalid path, got nil")
		}
	})
}

func TestReadToken(t *testing.T) {
	tmpDir := t.TempDir()

	t.Run("successfully reads token from file", func(t *testing.T) {
		tokenPath := filepath.Join(tmpDir, "read_token.json")
		expectedToken := &oauth2.Token{
			AccessToken:  "test-access-token",
			TokenType:    "Bearer",
			RefreshToken: "test-refresh-token",
			Expiry:       time.Now().Add(24 * time.Hour),
		}

		// Write token to file
		tokenBytes, err := json.MarshalIndent(expectedToken, "", "  ")
		if err != nil {
			t.Fatalf("Failed to marshal test token: %v", err)
		}
		err = os.WriteFile(tokenPath, tokenBytes, 0600)
		if err != nil {
			t.Fatalf("Failed to write test token file: %v", err)
		}

		// Read token
		token, err := readToken(tokenPath)
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		if token == nil {
			t.Fatal("Expected token, got nil")
		}

		if token.AccessToken != expectedToken.AccessToken {
			t.Errorf("Expected access token %s, got %s", expectedToken.AccessToken, token.AccessToken)
		}
		if token.RefreshToken != expectedToken.RefreshToken {
			t.Errorf("Expected refresh token %s, got %s", expectedToken.RefreshToken, token.RefreshToken)
		}
	})

	t.Run("returns nil when file does not exist", func(t *testing.T) {
		nonexistentPath := filepath.Join(tmpDir, "nonexistent.json")

		token, err := readToken(nonexistentPath)
		if err != nil {
			t.Errorf("Expected no error for nonexistent file, got: %v", err)
		}
		if token != nil {
			t.Errorf("Expected nil token for nonexistent file, got: %v", token)
		}
	})

	t.Run("fails with invalid JSON", func(t *testing.T) {
		tokenPath := filepath.Join(tmpDir, "invalid_token.json")
		invalidJSON := []byte(`{"access_token": "test", invalid json}`)
		err := os.WriteFile(tokenPath, invalidJSON, 0600)
		if err != nil {
			t.Fatalf("Failed to write invalid JSON file: %v", err)
		}

		token, err := readToken(tokenPath)
		if err == nil {
			t.Error("Expected error for invalid JSON, got nil")
		}
		if token != nil {
			t.Errorf("Expected nil token for invalid JSON, got: %v", token)
		}
	})

	t.Run("fails with permission denied", func(t *testing.T) {
		// This test only works on Unix-like systems
		if runtime.GOOS == "windows" {
			t.Skip("Skipping permission test on Windows")
		}

		tokenPath := filepath.Join(tmpDir, "noperm_token.json")
		tokenBytes := []byte(`{"access_token": "test"}`)
		// Set permissions to 0000 (no permissions) to test permission denied scenario
		err := os.WriteFile(tokenPath, tokenBytes, 0000)
		if err != nil {
			t.Fatalf("Failed to write no-permission file: %v", err)
		}
		defer os.Chmod(tokenPath, 0600) // Cleanup

		token, err := readToken(tokenPath)
		if err == nil {
			t.Error("Expected error for permission denied, got nil")
		}
		if token != nil {
			t.Errorf("Expected nil token for permission denied, got: %v", token)
		}
	})
}

func TestCreateTokenUpdateCallback(t *testing.T) {
	tmpDir := t.TempDir()

	t.Run("callback successfully updates token", func(t *testing.T) {
		tokenPath := filepath.Join(tmpDir, "callback_token.json")
		var diagnostics diag.Diagnostics

		callback := createTokenUpdateCallback(tokenPath, &diagnostics)
		testToken := &oauth2.Token{
			AccessToken:  "callback-test-token",
			RefreshToken: "callback-refresh-token",
		}

		// Execute callback
		callback(testToken)

		// Verify no diagnostics were added
		if diagnostics.HasError() {
			t.Errorf("Expected no errors in diagnostics, got: %v", diagnostics.Errors())
		}
		if len(diagnostics.Warnings()) > 0 {
			t.Errorf("Expected no warnings in diagnostics, got: %v", diagnostics.Warnings())
		}

		// Verify token was written
		content, err := os.ReadFile(tokenPath)
		if err != nil {
			t.Fatalf("Failed to read token file: %v", err)
		}

		var readToken oauth2.Token
		if err := json.Unmarshal(content, &readToken); err != nil {
			t.Fatalf("Failed to unmarshal token: %v", err)
		}

		if readToken.AccessToken != testToken.AccessToken {
			t.Errorf("Expected access token %s, got %s", testToken.AccessToken, readToken.AccessToken)
		}
	})

	t.Run("callback adds warning on error", func(t *testing.T) {
		// Use invalid path that will cause error
		invalidPath := filepath.Join(tmpDir, "nonexistent", "subdir", "token.json")
		var diagnostics diag.Diagnostics

		callback := createTokenUpdateCallback(invalidPath, &diagnostics)
		testToken := &oauth2.Token{AccessToken: "test-token"}

		// Execute callback
		callback(testToken)

		// Verify warning was added
		if !diagnostics.HasError() && len(diagnostics.Warnings()) == 0 {
			t.Error("Expected warning in diagnostics, got none")
		}

		warnings := diagnostics.Warnings()
		if len(warnings) > 0 {
			warningFound := false
			for _, w := range warnings {
				if w.Summary() == "Unable to update token" {
					warningFound = true
					break
				}
			}
			if !warningFound {
				t.Error("Expected 'Unable to update token' warning, but it was not found")
			}
		}
	})

	t.Run("callback can be called multiple times", func(t *testing.T) {
		tokenPath := filepath.Join(tmpDir, "multi_callback_token.json")
		var diagnostics diag.Diagnostics

		callback := createTokenUpdateCallback(tokenPath, &diagnostics)

		// First call
		firstToken := &oauth2.Token{AccessToken: "first-callback-token"}
		callback(firstToken)

		// Second call
		secondToken := &oauth2.Token{AccessToken: "second-callback-token"}
		callback(secondToken)

		// Verify latest token was written
		content, err := os.ReadFile(tokenPath)
		if err != nil {
			t.Fatalf("Failed to read token file: %v", err)
		}

		var readToken oauth2.Token
		if err := json.Unmarshal(content, &readToken); err != nil {
			t.Fatalf("Failed to unmarshal token: %v", err)
		}

		if readToken.AccessToken != "second-callback-token" {
			t.Errorf("Expected access token 'second-callback-token', got %s", readToken.AccessToken)
		}

		// Verify no warnings
		if len(diagnostics.Warnings()) > 0 {
			t.Errorf("Expected no warnings, got: %v", diagnostics.Warnings())
		}
	})
}
