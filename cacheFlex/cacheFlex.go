package cacheFlex

import (
	"fmt"
	"log"
	"sync"
)

/*
	The following three guys are used together to work as a callback function,
	support retrieval of resource from different types of source
	when resource are not found in cache
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

/*
A Group is a cache namespace and associated data loaded spread over
*/
type Group struct {
	name      string // a unique name
	getter    Getter // a callback function used to retrieve data in cache miss
	mainCache cache  // a concurrent cache
	peers     PeerPicker
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
// RLock() and RUnlock() provide shared access, allowing multiple goroutines to
// read from the shared data concurrently while preventing writes during the read phase

func GetGroup(name string) *Group {
	mu.RLock()
	g := groups[name]
	mu.RUnlock()
	return g
}

// Get value for a key from cache
func (g *Group) Get(key string) (ByteView, error) {
	// empty key
	if key == "" {
		return ByteView{}, fmt.Errorf("key is required")
	}

	// key is found in the cache
	if v, ok := g.mainCache.get(key); ok {
		log.Println("[CacheFlex] hit")
		return v, nil
	}

	// key is not found in the cahce
	return g.load(key)
}

// func (g *Group) load(key string) (ByteView, error) {
// 	return g.getLocally(key);
// }

// retrieve data from local source, and add it to the cache
func (g *Group) getLocally(key string) (ByteView, error) {
	bytes, err := g.getter.Get(key)
	if err != nil {
		return ByteView{}, err
	}

	value := ByteView{bytes: cloneBytes(bytes)}
	g.populateCache(key, value)
	return value, nil
}

func (g *Group) populateCache(key string, value ByteView) {
	g.mainCache.add(key, value)
}

/*
	------- day5 : Distributed Nodes --------------------
*/
// registers a PeerPicker for choosing remote peer
func (g *Group) RegisterPeers(peers PeerPicker) {
	if g.peers != nil {
		panic("RegisterPeerPicker called more than once")
	}
	g.peers = peers
}

func (g *Group) load(key string) (value ByteView, err error) {
	if g.peers != nil {
		if peer, ok := g.peers.PickPeer(key); ok {
			if value, err = g.getFromPeer(peer, key); err == nil {
				return value, nil
			}
			log.Println("[GeeCache] Failed to get from peer", err)
		}
	}

	return g.getLocally(key)
}

func (g *Group) getFromPeer(peer PeerGetter, key string) (ByteView, error) {
	bytes, err := peer.Get(g.name, key)
	if err != nil {
		return ByteView{}, err
	}
	return ByteView{bytes: bytes}, nil
}
