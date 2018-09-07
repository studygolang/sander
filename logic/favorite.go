// Copyright 2016 The StudyGolang Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// http://studygolang.com
// Author:polaris	polaris@studygolang.com

package logic

import (
	"errors"

	"sander/db"
	"sander/logger"
	"sander/model"

	"golang.org/x/net/context"
)

type FavoriteLogic struct{}

var DefaultFavorite = FavoriteLogic{}

func (FavoriteLogic) Save(ctx context.Context, uid, objid, objtype int) error {
	favorite := &model.Favorite{}
	favorite.Uid = uid
	favorite.Objid = objid
	favorite.Objtype = objtype

	affectedNum, err := db.MasterDB.Insert(favorite)
	if err != nil {
		logger.Error("save favorite error:", err)
		return errors.New("内部服务错误")
	}

	if affectedNum == 0 {
		logger.Error("FavoriteLogic Save error: 插入数据库失败！")
		return errors.New("收藏失败！")
	}

	return nil
}

func (FavoriteLogic) Cancel(ctx context.Context, uid, objid, objtype int) error {
	_, err := db.MasterDB.Where("uid=? AND objtype=? AND objid=?", uid, objtype, objid).Delete(new(model.Favorite))
	return err
}

// HadFavorite 某个用户是否已经收藏某个对象
func (FavoriteLogic) HadFavorite(ctx context.Context, uid, objid, objtype int) int {
	favorite := &model.Favorite{}
	_, err := db.MasterDB.Where("uid=? AND objid=? and objtype=?", uid, objid, objtype).Get(favorite)
	if err != nil {
		logger.Error("FavoriteLogic HadFavorite error:", err)
		return 0
	}

	if favorite.Uid != 0 {
		return 1
	}

	return 0
}

func (FavoriteLogic) FindUserFavorites(ctx context.Context, uid, objtype, start, rows int) ([]*model.Favorite, int64) {
	favorites := make([]*model.Favorite, 0)
	err := db.MasterDB.Where("uid=? AND objtype=?", uid, objtype).Limit(rows, start).OrderBy("objid DESC").Find(&favorites)
	if err != nil {
		logger.Error("FavoriteLogic FindUserFavorites error:", err)
		return nil, 0
	}

	total, err := db.MasterDB.Where("uid=? AND objtype=?", uid, objtype).Count(new(model.Favorite))
	if err != nil {
		logger.Error("FavoriteLogic FindUserFavorites count error:", err)
		return nil, 0
	}

	return favorites, total
}
