package data

// GkvMap 映射数据结构
// @author xuyang
// @datetime 2025-7-16 21:00
type GkvMap struct {
	// 全部数据 key - filed - value
	data        map[string]map[string]string
	// 全部数据的过期时间
	expireTimes map[string]time.Time
	// 锁实例
	keyLock     *KeyLock
}

// DataGkvMap 全局数据实例
// @author xuyang
// @datetime 2025-7-16 21:00
var DataGkvMap = &GkvMap{
	data:        make(map[string]map[string]string),
	expireTimes: make(map[string]time.Time),
	keyLock:     NewKeyLock(),
}

// MSet 设置数据
// @author xuyang
// @datetime 2025-7-16 21:00
// @param key
// @param field
// @param value
// @return bool 是否设置成功
func (gkvMap *GkvMap) MSet (key, field, value string) bool {

}

// MGet 获取数据
// @author xuyang
// @datetime 2025-7-16 21:00
// @param key
// @param field
// @return value string
// @return ok bool 是否获取成功
func (gkvMap *GkvMap) MGet (key, field string) (value string, ok bool) {
	
}