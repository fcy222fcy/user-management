package handler

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"user-management/internal/service"
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
	mux.HandleFunc("/users", LoginHandler)
	mux.HandleFunc("/users/", RegisterHandler)

}

// req 是DTO(数据传输对象) 接收前端发过来的json数据
var resp struct {
	User     string `json:"user"`
	Password string `json:"password"`
	Email    string `json:"email"`
}

// RegisterHandler 处理用户注册
func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	// 只接受post请求
	if r.Method != http.MethodPost {
		//当服务器执行这行代码时，HTTP 响应头中的状态行会被设置为 405 Method Not Allowed
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, `{"code":405,"message":"只支持POST方法"`)
		return
	}
	// 解析表单
	if err := r.ParseForm(); err != nil {
		// 执行这行代码后，HTTP 响应的状态行变为 400 Bad Request,这通常意味着错误在于客户端(发送请求的一方,而不是服务器内部错误)
		w.WriteHeader(http.StatusBadRequest)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, `{"code":400,"message":"表单解析失败"}`)
		return
	}
	username := r.FormValue("username")
	password := r.FormValue("password")
	email := r.FormValue("email")

	// 调用 Service层注册
	user, err := UserSvc.Register(username, password, email)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintln(w, "注册失败:", err)
		return
	}
	// 返回成功响应
	fmt.Fprintf(w, "注册成功!用户名 : %s,用户ID: %d\n", user.Username, user.ID)
}

// LoginHandler 主要功能是解析获取字段,然后传给service层
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		// 渲染登录页面
		t, err := template.ParseFiles("web/templates/login.html")
		if err != nil {
			log.Fatal(err)
		}
		err = t.Execute(w, nil)
		if err != nil {
			log.Fatal(err)
		}
	} else if r.Method == http.MethodPost {
		// 处理登录表单提交
		err := r.ParseForm()
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintln(w, "表单解析失败:", err)
			return
		}

		username := r.FormValue("username")
		password := r.FormValue("password")

		// 调用 Service 层登录
		user, err := UserSvc.Login(username, password)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			fmt.Fprintln(w, "登录失败:", err)
			return
		}

		// 登录成功
		fmt.Fprintf(w, "登录成功！欢迎 %s (ID: %d)\n", user.Username, user.ID)
	} else {
		// 不支持的请求方法
		w.WriteHeader(http.StatusMethodNotAllowed)
		fmt.Fprintln(w, "只支持 GET 和 POST 方法")
	}
}
