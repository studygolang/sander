// Copyright 2017 The StudyGolang Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// http://studygolang.com
// Author: polaris	polaris@studygolang.com

package app

import (
	xhttp "sander/http"
	"sander/http/internal/helper"
	"sander/logic"
	"sander/model"

	"github.com/labstack/echo"
)

// UserController .
type UserController struct{}

// RegisterRoute 注册路由
func (u UserController) RegisterRoute(g *echo.Group) {
	g.GET("/user/center", u.Center)
	g.GET("/user/me", u.Me)
	g.POST("/user/modify", u.Modify)
	g.POST("/user/login", u.Login)
}

// Center 用户自己个人中心
func (UserController) Center(ctx echo.Context) error {
	if user, ok := ctx.Get("user").(*model.Me); ok {
		data := map[string]interface{}{
			"user": user,
		}
		return success(ctx, data)
	}

	return success(ctx, nil)
}

// Me 用户信息
func (UserController) Me(ctx echo.Context) error {
	if me, ok := ctx.Get("user").(*model.Me); ok {
		user := logic.DefaultUser.FindOne(ctx, "uid", me.Uid)
		return success(ctx, map[string]interface{}{
			"user":            user,
			"default_avatars": logic.DefaultAvatars,
		})
	}

	return success(ctx, nil)
}

// Login .
func (UserController) Login(ctx echo.Context) error {
	if _, ok := ctx.Get("user").(*model.Me); ok {
		return success(ctx, nil)
	}

	username := ctx.FormValue("username")
	if username == "" {
		return fail(ctx, "用户名为空")
	}

	// 处理用户登录
	passwd := ctx.FormValue("passwd")
	userLogin, err := logic.DefaultUser.Login(ctx, username, passwd)
	if err != nil {
		return fail(ctx, err.Error())
	}

	data := map[string]interface{}{
		"token":    xhttp.GenToken(userLogin.Uid),
		"uid":      userLogin.Uid,
		"username": userLogin.Username,
	}
	return success(ctx, data)
}

// Modify .
func (UserController) Modify(ctx echo.Context) error {
	me, ok := ctx.Get("user").(*model.Me)
	if !ok {
		return fail(ctx, "请先登录", xhttp.NeedReLoginCode)
	}

	// 更新信息
	errMsg, err := logic.DefaultUser.Update(ctx, me, ctx.Request().FormParams())
	if err != nil {
		return fail(ctx, errMsg)
	}

	email := ctx.FormValue("email")
	if me.Email != email {
		boolHTTP := xhttp.CheckIsHttps(ctx)
		go logic.DefaultEmail.SendActivateMail(email, helper.RegActivateCode.GenUUID(email), boolHTTP)
	}

	return success(ctx, nil)
}
