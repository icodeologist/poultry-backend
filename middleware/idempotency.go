package middleware

import (
	"bytes"
	"net/http"
	"time"

	"gorm.io/gorm"
)

type responseRecorder struct {
	http.ResponseWriter // embeds the real one — still does the real writing
	code                int
	body                bytes.Buffer
}

func IdempotencyMiddleware(db *gorm.DB, tableName string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost {
				next.ServeHTTP(w, r)
				return
			}
			key := r.Header.Get("Idempotency-Key")
			if key == "" {
				http.Error(w, `{"error": "Idempotency-Key header required"}`, http.StatusBadRequest)
				return
			}
			claim := map[string]interface{}{
				"key":         key,
				"status_code": 0,
				"created_at":  time.Now(),
			}

			err := db.Table(tableName).Create(&claim).Error
			if err != nil {
				var existing struct {
					StatusCode int
					Body       []byte
				}
				db.Table(tableName).Where("key=?", key).First(&existing)
				if existing.StatusCode == 0 {
					http.Error(w, `{"error": "request already in progress, try again shortly"}`, http.StatusConflict)
					return
				}
				w.Header().Set("X-Idempotent-Replayed", "true")
				w.WriteHeader(existing.StatusCode)
				w.Write(existing.Body)
				return
			}
			rec := &responseRecorder{ResponseWriter: w, code: http.StatusOK}
			next.ServeHTTP(rec, r)

			db.Table(tableName).Where("key = ?", key).Updates(map[string]interface{}{
				"status_code": rec.code,
				"body":        rec.body.Bytes(),
			})

		})
	}
}
