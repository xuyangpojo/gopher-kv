package data

// // KeySafeDataRW 支持键级读写锁的安全数据结构
// type KeySafeDataRW struct {
// 	data    map[string]interface{}
// 	keyLock *KeyRWMutex
// }

// // GlobalKeySafeDataRW 全局键级安全数据(读写锁版本)
// var GlobalKeySafeDataRW = &KeySafeDataRW{
// 	data:    make(map[string]interface{}),
// 	keyLock: NewKeyRWMutex(),
// }

// // Set 安全写入数据（键级写锁）
// func (ksd *KeySafeDataRW) Set(key string, value interface{}) {
// 	ksd.keyLock.Lock(key)
// 	defer ksd.keyLock.Unlock(key)
// 	ksd.data[key] = value
// }

// // Get 安全读取数据（键级读锁）
// func (ksd *KeySafeDataRW) Get(key string) (interface{}, bool) {
// 	ksd.keyLock.RLock(key)
// 	defer ksd.keyLock.RUnlock(key)
// 	val, exists := ksd.data[key]
// 	return val, exists
// }

// // Delete 安全删除数据（键级写锁）
// func (ksd *KeySafeDataRW) Delete(key string) {
// 	ksd.keyLock.Lock(key)
// 	defer ksd.keyLock.Unlock(key)
// 	delete(ksd.data, key)
// }

// // Update 安全更新数据（键级写锁）
// func (ksd *KeySafeDataRW) Update(key string, updateFunc func(interface{}) interface{}) {
// 	ksd.keyLock.Lock(key)
// 	defer ksd.keyLock.Unlock(key)

// 	current, exists := ksd.data[key]
// 	if exists {
// 		ksd.data[key] = updateFunc(current)
// 	}
// }

// // GetOrCreate 获取或创建数据（避免写入放大）
// func (ksd *KeySafeDataRW) GetOrCreate(key string, createFunc func() interface{}) interface{} {
// 	// 首先尝试读锁获取
// 	if val, exists := ksd.Get(key); exists {
// 		return val
// 	}

// 	// 获取写锁
// 	ksd.keyLock.Lock(key)
// 	defer ksd.keyLock.Unlock(key)

// 	// 双重检查避免重复创建
// 	if val, exists := ksd.data[key]; exists {
// 		return val
// 	}

// 	// 创建新值
// 	newVal := createFunc()
// 	ksd.data[key] = newVal
// 	return newVal
// }

// // Snapshot 获取整个数据集的快照（无锁，可能不一致）
// func (ksd *KeySafeDataRW) Snapshot() map[string]interface{} {
// 	snapshot := make(map[string]interface{})

// 	// 获取所有键的列表（需要全局锁保护）
// 	keys := ksd.getAllKeys()

// 	for _, key := range keys {
// 		if val, exists := ksd.Get(key); exists {
// 			snapshot[key] = val
// 		}
// 	}
// 	return snapshot
// }

// // getAllKeys 安全获取所有键列表（使用全局锁）
// func (ksd *KeySafeDataRW) getAllKeys() []string {
// 	ksd.keyLock.globalLock.Lock()
// 	defer ksd.keyLock.globalLock.Unlock()

// 	keys := make([]string, 0, len(ksd.data))
// 	for k := range ksd.data {
// 		keys = append(keys, k)
// 	}
// 	return keys
// }

// // hashKey 计算键的哈希值
// func hashKey(key string) uint32 {
// 	h := fnv.New32a()
// 	h.Write([]byte(key))
// 	return h.Sum32()
// }

// func main() {
// 	// 初始化测试数据
// 	GlobalKeySafeDataRW.Set("config:timeout", 30)
// 	GlobalKeySafeDataRW.Set("config:retries", 3)

// 	var wg sync.WaitGroup

// 	// 模拟并发读取（可以并行）
// 	for i := 0; i < 5; i++ {
// 		wg.Add(1)
// 		go func(id int) {
// 			defer wg.Done()
// 			start := time.Now()
// 			if val, exists := GlobalKeySafeDataRW.Get("config:timeout"); exists {
// 				fmt.Printf("Reader %d got timeout: %d (took %v)\n",
// 					id, val, time.Since(start))
// 			}
// 		}(i)
// 	}

// 	// 模拟并发写入（互斥）
// 	for i := 0; i < 3; i++ {
// 		wg.Add(1)
// 		go func(id int) {
// 			defer wg.Done()
// 			key := fmt.Sprintf("counter:%d", id%2) // 两个不同的键
// 			GlobalKeySafeDataRW.Update(key, func(val interface{}) interface{} {
// 				if num, ok := val.(int); ok {
// 					return num + 1
// 				}
// 				return 1
// 			})
// 			fmt.Printf("Writer %d updated %s\n", id, key)
// 		}(i)
// 	}

// 	// 模拟GetOrCreate操作
// 	for i := 0; i < 4; i++ {
// 		wg.Add(1)
// 		go func(id int) {
// 			defer wg.Done()
// 			key := fmt.Sprintf("object:%d", id%3)
// 			val := GlobalKeySafeDataRW.GetOrCreate(key, func() interface{} {
// 				fmt.Printf("Creating new object for %s\n", key)
// 				return fmt.Sprintf("object#%d", time.Now().UnixNano())
// 			})
// 			fmt.Printf("Processor %d got %s: %v\n", id, key, val)
// 		}(i)
// 	}

// 	wg.Wait()

// 	// 显示最终结果
// 	fmt.Println("\nFinal counters:")
// 	for i := 0; i < 2; i++ {
// 		key := fmt.Sprintf("counter:%d", i)
// 		if val, exists := GlobalKeySafeDataRW.Get(key); exists {
// 			fmt.Printf("%s: %d\n", key, val)
// 		}
// 	}

// 	fmt.Println("\nAll objects:")
// 	for i := 0; i < 3; i++ {
// 		key := fmt.Sprintf("object:%d", i)
// 		if val, exists := GlobalKeySafeDataRW.Get(key); exists {
// 			fmt.Printf("%s: %v\n", key, val)
// 		}
// 	}
// }
