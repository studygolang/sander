// Copyright 2018 The StudyGolang Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// https://studygolang.com
// Author:polaris	polaris@studygolang.com

package logic

import (
	"sander/db"
	"sander/model"

	"github.com/polaris1119/logger"
	"golang.org/x/net/context"
)

type DownloadLogic struct{}

var DefaultDownload = DownloadLogic{}

func (DownloadLogic) FindAll(ctx context.Context) []*model.Download {
	downloads := make([]*model.Download, 0)
	err := db.MasterDB.Desc("seq").Find(&downloads)
	if err != nil {
		logger.Errorln("DownloadLogic FindAll Error:", err)
	}

	return downloads
}
