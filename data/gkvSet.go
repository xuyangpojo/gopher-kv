import (
	"time"
)

// GkvSet 集合结构
// @author xuyang
// @datetime 2025-7-16 21:00
type GkvSet struct {
	// 全部数据 key -> set成员集合
	data        map[string]map[string]struct{}
	// 全部数据的过期时间
	expireTimes map[string]time.Time
	// 锁实例
	keyLock     *KeyLock
}

// DataGkvSet 全局数据实例
// @author xuyang
// @datetime 2025-7-16 21:00
var DataGkvSet = &GkvSet{
	data:        make(map[string]map[string]struct{}),
	expireTimes: make(map[string]time.Time),
	keyLock:     NewKeyLock(),
}

// Add 向集合添加成员
// @param key string 集合名
// @param member string 成员
func (gkvSet *GkvSet) Add(key, member string) {
	gkvSet.keyLock.WLockRow(key)
	defer gkvSet.keyLock.WUnLockRow(key)
	if _, exists := gkvSet.data[key]; !exists {
		gkvSet.data[key] = make(map[string]struct{})
	}
	gkvSet.data[key][member] = struct{}{}
	delete(gkvSet.expireTimes, key)
}

// Remove 从集合移除成员
// @param key string 集合名
// @param member string 成员
func (gkvSet *GkvSet) Remove(key, member string) {
	gkvSet.keyLock.WLockRow(key)
	defer gkvSet.keyLock.WUnLockRow(key)
	if members, exists := gkvSet.data[key]; exists {
		delete(members, member)
		if len(members) == 0 {
			delete(gkvSet.data, key)
			delete(gkvSet.expireTimes, key)
		}
	}
}

// IsMember 判断成员是否存在
// @param key string 集合名
// @param member string 成员
// @return bool 是否存在
func (gkvSet *GkvSet) IsMember(key, member string) bool {
	gkvSet.keyLock.RLockRow(key)
	defer gkvSet.keyLock.RUnLockRow(key)
	if expireTime, exists := gkvSet.expireTimes[key]; exists {
		if time.Now().After(expireTime) {
			return false
		}
	}
	members, exists := gkvSet.data[key]
	if !exists {
		return false
	}
	_, ok := members[member]
	return ok
}

// GetAllMembers 获取集合所有成员
// @param key string 集合名
// @return []string 所有成员
func (gkvSet *GkvSet) GetAllMembers(key string) []string {
	gkvSet.keyLock.RLockRow(key)
	defer gkvSet.keyLock.RUnLockRow(key)
	if expireTime, exists := gkvSet.expireTimes[key]; exists {
		if time.Now().After(expireTime) {
			return nil
		}
	}
	members, exists := gkvSet.data[key]
	if !exists {
		return nil
	}
	result := make([]string, 0, len(members))
	for m := range members {
		result = append(result, m)
	}
	return result
}

// SetTime 设置过期时间(毫秒为单位)
// @param key string 集合名
// @param timeMs int 毫秒
// @return bool 是否设置成功
func (gkvSet *GkvSet) SetTime(key string, timeMs int) bool {
	gkvSet.keyLock.WLockRow(key)
	defer gkvSet.keyLock.WUnLockRow(key)
	if _, exists := gkvSet.data[key]; exists {
		expireTime := time.Now().Add(time.Duration(timeMs) * time.Millisecond)
		gkvSet.expireTimes[key] = expireTime
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
func (gkvSet *GkvSet) GetTTL(key string) int64 {
	gkvSet.keyLock.RLockRow(key)
	defer gkvSet.keyLock.RUnLockRow(key)
	if _, exists := gkvSet.data[key]; !exists {
		return -1
	}
	expireTime, exists := gkvSet.expireTimes[key]
	if !exists {
		return -2
	}
	remaining := time.Until(expireTime)
	if remaining <= 0 {
		return 0
	}
	return int64(remaining.Milliseconds())
}
