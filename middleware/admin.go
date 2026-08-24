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
func AdminMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("admin-authorization")
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
		// parse  the token
		token, err := jwt.Parse(cookie.Value, func(token *jwt.Token) (any, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte("hello"), nil
		})
		log.Printf("token : %v\n", token)

		switch {
		case token.Valid:
			slog.Info("Success", "token expires in ", token.Claims)

		case errors.Is(err, jwt.ErrTokenMalformed):
			slog.Warn("malformed token received")
			http.Error(w, "unauthorized 1", http.StatusUnauthorized)
			return

		case errors.Is(err, jwt.ErrTokenSignatureInvalid):
			slog.Warn("invalid token signature")
			http.Error(w, "unauthorized 2", http.StatusUnauthorized)
			return

		case errors.Is(err, jwt.ErrTokenExpired):
			slog.Info("expired token")
			http.Error(w, "unauthorized 3", http.StatusUnauthorized)
			return

		case errors.Is(err, jwt.ErrTokenNotValidYet):
			slog.Warn("token used before valid")
			http.Error(w, "unauthorized 4", http.StatusUnauthorized)
			return

		default:
			slog.Error("unexpected token validation error", "err", err)
			http.Error(w, "unauthorized 5", http.StatusUnauthorized)
			return
		}
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			http.Error(w, "unauthorized 6", http.StatusUnauthorized)
			return
		}

		adminID, ok := claims["admin_id"].(float64)
		if !ok || adminID == 0.0 {
			log.Println("admin id : ", adminID)
			http.Error(w, "unauthorized 7", http.StatusUnauthorized)
			return
		}
		ctx := context.WithValue(r.Context(), "adminID", adminID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
