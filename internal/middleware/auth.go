package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/driver-hub/driver-hub/internal/auth"
	"github.com/driver-hub/driver-hub/internal/domain"
)

type ctxKey string

const ClaimsKey ctxKey = "claims"

func Auth(tm *auth.TokenManager, roles ...domain.Role) func(http.Handler) http.Handler {
	allowed := map[domain.Role]bool{}
	for _, r := range roles {
		allowed[r] = true
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			h := r.Header.Get("Authorization")
			if !strings.HasPrefix(h, "Bearer ") {
				http.Error(w, `{"error":"требуется авторизация"}`, http.StatusUnauthorized)
				return
			}
			claims, err := tm.Parse(strings.TrimPrefix(h, "Bearer "))
			if err != nil {
				http.Error(w, `{"error":"недействительный токен"}`, http.StatusUnauthorized)
				return
			}
			if len(allowed) > 0 && !allowed[claims.Role] {
				http.Error(w, `{"error":"недостаточно прав"}`, http.StatusForbidden)
				return
			}
			ctx := context.WithValue(r.Context(), ClaimsKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func ClaimsFrom(ctx context.Context) *auth.Claims {
	v, _ := ctx.Value(ClaimsKey).(*auth.Claims)
	return v
}
