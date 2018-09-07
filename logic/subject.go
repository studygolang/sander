// Copyright 2017 The StudyGolang Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// http://studygolang.com
// Author:polaris	polaris@studygolang.com

package logic

import (
	"errors"
	"net/url"
	"strings"

	"sander/db"
	"sander/global"
	"sander/logger"
	"sander/model"
	"sander/util"

	"github.com/polaris1119/goutils"
	"github.com/polaris1119/set"
	"github.com/polaris1119/slices"
	"golang.org/x/net/context"
)

type SubjectLogic struct{}

var DefaultSubject = SubjectLogic{}

func (self SubjectLogic) FindBy(ctx context.Context, paginator *Paginator) []*model.Subject {

	subjects := make([]*model.Subject, 0)
	err := db.MasterDB.OrderBy("article_num DESC").Limit(paginator.PerPage(), paginator.Offset()).
		Find(&subjects)
	if err != nil {
		logger.Error("SubjectLogic FindBy error:", err)
	}

	if len(subjects) > 0 {

		uidSet := set.New(set.NonThreadSafe)

		for _, subject := range subjects {
			uidSet.Add(subject.Uid)
		}

		usersMap := DefaultUser.FindUserInfos(ctx, set.IntSlice(uidSet))

		for _, subject := range subjects {
			subject.User = usersMap[subject.Uid]
		}
	}

	return subjects
}

func (self SubjectLogic) FindOne(ctx context.Context, sid int) *model.Subject {

	subject := &model.Subject{}
	_, err := db.MasterDB.Id(sid).Get(subject)
	if err != nil {
		logger.Error("SubjectLogic FindOne get error:", err)
	}

	if subject.Uid > 0 {
		subject.User = DefaultUser.findUser(ctx, subject.Uid)
	}

	return subject
}

func (self SubjectLogic) findByIds(ids []int) map[int]*model.Subject {
	if len(ids) == 0 {
		return nil
	}

	subjects := make(map[int]*model.Subject)
	err := db.MasterDB.In("id", ids).Find(&subjects)
	if err != nil {
		return nil
	}

	return subjects
}

func (self SubjectLogic) FindArticles(ctx context.Context, sid int, paginator *Paginator, orderBy string) []*model.Article {

	order := "subject_article.created_at DESC"
	if orderBy == "commented_at" {
		order = "articles.lastreplytime DESC"
	}

	subjectArticles := make([]*model.SubjectArticles, 0)
	err := db.MasterDB.Join("INNER", "subject_article", "subject_article.article_id = articles.id").
		Where("sid=? AND state=?", sid, model.ContributeStateOnline).
		Limit(paginator.PerPage(), paginator.Offset()).
		OrderBy(order).Find(&subjectArticles)
	if err != nil {
		logger.Error("SubjectLogic FindArticles Find subject_article error:", err)
		return nil
	}

	articles := make([]*model.Article, 0, len(subjectArticles))
	for _, subjectArticle := range subjectArticles {
		if subjectArticle.Status == model.ArticleStatusOffline {
			continue
		}

		articles = append(articles, &subjectArticle.Article)
	}

	DefaultArticle.fillUser(articles)
	return articles
}

// FindArticleTotal 专栏收录的文章数
func (self SubjectLogic) FindArticleTotal(ctx context.Context, sid int) int64 {

	total, err := db.MasterDB.Where("sid=?", sid).Count(new(model.SubjectArticle))
	if err != nil {
		logger.Error("SubjectLogic FindArticleTotal error:", err)
	}

	return total
}

// FindFollowers 专栏关注的用户
func (self SubjectLogic) FindFollowers(ctx context.Context, sid int) []*model.SubjectFollower {

	followers := make([]*model.SubjectFollower, 0)
	err := db.MasterDB.Where("sid=?", sid).OrderBy("id DESC").Limit(8).Find(&followers)
	if err != nil {
		logger.Error("SubjectLogic FindFollowers error:", err)
	}

	if len(followers) == 0 {
		return followers
	}

	uids := slices.StructsIntSlice(followers, "Uid")
	usersMap := DefaultUser.FindUserInfos(ctx, uids)
	for _, follower := range followers {
		follower.User = usersMap[follower.Uid]
		follower.TimeAgo = util.TimeAgo(follower.CreatedAt)
	}

	return followers
}

func (self SubjectLogic) findFollowersBySid(sid int) []*model.SubjectFollower {
	followers := make([]*model.SubjectFollower, 0)
	db.MasterDB.Where("sid=?", sid).Find(&followers)
	return followers
}

// FindFollowerTotal 专栏关注的用户数
func (self SubjectLogic) FindFollowerTotal(ctx context.Context, sid int) int64 {

	total, err := db.MasterDB.Where("sid=?", sid).Count(new(model.SubjectFollower))
	if err != nil {
		logger.Error("SubjectLogic FindFollowerTotal error:", err)
	}

	return total
}

// Follow 关注或取消关注
func (self SubjectLogic) Follow(ctx context.Context, sid int, me *model.Me) (err error) {

	follower := &model.SubjectFollower{}
	_, err = db.MasterDB.Where("sid=? AND uid=?", sid, me.Uid).Get(follower)
	if err != nil {
		logger.Error("SubjectLogic Follow Get error:", err)
	}

	if follower.Id > 0 {
		_, err = db.MasterDB.Where("sid=? AND uid=?", sid, me.Uid).Delete(new(model.SubjectFollower))
		if err != nil {
			logger.Error("SubjectLogic Follow Delete error:", err)
		}

		return
	}

	follower.Sid = sid
	follower.Uid = me.Uid
	_, err = db.MasterDB.Insert(follower)
	if err != nil {
		logger.Error("SubjectLogic Follow insert error:", err)
	}
	return
}

func (self SubjectLogic) HadFollow(ctx context.Context, sid int, me *model.Me) bool {

	num, err := db.MasterDB.Where("sid=? AND uid=?", sid, me.Uid).Count(new(model.SubjectFollower))
	if err != nil {
		logger.Error("SubjectLogic Follow insert error:", err)
	}

	return num > 0
}

// Contribute 投稿
func (self SubjectLogic) Contribute(ctx context.Context, me *model.Me, sid, articleId int) error {

	subject := self.FindOne(ctx, sid)
	if subject.Id == 0 {
		return errors.New("该专栏不存在")
	}

	count, _ := db.MasterDB.Where("article_id=?", articleId).Count(new(model.SubjectArticle))
	if count >= 5 {
		return errors.New("该文超过 5 次投稿")
	}

	subjectArticle := &model.SubjectArticle{
		Sid:       sid,
		ArticleId: articleId,
		State:     model.ContributeStateNew,
	}

	// TODO: 非创建管理员投稿不需要审核
	if subject.Uid == me.Uid {
		subjectArticle.State = model.ContributeStateOnline
	} else {
		if !subject.Contribute {
			return errors.New("不允许投稿")
		}

		// 不需要审核
		if !subject.Audit {
			subjectArticle.State = model.ContributeStateOnline
		}
	}

	session := db.MasterDB.NewSession()
	defer session.Close()
	session.Begin()

	_, err := session.Insert(subjectArticle)
	if err != nil {
		session.Rollback()
		logger.Error("SubjectLogic Contribute insert error:", err)
		return errors.New("投稿失败:" + err.Error())
	}

	_, err = session.Id(sid).Incr("article_num", 1).Update(new(model.Subject))
	if err != nil {
		session.Rollback()
		logger.Error("SubjectLogic Contribute update subject article num error:", err)
		return errors.New("投稿失败:" + err.Error())
	}

	if err := session.Commit(); err == nil {
		// 成功，发送站内系统消息给关注者
		go self.sendMsgForFollower(ctx, subject, sid, articleId)
	}

	return nil
}

// sendMsgForFollower 专栏投稿发送消息给关注者
func (self SubjectLogic) sendMsgForFollower(ctx context.Context, subject *model.Subject, sid, articleId int) {
	followers := self.findFollowersBySid(sid)
	for _, f := range followers {
		DefaultMessage.SendSystemMsgTo(ctx, f.Uid, model.MsgtypeSubjectContribute, map[string]interface{}{
			"uid":   subject.Uid,
			"objid": articleId,
			"sid":   sid,
		})
	}
}

// RemoveContribute 删除投稿
func (self SubjectLogic) RemoveContribute(ctx context.Context, sid, articleId int) error {

	session := db.MasterDB.NewSession()
	defer session.Close()
	session.Begin()

	_, err := session.Where("sid=? AND article_id=?", sid, articleId).Delete(new(model.SubjectArticle))
	if err != nil {
		session.Rollback()
		logger.Error("SubjectLogic RemoveContribute delete error:", err)
		return errors.New("删除投稿失败:" + err.Error())
	}

	_, err = session.Id(sid).Decr("article_num", 1).Update(new(model.Subject))
	if err != nil {
		session.Rollback()
		logger.Error("SubjectLogic RemoveContribute update subject article num error:", err)
		return errors.New("删除投稿失败:" + err.Error())
	}

	session.Commit()

	return nil
}

func (self SubjectLogic) ExistByName(name string) bool {
	exist, _ := db.MasterDB.Where("name=?", name).Exist(new(model.Subject))
	return exist
}

// Publish 发布专栏。
func (self SubjectLogic) Publish(ctx context.Context, me *model.Me, form url.Values) (sid int, err error) {

	sid = goutils.MustInt(form.Get("sid"))
	if sid != 0 {
		subject := &model.Subject{}
		_, err = db.MasterDB.Id(sid).Get(subject)
		if err != nil {
			logger.Error("Publish Subject find error:", err)
			return
		}

		_, err = self.Modify(ctx, me, form)
		if err != nil {
			logger.Error("Publish Subject modify error:", err)
			return
		}

	} else {
		subject := &model.Subject{}
		err = schemaDecoder.Decode(subject, form)
		if err != nil {
			logger.Error("SubjectLogic Publish decode error:", err)
			return
		}
		subject.Uid = me.Uid

		_, err = db.MasterDB.Insert(subject)
		if err != nil {
			logger.Error("SubjectLogic Publish insert error:", err)
			return
		}
		sid = subject.Id
	}
	return
}

// Modify 修改专栏
func (SubjectLogic) Modify(ctx context.Context, user *model.Me, form url.Values) (errMsg string, err error) {
	change := map[string]interface{}{}

	fields := []string{"name", "description", "cover", "contribute", "audit"}
	for _, field := range fields {
		change[field] = form.Get(field)
	}

	sid := form.Get("sid")
	_, err = db.MasterDB.Table(new(model.Subject)).Id(sid).Update(change)
	if err != nil {
		logger.Error("更新专栏 【%s】 信息失败：%s\n", sid, err)
		errMsg = "对不起，服务器内部错误，请稍后再试！"
		return
	}

	return
}

func (self SubjectLogic) FindArticleSubjects(ctx context.Context, articleId int) []*model.Subject {

	subjectArticles := make([]*model.SubjectArticle, 0)
	err := db.MasterDB.Where("article_id=?", articleId).Find(&subjectArticles)
	if err != nil {
		logger.Error("SubjectLogic FindArticleSubjects find error:", err)
		return nil
	}

	subjectLen := len(subjectArticles)
	if subjectLen == 0 {
		return nil
	}

	sids := make([]int, subjectLen)
	for i, subjectArticle := range subjectArticles {
		sids[i] = subjectArticle.Sid
	}

	subjects := make([]*model.Subject, 0)
	err = db.MasterDB.In("id", sids).Find(&subjects)
	if err != nil {
		logger.Error("SubjectLogic FindArticleSubjects find subject error:", err)
		return nil
	}

	return subjects
}

// FindMine 获取我管理的专栏列表
func (self SubjectLogic) FindMine(ctx context.Context, me *model.Me, articleId int, kw string) []map[string]interface{} {

	subjects := make([]*model.Subject, 0)
	// 先是我创建的专栏
	session := db.MasterDB.Where("uid=?", me.Uid)
	if kw != "" {
		session.Where("name LIKE ?", "%"+kw+"%")
	}
	err := session.Find(&subjects)
	if err != nil {
		logger.Error("SubjectLogic FindMine find subject error:", err)
		return nil
	}

	adminSubjects := make([]*model.Subject, 0)
	// 获取我管理的专栏
	strSql := "SELECT s.* FROM subject s,subject_admin sa WHERE s.id=sa.sid AND sa.uid=?"
	if kw != "" {
		strSql += " AND s.name LIKE '%" + kw + "%'"
	}
	err = db.MasterDB.Sql(strSql, me.Uid).Find(&adminSubjects)
	if err != nil {
		logger.Error("SubjectLogic FindMine find admin subject error:", err)
	}

	subjectArticles := make([]*model.SubjectArticle, 0)
	err = db.MasterDB.Where("article_id=?", articleId).Find(&subjectArticles)
	if err != nil {
		logger.Error("SubjectLogic FindMine find subject article error:", err)
	}
	subjectArticleMap := make(map[int]struct{})
	for _, sa := range subjectArticles {
		subjectArticleMap[sa.Sid] = struct{}{}
	}

	uidSet := set.New(set.NonThreadSafe)
	for _, subject := range subjects {
		uidSet.Add(subject.Uid)
	}
	for _, subject := range adminSubjects {
		uidSet.Add(subject.Uid)
	}
	usersMap := DefaultUser.FindUserInfos(ctx, set.IntSlice(uidSet))

	subjectMapSlice := make([]map[string]interface{}, 0, len(subjects)+len(adminSubjects))

	for _, subject := range subjects {
		self.genSubjectMapSlice(subject, &subjectMapSlice, subjectArticleMap, usersMap)
	}

	for _, subject := range adminSubjects {
		self.genSubjectMapSlice(subject, &subjectMapSlice, subjectArticleMap, usersMap)
	}

	return subjectMapSlice
}

func (self SubjectLogic) genSubjectMapSlice(subject *model.Subject, subjectMapSlice *[]map[string]interface{}, subjectArticleMap map[int]struct{}, usersMap map[int]*model.User) {
	hadAdd := 0
	if _, ok := subjectArticleMap[subject.Id]; ok {
		hadAdd = 1
	}

	cover := subject.Cover
	if cover == "" {
		user := usersMap[subject.Uid]
		cover = util.Gravatar(user.Avatar, user.Email, 48, true)
	} else if !strings.HasPrefix(cover, "http") {
		cdnDomain := global.App.CanonicalCDN(true)
		cover = cdnDomain + subject.Cover
	}

	*subjectMapSlice = append(*subjectMapSlice, map[string]interface{}{
		"id":       subject.Id,
		"name":     subject.Name,
		"cover":    cover,
		"username": usersMap[subject.Uid].Username,
		"had_add":  hadAdd,
	})
}
