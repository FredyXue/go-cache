# go-cache
a simple go cache library


### Map
> map[interface{}]interface{} cache
```go
type MapSource struct{}

func (s *MapSource) Build() map[interface{}]interface{} {
	return map[interface{}]interface{}{
		"1": 1,
		"2": int64(2),
		"3": "3",
		"4": true,
		"5": 5.0,
	}
}

maps := NewMap(&MapSource{}, time.Minute*30)
val, has := store.Get(1)  // val = 1
```

### Set
> map[interface{}]struct{} cache, support Intersect,Union,Diff
```go
type SetSource struct{}

func (s *SetSource) Build() []interface{} {
	return []interface{}{1, 2, 3, 4, 5}
}

sets := NewSet(&SetSource{}, time.Minute*30)
has := sets.Has(1)  // has = true
```

### List
> []interface{} cache
```go
type ListSource struct{}

func (s *ListSource) Build() []interface{} {
	return []interface{}{0, 1, 2, 3, 4, 5}
}

list := NewList(&ListSource{}, time.Minute*30)
slice := list.Copy()  // slice = [0, 1, 2, 3, 4, 5]
```


### Store
> key interface{} -> value interface{} cache
```go
type StoreSource struct{}

func (s *StoreSource) Build(key interface{}, opts ...interface{}) interface{} {
	i := key.(int)
	return []interface{}{i, i + 1}
}

store := NewStore(&StoreSource{}, time.Minute*30)
val, has := store.Get(1)  // val = [1, 2]
```