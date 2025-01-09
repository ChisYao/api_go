package web

import (
	regexp "github.com/dlclark/regexp2"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	jwt "github.com/golang-jwt/jwt/v5"
	"go_web/internal/domain"
	"go_web/internal/service"
	"net/http"
	"time"
	//"regexp"
)

const (
	// 定义正则表达式
	emailRegexPattern    = "^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+.[a-zA-Z]{2,}$"
	passwordRegexPattern = "^(?=.*[a-z])(?=.*[A-Z])(?=.*\\d)[a-zA-Z\\d]{8,32}$"
)

type UserHandler struct {
	emailRegExp    *regexp.Regexp
	passwordRegExp *regexp.Regexp
	svc            *service.UserService
}

func NewUserHandler(svc *service.UserService) *UserHandler {
	return &UserHandler{
		// 性能优化:通过预编译正则表达式的手段,提高校验速度.
		emailRegExp:    regexp.MustCompile(emailRegexPattern, regexp.None),
		passwordRegExp: regexp.MustCompile(passwordRegexPattern, regexp.None),
		svc:            svc,
	}
}

func (h *UserHandler) RegistryRoutes(server *gin.Engine) {
	// 使用路由组
	group := server.Group("/user")
	// 分散注册路由
	group.POST("/signup", h.SignUp)
	//group.POST("/login", h.Login)
	group.POST("/login", h.LoginWithJWT)
	group.POST("/edit", h.Edit)
	group.GET("/profile", h.Profile)

}

func (h *UserHandler) SignUp(context *gin.Context) {
	// 请求参数
	type SignUpReq struct {
		Email         string `json:"email"`
		Password      string `json:"password"`
		CheckPassword string `json:"checkPassword"`
	}

	var req SignUpReq
	// bind方法会根据传入内容自动解析,若出现问题会自动处理返回400状态.
	if err := context.Bind(&req); err != nil {
		return
	}

	// 正则表达式-校验邮箱
	isEmail, err := h.emailRegExp.MatchString(req.Email)
	if !isEmail {
		context.String(http.StatusOK, "邮箱格式并不合法")
		return
	}

	// 正则表达式-校验密码
	isPassword, err := h.passwordRegExp.MatchString(req.Password)
	if !isPassword {
		context.String(http.StatusOK, "密码格式不正确")
		return
	}

	// 调用Repository层函数
	err = h.svc.SignUp(context, domain.User{
		Email:    req.Email,
		Password: req.Password,
	})

	switch err {
	case nil:
		context.String(http.StatusOK, "账户注册成功!")
	case service.ErrorDuplicateEmail:
		context.String(http.StatusOK, "注册邮箱已经存在,请更换后重试!")
	default:
		context.String(http.StatusOK, "系统错误")
	}
}

func (h *UserHandler) Login(context *gin.Context) {
	// 请求参数
	type LoginReq struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	var req LoginReq
	// 解析请求参数
	if err := context.Bind(&req); err != nil {
		return
	}

	// 调用Repository层函数
	u, err := h.svc.Login(context, req.Email, req.Password)

	switch err {
	case nil:
		// 加载缓存
		sess := sessions.Default(context)
		// 设置缓存
		sess.Set("userId", u.Id)
		// 缓存相关配置
		sess.Options(sessions.Options{
			MaxAge: 15 * 60, // 有效时间.
			//HttpOnly: true,
		})
		// 保存缓存
		err = sess.Save()
		if err != nil {
		}
		context.String(http.StatusOK, "登录成功!")
	case service.ErrorInvalidEmailOrPassword:
		context.String(http.StatusOK, "邮箱名称或密码不正确,请更改后重试!")
	default:
		context.String(http.StatusOK, "系统错误!")
	}
}

func (h *UserHandler) Edit(context *gin.Context) {
	sess := sessions.Default(context)
	userId := sess.Get("userId")

	// 请求参数
	type EditReq struct {
		Name   string `json:"name"`
		Gender string `json:"gender"`
		Phone  string `json:"phone"`
	}

	var req EditReq
	// 解析请求参数
	if err := context.Bind(&req); err != nil {
		return
	}

	err := h.svc.Edit(context, userId.(int64), domain.User{
		Name:   req.Name,
		Gender: req.Gender,
		Phone:  req.Phone,
	})

	switch err {
	case nil:
		context.String(http.StatusOK, "操作成功!")
	default:
		context.String(http.StatusOK, "系统错误!")
	}

}

func (h *UserHandler) Profile(context *gin.Context) {
	sess := sessions.Default(context)
	userId := sess.Get("userId")

	u, err := h.svc.Profile(context, userId.(int64))

	type responseData struct {
		Id     int64
		Email  string
		Name   string
		Phone  string
		Gender string
	}

	switch err {
	case nil:
		context.JSON(http.StatusOK, gin.H{
			"data": &responseData{Id: u.Id, Email: u.Email, Name: u.Name, Phone: u.Phone, Gender: u.Gender},
		})
	default:
		context.String(http.StatusOK, "系统错误!")
	}

}

func (h *UserHandler) LoginWithJWT(context *gin.Context) {
	// 请求参数
	type LoginReq struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	var req LoginReq
	// 解析请求参数
	if err := context.Bind(&req); err != nil {
		return
	}

	// 调用Repository层函数
	u, err := h.svc.Login(context, req.Email, req.Password)

	switch err {
	case nil:
		// 使用JWT的方案
		uc := UserClaims{
			Uid:       u.Id,
			UserAgent: context.GetHeader("User-Agent"),
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute * 30)), // 过期时间30分钟
			},
		}
		token := jwt.NewWithClaims(jwt.SigningMethodHS512, uc)
		tokenStr, err := token.SignedString([]byte(JWTKey))

		if err != nil {
			context.String(http.StatusOK, "系统错误!")
			return
		}
		context.Header("x-jwt-token", tokenStr)
		context.String(http.StatusOK, "登录成功!")
	case service.ErrorInvalidEmailOrPassword:
		context.String(http.StatusOK, "邮箱名称或密码不正确,请更改后重试!")
	default:
		context.String(http.StatusOK, "系统错误!")
	}
}

var JWTKey = []byte("iyGFLRg6BaBOKbbMaTldalPWn3RaS1r0oACtwaT4IrraopfBWQ095paBVUgr88UV")

type UserClaims struct {
	jwt.RegisteredClaims
	Uid       int64
	UserAgent string
}
