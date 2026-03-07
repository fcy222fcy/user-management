package handler

import (
	"html/template"
	"log"
	"net/http"
	"strconv"
	"user-management/pkg/response"
)

func ListHandler(w http.ResponseWriter, r *http.Request) {

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
	users, totalCount, totalPages, err := UserSvc.GetUserList(page, pageSize)

	if err != nil {
		// 测试
		log.Println("获取用户列表失败:", err)
		http.Error(w, "获取用户列表失败", http.StatusInternalServerError)
		return
	}

	t, err := template.ParseFiles("web/templates/userList.html")
	if err != nil {
		// 测试
		log.Println("模版解析失败:", err)
		response.JsonResponse(w, http.StatusInternalServerError, "页面渲染失败", nil)
		return
	}
	data := map[string]interface{}{
		"Users":      users,
		"Page":       page,
		"TotalCount": totalCount,
		"TotalPages": totalPages,
		// 暂时写死,开发完上传头像功能再做修改
		"NowAvatar": "/static/images/default-avatar.png",
		"NowRole":   "管理员",

		"StartPage": 1,
		"EndPage":   totalPages,
	}
	err = t.Execute(w, data)
	if err != nil {
		log.Println("模版执行失败:", err)
		http.Error(w, "页面渲染失败", 500)
		return
	}
}
