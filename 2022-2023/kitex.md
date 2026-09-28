# kitex

# kitex

## **概述**

 ![](https://internal-api-drive-stream.feishu.cn/space/api/box/stream/download/authcode/?code=ODQ4NjhmYWE5YjNiNWZiN2NkYzIzOGIxZDE4ZTIyMjdfM2Q1OTI4Nzc0NWZhNTIwODA5ZGQzNzIyYjdlYjMwOThfSUQ6NzI1Mjc0MDI2ODY3MDE4OTU2OV8xNzgxNzk1MDgxOjE3ODE3OTg2ODFfVjM)

## **kitex**

### **消息类型**

RPC请求一般分为：PingPong、Oneway、Streaming这几种类型

* PingPong：客户端发起一个请求后会等待一个响应才可以进行下一次请求
* Oneway：客户端发起一个请求后不等待一个响应
* Streaming：客户端发起一个或多个请求 , 等待一个或多个响应

### **序列化协议**

kitex支持thrif和protobuf这两种编码。thrift较早出现，有一些历史遗留问题。

* [begonia](https://github.com/MashiroC/begonia/tree/master)
* 网校自研
* avro序列化
* tcp上设计的自有协议

#### **thrift**

##### **协议种类**

* binary协议 相当简单的二进制编码——字段的长度和类型被编码为字节，后跟字段的实际值。
* compact协议 就是在二进制的基础上使用zigzag和varint压缩编码实现
* json json编码就非常的简易了，go的官方包里也有，一般来说，不会使用json来作为RPC序列化协议。RPC的应用场景多为分布式，高性能。
* json编解码的数据相较于其他两种方式要**更大**。
* 而且编解码的实现，也是利用反射api递归取值，这样就**更慢**了。 一般都被用作前后端交互的数据标准形式。

##### **语法参考**

* [docs-one](https://thrift.apache.org/docs/)idl可能稍微晦涩一点
* [docs-other](https://diwakergupta.github.io/thrift-missing-guide/)快速熟悉thrift协议的idl

##### **thrift文件示例**

```Thrift
 // include 'hello.thrift'
 const i8 a = 1
 typedef i8 id_typ
 enum gender {
     man = -1;
     woman =1
 }
 const gender b = gender.man
 //struct student xsd_all {
 // // xsd_all是facebook内部使用
 //}
 struct student {
     1: required i8 stuno,
     2: required string stuname,
     3: optional i8 ssex,
     optional set<string> hobby,
     optional map<i8,student> roommates
 }
 struct teacher {
     1: required i8 tid
     2: optional string tname
 }
 // 只需要其中的一个字段
 union api_response {
     student s
     teacher t
 }
 exception exp{
     i8 exp_code
     string exp_info
 }
 service  QueryCenter{
     oneway api_response findPerson(1:i8 no,2:optional string name) throws (1: required exp error)
     string ping() throws (1: optional exp error)
 }
```

##### **thrift怎么RPC的**


1. 客户端发送`Message`（类型`Call`或`Oneway`）。TMessage 包含一些元数据和要调用的方法的名称。
2. 客户端发送方法参数（由生成代码定义的结构）。
3. 服务器发送一个`Message`（类型`Reply`或`Exception`）来开始响应。
4. 服务器发送包含方法结果或异常的结构。

##### **thrift一些细节**

* [binary structure](https://github.com/apache/thrift/blob/master/doc/specs/thrift-binary-protocol.md)
* [thrift header](https://github.com/apache/thrift/blob/master/doc/specs/HeaderFormat.md)
* [kitex ttheader](https://www.cloudwego.io/zh/docs/kitex/reference/transport_protocol_ttheader/)
* [thrift rpc](https://github.com/apache/thrift/blob/master/doc/specs/thrift-rpc.md)

#### **protobuf**

略

#### **different & same points**

* thrift
  * 向前和向后的兼容性很好
  * facebook较早开源的序列化/反序列化协议，广泛应用于高性能、分布式的服务场景，支持非常多的语言绑定
  * 传输数据采用二进制格式，相对 XML 和 JSON 体积更小
  * 全套RPC解决方案，序列化、传输层、并发处理等，开箱即用
  * 文档内容太少了
* protobuf
  * 向前和向后的兼容性很好
  * 传输数据采用二进制格式，相对 XML 和 JSON 体积更小
  * 仅仅只是一个序列化/反序列化协议，并不是一个RPC框架，和GRPC配套使用。grpc的生态还是可以的
* protobuf的数据压缩率要比thrift编解码的方式效率更高

### **直连访问**

* 支持IP+Port建立连接
* 指定域名建立连接

### **连接类型**

* 短链接（不是短链，单纯表示连接不复用，用完就删了，请求就创建新的连接对象
* 长连接池，与短链接相反，就是支持连接的复用，减少TCP连接建立的开销

### **业务异常**

RPC方法直接返回没有实现`BizStatusErrorIface`或`GRPCStatusIface`的`Error`，会被kitex识别为RPC错误，也就是在RPC层面请求失败，但是实际上RPC层面上是成功的，也就是服务端收到了RPC。 而且，一旦被识别为RPC错误，可能会触发一些RPC的降级措施：服务直接熔断。所以，直接返回RPC的error需要慎重考虑！ 如果想通过RPC方法的error来传递一些业务异常，就需要实现前面的接口或者你也可以不通过error来传递业务异常，完全可以在response中携带业务异常。 更推荐后者实现，当然如果需要服务监控的话，还是使用前者。

### **预热**

预热可以避免首次请求的延迟。kitex支持客户端预热：创建客户端的时候预先初始化服务发现和连接池的相关组件，避免在首次请求时产生较大的延迟。

### **panic处理**

* 业务代码使用 go 关键字创建的 goroutine 里发生的 panic，需要业务自行 recover；受限于语言提供的能力, 无法由框架 Recover；
* 为了保证服务的稳定，Kitex 框架会自动 recover 其他所有 panic

### **请求重试**

目前有三类重试：异常重试、Backup Request，建连失败重试（默认）。其中建连失败是网络层面问题，由于请求未发出，框架会默认重试。 本文档介绍前两类重试的使用：

* 异常重试：提高服务整体的成功率
* Backup Request：减少服务的延迟波动 因为很多的业务请求不具有幂等性，这两类重试不会作为默认策略。

> 幂等性：指的是多次请求的RPC方法或其他类型的接口的结果和一次请求的结果是一样的

### **服务发现**

kitex支持很多服务发现组件，etcd、nacos、zookeeper、console等

* [kitex-etcd](https://github.com/kitex-contrib/registry-etcd)

### **Fallback**

业务在 RPC 请求失败后通常会有一些降级措施保证有效返回（比如请求超时、熔断后，构造默认返回），Kitex 的 Fallback 支持对所有异常请求进行处理。 同时，因为业务异常通常会通过 Resp（BaseResp） 返回，所以也支持对 Resp 进行处理。

#### **fallback类型**


1. **RPC** **Error**：RPC 请求异常，如超时、熔断、限流、协议等 RPC 层面的异常
2. 一般来说，我们主要是需要对RPC错误做一些响应的构造
3. **业务 Error**：业务自定义的异常，区别于 RPC 异常，具体是 [Kitex - 业务异常处理使用文档](https://www.cloudwego.io/zh/docs/kitex/tutorials/basic-feature/bizstatuserr/)
4. **Resp**：在没有使用业务异常的情况下，用户会在 Resp（BaseResp） 中定义错误返回，所以也支持对 Resp 判断做 fallback

#### **监控上报**

Fallback 后可能直接返回成功的 Resp，对用户而言是一次成功请求，但 RPC 层面还是失败请求，所以监控默认以原来的结果上报，但支持配置化调整为以 Fallback 结果上报。

### **限流**

限制上游服务对下游服务的流量，避免过载

* QPS 限流器
* 连接数限流器

### **自定义访问控制**

Kitex 框架提供了一个简单的中间件构造器，可以支持用户自定义访问控制的逻辑，在特定条件下拒绝请求。

### **兼容性问题**

* 下游的RPC服务更新
* 我们需要在上游服务中，更新受影响RPC调用。


---

## **参考**

* [quick start](https://www.cloudwego.io/zh/docs/kitex/getting-started/)
* [thrift](https://thrift.apache.org/docs/)
* [thrift docs](https://github.com/apache/thrift/tree/master/doc/specs)
* [cloudwego](https://www.cloudwego.io/zh/)

## **作业**

* 使用kitex编写简单的RPC服务

> kitex代码生成的`kitex_gen`目录，主要给客户端进行调用。
>
> 所以，如果远端编写RPC客户端时，需要 提供`kitex_gen`目录直接调用 或者 提供idl文件后，通过`kitex -module='xxx' xxx.thrift`生成`kitex_gen`。kitex是通过代码生成来进行调用的
>
> 我们需要维护idl的版本一致性，如果下游服务的idl更新了，上游的服务也需要同步，否则可能会出现RPC调用错误。
>
> 网校以前自主研发的begonia框架，有两种方式：一是代码生成调用，比较快；另一种是反射模式，这种方式有个好处就是可以不用保留代码生成文件，直接从远端拿RPC方法。当然，反射牺牲了些性能。