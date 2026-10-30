# 第九课：缓存机制与多级缓存层实现

## 第一部分：为什么需要缓存

### 1. 数据库扛不住的读流量

先说一个电商大促时非常经典的场景：某爆款商品详情页，每秒被点击 10 万次。这些点击大部分都转化为读请求——读商品、读库存、读评价。你的 MySQL 单机每秒查询量通常几千到几万，10 万+的查询请求直接打上去，连接池瞬间打满，后续请求排队超时，页面白屏。

问题不在于 MySQL 不行，而在于你让它做了它最不擅长的事：用少量的资源去重复做同一件事情：查询相同的数据。10 万次查询里，可能有9.5 万次拿到的是同一份商品详情——毫无以为，这 9.5 万次让数据库完成根本就没有必要。

这就引出了缓存：在数据库前面挡一层更快的存储，把大量的读请求截在它那里，只让少量其他请求走到数据库。

### 2. 缓存是什么：高速可读的数据库的小副本

缓存不是一种新的数据库，它是一部分主存数据的高速可读副本。说它快，是因为它常驻内存——内存访问在百纳秒级，而磁盘一次随机读在毫秒级（SSD 稍快，机械盘更慢），两者差一到两个数量级。Redis 一次读通常亚毫秒级，MySQL 一次读通常几毫秒，这就是缓存能加速的物理基础。

但快是有代价的，代价有三：

- **容量小**：内存贵，因此内存往往远远小于主存。故此，主存只存储需要被高速访问的数据。
- **掉电不能留**：内存是易失的，断电、重启、进程崩溃，缓存里的数据全部消失。磁盘靠预写日志（WAL，Write-Ahead Logging）才能掉电不丢，缓存却没有。
- **不能当源头**：缓存里的值来自数据库，是数据库的一个副本。首次启动、缓存全失效时，数据还是要从数据库去读，不能替代数据库。

那为什么"小容量"的缓存能加速"大容量"的数据库？靠的是**局部性原理**：程序访问的数据有聚集性，刚被访问的数据短期内很可能再被访问（时间局部性），热点数据就那么几个。把这一小撮热点放进缓存，绝大多数读就能在缓存命中，不必打到数据库。

具体到开头的场景：10 万次查询加一层缓存，假设命中率 95%（命中率 = 命中次数 / 总查询次数），真正落到数据库的只剩 5000 次，数据库轻松承担。一句话：缓存让面对"读多写少"场景的系统也能扛住。

## 第二部分：缓存与数据库的分工

缓存快但易失、数据库慢但持久，两者各有所长也各有所短——明确了这一点，就能定下分工：持久存储保存最终真实数据，承载写和复杂查询；缓存承载高频读，存的是副本。

### 1. 旁路缓存：读查缓存，写落库

分工落到代码上，最常用的是**旁路缓存（Cache-Aside）**模式——应用自己决定先查缓存还是数据库：

- **读**：先查缓存，命中直接返回；未命中查数据库，把结果回填缓存，再返回。
- **写**：先更新数据库，再**删除**缓存。

```text
读路径：缓存 ──未命中──> 数据库 ──回填──> 缓存 ──> 返回调用方
写路径：数据库 ──成功──> 删缓存
```

注意写路径是"删除"缓存而不是"更新"缓存——暂且按下不表，第三部分解释。

### 2. 当单层缓存不够：多级缓存

一层缓存挡不住时，再叠成多级：

```text
请求 → 进程内缓存(第一级) → Redis(第二级) → 数据库(第三级)
```

- 第一级进程内缓存：纳秒级，无网络，但多实例不共享、容量受堆限制；
- 第二级 Redis：亚毫秒级，多实例共享，可持久化，是主力缓存；
- 第三级数据库：兜底。

每一层只挡"上一层没挡住的流量"。第一级命中率 80%，剩下 20% 打到第二级；第二级再挡掉其中 90%，只剩 2% 打到数据库。层层分摊，数据库压力骤降。这种"分层兜底"的思路，是第六部分要把所有手段串起来的基础。

## 第三部分：缓存与数据库共用，不一致问题

把缓存和数据库放一起用，最大的问题是**数据不一致**：数据库更新了，缓存还是旧值。这一部分拆开看不一致的两个来源和各自的解法。

### 1. 写后"更新"缓存的并发覆盖

第一个不一致来自"写后更新缓存"这个看似自然的动作。线程 A 更新数据库为 1，线程 B 更新数据库为 2，但缓存更新的顺序可能和数据库相反：

```text
A 更新数据库 = 1
B 更新数据库 = 2
B 更新缓存   = 2
A 更新缓存   = 1      ← 顺序错乱

结果：数据库 = 2，缓存 = 1，脏数据
```

**解法一：写后删除，而不是更新。** 删除是幂等的——多次执行效果和一次相同。写后把缓存删掉，下次读自然从数据库拉最新值回填，把不一致窗口压缩到"一次读周期"。这就是第二部分写路径用"删"不用"更新"的原因。

**解法二：写后更新加版本号约束。** 写入时带上递增版本号，更新缓存前比对版本号，只有版本号更新才写入。线程 A 带 v=1、B 带 v=2，A 准备更新时发现 v=1 小于缓存里的 v=2，放弃更新。即使顺序错乱，旧版本也不会覆盖新版本。

### 2. 写后删除的"读旧值回填"窗口

换成"写后删除"之后，第二个不一致又冒出来：线程 A 更新数据库并删缓存，但在 A 删缓存**之前**，线程 B 读到旧缓存未命中、去数据库查了**旧值**（此时 A 还没提交或刚提交但 B 的查询已发出），然后 A 删缓存，B 把旧值回填缓存——结果缓存又是旧值。

**解法一：延迟双删。** 更新数据库前先删一次，更新后再异步延迟（如 500 毫秒）删一次，覆盖"B 回填旧值"的窗口：

```text
删缓存 → 更新数据库 → 等 500 毫秒 → 再删缓存
```

**解法二：版本号/时间戳。** 缓存带上版本号，读时比对，旧版本自动失效。写数据库时版本号递增，缓存写入时携带当前版本号，读取时若缓存版本号小于数据库版本号就重新加载。适合对一致性要求较高的场景。

### 3. 实战取舍：最终一致就够

强一致（同步双写 + 分布式锁）代价高、性能差，多数业务接受**最终一致**——短时间内可能不一致，但最终会一致。写后删 + 短存活时间 + 延迟双删，足够覆盖 99% 场景。只有金融账户这类才上强一致。

### 4. 延迟双删代码

下面是延迟双删的 Go 实现，流程就是上面那张图：

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

## 第四部分：三类缓存灾难

缓存用不好会引发三类经典故障：穿透、击穿、雪崩。名字像，成因不同，解法也不同。

### 1. 缓存穿透：查一个根本不存在的键

**成因**：请求查一个根本不存在的键（如不存在的用户 ID），缓存没有、数据库也没有，每次都打到数据库。恶意攻击时危害极大——攻击者拿一堆不存在的 ID 轮番请求，数据库被空查询打满。

**应对一：缓存空值。** 数据库查不到也缓存一个短存活时间的空标记，下次同一个键直接返回"不存在"，不再打数据库。

**应对二：布隆过滤器。** 在缓存前面加一层布隆过滤器——一种概率型数据结构，用一个位数组 + 多个哈希函数来判断元素"一定不在"或"可能在"集合中。它不存实际数据，只记录"哪些键曾经存在过"。查询时先问布隆过滤器：如果它说"一定不在"，就直接返回不存在，根本不去查缓存和数据库；如果它说"可能在"，才继续往下查。它会有误判（把不存在的判成可能在），但概率可控（如设为 1%），且绝不漏判。空间极省，100 万元素、1% 误判率只需约 1.2MB。

下面是缓存空值的 Go 模拟：

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
```

### 2. 缓存击穿：热点键突然过期

**成因**：一个热点键突然过期，瞬间大量并发请求同时未命中，全部打到数据库重建缓存，数据库瞬时压力激增。和穿透的区别：穿透是"键不存在"，击穿是"键存在但刚好过期"。

**应对一：互斥锁。** 回源时对每个键加锁，同一键只允许一个请求去查数据库重建缓存，其余请求等它建好后直接读缓存。

**应对二：热点键永不过期。** 热点数据只做逻辑过期，物理上不设过期时间，发现逻辑过期时由后台异步刷新。

下面是互斥锁的 Go 模拟：

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

### 3. 缓存雪崩：大量键同时过期

**成因**：大量键在同一时刻过期，或缓存服务整体宕机，请求大面积穿透到数据库，数据库被压垮。和击穿的区别：击穿是一个键，雪崩是一批键。

**应对一：存活时间加随机抖动。** 过期时间 = 基础存活时间 + 随机量，让键失效时间分散开，避免集体过期。批量预热时尤其要带抖动——否则同时过期等于人为制造雪崩。

**应对二：多级缓存兜底。** 第一级进程内缓存作为第二级 Redis 之后的防线，Redis 挂了第一级还能挡一阵。

下面是随机抖动 + 多级兜底的 Go 实现：

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

## 第五部分：缓存也扛不住的高 IO——消息队列

前三部分都在讲缓存如何挡"读"。但有一类高 IO 场景缓存帮不上忙——**高并发写**。这时候要请出另一个组件：消息队列（Message Queue，MQ）。

### 1. 缓存的边界：挡得住读，挡不住写

回顾一下：缓存把热点读从数据库挪走，数据库读压力骤降。但写请求不能走缓存——

- 缓存是易失的，写缓存不等于写库，掉电就丢；
- 写入要保证 ACID，必须落持久存储；
- 高并发写（如秒杀下单 10 万 TPS）直接打数据库，连接池瞬间打满、行锁激烈竞争、主从延迟拉大，系统崩溃。

所以：**缓存挡读，但写的高 IO 需要另一个组件——消息队列**。两者解决的是不同方向的流量问题。

### 2. 消息队列：削峰填谷

消息队列的核心思想：**把同步的"请求 → 立即处理"变成异步的"请求 → 入队 → 慢慢处理"**。

```text
传统同步：请求 ──> 立即写数据库 ──> 返回     （数据库被尖峰冲垮）

消息队列：请求 ──> 入队 ──> 立即返回
                    │
                    ▼
              消费者按数据库承受速率取消息 ──> 写数据库
```

三个角色：生产者接收请求，只把消息塞进队列就立即返回成功；队列是缓冲区，暂存海量消息，可持久化（掉电不丢）；消费者按数据库能承受的速率从队列取消息，匀速写入数据库。

效果是双向的：用户侧请求瞬间被确认，响应快，不感知后端压力；数据库侧写入速率被消费者控制，恒定在安全水位，不被尖峰冲垮。高峰被队列"削平"，低谷时消费者继续处理"填谷"，整体流量平滑。

类比一下：缓存像"读的漏斗"，把海量读收拢成少量读打到库；MQ 像"写的蓄水池"，把海量写暂存起来，按下游能力缓慢放出。两者形态相似，但方向相反。

### 3. 缓存 + 消息队列：一读一写保护数据库

实际高 IO 系统（如电商大促）是两者配合，各管一边：

```text
                  ┌─→ 缓存(Redis) ──命中──→ 返回
用户请求 ─┬─读──→ ┤
           │       └─未命中→ 数据库 → 回填缓存 → 返回
           │
           └─写──→ 消息队列(Kafka) ─→ 消费者(控速) ─→ 数据库
                                           ↓
                                       写后删缓存
```

读路径：缓存挡在数据库前，命中直接返回，未命中才查库并回填。写路径：请求先入 MQ 立即返回，消费者按数据库承受能力（如 2000 TPS）匀速消费，写库后删缓存保证一致性。

这样数据库的**读压力被缓存吸走、写压力被 MQ 整形**，始终运行在安全水位。缓存和 MQ 一读一写，共同把数据库保护起来。

### 4. 配合时的六个坑

两者配合威力大，但坑也多，工程上必须注意：

1. **消费顺序**：同一资源的写必须有序（如订单状态 待支付 → 已发货 → 已完成），要用有序队列（Kafka 单分区、RocketMQ 顺序消息），不能乱序消费，否则状态机错乱。
2. **消息丢失**：生产者要确认入队成功（ACK），消费者处理完再 ACK，中间宕机靠 MQ 持久化重投，不能丢。
3. **幂等消费**：消费者可能重复收到同一条消息（网络重试、重平衡），写入必须幂等——用业务 ID 去重，如 `INSERT ... ON DUPLICATE KEY UPDATE`，或消费前查去重表。幂等的更多实现模式（唯一键约束、状态机校验、token 去重）感兴趣的同学可以去查阅分布式事务资料，这里点到为止。
4. **缓存与 MQ 的一致性**：消费者写库后要删缓存，删缓存失败则缓存残留旧值——和第三部分一样需要延迟双删兜底。
5. **背压降级**：队列积压超过阈值时，生产者要降级（拒绝新请求或返回"系统繁忙"），不能让队列无限增长撑爆 MQ 内存或磁盘。
6. **消费速率与数据库能力匹配**：消费者并发数要小于数据库连接数，否则消费者自己就把数据库打满，MQ 的整形就失去意义。

### 5. 何时用 MQ、何时只用缓存

不是所有场景都要上 MQ，按写量特征选择：

| 场景 | 选择 |
|---|---|
| 读多写少，写量数据库能扛 | 只用缓存 |
| 写量瞬时尖峰，平均量数据库能扛 | 缓存 + MQ 削峰 |
| 写量持续高于数据库能力 | MQ 也救不了，要分库分表 |
| 强同步、要求立即看到写入结果 | 不能用 MQ 异步，只能同步写 + 缓存 |

一句话总结：**缓存解决"读不动"，MQ 解决"写不动"，分库分表解决"存不下"**。三者递进，按需选用。分库分表的具体拆分策略（水平/垂直/哈希/范围分片）及跨片查询、分布式 ID 等难题，感兴趣的同学可以去单独深入了解，本课不展开。

## 第六部分：多级缓存——把所有手段组合起来

前面分别讲了穿透、击穿、雪崩的单一应对，也讲了消息队列处理高并发写。读这一侧的多种手段，实际工程中往往**组合在一个多级缓存结构里**，让每一层各司其职、互相兜底。这一部分给出一个整合的多级缓存实现，把前面的思路串起来。

### 1. 每一层各司其职

| 问题	 | 多级缓存中的应对								    |
|		 |												    |
| 穿透      | 前置布隆过滤器拦截，漏网用缓存空值兜底			    |
| 击穿      | 回源时用单飞（singleflight）合并，同一键只重建一次 |
| 雪崩      | 每层写入带随机抖动，上层失效有下层兜底			    |
| 一致性   | 写后删每一层 + 延迟双删 						    |
| 容量      | 第一级用最近最少使用（LRU）限制大小 			    |

### 2. 整合实现

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

### 3. 使用示例

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

## 总结

本课沿着"缓存为什么快、快了为什么还不够、不够时怎么办"这条主线，讲了六件事：

1. **为什么需要缓存**（第一部分）：数据库扛不住高并发读，缓存在内存里挡一层高速副本，靠局部性原理让小容量加速大流量；
2. **缓存与数据库的分工**（第二部分）：旁路缓存读查缓存、写落库，单层不够就叠多级；
3. **不一致从哪来**（第三部分）：写后更新会并发覆盖，写后删除有读旧值回填窗口，靠写后删 + 延迟双删 + 短存活时间达成最终一致；
4. **三类缓存灾难**（第四部分）：穿透用缓存空值 + 布隆过滤器，击穿用互斥锁 + 永不过期，雪崩用随机抖动 + 多级兜底；
5. **消息队列**（第五部分）：缓存挡读、MQ 挡写，削峰填谷 + 幂等消费共同保护数据库；
6. **多级缓存整合**（第六部分）：把所有手段组合在一个结构里，层层拦截、互相兜底。

把这节课的三句话带走：

> 缓存是真相的高速副本，不是源头——快是用一致性、复杂度换来的，所以写后删而非更新，存活时间必带抖动。

> 缓存挡读、消息队列挡写、分库分表挡存不下——三者递进，按写量特征按需选用。

> 多级缓存不是炫技，是让每一层各司其职、互相兜底：布隆挡穿透，单飞挡击穿，抖动挡雪崩，延迟双删挡不一致。