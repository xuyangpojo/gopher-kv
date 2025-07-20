package data

import (
	"time"
)

// GkvBitMap 位图结构
// @author xuyang
// @datetime 2025-7-16 21:00
type GkvBitMap struct {
	data        map[string][]byte
	expireTimes map[string]time.Time
	keyLock     *KeyLock
}

// DataGkvBitMap 全局数据实例
var DataGkvBitMap = &GkvBitMap{
	data:        make(map[string][]byte),
	expireTimes: make(map[string]time.Time),
	keyLock:     NewKeyLock(),
}

// SetBit 设置某一位
// @param key string
// @param offset int
// @param value bool
func (bm *GkvBitMap) SetBit(key string, offset int, value bool) {
	bm.keyLock.WLockRow(key)
	defer bm.keyLock.WUnLockRow(key)
	byteIdx := offset / 8
	bitIdx := offset % 8
	if len(bm.data[key]) <= byteIdx {
		newBytes := make([]byte, byteIdx+1)
		copy(newBytes, bm.data[key])
		bm.data[key] = newBytes
	}
	if value {
		bm.data[key][byteIdx] |= 1 << bitIdx
	} else {
		bm.data[key][byteIdx] &^= 1 << bitIdx
	}
	delete(bm.expireTimes, key)
}

// GetBit 获取某一位
// @param key string
// @param offset int
// @return bool
func (bm *GkvBitMap) GetBit(key string, offset int) bool {
	bm.keyLock.RLockRow(key)
	defer bm.keyLock.RUnLockRow(key)
	byteIdx := offset / 8
	bitIdx := offset % 8
	if len(bm.data[key]) <= byteIdx {
		return false
	}
	return bm.data[key][byteIdx]&(1<<bitIdx) != 0
}

// Count 统计位图中为1的位数
// @param key string
// @return int
func (bm *GkvBitMap) Count(key string) int {
	bm.keyLock.RLockRow(key)
	defer bm.keyLock.RUnLockRow(key)
	data, exists := bm.data[key]
	if !exists {
		return 0
	}
	count := 0
	for _, b := range data {
		for i := 0; i < 8; i++ {
			if b&(1<<i) != 0 {
				count++
			}
		}
	}
	return count
}

// SetTime 设置过期时间(毫秒为单位)
func (bm *GkvBitMap) SetTime(key string, timeMs int) bool {
	bm.keyLock.WLockRow(key)
	defer bm.keyLock.WUnLockRow(key)
	if _, exists := bm.data[key]; exists {
		expireTime := time.Now().Add(time.Duration(timeMs) * time.Millisecond)
		bm.expireTimes[key] = expireTime
		return true
	}
	return false
}

// GetTTL 获取key的剩余生存时间(毫秒数)
func (bm *GkvBitMap) GetTTL(key string) int64 {
	bm.keyLock.RLockRow(key)
	defer bm.keyLock.RUnLockRow(key)
	if _, exists := bm.data[key]; !exists {
		return -1
	}
	expireTime, exists := bm.expireTimes[key]
	if !exists {
		return -2
	}
	remaining := time.Until(expireTime)
	if remaining <= 0 {
		return 0
	}
	return int64(remaining.Milliseconds())
}
