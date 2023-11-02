package lru

import "container/list"

// Cache is a LRU cache.
// NOTE: It is not safe for concurrent access.
type Cache struct {
	maxBytes  int64                    // max storage
	currBytes int64                    // current storage
	ll        *list.List               // a doubly linked list
	mmap      map[string]*list.Element // a map mapping key(string) to list element
	// optional and executed when an entry is purged.
	OnEvicted func(key string, value Value)
}

// entry is the data type for the node in doubly linked list,
// which is a <key,value> pair
type entry struct {
	key   string
	value Value
}

// Value use Len to count how many bytes it stores
type Value interface {
	Len() int
}

// Len the number of cache entries
/* NOTES for syntax
	When a type (whether it's a struct or any other type) has methods that 
	match the signature of all the methods of an interface, we say that the 
	type implements the interface. 
	We have a method 'Len( ) int' defined on the 'Cache' type. 
*/
func (c *Cache) Len() int {
	return c.ll.Len()
}

// Constructor of Cache
func New(maxBytes int64, onEvicted func(string, Value)) *Cache {
	return &Cache{
		maxBytes:  maxBytes,
		ll:        list.New(),
		mmap:      make(map[string]*list.Element),
		OnEvicted: onEvicted,
	}
}

// adds a <key, value> pair to the cache
/*	NOTES for syntax using in the Add function:
	'ele' is a variable of type '*list.Element', which is a node in doubly linked list
	'list.Element' has a field named 'Value' which is of type interface{}, which can hold any type,
	so it can be any type.
	'ele.Value.(*entry)' is a type assertion is Go. It tried to assert that the underlying type
	of 'ele.Value' is '*entry'.

	There is no explicit 'while' key word in Go, 'for' serves both purposes.
*/
func (c *Cache) Add(key string, value Value) {

	if ele, ok := c.mmap[key]; ok {        // the key is found map
		c.ll.MoveToFront(ele)
		kv := ele.Value.(*entry)                                   // the node we need to update
		c.currBytes += int64(value.Len()) - int64(kv.value.Len())  // update the overall size of the Cache 
		kv.value = value										   // assign new value to the node
	} else {                              // the key is not found in map
		ele := c.ll.PushFront(&entry{key, value})                  // create, and push the new node to the front
		c.mmap[key] = ele										   // map the key to the node in mmap
		c.currBytes += int64(len(key)) + int64(value.Len())		   // update the overall size of the Cache 
	}
	for c.maxBytes != 0 && c.maxBytes < c.currBytes {			   // check if the current size exceeds the limit
		c.RemoveOldest()										  
	}
}

// Get look ups a key's value
/* NOTEs for syntax:
	if Go, if a function has named return values in the signature and we don't explicity 
	specify the return values in a 'return' statement, the function will return the 
	current value of those named variables.

	We cannot directly access 'ele.Value.value', because 'ele.Value' is of type 'interface{}',
	we need to first assert its underlying type before we can access any if its field
*/
func (c *Cache) Get(key string) (value Value, ok bool) {
	if ele, ok := c.mmap[key]; ok {
		c.ll.MoveToFront(ele)
		kv := ele.Value.(*entry)
		return kv.value, true
	}
	return
}

// removes the oldest item
func (c *Cache) RemoveOldest() {
	ele := c.ll.Back()
	if ele != nil {
		c.ll.Remove(ele)           // remove the node in the linked list
		kv := ele.Value.(*entry)   // get the node
		delete(c.mmap, kv.key)     // delete the <key,value> pair in the mmap
		c.currBytes -= int64(len(kv.key)) + int64(kv.value.Len())    // update the overall size of the cache
		if c.OnEvicted != nil {    
			c.OnEvicted(kv.key, kv.value)         // execute the callback function if it is not nil
		}
	}
}


