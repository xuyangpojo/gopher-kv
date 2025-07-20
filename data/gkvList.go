package data

import (
	"time"
)

// GkvList 链表结构
// @author xuyang
// @datetime 2025-6-24 5:00
type GkvList struct {
	data        map[string][]string
	expireTimes map[string]time.Time
	k859
}                                                                                                                                                                                                                                                                                                                                                                                                                                             

// DataGkvList 全局数据实例
// @author xuyang
// @datetime 2025-6-24 6:00
var DataGkvList = &GkvList{
	data:        make(map[string][]string),
	expireTimes: make(map[string]time.Time),
	keyLock:     NewKeyLock(),
}

// LRPush 从右侧推入数据
// @author xuyang
// @datetime 2025-7-20 16:00
// @param key string 键
// @value value []string 值
func (gkvList *GkvList) LRPush(key string, value []string) {
	gkvList.keyLock.WLockRow(key)
	defer gkvList.keyLock.WUnLockRow(key)
	gkvList.date[key] = append(gkvList.data[key], value)
}

// LLPush 从左侧推入数据
// @author xuyang
// @datetime 2025-7-20 16:00
// @param key string 键
// @value value []string 值
func (gkvList *GkvList) LLPush(key string, value []string) {
	gkvList.keyLock.WLockRow(key)
	defer gkvList.keyLock.WUnLockRow(key)
	gkvList.date[key] = append(value, gkvList.data[key])
}

// LRPop 从右侧弹出数据
// @author xuyang
// @datetime 2025-7-20 16:00
// @param key string
// @return valueElement string
func (gkvList *GkvList) LRPop(key string) (string, bool) {
	gkvList.keyLock.WLockRow(key)
	defer gkvList.keyLock.WUnLockRow(key)
	n := len(gkvList.data[key])
	if n == 0{
		return "", false
	}
	element := gkvList.data[key][n-1]
	gkvList.data[key] := gkvList.data[key][:n]
	return element, true
}

// LLPop 从左侧弹出数据
// @author xuyang
// @datetime 2025-7-20 19:00
// @param key string
// @return valueElement string
func (gkvList *GkvList) LLPop(key string) (string, bool) {
	gkvList.keyLock.WLockRow(key)
	defer gkvList.keyLock.WUnLockRow(key)
	if len(gkvList.data[key]) == 0 {
		return "", false
	}
	element := gkvList.data[key][0]
	gkvList.data[key] = gkvList.data[key][1:]
	return element, true
}

// LRTop 查看最右侧数据
// @author xuyang
// @datetime 2025-7-20 16:00
// @param key string
// @return valueElement string
func (gkvList *GkvList) LRTop(key string) (string, bool) {
	gkvList.keyLock.WLockRow(key)
	defer gkvList.keyLock.WUnLockRow(key)
	n := len(gkvList.data[key])
	if n == 0{
		return "", false
	}
	element := gkvList.data[key][n-1]
	return element, true
}

// LLTop 查看最左侧数据
// @author xuyang
// @datetime 2025-7-20 19:00
// @param key string
// @return valueElement string
func (gkvList *GkvList) LLTop(key string) (string, bool) {
	gkvList.keyLock.WLockRow(key)
	defer gkvList.keyLock.WUnLockRow(key)
	if len(gkvList.data[key]) == 0 {
		return "", false
	}
	element := gkvList.data[key][0]
	return element, true
}

// GetAllKeys 获取所有key
// @author xuyang
// @datetime 2025-6-24 6:00
func (gkvList *GkvList) GetAllKeys() []string {
	gkvList.keyLock.tableLock.Lock()
	defer gkvList.keyLock.tableLock.Unlock()
	keys := make([]string, 0, len(gkvList.data))
	for key := range gkvList.data {
		keys = append(keys, key)
	}
	return keys
}

// LSetTime 设置过期时间(毫秒为单位)
func (gkvList *GkvList) LSetTime(key string, timeMs int) bool {
	gkvList.keyLock.WLockRow(key)
	defer gkvList.keyLock.WUnLockRow(key)
	// 检查键是否存在
	if _, exists := gkvList.data[key]; exists {
		expireTime := time.Now().Add(time.Duration(timeMs) * time.Millisecond)
		gkvList.expireTimes[key] = expireTime
		return true
	} else {
		return false
	}
}

// LGetTTL 获取键的剩余生存时间(毫秒数)
// @author xuyang
// @datetime 2025-7-16
// @param key string 键
// @return int64 剩余生存时间
// @return 0 键已过期
// @return -1 键不存在
// @return -2 键没有设置过期时间
func (gkv *GkvList) GetTTL(key string) int64 {
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
