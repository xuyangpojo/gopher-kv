package data

import (
	"time"
)

// GkvMap 映射数据结构
// @author xuyang
// @datetime 2025-7-16 21:00
type GkvMap struct {
	// 全部数据 key - filed - value
	data        map[string]map[string]string
	// 全部数据的过期时间
	expireTimes map[string]time.Time
	// 锁实例
	keyLock     *KeyLock
}

// DataGkvMap 全局数据实例
// @author xuyang
// @datetime 2025-7-16 21:00
var DataGkvMap = &GkvMap{
	data:        make(map[string]map[string]string),
	expireTimes: make(map[string]time.Time),
	keyLock:     NewKeyLock(),
}

// MSet 设置数据
// @author xuyang
// @datetime 2025-7-16 21:00
// @param key
// @param field
// @param value
// @return bool 是否设置成功
func (gkvMap *GkvMap) MSet(key, field, value string) bool {
	gkvMap.keyLock.WLockRow(key)
	defer gkvMap.keyLock.WUnLockRow(key)
	if _, exists := gkvMap.data[key]; !exists {
		gkvMap.data[key] = make(map[string]string)
	}
	gkvMap.data[key][field] = value
	delete(gkvMap.expireTimes, key)
	return true
}

// MGet 获取数据
// @author xuyang
// @datetime 2025-7-16 21:00
// @param key
// @param field
// @return value string
// @return ok bool 是否获取成功
func (gkvMap *GkvMap) MGet(key, field string) (value string, ok bool) {
	gkvMap.keyLock.RLockRow(key)
	defer gkvMap.keyLock.RUnLockRow(key)
	if expireTime, exists := gkvMap.expireTimes[key]; exists {
		if time.Now().After(expireTime) {
			return "", false
		}
	}
	fields, exists := gkvMap.data[key]
	if !exists {
		return "", false
	}
	value, ok = fields[field]
	return
}

// Delete 删除某个key或field
// @param key string
// @param field string
func (gkvMap *GkvMap) Delete(key, field string) {
	gkvMap.keyLock.WLockRow(key)
	defer gkvMap.keyLock.WUnLockRow(key)
	if fields, exists := gkvMap.data[key]; exists {
		delete(fields, field)
		if len(fields) == 0 {
			delete(gkvMap.data, key)
			delete(gkvMap.expireTimes, key)
		}
	}
}

// GetAllFields 获取某个key下所有field
// @param key string
// @return []string
func (gkvMap *GkvMap) GetAllFields(key string) []string {
	gkvMap.keyLock.RLockRow(key)
	defer gkvMap.keyLock.RUnLockRow(key)
	if expireTime, exists := gkvMap.expireTimes[key]; exists {
		if time.Now().After(expireTime) {
			return nil
		}
	}
	fields, exists := gkvMap.data[key]
	if !exists {
		return nil
	}
	result := make([]string, 0, len(fields))
	for f := range fields {
		result = append(result, f)
	}
	return result
}

// SetTime 设置过期时间(毫秒为单位)
// @param key string
// @param timeMs int
// @return bool
func (gkvMap *GkvMap) SetTime(key string, timeMs int) bool {
	gkvMap.keyLock.WLockRow(key)
	defer gkvMap.keyLock.WUnLockRow(key)
	if _, exists := gkvMap.data[key]; exists {
		expireTime := time.Now().Add(time.Duration(timeMs) * time.Millisecond)
		gkvMap.expireTimes[key] = expireTime
		return true
	}
	return false
}

// GetTTL 获取key的剩余生存时间(毫秒数)
// @param key string
// @return int64
// @return 0 已过期
// @return -1 不存在
// @return -2 没有设置过期时间
func (gkvMap *GkvMap) GetTTL(key string) int64 {
	gkvMap.keyLock.RLockRow(key)
	defer gkvMap.keyLock.RUnLockRow(key)
	if _, exists := gkvMap.data[key]; !exists {
		return -1
	}
	expireTime, exists := gkvMap.expireTimes[key]
	if !exists {
		return -2
	}
	remaining := time.Until(expireTime)
	if remaining <= 0 {
		return 0
	}
	return int64(remaining.Milliseconds())
}