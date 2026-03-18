package middleware

import (
	"context"
	"net/http"
	"strings"
	"user-management/internal/model"
	"user-management/internal/service"
	"user-management/pkg/util"
)

func AuthMiddleware(userService *service.UserService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := r.Header.Get("Authorization")
			if token == "" {
				model.JsonResponse(w, http.StatusUnauthorized, "unauthorized", nil)
				return
			}
			token = strings.TrimPrefix(token, "Bearer ")
			if token == "" {
				model.JsonResponse(w, http.StatusUnauthorized, "unauthorized", nil)
				return
			}

			claims, err := util.ParseToken(token)
			if err != nil {
				model.JsonResponse(w, http.StatusUnauthorized, "invalid token", nil)
				return
			}

			userID := claims.UserID
			user, err := userService.GetUserByID(userID)
			if err != nil {
				model.JsonResponse(w, http.StatusUnauthorized, "user not found", nil)
				return
			}

			if user.Status == 0 {
				model.JsonResponse(w, http.StatusForbidden, "你的账号已被封禁", nil)
				return
			}

			ctx := context.WithValue(r.Context(), "userID", userID)
			ctx = context.WithValue(ctx, "user", user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
