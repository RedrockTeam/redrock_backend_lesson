# 包-变量-循环-JSON处理

包的作用

* 使用go mod init 项目名来为你自己的项目命名并初始化项目
* 会生成go.mod（项目依赖配置）go.sum（依赖校验文件） 项目初始化后，就可以使用go mod tidy来快速地下载需要的但还没下载的第三方包，清理项目未使用的包 初始化项目后，我们可以快捷地使用本项目里的包 go语言里的包，大多数情况下是文件夹的名称 比如我在code文件夹下，使用go mod init my-project(中文意思是我的项目)  然后code文件夹下，有以下几个文件夹bob alice 。 bob文件夹下有hello.go、name.go两个储存源代码的文件 那么通常情况下这两个文件属于包"bob" hello.go的开头的package bob的意思就是说， 我属于包bob，我身体里的所有东西都是属于bob的 项目->包->函数 

 ![](uploads/115597c2-8959-4e30-9c07-03f756946100/1518506a-769f-46b1-b148-5111c3802d89/ChatGPT%20Image.png " =1402x1122")


那么如何在alice里调用bob包里的变量，常量或者是函数呢？ 首先，不管是变量，常量，还是函数，都必须保证它是对包外可见的

**在go语言中，如果一个变量名的第一个字母是大写的，就代表它可以被包外访问，如果是小写的，证明它不可以被包外访问，常量，函数，结构体等同理**

然后，我们要在improt里导入包 此时为项目命名的作用就体现出来了 比如我这个项目的名字是my-project，那么我想在alice包里的hello.go里调用bob包里的函数，就要在improt里写上improt "my-project/bob" 假如bob包里的hello.go里有个函数是这么写的

```go
package bob
func Hello(name string) string {
     return fmt.Sprint("Hello", name)
}
```

那我们在alice包里就要这么调用

```go
package alice
improt "my-projet/bob"
func printName() {
    name := "Alice"
    info := alice.Hello(name)
    fmt.Println(info)
}
```

不出所料输出结果是Hello Alice

# 变量作用域

在go语言中，没有声明此变量是全局变量还是局部变量的特殊关键词，go的开发者认为，这完全是多余的！ 所以，你可以不用了解全局变量和局部变量，你只需要知道：

* 在函数外面声明的变量，整个包都可以调用。 *注：包变量不可用:=快速赋值！这是因为直接暴露在包下的所有成员都需要声明其类型，也就是以func type var之类的类型声明语句开头*
* 在函数里面声明的变量，只有函数里面使用。
* 在代码块中声明的变量，只能在代码块中使用。 什么是代码块？ 用大括号括起来的(if和for后面的语句声明的变量只能在其代码块中使用)
* if{}
* else if{}
* else{}
* for{} 举个例子

```go
number := 2
if number == 2{
    numberInIf := number
    fmt.Println(numberInIf)
}
fmt.Println(number)
fmt.Println(numberInIf) 
//会报错，因为对于块外部来说numberInIf是未声明的变量，是陌生的
}
```

# for循环

## for循环参数数量不同时不同的语法意义

### 无参数

```go
for{
    fmt.Println("bob")
}
```

将会一直输出bob

### 一个参数

```go
for x == 5{
     fmt.Println("hello")
     x++
}
fmt.Println("bob")
```

将会在x为5时输出hello，输出一次后，将x加一，再次进行判断，直到x不等于5时，停止输出hello，继续执行下面的输出bob的语句

### 两个参数

```go
for i := 0; i < 5{
    fmt.Println("bob")
    i++
}
```

将会声明并赋值一个变量，并且当其身后的判断语句成立时，执行代码块中的代码 以上示例会输出5句bob

### 三个参数

```go
for i := 0; i < 5; i++{
    fmt.Println("bob")
}
```

将会声明并赋值一个变量，并且当其身后的判断语句成立时，执行下面的语句，执行完后执行for参数中的第三个语句，再使用第二个语句进行判断，与上个示例的效果一样 也会输出5句bob

# break continue

**break用于跳出当前的for循环，continue用于跳过后面的语句，进入下一次循环**

```go
for i := 0; i < 10; i++{
    if i == 4{
        continue
    }
    if i == 7{
        break
    }
    fmt.Println(i)
}
fmt.Println("结束输出！")
```

以上代码会输出

```bash
0
1
2
3
5
6
结束输出！
```

大家可以自己理会一下

# 错误处理

当函数的返回值类型中包含error时，你就要考虑进行错误处理了 go语言的错误处理非常方便 来看以下代码

```go
content, err := os.ReadFile("test.txt")
if err != nil {
        fmt.Printf("读取文件失败：%v\n", err)
        return
}
fmt.Printf("文件内容：%s\n", content)
```

这是一个经典的错误处理案例，当出现错误时，err不为nil，程序捕捉到这个错误并进行if代码块里的逻辑处理并提前返回，反之，当err的值为nil时，代表函数执行成功，此时执行if代码块之后的数据

可以简写成以下形式

```go
if content, err := os.ReadFile("test.txt"); err != nil {
        fmt.Printf("读取文件失败：%v\n", err)
        return
}
fmt.Printf("文件内容：%s\n", content)
```

# 空接口，类型断言

go语言提供了一个可以储存任意的类型的类型——空接口 来看以下代码

```go
map[string]interface{} {
    "name": "bob",
    "age": 11,
    "score" 60.1,
}
```

任何数据类型都可以丢给空接口，但是空接口不可以直接转化为其他数据类型，也不可以直接作为其他数据类型使用 若要转化需要用到类型断言

**类型断言，用于判断一个不确定类型的成员是否为某一特定类型，多用于空接口，泛型**

```go
value, ok := 接口变量.(目标类型)
//假设x是一个interface{}类型，储存了int类型的数据，其值为10


value,ok:=x.(int)
//会将value声明为int类型数据，并为其赋值10
//ok会被声明为bool类型，此时的值会是true


value,ok:=x.(string)
//会将value声明为字符串类型，但是由于出现错误，所以其值会是默认值空字符串
//此时ok的值会是false

//如果写成这样
value:=x.(string)
//程序在类型断言失败时就会直接抛出一个恐慌(panic)，
//如果没有从恐慌中平静下来的逻辑，就会直接中断程序的运行
```

多个类型判断可以使用switch type

```go
switch t := v.(type) {
case int:
    //空接口为int类型时的处理
case string:
    //空接口为string类型时的处理
case float64:
    //空接口为float64类型时的处理
default:
    //空接口为其他类型时的处理
}
```

**当某一函数可以传入的一个参数为空接口时，并不代表可以把所有类型的值丢给它，在这种函数内部往往会进行不同的类型断言，如果你传入了函数不支持的类型，函数会抛出一个错误。**

# JSON和结构体互相转化

JSON是什么？ JSON（JavaScript Object Notation，JavaScript 对象表示法）是一种轻量级的文本数据交换格式，核心作用是在不同系统（如前后端、不同服务）间高效传递结构化数据，因其易读、易解析、跨语言的特性，成为当前主流的数据交换标准。

因为JSON是键值对，所以多用map和结构体与其互相转化 首先我们导入"encoding/json"包

* 将结构体、map等转化为JSON数据(转化为01010101这样的字节数据以便传输) **json.Marshal(v interface{}) (\[\]byte, error)**  将输入的数据转为紧凑格式的 JSON 字节切片；输出无缩进，体积小，适合网络传输；返回字节切片以及可能的错误
* 将JSON数据(需要传入字节切片)转化为结构体、map等 **json.Unmarshal(data \[\]byte, v interface{}) error**  需传入变量指针（否则无法修改数据）；不匹配的 JSON 键会被忽略；返回可能出现的错误

# 用标签标记结构体数据

```go
type User struct {
    Name string 
    UserAge  int    
    Sex  string
}
user:=User{
    Name:       "bob",
    UserAge:    18, 
    Sex:        "female",
}
```

生成json数据，看看是什么样的

用以下简单代码生成JSON字符串

```go
jsonData, _ := json.Marshal(user)
// 这里为了演示方便忽略了错误处理
fmt.Println(string(jsonData))
```

*因为json数据是字节切片，所以要用string()转化为可供人类阅读的字符串*

```json
{"Name":"bob","UserAge":18,"Sex":"female"}
```

以上是一个结构体，如上文所讲，结构体中的字段需要首字母大写才能被被结构体之外的成员访问，也就是说，只有大写的字段才能被转化为JSON，但是JSON数据里的键一般都是小写字母，而且json一般用下划线分割单词，而go语言一般采用将单词首字母大写的方式来命名（即驼峰命名法） 如何处理这样的冲突？ 这就引申出了一个小特性，标签。 json标签可以告诉程序，将结构体的这个字段翻译成相应的json数据。 标签用\`\`引用，多个标签之间用空格隔开

```go
type User struct {
    Name     string   `json:"name"`     // 结构体标签：指定JSON的键为name
    UserAge  int      `json:"user_age"` //小写+下划线
    Sex      string   `json:"-"`        // 标签"-"：忽略该字段，不转为JSON
}
user:=User{
    Name:       "bob",
    UserAge:    18, 
    Sex:        "female",
}
```

```go
jsonData, _ := json.Marshal(user)
fmt.Println(string(jsonData))
```

运行以上代码，观察生成的json数据

```json
{"name":"bob","user_age":18}
```

如何将json数据中的嵌套数据解析成结构体？或者反过来想，如何设计结构体才能使其生成嵌套的json数据？

# 作业

## lv1

我们知道fmt.Print函数可以接收任何值并输出 使用空接口写一个可以接收int，float64，string类型的函数，并使用fmt.Printf输出，要求进行类型断言，并在不符合上述类型时返回错误。 禁止使用%v占位符

函数签名：myPrint(interface{})error{} 例：

```go
myPrint("你好")  // 输出你好并换行
myPrint(11224) // 输出11224并换行
myPrint(59.5731) // 输出
```

## lv2

将以下map转化为json数据输出，要求人类可读，并处理可能出现的错误(简单输出发生的错误)

```go
map[interface{}]interface{} {
    1424: "你好",
    "你好": 65.8,
    986.468: 98755,
}
```

## lv3

将以下结构体转化为json数据输出，要求json数据全部小写，不同单词之间用下划线(_)隔开，要求人类可读，并处理可能出现的错误(简单输出发生的错误)

```go
type Student struct{
    StudentID       int64
    Name            string
    NickName        string
    Score           float64
}
```

## lv4

输入一个数，判断是否为质数(不能被除1和它本身以外所有数除尽的数) 函数签名：isPrimeNumber(num int64)bool{}

```go
isPrimeNumber(14) // 返回false，14可以被2和7整除，不是质数
isPrimeNumber(17) // 返回true，17不可以被除了1和17以外的数整除
```

*提示：一个数永远不可能被大于它的数整除*

## lv5

请根据以下json数据设计结构体，并将其存入设计的结构体里，打印出来

```json
{
    "class":"class two",
    "student":
        [
            {"name":"bob","age":14,"score":[58.9,65.4,95.4]},    
            {"name":"alice","age":15,"score":[54.6,97.5,29.1]}
        ]
}
```