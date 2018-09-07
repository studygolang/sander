// Copyright 2017 The StudyGolang Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// https://studygolang.com
// Author:polaris	polaris@studygolang.com

package logic

import (
	"context"
	"time"

	"sander/db"
	"sander/logger"
	"sander/model"
)

type GCTTLogic struct{}

var DefaultGCTT = GCTTLogic{}

func (self GCTTLogic) FindTranslator(ctx context.Context, me *model.Me) *model.GCTTUser {
	gcttUser := &model.GCTTUser{}
	_, err := db.MasterDB.Where("uid=?", me.Uid).Get(gcttUser)
	if err != nil {
		logger.Error("GCTTLogic FindTranslator error:", err)
		return nil
	}

	return gcttUser
}

func (self GCTTLogic) FindOne(ctx context.Context, username string) *model.GCTTUser {
	gcttUser := &model.GCTTUser{}
	_, err := db.MasterDB.Where("username=?", username).Get(gcttUser)
	if err != nil {
		logger.Error("GCTTLogic FindOne error:", err)
		return nil
	}

	return gcttUser
}

func (self GCTTLogic) BindUser(ctx context.Context, gcttUser *model.GCTTUser, uid int, githubUser *model.BindUser) error {
	var err error

	if gcttUser.Id > 0 {
		gcttUser.Uid = uid
		_, err = db.MasterDB.Id(gcttUser.Id).Update(gcttUser)
	} else {
		gcttUser = &model.GCTTUser{
			Username: githubUser.Username,
			Avatar:   githubUser.Avatar,
			Uid:      uid,
			JoinedAt: time.Now().Unix(),
		}
		_, err = db.MasterDB.Insert(gcttUser)
	}

	if err != nil {
		logger.Error("GCTTLogic BindUser error:", err)
	}

	return err
}

func (self GCTTLogic) FindCoreUsers(ctx context.Context) []*model.GCTTUser {
	gcttUsers := make([]*model.GCTTUser, 0)
	err := db.MasterDB.Where("role!=?", model.GCTTRoleTranslator).OrderBy("role ASC").Find(&gcttUsers)
	if err != nil {
		logger.Error("GCTTLogic FindUsers error:", err)
	}

	return gcttUsers
}

func (self GCTTLogic) FindUsers(ctx context.Context) []*model.GCTTUser {
	gcttUsers := make([]*model.GCTTUser, 0)
	err := db.MasterDB.Where("num>0").OrderBy("num DESC,words DESC").Find(&gcttUsers)
	if err != nil {
		logger.Error("GCTTLogic FindUsers error:", err)
	}

	return gcttUsers
}

func (self GCTTLogic) FindUnTranslateIssues(ctx context.Context, limit int) []*model.GCTTIssue {
	gcttIssues := make([]*model.GCTTIssue, 0)

	err := db.MasterDB.Where("state=?", model.IssueOpened).
		Limit(limit).OrderBy("id DESC").Find(&gcttIssues)
	if err != nil {
		logger.Error("GCTTLogic FindUnTranslateIssues error:", err)
	}

	return gcttIssues
}

func (self GCTTLogic) FindIssues(ctx context.Context, paginator *Paginator, querysring string, args ...interface{}) []*model.GCTTIssue {
	gcttIssues := make([]*model.GCTTIssue, 0)

	session := db.MasterDB.Limit(paginator.PerPage(), paginator.Offset())
	if args[0] == model.LabelClaimed {
		session.OrderBy("translating_at DESC")
	} else {
		session.OrderBy("id DESC")
	}

	if querysring != "" {
		session.Where(querysring, args...)
	}
	err := session.Limit(paginator.PerPage(), paginator.Offset()).Find(&gcttIssues)
	if err != nil {
		logger.Error("GCTTLogic FindIssues error:", err)
	}

	return gcttIssues
}

func (self GCTTLogic) IssueCount(ctx context.Context, querystring string, args ...interface{}) int64 {
	var (
		total int64
		err   error
	)
	if querystring == "" {
		total, err = db.MasterDB.Count(new(model.GCTTIssue))
	} else {
		total, err = db.MasterDB.Where(querystring, args...).Count(new(model.GCTTIssue))
	}

	if err != nil {
		logger.Error("GCTTLogic Count error:", err)
	}

	return total
}

func (self GCTTLogic) FindNewestGit(ctx context.Context) []*model.GCTTGit {
	gcttGits := make([]*model.GCTTGit, 0)
	err := db.MasterDB.Where("translated_at!=0").OrderBy("translated_at DESC").
		Limit(10).Find(&gcttGits)
	if err != nil {
		logger.Error("GCTTLogic FindNewestGit error:", err)
	}

	return gcttGits
}

func (self GCTTLogic) FindTimeLines(ctx context.Context) []*model.GCTTTimeLine {
	gcttTimeLines := make([]*model.GCTTTimeLine, 0)
	err := db.MasterDB.Find(&gcttTimeLines)
	if err != nil {
		logger.Error("GCTTLogic FindTimeLines error:", err)
	}
	return gcttTimeLines
}
