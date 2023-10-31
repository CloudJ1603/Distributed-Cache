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
    |--cacheFlex/
        |--lru/
            |--lru.go  
            |--lru_test.go
        |--byteview.go 
        |--cache.go    
        |--geecache.go 
        |--http.go
    |--go.mod
    |--go.sum
    |--main.go
    |--README.md


```
## Test
```go
go test
```


## ToDo