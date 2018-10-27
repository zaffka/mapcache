# Simple in-memory cache example using Golang
Cache built using singleton design pattern and a map as a storage.

## Subtle point
```go
func (s *singleton) Get(key string) (interface{}, bool) {
	s.mu.RLock()
	item, ok := s.cache[key]
	s.mu.RUnlock()

	if ok {
		item.cancel()              <--------- look here
		return item.data, ok
	}

	return nil, ok
}
```
Context will be cancelled with a little delay, so it's possible to read a cache item for quite a while.

To avoid this we can directly call delete here with mutex lock.  
If we do, we need to add switch selection like this at `lockAndDelete` func.
```go
func lockAndDelete(ctx context.Context, key string) {
	<-ctx.Done()
	switch ctx.Err() {
	case context.DeadlineExceeded:
		i := GetInstance()
		i.Delete(key)
	default:
	}
}
```