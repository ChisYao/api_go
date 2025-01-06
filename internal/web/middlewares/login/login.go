package login

import (
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"net/http"
)

type MiddlewareBuilderLogin struct {
}

func (m *MiddlewareBuilderLogin) CheckLogin() gin.HandlerFunc {
	return func(context *gin.Context) {
		path := context.Request.URL.Path
		if path == "/user/login" || path == "/user/signup" {
			return // 指定的路由不需要校验
		}

		sess := sessions.Default(context)
		if sess.Get("userId") == nil {
			// 中断执行
			context.AbortWithStatus(http.StatusUnauthorized)
		}
	}
}
