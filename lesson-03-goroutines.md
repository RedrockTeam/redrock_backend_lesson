# 第三课：计算机基础与 goroutine

## 目标

学生能用进程、线程、调度、栈/堆和数据竞争解释并发行为，知道 goroutine 不是“免费的 CPU”，并能等待和收集并发任务结果。

## 流程（120 分钟）

1. 20 分钟：CPU 核、进程地址空间、线程、调用栈/堆；
2. 15 分钟：C 线程/`pthread` 与 Go goroutine 的创建和调度成本对比；
3. 20 分钟：Go runtime 的 M/P/G 模型（只讲直觉，不要求背缩写）；
4. 25 分钟：`go` 关键字、闭包捕获、顺序打印与并发打印；
5. 25 分钟：把慢任务拆成 goroutine，使用结果 slice；
6. 10 分钟：`go test -race` 识别共享变量；
7. 5 分钟：未等待 goroutine 导致结果缺失，预告 channel/sync。

## 现场代码

```go
func run(name string, done chan<- string) {
    done <- "finished: " + name
}

func main() {
    done := make(chan string)
    go run("compile", done)
    fmt.Println(<-done)
}
```

先展示没有接收方时发送会阻塞，再展示用 `WaitGroup` 等待但仍需安全传递结果。强调 goroutine 的栈会按需增长，真正的瓶颈可能是 I/O、锁或 CPU，而不是“线程数量”。

## 练习与转场

- `01-goroutine-order`：并发执行任务并保持结果顺序；
- `02-parallel-sum`：比较顺序和并行版本；
- `03-missing-results`：只启动 goroutine、不可靠等待，测试超时或结果缺失。

提问：“谁拥有结果？什么时候可以关闭出口？如果一个任务失败，其他任务还要继续吗？”第四课用 channel、`WaitGroup`、锁和取消机制回答。
