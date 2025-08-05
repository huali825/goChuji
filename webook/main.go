package main

import (
	"context"
	"crypto/rand"
	"fmt"

	"strings"
	"time"

	"github.com/gin-contrib/cors"
	sredis "github.com/gin-contrib/sessions/redis"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"gochuji/webook/internal/repository"
	"gochuji/webook/internal/repository/cache"
	"gochuji/webook/internal/repository/dao"
	"gochuji/webook/internal/service"
	"gochuji/webook/internal/web"
	"gochuji/webook/internal/web/middleware"
)

func main() {
	db := initDB()
	rdb, err := initRedis("81.71.139.129:6379", "qsgctys711!@#", 0)
	if err != nil {
		panic(err)
	}
	server := initWebServer()
	initUserHdl(db, rdb, server)
	err = server.Run(":8080")
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

func initUserHdl(db *gorm.DB, rdb redis.Cmdable, server *gin.Engine) {
	ud := dao.NewUserDAO(db)
	uredis := cache.NewUserCache(rdb)
	ur := repository.NewUserRepository(ud, uredis)
	us := service.NewUserService(ur)
	hdl := web.NewUserHandler(us)
	hdl.RegisterRoutes(server)
}

func initWebServer() *gin.Engine {
	server := gin.Default()

	server.Use(
		cors.New(cors.Config{
			//AllowAllOrigins:  true,                              //允许所有来源
			//AllowOrigins:     []string{"http://localhost:3000"}, //允许跨域请求的域名
			AllowCredentials: true, //允许跨域请求携带cookie

			AllowHeaders: []string{"Content-Type", "Authorization"}, //允许跨域请求携带的header
			//AllowHeaders: []string{"Content-Type"}, //允许跨域请求携带的header

			//AllowMethods: []string{"POST"},				//允许跨域请求的方法
			ExposeHeaders: []string{"x-jwt-token"},

			//自定义的校验规则
			AllowOriginFunc: func(origin string) bool {
				if strings.HasPrefix(origin, "http://localhost") {
					//if strings.Contains(origin, "localhost") {
					return true
				}
				return strings.Contains(origin, "your_company.com")
			},
			MaxAge: 12 * time.Hour,
		}),
		func(ctx *gin.Context) {
			println("这是我的 Middleware 01")
		})

	server.Use(func(ctx *gin.Context) {
		println("这是我的 Middleware 02")
	})

	//// 创建一个cookie存储，使用"secret"作为密钥
	//store := cookie.NewStore([]byte("secret"))
	//// 使用sessions中间件，将"ssid"作为session的名称，使用store作为存储
	//server.Use(sessions.Sessions("ssid", store))

	//// session生成和校验， 基于redis的
	//store := sessionStore()
	//server.Use(sessions.Sessions("ssid", store))
	//login := &middleware.LoginMiddlewareBuilder{}
	//server.Use(login.CheckLogin())

	//jwt login 校验
	login := middleware.LoginJWTMiddlewareBuilder{}
	server.Use(login.CheckLogin())

	return server
}

func initRedis(addr, password string, db int) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	// 测试连接
	_, err := client.Ping(context.Background()).Result()
	if err != nil {
		return nil, fmt.Errorf("无法连接到Redis: %v", err)
	}

	return client, nil
}

// sessionStore 函数用于生成一个redis.Store对象
func sessionStore() sredis.Store {
	// 生成32字节随机认证密钥
	authKey := make([]byte, 32)
	_, _ = rand.Read(authKey) // 需导入 "crypto/rand" 包
	// 生成16字节随机加密密钥
	encKey := make([]byte, 16)
	_, _ = rand.Read(encKey)

	// 创建一个redis.Store对象，参数分别为：最大连接数、地址、密码、认证密钥、加密密钥
	store, err := sredis.NewStore(10,
		"tcp", "81.71.139.129:6379",
		"", "qsgctys711!@#",
		authKey,
		encKey)
	if err != nil {
		panic(err)
	}
	return store
}
