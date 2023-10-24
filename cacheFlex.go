package distributed_cache

import (
	"sync"
)

/*
	interface-based programming
*/

// A Getter loads data for a key.
type Getter interface {
	Get(key string) ([]byte, error)
}

// A GetterFunc implements Getter with a function.
type GetterFunc func(key string) ([]byte, error)

// Get implements Getter interface function
func (f GetterFunc) Get(key string) ([]byte, error) {
	return f(key)
}

// A Group is a cache namespace and associated data loaded spread over
type Group struct {
	name      string
	getter    Getter
	mainCache cache
}

// when use 'var', the variable's scope extends to the entire package
var (
	mu     sync.RWMutex // read-write locks, to synchronize access to shared data structure
	groups = make(map[string]*Group)
)

// NewGroup create a new instance of Group and return its pointer
func NewGroup(name string, cacheBytes int64, getter Getter) *Group {
	if getter == nil {
		panic("nil Getter")
	}
	mu.Lock()
	defer mu.Unlock()

	g := &Group{
		name:      name,
		getter:    getter,
		mainCache: cache{cacheBytes: cacheBytes},
	}

	groups[name] = g
	return g
}

// GetGroup returns the named group previously created with NewGroup, or
// nil if there's no such group.
/*
	RLock() and RUnlock() provide shared access, allowing multiple goroutines to
	read from the shared data concurrently while preventing writes during the read phase
*/
func GetGroup(name string) *Group {
	mu.RLock()
	g := groups[name]
	mu.RUnlock()
	return g
}
