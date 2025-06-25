package data

import (
	"time"
)

// GkvList 字符串结构
// @author xuyang
// @datetime 2025-6-24 5:00
type GkvList struct {
	data        map[string][]byte
	expireTimes map[string]time.Time
	keyLock     *KeyLock
}

// DataGkvList 全局数据实例
// @author xuyang
// @datetime 2025-6-24 6:00
var DataGkvList = &GkvList{
	data:        make(map[string][]byte),
	expireTimes: make(map[string]time.Time),
	keyLock:     NewKeyLock(),
}

// Set 设置值
// @author xuyang
// @datetime 2025-6-24 6:00
func (GkvList *GkvList) Set(key string, value []byte) {
	GkvList.keyLock.WLockRow(key)
	defer GkvList.keyLock.WUnLockRow(key)
	GkvList.data[key] = value
	// 清除过期时间
	delete(GkvList.expireTimes, key)
}

// Get 获取值
// @author xuyang
// @datetime 2025-6-24 6:00
func (GkvList *GkvList) Get(key string) (val []byte, ok bool) {
	GkvList.keyLock.RLockRow(key)

	// 检查是否过期
	if expireTime, exists := GkvList.expireTimes[key]; exists {
		if time.Now().After(expireTime) {
			// 已过期，删除数据
			GkvList.keyLock.RUnLockRow(key)
			GkvList.keyLock.WLockRow(key)
			defer GkvList.keyLock.WUnLockRow(key)
			delete(GkvList.data, key)
			delete(GkvList.expireTimes, key)
			return nil, false
		}
	}

	val, ok = GkvList.data[key]
	GkvList.keyLock.RUnLockRow(key)
	return
}

// Delete 删除值
// @author xuyang
// @datetime 2025-6-24 6:00
func (GkvList *GkvList) Delete(key string) {
	GkvList.keyLock.WLockRow(key)
	defer GkvList.keyLock.WUnLockRow(key)
	delete(GkvList.data, key)
	delete(GkvList.expireTimes, key)
}

// GetAllKeys 获取所有key
// @author xuyang
// @datetime 2025-6-24 6:00
func (GkvList *GkvList) GetAllKeys() []string {
	GkvList.keyLock.tableLock.Lock()
	defer GkvList.keyLock.tableLock.Unlock()
	keys := make([]string, 0, len(GkvList.data))
	for key := range GkvList.data {
		keys = append(keys, key)
	}
	return keys
}

// GetAllKVs 获取所有数据
func (GkvList *GkvList) GetAllKVs() map[string][]byte {
	GkvList.keyLock.tableLock.Lock()
	defer GkvList.keyLock.tableLock.Unlock()
	return GkvList.data
}

// SetTime 设置过期时间(毫秒为单位)
func (GkvList *GkvList) SetTime(key string, timeMs int) {
	GkvList.keyLock.WLockRow(key)
	defer GkvList.keyLock.WUnLockRow(key)
	// 检查键是否存在
	if _, exists := GkvList.data[key]; exists {
		expireTime := time.Now().Add(time.Duration(timeMs) * time.Millisecond)
		GkvList.expireTimes[key] = expireTime
	}
}

// SetNX 仅当键不存在时才设置
func (GkvList *GkvList) SetNX(key string, value []byte) bool {
	GkvList.keyLock.WLockRow(key)
	defer GkvList.keyLock.WUnLockRow(key)

	// 检查键是否已存在
	if _, exists := GkvList.data[key]; exists {
		return false
	}

	// 设置新值
	GkvList.data[key] = value
	delete(GkvList.expireTimes, key)
	return true
}

// SetXX 仅当键存在时才设置
func (GkvList *GkvList) SetXX(key string, value []byte) bool {
	GkvList.keyLock.WLockRow(key)
	defer GkvList.keyLock.WUnLockRow(key)

	// 检查键是否存在
	if _, exists := GkvList.data[key]; !exists {
		return false
	}

	// 更新值
	GkvList.data[key] = value
	delete(GkvList.expireTimes, key)
	return true
}

// GetTTL 获取键的剩余生存时间（毫秒）
// 如果键不存在或没有设置过期时间，返回 -1
// 如果键已过期，返回 -2
func (GkvList *GkvList) GetTTL(key string) int64 {
	GkvList.keyLock.RLockRow(key)
	defer GkvList.keyLock.RUnLockRow(key)
	// 检查键是否存在
	if _, exists := GkvList.data[key]; !exists {
		return -1
	}
	// 检查是否有过期时间
	expireTime, exists := GkvList.expireTimes[key]
	if !exists {
		return -1
	}
	// 计算剩余时间
	remaining := time.Until(expireTime)
	if remaining <= 0 {
		return -2
	}
	return int64(remaining.Milliseconds())
}
