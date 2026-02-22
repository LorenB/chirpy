package auth

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestValidateJWT(t *testing.T) {
	// Setup: create a real token to test with
	testUserID := uuid.New()
	testSecret := "test-secret-key"
	expiresIn := time.Hour

	validToken, err := MakeJWT(testUserID, testSecret, expiresIn)
	if err != nil {
		t.Fatalf("failed to create test token: %v", err)
	}

	tests := []struct {
		name        string
		tokenString string
		tokenSecret string
		wantID      uuid.UUID
		wantErr     error
	}{
		{
			name:        "valid token with correct secret",
			tokenString: validToken,
			tokenSecret: testSecret,
			wantID:      testUserID,
			wantErr:     nil,
		},
		// {
		// 	name:        "valid token with worng secret",
		// 	tokenString: validToken,
		// 	tokenSecret: "wrong-secret",
		// 	wantID:      uuid.Nil,
		// 	wantErr:     errors.New(""), // TODO: fix
		// },
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ValidateJWT(tt.tokenString, tt.tokenSecret)
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("expected error %v, got %v", tt.wantErr, err)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}

			if got != tt.wantID {
				t.Errorf("expected ID: %v, got %v", tt.wantID, got)
			}
		})
	}

}
