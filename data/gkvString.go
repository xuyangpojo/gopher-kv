package data

import (
	"time"
)

// GkvString 字符串结构
// @author xuyang
// @datetime 2025-6-24 5:00
type GkvString struct {
	// 全部数据
	data        map[string][]byte
	// 全部数据的过期时间
	expireTimes map[string]time.Time
	// 锁实例
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

// Set 设置键值对
// @author xuyang
// @datetime 2025-6-24 6:00
// @param key string 键
// @param value []byte 值
func (gkvString *GkvString) Set(key string, value []byte) {
	gkvString.keyLock.WLockRow(key)
	defer gkvString.keyLock.WUnLockRow(key)
	gkvString.data[key] = value
	// 清除旧的过期时间
	delete(gkvString.expireTimes, key)
}

// Get 获取某个键对应的值
// @author xuyang
// @datetime 2025-6-24 6:00
// @param key string 键
// @return val []byte 值
// @return ok bool 是否成功获取到值
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

// Delete 删除某个键对应的值
// @author xuyang
// @datetime 2025-6-24 6:00
// @param key string 键
// 注意: 键不存在时不报错
func (gkvString *GkvString) Delete(key string) {
	gkvString.keyLock.WLockRow(key)
	defer gkvString.keyLock.WUnLockRow(key)
	delete(gkvString.data, key)
	delete(gkvString.expireTimes, key)
}

// GetAllKeys 获取所有key
// @author xuyang
// @datetime 2025-6-24 6:00
// @return []string 所有的key
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
// @author xuyang
// @datetime 2025-7-16 21:00
// @return map[string]string 所有的键值对数据
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
// @author xuyang
// @datetime 2025-7-16 21:00
// @param key string 键
// @param timeMs int 毫秒
// @return bool 是否设置成功
func (gkvString *GkvString) SetTime(key string, timeMs int) bool {
	gkvString.keyLock.WLockRow(key)
	defer gkvString.keyLock.WUnLockRow(key)
	if _, exists := gkvString.data[key]; exists {
		expireTime := time.Now().Add(time.Duration(timeMs) * time.Millisecond)
		gkvString.expireTimes[key] = expireTime
		return true
	} else {
		return false
	}
}

// SetNX 仅当键不存在时才设置
// @author xuyang
// @datetime 2025-7-16 21:00
// @param key string 键
// @param value []byte 值
// @return bool 是否设置成功
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
// @author xuyang
// @datetime 2025-7-16 21:00
// @param key string 键
// @param value []byte 值
// @return bool 是否设置成功
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

// GetTTL 获取键的剩余生存时间(毫秒数)
// @author xuyang
// @datetime 2025-7-16
// @param key string 键
// @return int64 剩余生存时间
// @return 0 键已过期
// @return -1 键不存在
// @return -2 键没有设置过期时间
func (gkv *GkvString) GetTTL(key string) int64 {
	gkv.keyLock.RLockRow(key)
	defer gkv.keyLock.RUnLockRow(key)
	if _, exists := gkv.data[key]; !exists {
		return -1
	}
	expireTime, exists := gkv.expireTimes[key]
	if !exists {
		return -2
	}
	remaining := time.Until(expireTime)
	if remaining <= 0 {
		return 0
	}
	return int64(remaining.Milliseconds())
}
