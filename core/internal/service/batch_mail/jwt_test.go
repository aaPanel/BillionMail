package batch_mail

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestUnsubscribeJWT_RoundTrip tests JWT generation and parsing with a known secret.
// Since getConfig() depends on DB, we directly construct tokens using jwt-go
// to verify the parsing logic in ParseUnsubscribeJWT-compatible fashion.
func TestUnsubscribeJWT_RoundTrip(t *testing.T) {
	secret := "test-secret-key-for-unit-tests"

	tests := []struct {
		name       string
		email      string
		templateId int
		taskId     int
		groupId    int
	}{
		{
			name:       "basic email",
			email:      "user@example.com",
			templateId: 1,
			taskId:     100,
			groupId:    5,
		},
		{
			name:       "special characters in email",
			email:      "user+tag@sub.domain.com",
			templateId: 999,
			taskId:     0,
			groupId:    0,
		},
		{
			name:       "zero IDs",
			email:      "test@test.com",
			templateId: 0,
			taskId:     0,
			groupId:    0,
		},
		{
			name:       "large IDs",
			email:      "big@ids.org",
			templateId: 999999,
			taskId:     888888,
			groupId:    777777,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Generate token manually (same logic as GenerateUnsubscribeJWT but with known secret)
			claims := UnsubscribeClaims{
				Email:      tt.email,
				TemplateId: tt.templateId,
				TaskId:     tt.taskId,
				GroupId:    tt.groupId,
				RegisteredClaims: jwt.RegisteredClaims{
					ExpiresAt: jwt.NewNumericDate(time.Now().Add(30 * 24 * time.Hour)),
					IssuedAt:  jwt.NewNumericDate(time.Now()),
				},
			}

			token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
			tokenString, err := token.SignedString([]byte(secret))
			require.NoError(t, err)
			assert.NotEmpty(t, tokenString)

			// Parse back
			parsed, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
				return []byte(secret), nil
			})
			require.NoError(t, err)
			assert.True(t, parsed.Valid)

			mapClaims, ok := parsed.Claims.(jwt.MapClaims)
			require.True(t, ok)
			assert.Equal(t, tt.email, mapClaims["email"])
			assert.Equal(t, float64(tt.templateId), mapClaims["template_id"])
			assert.Equal(t, float64(tt.taskId), mapClaims["task_id"])
			assert.Equal(t, float64(tt.groupId), mapClaims["group_id"])
		})
	}
}

func TestUnsubscribeJWT_ExpiredToken(t *testing.T) {
	secret := "test-secret-expired"
	claims := UnsubscribeClaims{
		Email:      "expired@example.com",
		TemplateId: 1,
		TaskId:     1,
		GroupId:    1,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)), // expired
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(secret))
	require.NoError(t, err)

	_, err = jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})
	assert.Error(t, err, "expired token should fail parsing")
}

func TestUnsubscribeJWT_WrongSecret(t *testing.T) {
	claims := UnsubscribeClaims{
		Email:      "user@example.com",
		TemplateId: 1,
		TaskId:     1,
		GroupId:    1,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte("correct-secret"))
	require.NoError(t, err)

	_, err = jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte("wrong-secret"), nil
	})
	assert.Error(t, err, "wrong secret should fail validation")
}

func TestSubscribeConfirmJWT_RoundTrip(t *testing.T) {
	secret := "test-subscribe-secret"

	tests := []struct {
		name       string
		email      string
		groupToken string
	}{
		{
			name:       "basic",
			email:      "subscriber@example.com",
			groupToken: "abc123",
		},
		{
			name:       "long group token",
			email:      "sub@domain.org",
			groupToken: "very-long-group-token-with-special-chars-123!@#",
		},
		{
			name:       "empty group token",
			email:      "empty@test.com",
			groupToken: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims := SubscribeConfirmClaims{
				Email:      tt.email,
				GroupToken: tt.groupToken,
				RegisteredClaims: jwt.RegisteredClaims{
					ExpiresAt: jwt.NewNumericDate(time.Now().Add(30 * 24 * time.Hour)),
					IssuedAt:  jwt.NewNumericDate(time.Now()),
				},
			}

			token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
			tokenString, err := token.SignedString([]byte(secret))
			require.NoError(t, err)

			parsed, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
				return []byte(secret), nil
			})
			require.NoError(t, err)

			mapClaims, ok := parsed.Claims.(jwt.MapClaims)
			require.True(t, ok)
			assert.Equal(t, tt.email, mapClaims["email"])
			assert.Equal(t, tt.groupToken, mapClaims["group_token"])
		})
	}
}

func TestUnsubscribeClaims_SigningMethod(t *testing.T) {
	secret := "test-signing-method"
	claims := UnsubscribeClaims{
		Email:      "user@test.com",
		TemplateId: 1,
		TaskId:     1,
		GroupId:    1,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	// Sign with HS256
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(secret))
	require.NoError(t, err)

	// Parse and validate signing method matches what ParseUnsubscribeJWT expects
	parsed, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			t.Fatalf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})
	require.NoError(t, err)
	assert.True(t, parsed.Valid)
}
