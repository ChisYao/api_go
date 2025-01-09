package login

import (
	"encoding/gob"
	"fmt"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
)

type MiddlewareBuilderLogin struct {
}

func (m *MiddlewareBuilderLogin) CheckLogin() gin.HandlerFunc {
	// 注册类型
	gob.Register(time.Now())
	// 返回Middleware
	return func(context *gin.Context) {
		path := context.Request.URL.Path
		if path == "/user/login" || path == "/user/signup" {
			return // 指定的路由不需要校验
		}

		sess := sessions.Default(context)
		userId := sess.Get("userId")
		if userId == nil {
			// 中断执行
			context.AbortWithStatus(http.StatusUnauthorized)
		}

		// 对于Session进行刷新,定义刷新策略
		now := time.Now()
		const UpdateTimeKey = "update_time"
		val := sess.Get(UpdateTimeKey)
		// 最后的操作时间
		lastUpdateTime, ok := val.(time.Time)
		// 设置的刷新标记不存在,表明:第一次进入系统 || ... || 距离上次操作超过一定时间(一分钟)
		if val == nil || !ok || now.Sub(lastUpdateTime) > time.Minute {
			sess.Set(UpdateTimeKey, now)
			sess.Set("userId", userId)
			err := sess.Save()
			if err != nil {
				// 日志输出
				fmt.Println(err)
			}
		}

	}
}
