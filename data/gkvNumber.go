package data

import (
	"time"
)

// GkvNumber 字符串结构
// @author xuyang
// @datetime 2025-6-24 5:00
type GkvNumber struct {
	data        map[string][]byte
	expireTimes map[string]time.Time
	keyLock     *KeyLock
}

// DataGkvNumber 全局数据实例
// @author xuyang
// @datetime 2025-6-24 6:00
var DataGkvNumber = &GkvNumber{
	data:        make(map[string][]byte),
	expireTimes: make(map[string]time.Time),
	keyLock:     NewKeyLock(),
}

// Set 设置值
// @author xuyang
// @datetime 2025-6-24 6:00
func (GkvNumber *GkvNumber) Set(key string, value []byte) {
	GkvNumber.keyLock.WLockRow(key)
	defer GkvNumber.keyLock.WUnLockRow(key)
	GkvNumber.data[key] = value
	// 清除过期时间
	delete(GkvNumber.expireTimes, key)
}

// Get 获取值
// @author xuyang
// @datetime 2025-6-24 6:00
func (GkvNumber *GkvNumber) Get(key string) (val []byte, ok bool) {
	GkvNumber.keyLock.RLockRow(key)

	// 检查是否过期
	if expireTime, exists := GkvNumber.expireTimes[key]; exists {
		if time.Now().After(expireTime) {
			// 已过期，删除数据
			GkvNumber.keyLock.RUnLockRow(key)
			GkvNumber.keyLock.WLockRow(key)
			defer GkvNumber.keyLock.WUnLockRow(key)
			delete(GkvNumber.data, key)
			delete(GkvNumber.expireTimes, key)
			return nil, false
		}
	}

	val, ok = GkvNumber.data[key]
	GkvNumber.keyLock.RUnLockRow(key)
	return
}

// Delete 删除值
// @author xuyang
// @datetime 2025-6-24 6:00
func (GkvNumber *GkvNumber) Delete(key string) {
	GkvNumber.keyLock.WLockRow(key)
	defer GkvNumber.keyLock.WUnLockRow(key)
	delete(GkvNumber.data, key)
	delete(GkvNumber.expireTimes, key)
}

// GetAllKeys 获取所有key
// @author xuyang
// @datetime 2025-6-24 6:00
func (GkvNumber *GkvNumber) GetAllKeys() []string {
	GkvNumber.keyLock.tableLock.Lock()
	defer GkvNumber.keyLock.tableLock.Unlock()
	keys := make([]string, 0, len(GkvNumber.data))
	for key := range GkvNumber.data {
		keys = append(keys, key)
	}
	return keys
}

// GetAllKVs 获取所有数据
func (GkvNumber *GkvNumber) GetAllKVs() map[string][]byte {
	GkvNumber.keyLock.tableLock.Lock()
	defer GkvNumber.keyLock.tableLock.Unlock()
	return GkvNumber.data
}

// SetTime 设置过期时间(毫秒为单位)
func (GkvNumber *GkvNumber) SetTime(key string, timeMs int) {
	GkvNumber.keyLock.WLockRow(key)
	defer GkvNumber.keyLock.WUnLockRow(key)
	// 检查键是否存在
	if _, exists := GkvNumber.data[key]; exists {
		expireTime := time.Now().Add(time.Duration(timeMs) * time.Millisecond)
		GkvNumber.expireTimes[key] = expireTime
	}
}

// SetNX 仅当键不存在时才设置
func (GkvNumber *GkvNumber) SetNX(key string, value []byte) bool {
	GkvNumber.keyLock.WLockRow(key)
	defer GkvNumber.keyLock.WUnLockRow(key)

	// 检查键是否已存在
	if _, exists := GkvNumber.data[key]; exists {
		return false
	}

	// 设置新值
	GkvNumber.data[key] = value
	delete(GkvNumber.expireTimes, key)
	return true
}

// SetXX 仅当键存在时才设置
func (GkvNumber *GkvNumber) SetXX(key string, value []byte) bool {
	GkvNumber.keyLock.WLockRow(key)
	defer GkvNumber.keyLock.WUnLockRow(key)

	// 检查键是否存在
	if _, exists := GkvNumber.data[key]; !exists {
		return false
	}

	// 更新值
	GkvNumber.data[key] = value
	delete(GkvNumber.expireTimes, key)
	return true
}

// GetTTL 获取键的剩余生存时间（毫秒）
// 如果键不存在或没有设置过期时间，返回 -1
// 如果键已过期，返回 -2
func (GkvNumber *GkvNumber) GetTTL(key string) int64 {
	GkvNumber.keyLock.RLockRow(key)
	defer GkvNumber.keyLock.RUnLockRow(key)
	// 检查键是否存在
	if _, exists := GkvNumber.data[key]; !exists {
		return -1
	}
	// 检查是否有过期时间
	expireTime, exists := GkvNumber.expireTimes[key]
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
