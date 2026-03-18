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
		log.Fatalf("数据库初始化失败: %v", err)
	}
	log.Println("数据库初始化成功")

	// 3. defer关闭连接
	defer database.DB.Close()

	// 初始化各层
	userRepo := repository.NewUserRepository(database.DB)
	userService := service.NewUserService(userRepo)
	userHandler := handler.NewUserHandler(userService)

	// 创建路由
	mux := http.NewServeMux()

	// 注册路由
	handler.RegisterUserRoutes(mux, userHandler)

	// 注册静态资源(css,js,images)
	staticFS := http.FileServer(http.Dir("web/static"))
	mux.Handle("/static/", http.StripPrefix("/static/", staticFS))

	// 中间件
	handlerWithMiddleware := middleware.CORS(mux)

	// 启动服务器
	port := os.Getenv("PORT")
	if port == "" {
		port = "8899"
	}
	log.Printf("系统已经准备好了,持续监听端口 %s ...", port)
	log.Printf("访问地址: https://localhost:%s", port)
	log.Fatal(http.ListenAndServe(":"+port, handlerWithMiddleware))
	//log.Fatal(http.ListenAndServeTLS(":"+port, "D:\\server.crt", "D:\\server.key", handlerWithMiddleware))
}
