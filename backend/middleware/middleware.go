package middleware

import (
	"context"
	"net/http"
	"strings"
	"time"

	"example.com/collaborative-coding-editor/config"
	"github.com/golang-jwt/jwt/v5"
)

type key string

const UserKey key = "user"

// Validates the access token provided in the Auth Header
func JWTAuthentication(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Missing auth token", http.StatusUnauthorized)
			return
		}

		// Bearer Token header
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, "Invalid auth token", http.StatusUnauthorized)
			return
		}

		tokenString := parts[1]

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(config.AppConfig.JWTSecret), nil
		})
		if err != nil || !token.Valid {
			http.Error(w, "Invalid refresh token", http.StatusUnauthorized)
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			http.Error(w, "Invalid token claims", http.StatusUnauthorized)
			return
		}

		if exp, ok := claims["exp"].(float64); ok {
			if int64(exp) < time.Now().Unix() {
				http.Error(w, "Token expired", http.StatusUnauthorized)
				return
			}
		}

		ctx := context.WithValue(r.Context(), UserKey, claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
