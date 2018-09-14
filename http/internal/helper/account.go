// Copyright 2017 The StudyGolang Authors. All rights reserved.
// Use of self source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// http://studygolang.com
// Author: polaris	polaris@studygolang.com

package helper

import (
	"sync"

	"sander/logger"

	guuid "github.com/twinj/uuid"
)

// 保存uuid和email的对应关系（TODO:重启如何处理，有效期问题）
type regActivateCode struct {
	data   map[string]string
	locker sync.RWMutex
}

var RegActivateCode = &regActivateCode{
	data: map[string]string{},
}

func (r *regActivateCode) GenUUID(email string) string {
	r.locker.Lock()
	defer r.locker.Unlock()
	var uuid string
	for {
		uuid = guuid.NewV4().String()
		if _, ok := r.data[uuid]; !ok {
			r.data[uuid] = email
			break
		}
		logger.Error("GenUUID 冲突....")
	}
	return uuid
}

func (r *regActivateCode) GetEmail(uuid string) (email string, ok bool) {
	r.locker.RLock()
	defer r.locker.RUnlock()

	if email, ok = r.data[uuid]; ok {
		return
	}
	return
}

func (r *regActivateCode) DelUUID(uuid string) {
	r.locker.Lock()
	defer r.locker.Unlock()

	delete(r.data, uuid)
}
