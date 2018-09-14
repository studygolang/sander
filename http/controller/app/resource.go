// Copyright 2017 The StudyGolang Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// http://studygolang.com
// Author: polaris	polaris@studygolang.com

package app

import (
	xhttp "sander/http"
	"sander/logic"
	"sander/model"

	"github.com/labstack/echo"
	"github.com/polaris1119/goutils"
)

// ResourceController .
type ResourceController struct{}

// RegisterRoute 注册路由
func (r ResourceController) RegisterRoute(g *echo.Group) {
	g.GET("/resources", r.ReadList)
	g.GET("/resource/detail", r.Detail)
}

// ReadList 资源索引页
func (ResourceController) ReadList(ctx echo.Context) error {
	curPage := goutils.MustInt(ctx.QueryParam("p"), 1)
	paginator := logic.NewPaginatorWithPerPage(curPage, perPage)

	resources, total := logic.DefaultResource.FindAll(ctx, paginator, "resource.mtime", "")
	hasMore := paginator.SetTotal(total).HasMorePage()

	data := map[string]interface{}{
		"resources": resources,
		"has_more":  hasMore,
	}

	return success(ctx, data)
}

// Detail 某个资源详细页
func (ResourceController) Detail(ctx echo.Context) error {
	id := goutils.MustInt(ctx.QueryParam("id"))
	resource, comments := logic.DefaultResource.FindById(ctx, id)
	if len(resource) == 0 {
		return fail(ctx, "获取失败")
	}

	logic.Views.Incr(xhttp.Request(ctx), model.TypeResource, id)

	return success(ctx, map[string]interface{}{"resource": resource, "comments": comments})
}
