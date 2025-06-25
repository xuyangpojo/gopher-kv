package data

import (
	"testing"
	"time"
)

func TestSetAndGet(t *testing.T) {
	g := &GkvString{
		data:        make(map[string][]byte),
		expireTimes: make(map[string]time.Time),
		keyLock:     NewKeyLock(),
	}
	key := "foo"
	val := []byte("bar")
	g.Set(key, val)
	v, ok := g.Get(key)
	if !ok || string(v) != "bar" {
		t.Errorf("Get failed, want bar, got %v", v)
	}
}

func TestDelete(t *testing.T) {
	g := &GkvString{
		data:        make(map[string][]byte),
		expireTimes: make(map[string]time.Time),
		keyLock:     NewKeyLock(),
	}
	key := "foo"
	g.Set(key, []byte("bar"))
	g.Delete(key)
	_, ok := g.Get(key)
	if ok {
		t.Errorf("Delete failed, key still exists")
	}
}

func TestGetAllKeys(t *testing.T) {
	g := &GkvString{
		data:        make(map[string][]byte),
		expireTimes: make(map[string]time.Time),
		keyLock:     NewKeyLock(),
	}
	g.Set("a", []byte("1"))
	g.Set("b", []byte("2"))
	keys := g.GetAllKeys()
	if len(keys) != 2 {
		t.Errorf("GetAllKeys failed, want 2, got %d", len(keys))
	}
}

func TestGetAllKVs(t *testing.T) {
	g := &GkvString{
		data:        make(map[string][]byte),
		expireTimes: make(map[string]time.Time),
		keyLock:     NewKeyLock(),
	}
	g.Set("a", []byte("1"))
	g.Set("b", []byte("2"))
	kvs := g.GetAllKVs()
	if kvs["a"] != "1" || kvs["b"] != "2" {
		t.Errorf("GetAllKVs failed, got %v", kvs)
	}
}

func TestSetTimeAndExpire(t *testing.T) {
	g := &GkvString{
		data:        make(map[string][]byte),
		expireTimes: make(map[string]time.Time),
		keyLock:     NewKeyLock(),
	}
	key := "foo"
	g.Set(key, []byte("bar"))
	g.SetTime(key, 10) // 10ms
	time.Sleep(20 * time.Millisecond)
	_, ok := g.Get(key)
	if ok {
		t.Errorf("SetTime or expire failed, key should be expired")
	}
}

func TestSetNX(t *testing.T) {
	g := &GkvString{
		data:        make(map[string][]byte),
		expireTimes: make(map[string]time.Time),
		keyLock:     NewKeyLock(),
	}
	ok := g.SetNX("foo", []byte("bar"))
	if !ok {
		t.Errorf("SetNX failed on new key")
	}
	ok = g.SetNX("foo", []byte("baz"))
	if ok {
		t.Errorf("SetNX should fail on existing key")
	}
}

func TestSetXX(t *testing.T) {
	g := &GkvString{
		data:        make(map[string][]byte),
		expireTimes: make(map[string]time.Time),
		keyLock:     NewKeyLock(),
	}
	ok := g.SetXX("foo", []byte("bar"))
	if ok {
		t.Errorf("SetXX should fail on non-existing key")
	}
	g.Set("foo", []byte("bar"))
	ok = g.SetXX("foo", []byte("baz"))
	if !ok || string(g.data["foo"]) != "baz" {
		t.Errorf("SetXX failed on existing key")
	}
}

func TestGetTTL(t *testing.T) {
	g := &GkvString{
		data:        make(map[string][]byte),
		expireTimes: make(map[string]time.Time),
		keyLock:     NewKeyLock(),
	}
	if g.GetTTL("foo") != -1 {
		t.Errorf("GetTTL should return -1 for non-existing key")
	}
	g.Set("foo", []byte("bar"))
	if g.GetTTL("foo") != -1 {
		t.Errorf("GetTTL should return -1 for no expire key")
	}
	g.SetTime("foo", 10)
	if g.GetTTL("foo") <= 0 {
		t.Errorf("GetTTL should return positive for valid expire key")
	}
	time.Sleep(20 * time.Millisecond)
	if g.GetTTL("foo") != -2 {
		t.Errorf("GetTTL should return -2 for expired key")
	}
}
