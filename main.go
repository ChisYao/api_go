package main

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/redis"
	"github.com/gin-gonic/gin"
	"go_web/internal/repository"
	"go_web/internal/repository/dao"
	"go_web/internal/service"
	"go_web/internal/web"
	"go_web/internal/web/middlewares/login"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"net/http"
	"time"
)

func main() {
	// 初始化表格.数据库驱动
	//db := initDB()
	// 初始化引擎,加载middleware
	engine := initEngine()

	engine.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "hello world!")
	})
	// 注册校验方式
	//initNormalLoginCheck(engine)

	// 加载User区块的配置信息
	//initUserHandler(db, engine)
	// 引擎挂载端口
	engine.Run(":8080")
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

// 初始化驱动引擎
func initEngine() *gin.Engine {
	engine := gin.Default()

	// 加入cors[跨域处理]的middleware.
	engine.Use(cors.New(cors.Config{
		//AllowMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD"}, // 允许通过的请求方法类型.
		//AllowOrigins:     []string{"https://foo.com"}, // 允许通过的请求源.
		AllowHeaders:  []string{"Content-Type", "Authorization"}, // 允许通过的请求头.
		ExposeHeaders: []string{"x-jwt-token"},                   // 允许前台访问的后台设置的请求头
		//AllowCredentials: true, // 请求时是否允许附带信息(例:cookies)
		AllowOriginFunc: func(origin string) bool {
			return true // 通过函数处理
		},
		MaxAge: 12 * time.Hour, // 预检请求有效期限.
	}))

	return engine
}

// 加载校验方式(Redis/Memstore...)
func initNormalLoginCheck(engine *gin.Engine) {
	login := &login.MiddlewareBuilderLogin{}
	// 直接存储在Cookie中(不建议)
	//store := cookie.NewStore([]byte("secret"))

	// 变更 memstore 存储方式(单实例部署)
	//store := memstore.NewStore([]byte("iyGFLRg6BaBOKbbMaTldalPWn3RaS1r8oACtwaT4IrraopfBWQ095paBVUgr37UV"),
	//	[]byte("RSNRR1Reo827M3QzlG5Epk2ckgSlfESPupteSuXN1mFoIeQZhiVGT2XuymaEAS6h"))

	// 变更redis存储方式(多实例部署)
	// size:最大空闲链接数量, network:链接方式 TCP/UDP, address:链接地址, password:链接密码
	store, err := redis.NewStore(16, "tcp", "localhost:6379", "",
		// authentication key: 身份认证密钥
		[]byte("iyGFLRg6BaBOKbbMaTldalPWn3RaS1r8oACtwaT4IrraopfBWQ095paBVUgr37UV"),
		// encryption key: 数据加密密钥
		[]byte("RSNRR1Reo827M3QzlG5Epk2ckgSlfESPupteSuXN1mFoIeQZhiVGT2XuymaEAS6h"))
	if err != nil {
		panic("Redis初始化错误!")
	}
	// 信息安全的三个核心概念: 身份认证,数据加密,授权

	engine.Use(sessions.Sessions("ssid", store), login.CheckLogin())
}

// 返回WEB驱动引擎,具备JWT的校验方式 | 限流
func initJWTLoginCheck(engine *gin.Engine) {

	//redisClient, _ := redis.NewStore(16, "tcp", "localhost:6379", "",
	//	// authentication key: 身份认证密钥
	//	[]byte("iyGFLRg6BaBOKbbMaTldalPWn3RaS1r8oACtwaT4IrraopfBWQ095paBVUgr37UV"),
	//	// encryption key: 数据加密密钥
	//	[]byte("RSNRR1Reo827M3QzlG5Epk2ckgSlfESPupteSuXN1mFoIeQZhiVGT2XuymaEAS6h"))

	//engine.Use(middleware.NewBuilder(redisClient, time.Second, 100).Build())

	login := &login.MiddlewareBuilderJWTLogin{}
	engine.Use(login.CheckLogin())
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
