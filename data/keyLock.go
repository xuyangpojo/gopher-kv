package data

import (
	"hash/fnv"
	"sync"
)

// KeyLock 通用锁结构
// @author xuyang
// @datetime 2025-6-24 5:00
type KeyLock struct {
	// 表锁: 维持映射关系
	tableLock sync.Mutex
	// 行锁: 特定键的锁
	rowLocks map[uint32]*sync.RWMutex
}

// hashS 计算哈希值
// @param key string 待计算值
// @return uint32 哈希值
// @author xuyang
// @datetime 2025-6-24 5:00
func hashS(s string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(s))
	return h.Sum32()
}

// NewKeyLock 创建新的KeyLock实例
// @param nil
// @return *KeyLock 新的KeyLock实例
// @author xuyang
// @datetime 2025-6-24 5:00
func NewKeyLock() *KeyLock {
	return &KeyLock{
		rowLocks: make(map[uint32]*sync.RWMutex),
	}
}

// RLockRow(key string) 获取行读锁
// @param key string 行键
// @return nil
// @author xuyang
// @datetime 2025-6-24 5:00
func (keyLock *KeyLock) RLockRow(key string) {
	lockID := hashS(key)
	keyLock.tableLock.Lock()
	rowLock, ok := keyLock.rowLocks[lockID]
	// 不是已存在锁时, 创建新锁
	if !ok {
		rowLock = &sync.RWMutex{}
		keyLock.rowLocks[lockID] = rowLock
	}
	keyLock.tableLock.Unlock()
	rowLock.RLock()
}

// RUnLockRow(key string) 释放行读锁
// @param key string 行键
// @return nil
// @author xuyang
// @datetime 2025-6-24 5:00
func (keyLock *KeyLock) RUnLockRow(key string) {
	lockID := hashS(key)
	keyLock.tableLock.Lock()
	if rowLock, ok := keyLock.rowLocks[lockID]; ok {
		rowLock.RUnlock()
	}
	keyLock.tableLock.Unlock()
}

// WLockRow(key string) 获取行写锁
// @param key string 行键
// @return nil
// @author xuyang
// @datetime 2025-6-24 5:00
func (keyLock *KeyLock) WLockRow(key string) {
	lockID := hashS(key)
	keyLock.tableLock.Lock()
	rowLock, ok := keyLock.rowLocks[lockID]
	// 不是已存在锁时, 创建新锁
	if !ok {
		rowLock = &sync.RWMutex{}
		keyLock.rowLocks[lockID] = rowLock
	}
	keyLock.tableLock.Unlock()
	rowLock.RLock()
}

// WUnLockRow(key string) 释放行写锁
// @param key string 行键
// @return nil
// @author xuyang
// @datetime 2025-6-24 5:00
func (keyLock *KeyLock) WUnLockRow(key string) {
	lockID := hashS(key)
	keyLock.tableLock.Lock()
	if rowLock, ok := keyLock.rowLocks[lockID]; ok {
		rowLock.RUnlock()
	}
	keyLock.tableLock.Unlock()
}
