package handler

import (
	"html/template"
	"net/http"
	"user-management/pkg/response"
)

// RegisterHandler 处理用户注册
func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	// GET 使用 template渲染注册页面
	// POST 使用 Parse+JSONResponse
	case http.MethodGet:
		// 渲染注册页面
		t, err := template.ParseFiles("web/templates/login.html")
		if err != nil {
			response.JsonResponse(w, http.StatusInternalServerError, "页面渲染失败", nil)
			return
		}
		// 传入nil,说明模版里面没有占位符
		t.Execute(w, nil)
	case http.MethodPost:
		// 解析表单
		//--前端发送来的请求是 multipart/form-data格式的,不能直接用这个函数
		//if err := r.ParseForm(); err != nil {
		if err := r.ParseMultipartForm(10 << 20); err != nil {
			response.JsonResponse(w, http.StatusBadRequest, "表单解析失败", nil)
			return
		}

		username := r.FormValue("username")
		email := r.FormValue("email")
		password := r.FormValue("password")

		// 调用 Service层注册
		user, err := UserSvc.Register(username, password, email)
		if err != nil {
			response.JsonResponse(w, http.StatusBadRequest, err.Error(), nil)
			return
		}

		// 返回成功响应
		response.JsonResponse(w, http.StatusOK, "注册成功", map[string]interface{}{
			"user_id":  user.ID,
			"username": user.Username,
			"email":    user.Email,
		})
	default:
		response.JsonResponse(w, http.StatusMethodNotAllowed, "只支持 GET 和 POST 方法", nil)
	}
}

// LoginHandler 主要功能是解析获取字段,然后传给service层
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		// 渲染登录页面
		t, err := template.ParseFiles("web/templates/login.html")
		if err != nil {
			response.JsonResponse(w, http.StatusInternalServerError, "页面渲染失败", nil)
			return
		}
		t.Execute(w, nil)
	case http.MethodPost:
		// 解析表单
		//--前端发送来的请求是 multipart/form-data格式的,不能直接用这个函数
		//if err := r.ParseForm(); err != nil {
		if err := r.ParseMultipartForm(10 << 20); err != nil {
			response.JsonResponse(w, http.StatusBadRequest, "表单解析失败", nil)
			return
		}
		// 验证类型
		//fmt.Println("Content-Type:", r.Header.Get("Content-Type"))

		username := r.FormValue("username")
		password := r.FormValue("password")

		// 调用 Service 层登录
		user, errLogin := UserSvc.Login(username, password)
		if errLogin != nil {
			response.JsonResponse(w, http.StatusUnauthorized, errLogin.Error(), nil)
			return
		}

		// 返回成功响应
		response.JsonResponse(w, http.StatusOK, "登录成功", map[string]interface{}{
			"success":  true,
			"message":  "登录成功",
			"redirect": "/", // 登录成功后跳转到首页
			"user_id":  user.ID,
			"username": user.Username,
			"email":    user.Email,
		})
	default:
		response.JsonResponse(w, http.StatusMethodNotAllowed, "只支持 GET 和 POST 方法", nil)
	}
}
