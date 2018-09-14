// Copyright 2013 The StudyGolang Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// http://studygolang.com
// Author: polaris	polaris@studygolang.com

package admin

import (
	"sander/logic"

	"github.com/labstack/echo"
	"github.com/polaris1119/goutils"
)

// UserController .
type UserController struct{}

// RegisterRoute 注册路由
func (u UserController) RegisterRoute(g *echo.Group) {
	g.GET("/user/user/list", u.UserList)
	g.POST("/user/user/query.html", u.UserQuery)
	g.GET("/user/user/detail", u.Detail)
	g.POST("/user/user/modify", u.Modify)
}

// UserList 所有用户（分页）
func (UserController) UserList(ctx echo.Context) error {
	curPage, limit := parsePage(ctx)

	users, total := logic.DefaultUser.FindUserByPage(ctx, nil, curPage, limit)

	data := map[string]interface{}{
		"datalist":   users,
		"total":      total,
		"totalPages": (total + limit - 1) / limit,
		"page":       curPage,
		"limit":      limit,
	}

	return render(ctx, "user/list.html,user/query.html", data)
}

// UserQuery .
func (UserController) UserQuery(ctx echo.Context) error {
	curPage, limit := parsePage(ctx)
	conds := parseConds(ctx, []string{"uid", "username", "email"})

	users, total := logic.DefaultUser.FindUserByPage(ctx, conds, curPage, limit)

	data := map[string]interface{}{
		"datalist":   users,
		"total":      total,
		"totalPages": (total + limit - 1) / limit,
		"page":       curPage,
		"limit":      limit,
	}

	return renderQuery(ctx, "user/query.html", data)
}

// Detail .
func (UserController) Detail(ctx echo.Context) error {
	user := logic.DefaultUser.FindOne(ctx, "uid", ctx.QueryParam("uid"))

	data := map[string]interface{}{
		"user": user,
	}

	return render(ctx, "user/detail.html", data)
}

// Modify .
func (UserController) Modify(ctx echo.Context) error {
	uid := ctx.FormValue("uid")

	amount := goutils.MustInt(ctx.FormValue("amount"))
	if amount > 0 {
		logic.DefaultUserRich.Recharge(ctx, uid, ctx.FormParams())
	} else {
		logic.DefaultUser.SetDauAuth(ctx, uid, ctx.FormParams())
	}
	return success(ctx, nil)
}
