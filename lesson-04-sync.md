# 第四课：channel、sync 与可控并发

## 目标

学生能区分通信和共享内存两种同步方式，正确关闭 channel，使用 `Mutex`/`RWMutex` 保护状态，用 `WaitGroup` 等待 worker，并通过 `go test -race` 验证。

## 流程（120 分钟）

1. 20 分钟：无缓冲/有缓冲 channel、发送/接收阻塞和关闭规则；
2. 20 分钟：`select`、超时和取消；
3. 20 分钟：`sync.Mutex`、`RWMutex`、`WaitGroup`，锁的临界区；
4. 25 分钟：实现固定 worker 数量的任务池；
5. 25 分钟：完成结课题并运行功能测试与竞态测试；
6. 10 分钟：回顾四课数据流和提交到榜单的闭环。

## 现场代码

```go
func worker(jobs <-chan Job, results chan<- Result, wg *sync.WaitGroup) {
    defer wg.Done()
    for job := range jobs {
        results <- process(job)
    }
}
```

由发送方负责关闭 channel；接收方使用 `for range` 直到关闭。锁只保护共享状态，不要把网络请求放在锁内。`WaitGroup` 负责“何时结束”，channel 负责“如何传值”，两者经常一起使用。

## 练习与验收

- `01-channel-pipeline`：合并多个输入流并在所有输入关闭后关闭输出；
- `02-safe-counter`：实现并发安全的计数器；
- `03-worker-pool`：限制并发数、支持错误结果和超时。

结课验收命令：

```bash
go test ./...
go test -race ./...
```

把测试结果上传评分服务；榜单只显示最高分，不把速度奖励置于正确性之前。
