package main

import (
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"gochuji/webook/internal/repository"
	"gochuji/webook/internal/repository/dao"
	"gochuji/webook/internal/service"
	"gochuji/webook/internal/web"
	"gochuji/webook/internal/web/middleware"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func main() {
	db := initDB()
	server := initWebServer()
	initUserHdl(db, server)
	err := server.Run(":8080")
	if err != nil {
		panic(err)
	}
}

func initDB() *gorm.DB {
	//db, err := gorm.Open(mysql.Open("root:root@tcp(localhost:13316)/webook"))
	db, err := gorm.Open(mysql.Open("root:qsgctys711@tcp(81.71.139.129:3306)/webook"))
	if err != nil {
		panic(err)
	}

	err = dao.InitTables(db)
	if err != nil {
		panic(err)
	}
	return db
}

func initUserHdl(db *gorm.DB, server *gin.Engine) {
	ud := dao.NewUserDAO(db)
	ur := repository.NewUserRepository(ud)
	us := service.NewUserService(ur)
	hdl := web.NewUserHandler(us)
	hdl.RegisterRoutes(server)
}

func initWebServer() *gin.Engine {
	server := gin.Default()

	//server.Use(cors.New(cors.Config{
	//	//AllowAllOrigins:  true,                              //允许所有来源
	//	//AllowOrigins:     []string{"http://localhost:3000"}, //允许跨域请求的域名
	//	AllowCredentials: true,                     //允许跨域请求携带cookie
	//	AllowHeaders:     []string{"Content-Type"}, //允许跨域请求携带的header
	//	//AllowMethods: []string{"POST"},				//允许跨域请求的方法
	//
	//	//自定义的校验规则
	//	AllowOriginFunc: func(origin string) bool {
	//		if strings.HasPrefix(origin, "http://localhost") {
	//			//if strings.Contains(origin, "localhost") {
	//			return true
	//		}
	//		return strings.Contains(origin, "your_company.com")
	//	},
	//	MaxAge: 12 * time.Hour,
	//}),
	//func(ctx *gin.Context) {
	//	println("这是我的 Middleware 01")
	//})

	server.Use(func(ctx *gin.Context) {
		println("这是我的 Middleware 02")
	})

	login := &middleware.LoginMiddlewareBuilder{}
	// 存储数据的，也就是你 userId 存哪里
	// 直接存 cookie
	store := cookie.NewStore([]byte("secret"))

	server.Use(sessions.Sessions("ssid", store), login.CheckLogin())
	return server
}
