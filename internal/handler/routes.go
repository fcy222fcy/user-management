package handler

// 专门管理路由
import (
	"net/http"
	"user-management/internal/service"
)

// UserSvc 注册全局变量,方便注册handler时使用
var UserSvc *service.UserService

// RegisterUserRoutes 注册路由函数
func RegisterUserRoutes(mux *http.ServeMux, svc *service.UserService) {
	UserSvc = svc

	mux.HandleFunc("/login", LoginHandler)
	mux.HandleFunc("/register", RegisterHandler)
	mux.HandleFunc("/userList", ListHandler)
}
