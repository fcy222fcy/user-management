package handler

// 专门管理路由
import (
	"net/http"
	"user-management/internal/middleware"
	"user-management/internal/service"
)

// UserHandler 注册全局变量,方便注册handler时使用-依赖注入
type UserHandler struct {
	UserService *service.UserService
}

func NewUserHandler(userService *service.UserService) *UserHandler {
	return &UserHandler{
		UserService: userService,
	}
}

// RegisterUserRoutes 注册路由函数
func RegisterUserRoutes(mux *http.ServeMux, h *UserHandler) {

	// 页面
	mux.HandleFunc("/", Index)
	mux.HandleFunc("/login", h.LoginPage)
	mux.HandleFunc("/userList", h.UserListPage)

	// API
	mux.HandleFunc("/api/login", h.LoginHandler)
	mux.HandleFunc("/api/register", h.RegisterHandler)
	mux.HandleFunc("/api/userList", h.ListHandler)
	// jwt保护
	mux.Handle("/api/stats", Wrap(h.StatsHandler, h.UserService))
	mux.Handle("/api/profile", Wrap(h.ProfileHandler, h.UserService))
	mux.Handle("/api/user/create", Wrap(h.CreateUserHandler, h.UserService))
	mux.Handle("/api/user/update", Wrap(h.UpdateUserHandler, h.UserService))
	mux.Handle("/api/user/delete", Wrap(h.DeleteUserHandler, h.UserService))
}

// LoginPage 登录页
func (h *UserHandler) LoginPage(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "web/templates/login.html")
}

// RegisterPage 注册页
func (h *UserHandler) RegisterPage(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "web/templates/register.html")
}

// UserListPage 用户列表页
func (h *UserHandler) UserListPage(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "web/templates/userList.html")
}

// Wrap 用于保护接口
func Wrap(handler http.HandlerFunc, userService *service.UserService) http.Handler {
	return middleware.AuthMiddleware(userService)(handler)
}

func Index(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	http.ServeFile(w, r, "web/templates/index.html")
}
