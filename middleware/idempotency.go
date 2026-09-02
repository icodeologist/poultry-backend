package middleware

import (
	"errors"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/icodeologist/poultry-backend/models"
	"gorm.io/gorm"
)

func IdempotencyMiddleware(db *gorm.DB) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost {
				log.Println("Method is not post")
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
				return
			}
			key := r.Header.Get("Idempotency-Key")
			if key == "" {
				slog.Error("Key is empty", "Key", key)
				http.Error(w, "key is empty. Unauthorized", http.StatusUnauthorized)
				// may be check if the key is correct here
				// liek the format or something
				// we will do that once we have key format set in front end
				return
			}
			// now we assume key is valid
			// is this new key or do we already have registered this key
			// fecth the key data from db and if it exists then we have already processing or processed this request
			var ip models.OrderAndPaymentIdempotency
			ip.Key = key
			ip.CreatedAt = time.Now()
			err := db.Create(&ip).Error
			// one ip is already registed with this key
			// so reject this but dont send error send the status code and staus response from the original request
			if errors.Is(err, gorm.ErrDuplicatedKey) {
				log.Printf("Duplicated key reqeust : %v\n", err)
				var existing models.OrderAndPaymentIdempotency
				res := db.Where("key=?", key).First(&existing)
				// no chnace of having no keys since we already know it exists with the key
				if res.Error != nil {
					http.Error(w, res.Error.Error(), http.StatusInternalServerError)
					return
				} else {
					if existing.ResponseStatus == 0 {
						fmt.Fprintf(w, fmt.Sprintf("Request with key :%v is still processing.\n", key))
						return
					}
					w.WriteHeader(existing.ResponseStatus)
					w.Write([]byte(existing.ResponseBody))
					return
				}
			} else if err != nil {
				log.Printf("Error server side while creating Idempotency-Key : %v\n", err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			} else {
				rr := httptest.NewRecorder()
				next.ServeHTTP(rr, r)
				ip.ResponseStatus = rr.Code
				ip.ResponseBody = rr.Body.String()
				db.Save(&ip)
				w.WriteHeader(ip.ResponseStatus)
				w.Write([]byte(ip.ResponseBody))
			}
		})
	}
}
