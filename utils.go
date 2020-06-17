package cache

// 预重建 时间
// 默认 10分之一， 最少 1s
func prebuildDuration(expire int64) int64 {
	duration := expire / 10
	if duration < 1 {
		duration = 1
	}
	return duration
}
