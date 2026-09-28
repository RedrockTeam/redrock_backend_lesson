# Oauth &amp; OIDC

# Oauth&OIDC

# Oauth2

## 为什么需要oauth2

阿海想要将自己网盘里的存的一些精美的照片打印出来，他发现有一个"云冲印"的网站，可以在线冲印照片。目前该网站提供两个方案：


1. 用户自己上传提供照片，这要求阿海要先将照片从网盘里下载下来，然后再上传到网站
2. 将网盘的账号密码交给网站，这样就可以不用自己下载照片，直接选择好需要下载的照片让网站打印即可。 第一种方案阿海觉得这样太麻烦了，不符合他懒狗的气质。但第二种方案网站有可能会擅自查看或者保存网盘上的其他资源，阿海觉得这样太不安全了。有没有一种方法可以不需要告知账号密码又能限制网站能够获取资源的方法呢？ 接下来要讲的oauth2就是一个非常好的解决方法

## OAuth简介

OAuth是一个**开放授权标准**，允许用户授权第三方网站访问他们存储在另外的服务提供者上的信息，而不需要将用户名和密码提供给第三方网站或分享他们数据的所有内容。为了保护用户数据的安全和隐私，第三方网站访问用户数据前都需要显式的向用户征求授权。我们常见的提供OAuth认证服务的厂商有支付宝、QQ、微信等。\nOAuth协议有两个版本，这里我们只介绍2.0版本，2.0版整个授权验证流程更简单更安全，关注客户端开发者的简易性，同时为Web应用、桌面应用、手机和智能设备提供专门的认证流程，也是目前最主要的用户身份验证和授权方式。

### OAuth2工作流程

OAuth 2.0的运行流程如下图  ![](https://internal-api-drive-stream.feishu.cn/space/api/box/stream/download/authcode/?code=YjQwNTdhZmQyODU4ODVlNmYwZGJmNDQwZTAzZDM4YThfYzgwODYxNDZhYzMxODdhNjdhZjQyYWVkZWNiNGJiOWFfSUQ6NzIwNjM0OTQwNjEwMTA2MTYzM18xNzgxNzk1MDkyOjE3ODE3OTg2OTJfVjM)

> 名词解释：

* **client**：第三方应用程序，代指任何消费资源服务器的第三方应用
* **Resource Owner**：资源所有者，本文中又称用户（user）。
* **User Agent**：用户代理，本文中就是指浏览器。
* **Authorization server**：认证服务器，即服务提供商专门用来处理认证的服务器。
* **Resource server**：资源服务器，即服务提供商存放用户生成的资源的服务器。它与认证服务器，可以是同一台服务器，也可以是不同的服务器。 （A）用户打开客户端以后，客户端要求用户给予授权。 （B）用户同意给予客户端授权。 （C）客户端使用上一步获得的授权，向认证服务器申请令牌。 （D）认证服务器对客户端进行认证以后，确认无误，同意发放令牌。 （E）客户端使用令牌，向资源服务器申请获取资源。 （F）资源服务器确认令牌无误，同意向客户端开放资源。 这其中最重要的东西是**Access Token**，它是整个OAuth2的核心。**Access Token**是客户端可以在资源服务器访问用户的哪些信息的权限令牌。这一般包含三类信息：
* 客户端标识
* 用户标识
* 客户端能访问资源所有者的哪些资源以及其相应的权限 有了这三类信息，那么资源服务器就可以区分出来是哪个客户端要访问哪个用户的哪些资源（以及有没有权限）。

### OAuth2的授权模式

客户端必须得到用户的授权（authorization grant），才能获得令牌（access token）。OAuth 2.0定义了四种授权方式。

* 授权码模式（authorization code）
* 简化模式（implicit）
* 密码模式（resource owner password credentials）
* 客户端模式（client credentials） 这里我们只讲授权码模式

#### authorization code

授权码模式（authorization code）是功能最完整、流程最严密的授权模式。它的特点就是通过客户端的后台服务器，与"服务提供商"的认证服务器进行互动。  ![](https://internal-api-drive-stream.feishu.cn/space/api/box/stream/download/authcode/?code=Yjc5M2UwMTY3N2E2YjA3MWQxZTA5MWU2OTUwOTM2YmZfZDRmYTZmOTI1YTk4Nzk1YjI3ODJmODZlMjhkNzdiOTNfSUQ6NzIwNjM0OTQwNjE3NjYwODI1N18xNzgxNzk1MDkyOjE3ODE3OTg2OTJfVjM) 它的步骤如下： （A）用户访问客户端，后者将前者导向认证服务器 （B）用户选择是否给予客户端授权 （C）假设用户给予授权，认证服务器将用户导向客户端事先指定的重定向URI，同时附上一个授权码 （D）客户端收到授权码，附上早先的"重定向URI"，向认证服务器申请令牌。这一步是在客户端的后台的服务器上完成的，对用户不可见 （E）认证服务器核对了授权码和重定向URI，确认无误后，向客户端发送访问令牌和可选的更新令牌 A步骤中，客户端申请认证的URI，包含以下参数：

* response_type：表示授权类型，必选项，此处的值固定为"code"
* client_id：表示客户端的ID，必选项
* redirect_uri：表示重定向URI，可选项
* scope：表示申请的权限范围，可选项
* state：**推荐项**，客户端应当指定一个随机字符串，认证服务器会原封不动地返回这个值

```HTTP
GET http://oauth.com/authorize?
    response_type=code
    &client_id=1
    &state=xyz
    &redirect_uri=https%3A%2F%2Fclient%2Eexample%2Ecom%2Foauth2
    &scope=user,photo
```

C步骤中，服务器回应客户端的URI，包含以下参数：

* code：表示授权码，必选项。该码的有效期应该很短，通常设为10分钟，客户端只能使用该码一次，否则会被授权服务器拒绝。该码与客户端ID和重定向URI，是一一对应关系。
* state：如果客户端的请求中包含这个参数，认证服务器的回应也必须一模一样包含这个参数。

```HTTP
HTTP/1.1 302 Found
Location: https://client.example.com/callback?
          code=SplxlOBeZQQYbYS6WxSbIA
          &state=dfDFjmlFV6PGF1963
```

D步骤中，客户端向认证服务器申请令牌的HTTP请求，包含以下参数：

* grant_type：表示使用的授权模式，必选项，此处的值固定为"authorization_code"
* code：表示上一步获得的授权码，必选项
* redirect_uri：表示重定向URI，必选项，且必须与A步骤中的该参数值保持一致
* client_id：表示客户端ID，必选项
* client_secret：表示客户端密钥，必选项。该参数是保密的，因此只能在后端发请求

```HTTP
POST http://oauth.com/token HTTP/1.1
Content-Type: application/x-www-form-urlencoded
grant_type=authorization_code&code=SplxlOBeZQQYbYS6WxSbIA
&redirect_uri=https%3A%2F%2Fclient%2Eexample%2Ecom%2Fcb&client_id=111111&client_secret=22222
```

E步骤中，认证服务器发送的HTTP回复，包含以下参数：

* access_token：表示访问令牌，必选项。
* token_type：表示令牌类型，该值大小写不敏感，必选项，可以是bearer类型或mac类型。
* expires_in：表示过期时间，单位为秒。如果省略该参数，必须其他方式设置过期时间。
* refresh_token：表示更新令牌，用来获取下一次的访问令牌，可选项。
* scope：表示权限范围，如果与客户端申请的范围一致，此项可省略。

```HTTP
HTTP/1.1 200 OK
Content-Type: application/json;charset=UTF-8
{
        "access_token":"2YotnFZFEjr1zCsicMWpAA",
        "token_type":"example",
        "expires_in":3600,
        "refresh_token":"tGzv3JOkF0XG5Qx2TlKWIA",
        "example_parameter":"example_value"
}
```

拿到Access Token后，客户端的后端就可以向资源服务器获取用户的资源。

### 刷新令牌

在上述得到访问令牌时，一般会提供一个过期时间和刷新令牌。以便在访问令牌过期失效的时候可以由客户端自动获取新的访问令牌，而不是让用户再次登陆授权。那么问题来了，是否可以把过期时间设置的无限大呢，答案是可以的，但是不推荐。如下是刷新令牌的收客户端需要提供给Authorization Server的参数：

* grant_type：必选。固定值"refresh_token"。
* refresh_token：必选。客户端得到access_token的同时拿到的刷新令牌
* scope：表示申请的授权范围，不可以超出上一次申请的范围，如果省略该参数，则表示与上一次一致
* client_id：表示客户端ID，必选项
* client_secret：表示客户端密钥，必选项

```HTTP
POST http://oauth.com/token HTTP/1.1
Content-Type: application/x-www-form-urlencoded
grant_type=refresh_token&refresh_token=tGzv3JOkF0XG5Qx2TlKWIA&client_id=111111&client_secret=22222
```

**这里留给几个问题大家思考：**

* 为什么需要code验证码？
* 授权码模式，步骤A中如果不设置state/设置固定的state会有什么后果？

# OIDC（OpenID Connect）

## Authentication(认证) 与 Authorization(授权)

[OAuth 2.0](http://tools.ietf.org/html/rfc6749) 规范定义了一个**授权（delegation）协议，对于使用Web的应用程序和API在网络上传递授权决策**非常有用。OAuth被用在各钟各样的应用程序中，包括提供用户认证的机制。这导致许多的开发者和API提供者得出一个OAuth本身是一个**认证**协议的错误结论，并将其错误的使用于此。让我们再次明确的指出：

> **OAuth2.0 不是认证协议。** **OAuth2.0 不是认证协议。** **OAuth2.0 不是认证协议。** 接下来我们将解决这样一个问题：我有OAuth2，并且我需要身份认证，该怎么办？ 在用户访问一个应用程序的上下文环境中，**认证**会告诉应用程序**当前用户是谁**以及其是否存在。一个完整的认证协议可能还会告诉你一些关于此用户的相关属性，比如唯一标识符、电子邮件地址以及应用程序说"早安"时所需要的内容。认证是关于应用程序中存在的用户，而互联网规模的认证协议需要能够跨网络和安全边界来执行此操作。 在OAuth 中，身份认证通常发生在颁发Access Token的之前，因此使用Access Token作为身份认证的证明是非常诱人的。然而,，仅仅拥有一个Access Token并没有告诉Client任何东西。在OAuth 中，token被设计为对Client不透明，OAuth没有告诉应用程序上述任何信息，但在用户身份认证的上下文环境中，Client需要能够从token中派生一些信息。 这还不是Oauth不能用于认证的最大问题，基于OAuth 身份API的最大问题在于，即使使用完全符合OAuth的机制，不同的提供程序不可避免的会使用不同的方式实现身份API。比如，在一个提供程序中，用户标识符可能是用user_id字段来表示的，但在另外的提供程序中则是用subject字段来表示的。即使这些语义是等效的，也需要两份代码来处理。换句话说，虽然发生在每个提供程序中的授权是相同的，但是身份认证信息的传输可能是不同的。

## 什么是OIDC

OIDC是OpenID Connect的简称，它在OAuth2上构建了一个身份层，是一个基于OAuth2协议的身份认证标准协议。我们都知道OAuth2是一个授权协议，它无法提供完善的身份认证功能，OIDC使用OAuth2的授权服务器来为第三方客户端提供用户的身份认证，并把对应的身份认证信息传递给客户端，且可以适用于各种类型的客户端（比如服务端应用，移动APP，JS应用），且完全兼容OAuth2，也就是说你搭建了一个OIDC的服务后，也可以当作一个OAuth2的服务来用。  ![](https://internal-api-drive-stream.feishu.cn/space/api/box/stream/download/authcode/?code=NjZiMDEyYmFkYTU5MDk0MGUxOWVmZDljZDkwYmY5YTJfNjE3MmVjNDg5ZTE4YTlkYWJlZjQ0OThkNDViM2FkZGJfSUQ6NzIwNjM0OTQwMzk3MDYwMDk2Ml8xNzgxNzk1MDkyOjE3ODE3OTg2OTJfVjM) OIDC本身是由多个规范构成，其中包含一个核心的规范，多个可选支持的规范来提供扩展支持，简单的来看一下：


1. [Core](http://openid.net/specs/openid-connect-core-1_0.html)：必选。定义OIDC的核心功能，在OAuth 2.0之上构建身份认证，以及如何使用Claims来传递用户的信息。
2. [Discovery](http://openid.net/specs/openid-connect-discovery-1_0.html)：可选。发现服务，使客户端可以动态的获取OIDC服务相关的元数据描述信息（比如支持那些规范，接口地址是什么等等）。
3. [Dynamic Registration](http://openid.net/specs/openid-connect-registration-1_0.html) ：可选。动态注册服务，使客户端可以动态的注册到OIDC的OP（这个缩写后面会解释）。
4. [OAuth 2.0 Multiple Response Types](http://openid.net/specs/oauth-v2-multiple-response-types-1_0.html) ：可选。针对OAuth2的扩展，提供几个新的response_type。
5. [OAuth 2.0 Form Post Response Mode](http://openid.net/specs/oauth-v2-form-post-response-mode-1_0.html)：可选。针对OAuth2的扩展，OAuth2回传信息给客户端是通过URL的querystring和fragment这两种方式，这个扩展标准提供了一基于form表单的形式把数据post给客户端的机制。
6. [Session Management](http://openid.net/specs/openid-connect-session-1_0.html) ：可选。Session管理，用于规范OIDC服务如何管理Session信息。
7. [Front-Channel Logout](http://openid.net/specs/openid-connect-frontchannel-1_0.html)：可选。基于前端的注销机制，使得RP（这个缩写后面会解释）可以不使用OP的iframe来退出。
8. [Back-Channel Logout](http://openid.net/specs/openid-connect-backchannel-1_0.html)：可选。基于后端的注销机制，定义了RP和OP直接如何通信来完成注销。 除了上面这8个之外，还有其他的正在制定中的扩展。看起来是挺多的，不要被吓到，其实并不是很复杂，除了Core核心规范内容多一点之外，另外7个都是很简单且简短的规范，另外Core是基于OAuth2的，也就是说其中很多东西在复用OAuth2，所以说你理解了OAuth2之后，OIDC就是非常容易理解的了，我们这里就只关注OIDC引入了哪些新的东西（Core，其余7个可选规范不做介绍，但是可能会提及到）。 下图是官方给出的一个OIDC组成结构图，我们暂时只关注Core的部分，其他的部分了解是什么东西就可以了，当作黑盒来用。 ![](https://internal-api-drive-stream.feishu.cn/space/api/box/stream/download/authcode/?code=YzE2NmFlNDE5ZTViYzBjOTIwOTNhOGZkYzI2YmIzMmVfMTRjNWZhZmZlODEyMDM1MjlmODQyNWUyY2M3OWU5NTVfSUQ6NzIwNjM0OTQwNjM1MjcxOTg3M18xNzgxNzk1MDkyOjE3ODE3OTg2OTJfVjM)

## OIDC核心流程

主要的术语以及概念介绍（完整术语参见[Terminology](http://openid.net/specs/openid-connect-core-1_0.html#Terminology)）：

* EU(End User)：一个人类用户
* RP(Relying Party)：用来代指OAuth2中的受信任的客户端，身份认证和授权信息的消费方
* OP(OpenID Provider)：有能力提供EU认证的服务（比如OAuth2中的授权服务），用来为RP提供EU的身份认证信息
* ID Token：JWT格式的数据，包含EU身份认证的信息
* UserInfo Endpoint：用户信息接口（受OAuth2保护），当RP使用Access Token访问时，返回授权用户的信息，此接口必须使用HTTPS
* \n**OIDC** 被抽象为以下5个步骤


1. RP发送一个认证请求给OP；
2. OP对EU进行身份认证，然后提供授权；
3. OP把ID Token和Access Token（需要的话）返回给RP；
4. RP使用Access Token发送一个请求UserInfo EndPoint；
5. UserInfo EndPoint返回EU的Claims。

```Plaintext
+--------+                                   +--------+
|        |                                   |        |
|        |---------(1) AuthN Request-------->|        |
|        |                                   |        |
|        |  +--------+                       |        |
|        |  |        |                       |        |
|        |  |  End-  |<--(2) AuthN & AuthZ-->|        |
|        |  |  User  |                       |        |
|   RP   |  |        |                       |   OP   |
|        |  +--------+                       |        |
|        |                                   |        |
|        |<--------(3) AuthN Response--------|        |
|        |                                   |        |
|        |---------(4) UserInfo Request----->|        |
|        |                                   |        |
|        |<--------(5) UserInfo Response-----|        |
|        |                                   |        |
+--------+                                   +--------+
```

上图取自Core规范文档，其中AuthN=Authentication，表示认证；AuthZ=Authorization，代表授权。注意这里面RP发往OP的请求，是属于**Authentication**类型的请求，虽然在OIDC中是复用OAuth2的**Authorization**请求通道，但是用途是不一样的，且OIDC的AuthN请求中**scope参数必须要有一个值为的openid的参数**（后面会详细介绍AuthN请求所需的参数），用来区分这是一个OIDC的**Authentication**请求，而不是OAuth2的**Authorization**请求。 上面提到过**OIDC对OAuth2最主要的扩展就是提供了ID Token**。ID Token是一个安全令牌，是一个授权服务器提供的包含用户信息（由一组Cliams构成以及其他辅助的Cliams）的JWT格式的数据结构。ID Token的主要构成部分如下（使用OAuth2流程的OIDC）。


 1. iss = Issuer Identifier：必须。提供认证信息者的唯一标识。一般是一个https的url（不包含querystring和fragment部分）。
 2. sub = Subject Identifier：必须。iss提供的EU的标识，在iss范围内唯一。它会被RP用来标识唯一的用户。最长为255个ASCII个字符。
 3. aud = Audience(s)：必须。标识ID Token的受众。必须包含OAuth2的client_id。
 4. exp = Expiration time：必须。过期时间，超过此时间的ID Token会作废不再被验证通过。
 5. iat = Issued At Time：必须。JWT的构建的时间。
 6. auth_time = AuthenticationTime：EU完成认证的时间。如果RP发送AuthN请求的时候携带max_age的参数，则此Claim是必须的。
 7. nonce：RP发送请求的时候提供的随机字符串，用来减缓重放攻击，也可以来关联ID Token和RP本身的Session信息。
 8. acr = Authentication Context Class Reference：可选。表示一个认证上下文引用值，可以用来标识认证上下文类。
 9. amr = Authentication Methods References：可选。表示一组认证方法。
10. azp = Authorized party：可选。结合aud使用。只有在被认证的一方和受众（aud）不一致时才使用此值，一般情况下很少使用。 ID Token通常情况下还会包含其他的Claims。另外ID Token必须使用JWS进行签名和JWE加密，从而提供认证的完整性、不可否认性以及可选的保密性。一个ID Token的例子如下：

```JSON
{
        "iss": "https://server.example.com",
        "sub": "24400320",
        "aud": "s6BhdRkqt3",
        "nonce": "n-0S6_WzA2Mj",
        "exp": 1311281970,
        "iat": 1311280970,
        "auth_time": 1311280969,
        "acr": "urn:mace:incommon:iap:silver"
}
```

## UserInfo Endpoint

上述claim中只有sub是和EU相关的，这在一般情况下是不够的，必须还需要EU的用户名，头像等其他的资料，OIDC提供了一组公共的cliams，来提供更多用户的信息，这就是——UserIndo Endpoint。 除了ID Token包含的信息之外，还定义了一个包含当前用户信息的标准的受保护的资源。如上所述，这些信息不是身份认证的一部分，而是提供附加的标识信息。比如说应用程序提示说"早上好：Jane Doe"，总比说"早上好：9XE3-JI34-00132A"要友好的多。它提供了一组标准化的属性：比如profile、email、phone和address。OpenId Connect定义了一个特殊的openid scope，可以通过access token来开启ID Token的颁发以及对UserInfo Endpoint的访问。它可以和其他scope一起使用而不发生冲突。这允许OpenId Connect和OAuth平滑的共存。 在RP得到Access Token后可以请求此资源，然后获得一组EU相关的Claims，这些信息可以说是ID Token的扩展，ID Token中只需包含EU的唯一标识sub即可（避免ID Token过于庞大和暴露用户敏感信息），然后在通过此接口获取完整的EU的信息。此资源**必须部署在TLS之上**。 其中一个例子如下：

```JSON
{
   "sub": "248289761001",
   "name": "Jane Doe",
   "given_name": "Jane",
   "family_name": "Doe",
   "preferred_username": "j.doe",
   "email": "janedoe@example.com",
   "picture": "http://example.com/janedoe/me.jpg"
}
```

其中sub代表EU的唯一标识，这个claim是必须的，其他的都是可选的。

## OIDC的认证流程

解释完了ID Token是什么，下面就看一下OIDC如何获取到ID Token，因为OIDC基于OAuth2，所以OIDC的认证流程主要是由OAuth2的几种授权流程延伸而来的，有以下3种：

* Authorization Code Flow：使用OAuth2的授权码来换取Id Token和Access Token
* Implicit Flow：使用OAuth2的Implicit流程获取Id Token和Access Token
* Hybrid Flow：混合Authorization Code Flow+Implici Flow

# sso

## 背景

在企业发展初期，企业使用的系统很少，通常一个或者两个，每个系统都有自己的登录模块，运营人员每天用自己的账号登录，很方便。\n但随着企业的发展，用到的系统随之增多，运营人员在操作不同的系统时，需要多次登录，而且每个系统的账号都不一样，这对于运营人员来说，很不方便。于是，就想到是不是可以在一个系统登录，其他系统就不用登录了呢？这就是单点登录要解决的问题。 单点登录英文全称Single Sign On，简称就是SSO。它的解释是：**在多个应用系统中，只需要登录一次，就可以访问其他相互信任的应用系统。**  ![](https://internal-api-drive-stream.feishu.cn/space/api/box/stream/download/authcode/?code=MTI2ZWM3YzRhY2E0ODU5ZGMxYmI3MmIxODlkZTE0MTFfYmMwMmE4YjBjY2JlOGRlOTI0ZTc4MzJhMDdiY2JiOWFfSUQ6NzIwNjM0OTQwMjk4MDc0NTIxOF8xNzgxNzk1MDkyOjE3ODE3OTg2OTJfVjM) 如图所示，图中有4个系统，分别是Application1、Application2、Application3、和SSO。Application1、Application2、Application3没有登录模块，而SSO只有登录模块，没有其他的业务模块，当Application1、Application2、Application3需要登录时，将跳到SSO系统，SSO系统完成登录，其他的应用系统也就随之登录了。这完全符合我们对单点登录（SSO）的定义。

## 技术实现

在说单点登录（SSO）的技术实现之前，我们先说一说普通的登录认证机制。  ![](https://internal-api-drive-stream.feishu.cn/space/api/box/stream/download/authcode/?code=NGUxNDM0MTQ1OGE1NGZmMDc2ODY4MjZhNmY5M2YxZTBfNGM3ZjA4NWM5OGI5OWRjMmQwOWQyZWRhOWM2YWZlMDZfSUQ6NzIwNjM0OTQwNDQyNzU4MzQ5MF8xNzgxNzk1MDkyOjE3ODE3OTg2OTJfVjM) 如上图所示，我们在浏览器（Browser）中访问一个应用，这个应用需要登录，我们填写完用户名和密码后，完成登录认证。这时，我们在这个用户的session中标记登录状态为yes（已登录），同时在浏览器（Browser）中写入Cookie，这个Cookie是这个用户的唯一标识。下次我们再访问这个应用的时候，请求中会带上这个Cookie，服务端会根据这个Cookie找到对应的session，通过session来判断这个用户是否登录。如果不做特殊配置，这个Cookie的名字叫做jsessionid，值在服务端（server）是唯一的。

### 同域下的单点登录

一个企业一般情况下只有一个域名，通过二级域名区分不同的系统。比如我们有个域名叫做：a.com，同时有两个业务系统分别为：app1.a.com和app2.a.com。我们要做单点登录（SSO），需要一个登录系统，叫做：sso.a.com。 我们只要在sso.a.com登录，app1.a.com和app2.a.com就也登录了。通过上面的登陆认证机制，我们可以知道，在sso.a.com中登录了，其实是在sso.a.com的服务端的session中记录了登录状态，同时在浏览器端（Browser）的sso.a.com下写入了Cookie。那么我们怎么才能让app1.a.com和app2.a.com登录呢？这里有两个问题：

* Cookie是不能跨域的，我们Cookie的domain属性是sso.a.com，在给app1.a.com和app2.a.com发送请求是带不上的。
* sso、app1和app2是不同的应用，它们的session存在自己的应用内，是不共享的。 ![](https://internal-api-drive-stream.feishu.cn/space/api/box/stream/download/authcode/?code=YmQ3MGYxYzcxYWYxZmEzZmJmZDk3NTRhYTFlNGQ2MTRfMGJjZjEwNGY5YjJkNGM3YzEzZjk0ODI2YmMwM2JmYTdfSUQ6NzIwNjM1MDA3NjM2NzcxNjM1M18xNzgxNzk1MDkyOjE3ODE3OTg2OTJfVjM) 那么我们如何解决这两个问题呢？针对第一个问题，sso登录以后，可以将Cookie的域设置为顶域，即.a.com，这样所有子域的系统都可以访问到顶域的Cookie。我们在设置Cookie时，**只能设置顶域和自己的域**，不能设置其他的域。比如：我们不能在自己的系统中给baidu.com的域设置Cookie。 Cookie的问题解决了，我们再来看看session的问题。我们在sso系统登录了，这时再访问app1，Cookie也带到了app1的服务端（Server），app1的服务端怎么找到这个Cookie对应的Session呢？这里就要把3个系统的Session共享，这样第2个问题也解决了。 同域下的单点登录就实现了，但这还不是真正的单点登录。

### 不同域下的单点登录

同域下的单点登录是巧用了Cookie顶域的特性。如果是不同域呢？不同域之间Cookie是不共享的，怎么办？ 这里我们就要说一说CAS流程了，下图是单点登录的标准流程：  ![](https://internal-api-drive-stream.feishu.cn/space/api/box/stream/download/authcode/?code=Y2Q5ZWNjZjNlNWFhNTkyM2Q4M2E3ZTI4MjI1ODEwNzlfY2Q0NjY5MTUxNjk4ZmNlOGU3Y2Y0MTgzODQyNWUyNDdfSUQ6NzIwNjM1MDMyNjY4ODA1NTMyNF8xNzgxNzk1MDkyOjE3ODE3OTg2OTJfVjM)


 1. 用户访问app系统，app系统是需要登录的，但用户现在没有登录。
 2. 跳转到CAS server，即SSO登录系统，**以后图中的CAS Server我们统一叫做SSO系统。** SSO系统也没有登录，弹出用户登录页。
 3. 用户填写用户名、密码，SSO系统进行认证后，将登录状态写入SSO的session，浏览器（Browser）中写入SSO域下的Cookie。
 4. SSO系统登录完成后会生成一个ST（Service Ticket），然后跳转到app系统，同时将ST作为参数传递给app系统。
 5. app系统拿到ST后，从后台向SSO发送请求，验证ST是否有效。
 6. 验证通过后，app系统将登录状态写入session并设置app域下的Cookie。 至此，跨域单点登录就完成了。以后我们再访问app系统时，app就是登录的。接下来，我们再看看访问app2系统时的流程。
 7. 用户访问app2系统，app2系统没有登录，跳转到SSO。
 8. 由于SSO已经登录了，不需要重新登录认证。
 9. SSO生成ST，浏览器跳转到app2系统，并将ST作为参数传递给app2。
10. app2拿到ST，后台访问SSO，验证ST是否有效。
11. 验证成功后，app2将登录状态写入session，并在app2域下写入Cookie。 这样，app2系统不需要走登录流程，就已经是登录了。SSO，app和app2在不同的域，它们之间的session不共享也是没问题的。

## 单点登出

单点登出则是指用户只要在app1上进行登出操作，则在其他业务服务器如app2上也应处于未登录状态。 单点登出功能跟单点登录功能是相对应的，旨在通过Cas Server的登出使所有的Cas Client都登出。每一个接入单点登录的服务都需要实现一个单点登出回调API，并保证单点登录服务可以调用这个地址以结束指定会话。而单点登录服务需记录用户所访问的应用列表，并提供一个统一登出地址，一旦该地址被访问，则调用涉及到的应用的登出回调API来登出当前用户的所有会话。 单点登出的流程一般如下：


1. 用户在app1上执行登出操作
2. app1访问app1的后端logout接口
3. app1的后端清除用户在app1上的局部会话，并且向SSO系统的logout接口发出请求
4. SSO系统分别清除用户的全局会话，向用户在登录过的其他系统的logout接口发出请求
5. 每个系统中的局部会话都已经销毁之后，跳转到登陆页面

# 作业

**LV1** 以github作为OAuth2的Resource server，在你的寒假作业基础上实现一个OAuth2认证登录与注册功能\n即让你的寒假作业能够通过github登录 **LV2** 自行设计实现OIDC，让你的寒假作业支持OIDC。也就是说**RP**、**OP**都是要你自己写，前端部分html5简单写下能用就行（有前端做苦力的也行） **LV3** 有余力的话自行设计一套SSO并实现 **截止日期** 下一次上课前 **格式与方式** 和上学期一样

# 参考

[RFC 6749](https://datatracker.ietf.org/doc/html/rfc6749) [RFC 7519](https://datatracker.ietf.org/doc/html/rfc7519) [OAuth 2.0](https://oauth.net/2/) [理解OAuth 2.0](https://www.ruanyifeng.com/blog/2014/05/oauth_2_0.html) [CAS](https://www.apereo.org/projects/cas) [openid.net](https://openid.net/specs/openid-connect-core-1_0.html)