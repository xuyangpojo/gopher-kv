package data

import (
	"hash/fnv"
	"time"
)

// GkvHyperLoglog HyperLogLog结构
// @author xuyang
// @datetime 2025-7-16 21:00
type GkvHyperLoglog struct {
	data        map[string][]uint8 // key -> register array
	expireTimes map[string]time.Time
	keyLock     *KeyLock
	precision   uint8 // 桶数量为 2^precision
}

// DataGkvHyperLoglog 全局数据实例
var DataGkvHyperLoglog = &GkvHyperLoglog{
	data:        make(map[string][]uint8),
	expireTimes: make(map[string]time.Time),
	keyLock:     NewKeyLock(),
	precision:   14, // 16384 桶
}

// Add 添加元素
// @author xuyang
// @datetime 2025-7-20 23:00
// @param key string
// @param element string
func (hll *GkvHyperLoglog) Add(key, element string) {
	hll.keyLock.WLockRow(key)
	defer hll.keyLock.WUnLockRow(key)
	if _, exists := hll.data[key]; !exists {
		hll.data[key] = make([]uint8, 1<<hll.precision)
	}
	// 哈希值计算
	h := fnv.New64a()
	h.Write([]byte(element))
	hash := h.Sum64()
	idx := hash & ((1 << hll.precision) - 1)
	w := hash >> hll.precision
	zeros := uint8(1)
	for w != 0 && (w&1) == 0 {
		zeros++
		w >>= 1
	}
	if hll.data[key][idx] < zeros {
		hll.data[key][idx] = zeros
	}
	delete(hll.expireTimes, key)
}

// Count 估算基数
// @author xuyang
// @datetime 2025-7-20 23:00
// @param key string 键
// @return uint64 估算得到的基数(误差0.81%)
func (hll *GkvHyperLoglog) Count(key string) uint64 {
	hll.keyLock.RLockRow(key)
	defer hll.keyLock.RUnLockRow(key)
	registers, exists := hll.data[key]
	if !exists {
		return 0
	}
	m := float64(len(registers))
	sum := 0.0
	for _, v := range registers {
		sum += 1.0 / float64(uint64(1)<<v)
	}
	alpha := 0.7213 / (1 + 1.079/m)
	est := alpha * m * m / sum
	return uint64(est)
}

// Merge 合并多个 HyperLogLog
// @param dest string 目标HLL
// @param srcs ...string 要被合并的若干个HLL
func (hll *GkvHyperLoglog) Merge(dest string, srcs ...string) {
	hll.keyLock.WLockRow(dest)
	if _, exists := hll.data[dest]; !exists && len(srcs) > 0 {
		hll.data[dest] = make([]uint8, 1<<hll.precision)
	}
	for _, src := range srcs {
		hll.keyLock.RLockRow(src)
		srcReg, exists := hll.data[src]
		hll.keyLock.RUnLockRow(src)
		if !exists {
			continue
		}
		for i, v := range srcReg {
			if hll.data[dest][i] < v {
				hll.data[dest][i] = v
			}
		}
	}
	delete(hll.expireTimes, dest)
	hll.keyLock.WUnLockRow(dest)
}

// HSetTime 设置过期时间(毫秒为单位)
// @author xuyang
// @datetime 2025-7-20 23:00
// @param key string 键
// @param timeMs int 过期时间(毫秒数)
// @return bool 是否设置成功
func (hll *GkvHyperLoglog) HSetTime(key string, timeMs int) bool {
	hll.keyLock.WLockRow(key)
	defer hll.keyLock.WUnLockRow(key)
	if _, exists := hll.data[key]; exists {
		expireTime := time.Now().Add(time.Duration(timeMs) * time.Millisecond)
		hll.expireTimes[key] = expireTime
		return true
	}
	return false
}

// HGetTTL 获取键的剩余生存时间(毫秒数)
// @author xuyang
// @datetime 2025-7-16
// @param key string 键
// @return int64 剩余生存时间
// @return 0 键已过期
// @return -1 键不存在
// @return -2 键没有设置过期时间
func (gkv *GkvHyperLoglog) HGetTTL(key string) int64 {
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
