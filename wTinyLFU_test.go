package main

import (
	"fmt"
	"testing"
)

// TestWTinyLFUBasic 基本功能测试
func TestWTinyLFUBasic(t *testing.T) {
	cache := NewWTinyLFU(100)

	// 测试基本设置和获取
	cache.Put("key1", "value1")
	if value, found := cache.Get("key1"); !found || value != "value1" {
		t.Errorf("期望找到 'value1'，实际得到 %v, found=%v", value, found)
	}

	// 测试不存在的键
	if value, found := cache.Get("nonexistent"); found {
		t.Errorf("期望未找到，实际得到 %v", value)
	}

	// 测试删除
	cache.Delete("key1")
	if value, found := cache.Get("key1"); found {
		t.Errorf("期望删除后未找到，实际得到 %v", value)
	}
}

// TestWTinyLFUEviction 淘汰策略测试
func TestWTinyLFUEviction(t *testing.T) {
	cache := NewWTinyLFU(10) // 小容量缓存

	// 填充缓存
	for i := 0; i < 15; i++ {
		key := fmt.Sprintf("key%d", i)
		cache.Put(key, fmt.Sprintf("value%d", i))
	}

	// 验证缓存大小不超过容量
	if cache.Size() > 10 {
		t.Errorf("缓存大小 %d 超过容量 10", cache.Size())
	}

	// 验证窗口缓存和主缓存都有数据
	stats := cache.GetStats()
	if stats["window_size"].(int) == 0 && stats["main_size"].(int) == 0 {
		t.Error("缓存应该包含数据")
	}
}

// TestWTinyLFUFrequency 频率统计测试
func TestWTinyLFUFrequency(t *testing.T) {
	cache := NewWTinyLFU(50)

	// 多次访问同一个键
	key := "frequent_key"
	cache.Put(key, "value")

	// 多次访问以增加频率
	for i := 0; i < 10; i++ {
		cache.Get(key)
	}

	// 添加一些其他键
	for i := 0; i < 5; i++ {
		cache.Put(fmt.Sprintf("other_key%d", i), fmt.Sprintf("value%d", i))
	}

	// 验证频繁访问的键仍然在缓存中
	if value, found := cache.Get(key); !found || value != "value" {
		t.Errorf("频繁访问的键应该仍在缓存中")
	}
}

// TestWTinyLFUWindowCache 窗口缓存测试
func TestWTinyLFUWindowCache(t *testing.T) {
	cache := NewWTinyLFU(100)

	// 添加一些数据到窗口缓存
	for i := 0; i < 5; i++ {
		key := fmt.Sprintf("window_key%d", i)
		cache.Put(key, fmt.Sprintf("window_value%d", i))
	}

	// 验证窗口缓存大小
	windowSize := cache.WindowSize()
	if windowSize == 0 {
		t.Error("窗口缓存应该包含数据")
	}

	fmt.Printf("窗口缓存大小: %d\n", windowSize)
}

// TestWTinyLFUMainCache 主缓存测试
func TestWTinyLFUMainCache(t *testing.T) {
	cache := NewWTinyLFU(100)

	// 添加足够多的数据以触发窗口缓存到主缓存的迁移
	for i := 0; i < 20; i++ {
		key := fmt.Sprintf("main_key%d", i)
		cache.Put(key, fmt.Sprintf("main_value%d", i))
	}

	// 验证主缓存大小
	mainSize := cache.MainSize()
	if mainSize == 0 {
		t.Error("主缓存应该包含数据")
	}

	fmt.Printf("主缓存大小: %d\n", mainSize)
}

// TestWTinyLFUStats 统计信息测试
func TestWTinyLFUStats(t *testing.T) {
	cache := NewWTinyLFU(100)

	// 添加一些数据
	for i := 0; i < 10; i++ {
		key := fmt.Sprintf("stats_key%d", i)
		cache.Put(key, fmt.Sprintf("stats_value%d", i))
	}

	stats := cache.GetStats()

	// 验证统计信息
	if stats["total_size"].(int) != cache.Size() {
		t.Errorf("总大小不匹配: 期望 %d, 实际 %d", cache.Size(), stats["total_size"].(int))
	}

	if stats["window_size"].(int) != cache.WindowSize() {
		t.Errorf("窗口大小不匹配: 期望 %d, 实际 %d", cache.WindowSize(), stats["window_size"].(int))
	}

	if stats["main_size"].(int) != cache.MainSize() {
		t.Errorf("主缓存大小不匹配: 期望 %d, 实际 %d", cache.MainSize(), stats["main_size"].(int))
	}

	fmt.Printf("统计信息: %+v\n", stats)
}

// TestWTinyLFUResetFrequencies 频率重置测试
func TestWTinyLFUResetFrequencies(t *testing.T) {
	cache := NewWTinyLFU(50)

	// 添加数据并多次访问
	key := "reset_key"
	cache.Put(key, "value")

	for i := 0; i < 5; i++ {
		cache.Get(key)
	}

	// 重置频率
	cache.ResetFrequencies()

	// 验证键仍然在缓存中（频率重置不应该影响缓存内容）
	if value, found := cache.Get(key); !found || value != "value" {
		t.Errorf("频率重置后键应该仍在缓存中")
	}
}

// BenchmarkWTinyLFU 性能基准测试
func BenchmarkWTinyLFU(b *testing.B) {
	cache := NewWTinyLFU(1000)

	b.ResetTimer()

	// 测试Put性能
	b.Run("Put", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			key := fmt.Sprintf("bench_key%d", i)
			cache.Put(key, fmt.Sprintf("bench_value%d", i))
		}
	})

	// 测试Get性能
	b.Run("Get", func(b *testing.B) {
		// 先添加一些数据
		for i := 0; i < 100; i++ {
			key := fmt.Sprintf("bench_get_key%d", i)
			cache.Put(key, fmt.Sprintf("bench_get_value%d", i))
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			key := fmt.Sprintf("bench_get_key%d", i%100)
			cache.Get(key)
		}
	})
}

// ExampleWTinyLFU 使用示例
func ExampleWTinyLFU() {
	// 创建一个容量为100的W-TinyLFU缓存
	cache := NewWTinyLFU(100)

	// 添加数据
	cache.Put("name", "张三")
	cache.Put("age", 25)
	cache.Put("city", "北京")

	// 获取数据
	if value, found := cache.Get("name"); found {
		fmt.Printf("姓名: %s\n", value)
	}

	// 获取统计信息
	stats := cache.GetStats()
	fmt.Printf("缓存统计: 总大小=%d, 窗口大小=%d, 主缓存大小=%d\n",
		stats["total_size"], stats["window_size"], stats["main_size"])

	// 输出:
	// 姓名: 张三
	// 缓存统计: 总大小=3, 窗口大小=3, 主缓存大小=0
}
