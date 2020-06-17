package cache

import (
	"sync"
	"time"
)

// MapSource map data source
// Build() failed return nil.  判断 nil, 不会更新数据.  SetSource  ListSource 同
type MapSource interface {
	Build() map[interface{}]interface{}
}

// Map map 缓存
type Map struct {
	sync.RWMutex
	cache     map[interface{}]interface{}
	expiredAt int64
	expire    int64
	source    MapSource
}

// NewMap 创建 map 缓存
// expire 缓存保留时间
// opts[0]  check duration   默认 1h
func NewMap(source MapSource, expire time.Duration, opts ...interface{}) *Map {
	obj := &Map{
		expire: int64(expire.Seconds()),
		source: source,
		cache:  make(map[interface{}]interface{}),
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
func (m *Map) check(duration time.Duration) {
	c := time.Tick(duration)
	for next := range c {
		if m.expiredAt <= next.Unix() {
			m.Lock()
			if m.expiredAt <= next.Unix() {
				m.cache = make(map[interface{}]interface{})
			}
			m.Unlock()
		}
	}
}

// build cache
func (m *Map) build() {
	maps := m.source.Build()
	if maps != nil {
		m.cache = make(map[interface{}]interface{}, len(maps))
		for k, v := range maps {
			m.cache[k] = v
		}
	}
	m.expiredAt = time.Now().Unix() + m.expire
}

// Build .
// force force build
func (m *Map) Build(force ...bool) {
	if len(force) > 0 && force[0] {
		m.Lock()
		m.build()
		m.Unlock()
		return
	}
	now := time.Now().Unix()
	exp := m.expiredAt
	if exp <= now {
		m.Lock()
		if exp == m.expiredAt { // 二次确认   for  parallel build
			m.build()
		}
		m.Unlock()
		return
	}
	// 预重建
	preduration := prebuildDuration(m.expire)
	if exp-now <= preduration { // min 1s
		go func() {
			m.Lock()
			if exp == m.expiredAt {
				m.build()
			}
			m.Unlock()
		}()
	}
}

// Get get value
func (m *Map) Get(key interface{}) (interface{}, bool) {
	m.Build()
	m.RLock()
	val, has := m.cache[key]
	m.RUnlock()
	return val, has
}

// GetBool .
func (m *Map) GetBool(key interface{}) bool {
	val, has := m.Get(key)
	if has {
		return val.(bool)
	}
	return false
}

// GetFloat64 .
func (m *Map) GetFloat64(key interface{}) float64 {
	val, has := m.Get(key)
	if has {
		return val.(float64)
	}
	return 0
}

// GetInt64 .
func (m *Map) GetInt64(key interface{}) int64 {
	val, has := m.Get(key)
	if has {
		return val.(int64)
	}
	return 0
}

// GetInt .
func (m *Map) GetInt(key interface{}) int {
	val, has := m.Get(key)
	if has {
		return val.(int)
	}
	return 0
}

// GetString .
func (m *Map) GetString(key interface{}) string {
	val, has := m.Get(key)
	if has {
		return val.(string)
	}
	return ""
}

// Size .
func (m *Map) Size() int {
	m.RLock()
	lens := len(m.cache)
	m.RUnlock()
	return lens
}

// Set set value
func (m *Map) Set(key interface{}, val interface{}) {
	m.Lock()
	m.cache[key] = val
	m.Unlock()
}

// Delete delete value
func (m *Map) Delete(key interface{}) {
	m.Lock()
	delete(m.cache, key)
	m.Unlock()
}
