package util

import (
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"net/http"
	"os"
	"time"
	"user-management/internal/model"
	"user-management/internal/service"
)

var jwtSecret = []byte(getEnvOrDefault("JWT_SECRET", "secret"))

// 从env中读取
func getEnvOrDefault(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

type Claims struct {
	UserID int64 `json:"user_id"`
	jwt.RegisteredClaims
}

// GenerateToken 生成token
func GenerateToken(userID int64) (string, error) {
	claims := Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			// 24小时过期
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	// 根据Header和Payload生成签名,最终拼接成完整的字符串
	return token.SignedString(jwtSecret)

}

// ParseToken 解析 Token
func ParseToken(tokenStr string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
		// 校验token的算法签名
		if token.Method.Alg() != jwt.SigningMethodHS256.Alg() {
			return nil, fmt.Errorf("unexpected signing method: %s", token.Method.Alg())
		}
		return jwtSecret, nil
	})
	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, errors.New("token invalid")
	}
	return claims, nil
}

// GetUserID 从请求中获取 userID
func GetUserID(r *http.Request) int64 {
	userID, ok := r.Context().Value("userID").(int64)
	if !ok {
		return 0
	}
	return userID
}

// GetCurrentUser 从请求中获取当前用户
func GetCurrentUser(r *http.Request, userService *service.UserService) (*model.User, error) {
	userID := GetUserID(r)
	if userID == 0 {
		return nil, nil
	}
	return userService.GetUserByID(userID)
}
