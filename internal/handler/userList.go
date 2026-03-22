package handler

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
	"user-management/internal/model"
	"user-management/pkg/util"
)

// ListHandler 获取用户列表数据
func (h *UserHandler) ListHandler(w http.ResponseWriter, r *http.Request) {
	// 获取分页参数
	pageStr := r.URL.Query().Get("page")
	page := 1
	if pageStr != "" {
		p, err := strconv.Atoi(pageStr)
		if err == nil && p > 0 {
			page = p
		}
	}
	pageSize := 10

	// 获取搜索参数
	keyword := r.URL.Query().Get("keyword")
	statusStr := r.URL.Query().Get("status")

	var users []model.User
	var totalCount, totalPages int
	var err error

	// 如果有搜索条件，使用搜索功能
	if keyword != "" || statusStr != "" {
		status := -1 // -1 表示不筛选状态
		if statusStr != "" {
			s, _ := strconv.Atoi(statusStr)
			status = s
		}
		users, totalCount, totalPages, err = h.UserService.SearchUsers(keyword, status, page, pageSize)
	} else {
		// 否则获取全部列表
		users, totalCount, totalPages, err = h.UserService.GetUserList(page, pageSize)
	}

	if err != nil {
		log.Println("获取用户列表失败:", err)
		http.Error(w, "获取用户列表失败", http.StatusInternalServerError)
		return
	}

	model.JsonResponse(w, http.StatusOK, "success", map[string]interface{}{
		"users":       users,
		"page":        page,
		"total_count": totalCount,
		"total_pages": totalPages,
	})
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

// DeleteUserHandler 删除用户处理
func (h *UserHandler) DeleteUserHandler(w http.ResponseWriter, r *http.Request) {

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

	// 先获取被删除用户的信息（用于记录日志）
	targetUser, _ := h.UserService.GetUserByID(targetUserID)

	// 调用Service层执行删除
	err = h.UserService.DeleteUser(currentUser.ID, currentUser.Role, targetUserID)
	if err != nil {
		model.JsonResponse(w, http.StatusBadRequest, err.Error(), nil)
		return
	}

	// 删除成功就记录日志
	if targetUser != nil {
		_ = h.UserService.Repo.CreateOperationLog(targetUserID, targetUser.Username, "delete")
	}

	model.JsonResponse(w, http.StatusOK, "删除成功", nil)
}
