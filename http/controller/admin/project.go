// Copyright 2014 The StudyGolang Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// http://studygolang.com
// Author: polaris	polaris@studygolang.com

package admin

import (
	"net/http"
	"strings"

	"sander/logic"
	"sander/model"

	"github.com/labstack/echo"
)

// ProjectController .
type ProjectController struct{}

// RegisterRoute 注册路由
func (p ProjectController) RegisterRoute(g *echo.Group) {
	g.GET("/crawl/project/list", p.ProjectList)
	g.POST("/crawl/project/query.html", p.ProjectQuery)
	g.Match([]string{"GET", "POST"}, "/crawl/project/new", p.CrawlProject)
	g.Match([]string{"GET", "POST"}, "/crawl/project/modify", p.Modify)
}

// ProjectList 所有文章（分页）
func (ProjectController) ProjectList(ctx echo.Context) error {
	curPage, limit := parsePage(ctx)
	articles, total := logic.DefaultArticle.FindArticleByPage(ctx, nil, curPage, limit)

	if articles == nil {
		return ctx.HTML(http.StatusInternalServerError, "500")
	}

	data := map[string]interface{}{
		"datalist":   articles,
		"total":      total,
		"totalPages": (total + limit - 1) / limit,
		"page":       curPage,
		"limit":      limit,
	}

	return render(ctx, "article/list.html,article/query.html", data)
}

// ProjectQuery .
func (ProjectController) ProjectQuery(ctx echo.Context) error {
	curPage, limit := parsePage(ctx)
	conds := parseConds(ctx, []string{"id", "domain", "title"})

	articles, total := logic.DefaultArticle.FindArticleByPage(ctx, conds, curPage, limit)

	if articles == nil {
		return ctx.HTML(http.StatusInternalServerError, "500")
	}

	data := map[string]interface{}{
		"datalist":   articles,
		"total":      total,
		"totalPages": (total + limit - 1) / limit,
		"page":       curPage,
		"limit":      limit,
	}

	return renderQuery(ctx, "article/query.html", data)
}

// CrawlProject .
func (ProjectController) CrawlProject(ctx echo.Context) error {
	var data = make(map[string]interface{})

	if ctx.FormValue("submit") == "1" {
		urls := strings.Split(ctx.FormValue("urls"), "\n")

		var errMsg string
		for _, url := range urls {
			err := logic.DefaultProject.ParseOneProject(strings.TrimSpace(url))
			if err != nil {
				errMsg = err.Error()
			}
		}

		if errMsg != "" {
			return fail(ctx, 1, errMsg)
		}
		return success(ctx, nil)
	}

	return render(ctx, "project/new.html", data)
}

// Modify .
func (p ProjectController) Modify(ctx echo.Context) error {
	var data = make(map[string]interface{})

	if ctx.FormValue("submit") == "1" {
		user := ctx.Get("user").(*model.Me)
		errMsg, err := logic.DefaultArticle.Modify(ctx, user, ctx.FormParams())
		if err != nil {
			return fail(ctx, 1, errMsg)
		}
		return success(ctx, nil)
	}
	article, err := logic.DefaultArticle.FindById(ctx, ctx.QueryParam("id"))
	if err != nil {
		return ctx.Redirect(http.StatusSeeOther, ctx.Echo().URI(echo.HandlerFunc(p.ProjectList)))
	}

	data["article"] = article
	data["statusSlice"] = model.ArticleStatusSlice
	data["langSlice"] = model.LangSlice

	return render(ctx, "article/modify.html", data)

}

// /crawl/article/del
// func DelArticleHandler(rw http.ResponseWriter, req *http.Request) {
// 	var data = make(map[string]interface{})

// 	id := req.FormValue("id")

// 	if _, err := strconv.Atoi(id); err != nil {
// 		data["ok"] = 0
// 		data["error"] = "id不是整型"

// 		filter.SetData(req, data)
// 		return
// 	}

// 	if err := service.DelArticle(id); err != nil {
// 		data["ok"] = 0
// 		data["error"] = "删除失败！"
// 	} else {
// 		data["ok"] = 1
// 		data["msg"] = "删除成功！"
// 	}

// 	filter.SetData(req, data)
// }
