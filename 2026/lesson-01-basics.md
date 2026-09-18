# 第一课：什么是后端与 Go 基础

## 第一部分：什么是后端

### 1. 后端负责什么

后端开发，就是写运行在服务器上的程序，让网站或 App 能存数据、查数据、做判断、处理业务，并把结果返回给前端。

网站中一次请求与响应的基本流程如下：

用户 → 前端 → HTTP/API → 后端 → 数据库

用户 ← 前端 ← JSON/数据 ← 后端 ← 数据库

### 2. 从两个实际例子理解后端

列举两个例子：

1. 微博发帖子：用户写完文字，点“发布”。页面把文字和“我是谁”一起交给网站后台。后台先确认这个人能不能发、内容有没有问题，然后把这条内容放进长期保存的资料库。同时，后台把“有新内容了”放进一个排队箱，稍后再通知粉丝、更新搜索。最后，后台告诉页面“发布成功”，页面把新内容显示出来。
2. 提交订单：用户选好商品，点“提交订单”。页面把商品、数量、收货地址交给网站后台。后台检查库存够不够、价格对不对、优惠能不能用，然后创建订单、扣减库存，并把订单放进资料库。接着，后台把“该付款了”放进排队箱，稍后再通知付款和物流。最后，后台返回订单号，页面跳到付款页。

### 3. 后端由哪些部分组成

后端不是只做一件事，而是几块配合：

* 接待入口：接住页面发来的请求，分给对应处理的人，挡住异常流量。
* 业务处理：真正做判断，比如能不能发、库存够不够、有没有点过赞、密码对不对。
* 资料库：长期保存用户、帖子、订单、点赞这些数据。
* 临时小本子：存登录状态、热门内容、点赞数，读得快。
* 排队箱：把不急的事先排队，比如发通知、更新搜索、发推荐。
* 搜索柜：专门按关键词找内容。
* 文件柜：存图片、视频、附件。
* 监控记录：记下发生了什么，出问题能报警。

下面用更具体的技术栈来表示这些功能块，以及由它们组成的完整后端体系：

```text
+-----------------------------------------------------------------------+
| 用户层：浏览器 / App / 小程序                                          |
| 操作：发布、评论、点赞、下单、支付、登录、搜索                          |
+-----------------------------------+-----------------------------------+
                                    | ① 请求
                                    v
+-----------------------------------------------------------------------+
| 接入层：CDN / WAF / 负载均衡                                           |
| Nginx / Traefik(Go) / APISIX / Kong                                    |
+-----------------------------------+-----------------------------------+
                                    |
                                    v
+-----------------------------------------------------------------------+
| Go Web/API 层                                                         |
| Gin / Echo / Fiber / Chi                                              |
| 中间件：golang-jwt/jwt、Casbin、x/time/rate、Zap、Prometheus           |
+-----------------------------------+-----------------------------------+
                                    |
                                    v
+-----------------------------------------------------------------------+
| Go Service 业务层                                                     |
| 用户服务、内容服务、订单服务、点赞服务、搜索服务                       |
| Wire / Fx 依赖注入、Viper 配置、DTO/VO、错误处理                       |
+------+----------------+----------------+----------------+-------------+
       |                |                |                |
       v                v                v                v
+-------------+  +-------------+  +-------------+  +-------------------+
| Go DAO/Repo |  | Go 缓存     |  | Go 消息     |  | Go RPC/搜索/存储  |
| GORM/sqlc   |  | go-redis    |  | Sarama      |  | gRPC-Go/protobuf  |
| sqlx/ent    |  | Redis       |  | amqp091-go  |  | go-elasticsearch  |
|             |  |             |  | NSQ/NATS    |  | minio-go          |
+------+------+  +------+------+  +------+------+  +---------+---------+
       |                |                |                   |
       v                v                v                   v
+-------------+  +-------------+  +-------------+  +-------------------+
| MySQL/PG    |  | Redis       |  | Kafka/Rabbit|  | ES / MinIO / S3   |
| 主从/分库   |  | 缓存/Session|  | MQ          |  | 搜索/对象存储     |
+-------------+  +-------------+  +-------------+  +-------------------+
       |                |                |                   |
       +----------------+----------------+-------------------+
                                    |
                                    | ② JSON/数据响应
                                    v
+-----------------------------------------------------------------------+
| Go Web 层：c.JSON / json.Marshal / gRPC 响应                           |
+-----------------------------------+-----------------------------------+
                                    |
                                    v
+-----------------------------------------------------------------------+
| 前端：解析 JSON，更新 DOM / 状态 / 列表 / 按钮                         |
+-----------------------------------+-----------------------------------+
                                    v
                                  用户看到结果
```

### 4. 从一个简单的三层项目开始

这样一个成熟、完整的单服务架构看起来有点吓人。直接要求你们掌握整个架构体系并不现实，更实际的路径是先掌握一些基础技术栈，并基于它们搓一个简单的小项目：

```text
┌──────────────────────────────────────────────┐
│ 第一层：用户 / 前端请求                       │
│ 浏览器、App、页面、按钮、表单、列表            │
│ 职责：展示内容、收集输入、发请求、显示结果      │
│ 不负责：业务规则、存数据、直接连数据库          │
└──────────────────────────────────────────────┘
=============== 通信边界：HTTP / JSON ===============
┌──────────────────────────────────────────────┐
│ 第二层：Web API 层                            │
│ 接请求、验参数、做业务判断、读写数据库、返回数据 │
│ 新手推荐：FastAPI / Express / Gin              │
│ 不负责：页面样式、长期保存数据                  │
└──────────────────────────────────────────────┘
============ 通信边界：SQL / 数据库驱动 ============
┌──────────────────────────────────────────────┐
│ 第三层：数据库存储层                          │
│ 表、记录、增删改查、长期保存                   │
│ 新手推荐：SQLite → MySQL / PostgreSQL          │
│ 不负责：业务规则、页面展示                     │
└──────────────────────────────────────────────┘
```

这个相对简单、基础的架构看起来就可爱得多。后续很长一段时间里，我们会聚焦于 Go 语言的实现和相关技术栈的掌握，让你具备独自实现这样一个小项目的能力。

### 5. 授课计划

预计的授课计划如下：Go 语言的基础语法教学会持续一个月，Web 框架（第二层）会持续两周，数据库（第三层）会持续两周。大概在三个月内，你们就能具备搭建一个小型后端项目的能力。

## 第二部分：从 C 到 Go 的语法迁移

在学校这几周的课程中，我想你们已经简单学习了一些 C 语言语法，比如 `struct`、变量和函数定义等。学完这些内容后，最直观的感受可能是：“我不知道这些东西能做什么，也很难独自写出一些东西。”前一个问题我们已经在第一部分讨论过，后一个问题则会在后续课程中逐步给出答案。

我们先从 Hello World 开始，看看同一个程序在 C 和 Go 中分别怎样编写。

### 1. 从 Hello World 看 C 和 Go 的差异

```c
#include <stdio.h>

int main(void) {
    printf("hello world!\n");
    return 0;
}
```

```go
package main

import "fmt"

func main() {
    fmt.Printf("hello, world\n")
}
```

这两个 Hello World 程序已经反映出一些最直观的语法差异：

| 部分 | C | Go | 注意 |
|---|---|---|---|
| 引入功能 | `#include <stdio.h>` | `import "fmt"` | C 是头文件，Go 是包 |
| 主函数 | `int main(void) {` | `func main() {` | Go 用 `func` 开头 |
| 返回值 | `return 0;` | 不写 | Go 的 `main` 无返回值 |
| 打印 | `printf("hello\n");` | `fmt.Println("hello")` | Go 用包名点函数 |
| 语句结尾 | `;` | 通常不写 | Go 自动插入分号 |
| 大括号 | 可以另起一行 | 必须和函数头同一行 | 否则自动分号会出错 |

### 2. 最基本的语法差异

下面来看一些更具体的语法差异。

#### 2.1 函数定义

C 的函数定义：

```c
返回类型 函数名(参数列表) {
    函数体
}
```

例子：

```c
int add(int a, int b) {
    return a + b;
}
```

Go 的函数定义：

```go
func 函数名(参数列表) 返回类型 {
    函数体
}
```

例子：

```go
func add(a int, b int) int {
    return a + b
}
```

同类型参数可以合并：

```go
func add(a, b int) int {
    return a + b
}
```

需要记住的是：

- C：返回类型写在最前面。
- Go：用 `func` 开头，返回类型写在参数后面。
- C：参数是 `类型 名字`。
- Go：参数是 `名字 类型`。
- Go 的 `main` 很特殊：`func main()`，没有参数，也没有返回值。

#### 2.2 分号

C 每条语句末尾都要分号：

```c
int a = 1;
printf("%d\n", a);
return 0;
```

Go 通常不写分号：

```go
a := 1
fmt.Println(a)
```

但 Go 不是“完全不能写分号”。编译器会在符合规则的行尾自动插入分号，所以一般不需要手写。即使在普通语句末尾显式写出分号，通常也不影响结果；`gofmt`（用于规范化 Go 代码格式的工具）会去掉不必要的分号。

```go
func main()
{
    fmt.Println("hello")
}
```

这样写在 Go 里会出错。因为 `func main()` 后面发生换行时，编译器会自动插入分号，导致后面的 `{` 无法与函数声明连在一起。在实际 Go 代码中，大家通常都不会在语句末尾手写分号。

```go
func main() {
    fmt.Println("hello")
}
```

左大括号 `{` 必须和 `func main()` 在同一行。`if`、`for` 也一样。原因同上。

#### 2.3 打印和大小写

C：

```c
printf("hello world!\n");
```

Go：

```go
fmt.Println("hello, world")
```

或者：

```go
fmt.Printf("hello, world\n")
```

注意：

- Go 里的包名是小写的 `fmt`。`fmt` 是 Go 标准库中最常用的包之一，主要负责格式化输入和输出。你可以先把它理解成 C 语言里的 `stdio.h` 加上 `printf`、`scanf`，但 Go 的 `fmt` 更安全，也更好用。
- `Println`、`Printf` 的首字母大写。Go 使用首字母大小写控制一个名称能否被其他包使用，第三部分会在讲解包级组织方式时进一步说明这一点。
- Go 中普通字符串字面量使用双引号 `"..."`。
- 单引号 `'a'` 在 Go 里是字符，不是字符串。
- `fmt.Println` 会自动换行。
- `fmt.Printf` 不会自动换行，需要自己写 `\n`。

#### 2.4 变量声明

C：

```c
int age = 18;
float score = 90.5;
```

Go：

```go
var age int = 18
var score float64 = 90.5
```

函数内部可以简写：

```go
age := 18
score := 90.5
```

需要记住：

- C：类型在前，变量名在后。
- Go：变量名在前，类型在后。
- Go 函数内常用 `:=` 自动推断类型。
- Go 变量未初始化时有零值，比如 `int` 是 `0`，`string` 是 `""`。

#### 2.5 条件判断和循环

C：

```c
if (x > 0) {
    printf("正数\n");
}

for (int i = 0; i < 3; i++) {
    printf("%d\n", i);
}
```

Go：

```go
if x > 0 {
    fmt.Println("正数")
}

for i := 0; i < 3; i++ {
    fmt.Println(i)
}
```

区别：

- Go 的 `if` 条件不需要括号。
- Go 的 `for` 条件也不需要括号。
- 但 `for` 的三段之间仍然用分号：`for i := 0; i < 3; i++`。
- 左大括号仍然要和 `if`、`for` 同一行。

### 3. 把 C 代码改成 Go 代码

把 C 代码改成 Go，可以按这个顺序检查：

1. 新建 `.go` 文件，第一行写 `package main`，先当模板。
2. 需要打印就写 `import "fmt"`。
3. 把 `int main()` 改成 `func main()`。
4. 删掉 `return 0`。
5. 把 `printf(...)` 改成 `fmt.Printf(...)` 或 `fmt.Println(...)`。
6. 删掉所有行尾分号。
7. 确保 `{` 跟在函数、`if`、`for` 同一行。
8. 变量声明改成 `var 名字 类型`，函数内可用 `名字 := 值`。
9. 函数定义改成 `func 函数名(参数) 返回类型`。
10. 用 `go run 文件名.go` 运行。

语法迁移解决的是“单个程序怎么写”的问题。接下来再看另一个重要问题：当代码越来越多时，C 和 Go 分别怎样组织一个项目。

## 第三部分：从 C 的项目组织到 Go 的 package

### 1. C 的项目组织：头文件、源文件与库

C 的项目组织形式主要依靠：

- `.h` 头文件：放声明，给别人看。
- `.c` 源文件：放定义，真正实现。
- 静态库 `.a` / `.lib`：编译后打包，链接进程序。
- 动态库 `.so` / `.dll`：运行时加载。
- `#include`：文本包含头文件。
- `static`：限制函数或变量只在当前文件可见。

典型结构：

```text
project/
  main.c
  math.h
  math.c
  libmath.a
  libmath.so
```

例子：

```c
// math.h
#ifndef MATH_H
#define MATH_H

int add(int a, int b);

#endif
```

```c
// math.c
#include "math.h"

int add(int a, int b) {
    return a + b;
}

static int helper(int x) {
    return x * 2;
}
```

```c
// main.c
#include <stdio.h>
#include "math.h"

int main(void) {
    printf("%d\n", add(1, 2));
    return 0;
}
```

编译：

```bash
gcc main.c math.c -o app
```

静态库：

```bash
gcc -c math.c -o math.o
ar rcs libmath.a math.o
gcc main.c -L. -lmath -o app
```

动态库：

```bash
gcc -shared -fPIC math.c -o libmath.so
gcc main.c -L. -lmath -o app
```

C 的可见性靠：

- 头文件里声明：相当于公开接口。
- `static` 函数：只在当前 `.c` 文件可见。
- `extern` 变量：跨文件共享。
- 链接器决定静态库、动态库怎么连。

我们暂且不深入讨论动态库和静态库，先关注 C 中的函数对于同一项目内其他文件是否可见：

- C 中的函数默认可以被其他源文件链接使用；添加 `static` 后，函数会被限制在当前 `.c` 文件内。
- 如果多个源文件需要使用同一组函数，通常会把这些函数的声明集中写在头文件中，再由需要它们的 `.c` 文件通过 `#include` 引入。
- 在更大的项目中，一组功能相关的源文件往往会对应一个对外头文件。这种做法既划分了功能模块，也减少了开发者逐个维护和查找声明的负担。

例如，一个项目可以按照用户和订单两组功能来组织：

```c
user/
  user.h              // 对外公共头，只放公开 API
  user_internal.h     // 模块内部共享头，只给本模块 .c 用
  user_register.c
  user_login.c
  user_store.c

order/
  order.h
  order_internal.h
  order_create.c
  order_cancel.c

main.c
```

尽管 C 语言本身没有严格规定项目必须这样组织，但随着项目逐渐演进、架构逐渐产生层级，把核心功能的声明集中到一个头文件中，几乎是一种自然且必然的做法。这种方式实际上把整个项目划分成若干功能块，并通过文件树表现出来。

在上面的例子中，`user` 目录负责用户相关的逻辑实现，`user.h` 负责提供对外接口。原则上，其他功能单元只应调用 `user.h` 中声明的公开函数，不应绕过这层接口去依赖内部实现；这通常需要通过实际代码和工程约定共同约束。

### 2. Go 使用 package 组织代码

Go 使用 `package` 组织代码，不需要头文件。它通过 `package` 和标识符首字母的大小写来规定可见性。

对应前面的 C 项目，在 Go 中通常让一个功能块对应一个 `package`。同一目录下的 `.go` 文件共享同一个包，因此通常可以简单理解为“一个目录就是一个包”。除测试包等特殊情况外，同一目录下的所有 `.go` 文件必须属于同一个包。

我们可以复用前面的 `user`、`order` 例子：

```text
project/
  go.mod              // module example.com/project
  main.go             // package main

  user/
    user.go           // package user：对外公开 API，导出标识符首字母大写
    register.go       // package user：注册逻辑
    login.go          // package user：登录逻辑
    store.go          // package user：存储逻辑

  order/
    order.go          // package order：对外公开 API，导出标识符首字母大写
    create.go         // package order：创建订单
    cancel.go         // package order：取消订单
```

这个项目中的对应关系可以这样理解：

- `user/` 下所有 `.go` 文件都声明 `package user`，`order/` 下所有 `.go` 文件都声明 `package order`。
- 其他包通过 `import "example.com/project/user"` 使用 `user.Register`、`user.Login` 等；只有首字母大写的标识符才能被外部包访问。
- 小写标识符只在同一个包内可见。因此，C 中 `user_internal.h` 的“模块内部共享”作用，在 Go 中通常由同包内的小写标识符承担，不需要单独的头文件。
- C 中 `user.h` 承担的“对外公共接口”作用，在 Go 中通常由同包内首字母大写的声明承担；可以把它们集中放在 `user.go`，但这不是语言强制。
- `main.go` 属于 `package main`，它通过 `import` 使用其他包。
- 如果确实需要限制某些包只能被特定目录树导入，可以使用 Go 的 `internal/` 目录机制，而不是 `_internal.h` 这类头文件。

### 3. Go module：标记项目边界并管理依赖

`package` 和 `module` 解决的问题不同：

- `package` 是代码的组织和编译单位。一个项目通常包含多个 package。
- `module` 是一个 Go 项目的依赖管理单位，通常对应一个包含 `go.mod` 文件的项目目录。
- 一个 module 可以包含多个 package，而 package 的完整导入路径通常由“module 路径 + package 所在子目录”组成。

请注意，module 不是要求我们额外创建一个名为 `module` 的目录。通常只需要进入项目根目录，执行：

```bash
go mod init example.com/hello
```

这条命令会在当前目录生成如下 `go.mod` 文件：

```go
module example.com/hello

go 1.22
```

其中：

- `module example.com/hello` 声明当前 module 的路径。这里使用的是示例路径；如果项目将来发布到 GitHub，通常可以写成 `github.com/用户名/仓库名`。
- `go 1.22` 声明这个 module 使用的 Go 语言版本基线。

在实际的编码过程中，一般不需要你手动创建 module，本身ide 已经将这个过程大大简化了，我们在这里去花时间讲解 module 的原因是module 才是一个 go 项目的边界，而 package 只是组织单位

下面用一个完整的小项目看看 module 路径和 package 导入路径如何对应：

```text
hello/
  go.mod
  main.go
  calc/
    add.go
```

`calc/add.go`：

```go
package calc

func Add(a, b int) int {
    return a + b
}

func helper(x int) int {
    return x * 2
}
```

`main.go`：

```go
package main

import (
    "fmt"

    "example.com/hello/calc"
)

func main() {
    fmt.Println(calc.Add(1, 2))
}
```

`go.mod` 声明的 module 路径是 `example.com/hello`，而 `calc` 包位于项目的 `calc/` 子目录，所以它的完整导入路径就是 `example.com/hello/calc`。

在 `hello/` 项目根目录运行：

```bash
go run .
```

或者编译：

```bash
go build
```

这里的 `.` 表示当前目录中的包。`go run .` 会编译并运行当前的 `main` 包，`go build` 则会编译项目。

项目中使用的包大致可以分为三类：

- 标准库：Go 自带，例如 `fmt`、`net/http`。
- 第三方库：由 `go.mod` 管理，例如 Gin。
- 自己写的包：放在项目目录中，并通过完整导入路径引用，例如 `example.com/hello/calc`。

对于第三方依赖，可以使用 `go get` 把依赖加入当前 module。以 Gin 框架为例：

```bash
go get github.com/gin-gonic/gin
```

随后便可以在 Go 文件中通过 `import` 使用它：

```go
import (
	"net/http" // Go 标准库自带的包

	"github.com/gin-gonic/gin" // 刚刚加入的第三方包
)
```

执行 `go get` 后，依赖及其版本会记录到 `go.mod` 中，同时项目通常还会生成 `go.sum`。`go.sum` 用来记录依赖内容的校验信息，帮助 Go 确认下载到的依赖内容一致，因此它也应该和 `go.mod` 一起保留。

当代码中的依赖发生变化时，还可以运行：

```bash
go mod tidy
```

它会根据代码中的 `import` 补充缺少的依赖，并清理已经不再使用的依赖。初学阶段先记住：`go mod init` 用来初始化项目，`go get` 用来加入依赖，`go mod tidy` 用来整理依赖。

### 4. C 和 Go 的项目组织对比

| 对比项 | C | Go |
|---|---|---|
| 组织单位 | 文件、头文件、库 | package 组织代码，module 管理项目与依赖 |
| 接口声明 | `.h` 头文件 | 不需要，直接看导出标识符 |
| 实现 | `.c` 文件 | `.go` 文件 |
| 引入别人 | `#include` | `import` |
| 公开方式 | 头文件里声明 | 标识符首字母大写 |
| 私有方式 | `static` 文件内可见 | 标识符首字母小写，包内可见 |
| 库 | 静态库 `.a`、动态库 `.so/.dll` | 包、模块，通常静态链接 |
| 编译单位 | 单个 `.c` 文件 | 一个包 |
| 依赖管理 | Makefile、CMake、手动链接 | `go.mod`、`go.sum` |
| 初始化 | 手动调用 | 包级变量初始化、`init()` |
| 循环依赖 | 链接器可能报错 | 语言层面禁止循环导入 |

新手可以这样记：

- C 常用头文件声明接口，用源文件实现功能，再通过编译和链接把它们组合起来。
- Go 用 package 划分代码功能，用名称的首字母大小写决定是否对外可见。
- Go module 以 `go.mod` 标记项目边界，并记录项目需要的第三方依赖。

## 第四部分：Go 的包规范

前面已经从整体上理解了 package 和 module。下面再集中整理 Go 对包的具体规定和常用约定。

### 1. 一个目录一个包

同一目录下的所有 `.go` 文件必须使用同一个包名；测试文件还可以使用 `包名_test`。

正确：

```text
calc/
  add.go
  sub.go
```

```go
// add.go
package calc
```

```go
// sub.go
package calc
```

错误：

```go
// add.go
package calc
```

```go
// sub.go
package math  // 同一目录不同包名，编译报错
```

### 2. 包名规范

包名应该：

- 使用小写字母。
- 简短并且有意义。
- 不使用下划线或驼峰命名。
- 避免使用 `util`、`common` 这类含义过于宽泛的名字。

推荐：

```go
package user
package order
package calc
```

不推荐：

```go
package User
package user_service
package common
```

### 3. 导入路径和包名

导入路径由 module 路径和包所在的目录决定，代码中使用的包名则由文件开头的 `package` 声明决定。

目录：

```text
example.com/hello/calc
```

文件：

```go
package calc
```

导入：

```go
import "example.com/hello/calc"
```

使用时：

```go
calc.Add(1, 2)
```

包名通常与目录名一致，但语言没有强制要求。初学阶段建议让二者保持一致，避免混淆。

### 4. `main` 包

`package main` 是一个特殊的包，表示这里要构建的是可执行程序，而不是供其他代码使用的普通库。

要运行这个程序，包中必须有一个：

```go
func main() {
}
```

`main` 包不能被其他包导入。

### 5. `internal` 包

`internal` 是 Go 的特殊目录名。

```text
hello/
  internal/
    secret/
      secret.go
```

`internal` 下面的包，只能被 `internal` 的父目录所对应的目录树中的代码导入。以上面的结构为例，`secret` 可以被 `hello/` 及其子目录中的代码导入，却不能被 `hello/` 目录之外的代码导入。它适合存放不希望对外暴露的代码。

新手先了解，不用急着用。

### 6. 循环导入禁止

如果 A 包导入 B 包，而 B 包又导入 A 包，就形成了循环导入，编译器会直接报错。

```text
package a imports package b
package b imports package a
```

解决办法是抽出一个公共包，或者重新设计包之间的依赖方向。

## 第五部分：Go 的可见性与不可见性

Go 没有 `public`、`private` 关键字，而是使用一条统一的规则控制标识符的可见性：

> 标识符的首字母是不是大写。

- 首字母大写：标识符是“导出的”，其他包可以访问。
- 首字母小写：标识符是“未导出的”，只有同一个包内可以访问。

这个规则适用于：

- 变量
- 常量
- 函数
- 类型
- 结构体字段
- 方法
- 接口方法

### 1. 函数可见性

```go
package calc

func Add(a, b int) int {  // 大写 A，导出
    return a + b
}

func helper(x int) int {  // 小写 h，包内可见
    return x * 2
}
```

在其他包中：

```go
calc.Add(1, 2)    // 可以
calc.helper(3)    // 错误，helper 不可见
```

### 2. 变量和常量

```go
package config

var Version = "1.0"   // 导出
var debug = true      // 未导出

const MaxSize = 100   // 导出
const minSize = 1     // 未导出
```

其他包只能访问 `Version` 和 `MaxSize`。

### 3. 类型和结构体字段

```go
package user

type User struct {
    Name string // 导出字段
    age  int    // 未导出字段
}
```

在其他包中：

```go
u := user.User{Name: "小明"} // 可以设置 Name
fmt.Println(u.Name)          // 可以读取 Name
fmt.Println(u.age)           // 错误，age 不可见
```

注意：

- 结构体类型名 `User` 大写，所以外部能创建。
- 字段 `Name` 大写，外部能访问。
- 字段 `age` 小写，外部不能访问。
- 但包内可以访问 `age`。

### 4. 方法可见性

方法同样遵循首字母大小写规则：

```go
package user

func (u *User) Age() int { // 大写 A，导出
    return u.age
}

func (u *User) grow() { // 小写 g，包内可见
    u.age++
}
```

在其他包中：

```go
u.Age()   // 可以
u.grow()  // 错误
```

### 5. 包内可见，而不是文件内可见

C 的 `static` 将可见范围限制在当前文件内。

Go 的小写标识符则是包内可见。

也就是说，同一个包的不同文件，可以互相访问小写标识符。

`calc/add.go`：

```go
package calc

func helper(x int) int {
    return x * 2
}
```

`calc/use.go`：

```go
package calc

func UseHelper(x int) int {
    return helper(x)  // 可以，同一个包
}
```

### 6. 一个完整例子

目录：

```text
hello/
  go.mod
  main.go
  user/
    user.go
```

`user/user.go`：

```go
package user

type User struct {
    Name string
    age  int
}

func New(name string, age int) *User {
    return &User{Name: name, age: age}
}

func (u *User) Age() int {
    return u.age
}

func (u *User) grow() {
    u.age++
}
```

`main.go`：

```go
package main

import (
    "fmt"

    "example.com/hello/user"
)

func main() {
    u := user.New("小明", 18)

    fmt.Println(u.Name)  // 可以
    fmt.Println(u.Age()) // 可以

    // fmt.Println(u.age) // 错误
    // u.grow()           // 错误
}
```

### 7. 可见性速查表

| 标识符 | 首字母 | 可见范围 |
|---|---|---|
| `Add` | 大写 | 所有包 |
| `add` | 小写 | 当前包 |
| `User` | 大写 | 所有包 |
| `user` | 小写 | 当前包 |
| `Name` 字段 | 大写 | 所有包可访问 |
| `age` 字段 | 小写 | 当前包可访问 |
| `Age()` 方法 | 大写 | 所有包可调用 |
| `grow()` 方法 | 小写 | 当前包可调用 |

### 8. 常见错误

- 想从别的包调用小写函数，报 `cannot refer to unexported name`。
- 结构体字段小写，外部无法直接读写。
- 包名用大写或下划线，不符合规范。
- 同目录下 `.go` 文件写了不同 `package`。
- `main` 包被其他包导入。
- 循环导入。

## 第六部分：练习 Go 语言

在 go 语言的学习阶段，我们会在 github classroom 中进行习题发送来帮助你们进行练习

//todo 进行练习仓库的建立和这部分的编写
//todo 这部分应该先从实际的生产问题出发，引出 git 的必要性，最后进行 github 进行讲解，应该在课堂上实际演示 fork -> pull -> push 的三个步骤，应该会做对应的排名网站。
