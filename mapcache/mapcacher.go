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

//holds time to live for cache item
const ttlInSecond = 30

var (
	instance *singleton //single instance of the cache
	once     sync.Once
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

//Set method sets cache item
func (s *singleton) Set(key string, value interface{}) {
	//cache write lock
	s.mu.Lock()

	//produce new context froma the background one with timeout and cancel func
	ctxWithTimeout, cFunc := context.WithTimeout(s.ctxBck, getTimeDurationFunc(ttlInSecond))

	//set cache item with context.CancelFunc attached
	s.cache[key] = &item{
		cancel: cFunc,
		data:   value,
	}

	//start locked func awaited for TTL to be reached
	go lockAndDelete(ctxWithTimeout, key)

	//cache write ulock
	s.mu.Unlock() //explicit call instead of defer call because of an execution overhead
}

//Get method reads item from cache and then remove it
func (s *singleton) Get(key string) (item interface{}, itemExist bool) {
	s.mu.RLock()
	item, itemExist = s.cache[key]
	s.mu.RUnlock()

	if itemExist {
		s.Delete(key)
	}
	return
}

//Delete method removes item from cache
func (s *singleton) Delete(key string) {
	s.mu.Lock()
	delete(s.cache, key)
	s.mu.Unlock()
}

//lockAndDelete func to be spawned locked during Set() execution
func lockAndDelete(ctx context.Context, key string) {

	//lock reading from the ctx.Done() channel
	<-ctx.Done()

	switch ctx.Err() {
	case context.DeadlineExceeded: //context's timeout reached, cache item to be removed
		i := GetInstance()
		i.Delete(key)
	default: //context cancelled for other reason - exiting func
		return
	}

}
