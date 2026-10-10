# Lesson 9：缓存机制与多级缓存层实现

## 零、这节课看完，应该能回答什么问题？

1. 什么是缓存，为什么快？
2. 为什么快还要持久存储？
3. 为什么只用持久存储不行？
4. 缓存与持久存储分别在做什么？
5. 缓存与持久存储共用时有什么问题，如何解决？
6. 常见的缓存灾难有哪些？我们该如何应对？
7. 如何用多级缓存：用分层兜底解决难题？

## 一、缓存

**缓存**是用高速介质临时存放"高频访问数据"的存储层。它不负责"保存真相"，只负责回答已经被
问过的问题。

缓存之所以快，本质是**介质和访问方式**与持久存储不同：

1. **介质不同**：缓存用内存，持久存储用磁盘。内存直接电信号寻址，磁盘要机械寻道（机械硬盘）或闪存块
   寻址（固态硬盘）。
2. **寻址代价不同**：内存访问是随机寻址，纳秒级；磁盘要先定位磁道或存储块，微秒到毫秒级。
3. **有无网络**：分布式缓存（如Redis）虽在内存，但跨网络，多一跳往返时延（RTT[Round-Trip Time]，即请求一来一回的时间）；进程内缓存
   连网络都省了，纳秒级直达。
4. **有无并发争用**：数据库要维护锁、事务、多版本并发控制（MVCC[Multi-Version Concurrency Control]，即让读写不互相阻塞的机制），高并发时连接
   池和行锁会成为瓶颈；缓存大多是单键原子操作，争用小。

**量化对比（典型延迟）**：

| 操作 | 延迟量级 |
|------|----------|
| 进程内缓存（内存读） | 约 10 纳秒 |
| Redis 单次读取（同机房） | 约 0.2 毫秒 |
| 固态硬盘随机读 | 约 0.1 毫秒 |
| MySQL 单行查询（走索引） | 约 1~5 毫秒 |
| 机械硬盘随机读 | 约 5~10 毫秒 |

进程内缓存比MySQL快约**10万倍**，Redis比MySQL快约**10~25倍**。这就是缓存的价值——把"读"从毫秒拉到亚毫秒甚至纳秒。

---

## 二、持久存储的意义

既然缓存快几个数量级，为什么不全部用缓存、丢掉数据库？三个现实原因：

1. **内存易失**：断电、进程崩溃、重启，内存数据全部消失。持久存储靠磁盘+预写日志（WAL[Write-Ahead Logging]，即先把修改写日志再
   改数据，保证掉电可恢复）才能保证掉电不丢。缓存可以丢，账本不能丢。
2. **内存贵且容量有限**：1TB数据放内存要几十台机器，放固态硬盘一台就够，成本差一两个数量级。缓存只
   能放"值得放"的那一小部分。
3. **缓存是"派生数据"**：缓存里的值来自持久存储，是副本。没有持久存储这个"源头"，缓存就成了无源之水——
   首次启动、缓存全失效时，数据从哪来？

所以持久存储是**真相的归属地**，缓存是**真相的高速副本**。两者缺一不可。

---

## 三、缓存的意义

那能不能反过来，只用数据库、不要缓存？小流量系统可以，但流量上来就不行：

1. **高并发扛不住**：数据库单机每秒查询量通常几千到几万，而热点页面读每秒可能十几万。没有缓存
   挡在前面，连接池瞬间打满，请求排队超时。
2. **延迟高，体验差**：每次读都1~5毫秒，叠加网络和渲染，页面响应几百毫秒，用户体感卡顿。缓存能把热点读
   压到亚毫秒。
3. **数据库资源宝贵**：数据库CPU、IO、连接都是稀缺资源，应该留给"写"和"复杂查询"。把海量重复读也压给数据库，
   是浪费昂贵资源去做最廉价的事。
4. **故障放大**：数据库一旦被打满，整个系统不可用，连写请求也受影响。缓存相当于给数据库加了一道"
   读流量保险丝"。

具体场景：电商商品详情页，大促时某爆款每秒查询10万。直接查库，数据库早挂了；加一层Redis缓存，命中率95%，
落到数据库的只剩5000，数据库轻松扛住。**缓存让"读多写少"的系统在不扩数据库的前提下扛住几十倍
流量**。

---

## 四、缓存与持久存储的分工

理解了"谁也替代不了谁"，就能明确分工：

- **持久存储**：保存权威数据，保证事务的原子性、一致性、隔离性、持久性（ACID[Atomicity, Consistency, Isolation, Durability]），承载写和复杂查询，是"源头"。
- **缓存**：承载高频读，存的是持久存储的副本，可丢可重建，是"加速层"。

分工落到代码上，最常用的是**旁路缓存（Cache-Aside，即应用自己决定先查缓存还是数据库）**模式：

- **读**：先查缓存，命中直接返回；未命中查数据库，把结果回填缓存，再返回。
- **写**：先更新数据库，再**删除**缓存（不是更新缓存，原因下一节讲）。

```
读路径：缓存--未命中-->
数据库--回填-->缓存-->
返回-->调用方
写路径：数据库--成功-->
删缓存
```

当单层缓存不够时，再叠加成**多级缓存**：

```
请求→进程内缓存(第一级)
→Redis(第二级)
→数据库(第三级)
```

- 第一级进程内缓存：纳秒级，无网络，但多实例不共享、容量受堆限制。
- 第二级Redis：亚毫秒级，多实例共享，可持久化，是主力缓存。
- 第三级数据库：兜底源头。

每一层只挡"上一层没挡住的流量"。第一级命中率80%，剩下20%打到第二级；第二级再挡掉其中90%，只剩2%打到数
据库。层层分摊，数据库压力骤降。

---

## 五、缓存与持久存储共用而诞生的问题

把缓存和数据库放一起用，最大的问题是**数据不一致**：数据库更新了，缓存还是旧值。

### 5.1 不一致从哪来

**问题一：写后"更新"缓存的并发覆盖**

线程A更新数据库=1，线程B更新数据库=2，但缓存更新顺序可能错乱：
- A更新数据库=1
- B更新数据库=2
- B更新缓存=2
- A更新缓存=1

结果：数据库=2，缓存=1，**脏数据**。

**解法一：写后删除**写后**删除**缓存，而不是更新。删除是幂等的（多次执行效果和一次相同），下次读自然
从数据库拉最新值回填，把不一致窗口压缩到"一次读周期"。

**解法二：写后更新加版本号约束**写入时带上递增版本号，更新缓存前比对版本号，只有版本号更新
才写入。流程如下：
1. 线程A写数据库，版本号v=1
2. 线程B写数据库，版本号v=2
3. B准备更新缓存，带版本号v=2
4. A准备更新缓存，带版本号v=1
5. A发现v=1小于缓存里的v=2，放弃更新
6. 缓存最终为v=2的值，与数据库一致这样即使更新顺序错乱，旧版本也不会覆盖新版本。

**问题二：写后删除的"读旧值回填"窗口**

线程A更新数据库并删缓存，但在A删缓存**之前**，线程B读到旧缓存未命中、去数据库查了**旧值**（此时A还没
提交或刚提交但B的查询已发出），然后A删缓存，B把旧值回填缓存。结果缓存又是旧值。

**解法一：延迟双删**更新数据库前先删一次，更新后再异步延迟（如500毫秒）删一次，覆盖"B回填旧值"的窗口：

```
删缓存→更新数据库
→等500毫秒→再删缓存
```

**解法二：版本号/时间戳**缓存带上版本号，读时比对，旧版本自动失效。流程如下：
1. 写数据库时，版本号递增
2. 缓存写入时，携带当前版本号
3. 读取时先查缓存，得到值和版本号
4. 再查数据库的当前版本号
5. 若缓存版本号小于数据库版本号，说明缓存已过期
6. 重新从数据库加载，回填新版本号
7. 若版本号相等，缓存有效，直接返回适合对一致性要求较高的场景。

### 5.2 实战取舍

强一致（同步双写+分布式锁）代价高、性能差，多数业务接受**最终一致**（短时间内可能不一致，但最终会
一致）：写后删+短存活时间+延迟双删，足够覆盖99%场景。只有金融账户这类才上强一致。

### 5.3 一致性补救：延迟双删代码

```go
package main

import (
	"context"
	"time"
)

// DelayedDoubleDelete 写后延迟双删
// 流程：删缓存→更新数据库
// →延迟500毫秒→再删缓存
func DelayedDoubleDelete(ctx context.Context, c *LocalCache, key string,
	newVal string, dbUpdate func(context.Context, string, string) error) error {
	// 第一次删除
	c.Del(key)
	// 更新数据库
	if err := dbUpdate(ctx, key, newVal); err != nil {
		return err
	}
	// 延迟第二次删除（异步）
	go func() {
		time.Sleep(500 * time.Millisecond)
		c.Del(key)
	}()
	return nil
}
```

---

## 六、常见的缓存灾难与应对

缓存用不好会引发三类经典故障

### 6.1 缓存穿透

**成因**：请求查一个**根本不存在**的键（如不存在的用户ID），缓存没有、数据库也没有，每次都打到数据库。恶
意攻击时危害极大。

**应对一：缓存空值**数据库查不到也缓存一个短存活时间的空标记，下次同一个键直接返回"不存在"，
不再打数据库。

**应对二：布隆过滤器**在缓存前面加一层布隆过滤器（一种概率型数据结构，用一个位数组+多个哈希
函数来判断元素"一定不在"或"可能在"集合中）。它不存实际数据，只记录"哪些键曾经存在过"。查询时先
问布隆过滤器：如果它说"一定不在"，就直接返回不存在，根本不去查缓存和数据库；如果它说"可能在"，
才继续往下查。它会有误判（把不存在的判成可能在），但概率可控（如设为1%），且绝不漏判（存在的不会判
成一定不在）。空间极省，100万元素、1%误判率只需约1.2MB。

下面给出解法一:缓存空值的模拟
```go
package main

import (
	"context"
	"errors"
	"time"
)

var ErrNotFound = errors.New("not found")

// ---------- 简单的进程内缓存 ----------
type LocalCache struct {
	data map[string]entry
}
type entry struct {
	val interface{}
	exp time.Time
}

func NewLocalCache() *LocalCache {
	return &LocalCache{data: make(map[string]entry)}
}
func (c *LocalCache) Set(key string, val interface{}, ttl time.Duration) {
	c.data[key] = entry{val: val, exp: time.Now().Add(ttl)}
}
func (c *LocalCache) Get(key string) (interface{}, bool) {
	e, ok := c.data[key]
	if !ok || time.Now().After(e.exp) {
		return nil, false
	}
	return e.val, true
}
func (c *LocalCache) Del(key string) { delete(c.data, key) }

// ---------- 模拟数据库 ----------
func dbQuery(ctx context.Context, id string) (string, error) {
	if id == "404" {
		return "", ErrNotFound // 模拟不存在
	}
	return "user-" + id, nil
}

// ---------- 缓存空值防穿透 ----------
// nullSentinel 标记"查过了，不存在"
type nullSentinel struct{}

func GetWithNullCache(ctx context.Context, c *LocalCache, id string) (string, error) {
	if v, ok := c.Get(id); ok {
		if _, isNull := v.(nullSentinel); isNull {
			// 命中空值标记，直接返回
			return "", ErrNotFound
		}
		return v.(string), nil
	}
	// 未命中，查数据库
	val, err := dbQuery(ctx, id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			// 把"不存在"也缓存起来
			c.Set(id, nullSentinel{}, 30*time.Second)
			return "", ErrNotFound
		}
		return "", err
	}
	c.Set(id, val, 5*time.Minute)
	return val, nil
}
### 6.2 缓存击穿

**成因**：一个**热点键突然过期**，瞬间大量并发请求同时未命中，全部打到数据库重建缓存，数据库瞬时
压力激增。

**应对**：
- **互斥锁**：手动对每个键加锁，效果类似。
- **热点键永不过期**：热点数据只做逻辑过期，热点数据实际上没有过期时间，出现逻辑过期数据时由后台异步刷新。

下面给出解法一:互斥锁的代码模拟
```go
package main

import (
	"context"
	"sync"
	"time"
)

// ---------- 互斥锁防击穿 ----------
type MutexCache struct {
	c      *LocalCache
	mu     sync.Mutex
	locks  map[string]*sync.Mutex
	loader func(ctx context.Context, key string) (string, error)
	ttl    time.Duration
}

func NewMutexCache(c *LocalCache, loader func(context.Context, string) (string, error), ttl time.Duration) *MutexCache {
	return &MutexCache{c: c, locks: make(map[string]*sync.Mutex), loader: loader, ttl: ttl}
}

func (m *MutexCache) Get(ctx context.Context, key string) (string, error) {
	if v, ok := m.c.Get(key); ok {
		return v.(string), nil
	}
	// 取/创建该键专属的锁
	m.mu.Lock()
	lk, ok := m.locks[key]
	if !ok {
		lk = &sync.Mutex{}
		m.locks[key] = lk
	}
	m.mu.Unlock()

	lk.Lock()
	defer lk.Unlock()
	// 二次检查
	if v, ok := m.c.Get(key); ok {
		return v.(string), nil
	}
	val, err := m.loader(ctx, key)
	if err != nil {
		return "", err
	}
	m.c.Set(key, val, m.ttl)
	return val, nil
}
```

### 6.3 缓存雪崩

**成因**：**大量键在同一时刻过期**，或缓存服务整体宕机，请求大面积穿透到数据库，数据库被压垮。

**应对**：
- **存活时间加随机抖动**：过期时间=基础存活时间+随机量，让键失效时间分散开，避免集体过期。
- **多级缓存兜底**：第一级进程内缓存作为第二级Redis之后的防线，Redis挂了第一级还能挡一阵。

```go
package main

import (
	"context"
	"math/rand"
	"time"
)

// ---------- 方案一：随机抖动防雪崩 ----------
// 给基础存活时间加随机量
func SetWithJitter(c *LocalCache, key string, val interface{}, baseTTL, jitter time.Duration) {
	exp := baseTTL + time.Duration(rand.Int63n(int64(jitter)))
	c.Set(key, val, exp)
}

// 批量预热务必带抖动
// 否则同时过期=人为雪崩
func WarmupWithJitter(c *LocalCache, keys []string, loader func(string) (string, error)) {
	for _, k := range keys {
		if v, err := loader(k); err == nil {
			// 基础5分钟+0~60秒抖动
			SetWithJitter(c, k, v, 5*time.Minute, 60*time.Second)
		}
	}
}

// ---------- 方案二：多级兜底防雪崩 ----------
type TwoLevelCache struct {
	local  *LocalCache
	remote *LocalCache // 模拟Redis，实际替换
	loader func(ctx context.Context, key string) (string, error)
	ttl    time.Duration
}

func NewTwoLevel(local, remote *LocalCache, loader func(context.Context, string) (string, error), ttl time.Duration) *TwoLevelCache {
	return &TwoLevelCache{local: local, remote: remote, loader: loader, ttl: ttl}
}

func (t *TwoLevelCache) Get(ctx context.Context, key string) (string, error) {
	// 第一级
	if v, ok := t.local.Get(key); ok {
		return v.(string), nil
	}
	// 第二级
	if v, ok := t.remote.Get(key); ok {
		// 回填第一级，带抖动
		SetWithJitter(t.local, key, v, t.ttl, 60*time.Second)
		return v.(string), nil
	}
	// 回源
	val, err := t.loader(ctx, key)
	if err != nil {
		return "", err
	}
	SetWithJitter(t.local, key, val, t.ttl, 60*time.Second)
	SetWithJitter(t.remote, key, val, t.ttl, 60*time.Second)
	return val, nil
}
```


---

## 七、多级缓存：用分层兜底解决难题

前面分别讲了穿透、击穿、雪崩的单一应对。实际工程中，这些手段往往**组合在一个多级缓存结构里**，
让每一层各司其职、互相兜底。本节给出一个整合的多级缓存实现，把前面的思路串起来。

### 7.1 多级缓存如何解决问题

| 问题 | 多级缓存中的应对 |
|------|------|
| 穿透 | 前置布隆过滤器拦截，漏网用缓存空值兜底 |
| 击穿 | 回源时用单飞合并，同一键只重建一次 |
| 雪崩 | 每层写入带随机抖动，上层失效有下层兜底 |
| 一致性 | 写后删每一层 + 延迟双删 |
| 容量 | 第一级用最近最少使用（LRU[Least Recently Used]）限制大小 |

### 7.2 整合实现

```go
package cache

import (
	"context"
	"errors"
	"hash/fnv"
	"math"
	"math/rand"
	"sync"
	"time"

	"golang.org/x/sync/singleflight"
)

var ErrNotFound = errors.New("cache: key not found")

// ---------- 第一级：进程内缓存 ----------
type item struct {
	val interface{}
	exp time.Time
}

type LocalCache struct {
	data map[string]item
	mu   sync.RWMutex
}

func NewLocalCache() *LocalCache {
	return &LocalCache{data: make(map[string]item)}
}

func (c *LocalCache) Set(key string, val interface{}, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data[key] = item{val: val, exp: time.Now().Add(ttl)}
}

func (c *LocalCache) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	it, ok := c.data[key]
	c.mu.RUnlock()
	if !ok || time.Now().After(it.exp) {
		return nil, false
	}
	return it.val, true
}

func (c *LocalCache) Del(key string) {
	c.mu.Lock()
	delete(c.data, key)
	c.mu.Unlock()
}

// ---------- 布隆过滤器：前置拦截 ----------
type BloomFilter struct {
	bits []uint64
	k    int
	m    int
}

func NewBloomFilter(n int, fp float64) *BloomFilter {
	if n <= 0 {
		n = 1000
	}
	if fp <= 0 || fp >= 1 {
		fp = 0.01
	}
	m := int(math.Ceil(-1 * float64(n) * math.Log(fp) / (math.Log(2) * math.Log(2))))
	k := int(math.Ceil(float64(m) / float64(n) * math.Log(2)))
	return &BloomFilter{bits: make([]uint64, (m+63)/64), k: k, m: m}
}

func (b *BloomFilter) idx(data []byte) []int {
	h := fnv.New128a()
	h.Write(data)
	s := h.Sum128()
	h1, h2 := uint64(s[0])^uint64(s[1]), uint64(s[1])
	out := make([]int, b.k)
	for i := 0; i < b.k; i++ {
		out[i] = int((h1 + uint64(i)*h2) % uint64(b.m))
	}
	return out
}

func (b *BloomFilter) Add(key string) {
	for _, p := range b.idx([]byte(key)) {
		b.bits[p/64] |= 1 << uint(p%64)
	}
}

func (b *BloomFilter) MayContain(key string) bool {
	for _, p := range b.idx([]byte(key)) {
		if (b.bits[p/64] & (1 << uint(p%64))) == 0 {
			return false
		}
	}
	return true
}

// ---------- 多级缓存管理器 ----------
type Loader func(ctx context.Context, key string) (interface{}, error)

type MultiLevel struct {
	local  *LocalCache
	remote *LocalCache // 模拟第二级Redis
	bloom  *BloomFilter
	group  singleflight.Group
	loader Loader
	ttl    time.Duration
}

func NewMultiLevel(local, remote *LocalCache, bloom *BloomFilter,
	loader Loader, ttl time.Duration) *MultiLevel {
	return &MultiLevel{local: local, remote: remote, bloom: bloom, loader: loader, ttl: ttl}
}

// jitter 给存活时间加随机抖动
func jitter(base time.Duration) time.Duration {
	return base + time.Duration(rand.Int63n(int64(base/10)))
}

// nullSentinel 标记"不存在"，防穿透
type nullSentinel struct{}

// Get 多级查找：布隆→第一级→第二级→回源
func (m *MultiLevel) Get(ctx context.Context, key string) (interface{}, error) {
	// 布隆过滤器前置拦截
	if m.bloom != nil && !m.bloom.MayContain(key) {
		return nil, ErrNotFound
	}
	// 第一级
	if v, ok := m.local.Get(key); ok {
		return v, nil
	}
	// 第二级
	if v, ok := m.remote.Get(key); ok {
		m.local.Set(key, v, jitter(m.ttl)) // 回填第一级
		return v, nil
	}
	// 单飞回源，防击穿
	v, err, _ := m.group.Do(key, func() (interface{}, error) {
		// 二次检查
		if v, ok := m.local.Get(key); ok {
			return v, nil
		}
		val, err := m.loader(ctx, key)
		if err != nil {
			if errors.Is(err, ErrNotFound) {
				// 缓存空值，短存活时间
				m.local.Set(key, nullSentinel{}, 30*time.Second)
				return nil, ErrNotFound
			}
			return nil, err
		}
		// 回填两级，都带抖动
		m.local.Set(key, val, jitter(m.ttl))
		m.remote.Set(key, val, jitter(m.ttl))
		if m.bloom != nil {
			m.bloom.Add(key)
		}
		return val, nil
	})
	return v, err
}

// Invalidate 写后删两级+延迟双删
func (m *MultiLevel) Invalidate(key string) {
	m.local.Del(key)
	m.remote.Del(key)
	go func() {
		time.Sleep(500 * time.Millisecond)
		m.local.Del(key)
		m.remote.Del(key)
	}()
}
```

### 7.3 使用示例

```go
package main

import (
	"context"
	"fmt"
	"time"

	"yourpkg/cache"
)

func main() {
	ctx := context.Background()
	local := cache.NewLocalCache()
	remote := cache.NewLocalCache() // 实际替换为Redis
	bloom := cache.NewBloomFilter(100000, 0.01)
	bloom.Add("user:1") // 预登记存在的键

	loader := func(ctx context.Context, key string) (interface{}, error) {
		if key == "user:404" {
			return nil, cache.ErrNotFound
		}
		return "Alice", nil
	}

	ml := cache.NewMultiLevel(local, remote, bloom, loader, 5*time.Minute)

	v, err := ml.Get(ctx, "user:1")
	fmt.Println("命中:", v, err)

	ml.Invalidate("user:1") // 写后失效
	fmt.Println("已失效，下次读会回源")
}
```

---

## 八、总结

缓存的本质是**用"一致性"换"性能"、用"复杂度"换"吞吐"**。它和持久存储是一对分工明确的搭档：

- 持久存储负责"对"——掉电不丢、强一致、是真相归属地。
- 缓存负责"快"——把热点读从毫秒压到亚毫秒甚至纳秒，是真相的高速副本。

两者必须共用，于是带来两类问题：**一致性**与**灾难性故障**。

- 一致性靠"写后删+延迟双删+短存活时间"解决，多数业务接受最终一致。
- 穿透靠"缓存空值+布隆过滤器"解决。
- 击穿靠"单飞合并+互斥锁"解决。
- 雪崩靠"存活时间随机抖动+多级兜底"解决。
- 多级缓存把这些手段组合在一个结构里，层层拦截、互相兜底，是工程上的综合解。

记住几条铁律：**存活时间必带抖动、写后删除而非更新、热点键永不过期加异步刷新、缓存故障要有
降级预案**。做到这些，缓存才能稳定地发挥它"快几个数量级"的价值，而不是在某个高峰时刻反噬整
个系统。