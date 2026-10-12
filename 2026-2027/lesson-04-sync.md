# 第四课：channel、sync 与锁

第三课我们弄清了写下 `go` 时发生了什么，但第三课的练习会暴露一件事——**能并发跑起来和结果是正确的是两个完全独立的问题**

这节课要回答四个问题：

1. **并发程序到底会以哪些方式出错？**（第一部分）
2. **Go 提供了哪些保证正确性的工具？每个工具为什么存在、怎么用、有什么代价？**（第二到第六部分）
3. **这些工具组合起来能实现什么？**（第七部分）
4. **出现并发相关问题该如何排查？**（第八、九部分）

---

## 一、并发程序会怎么错

### 1. main 返回，goroutine 被静默杀死

```go
package main

import "fmt"

func main() {
    go fmt.Println("hello from goroutine")
}
```

这段程序大多数时候什么都不打印。原因是 `go` 只做一件事：**创建一个 goroutine 并让它变成可运行状态**，它不等待，不承诺执行时机。`main` 一旦返回，runtime 就结束整个进程，所有其他 goroutine 连同它们没跑完的代码一起消失——没有报错，日志或者异常。

想让协程执行完，你可能会直接加延迟

```go
func main() {
    go fmt.Println("hello from goroutine")
    time.Sleep(time.Second) // 猜测一秒内它能跑完
}
```
但是问题显而易见，这样只能让当前的代码得到期望的输出，但是并没有真的确保任何情况下一个进程都能完整的执行。第二部分到第四部分的 `channel` 和 `WaitGroup` 提供的正是这件事。

### 2. 共享变量的读改写不是一步完成的

```go
func main() {
    var n int
    var wg sync.WaitGroup
    wg.Add(2)
    for i := 0; i < 2; i++ {
        //启用两个携程
        go func() {
            defer wg.Done()
            for j := 0; j < 100000; j++ {
                n++
                //分别对n+100000
            }
        }()
    }
    wg.Wait()
    fmt.Println(n) // 期望 200000，实际经常小于它
}
```

`n++` 实际有三步：**读 `n` 的值 → +1 → 写回 `n`**。两个 goroutine 完全可能同时读到 `100`，各自算出 `101`，再各自写回 `101`。每次运行都有可能得到不同的结果。

这类问题叫**数据竞争（data race）**：两个 goroutine 并发访问同一块内存，其中至少一个是写，并且没有用同步原语建立访问顺序。Go 自带的检测器可以直接发现它：

```bash
go run -race main.go
go test -race ./...
```

报告长这样（节选）：

```text
==================
WARNING: DATA RACE
Read at 0x00c0000ba010 by goroutine 7:
  main.main.func1()
      /home/you/race/main.go:12 +0x2c
Previous write at 0x00c0000ba010 by goroutine 6:
  main.main.func1()
      /home/you/race/main.go:12 +0x44
==================
Found 1 data race(s)
exit status 66
```

读报告的方法是**看两处访问的位置**：谁在读、谁在写、分别在代码哪一行。


关于数据竞争有两个重点：

- **某次输出正确，不等于程序没有竞态。** 竞争是概率事件，竞态检测器看的是有没有**建立顺序**，不是**结果对不对**。
- 数据竞争在 C 里也许只是**结果有问题**，但在 Go 里更严重：**并发的 map 写会让整个进程直接 `fatal error: concurrent map writes` 退出**


### 3. 三个问题，三个工具

上面两个问题之外，还有第三个：**多个 goroutine 的任务什么时候算全部做完**，这在第三课的练习里已经遇到过。

```go
for i := 0; i < 3; i++ {
    go work(i)
}
// 这里怎么知道三个 work 都结束了？
```

把三类问题放在一起看，工具的分工就清楚了：

| 问题现象 | 本质 | 需要的工具 |
| --- | --- | --- |
| 什么都没打印 / 结果不完整 | 没有等待协程结束 | `sync.WaitGroup`、`channel` |
| 计数丢失、map 崩溃 | 共享内存被并发读写 | `sync.Mutex` / `RWMutex` / `atomic` |
| 结果传不出来 | 没有传递数据的通道 | `channel` |
| 无法中途停下 | 没有取消机制 | `select` + `context` |
| 初始化被重复执行 | 没有只做一次的表达 | `sync.Once` |

```text
                    一个 goroutine 需要和别的 goroutine 建立的关系
                                    │
        ┌───────────────┬───────────┼───────────┬───────────────┐
        ▼               ▼           ▼           ▼               ▼
    传数据/传任务     等它结束    通知它停止   保护同一块内存   只初始化一次
        │               │           │           │               │
        ▼               ▼           ▼           ▼               ▼
     channel        WaitGroup   close(done)   Mutex/RWMutex    Once
                                context.Done()  atomic        Pool
        │               │           │           │
        └───────────────┴───────────┴───────────┘
                     它们的共同点：
              都在两个 goroutine 之间建立先后顺序
```

### 4. 为什么**看起来对**不算对：happens-before

同步原语的真正价值是**建立顺序保证**。

可以这样理解：编译器和 CPU 都会为了性能重排指令；每个 CPU 核心还有自己的缓存，一个核心写下的值不保证能立刻被另一个核心看到。

所以**A 先执行、B 后执行**这件事，如果没有任何同步手段，程序里是不保证成立的。Go 的内存模型给出的说法是：**只有当操作 X happens before 操作 Y 时，X 写入的值才保证能被 Y 看到。**

| 操作 A | 操作 B |
| --- | --- |
| 向 channel 发送 | 对应的接收完成 |
| `close(ch)` | 从已关闭的 `ch` 收到零值 | 
| `mu.Unlock()` | 同一把 `mu` 的下一次 `Lock()` 返回 | 
| `wg.Done()` | 对应的 `wg.Wait()` 返回 | 
| `once.Do(f)` 里的 `f` 返回 | 任何 `once.Do` 返回 | 
| `atomic` 的某次写 | 后续观察到这个值的 `atomic` 读 |

以上的同步原语已经确保实现了 A happens before B

反过来说，**只要用了这些原语中的任意一种把顺序建立起来，就能保证A 先执行、B 后执行**，不会出现前面提到的问题。

---

## 二、channel：在 goroutine 之间传递数据

### 1. 为什么需要 channel

假设 Go 没有 channel。你想让一个 worker 算完把结果交给 main，只能这样：

```go
var (
    mu      sync.Mutex
    result  int
    hasData bool
)

func worker() {
    v := compute()
    mu.Lock()
    result = v
    hasData = true
    mu.Unlock()
}

func main() {
    go worker()
    for { // 轮询
        mu.Lock()
        ok := hasData
        mu.Unlock()
        if ok {
            break
        }
        time.Sleep(time.Millisecond) // 要么浪费 CPU，要么增加延迟
    }
}
```

这个方法能跑，但暴露了三个问题：

1. **有数据了这件事本身要靠一个额外变量和时间轮询来表达。** 延迟和 CPU 消耗二选一。
2. **多个 worker 时没法直接扩展**：得加计数、加锁、再判断。
3. **没法表达任务结束了**：`hasData` 只能说明**有过数据**，不能说明**不会再有数据了**。

channel 把**通信**做成了语言的一等公民，Go 推崇的写法是：

> **不要通过共享内存来通信，而要通过通信来共享内存。**

`channel` 是一条**带类型的管道**，同时具备四个属性：

| 属性 | 含义 | 带来的能力 |
| --- | --- | --- |
| 类型化 | `chan T` 只能传 `T` | 不需要类型断言，编译期就能查错 |
| 同步点 | 发送和接收构成一次配对 | 天然建立了 happens-before |
| 可关闭 | 关闭是一个**对所有接收者同时生效**的信号 | 广播，且能表达**没有更多数据了** |
| 有容量 | 无缓冲 / 有缓冲 | 决定生产者能跑多快、要不要等消费者 |

### 2. 创建 channel：`make(chan T)` 与 `make(chan T, n)`

```go
ch := make(chan int)      // 无缓冲：容量为 0
buf := make(chan int, 3)  // 有缓冲：容量为 3
```

为什么必须用 `make`？回想第二课的 `make` vs `new`：`make` 负责初始化那些**需要内部结构**的类型（切片、map、channel）。channel 内部有一把锁、一个等待接收者队列、一个等待发送者队列和（可能的）缓冲区，用零值当不了用：

```go
var ch chan int // nil channel
ch <- 1         // 永久阻塞
```

**缓冲区到底解决什么问题。** 无缓冲 channel 要求发送和接收**同时就位**，它把两个 goroutine 的节奏焊死在一起。有缓冲 channel 则允许发送方在缓冲区没满时先走一步，从而：

- 削峰：消费端偶尔慢一点，不至于立刻把生产端卡住；
- 减少调度：一批小数据不用每一条都做一次**交接**；
- 但**不解决根本的速率差**：生产长期快于消费，缓冲区迟早会满，发送照样阻塞。

| 类型 | 发送阻塞条件 | 接收阻塞条件 | `cap` | 典型用途 |
| --- | --- | --- | --- | --- |
| `nil` | 永久阻塞 | 永久阻塞 | 0 | 在 `select` 里临时禁用分支 |
| `make(chan T)` | 没有接收者就绪 | 没有发送者就位 | 0 | 严格交接：任务分发、握手 |
| `make(chan T, n)` | 缓冲区已满且无接收者 | 缓冲区为空且无发送者 | n | 削峰、限流、异步投递 |

`len(ch)` 返回缓冲区里当前待取的元素个数，`cap(ch)` 返回容量：

```go
ch := make(chan int, 3)
ch <- 1
fmt.Println(len(ch), cap(ch)) // 1 3
```

**但请不要用 `len(ch)` 做逻辑判断。** 在你看完这个数字到用它做决定之间，别的 goroutine 可能已经改了它，而 `len` 本身不参与任何同步。想表达**有就取，没有就走开**，用 `select` + `default`（第三部分）。

### 3. 发送与接收：`<-` 的四种写法

`<-` 是 channel 唯一的操作符，方向决定含义：箭头指向 channel 是发送，从 channel 指出来是接收。

| 写法 | 含义 | 什么时候用 |
| --- | --- | --- |
| `ch <- v` | 发送 | 投递任务、结果、信号 |
| `v := <-ch` | 接收，忽略状态 | 确定 channel 不会关闭 |
| `v, ok := <-ch` | 接收并判断是否已关闭 | 区分**零值**和**已关闭** |
| `<-ch` | 只接收，丢弃值 | 只关心**事件发生了**，例如等一个信号 |

`ok` 这个返回值为什么必须存在？因为接收一个已关闭的 channel **不会报错**，它立刻返回零值。看这段代码：

```go
func consume(in <-chan int) {
    for {
        v := <-in
        if v == 0 { // 想用 0 当作结束标记
            return
        }
        handle(v)
    }
}
```

如果上游真的发来一个 `0`，这个函数就提前退出了。正确做法是让**结束**由 `close` 表达，而不是占用一个业务值：

```go
func consume(in <-chan int) {
    for {
        v, ok := <-in
        if !ok { // channel 已关闭且数据取完
            return
        }
        handle(v)
    }
}
```

这段其实就是 `for v := range in` 展开后的样子，所以日常写 `range` 就够了（见第 8 小节）。

### 4. 通道方向：`chan<- T` 与 `<-chan T`

```go
func produce(out chan<- int) { out <- 1 }        // 只能发送
func consume(in <-chan int)  { v := <-in; _ = v } // 只能接收
```

方向类型存在的理由是**把使用契约写进函数签名**：函数作者一旦声明 `<-chan int`，函数体里就不可能误发数据，编译器直接拦住；调用方看签名就知道这个 channel 属于**谁**。这在多人协作和后续维护中省掉很多口头约定。

转换规则是不对称的，这一点和第二课**方法域**的不对称性同源：

| 从 | 到 | 允许吗 |
| --- | --- | --- |
| `chan T` | `chan<- T` | 允许（双向收窄为只写） |
| `chan T` | `<-chan T` | 允许（双向收窄为只读） |
| `chan<- T` | `chan T` | 不允许 |
| `<-chan T` | `chan T` | 不允许 |

### 5. `nil` channel：不是 bug，是特性

`nil` channel 的发送和接收会**永久阻塞**，`close(nil)` 会 panic。听起来像陷阱，但它有一个非常重要的用途：**在 `select` 中禁用某个分支**。

```go
func merge2(a, b <-chan int, out chan<- int) {
    for a != nil || b != nil {
        select {
        case v, ok := <-a:
            if !ok {
                a = nil // 第一个源读完了，从此这个 case 永远不会被选中
                continue
            }
            out <- v
        case v, ok := <-b:
            if !ok {
                b = nil
                continue
            }
            out <- v
        }
    }
}
```

把 `a` 置为 `nil` 之后，`case v, ok := <-a` 变成一个永远不就绪的分支，`select` 自然只从 `b` 上取——不需要额外的布尔标志，也不需要写两遍循环。这是**用数据流的状态控制控制流**的典型手法。

### 6. 关闭通道：`close()` 的完整语义

**为什么需要关闭。** 没有 `close`，接收方永远无法区分两种情况：**上游暂时没有数据**和**上游永远不会再有数据**。它只能靠约定（发一个哨兵值，例如上面的 `0`）、超时或者干脆不退出——这就是 goroutine 泄漏的源头。`close` 用一条语句表达**没有更多数据了**。

完整的语义如下，这张表值得记牢：

| 操作 | 结果 |
| --- | --- |
| `close(ch)` | 标记 channel 为已关闭 |
| 关闭后再接收 | 先取完缓冲区里剩余的元素，**取完后立刻返回零值 + `false`**（不阻塞） |
| 已关闭的 channel 上 `for range` | 取完剩余元素后循环自动结束 |
| `close(ch)` 第二次 | **panic: close of closed channel** |
| 向已关闭的 channel 发送 | **panic: send on closed channel** |
| 从已关闭的 channel 接收 | 永不阻塞（这是**广播**能力的来源） |
| `close(nil)` | panic |

注意**从已关闭 channel 接收永远不阻塞**这一条：**关闭是一个对所有接收者同时生效的信号**。这就是为什么它天然适合做**停止**广播——一次 `close(done)`，所有正在 `<-done` 或 `select` 中等待 `done` 的 goroutine 会**同时**被唤醒：

```go
done := make(chan struct{})
for i := 0; i < 5; i++ {
    go func() {
        <-done // 5 个 goroutine 一起等
        fmt.Println("stop")
    }()
}
close(done) // 一条语句唤醒全部
```

对比一下**发送一个值来通知**的做法：

| 通知方式 | 能否唤醒多个等待者 | 是否需要知道等待者数量 | 能否重复通知 |
| --- | --- | --- | --- |
| `close(done)` | 能，全部同时 | 不需要 | 一次（再关就 panic） |
| `done <- struct{}{}` | 只能唤醒一个 | 需要 | 可以发多次，但没等待者时会阻塞 |
| `context.Context` | 能（内部就是关 channel） | 不需要 | 一次，但可带超时和原因 |

### 7. 谁有权关闭：所有权规则

`close` 会 panic 的两条规则（重复关闭、向已关闭发送）都指向同一个结论：

> **由发送方关闭，接收方永远不关闭；有多个发送方时，由统一的协调者关闭。**

理由是信息不对称：只有写数据的人才知道**我发完了**。接收方不知道还有没有别的发送者，它一关，其他发送者继续发送就会 panic。看这个错误示范：

```go
func bad() {
    ch := make(chan int, 1)
    go func() { ch <- 1 }() // 发送方
    go func() {
        <-ch
        close(ch) // 接收方关闭：发送方还可能再发，直接 panic
    }()
}
```

多个发送方的正确写法是：让所有生产者结束后，由一个专门的 goroutine 关闭公共出口。

```go
func merge(inputs ...<-chan int) <-chan int {
    out := make(chan int)
    var wg sync.WaitGroup
    wg.Add(len(inputs))
    for _, in := range inputs {
        go func(in <-chan int) {
            defer wg.Done()
            for v := range in {
                out <- v
            }
        }(in)
    }
    go func() {
        wg.Wait()   // 等所有生产者退出
        close(out)  // 只有这一个 goroutine close
    }()
    return out
}
```

这段代码里有三个关键点，值得逐个拆开：

1. `in` 被当作**参数传入**闭包，而不是直接捕获循环变量。Go 1.22 起 `for range` 的循环变量是每轮独立的，可以不写参数；但在 1.22 之前必须这样传，否则所有 goroutine 会读到同一个（循环结束后最终的）值。写参数是兼容性更好的习惯。
2. `wg.Wait()` 在**另一个 goroutine** 里执行。如果直接在 `merge` 里 `wg.Wait()` 再 `close`，函数就会阻塞到所有输入处理完才能返回 `out`——那就退化成同步了。
3. `close(out)` 只可能被执行一次，因为 `wg.Wait()` 只返回一次。

### 8. 遍历：`for range` 与 `for select` 的分工

| 写法 | 何时结束 | 能否同时处理取消 | 适用场景 |
| --- | --- | --- | --- |
| `for v := range ch` | channel 关闭且取完 | 不能 | 简单消费，明确知道上游会关闭 |
| `for { select { ... } }` | 由 `return` 决定 | 能 | 需要响应 `done` / `ctx.Done()` |

```go
// 简单版
for v := range ch {
    handle(v)
}

// 完整版：能被打断
for {
    select {
    case <-done:
        return
    case v, ok := <-ch:
        if !ok {
            return // 上游关闭且取完
        }
        handle(v)
    }
}
```

一旦代码里出现**取消**，就该用 `select` 版本；反之能用 `range` 就说明这条链路是简单直线，是好代码的标志。

### 9. channel 的五种典型用法

channel 不只是**传数据**，下面五种用法都来自同一套语义：

| 用法 | 怎么写 | 场景 |
| --- | --- | --- |
| 数据流 / 任务队列 | `ch <- job` / `job, ok := <-ch` | worker pool、流水线 |
| 结果回收 | 每个 worker 把结果发到同一个 `results` | 并发任务收集 |
| 停止广播 | `close(done)`，全员 `<-done` | 优雅关闭、取消 |
| 计数信号量（限流） | 带缓冲 channel 当令牌桶，`sem <- struct{}{}` / `<-sem` | 限制并发数 |
| 互斥 / 单槽缓存 | `cap = 1` 的 channel | 保护单个值、最新值缓存 |

限流（信号量）的写法值得单独看，因为在真实项目里出现频率极高：

```go
func limit(limit int, jobs []int) {
    sem := make(chan struct{}, limit) // 最多同时 limit 个
    var wg sync.WaitGroup
    for _, job := range jobs {
        wg.Add(1)
        go func(job int) {
            defer wg.Done()
            sem <- struct{}{}        // 取令牌，满了就在这里等
            defer func() { <-sem }() // 还令牌
            handle(job)
        }(job)
    }
    wg.Wait()
}
```

用 `cap = 1` 的 channel 也能做互斥，但**不推荐**作为通用锁：

| 需求 | `cap=1` channel | `sync.Mutex` |
| --- | --- | --- |
| 加锁 / 解锁 | `ch <- struct{}{}` / `<-ch` | `Lock` / `Unlock` |
| 非阻塞尝试加锁 | `select` + `default` 可以 | `TryLock`（1.18+） |
| 读写锁（多读单写） | 做不到 | `RWMutex` |
| 跨 goroutine 解锁 | 可以（这正是问题） | 可以（同样的坑） |
| 性能 | 更差（要过调度） | 更好 |

### 10. channel 的代价与边界

- **每次传递都有开销**：涉及 runtime 的调度与锁，比直接读写内存慢得多。高频计数、超小数据用 `atomic` 或 `Mutex` 更合适。
- **它解决的是通信，不是保护**。多个 goroutine 读写同一个 map 时，你不能**用 channel 保护 map**，只能把 map 的所有权收拢到一个 goroutine 里，其他人通过 channel 请求它代劳（这叫用通信代替共享）。如果做不到，就老老实实上锁。
- **它带来两类新 bug**：发送没人接收会永久阻塞（死锁），接收方提前退出会让发送方永远卡住（goroutine 泄漏）。第三部分的 `select` + `context`，就是为了收拾这两类问题而存在的。

---
## 三、select：同时等多个事件

### 1. 为什么需要 select

一个 worker 通常有两个等待对象：**新任务来了要处理**，**收到取消要退出**。只有 channel 没有 `select` 时，你只有三条路：

| 笨办法 | 问题 |
| --- | --- |
| 阻塞在一个 channel 上 | 无法同时响应取消，退出只能靠上游关闭 |
| `for` 循环 + `default` 轮询 | 忙等，把 CPU 烧在一个什么都不做的循环上 |
| 为每个 channel 各开一个 goroutine，转发到同一个 channel | 多出一层 goroutine，而这些**搬运工**自己也需要退出条件 |

`select` 把**多路等待**做成一条语句，并且直接由 runtime 实现——**它不会为每个 `case` 创建 goroutine**。

### 2. 语义要点

```go
select {
case v := <-results:
    fmt.Println("result:", v)
case <-done:
    fmt.Println("cancelled")
default:
    fmt.Println("no event right now")
}
```

| 规则 | 说明 |
| --- | --- |
| `case` 必须是 channel 操作 | 接收或发送，普通表达式不行 |
| 多个 `case` 同时就绪 | **伪随机**选一个执行（防止某个 case 永远被饿死） |
| 都不就绪 | 阻塞等待；若有 `default`，立刻执行 `default` 并结束 |
| `nil` channel 的 `case` | 永远不就绪 → 可用于动态禁用分支（见第二部分第 5 小节） |
| `case` 的求值 | 进入 `select` 时先求出所有 channel 操作数，再判断就绪 |
| `select {}` | 永久阻塞（如果所有 goroutine 都阻塞，runtime 会报 `fatal error: all goroutines are asleep - deadlock!`） |
| `break` | 只跳出 `select`，不会跳出外层 `for` |

**伪随机**这条规则有一个重要推论：**不要依赖 `case` 的书写顺序。** 想要**取消优先于处理**，必须写两段 `select`：

```go
for {
    select {
    case <-done: // 第一段只查取消
        return
    default:
    }
    select { // 第二段正常等待
    case <-done:
        return
    case v := <-ch:
        handle(v)
    }
}
```

### 3. 六种写法对照

| 写法 | 效果 | 典型用途 |
| --- | --- | --- |
| `select { case v := <-ch: }` | 阻塞，直到 ch 可接收 | 等结果 |
| `select { case ch <- v: }` | 阻塞，直到 ch 可发送 | 投递结果 |
| `select { case v := <-ch: default: }` | 非阻塞尝试，拿不到就走 | **有就取**、丢旧值 |
| `select { case <-time.After(d): }` | 等一个**定时事件** | 超时 |
| `select { case <-ctx.Done(): }` | 等一个**已取消**事件 | 取消、优雅关闭 |
| `for { select { ... } }` | 事件循环 | worker 主体 |

### 4. 超时：`time.After`、`time.NewTimer`、`time.Ticker`

**为什么需要超时。** 网络对端可能永远不回应，数据库可能锁住，第三方接口可能挂起。没有超时，这次等待就会永久占用一个 goroutine、一个连接、一段内存；调用链上游也会跟着挂起，最后整条链路都被拖死——这是生产环境最常见的故障形态之一。

```go
select {
case v := <-results:
    use(v)
case <-time.After(500 * time.Millisecond):
    return errors.New("等待超时")
}
```

| 工具 | 语义 | 特点 |
| --- | --- | --- |
| `time.After(d)` | 返回一个 `<-chan Time`，d 后收到一个值 | 写法最省事；每次调用都会创建一个定时器 |
| `time.NewTimer(d)` | 返回 `*Timer`，可 `Stop()` / `Reset()` | 需要提前结束或复用定时器时用 |
| `time.NewTicker(d)` | 每隔 d 发一次，可 `Stop()` | 周期任务，如心跳、定时刷新 |
| `context.WithTimeout` | 把超时沿调用链传下去 | 真正解决**让上游停下来**的问题 |

两个必须说清的陷阱：

1. **`time.After` 不适合放在热循环里**（例如每秒几十万次的分支判断）。每次都会创建一个定时器对象；Go 1.23 起未被引用的定时器可以被 GC 回收，压力比过去小，但**为一次判断造一个定时器**本身仍是浪费。这种位置用 `time.NewTimer` + `Reset` 或直接把超时提出去。
2. **超时只结束**这次等待**，不会终止对方的工作。** 上面那段代码在超时后直接返回，但产生 `results` 的 goroutine 仍然在跑。想要它停下来，必须把取消信号传进去——这就是下一小节的事。

### 5. 取消：`done` channel 与 `context.Context`

Go **没有** **杀死一个 goroutine** 的 API，这是刻意的设计：goroutine 可能正持有锁、正在写共享状态、正在写文件，被强行中断会留下不一致的数据；而且它的栈由 runtime 管理，外部无法安全地把它撕掉。所以 Go 的取消是**协作式**的：

> 取消 = 通知 + 对方自己检查并退出。

最小实现就是一个 `done` channel：

```go
func worker(done <-chan struct{}, jobs <-chan int, results chan<- int) {
    for {
        select {
        case <-done:
            return
        case job, ok := <-jobs:
            if !ok {
                return
            }
            // 发送结果时也要监听取消
            select {
            case results <- job * job:
            case <-done:
                return
            }
        }
    }
}
```

注意内层那个 `select`：**发送也要能被取消**。如果没有它，当没有消费者时，worker 会永久阻塞在 `results <-` 上，`done` 也救不了它——这是最常见的泄漏形态。

手写 `done` 能表达**停**，但表达不了**为什么停**、**什么时候停**：

| 需求 | 手写 `done` channel | `context.Context` |
| --- | --- | --- |
| 通知停止 | 要，`close(done)` | `ctx.Done()` |
| 超时自动停止 | 自己配 `time.After` | `context.WithTimeout` |
| 上游传下来、往下传下去 | 每个函数都要多加一个参数 | 标准做法，一路传 `ctx` |
| 知道是超时还是主动取消 | 自己记 | `ctx.Err()` 区分 |
| 与标准库打通（HTTP、数据库） | 不行 | 绝大多数库都接受 `ctx` |

`context` 家族全表：

| API | 作用 |
| --- | --- |
| `context.Background()` | 根 context，main 里用 |
| `context.TODO()` | 暂时不知道该传什么时的占位 |
| `WithCancel(parent)` | 得到 `ctx` 和 `cancel`，手动取消 |
| `WithTimeout(parent, d)` | d 之后自动取消 |
| `WithDeadline(parent, t)` | 到时间点自动取消 |
| `WithValue(parent, k, v)` | 携带请求域的值（trace id、用户标识） |
| `ctx.Done()` | 返回一个 channel，取消后关闭 |
| `ctx.Err()` | `context.Canceled` 或 `context.DeadlineExceeded` |
| `ctx.Deadline()` | 是否设了截止时间、什么时候到 |

```go
func fetch(ctx context.Context) error {
    req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
    if err != nil {
        return err
    }
    _, err = http.DefaultClient.Do(req)
    return err
}

ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
defer cancel() // 必须调用，否则计时器与子 context 无法释放
if err := fetch(ctx); err != nil {
    if errors.Is(err, context.DeadlineExceeded) {
        // 超时，不是别的错
    }
}
```

使用规范，每条都有具体的失败后果：

| 规范 | 不遵守会怎样 |
| --- | --- |
| `ctx` 作为函数第一个参数 | 无法沿调用链取消 |
| 不把 `ctx` 存进结构体 | 生命周期错配，超时无法传递 |
| 不传 `nil` | panic 或永远不取消 |
| `cancel()` 必须调用（通常 `defer cancel()`） | 计时器与子 context 泄漏，直到父级结束 |
| `WithValue` 只放请求域数据，key 用自定义类型 | 用字符串 key 会和其他包撞车 |
| `ctx.Done()` 只在 `select` 里等，不要反复查询 | 查询式检查会漏掉阻塞中的操作 |
| 区分 `Canceled` 和 `DeadlineExceeded` | 把**上游主动取消**当成错误上报，日志里全是噪声 |

### 6. 循环骨架：结果或取消

把第二部分和第三部分拼起来，就是一个几乎可以套用到所有并发任务的骨架：

```go
func run(ctx context.Context, jobs []int) []int {
    results := make(chan int)
    go func() {
        defer close(results) // 生产者负责关闭
        for _, job := range jobs {
            select {
            case results <- job * job:
            case <-ctx.Done(): // 取消时停止生产
                return
            }
        }
    }()

    var out []int
    for {
        select {
        case <-ctx.Done():
            return out // 返回已经拿到的部分结果
        case v, ok := <-results:
            if !ok {
                return out // 正常结束
            }
            out = append(out, v)
        }
    }
}
```

这个骨架同时具备四个性质：生产者会关闭 channel、消费者不会泄漏、取消能双向传播、能返回部分结果。后面第七部分的所有功能，都是在这个骨架上加东西。

---

## 四、sync.WaitGroup：等待一组工作结束

### 1. 为什么需要 WaitGroup

**等一组 goroutine 全部结束**这个需求，手动实现有三种做法，各有各的毛病：

| 手写方案 | 问题 |
| --- | --- |
| `time.Sleep` | 不知道要睡多久，只是赌 |
| 每个 worker 完成后往一个 channel 发一个值，主 goroutine 收 n 个 | 可行，但要预先知道 n；收完之后如果还有人不退出也无法察觉 |
| `atomic` 计数 + 忙等 | 烧 CPU；等待逻辑自己写容易错 |

`WaitGroup` 就是第三种做法的正确实现：一个**计数器**加一个**等待队列**。它恰好描述了我们能说清的那部分信息——**我不知道谁先结束，但我知道一共有几个**。

### 2. 三个方法（外加一个新方法）

| 方法 | 作用 |
| --- | --- |
| `wg.Add(delta int)` | 计数加 delta，通常 `Add(1)` 或 `Add(n)` |
| `wg.Done()` | 等价于 `Add(-1)`，表示**这一个完成了** |
| `wg.Wait()` | 阻塞，直到计数归零 |
| `wg.Go(f func())` | Go 1.25 新增：等价于 `Add(1)` + `go` + 结束时自动 `Done` |

```go
func main() {
    var wg sync.WaitGroup
    for i := 0; i < 3; i++ {
        wg.Add(1)
        go func(id int) {
            defer wg.Done() // 放在第一行，任何 return 路径都不会漏
            fmt.Println("worker", id, "done")
        }(i)
    }
    wg.Wait()
    fmt.Println("all workers finished")
}
```

状态机很简单，但它解释了所有规则：

```text
        Add(n)                      Done() 使计数归零
  0 ─────────────►  n  ─────────────────────────────►  Wait() 返回
  │                 │
  │ Wait() 立刻返回  │ Wait() 在这里阻塞
  │                 │
  └─────────────────┘
  计数不能为负：Done 多于 Add 会 panic
```

### 3. 五条硬规则

| 规则 | 违反的后果 |
| --- | --- |
| `Add` 要在启动 goroutine **之前**调用 | `Wait` 可能在计数还是 0 时提前返回，程序**没等就结束了** |
| `Add(1)` 与 `Done()` 必须一一对应 | 少一次 `Done` 就永久阻塞；多一次 `Done` 会 `panic: sync: negative WaitGroup counter` |
| 用 `defer wg.Done()` 而不是在末尾手动写 | 提前 `return` 或 panic 的路径上，`Done` 会被漏掉 |
| 不要复制 `WaitGroup`（用 `*sync.WaitGroup` 传递） | 两个计数器各自计数，`Wait` 永远等不到；`go vet` 会以 `copylocks` 报警 |
| 零值即可用，不需要初始化 | —— |

`Add` 的时机值得单独强调。下面这段是经典错误：

```go
for i := 0; i < 3; i++ {
    go func() {
        wg.Add(1) // 错：可能还没执行到，Wait 已经返回了
        defer wg.Done()
        work()
    }()
}
wg.Wait()
```

**`Add` 必须在 `Wait` 之前**这句话的精确版本是：**计数从 0 变成正数这件事，必须发生在 `Wait` 可能看到 0 之前**。所以习惯上把 `Add` 放在 `go` 之前，由启动者负责计数，goroutine 只负责 `Done`。

Go 1.25 的 `wg.Go` 把这三件事合成一件，从根上消灭了**忘写 `Done`**和**`Add` 放错位置**两个错误：

```go
var wg sync.WaitGroup
for _, job := range jobs {
    wg.Go(func() { // 自动 Add(1)，函数返回后自动 Done
        handle(job)
    })
}
wg.Wait()
```

> 用它之前先确认工具链版本（`go version`）。在旧版本上它不存在，本课程的练习仍以 `Add`/`Done`/`Wait` 为主——这三个是必须掌握的，`Go` 只是更省事的写法。

### 4. 特点与边界

`WaitGroup` 只做一件事，这也决定了它的边界：

| 它负责 | 它不负责 |
| --- | --- |
| 等待一组 goroutine 结束 | 传递结果（用 channel） |
| 计数 | 取消任务（用 `context`） |
| —— | 保护共享内存（用锁） |
| —— | 等待**第一个完成**（用 channel） |

还有一点容易被忽略：**`Wait` 等的是最慢的那个**。一个 worker 卡住，整组都等不到齐。所以真实的批量任务通常写成**等，但最多等 3 秒**——这就需要把 `Wait` 和一个 channel 组合起来：

```go
func waitTimeout(wg *sync.WaitGroup, d time.Duration) bool {
    done := make(chan struct{})
    go func() {
        wg.Wait()
        close(done)
    }()
    select {
    case <-done:
        return true // 全部完成
    case <-time.After(d):
        return false // 超时，调用方需要决定是放弃还是继续等
    }
}
```

### 5. 组合用法

**（1）多个生产者汇聚到一个出口**：第四部分开头的 `merge`。

**（2）按下标写结果，不加锁**。`append` 是读-改-写，多个 goroutine 并发 `append` 同一个切片是数据竞争；预分配 + 按下标写就完全不用锁：

```go
func fanOut(jobs []int, workers int) []int {
    partial := make([]int, workers) // 每个 worker 一个固定槽位
    var wg sync.WaitGroup
    wg.Add(workers)
    for i := 0; i < workers; i++ {
        go func(idx int) {
            defer wg.Done()
            sum := 0
            for _, job := range jobs {
                sum += job
            }
            partial[idx] = sum // 各写各的槽位，互不干扰
        }(i)
    }
    wg.Wait()
    return partial
}
```

这个手法在真实项目里非常常用：只要做到**每个 goroutine 只写属于自己的那份**，就不需要任何锁。反过来，如果连下标都可能撞车（例如结果条数不定），就得回到 channel 或 `Mutex`。

**（3）与 `context` 组合**：`WaitGroup` 只告诉你**完了**，`context` 告诉你**该停了**，两者一起用构成第七部分的批量任务模式。

---

## 五、锁：保护共享内存

### 1. 为什么需要锁

`channel` 传递的是**值**，但有些状态天生就是**共享的**：一个服务里的配置、一个缓存 map、一个连接池、一个累计计数器。这些数据不属于某一个 goroutine，它们被多方读写，此时用 channel 传递是别扭的。

没有锁，共享内存会以三种方式出错：

| 现象 | 例子 | 严重程度 |
| --- | --- | --- |
| 更新丢失 | 两个 goroutine `n++`，结果少了几千 | 静默错误，最危险 |
| 结构损坏 | 两个 goroutine 同时写一个 map | `fatal error: concurrent map writes`，进程直接退出 |
| 读到中间状态 | 一个 goroutine 更新 `{账号, 余额}` 两个字段，另一个读到一半新的 | 业务逻辑错乱，极难复现 |

锁提供了两件事：

1. **互斥**：同一时刻只有一个 goroutine 在临界区里。
2. **顺序**：`Unlock()` 之后的 `Lock()` 必须等到前者完成，于是建立了一条 happens-before 边——临界区里的普通读写因此变得可见、有序。

第二点比第一点更本质。常见的误解是**锁只是防止同时写**，其实**读也必须在同一把锁的保护范围内**：读不到最新值，前面那些写就等于没发生。

### 2. `sync.Mutex` 的全部能力

```go
type Counter struct {
    mu    sync.Mutex
    value int
}

func (c *Counter) Add(delta int) {
    c.mu.Lock()
    defer c.mu.Unlock() // 保证任何路径都会解锁
    c.value += delta
}

func (c *Counter) Value() int {
    c.mu.Lock()
    defer c.mu.Unlock()
    return c.value // 读也要锁
}
```

| 方法 | 作用 | 备注 |
| --- | --- | --- |
| `Lock()` | 加锁，已有人持锁则阻塞 | |
| `Unlock()` | 解锁 | 对未加锁的 Mutex 解锁会 panic |
| `TryLock()` | 尝试加锁，成功返回 `true` | Go 1.18+；**不阻塞**，适合**拿不到就走别的路** |

特性与坑，逐条对应到具体现象：

| 特性 | 说明 | 不理解的后果 |
| --- | --- | --- |
| 零值可用 | `var mu sync.Mutex` 就能用，不需要 `New` | 多余地写 `sync.NewMutex()` |
| 不可重入 | 同一个 goroutine 重复 `Lock` 会自己把自己锁死 | 程序在 `sync.Mutex.Lock` 上卡住，runtime 报 `fatal error: all goroutines are asleep - deadlock!` |
| 不绑定 goroutine | A goroutine 加锁、B 解锁在语法上允许 | 误以为**锁属于某个 goroutine**，写出混乱的解锁逻辑 |
| 不可复制 | 锁内部有状态，复制会得到两把独立的锁 | `go vet` 报 `copylocks`；两个**同一把锁**实际不互斥 |
| 解锁未持有的锁 | panic: `sync: unlock of unlocked mutex` | —— |

不可重入这一条，新手最常撞上：

```go
func (c *Counter) AddTwice(delta int) {
    c.mu.Lock()
    defer c.mu.Unlock()
    c.Add(delta) // Add 里又 Lock 同一把锁 → 永久阻塞
    c.Add(delta)
}
```

解决方法不是换一把锁，而是**把需要锁的操作和不需要锁的内部逻辑分开**：内部函数不加锁，只在公开方法的边界上加锁。

`TryLock` 也不是**更安全的 Lock**，它的语义是**拿不到就算了**：

```go
if mu.TryLock() { // 拿到就更新缓存，拿不到说明有人正在更新，跳过即可
    defer mu.Unlock()
    cache[key] = value
}
```

如果代码变成 `for !mu.TryLock() {}`，那只是把你的手写轮询换了张皮。

### 3. 临界区怎么划

临界区的定义很简单：**从 `Lock` 到 `Unlock` 之间的代码**。难点在于它该有多宽。

| 临界区 | 后果 |
| --- | --- |
| 过小：只锁了写、没锁读 | 读到旧值，等于没同步 |
| 过小：漏掉某个字段的更新 | 出现**半个新、半个旧**的状态 |
| 合适：刚好覆盖共享状态的读写 | 正确且并发度高 |
| 过大：把网络请求、文件 I/O、耗时计算包进来 | 所有 goroutine 一起排队，并发退化为串行 |

```go
// 反例：锁里做了网络请求
mu.Lock()
data, err := fetchFromRemote() // 可能要几百毫秒
cache[name] = data
mu.Unlock()

// 正确：先做慢操作，只锁写回那一小段
data, err := fetchFromRemote()
mu.Lock()
cache[name] = data
mu.Unlock()
```

判断方法：问自己**这段时间里，除了我在改的那几个字段，还发生了什么**。如果答案是**一段与共享状态无关的计算或 I/O**，就应该挪出临界区。

### 4. 死锁：四种形态

死锁在教科书上有四个必要条件，落到 Go 代码里其实就是四类写法：

| 形态 | 例子 | 特征 |
| --- | --- | --- |
| 同一把锁重入 | 上面 `AddTwice` 的例子 | runtime 直接报 deadlock |
| 两把锁交叉 | A 持 `mu1` 等 `mu2`，B 持 `mu2` 等 `mu1` | 需要统一锁顺序 |
| 锁与 channel 交叉 | 持锁等待 channel，而 channel 的发送方正等这把锁 | 最隐蔽 |
| 等一个永远不会发生的事件 | 无缓冲 channel 发送没人接收；`wg.Wait()` 少一次 `Done` | 表现为**程序静静地卡住** |

**关于 runtime 的报错要有一个清醒认识**：`fatal error: all goroutines are asleep - deadlock!` 只在**所有** goroutine 都阻塞时才会出现。只要还有一个人在跑（例如一个 HTTP 服务的其他连接还在处理请求），runtime 就不会报，程序会静默卡死。所以真正的排查手段是第七部分那个方法——**对每个阻塞点写出谁会唤醒它**：

```text
阻塞点：ch <- result
谁会唤醒它？→ 消费者执行 <-ch
消费者在哪里？→ 只在 for range results 里
它有没有可能提前退出？→ 有，ctx.Done() 时它 return 了
→ 结论：一旦取消，这里就再也不会有人接收 → 必须给发送加 select
```

### 5. `sync.RWMutex`：读多写少的优化

`Mutex` 把并发的读也串行化了。如果一份数据**读远多于写**（配置、路由表、本地缓存），可以让多个读者同时进入：

```go
type ConfigStore struct {
    mu   sync.RWMutex
    data map[string]string
}

func (s *ConfigStore) Get(key string) (string, bool) {
    s.mu.RLock()
    defer s.mu.RUnlock()
    value, ok := s.data[key]
    return value, ok
}

func (s *ConfigStore) Set(key, value string) {
    s.mu.Lock()
    defer s.mu.Unlock()
    s.data[key] = value
}
```

| 方法 | 语义 |
| --- | --- |
| `RLock()` / `RUnlock()` | 读锁，可被多个 goroutine 同时持有 |
| `Lock()` / `Unlock()` | 写锁，独占；有读者时等待 |
| `TryLock()` / `TryRLock()` | 非阻塞版本（Go 1.18+） |

规则与代价：

| 要点 | 说明 |
| --- | --- |
| 读锁之间共享，读与写、写与写互斥 | 这是它唯一的收益来源 |
| 写者优先 | 有写者等待时，新的 `RLock` 会排队，避免写者被无休止的读者饿死 |
| **读锁不可递归** | 一个 goroutine 持 `RLock` 时再次 `RLock`，如果中间有写者插队，就是死锁 |
| 不要默认使用 | 没有读写差异、或写也很频繁时，它比 `Mutex` 更慢（多了一层读者计数） |
| 收益要用数据说话 | 先写基准测试（`go test -bench`）比较，再决定换不换 |
| 同样不可复制 | 传递时用指针 |

最后一条经验：**`RWMutex` 属于优化，不属于设计**。先用 `Mutex` 写出正确的代码，确认读确实是瓶颈后再换。

### 6. 锁还是 channel

这两者不是竞争关系，它们解决不同的问题。判据只有一问：

> **我是在传递一份数据或任务的所有权，还是在保护一份共同拥有的内存？**

| 场景 | 选择 | 理由 |
| --- | --- | --- |
| 把任务分给 worker | channel | 所有权在移动 |
| 把结果交回调用方 | channel | 同上 |
| 通知一组 goroutine 停止 | `close(done)` / `context` | 广播 |
| 等待一组任务结束 | `WaitGroup` | 与数据无关 |
| 保护计数器、缓存、配置 | `Mutex` / `RWMutex` | 内存被共同拥有 |
| 单个标志位、单变量计数 | `atomic` | 比锁轻 |

一句话：**`channel` 管流，锁管存。** 用它俩互相替代，通常意味着对数据的所有权还没想清楚。

---

## 六、`sync` 包的其余工具（进阶）

前面四个工具（`WaitGroup`、`Mutex`、`RWMutex`、`channel`）是本节课必须掌握的主体。`sync` 包里还有几件工具，它们各自解决一个很具体的问题。这一节的目标是**知道它们存在、知道它们解决什么问题**，遇到对应场景时能想起来。

### 1. `sync.Once`：只执行一次的初始化

**问题**：某个资源（数据库连接、全局配置、昂贵的索引）只需初始化一次，但要保证并发安全。手写**检查—加锁—再检查**很容易漏：先判断标志位再初始化，两个 goroutine 可能同时通过判断，初始化两次。

```go
var (
    instance *DB
    once     sync.Once
)

func GetDB() *DB {
    once.Do(func() {
        instance = connect() // 并发调用下只会执行一次
    })
    return instance
}
```

| 要点 | 说明 |
| --- | --- |
| 只有一个方法 | `Do(f func())`，没有 `Undo`、不能重置 |
| 阻塞语义 | `f` 执行期间，其他 `Do` 调用会阻塞，直到 `f` 返回；所有调用者返回后一定能看到 `f` 的结果 |
| `f` 里 panic | 这次调用仍被记为**已执行**；后续 `Do` 不会重试（Go 1.21 起 `sync.OnceFunc` 会让每次调用都重新 panic） |
| Go 1.21+ 的语法糖 | `OnceFunc`、`OnceValue`、`OnceValues`，省掉 `var instance` 这个全局变量 |

```go
getConfig := sync.OnceValue(func() *Config {
    return loadConfig() // 只执行一次，之后直接返回缓存结果
})
cfg := getConfig()
```

`Once` + `Mutex` 是常见组合：`Once` 负责**只创建一次**，`Mutex` 负责**创建之后，对它的并发读写**。

### 2. `sync.Map`：并发安全的键值表

**问题**：普通 map 并发写会让进程直接退出，而用 `Mutex` 包一层在**读多写少、key 集合稳定**的场景下锁竞争明显。`sync.Map` 内部用两份表（read / dirty）把这个场景优化了。

| 方法 | 作用 |
| --- | --- |
| `Store(k, v)` | 写入 |
| `Load(k)` | 读取，返回 `(any, bool)` |
| `LoadOrStore(k, v)` | 有则返回已有值，无则存入 |
| `LoadAndDelete(k)` | 取出并删除，常用于**一次性消费** |
| `Delete(k)` | 删除 |
| `Range(f func(k, v any) bool)` | 遍历，`f` 返回 `false` 时停止 |
| `CompareAndSwap(k, old, new)` | 比较并替换（Go 1.20+） |
| `CompareAndDelete(k, old)` | 比较并删除（Go 1.20+） |
| `Clear()` | 清空（Go 1.23+） |

代价与限制，这几条决定了它不能当默认选择：

| 限制 | 影响 |
| --- | --- |
| 键值类型是 `any` | 取值要类型断言，失去编译期检查 |
| 没有长度、不能原子地**读-改-写**多个键 | 复合操作仍要额外的锁 |
| `Range` 不保证一致性快照 | 遍历期间可能看到也可能看不到并发写入 |
| 写多时性能并不比 `Mutex + map` 好 | 用错场景反而更慢 |

选择标准很简单：**key 集合基本固定、读远多于写**（例如**已注册的连接**、**已加载的模块**）→ `sync.Map`；其他情况 → `map` + `Mutex`/`RWMutex`。

### 3. `sync.Pool`：临时对象复用

**问题**：高频请求里反复创建/丢弃大对象（例如 4KB 的缓冲区），会把压力全压给 GC。

```go
var bufPool = sync.Pool{
    New: func() any { return make([]byte, 0, 4096) },
}

func handle() {
    buf := bufPool.Get().([]byte)
    defer bufPool.Put(buf[:0]) // 用完还回去
    buf = append(buf, "hello"...)
}
```

| 要点 | 说明 |
| --- | --- |
| `Get()` / `Put(x)` | 取与还；`Get` 在没有可用对象时调用 `New` |
| 不保证命中 | 池里的对象可能在任意一次 GC 时被清空 |
| 对象状态未知 | `Get` 拿到的可能是别人用过的，**必须重置**（`buf[:0]` 就是重置长度） |
| 不是缓存 | 不能用来缓存有业务含义的数据，因为它随时会消失 |
| 返回值是 `any` | 需要类型断言 |

一句话：`Pool` 是**减少分配**的工具，不是**保存数据**的工具。

### 4. `sync.Cond`：条件等待

**问题**：一个 goroutine 要等到某个条件成立才继续（任务队列非空、缓冲区可写）。没有 `Cond`，只能**循环 + `Sleep` 轮询**：要么延迟高，要么空转烧 CPU。

```go
type Queue struct {
    mu    sync.Mutex
    cond  *sync.Cond
    items []int
}

func NewQueue() *Queue {
    q := &Queue{}
    q.cond = sync.NewCond(&q.mu)
    return q
}

func (q *Queue) Push(v int) {
    q.mu.Lock()
    defer q.mu.Unlock()
    q.items = append(q.items, v)
    q.cond.Signal() // 唤醒一个等待者
}

func (q *Queue) Pop() int {
    q.mu.Lock()
    defer q.mu.Unlock()
    for len(q.items) == 0 { // 必须用循环，不能只 if 一次
        q.cond.Wait()
    }
    v := q.items[0]
    q.items = q.items[1:]
    return v
}
```

| 方法 | 语义 |
| --- | --- |
| `NewCond(l Locker)` | 绑定一把锁（通常是 `*sync.Mutex` 或 `*sync.RWMutex`） |
| `Wait()` | **先解锁并阻塞，被唤醒后重新加锁**再返回 |
| `Signal()` | 唤醒一个等待者 |
| `Broadcast()` | 唤醒所有等待者 |

三条必须记住的规则：

1. `Wait` 必须在持锁状态下调用，否则 panic。
2. `Wait` 返回不代表条件已成立（可能被无关的 `Signal` 唤醒），所以**必须放在 `for` 循环里重新检查条件**。
3. 修改条件后要持锁调用 `Signal`/`Broadcast`。

**为什么现在用得少了**：大多数**等条件**的场景已经被 channel 覆盖——等一个事件用 `<-ch`，广播用 `close(ch)`。`Cond` 仍有它的位置：需要反复等待、且等待者与唤醒者共享同一份可变状态时（例如自建连接池的空闲列表）。

### 5. `sync/atomic`：单变量的无锁操作

**问题**：一个计数器、一个开关标志，用 `Mutex` 太重——加锁本身有开销，还会让本来无冲突的操作排队。

```go
var (
    requests atomic.Int64
    ready    atomic.Bool
)

requests.Add(1)
ready.Store(true)
if ready.Load() {
    n := requests.Swap(0) // 取旧值并写入新值
    fmt.Println("reset from", n)
}
if requests.CompareAndSwap(0, 1) { // 只有当前值等于 0 时才写入 1
    // ...
}
```

| 操作族 | 方法 | 用途 |
| --- | --- | --- |
| 加法 | `Add` | 计数 |
| 读 | `Load` | 读取 |
| 写 | `Store` | 写入 |
| 交换 | `Swap` | 取值并重置 |
| 比较并交换 | `CompareAndSwap` | 实现**只在值没变时才更新** |

Go 1.19 起提供了类型化的原子类型：`atomic.Bool`、`atomic.Int32`、`atomic.Int64`、`atomic.Uint32`、`atomic.Uint64`、`atomic.Uintptr`、`atomic.Pointer[T]`、`atomic.Value`（可存任意类型的值，但同一个 `Value` 里必须始终存同一具体类型）。

```go
var cfg atomic.Pointer[Config] // 配置热更新：读多写少，读侧无锁

func current() *Config { return cfg.Load() }
func reload(c *Config) { cfg.Store(c) } // 整个指针换掉，读方永远看到完整的旧配置或新配置
```

`atomic` 的边界：

| 能做 | 不能做 |
| --- | --- |
| 单个变量的原子读写改 | 保证多个变量之间的一致（要锁） |
| 发布不可变对象（换指针） | 保护一个会被逐步修改的结构体 |
| 状态标志、计数、一次性开关 | 复合逻辑（**先查再改**需要 CAS 循环，容易写错） |

### 6. 小结表

| 工具 | 没有它会怎样 | 一句话记住 |
| --- | --- | --- |
| `WaitGroup` | 不知道一组任务是否都结束 | 计数等待 |
| `Mutex` | 共享内存被并发读写，结果错或进程崩 | 互斥 + 建立顺序 |
| `RWMutex` | 读多写少时读被串行化 | 多读单写 |
| `Once` | 初始化重复执行 | 只做一次 |
| `Map` | 并发 map 操作让进程退出 | 读多写少的内建并发表 |
| `Pool` | 高频分配压垮 GC | 复用临时对象 |
| `Cond` | 等条件只能轮询 | 条件等待（现在少用） |
| `atomic` | 单变量也要用锁，开销大 | 单变量无锁操作 |

---
## 七、组合：把工具拼起来能得到什么

前面六部分讲的是零件。真实的并发功能都是**组合**出来的，而且组合方式相当固定——八成的场景就是下面这几种。这一节对每个功能都给出四件事：**要实现什么**、**组合了哪些 API**、**代码**、**去掉某个零件会怎样**。

### 0. 组合矩阵

| 功能 | channel | select / context | WaitGroup | Mutex / RWMutex | Once / atomic / Pool |
| --- | --- | --- | --- | --- | --- |
| 流水线 | ✓ 必需 | ✓ 取消时必需 | 可选 | —— | —— |
| fan-out / fan-in | ✓ 必需 | 可选 | ✓ 必需（汇聚时） | —— | —— |
| 固定 worker 任务池 | ✓ 必需 | ✓ 必需 | ✓ 必需 | —— | —— |
| 批量任务 + 超时 | ✓ 必需 | ✓ 必需 | ✓ 必需 | —— | —— |
| 并发聚合（保序） | 可选 | 可选 | ✓ 必需 | —— | —— |
| 错误传播与快速失败 | ✓ 必需 | ✓ 必需 | ✓ 必需 | —— | —— |
| 并发限流 | ✓ 必需（当令牌） | 可选 | ✓ 需要 | —— | —— |
| 优雅关闭 | ✓ 广播 | ✓ 兜底 | ✓ 必需 | ✓ 必需（状态标志） | —— |
| 单例与热更新 | —— | —— | —— | ✓ 可选 | ✓ 必需 |
| 临时对象复用 | 可选 | —— | —— | —— | ✓ 必需 |

结论可以先说：**`channel` + `select`/`context` + `WaitGroup` 这三件套覆盖了绝大多数并发场景**，锁只在**有真正被共享的内存**时出场。

> **版本提示**：下面示例里出现的 `wg.Go(...)` 是 Go 1.25 新增的写法（等价于 `wg.Add(1)` + `go func(){ defer wg.Done(); ... }()`）。如果你的工具链更旧，用 `go version` 确认后改回三件套即可，逻辑完全一样。

### 1. 流水线（pipeline）

**要实现什么**：把一个大任务拆成若干级，每一级独立 goroutine 处理，数据像流水线一样流过；上游处理完自动告诉下游**没有更多了**。

**组合**：`channel`（级与级之间的传送带）+ `close`（结束信号）+ `for range`（分级循环）。

```text
 generate ──ch1──►  square ──ch2──►  filter ──ch3──►  消费
    │                 │                 │
 close(ch1)        close(ch2)        close(ch3)
            上游关闭 → 本级的 range 结束 → 本级关闭自己的出口
```

```go
func generate(nums ...int) <-chan int {
    out := make(chan int)
    go func() {
        defer close(out) // 发送方负责关闭
        for _, n := range nums {
            out <- n
        }
    }()
    return out
}

func square(in <-chan int) <-chan int {
    out := make(chan int)
    go func() {
        defer close(out)
        for n := range in { // 上游关闭时自动退出
            out <- n * n
        }
    }()
    return out
}

func main() {
    for v := range square(generate(1, 2, 3, 4, 5)) {
        fmt.Println(v) // 1 4 9 16 25
    }
}
```

| 去掉某个零件 | 后果 |
| --- | --- |
| 没有 `defer close(out)` | 下游 `range` 永远不结束，`main` 卡在最后一行 |
| 由下游 `close` | 上游发送时 panic |
| 只有一级就 `close`，漏了中间级 | 中间级后的所有消费者一起卡住 |
| 下游提前退出（例如只想取前 3 个） | 上游卡在发送上 → goroutine 泄漏，需要把 `ctx` 接进每一级的收发 |

流水线最容易被忽略的一点是：**它的价值不在快，而在可组合**。每一级都是**输入 channel → 输出 channel**的纯函数形状，所以能像搭积木一样加减级数，也能把某一级换成 fan-out。

### 2. fan-out / fan-in（分发与汇聚）

**要实现什么**：同一个输入被多个 worker 并行消费（fan-out，天然负载均衡：谁先空下来谁先拿），多个输出的结果再汇总成一个（fan-in）。

```text
                        ┌──► worker 1 ──┐
  生成器 ──jobs────────►├──► worker 2 ──┼──results──► 汇总
                        └──► worker 3 ──┘
                         fan-out        fan-in
```

**组合**：`channel`（分发 + 汇聚）+ `WaitGroup`（判断**所有生产者都结束了**）+ 单点 `close`（汇聚出口）。

```go
func merge(inputs ...<-chan int) <-chan int {
    out := make(chan int)
    var wg sync.WaitGroup
    wg.Add(len(inputs))
    for _, in := range inputs {
        go func(in <-chan int) {
            defer wg.Done()
            for v := range in {
                out <- v
            }
        }(in)
    }
    go func() {
        wg.Wait()  // 等所有输入都排空
        close(out) // 唯一一次 close
    }()
    return out
}
```

| 去掉某个零件 | 后果 |
| --- | --- |
| 没有 `WaitGroup` | 不知道什么时候所有输入都排空了，`close(out)` 无从下手 |
| 在每个生产者里 `close(out)` | 第二个生产者也 close 时 panic |
| 在 `merge` 里直接 `wg.Wait()` | 函数阻塞到全部处理完才返回，失去了**边跑边消费**的意义 |

**顺序会变**：fan-out 之后结果的到达顺序不再等于输入顺序。如果要保序，见第 5 小节。

### 3. 固定 worker 数量的任务池

**要实现什么**：限制**同一时刻有多少个任务在跑**。不限制时，一万个请求就起一万个 goroutine，下游数据库连接池会被打爆；限制后，goroutine 数量固定，任务排队等待。

**组合**：`jobs` channel（任务队列）+ `results` channel（结果回收）+ `WaitGroup`（等 worker 全退出）+ `context`（取消）。

```go
type Result struct {
    JobID int
    Value int
}

func worker(ctx context.Context, jobs <-chan int, results chan<- Result, wg *sync.WaitGroup) {
    defer wg.Done()
    for {
        select {
        case <-ctx.Done():  // 取消优先
            return
        case job, ok := <-jobs:
            if !ok {        // 任务队列关闭且取完
                return
            }
            r := Result{JobID: job, Value: job * job}
            select {
            case results <- r: // 发送也要能被取消
            case <-ctx.Done():
                return
            }
        }
    }
}

func run(ctx context.Context, jobs []int, workerCount int) []Result {
    if workerCount <= 0 {
        return nil
    }
    jobCh := make(chan int)
    resultCh := make(chan Result)

    var wg sync.WaitGroup
    wg.Add(workerCount) // 在启动之前 Add
    for i := 0; i < workerCount; i++ {
        go worker(ctx, jobCh, resultCh, &wg)
    }

    go func() { // 投递者：负责关闭 jobCh
        defer close(jobCh)
        for _, job := range jobs {
            select {
            case jobCh <- job:
            case <-ctx.Done():
                return
            }
        }
    }()

    go func() { // 协调者：所有 worker 退出后关闭 resultCh
        wg.Wait()
        close(resultCh)
    }()

    var results []Result
    for r := range resultCh { // 唯一的消费者：消费到 resultCh 关闭
        results = append(results, r)
    }
    return results
}
```

这个结构里有三个关闭点，各自的责任必须分清，这张表是任务池的核心：

| 关闭点 | 由谁执行 | 为什么是它 | 写错的后果 |
| --- | --- | --- | --- |
| `close(jobCh)` | 投递任务的 goroutine | 只有它知道**没有更多任务了** | 由 worker 关闭：互相 close → panic |
| `close(resultCh)` | `wg.Wait()` 之后的协调 goroutine | 只有它知道**所有 worker 都停了** | 由调用方关闭：worker 还在发送 → panic |
| `ctx` 的 `cancel()` | 调用方（`defer cancel()`） | 生命周期由调用方拥有 | 忘记调用：计时器与子 context 泄漏 |

| 去掉某个零件 | 后果 |
| --- | --- |
| 去掉 `ctx.Done()` 分支 | 取消后 worker 仍会继续取任务；调用方退出后它们阻塞在 `results <-` 上 |
| 去掉 `WaitGroup` | 无法知道 worker 何时全部退出，`resultCh` 无法安全关闭 |
| 不用 channel 而让每个任务 `go` 一下 | goroutine 数量随任务量线性增长，失去限流意义 |

**一个必须接受的代价**：结果顺序不确定。谁先算完谁先发。需要顺序就配合第 5 小节的按下标写入。

### 4. 带超时与取消的批量任务

**要实现什么**：一次并发做 N 件事，整体最多等 3 秒；到点就返回已经拿到的部分结果，不让调用方无限期挂起。

**组合**：`context.WithTimeout`（谁该停）+ 任务内检查 `ctx`（协作式退出）+ `WaitGroup` + `channel`（结果）。

```go
func gather(ctx context.Context, jobs []int, workers int) []Result {
    ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
    defer cancel() // 必须 defer，否则超时计时器无法释放

    results := run(ctx, jobs, workers) // 复用上一小节的任务池
    if err := ctx.Err(); err != nil {
        fmt.Printf("批量任务未全部完成：%v，已收集 %d 条\n", err, len(results))
    }
    return results
}
```

| 关键点 | 说明 |
| --- | --- |
| 超时是**协作式**的 | `ctx` 只负责广播，任务的退出必须自己在 `select` 里检查 |
| 拿到的结果数量不确定 | 这是特性不是 bug：部分结果 + 明确的错误信息，比**什么都没有**有用 |
| `cancel()` 一定要调用 | 即使已经超时，也释放资源 |
| 区分 `Canceled` 与 `DeadlineExceeded` | 前者是上游主动取消，后者是超时，日志里要能看出区别 |

### 5. 并发聚合（需要保序时）

**要实现什么**：并发处理一个切片，结果**按输入顺序**放进结果切片，且不需要任何锁。

**组合**：`WaitGroup`（等待）+ 预分配切片 + **按下标写入**。

```go
func mapAll(inputs []int, workers int) ([]int, error) {
    out := make([]int, len(inputs)) // 预先分配，长度固定
    errs := make([]error, len(inputs))
    sem := make(chan struct{}, workers) // 限制并发度
    var wg sync.WaitGroup

    for i, in := range inputs {
        wg.Add(1)
        go func(idx, val int) {
            defer wg.Done()
            sem <- struct{}{}
            defer func() { <-sem }()
            out[idx], errs[idx] = compute(val) // 各写各的槽位
        }(i, in)
    }
    wg.Wait()

    for _, err := range errs {
        if err != nil {
            return nil, err
        }
    }
    return out, nil
}
```

| 写法 | 是否安全 | 说明 |
| --- | --- | --- |
| 多个 goroutine `append(out, v)` | 不安全 | `append` 是读-改-写，并发调用会丢元素甚至崩 |
| 按下标 `out[idx] = v` | 安全 | 每个 goroutine 只写自己那一个位置，无需同步 |
| 用 `Mutex` 保护 `append` | 安全但慢 | 所有结果汇聚到一把锁上 |
| 通过 channel 回收 | 安全，但不保序 | 需要额外记录下标来恢复顺序 |

这三种手法（按下标写、加锁、走 channel）在真实项目里都会出现，判断标准就是：**能不能让每个 goroutine 只碰自己的那一份**。能，就不用锁。

### 6. 错误传播与快速失败

**要实现什么**：一批并发任务里任何一个失败，立刻取消其余的，并把第一个错误返回给调用方。

**组合**：`context.WithCancel`（一个失败就广播停止）+ 容量为 1 的 `error` channel（只记第一个错误）+ `select` + `default`（有人记过错误就不必阻塞）+ `WaitGroup`（等大家收摊）。

```go
func runAll(ctx context.Context, jobs []int) error {
    ctx, cancel := context.WithCancel(ctx)
    defer cancel()

    errCh := make(chan error, 1) // 容量 1：只保存第一个错误
    var wg sync.WaitGroup
    for _, job := range jobs {
        wg.Go(func() {
            if err := handle(ctx, job); err != nil {
                select {
                case errCh <- err: // 我是第一个出错的
                    cancel()       // 通知其余任务停下
                default:           // 已经有人报过错，直接放弃上报
                }
            }
        })
    }
    wg.Wait() // 等所有任务真正退出，避免 goroutine 泄漏

    select {
    case err := <-errCh:
        return err
    default:
        return nil
    }
}
```

| 写法 | 后果 |
| --- | --- |
| 只收集错误、不 `cancel()` | 一个任务失败后其余的还在跑，浪费资源、延迟返回 |
| `cancel()` 但不等 `wg.Wait()` | 函数返回后还有 goroutine 活着，泄漏 |
| `errCh` 不带缓冲、也不用 `select`/`default` | 第二个出错的任务阻塞在发送上，`wg.Wait()` 永远等不完 |
| 用普通变量 `firstErr = err` 记录 | 数据竞争，`-race` 会直接报出来 |

`golang.org/x/sync/errgroup` 把这套模式封装好了，等价写法是：

```go
g, ctx := errgroup.WithContext(ctx)
for _, job := range jobs {
    g.Go(func() error { return handle(ctx, job) })
}
err := g.Wait() // 返回第一个错误，并保证所有 goroutine 已退出
```

理解上面的手写版本之后，`errgroup` 就只是省事的语法糖——**它没有引入任何新概念，只是把错误 channel + cancel + WaitGroup 打包**。

### 7. 并发限流（信号量）

**要实现什么**：不管有多少任务，同一时刻最多只有 K 个在跑。这是保护下游（数据库、第三方接口）最直接的手段。

**组合**：容量为 K 的 `channel` 当令牌桶 + `WaitGroup`（等待）。可选地加上 `context`。

```go
func limited(jobs []int, limit int) []int {
    sem := make(chan struct{}, limit) // 最多放 limit 个运行中的标记
    out := make([]int, len(jobs))
    var wg sync.WaitGroup

    for i, job := range jobs {
        wg.Go(func() {
            sem <- struct{}{}        // 取令牌：满了就在这里等
            defer func() { <-sem }() // 必须 defer，否则 panic 路径会漏还
            out[i] = compute(job)
        })
    }
    wg.Wait()
    return out
}
```

| 方案 | 并发上限 | 缺点 |
| --- | --- | --- |
| 每个任务一个 goroutine | 无限 | 内存、文件描述符、下游连接被打爆 |
| 固定 worker 任务池 | worker 数 | 需要多写一层队列 |
| 信号量 | limit | 每个任务仍然是一个 goroutine，只是被挡住；任务量极大时仍要小心 |
| 令牌桶 + `context` | limit | 等令牌时也要能被取消，`select` 要写在取令牌那一步 |

信号量适合**任务数不算巨大、但每个任务要独占一个外部资源**的场景；任务池适合**任务源源不断**的场景。

### 8. 优雅关闭

**要实现什么**：收到停止信号后，第一，不再接受新任务；第二，通知在途任务尽快收尾；第三，等它们真正结束；第四，如果超过给定时间还没结束，就报错退出而不是无限等。

**组合**：`Mutex`（保护**是否已关闭**这个共享标志）+ `close(done)`（广播收尾）+ `WaitGroup`（在途任务计数）+ `context`（兜底超时）。

```text
   正常运行
      │  收到停止信号
      ▼
  ① 拒绝新任务         closed = true（必须在锁里）
      │
      ▼
  ② 通知在途任务收尾    close(done)  ── 广播，一次唤醒全部
      │
      ▼
  ③ 等在途任务结束      wg.Wait()    ── 不在持锁状态下等
      │
      ▼
  ④ 收尾也不能无限等    select { <-finished / <-ctx.Done() }
      │
      ▼
    退出（正常 / 报超时）
```

```go
type Service struct {
    mu     sync.Mutex
    closed bool
    wg     sync.WaitGroup
    done   chan struct{}
}

func (s *Service) Handle(ctx context.Context, req int) error {
    s.mu.Lock()
    if s.closed {
        s.mu.Unlock()
        return errors.New("service is closing") // ① 拒绝新任务
    }
    s.wg.Add(1) // 与 closed 检查在同一段临界区里，否则会有竞态
    s.mu.Unlock()

    defer s.wg.Done()
    select {
    case <-ctx.Done():
        return ctx.Err()
    case <-s.done: // ② 收到停止信号，尽快收尾
    case <-time.After(20 * time.Millisecond): // 正常处理
    }
    return nil
}

func (s *Service) Shutdown(ctx context.Context) error {
    s.mu.Lock()
    s.closed = true
    close(s.done) // 广播
    s.mu.Unlock()

    finished := make(chan struct{})
    go func() {
        s.wg.Wait() // ③ 不能持锁等，否则拒绝新任务和等待互相锁死
        close(finished)
    }()

    select {
    case <-finished:
        return nil
    case <-ctx.Done(): // ④ 收尾超时
        return ctx.Err()
    }
}
```

三个易错点：

| 易错点 | 后果 |
| --- | --- |
| `closed` 检查与 `wg.Add(1)` 不在一段临界区里 | 关闭流程可能已经 `Wait` 完，之后才 `Add`，出现**没人等的在途任务** |
| 持锁调用 `wg.Wait()` | 在途任务若需要这把锁 → 死锁；同时新任务还进不来 |
| 只 `close(done)` 不 `Wait` | 函数返回时还有任务在跑，程序退出时数据可能只写了一半 |

### 9. 单例初始化与配置热更新

**要实现什么**：连接池/配置只初始化一次；配置可以原子地整体替换，读侧（每个请求）不加锁、不阻塞，也绝不会读到**改了一半**的配置。

**组合**：`sync.Once`（只初始化一次）+ `atomic.Pointer`（读多写少的原子替换）。

```go
var (
    once sync.Once
    cfg  atomic.Pointer[Config]
)

func Config() *Config {
    once.Do(func() {
        c := loadConfig()
        cfg.Store(c) // 先完整构造好，再一次性发布
    })
    return cfg.Load()
}

func Reload() {
    cfg.Store(loadConfig()) // 整体换掉指针
}
```

| 方案 | 读并发 | 写代价 | 适用 |
| --- | --- | --- | --- |
| `Mutex` 包裹 | 串行 | 小 | 读写都少，最简单 |
| `RWMutex` | 并行 | 中 | 读明显多于写 |
| `atomic.Value` / `atomic.Pointer` | 并行且无锁 | 极小 | 值是**不可变对象**，整体替换 |
| 什么都不加 | 数据竞争 | —— | 一定不要 |

关键在**先构造完整，再一次性发布**：`loadConfig()` 返回的对象不再被任何人修改，指针替换是原子操作，读方要么看到整套旧配置、要么看到整套新配置，不会看到中间状态。这也是并发编程里最省心的一条路——**用不可变对象 + 原子发布，代替锁**。

### 10. 每个功能的验收方式

写完不算完，下面这张表给出**怎么证明它是对的**：

| 功能 | 验证方式 |
| --- | --- |
| 流水线 | 输入 N 个元素，输出正好 N 个；结束不阻塞 |
| fan-out / fan-in | 结果集合与串行版本一致（顺序不同，用排序后比较） |
| 任务池 | 并发数不超过 worker 数；`-race` 无报告；取消后 1 秒内返回 |
| 批量任务 + 超时 | 人为让某个任务变慢，确认整体在 T 内返回 |
| 并发聚合 | 结果顺序与输入一致；`-race` 无报告 |
| 错误传播 | 注入一个必失败的任务，确认返回该错误且其余任务被取消 |
| 并发限流 | 用计数器记录峰值并发数（在 `-race` 下断言峰值 ≤ limit） |
| 优雅关闭 | 关闭期间新任务被拒绝；在途任务跑完；超时分支返回错误 |
| 单例 | 并发调用 100 次，初始化只发生一次（在初始化里计数） |
| 对象池 | 复用率不是正确性问题，用 `-benchmem` 看分配次数下降 |

---

## 八、怎么选：一张决策流程

把所有工具放在一起，选择其实只需回答三个问题：

```text
                   我要处理的东西是什么？
                            │
        ┌───────────────────┴───────────────────┐
        ▼                                       ▼
  一份数据/任务的所有权                一份共用的内存状态
  （谁处理、处理完交给谁）               （缓存、配置、计数器）
        │                                       │
        ▼                                       ▼
      channel                              它有多大？
                                                │
                          ┌─────────────────────┼─────────────────────┐
                          ▼                     ▼                     ▼
                    单个变量              多个字段/结构体        读远多于写
                          │                     │                     │
                          ▼                     ▼                     ▼
                      atomic                Mutex                RWMutex
                                                │                     │
                                                └────────┬────────────┘
                                                         ▼
                                        只想初始化一次       → Once
                                        只想复用对象         → Pool
                                        等条件成立           → Cond（或 channel）

  另外三个与上面正交的问题：
  · 要等一组任务全部结束吗？      → WaitGroup
  · 要能中途停下/超时吗？         → select + context
  · 要通知所有 goroutine 停止吗？ → close(done) / ctx.Done()
```

常见场景速查：

| 场景 | 推荐 | 一句话理由 |
| --- | --- | --- |
| 把请求分给 N 个 worker | `channel` + `WaitGroup` | 任务所有权在移动 |
| 保护一个 map 缓存 | `Mutex` / `RWMutex` | 内存被共享 |
| 统计 QPS | `atomic.Int64` | 单个计数器 |
| 并发请求第三方接口，限制并发 | 信号量（带缓冲 channel） | 保护下游 |
| 加载一次配置 | `Once` + `atomic.Pointer` | 只做一次 + 无锁读 |
| 批量任务，最多等 3 秒 | `context.WithTimeout` + `WaitGroup` | 需要可中断 |
| 服务退出 | `close(done)` + `WaitGroup` | 广播 + 计数 |
| 一个连接池 | `Cond` 或带缓冲 channel | 等**有空闲**条件 |

---

## 九、排查清单

并发 bug 的特点是**不报错、偶发、重跑就好**。所以排查必须靠方法，不能靠感觉。

### 1. 按症状查

| 症状 | 优先检查 | 解释 |
| --- | --- | --- |
| 程序不退出、一直卡住 | 每个阻塞点：**谁会唤醒它**？ | 死锁、通道没人接收、`Wait` 少一次 `Done` |
| 结果少了 / 数字对不上 | 共享变量是否被并发读写、是否漏了锁 | 数据竞争 |
| 进程被 `fatal error` 干掉 | 是否并发写 map、是否 `close` 了两次、是否向已关闭通道发送 | runtime 直接终止，`recover` 无效 |
| goroutine 数量一直涨 | 每个 goroutine 的退出条件是什么 | 泄漏：没有人接收、没人取消、上游不关闭 |
| 偶发失败，重跑就过 | 用 `-race` 和 `-count=10` 复现 | 时序相关的问题 |
| 关闭服务时数据写了一半 | 是否只广播了停止，没有等 `WaitGroup` | 优雅关闭缺了第三步 |

### 2. 三个必问的问题

写并发代码时，这三句话能挡掉绝大多数事故：

1. **这个 channel 的发送方、接收方、关闭方分别是谁？**（答不上来就一定会出问题）
2. **这个 goroutine 靠什么退出？**（没有答案就是泄漏）
3. **这个阻塞点会被谁唤醒？**（写不出答案就是死锁）

### 3. 命令与工具

```bash
gofmt -l .            # 看哪些文件没格式化
go vet ./...          # 静态检查：copylocks（复制锁）、lostcancel（忘记 cancel）等
go test ./...         # 跑测试
go test -race ./...   # 竞态检测：最有用的一条命令
go test -count=10 ./...   # 重复跑，逼出偶发问题
go test -timeout 10s ./...  # 卡死时立刻失败，而不是一直挂着
go test -run TestPool -v ./...  # 只跑某一个测试
```

关于 runtime 的死锁提示要有正确的预期：`fatal error: all goroutines are asleep - deadlock!` **只在所有 goroutine 都阻塞时才出现**。一个真实的服务里总有别的 goroutine 在跑，所以最常见的现象是**程序静静地卡在那里**。这时用 `-timeout` 让测试失败，或者打印 goroutine 栈：

```go
// 定位泄漏：把当前所有 goroutine 的调用栈打出来
buf := make([]byte, 1<<20)
n := runtime.Stack(buf, true) // true：包含所有 goroutine
fmt.Println(string(buf[:n]))
```

### 4. 提交前的四个问题

做完练习后，逐条写下答案：

1. 每一个 channel 的发送方、接收方、关闭方是谁？
2. 每一个 goroutine 的退出条件是什么？
3. 共享状态的临界区在哪里？锁为什么覆盖了完整的访问（读和写都在里面）？
4. 如果消费者提前返回，生产者会不会永远阻塞？怎么验证？

---

## 十、课堂练习

每一级的验收标准都写清楚了，做完必须能通过 `go test -race ./...`。

### LV0：观察行为（理解语义）

1. 分别用 `WaitGroup` 和 `channel` 改写**`main` 直接退出**的示例，让 `hello` 一定被打印。
2. 创建容量为 0、1、2 的 channel，记录发送和接收分别在什么时刻阻塞，写成一张表。
3. 解释为什么**接收方关闭 channel**可能导致发送方 panic，并写出一个能复现的最小程序。
4. 用 `select` + `default` 实现**有就取，没有立刻返回**，对比用 `len(ch)` 判断的区别。

### LV1：安全计数器（`Mutex` / `atomic`）

实现 `Counter`，提供：

```go
func (c *Counter) Add(delta int)
func (c *Counter) Value() int
func (c *Counter) Reset()
```

验收标准：

- 100 个 goroutine 各加 1000 次，`Value()` 必须等于 100000；
- 零值 `Counter` 可直接使用；
- `go test -race` 无报告；
- 再用 `atomic.Int64` 实现一版，并用 `go test -bench` 比较两者性能。

### LV2：可取消的流水线

把第七部分第 1 小节的流水线改成接收 `context.Context`：

- 每一级的发送和接收都要能被取消；
- 调用方收到前两个结果后调用 `cancel()`；
- 测试必须在 1 秒内结束；
- 结束后用 `runtime.NumGoroutine()` 确认没有遗留 goroutine。

### LV3：worker pool（结课题）

实现一个 `Pool` 类型：

```go
func NewPool(workers int) *Pool
func (p *Pool) Submit(ctx context.Context, job int) error
func (p *Pool) Results() <-chan Result
func (p *Pool) Shutdown(ctx context.Context) error
```

要求：

- 指定 worker 数量，并发数不超过它；
- `Submit` 之后可以通过 channel 收集结果；
- 支持 `context` 取消；
- 每个 channel 只关闭一次，且关闭方明确；
- `Shutdown` 会拒绝新任务、等待在途任务、支持超时；
- 任何一个任务出错时取消其余任务，并把错误返回给调用方。

**动笔之前先画两张图**：任务 channel / 结果 channel / 关闭方 / 等待方分别是谁；每个 goroutine 的退出条件是什么。图画不出来，代码一定写不对。

---

## 十一、验收与提交

在项目根目录执行：

```bash
gofmt -w .
go vet ./...
go test ./...
go test -race ./...
```

四项全过再提交。第一次跑 `-race` 时如果编译报错，先回头看第一部分末尾关于 C 工具链的提示。

提交时在 PR 描述里回答第九部分那四个问题。**这四问不是形式**：它们正是三年后你 review 别人并发代码时会问的东西，而且比代码本身更难写对。

---

## 总结

回头看这一课，所有的工具都在为**建立顺序**和**明确所有权**服务：

| 工具 | 没有它会怎样 | 它建立的保证 |
| --- | --- | --- |
| `channel` | 数据传不出来，结束无法表达 | 发送 → 接收 的顺序；`close` 广播**没有更多了** |
| `select` | 只能等一个事件，无法超时 | 多路等待，`default` 为非阻塞 |
| `context` | 无法取消，goroutine 越积越多 | 广播**该停了**，并携带截止时间 |
| `WaitGroup` | 不知道一组任务何时结束 | `Done` → `Wait` 的顺序 |
| `Mutex` / `RWMutex` | 共享内存被并发读写，结果错或崩溃 | 互斥 + `Unlock` → `Lock` 的顺序 |
| `Once` / `atomic` | 初始化重复、单变量也要锁 | 只执行一次；单变量的原子读写 |

三句话带走：

1. **`channel` 管流，锁管存。** 先想清楚数据的所有权，再决定用哪个。
2. **每个 goroutine 都必须有退出条件，每个阻塞点都必须有人唤醒。** 写不出来就说明代码有问题。
3. **正确性是可以被检验的。** `go test -race` 不是可选项，它是这门课往后每一次提交的底线——第三课的练习题让你看见了问题，这一课让你有了证明的能力。

下一课开始，这些工具会从**练习题**变成**业务代码的一部分**：处理一个 HTTP 请求要用 `context` 传递取消和超时，访问数据库要通过连接池限制并发，返回结果要经过一层层 `channel` 或锁保护的缓存。等你写 Web 框架和数据库那两节课时，会不断回到这一课的表格上来。

