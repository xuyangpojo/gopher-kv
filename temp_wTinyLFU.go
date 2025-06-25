package main

import (
	"container/list"
	"sync"
	"time"
)

// W-TinyLFU 缓存淘汰算法实现
// 结合了TinyLFU的频率统计和Window LRU的窗口机制
// @author xuyang
// @datetime 2025-1-27

// TinyLFU 频率统计器
type TinyLFU struct {
	freq    map[string]uint8 // 频率计数器
	maxFreq uint8            // 最大频率
	mu      sync.RWMutex
}

// NewTinyLFU 创建新的TinyLFU实例
func NewTinyLFU() *TinyLFU {
	return &TinyLFU{
		freq:    make(map[string]uint8),
		maxFreq: 255, // 8位无符号整数最大值
	}
}

// Increment 增加键的频率计数
func (t *TinyLFU) Increment(key string) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.freq[key] < t.maxFreq {
		t.freq[key]++
	}
}

// Estimate 估算键的频率
func (t *TinyLFU) Estimate(key string) uint8 {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.freq[key]
}

// Reset 重置所有频率计数（用于老化）
func (t *TinyLFU) Reset() {
	t.mu.Lock()
	defer t.mu.Unlock()

	// 将所有频率减半，实现老化效果
	for key, freq := range t.freq {
		t.freq[key] = freq / 2
	}
}

// CacheEntry 缓存条目
type CacheEntry struct {
	Key        string
	Value      interface{}
	Frequency  uint8
	AccessTime time.Time
}

// WindowCache 窗口缓存（LRU实现）
type WindowCache struct {
	capacity int
	list     *list.List
	cache    map[string]*list.Element
	mu       sync.RWMutex
}

// NewWindowCache 创建新的窗口缓存
func NewWindowCache(capacity int) *WindowCache {
	return &WindowCache{
		capacity: capacity,
		list:     list.New(),
		cache:    make(map[string]*list.Element),
	}
}

// Get 从窗口缓存获取值
func (w *WindowCache) Get(key string) (interface{}, bool) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if element, exists := w.cache[key]; exists {
		w.list.MoveToFront(element)
		entry := element.Value.(*CacheEntry)
		entry.AccessTime = time.Now()
		return entry.Value, true
	}
	return nil, false
}

// Put 向窗口缓存添加值
func (w *WindowCache) Put(key string, value interface{}) *CacheEntry {
	w.mu.Lock()
	defer w.mu.Unlock()

	// 如果键已存在，更新值
	if element, exists := w.cache[key]; exists {
		w.list.MoveToFront(element)
		entry := element.Value.(*CacheEntry)
		entry.Value = value
		entry.AccessTime = time.Now()
		return entry
	}

	// 如果容量已满，移除最久未使用的项
	if w.list.Len() >= w.capacity {
		lastElement := w.list.Back()
		if lastElement != nil {
			w.list.Remove(lastElement)
			entry := lastElement.Value.(*CacheEntry)
			delete(w.cache, entry.Key)
			return entry // 返回被淘汰的条目
		}
	}

	// 添加新条目
	entry := &CacheEntry{
		Key:        key,
		Value:      value,
		AccessTime: time.Now(),
	}
	element := w.list.PushFront(entry)
	w.cache[key] = element
	return nil
}

// Remove 从窗口缓存移除项
func (w *WindowCache) Remove(key string) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if element, exists := w.cache[key]; exists {
		w.list.Remove(element)
		delete(w.cache, key)
	}
}

// Size 获取窗口缓存大小
func (w *WindowCache) Size() int {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.list.Len()
}

// MainCache 主缓存（LFU实现）
type MainCache struct {
	capacity int
	cache    map[string]*CacheEntry
	mu       sync.RWMutex
}

// NewMainCache 创建新的主缓存
func NewMainCache(capacity int) *MainCache {
	return &MainCache{
		capacity: capacity,
		cache:    make(map[string]*CacheEntry),
	}
}

// Get 从主缓存获取值
func (m *MainCache) Get(key string) (interface{}, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if entry, exists := m.cache[key]; exists {
		entry.AccessTime = time.Now()
		return entry.Value, true
	}
	return nil, false
}

// Put 向主缓存添加值
func (m *MainCache) Put(key string, value interface{}, frequency uint8) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 如果键已存在，更新值
	if entry, exists := m.cache[key]; exists {
		entry.Value = value
		entry.Frequency = frequency
		entry.AccessTime = time.Now()
		return true
	}

	// 如果容量已满，需要淘汰一个条目
	if len(m.cache) >= m.capacity {
		// 找到频率最低的条目
		var minFreq uint8 = 255
		var minKey string

		for k, entry := range m.cache {
			if entry.Frequency < minFreq {
				minFreq = entry.Frequency
				minKey = k
			}
		}

		// 如果新条目的频率更高，则淘汰最低频率的条目
		if frequency > minFreq {
			delete(m.cache, minKey)
		} else {
			return false // 无法插入，频率太低
		}
	}

	// 添加新条目
	m.cache[key] = &CacheEntry{
		Key:        key,
		Value:      value,
		Frequency:  frequency,
		AccessTime: time.Now(),
	}
	return true
}

// Remove 从主缓存移除项
func (m *MainCache) Remove(key string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.cache, key)
}

// Size 获取主缓存大小
func (m *MainCache) Size() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.cache)
}

// WTinyLFU W-TinyLFU缓存实现
type WTinyLFU struct {
	windowCache *WindowCache
	mainCache   *MainCache
	tinyLFU     *TinyLFU
	windowSize  int
	mainSize    int
	mu          sync.RWMutex
}

// NewWTinyLFU 创建新的W-TinyLFU缓存
func NewWTinyLFU(totalCapacity int) *WTinyLFU {
	// 窗口缓存占1%，主缓存占99%
	windowSize := totalCapacity / 100
	if windowSize < 1 {
		windowSize = 1
	}
	mainSize := totalCapacity - windowSize

	return &WTinyLFU{
		windowCache: NewWindowCache(windowSize),
		mainCache:   NewMainCache(mainSize),
		tinyLFU:     NewTinyLFU(),
		windowSize:  windowSize,
		mainSize:    mainSize,
	}
}

// Get 从缓存获取值
func (w *WTinyLFU) Get(key string) (interface{}, bool) {
	w.mu.Lock()
	defer w.mu.Unlock()

	// 增加频率计数
	w.tinyLFU.Increment(key)

	// 先从窗口缓存查找
	if value, found := w.windowCache.Get(key); found {
		return value, true
	}

	// 再从主缓存查找
	if value, found := w.mainCache.Get(key); found {
		return value, true
	}

	return nil, false
}

// Put 向缓存添加值
func (w *WTinyLFU) Put(key string, value interface{}) {
	w.mu.Lock()
	defer w.mu.Unlock()

	// 增加频率计数
	w.tinyLFU.Increment(key)

	// 先尝试放入窗口缓存
	if evicted := w.windowCache.Put(key, value); evicted != nil {
		// 窗口缓存已满，被淘汰的条目需要处理
		evicted.Frequency = w.tinyLFU.Estimate(evicted.Key)

		// 尝试将淘汰的条目放入主缓存
		if !w.mainCache.Put(evicted.Key, evicted.Value, evicted.Frequency) {
			// 主缓存也放不下，完全淘汰
			w.tinyLFU.mu.Lock()
			delete(w.tinyLFU.freq, evicted.Key)
			w.tinyLFU.mu.Unlock()
		}
	}
}

// Delete 从缓存删除值
func (w *WTinyLFU) Delete(key string) {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.windowCache.Remove(key)
	w.mainCache.Remove(key)

	// 清除频率统计
	w.tinyLFU.mu.Lock()
	delete(w.tinyLFU.freq, key)
	w.tinyLFU.mu.Unlock()
}

// Size 获取缓存总大小
func (w *WTinyLFU) Size() int {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.windowCache.Size() + w.mainCache.Size()
}

// WindowSize 获取窗口缓存大小
func (w *WTinyLFU) WindowSize() int {
	return w.windowCache.Size()
}

// MainSize 获取主缓存大小
func (w *WTinyLFU) MainSize() int {
	return w.mainCache.Size()
}

// ResetFrequencies 重置频率统计（用于老化）
func (w *WTinyLFU) ResetFrequencies() {
	w.tinyLFU.Reset()
}

// GetStats 获取缓存统计信息
func (w *WTinyLFU) GetStats() map[string]interface{} {
	w.mu.RLock()
	defer w.mu.RUnlock()

	return map[string]interface{}{
		"total_size":      w.Size(),
		"window_size":     w.windowCache.Size(),
		"main_size":       w.mainCache.Size(),
		"window_capacity": w.windowSize,
		"main_capacity":   w.mainSize,
	}
}
