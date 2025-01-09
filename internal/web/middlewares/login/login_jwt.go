package login

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"go_web/internal/web"
	"net/http"
	"strings"
	"time"
)

type MiddlewareBuilderJWTLogin struct {
}

func (m *MiddlewareBuilderJWTLogin) CheckLogin() gin.HandlerFunc {

	return func(context *gin.Context) {
		path := context.Request.URL.Path
		if path == "/user/login" || path == "/user/signup" {
			return // 指定的路由不需要校验
		}

		// 根据约定,token在请求头中的Authorization中
		authCode := context.GetHeader("Authorization")
		if authCode == "" {
			context.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		segs := strings.Split(authCode, " ")
		// Authorization中的内容是无效的,乱传的
		if len(segs) != 2 {
			context.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		tokenString := segs[1]
		var uc web.UserClaims
		token, err := jwt.ParseWithClaims(tokenString, &uc, func(token *jwt.Token) (interface{}, error) {
			return web.JWTKey, nil
		})
		// 传递的token是伪造的,无法解析
		if err != nil {
			context.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		// 传递的token是非法的,或者已经过期的
		if !token.Valid {
			context.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		if uc.UserAgent != context.GetHeader("User-Agent") {
			// 后期接入监控告警的埋点区
			// 正常用户绝大多数情况不可能进入当前分支,进入当前分支的大概率是攻击者.
			context.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		expireTime := uc.ExpiresAt
		// token 已经过期了
		if expireTime.Before(time.Now()) {
			context.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		// 刷新Token,剩余过期实现小于XXX时间,则需要刷新
		if expireTime.Sub(time.Now()) < time.Minute*5 {
			uc.ExpiresAt = jwt.NewNumericDate(time.Now().Add(time.Minute * 30))
			tokenStr, err := token.SignedString(web.JWTKey)
			context.Header("x-jwt-token", tokenStr)
			if err != nil {
				// 不需要终止,仅过期时间刷新失败,但已经登录了
				fmt.Println(err)
			}

		}

	}
}
