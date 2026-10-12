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

下面用到了第四部分的 `sync.WaitGroup`——它只负责**等这组 goroutine 全部结束**，细节后面再讲，这里先看现象。

```go
func main() {
    var n int
    var wg sync.WaitGroup
    wg.Add(2)
    for i := 0; i < 2; i++ {
        // 启动两个协程
        go func() {
            defer wg.Done()
            for j := 0; j < 100000; j++ {
                n++
                // 每个协程加 100000 次
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
- 数据竞争在 C 里也许只是**结果有问题**，但在 Go 里更严重：**并发的 map 写会让整个进程直接 `fatal error: concurrent map writes` 退出**。


### 3. 三个问题，一张分工表

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

### 4. 为什么**看起来对**不算对：happens-before

同步原语的真正价值是**建立顺序保证**。

可以这样理解：程序中普通读写只有在同步操作建立顺序之后，才具有可依赖的可见性和顺序保证。编译器和 CPU 可能为了性能重排指令，硬件也有缓存层次；但学习这部分时，重点不是猜测某个缓存何时刷新，而是判断代码是否建立了明确的同步关系。

Go 的内存模型给出的说法是：**只有当操作 X happens before 操作 Y 时，X 写入的值才保证能被 Y 看到。** 日常用到的对应关系就是下表（左列 happens before 右列）：

| 操作 A | 操作 B |
| --- | --- |
| 向 channel 发送 | 对应的接收完成 |
| `close(ch)` | 从已关闭的 `ch` 收到零值 | 
| `mu.Unlock()` | 同一把 `mu` 的下一次 `Lock()` 返回 | 
| `wg.Done()` | 对应的 `wg.Wait()` 返回 | 
| `once.Do(f)` 里的 `f` 返回 | 任何 `once.Do` 返回 | 
| `atomic` 的某次写 | 后续观察到这个值的 `atomic` 读 |

**正确使用表中的同步原语，并且同步关系确实连接了相关访问时，才能建立对应的 happens-before 关系。** 才能防止前面提到的问题发生。

---

## 二、channel：在 goroutine 之间传递数据

### 1. 为什么需要 channel

并发程序需要把数据（结果、任务、信号）从**一个 goroutine 交给另一个**，还要能明确表达**不会再有数据了**，好让接收方知道何时退出。用共享内存做这两件事很麻烦：得额外维护一个标志位、再轮询它，既费 CPU 又加延迟；而且**有过数据**和**不会再有数据**是两回事，一个布尔量说不清后者。

`channel` 负责的正是这件事：它是一条**带类型的管道**。Go 推崇的写法是：

> **不要通过共享内存来通信，而要通过通信来共享内存。**

`channel` 同时具备四个属性：

| 属性 | 含义 |
| --- | --- | 
| 类型化 | `chan T` 只能传 `T` |
| 同步点 | 发送和接收构成一次配对 | 
| 可关闭 | 关闭是一个**对所有接收者同时生效**的信号 | 
| 有容量 | 无缓冲 / 有缓冲 |

### 2. 创建 channel：`make(chan T)` 与 `make(chan T, n)`

```go
ch := make(chan int)      // 无缓冲：容量为 0
buf := make(chan int, 3)  // 有缓冲：容量为 3
```

不能用 `new` 只能用 `make` ，`make` 可以初始化那些**需要内部结构**的类型（切片、map、channel）。channel 内部有一把锁、一个等待接收者队列、一个等待发送者队列和（可能的）缓冲区，用零值无法使用：

```go
var ch chan int // nil channel
ch <- 1         // 永久阻塞
```

**缓冲区解决什么问题。** 无缓冲 channel 要求发送和接收**同时就位**，有缓冲 channel 则允许发送方在缓冲区没满时先走一步，从而削减短期峰值，减少调度次数，但是不能解决收发速度差异，发送稳定快于接收仍会阻塞。


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

`ok` 这个返回值为什么必须存在？因为接收一个已关闭的 channel **不会报错**，它立刻返回零值：

```go
v := <-in
if v == 0 {   // 错：零值可能真是业务数据，也可能是上游已经关闭
    return
}

v, ok := <-in
if !ok {      // 对：用第二个返回值区分两种情况
    return
}
```

所以：**如果代码需要区分“收到零值”和“channel 已关闭”，就用 `v, ok := <-ch`**；如果明确知道 channel 不会关闭，或者使用 `for range` 消费，也可以采用更简洁的写法。


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

`nil` channel 的发送和接收会**永久阻塞**，`close(nil)` 会 panic。

这个特性本身就有用：一个永远阻塞的 channel 恰好表示**一个永远不触发的分支**。读完一个数据源后把它置为 `nil`，就能让后续只从剩下的源取数据——不需要额外的布尔标志，也不用写两遍循环，这是**用数据流的状态控制控制流**的典型手法。

要真正用上它得配合 `select`：`select` 会跳过永远不就绪的 `case`，于是把某个源设为 `nil` 就等于**关掉它的分支**。具体写法见第三部分。

### 6. 关闭通道：`close()` 的完整语义

**为什么需要关闭。** 没有 `close`，接收方永远无法区分两种情况：**上游暂时没有数据**和**上游永远不会再有数据**。只能靠约定（发一个哨兵值，例如上面的 `0`）、超时或者干脆不退出——这就是 goroutine 泄漏的源头。而 `close` 用一条语句表达**没有更多数据了**。

完整的语义如下：

| 操作 | 结果 |
| --- | --- |
| `close(ch)` | 标记 channel 为已关闭 |
| 关闭后再接收 | 先取完缓冲区里剩余的元素，**取完后立刻返回零值 + `false`**（不阻塞） |
| 已关闭的 channel 上 `for range` | 取完剩余元素后循环自动结束 |
| `close(ch)` 第二次 | **panic: close of closed channel** |
| 向已关闭的 channel 发送 | **panic: send on closed channel** |
| 从已关闭的 channel 接收 | 永不阻塞（这是**广播**能力的来源） |
| `close(nil)` | panic |

注意**从已关闭 channel 接收永远不阻塞**这一条：**关闭是一个对所有接收者同时生效的信号**。天然适合做**停止**广播：一次 `close(done)`，所有正在 `<-done` 或 `select` 中等待 `done` 的 goroutine 会**同时**被唤醒：

```go
done := make(chan struct{})
for i := 0; i < 5; i++ {
    go func() {
        <-done // 5 个 goroutine 一起等
        fmt.Println("stop")
    }()
}
close(done) // 一条语句唤醒全部

time.Sleep(time.Second) // 只是为了让上面那 5 行打印出来——main 一返回，它们就会像第一部分那样被静默杀死
```

对比一下**发送一个值来通知**的做法：

| 通知方式 | 能否唤醒多个等待者 | 是否需要知道等待者数量 | 能否重复通知 |
| --- | --- | --- | --- |
| `close(done)` | 能，全部同时 | 不需要 | 一次（再关就 panic） |
| `done <- struct{}{}` | 只能唤醒一个 | 需要 | 可以发多次，但没等待者时会阻塞 |
| `context.Context` | 能（内部就是关 channel） | 不需要 | 一次，但可带超时和原因 |

### 7. 谁有权关闭：所有权规则

`close` 的两条 panic 规则指向同一个结论：

> **由发送方关闭，接收方永远不关闭；有多个发送方时，由统一的协调者关闭。**

只有写数据的人才知道**我发完了**。接收方不知道还有没有别的发送者，如果接收方关闭其他发送者继续发送就会 panic

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

多个发送方的正确写法是：让所有生产者都结束后，由一个专门的 goroutine 关闭公共出口。关键是**等所有生产者退出**，通常由 `WaitGroup` 计数；完整组合见第七部分的 fan-in。

### 8. 遍历：`for range` 与 `for select` 的分工

| 写法 | 何时结束 | 能否同时处理取消 | 适用场景 |
| --- | --- | --- | --- |
| `for v := range ch` | channel 关闭且取完 | 不能 | 简单消费，明确知道上游会关闭 |
| `for { select { ... } }` | 由 `return` 决定 | 能 | 需要响应 `done` / `ctx.Done()` |

消费 channel 最常见的形式就是 `range`，它会一直取到 channel 关闭且缓冲区取空：

```go
for v := range ch {
    handle(v)
}
```

它其实就是第 3 小节那个 `v, ok := <-ch` 循环的语法糖。一旦这条链路需要**能被取消**——一边收数据一边等 `done`——就得把 `range` 换成 `select` 版，写法在第三部分。

反之，能用 `range` 往往说明这条链路是简单直线；如果还需要处理取消或多个事件，就应改用 `select`。

### 9. channel 的五种典型用法

channel 不只是**传数据**，下面五种用法都来自同一套语义：

| 用法 | 怎么写 | 场景 |
| --- | --- | --- |
| 数据流 / 任务队列 | `ch <- job` / `job, ok := <-ch` | worker pool、流水线 |
| 结果回收 | 每个 worker 把结果发到同一个 `results` | 并发任务收集 |
| 停止广播 | `close(done)`，全员 `<-done` | 优雅关闭、取消 |
| 计数信号量（限流） | 带缓冲 channel 当令牌桶，`sem <- struct{}{}` / `<-sem` | 限制并发数 |
| 互斥 / 单槽缓存 | `cap = 1` 的 channel | 保护单个值、最新值缓存 |

限流（信号量）的写法值得单独看，因为在真实项目里出现频率极高：**用带缓冲的 channel 当令牌桶**，容量就是并发上限。想开工先取一个令牌，做完再还回去；令牌被取空时，后来者自动挡在取令牌那一步排队：

```go
func withSem(sem chan struct{}, job int) {
    sem <- struct{}{}        // 取令牌：满了就在这里等
    defer func() { <-sem }() // 还令牌，panic 路径也不会漏
    handle(job)
}
```

创建时给定容量即可：`sem := make(chan struct{}, limit)`。要**给一批任务都套上它、并等它们全部结束**，还需要第四部分的 `sync.WaitGroup` 来计数（完整写法见第七部分的并发限流）。

用 `cap = 1` 的 channel 也能实现单槽同步，但它表达的是令牌传递；通用互斥通常优先使用 `sync.Mutex`，因为意图更清楚，也便于维护。

### 10. channel 的代价与边界

- **每次传递都有开销**：涉及 runtime 的调度与锁，比直接读写内存慢得多。高频计数、超小数据用 `atomic` 或 `Mutex` 更合适。
- **它带来两类新 bug**：发送没人接收会永久阻塞（死锁），接收方提前退出会让发送方永远卡住（goroutine 泄漏）——解决办法见第三部分。

---
## 三、select：同时等多个事件

### 1. 为什么需要 select

一个 worker 常常要**同时等好几个事件**：新任务来了要处理，收到取消要退出。只有 `channel` 时这很难两全——阻塞在一个 channel 上就没法再响应取消。
`select` 负责的就是**同时等多路事件**：它把多路等待做成一条语句，直接由 runtime 实现。

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
| 多个 `case` 同时就绪 | **伪随机**选一个执行（防止某个 case 永远等待） |
| 都不就绪 | 阻塞等待；若有 `default`，立刻执行 `default` 并结束 |
| `nil` channel 的 `case` | 永远不就绪 → 可用于动态禁用分支（见下例） |
| `case` 的求值 | 进入 `select` 时先求出所有 channel 操作数，再判断就绪 |
| `select {}` | 永久阻塞（什么时候才会报死锁，见第九部分） |
| `break` | 只跳出 `select`，不会跳出外层 `for` |

**`nil` channel 的用法**：把某个 `case` 的 channel 置为 `nil`，这个分支就永远不会被选中，等于**动态关掉它**。合并两个数据源时特别顺手——读完一个就把它设 `nil`，剩下的继续读：

```go
func merge2(a, b <-chan int, out chan<- int) {
    defer close(out) // 两个输入都结束后，统一关闭输出
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

如果下游可能只取前几个值就返回，`out <- v` 仍可能让这个合并 goroutine 阻塞；实际项目中应再传入 `context.Context` 或 `done`，让发送也能响应取消。

**伪随机**这条规则有一个重要推论：**不要依赖 `case` 的书写顺序。** 如果业务希望在已经取消时尽快退出，可以先做一次非阻塞取消检查，再进入正常等待；这能减少取消与业务事件同时就绪时继续处理的机会，但不是对所有竞态时序的绝对优先级保证：

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

### 3. 五种写法对照

| 写法 | 效果 | 典型用途 |
| --- | --- | --- |
| `select { case v := <-ch: }` | 阻塞，直到 ch 可接收 | 等结果 |
| `select { case ch <- v: }` | 阻塞，直到 ch 可发送 | 投递结果 |
| `select { case v := <-ch: default: }` | 非阻塞尝试，拿不到就走 | **有就取**、丢旧值 |
| `select { case <-time.After(d): }` | 等一个**定时事件** | 超时 |
| `select { case <-ctx.Done(): }` | 等一个**已取消**事件 | 取消、优雅关闭 |

### 4. 超时：`time.After`、`time.NewTimer`、`time.Ticker`

**为什么需要超时。** 网络对端可能永远不回应，数据库可能锁住，第三方接口可能挂起。等不到结果的等待必须有个上限，否则一个挂起的下游会把整条请求链拖住。

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

两个必要的提醒：

1. **`time.After` 不适合放在热循环里**（如每秒几十万次的分支判断）。每次都会创建一个定时器对象，这种位置用 `time.NewTimer` + `Reset` 或直接把超时提出去。
2. 超时只结这次等待，**不会终止对方的工作。** 示例代码在超时后直接返回，但产生 `results` 的 goroutine 仍然在跑。想要它停下需要把取消信号传进去，这是下一小节的事。

### 5. 取消：`done` channel 与 `context.Context`

Go **没有** **杀死一个 goroutine** 的 API，这是刻意的设计：goroutine 可能正持有锁、正在写共享状态、正在写文件，被强行中断会留下不一致的数据；而且它的栈由 runtime 管理，外部无法安全地把它关掉。所以 Go 的取消是**协作式**的：

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

注意内层那个 `select`：**发送也要能被取消**。如果没有它，当没有消费者时，worker 会永久阻塞在 `results <-` 上，`done` 也无法解决，这是最常见的泄漏形态。

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
ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
defer cancel() // 必须调用，否则计时器与子 context 无法释放

doSomething(ctx) // HTTP / 数据库这些库的第一个参数都是 ctx，取消会一路传下去

if errors.Is(ctx.Err(), context.DeadlineExceeded) {
    // 超时，不是别的错
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

启动了一组 goroutine，就需要知道**它们什么时候全部结束**，以便进行后续的业务。`WaitGroup` 负责的就是**等待一组 goroutine 结束**：一个**计数器**加一个**等待队列**。它描述了我们已知的信息，**不知道谁先结束，但知道一共有几个**。

### 2. 三个方法（外加一个新方法）

| 方法 | 作用 |
| --- | --- |
| `wg.Add(delta int)` | 计数加 delta，通常 `Add(1)` 或 `Add(n)` |
| `wg.Done()` | 等价于 `Add(-1)`，表示**这一个完成了** |
| `wg.Wait()` | 阻塞，直到计数归零 |
| `wg.Go(f func())` | Go 1.25 新增：等价于 `Add(1)` + `go` + 结束时自动 `Done` |

**零值即可用**：`var wg sync.WaitGroup` 直接就能用，不需要初始化。

```go
func main() {
    var wg sync.WaitGroup
    for i := 0; i < 3; i++ {
        wg.Add(1) // 新建 goroutine 前计数加 1
        go func(id int) {
            defer wg.Done() // 本协程结束时自动释放计数
            fmt.Println("worker", id, "done")
        }(i)
    }
    wg.Wait()//阻塞直到计数器归零
    fmt.Println("all workers finished")
}
```

状态机很简单，但它解释了所有规则：

```text
        Add(n)                  Done() 使计数归零
  0 ────────────────────►  n  ──────────────────────►  Wait() 返回
  │                         │
  │ Wait() 立刻返回          │ Wait() 在这里阻塞
  │                         │
  └─────────────────────────┘
```

### 3. 四条规则

| 规则 | 违反后果 |
| --- | --- |
| `Add` 要在启动 goroutine **之前**调用 | `Wait` 可能在计数还是 0 时提前返回，程序**没等就结束了** |
| `Add(1)` 与 `Done()` 必须一一对应 | 少一次 `Done` 就永久阻塞；多一次 `Done` 会 `panic: sync: negative WaitGroup counter` |
| 用 `defer wg.Done()` 而不是在末尾手动写 | 提前 `return` 或 panic 的路径上，`Done` 会被漏掉 |
| 不要复制 `WaitGroup`（用 `*sync.WaitGroup` 传递） | 两个计数器各自计数，`Wait` 永远等不到；`go vet` 会以 `copylocks` 报警 |

`Add` 的时机值得单独强调。下面这段是经典错误：

```go
for i := 0; i < 3; i++ {
    go func() {//创建了goroutine，同时 Wait 开始
        wg.Add(1) // 可能还没执行到计数器增加，Wait 已经返回了
        defer wg.Done()
        work()
    }()
}
wg.Wait()
```
**计数从 0 变成正数必须发生在 `Wait` 可能看到 0 之前**——这就是上表第一条规则的由来。

Go 1.25 的 `wg.Go` 把这三件事合成一件，减少了**忘写 `Done`**和**`Add` 放错位置**的机会。传入的函数不应 panic；如果任务可能失败，应在任务内部捕获并通过结果或错误 channel 传递错误，而不是依赖 panic：

```go
var wg sync.WaitGroup
for _, job := range jobs {
    wg.Go(func() { // 自动 Add(1)，函数返回后自动 Done
        handle(job)
    })
}
wg.Wait()
```


还有一点容易被忽略：**`Wait` 等的是最慢的那个**。一个 worker 卡住整组都等不到结束。所以真实的批量任务通常写成**等，但最多等一段时间**——这就需要把 `Wait` 和一个 channel 组合起来：

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

`waitTimeout` 只负责给等待增加时间上限，不会取消或终止正在执行的任务。超时后，内部等待 goroutine 仍可能继续存在，直到所有任务调用 `Done`。如果任务需要真正停止，必须让任务接收 `context.Context`，并在阻塞点检查取消信号。

---

## 五、锁：保护共享内存

### 1. 为什么需要锁

`channel` 传递的是**值**，但有些状态天生就是**共享的**：一个服务里的配置、一个缓存 map、一个连接池、一个累计计数器。这些数据不属于某一个 goroutine，它们被多方读写，此时用 channel 传递并不合适。

没有锁，共享内存会以三种方式出错：**更新丢失**（两个 goroutine `n++`，结果少了）、**结构损坏**（并发写同一个 map 直接 `fatal error: concurrent map writes`，进程退出）、**读到中间状态**（一个 goroutine 更新 `{账号, 余额}` 两个字段时，另一个只读到了半个更新的值，极难复现）。

锁提供的正是两件事：**互斥**和**顺序**。互斥是同一时刻只有一个 goroutine 在临界区里；顺序是上锁后，后面要上锁的 goroutine 必须等前者完成，于是建立一条 happens-before，临界区里的普通读写因此变得可见、有序。


### 2. `sync.Mutex`

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

它只有两个方法：`Lock()` 加锁（已有人持锁则阻塞）、`Unlock()` 解锁（对未加锁的 Mutex 解锁会 panic）。

特性与坑：

| 特性 | 说明 | 
| --- | --- | 
| 零值可用 | `var mu sync.Mutex` 就能用，不需要 `New` |
| 不可重入 | 同一个 goroutine 重复 `Lock` 会把自己锁死 | 
| 不绑定 goroutine | A 加锁、B 解锁语法上允许 | 
| 不可复制 | 复制会得到两把独立的锁 |

不可重入的解法不是换锁，而是**把需要锁的操作和不需要锁的内部逻辑分开**：内部函数不加锁，只在公开方法的边界上加锁。

### 3. 临界区怎么划

临界区就是**从 `Lock` 到 `Unlock` 之间的代码**，难点在于它该有多宽：

| 临界区 | 后果 |
| --- | --- |
| 只锁了写、没锁读 | 读到旧值，等于没同步 |
| 漏掉某个字段的更新 | 出现**半个新、半个旧**的状态 |
| 把网络请求、文件 I/O、耗时计算包进来 | 所有 goroutine 一起排队，并发退化为串行 |
| 刚好覆盖共享状态的读写 | 正确且并发度高 |

判断方法只有一问：**这段时间里除了我在改的那几个字段，还发生了什么**——只要包含与共享状态无关的计算或 I/O，就该挪出临界区。

### 4. 死锁：四种形态

落到 Go 代码里就是四类写法：

| 形态 | 例子 | 特征 |
| --- | --- | --- |
| 同一把锁重入 | 同一个 goroutine 重复 `Lock` | runtime 直接报 deadlock |
| 两把锁交叉 | A 持 `mu1` 等 `mu2`，B 持 `mu2` 等 `mu1` | 需要统一锁顺序 |
| 锁与 channel 交叉 | 持锁等待 channel，而发送方正等这把锁 | 最隐蔽 |
| 等一个永远不会发生的事件 | 无缓冲 channel 没人接收；`wg.Wait()` 少一次 `Done` | 表现为**程序静静地卡住** |

后两种**不会报错**（原因见第九部分），只能主动排查。方法只有一个——**对每个阻塞点写出谁会唤醒它**：

```text
阻塞点：ch <- result
谁会唤醒它？→ 消费者执行 <-ch
消费者在哪里？→ 只在 for range results 里
它有没有可能提前退出？→ 有，ctx.Done() 时它 return 了
→ 结论：一旦取消，这里就再也不会有人接收 → 必须给发送加 select
```

### 5. `sync.RWMutex`：读多写少的优化

`Mutex` 把并发的读也串行化了。如果一份数据**读远多于写**（配置、路由表、本地缓存），可以让多个读者同时进入：读用 `RLock()` / `RUnlock()`（可被多个 goroutine 同时持有），写用 `Lock()` / `Unlock()`（独占，有读者时等待）。

```go
type ConfigStore struct {
    mu   sync.RWMutex
    data map[string]string
}

func NewConfigStore() *ConfigStore {
    return &ConfigStore{data: make(map[string]string)}
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

`sync.RWMutex` 本身零值可用，但包含 map 的结构体不一定能直接使用：上面的 `data` 仍需通过 `NewConfigStore` 初始化。

| 要点 | 说明 |
| --- | --- |
| 读之间共享，读与写、写与写互斥 | 唯一的收益来源 |
| 写者优先 | 有写者等待时，新的 `RLock` 会排队，避免写者被饿死 |
| 读锁不可递归 | 持 `RLock` 时再次 `RLock`，中间有写者插队就是死锁 |

**`RWMutex` 属于优化，不属于设计。** 先用 `Mutex` 写出正确的代码，用 `go test -bench` 确认读取确实是瓶颈，再考虑换成 `RWMutex`。


---

## 六、`sync` 包的其余工具（进阶）

前面的核心工具包括 `channel`、`WaitGroup`、`Mutex` 和 `RWMutex`。`sync` 包里还有几件工具，它们各自解决一个很具体的问题。本节只要求知道它们存在、知道它们解决什么问题，遇到对应场景时再查文档。

下表是它们的速览，遇到对应场景再查文档：

| 工具 | 解决什么问题 | 关键 API | 使用要点 |
| --- | --- | --- | --- |
| `sync.Once` | 某资源只需初始化一次，且要并发安全 | `Do(f)`（Go 1.21+ 另有 `OnceFunc`/`OnceValue`） | `f` 里 panic 也算**已执行**，不会重试；`Once` 只管创建一次，之后的并发读写仍要锁 |
| `sync.Map` | 读多写少、key 集合稳定的并发键值表 | `Load`/`Store`/`LoadOrStore`/`Range` | 键值是 `any`，取值要类型断言；`Range` 不是一致性快照；**默认仍用 `map` + `RWMutex`** |
| `sync.Pool` | 高频创建/丢弃大对象，压力全给 GC | `Get`/`Put` | 对象随时可能在 GC 时被清空，拿到后**必须重置**；是**减少分配**的工具，不是缓存 |
| `sync.Cond` | 等一个条件成立才继续 | `NewCond`/`Wait`/`Signal`/`Broadcast` | `Wait` 必须持锁调用，且必须放在 `for` 循环里重新检查条件；等条件大多已被 channel 取代 |
| `sync/atomic` | 单个变量（计数、标志、指针）的无锁读写 | `Add`/`Load`/`Store`/`Swap`/`CompareAndSwap` | 只保证**单个变量**的原子性；多个变量的联合一致仍要锁 |

Go 1.19 起提供了类型化的原子类型：`atomic.Bool`、`atomic.Int32/Int64`、`atomic.Uint32/Uint64`、`atomic.Pointer[T]`、`atomic.Value`。

```go
var n atomic.Int64
n.Add(1)                    // 计数

var once sync.Once
once.Do(func() { init() })  // 并发调用下只执行一次
```

---
## 七、组合：把工具拼起来能得到什么

前面六部分讲的是零件。真实的并发功能都是**组合**出来的，而且组合方式相当固定——下面这五种覆盖了绝大多数场景。这一节对每个功能都给出三件事：**要实现什么**、**组合了哪些 API**、**代码**，最后用一句话点出**少了某个零件会怎样**。


### 1. 流水线（pipeline）

**要实现什么**：把一个大任务拆成若干级，每级一个 goroutine，数据像流水线一样流过；上游处理完自动告诉下游**没有更多了**。**组合**：`channel` + `close` + `for range`。

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

少了任何一级的 `defer close(out)`，下游的 `range` 就永远不结束；下游提前退出（只想取前几个），上游会卡在发送上泄漏。流水线的价值不在快，而在**可组合**：每级都是**输入 channel → 输出 channel** 的纯函数形状，能像搭积木一样增删。

### 2. fan-out / fan-in（分发与汇聚）

**要实现什么**：同一个输入被多个 worker 并行消费（fan-out，天然负载均衡：谁先空下来谁先拿），多个输出再汇总成一个（fan-in）。**组合**：`channel` + `WaitGroup`（判断所有生产者都结束了）+ 单点 `close`。

```text
                        ┌──► worker 1 ──┐
  生成器 ──jobs────────►├──► worker 2 ──┼──results──► 汇总
                        └──► worker 3 ──┘
                         fan-out        fan-in
```

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

`close(out)` 遵循前文的单点关闭规则；`wg.Wait()` 必须放在另一个 goroutine 里，否则函数会等到全部处理完才返回。`in` 作为闭包参数传入，可避免旧版本 Go 的循环变量捕获问题。结果顺序会变化；要保序，就按下标写回预先分配的切片。

### 3. 固定 worker 数量的任务池

**要实现什么**：限制**同一时刻有多少个任务在跑**。不限制一万个请求就起一万个 goroutine，下游连接池会被打爆。**组合**：`jobs` channel（任务队列）+ `results` channel（结果回收）+ `WaitGroup`（等 worker 全退出）+ `context`（取消）。

```go
type Result struct {
    JobID int
    Value int
}

func worker(ctx context.Context, jobs <-chan int, results chan<- Result, wg *sync.WaitGroup) {
    defer wg.Done()
    for {
        select {
        case <-ctx.Done():  // 先检查取消信号
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

三个关闭点分别由发送任务的人、等待 worker 结束的协调者和调用方负责：`close(jobCh)`、`close(resultCh)`、`cancel()`。它们不能混用，否则会导致 panic 或泄漏。当前示例只演示成功结果；如果任务可能失败，应把 `Result` 扩展为包含 `Err` 字段的类型，协调者收到第一个错误后调用派生 context 的 `cancel()`，其余 worker 在取任务和发结果时都监听取消。即使已经发现错误，也要继续负责关闭结果 channel，并等待所有 worker 退出，避免留下 goroutine。第十部分的结课题要求把这条错误传播链路完整实现。

### 4. 并发限流（信号量）

**要实现什么**：不管有多少任务，同一时刻最多 K 个在跑——保护下游（数据库、第三方接口）最直接的手段。**组合**：容量为 K 的 `channel` 当令牌桶 + `WaitGroup`。

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

顺带一提：`out[i] = compute(job)` 这个写法**天然保序**——每个任务只写自己那一格，结果顺序与输入顺序一致，不需要额外的锁或排序。

和固定 worker 任务池的区别只有一问：**任务源源不断**（队列式）用任务池，**任务数不算巨大但要独占外部资源**用信号量。信号量里每个任务仍是一个 goroutine，只是被挡住，任务量极大时仍要小心；要在等令牌时也能取消，就把 `select` 写在取令牌那一步。

### 5. 每个功能的验收方式

下面这张表给出**怎么证明它是对的**：

| 功能 | 验证方式 |
| --- | --- |
| 流水线 | 输入 N 个元素，输出正好 N 个；结束不阻塞 |
| fan-out / fan-in | 结果集合与串行版本一致（顺序不同，用排序后比较） |
| 任务池 | 并发数不超过 worker 数；`-race` 无报告；取消后 1 秒内返回 |
| 并发限流 | 用计数器记录峰值并发数（在 `-race` 下断言峰值 ≤ limit） |

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
| 服务退出 | `close(done)` + `WaitGroup` | 广播停止信号，并等待所有 goroutine 退出 |
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
| 关闭服务时数据写了一半 | 是否只广播了停止，没有等 `WaitGroup` | 只 `close(done)` 不 `Wait`，进程就已经退出 |

### 2. 提交前的四问

写并发代码时，这四句话能挡掉绝大多数事故；

1. 每个 channel 的发送方、接收方、关闭方分别是谁？
2. 每个 goroutine 靠什么退出？每个阻塞点会被谁唤醒？
3. 共享状态的临界区在哪里？锁有没有覆盖完整的读和写？
4. 消费者提前返回时，生产者会不会永远阻塞？怎么验证？

### 3. 命令与工具

```bash
gofmt -l .                 # 看哪些文件没格式化
go vet ./...               # 静态检查：copylocks（复制锁）、lostcancel（忘记 cancel）等
go test -race ./...        # 竞态检测：最有用的一条
go test -count=10 ./...    # 重复跑，逼出偶发问题
go test -timeout 10s ./... # 卡死时立刻失败，而不是一直挂着
```

关于 runtime 的死锁提示要有正确的预期：`fatal error: all goroutines are asleep - deadlock!` **只在所有 goroutine 都阻塞时才出现**。真实服务里总有别的 goroutine 在跑，所以最常见的现象是**程序静静地卡在那里**——用 `-timeout` 让测试失败，再用第五部分那个**阻塞点会被谁唤醒**的方法定位。

---

## 十、课堂练习

做完必须能通过 `go test -race ./...`。

### LV1：安全计数器（`Mutex` / `atomic`）

实现 `Counter`，提供：

```go
func (c *Counter) Add(delta int)
func (c *Counter) Value() int
func (c *Counter) Reset()
```

验收标准：

- 100 个 goroutine 各加 100 次，`Value()` 必须等于 10000；
- 零值 `Counter` 可直接使用；
- `go test -race` 无报告；

### LV2：可取消的流水线

把第七部分第 1 小节的流水线改成接收 `context.Context`：

- 每一级的发送和接收都要能被取消；
- 调用方收到前两个结果后调用 `cancel()`；
- 测试必须在 1 秒内结束；
- 结束后用 `runtime.NumGoroutine()` 辅助观察是否有遗留 goroutine；它是进程级瞬时数量，不能单独作为严格断言，还要配合超时和明确的退出信号。

### LV3：worker pool（结课题）

实现一个 `Pool` 类型：

```go
func NewPool(workers int) *Pool
func (p *Pool) Submit(ctx context.Context, job int) error
func (p *Pool) Results() <-chan Result
func (p *Pool) Shutdown(ctx context.Context) error
```

要求（骨架可参考第七部分第 3、5 小节）：

- 指定 worker 数量，并发数不超过它；
- `Submit` 之后可以通过 channel 收集结果；
- 支持 `context` 取消；
- 每个 channel 只关闭一次，且关闭方明确；
- `Shutdown` 会拒绝新任务、等待在途任务、支持超时；
- 任何一个任务出错时取消其余任务，并把错误返回给调用方。


---

## 十一、总结

回头看这一课，所有的工具都在为**建立顺序**和**明确所有权**服务：

| 工具 | 没有它会怎样 | 它建立的保证 |
| --- | --- | --- |
| `channel` | 数据传不出来，结束无法表达 | 发送 → 接收 的顺序；`close` 广播**没有更多了** |
| `select` / `context` | 只能等一个事件；无法取消，goroutine 越积越多 | 多路等待、`default` 非阻塞；广播**该停了**并带截止时间 |
| `WaitGroup` | 不知道一组任务何时结束 | `Done` → `Wait` 的顺序 |
| `Mutex` / `RWMutex` | 共享内存被并发读写，结果错或崩溃 | 互斥 + `Unlock` → `Lock` 的顺序 |


下一课开始，这些工具会从**练习题**变成**业务代码的一部分**：处理一个 HTTP 请求要用 `context` 传递取消和超时，访问数据库要通过连接池限制并发，返回结果要经过一层层 `channel` 或锁保护的缓存。有很多业务代码因为需求都有高度结构化且相似的内容，理解并熟练掌握本节课的内容会对以后的业务编写（或者单纯阅读理解ai的代码）有很大帮助

