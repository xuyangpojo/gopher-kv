package main

import (
	"encoding/gob"
	"os"
	"gopherkv/data"
	"time"
)

// SaveGkvStringToFile 将DataGkvString的数据持久化到文件
func SaveGkvStringToFile(filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := gob.NewEncoder(file)
	data.DataGkvString.keyLock.tableLock.Lock()
	defer data.DataGkvString.keyLock.tableLock.Unlock()
	tmp := struct {
		Data        map[string][]byte
		ExpireTimes map[string]time.Time
	}{
		Data:        data.DataGkvString.DataCopy(),
		ExpireTimes: data.DataGkvString.ExpireTimesCopy(),
	}
	return encoder.Encode(tmp)
}

// LoadGkvStringFromFile 从文件加载数据到DataGkvString
func LoadGkvStringFromFile(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	decoder := gob.NewDecoder(file)
	tmp := struct {
		Data        map[string][]byte
		ExpireTimes map[string]time.Time
	}{}
	if err := decoder.Decode(&tmp); err != nil {
		return err
	}
	data.DataGkvString.keyLock.tableLock.Lock()
	defer data.DataGkvString.keyLock.tableLock.Unlock()
	data.DataGkvString.SetAll(tmp.Data, tmp.ExpireTimes)
	return nil
}

// DataCopy 返回data的深拷贝
func (g *data.GkvString) DataCopy() map[string][]byte {
	result := make(map[string][]byte, len(g.data))
	for k, v := range g.data {
		copyV := make([]byte, len(v))
		copy(copyV, v)
		result[k] = copyV
	}
	return result
}

// ExpireTimesCopy 返回expireTimes的深拷贝
func (g *data.GkvString) ExpireTimesCopy() map[string]time.Time {
	result := make(map[string]time.Time, len(g.expireTimes))
	for k, v := range g.expireTimes {
		result[k] = v
	}
	return result
}

// SetAll 用于批量设置data和expireTimes
func (g *data.GkvString) SetAll(dataMap map[string][]byte, expireMap map[string]time.Time) {
	g.data = dataMap
	g.expireTimes = expireMap
}

