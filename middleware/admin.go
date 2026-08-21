package middleware

import (
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/golang-jwt/jwt/v5"
)

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
			fmt.Println("You look nice today")
		case errors.Is(err, jwt.ErrTokenMalformed):
			fmt.Println("That's not even a token")
			http.Error(w, "invalid token", http.StatusBadRequest)
			return
		case errors.Is(err, jwt.ErrTokenSignatureInvalid):
			// Invalid signature
			fmt.Println("Invalid signature")
			http.Error(w, "invalid signature", http.StatusBadRequest)
			return
		case errors.Is(err, jwt.ErrTokenExpired) || errors.Is(err, jwt.ErrTokenNotValidYet):
			// Token is either expired or not active yet
			fmt.Println("Timing is everything")
			http.Error(w, "expired token", http.StatusBadRequest)
			return
		default:
			fmt.Println("Couldn't handle this token:", err)
			http.Error(w, "unexpected error", http.StatusBadRequest)
			return
		}
		next.ServeHTTP(w, r)
	})
}
