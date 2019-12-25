// Copyright 2019 orivil.com. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found at https://mit-license.org.

package elector

import (
	"github.com/go-redis/redis"
	"time"
)

type Connector interface {
	// 设置集群 master 名称, 且 master 会自动过期
	SetMaster(name string, expire time.Duration) error

	// 获得集群 master 名称, 如果未设置或设置的 master 已过期则返回空 name
	GetMaster() (name string, err error)

	// 设置集群为锁住状态, expire 参数用于设置上锁最大时间,
	// 返回值 success 为 true 则表示集群上锁成功, false 表示集群已上锁不可操作
	// 该方法不会阻止程序执行
	Lock(expire time.Duration) (success bool, err error)

	// 设置集群为开放状态, Unlock 并不能保证一定成功或一定被执行, 当 Unlock 失效时
	// 可根 Lock 方法的 expire 时间自动解锁, 避免整个集群被锁住
	Unlock() error
}

type RedisConnector struct {
	rds    *redis.Client
	prefix string
}

func NewRedisConnector(keyPrefix string, rds *redis.Client) *RedisConnector {
	return &RedisConnector{rds: rds, prefix: keyPrefix}
}

func (r *RedisConnector) SetMaster(name string, expire time.Duration) error {
	return r.rds.Set(r.prefix+"master", name, expire).Err()
}

func (r *RedisConnector) GetMaster() (name string, err error) {
	return r.rds.Get(r.prefix + "master").Result()
}

func (r *RedisConnector) Lock(expire time.Duration) (success bool, err error) {
	return r.rds.SetNX(r.prefix+"locker", 1, expire).Result()
}

func (r *RedisConnector) Unlock() error {
	return r.rds.Del(r.prefix + "locker").Err()
}
