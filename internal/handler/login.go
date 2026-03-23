package handler

import (
	"encoding/json"
	"net/http"
	"user-management/internal/model"
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

// LoginHandler 处理用户登录
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
	// 记录登录日志
	_ = h.UserService.Repo.CreateOperationLog(user.ID, user.Username, "login")

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
