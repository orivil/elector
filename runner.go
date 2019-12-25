// Copyright 2019 orivil.com. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found at https://mit-license.org.

package elector

import "time"

type Runner struct {
	Name string
	conn Connector
}

// 创建集群 runner, uniqueName 在集群中必须具有唯一性, conn 用于存取集群 master 名称
// 以及维持集群上锁状态
func NewRunner(uniqueName string, conn Connector) *Runner {
	return &Runner{
		Name: uniqueName,
		conn: conn,
	}
}

// 在集群中运行一次 call 函数, 集群会自动选择一个 master 程序运行该函数, 该方法不会阻塞程序.
// 参数 lockerExpire 用于设置集群锁的最大上锁时间, masterExpire 用于设置 master 状态的最大时间,
// lockerExpire 及 masterExpire 参数都是为了避免 master 退出或阻塞,
func (rer *Runner) MasterRun(lockerExpire, masterExpire time.Duration, call func()) error {
	success, err := rer.conn.Lock(lockerExpire)
	if err != nil {
		return err
	}
	if !success {
		return nil
	}
	defer rer.conn.Unlock()
	master, err := rer.conn.GetMaster()
	if err != nil {
		return err
	}
	if master == "" || master == rer.Name {
		call()
		return rer.conn.SetMaster(rer.Name, masterExpire)
	} else {
		return nil
	}
}
