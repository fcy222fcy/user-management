package model

import "time"

type User struct { // 直接模版渲染不需要加json,但是如果是返回json就需要加
	ID       int64  `json:"ID"`
	Username string `json:"Username"`
	// password 不能加json,不然会暴露给前端
	Password  string    `json:"-"`
	Email     string    `json:"Email"`
	AvatarURL string    `json:"AvatarURL"`
	Role      string    `json:"Role"`
	Status    int       `json:"Status"`
	CreatedAt time.Time `json:"CreatedAt"`
	UpdatedAt time.Time `json:"UpdatedAt"`
}

type RegisterRequest struct {
	Username string
	Password string
	Email    string
}
type LoginRequest struct {
	Username string
	Password string
}
type UserResponse struct {
	ID       int64
	Username string
	Email    string
	Role     string
}
