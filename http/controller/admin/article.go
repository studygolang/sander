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
	"github.com/polaris1119/goutils"
)

// ArticleController .
type ArticleController struct{}

// RegisterRoute 注册路由.
func (a ArticleController) RegisterRoute(g *echo.Group) {
	g.GET("/crawl/article/list", a.ArticleList)
	g.POST("/crawl/article/query.html", a.ArticleQuery)
	g.POST("/crawl/article/move", a.MoveToTopic)
	g.Match([]string{"GET", "POST"}, "/crawl/article/new", a.CrawlArticle)
	g.Match([]string{"GET", "POST"}, "/crawl/article/publish", a.Publish)
	g.Match([]string{"GET", "POST"}, "/crawl/article/modify", a.Modify)
}

// ArticleList 所有文章（分页）
func (ArticleController) ArticleList(ctx echo.Context) error {
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

// ArticleQuery .
func (ArticleController) ArticleQuery(ctx echo.Context) error {
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

// CrawlArticle .
func (ArticleController) CrawlArticle(ctx echo.Context) error {
	var data = make(map[string]interface{})

	if ctx.FormValue("submit") == "1" {
		urls := strings.Split(ctx.FormValue("urls"), "\n")

		var (
			errMsg string
			err    error
		)
		for _, url := range urls {
			url = strings.TrimSpace(url)

			if strings.HasPrefix(url, "http") {
				_, err = logic.DefaultArticle.ParseArticle(ctx, url, false)
			} else {
				isAll := false
				websiteInfo := strings.Split(url, ":")
				if len(websiteInfo) >= 2 {
					isAll = goutils.MustBool(websiteInfo[1])
				}
				err = logic.DefaultAutoCrawl.CrawlWebsite(strings.TrimSpace(websiteInfo[0]), isAll)
			}

			if err != nil {
				errMsg = err.Error()
			}
		}

		if errMsg != "" {
			return fail(ctx, 1, errMsg)
		}
		return success(ctx, nil)
	}

	return render(ctx, "article/new.html", data)
}

// Publish .
func (a ArticleController) Publish(ctx echo.Context) error {
	var data = make(map[string]interface{})

	if ctx.FormValue("submit") == "1" {
		user := ctx.Get("user").(*model.Me)
		err := logic.DefaultArticle.PublishFromAdmin(ctx, user, ctx.FormParams())
		if err != nil {
			return fail(ctx, 1, err.Error())
		}
		return success(ctx, nil)
	}

	data["statusSlice"] = model.ArticleStatusSlice
	data["langSlice"] = model.LangSlice

	return render(ctx, "article/publish.html", data)
}

// Modify .
func (a ArticleController) Modify(ctx echo.Context) error {
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
		return ctx.Redirect(http.StatusSeeOther, ctx.Echo().URI(echo.HandlerFunc(a.ArticleList)))
	}

	data["article"] = article
	data["statusSlice"] = model.ArticleStatusSlice
	data["langSlice"] = model.LangSlice

	return render(ctx, "article/modify.html", data)
}

// MoveToTopic 放入 Topic 中
func (a ArticleController) MoveToTopic(ctx echo.Context) error {
	user := ctx.Get("user").(*model.Me)
	err := logic.DefaultArticle.MoveToTopic(ctx, ctx.QueryParam("id"), user)

	if err != nil {
		return fail(ctx, 1, err.Error())
	}
	return success(ctx, nil)
}
