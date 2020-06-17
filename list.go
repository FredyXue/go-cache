package cache

import (
	"sync"
	"time"
)

// ListSource  list data source
type ListSource interface {
	Build() []interface{}
}

// List list 缓存
type List struct {
	sync.RWMutex
	cache     []interface{}
	expiredAt int64
	expire    int64
	source    ListSource
}

// NewList 创建 list 缓存
func NewList(source ListSource, expire time.Duration, opts ...interface{}) *List {
	obj := &List{
		expire: int64(expire.Seconds()),
		source: source,
		cache:  make([]interface{}, 0),
	}
	duration := time.Hour // 默认 1h
	if len(opts) > 0 {
		param, ok := opts[0].(time.Duration)
		if !ok {
			panic("params must be time.Duration")
		}
		duration = param
	}
	go obj.check(duration)
	return obj
}

// check cache list
func (s *List) check(duration time.Duration) {
	c := time.Tick(duration)
	for next := range c {
		if s.expiredAt <= next.Unix() {
			s.Lock()
			if s.expiredAt <= next.Unix() {
				s.cache = make([]interface{}, 0)
			}
			s.Unlock()
		}
	}
}

// build cache
func (s *List) build() {
	slice := s.source.Build()
	if slice != nil {
		s.cache = make([]interface{}, len(slice))
		s.cache = slice
	}
	s.expiredAt = time.Now().Unix() + s.expire
}

// Build .
func (s *List) Build(force ...bool) {
	if len(force) > 0 && force[0] {
		s.Lock()
		s.build()
		s.Unlock()
		return
	}
	now := time.Now().Unix()
	exp := s.expiredAt
	if exp <= now {
		s.Lock()
		if exp == s.expiredAt {
			s.build()
		}
		s.Unlock()
		return
	}
	preduration := prebuildDuration(s.expire)
	if exp-now <= preduration {
		go func() {
			s.Lock()
			if exp == s.expiredAt {
				s.build()
			}
			s.Unlock()
		}()
	}
}

// Get 获取原 slice
func (s *List) Get() []interface{} {
	s.Build()
	return s.cache
}

// Copy 获取副本
func (s *List) Copy() []interface{} {
	s.Build()
	s.RLock()
	slice := make([]interface{}, len(s.cache))
	copy(slice, s.cache)
	s.RUnlock()
	return slice
}

// Length .
func (s *List) Length() int {
	s.RLock()
	lens := len(s.cache)
	s.RUnlock()
	return lens
}
