// Copyright 2017 The StudyGolang Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// http://studygolang.com
// Author: polaris	polaris@studygolang.com

package admin

import (
	"expvar"
	"net/http"
	"strconv"
	"time"

	"sander/global"
	xhttp "sander/http"
	"sander/logic"

	"github.com/labstack/echo"
	"github.com/polaris1119/goutils"
)

var (
	onlineStats    = expvar.NewMap("online_stats")
	loginUserNum   expvar.Int
	visitorUserNum expvar.Int
)

// MetricsController  .
type MetricsController struct{}

// RegisterRoute 注册路由
func (m MetricsController) RegisterRoute(g *echo.Group) {
	g.GET("/debug/vars", m.DebugExpvar)
	g.GET("/user/is_online", m.IsOnline)
}

// DebugExpvar .
func (m MetricsController) DebugExpvar(ctx echo.Context) error {
	loginUserNum.Set(int64(logic.Book.LoginLen()))
	visitorUserNum.Set(int64(logic.Book.Len()))

	onlineStats.Set("login_user_num", &loginUserNum)
	onlineStats.Set("visitor_user_num", &visitorUserNum)
	onlineStats.Set("uptime", expvar.Func(m.calculateUptime))
	onlineStats.Set("login_user_data", logic.Book.LoginUserData())

	handler := expvar.Handler()
	handler.ServeHTTP(xhttp.ResponseWriter(ctx), xhttp.Request(ctx))
	return nil
}

// IsOnline .
func (m MetricsController) IsOnline(ctx echo.Context) error {
	uid := goutils.MustInt(ctx.FormValue("uid"))
	onlineInfo := map[string]int{"online": logic.Book.Len()}
	message := logic.NewMessage(logic.WsMsgOnline, onlineInfo)
	logic.Book.PostMessage(uid, message)
	return ctx.HTML(http.StatusOK, strconv.FormatBool(logic.Book.UserIsOnline(uid)))
}

func (m MetricsController) calculateUptime() interface{} {
	return time.Since(global.App.LaunchTime).String()
}
