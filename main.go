package main

import (
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/redis"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"github.com/liliang-cn/webook/config"
	"github.com/liliang-cn/webook/internal/repository"
	"github.com/liliang-cn/webook/internal/repository/dao"
	"github.com/liliang-cn/webook/internal/service"
	"github.com/liliang-cn/webook/internal/web"
	"github.com/liliang-cn/webook/internal/web/middleware"
)

func main() {
	db := initDB()
	server := initWebServer()

	u := initUser(db)
	u.RegisterRoutesV1(server.Group("/users"))

	server.Run(":8080")
}

func initWebServer() *gin.Engine {
	server := gin.Default()
	server.Use(cors.New(cors.Config{
		AllowMethods:     []string{"PUT", "PATCH"},
		AllowHeaders:     []string{"Content-Type", "Authorization"},
		ExposeHeaders:    []string{"x-jwt-token"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
		AllowOriginFunc: func(origin string) bool {
			if strings.HasPrefix(origin, "http://localhost") {
				return true
			}
			return strings.Contains(origin, "webook.com")
		},
	}))

	// init session store
	store, err := redis.NewStore(16, "tcp", config.Config.Redis.Addr, "", []byte("secret"))
	if err != nil {
		panic(err)
	}

	server.Use(sessions.Sessions("mysession", store))

	server.Use(middleware.NewLoginJWTMiddlewareBuilder().
		IgnorePaths("/users/status").
		IgnorePaths("/users/signup").
		IgnorePaths("/users/login").
		Build())

	return server
}

func initUser(db *gorm.DB) *web.UserHandler {
	ud := dao.NewUserDAO(db)
	ur := repository.NewUserRepository(ud)
	us := service.NewUserService(ur)
	u := web.NewUserHandler(us)

	return u
}

func initDB() *gorm.DB {
	db, err := gorm.Open(mysql.Open(config.Config.DB.DSN))

	if err != nil {
		panic(err)
	}

	err = dao.InitTables(db)

	if err != nil {
		panic(err)
	}
	return db
}
