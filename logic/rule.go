// Copyright 2016 The StudyGolang Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// http://studygolang.com
// Author:polaris	polaris@studygolang.com

package logic

import (
	"net/url"

	"sander/db"
	"sander/logger"
	"sander/model"

	"golang.org/x/net/context"
)

type RuleLogic struct{}

var DefaultRule = RuleLogic{}

// 获取抓取规则列表（分页）
func (RuleLogic) FindBy(ctx context.Context, conds map[string]string, curPage, limit int) ([]*model.CrawlRule, int) {

	session := db.MasterDB.NewSession()

	for k, v := range conds {
		session.And(k+"=?", v)
	}

	totalSession := session.Clone()

	offset := (curPage - 1) * limit
	ruleList := make([]*model.CrawlRule, 0)
	err := session.OrderBy("id DESC").Limit(limit, offset).Find(&ruleList)
	if err != nil {
		logger.Error("rule find error:", err)
		return nil, 0
	}

	total, err := totalSession.Count(new(model.CrawlRule))
	if err != nil {
		logger.Error("rule find count error:", err)
		return nil, 0
	}

	return ruleList, int(total)
}

func (RuleLogic) FindById(ctx context.Context, id string) *model.CrawlRule {

	rule := &model.CrawlRule{}
	_, err := db.MasterDB.Id(id).Get(rule)
	if err != nil {
		logger.Error("find rule error:", err)
		return nil
	}

	if rule.Id == 0 {
		return nil
	}

	return rule
}

func (RuleLogic) Save(ctx context.Context, form url.Values, opUser string) (errMsg string, err error) {

	rule := &model.CrawlRule{}
	err = schemaDecoder.Decode(rule, form)
	if err != nil {
		logger.Error("rule Decode error", err)
		errMsg = err.Error()
		return
	}

	rule.OpUser = opUser

	if rule.Id != 0 {
		_, err = db.MasterDB.Id(rule.Id).Update(rule)
	} else {
		_, err = db.MasterDB.Insert(rule)
	}

	if err != nil {
		errMsg = "内部服务器错误"
		logger.Error("rule save:", errMsg, ":", err)
		return
	}

	return
}

func (RuleLogic) Delete(ctx context.Context, id string) error {
	_, err := db.MasterDB.Id(id).Delete(new(model.CrawlRule))
	return err
}
