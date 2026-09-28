# 第一节课：redis

### 什么是 Redis？

Redis（Remote Dictionary Server）是一个开源的内存数据结构存储系统，常用作数据库、缓存和消息中间件。它支持多种数据结构，如字符串、哈希、列表、集合、有序集合等，并提供了丰富的操作命令。

### Redis 主要特性

* **内存存储**：数据主要存储在内存中，读写性能极高（通常达到微秒级）。
* **持久化**：支持 RDB（快照）和 AOF（追加日志）两种持久化方式，可确保数据在重启后不丢失。
* **数据结构丰富**：除了基本的键值对，还提供哈希、列表、集合、有序集合、位图、地理位置等。
* **复制与高可用**：通过主从复制和 Redis Sentinel 实现高可用，通过 Redis Cluster 实现分布式存储。
* **发布/订阅**：支持消息的发布与订阅模式。
* **事务**：支持简单的事务（multi/exec）和 Lua 脚本。
* **过期策略**：可以为键设置生存时间（TTL），自动删除过期数据。

### Redis 部署模式

Redis 支持多种部署架构，以适应不同的可用性、扩展性需求。


1. **单机模式（Standalone）**\n最简单的部署方式，单个 Redis 实例处理所有请求。适用于开发、测试或数据量较小的场景。缺点是无高可用保障。
2. **主从复制模式（Master‑Slave Replication）**\n一个主节点（master）负责写操作，一个或多个从节点（slave）复制主节点的数据，提供读操作的扩展性和数据冗余。主节点故障时需要手动切换。
3. **哨兵模式（Sentinel）**\nRedis Sentinel 是官方提供的高可用解决方案。它会自动监控主从节点，并在主节点故障时通过投票机制将一个从节点提升为新的主节点，同时更新客户端的连接信息。
4. **集群模式（Cluster）**\nRedis Cluster 是分布式方案，将数据分片（sharding）存储在不同的节点上，每个节点负责一部分哈希槽（hash slot）。支持水平扩展，同时提供一定程度的高可用性（每个分片可配置主从复制）。

选择合适的部署模式取决于业务对性能、可用性、扩展性以及运维复杂度的要求。

### Redis 常用数据结构

* **String（字符串）**：最基本的数据类型，可以存储文本、整数或二进制数据。
* **Hash（哈希）**：适合存储对象，每个哈希可以包含多个字段和值。
* **List（列表）**：按照插入顺序排序的字符串列表，支持从两端推入/弹出。
* **Set（集合）**：无序且不重复的字符串集合，支持交并差等集合运算。
* **Sorted Set（有序集合）**：每个成员关联一个分数，根据分数排序，适合排行榜等场景。

### Redis 常用环境

* **缓存**：热点数据（如商品信息、用户信息）缓存，减轻数据库压力
* **分布式锁**：用 SETNX/RedLock 实现分布式环境下的互斥操作
* **计数器**：文章阅读量、点赞数（INCR/DECR 原子操作）
* **限流**：接口限流（结合 ZSet 或计数器）

### 在 Go 中使用 Redis

#### 安装 go-redis 库

Go 社区最常用的 Redis 客户端是 go-redis。使用以下命令安装：

```bash
go get github.com/go-redis/redis/v8
```

#### 连接到 Redis

以下代码展示了如何创建一个 Redis 客户端并连接到一个本地 Redis 服务器（默认端口 6379）。

```go
package main

import (
    "context"
    "fmt"
    "github.com/go-redis/redis/v8"
)

var ctx = context.Background()

func main() {
    rdb := redis.NewClient(&redis.Options{
        Addr:     "localhost:6379",
        Password: "", // 无密码
        DB:       0,  // 使用默认 DB
    })

    pong, err := rdb.Ping(ctx).Result()
    if err != nil {
        panic(err)
    }
    fmt.Println(pong) // 输出 PONG 表示连接成功
}
```

#### 连接到不同部署模式的 Redis

在实际生产环境中，Redis 有多种部署模式，每种模式对应不同的客户端创建方式。

##### 1. 单机模式（Standalone）

即上面示例中的标准连接方式，适用于单个 Redis 实例。

```go
rdb := redis.NewClient(&redis.Options{
    Addr:     "localhost:6379",
    Password: "",   // 密码，如果没有则留空
    DB:       0,    // 数据库编号
})
```

##### 2. 哨兵模式（Sentinel）

当使用 Redis Sentinel 实现高可用时，可以使用 `NewFailoverClient` 连接。

```go
rdb := redis.NewFailoverClient(&redis.FailoverOptions{
    MasterName:    "mymaster",
    SentinelAddrs: []string{"sentinel1:26379", "sentinel2:26379", "sentinel3:26379"},
    Password:      "", // 主节点密码
    DB:            0,
})
```

##### 3. 集群模式（Cluster）

对于 Redis Cluster 分布式部署，需要使用 `NewClusterClient`。

```go
rdb := redis.NewClusterClient(&redis.ClusterOptions{
    Addrs:    []string{"node1:6379", "node2:6379", "node3:6379"},
    Password: "", // 集群密码（如果所有节点密码相同）
    // 若需路由，客户端会自动处理
})
```

##### 4. 主从模式（主从复制）

在 Redis 主从复制架构中，通常只需连接主节点进行写操作，从节点用于读操作。go-redis 支持为只读操作自动路由到从节点，可使用 `NewFailoverClient`（配合 Sentinel）或手动管理多个客户端。

若使用简单的静态主从（无 Sentinel），可以创建两个客户端：

```go
master := redis.NewClient(&redis.Options{Addr: "master:6379"})
slave  := redis.NewClient(&redis.Options{Addr: "slave:6379"})
// 写操作使用 master，读操作可选用 slave
```

在实际生产环境，推荐使用 Sentinel 或 Cluster 来自动管理故障转移。

#### 基本数据操作

##### 字符串（String）操作

```go
// 设置键值
err := rdb.Set(ctx, "key", "value", 0).Err()
if err != nil {
    panic(err)
}

// 获取键值
val, err := rdb.Get(ctx, "key").Result()
if err != nil {
    panic(err)
}
fmt.Println("key", val)

// 删除键
rdb.Del(ctx, "key")
```

##### 哈希（Hash）操作

```go
// 设置哈希字段
rdb.HSet(ctx, "user:1000", "name", "Alice", "age", 30)

// 获取单个字段
name, err := rdb.HGet(ctx, "user:1000", "name").Result()
if err != nil {
    panic(err)
}
fmt.Println("name:", name)

// 获取所有字段
fields, err := rdb.HGetAll(ctx, "user:1000").Result()
if err != nil {
    panic(err)
}
for k, v := range fields {
    fmt.Printf("%s: %s\n", k, v)
}
```

##### 列表（List）操作

```go
// 从右侧推入
rdb.RPush(ctx, "mylist", "a", "b", "c")

// 获取列表长度
length, err := rdb.LLen(ctx, "mylist").Result()
if err != nil {
    panic(err)
}
fmt.Println("length:", length)

// 获取范围元素
items, err := rdb.LRange(ctx, "mylist", 0, -1).Result()
if err != nil {
    panic(err)
}
for _, item := range items {
    fmt.Println(item)
}
```

##### 集合（Set）操作

```go
// 添加成员
rdb.SAdd(ctx, "myset", "apple", "banana", "orange")

// 判断成员是否存在
exists, err := rdb.SIsMember(ctx, "myset", "apple").Result()
if err != nil {
    panic(err)
}
fmt.Println("exists:", exists)

// 获取所有成员
members, err := rdb.SMembers(ctx, "myset").Result()
if err != nil {
    panic(err)
}
for _, m := range members {
    fmt.Println(m)
}
```

##### 有序集合（Sorted Set）操作

```go
// 添加成员及分数
rdb.ZAdd(ctx, "leaderboard", &redis.Z{Score: 100, Member: "player1"},
    &redis.Z{Score: 200, Member: "player2"})

// 按分数范围获取成员
zs, err := rdb.ZRangeByScoreWithScores(ctx, "leaderboard", &redis.ZRangeBy{
    Min: "0",
    Max: "300",
}).Result()
if err != nil {
    panic(err)
}
for _, z := range zs {
    fmt.Println(z.Member, z.Score)
}
```

##### 管道（Pipeline）

管道用于将多个命令一次性发送，减少网络往返延迟。

```go
pipe := rdb.Pipeline()
pipe.Set(ctx, "key1", "value1", 0)
pipe.Set(ctx, "key2", "value2", 0)
pipe.Get(ctx, "key1")
cmds, err := pipe.Exec(ctx)
if err != nil {
    panic(err)
}
for _, cmd := range cmds {
    fmt.Println(cmd.String())
}
```

##### 事务（Transaction）

Redis 事务通过 MULTI / EXEC 实现，go-redis 提供了 `TxPipeline`。

```go
tx := rdb.TxPipeline()
tx.Set(ctx, "tx_key1", "tx_val1", 0)
tx.Set(ctx, "tx_key2", "tx_val2", 0)
_, err := tx.Exec(ctx)
if err != nil {
    panic(err)
}
```

##### 发布/订阅（Pub/Sub）

**发布消息**

```go
err := rdb.Publish(ctx, "mychannel", "hello world").Err()
if err != nil {
    panic(err)
}
```

**订阅消息**

```go
pubsub := rdb.Subscribe(ctx, "mychannel")
defer pubsub.Close()

// 接收消息
ch := pubsub.Channel()
for msg := range ch {
    fmt.Println(msg.Channel, msg.Payload)
}
```

##### 过期（Expiration）

可以为键设置生存时间（TTL）。

```go
// 设置键并指定 10 秒过期
rdb.SetEX(ctx, "temp_key", "temp_value", 10*time.Second)

// 查询剩余时间
ttl, err := rdb.TTL(ctx, "temp_key").Result()
if err != nil {
    panic(err)
}
fmt.Println("ttl:", ttl)
```

## Redis 缓存问题及解决方案

在基于 Redis 的缓存系统中，我们常常会遇到三类典型问题：缓存穿透、缓存击穿和缓存雪崩。

### 1. 缓存穿透（Cache Penetration）

**定义**\n缓存穿透是指查询一个数据库中根本不存在的数据。由于缓存中查不到，每次请求都会穿透到数据库，导致数据库压力剧增。恶意攻击者可能会利用此漏洞，用大量不存在的 key 发起请求，从而拖垮数据库。

**解决方案**


1. **布隆过滤器（Bloom Filter）**：将所有可能存在的数据哈希到一个足够大的 bitmap 中，查询时先经过布隆过滤器，如果过滤器判断不存在，则直接返回，避免查询数据库。
2. **缓存空值**：即使查询结果为空，也将空结果（例如 nil）进行缓存，并设置一个较短的过期时间（比如 5 分钟），这样后续相同的请求可以直接从缓存中获取空值，保护数据库。

**Go 代码示例**\n我们使用 go-redis 客户端库来实现缓存空值的方案。首先需要初始化 Redis 客户端：

```go
// demo.go
package main

import (
    "context"
    "fmt"
    "time"

    "github.com/go-redis/redis/v8"
)

var rdb *redis.Client

func initRedis() {
    rdb = redis.NewClient(&redis.Options{
        Addr:     "localhost:6379",
        Password: "", // no password set
        DB:       0,  // use default DB
    })
}

// GetDataWithCachePenetration 模拟缓存穿透防护
func GetDataWithCachePenetration(ctx context.Context, key string) (string, error) {
    // 1. 尝试从缓存获取
    val, err := rdb.Get(ctx, key).Result()
    if err == redis.Nil {
        // 2. 缓存不存在，查询数据库（模拟）
        dbResult, dbErr := fetchFromDB(key)
        if dbErr != nil {
            // 数据库查询失败（例如数据不存在），将空值缓存起来
            // 设置较短的过期时间，防止长期占用缓存
            rdb.SetEX(ctx, key, "", 5*time.Minute)
            return "", dbErr
        }
        // 3. 将查询到的数据写入缓存
        rdb.SetEX(ctx, key, dbResult, 30*time.Minute)
        return dbResult, nil
    } else if err != nil {
        return "", err
    }
    // 如果缓存中存储的是空字符串（代表数据库无此数据），这里可以根据业务逻辑返回空或错误
    if val == "" {
        return "", fmt.Errorf("data not found")
    }
    return val, nil
}

func fetchFromDB(key string) (string, error) {
    // 模拟数据库查询，这里假设只有 key=="exist" 时存在数据
    if key == "exist" {
        return "some data", nil
    }
    return "", fmt.Errorf("data not found in DB")
}
```

### 2. 缓存击穿（Cache Breakdown）

**定义**\n缓存击穿是指某个热点数据过期的瞬间，大量并发请求同时无法从缓存中获取数据，从而全部涌向数据库，导致数据库瞬时压力过大。

**解决方案**


1. **互斥锁（Mutex）**：当缓存失效时，只允许一个线程去查询数据库并重建缓存，其他线程等待该线程完成后再从缓存中获取数据。
2. **逻辑过期**：不对热点数据设置物理过期时间，而是将过期时间信息存储在 value 中。当发现数据逻辑过期时，另一个线程负责异步刷新缓存，当前线程返回旧数据。

**Go 代码示例**\n以下是使用互斥锁（sync.Mutex）实现的缓存击穿防护：

```go
// demo.go 续
import "sync"

var mutexMap = make(map[string]*sync.Mutex)
var mapMu sync.Mutex

func getMutex(key string) *sync.Mutex {
    mapMu.Lock()
    defer mapMu.Unlock()
    if _, ok := mutexMap[key]; !ok {
        mutexMap[key] = &sync.Mutex{}
    }
    return mutexMap[key]
}

// GetDataWithCacheBreakdown 使用互斥锁防止缓存击穿
func GetDataWithCacheBreakdown(ctx context.Context, key string) (string, error) {
    // 1. 尝试从缓存获取
    val, err := rdb.Get(ctx, key).Result()
    if err == nil {
        return val, nil
    }
    // 2. 缓存未命中，获取该 key 专用的互斥锁
    mu := getMutex(key)
    mu.Lock()
    defer mu.Unlock()

    // 3. 再次检查缓存（双检锁），因为在等待锁期间可能已经有其他线程写入了缓存
    val, err = rdb.Get(ctx, key).Result()
    if err == nil {
        return val, nil
    }
    // 4. 查询数据库并重建缓存
    dbResult, dbErr := fetchFromDB(key)
    if dbErr != nil {
        return "", dbErr
    }
    rdb.SetEX(ctx, key, dbResult, 30*time.Minute)
    return dbResult, nil
}
```

### 3. 缓存雪崩（Cache Avalanche）

**定义**\n缓存雪崩是指在同一时间大量缓存数据集中过期，导致所有请求都直接打到数据库，造成数据库瞬时压力过大甚至宕机。

**解决方案**


1. **随机过期时间**：为缓存数据的过期时间添加随机值（例如基础时间 ± 随机数），避免大量 key 同时过期。
2. **高可用架构**：通过 Redis 集群、主从复制、哨兵模式等提高缓存可用性。
3. **限流降级**：在应用层使用限流算法（如令牌桶、漏桶）控制访问数据库的并发量，或对非核心业务进行降级处理。

**Go 代码示例**\n以下是设置随机过期时间的实现：

```go
// demo.go 续
import "math/rand"

// SetWithRandomExpire 设置缓存，并添加随机过期时间以避免雪崩
func SetWithRandomExpire(ctx context.Context, key, value string) error {
    // 基础过期时间 30 分钟
    baseExpire := 30 * time.Minute
    // 随机增加或减少 0~5 分钟的偏移量
    offset := time.Duration(rand.Intn(10)-5) * time.Minute // -5 ~ +5 分钟
    expire := baseExpire + offset
    if expire < 0 {
        expire = baseExpire
    }
    return rdb.SetEX(ctx, key, value, expire).Err()
}

// GetDataWithCacheAvalanche 模拟缓存雪崩防护的读取流程
func GetDataWithCacheAvalanche(ctx context.Context, key string) (string, error) {
    val, err := rdb.Get(ctx, key).Result()
    if err == redis.Nil {
        // 缓存未命中，从数据库加载
        dbResult, dbErr := fetchFromDB(key)
        if dbErr != nil {
            return "", dbErr
        }
        // 使用随机过期时间写入缓存
        if err := SetWithRandomExpire(ctx, key, dbResult); err != nil {
            return "", err
        }
        return dbResult, nil
    } else if err != nil {
        return "", err
    }
    return val, nil
}
```