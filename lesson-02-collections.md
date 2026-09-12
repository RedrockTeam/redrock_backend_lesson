# 第二课：map 与 interface 让数据结构可扩展

## 目标

学生能选择 slice 或 map，判断 key 不存在的情况，定义带方法的类型，并使用最小接口隔离“数据”和“展示”。

## 流程（120 分钟）

1. 10 分钟：复盘第一课 bridge 的线性查找；
2. 25 分钟：map 的创建、零值、`value, ok`、删除和遍历无序性；
3. 25 分钟：方法、隐式实现接口、`fmt.Stringer` 和错误值；
4. 20 分钟：把成绩榜改成 `map[string]Student`，用接口输出文本/JSON；
5. 25 分钟：练习与 `go test -race` 初体验；
6. 15 分钟：观察并发读写 map 的失败，导入第三课。

## 现场代码

```go
type Formatter interface { Format(Student) string }

type PlainFormatter struct{}
func (PlainFormatter) Format(s Student) string { return s.Name }

func Find(students map[string]Student, name string) (Student, bool) {
    student, ok := students[name]
    return student, ok
}
```

说明 interface 只描述使用者需要的行为；不要为了“看起来高级”给每个类型都套接口。演示 `nil` map 不能写入、map 遍历顺序不保证。

## 练习与转场

- `01-scoreboard`：统计、排序和缺失 key；
- `02-formatter`：为文本和 JSON 实现同一接口；
- `03-map-race`：两个 goroutine 读写共享 map。普通测试可能通过，`go test -race` 会报告数据竞争。

在失败输出上停留一分钟：问题不是 map 查找慢，而是多个执行流同时访问共享内存。第三课先解释“执行流”如何出现，再决定怎样拆任务。
