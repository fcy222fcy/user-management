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
			// 从请求头提取并验证 token
			token := r.Header.Get("Authorization")
			if token == "" {
				model.JsonResponse(w, http.StatusUnauthorized, "unauthorized", nil)
				return
			}
			// 去除前缀
			token = strings.TrimPrefix(token, "Bearer ")
			if token == "" {
				model.JsonResponse(w, http.StatusUnauthorized, "unauthorized", nil)
				return
			}
			// 解析token
			claims, err := util.ParseToken(token)
			if err != nil {
				model.JsonResponse(w, http.StatusUnauthorized, "invalid token", nil)
				return
			}
			// 查询用户信息
			userID := claims.UserID
			user, CheckErr := userService.GetUserByID(userID)
			if CheckErr != nil {
				model.JsonResponse(w, http.StatusUnauthorized, "user not found", nil)
				return
			}
			// 判断用户状态
			if user.Status == 0 {
				model.JsonResponse(w, http.StatusForbidden, "你的账号已被封禁", nil)
				return
			}
			// 将用户信息存入到HTTP请求的 context
			ctx := context.WithValue(r.Context(), "userID", userID)
			ctx = context.WithValue(ctx, "user", user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
