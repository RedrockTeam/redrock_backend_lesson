# Gin框架源码解析

# Gin 框架源码解析

# Gin 框架源码解析

## Why Gin

为什么我们会使用 Gin 框架呢？ 对比 Go 官方原生的 HTTP 库 net/http

### 良好的路由匹配机制

net.http是通过map实现的路由匹配的，这样的话对带有参数的路由支持不好，而且也会匹配错误。 net/http 是 Go 语言中用来创建 HTTP 服务器的标准库之一，它通过 `http.ServeMux` 来进行路由匹配。 `http.ServeMux` 是一个 HTTP 请求路由器，它会根据请求的 URL 路径来匹配对应的处理函数。在创建 `http.ServeMux` 实例后，可以通过调用 `mux.Handle()` 或 `mux.HandleFunc()` 方法来为特定的 URL 路径注册处理函数。 直接上代码

```Go
type ServeMux struct {
        mu    sync.RWMutex
        m     map[string]muxEntry 
        es    []muxEntry 
        hosts bool       // whether any patterns contain hostnames
}
// 用于存储 URL 路径和处理函数之间的映射关系
type muxEntry struct {
        h       Handler // 处理 HTTP 请求的处理函数，可以是一个函数或实现了 http.Handler 接口的结构体
        pattern string  // URL 路径的模式字符串，用于匹配 HTTP 请求的 URL 路径
}
```

其中，最关键的是 `m` ，它是一个 map 类型， 用于保存 URL 路径和对应的处理函数的映射关系。键是字符串类型的 URL 路径，值是一个 muxEntry 类型的结构体。 下面我们直接看它的匹配规则

```Go
// 在给定路径字符串的 handler map 上查找一个 handler。
func (mux *ServeMux) match(path string) (h Handler, pattern string) {
        // 首先检查精确匹配。即从 map 中获取
        v, ok := mux.m[path]
        if ok {
                return v.h, v.pattern
        }
        // 如果匹配不到，就遍历 es 切片。检查最长有效匹配。es 包含所有模式
        // 以 / 结尾，从长到短排序。
        for _, e := range mux.es {
                if strings.HasPrefix(path, e.pattern) {
                        return e.h, e.pattern
                }
        }
        return nil, ""
}
```

所以为什么它对带有参数的路由支持不好且会匹配错误呢？ 无奖竞猜。 而 Gin 框架采用了 httprouter 进行路由匹配，httprouter 是通过 Tire tree 来进行高效的路径查找；同时路径还支持两种通配符匹配。Gin 会将请求路径和路由路径进行比较，如果路径完全匹配，则直接返回匹配的处理函数；如果路由路径包含参数，则将参数的值保存到上下文（`Context`）中，供后续的处理函数使用；如果路由路径包含通配符，则继续递归查找，直到找到最后一个节点。我们下面讲 Engine 和 RouterGroup 的时候会提到。

### 简单易用

Gin 框架对 net/http 库进行了良好的封装，并且它采用了类似于 HTTP 标准库的 API 设计风格，使用起来非常直观和简单，不需要过多的学习成本。

## 从一个简单的 Demo 开始

下面是一个最简单的 Demo

```Go
package main
import (
  "net/http"
  "github.com/gin-gonic/gin"
)
func main() {
  r := gin.Default()
  r.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}
```

`r := gin.Default()`

```Go
func Default() *Engine {
        debugPrintWARNINGDefault()
        engine := New()
        engine.Use(Logger(), Recovery())
        return engine
}
```

我们可以看到，调用这个函数，我们可以得到一个结构体指针，结构体为 Engine。 其中， Logger 与 Recovery 是 gin 的两个中间件，前者用来打印日志， 后者用来从任何异常中恢复，如果有异常，则写入500状态码（服务器内部错误）。 `r.Run()`

```Go
(engine *Engine) Run(addr ...string) (err error) {
   ……
   err = http.ListenAndServe(address, engine)
   return
}
```

可以看到我们把engine传入了http.ListenAndServe，让我们再康康这个的源码

```Go
func ListenAndServe(addr string, handler Handler) error {
        server := &Server{Addr: addr, Handler: handler}
        return server.ListenAndServe()
}
```

engine满足了Handler接口，传入了server 然后开始监听

```Go
func (srv *Server) ListenAndServe() error {
        ......
        addr := srv.Addr
        ......
        ln, err := net.Listen("tcp", addr)
        ......
        return srv.Serve(ln)
}
```

```Go
func (srv *Server) Serve(l net.Listener) error {
        ......
        for {
                // Accept等待并返回到侦听器的下一个连接。
                rw, err := l.Accept()
                ......
                tempDelay = 0
                c := srv.newConn(rw)
                c.setState(c.rwc, StateNew, runHooks) // before Serve can return
                go c.serve(connCtx)
        }
}
```

然后用 go 关键字起一个协程处理请求。 从这里把消息处理交给gin，让我们康康gin具体是怎么满足`Handler`接口的

```Go
func (c *conn) serve(ctx context.Context) {
   ………
      // HTTP cannot have multiple simultaneous active requests.[*]
      // Until the server replies to this request, it can't read another,
      // so we might as well run the handler in this goroutine.
      // [*] Not strictly true: HTTP pipelining. We could let them all process
      // in parallel even if their responses need to be serialized.
      // But we're not going to implement HTTP pipelining because it
      // was never deployed in the wild and the answer is HTTP/2.
      serverHandler{c.server}.ServeHTTP(w, w.req)
}
```

```Go
func (engine *Engine) ServeHTTP(w http.ResponseWriter, req *http.Request) {
   c := engine.pool.Get().(*Context)
   c.writermem.reset(w)
   c.Request = req
   c.reset()
   engine.handleHTTPRequest(c)
   engine.pool.Put(c)
}
```

这个代码就很简单了，可能大家没学过sync包里的东西,这里复制了一段话（我们通常用golang来构建高并发场景下的应用，但是由于golang内建的GC机制会影响应用的性能，为了减少GC，golang提供了对象重用的机制，也就是sync.Pool对象池。 sync.Pool是可伸缩的，并发安全的。其大小仅受限于内存的大小，可以被看作是一个存放可重用对象的值的容器。 设计的目的是存放已经分配的但是暂时不用的对象，在需要用到的时候直接从pool中取。）但这不是重点。 我们先从池中拿出一个context,初始化开始处理HTTP请求，结束后放回池中 这就是一个请求进入gin处理的过程

### 总结

net/http 根据请求的URL地址找到对应的处理器，调用处理器对应的`ServeHTTP()`方法处理请求。而Gin 框架构建一个 Web 应用的流程就是：先通过 gin.Deafult() （如果调用这个方法，会使用 Logger 和 Rcovery 这两个中间件）或者 gin.New() （不会使用任何中间件）来获取一个 engine 结构体指针，这个 engine 实现了 `ServeHTTP()` 方法， 然后用户操作这个 engine 来进行路由注册，中间件注册等操作。而`ServeHTTP()`的实现，是获取对象池中的一个 context，初始化并处理请求， 结束后我们调用 put 把它放回到池中。

### Engine

在 Gin 框架中，`Engine` 是整个框架的核心组件，负责管理所有的路由和中间件，并处理所有的 HTTP 请求和响应。`Engine` 是一个结构体类型，定义如下：

```Go
goCopy codetype Engine struct {
    RouterGroup // 是一个路由组，用于管理一组相关的路由规则和中间件。
    // 如果true，当前路由匹配失败但将路径最后的 / 去掉时匹配成功时自动匹配后者
    // 比如：请求是 /foo/ 但没有命中，而存在 /foo，
    // 对get method请求，客户端会被301重定向到 /foo
    // 对于其他method请求，客户端会被307重定向到 /foo
    RedirectTrailingSlash bool 
    // 是一个字符串到 radix 树的映射，用于实现路由匹配。
    trees map[string]*node 
    // 控制是否自动处理 HTTP 方法不允许的情况。
    HandleMethodNotAllowed bool 
    // 是中间件列表，用于处理所有的 HTTP 请求和响应。
    Middleware []HandlerFunc
    ......
    // 用于对象池的 sync.Pool 类型
    Pool sync.Pool 
    ......
}
```

我们通常用golang来构建高并发场景下的应用，但是由于golang内建的GC机制会影响应用的性能，为了减少GC，golang提供了对象重用的机制，也就是sync.Pool对象池。 sync.Pool是可伸缩的，并发安全的。其大小仅受限于内存的大小，可以被看作是一个存放可重用对象的值的容器。 设计的目的是存放已经分配的但是暂时不用的对象，在需要用到的时候直接从pool中取。）但这不是重点。 通过对 `Engine` 结构体的设置，可以控制框架的路由匹配、中间件处理、模板渲染等方面的行为，使得开发者能够更加方便地使用 Gin 框架构建 Web 应用程序。

### RouterGroup

在 Gin 框架中，`RouterGroup` 是一个路由组，用于对路由进行分组，以便于管理和维护。`RouterGroup` 对象包含了一些公共方法，如 `GET`、`POST`、`PUT`、`PATCH`、`DELETE` 等方法，这些方法可以用于注册对应的 HTTP 请求方法和路由，然后在该分组中共享中间件和参数。 在 Gin 框架中，我们可以通过 `engine.Group` 方法来创建一个 `RouterGroup` 对象，用于分组处理路由。 我们在注册路由时，使用的是路由组对象的方法，而不是 `engine` 对象的方法。 如下：

```Go
func (group *RouterGroup) Handle(httpMethod, relativePath string, handlers ...HandlerFunc) IRoutes {
        // 这里是进行传参校验， 如果传入的参数无效（就是没有这种请求方法），panic
        if matched := regEnLetter.MatchString(httpMethod); !matched {
                panic("http method " + httpMethod + " is not valid")
        }
        // 检查无误调用 handle 方法
        return group.handle(httpMethod, relativePath, handlers)
}
func (group *RouterGroup) handle(httpMethod, relativePath string, handlers HandlersChain) IRoutes {
        // 算出绝对路径，就是把路由组的路径和现在这个相对路径拼起来，得到一个完整的路径
        absolutePath := group.calculateAbsolutePath(relativePath)
        // 获取这个路径下的处理方法（比如用了什么中间件之类的）
        handlers = group.combineHandlers(handlers)
        // 然后进行路由注册
        group.engine.addRoute(httpMethod, absolutePath, handlers)
        return group.returnObj()
}
func (engine *Engine) addRoute(method, path string, handlers HandlersChain) {
        ......
        // 打印日志
        debugPrintRoute(method, path, handlers)
        // 获取对应方法的前缀树，每个方法对应一颗前缀树
        root := engine.trees.get(method)
        // 如果没有获取到这棵树（就是没注册过这个方法，就创建一个新的树并加入到那个树的切片中。
        if root == nil {
                root = new(node)
                root.fullPath = "/"
                engine.trees = append(engine.trees, methodTree{method: method, root: root})
        }
        // 然后进行添加节点
        root.addRoute(path, handlers)
        // 这里的话就是更新参数的信息啥的
        if paramsCount := countParams(path); paramsCount > engine.maxParams {
                engine.maxParams = paramsCount
        }
        if sectionsCount := countSections(path); sectionsCount > engine.maxSections {
                engine.maxSections = sectionsCount
        }
}
```

`**handle**` **干了什么：** 我们调用 `GET` `POST` 这些方法时，其实也是调用了 group.Handle 方法，只不过 Gin 把它封装了一下。 先通过`calculateAbsolutePath`计算出绝对路径, 这个路径是由路由组的路径和当前请求的路径拼接而成的。 然后，调用 `group.combineHandlers(handlers)` 方法将处理请求的函数列表合并成一个函数列表。 最后，调用 `group.engine.addRoute(httpMethod, absolutePath, handlers)` 方法将路由和处理函数添加到路由器中。 最后一步是比较关键的，它将路由和对应的处理函数保存在路由器中，以便之后能够根据请求的 URL 和请求方法，匹配到对应的处理函数进行处理。同时，`group.returnObj()` 方法返回当前路由组的 `IRoutes` 接口，从而实现链式调用。 **其中** `**addRoute**` **主要干了什么：** `debugPrintRoute()` 函数输出日志信息，记录新增的路由信息。 调用 `engine.trees.get(method)` 方法获取相应 HTTP 请求方法的路由树。如果该路由树不存在，就创建一个新的根节点，其 `fullPath` 字段为 `/`，并将该根节点添加到路由树列表 `engine.trees` 中。接着，调用根节点的 `addRoute()` 方法，将路由和对应的处理函数添加到该节点下面的子节点中。 最后，如果新增的路由路径中包含的参数数量大于 `engine.maxParams`，则更新 `engine.maxParams`；如果新增的路由路径中的路径段数量大于 `engine.maxSections`，则更新 `engine.maxSections`。

#### 路由树的实现

##### 大概流程

它是个前缀树 这个前缀树的每个节点包含以下主要字段：

* `path`：节点代表的路径段（相对路径，与祖先节点拼接可得完整路径）。
* `wildChild`：标识是否有子节点是通配符节点。
* `indices`：索引数组，每个元素表示该节点下面有哪些子节点，每个元素对应一个字符，取值为 0 到 255。
* `children`：子节点数组，存放所有子节点。
* `fullpath`: 是从 root 节点到当前节点的全部 path 部分
* `handlers`：如果此节点为终结节点，则此为它的处理函数，否则为 nil
* `nType` : 它表示这个节点是个什么节点，没有参数的静态节点（static），根节点（root），有参数的节点（param），有通配符`*`的节点。 其中，通配符节点是一种特殊的节点，它的 `path` 字段以 `:` 或 `*` 开头，表示可以匹配任意值。例如，路径 `/users/:id` 中的 `:id` 就是一个通配符节点。 Gin 框架中，每个 HTTP 请求方法都对应一棵前缀树。`Engine` 对象中的 `trees` 字段是一个 `methodTrees` 类型的切片，其中每个元素都包含一个 HTTP 请求方法和对应的前缀树。每个前缀树的根节点都表示根路径 `/`，所有添加的路由都是该节点的子节点，最终形成一棵完整的路由树。 在添加路由时，Gin 框架会根据请求方法获取相应的前缀树，并从根节点开始逐层查找对应的路由节点。对于包含通配符的路由，Gin 框架会将通配符节点的 `handle` 字段设置为对应的处理函数，并将 `wildChild` 字段设置为 `true`，表示该节点下面有子节点是通配符节点。这样，当请求路径与路由匹配时，Gin 框架会沿着路由树向下遍历，查找对应的处理函数。

##### addRoute

懒得写了，cv了

###### 流程

 ![](https://internal-api-drive-stream.feishu.cn/space/api/box/stream/download/authcode/?code=NmFkNWY2MTA0OTM2ZTIzZmJlNDEwN2JlNmZiOWJmYjRfNjllNDA0ODAyMjEwMTQzMjFlZjAxYTkyZDVlYWM1NTZfSUQ6NzIxOTU1MDcxNTIwMTUzNjAwMV8xNzgxNzk1MDkyOjE3ODE3OTg2OTJfVjM) case1：单条包含两个冒号通配path的一颗前缀树 路由信息： "/user/:name/:age" name和age是通配符的名字；这条路径形成的前缀树如下。  ![](https://internal-api-drive-stream.feishu.cn/space/api/box/stream/download/authcode/?code=NzAyNmQ1MGQwZjY1YjM1MjFmZmRmNDhmNzlkYWJlNmVfZWE5OGFlYjQ1ZTlhMDRiMGZmYjNhMzM5NDNhZTI1NjVfSUQ6NzIxOTU1MDcxMTM5ODIwMzQyMF8xNzgxNzk1MDkyOjE3ODE3OTg2OTJfVjM) 形成的过程是这样的：


 1. 查找到当前path /user/:name/:age的wildCard为 \[:name\] ,开始的位置是第7个字符
 2. 判断wildCard是冒号通配符类型且起始位置大于0，那么path之前的部分 \[/user/\] 设置为当前节点（上面树中的第一层节点）的path，然后更新要处理的path为后半部分，即 \[:name/:age\]
 3. 设置当前node wildChild为true。
 4. 新建孩子节点，path为wilCard \[:name\],并且挂接到第一层节点，并将当前节点指向第二层的节点。
 5. 判断path是否处理结束，如果处理结束，则可以退出
 6. 如果path还没有处理结束，更新path为除去wildCard剩余部分 \[/:age\] , 新建孩子节点，挂接到当前的节点（当前还指向第二层）并且设置当前节点为第三层结点
 7. 更新后的path为 \[/:age\]，当前的节点为第三层节点，继续从第一步开始处理。再次循环一遍，两个冒号通配符处理完成。 case2：单条包含星号通配path的一颗前缀树 路由信息："/user/:name/\*age" ，其中包含一个冒号通配符和一个星号通配符；这条路径形成的前缀树如下： ![](https://internal-api-drive-stream.feishu.cn/space/api/box/stream/download/authcode/?code=OWUzYzg1NzJjM2IxODRjNjA4ZDM3ZTIwNjljMDQ1ZDhfMzVkMWVlZDViMjkzMDZhZmVmZWZhMzM1ODg0YmMwMTNfSUQ6NzIxOTU1MDcxMjk3ODUzODQ5N18xNzgxNzk1MDkyOjE3ODE3OTg2OTJfVjM) 对于这个路由，形成的过程如下：
 8. 查找当前path的wildCard，当前node指向第一层的node，wildCard的查找结果为【:name】位置在第七个
 9. 按照上面的处理步骤，处理完wildCard，path更新为【/\*age】，当前node指向第三层，进行下一次循环处理
10. 再次查找wildCard结果为【\*age】位置是第二个符号，进入insertChild插入星号通配符的逻辑
11. 查看通配符之前的是否是下划线，不是的话报错，将当前的节点path设置为下划线之前的path，这个时候当前节点path为空
12. 生成子节点，即图中的第四层节点，设置节点类型为catchAll，挂接到当前节点，设置单前节点的indices为【/】，当前节点指向第四层节点
13. 新建第五层节点，节点path设置为剩下的path（此处，星号通配符之后不再处理），并挂接到当前当前节点
14. 处理结束 case3：多种path的寻找过程 路由信息：第一条【"/user/info"】，第二条【"/user/info/:name"】，第三条【"/user/info/:name/\*age"】 ![](https://internal-api-drive-stream.feishu.cn/space/api/box/stream/download/authcode/?code=NjljZDQzMTUzZTMxYmRmNDlkOTE2ZDA3OTQ0NWVmMjlfYTg0MmQxYzI0OWJiMDVlMDgzNWVkM2Y1ZmYxMGNjNWNfSUQ6NzIxOTU1MDcxNDE0ODgzMTIzM18xNzgxNzk1MDkyOjE3ODE3OTg2OTJfVjM) 对于上述的路由信息，前缀树的形成过程为：
15. 第一条路由信息，root树为空树并且没有通配符，会在insertChild方法中直接设置当前节点（第一层节点）为path，即在第一层节点设置path为【"/user/info"】
16. 对于第二条路由，先找到与当前节点的最长匹配长度，长度小于单前path长度，重置path为【/:name】且当前节点的wilChild为false，取出剩余path第一个字符进行判断
17. 根据第一个字符逻辑 在当前节点新增孩子（图中第二级孩子），并将当前node指向第二层节点，调用insertChild将【/:name】插入到第二层节点，逻辑同case1，这样形成了第二层和第三层节点。
18. 对于第三条路由，先找到与第一层节点的共同前缀，path更新为【/:name/\*age】，取出剩余path第一个字符【/】，【/】在当前的indices里，对孩子节点做优先级调整，并且将当前节点更新下沉到孩子节点,也就是图中第二层节点，继续下一次循环查找。
19. path【/:name/\*age】与第二层节点的共同前缀为【/】，当前第二层节点wildCard为true，这个时候将当前节点指定到孩子节点，即图中第三层节点，并且判断path在下划线之前的部分，即【:name】与三层节点的path相同，然后继续循环查找。
20. path剩余部分为【:name/\*\*age】，当前node指向第三层节点，继续查找插入位置；共同前缀为【:name】, 更新path变为【/\*\*age】，由于在第二部分处理完以后只有第三层节点。这个时候直接新增节点，进入insertChild逻辑。与case2的后半部分按照相同的逻辑，形成第四五六层节点

###### 代码实现

添加

```Go
func (n *node) addRoute(path string, handlers HandlersChain) {
   fullPath := path
   n.priority++
   // 如果是空树，我们就直接创一个新的
   if len(n.path) == 0 && len(n.children) == 0 {
      n.insertChild(path, fullPath, handlers)
      n.nType = root
      return
   }
   // 它们公共前缀的长度
   parentFullPathIndex := 0
walk:
   for {
       //查找最长的公共前缀。
       //这也意味着公共前缀不包含"："或"*"
       //因为现有的键不能包含这些字符。
       //大概意思是：如果公共前缀包含通配符，假如现在注册的是 ":name" 和 ":ages"
       //这俩应该是单独的俩节点，不应该同时出现在这里。
      i := longestCommonPrefix(path, n.path)
      // 如果说这个i小于这个节点的path，意味着这个节点要分裂
      if i < len(n.path) {
         child := node{
            //新的孩子节点就取后面
            path:      n.path[i:],
            wildChild: n.wildChild,
            indices:   n.indices,
            children:  n.children,
            handlers:  n.handlers,
            priority:  n.priority - 1,
            fullPath:  n.fullPath,
         }
         //将新孩子挂上去
         n.children = []*node{&child}
         //BytesToString 强制转换，性能更高。因为string的底层就是[]byte
         //此时这个节点就只挂了一个孩子。就把indices设置为新的孩子的
         n.indices = bytesconv.BytesToString([]byte{n.path[i]})
         //他的路径取前半部分
         n.path = path[:i]
         //把 handlers 挂到孩子节点
         n.handlers = nil
         n.wildChild = false
         n.fullPath = fullPath[:parentFullPathIndex+i]
      }
      //这个新节点前半部分与这个节点全部重合了，直接在这个节点下面挂
      if i < len(path) {
         path = path[i:]
         c := path[0]
         // '/' after param
         //冒号通配符后面的 下划线处理
         if n.nType == param && c == '/' && len(n.children) == 1 {
            parentFullPathIndex += len(n.path)
            n = n.children[0]
            n.priority++
            continue walk
         }
         // 当前节点的某个孩子第一个字符与path的第一个字符相同
         for i, max := 0, len(n.indices); i < max; i++ {
            if c == n.indices[i] {
               parentFullPathIndex += len(n.path)
                //修改优先级，重新排序
               i = n.incrementChildPrio(i)
               n = n.children[i]
               continue walk
            }
         }
         //其他情况就插入就完事
         if c != ':' && c != '*' && n.nType != catchAll {
            // 首字母加入节点
            n.indices += bytesconv.BytesToString([]byte{c})
            child := &node{
               fullPath: fullPath,
            }
            n.addChild(child)
            n.incrementChildPrio(len(n.indices) - 1)
            n = child
         } else if n.wildChild {
            // 插入通配符节点时，需要检查它是否与现有通配符冲突
            n = n.children[len(n.children)-1]
            n.priority++
            // Check if the wildcard matches
            if len(path) >= len(n.path) && n.path == path[:len(n.path)] &&
               // 向catchAll中添加子项是不可能的
               n.nType != catchAll &&
               // Check for longer wildcard, e.g. :name and :names
               (len(n.path) >= len(path) || path[len(n.path)] == '/') {
               continue walk
            }
            // 冲突了！
            // 举个例子就是 注册了 :name 后注册 :names
            pathSeg := path
            if n.nType != catchAll {
               pathSeg = strings.SplitN(pathSeg, "/", 2)[0]
            }
            prefix := fullPath[:strings.Index(fullPath, pathSeg)] + n.path
            panic("'" + pathSeg +
               "' in new path '" + fullPath +
               "' conflicts with existing wildcard '" + n.path +
               "' in existing prefix '" + prefix +
               "'")
         }
         n.insertChild(path, fullPath, handlers)
         return
      }
      // Otherwise add handle to current node
      if n.handlers != nil {
         panic("handlers are already registered for path '" + fullPath + "'")
      }
      n.handlers = handlers
      n.fullPath = fullPath
      return
   }
}
```

插入

```Go
func (n *node) insertChild(path string, fullPath string, handlers HandlersChain) {
   for {
      // 找到第一个通配符的之前path
      wildcard, i, valid := findWildcard(path)
      if i < 0 { // No wildcard found
         break
      }
      // The wildcard name must not contain ':' and '*'
       //因为i不是-1，而且vaild为false，直接panic掉
      if !valid {
         panic("only one wildcard per path segment is allowed, has: '" +
            wildcard + "' in path '" + fullPath + "'")
      }
      // check if the wildcard has a name
      if len(wildcard) < 2 {
         panic("wildcards must be named with a non-empty name in path '" + fullPath + "'")
      }
      if wildcard[0] == ':' { // param
         if i > 0 {
            // Insert prefix before the current wildcard
             //更新节点的path
            n.path = path[:i]
             //将path设置为通配符开始的后面一段
            path = path[i:]
         }
         child := &node{
            nType:    param,
            path:     wildcard,
            fullPath: fullPath,
         }
                 //将孩子节点挂到这个节点下
         n.addChild(child)
          //这两个顺序不能变
          //新孩子节点是通配符，当前节点设置为true
         n.wildChild = true
          //现在将节点设置为他的新孩子
         n = child
         n.priority++
                //如果路径没有以通配符结尾，那么
                //将是另一个以"/"开头的非通配符子路径
          //也就是path中除了通配符段还有path，继续循环
         if len(wildcard) < len(path) {
            path = path[len(wildcard):]
            child := &node{
               priority: 1,
               fullPath: fullPath,
            }
            n.addChild(child)
            n = child
            continue
         }
         // Otherwise we're done. Insert the handle in the new leaf
          //将处理函数挂到叶子节点
         n.handlers = handlers
         return
      }
       //不是'：'是'*'
      // catchAll
       //星号通配符后面不能再有路径
      if i+len(wildcard) != len(path) {
         panic("catch-all routes are only allowed at the end of the path in path '" + fullPath + "'")
      }
                //与已有路径冲突
      if len(n.path) > 0 && n.path[len(n.path)-1] == '/' {
         panic("catch-all conflicts with existing handle for the path segment root in path '" + fullPath + "'")
      }
                 //星号通配符的前一个字符必须为下划线，否则panic
      // currently fixed width 1 for '/'
      i--
      if path[i] != '/' {
         panic("no / before catch-all in path '" + fullPath + "'")
      }
      n.path = path[:i]
      // First node: catchAll node with empty path
      child := &node{
         wildChild: true,
         nType:     catchAll,
         fullPath:  fullPath,
      }
      n.addChild(child)
      n.indices = string('/')
      n = child
      n.priority++
      // second node: node holding the variable
      child = &node{
         path:     path[i:],
         nType:    catchAll,
         handlers: handlers,
         priority: 1,
         fullPath: fullPath,
      }
      n.children = []*node{child}
      return
   }
   // If no wildcard was found, simply insert the path and handle
   n.path = path
   n.handlers = handlers
   n.fullPath = fullPath
}
```

寻找通配符 wildCard是指包含通配符的path段，例如path是  /:id/info  那么id代表了通配符的名字；这个方法用于查找path中是否包含 wildCard；通配支持了path上传参数，但是也增加了path设计的复杂性；如果没有通配符的设计，程序员需要定义每一个path。

```Go
func findWildcard(path string) (wildcard string, i int, valid bool) {
        for start, c := range []byte(path) {
        //如果没有遇到通配符就继续向后查找
               if c != ':' && c != '*' {
                       continue
               }
               //找到通配符设置valid为true，那么通配符在path的起始位置就是start
               valid = true
        //从通配符后面继续查找
               for end, c := range []byte(path[start+1:]) {
                       switch c {
            //如果遇到下划线，返回wildCard（不包括下划线）、start、true
                       case '/':
                               return path[start : start+1+end], start, valid
             //如果遇到通配符，valid设置为false
                       case ':', '*':
                               valid = false
                       }
               }
        //在这个位置返回，遍历完了path，valid为true和false的可能性都有
               return path[start:], start, valid
        }
    //在path里没有找到通配符
        return "", -1, false
}
```

##### getValue

```Go
// 查找并返回信息
// 大概流程就是：通过完整的路径去构成的前缀树一层一层的查找，查找到了返回它的信息（比如handler方法之类的）
func (n *node) getValue(path string, params *Params, skippedNodes *[]skippedNode, unescape bool) (value nodeValue) {
  var globalParamsCount int16
walk:
   for {
      // 获取当前节点的路径
      prefix := n.path
      if len(path) > len(prefix) {
         if path[:len(prefix)] == prefix {
            path = path[len(prefix):]
            // Try all the non-wildcard children first by matching the indices
            idxc := path[0]
            for i, c := range []byte(n.indices) {
               if c == idxc {
                  //  strings.HasPrefix(n.children[len(n.children)-1].path, ":") == n.wildChild
                  if n.wildChild {
                     index := len(*skippedNodes)
                     *skippedNodes = (*skippedNodes)[:index+1]
                     (*skippedNodes)[index] = skippedNode{
                        path: prefix + path,
                        node: &node{
                           path:      n.path,
                           wildChild: n.wildChild,
                           nType:     n.nType,
                           priority:  n.priority,
                           children:  n.children,
                           handlers:  n.handlers,
                           fullPath:  n.fullPath,
                        },
                        paramsCount: globalParamsCount,
                     }
                  }
                  n = n.children[i]
                  continue walk
               }
            }
            if !n.wildChild {
               // If the path at the end of the loop is not equal to '/' and the current node has no child nodes
               // the current node needs to roll back to last vaild skippedNode
               if path != "/" {
                  for l := len(*skippedNodes); l > 0; {
                     skippedNode := (*skippedNodes)[l-1]
                     *skippedNodes = (*skippedNodes)[:l-1]
                     if strings.HasSuffix(skippedNode.path, path) {
                        path = skippedNode.path
                        n = skippedNode.node
                        if value.params != nil {
                           *value.params = (*value.params)[:skippedNode.paramsCount]
                        }
                        globalParamsCount = skippedNode.paramsCount
                        continue walk
                     }
                  }
               }
               // Nothing found.
               // We can recommend to redirect to the same URL without a
               // trailing slash if a leaf exists for that path.
               value.tsr = path == "/" && n.handlers != nil
               return
            }
            // Handle wildcard child, which is always at the end of the array
            n = n.children[len(n.children)-1]
            globalParamsCount++
            switch n.nType {
            case param:
               // fix truncate the parameter
               // tree_test.go  line: 204
               // Find param end (either '/' or path end)
               end := 0
               for end < len(path) && path[end] != '/' {
                  end++
               }
               // Save param value
               if params != nil && cap(*params) > 0 {
                  if value.params == nil {
                     value.params = params
                  }
                  // Expand slice within preallocated capacity
                  i := len(*value.params)
                  *value.params = (*value.params)[:i+1]
                  val := path[:end]
                  if unescape {
                     if v, err := url.QueryUnescape(val); err == nil {
                        val = v
                     }
                  }
                  (*value.params)[i] = Param{
                     Key:   n.path[1:],
                     Value: val,
                  }
               }
               // we need to go deeper!
               if end < len(path) {
                  if len(n.children) > 0 {
                     path = path[end:]
                     n = n.children[0]
                     continue walk
                  }
                  // ... but we can't
                  value.tsr = len(path) == end+1
                  return
               }
               if value.handlers = n.handlers; value.handlers != nil {
                  value.fullPath = n.fullPath
                  return
               }
               if len(n.children) == 1 {
                  // No handle found. Check if a handle for this path + a
                  // trailing slash exists for TSR recommendation
                  n = n.children[0]
                  value.tsr = n.path == "/" && n.handlers != nil
               }
               return
            case catchAll:
               // Save param value
               if params != nil {
                  if value.params == nil {
                     value.params = params
                  }
                  // Expand slice within preallocated capacity
                  i := len(*value.params)
                  *value.params = (*value.params)[:i+1]
                  val := path
                  if unescape {
                     if v, err := url.QueryUnescape(path); err == nil {
                        val = v
                     }
                  }
                  (*value.params)[i] = Param{
                     Key:   n.path[2:],
                     Value: val,
                  }
               }
               value.handlers = n.handlers
               value.fullPath = n.fullPath
               return
            default:
               panic("invalid node type")
            }
         }
      }
      if path == prefix {
         // If the current path does not equal '/' and the node does not have a registered handle and the most recently matched node has a child node
         // the current node needs to roll back to last vaild skippedNode
         if n.handlers == nil && path != "/" {
            for l := len(*skippedNodes); l > 0; {
               skippedNode := (*skippedNodes)[l-1]
               *skippedNodes = (*skippedNodes)[:l-1]
               if strings.HasSuffix(skippedNode.path, path) {
                  path = skippedNode.path
                  n = skippedNode.node
                  if value.params != nil {
                     *value.params = (*value.params)[:skippedNode.paramsCount]
                  }
                  globalParamsCount = skippedNode.paramsCount
                  continue walk
               }
            }
            // n = latestNode.children[len(latestNode.children)-1]
         }
         // We should have reached the node containing the handle.
         // Check if this node has a handle registered.
         if value.handlers = n.handlers; value.handlers != nil {
            value.fullPath = n.fullPath
            return
         }
         // If there is no handle for this route, but this route has a
         // wildcard child, there must be a handle for this path with an
         // additional trailing slash
         if path == "/" && n.wildChild && n.nType != root {
            value.tsr = true
            return
         }
         // No handle found. Check if a handle for this path + a
         // trailing slash exists for trailing slash recommendation
         for i, c := range []byte(n.indices) {
            if c == '/' {
               n = n.children[i]
               value.tsr = (len(n.path) == 1 && n.handlers != nil) ||
                  (n.nType == catchAll && n.children[0].handlers != nil)
               return
            }
         }
         return
      }
      // Nothing found. We can recommend to redirect to the same URL with an
      // extra trailing slash if a leaf exists for that path
      value.tsr = path == "/" ||
         (len(prefix) == len(path)+1 && prefix[len(path)] == '/' &&
            path == prefix[:len(prefix)-1] && n.handlers != nil)
      // roll back to last valid skippedNode
      if !value.tsr && path != "/" {
         for l := len(*skippedNodes); l > 0; {
            skippedNode := (*skippedNodes)[l-1]
            *skippedNodes = (*skippedNodes)[:l-1]
            if strings.HasSuffix(skippedNode.path, path) {
               path = skippedNode.path
               n = skippedNode.node
               if value.params != nil {
                  *value.params = (*value.params)[:skippedNode.paramsCount]
               }
               globalParamsCount = skippedNode.paramsCount
               continue walk
            }
         }
      }
      return
   }
}
```

##### 小总结

通过使用 `RouterGroup` 对象，我们可以更好地管理路由，以便于维护和扩展。我们可以为不同的路由分组，共享中间件和参数，从而减少了代码的重复性，并提高了代码的可维护性。

### Context

在 Gin 框架中，`gin.Context` 是一个上下文对象，它包含了当前 HTTP 请求的所有信息和操作，如请求参数、请求头、响应状态码等等。`gin.Context` 对象是每个 HTTP 请求独立的，因此在处理请求时，我们需要从 HTTP 请求中获取到对应的 `gin.Context` 对象，然后在该对象上进行操作和处理。

```Go
func (engine *Engine) handleHTTPRequest(c *Context) {
   …………
   // Find root of the tree for the given HTTP method
   t := engine.trees
   for i, tl := 0, len(t); i < tl; i++ {
      if t[i].method != httpMethod {
         continue
      }
      root := t[i].root
      // Find route in tree
      value := root.getValue(rPath, c.params, c.skippedNodes, unescape)
      if value.params != nil {
         c.Params = *value.params
      }
      if value.handlers != nil {
         c.handlers = value.handlers
         c.fullPath = value.fullPath
         c.Next()
         c.writermem.WriteHeaderNow()
         return
      }
      ……
      break
   }
    // 如果true，当路由没有被命中时，去检查是否有其他method命中
    //  如果命中，响应405 （Method Not Allowed）
    //  如果没有命中，请求将由 NotFound handler 来处理
   if engine.HandleMethodNotAllowed {
      for _, tree := range engine.trees {
         if tree.method == httpMethod {
            continue
         }
         if value := tree.root.getValue(rPath, nil, c.skippedNodes, unescape); value.handlers != nil {
            c.handlers = engine.allNoMethod
            serveError(c, http.StatusMethodNotAllowed, default405Body)
            return
         }
      }
   }
   c.handlers = engine.allNoRoute
   serveError(c, http.StatusNotFound, default404Body)
}
```

遍历所有的树，拿到对应的处理函数，调用c.Next()开始执行

```Go
func (c *Context) Next() {
   c.index++
   for c.index < int8(len(c.handlers)) {
      c.handlers[c.index](c)
      c.index++
   }
}
```

可以看到判断条件让执行次数等于了函数链的函数个数

#### 参数获取

* QueryString Parameter 利用 net/url 的相关函数
* Param 路由树的时候就写进 context 了，直接拿
* Form , 使用想要的 decoder

#### 返回

```Go
// JSON serializes the given struct as JSON into the response body.
// It also sets the Content-Type as "application/json".
func (c *Context) JSON(code int, obj interface{}) {
   c.Render(code, render.JSON{Data: obj})
}
```

```Go
// Render writes the response headers and calls render.Render to render data.
func (c *Context) Render(code int, r render.Render) {
   c.Status(code)
   if !bodyAllowedForStatus(code) {
      r.WriteContentType(c.Writer)
      c.Writer.WriteHeaderNow()
      return
   }
   if err := r.Render(c.Writer); err != nil {
      panic(err)
   }
}
```

如果不是允许的渲染方式，返回就直接写入错误的返回 开始渲染 后面就是具体的渲染了

```Go
// Render interface is to be implemented by JSON, XML, HTML, YAML and so on.
type Render interface {
   // Render writes data with custom ContentType.
   Render(http.ResponseWriter) error
   // WriteContentType writes custom ContentType.
   WriteContentType(w http.ResponseWriter)
}
```

## 作业

要不试试自己搓一个 gin ？ 复习，然后把 websocket 的作业写完。

## 参考

[关于 GIN 的路由树 - 掘金 (juejin.cn)](https://juejin.cn/post/7111879694328266788) [数据结构与算法：字典树（前缀树） - 知乎 (zhihu.com)](https://zhuanlan.zhihu.com/p/28891541) [Golang-gin框架路由原理 - 知乎 (zhihu.com)](https://zhuanlan.zhihu.com/p/491337692) [gin 源码阅读(1) - gin 与 net/http 的关系 (qq.com)](https://mp.weixin.qq.com/s?__biz=MzAwMDY4ODg5MA==&mid=2247485557&idx=1&sn=1e7cb52fe419ba57d452e5ea663a8610&chksm=9ae45fe0ad93d6f6db92b3c7d9cf8d7bc2cef5a8c00dc1e6d34304f57a3556ae38e7bcab91b2&scene=21#wechat_redirect) [gin 源码阅读（一）-- 启动 - 知乎 (zhihu.com)](https://zhuanlan.zhihu.com/p/133208366) [gin框架之路由前缀树初始化分析 (qq.com)](https://mp.weixin.qq.com/s/lLgeKMzT4Q938Ij0r75t8Q)