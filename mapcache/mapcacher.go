package mapcache

import (
	"context"
	"sync"
)

//holds cached value
type item struct {
	cancel context.CancelFunc
	data   interface{}
}

//map of the cached values
type cache map[string]*item

//struct to hold a single instance of the cache
type singleton struct {
	cache
	ctxBck context.Context
	mu     sync.RWMutex
}

var (
	ttlInSecond = 30       //holds time to live for cache item
	instance    *singleton //single instance of the cache
	once        sync.Once
)

//Mapcacher interface wrapper for singletone struct
type Mapcacher interface {
	Set(key string, value interface{})
	Get(key string) (item interface{}, itemExist bool)
	Delete(key string)
}

//GetInstance creates a single instance of a singletone struct
func GetInstance() Mapcacher {
	once.Do(func() {
		ctx := context.Background()
		instance = &singleton{
			ctxBck: ctx,
			cache:  make(map[string]*item),
		}
	})
	return instance
}

//Set method sets cache item (!possible to set an item with empty key)
func (s *singleton) Set(key string, value interface{}) {
	//produce new context from the background one with timeout and a cancel func
	ctxWithTimeout, cFunc := context.WithTimeout(s.ctxBck, getTimeDurationFunc(ttlInSecond))

	//cache write lock
	s.mu.Lock()

	//set cache item with context.CancelFunc attached
	s.cache[key] = &item{
		cancel: cFunc,
		data:   value,
	}

	//cache write ulock
	s.mu.Unlock() //explicit call instead of defer call because of a defer execution overhead

	//start channel locked func awaited for timeout or cancel call
	go lockAndDelete(ctxWithTimeout, key)
}

//Get method reads item from cache and then remove it
func (s *singleton) Get(key string) (interface{}, bool) {
	s.mu.RLock()
	item, ok := s.cache[key]
	s.mu.RUnlock()

	if ok {
		item.cancel()
		return item.data, ok
	}

	return nil, ok
}

//Delete method removes item from the cache
func (s *singleton) Delete(key string) {
	s.mu.Lock()
	delete(s.cache, key)
	s.mu.Unlock()
}

//lockAndDelete func to be spawned locked during Set() execution
func lockAndDelete(ctx context.Context, key string) {

	//lock while reading from the ctx.Done() channel
	<-ctx.Done()
	switch ctx.Err() {
	case context.DeadlineExceeded:
		i := GetInstance()
		i.Delete(key)
	default:
	}

}
