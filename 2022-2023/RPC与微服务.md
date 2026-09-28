# RPC与微服务

# RPC与微服务

<callout emoji="📌"> 时间：`4月15日 (周六) 19:00 - 20:40 (GMT+8)` 地点：`红岩网校工作站 A区会议室` </callout>

## **RPC**

> *RPC (**远程过程调用***  *-**R**emote* ***P***rocedure ***C****all，**RPC**)应用于分布式计算中 ，允许允许一台计算机的程序调用另一个地址空间的子程序，就像调用本地程序一样，是一种C/S 模式。通过发送请求-接收回应进行信息交互的系统。* 既然有了HTTP为什么还要有RPC? 在 `TCP` 和 `UDP` 传输层协议的基础上产生的各种协议  ![](https://internal-api-drive-stream.feishu.cn/space/api/box/stream/download/authcode/?code=NzgyNWFlY2U3ZDAzMzM2MmU1NDE4MWYzYjc2MWUyM2RfYTdhM2M2N2I2YWUzNDRiYTg0MGQyMmY2MTkxMjA4ZWFfSUQ6NzIyMTcwMDA3MTQ1NTMyNjIzNl8xNzgxNzk1MDg0OjE3ODE3OTg2ODRfVjM) **HTTP** 协议（**H**yper **T**ext **T**ransfer **P**rotocol），又叫做**超文本传输协议**。  ![](https://internal-api-drive-stream.feishu.cn/space/api/box/stream/download/authcode/?code=YzUxZGVhMzdiYjg5NDZlYWJkYjMzNzU4N2UxODUzZmRfZjFiZWNkYjRhMWM5YTc5N2ZhODcwYmU5MTFkMmRjNTVfSUQ6NzIyMTcwMDA3MTE0OTMyMjI2OF8xNzgxNzk1MDg0OjE3ODE3OTg2ODRfVjM) 而 **RPC**（**R**emote **P**rocedure **C**all），又叫做**远程过程调用**。它本身并不是一个具体的协议，而是一种**调用方式**。 对于一般我们调用本地函数的时候。 `response:=LoccalFunction(request)` 其实没有什么特别的，但是你要是这样想，别人在一个远端服务器上写好了你这个本地函数的逻辑，要是能够直接调用的话，像调用本地函数一样，岂不是很舒服。基于这个思路，就产生了 RPC  ![](https://internal-api-drive-stream.feishu.cn/space/api/box/stream/download/authcode/?code=MzY1OGRiMzRjZmMxOTYzMWUwM2I5ODNiNTkzZTU3ZjFfNzQxMTdjNjg2NTVmMjgyZGQ0YzgwNGY5NmE4MWVlYzlfSUQ6NzIyMTcwMDA3MTk0NTk2MTQ3NF8xNzgxNzk1MDg0OjE3ODE3OTg2ODRfVjM) 回到之前的问题，既然有了HTTP而且还那么方便，为什么还要RPC呢 对于历史而言 TCP产生于70年代，HTTP流行于90年代。所以 在这个空隙，也就是80年代 RPC协议出现了 *在 20 世纪 80 年代初期，传奇的*[*施乐 Palo Alto 研究中心* ](https://en.wikipedia.org/wiki/PARC_(company))*发布了基于 Cedar 语言的 RPC 框架 Lupine，并实现了世界上第一个基于 RPC 的商业应用 Courier* 这样的话，另外一个问题就出现了 既然有了RPC为啥还要HTTP呢？ 对于现在的电脑上的各类软件 比如QQ,他们都作为客户端 需要和服务端建立连接发送消息。对于这种 `Client/Server (C/S)` 模式的架构，各自的厂家有自己的通信协议，就可以实现自己的RPC来通信 但是对于一类特殊的软件，浏览器，他们不仅需要和自家服务器通信，还需要和别人家的服务器通信，因此就需要一个统一的协议来实现通信，HTTP 就是用于统一 `Browser/Server (B/S) `的协议。

### RPC和HTTP

对于两者，有着明显的区别。

#### **服务发现**

为了建立连接，必须要知道对方的IP地址和端口号，对于HTTP而言 通过默认端口加上DNS就可以实现。 而对于RPC,则需要专门的中间服务来保存这一部分的信息，被称为配置中心，提供服务注册和服务发现的作用

* 服务注册:一个服务在运行的时候需要告诉中间服务自己的地址和端口，以便客户端能够连接到服务端。
* 服务发现：一个调用者需要先从注册中心获取到服务端的地址和端口，以便建立连接。

#### **传输内容**

```
    HTTP的传输内容包含头部和Body两部分，这里就不再展示了，对于RPC而言，大体上也是两部分，控制信息和负载信息。
    对于计算机和计算机网络而言，他们只认识01串，为了实现网络传输，需要将交互双方所涉及的数据转换为某种事先约定好的中立数据流格式来进行传输，将数据流转换回不同语言中对应的数据类型来进行使用，就是序列化与反序列化，这样的方案现在也有很多现成的，比如 `Json，Protobuf`。
```

 ![](https://internal-api-drive-stream.feishu.cn/space/api/box/stream/download/authcode/?code=NDY3ZjEzZTc2NGJmZTExNTY4MWVhMGNjNGI3MzAwNDRfMGViMTY1MGI5MDllOGNjY2M5Mzk0NzA4M2JkNjgzMDZfSUQ6NzIyMTcwMDA3MTE0OTMwNTg4NF8xNzgxNzk1MDg0OjE3ODE3OTg2ODRfVjM) 对于HTTP而言，他的头部和Body包含很多控制信息，对于RPC而言具备更高的定制化程度，可以采用体积更加小的Protobuf来实现序列化 HTTP:  ![](https://internal-api-drive-stream.feishu.cn/space/api/box/stream/download/authcode/?code=OGZiYjVmYTE4MWIwYjBlOGMxMmU4ZGIwYWNiNjE4N2NfYjM5MDQxZTYzMWE2YjdiOGJlMjIxM2JjZjc4NWVkMDVfSUQ6NzIyMTcwMDA3MTM4NDAyMzA0Ml8xNzgxNzk1MDg0OjE3ODE3OTg2ODRfVjM) RPC：  ![](https://internal-api-drive-stream.feishu.cn/space/api/box/stream/download/authcode/?code=NDlkNjQ4MTk3YjA1YjFmMjVkZjc4ZmNjOTdhMmFiZDFfYzdiMDk2OTRhZTllNjk0ZjMyODU0NjZjNTA5YTczNjNfSUQ6NzIyMTcwMDA3MTEzNjU0MjcyMl8xNzgxNzk1MDg0OjE3ODE3OTg2ODRfVjM)

#### **连接形式**

```
    对于HTTP是基于TCP协议 建立连接， 对于RPC 就可以根据自生需要建立不同的连接协议，可以是HTTP TCP UDP 甚至QUICK 还有一些其他的自定义协议。两个服务交互不是只扔个序列化数据流来表示参数和结果就行的，许多在此之外信息，譬如异常、超时、安全、认证、授权、事务，等等，都可能产生双方需要交换信息的需求。对此都需要进项考虑。
```

### Protobuf

```
    **Protocol Buffers**（简称：ProtoBuf）是一种开源跨平台的[序列化](https://zh.wikipedia.org/wiki/%E5%BA%8F%E5%88%97%E5%8C%96)数据结构的协议。其对于存储资料或在网络上进行通信的程序是很有用的。这个方法包含一个[接口描述语言](https://zh.wikipedia.org/wiki/%E6%8E%A5%E5%8F%A3%E6%8F%8F%E8%BF%B0%E8%AF%AD%E8%A8%80)，描述一些数据结构，并提供程序工具根据这些描述产生代码，这些代码将用来生成或解析代表这些数据结构的字节流。这里说grpc的那一部分。
```

#### **文件定义**

文件格式`xxx.proto` 定义语言格式[proto3](https://protobuf.dev/programming-guides/proto3/)这里的是**proto3**版本

```ProtoBuf
//声明使用的版本
syntax = "proto3";
//定义message这里的message类似go里面的结构体。
//需要注意的是需要给他们附一个唯一的数字标识。
//这个字段用于在消息的二进制格式中识别每个字段。
package xxx//定义包名 防止冲突
option go_package="./rpc"; //需要加上用于指定生成go文件的目录和package
message Request{
    int64 id=1;
    string name=2;
    repeated string order=3;//repeated 可重复的意思，可以理解为数组。
    map<string,int32> info=4; //map
    optional string opt=6;//可选字段。
    //不想再去看文章了附上链接地址：
    //https://protobuf.dev/programming-guides/proto3/
}
service server {
  rpc Serv (Request) returns ();
}
```

对于不同的proto类型转换为对应语言的类型：

#### **Protoc**

安装 protoc [github release](https://github.com/protocolbuffers/protobuf/releases) 选择合适的系统的下载下来 解压到自己喜欢的目录 然后添加环境变量到`path` 网上都有教程的 此外还需要下载两个插件 以供go代码生成

```Bash
go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.28
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.2
```

在安装好之后 写好protoc文件后就可以生成代码了

```Bash
protoc --go_out=. --go-grpc_out=. rpc.proto //关于这个指令 可以自己下去了解
```

代码结构

```Bash
➜  dev tree
.
├── rpc
│   ├── rpc_grpc.pb.go
│   └── rpc.pb.go
└── rpc.proto
```

`rpc.pb.go`  纯粹的Protocol Buffer消息代码，这包括Go语言的消息结构体和一些辅助方法 `rpc_grpc.pb.go`  gRPC服务端和客户端代码 现在我们来解析这两个文件

* rpc_grpc.pb.go 其中最主要的是这个接口类型

```Go
// ServerServer is the server API for Server service.
// All implementations must embed UnimplementedServerServer
// for forward compatibility
type ServerServer interface {
    Serv(context.Context, *Request) (*OK, error)
    mustEmbedUnimplementedServerServer()
}
```

在注册服务的函数里 将这个`srv`入注册函数

```Go
func RegisterServerServer(s grpc.ServiceRegistrar, srv ServerServer) {
    s.RegisterService(&Server_ServiceDesc, srv)
}
```

所以我们在写服务端的时候 就需要实现这个接口 这个接口包含两个方法  一个`Serv()`一个`mustEmbedUnimplementedServerServer()[必须嵌入未实现的服务器]` 前面的是我们需要实现的接口，后面的那个根据其翻译 可以知道，必须把`UnimplementedServerServer[未实现的服务器]`嵌入到结构体内

```Go
// UnimplementedServerServer must be embedded to have forward compatible implementations.
type UnimplementedServerServer struct {
}
func (UnimplementedServerServer) Serv(context.Context, *Request) (*OK, error) {
    return nil, status.Errorf(codes.Unimplemented, "method Serv not implemented")
}
func (UnimplementedServerServer) mustEmbedUnimplementedServerServer() {}
// UnsafeServerServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to ServerServer will
// result in compilation errors.
type UnsafeServerServer interface {
    mustEmbedUnimplementedServerServer()
}
```

好在他已经导出了字段，并且已经实现这个接口 当我们没有实现Serv方法的时候 在客户端调用就会出现这样的报错，提示服务没有实现

```Bash
rpc error: code = Unimplemented desc = method Serv not implemented
```

我们通过在自己定义的结构体上实现就行

```Go
type Inter struct {
    rpc.UnimplementedServerServer
}
func (t *Inter) Serv(ctx context.Context, req *rpc.Request) (*rpc.OK, error) {
    log.Println(req)
    return &rpc.OK{Ok: "OK It's from " + req.Name}, nil
}
```

服务端的整体的代码看起来是这样的

```Go
package main
import (
    "context"
    "google.golang.org/grpc"
    "grpc/rpc"
    "log"
    "net"
)
type Inter struct {
    rpc.UnimplementedServerServer
}
func (t *Inter) Serv(ctx context.Context, req *rpc.Request) (*rpc.OK, error) {
    log.Println(req)
    return &rpc.OK{Ok: "OK It's from " + req.Name}, nil
}
func main() {
    Li, err := net.Listen("tcp", ":8080")
    if err != nil {
        panic(err)
    }
    server := grpc.NewServer()
    rpc.RegisterServerServer(server, &Inter{})
    err = server.Serve(Li)
    if err != nil {
        panic(err)
        return
    }
}
```

由此一个服务端就建立起来了

* 客户端 对于客户端来说 使用同一份`proto`文件 同样生成同样的`pb`文件 `rpc_grpc.pb.go`

```Go
type ServerClient interface {
    Serv(ctx context.Context, in *Request, opts ...grpc.CallOption) (*OK, error)
}
type serverClient struct {
    cc grpc.ClientConnInterface
}
func NewServerClient(cc grpc.ClientConnInterface) ServerClient {
    return &serverClient{cc}
}
func (c *serverClient) Serv(ctx context.Context, in *Request, opts ...grpc.CallOption) (*OK, error) {
    out := new(OK)
    err := c.cc.Invoke(ctx, Server_Serv_FullMethodName, in, out, opts...)
    if err != nil {
        return nil, err
    }
    return out, nil
}
```

通过`NewServerClient()` 实例化一个`ServerClient` 调用`ServerClient`的`Serv`方法，实现服务的调用。 这里的一份简单的示例

```Go
package main
import (
    "client/rpc"
    "context"
    "google.golang.org/grpc"
    "google.golang.org/grpc/credentials/insecure"
    "log"
    "time"
)
func main() {
    con, err := grpc.Dial("localhost:8080", grpc.WithTransportCredentials(insecure.NewCredentials()))
    defer con.Close()
    if err != nil {
        panic(err)
        return
    }
    cli := rpc.NewServerClient(con)
    t := time.NewTicker(time.Second * 2)
    for {
        select {
        case <-t.C:
            re, err := cli.Serv(context.Background(), &rpc.Request{
                Id:    time.Now().Unix(),
                Name:  "cqupt",
                Order: nil,
                Info:  nil,
                Opt:   nil,
            })
            if err != nil {
                log.Println(err.Error())
            }
            log.Println("call serv return", re.Ok)
        }
    }
}
```

分别把服务端和客户端运行起来，就实现了服务的一个完整运行。 `rpc.pb.go`

```Go
type Request struct {
    state         protoimpl.MessageState
    sizeCache     protoimpl.SizeCache
    unknownFields protoimpl.UnknownFields
    Id    int64            `protobuf:"varint,1,opt,name=id,proto3" json:"id,omitempty"`
    Name  string           `protobuf:"bytes,2,opt,name=name,proto3" json:"name,omitempty"`
    Order []string         `protobuf:"bytes,3,rep,name=order,proto3" json:"order,omitempty"`                                                                                        //repeated 可重复的意思，可以理解为数组。
    Info  map[string]int32 `protobuf:"bytes,4,rep,name=info,proto3" json:"info,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"varint,2,opt,name=value,proto3"` //map
    Opt   *string          `protobuf:"bytes,6,opt,name=opt,proto3,oneof" json:"opt,omitempty"`                                                                                      //可选字段。
}
type OK struct {
    state         protoimpl.MessageState
    sizeCache     protoimpl.SizeCache
    unknownFields protoimpl.UnknownFields
    Ok string `protobuf:"bytes,1,opt,name=ok,proto3" json:"ok,omitempty"`
}
```

这里有一段代码来展示这个序列化

```Go
package main
import (
    "context"
    "encoding/json"
    "fmt"
    "google.golang.org/protobuf/proto"
    "grpc/rpc"
    "log"
)
func main() {
    var s rpc.Request
    s.Id = 1
    s.Name = "cqupt"
    s.Order = []string{
        "红烧大鲤鱼",
        "清蒸排骨",
    }
    Proto, _ := proto.Marshal(&s)
    Json, _ := json.Marshal(&s)
    fmt.Println("Length of json", len(Json), "\n Length of proto", len(Proto))
}
```

运行结果

```Bash
➜  grpc go run ./main.go
Length of json 66 
Length of proto 40
```

可以看出来`proto`的体积更小。

## **微服务简介**

```
    微服务是一种开发软件的架构和组织方法，其中软件由通过明确定义的 API 进行通信的小型独立服务组成。这些服务由各个小型独立团队负责。微服务架构使应用程序更易于扩展和更快地开发，从而加速创新并缩短新功能的上市时间。
```

* 整体式架构 早期的软件，所有功能都写在一起，这称为**单体架构**（ Monolithic Application 巨石系统）。 通过整体式架构，所有进程紧密耦合，并可作为单项服务运行。这意味着，如果应用程序的一个进程遇到需求峰值，则必须扩展整个架构。随着代码库的增长，添加或改进整体式应用程序的功能变得更加复杂。这种复杂性限制了试验的可行性，并使实施新概念变得困难。整体式架构增加了应用程序可用性的风险，因为许多依赖且紧密耦合的进程会扩大单个进程故障的影响。
* 集群 指将多台服务器集中在一起，每台服务器都实现相同的业务，做相同的事情。但是每台服务器并不是缺一不可，存在的作用主要是缓解并发压力和单点故障转移问题。可以利用一些廉价的符合工业标准的硬件构造高性能的系统。实现：高扩展、高性能、低成本、高可用！
* 微服务架构 使用微服务架构，将应用程序构建为独立的组件，并将每个应用程序进程作为一项服务运行。这些服务使用轻量级 API  通过明确定义的接口进行通信。这些服务是围绕业务功能构建的，每项服务执行一项功能。由于它们是独立运行的，因此可以针对各项服务进行更新、部署和扩展，以满足对应用程序特定功能的需求。 我觉得这张图很形象： ![](https://internal-api-drive-stream.feishu.cn/space/api/box/stream/download/authcode/?code=NzFlM2YxMjEyNmM1MTRmNDY3ZDEwY2M1YWJmNDQ2NmZfZmI2NDFjNTU0NzQyNjI1MDBkYzE3OGRiOTNkYWZiODlfSUQ6NzIyMTU2MDA2NDU4MjEzOTkwNl8xNzgxNzk1MDg0OjE3ODE3OTg2ODRfVjM) 微服务拆分： 将集中的服务按照业务逻辑拆分 实现解耦或者低耦合 ![](https://internal-api-drive-stream.feishu.cn/space/api/box/stream/download/authcode/?code=YmQzY2I0MTE2NzgyMTBmNjIzYWU2NDE1ZWM1YjJlMTNfNGJlYTY1YjZhNmNkZTkwOTk2MTQ3MjAxMDVhOTY3NTRfSUQ6NzIyMTU2MDA2NTIyNzgzMzM0NV8xNzgxNzk1MDg0OjE3ODE3OTg2ODRfVjM) **微服务的特性**
* 自主性 可以对微服务架构中的每个组件服务进行开发、部署、运营和扩展，而不影响其他服务的功能。这些服务不需要与其他服务共享任何代码或实施。各个组件之间的任何通信都是通过明确定义的 API 进行的。
* 专用性 每项服务都是针对一组功能而设计的，并专注于解决特定的问题。如果开发人员逐渐将更多代码增加到一项服务中并且这项服务变得复杂，那么可以将其拆分成多项更小的服务。 这里就像计算机网络的分层的思想。 **微服务的优势**
* 敏捷性
* 灵活扩展
* 轻松部署
* 技术自由
* 可重复用的代码
* 弹性 **康威定律** (康威法则 , Conway's Law) 是[马尔文·康威](https://zh.wikipedia.org/wiki/%E9%A9%AC%E5%B0%94%E6%96%87%C2%B7%E5%BA%B7%E5%A8%81)1967年提出的：

> *"设计系统的架构受制于产生这些设计的组织的沟通结构。"——M. Conway* 一个好的架构，不仅取决于架构模式，更加取决于组织形式，盲目的拆分服务的话，最终都要导致一个单体服务到微服务最后到分布式单体， 在微服务的设计中，不仅要考虑到服务的拆分（粒度的大小），分布式事务，基础设施（CI/CD) ，日志聚合，链路追踪)，CAP等多方面考虑。

## **go-zero**

### 介绍

在这里介绍一下一个微服务框架  [go-zero](https://go-zero.dev/cn/) go-zero 是一个集成了各种工程实践的 web 和 rpc 框架。通过弹性设计保障了大并发服务端的稳定性，经受了充分的实战检验。 go-zero 包含极简的 API 定义和生成工具 goctl，可以根据定义的 api 文件一键生成 Go, iOS, Android, Kotlin, Dart, TypeScript, JavaScript 代码，并可直接运行。 使用 go-zero 的好处：

* 轻松获得支撑千万日活服务的稳定性
* 内建级联超时控制、限流、自适应熔断、自适应降载等微服务治理能力，无需配置和额外代码
* 微服务治理中间件可无缝集成到其它现有框架使用
* 极简的 API 描述，一键生成各端代码
* 自动校验客户端请求参数合法性
* 大量微服务治理和并发工具包 ![](https://internal-api-drive-stream.feishu.cn/space/api/box/stream/download/authcode/?code=ZWNkZTM4NTBlMjYzN2JlYzcwNDUzMDE0ZTM5Yzg5NzhfODJjMDFmNDgwZTM1YWQyYTRlM2JhMjgzOWU1YWJhMzRfSUQ6NzIyMTY5OTQ0NjY5MjcwODM1Nl8xNzgxNzk1MDg0OjE3ODE3OTg2ODRfVjM) 如下图，从多个层面保障了整体服务的高可用： ![](https://internal-api-drive-stream.feishu.cn/space/api/box/stream/download/authcode/?code=NjVjMTg1OGRmYjhkZDMwYTIzZjhhN2UxZjRlMWNhMDRfMDk1ZTQ4MTY5ZmFkZDczMTVkYTAwYWZlODFlOWM1YmZfSUQ6NzIyMTY5OTQ5NzI3MTkxODU5NF8xNzgxNzk1MDg0OjE3ODE3OTg2ODRfVjM) 安装

```Bash
go install github.com/zeromicro/go-zero/tools/goctl@latest
```

这里按照官方示例构在他的基础之上构建一个微服务

### 情景提要

假设我们在开发一个商城项目，而开发者小明负责用户模块(user)和订单模块(order)的开发，我们姑且将这两个模块拆分成三个微服务

* 订单服务(order)提供一个查询接口
* 用户服务(user)提供一个方法供订单服务获取用户信息
* 订单添加服务（add order) 提供一个添加订单

### 设计分析

根据情景提要我们可以得知，订单是直接面向用户，通过http协议访问数据，而订单内部需要获取用户的一些基础数据，既然我们的服务是采用微服务的架构设计， 那么三个服务（user, order）就必须要进行数据交换，服务间的数据交换即服务间的通讯，到了这里，采用合理的通讯协议也是一个开发人员需要 考虑的事情，可以通过http，rpc等方式来进行通讯，这里我们选择rpc来实现服务间的通讯，相信这里我已经对"rpc服务存在有什么作用？"已经作了一个比较好的场景描述。 当然，一个服务开发前远不止这点设计分析，我们这里就不详细描述了。从上文得知，我们需要一个

* order api
* add-order api 这三部分来实现

### 实现RPC

定义rpc服务

```Bash
goctl rpc template -o get.proto
```

protoc定义

```ProtoBuf
syntax = "proto3";
package get;
option go_package="./rpc";
message Request {
  string uid = 1;
}
message Response {
  repeated string oname = 1;
}
service Get {
  rpc order(Request) returns(Response);
}
```

生成代码

```Bash
goctl rpc protoc get.proto --go_out=. --go-grpc_out=. --zrpc_out=.
```

这里的 `protoc get.proto --go_out=. --go-grpc_out=.` 这部分是protoc 的的命令 后面那部分是zrpc的

```Bash
➜  rpcs tree
.
├── etc
│   └── get.yaml
├── get
│   └── get.go
├── get.go
├── get.proto
├── go.mod
├── go.sum
├── internal
│   ├── config
│   │   └── config.go
│   ├── logic
│   │   └── orderlogic.go
│   ├── server
│   │   └── getserver.go
│   └── svc
│       └── servicecontext.go
└── rpc
    ├── get_grpc.pb.go
    └── get.pb.go
```

配置config

```YAML
Name: get.rpc
ListenOn: 0.0.0.0:8080
Etcd:
  Hosts:
  - 127.0.0.1:2379
  Key: get.rpc
DB: "root:sianao@tcp(127.0.0.1:3306)/orde?charset=utf8mb4&parseTime=True&loc=Local"
```

```Go
package config
import "github.com/zeromicro/go-zero/zrpc"
type Config struct {
    zrpc.RpcServerConf
    DB string
}
```

配置context

```Go
import (
    "gorm.io/driver/mysql"
    "gorm.io/gorm"
    "rpcs/internal/config"
)
type ServiceContext struct {
    Config config.Config
    DB     *gorm.DB
}
func NewServiceContext(c config.Config) *ServiceContext {
    db, _ := gorm.Open(mysql.Open(c.DB))
    return &ServiceContext{
        Config: c,
        DB:     db,
    }
}
```

实现get-order loginc

```Go
func (l *OrderLogic) Order(in *rpc.Request) (*rpc.Response, error) {
    var s []string
    l.svcCtx.DB.Table("orde").Select("oname").Where("uid=?", in.Uid).Find(&s)
    return &rpc.Response{Oname: s}, nil
}
```

至此 rpc服务就配置好了

### 实现REST

首先定义api api一个接口直接连接数据库提供服务，另一个接口通过调用rpc提供服务，这里只做演示，不涉及服务拆分。 生成api文档

```Bash
 goctl api -o order.api 
```

定义api

```ProtoBuf
syntax = "v1"
info (
    title:
    desc:
    author: "zhengjinkun"
    email: "sansermail@163.com"
)
type geterequest {
    Uid int `json:"uid"`
}
type getresposne {
    Message string `json:"message"`
}
type addrequest {
    Uid   string `json:"uid"`
    Oname string `json:"oname"`
}
type addresponse {
    Result string `json:"result"`
}
service order-api {
    @handler getorder
    get /order/id(geterequest) returns(getresposne)
    @handler addorder 
    post /users/create(addrequest) returns(addresponse)
}
```

生成代码

```Bash
goctl api go --api order.api --dir . 
```

```Bash
➜  rest tree 
.
├── etc
│   └── order-api.yaml
├── go.mod
├── go.sum
├── internal
│   ├── config
│   │   └── config.go
│   ├── handler
│   │   ├── addorderhandler.go
│   │   ├── getorderhandler.go
│   │   └── routes.go
│   ├── logic
│   │   ├── addorderlogic.go
│   │   └── getorderlogic.go
│   ├── svc
│   │   └── servicecontext.go
│   └── types
│       └── types.go
├── order.api
└── order.go
```

创建数据库表

```SQL
MariaDB [orde]> desc orde;
+-------+-------------+------+-----+---------+-------+
| Field | Type        | Null | Key | Default | Extra |
+-------+-------------+------+-----+---------+-------+
| uid   | varchar(25) | YES  |     | NULL    |       |
| oname | varchar(25) | YES  |     | NULL    |       |
+-------+-------------+------+-----+---------+-------+
```

配置数据库和rpc

```YAML
Name: order-api
Host: 0.0.0.0
Port: 8888
DB: "root:sianao@tcp(127.0.0.1:3306)/orde?charset=utf8mb4&parseTime=True&loc=Local"
GetRpc:
  Etcd:
    Hosts:
      - 127.0.0.1:2379
    Key: get.rpc
```

配置conf

```Go
type Config struct {
    rest.RestConf
    DB     string
    GetRpc zrpc.RpcClientConf //手动添加
}
```

为了调用`rpc`需要把两个`pb`文件再复制一份或者复制`protoc` 再生成一份 配置context

```Go
package svc
import (
    "gorm.io/driver/mysql"
    "gorm.io/gorm"
    "rest/internal/config"
)
type ServiceContext struct {
    Config config.Config
    DB     *gorm.DB
    Rpc  rpc.GetClient
}
func NewServiceContext(c config.Config) *ServiceContext {
    // 配置context
    db, err := gorm.Open(mysql.Open(c.DB))
    if err != nil {
        return nil
    }
    return &ServiceContext{
        Config: c,
        DB:     db,
        Rpc:  rpc.NewGetClient(zrpc.MustNewClient(c.GetRpc).Conn()),
    }
}
```

实现addorder-logic

```Go
func (l *AddorderLogic) Addorder(req *types.Addrequest) (resp *types.Addresponse, err error) {
    //自定义逻辑
    type Orde struct {
        Uid   string `gorm:"column:uid"`
        Oname string `gorm:"column:oname"`
    }
    re := l.svcCtx.DB.Table("orde").Create(&Orde{
        Uid:   req.Uid,
        Oname: req.Oname,
    })
    if re.RowsAffected == 1 && re.Error == nil {
        return &types.Addresponse{Result: "ok"}, nil
    }
    err = errors.New("add failed")
    return
}
```

在这里 就已经实现了addorder 的接口 再实现get-logic这里的get-order 是直接调用rpc来实现

```Go
func (l *GetorderLogic) Getorder(req *types.Geterequest) (resp *types.Getresposne, err error) {
    re, err := l.svcCtx.Rpc.Order(
        context.Background(),
        &rpc.Request{Uid: strconv.Itoa(req.Uid)})
    return &types.Getresposne{Message: re.Oname}, nil
}
```

至此，所有的服务都定义好了，虽然是一个很简单的配置，但是实现了他的rpc和rest api 启动etcd,我这里为了方便，就直接docker跑起来了。 先启动rpc服务

```Bash
➜  rpcs go run ./get.go -f etc/get.yaml
Starting rpc server at 0.0.0.0:8080...
```

可以去etcd 里面看一下这个服务注册

```Bash
/ # etcdctl get --prefix ""
get.rpc/7587869918583358725
10.20.136.169:8080e
```

启动rest服务

```Bash
➜  rest go run ./order.go -f etc/order-api.yaml
Starting server at 0.0.0.0:8888...
```

至此，实现了一个迷你的微服务。

## 作业

* Lv0 : 阅读课件，照着示例实现一个`grpc` 的服务端和客户端
* Lv1: 将`grpc client`端集成到`gin`框架
* Lv2: 尝试使用`go-zero`实现一个`api`和 RPC