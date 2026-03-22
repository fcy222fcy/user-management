package handler

import (
	"net/http"
	"user-management/internal/model"
)

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
