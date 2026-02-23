package middleware

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"strings"
	"time"

	"github.com/jizo-hr/sharepwd/internal/model"
	"github.com/jizo-hr/sharepwd/internal/repository"
)

type apiKeyContextKey string

const APIKeyContextKey apiKeyContextKey = "api_key"

func APIKeyAuth(repo *repository.APIKeyRepository) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte("{\"error\":\"authorization header required\"}"))
				return
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte("{\"error\":\"invalid authorization format\"}"))
				return
			}

			token := parts[1]
			h := sha256.Sum256([]byte(token))
			keyHash := hex.EncodeToString(h[:])

			apiKey, err := repo.GetByKeyHash(r.Context(), keyHash)
			if err != nil || apiKey == nil {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte("{\"error\":\"invalid api key\"}"))
				return
			}

			if apiKey.ExpiresAt != nil && time.Now().After(*apiKey.ExpiresAt) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte("{\"error\":\"api key expired\"}"))
				return
			}

			go repo.UpdateLastUsed(context.Background(), apiKey.ID)

			ctx := context.WithValue(r.Context(), APIKeyContextKey, apiKey)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func GetAPIKey(ctx context.Context) *model.APIKey {
	key, _ := ctx.Value(APIKeyContextKey).(*model.APIKey)
	return key
}
