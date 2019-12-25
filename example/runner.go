// Copyright 2019 orivil.com. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found at https://mit-license.org.

package main

import (
	"fmt"
	"github.com/orivil/elector"
	"strconv"
	"sync"
	"time"
)

type expireValue struct {
	expireAt time.Time
	v        interface{}
}

func newExpireValue(expire time.Duration, val interface{}) *expireValue {
	return &expireValue{
		expireAt: time.Now().Add(expire),
		v:        val,
	}
}

func (ev *expireValue) Val() interface{} {
	if ev.expireAt.After(time.Now()) {
		return ev.v
	} else {
		return nil
	}
}

type connector struct {
	master *expireValue
	locker *expireValue
	mu     sync.Mutex
}

func (c *connector) SetMaster(name string, expire time.Duration) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.master = newExpireValue(expire, name)
	return nil
}

func (c *connector) GetMaster() (name string, err error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.master != nil {
		name, _ = c.master.Val().(string)
	}
	return name, nil
}

func (c *connector) Lock(expire time.Duration) (success bool, err error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	locked := c.locker != nil && c.locker.Val() != nil
	if locked {
		return false, nil
	} else {
		c.locker = newExpireValue(expire, true)
		return true, nil
	}
}

func (c *connector) Unlock() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.locker = nil
	return nil
}

func main() {
	conn := &connector{}
	times := make(map[int]int, 10)
	for i := 0; i < 10; i++ {
		i := i
		runner := elector.NewRunner(strconv.Itoa(i), conn)
		go func() {
			ticker := time.NewTicker(2 * time.Second)
			for range ticker.C {
				success := false
				_ = runner.MasterRun(5*time.Second, 6*time.Second, func() {
					success = true
				})
				if success {
					fmt.Printf("%d %v\n", i, success)
				}
				if success {
					times[i]++
					if times[i] >= 2 {
						return
					}
				}
			}
		}()
	}
	time.Sleep(60 * time.Second)
}
