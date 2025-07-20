import (
	"time"
	"sort"
)

// GkvZSet 有序集合结构
// @author xuyang
// @datetime 2025-7-16 21:00
type GkvZSet struct {
	// 全部数据 key -> member -> score
	data        map[string]map[string]float64
	// 全部数据的过期时间
	expireTimes map[string]time.Time
	// 锁实例
	keyLock     *KeyLock
}

// DataGkvZSet 全局数据实例
// @author xuyang
// @datetime 2025-7-16 21:00
var DataGkvZSet = &GkvZSet{
	data:        make(map[string]map[string]float64),
	expireTimes: make(map[string]time.Time),
	keyLock:     NewKeyLock(),
}

// Add 添加成员及分数
// @param key string 集合名
// @param member string 成员
// @param score float64 分数
func (gkvZSet *GkvZSet) Add(key, member string, score float64) {
	gkvZSet.keyLock.WLockRow(key)
	defer gkvZSet.keyLock.WUnLockRow(key)
	if _, exists := gkvZSet.data[key]; !exists {
		gkvZSet.data[key] = make(map[string]float64)
	}
	gkvZSet.data[key][member] = score
	delete(gkvZSet.expireTimes, key)
}

// Remove 移除成员
// @param key string 集合名
// @param member string 成员
func (gkvZSet *GkvZSet) Remove(key, member string) {
	gkvZSet.keyLock.WLockRow(key)
	defer gkvZSet.keyLock.WUnLockRow(key)
	if members, exists := gkvZSet.data[key]; exists {
		delete(members, member)
		if len(members) == 0 {
			delete(gkvZSet.data, key)
			delete(gkvZSet.expireTimes, key)
		}
	}
}

// Score 获取成员分数
// @param key string 集合名
// @param member string 成员
// @return float64 分数
// @return bool 是否存在
func (gkvZSet *GkvZSet) Score(key, member string) (float64, bool) {
	gkvZSet.keyLock.RLockRow(key)
	defer gkvZSet.keyLock.RUnLockRow(key)
	if expireTime, exists := gkvZSet.expireTimes[key]; exists {
		if time.Now().After(expireTime) {
			return 0, false
		}
	}
	members, exists := gkvZSet.data[key]
	if !exists {
		return 0, false
	}
	score, ok := members[member]
	return score, ok
}

// RangeByScore 按分数区间获取成员（升序）
// @param key string 集合名
// @param min, max float64 分数区间
// @return []string 成员
func (gkvZSet *GkvZSet) RangeByScore(key string, min, max float64) []string {
	gkvZSet.keyLock.RLockRow(key)
	defer gkvZSet.keyLock.RUnLockRow(key)
	if expireTime, exists := gkvZSet.expireTimes[key]; exists {
		if time.Now().After(expireTime) {
			return nil
		}
	}
	members, exists := gkvZSet.data[key]
	if !exists {
		return nil
	}
	type kv struct {
		member string
		score  float64
	}
	var arr []kv
	for m, s := range members {
		if s >= min && s <= max {
			arr = append(arr, kv{m, s})
		}
	}
	sort.Slice(arr, func(i, j int) bool {
		return arr[i].score < arr[j].score
	})
	result := make([]string, len(arr))
	for i, v := range arr {
		result[i] = v.member
	}
	return result
}

// SetTime 设置过期时间(毫秒为单位)
// @param key string 集合名
// @param timeMs int 毫秒
// @return bool 是否设置成功
func (gkvZSet *GkvZSet) SetTime(key string, timeMs int) bool {
	gkvZSet.keyLock.WLockRow(key)
	defer gkvZSet.keyLock.WUnLockRow(key)
	if _, exists := gkvZSet.data[key]; exists {
		expireTime := time.Now().Add(time.Duration(timeMs) * time.Millisecond)
		gkvZSet.expireTimes[key] = expireTime
		return true
	}
	return false
}

// GetTTL 获取key的剩余生存时间(毫秒数)
// @param key string 集合名
// @return int64 剩余生存时间
// @return 0 已过期
// @return -1 不存在
// @return -2 没有设置过期时间
func (gkvZSet *GkvZSet) GetTTL(key string) int64 {
	gkvZSet.keyLock.RLockRow(key)
	defer gkvZSet.keyLock.RUnLockRow(key)
	if _, exists := gkvZSet.data[key]; !exists {
		return -1
	}
	expireTime, exists := gkvZSet.expireTimes[key]
	if !exists {
		return -2
	}
	remaining := time.Until(expireTime)
	if remaining <= 0 {
		return 0
	}
	return int64(remaining.Milliseconds())
}
