// Copyright 2017 The StudyGolang Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// http://studygolang.com
// Author:polaris	polaris@studygolang.com

package logic

import (
	"encoding/json"
	"errors"
	"io/ioutil"

	"sander/config"
	"sander/db"
	"sander/logger"
	"sander/model"

	"golang.org/x/net/context"
	"golang.org/x/oauth2"
)

var githubConf *oauth2.Config

const GithubAPIBaseUrl = "https://api.github.com"

func init() {
	githubConf = &oauth2.Config{
		ClientID:     config.ConfigFile.MustValue("github", "client_id"),
		ClientSecret: config.ConfigFile.MustValue("github", "client_secret"),
		Scopes:       []string{"user:email"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://github.com/login/oauth/authorize",
			TokenURL: "https://github.com/login/oauth/access_token",
		},
	}
}

type ThirdUserLogic struct{}

var DefaultThirdUser = ThirdUserLogic{}

func (ThirdUserLogic) GithubAuthCodeUrl(ctx context.Context, redirectURL string) string {
	// Redirect user to consent page to ask for permission
	// for the scopes specified above.
	githubConf.RedirectURL = redirectURL
	return githubConf.AuthCodeURL("state", oauth2.AccessTypeOffline)
}

func (self ThirdUserLogic) LoginFromGithub(ctx context.Context, code string) (*model.User, error) {

	githubUser, token, err := self.githubTokenAndUser(ctx, code)
	if err != nil {
		logger.Error("LoginFromGithub githubTokenAndUser error:", err)
		return nil, err
	}

	bindUser := &model.BindUser{}
	// 是否已经授权过了
	_, err = db.MasterDB.Where("username=? AND type=?", githubUser.Login, model.BindTypeGithub).Get(bindUser)
	if err != nil {
		logger.Error("LoginFromGithub Get BindUser error:", err)
		return nil, err
	}

	if bindUser.Uid > 0 {
		// 更新 token 信息
		bindUser.AccessToken = token.AccessToken
		bindUser.RefreshToken = token.RefreshToken
		if !token.Expiry.IsZero() {
			bindUser.Expire = int(token.Expiry.Unix())
		}
		_, err = db.MasterDB.Where("uid=?", bindUser.Uid).Update(bindUser)
		if err != nil {
			logger.Error("LoginFromGithub update token error:", err)
			return nil, err
		}

		user := DefaultUser.FindOne(ctx, "uid", bindUser.Uid)
		return user, nil
	}

	exists := DefaultUser.EmailOrUsernameExists(ctx, githubUser.Email, githubUser.Login)
	if exists {
		// TODO: 考虑改进？
		logger.Error("LoginFromGithub Github 对应的用户信息被占用")
		return nil, errors.New("Github 对应的用户信息被占用，可能你注册过本站，用户名密码登录试试！")
	}

	session := db.MasterDB.NewSession()
	defer session.Close()
	session.Begin()

	// 有可能获取不到 email？加上 @github.com做邮箱后缀
	if githubUser.Email == "" {
		githubUser.Email = githubUser.Login + "@github.com"
	}
	// 生成本站用户
	user := &model.User{
		Email:    githubUser.Email,
		Username: githubUser.Login,
		Name:     githubUser.Name,
		City:     githubUser.Location,
		Company:  githubUser.Company,
		Github:   githubUser.Login,
		Website:  githubUser.Blog,
		Avatar:   githubUser.AvatarUrl,
		IsThird:  1,
		Status:   model.UserStatusAudit,
	}
	err = DefaultUser.doCreateUser(ctx, session, user)
	if err != nil {
		session.Rollback()
		logger.Error("LoginFromGithub doCreateUser error:", err)
		return nil, err
	}

	bindUser = &model.BindUser{
		Uid:          user.Uid,
		Type:         model.BindTypeGithub,
		Email:        user.Email,
		Tuid:         githubUser.Id,
		Username:     githubUser.Login,
		Name:         githubUser.Name,
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		Avatar:       githubUser.AvatarUrl,
	}
	if !token.Expiry.IsZero() {
		bindUser.Expire = int(token.Expiry.Unix())
	}
	_, err = session.Insert(bindUser)
	if err != nil {
		session.Rollback()
		logger.Error("LoginFromGithub bindUser error:", err)
		return nil, err
	}

	session.Commit()

	return user, nil
}

func (self ThirdUserLogic) BindGithub(ctx context.Context, code string, me *model.Me) error {

	githubUser, token, err := self.githubTokenAndUser(ctx, code)
	if err != nil {
		logger.Error("LoginFromGithub githubTokenAndUser error:", err)
		return err
	}

	bindUser := &model.BindUser{}
	// 是否已经授权过了
	_, err = db.MasterDB.Where("username=? AND type=?", githubUser.Login, model.BindTypeGithub).Get(bindUser)
	if err != nil {
		logger.Error("LoginFromGithub Get BindUser error:", err)
		return err
	}

	if bindUser.Uid > 0 {
		// 更新 token 信息
		bindUser.AccessToken = token.AccessToken
		bindUser.RefreshToken = token.RefreshToken
		if !token.Expiry.IsZero() {
			bindUser.Expire = int(token.Expiry.Unix())
		}
		_, err = db.MasterDB.Where("uid=?", bindUser.Uid).Update(bindUser)
		if err != nil {
			logger.Error("LoginFromGithub update token error:", err)
			return err
		}

		return nil
	}

	bindUser = &model.BindUser{
		Uid:          me.Uid,
		Type:         model.BindTypeGithub,
		Email:        githubUser.Email,
		Tuid:         githubUser.Id,
		Username:     githubUser.Login,
		Name:         githubUser.Name,
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		Avatar:       githubUser.AvatarUrl,
	}
	if !token.Expiry.IsZero() {
		bindUser.Expire = int(token.Expiry.Unix())
	}
	_, err = db.MasterDB.Insert(bindUser)
	if err != nil {
		logger.Error("LoginFromGithub insert bindUser error:", err)
		return err
	}

	return nil
}

func (ThirdUserLogic) UnBindUser(ctx context.Context, bindId interface{}, me *model.Me) error {
	if !DefaultUser.HasPasswd(ctx, me.Uid) {
		return errors.New("请先设置密码！")
	}
	_, err := db.MasterDB.Where("id=? AND uid=?", bindId, me.Uid).Delete(new(model.BindUser))
	return err
}

func (ThirdUserLogic) findUid(thirdUsername string, typ int) int {
	bindUser := &model.BindUser{}
	_, err := db.MasterDB.Where("username=? AND `type`=?", thirdUsername, typ).Get(bindUser)
	if err != nil {
		logger.Error("ThirdUserLogic findUid error:%+v", err)
	}

	return bindUser.Uid
}

func (ThirdUserLogic) githubTokenAndUser(ctx context.Context, code string) (*model.GithubUser, *oauth2.Token, error) {
	token, err := githubConf.Exchange(ctx, code)
	if err != nil {
		return nil, nil, err
	}

	httpClient := githubConf.Client(ctx, token)
	resp, err := httpClient.Get(GithubAPIBaseUrl + "/user")
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()

	respBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, err
	}

	githubUser := &model.GithubUser{}
	err = json.Unmarshal(respBytes, githubUser)
	if err != nil {
		return nil, nil, err
	}

	if githubUser.Id == 0 {
		return nil, nil, errors.New("get github user info error")
	}

	return githubUser, token, nil
}
