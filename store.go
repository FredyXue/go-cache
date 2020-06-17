package cache

import (
	"sync"
	"time"
)

// StoreSource map data source
// Build() failed return nil, 不会更新数据.  SetSource  ListSource 同
type StoreSource interface {
	Build(key interface{}, opts ...interface{}) interface{}
}

// Store kv 缓存
type Store struct {
	sync.RWMutex
	cache  map[interface{}]*storeEelment
	expire int64
	source StoreSource
}

type storeEelment struct {
	sync.RWMutex
	expiredAt int64
	value     interface{}
}

// NewStore 创建 kv 缓存
func NewStore(source StoreSource, expire time.Duration, opts ...interface{}) *Store {
	obj := &Store{
		expire: int64(expire.Seconds()),
		source: source,
		cache:  make(map[interface{}]*storeEelment),
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

// check cache map
func (m *Store) check(duration time.Duration) {
	c := time.Tick(duration)
	for next := range c {
		now := next.Unix()
		for k, v := range m.cache {
			if v.expiredAt <= now {
				m.Lock()
				if v.expiredAt <= now {
					delete(m.cache, k)
				}
				m.Unlock()
			}
		}
	}
}

// build cache
func (m *Store) build(val *storeEelment, key interface{}, opts ...interface{}) {
	result := m.source.Build(key, opts)
	if result != nil {
		val.value = result
	}
	val.expiredAt = time.Now().Unix() + m.expire // 延续之前的值 or 保留 nil 值
}

// Build .
func (m *Store) Build(force bool, key interface{}, opts ...interface{}) {
	// check exist
	val, has := m.cache[key]
	if !has {
		m.Lock()
		val, has = m.cache[key]
		if !has { // check value
			val = &storeEelment{
				expiredAt: 0,
				value:     nil, // 预创建，避免 不存在的数据 频繁 build
			}
			m.cache[key] = val
		}
		m.Unlock()
	}

	// force build
	if force {
		val.Lock()
		m.build(val, key, opts...)
		val.Unlock()
		return
	}

	// check expireAt
	now := time.Now().Unix()
	exp := val.expiredAt
	if exp <= now {
		val.Lock()
		if exp == val.expiredAt { // check value
			m.build(val, key, opts...)
		}
		val.Unlock()
		return
	}
	preduration := prebuildDuration(m.expire)
	if exp-now <= preduration {
		go func() {
			val.Lock()
			if exp == val.expiredAt {
				m.build(val, key, opts...)
			}
			val.Unlock()
		}()
	}
}

// Get get value
func (m *Store) Get(key interface{}, opts ...interface{}) (interface{}, bool) {
	m.Build(false, key, opts...)
	m.RLock()
	val, has := m.cache[key]
	m.RUnlock()
	if has && val.value == nil {
		has = false
	}
	return val.value, has
}

// Size .
func (m *Store) Size() int {
	m.RLock()
	lens := len(m.cache)
	m.RUnlock()
	return lens
}
