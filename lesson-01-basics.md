# 第一课：从红岩网校协作到第一个 Go 包

## 目标

学生能说清后端请求从路由到数据库的大致路径，完成一次 GitHub Classroom 提交，并读写 `package`、`import`、`var`、`func`、`struct`、slice 和错误返回。

## 流程（120 分钟）

1. 10 分钟：红岩网校后端研发部做什么——接口、数据、部署和协作边界；
2. 15 分钟：浏览器请求如何到达 HTTP handler，展示本课程评分服务的 `/api/health`；
3. 20 分钟：Git 三个区域（工作区、暂存区、提交），分支、PR 和 Classroom 接受作业；
4. 30 分钟：从 C 映射到 Go，现场编写 `Student`、`Average` 和 `Greeting`；
5. 25 分钟：`go fmt`、`go test -v`、读失败信息；
6. 15 分钟：完成 `lesson-01-basics` 的 warm-up 与 core；
7. 5 分钟：bridge 题失败复盘，预告 map/interface。

## 现场代码

```go
package student

type Student struct {
    Name   string
    Scores []int
}

func Average(s Student) float64 {
    if len(s.Scores) == 0 { return 0 }
    total := 0
    for _, score := range s.Scores { total += score }
    return float64(total) / float64(len(s.Scores))
}
```

强调三点：slice 是带长度和容量的描述符；结构体按值复制，必要时使用 `*Student`；Go 没有 C 风格的头文件，导出名首字母大写。

## Git 演示脚本

```bash
git clone <classroom-repo>
git switch -c work
go test ./...
git add . && git commit -m "solve greeting"
git push -u origin work
```

展示一个小而清楚的提交，并在 Classroom 的 Actions 中找到测试 artifact。不要现场展示真实 token 或学生邮箱。

## 练习与转场

- `01-basics`：实现问候、结构体平均分和边界条件；
- `02-linear-lookup`：用 slice 线性查找学生并格式化输出；
- bridge 故意保留重复遍历和字符串拼接。提问：“如果有 10 万名学生，按姓名查找还要从头扫吗？如果不同输出格式越来越多，函数会变成什么样？”第二课用 map 和 interface 回答。
