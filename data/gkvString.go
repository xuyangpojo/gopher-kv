package data

import (
	"time"
)

// GkvString 字符串结构
// @author xuyang
// @datetime 2025-6-24 5:00
type GkvString struct {
	data        map[string][]byte
	expireTimes map[string]time.Time
	keyLock     *KeyLock
}

// DataGkvString 全局数据实例
// @author xuyang
// @datetime 2025-6-24 6:00
var DataGkvString = &GkvString{
	data:        make(map[string][]byte),
	expireTimes: make(map[string]time.Time),
	keyLock:     NewKeyLock(),
}

// Set 设置值
// @author xuyang
// @datetime 2025-6-24 6:00
func (gkvString *GkvString) Set(key string, value []byte) {
	gkvString.keyLock.WLockRow(key)
	defer gkvString.keyLock.WUnLockRow(key)
	gkvString.data[key] = value
	// 清除过期时间
	delete(gkvString.expireTimes, key)
}

// Get 获取值
// @author xuyang
// @datetime 2025-6-24 6:00
func (gkvString *GkvString) Get(key string) (val []byte, ok bool) {
	gkvString.keyLock.RLockRow(key)
	// 检查是否过期
	if expireTime, exists := gkvString.expireTimes[key]; exists {
		if time.Now().After(expireTime) {
			// 已过期，删除数据
			gkvString.keyLock.RUnLockRow(key)
			gkvString.keyLock.WLockRow(key)
			defer gkvString.keyLock.WUnLockRow(key)
			delete(gkvString.data, key)
			delete(gkvString.expireTimes, key)
			return nil, false
		}
	}
	val, ok = gkvString.data[key]
	gkvString.keyLock.RUnLockRow(key)
	return
}

// Delete 删除值
// @author xuyang
// @datetime 2025-6-24 6:00
func (gkvString *GkvString) Delete(key string) {
	gkvString.keyLock.WLockRow(key)
	defer gkvString.keyLock.WUnLockRow(key)
	delete(gkvString.data, key)
	delete(gkvString.expireTimes, key)
}

// GetAllKeys 获取所有key
// @author xuyang
// @datetime 2025-6-24 6:00
func (gkvString *GkvString) GetAllKeys() []string {
	gkvString.keyLock.tableLock.Lock()
	defer gkvString.keyLock.tableLock.Unlock()
	keys := make([]string, 0, len(gkvString.data))
	for key := range gkvString.data {
		keys = append(keys, key)
	}
	return keys
}

// GetAllKVs 获取所有数据
func (gkvString *GkvString) GetAllKVs() (result map[string]string) {
	gkvString.keyLock.tableLock.Lock()
	defer gkvString.keyLock.tableLock.Unlock()
	result = make(map[string]string)
	for s, bs := range gkvString.data {
		result[s] = string(bs)
	}
	return
}

// SetTime 设置过期时间(毫秒为单位)
func (gkvString *GkvString) SetTime(key string, timeMs int) {
	gkvString.keyLock.WLockRow(key)
	defer gkvString.keyLock.WUnLockRow(key)
	if _, exists := gkvString.data[key]; exists {
		expireTime := time.Now().Add(time.Duration(timeMs) * time.Millisecond)
		gkvString.expireTimes[key] = expireTime
	}
}

// SetNX 仅当键不存在时才设置
func (gkvString *GkvString) SetNX(key string, value []byte) bool {
	gkvString.keyLock.WLockRow(key)
	defer gkvString.keyLock.WUnLockRow(key)
	if _, exists := gkvString.data[key]; exists {
		return false
	}
	gkvString.data[key] = value
	delete(gkvString.expireTimes, key)
	return true
}

// SetXX 仅当键存在时才设置
func (gkvString *GkvString) SetXX(key string, value []byte) bool {
	gkvString.keyLock.WLockRow(key)
	defer gkvString.keyLock.WUnLockRow(key)
	if _, exists := gkvString.data[key]; !exists {
		return false
	}
	gkvString.data[key] = value
	delete(gkvString.expireTimes, key)
	return true
}

// GetTTL 获取键的剩余生存时间（毫秒）
// 如果键不存在或没有设置过期时间，返回 -1
// 如果键已过期，返回 -2
func (gkvString *GkvString) GetTTL(key string) int64 {
	gkvString.keyLock.RLockRow(key)
	defer gkvString.keyLock.RUnLockRow(key)
	if _, exists := gkvString.data[key]; !exists {
		return -1
	}
	expireTime, exists := gkvString.expireTimes[key]
	if !exists {
		return -1
	}
	remaining := time.Until(expireTime)
	if remaining <= 0 {
		return -2
	}
	return int64(remaining.Milliseconds())
}
