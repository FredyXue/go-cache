package cache

import (
	"sync"
	"time"
)

// SetSource  set data source
type SetSource interface {
	Build() []interface{}
}

// Set set 缓存
type Set struct {
	sync.RWMutex
	cache     map[interface{}]struct{}
	expiredAt int64
	expire    int64
	source    SetSource
}

// NewSet 创建 set 缓存
func NewSet(source SetSource, expire time.Duration, opts ...interface{}) *Set {
	obj := &Set{
		expire: int64(expire.Seconds()),
		source: source,
		cache:  make(map[interface{}]struct{}),
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

// check cache set
func (s *Set) check(duration time.Duration) {
	c := time.Tick(duration)
	for next := range c {
		if s.expiredAt <= next.Unix() {
			s.Lock()
			if s.expiredAt <= next.Unix() {
				s.cache = make(map[interface{}]struct{})
			}
			s.Unlock()
		}
	}
}

// build cache
func (s *Set) build() {
	slice := s.source.Build()
	if slice != nil {
		s.cache = make(map[interface{}]struct{}, len(slice))
		for _, v := range slice {
			s.cache[v] = struct{}{}
		}
	}
	s.expiredAt = time.Now().Unix() + s.expire
}

// Build .
func (s *Set) Build(force ...bool) {
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

// Has .
func (s *Set) Has(key interface{}) bool {
	s.Build()
	s.RLock()
	_, has := s.cache[key]
	s.RUnlock()
	return has
}

// Size .
func (s *Set) Size() int {
	s.RLock()
	lens := len(s.cache)
	s.RUnlock()
	return lens
}

// Add .
func (s *Set) Add(key interface{}) {
	s.Lock()
	s.cache[key] = struct{}{}
	s.Unlock()
}

// Delete .
func (s *Set) Delete(key interface{}) {
	s.Lock()
	delete(s.cache, key)
	s.Unlock()
}

// Intersect 取交集
func (s *Set) Intersect(arr []interface{}) []interface{} {
	s.Build()
	s.RLock()
	result := make([]interface{}, 0, len(arr))
	for _, v := range arr {
		_, has := s.cache[v]
		if has {
			result = append(result, v)
		}
	}
	s.RUnlock()
	return result
}

// Union 取并集
func (s *Set) Union(arr []interface{}) []interface{} {
	s.Build()
	s.RLock()
	result := make([]interface{}, 0, len(arr)+len(s.cache))
	for k := range s.cache {
		result = append(result, k)
	}
	for _, v := range arr {
		_, has := s.cache[v]
		if !has {
			result = append(result, v)
		}
	}
	s.RUnlock()
	return result
}

// Diff 取差集
func (s *Set) Diff(arr []interface{}) []interface{} {
	s.Build()
	s.RLock()
	arrSet := make(map[interface{}]struct{}, len(arr))
	for _, v := range arr {
		arrSet[v] = struct{}{}
	}
	result := make([]interface{}, 0, len(s.cache))
	for k := range s.cache {
		_, has := arrSet[k]
		if !has {
			result = append(result, k)
		}
	}
	s.RUnlock()
	return result
}
