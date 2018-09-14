// Copyright 2016 The StudyGolang Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// http://studygolang.com
// Author: polaris	polaris@studygolang.com

package db

import (
	"database/sql"
	"errors"
	"fmt"

	"sander/config"

	_ "github.com/go-sql-driver/mysql"
	"github.com/go-xorm/core"
	"github.com/go-xorm/xorm"
)

var (
	dns string
	// MasterDB .
	MasterDB *xorm.Engine
)

func init() {
	mysqlConfig, err := config.ConfigFile.GetSection("mysql")
	if err != nil {
		fmt.Println("get mysql config error:", err)
		return
	}
	fillDNS(mysqlConfig)
	// 启动时就打开数据库连接
	if err = initEngine(); err != nil {
		panic(err)
	}
}

var (
	// ErrConnectDB .
	ErrConnectDB = errors.New("connect db error")
	// ErrUseDB .
	ErrUseDB = errors.New("use db error")
)

// TestDB 测试数据库
func TestDB() error {
	mysqlConfig, err := config.ConfigFile.GetSection("mysql")
	if err != nil {
		fmt.Println("get mysql config error:", err)
		return err
	}

	tmpDNS := fmt.Sprintf("%s:%s@tcp(%s:%s)/?charset=%s&parseTime=True&loc=Local",
		mysqlConfig["user"],
		mysqlConfig["password"],
		mysqlConfig["host"],
		mysqlConfig["port"],
		mysqlConfig["charset"])
	egnine, err := xorm.NewEngine("mysql", tmpDNS)
	if err != nil {
		fmt.Println("new engine error:", err)
		return err
	}
	defer egnine.Close()

	// 测试数据库连接是否 OK
	if err = egnine.Ping(); err != nil {
		fmt.Println("ping db error:", err)
		return ErrConnectDB
	}

	_, err = egnine.Exec("use " + mysqlConfig["dbname"])
	if err != nil {
		fmt.Println("use db error:", err)
		_, err = egnine.Exec("CREATE DATABASE " + mysqlConfig["dbname"] + " DEFAULT CHARACTER SET " + mysqlConfig["charset"])
		if err != nil {
			fmt.Println("create database error:", err)

			return ErrUseDB
		}
		fmt.Println("create database successfully!")
	}
	// 初始化 MasterDB
	Init()
	return nil
}

// Init .
func Init() error {
	mysqlConfig, err := config.ConfigFile.GetSection("mysql")
	if err != nil {
		fmt.Println("get mysql config error:", err)
		return err
	}
	fillDNS(mysqlConfig)
	// 启动时就打开数据库连接
	if err = initEngine(); err != nil {
		fmt.Println("mysql is not open:", err)
		return err
	}
	return nil
}

func fillDNS(mysqlConfig map[string]string) {
	dns = fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=%s&parseTime=True&loc=Local",
		mysqlConfig["user"],
		mysqlConfig["password"],
		mysqlConfig["host"],
		mysqlConfig["port"],
		mysqlConfig["dbname"],
		mysqlConfig["charset"])
}

func initEngine() error {
	var err error

	MasterDB, err = xorm.NewEngine("mysql", dns)
	if err != nil {
		return err
	}

	maxIdle := config.ConfigFile.MustInt("mysql", "max_idle", 2)
	maxConn := config.ConfigFile.MustInt("mysql", "max_conn", 10)

	MasterDB.SetMaxIdleConns(maxIdle)
	MasterDB.SetMaxOpenConns(maxConn)

	showSQL := config.ConfigFile.MustBool("xorm", "show_sql", false)
	logLevel := config.ConfigFile.MustInt("xorm", "log_level", 1)

	MasterDB.ShowSQL(showSQL)
	MasterDB.Logger().SetLevel(core.LogLevel(logLevel))

	// 启用缓存
	// cacher := xorm.NewLRUCacher(xorm.NewMemoryStore(), 1000)
	// MasterDB.SetDefaultCacher(cacher)

	return nil
}

// StdMasterDB .
func StdMasterDB() *sql.DB {
	return MasterDB.DB().DB
}
