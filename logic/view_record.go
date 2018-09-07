// Copyright 2017 The StudyGolang Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// http://studygolang.com
// Author:polaris	polaris@studygolang.com

package logic

import (
	"sander/db"
	"sander/logger"
	"sander/model"

	"golang.org/x/net/context"
)

type ViewRecordLogic struct{}

var DefaultViewRecord = ViewRecordLogic{}

func (ViewRecordLogic) Record(objid, objtype, uid int) {

	total, err := db.MasterDB.Where("objid=? AND objtype=? AND uid=?", objid, objtype, uid).Count(new(model.ViewRecord))
	if err != nil {
		logger.Error("ViewRecord logic Record count error:%+v", err)
		return
	}

	if total > 0 {
		return
	}

	viewRecord := &model.ViewRecord{
		Objid:   objid,
		Objtype: objtype,
		Uid:     uid,
	}

	if _, err = db.MasterDB.Insert(viewRecord); err != nil {
		logger.Error("ViewRecord logic Record insert Error:%+v", err)
		return
	}

	return
}

func (ViewRecordLogic) FindUserNum(ctx context.Context, objid, objtype int) int64 {

	total, err := db.MasterDB.Where("objid=? AND objtype=?", objid, objtype).Count(new(model.ViewRecord))
	if err != nil {
		logger.Error("ViewRecordLogic FindUserNum error:%+v", err)
	}

	return total
}
