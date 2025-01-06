package main

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"go_web/internal/repository"
	"go_web/internal/repository/dao"
	"go_web/internal/service"
	"go_web/internal/web"
	"go_web/internal/web/middlewares/login"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"time"
)

func main() {
	// 初始化表格.数据库驱动
	db := initDB()
	// 初始化引擎,加载middleware
	engine := initEngine()
	// 加载User区块的配置信息
	initUserHandler(db, engine)
	// 引擎挂载端口
	engine.Run(":8081")
}

// 初始化数据库链接驱动,创建表格,返回驱动对象
func initDB() *gorm.DB {
	// Docker中数据库 => U:root,P:root,PT:13316
	dsn := "root:root@tcp(localhost:13316)/mydb"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("数据库驱动加载错误!")
	}
	err = dao.InitTable(db)
	if err != nil {
		panic(err)
	}
	return db
}

// 返回WEB驱动引擎
func initEngine() *gin.Engine {
	engine := gin.Default()

	// 加入cors[跨域处理]的middleware.
	engine.Use(cors.New(cors.Config{
		//AllowMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD"}, // 允许通过的请求方法类型.
		//AllowOrigins:     []string{"https://foo.com"}, // 允许通过的请求源.
		//AllowHeaders:     []string{"Origin"}, // 允许通过的请求头.
		//AllowCredentials: true, // 请求时是否允许附带信息(例:cookies)
		AllowOriginFunc: func(origin string) bool {
			return true // 通过函数处理
		},
		MaxAge: 12 * time.Hour, // 预检请求有效期限.
	}))

	// 加入登录状态校验

	login := &login.MiddlewareBuilderLogin{}
	// 直接存储在Cookie中(不建议)
	store := cookie.NewStore([]byte("secret"))
	engine.Use(sessions.Sessions("ssid", store), login.CheckLogin())
	return engine
}

// 加载User区块相关路由
func initUserHandler(db *gorm.DB, engine *gin.Engine) {
	ud := dao.NewUserDao(db)
	ur := repository.NewUserRepository(ud)
	us := service.NewUserService(ur)

	userHandler := web.NewUserHandler(us)

	// 注册UserHandler中定义的路由.
	userHandler.RegistryRoutes(engine)

}
