# 第三节课-gin hertz 计网

## 通信协议

通信协议是设备/系统间数据传输的"共同规则集"，规定了数据的格式、传输时序、纠错方式等，就像两人对话要统一语言和语法，确保信息能准确互通。

## TCP/IP

TCP/IP 是互联网的核心通信协议簇，它定义了设备如何在网络中传输数据、寻址和路由。简单来说，IP地址用于在网络找设备，类似于现实生活中的门牌号。

## 公网和内网

* **公网 IP**：互联网上唯一的"身份证"，全球可访问（比如百度服务器的 IP）；
* **私网 IP**：局域网内专用的 IP（比如家里 Wi-Fi 下手机、电脑的 IP，通常是 192.168.x.x、10.x.x.x、172.16.x.x-172.31.x.x），无法直接被互联网访问，且不同局域网可重复使用。

由于 IPv4 设计的原因，现在的 IP 地址极其紧缺，现在大多数运营商使用 NAT 来中继 IP。

NAT 有以下优点：

* 节省公网 IP 资源：全世界公网 IP 数量有限，但一个家庭/公司只需 1 个公网 IP，内部所有设备（手机、电脑、智能家电）用私网 IP 即可上网，极大降低了公网 IP 的消耗；
* 隔离内网，提升安全性：互联网上的设备无法直接访问内网的私网 IP（只能看到路由器的公网 IP），相当于给内网加了一道"防火墙"，减少了被攻击的风险；
* 简化内网配置：内网设备无需手动设置公网 IP，只需通过路由器自动获取私网 IP（DHCP），就能直接上网，配置更简单。

以下是 NAT 网络简单拓扑图：

 ![](uploads/115597c2-8959-4e30-9c07-03f756946100/28a614b3-098d-4944-9b09-a71f24c20cb7/ChatGPT%20Image%202026%E5%B9%B48%E6%9C%886%E6%97%A5%2011_51_02.png " =1402x546")


如图，不同私网（局域网）设备之间不能直接通信，需要 NAT 进行中转。但是同一个局域网中的设备可以使用局域网 IP 直接互相访问。

* *校园网就是一个局域网*
* *127.0.0.1（localhost）为本地 IP，只能本地访问，不能通过网络访问*

## 端口

在 TCP/IP 协议中，端口（Port）是"IP 地址的延伸"，用于区分同一设备上的不同网络服务（比如一台服务器的 80 端口对应网页服务，443 端口对应加密网页服务）。如果说 IP 是门牌号，那么端口就是房间号。通过端口号可以定位到服务器的具体服务（应用）上。

常用的端口如下：

| 端口号 | 对应服务/应用 | 说明  |
|-----|---------|-----|
| 80  | HTTP    | 超文本传输协议，普通网页访问（比如打开百度默认用这个端口） |
| 443 | HTTPS   | 加密版 HTTP（HTTPS），网银、购物、微信网页等安全场景必用 |
| 21  | FTP     | 文件传输协议，用于服务器和本地之间上传/下载文件 |
| 22  | SSH     | 安全外壳协议，远程登录服务器（比如 Linux 服务器、路由器管理） |
| 23  | Telnet  | 老式远程登录协议（无加密，现在基本被 SSH 替代） |
| 53  | DNS     | 域名解析协议（比如输入 [www.baidu.com](http://www.baidu.com) 时，通过 53 端口查询对应 IP） |

## DNS

DNS（Domain Name System，域名系统）是 TCP/IP 协议簇中应用层的核心协议，本质是"域名 ↔ IP 地址的翻译系统"——解决了"人类记不住复杂 IP 地址，机器只认 IP 地址"的矛盾，是互联网正常运转的"基础设施"。

DNS 服务器用于将域名转化为 IP 地址，这样可以通过简单好记的网址来访问服务器，DNS 请求是隐式的，用户感知不到这个过程。

DNS 服务器分布全世界各地，且因为可以用很廉价的服务器来运行 DNS 服务，所以数量足够。大家可以理解为它就是个 go 语言中的 `map[string]string` 储存器，可以快速地通过 key（网址）来查询 value（IP）。（称之为 A 记录）

## HTTP

HTTP 协议是超文本传输协议（HyperText Transfer Protocol）的缩写，是客户端（如浏览器、手机 APP）与服务器之间传输网页、图片、数据等内容的"通用语言"，规定了双方如何请求数据和响应数据。

它的核心特点是"请求-响应"模式，且默认以明文形式传输（需搭配 SSL/TLS 加密形成 HTTPS）。

> *请求者发出一条信息，接受者收到信息发出一条回信，这就是"请求-响应"模式。*

### HTTP 方法

* **GET**：获取服务器资源。提交的数据有长度限制，且明文放在 URL 中。常用于请求数据，如查询所有商品。\n*在浏览器中直接输入网址访问用的就是 GET 方法。*
* **POST**：向服务器提交数据（创建资源）。请求参数放在请求体（Body）中（不可见，无长度限制）。常用于创建（如新增用户，需要获取账号密码等敏感信息）。
* **PUT**：全量更新服务器资源。需指定完整资源内容（覆盖式更新，缺少字段会被清空）。常用于更新用户信息（需传所有用户字段）、替换文件。
* **DELETE**：删除服务器资源。请求通常无请求体，和 GET 极为相似。常用于删除用户、删除订单。

### HTTP 请求

HTTP 请求由三部分组成：


1. **请求行**：用于标识方法、版本等信息，如使用 GET 方法，目标 URL 是 `"www.baidu.com/s"`，所用 HTTP 协议版本是 HTTP/1.1。
2. **请求头**：用于传递一些附加信息，如所使用的浏览器版本，接受的数据格式，鉴权有关信息等。
3. **请求体**：也就是常说的 body，用于 POST 方法存放发送给接收者的数据。

### HTTP 响应

HTTP 响应由三部分组成：

* **状态行**：包含协议版本、状态码、状态描述。
* **响应头**：用于标识服务器信息，响应体的数据格式（Content-Type）字段。
* **响应体**：存放实际返回的数据（如网页 HTML 代码、图片二进制数据、JSON 结果），发送者依靠响应头中的 Content-Type 来判断响应体中的内容是什么类型。

#### 状态码

状态码用于快速告诉请求者这次的请求状态如何：

* **1xx（信息类）**：临时响应，表示服务器已接收请求，需继续处理，实际开发中极少用到，如 `100 Continue`（告知客户端可继续发送请求体）。
* **2xx（成功类）**：请求正常处理完成，最常用的是 `200 OK`（请求成功，返回对应数据）和 `204 No Content`（请求成功，但无数据返回，如删除操作）。
* **3xx（重定向类）**：需客户端进一步操作才能完成请求，常见 `301 Moved Permanently`（资源永久迁移，如旧域名跳新域名）和 `302 Found`（资源临时迁移，临时跳转）。
* **4xx（客户端错误类）**：客户端请求有误，服务器无法处理，核心是 `400 Bad Request`（请求参数错误）、`401 Unauthorized`（未登录/身份验证失败）、`403 Forbidden`（登录后无权限访问）、`404 Not Found`（请求的资源不存在，如网址输错）。
* **5xx（服务器错误类）**：服务器处理请求时出错，常见 `500 Internal Server Error`（服务器未知错误，如代码 BUG）和 `503 Service Unavailable`（服务器暂时不可用，如维护中）。

#### 常见的 Content-Type 字段对应的数据类型

| Content-Type | 说明  |
|--------------|-----|
| `text/html`  | 返回网页内容，浏览器会按 HTML 规则解析渲染页面（如打开网页时的默认类型）。 |
| `application/json` | 返回 JSON 格式数据，是前后端 API 通信的主流类型（如 APP 获取列表数据、表单提交后的响应）。 |
| `application/x-www-form-urlencoded` | 表单数据默认提交格式，数据以"键=值&键=值"的字符串形式传输（如传统网页表单提交）。 |
| `multipart/form-data` | 用于上传文件或混合数据（如同时提交文字描述和图片/视频，常见于头像上传、文件上传功能）。 |
| `text/plain` | 返回纯文本数据，无特殊格式，仅作为普通文字展示（如 TXT 文件内容、简单的字符串响应）。 |
| `image/jpeg` / `image/png` | 分别对应 JPG 和 PNG 格式图片，服务器返回图片数据时使用（如网页加载图片、获取验证码图片）。 |

> 当返回 JSON 数据时，要设置响应头中的 Content-Type 为 `application/json`。

## JSON

JSON（JavaScript Object Notation，JavaScript 对象表示法）是一种轻量级的文本数据交换格式，核心作用是在不同系统（如前后端、不同服务）间高效传递结构化数据，因其易读、易解析、跨语言的特性，成为当前主流的数据交换标准。所以我们使用 HTTP 协议时，通常请求与返回的 body 里都遵循 JSON 格式，这是属于前端和后端的业务层面的通信协议。

## 路径（path）

我们打开最常用的百度搜索"搜索 helloworld"，浏览器最上方会显示这样的字符串：

`https://www.baidu.com/s?wd=helloworld`

这是什么？这其实就是我们常说的网址，网络的地址，也叫 URL。

网址（Uniform Resource Locator，简称 URL，统一资源定位符），本质是互联网上资源的"唯一地址"，通过它能精准找到网页、图片、接口、文件等各类网络资源，相当于网络世界的"门牌号"。

我们来拆分一下，看看它是如何运作的：

* 前面的 `"https://"` 就是我们前面讲的 HTTP 协议的加密版。
* 我们来看后面的 `"www.baidu.com"`，这是百度的基础网址（域名），也叫 base URL。
* 根据 `www.baidu.com` 定位到了服务器的地址，那么如何知道要调用哪个函数呢？\n`/s` —— baseurl 后面跟的斜杠加一串字符串，就告诉服务器请求此 URL 时具体该用哪个函数处理。这就是路径，也称作路由（后端意义上的路由）。\n网站后端的"软件路由"：是服务器上的一套"路径匹配规则"，比如后端会配置"当收到 `/s` 路径的请求时，调用搜索相关的代码逻辑；当收到 `/index` 路径的请求时，调用首页相关的代码逻辑"。

## 搭建你的第一台 web 服务器

为了方便搭建 web 服务器，也就是广义上的后端，框架工程师创建了很多框架供我们使用，框架就是把很麻烦的事情变得很容易的一个工具。我们先讲一下基于 Go 的 web 服务器最常用的框架之一 —— Gin。关于如何方便地处理依赖，请使用 `go mod tidy`，具体用法以前讲过，这里就不多赘述，我们默认你安装了 Gin 相关的依赖。

### Gin

以下是一个用 Gin 实现的简单的 ping pong 示例：

```go
package main

import "github.com/gin-gonic/gin"

func main() {
    r := gin.Default()
    r.GET("/", ping)
    r.Run()
}

func ping(c *gin.Context) {
    c.JSON(200, gin.H{
        "message": "pong!",
    })
}
```

* `gin.Default()` 用于创建 Gin 引擎实例，会帮你配置好一些默认的配置，初学者不必关心内部具体实现，只需知道我们使用它来创建一个具有很多方法的结构体。`r` 就是这个结构体的实例。


* `r.GET()` 方法用于注册 GET 路由。因为我们写的 web 服务器是基于 HTTP 协议的，所以有以下几种常用方法：
  * `r.GET()` 将函数映射到路径（接口）
  * `r.POST()` 如上
  * `r.DELETE()` 如上
  * `r.PUT()` 如上
  * `r.Static()` 将本地文件夹映射到路由
  * `r.StaticFile()` 将本地文件映射到路由

> **注意**：同一个路径不同的方法会被认为是不同的路由！用 GET 方法请求 `baseurl/user` 和用 POST 方法请求 `baseurl/user` 会使用完全不同的函数！

`r.GET()` 的第一个参数是注册的路由，是字符串类型，第二个参数是一个符合特定函数签名的函数。（当一个函数的传入值的顺序、类型与规定的一样，返回值的顺序、类型也和规定的一样，我们就认为它符合特定的函数签名。）这个方法其实就是把这个路由与函数连起来，告诉服务器要用这个函数处理这个路由收到的数据。

路由函数签名：`func 函数名(变量名 *gin.Context) {}`，这就是路由要绑定的函数的函数签名，gin 框架规定了这样的形式。

映射完路由后使用 `r.Run()` 开始运行后端服务。`r.Run()` 默认将路由部署在本地（localhost）的 8080 端口上。

> `r.Run()` 会堵塞后续函数运行！相当于内部有一个无限循环，所以要在最后执行。

注意到百度搜索的网址后面还有一串奇怪的字符：`/s?wd=helloworld`。在网络中，大多数符号的作用是用来分割，在 HTTP 请求中，`?` 用来在 URL 中分割请求数据与接口。`s` 代表接口名称，后面跟着的是请求数据。

这下好理解为什么用 POST 注册和登录了。如果用 GET 注册，所有的数据都要写在 URL 里，而这些都会被写入浏览器历史记录，你也不想别人随便一翻浏览器历史就知道你账号密码吧？

### 中间件

中间件是"在请求到达处理器（Handler）之前/之后执行的函数"，用于统一处理通用逻辑，避免在每个 Handler 中重复写代码。常见使用场景：

* 日志记录（记录请求方法、路径、耗时）；
* 身份认证（验证 Token 是否有效，无效则直接返回）；
* 跨域处理（设置 CORS 响应头）；
* 请求限流（限制同一 IP 的访问频率）；
* 错误捕获（捕获 Handler 中的异常，返回统一错误格式）。

#### 注册全局中间件

```go
// 定义一个日志中间件：记录请求信息
func loggerMiddleware(c *gin.Context) {
    // 1. 请求到达前的逻辑（前半部分）
    startTime := time.Now()   // 记录请求开始时间
    method := c.Request.Method // 获取请求方法（GET/POST）
    path := c.Request.URL.Path // 获取请求路径

    // 2. 执行后续的中间件/Handler（必须调用，否则流程中断）
    c.Next()

    // 3. Handler 执行后的逻辑（后半部分）
    costTime := time.Since(startTime) // 计算请求耗时
    statusCode := c.Writer.Status()   // 获取响应状态码
    fmt.Printf("[%s] %s → 耗时：%v，状态码：%d\n", method, path, costTime, statusCode)
}

func main() {
    r := gin.Default() // gin.Default() 已内置 2 个全局中间件：日志（Logger）和恢复（Recovery，捕获 panic）

    // 注册自定义全局中间件（所有路由都会经过）
    r.Use(loggerMiddleware)

    // 普通路由（自动注册全局中间件）
    r.GET("/hello", func(c *gin.Context) {
        c.JSON(200, gin.H{"msg": "hello gin"})
    })

    // 单个路由使用（在使用全局中间件后再使用单独给它的中间件）
    r.GET("/hello2", hello2, loggerMiddleware2) // 这里假设 hello2 是另一个路由函数，loggerMiddleware2 是另一个中间件

    r.Run()
}
```

### 路由组

路由组本质是"带有公共前缀的一组路由集合"，用于将同一模块的路由归类管理（比如"用户模块"的所有路由都以 `/user` 为前缀，"订单模块"以 `/order` 为前缀），避免重复写公共路径，同时方便统一配置（如给整个模块加中间件）。

```go
package main

import "github.com/gin-gonic/gin"

func main() {
    // 1. 创建默认的 Gin 引擎（包含默认中间件）
    r := gin.Default()

    // 2. 创建路由组：公共前缀 /user（用户模块）
    userGroup := r.Group("/user")
    // 可以通过给路由组注册中间件实现对路由批量注册中间件
    userGroup.Use(middleware) // 这里假设 middleware 是一个中间件
    {
        // 实际访问路径：/user/info（GET 请求）
        userGroup.GET("/info", func(c *gin.Context) {
            c.JSON(200, gin.H{"msg": "获取用户信息"})
        })
        // 实际访问路径：/user/login（POST 请求）
        userGroup.POST("/login", func(c *gin.Context) {
            c.JSON(200, gin.H{"msg": "用户登录成功"})
        })
        // 实际访问路径：/user/logout（GET 请求）
        userGroup.GET("/logout", func(c *gin.Context) {
            c.JSON(200, gin.H{"msg": "用户退出登录"})
        })
    }

    // 3. 再创建路由组：公共前缀 /order（订单模块）
    orderGroup := r.Group("/order")
    {
        // 实际访问路径：/order/detail（GET 请求）
        orderGroup.GET("/detail", func(c *gin.Context) {
            c.JSON(200, gin.H{"msg": "获取订单详情"})
        })
        // 实际访问路径：/order/create（POST 请求）
        orderGroup.POST("/create", func(c *gin.Context) {
            c.JSON(200, gin.H{"msg": "创建订单成功"})
        })
    }

    // 启动服务（默认 8080 端口）
    r.Run()
}
```

### Gin 的请求上下文

Gin 会在接收到请求时，把一些杂七杂八的东西都丢到 `gin.Context` 里面，所以想要提取请求传递的信息，我们要使用 `gin.Context` 的方法。

以下是一些常见的用于将请求的数据读取的方法（我们假设变量 `c` 就是那个传入的 `gin.Context`）：

* `c.Query(key string) string`：获取 URL 中的查询参数（`?key=value`）。若 URL 是 `?name=test`，则 `c.Query("name")` 返回 `"test"`。
* `c.Param(key string) string`：获取路由中的路径参数（`:key` 定义的动态路由）。路由 `GET /user/:id` 匹配 `user/123` 时，`c.Param("id")` 返回 `"123"`。
* `c.ShouldBind(obj interface{}) error`：根据请求头中的 Content-Type 来自动解析请求体中的数据到结构体实例。
* `c.ShouldBindJSON(obj interface{}) error`：解析请求体中的 JSON 数据到结构体实例。例如绑定登录参数：`c.ShouldBindJSON(&loginReq)`（需加入结构体 json 标签）。
* `c.GetHeader(key string) string`：获取请求头中的信息。

因为 JSON 是键值对，所以多用 map 和结构体绑定或生成 JSON。

#### 示例：使用 `c.ShouldBindJSON()` 处理请求体数据

请求体：

```json
{"data":{"name":"bob","age":11}}
```

```go
package main

import (
    "log"
    "github.com/gin-gonic/gin"
)

// 根据接口文档约定好的形式设计结构体
type student struct {
    Name string `json:"name"`
    Age  int    `json:"age"`
}

func main() {
    r := gin.Default()
    r.GET("/", printStudent)
    r.Run()
}

func printStudent(c *gin.Context) {
    // 初始化一个 student 结构体实例
    var s student
    // 传入方法，绑定结构体
    err := c.ShouldBindJSON(&s)
    if err != nil {
        log.Println("请求参数错误：", err)
        return
    }
    log.Println(s)
}
```

更多有关 Go 语言处理 JSON 的方法，大家可以去看一下群文件。

### 回复

接收到信息后，我们如何回复呢？还是用 `gin.Context`（假设 `c` 是那个 `gin.Context`）。

常用 `c.JSON()` 快速回复：

`c.JSON(code int, obj interface{})` 返回 JSON 格式响应，且会在响应头自动配置 Content-Type 字段。

> 使用 `c.JSON()` 会发出响应报文，但并不会终止函数运行，往往在其后加入 `return` 来防止继续执行后续逻辑。

其中第一个传参是状态码，常用 `http` 包中的常量代替，如 `http.StatusOK` 常量值是 200。

示例：

```go
type user struct {
    Name string `json:"name"`
    Age  int    `json:"age"`
}
u := user{
    Name: "bob",
    Age:  11,
}
```

成功响应：

```go
c.JSON(200, gin.H{
    "msg":  "success",
    "data": u,
})
```

生成以下 JSON：

```json
{"msg":"success","data":{"name":"bob","age":11}}
```

失败响应：

```go
c.JSON(400, gin.H{
    "msg": "参数错误",
})
```

生成以下 JSON：

```json
{"msg":"参数错误"}
```

> 在编程中尽量少出现突兀的数字，最好用常量来标识，这样后人可以通过常量名来快速理解数字的意义。 `gin.H` 是 `map[string]interface{}` 的一个别名，用于快捷生成 body 里的 JSON 数据。 在 web 编程中，我们常在 JSON 中使用 `status` 字段来标识业务层面的状态码，使用 `data` 字段来传递具体数据。

## Hertz

Hertz 是字节跳动（研发抖音的那个）开源的高性能 HTTP 框架，专为 Go 语言设计。Hertz 框架在性能方面有很大的优势，且在字节跳动内部应用广泛。但由于资料比较少，相较于 Gin 来说更难上手，且其与 Gin 高度相似，所以建议先学 Gin 框架，再逐渐转为 Hertz。

以下是一段 Hertz 的路由函数的简单示例：

```go
func CreateUser(c context.Context, ctx *app.RequestContext) {
    // 定义请求体结构体
    type CreateUserReq struct {
        Name  string `json:"name"`
        Age   int    `json:"age"`
        Email string `json:"email"`
    }

    // 初始化请求体变量
    var req CreateUserReq

    // 使用 BindJSON 绑定 JSON 请求体（仅解析 application/json 格式）
    // 绑定失败场景：JSON 格式错误、请求头 Content-Type 不是 application/json 等
    if err := ctx.BindJSON(&req); err != nil {
        ctx.JSON(http.StatusBadRequest, map[string]interface{}{
            "code": -1,
            "msg":  fmt.Sprintf("JSON绑定失败：%v（请检查JSON格式或Content-Type）", err),
        })
        return
    }

    ctx.JSON(consts.StatusCreated, map[string]interface{}{
        "code": 0,
        "msg":  "用户创建成功",
        "data": map[string]interface{}{
            "id":    newUserID,
            "name":  req.Name,
            "age":   req.Age,
            "email": req.Email,
        },
    })
}
```

这里不过多介绍，感兴趣的同学可以去官网看看：<https://cloudwego.cn/zh/docs/hertz/>

## 接口测试

可以试试以下几款接口测试软件：

* Postman：<https://www.postman.com>
* Apifox：<https://apifox.com/>

## 作业

### lv1

实现以下功能：

* 当浏览器访问 `http://localhost/talk?msg=ping` 时，在 JSON 的 `data` 字段返回 `pong`。
* 当浏览器访问 `http://localhost/talk?msg=helloserver` 时，在 JSON 的 `data` 字段返回 `helloclient`。

### lv2

将群文件里的文件下载到项目目录，并将此映射为路由，使用浏览器访问 `http://localhost/cat.jpg`，成功后会显示图片。

### lv3

实现一个计算学生评论成绩的接口，要求使用结构体进行绑定请求体数据。

请求体 JSON 示例：

```json
{"name":"bob","score":[68,97.4,94.2,75.4]}
```

返回体 JSON 示例：

```json
{"average":83.75}
```


\