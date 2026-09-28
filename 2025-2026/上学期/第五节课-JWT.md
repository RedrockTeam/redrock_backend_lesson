# 第五节课-JWT

# 前言

在第三节课，我们学习了两个http框架(gin和hertz)，你已经可以构建出一个简单的ping-pong反馈器； 在第四节课，我们学习了数据库操作，你已经能改写上面的ping-pong程序，将要响应的东西存入数据 库，或者从数据库取出需要响应的东西。

但是，只会这些还不够，你没法构建出一套完整的系统，即使这个系统只需要进行简单的增删改查操 作。我们面临如下难点：


1. 如何保证执行特定操作的用户具有特定权限？如何保证访问这个接口的用户就一定是该用户？
2. 把东西全部写在一个个 func Handler(ctx context.Context, c \*app.RequestContext){} 里 面似乎并不是很利于管理，那么如何更好的组织你想实现的业务逻辑，使得前期的开发和后期的维 护井井有条？
3. 当系统规模扩大后，我们是否还要每个接口都手动验证身份？是否可以把一些公共的逻辑（如鉴 权、日志、异常捕获）抽取出来统一处理？
4. 当项目不再是简单的单文件结构，而是涉及多个模块（用户、商品、订单等）时，如何进行合理的 项目分层和模块化，让代码既能快速迭代，又便于维护和扩展？

我们今天的教学分为如下三大块，学完今天的内容，上面的问题将迎刃而解：


1. 用户认证模型：从Cookie/Session/Token 的区别，到 JWT 结构（Header/Claim/签名）及其过期 与刷新；
2. 和中间件配合：将上面的 JWT认证流程和Gin/Hertz的中间件配合使用，这样就解决了用户鉴权的问 题；
3. 代码规范：如何从0开始新建一个项目、错误传递规范、项目结构规范等。

好的，我们开始上课。

# 1 如何实现对用户的认证？

## 1.1 为啥要认证？

首先，我们为何需要认证用户？就像你去坐飞机，机场需要确定你确实是购买了该机票的人；同理，你 访问网站时，网站也需要确定你就是你，从而为你提供专属于你的内容。

举个例子，在使用bilibili时，你只需要登录一次，下次你打开bilibili，就不需要再次登录，网站就能确定 你是谁，给你提供专属于你的服务。那服务器又怎么知道在登录之后访问的用户还是我？如果你有往后 学过计算机网络，你会知道，http请求是无状态的，每次传输之间都是独立的，没有关联性。所以，我 们需要找到一个方式，来维持我们的会话状态。

## 1.2 最早的认证方式:Cookie 与 Session

### 1.2.1 从「登录状态」的问题入手

在1.1里面，我们说了，在 HTTP 协议中，每一次请求都是独立的。 也就是说，服务器并不知道"这次请求"和上次请求是不是同一个用户发的。

举个例子： 用户在登录页面输入用户名和密码登录成功后，紧接着访问"个人主页"。 如果没有额外机制，服务器是无法知道"这次访问的用户"就是刚才登录的那个人。

所以问题是：

如何让服务器"记住"已经登录的用户？

这就引出了 Cookie 和 Session。

### 1.2.2 Cookie：保存在浏览器端的小纸条

#### 概念

Cookie 是浏览器存储在本地的小段文本数据，一般包含键值对。

通常由服务器在第一次响应时通过 Set-Cookie 头返回给浏览器。

例如：

```
Set-Cookie: sessionid=abc123; Path=/; HttpOnly
```

浏览器接收到后，会在后续对同一域名的请求里，自动带上这个 Cookie：

```
Cookie: sessionid=abc123
```

#### 特点

存在客户端（浏览器），用户可以查看、修改甚至伪造；

每次请求会自动附带给服务器；

可以设置过期时间、作用域、安全标志（如 HttpOnly、Secure）。

💡 类比： Cookie 就像"进入系统时服务器给你的小纸条"，上面写着你的"身份编号"，你每次访问都把它带回 来。

### 1.2.3 Session：存储在服务器端的状态信息

#### 概念

Session 是服务器端维护的"用户会话"；

当用户第一次登录成功时，服务器会创建一个 Session（通常存在内存或 Redis 中）；

服务器生成一个唯一标识符（SessionID），并通过 Cookie 返回给客户端。

这样：


1. 浏览器保存 sessionid ；
2. 每次请求时自动带上；
3. 服务器通过这个 id 找回对应的 Session 数据（例如用户 id、角色等）。

#### 优点

数据安全（保存在服务器端）；

可以存储较多信息；

Cookie 只保存一个 id，而不是整个状态。

### 1.2.4 问题与瓶颈：Session 的局限性

当项目变大后，问题逐渐出现：

#### 1. 分布式部署问题

多台服务器如何共享 Session 数据？

如果 Session 存在单台服务器的内存里，下一次请求被转发到另一台机器，就"找不到登录状态"。

解决办法通常是：

Session 共享（Redis）

Session 粘性（固定路由）

但这些方案都增加了复杂性。

#### 2. 跨端访问问题

如果不是浏览器访问（比如小程序、移动端、第三方接口），还会自动带 Cookie 吗？

移动端和第三方接口调用时，需要手动管理 Cookie，非常不便；

而且不同端、不同域之间存在跨域限制，Cookie 不好传递。

#### 3. 伸缩性问题

微服务架构中，每个服务都需要知道用户是谁，是否登录？

如果每个服务都依赖 Session，就要共享 Session 数据；

随着服务增多，Session 同步和验证变得繁琐。

### 1.2.5 思考过渡：为什么需要 Token？

如果我们希望服务端不再保存任何登录状态，每次请求只凭借请求头中的一段数据就能验证身份，会不 会更简单？

这段数据就是 Token。而 JWT（JSON Web Token）就是其中最常见、最标准的一种。

## 1.3 Token 思想的出现

### 1.3.1 Token 的核心思路：让客户端自己带身份信息

Token（令牌）本质上是一段字符串，里面包含了足够的信息，能够让服务器判断出：

这个请求来自谁；

是否经过授权；

Token 是否被伪造；

是否已经过期。

关键思想：

"服务器不保存状态，状态放在客户端的 Token 里。"

也就是说，用户登录成功后：


1. 服务器生成一段 Token（通常包含用户信息、时间戳等）；
2. 返回给客户端（浏览器、App、小程序等）；
3. 客户端保存起来（通常放在 localStorage、sessionStorage 或 Authorization 头，我们在使用 hertz开发时使用的是Authorization 头）；
4. 以后每次请求都带上这段 Token；
5. 服务端只需要验证 Token 是否合法，就能确认身份。

### 1.3.2 Token 与 Session 的对比理解

| 对比项 | Session 机制 | Token 机制 |
|:----|:-----------|:---------|
| 状态存储位置 | 服务器端       | 客户端      |
| 是否依赖 Cookie | 是（自动传递）    | 否（可放在任意位置） |
| 分布式支持 | 需共享 Session 或粘性路由 | 天然支持（服务无状态） |
| 跨平台 | 较困难（浏览器依赖） | 方便（适合移动端/第三方） |
| 安全性 | 较安全（服务端控制） | 需加签、防伪造  |
| 服务端扩展 | 较难         | 容易横向扩容   |

💡 可以打个比方：

Session：像是学校保存在档案室的"学生记录"，每次查身份都要回档案室；

Token：像是发给学生的"学生证"，上面有照片、有签章，保安一看就能确认真假。

### 1.3.3 Token 的基本形式与工作流程

#### 登录阶段

用户提交用户名和密码；

服务器验证成功后，生成一个 Token；

返回给客户端。

#### 访问阶段

客户端在请求时带上 Token（通常放在 HTTP Header 中）；

```
Authorization: Bearer <token>
```

服务端解析 Token 并验证合法性；

若合法，则继续处理请求，否则返回 401。

#### 无状态的好处

服务端不再存储登录状态；

任意一台服务器都能独立处理请求；

横向扩展简单；

客户端可以同时登录多个端（网页、App、小程序）。

### 1.3.4 Token 的安全性问题

由于 Token 存在客户端，一旦泄露，就等同于"身份被盗"。 因此需要额外考虑：

加密与签名：防止被伪造或篡改；

有效期控制：Token 不能永久有效；

刷新机制：过期后可以使用 refresh token 重新获取；

传输安全：始终使用 HTTPS；

存储安全：避免放在不安全的位置（如浏览器 Cookie）。

## 1.4 JWT 的出现与演化

### 1.4.1 JWT 是什么？

JWT（JSON Web Token） 是一种开放标准（RFC 7519），用于在网络应用环境中以 JSON 对象 的形 式安全地传递信息。 这些信息经过 数字签名（HMAC 或 RSA 等算法），因此是 可验证且防篡改 的。

JWT 最常见的使用场景是：

用户登录后，服务端生成 JWT 返回给客户端；

客户端保存 JWT（一般放在请求头中）；

之后每次请求都携带 JWT，服务端验证后即可识别用户身份。

### 1.4.2 JWT 的三部分结构

JWT 的结构由三部分组成，中间用 . 分隔：

```
Header.Payload.Signature 头部.载荷.签名
```

#### 1. Header（头部）

说明加密算法和类型，示例：

```json
{ "alg": "HS256", "typ": "JWT" }
```

Base64URL 编码后得到第一段。

#### 2. Payload（负载）

包含需要传递的用户信息（claims），如：

```json
{ "sub": "user_123", "name": "Alice", "role": "admin", "exp": 1731379200 }
```

常见的标准字段：

| 字段  | 含义  |
|:----|:----|
| iss | 签发者 |
| sub | 主题（用户唯一标识） |
| iat | 签发时间 |
| exp | 过期时间 |
| nbf | 生效时间 |

#### 3. Signature（签名）

用于防止被篡改。 签名的生成方式：

```
HMACSHA256(base64urlEncode(header) + "." + base64urlEncode(payload), secret)
```

解释如下：


1. 先把前两段分别做 Base64URL 编码

header ：一个 JSON（如 {"alg":"HS256","typ":"JWT"} ）。

payload ：一个 JSON（如 {"sub":"123","exp":...} ）。

分别做 Base64URL 编码（和普通 Base64 类似，但用 - 、 _ 替换 + 、 / ，且通常去掉 = ，以便 放进 URL/HTTP 头里）。


2. 用 "点" 连接这两段

得到字符串： <header_b64url>.<payload_b64url> 例子：

```
eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0IiwiZXhwIjoxNzMxMzc5MjAwfQ
```


3. 对这个字符串做 HMAC-SHA256

HMAC-SHA256 是一种"带密钥的哈希"，需要一个只有服务器知道的 secret 。

计算： HMAC_SHA256(message=<header>.<payload>, key=secret)

结果是一串二进制摘要（32 字节）。


4. 把摘要再做一次 Base64URL 编码

这就是 JWT 的第三段 Signature。

最终 JWT 是： header_b64url.payload_b64url.signature_b64url

完整 JWT 示例：

```
eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9. eyJzdWIiOiIxMjM0IiwibmFtZSI6IkFsaWNlIiwiZXhwIjoxNzMxMzc5MjAwfQ. TJVA95OrM7E2cBab30RMHrHDcEfxjoYZgeFONFh7HgQ
```

### 1.4.3 JWT 的优点与典型应用

#### ✅ 优点


1. 无状态：服务端无需保存 Session；
2. 跨平台：浏览器、App、小程序、第三方 API 均可使用；
3. 可扩展性好：天然支持分布式和微服务；
4. 传递丰富信息：Payload 可携带自定义用户数据。

#### ⚠ 缺点


1. 无法主动失效（除非实现黑名单机制）；
2. Payload 可解码（虽然有签名，但不是加密）；
3. 体积较大，每次请求头都会变长；
4. 过期与刷新机制实现复杂。

### 1.4.4 JWT 的过期与刷新机制

Access Token（短期有效）

用于实际请求；

有效期通常较短（几分钟至几小时）；

减少泄露风险。

Refresh Token（长期有效）

存放在安全位置；

当 Access Token 过期时，用 Refresh Token 获取新的 Access Token；

服务器可以在此阶段验证用户是否被封禁或登出。

这种"双 token 模型"是现代认证的主流实践方式。

## 1.5 生成与验证JWT实战

先纵览全部代码：

```go
// main.go package main

import ( "errors" "fmt" "strings" "time"

"github.com/golang-jwt/jwt/v5"

)

/* 需求：
1) 生成刷新和访问 token：GenerateTokens
2) 验证刷新 token：VerifyRefreshToken
3) 验证访问 token：VerifyAccessToken */

// 建议在生产环境通过环境变量或配置文件管理密钥 var ( accessSecret = []byte("access_secret_example_change_me") refreshSecret = []byte("refresh_secret_example_change_me") issuer = "demo.jwt.singlefile" accessTTL = 15 * time.Minute // 访问令牌有效期 refreshTTL = 7 * 24 * time.Hour // 刷新令牌有效期 )

// 自定义声明（访问/刷新可共用），并用 Type 区分 token 类型 type CustomClaims struct { UserID uint64 `json:"uid"` Role string `json:"role"` Type string `json:"type"` // "access" or "refresh" jwt.RegisteredClaims }

// GenerateTokens 生成访问/刷新 Token func GenerateTokens(userID uint64, role string) (accessToken string, refreshToken string, err error) { now := time.Now()

// Access Token accessClaims := CustomClaims{ UserID: userID, Role: role, Type: "access", RegisteredClaims: jwt.RegisteredClaims{ Issuer: issuer, Subject: fmt.Sprintf("%d", userID), Audience: []string{"user"}, ExpiresAt: jwt.NewNumericDate(now.Add(accessTTL)), NotBefore: jwt.NewNumericDate(now.Add(-5 * time.Second)), // 容忍少量时 钟偏差 IssuedAt: jwt.NewNumericDate(now), }, } accessTok := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims) accessToken, err = accessTok.SignedString(accessSecret) if err != nil { return "", "", fmt.Errorf("sign access token: %w", err) }

// Refresh Token refreshClaims := CustomClaims{ UserID: userID, Role: role,

Type: "refresh", RegisteredClaims: jwt.RegisteredClaims{ Issuer: issuer, Subject: fmt.Sprintf("%d", userID), Audience: []string{"user"}, ExpiresAt: jwt.NewNumericDate(now.Add(refreshTTL)), NotBefore: jwt.NewNumericDate(now.Add(-5 * time.Second)), IssuedAt: jwt.NewNumericDate(now), }, } refreshTok := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims) refreshToken, err = refreshTok.SignedString(refreshSecret) if err != nil { return "", "", fmt.Errorf("sign refresh token: %w", err) }

return accessToken, refreshToken, nil }

// VerifyAccessToken 验证访问 Token（支持传入裸 token 或 "Bearer xxx"） func VerifyAccessToken(tokenStr string) (*CustomClaims, error) { raw := stripBearer(tokenStr)

token, err := jwt.ParseWithClaims(raw, &CustomClaims{}, func(t *jwt.Token) (interface{}, error) { // 只接受 HS256 if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok || t.Method.Alg() != jwt.SigningMethodHS256.Alg() { return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"]) } return accessSecret, nil }, jwt.WithLeeway(5*time.Second)) if err != nil { return nil, err }

claims, ok := token.Claims.(*CustomClaims) if !ok || !token.Valid { return nil, errors.New("invalid access token") } if claims.Type != "access" { return nil, errors.New("token type mismatch: not an access token") } return claims, nil }

// VerifyRefreshToken 验证刷新 Token（支持传入裸 token 或 "Bearer xxx"） func VerifyRefreshToken(tokenStr string) (*CustomClaims, error) { raw := stripBearer(tokenStr)

token, err := jwt.ParseWithClaims(raw, &CustomClaims{}, func(t *jwt.Token) (interface{}, error) { if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok || t.Method.Alg() != jwt.SigningMethodHS256.Alg() {

return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"]) } return refreshSecret, nil }, jwt.WithLeeway(5*time.Second)) if err != nil { return nil, err }

claims, ok := token.Claims.(*CustomClaims) if !ok || !token.Valid { return nil, errors.New("invalid refresh token") } if claims.Type != "refresh" { return nil, errors.New("token type mismatch: not a refresh token") } return claims, nil }

// ---- 演示入口 ----

func main() { fmt.Println("== JWT single-file demo ==")

// 1) 生成访问/刷新 token access, refresh, err := GenerateTokens(12345, "admin") if err != nil { panic(err) } fmt.Println("\nGenerated Access Token:\n", access) fmt.Println("\nGenerated Refresh Token:\n", refresh)

// 2) 验证访问 token（模拟 Authorization 头的 Bearer 形式） fmt.Println("\n-- Verify Access Token --") if ac, err := VerifyAccessToken("Bearer " + access); err != nil { fmt.Println("Access verify error:", err) } else { fmt.Printf("Access OK. uid=%d role=%s exp=%s\n", ac.UserID, ac.Role, ac.ExpiresAt.Time.Format(time.RFC3339)) }

// 3) 验证刷新 token fmt.Println("\n-- Verify Refresh Token --") if rc, err := VerifyRefreshToken(refresh); err != nil { fmt.Println("Refresh verify error:", err) } else { fmt.Printf("Refresh OK. uid=%d role=%s exp=%s\n", rc.UserID, rc.Role, rc.ExpiresAt.Time.Format(time.RFC3339)) }

// 4) （可选）模拟过期校验：将 accessTTL 改很短或手动解析并检查 claims.ExpiresAt fmt.Println("\nDone.") }

// stripBearer 去掉可能的 "Bearer " 前缀 func stripBearer(s string) string {

if strings.HasPrefix(strings.ToLower(strings.TrimSpace(s)), "bearer ") { return strings.TrimSpace(s[len("Bearer "):]) } return strings.TrimSpace(s) }
```

为了更易于理解，我将代码拆成一块一块的来讲。这个示例聚焦于生成和验证本身，所以采用单文件的 形式。

第2-11行就不讲了，基本的库导入。

第21-27行：

```go
var ( accessSecret = []byte("access_secret_example_change_me") refreshSecret = []byte("refresh_secret_example_change_me") issuer = "demo.jwt.singlefile" accessTTL = 15 * time.Minute // 访问令牌有效期 refreshTTL = 7 * 24 * time.Hour // 刷新令牌有效期 )
```

我在这里初始化了一些变量，其中 accessSecret 和 refreshSecret 分别代表AccessToken和 RefreshToken的Secret。这是服务器生成Token签名的依据，并且这个东西不能泄露，否则签名就会被 别人伪造。

issuer 是用来标明 token 的"签发者身份"的，方便溯源Token的来源。这样如果某人用别的系统生成了 签名正确但 iss 不对的 token，你就可以直接拒绝。

第30-35行：

```go
type CustomClaims struct { UserID uint64 `json:"uid"` Role string `json:"role"` Type string `json:"type"` // "access" or "refresh" jwt.RegisteredClaims }
```

这里是我自定义的生成Token用的结构体，其中 UserID 和 Role 分别代表用户id和用户职责， Type 代表 Token类型，用于区分AccessToken和RefreshToken。

但是其实，这里的type就算去掉，也能完成二者的区分。因为我生成二者所用的secret不同，如果混用 二者，会直接校验失败。这里的type字段主要是为了增强可维护性和拓展性：


1. 防止以后有人把 secret 写成一样的

一开始你是两个 secret，后来某人图方便改成同一个： 这时如果没有 Type 校验，Access/Refresh 在代码逻辑上就可以混用了。

有 Type ：就算 secret 一样， claims.Type != "access"/"refresh" 也会拦住。


2. 帮助日志排查和调试

打印日志时能看到："收到一个 type=refresh 的 token 却被拿去访问受保护资源"，比光看 exp/uid 直观多了。


3. 对未来架构变更更友好

比如以后你改成 RS256，用同一套公钥来验证两种 token，只靠密钥已经没办法区分类型了， Type 就变成了硬需求。

或者以后做统一 Token 解析服务（一个函数解析各种 token），靠 Type 决定后续逻辑更清 晰。

那么最后这个 jwt.RegisteredClaims 字段又是干啥的呢？

我使用的是GoLand，我将鼠标悬停在 jwt.RegisteredClaims 上面能看到它的结构(这个小技巧你一定 要会，写项目时很有帮助)：

提取出来是这样的：

```go
type RegisteredClaims struct { Issuer string `json:"iss,omitempty"` // 签发者 Subject string `json:"sub,omitempty"` // 用户ID/主题 Audience ClaimStrings `json:"aud,omitempty"` // 接收方 ExpiresAt *NumericDate `json:"exp,omitempty"` // 过期时间 NotBefore *NumericDate `json:"nbf,omitempty"` // 生效时间 IssuedAt *NumericDate `json:"iat,omitempty"` // 签发时间 ID string `json:"jti,omitempty"` // Token ID(唯一) }
```

这些字段全部来自 JWT 标准 RFC 7519，是 JWT 标准里定义的一组"官方字段"。它的作用是：提供 JWT 中最常用、最规范化的那些 Claim 字段（如过期时间、签发时间、签发者等），避免你自己重复造轮 子。

在上面的代码里：

```go
type CustomClaims struct { UserID uint64 `json:"uid"` Role string `json:"role"` Type string `json:"type"` jwt.RegisteredClaims // <--- 就是把官方字段嵌入进来 }
```

这相当于：

我的 Token 既能包含我自定义的信息（UserID、Role）

又能自动带上 JWT 标准字段（exp、iat、iss 等）

这是最推荐的使用方式。

当然，你也可以不用它，但是没必要。

接着我们来看第37-82行：

```go
// GenerateTokens 生成访问/刷新 Token func GenerateTokens(userID uint64, role string) (accessToken string, refreshToken string, err error) { now := time.Now()

// Access Token accessClaims := CustomClaims{ UserID: userID, Role: role, Type: "access", RegisteredClaims: jwt.RegisteredClaims{ Issuer: issuer, Subject: fmt.Sprintf("%d", userID), Audience: []string{"user"}, ExpiresAt: jwt.NewNumericDate(now.Add(accessTTL)), NotBefore: jwt.NewNumericDate(now.Add(-5 * time.Second)), // 容忍少量时 钟偏差 IssuedAt: jwt.NewNumericDate(now), }, } accessTok := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims) accessToken, err = accessTok.SignedString(accessSecret) if err != nil { return "", "", fmt.Errorf("sign access token: %w", err) }

// Refresh Token refreshClaims := CustomClaims{ UserID: userID, Role: role, Type: "refresh", RegisteredClaims: jwt.RegisteredClaims{ Issuer: issuer, Subject: fmt.Sprintf("%d", userID),

Audience: []string{"user"}, ExpiresAt: jwt.NewNumericDate(now.Add(refreshTTL)), NotBefore: jwt.NewNumericDate(now.Add(-5 * time.Second)), IssuedAt: jwt.NewNumericDate(now), }, } refreshTok := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims) refreshToken, err = refreshTok.SignedString(refreshSecret) if err != nil { return "", "", fmt.Errorf("sign refresh token: %w", err) }

return accessToken, refreshToken, nil }
```

1\.函数整体干嘛用的？

```go
// GenerateTokens 生成访问/刷新 Token func GenerateTokens(userID uint64, role string) (accessToken string, refreshToken string, err error) { ... }
```

作用： 给一个用户（ userID + role ）生成一对 JWT：

accessToken : 访问接口用的短期令牌（有效期短）

refreshToken : 专门用来"续命"生成新 access token 的长期令牌（有效期长）

返回值有三个：

accessToken string

refreshToken string

err error

2\.拿当前时间

```go
now := time.Now()
```

之后所有 ExpiresAt / IssuedAt / NotBefore 都要依赖这个时间。

好处：统一时间基准，方便调试和测试。


3. 构造 Access Token 的 Claims

```go
accessClaims := CustomClaims{ UserID: userID, Role: role, Type: "access", RegisteredClaims: jwt.RegisteredClaims{ Issuer: issuer, Subject: fmt.Sprintf("%d", userID), Audience: []string{"user"}, ExpiresAt: jwt.NewNumericDate(now.Add(accessTTL)), NotBefore: jwt.NewNumericDate(now.Add(-5 * time.Second)), // 容忍少量时钟偏 差 IssuedAt: jwt.NewNumericDate(now), }, }
```

3\.1 CustomClaims 是啥？

一般我们会自定义一个结构体，比如：

```go
type CustomClaims struct { UserID uint64 `json:"user_id"` Role string `json:"role"` Type string `json:"type"` // "access" 或 "refresh" jwt.RegisteredClaims }
```

前三项是你自己业务里的字段

jwt.RegisteredClaims 是 JWT 标准里的一些通用字段（iss、sub、exp、nbf、iat 等）

3\.2 自己的业务字段

```go
UserID: userID, Role: role, Type: "access",
```

UserID ：标识是哪一个用户

Role ：用户角色，比如 "admin" / "user"

Type ：标识这个 token 是 access 还是 refresh，这样校验时可以区分。

3\.3 标准字段 RegisteredClaims

```go
RegisteredClaims: jwt.RegisteredClaims{ Issuer: issuer, Subject: fmt.Sprintf("%d", userID), Audience: []string{"user"}, ExpiresAt: jwt.NewNumericDate(now.Add(accessTTL)), NotBefore: jwt.NewNumericDate(now.Add(-5 * time.Second)), IssuedAt: jwt.NewNumericDate(now), },
```

逐个看：

Issuer ：谁签发的（你的系统名字），比如 "my-auth-service"

Subject ：这个 token 属于谁，这里用 userID 字符串

Audience ：发给谁用的，这里简单写 \["user"\]

ExpiresAt ：过期时间

now.Add(accessTTL) ：当前时间 + 一个短的有效期，比如 15 分钟

NotBefore ：在这个时间点之前，token 无效

now.Add(-5 \* time.Second) ：比现在早 5 秒，容忍一点前后端时间差（有些服务器时间不 完全同步）

IssuedAt ：签发时间，就是 now

4\.生成 Access Token 并用 HS256 签名

```go
accessTok := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims) accessToken, err = accessTok.SignedString(accessSecret) if err != nil { return "", "", fmt.Errorf("sign access token: %w", err) }
```

jwt.NewWithClaims ：

指定签名算法： HS256 （HMAC-SHA256，对称加密）

指定载荷： accessClaims

SignedString(accessSecret) ：

用一个服务器端保存的密钥字符串 accessSecret 对 token 进行签名

一定要妥善保管，不能泄露到前端或公开仓库

失败的话：

返回空字符串和错误： return "", "", fmt.Errorf("sign access token: %w", err)

5\.构造 Refresh Token 的 Claims

```go
refreshClaims := CustomClaims{ UserID: userID, Role: role, Type: "refresh", RegisteredClaims: jwt.RegisteredClaims{ Issuer: issuer, Subject: fmt.Sprintf("%d", userID), Audience: []string{"user"}, ExpiresAt: jwt.NewNumericDate(now.Add(refreshTTL)), NotBefore: jwt.NewNumericDate(now.Add(-5 * time.Second)), IssuedAt: jwt.NewNumericDate(now), }, }
```

和 access 基本一样，有两个关键区别：


1. Type: "refresh"

方便后面校验时，如果是 refresh token 就只允许干"刷新 token"这件事。


2. ExpiresAt: now.Add(refreshTTL)

refreshTTL 通常会比 accessTTL 长很多，比如 7 天、30 天。

6\.生成 Refresh Token 并签名

```go
refreshTok := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims) refreshToken, err = refreshTok.SignedString(refreshSecret) if err != nil { return "", "", fmt.Errorf("sign refresh token: %w", err) }
```

和 access token 完全一样的流程，只是用了：

不同的 claims ( refreshClaims )

一般会用不同的密钥 refreshSecret （更安全）


7. 正常返回两个 Token

```go
return accessToken, refreshToken, nil
```

一切顺利的话，把两个 token 字符串都返回给调用方。

接下来看第85-107行：

```go
func VerifyAccessToken(tokenStr string) (*CustomClaims, error) { raw := stripBearer(tokenStr)

token, err := jwt.ParseWithClaims(raw, &CustomClaims{}, func(t *jwt.Token) (interface{}, error) { // 只接受 HS256 if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok || t.Method.Alg() != jwt.SigningMethodHS256.Alg() { return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"]) } return accessSecret, nil }, jwt.WithLeeway(5*time.Second)) if err != nil { return nil, err }

claims, ok := token.Claims.(*CustomClaims) if !ok || !token.Valid { return nil, errors.New("invalid access token") } if claims.Type != "access" { return nil, errors.New("token type mismatch: not an access token") } return claims, nil }
```


1. 这个函数是干嘛的？

```go
func VerifyAccessToken(tokenStr string) (*CustomClaims, error)
```

作用： 对前端传来的 访问 token（access token） 做验证，验证通过就把里面的 CustomClaims （用户 ID、角色等）拿出来给后续业务使用。

典型场景： 在 HTTP 中间件里，拿 Authorization 头里的值调用它：

```go
auth := r.Header.Get("Authorization") // "Bearer xxx.yyy.zzz" claims, err := VerifyAccessToken(auth)
```


1. 去掉 "Bearer " 前缀

注：下面的stripBearer函数在全代码文件的最下面，为了便于这里的理解我挪到了这里。

那么这里就有同学要问了：函数放在文件的最底下，我记得我学c语言的时候必须得先声明后调用的啊， 这样语法不会出错吗？

其实，Go语言是支持先调用后定义的。所以这样写没问题。

```go
// stripBearer 去掉可能的 "Bearer " 前缀 func stripBearer(s string) string { if strings.HasPrefix(strings.ToLower(strings.TrimSpace(s)), "bearer ") { return strings.TrimSpace(s[len("Bearer "):]) } return strings.TrimSpace(s) } raw := stripBearer(tokenStr)
```

一般前端会这样带 token：

```
Authorization: Bearer xxx.yyy.zzz
```

所以后端收到的是 "Bearer xxx.yyy.zzz" ，我们要先把 "Bearer " 去掉，只要后面的那一串。


2. 用 jwt.ParseWithClaims 解析并验证

```go
token, err := jwt.ParseWithClaims(raw, &CustomClaims{}, func(t *jwt.Token) (interface{}, error) { // 只接受 HS256 if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok || t.Method.Alg() != jwt.SigningMethodHS256.Alg() { return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"]) } return accessSecret, nil }, jwt.WithLeeway(5*time.Second))
```

2\.1 参数一： raw

就是刚才去掉 "Bearer " 后的纯 token 字符串： xxx.yyy.zzz

2\.2 参数二： &CustomClaims{}

告诉 jwt 库：把 payload 解析成你自定义的 CustomClaims 结构体。

这个结构体就是前一节里定义的：

```go
type CustomClaims struct { UserID uint64 `json:"user_id"` Role string `json:"role"` Type string `json:"type"` jwt.RegisteredClaims }
```

2\.3 参数三：keyFunc 回调 —— 提供密钥并校验算法

```go
func(t *jwt.Token) (interface{}, error) { // 只接受 HS256 if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok || t.Method.Alg() != jwt.SigningMethodHS256.Alg() { return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"]) } return accessSecret, nil }
```

这是非常关键的安全点，细讲一下：


1. t.Method ：当前 token 声明使用的签名算法，比如 "HS256" 、 "HS384" 、 "RS256" 等。
2. 这里做了两层检查：

```go
if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok ...
```

必须是 HMAC 系列的方法（HS256/384/512）

```go
|| t.Method.Alg() != jwt.SigningMethodHS256.Alg()
```

并且算法名必须就是 "HS256"

这样做是为了避免有人把 JWT 的 alg 改成其他算法，比如 "none" 或者你没打算支持的算法，从 而绕过校验。


1. 如果算法检查通过：

```go
return accessSecret, nil
```

返回用来校验签名的密钥： accessSecret （跟你生成 access token 时用的是同一个）

2\.4 参数四： jwt.WithLeeway(5\*time.Second)

允许时间相关字段（ exp 、 nbf 等）有最多 5 秒的误差。

就是容忍一点服务之间的时钟不同步，避免刚好差几秒就被误判为过期或者未生效。


3. 错误处理：解析/验证失败

```go
if err != nil { return nil, err }
```

ParseWithClaims 会同时做这些事：

检查签名是否正确

检查标准 claims，如：

exp ：是否过期

nbf ：是否已到生效时间

iat ：是否在合理时间内（某些配置会用）

只要有一个不通过，就会返回 err （例如 "token is expired"），这时我们直接返回错误就行了。


4. 从 token 中拿出自定义 Claims 并进一步校验

```go
claims, ok := token.Claims.(*CustomClaims) if !ok || !token.Valid { return nil, errors.New("invalid access token") }
```

这里分两步：


1. 类型断言

```go
claims, ok := token.Claims.(*CustomClaims)
```

把 token.Claims 转成我们自己的 \*CustomClaims

如果不是这个类型（例如解析用的结构不对）， ok 会是 false


2. 检查 token 是否标记为有效

```go
!token.Valid
```

这个字段会在 ParseWithClaims 里根据各种校验结果自动设置

只要有任意一个不通过，就认为是"无效的 access token"。


5. 再次校验：必须是 access token，而不是别的类型

```go
if claims.Type != "access" { return nil, errors.New("token type mismatch: not an access token") }
```

这里用到了我们在生成 token 时自定义的字段 Type ：

在 GenerateTokens 里：

access token 是： Type: "access"

refresh token 是： Type: "refresh"

所以这里检查：

如果有人把 refresh token 拿来当访问凭证，就会被拦下来。

或者你未来还有其他类型的 token，也能靠这个字段区分用途。


6. 最终返回：解析成功、验证通过的 claims

```go
return claims, nil
```

调用方拿到 claims 后，一般会取出其中的用户id、职级等信息，供后续业务逻辑的处理。这个在下面 讲解如何与中间件配合时会讲。

接着来讲第110-131行，这里和上面的唯一区别就是第127-129行：

```go
if claims.Type != "refresh" { return nil, errors.New("token type mismatch: not a refresh token") }
```

除了识别type时需要将它当成RefreshToken来处理，其他逻辑是一模一样的，在这里就不过多赘述。

# 2 如何从0开始新建一个项目？

以Hertz框架为例(Gin也差不多)，讲讲项目结构规范、错误传递规范、以及如何从0开始新建一个项目。

## 2.1 先弄清楚项目之间的结构应该是咋样的

1 Hertz 项目的推荐分层结构（概念图）

```
API 层（Handler） │ ▼ Service 层（业务逻辑） │ ▼ DAO 层（数据访问） │ ▼ 数据库 / 缓存
```

还有三个辅助部分：

model：所有层共享的数据结构

routers：统一注册 API 路由

utils：工具函数（如 id 生成、加密、jwt）

2 各模块的职责与协作关系（重点）

下面这张图是整个框架的"协作流转图"：

```
【routers】 │ 注册路由 ▼ 【api】 │ 解析请求、返回响应（HTTP 处理） ▼ 【sv】 │ 实现业务逻辑，校验数据、组合流程 ▼ 【dao】 │ 与数据库交互（CRUD）

▼ 【model】 数据结构在各层之间流转 【utils】 各层可能会用到
```

下面逐层讲清楚它们是怎么连起来的。

3 routers 层：连接 HTTP 到 api 层

routers 只做一件事： 把 URL 和 Handler（api 层的函数）连接起来。

示意：

```go
func RegisterRoutes(h *server.Hertz) { group := h.Group("/api") group.POST("/messages", api.CreateMessage) }
```

/api/messages → api.CreateMessage

routers 层本身不写业务、不访问数据库

它是 HTTP 世界的入口。

4 api 层（Handler 层）：把 HTTP 转为业务调用

api 是 HTTP Handler，职责：


1. 解析 HTTP 请求（JSON、查询参数、Header 等）
2. 调用 service（sv）执行业务逻辑
3. 根据业务结果生成响应 JSON

示意：

```
HTTP Request → api → sv → dao → 数据库
```

api 层不写业务、不写数据操作。 它只是"对外的接口层"。

5 sv 层（Service 层）：把业务逻辑组合起来

Service 层是整个项目的"核心逻辑层"。

职责：

做业务校验

调用 DAO 获取/写入数据

组装业务流程逻辑

与 api 解耦，方便复用/测试

从协作关系看：

api 调 sv sv 调 dao

sv 层不关心 HTTP，也不关心数据库，只关心"业务应该怎么运行"。

例如"创建留言"这个动作的逻辑都属于 sv 层。

6 dao 层（Data Access Layer）：对数据库的抽象

DAO 的职责：

只负责数据存取（CRUD）

不做业务校验

不关心 HTTP 入参

不关心业务流程

DAO 是 sv 层的"工具人"： 要什么数据、查什么条件、写入什么，sv 说了算，dao 负责实现。

未来你要换 MySQL、Redis、MongoDB，只改 dao：

api 不变 sv 不变 model 不变 routers 不变

这是分层的最大优势。

7 model 层：统一数据结构定义

model 层是：

多层之间共享的数据结构

把数据结构从任一层中抽离出来

避免循环依赖

model 解决的问题：

如果不拆分 model，可能导致：

api → sv → dao 之间互相 import

产生循环引用

现在统一放 model，各层只 import model：

api → model sv → model dao → model

model 就像一个"数据中心"，定义所有结构体。

8 utils 层：工具库，所有层都可用

utils 一般包含：

id 生成器

配置加载

时间工具

JWT 工具

加密哈希

utils 不属于业务逻辑，因此：

被 api 使用

被 sv 使用

被 dao 使用

都正常。

但要注意：

utils 不应该依赖 api/sv/dao，避免形成反向依赖。

9 整体协作的实际流程（留言板示例）

假设用户请求：

```
POST /api/messages { "author": "Tom", "content": "Hello Hertz!" }
```

对应的调用链是：

① routers：匹配到 Handler

POST /api/messages → api.CreateMessage

② api：解析 HTTP + 调 sv

解析 JSON

组合输入参数

调用 sv.CreateMessage(author, content)

③ sv：执行业务逻辑

校验内容

调用 dao 写入数据：

dao.InsertMessage(...)

返回 Message 对象（model）

④ dao：真实写数据

写入数据库 / 内存

返回数据对象（model）

⑤ api：返回 HTTP 响应

把 sv 返回的 model 对象封装成 JSON

返回给客户端

10 为什么要强制 api → sv → dao 这样的分层？

总结一下优点：


1. api 与业务逻辑解耦

接口层变化（HTTP 协议、格式变了）不影响业务


2. sv 层复用性高

sv 不依赖 HTTP，意味着：

业务可以被 RPC 调用

可以直接被单元测试


3. dao 独立，方便切换数据库

把内存版换成 MySQL：

只改 dao

api、sv 全不动


4. model 统一定义数据结构，避免循环引用

严谨的大型项目都会这样做。

## 2.2 如何规范与优雅的传递错误？

PS:下面是以我自己的项目的错误传递方式为示例的，在此基础上稍作更改也是没问题的。

1 respond 包的核心设计思想

目前的 respond 包包含三类东西：

(1) Response（业务级错误）

轻量结构，描述错误与状态：

```go
type Response struct { Status string `json:"status"` Info string `json:"info"` }
```

它还实现了 error 接口：

```go
func (r Response) Error() string { return r.Info }
```

所以 Response 既是"业务语义"，又能作为 error 向上抛。

(2) FinalResponse（HTTP 最终响应）

```go
type FinalResponse struct { Status string `json:"status"` Info string `json:"info"` Data interface{} `json:"data"` }
```

只有 api 层会用它返回给前端

sv/dao 不需要知道 HTTP 格式

(3) 全局错误与内部错误构造器

如：

```go
var ( Ok = Response{Status: "10000", Info: "success"} WrongName = Response{Status: "40001", Info: "wrong username"} WrongPwd = Response{Status: "40002", Info: "wrong password"} )
```

以及内部错误：

```go
func InternalError(err error) Response { return Response{ Status: "500", Info: err.Error(), } }
```

2 错误传递规范 —— 推荐在项目中使用的标准流程

为了让项目层次清晰，我们规范：

dao → sv → api 按照固定方式传递错误 api 层统一转成 FinalResponse 返回前端

这是一套大多数中大型 Go 项目会采用的标准。

整体规则图:

```
【dao】返回 error（可能为 Response 或 InternalError） │ ▼ 【sv】判断 business error 或 internal error │ ▼ 【api】统一包装为 FinalResponse 返回前端
```

3 dao 层错误传递规范

规则：

dao 层返回两类错误：


1. 内部错误（数据库、网络、空指针等） → 用 InternalError(err) 包装
2. 正常业务错误（如用户不存在、重复） → 直接返回一个 Response（如 respond.WrongName）

例如(伪代码，理解意思为主)：

```go
if err := db.Save(&user).Error; err != nil { return InternalError(err) }

if userNotFound { return respond.WrongName }
```

dao 层不需要构造 FinalResponse（HTTP 的），也不负责决定 Status Code。

4 sv（service）层错误传递规范

规则：

sv 层只做两件事：

(1) 接住 dao 层抛上来的错误

并向上"原样 return"。

(2) 做业务校验并返回 Response 类型错误

例如：

```go
if name == "" { return respond.WrongName }

u, err := dao.GetUser(name) if err != nil { return nil, err // err 可能是 Response，也可能是 InternalError }
```

sv 层绝不返回 FinalResponse。

sv 层也不把 error 写死成 string，而是使用 Response。

5 api（handler）层错误规范（最重要）

api 层是 错误体系的最终输出口，负责：

统一处理 error

统一转成 FinalResponse

统一返回 JSON

流程如下：

```
sv returns (data, error)

if error != nil { if err 是 Response： 用 Response 转 FinalResponse else： err 是普通 error → respond.InternalError(err) }
```

示例（概念代码）：

```go
data, err := sv.Login(req.Name, req.Password) if err != nil { if r, ok := err.(respond.Response); ok { // 是业务错误 c.JSON(200, respond.Respond(r, nil)) return } // 是普通错误 → 包成内部错误 c.JSON(500, respond.Respond(respond.InternalError(err), nil)) return }

// 无错误 → OK c.JSON(200, respond.Respond(respond.Ok, data))
```

6 完整错误传递流程图（非常适合 PPT 展示）

```
┌──────────────┐ │ dao │ │ 返回 Response │ │ 或 InternalError │ └───────▲────────┘ │ error │ ┌───────┴────────┐ │ sv │ │ 不修改 error │ │ 继续上抛 │ └───────▲────────┘ │ error │ ┌───────┴────────┐ │ api │ │ 根据 error │ │ 生成 FinalResponse │ └───────▲────────┘ │ ▼ 给前端统一 JSON 格式响应
```

7 总结

在 Hertz 的错误处理体系中： dao 和 sv 使用 Response 作为业务错误，内部错误用 InternalError ，所有错误最后由 api 层统 一包装成 FinalResponse 返回给前端。 这样每一层职责单一、协作清晰、可维护性高。

## 2.3 如何从0构建起一个项目？

在掌握了项目的结构和错误的传递后，你基本上已经可以做项目了，下面推荐一下各个模块的编写顺 序，防止做项目时手忙脚乱：


1. go mod init，装好 Hertz 依赖
2. 建好目录结构： cmd/api/sv/dao/model/routers/utils
3. 先写 model ：所有层要用的结构体
4. 再写 utils/respond ：统一响应与错误
5. 写 dao 函数签名（实现可以先简单占位）
6. 写 sv ：调用 dao，实现业务逻辑
7. 写 api ：处理 HTTP 请求，调用 sv，返回 FinalResponse
8. 写 routers ：把 URL 映射到 api
9. 最后写 main ：创建 Hertz 实例，注册路由，启动

# 3 JWT认证如何在中间件中实现

## 3.1 为什么要在中间件处理JWT认证

统一入口，避免重复代码 每个需要登录的接口都要做"解析 token → 校验 → 取 userID"这一套，如果在 handler 里一遍遍 写，会非常啰嗦、难维护。放在中间件里，所有受保护的路由只要挂上这个中间件(如下方示例代码 所示，来自我的项目)，就自动拥有认证能力。

```go
package routers

import ( "OnlineMall/api" "OnlineMall/middleware" "github.com/cloudwego/hertz/pkg/app/server" )

func RegisterRouters() { h := server.Default()

userGroup := h.Group("/user") //...此处省略

//分组依据为使用对应功能需要的最低权限

h.GET("/search_products", api.SearchForProducts) //搜索商品

h.GET("/show_all_products", api.ShowAllProducts) //展示所有商品 h.GET("/show_category_products", api.ShowACategoryProducts) //展示某一类商品 h.GET("/view_product", middleware.JWTTokenAuthTokenNotAMust(), api.ShowSingleProduct) //查看商品 h.GET("/show_product_reviews", api.ShowAProductReviews) //展示商品评论 h.GET("/homepage", middleware.JWTTokenAuthTokenNotAMust(), api.ShowHomePage) //展示主页 //上方的 middleware.JWTTokenAuthTokenNotAMust() 就是验证jwt的中间件，只需写这 个h.GET()函数的中间就自动拥有验证能力 //...此处省略 h.Spin() }
```

把"认证"从"业务"里分离出来 业务 handler 应该只关心"这个用户要干什么"，而不关心"这个用户是谁、是不是合法"。 认证逻辑放在中间件：

中间件负责：鉴别身份（JWT 校验）

handler 负责：执行业务（比如发留言、删留言）

方便做统一的权限控制/日志/审计 中间件可以在放行之前，先把 userID / role 放进 context：

方便后面的 handler 做权限判断（例如：管理员才能访问某些接口）

方便日志记录"是哪个用户访问的这个接口"

## 3.2 怎么做

概括："拦头、验 token、塞用户信息、再放行"。大致步骤如下：


1. 写一个中间件函数（例如 AuthMiddleware ）

形参是框架规定的 next handler 或者 ctx, c 之类（Hertz / Gin 风格都类似）

在路由注册时，把它挂到需要保护的路由或路由组上。


2. 在中间件里拿到 Authorization 头

从 Header 中读取： Authorization

预期格式： Bearer xxx.yyy.zzz

如果没带，或者格式不对，直接返回一个未登录/未授权的响应（比如用你定义的 Response + FinalResponse 返回）。


3. 调用封装好的 VerifyAccessToken 函数

封装格式如下(具体的token处理逻辑上面讲过了，此处不再赘述)：

```go
func JWTTokenAuth() app.HandlerFunc { return func(ctx context.Context, c *app.RequestContext) { // 1. 中间件的入参： // - ctx: 上下文（可用于超时控制、跨服务传递信息等） // - c: Hertz 的 RequestContext，封装了请求和响应

// 2. 从请求头中获取 Authorization 字段

tokenString := c.GetHeader("Authorization")

// 3. 如果没有带 token，直接返回 401，并 Abort 中断后续流程 if len(tokenString) == 0 { c.JSON(consts.StatusUnauthorized, respond.MissingToken) c.Abort() // ✅ 中断：后面的 handler 不会再执行 return }

// ========================= // 这里原本应该写： // - 解析 token // - 验证签名、过期时间、类型等 // - 验证失败时： // c.JSON(consts.StatusUnauthorized, respond.InvalidToken) // c.Abort() // return // =========================

// ========================= // 验证成功后常见做法（示意）： // - 从 claims 里拿 user_id / role // - 写入到 context，让后续 handler 可以使用 // // userID := ... // c.Set("user_id", userID) // // 注意：这里我们不写具体逻辑，只说明"通常会在这里塞东西到 context" // =========================

// 4. 不调用 Abort，同时函数正常 return // => Hertz 会继续执行后续的 handler 链（真正的业务处理函数） } }
```


4. 校验成功后，把用户信息放入 context

从 claims 中拿出： UserID 、 Role 等

使用上下文（例如 context.WithValue 或框架自带的 c.Set("userID", claims.UserID) ）

这样后面的 handler 不用再解析 token，直接从 context 里拿用户信息即可。


5. 最后调用"下一个 handler"继续执行

如果一切正常，中间件调用 next(ctx, c) 或类似 API

让请求流继续传到真正的业务 handler（例如你的留言板 API）。


6. 在路由层挂中间件

比如：

公共路由： /login , /register 不需要 AuthMiddleware

需要登录才能用的路由组： /api/messages 挂上 AuthMiddleware

这样就实现了"部分接口需要 JWT 认证"的效果。

# 4 期中项目作业

经过今天课程的学习，再加上之前的知识，你已经完全具备构建CRUD项目的能力了！

下面，到你大展身手的时候了。你有2星期的时间，来完成下面这个简单的小项目。

## 4.1 项目主题

一个简单的选课系统。

## 4.2 项目需要实现的功能

用户注册

用户登录

管理员和普通用户的区分

管理员可以向数据库内新增课程，而普通用户不行

Token刷新

获取课程列表

选课(一般来说是抢课，需要用到互斥锁来保证操作的原子性)

获取已选课程列表

退课

## 4.3 注意事项

由于是小demo，各种数据结构从简，只需要能实现基本功能就行。

将项目上传到github，然后发送仓库链接到：2810873701@qq.com

截止时间：下节课(第六节课)上课前。