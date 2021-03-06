// Copyright 2014 The StudyGolang Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// http://studygolang.com
// Author: polaris	polaris@studygolang.com

package admin

import (
	"sander/logic"

	"github.com/labstack/echo"
)

// SettingController .
type SettingController struct{}

// RegisterRoute 注册路由
func (s SettingController) RegisterRoute(g *echo.Group) {
	g.Match([]string{"GET", "POST"}, "/setting/genneral/modify", s.GenneralModify)
	g.Match([]string{"GET", "POST"}, "/setting/nav/modify", s.NavModify)
	g.Match([]string{"GET", "POST"}, "/setting/index_tab/children", s.IndexTabChildren)
}

// GenneralModify 常规选项修改
func (s SettingController) GenneralModify(ctx echo.Context) error {
	if ctx.FormValue("submit") == "1" {
		err := logic.DefaultSetting.Update(ctx, ctx.FormParams())
		if err != nil {
			return fail(ctx, 1, err.Error())
		}

		return success(ctx, nil)
	}

	return render(ctx, "setting/genneral.html", nil)
}

// NavModify 菜单、导航修改
func (s SettingController) NavModify(ctx echo.Context) error {
	if ctx.FormValue("submit") == "1" {
		err := logic.DefaultSetting.Update(ctx, ctx.FormParams())
		if err != nil {
			return fail(ctx, 1, err.Error())
		}

		return success(ctx, nil)
	}
	return render(ctx, "setting/menu_nav.html", nil)
}

// IndexTabChildren .
func (s SettingController) IndexTabChildren(ctx echo.Context) error {
	if ctx.FormValue("submit") == "1" {
		err := logic.DefaultSetting.UpdateIndexTabChildren(ctx, ctx.FormParams())
		if err != nil {
			return fail(ctx, 1, err.Error())
		}

		return success(ctx, nil)
	}

	tab := ctx.QueryParam("tab")
	name := ctx.QueryParam("name")

	return render(ctx, "setting/index_tab.html", map[string]interface{}{"tab": tab, "name": name})
}
