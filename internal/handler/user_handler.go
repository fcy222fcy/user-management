package handler

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"user-management/internal/service"
	"user-management/pkg/response"
)

// router 路由注册

// Handler层--处理HTTP请求和响应----改为controller
// Service层--核心业务逻辑,查询判断用户
// Repository层--和数据库交互,插入删除

// UserSvc 注册全局变量,方便注册handler时使用
var UserSvc *service.UserService

// RegisterUserRoutes 注册路由函数
func RegisterUserRoutes(mux *http.ServeMux, svc *service.UserService) {
	UserSvc = svc

	// restful风格
	log.Printf("[路由] 注册 /users Handler")
	mux.HandleFunc("/users", LoginHandler)
	log.Printf("[路由] 注册 /users/ Handler")
	mux.HandleFunc("/users/", RegisterHandler)

}

// req 是DTO(数据传输对象) 接收前端发过来的json数据
//var resp struct {
//	User     string `json:"user"`
//	Password string `json:"password"`
//	Email    string `json:"email"`
//}

// RegisterHandler 处理用户注册
func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	// 只接受post请求
	if r.Method != http.MethodPost {
		response.JsonResponse(w, http.StatusMethodNotAllowed, "只支持POST方法", nil)
		return
	}
	// 解析表单
	if err := r.ParseForm(); err != nil {
		response.JsonResponse(w, http.StatusBadRequest, "表单解析失败", nil)
		return
	}
	username := r.FormValue("username")
	password := r.FormValue("password")
	email := r.FormValue("email")

	log.Printf("获取到表单数据: username=%s, email=%s", username, email)
	log.Printf("开始注册，用户名: %s, 邮箱: %s", username, email)
	// 调用 Service层注册
	user, err := UserSvc.Register(username, password, email)
	if err != nil {
		log.Printf("注册失败: %v", err)
		response.JsonResponse(w, http.StatusBadRequest, "注册失败"+err.Error(), nil)
		return
	}
	log.Printf("注册成功!用户ID: %d, 用户名: %s", user.ID, user.Username)
	// 返回成功响应
	response.JsonResponse(w, http.StatusOK, "注册成功", map[string]interface{}{
		"user_id":  user.ID,
		"username": user.Username,
		"email":    user.Email,
	})
}

// LoginHandler 主要功能是解析获取字段,然后传给service层
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		// 渲染登录页面
		t, err := template.ParseFiles("web/templates/login.html")
		if err != nil {
			log.Printf("渲染登录页卖你失败:%v", err)
			response.JsonResponse(w, http.StatusInternalServerError, "页面渲染失败", nil)
			return
		}
		t.Execute(w, nil)
	case http.MethodPost:
		// 解析表单
		if err := r.ParseForm(); err != nil {
			response.JsonResponse(w, http.StatusBadRequest, "表单解析失败", nil)
			return
		}

		username := r.FormValue("username")
		password := r.FormValue("password")

		// 调用 Service 层登录
		user, errLogin := UserSvc.Login(username, password)
		if errLogin != nil {
			w.WriteHeader(http.StatusUnauthorized)
			fmt.Fprintln(w, "登录失败:", errLogin)
			return
		}

		response.JsonResponse(w, http.StatusOK, "登录成功", map[string]interface{}{
			"user_id":  user.ID,
			"username": user.Username,
			"email":    user.Email,
		})
	default:
		// 不支持的请求方法
		response.JsonResponse(w, http.StatusMethodNotAllowed, "只支持 GET 和 POST 方法", nil)
	}
}
