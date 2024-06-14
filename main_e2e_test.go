package main

import (
	"bytes"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"testing"
	"time"

	_ "github.com/lib/pq"
)

const (
	dbUser     = "postgres"
	dbPassword = "localdb"
	dbName     = "nakama"
	dbHost     = "postgres"
	dbPort     = 5432
)

func setupDatabase() error {
	connStr := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable", dbUser, dbPassword, dbHost, dbPort, dbName)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return err
	}
	defer db.Close()

	_, err = db.Exec("DELETE FROM file_data")
	return err
}

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}

func TestProcessFilePayloadE2E(t *testing.T) {
	// Read the file content from the given path
	filePath := "./sample_files/core/1.0.0.json"
	fileContent, err := ioutil.ReadFile(filePath)
	if err != nil {
		t.Fatalf("Error reading file: %v", err)
	}

	// Calculate the correct hash for the provided file content
	hash := sha256.Sum256(fileContent)
	correctHash := hex.EncodeToString(hash[:])

	tests := []struct {
		name            string
		setupFunc       func() error
		payload         map[string]string
		expectedStatus  int
		expectedHash    string
		expectedContent string
		expectedMessage string
	}{
		{
			name: "File does not exist and will be added",
			setupFunc: func() error {
				return setupDatabase() // Ensure the file_data table is empty
			},
			payload:         map[string]string{"type": "core", "version": "1.0.0", "hash": "null"},
			expectedStatus:  200,
			expectedHash:    correctHash,
			expectedContent: string(fileContent),
			expectedMessage: "",
		},
		{
			name: "File exists and hash matches",
			setupFunc: func() error {
				err := setupDatabase() // Ensure the file_data table is empty
				if err != nil {
					return err
				}
				return addTestFileToDatabase(correctHash, string(fileContent))
			},
			payload:         map[string]string{"type": "core", "version": "1.0.0", "hash": correctHash},
			expectedStatus:  200,
			expectedHash:    correctHash,
			expectedContent: string(fileContent),
			expectedMessage: "",
		},
		{
			name: "File exists and hash does not match",
			setupFunc: func() error {
				err := setupDatabase() // Ensure the file_data table is empty
				if err != nil {
					return err
				}
				return addTestFileToDatabase(correctHash, string(fileContent))
			},
			payload:         map[string]string{"type": "core", "version": "1.0.0", "hash": "incorrect_hash"},
			expectedStatus:  200,
			expectedHash:    "null",
			expectedContent: "null",
			expectedMessage: "",
		},
		{
			name: "File does not exist",
			setupFunc: func() error {
				return setupDatabase() // Ensure the file_data table is empty
			},
			payload:         map[string]string{"type": "nonexistent", "version": "1.0.0", "hash": "null"},
			expectedStatus:  404,
			expectedHash:    "",
			expectedContent: "",
			expectedMessage: "file not found",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if err := tc.setupFunc(); err != nil {
				t.Fatalf("failed to set up test: %v", err)
			}

			payloadBytes, err := json.Marshal(tc.payload)
			if err != nil {
				t.Fatalf("failed to marshal payload: %v", err)
			}

			// Convert the payload to a JSON string inside a JSON string
			jsonPayload := fmt.Sprintf(`"%s"`, escapeJSONString(string(payloadBytes)))

			res, err := http.Post("http://nakama:7350/v2/rpc/process_file_payload?http_key=defaulthttpkey", "application/json", bytes.NewBuffer([]byte(jsonPayload)))
			if err != nil {
				t.Fatalf("failed to make POST request: %v", err)
			}
			defer res.Body.Close()

			body, err := io.ReadAll(res.Body)
			if err != nil {
				t.Fatalf("failed to read response body: %v", err)
			}

			if res.StatusCode != tc.expectedStatus {
				t.Fatalf("expected status code %d, got %d, response: %s", tc.expectedStatus, res.StatusCode, string(body))
			}

			var response map[string]interface{}
			if err := json.Unmarshal(body, &response); err != nil {
				t.Fatalf("failed to unmarshal response: %v", err)
			}

			// Log the full response for debugging
			t.Logf("response: %v", response)

			// Extract the payload if present
			var payload map[string]interface{}
			if p, ok := response["payload"].(string); ok {
				if err := json.Unmarshal([]byte(p), &payload); err != nil {
					t.Fatalf("failed to unmarshal payload: %v", err)
				}
			} else {
				payload = response
			}

			if res.StatusCode == 200 {
				if hash, ok := payload["hash"].(string); ok {
					if hash != tc.expectedHash {
						t.Errorf("expected hash %s, got %s", tc.expectedHash, hash)
					}
				} else if tc.expectedHash != "" {
					t.Errorf("expected hash %s, got none", tc.expectedHash)
				}

				if content, ok := payload["content"].(string); ok {
					if content != "null" {
						if content != tc.expectedContent {
							t.Errorf("expected content %s, got %s", tc.expectedContent, content)
						}
					} else if content != "null" && tc.expectedContent != "null" {
						t.Errorf("expected content null, got %s", content)
					}
				} else if tc.expectedContent != "" {
					t.Errorf("expected content %s, got none", tc.expectedContent)
				}
			} else {
				if message, ok := response["message"].(string); ok {
					if message != tc.expectedMessage {
						t.Errorf("expected message %s, got %s", tc.expectedMessage, message)
					}
				} else if tc.expectedMessage != "" {
					t.Errorf("expected message %s, got none", tc.expectedMessage)
				}
			}
		})
	}
}

func addTestFileToDatabase(hash, content string) error {
	connStr := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable", dbUser, dbPassword, dbHost, dbPort, dbName)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return err
	}
	defer db.Close()

	_, err = db.Exec("INSERT INTO file_data (type, version, content, hash, processed_at) VALUES ($1, $2, $3, $4, $5)",
		"core", "1.0.0", content, hash, time.Now())
	return err
}

// escapeJSONString escapes special characters in a JSON string
func escapeJSONString(s string) string {
	escaped := ""
	for _, c := range s {
		switch c {
		case '"':
			escaped += `\"`
		case '\\':
			escaped += `\\`
		default:
			escaped += string(c)
		}
	}
	return escaped
}
