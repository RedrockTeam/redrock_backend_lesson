# 第四课：channel、sync 与锁

## 这节课讲什么

从第三课的丢结果和竞态出发，讲无缓冲/有缓冲 channel、关闭规则、`select`、超时、`sync.Mutex`、`RWMutex` 和 `WaitGroup`，最后组合成固定 worker 数量的任务池。

## 课堂安排

- 25 分钟：发送/接收阻塞、channel 关闭和 `for range`；
- 20 分钟：`select`、超时与取消；
- 25 分钟：锁保护临界区，WaitGroup 等待 worker；
- 30 分钟：完成流水线和 worker pool 结课题；
- 15 分钟：同时运行 `go test ./...` 与 `go test -race ./...`；
- 5 分钟：回顾四课的提交、测试和反馈闭环。

## 讲解重点

发送方负责关闭 channel；锁只保护共享状态，不要把慢 I/O 放在临界区。明确 channel 负责传递和结束信号，锁负责共享内存互斥，WaitGroup 负责等待，三者解决不同问题。
