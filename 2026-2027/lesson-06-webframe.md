# web框架基础

Web框架是一类用于简化Web应用开发的工具集，而Web应用是指通过互联网或内部网络在浏览器中运行的应用程序，用户可以直接通过浏览器访问，无需在本地安装特定软件。

过往的框架工程师们搭了很多框架供我们使用，在go语言这一块有诸如Gin，Echo，Fiber，Chi，GoFrame，Hertz等框架。其中最常用的无疑是Gin框架，除此之外就是Echo，Fiber，Hertz等框架了——我们从中挑选Echo和Hertz来着重讲解下。

# C1web框架基础

每个框架都有其各自的内容，如中间件，路由，请求上下文等，完整阐述出来的内容实在太多，而篇幅实在有限——因此，我们会着重选其中最基础，最常见的方法进行讲解。

如对其它的内容有兴趣\(或者找不到想要的方法\)，可以查看这些框架的官方文档。

gin框架：https://gin\-gonic\.com/zh\-cn/docs/

echo框架：https://echo\.laily\.net/或者https://echo\.labstack\.com/docs

hertz框架：https://www\.cloudwego\.io/zh/docs/hertz/

## C1\.1名词解释

### C1\.1\.1基础名词

1. **客户端（Client）**，即发起网络请求的一端，是请求的源头。

谁来访问服务端，谁就是客户端。比如浏览器、Postman/Apifox、curl、另一个 Go 服务、手机APP都属于客户端。

客户端只负责发请求、收响应，不处理业务逻辑。

> 比如go原生的`net/http`的`http.Client`发起HTTP调用，这个对象就是客户端。
> 
> 



2. **服务端（Server）**是监听端口，接收客户端请求、执行业务、返回响应的程序。

这是等待别人来访问，提供服务的程序，我们写的Go Web应用就是服务端。

> go原生的`http.ListenAndServe(":8080", router)`，启动的就是HTTP服务端。
> 
> 



3. **请求（Request）**是指客户端发给服务端**HTTP报文**\(的行为\)，这个数据包包含了客户端想要表达的全部信息。 包含：请求方法（GET/POST）、URL 路径、请求头Header、Cookie、Query参数、Body请求体等等。

> 比如Go原生的`*http.Request`对象就是原生请求。
> 
> 



4. **响应（Response）**是指服务端处理完成后，返回给客户端的**HTTP报文**\(的行为\)。 包含：状态码（200/404/500）、响应头Header、响应体（JSON、HTML、文本）等等。

> 比如Go原生的`http.ResponseWriter`是原生写入响应的接口。
> 
> 



5. **路由（Route）**是一套**匹配规则**，用来根据请求路径和请求方法找到对应的处理函数。比如我们有POST方法的`/user/login`，其就可以根据"POST方法"和"/user/login"，将请求交给login处理函数。



6. **Handler（处理器 / 处理函数）**是路由匹配成功后，**真正执行业务代码**的函数。

> 比如go原生的`func(w http.ResponseWriter, r *http.Request)`
> 
> 



7. **请求上下文（Context）**有两种，一个是go标准库自带的context，一个是框架封装的请求上下文。

- go标准库的**context\.Context**是携带了请求生命周期内的**元信息、超时、取消信号**的对象，其也可以用于管理协程。

- 框架封装的请求上下文，比如**gin\.Context**，是框架对`http.Request、http.ResponseWriter`的**封装对象**，其还通过`Request.Context()`暴露了标准`context.Context`的能力，并且聚合了本次请求所有相关资源，用于给Handler和中间件使用。



8. **web请求上下文**：我们知道，context是聚合了请求的所有相关资源的，用于给Handler和中间件使用——要获取这些资源，自然就需要采取某些方法。

- ①**路径参数Params**是URL路径的一部分，用于指定特定的资源，可以让我们直接从url中获得一些值。平常我们经常能看到某个网站的url里包含有某些信息，比如https://www\.bilibili\.com/video/BV164421Z73z和https://www\.bilibili\.com/video/BV1AGam6pEVh，url的末尾挂的有视频的bv号。

    如果这些信息都是硬写在上面的，那b站几百万几千万的视频，岂不是都要开发者一个个硬写？这不可能。

    为此我们可以采取**Params**参数，只需要以**`:key`**的形式指定url，就可以在后端动态获取一些信息——这就是**单段路径参数**。比如我们注册的路由是`/user/:id`，那么只需要访问`/user/123`，后端就会自动收到"123"

    除此之外，我们还能用星号来匹配某个前缀之后的所有内容，形式为`*key`，甚至连斜杠也能获取——这就是**剩余多段通配参数**。比如我们注册的路由是`/user/:id/*action`，那么只需要访问`/user/114514/send/`，后端就会自动收到"/send/"

- ②**查询字符串参数query**的用途和params类似，也是用于指定特定的url资源的。比如我们平常用百度的时候，能观察到url参数里包含有我们搜索目标的信息，比如https://cn\.bing\.com/search?q=baidu，直接就能通过`search?q=`来搜索"baidu"的相关信息。

    我们可以以**`?key=value`**的形式来指定query参数，有多个key可以用`&`符号隔开，比如**`?key1=value1&key2=value2`**，就可以同时指定两个query参数。

    比如前端调用的链接是`?name=test`，那么后端就可以凭借"name"这个key来得到"test"。

- ③**请求体body**是请求的主要传输部分，有很多种格式，比如表单格式，二进制格式，原始内容格式\(包括JSON，HTML等\)，GraphQL格式等等。请求体在直接的浏览器或者url中是不可见的，需要后端去手动解析。

- ④**请求头Header**是客户端在发送请求时附加的键值对信息，可以包含用户的鉴权信息，设备信息，请求体使用的格式等等。



9. **中间件（Middleware）**中间件是请求处理链中，在业务Handler前后执行的函数。在另一些框架中，中间件本质就是包装Handler的**高阶函数**，接收一个Handler，返回新的Handler。

    中间件可以**拦截**、**预处理**、**后处理**请求，可以用于鉴权认证，记录日志，拦截以及其它各种各样的小需求，最终形成链式调用。

    假设张三向李四发送了一个请求，有A和B两个中间件，那么该请求的流向就是：张三送出请求→A→B→业务Handler→B后置逻辑→A后置逻辑→李四返回响应。



### C1\.1\.2其它

在实际使用web框架时我们会发现貌似使用的都是Http协议，然而现在大多数网站地址的开头都是https——这是怎么回事？在这里我们就需要了解一下TLS和HTTPS了。

1. **TLS（传输层安全协议）**是在TCP之上对网络传输数据**加密、身份认证和完整性校验**的协议。HTTPS协议也就是HTTP协议\+TLS。通过TLS，我们得以给HTTP架设加密隧道，使得用户的密码等数据信息可以安全传输。

证书则用于证明服务端身份并携带公钥，其它内容感兴趣可自行查找。

> Go原生带有`http.ListenAndServeTLS()`来开启HTTPS协议。
> 
> 

2. **Hooks（钩子函数）**是框架预留的**回调入口**，其不同于中间件，Hooks是在框架生命周期特定时机自动执行我们注册的自定义函数。\(注意，某些框架中也会有请求级的Hook\)

中间件是根据请求的生命周期触发的，而Hooks是根据框架/服务的生命周期触发的，其常用于在服务启动/关闭前进行某些初始化操作，比如加载配置，释放资源等等。



3. **优雅关闭（Graceful Shutdown）**是指收到web服务的关闭信号后，不再接收新请求，而是等待正在处理的请求执行完毕再退出——它不会强行杀死正在处理的连接，因而可以减少数据丢失等问题。



4. **静态文件（Static Files）**指在服务器上以**物理**文件形式存在、内容不会随用户或请求变化的资源。

    有的时候，我们需要储存用户的头像或者是用户的主页封面，用户的作品等等。这时候肯定不能直接把图片或者文件转成base64字符串来储存——这时可能要将这些文件以物理形式储存在服务器中，也就是**静态文件**。用户只需要访问使用了静态文件方法的url，就可以直接读取服务器中的这些文件了。

    实现静态文件的访问技术多种多样，除了直接储存在服务器中以外，还有CDN，OSS以及反向代理等等，若感兴趣可自行查看相关文档。



## C1\.2Gin框架

Gin是一个用Go\(Golang\)编写的Web框架，十分甚至九分的好用。我们可以在终端，cmd之类的场景使用`go get ``github.com/gin-gonic/gin`命令来导入gin框架。

### C1\.2\.1快速开始

我们需要通过gin\.Default方法来新建一个**路由引擎**，随后便可以通过路由引擎来注册路由了。

在这里，我们用GET方法来注册一个测试路由，路由引擎的HTTP方法的第一个参数需要传入url，第二个参数则传入**handler函数**参数。gin框架的**handler函数**只需要在参数里包含一个`*gin.Context`即可，不需要含其它的参数。

在handler中我们可以处理业务逻辑，比如这里就是用gin\.Context的**JSON\(\)**方法给客户端返回一个状态码200的JSON响应\(注意，这个响应**不会终止**函数的运行\)，而响应体用**gin\.H\{\}**表示为`"message":"pong"`。

这里的**gin\.H\{\}**是`map[string]interface{}`的一个命名类型，用于快速表述响应体的JSON字段。



最后我们调用**Run\(addr \.\.\.string\) \(err error\)**方法来启动路由引擎，参数不填的话服务会优先使用环境变量PORT的值，如果PORT没设置则默认监听8080端口，本机所有网络接口的8080端口都可以来访问web服务。

我们也可以在参数中手动指定监听的端口和地址。

为什么Run\(\)方法是放在最后呢？仔细想想，如果要让一个服务稳定运行，它就不可能在调用完所有函数后直接结束。因此，这个Run\(\)方法必须是阻塞的，让整个程序不能直接结束，也因此，它通常放在最后运行，不然在Run\(\)方法之后的代码全部只能在服务退出后才能进行了。

一旦我们运行main函数，我们就可以在浏览器访问[http://localhost:8080/ping](http://localhost:8080/ping)来看到我们的成果了：`{"message":"pong"}`。

```Go
func main() {
    r := gin.Default()
    r.GET("/ping", ping)
    r.Run(":8080")
}
func ping(c *gin.Context) {
    c.JSON(200, gin.H{"message": "pong"})
}
//这样也可以
//func main() {
//    r := gin.Default()
//    r.GET("/ping", func(c *gin.Context) {
//        c.JSON(200, gin.H{"message": "pong"})
//    })
//    r.Run()
//}
```



### C1\.2\.2HTTP方法

Gin提供了直接映射到HTTP动词的方法，我们可以通过这些方法来注册路由。

注意，同一个路径不同的方法会被认为是不同的路由！比如GET /api/v1/user方法可以理解成获取某个用户的信息，而POST /api/v1/user则可以理解成注册用户，PUT /api/v1/user是更新用户信息，DELETE /api/v1/user则是注销用户——这四个方法虽然是同一个url，却可以使用不同的handler。

|方法|一般用途|
|---|---|
|**GET**|获取资源|
|**POST**|创建新资源|
|**PUT**|替换现有资源|
|**PATCH**|部分更新现有资源|
|**DELETE**|删除资源|
|**HEAD**|与GET相同但不返回响应体|
|**OPTIONS**|描述通信选项|



### C1\.2\.3请求上下文

在C1\.1\.1中我们得知，web框架会把请求中的一些杂七杂八的东西都丢到context里面，在gin里面也就是丢到`gin.Context`里头。所以想要提取客户端传递的信息，我们就要使用`gin.Context`相关的方法

1. **c\.Param\(key string\)string**用于获取路由中的路径参数，其同时支持单段路径参数和剩余多段通配参数。

    比如我们注册的路由是`/user/:id`，那么只需要访问`/user/123`，后端的`c.Param("id")`就会收到"123"；再比如注册了`/user/:id/*action`，那么只需要访问`/user/114514/send/`，后端的`c.Param("action")`方法就会自动返回"/send/"

    注意:gin框架中的剩余多端通配参数请放在url末尾



2. **c\.Query\(key string\)string**用于获取URL中的查询参数。若URL是`?name=test`，则`c.Query("name")`返回"test"。如果没有会返回空字符串，当然，用**c\.DefaultQuery\(key, defaultValue\)**可以在query为空时返回一个默认值。

    如果需要解析一些重复key，比如name=1\&name=2这种情况，就可以用**c\.QueryArray\(key string\) \(values \[\]string\)**方法来解析。

    除此之外，我们常常会约定一个方括号表示法，一般会以`key[subkey]=value`的形式来表述一些不知道键名的键值对，这时需要采用**c\.QueryMap\(key string\)\(dicts map\[string\]string\)**来解析这些键值对。比如url上写的是`http://xxxx/xxx?ids[a]=1234`，那么使用`c.QueryMap("ids")`转化后，就能得到键为a，值为1234的映射值。\(不支持嵌套\)



3. 请求体有很多格式，比如表单格式，原始内容格式等等，因此其的方法也分为很多种。

用一些方法解析请求体的时候，需要给参数传入结构体指针来将请求体信息反序列化。而这些结构体的属性最好带有相关的标签来表明其格式，不然会出现无法匹配字段的问题——如果在跑代码的时候发现莫名其妙报错又难以找到原因，可以看看你是不是忘记给结构体加标签了。

```Go
type User struct {
    ID       string `json:"user_id"`
    NickName string `json:"nick_name"`
    Avatar   string `json:"avatar"`
    Phone    string `json:"phone"`
    Pwd      string `json:"-"`
}
```

- **c\.ShouldBindJSON\(obj interface\{\}\)error**这个方法会默认请求使用JSON格式，将请求体数据解析为结构体。该方法的参数需要传入一个结构体指针。

- **c\.PostForm\(key string\)\(value string\)**则是通过对应的键来解析表单格式的数据，只需要传入键即可。如果要返回默认值，可以采用**c\.DefaultPostForm\(key, defaultValue string\)string**方法。

- 和查询参数一样，请求体中有的时候也可能会有`key[subkey]=value`的形式，这时需要使用**c\.PostFormMap\(key string\) \(dicts map\[string\]string\)**，如果请求体的表单写的是ids\[a\]=1234，同样也是得到键为a，值为1234的映射值。

- 最后，如果想要上传文件，我们可以用**c\.FormFile\(name string\)\(\*multipart\.FileHeader, error\)**方法来从表单格式的`multipart/form-data`中接收文件，随后将`*multipart.FileHeader`和目标路径传入**c\.SaveUploadedFile\(\)**方法保存文件即可。如果是多文件可以用**c\.MultipartForm\(\)**方法，更多注意事项和详细信息可见https://gin\-gonic\.com/zh\-cn/docs/routing/upload\-file



4. 解析请求头的方法有多种：

- **ShouldBindHeader\(obj interface\{\}\) error**方法会根据结构体标签来把多个请求头信息绑定到结构体中。

```Go
type testHeader struct {
  Rate   int    `header:"Rate"`
  Domain string `header:"Domain"`
}
......
h := testHeader{}
err := c.ShouldBindHeader(&h)
```

- **c\.GetHeader\(key string\) string**方法会根据传入的key值获取单个请求头中的信息。



5. **Static\(relativePath, root string\) IRoutes**方法可以把对relativePath的请求映射到root下的静态文件，比如`router.Static("/assets", "./assets")`会在`http://xxxx/assets/style.css`处提供`./assets/style.css`文件。

- 除此之外，还有`router.StaticFS(relativePath, fs)`，`router.StaticFile(relativePath, filePath)`等方法可用，若感兴趣可自行查看相关文档。

- 实现静态文件的访问技术多种多样，其中主流的除了现在提到的直接储存在服务器中以外，还有CDN，OSS以及反向代理等等，若感兴趣可自行查看相关文档。



6. **c\.ShouldBind\(obj interface\{\}\)error**方法会根据请求方法以及请求头中的`Content-Type`等信息来自动解析请求中的数据到结构体实例，比如query参数，表单，JSON，XML等等。该方法的参数需要传入一个结构体指针。

    比如收到了一个GET请求，那么ShouldBind就知道这次的请求没有请求体，其就会把请求数据当query参数解析。

    如果是POST或者PUT之类的请求，那么ShouldBind就会知道这次的请求有请求体，如果同时请求头的`Content-Type`是"application/json"，那么Gin框架就会知道请求体用的是JSON格式，从而将JSON反序列化为go的结构体。

    





### C1\.2\.4响应上下文

在 C1\.2\.3 中，我们介绍了如何从`gin.Context`中取出请求信息；在 C1\.2\.4 中，我们就要反过来看看怎么把响应信息写回去。

Gin把响应写入能力封装在`gin.Context`的`Writer`字段里，这个**Writer**实现了`http.ResponseWriter`接口，并额外提供了一些方法。我们通常不直接操作`c.Writer`，而是使用`c.JSON()`、`c.String()`、`c.Data()`等更顺手的方法。

1. **c\.Status\(code int\)**用于设置HTTP状态码，不写入响应体。常用于`204 No Content`这类没有响应体的响应。



2. 既然请求体有多种格式，响应体自然也有多种格式。我们来看看其中常见的几种：

- **c\.JSON\(code int, obj interface\{\}\)**会把任意对象序列化为JSON，设置响应头为`Content-Type: application/json; charset=utf-8`，并写入状态码和响应体。obj可以是`gin.H`、结构体、map、slice 等。\(注意：`c.JSON()`不会终止 Handler，需要手动return\)

- **c\.Data\(code int, contentType string, data \[\]byte\)**会直接写入原始字节，需要自己指定 `Content-Type`。适合返回HTML片段、图片、二进制数据或自定义格式。

```Go
c.Data(
    http.StatusOK,
    "text/html; charset=utf-8",
    []byte("<h1>Hello Gin</h1>"),
)
```

- **c\.File\(filepath string\)**可以返回服务器上的文件，底层通常使用`http.ServeFile`，会自动推断`Content-Type`，并支持Range请求等能力。

    当然，如果要做文件下载，可以使用**c\.FileAttachment\(filepath, filename\)**方法，它会设置`Content-Disposition: attachment`。

```Go
c.File("./uploads/avatar.png")
c.FileAttachment("./uploads/report.pdf", "report.pdf")
```



3. **c\.Header\(key, value string\)**用于设置响应头。它必须在写响应体或写状态码之前调用，否则可能不会生效。



4. **c\.Redirect\(code int, location string\)**用于重定向，它会设置`Location`响应头并写入重定向状态码。常用状态码有 `301`、`302`、`303`、`307`、`308`等等。



除此之外，还有一些常用的响应方法诸如渲染HTML模板，流式响应，SSE等等，若感兴趣可自行去了解。



### C1\.2\.5中间件

1. **中间件**是在请求到达Handler之前/之后执行的函数，用于统一处理通用逻辑，避免在每个Handler中重复写代码。相当于一个中介。

    **中间件**的常见使用场景有：

    日志记录（记录请求方法、路径、耗时）；

    身份认证（验证Token是否有效，无效则直接返回）；

    跨域处理（设置CORS响应头）；

    请求限流（限制同一IP的访问频率）；

    错误捕获（捕获Handler中的异常，返回统一错误格式）。



2. gin自带有一些中间件，当然我们也可以自己指定中间件——其格式需要返回一个**gin\.HandlerFunc**函数，这一函数不需要返回值，并且需要传入`*gin.Context`。

自定义中间件需要用**c\.Next\(\)**方法来分隔调用阶段：

- **`c.Next()`**之前的代码在请求到达主处理函数之前运行。用于设置任务，如记录开始时间、验证令牌或使用`c.Set()`设置上下文值。\(这里的c\.Set\(\)设置的值可以在同一响应的别处调用\)

- **`c.Next()`**：在这一阶段，执行会暂停，中间件会将控制权传递给下一个中间件（或最终处理函数）——直到所有下游处理函数完成。如果不调用**c\.Next\(\)**，那么中间件不会等待下游执行完，也无法在下游执行完后执行后置逻辑，而gin会仍然执行后续处理器，这可能造成一些不可名状的bug。

- **`c.Next()`**之后的代码在主处理函数完成后运行。用于清理、记录响应状态或测量延迟。

如果出现认证失败或者需要阻止后续处理器之类的情况，可以调用**c\.Abort\(\)**方法来终止调用。如果使用**c\.AbortWithStatusJSON\(\)**方法，还能在终止调用时给前端返回状态码之类的。

> 注意，终止调用不等于代码逻辑终止了，为了避免剩余代码继续执行，请在c\.Abort\(\)后面接个return
> 
> 

```Go
func Logger() gin.HandlerFunc {
  return func(c *gin.Context) {
    t := time.Now()
    // Set example variable
    c.Set("example", "12345")
    // before request
    c.Next()
    // after request
    latency := time.Since(t)
    log.Print(latency)
    // access the status we are sending
    status := c.Writer.Status()
    log.Println(status)
  }
}
......
r.Use(Logger())
r.GET("/test", func(c *gin.Context) {
    example := c.MustGet("example").(string)
    // 会打印"12345"
    log.Println(example)
})
```



3. 我们在C1\.2\.1中使用了**gin\.Default\(\)**方法来创建路由引擎，实际上使用**gin\.New\(\)**方法也可以——但是区别在于，gin\.Default\(\)方法创建的路由引擎会自带有Logger和Recovery两个中间件，而gin\.New\(\)则是完全空白的，不自带任何中间件。

其中，**Logger**中间件将请求日志写入标准输出（方法、路径、状态码、延迟）。而**Recovery**中间件会从handler函数中的任何panic中恢复并返回500响应，以此防止服务器崩溃。



4. Gin支持三个级别的中间件：

- **全局中间件**应用于路由引擎中的每个路由。使用`router.Use()`注册。适用于日志记录和panic恢复等普遍适用的关注点。

- **分组中间件**应用于路由组中的所有路由。使用`group.Use()`注册。适用于将认证或授权应用到路由子集——我们会在后面了解到路由组。

- **路由级中间件**仅应用于单个路由。作为额外参数传递给`router.GET()`、`router.POST()`等。适用于路由特定的逻辑，如自定义限流或输入验证。

它们的**执行顺序**是**：**中间件函数按注册顺序执行。当中间件调用`c.Next()`时，它将控制权传递给下一个中间件（或最终处理函数），然后在`c.Next()`返回后继续执行。

这就创建了一个类似栈的模式——第一个注册的中间件最先开始但最后结束。`c.Next()`的作用是让当前中间件把控制权交给下一个中间件或 Handler，并等待它们执行完，以便执行后置逻辑。
\(注意：不调用`c.Next()`不会自动跳过后续 Handler\)

> 注意：注册是**不可回溯**的——如果你先注册了一个路由，再调用Use方法，那么该路由是不会应用中间件的，如果出现什么难以排查的bug，也可以检查下是不是调用顺序出了问题。
> 
> 



### C1\.2\.6路由组

**路由组**允许你将相关路由组织在一个共享的URL前缀下，用于将同一模块的路由归类管理，可以更方便的管理各个路由和API版本。



注册路由组需要调用路由引擎的**Group\(\)**方法，参数需要传入共享路由，随后我们便可以用返回的group注册其它路由了；同时，调用group的**Use\(\)**方法可以为group旗下的所有路由加上中间件。

比如下述示例里，我们注册了v1v2两个路由组，我们便可以访问http://localhost:8080/v1/login和http://localhost:8080/v2/submit了。除此之外，[http://localhost:8080/v2](http://localhost:8080/v2)后的所有路由，不论是submit还是其它方法，只要它们处于`/v2`路由组下，就都自带有xxxMiddleware这一中间件。

```Go
func loginEndpoint(c *gin.Context) {
  c.JSON(http.StatusOK, gin.H{"action": "login"})
}
func submitEndpoint(c *gin.Context) {
  c.JSON(http.StatusOK, gin.H{"action": "submit"})
}
func main() {
  router := gin.Default()
  {
    v1 := router.Group("/v1")
    v1.POST("/login", loginEndpoint)
  }
  {
    v2 := router.Group("/v2")
    v2.Use(xxxMiddleware())
    v2.POST("/submit", submitEndpoint)
  }
  router.Run(":8080")
}
```



### C1\.2\.7示例

现在，我们差不多已经了解了Gin框架的基本用法了。除了先前演示的用法之外，Gin框架还有许多其它用法——比如给客户端返回文件，流式传输，重定向等等——这些需要自己去理解和探索。在此之前，让我们看几个示例来进一步加深对Gin框架的理解吧。

```Go
type LoginRequest struct {
    Username string `json:"username" form:"username" binding:"required"`
    Password string `json:"password" form:"password" binding:"required,min=6"`
}
func main() {
    r := gin.Default()
    r.POST("/login", func(c *gin.Context) {
        var req LoginRequest
        // ShouldBind 会根据 Content-Type 自动选择 JSON 或表单绑定
        if err := c.ShouldBind(&req); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{
                "error": err.Error(),
            })
            return
        }
        c.JSON(http.StatusOK, gin.H{
            "message":  "登录成功",
            "username": req.Username,
        })
    })
    r.Run(":8080")
}
```



```Go
func main() {
    r := gin.Default()
    // 限制 multipart 表单在内存中的缓存为 8 MB，超出部分会落盘。
    r.MaxMultipartMemory = 8 << 20
    // 确保上传目录存在
    _ = os.MkdirAll("./uploads", os.ModePerm)
    // 通过 /static/xxx 访问上传后的文件
    r.Static("/static", "./uploads")
    r.POST("/upload", func(c *gin.Context) {
        file, err := c.FormFile("file")
        if err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
            return
        }
        dst := "./uploads/" + file.Filename
        if err := c.SaveUploadedFile(file, dst); err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            return
        }
        c.JSON(http.StatusOK, gin.H{
            "filename": file.Filename,
            "size":     file.Size,
            "url":      "/static/" + file.Filename,
        })
    })
    r.Run(":8080")
}
```



## C1\.3Hertz框架

Hertz是字节跳动开源的Go HTTP框架，主打高性能、可扩展和易用性。它的API与Gin类似，但在Handler签名、请求上下文、绑定校验等方面有一些差异。

除此之外，在底层方面，Hertz框架是默认Netpoll\+自研HTTP实现的，其在高性能、低延迟场景更加突出，但是对于net/http兼容性便相对较低了。

我们可以在终端使用`go get ``github.com/cloudwego/hertz`命令来导入Hertz框架。

### C1\.3\.1快速开始

Hertz通过**server\.Default\(\)**方法创建服务端，默认监听`:8888`端口。这里返回的server\.Hertz服务端是路由引擎和signalWaiter的综合，后者用于实现优雅退出，前者则跟Gin框架的路由引擎类似。

不同于gin，Hertz的Handler签名是**func\(c context\.Context, ctx \*app\.RequestContext\)**：第一个参数是 Go 标准库的`context.Context`，第二个参数是Hertz的请求上下文。

响应值可以使用**ctx\.JSON\(\)**写入，响应体可以用**utils\.H\{\}**快速表示。后者实际上也是`map[string]interface{}`的便捷写法。

最后调用阻塞的**Spin\(\)**方法启动服务即可。

稍后我们在浏览器访问http://localhost:8888/ping即可得到结果。

```Go
func main() {
    h := server.Default()
    h.GET("/ping", func(c context.Context, ctx *app.RequestContext) {
        ctx.JSON(consts.StatusOK, utils.H{"message": "pong"})
    })
    h.Spin()
}
```



### C1\.3\.2HTTP方法和路由组

除去直接动词映射的HTTP方法，比如h\.GET\(\)，h\.POST\(\)，h\.PUT\(\)等等，hertz还有一些特殊的HTTP方法。

|方法名|说明|
|---|---|
|h\.Handle|这个方法支持用户手动传入注册方法，可以传入GET，POST之类的传统方法，这同时也支持用于注册自定义的HTTP方法|
|h\.Any|用于匹配所有的HTTP方法|



路由组方面Hertz基本上与Gin一致，**Group\(\)**接收共享URL前缀，**Use\(\)**也是不可回溯的。路由组也可以继续嵌套。

```Go
v1 := h.Group("/v1")
v1.GET("/login", loginEndpoint)
v2 := h.Group("/v2")
v2.Use(AuthMiddleware())
v2.POST("/submit", submitEndpoint)
```



### C1\.3\.3请求和响应上下文

Hertz把请求信息封装在`*app.RequestContext` 中。为了减少篇幅，这里用表格集中对比Gin和Hertz的常用方法，不再逐一展开解释。

注意：Hertz的`GetHeader()`返回的是`[]byte`，一般需要手动转换成`string`。

|**用途**|**Gin**|**Hertz**|
|---|---|---|
|路径参数|c\.Param\(key\)|c\.Param\(key\)|
|查询参数|c\.Query\(key\) / c\.DefaultQuery\(key, def\)|c\.Query\(key\) / c\.DefaultQuery\(key, def\)|
|查询数组|c\.QueryArray\(key\)|c\.QueryArray\(key\)|
|表单参数|c\.PostForm\(key\) / c\.DefaultPostForm\(key, def\)|c\.PostForm\(key\) / c\.DefaultPostForm\(key, def\)|
|自动绑定|c\.ShouldBind\(\&obj\)|c\.Bind\(\&obj\)|
|JSON 绑定|c\.ShouldBindJSON\(\&obj\)|c\.BindJSON\(\&obj\)|
|绑定并校验|c\.ShouldBind\(\&obj\) \+ binding 标签|c\.BindAndValidate\(\&obj\) \+ vd 标签|
|获取请求头|c\.GetHeader\(key\) string|c\.GetHeader\(key\) \[\]byte|
|绑定请求头|c\.ShouldBindHeader\(\&obj\)|c\.BindHeader\(\&obj\)|
|文件上传|c\.FormFile\(name\) / c\.SaveUploadedFile\(file, dst\)|c\.FormFile\(name\) / c\.SaveUploadedFile\(file, dst\)|
|多文件表单|c\.MultipartForm\(\)|c\.MultipartForm\(\)|
|静态文件|r\.Static\(prefix, root\) / r\.StaticFile\(prefix, file\)|h\.Static\(prefix, root\) / h\.StaticFile\(prefix, file\)|



Hertz同样把响应写入能力封装在`*app.RequestContext`中。常用响应方法如下。

与Gin一样，Hertz 的`JSON()`等方法写入响应后也不会自动终止Handler，需要自己`return`

|**用途**|**Gin**|**Hertz**|
|---|---|---|
|设置状态码|c\.Status\(code\)|c\.SetStatusCode\(code\)|
|JSON 响应|c\.JSON\(code, obj\)|c\.JSON\(code, obj\)|
|字符串响应|c\.String\(code, format, args\.\.\.\)|c\.String\(code, format, args\.\.\.\)|
|原始数据|c\.Data\(code, contentType, data\)|c\.Data\(code, contentType, data\)|
|设置响应头|c\.Header\(key, value\)|c\.Header\(key, value\)|
|重定向|c\.Redirect\(code, location\)|c\.Redirect\(code, location\)|
|返回文件|c\.File\(path\)|c\.File\(path\)|
|下载文件|c\.FileAttachment\(path, filename\)|c\.FileAttachment\(path, filename\)|
|中断并返回JSON|c\.AbortWithStatusJSON\(code, obj\)|c\.AbortWithStatusJSON\(code, obj\)|



### C1\.3\.4中间件

Hertz的中间件类型是**app\.HandlerFunc**，签名与普通Handler一致：**func\(context\.Context, \*app\.RequestContext\)**。

与gin不同，hertz的**Next\(\)**方法需要传入标准context。

- **ctx\.Next\(c\)**：把控制权交给下一个中间件或 Handler，并在它们执行完后继续执行后续代码。

- **ctx\.Abort\(\)**：阻止后续处理器执行，通常需要紧接着 `return`。

- **ctx\.AbortWithStatusJSON\(code, obj\)**：中断并返回 JSON。

hertz中间件的注册方式与Gin类似。

- 全局中间件使用`h.Use()`。

- 分组中间件使用`group.Use()`。

- 路由级中间件作为额外参数传给`h.GET()`、`h.POST()`等方法。

```Go
func AuthMiddleware() app.HandlerFunc {
    return func(c context.Context, ctx *app.RequestContext) {
        token := string(ctx.GetHeader("Authorization"))
        if token != "Bearer secret-token" {
            ctx.AbortWithStatusJSON(consts.StatusUnauthorized, utils.H{
                "error": "unauthorized",
            })
            return
        }
        ctx.Next(c)
    }
}

h.Use(AuthMiddleware())
```



### C1\.3\.5示例

我们随便弄一个登录的示例：

```Go
type LoginRequest struct {
    Username string `json:"username"`
    Password string `json:"password"`
}
func xxxMiddleware() app.HandlerFunc {
    return func(c context.Context, ctx *app.RequestContext) {
        ......
        ctx.Next(c)
    }
}

func main() {
    h := server.Default()
    h.POST("/login", func(c context.Context, ctx *app.RequestContext) {
        var req LoginRequest
        if err := ctx.BindJSON(&req); err != nil {
            ctx.JSON(consts.StatusBadRequest, utils.H{
                "error": err.Error(),
            })
            return
        }
        ctx.JSON(consts.StatusOK, utils.H{
            "message":  "登录成功",
            "username": req.Username,
        })
    })
    g1 := h.Group("/g1")
    g1.Use(xxxMiddleware())
    {
        g1.GET("/test", func(c context.Context, ctx *app.RequestContext) {
            ctx.JSON(consts.StatusOK, utils.H{
                "message": "test",
            })
        })
    }
    h.Spin()
}
```



除此之外，hertz还有很多其它用法，感兴趣可自行去查找。



## C1\.4Echo框架

Echo是Go语言中一款高性能、可扩展、极简的Web框架，构建在`net/http`之上。它的API风格与Gin相近，但在Handler签名、错误处理、中间件、参数校验上有明显差异——最大的区别就是，echo的很多方法都是会返回error的。

我们可以在终端使用`go get ``github.com/labstack/echo/v4`来导入Echo框架。

### C1\.4\.1快速开始

通过**echo\.New\(\)**创建`*echo.Echo`实例，它既是路由引擎，也是服务入口。Echo默认**不会**自动挂载Logger、Recover等中间件，需要自己注册。

注册路由使用 `e.GET(path, handler)`。而Echo的Handler签名是：**`func(c echo.Context) error`**

也就是说，echo的**Handler必须返回error**。也因此，echo的`c.JSON(code, obj)`会写入JSON并返回error，所以我们在这里直接`return c.JSON(...)`。

最后使用**Start\(\)**方法启动路由引擎即可。

接下来访问`http://localhost:8080/ping`即可得到成果。

```Go
func main() {
    e := echo.New()
    e.GET("/ping", func(c echo.Context) error {
        return c.JSON(http.StatusOK, map[string]string{
            "message": "pong",
        })
    })
    e.Logger.Fatal(e.Start(":8080"))
}
```



### C1\.4\.2HTTP方法和路由组

Echo 同样提供了直接映射 HTTP 动词的方法，例如e\.GET\(\)、e\.POST\(\)、e\.PUT\(\)等等。

除此之外，Echo还有一些特殊注册方法：

|方法|说明|
|---|---|
|`e.Any(path, handler)`|匹配所有 HTTP 方法|
|`e.Match(methods, path, handler)`|一次注册多个 HTTP 方法|
|`e.Add(method, path, handler)`|注册指定方法，也可用于自定义 HTTP 方法|

> 注意：为同一路径注册的特定方法路由优先于`Any`路由。
> 
> 



路由组方面，Echo与Gin很相似，`Group()`接收共享URL前缀，`Use()`对之后注册的路由生效，**不可回溯**。路由组也可以继续嵌套：

```Go
v1 := e.Group("/v1")
v1.GET("/login", loginHandler)
v2 := e.Group("/v2")
v2.Use(AuthMiddleware)
v2.POST("/submit", submitHandler)
```



### C1\.4\.3请求和响应上下文

Echo把请求信息封装在了**echo\.Context**中。注意，`echo.Context`是一个**接口**，不是Gin那样的具体结构体。它同时提供了`Request()`和`Response()`，以此方便拿到底层的`*http.Request`和`*http.Response`。

以下是相对于Gin，Echo存在的一些方法，感兴趣可自行查看。

注意：`c.Validate(&obj)`只负责校验，使用前需要给`e.Validator`注册校验器。

|用途|Gin|Echo|
|---|---|---|
|路径参数|`c.Param(key)`|`c.Param(key)`|
|查询参数|`c.Query(key)` / `c.DefaultQuery(key, def)`|`c.QueryParam(key)`；没有默认值方法|
|查询数组|`c.QueryArray(key)`|`c.QueryParams()[key]`|
|表单参数|`c.PostForm(key)` / `c.DefaultPostForm(key, def)`|`c.FormValue(key)` / `c.FormParams()`|
|自动绑定|`c.ShouldBind(&obj)`|`c.Bind(&obj)`|
|JSON 绑定|`c.ShouldBindJSON(&obj)`|`c.Bind(&obj)`，或结合 `Content-Type` 判断|
|绑定并校验|`c.ShouldBind(&obj)` \+ `binding` 标签|`c.Bind(&obj)` \+ `c.Validate(&obj)` \+ `validate` 标签|
|获取请求头|`c.GetHeader(key) string`|`c.Request().Header.Get(key)`|
|文件上传|`c.FormFile(name)` / `c.SaveUploadedFile(file, dst)`|`c.FormFile(name)` \+ `file.Open()` \+ `io.Copy`|
|多文件表单|`c.MultipartForm()`|`c.MultipartForm()`|
|静态文件|`r.Static(prefix, root)` / `r.StaticFile(prefix, file)`|`e.Static(prefix, root)` / `e.File(path, file)`|



与Gin一样，响应方面同样通过上下文写回。除此之外，Echo的`c.JSON()`等方法写入响应后也不会自动终止Handler，需要`return`，因此Echo一般直接写成`return c.JSON()`

|用途|Gin|Echo|
|---|---|---|
|设置状态码|`c.Status(code)`|`c.NoContent(code)` 或 `c.Response().WriteHeader(code)`|
|JSON 响应|`c.JSON(code, obj)`|`return c.JSON(code, obj)`|
|字符串响应|`c.String(code, format, args...)`|`c.String(code, s)`|
|原始数据|`c.Data(code, contentType, data)`|`c.Blob(code, contentType, b)`|
|设置响应头|`c.Header(key, value)`|`c.Response().Header().Set(key, value)`|
|重定向|`c.Redirect(code, location)`|`c.Redirect(code, url)`|
|返回文件|`c.File(path)`|`c.File(path)`|
|下载文件|`c.FileAttachment(path, filename)`|`c.Attachment(file, name)`|
|HTML 字符串|`c.Data(200, "text/html...", ...)`|`c.HTML(code, html)`|
|模板渲染|`c.HTML(code, name, obj)`|`c.Render(code, name, data)`|



### C1\.4\.4中间件和错误处理

1. echo的Handler类型和中间件类型如下，可以发现，Echo的中间件不是Gin那种在同一个c上调用`c.Next()`，而是包装**next handler**，通过调用`next(c)`表示继续执行后续链路。

```Go
type HandlerFunc func(c echo.Context) error
type MiddlewareFunc func(next echo.HandlerFunc) echo.HandlerFunc

func AuthMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
    return func(c echo.Context) error {
        token := c.Request().Header.Get("Authorization")
        if token != "Bearer secret-token" {
            return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
        }
        return next(c)
    }
}
```



2. echo的注册方式如下：

- 全局中间件使用`e.Use()`。

- 分组中间件使用`g.Use()`。

- 路由级中间件作为额外参数传给`e.GET()`、`e.POST()`等方法。

```Go
e.Use(middleware.Logger())
e.Use(middleware.Recover())
e.Use(xxxMiddleware)
```



3. Echo的错误处理很有特色：Handler直接`return err`，由`e.HTTPErrorHandler`统一处理。当然，你也可以自定义错误处理。

> 注意：如果Handler里已经调用了`c.JSON()`，通常直接`return nil`；如果返回error，就不要再写响应，否则可能重复写header。
> 
> 

```Go
return echo.NewHTTPError(http.StatusBadRequest, "invalid params")
e.HTTPErrorHandler = func(err error, c echo.Context) {
    if he, ok := err.(*echo.HTTPError); ok {
        _ = c.JSON(he.Code, map[string]interface{}{
                "error": he.Message,
        })
        return
    }
    _ = c.JSON(http.StatusInternalServerError, map[string]interface{}{
        "error": "internal error",
    })
}
```



### C1\.4\.5示例

最后，让我们来看一下几个Echo的例子吧：

```Go
type LoginRequest struct {
    Username string `json:"username" form:"username"`
    Password string `json:"password" form:"password"`
}
func main() {
    e := echo.New()
    e.POST("/login", func(c echo.Context) error {
        var req LoginRequest
        if err := c.Bind(&req); err != nil {
            return echo.NewHTTPError(http.StatusBadRequest, err.Error())
        }
        return c.JSON(http.StatusOK, map[string]string{
            "message":  "登录成功",
            "username": req.Username,
        })
    })
    e.Logger.Fatal(e.Start(":8080"))
}
```



```Go
e.POST("/upload", func(c echo.Context) error {
    file, err := c.FormFile("file")
    if err != nil {
        return echo.NewHTTPError(http.StatusBadRequest, err.Error())
    }
    src, err := file.Open()
    if err != nil {
        return err
    }
    defer src.Close()
    dst, err := os.Create("./uploads/" + file.Filename)
    if err != nil {
        return err
    }
    defer dst.Close()
    if _, err := io.Copy(dst, src); err != nil {
        return err
    }
    return c.JSON(http.StatusOK, map[string]interface{}{
        "filename": file.Filename,
        "size":     file.Size,
    })
})
```





## C1\.5对比

最后，让我们来看看Hertz框架，Gin框架和Echo框架的对比表吧：

|**维度**|**Gin**|**Echo**|**Hertz**|
|---|---|---|---|
|底层网络|net/http|net/http|默认Netpoll\+自研HTTP实现|
|定位|最流行、轻量、生态最大|功能完整、官方中间件丰富|高性能、可扩展、云原生微服务|
|Handler 风格|单请求上下文，不返回error|请求上下文\+返回error|标准context\+RequestContext双参数|
|错误处理|手动写响应，或中间件统一处理|Handler返回error，集中处理|手动写响应，或中间件统一处理|
|参数校验|binding标签|需注册Validator，validate标签|BindAndValidate\+vd 标签|
|中间件风格|c\.Next\(\)链式控制|包装next echo\.HandlerFunc|ctx\.Next\(c\)链式控制|
|net/http 兼容性|高|高|较低，通常需要适配|
|性能特点|高|高|通常更高，高并发、低延迟场景更突出|
|生态|最大，教程和第三方中间件最多|官方中间件完善，社区稳定|CloudWeGo/Kitex生态强，社区相对小|
|学习成本|低|低|中，签名、绑定、校验有差异|
|典型场景|通用Web、API、教学|中大型API、需要完整官方能力|高并发微服务、云原生、字节系技术栈|







