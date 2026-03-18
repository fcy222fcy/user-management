package handler

import (
	"log"
	"net/http"
	"strconv"
	"user-management/internal/model"
)

func (h *UserHandler) ListHandler(w http.ResponseWriter, r *http.Request) {

	pageStr := r.URL.Query().Get("page")
	page := 1

	if pageStr != "" {
		p, err := strconv.Atoi(pageStr)
		if err == nil && p > 0 {
			page = p
		}
	}
	pageSize := 10
	// 调用业务层
	users, totalCount, totalPages, err := h.UserService.GetUserList(page, pageSize)

	if err != nil {
		// 测试
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
