# Distributed-Cache

## Key Features
- Designed a distributed cache system in Golang that leverages single-node concurrent caching and HTTP-based multi-node caching
- Utilized the Least Recently Used (LRU) cache eviction strategy to optimize cache management
- Employed Golang's native locking mechanisms to prevent cache penetration and ensure data consistency
- Implemented a consistent hashing mechanism for node selection, achieving load balancing in the distributed environment
- Optimized inter-node binary communication using Protocol Buffers (protobuf)

## Directory
```bash
Distributed-Cache/
    |--lru/
        |--lru.go  // 
    |--byteview.go // 
    |--cache.go    // concurrent access control
    |--geecache.go // 
    |--http.go
```
## Test
```go
go test
```


## ToDo