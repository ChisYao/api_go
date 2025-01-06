package web

import (
	regexp "github.com/dlclark/regexp2"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"go_web/internal/domain"
	"go_web/internal/service"
	"net/http"
	//"regexp"
)

const (
	// 定义正则表达式
	emailRegexPattern    = ""
	passwordRegexPattern = ""
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
	group.POST("/login", h.Login)
	group.POST("/edit", h.Edit)
	group.GET("/profile", h.Profile)

}

func (h *UserHandler) SignUp(context *gin.Context) {
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
	// 正则表达式校验
	isEmail, err := h.emailRegExp.MatchString(req.Email)
	if !isEmail {
		context.String(http.StatusOK, "邮箱格式并不合法")
		return
	}

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
	type LoginReq struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	var req LoginReq
	if err := context.Bind(&req); err != nil {
		context.String(http.StatusOK, "系统错误!")
		return
	}

	u, err := h.svc.Login(context, req.Email, req.Password)

	switch err {
	case nil:
		sess := sessions.Default(context)

		sess.Set("userId", u)
		sess.Options(sessions.Options{
			MaxAge: 15 * 60,
			//HttpOnly: true,
		})
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

}
func (h *UserHandler) Profile(context *gin.Context) {

}
