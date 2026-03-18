package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
	"user-management/internal/middleware"
	"user-management/internal/model"
	"user-management/internal/service"
	"user-management/pkg/util"
)

// RegisterHandler 处理用户注册
func (h *UserHandler) RegisterHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPost {
		model.JsonResponse(w, http.StatusMethodNotAllowed, "只支持POST方法", nil)
		return
	}

	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
		Email    string `json:"email"`
	}
	// 解析 JSON
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		model.JsonResponse(w, http.StatusBadRequest, "请求参数错误", nil)
		return
	}

	// 参数校验
	if req.Username == "" || req.Password == "" || req.Email == "" {
		model.JsonResponse(w, http.StatusBadRequest, "参数不能为空", nil)
		return
	}

	// 调用 Service
	_, err := h.UserService.Register(req.Username, req.Password, req.Email)
	if err != nil {
		model.JsonResponse(w, http.StatusBadRequest, err.Error(), nil)
		return
	}
	model.JsonResponse(w, http.StatusOK, "注册成功", nil)

}

// LoginHandler 主要功能是解析获取字段,然后传给service层
func (h *UserHandler) LoginHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPost {
		model.JsonResponse(w, http.StatusMethodNotAllowed, "只支持 POST 方法", nil)
		return
	}

	// 处理登录
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	// 解析 JSON 请求
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		model.JsonResponse(w, http.StatusBadRequest, "请求参数错误", nil)
		return
	}

	// 调用 Service 层
	user, err := h.UserService.Login(req.Username, req.Password)
	if err != nil {
		_ = h.UserService.Repo.CreateLoginLog(0, req.Username, false, r.RemoteAddr, r.UserAgent(), err.Error())
		model.JsonResponse(w, http.StatusUnauthorized, err.Error(), nil)
		return
	}

	// 返回 JSON
	token, _ := util.GenerateToken(user.ID)
	//if err != nil {
	//	response.JsonResponse(w, http.StatusInternalServerError, "生成token失败", nil)
	//	return
	//}

	model.JsonResponse(w, http.StatusOK, "登陆成功", map[string]interface{}{
		"token": token,
	})
	_ = h.UserService.Repo.CreateLoginLog(user.ID, user.Username, true, r.RemoteAddr, r.UserAgent(), "login success")

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

// ProfilePage 用户个人中心页
func (h *UserHandler) ProfilePage(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "web/templates/profile.html")
}

// DeleteUserHandler 删除用户处理
func (h *UserHandler) DeleteUserHandler(w http.ResponseWriter, r *http.Request) {

	//测试
	log.Println("handler层,删除用户")

	if r.Method != http.MethodDelete {
		model.JsonResponse(w, http.StatusMethodNotAllowed, "只支持DELETE方法", nil)
		return
	}

	// 获取当前用户
	currentUser, err := util.GetCurrentUser(r, h.UserService)
	if err != nil || currentUser == nil {
		model.JsonResponse(w, http.StatusUnauthorized, "未登录", nil)
		return
	}

	// 获取要删除的用户ID
	idStr := r.URL.Query().Get("id")
	targetUserID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		model.JsonResponse(w, http.StatusBadRequest, "无效的用户ID", nil)
		return
	}

	// 调用Service层执行删除
	err = h.UserService.DeleteUser(currentUser.ID, currentUser.Role, targetUserID)
	if err != nil {
		model.JsonResponse(w, http.StatusBadRequest, err.Error(), nil)
		return
	}

	model.JsonResponse(w, http.StatusOK, "删除成功", nil)
}

// ProfileHandler 用户个人中心处理
func (h *UserHandler) ProfileHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		model.JsonResponse(w, http.StatusMethodNotAllowed, "只支持 GET 方法", nil)
		return
	}

	//从context 获取 userID
	userID := util.GetUserID(r)
	if userID == 0 {
		model.JsonResponse(w, http.StatusUnauthorized, "未登录", nil)
		return
	}
	// 查询用户
	user, err := h.UserService.GetUserByID(userID)
	if err != nil {
		model.JsonResponse(w, http.StatusInternalServerError, "获取用户信息失败", nil)
		return
	}
	model.JsonResponse(w, http.StatusOK, "success", user)

}

// StatsHandler 首页统计数据
func (h *UserHandler) StatsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		model.JsonResponse(w, http.StatusMethodNotAllowed, "只支持 GET 方法", nil)
		return
	}

	stats, err := h.UserService.GetDashboardStats()
	if err != nil {
		model.JsonResponse(w, http.StatusInternalServerError, "获取统计数据失败", nil)
		return
	}

	model.JsonResponse(w, http.StatusOK, "success", stats)
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

// UpdateUserHandler 更新用户处理
func (h *UserHandler) UpdateUserHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		model.JsonResponse(w, http.StatusMethodNotAllowed, "只支持POST方法", nil)
		return
	}

	// 获取当前用户
	currentUser, err := util.GetCurrentUser(r, h.UserService)
	if err != nil || currentUser == nil {
		model.JsonResponse(w, http.StatusUnauthorized, "未登录", nil)
		return
	}

	err = r.ParseMultipartForm(10 << 20)
	if err != nil {
		model.JsonResponse(w, http.StatusBadRequest, "解析表单失败", nil)
		return
	}

	idStr := r.FormValue("id")
	username := r.FormValue("username")
	password := r.FormValue("password")
	email := r.FormValue("email")
	statusStr := r.FormValue("status")
	role := r.FormValue("role")

	id, _ := strconv.ParseInt(idStr, 10, 64)
	status, _ := strconv.Atoi(statusStr)

	var avatarURL string
	file, header, err := r.FormFile("avatar")
	if err == nil {
		defer file.Close()

		uploadDir := "web/static/uploads/avatar"
		if err := os.MkdirAll(uploadDir, 0755); err != nil {
			model.JsonResponse(w, http.StatusInternalServerError, "创建上传目录失败", nil)
			return
		}

		filename := fmt.Sprintf("%s/%d_%s", uploadDir, time.Now().Unix(), header.Filename)
		out, err := os.Create(filename)
		if err != nil {
			model.JsonResponse(w, http.StatusInternalServerError, "保存文件失败", nil)
			return
		}
		defer out.Close()
		io.Copy(out, file)
		avatarURL = "/static/uploads/avatar/" + fmt.Sprintf("%d_%s", time.Now().Unix(), header.Filename)
	}

	err = h.UserService.UpdateUser(currentUser.ID, currentUser.Role, id, username, password, email, status, avatarURL, role)
	if err != nil {
		model.JsonResponse(w, http.StatusBadRequest, err.Error(), nil)
		return
	}

	model.JsonResponse(w, http.StatusOK, "修改成功", nil)
}

// CreateUserHandler 创建用户处理
func (h *UserHandler) CreateUserHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		model.JsonResponse(w, http.StatusMethodNotAllowed, "只支持POST方法", nil)
		return
	}

	// 获取当前用户
	currentUser, err := util.GetCurrentUser(r, h.UserService)
	if err != nil || currentUser == nil {
		model.JsonResponse(w, http.StatusUnauthorized, "未登录", nil)
		return
	}

	// 解析表单
	err = r.ParseMultipartForm(10 << 20)
	if err != nil {
		model.JsonResponse(w, http.StatusBadRequest, "解析表单失败", nil)
		return
	}

	// 获取表单数据
	username := r.FormValue("username")
	password := r.FormValue("password")
	email := r.FormValue("email")
	role := r.FormValue("role")
	statusStr := r.FormValue("status")
	status, _ := strconv.Atoi(statusStr)

	// 处理头像上传
	var avatarURL string
	file, header, err := r.FormFile("avatar")
	if err == nil {
		defer file.Close()

		uploadDir := "web/static/uploads/avatar"
		if err := os.MkdirAll(uploadDir, 0755); err != nil {
			model.JsonResponse(w, http.StatusInternalServerError, "创建上传目录失败", nil)
			return
		}

		filename := fmt.Sprintf("%s/%d_%s", uploadDir, time.Now().Unix(), header.Filename)
		out, err := os.Create(filename)
		if err != nil {
			model.JsonResponse(w, http.StatusInternalServerError, "保存文件失败", nil)
			return
		}
		defer out.Close()
		io.Copy(out, file)
		avatarURL = "/static/uploads/avatar/" + fmt.Sprintf("%d_%s", time.Now().Unix(), header.Filename)
	}

	// 调用Service层创建用户
	_, err = h.UserService.CreateUser(currentUser.Role, username, password, email, avatarURL, role, status)
	if err != nil {
		model.JsonResponse(w, http.StatusBadRequest, err.Error(), nil)
		return
	}

	model.JsonResponse(w, http.StatusOK, "创建成功", nil)
}
