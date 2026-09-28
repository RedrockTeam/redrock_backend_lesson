# redis

# redis

# redis

在了解redis相关内容之前，我们先来了解一下什么是关系型数据库，什么是非关系型数据库。

> NoSQL:NoSQL常见的解释有Non-Relational SQL或者Not Only SQL，不过Not Only SQL被更多人接受，一般泛指非关系型数据库。它和关系型数据库不同的是，不保证关系数据的ACID特性。 关系型数据库：简单来说，关系模型指的就是二维表格模型，而一个关系型数据库就是由二维表及其之间的联系所组成的一个数据组织。在数据库中有个比较重要的概念，**关系要比表结构更重要** 非关系型数据库：是一种数据结构化存储方法的集合，可以是文档或者键值对等。例如，以键值对存储，且结构不固定，每一个元组可以有不一样的字段，每个元组可以根据需要增加一些自己的键值对，这样就不会局限于固定的结构，可以减少一些时间和空间的开销。

## 什么是redis

Redis是现在最受欢迎的NoSQL数据库之一，Redis是一个使用ANSI C编写的开源、包含多种数据结构、支持网络、基于内存、可选持久性的键值对存储数据库，其具备如下特性：

* 基于内存运行，性能高效
* 支持分布式，理论上可以无限扩展
* key-value存储系统
* 开源的使用ANSI C语言编写、遵守BSD协议、支持网络、可基于内存亦可持久化的日志型、Key-Value数据库，并提供多种语言的API

## **常见的使用场景**

1\.缓存系统（热点数据：高频读取，低频写）

* 由于redis访问速度块、支持的数据类型比较丰富，所以redis很适合用来存储热点数据，另外结合expire，我们可以设置过期时间然后再进行缓存更新操作，这个功能最为常见 2.排行榜
* 关系型数据库在排行榜方面查询速度普遍偏慢，所以可以借助redis的SortedSet进行热点数据的排序。 3.MQ（消息队列）
* 由于redis有list push和list pop这样的命令，所以能够很方便的执行队列操作。 redis 的使用场景还有很多，这里就不一一列举了

## redis为什么快

* **数据存放在内存中**
* **采用了非阻塞I/O多路复用机制**
* **数据结构简单,对数据操作也简单**

## redis操作

### 字符串

#### set

语法：`set key value [ex seconds]/[px milliseconds]` \[nx/xx\]

* **EX_seconds**:表明过期时间，单位是秒
* **PX**:单位毫秒
* **NX** ： 只在键不存在时， 才对键进行设置操作
* **XX**： 与NX相反只在键已经存在时， 才对键进行设置操作 基本操作，注意当键值对已经存在时候，会替换掉其对应的value

```Plaintext
127.0.0.1:6379> set key value
OK
127.0.0.1:6379> get key
"value"
127.0.0.1:6379> set key 11
OK
127.0.0.1:6379> get key
"11"
```

使用EX、PX选项,TTL 命令显示过期时间s(PTTL显示ms)，这里无法同时使用EX,PX选项

```Plaintext
127.0.0.1:6379> set key value ex 1000
OK
127.0.0.1:6379> get key
"value"
127.0.0.1:6379> ttl key
(integer) 991
```

使用了XX与NX选项，NX就是当这个kv不存在的时候才会被设置 ,XX确实在这个kv存在的时候才会被设置

```Plaintext
127.0.0.1:6379> set key value
OK
127.0.0.1:6379> set key 111 nx
(nil)
127.0.0.1:6379> get key
"value"
127.0.0.1:6379> set key 111 xx
OK
127.0.0.1:6379> get key
"111"
```

#### MSET

语法：`mset k1 v1 k2 v2` 同时设置多个键值对，并且总是返回ok（有MSET就肯定会有MGET来得到多个值）

```Plaintext
127.0.0.1:6379> mset k1 v1 k2 v2 k3 v3
OK
127.0.0.1:6379> mget k1 k2 k3
1) "v1"
2) "v2"
3) "v3"
127.0.0.1:6379> set k4 v4
OK
127.0.0.1:6379> mget k1 k2 k3 k4
1) "v1"
2) "v2"
3) "v3"
4) "v4"
```

#### **SETNX**

SETNX（set if not exists）跟上述加上NX选项的SET功能一样

```Plaintext
127.0.0.1:6379> setnx key value
(integer) 1
127.0.0.1:6379> setnx key 11
(integer) 0
127.0.0.1:6379> get key
"value"
```

#### SETEX

语法: `setex key seconds value` (先跟时间再跟值)

```Plaintext
127.0.0.1:6379> setex key 10 val
OK
127.0.0.1:6379> get key
"val"
127.0.0.1:6379> ttl key
(integer) 2
```

#### PSETEX

跟SETEX差不过，只不过这个是以ms为单位了

#### GET

语法：`GET key` 返回值：若value存在，则返回value，若不存在则返回nil，若该键对应的值不为字符串类型，则返回错误

#### DEL

语法：`GET key` 返回值：返回删除的key数量

```Plaintext
127.0.0.1:6379> DEL k1
(integer) 1
```

#### GETSET

语法：GETSET key value 设置当前的key-value，并返回之前的值，如果之前的key并不存在，就返回nil 这里如果原来就不存在这个KV，则会返回nil

```Plaintext
127.0.0.1:6379> getset key value
(nil)
127.0.0.1:6379> get key
"value"
127.0.0.1:6379> getset key 111
"value"
```

#### STRLEN

语法：STRLEN key STRLEN 命令返回字符串值的长度,当键 key 不存在时， 命令返回 0 。 当 key 储存的不是字符串值时， 返回一个错误。

```Plaintext
127.0.0.1:6379> set key 111111
OK
127.0.0.1:6379> strlen key
(integer) 6
```

#### APPEND

语法：`APPEND key value` 追加字符串，若该key已存在，则往后追加，若不存在就跟简单的执行`SET`一样 返回追加后key的长度

```Plaintext
127.0.0.1:6379> set key value
OK
127.0.0.1:6379> append key 11
(integer) 7
127.0.0.1:6379> get key
"value11"
```

#### SETRANGE

语法：`SETRANGE key offset value` 从偏移量`offset`后覆盖原来的value，并返回当前key对应字符串的长度

```Plaintext
127.0.0.1:6379> set key value
OK
127.0.0.1:6379> setrange key 5 isgolang
(integer) 13
127.0.0.1:6379> get key
"valueisgolang"
```

#### GETRANGE

语法：`GETRANGE key start end` 获取\[start,end\]的子字符串，允许负偏移量，即-1代表最后一个，-2代表倒数第二个 数字规则和**SETRANGE**一样

```Plaintext
127.0.0.1:6379> getrange key 0 -1
"valueisgolang"
```

#### INCR

语法：INCR key 为键 `key` 储存的数字值加上一。 如果键 `key` 不存在， 那么它的值会先被初始化为 `0` ， 然后再执行 `INCR` 命令。 如果键 `key` 储存的值不能被**解释**为数字（因为redis中并没有专用的整数类型）， 那么 `INCR` 命令将返回一个错误。

```Plaintext
127.0.0.1:6379> incr key
(integer) 1
127.0.0.1:6379> get key
"1"
127.0.0.1:6379> incr key
(integer) 2
127.0.0.1:6379> get key
"2"
```

### 哈希

#### hset

语法：`HEST hash field value` 将hash表中field域设置为value，与之相配的为HGET(`HEGT hash field`) 返回值：当 `HSET` 命令在哈希表中新创建 `field` 域并成功为它设置值时， 命令返回 `1` ； 如果已经存在，那么命令返回 `0` 。

```Plaintext
127.0.0.1:6379> hset hash k1 v1
(integer) 1
127.0.0.1:6379> hget hash k1
"v1"
127.0.0.1:6379> hset hash k1 v1
(integer) 0
```

#### HLEN

语法：`HLEN hash` 返回哈希表hash中域的数量，若hash不存在，则返回0

```Plaintext
127.0.0.1:6379> hset hash k1 v1
(integer) 1
127.0.0.1:6379> hlen has
(integer) 0
127.0.0.1:6379> hlen hash
(integer) 1
```

#### HDEL

语法：`HDEL key field1 field2...` 删除哈希表key中的域，并返回成功删除的域的数量

```Plaintext
127.0.0.1:6379> hdel hash k1
(integer) 1
```

redis的其他哈希操作比如获取一个哈希表中所有的field，获取哈希表中所有的值等等就自己去试试吧

### 队列

#### LPUSH与LRANGE

将一个或者多个值插入**列表头部**，如果为空列表会被创建执行LPUSH，当key存在但是却不是列表，就返回错误 LRANGE这命令中0标识列表第一个元素，1则是第二个。 -1代表最后一个元素，-2则代表倒数第二个。

```Plaintext
127.0.0.1:6379> lpush list c java go
(integer) 3
127.0.0.1:6379> lrange list 0 -1
1) "go"
2) "java"
3) "c"
```

#### LPOP

语法：`LPOP key` 移除列表key表头元素，即出队列

```Plaintext
127.0.0.1:6379> lpop list
"go"
127.0.0.1:6379> lrange list 0 -1
1) "java"
2) "c"
```

#### LSET

语法：`LSET key index value` 将列表 key下标为 index 的元素的值设置为 value 若 `index` (这里的下标是从0开始的，也允许负数)参数超出范围，或对一个空列表( `key` 不存在)进行操作时，则会返回一个错误，成功返回OK

```Plaintext
127.0.0.1:6379> lset list 1 go
OK
127.0.0.1:6379> lrange list 0 -1
1) "java"
2) "go"
```

### 集合

#### SADD

语法：`SADD key m1 m2...` 将一或多个 `member` 元素加入到集合 `key` 当中，已经存在于集合的 `member` 元素将被忽略（而列表则不会忽略）。并返回成功添加到集合的数目 与之对应的SMEMBERS（`SMEMBERS key`）则是获取该集合

```Plaintext
127.0.0.1:6379> sadd set m1 m2 m2
(integer) 2
127.0.0.1:6379> smembers set
1) "m1"
2) "m2"
```

#### SPOP

语法：`SPOP key` 移除并返回集合中的一个**随机**元素,并返回被移除的元素，若key不存在或该集合为空时，就返回nil

```Plaintext
127.0.0.1:6379> spop set
"m1"
```

#### SMOVE

语法：`SMOVE source destination member` 如果 `source` 集合不存在或不包含指定的 `member` 元素，则该命令不执行任何操作，仅返回 `0` 。否则， `member` 元素从 `source` 集合中被**移除**，并**添加**到 `destination` 集合中去。 当 `destination` 集合已经包含 `member` 元素时，该命令只是简单地将 `source` 集合中的 `member` 元素删除。

```Plaintext
127.0.0.1:6379> sadd set m1
(integer) 1
127.0.0.1:6379> sadd set m2 m3
(integer) 2
127.0.0.1:6379> sadd set1 m1 m2
(integer) 2
127.0.0.1:6379> smove set set1 m2
(integer) 1
127.0.0.1:6379> smembers set1
1) "m1"
2) "m2"
127.0.0.1:6379> smembers set
1) "m1"
2) "m3"
127.0.0.1:6379> smove set set1 m4
(integer) 0
```

### 有序集合

#### ZADD

语法：`ZADD key score1 member1 score2 member2…` 将一个或多个 `member` 元素及其 `score` 值加入到有序集 `key` 当中，注意这里的member和score是必须配对的，而且score必须是浮点数或者整型，添加成功后返回被成功添加的新成员的数量

```Plaintext
127.0.0.1:6379> zadd zset 1 m1 2 m2 3 m3
(integer) 3
127.0.0.1:6379> zrange zset 0 -1
1) "m1"
2) "m2"
3) "m3"
127.0.0.1:6379> zrange zset 0 -1 WITHSCORES
1) 1 "m1"
2) 2 "m2"
3) 3 "m3"
127.0.0.1:6379> zrem zset m3
(integer) 1
127.0.0.1:6379> zrange zset 0 -1 WITHSCORES
1) 1 "m1"
2) 2 "m2"
```

#### ZRANGE

语法：`ZRANGE key start stop [WITHSCORES]` 返回有序集 `key` 中，并且其中成员的位置按 `score` 值递增(从小到大)来排序（`ZREVRANGE`则按从大到小排列）。 例子如上，这里便不举例了

#### zrem

语法：`ZREM key member1 member2` 移除有序集 `key` 中的一个或多个成员，不存在的成员将被忽略，并返回成功移除的数量 例子如上，这里便不举例了

### Redis事务

我们开始也说了NOSQL是不会保证关系数据的ACID的，详见 https://www.cnblogs.com/chenpingzhao/archive/2015/11/27/5001894.html redis中如何执行事务？ 在redis中multi标志着事务的开始，exec即执行事务

```Plaintext
127.0.0.1:6379> multi
OK
127.0.0.1:6379> set k1 v1                                                                                 <transaction>
QUEUED
127.0.0.1:6379> set k2 v1                                                                                 <transaction>
QUEUED
127.0.0.1:6379> exec                                                                                      <transaction>
1) "OK"
2) "OK"
3) "OK"
127.0.0.1:6379> mget k2 k1
1) "v1"
2) "v1"
```

看看接下来的这段段码有什么问题

```Plaintext
127.0.0.1:6379> multi
OK
127.0.0.1:6379> set k1 v2
QUEUED
127.0.0.1:6379> lpush k2 v2
QUEUED
127.0.0.1:6379> exec
1) OK
2) (error) WRONGTYPE Operation against a key holding the wrong kind of value
127.0.0.1:6379> get k1
"v2"
```

很明显，第一条执行成功，而第二条失败了，这并不复合事务的原子性。 因此，我们的redis事务是不满足平时所说的ACDI性质的，其没有回滚的机制（即事务中的命名执行错误，回滚到事务执行之前）

## go语言操作redis

go有两个比较常见的redis库go-redis和redigo，我们这边使用的是go-redis这个库

```Bash
go get github.com/redis/go-redis/v9
```

#### 连接到redis服务器

```Go
var C *redis.Client
func InitRedis() error {
        C = redis.NewClient(&redis.Options{
                Addr:     "localhost:6379",
                DB:       0,
                Password: "",
        })
        err := C.Ping().Err()
        return err
}
```

### SET

```Go
func Set(key string, value interface{}, expiration time.Duration) error {
        return C.Set(key, value, expiration).Err()
}
```

### GET

```Go
func Get(key string) (string, error) {
        return C.Get(key).Result()
}
```

### ZSET

```Go
func main() {
        err := InitRedis()
        if err != nil {
                log.Fatalln("init redis is err:", err)
                return
        }
        key := "languageRange"
        zset := []redis.Z{
                {
                        Score:  2,
                        Member: "c",
                },
                {
                        Score:  3,
                        Member: "go",
                },
                {
                        Score:  1,
                        Member: "jvav",
                },
        }
        err = ZSet(key, zset)
        if err != nil {
                log.Fatalln("zset redis is err:", err)
                return
        }
        op := redis.ZRangeBy{Min: "1", Max: "3"}
        scores := C.ZRangeByScoreWithScores(key, op)
        if scores.Err() != nil {
                fmt.Printf("zrangebyscore failed, err:%v\n", err)
                return
        }
        for _, z := range scores.Val() {
                fmt.Println(z.Member, z.Score)
        }
}
func ZSet(key string, val []redis.Z) error {
        _, err := C.ZAdd(key, val...).Result()
        if err != nil {
                log.Printf("zadd failed, err:%v\n", err)
        }
        return err
}
```

其他命令你们就自己下去看看吧，这pacakajgiopsdfhgui个框架应该非常简单的了，可以自己进行封装一下 go-redis这个框架可以让你很简单的实现redis的连接池，直接在newClient的时候对参数进行设置即可。看看他的官方说明即可。

## 作业

lv0: 复习一下今天所讲的代码，了解尝试一下连接池 lv1:使用redis实现一个发布订阅模型，比如当你订阅的主题发布了一个帖子，所有关注了这个主题的人都可以得到一个推送知道这个信息。 **截止日期** 下一次上课前 **格式与方式** 和上学期一样