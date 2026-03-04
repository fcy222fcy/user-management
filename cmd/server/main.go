package main

import (
	"github.com/joho/godotenv"
	"log"
	"net/http"
	"os"
	"user-management/internal/handler"
	"user-management/internal/middleware"
	"user-management/internal/repository"
	"user-management/internal/service"
	"user-management/pkg/database"
)

func main() {
	// 1. 加载 .env
	if err := godotenv.Load(); err != nil {
		log.Println("没有发现 .env 文件")
	}
	// 2. 初始化数据库
	if err := database.InitDB(); err != nil {
		log.Fatal("数据库初始化失败:", err)
	}
	log.Println("数据库初始化成功")

	// 3. defer关闭连接
	defer database.DB.Close()

	// 初始化各层
	userRepo := repository.NewUserRepository(database.DB)
	userSvc := service.NewUserService(userRepo)

	// 注册路由
	mux := http.NewServeMux()

	// 注册静态文件服务--根路径
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "web/templates/register.html")
	})

	// 先注册API路由（优先匹配）
	handler.RegisterUserRoutes(mux, userSvc)

	// 静态文件服务
	mux.Handle("/index", http.FileServer(http.Dir("web/templates/index.html")))
	mux.Handle("/login", http.FileServer(http.Dir("web/templates/login.html")))
	mux.Handle("/userList", http.FileServer(http.Dir("web/templates/userList.html")))
	mux.Handle("/static/", http.FileServer(http.Dir("web/static")))

	// 应用 CORS 中间件到整个服务器
	handlerWithMiddleware := middleware.CORS(mux)

	// 启动服务器
	port := os.Getenv("PORT")
	if port == "" {
		port = "8899"
	}
	log.Printf("系统已经准备好了,持续监听端口 %s ... \n", port)
	log.Fatal(http.ListenAndServe(":"+port, handlerWithMiddleware))
}
