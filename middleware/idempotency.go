package middleware

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/icodeologist/poultry-backend/db"
	"github.com/icodeologist/poultry-backend/models"
	"gorm.io/gorm"
)

// func Middleware(next http.Handler) http.Handler {
// 	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		log.Println("passing through middleware")
// 		// next.ServeHTTP(w, r)
// 	})
//
// }

func IdempotencyMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			next.ServeHTTP(w, r)
			return
		}

		key := r.Header.Get("Idempotency-Key")
		var paymentImp models.PaymentIdempotency

		res := db.DB.Where("key=?", key).First(&paymentImp)
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			// record does not exist
			pi := models.PaymentIdempotency{
				Key:            key,
				ResponseBody:   "",
				ResponseStatus: 0,
				CreatedAt:      time.Now(),
			}
			if err := db.DB.Create(&pi).Error; err != nil {
				http.Error(w, "could not claim the key", http.StatusInternalServerError)
				return
			}
			rec := &responseRecorder{ResponseWriter: w}
			next.ServeHTTP(rec, r)

			db.DB.Model(&pi).Updates(models.PaymentIdempotency{
				ResponseStatus: rec.Code,
				ResponseBody:   rec.Body,
			})

		} else if res.Error != nil {
			http.Error(w, fmt.Sprint(res.Error), http.StatusInternalServerError)
			return
		} else {
			if paymentImp.ResponseStatus == 0 {
				http.Error(w, `{"error": "request already in progress, try again shortly"}`, http.StatusConflict)
				return
			}
			w.WriteHeader(paymentImp.ResponseStatus)
			w.Write([]byte(paymentImp.ResponseBody))
			return
		}

	})
}

type responseRecorder struct {
	http.ResponseWriter
	Code int
	Body string
}
