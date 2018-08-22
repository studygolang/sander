package nosql

import (
	"bytes"
	"sync"
	"time"

	"github.com/polaris1119/goutils"
	"github.com/polaris1119/logger"

	"github.com/golang/groupcache/lru"
)

const defaultMaxEntryNum = 100

const CacheKey = "cache_key"

type Compressor interface {
	Compress([]byte) error
	UnCompress() ([]byte, error)
}

// 缓存的数据格式
type CacheData struct {
	StoreTime     time.Time
	compressValue []byte
}

func NewCacheData() *CacheData {
	return &CacheData{
		StoreTime: time.Now(),
	}
}

func (this *CacheData) Compress(value []byte) (err error) {
	buf := new(bytes.Buffer)
	_, err = goutils.Gzip(value, buf)
	if err != nil {
		logger.Errorln("Compress -> gzip error:", err)
		return
	}

	this.compressValue = buf.Bytes()

	return
}

func (this *CacheData) UnCompress() ([]byte, error) {
	return goutils.Gunzip(bytes.NewBuffer(this.compressValue))
}

var DefaultLRUCache = NewLRUCache(defaultMaxEntryNum)

type LRUCache struct {
	Cache  *lru.Cache // 可以通过它直接调用非锁定方法
	locker *sync.RWMutex
}

func NewLRUCache(maxEntries int) *LRUCache {
	return &LRUCache{Cache: lru.New(maxEntries), locker: new(sync.RWMutex)}
}

func (this *LRUCache) Add(key, value interface{}) {
	this.locker.Lock()
	defer this.locker.Unlock()

	this.Cache.Add(lru.Key(key), value)
}

func (this *LRUCache) Get(key interface{}) (value interface{}, ok bool) {
	this.locker.RLock()
	defer this.locker.RUnlock()

	return this.Cache.Get(lru.Key(key))
}

// CompressAndAdd 对数据进行 gzip 压缩之后再加入缓存中
func (this *LRUCache) CompressAndAdd(key interface{}, value []byte, compressor Compressor) {
	this.locker.Lock()
	defer this.locker.Unlock()

	defer func() {
		// 避免 cache 中对象被 gc
		if err := recover(); err != nil {
			logger.Errorln("lru RemoveOldest panic:", err)
		}
	}()

	if err := compressor.Compress(value); err != nil {
		return
	}

	this.Cache.Add(lru.Key(key), compressor)
}

// GetAndUnCompress 获取数据并解压缩（Gunzip）
func (this *LRUCache) GetAndUnCompress(key interface{}) (value []byte, compressor Compressor, ok bool) {
	this.locker.RLock()
	defer this.locker.RUnlock()

	valInter, ok := this.Cache.Get(lru.Key(key))
	if !ok {
		return
	}

	if compressor, ok = valInter.(Compressor); ok {
		var err error
		value, err = compressor.UnCompress()
		if err != nil {
			ok = false
			return
		}
	}

	return
}

func (this *LRUCache) Len() int {
	this.locker.RLock()
	defer this.locker.RUnlock()

	return this.Cache.Len()
}

func (this *LRUCache) Remove(key interface{}) {
	this.locker.Lock()
	defer this.locker.Unlock()

	this.Cache.Remove(lru.Key(key))
}

func (this *LRUCache) RemoveOldest() {
	this.locker.Lock()
	defer this.locker.Unlock()

	this.Cache.RemoveOldest()
}
