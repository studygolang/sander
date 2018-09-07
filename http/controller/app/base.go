// Copyright 2016 The StudyGolang Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// http://studygolang.com
// Author: polaris	polaris@studygolang.com

package app

import (
	"encoding/json"
	"net/http"

	"sander/db/nosql"
	xhttp "sander/http"
	"sander/logger"

	"github.com/labstack/echo"
)

const perPage = 12

func success(ctx echo.Context, data interface{}) error {
	result := map[string]interface{}{
		"code": 0,
		"msg":  "ok",
		"data": data,
	}

	b, err := json.Marshal(result)
	if err != nil {
		return err
	}

	go func(b []byte) {
		if cacheKey := ctx.Get(nosql.CacheKey); cacheKey != nil {
			nosql.DefaultLRUCache.CompressAndAdd(cacheKey, b, nosql.NewCacheData())
		}
	}(b)

	xhttp.AccessControl(ctx)

	if ctx.Response().Committed() {
		return nil
	}

	return ctx.JSONBlob(http.StatusOK, b)
}

func fail(ctx echo.Context, msg string, codes ...int) error {
	xhttp.AccessControl(ctx)

	if ctx.Response().Committed() {
		return nil
	}

	code := 1
	if len(codes) > 0 {
		code = codes[0]
	}
	result := map[string]interface{}{
		"code": code,
		"msg":  msg,
	}

	logger.Error("operate fail:%+v", result)

	return ctx.JSON(http.StatusOK, result)
}
