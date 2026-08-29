package middleware

import (
	"context"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"net/http"

	"github.com/golang-jwt/jwt/v5"
)

const adminIDKey = "adminID"

// admin middleware parses through incoming reqeust and gets the unique key kind  of like jwt
// you take the key and decrypt it using hashing with local stored secret key and if they match you edit the role to admin
// admin can
//
//	update prices
//	view all orders
//	view or edit stocks
//	have a detail of cash flow  (nice dashboard with daily income expense basic crud)
func UserMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("user-authorization")
		if err != nil {
			switch {
			case errors.Is(err, http.ErrNoCookie):
				http.Error(w, "cookie not found", http.StatusBadRequest)
				return
			default:
				log.Println(err)
				http.Error(w, "server error", http.StatusInternalServerError)
				return
			}
		}

		// parse the token
		token, err := jwt.Parse(cookie.Value, func(token *jwt.Token) (any, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte("hello"), nil
		})

		// check err BEFORE touching token.Valid — a parse error can return
		// a nil or unusable token, so checking .Valid first risks a nil panic
		switch {
		case errors.Is(err, jwt.ErrTokenMalformed):
			slog.Warn("malformed token received")
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		case errors.Is(err, jwt.ErrTokenSignatureInvalid):
			slog.Warn("invalid token signature")
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		case errors.Is(err, jwt.ErrTokenExpired):
			slog.Info("expired token")
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		case errors.Is(err, jwt.ErrTokenNotValidYet):
			slog.Warn("token used before valid")
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		case err != nil:
			slog.Error("unexpected token validation error", "err", err)
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		// only safe to check token.Valid once we know err == nil
		if !token.Valid {
			slog.Warn("token parsed but not valid")
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		slog.Info("Success", "token expires in", token.Claims)

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		role, ok := claims["role"]
		if !ok {
			slog.Error("no role mentioned", "role", role)
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		slog.Info("Current role logged in", "role", role)

		// fixed: read "user_id" — matches the key set at login time
		userID, ok := claims["user_id"].(float64)
		if !ok || userID == 0.0 {
			slog.Error("invalid or missing user id")
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		if role == "staff" {
			ctx := context.WithValue(r.Context(), "staffID", userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		} else if role == "admin" {
			ctx := context.WithValue(r.Context(), "adminID", userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		} else {
			slog.Error("invalid role", "role", role)
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
	})
}
