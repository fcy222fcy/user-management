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

	// 创建路由
	mux := http.NewServeMux()

	// 注册路由
	handler.RegisterUserRoutes(mux, userSvc)

	// 注册静态资源(css,js,images)
	staticFS := http.FileServer(http.Dir("web/static"))
	mux.Handle("/static/", http.StripPrefix("/static/", staticFS))

	// 首页
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "web/templates/index.html")
	})

	// 中间件
	handlerWithMiddleware := middleware.CORS(mux)

	// 启动服务器
	port := os.Getenv("PORT")
	if port == "" {
		port = "8899"
	}
	log.Printf("系统已经准备好了,持续监听端口 %s ... \n", port)
	log.Printf("访问地址: https://localhost:%s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, handlerWithMiddleware))
	//log.Fatal(http.ListenAndServeTLS(":"+port, "D:\\server.crt", "D:\\server.key", handlerWithMiddleware))
}
