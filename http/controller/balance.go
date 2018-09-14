// Copyright 2017 The StudyGolang Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// http://studygolang.com
// Author: polaris	polaris@studygolang.com

package controller

import (
	"sander/http/middleware"
	"sander/logic"
	"sander/model"

	"github.com/labstack/echo"
)

type UserRichController struct{}

// 注册路由
func (u UserRichController) RegisterRoute(g *echo.Group) {
	g.Get("/balance", u.MyBalance, middleware.NeedLogin())
	g.Get("/balance/add", u.Add, middleware.NeedLogin())
}

func (UserRichController) MyBalance(ctx echo.Context) error {
	me := ctx.Get("user").(*model.Me)
	balanceDetails := logic.DefaultUserRich.FindBalanceDetail(ctx, me)

	data := map[string]interface{}{
		"details": balanceDetails,
	}
	return render(ctx, "rich/balance.html", data)
}

func (UserRichController) Add(ctx echo.Context) error {
	me := ctx.Get("user").(*model.Me)
	balanceDetails := logic.DefaultUserRich.FindBalanceDetail(ctx, me, model.MissionTypeAdd)

	rechargeAmount := logic.DefaultUserRich.FindRecharge(ctx, me)

	data := map[string]interface{}{
		"details":         balanceDetails,
		"recharge_amount": rechargeAmount,
	}
	return render(ctx, "rich/add.html", data)
}
