# WebSocket

# WebSocket

# WebSocket

## 前言

初次接触 WebSocket 的人，都会问同样的问题：我们已经有了 HTTP 协议，为什么还需要另一个协议？它能带来什么好处？ 答案很简单，因为 HTTP 协议有一个缺陷：通信只能由客户端发起。 我们知道，通常的web应用的交互过程是：客户端发送请求，服务端接收和审核完成请求后进行处理并返回结果给客户端，然后客户端将信息呈现出来。 这种机制在处理一些简单信息传递，发送请求不频繁的应用中比较常用。但对于一些实时传递要求比较高的应用来说，比如在线聊天这样的需求来说就显得力不从心，一来是要处理实时信息，二来可能要在极短的时间内处理大量的数据。 在WebSocket之前，技术人员常采用的方法就是**轮询**(polling)和**服务器推送**（comet）技术。**服务器推送**技术就是轮询技术的改进，分为长轮询和流技术。这里不做详细介绍，感兴趣的同学可以课下自己了解一下。 举例来说


1. 我们想了解今天的天气，只能是客户端向服务器发出请求，服务器返回查询结果。HTTP 协议做不到服务器主动向客户端推送信息。这种单向请求的特点，注定了如果服务器有连续的状态变化，客户端要获知就非常麻烦。我们只能使用"轮询"：每隔一段时候，就发出一个询问，了解服务器有没有新的信息。最典型的场景就是聊天室。轮询的效率低，非常浪费资源（因为必须不停连接，或者 HTTP 连接始终打开）。
2. 很多网站为了实现推送技术，所用的技术都是轮询。轮询是在特定的的时间间隔（如每1秒），由浏览器对服务器发出HTTP请求，然后由服务器返回最新的数据给客户端的浏览器。这种传统的模式带来很明显的缺点，即浏览器需要不断的向服务器发出请求，然而HTTP请求可能包含较长的头部，其中真正有效的数据可能只是很小的一部分，显然这样会浪费很多的带宽等资源。 而WebSocket协议，能更好的节省服务器资源和带宽，并且能够更实时地进行通讯。

## HTTP 与 WebSocket 的主要区别

 ![](https://internal-api-drive-stream.feishu.cn/space/api/box/stream/download/authcode/?code=ZmJhMzQ1ZTQ5NGRhN2MyOTM4ZGRjYzhiMGI5MDg0YzBfMTM3NmY5YzA3NmQzZTkwMzM4YTRkMGFiOGEwMTY3Y2VfSUQ6NzIxMzc2MDQzNjE0Mjk4MTEyMl8xNzgxNzk1MDg2OjE3ODE3OTg2ODZfVjM)

## WebSocket是什么

WebSocket 是基于TCP/IP协议，**独立**于HTTP协议的通信协议。 WebSocket 是双向通讯，有状态的，可实现一对一，一对多双向实时响应（客户端 ⇄ 服务端）。

## 特点（优点）

* 较少的控制开销：在连接创建后，服务器和客户端之间交换数据时，用于协议控制的数据包头部相对较小；
* 更强的实时性：由于协议是全双工的，所以服务器可以随时主动给客户端下发数据。相对于 HTTP 请求需要等待客户端发起请求服务端才能响应，延迟明显更少；
* 保持连接状态：与 HTTP 不同的是，WebSocket 需要先创建连接，这就使得其成为一种有状态的协议，之后通信时可以省略部分状态信息；
* 更好的二进制支持：WebSocket 定义了二进制帧，相对 HTTP，可以更轻松地处理二进制内容；
* 可以支持扩展：WebSocket 定义了扩展，用户可以扩展协议、实现部分自定义的子协议。

## WebSocket的实现

### WebSocket Handshake

#### 客户端

客户端想要通过WebSocket协议与服务端通信时，先要确定服务端是否支持WebSocket协议，因此 WebSocket 协议的第一步是进行握手, WebSocket 握手采用 HTTP Upgrade 机制, 客户端可以发送如下所示的结构发起握手 (请注意 WebSocket 握手只允许使用 HTTP GET 方法):

```HTTP
GET /chat HTTP/1.1
Host: server.example.com
Upgrade: websocket
Connection: Upgrade
Sec-WebSocket-Key: dGhlIHNhbXBsZSBub25jZQ==
Origin: http://example.com
Sec-WebSocket-Protocol: chat, superchat
Sec-WebSocket-Version: 13
```

其中， `Sec-WebSocket-Key`, 必传, 由客户端随机生成的 16 字节值, 然后做 base64 编码, 客户端需要保证该值是足够随机, 不可被预测的 (换句话说, 客户端应使用熵足够大的随机数发生器), 在 WebSocket 协议中, 该头部字段必传, 若客户端发起握手时缺失该字段, 则无法完成握手 `Sec-WebSocket-Version`, 必传, 指示 WebSocket 协议的版本, [RFC 6455](https://tools.ietf.org/html/rfc6455) 的协议版本为 13, 在 [RFC 6455](https://tools.ietf.org/html/rfc6455) 的 Draft 阶段已经有针对相应的 WebSocket 实现, 它们当时使用更低的版本号, 若客户端同时支持多个 WebSocket 协议版本, 可以在该字段中以逗号分隔传递支持的版本列表 (按期望使用的程序降序排列), 服务端可从中选取一个支持的协议版本 `Sec-WebSocket-Protocol`, 可选, 客户端发起握手的时候可以在头部设置该字段, 该字段的值是一系列客户端希望在于服务端交互时使用的子协议 (subprotocol), 多个子协议之间用逗号分隔, 按客户端期望的顺序降序排列, 服务端可以根据客户端提供的子协议列表选择一个或多个子协议 `Sec-WebSocket-Extensions`, 可选, 客户端在 WebSocket 握手阶段可以在头部设置该字段指示自己希望使用的 WebSocket 协议拓展 并需要在 HTTP Header 中设置 Upgrade 字段, 其字段值为 websocket, 并在 Connection 字段指示 Upgrade

#### 服务端

服务端若支持 WebSocket 协议, 并同意握手, 可以返回如下所示的结构:

```HTTP
HTTP/1.1 101 Switching Protocols
Upgrade: websocket
Connection: Upgrade
Sec-WebSocket-Accept: s3pPLMBiTxaQ9kYGzzhZRbK+xOo=
Sec-WebSocket-Protocol: chat
Sec-WebSocket-Version: 13
```

服务端若支持 WebSocket 协议, 并同意与客户端握手, 则应返回 101 的 HTTP 状态码, 表示同意协议升级, 同时应设置 Upgrade 字段并将值设置为 websocket, 并将 Connection 字段的值设置为 Upgrade, 这些都是与标准 HTTP Upgrade 机制完全相同的, 除了这些以外, 服务端还应设置与 WebSocket 相关的头部字段:

* `Sec-WebSocket-Accept`, 必传, 客户端发起握手时通过 `Sec-WebSocket-Key` 字段传递了一个将随机生成的 16 字节做 base64 编码后的字符串, 服务端若接收握手, 则应将该值与 WebSocket 魔数 (Magic Number) `258EAFA5-E914-47DA- 95CA-C5AB0DC85B11`进行字符串连接, 将得到的字符串做 SHA-1 哈希, 将得到的哈希值再做 base64 编码, 最终的值便是该字段的值。并且当客户端收到服务端的握手响应后, 会做同样的运算来校验该值是否符合预期, 以便于判断服务端是否真的支持 WebSocket 协议, 设置这个环节的目的就是为了最终校验服务端对 WebSocket 协议的支持性, 因为单纯使用 Upgrade 机制, 对于一些没有正确实现 HTTP Upgrade 机制的 Web Server, 可能也会返回预期的 Upgrade, 但实际上它并不支持 WebSocket, 而引入 WebSocket 魔数并进行这一系列操作后，可以很大程度上确定服务端确实支持 WebSocket 协议
* `Sec-WebSocket-Protocol`, 可选, 若客户端在握手时传递了希望使用的 WebSocket 子协议, 则服务端可在客户端传递的子协议列表中选择其中支持的一个, 服务端也可以不设置该字段表示不希望或不支持客户端传递的任何一个 WebSocket 子协议
* `Sec-WebSocket-Extensions`, 可选, 与 Sec-WebSocket-Protocol 字段类似, 若客户端传递了拓展列表, 可服务端可从中选择其中一个做为该字段的值, 若服务端不支持或不希望使用这些扩展, 则不设置该字段
* `Sec-WebSocket-Version`, 必传, 服务端从客户端传递的支持的 WebSocket 协议版本中选择其中一个, 若客户端传递的所有 WebSocket 协议版本对服务端来说都不支持, 则服务端应立即终止握手, 并返回 HTTP 426 状态码, 同时在 Header 中设置 `Sec-WebSocket-Version` 字段向客户端指示自己所支持的 WebSocket 协议版本列表 至此，协议升级成功。

### WebSocket数据帧（frame）

升级成功后，怎么发送和接收信息呢？ WebSocket 以 frame 为单位传输数据, frame 是客户端和服务端数据传输的最小单元, 当一条消息过长时, 通信方可以将该消息拆分成多个 frame 发送, 接收方收到以后重新拼接、解码从而还原出完整的消息, 在 WebSocket 中, frame 有多种类型, frame 的类型由 frame 头部的 Opcode 字段指示, WebSocket frame 的结构如下所示:

```Plaintext
0                   1                   2                   3
0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
+-+-+-+-+-------+-+-------------+-------------------------------+
|F|R|R|R| opcode|M| Payload len |    Extended payload length    |
|I|S|S|S|  (4)  |A|     (7)     |             (16/64)           |
|N|V|V|V|       |S|             |   (if payload len==126/127)   |
| |1|2|3|       |K|             |                               |
+-+-+-+-+-------+-+-------------+ - - - - - - - - - - - - - - - +
|     Extended payload length continued, if payload len == 127  |
+ - - - - - - - - - - - - - - - +-------------------------------+
|                               |Masking-key, if MASK set to 1  |
+-------------------------------+-------------------------------+
| Masking-key (continued)       |          Payload Data         |
+-------------------------------- - - - - - - - - - - - - - - - +
:                     Payload Data continued ...                :
+ - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - +
|                     Payload Data continued ...                |
+---------------------------------------------------------------+
```

该结构的字段语义如下:

* FIN, 长度为 1 位, 该标志位用于指示当前的 frame 是否是消息分段的最后一段，为1则是最后一段，为0则表明后面还有数据包。
* RSV 1 \~ 3, 这三个字段（各一位）为保留字段, 只有在 WebSocket 扩展时用, 若不启用扩展, 则该三个字段应置为 1, 若接收方收到 RSV 1 \~ 3 不全为 0 的 frame, 并且双方没有协商使用 WebSocket 协议扩展, 则接收方应立即终止 WebSocket 连接。
* Opcode, 长度为 4 位, 该字段将指示 frame 的类型, [RFC 6455](https://tools.ietf.org/html/rfc6455) 定义的 Opcode 共有如下几种:

```Plaintext
0x0, 代表当前是一个 continuation frame
0x1, 代表当前是一个 text frame
0x2, 代表当前是一个 binary frame
0x3 ~ 7, 目前保留, 以后将用作更多的非控制类 frame
0x8, 代表当前是一个 connection close, 用于关闭 WebSocket 连接
0x9, 代表当前是一个 ping frame
0xA, 代表当前是一个 pong frame
0xB ~ F, 目前保留, 以后将用作更多的控制类 frame
```

* Mask, 长度为 1 位, 该字段是一个标志位, 用于指示 frame 的数据 (Payload) 是否使用掩码掩盖, [RFC 6455](https://tools.ietf.org/html/rfc6455) 规定当且仅当由客户端向服务端发送的 frame, 需要使用掩码覆盖, 掩码覆盖主要为了解决代理缓存污染攻击 。此位应为1。
* Payload Len, 以字节为单位指示 frame Payload 的长度, 该字段的长度可变, 可能是7位，7+16位，7+64位。
* 值为0-125，则这个数代表Payload的真实长度。
* 值为126，Payload的长度为后面两个字节形成的16位无符号整型数的值是payload的真实长度。
* 值为127，Payload的长度为后面四个字节形成的64位无符号整型数的值是payload的真实长度。
* Masking-key, 该字段为可选字段, 当 Mask 标志位为 1 时, 代表这是一个掩码覆盖的 frame, 此时 Masking-key 字段存在, 其长度为 32 位, [RFC 6455](https://tools.ietf.org/html/rfc6455) 规定**所有由客户端发往服务端的 frame 都必须使用掩码覆盖**, 即对于所有由客户端发往服务端的 frame, 该字段都必须存在, 该字段的值是由客户端使用熵值足够大的随机数发生器生成。
* Payload, 该字段的长度是任意的, 该字段即为 frame 的数据部分, 若通信双方协商使用了 WebSocket 扩展, 则该扩展数据 (Extension data) 也将存放在此处, 扩展数据 + 应用数据, 它们的长度和便为 Payload Len 字段指示的值

### WebSocket Closing Handshake

[RFC 6455](https://tools.ietf.org/html/rfc6455) 将连接关闭表述为 Closing Handshake,WebSocket 的连接关闭分为 CLOSING 和 CLOSED 两个阶段, 当发送完 Close frame 或接收到对方发来的 Close frame 后, WebSocket 连接便从 OPEN 状态转变为 CLOSING 状态, 此时可以称挥手已启动, 通信方接收到 Close frame 后应立即向对方发回 Close frame, 并关闭底层 TCP 连接, 此时 WebSocket 连接处于 CLOSED 状态。

## 应用

### 框架使用

我们通常使用[gorilla/websocket](https://github.com/gorilla/websocket)来编写相关代码。 Doc：https://godoc.org/github.com/gorilla/websocket tips：gorilla/websocket 并**不保证并发安全** 在升级websocket协议前，我们需要创捷一个类型为Upgrader的结构体

```Go
type Upgrader struct {
    // 指定websocket握手的超时时间
    HandshakeTimeout time.Duration
    // 指定 io 操作的缓存大小，如果不指定就会自动分配。
    ReadBufferSize, WriteBufferSize int
    // 写数据操作的缓存池，如果没有设置值，write buffers 将会分配到链接生命周期里。
    WriteBufferPool BufferPool
    //按顺序指定服务支持的协议，如值存在，则服务会从第一个开始匹配客户端的协议。
    Subprotocols []string
    // 指定 http 的错误响应函数，如果没有设置 Error 则，会生成 http.Error 的错误响应。
    Error func(w http.ResponseWriter, r *http.Request, status int, reason error)
    // 请求检查函数，用于统一的链接检查，以防止跨站点请求伪造。如果不检查，就设置一个返回值为true的函数。
    // 如果请求Origin标头可以接受，CheckOrigin将返回true。 如果CheckOrigin为nil，则使用安全默认值：Origin请求头存在且原始主机不等于请求主机头，则返回false
    CheckOrigin func(r *http.Request) bool
    // EnableCompression 指定服务器是否应尝试协商每个邮件压缩（RFC 7692）。 
    // 将此值设置为true并不能保证将支持压缩。 
    // 目前仅支持"无上下文接管"模式
    EnableCompression bool
}
```

然后调用Upgrader.Upgrade方法

```Go
// responseHeader包含在对客户端升级请求的响应中。 
// 使用responseHeader指定cookie（Set-Cookie）和应用程序协商的子协议（Sec-WebSocket-Protocol）。
// 如果升级失败，则升级将使用HTTP错误响应回复客户端
// 返回一个 Conn 指针，拿到他后，可使用 Conn 读写数据与客户端通信。
func (u *Upgrader) Upgrade(w http.ResponseWriter, r *http.Request, responseHeader http.Header) (*Conn, error)
```

到时候讲gin源码会讲到相关的知识，这里就简单的介绍一下这几个参数

* http.Request是http标准库里的结构体，当服务器收到请求时，可以通过调用这个结构体的方法来获得http请求里的数据。
* http.ResponseWriter是处理器用来创建 HTTP 响应的接口
* Conn是net.Conn的封装，而net.Conn是一个基本的接口类型，以数据流为向导的网络连接接口（即**面向流的通用网络连接**）。在这里你可以认为它就是TCP连接，拿到这个管道后，对这个管道进行操作就可以了

### 代码演示

十分非常很简易聊天室

```Go
package main
import (
        "github.com/gin-gonic/gin"
        "github.com/gorilla/websocket"
        "fmt"
        "log"
        "math/rand"
        "net/http"
        "strconv"
        "sync"
        "time"
)
var Upgrade = websocket.Upgrader{
        ReadBufferSize:  1024,
        WriteBufferSize: 1024,
        CheckOrigin: func(r *http.Request) bool {
                return true
        },
}
var (
        room = sync.Map{}
        lock = sync.Mutex{}
)
type client struct {
        conn     *websocket.Conn
        username string
        send     chan []byte
}
func main() {
        r := gin.Default()
        r.GET("/test", WsTest)
        r.Run()
}
func WsTest(ctx *gin.Context) {
        conn, err := Upgrade.Upgrade(ctx.Writer, ctx.Request, nil)
        if err != nil {
                log.Println("upgrade req failed, err:", err)
                ctx.JSON(http.StatusInternalServerError, "upgrade failed")
                return
        }
        rand.Seed(time.Now().UnixMicro())
        // 随机生成名字
        c := client{
                conn:     conn,
                username: "talker" + strconv.Itoa(rand.Intn(10000)+1000),
                send:     make(chan []byte, 1024),
        }
        // 防止并发
        lock.Lock()
        room.Store(c.username, c.send)
        lock.Unlock()
    // 开俩协程读写消息
        go c.Read()
        go c.Write()
}
func (c *client) Read() {
        defer func() {
                log.Printf("user %s exit the room\n", c.username)
                c.conn.Close()
        }()
        for {
                select {
                case msg := <-c.send:
                        err := c.conn.WriteMessage(websocket.TextMessage, msg)
                        if err != nil {
                                log.Println("write msg failed,err:", err)
                        }
                }
        }
}
func (c *client) Write() {
        defer func() {
                log.Printf("user %s exit the room\n", c.username)
                room.Delete(c.username)
                c.conn.Close()
        }()
        for {
                msgType, msgByte, err := c.conn.ReadMessage()
                if err != nil {
                        // 这里遇到错误一般是断开websocket链接，不管怎样，咱们关闭链接就是了
                        log.Println("read msg failed, err:", err)
                        break
                }
                // 这里只处理一个消息类型
                switch msgType {
                case websocket.TextMessage:
                        msg := []byte(fmt.Sprintf("%s %s说:%s", time.Now().Format("01/02 03:04"), c.username, string(msgByte)))
                        // 懒得改了，直接用这种方法实现同步消息
                        room.Range(func(key, value any) bool {
                                value.(chan []byte) <- msg
                                return true
                        })
                default:
                        log.Println("receive don't know msg type is ", msgType)
                        continue
                }
        }
}
```

### 不用框架？自己搓？

留成作业，给点xiou提示。

```Go
type Msg struct {
        Typ     int
        Content []byte
}
type MyConn struct {
        conn         net.Conn
        ReadLimit    int
        WriteLimit   int
        PongHandle   Handler
        PingHandle   Handler
}
type Upgrader struct {
        ReadBufferSize  int
        WriteBufferSize int
        CheckOrigin     func(r *http.Request) bool
}
func (u *Upgrader) Upgrade(w http.ResponseWriter, r *http.Request, responseHeader http.Header) (conn MyConn, err error) {
        // 创建一个MyConn
        //检查请求头 Connection
        //检查请求头 Upgrade
        //检查请求方式
        //检查请求头 Sec-Websocket-Version 是否为13
        //检查Origin是否是允许的
        //检查请求头 Sec-Websocket-Key
        // 处理 Sec-Websocket-Protocol 子协议字段
        // 处理协议拓展
        // 从http.ResponseWriter重新拿到conn
        // 调用 http.Hijacker 拿到这个连接现在开始就可以使用websocket通信了
    // Hijack的中文意思是劫持的意思。
        h, ok := w.(http.Hijacker)
        if !ok {
                err = errors.New("fail to hijacker the request")
                return
        }
        // 截获请求，建立websocket通信
        conn.conn, _, err = h.Hijack()
        if err != nil {
                return
        }
        // 回复报文 一系列请求头
        var resp []byte
        resp = append(resp, "HTTP/1.1 101 Switching Protocols\r\nUpgrade: websocket\r\nConnection: Upgrade\r\n "...)
        //Sec-WebSocket-Accept：
        //Sec-WebSocket-Protocol:
        //请求头写完别忘了换行
    resp = append.........(省略号是代表我懒得抄了)
        //将请求报文写入
        _, err = conn.conn.Write(resp)
        return
}
func (c *MyConn) ReadMsg() (m Msg, err error) {
        // 根据数据帧读取数据
        // 读取第一个字节
        firstByte := make([]byte, 1)
    // TODO 用位运算处理这些字节
    // FIN是否提示为最终消息
        // RSV1~3的协议拓展判断
        // 读取第二个字节
        // 检查是否使用mask
    // 然后掩码处理
        if mask != 1 {
                err = ErrNoMask
                return
        }
        // 处理payload len
        switch {
        case 125 >= payloadLen && payloadLen > 0:
        case payloadLen == 126:
        case payloadLen == 127:
        default:
                // 都不是？ 那发个锤
                return
        }
        // 如果你有ReadLimit这个功能 该咋搞呢
        m.Typ = int(opcode)
        switch opcode {
        case PingMessage:
                // 按照用户设置的执行
        case PongMessage:
        case TextMessage:
        case BinaryMessage:
        case CloseMessage:
        default:
        }
        return
}
func (c *MyConn) WriteMsg(m Msg) (err error) {
        // 按照数据帧写出数据
        // 消息内容
    // payloadLen怎么处理？
    // 啥时候应该消息分片？
        // 写出发送的数据
        _, err = c.conn.Write(data)
    if err != nil{
        // 写不进去，好寄
    }
        return
}
func (c *MyConn) Close() {
        // 应该怎么关闭？
}
```

## 作业

调课了所以下下周才是gin框架，这个作业就做三周（4.9前提交），能写多少就写多少。

* Lv1（必做\n参照\*\*[websocket/examples/chat](https://github.com/gorilla/websocket/tree/master/examples/chat)\*\*这个官方写的聊天室demo完善
  * 心跳
  * 防止并发问题
  * 多个房间
  * 限制发言时间
* Lv1 Plus（学有余力，选做
  * 断线重连
  * 聊天记录
  * 鉴别发言的颜色程度（可以屏蔽？发不出去？\*\* 你个 \*\*
  * 可以发表情包
  * 任何你想实现的功能（像qq群管理员有禁言功能啥的
* Lv2（最好琢磨一下\n完善框架（加入自己的思想，多学习gorilla/websocket的思想\n上面的例子简单得一批的同时还可能有错）
  * 能够运行
  * 能够连上在线测试网站，正常发送和接受消息
* Lv 2Plus（精力充沛，选做
  * 加入自己的拓展消息类型
  * 实现客户端部分

## 拓展阅读

[RFC 6455: The WebSocket Protocol (rfc-editor.org)](https://www.rfc-editor.org/rfc/rfc6455.html) [RFC6455机翻(rfc2cn.com)](https://rfc2cn.com/rfc6455.html) [RFC 7692：WebSocket 的压缩扩展 (rfc-editor.org)](https://www.rfc-editor.org/rfc/rfc7692)