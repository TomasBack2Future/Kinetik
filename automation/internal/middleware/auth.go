package middleware

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
	"strings"

	"github.com/TomasBack2Future/Kinetik/automation/pkg/logger"
)

// ValidateGitHubWebhook validates GitHub webhook signatures
func ValidateGitHubWebhook(secret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			signature := r.Header.Get("X-Hub-Signature-256")
			if signature == "" {
				logger.Warn("Missing X-Hub-Signature-256 header")
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			// Read body
			body, err := io.ReadAll(r.Body)
			if err != nil {
				logger.Error("Failed to read request body", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
			r.Body = io.NopCloser(bytes.NewBuffer(body))

			// Verify signature
			if !verifySignature(secret, signature, body) {
				logger.Warn("Invalid webhook signature")
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// verifySignature verifies HMAC-SHA256 signature
func verifySignature(secret, signature string, body []byte) bool {
	// GitHub sends signature as "sha256=<hex>"
	if !strings.HasPrefix(signature, "sha256=") {
		return false
	}

	signatureHex := strings.TrimPrefix(signature, "sha256=")
	expectedSignature, err := hex.DecodeString(signatureHex)
	if err != nil {
		return false
	}

	// Compute HMAC
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	actualSignature := mac.Sum(nil)

	// Compare signatures
	return hmac.Equal(expectedSignature, actualSignature)
}
