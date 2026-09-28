# 第二课：go语言的其他语法结构

## 指针和数组

我们首先来看一个更熟悉的概念，数组，下面是 c 语言的数组示例
```c
int a[3] = {1,2,3};
void f(int a[]) { a[0] = 100; }
```
在c 中我们常常需要数组名退化成指针的问题，而在 go 中赋值会直接传递整个数组
```go
func f(a [3]int) {
    a[0] = 100
}

func main() {
    x := [3]int{1, 2, 3}
    f(x)
    fmt.Println(x)
}
```

再来看一下指针，在 c 语言中，指针是一个非常灵活的工具，可以本身针对内存上的位置进行特定步长上的运算，这让 c 语言中的指针难以理解和掌握，而在 go 语言中，指针并非重头戏，直接依靠指针来操控内存地址的形式绝大多数在 go 中都是不允许的，在 go 语言中只有一些基本的指针使用形式。
```go
var p *int   // p 是 *int，零值是 nil
var q *int = &a
r := &a      // 类型推断为 *int
```

```go
type Student struct {
    Name string
    Age  int
}

s := Student{Name: "Tom", Age: 18}
p := &s

p.Age = 20        // 自动解引用，等价于 (*p).Age = 20
fmt.Println(s.Age) // 20
```

``` go
func inc(p *int) {
    *p = *p + 1
}

func main() {
    x := 10
    inc(&x)
    fmt.Println(x) // 11
}
```
从上面可以看到，go 中指针的使用只有很简单的形式，在很多时候可以很直观的进行操作，即使是第三个例子中将函数作为指针进行传递的方式也不需要对指针进行深究。总结一下，指针在 go 中最核心的时候就是决定是值传递还是靠一个句柄传递，当结构体过于复杂时，通过这种方式减少大字段复制带来的内存开销，或者用在其他地方的逻辑考量。

## 切片

对指针的限制让 go 中操作数组不像 c 里面一样灵活，在go 中， 我们实际很少使用数组，更多的是通过切片。

切片你可以理解是一个动态的数组，为了保证这种动态性切片除了具体的数据外同时储存长度信息和容量信息，
```c
struct slice {
    int *data;//指向底层数组
    size_t len;
    size_t cap;
};
```
也就是说，切片针对底层的数组进行了额外的一层抽象，本身底层还是依靠之前提到的数组。同时 len 和 cap 的概念可能会很让你疑惑，cap 提供总容量，len 记录目前已经使用的已经填充的部分，除了一些边界情况或者算法外，我们一般不需要对这两者进行精细化的考量

现在我们来讲解一下基本的切片使用方法:先记住一条规则：左闭右开

```go
s[low:high]
```
从索引 low 开始，到索引 high 之前结束,也就是包含 low，不包含 high。更直观的图形理解如下：
```text
arr:    [ 10 | 20 | 30 | 40 | 50 ]
索引:      0    1    2    3    4

arr[:3]  -> [ 10 | 20 | 30 ]        // 0,1,2
arr[2:]  ->           [ 30 | 40 | 50 ] // 2,3,4
arr[1:4] ->      [ 20 | 30 | 40 ]     // 1,2,3
```
如果 high或 low 没有，那么默认到达底层数组的切片上（存在元素），如果我们这样画可能会更容易理解，这就回到我们之前的讨论，切片是基于底层数组进行实现的。
```text
arr:    [ 10 | 20 | 30 | 40 | 50 | 0 | 0 | ...]
索引:      0    1    2    3    4

arr[:3]  -> [ 10 | 20 | 30 ]        // 0,1,2
arr[2:]  ->           [ 30 | 40 | 50 ] // 2,3,4
arr[1:4] ->      [ 20 | 30 | 40 ]     // 1,2,3
```

len 和 cap 怎么随着切片的变化而变化呢，对于简单切片 s[low:high]：
```text
len = high - low
cap = cap(s) - low
```
```go
arr := []int{10, 20, 30, 40, 50} // len=5, cap=5

s1 := arr[:3]   // len=3, cap=5
s2 := arr[2:]   // len=3, cap=3
s3 := arr[1:4]  // len=3, cap=4
```
len 是窗口里能看到几个元素。cap 是从窗口起点到底层数组末尾还有多少空间。如果你目前搞不清楚这一点那就先将他们抛掉，他们并不是很重要的内容

```go
arr := []int{1, 2, 3, 4, 5}
sub := arr[1:4] // sub = [2 3 4]

sub[0] = 99
fmt.Println(arr) // [1 99 3 4 5]
```
在这里可以看到，sub 和 arr 底层实际共享一个数组，注意我们在之前用 c 语言模拟 go 底层切片机制的时候使用的是指针，两个切片数据指针和指向同一个内存地址，你可以这样大致理解。

如果你需要独立进行切片复制需要：
```go
sub := make([]int, 3)
copy(sub, arr[1:4])
```
结合上面这些操作我们来讲解一下在 go 中能使用到的关键字，很多针对切片起到作用


### 1.make：

```go
make([]T, len)        // 长度 = 容量 = len
make([]T, len, cap)   // 长度 = len，容量 = cap
```

例子：

```go
s1 := make([]int, 3)      // len=3, cap=3, [0 0 0]
s2 := make([]int, 0, 5)   // len=0, cap=5, []
s3 := make([]int, 2, 5)   // len=2, cap=5, [0 0]
```

要点：

- `make` 只用于 `slice`、`map`、`chan`，不能用来创建普通变量或结构体（后两者我们会在后续课程中提到）；
- `make([]int, 3)` 里的 `3` 是长度，不是容量；
- 创建后元素是**零值**，不是垃圾值。


Go 帮你做了三件事：

1. 分配底层数组；
2. 把元素清零；
3. 返回一个带 `len`、`cap` 的切片头。

### 2. new
```go
p := new(T)
```
这个表达式意味着：分配一块能放下 T 的内存，把这块内存清零，变成 T 的零值；，返回一个指向它的指针，类型是 *T。

```go
p := new(int)
fmt.Println(*p) // 0

*p = 42
fmt.Println(*p) // 42
```
如果你想初始化一个切片也是同样的操作
### 3. `make` vs `new`

| | `new(T)` | `make(T, ...)` |
|---|---|---|
| 返回 | `*T` | `T` 本身 |
| 适用 | 任意类型 | 只能 slice、map、chan |
| 初始化 | 零值 | slice/map/chan 的内部结构已就绪 |
| 例子 | `p := new(int)` | `s := make([]int, 0, 10)` |

一句话：

> `new` 给你一个指向零值的指针；  
> `make` 给你一个可以直接用的切片/map/chan。

额外补充一个点就是为什么要用 new/make 而不是更直观的 var，var只能给一个零值，但 new/make 则能给出完整的初始状态，很多数据结构零值完全可用比如 int 等，但是在我们本节以及之后介绍的诸如map channel 指针他们则完全不可用，切片可以执行 append（后续介绍），但在一些情况下也难以使用

```go
var s []int
s[0] = 1 // panic: index out of range
```
因此，在复杂数据结构创建的时候，我们一般使用 new/make 的方式（在我们结束完讲解之后你可以想一想为什么很多数据结构零值不可用）

### 4. 为什么要写容量

```go
// 不预分配：多次扩容
var s []int
for i := 0; i < 1000; i++ {
    s = append(s, i)
}

// 预分配：一次到位
s := make([]int, 0, 1000)
for i := 0; i < 1000; i++ {
    s = append(s, i)
}
```

预分配的意义：

- 避免多次 `growslice`；
- 减少内存拷贝；
- 减少 GC 压力。

我们之前提到切片是动态扩容的，但是这个扩容机制是有 runtime 或者 go 底层去帮你做的，到达上限后（len>cap）：分配新数组，在堆上申请内存，进行旧数据拷贝，返回新的切片头。这本身会带来很大的开销，如果可以的话尽可能的预分配。当然，在学习阶段可以先不纠结这个

---

### copy：


```go
n := copy(dst, src)
```

规则：

- 按**较短的**复制，返回实际复制的元素个数；dst src那个短按短的来进行元素复制个数的取决
- `dst` 和 `src` 可以长度不同；
- 只复制元素，不复制底层数组本身。

例子：

```go
src := []int{1, 2, 3, 4, 5}
dst := make([]int, 3)

n := copy(dst, src)
fmt.Println(n)   // 3
fmt.Println(dst) // [1 2 3]
```

反过来：

```go
src := []int{1, 2, 3}
dst := make([]int, 5)

n := copy(dst, src)
fmt.Println(n)   // 3
fmt.Println(dst) // [1 2 3 0 0]
```



### . 复制切片的常见写法

```go
// 写法 1：make + copy
clone := make([]int, len(src))
copy(clone, src)

// 写法 2：append 到 nil
clone := append([]int(nil), src...)
```

两种都可以。第一种更明确，第二种更短。

### . 为什么切片不能直接用 `=`

```go
a := []int{1, 2, 3}
b := a       // b 和 a 共享底层数组
b[0] = 100
fmt.Println(a) // [100 2 3]
```

因为 `=` 只复制了切片头，底层数组是共享的。也就是说，这里只复制了上层的抽象，并没有在底层储存单元上实现真正的副本复制。

要独立副本：

```go
b := make([]int, len(a))
copy(b, a)
```

或者：

```go
b := append([]int(nil), a...)
```

记住：

> 切片赋值不是复制数据，是复制窗口。  
> 想复制数据，用 `copy`。


### append


```go
s = append(s, x)        // 追加一个
s = append(s, x, y, z)  // 追加多个
s = append(s, t...)     // 追加另一个切片的所有元素（...在 go 中常用语切片展开，或者表示不定参数）
```

例子：

```go
s := []int{1, 2, 3}
s = append(s, 4)          // [1 2 3 4]
s = append(s, 5, 6)       // [1 2 3 4 5 6]
s = append(s, []int{7, 8}...) // [1 2 3 4 5 6 7 8]
```


错误写法：

```go
s := []int{1, 2, 3}
append(s, 4)
fmt.Println(s) // [1 2 3]，4 丢了
```

正确写法：

```go
s = append(s, 4)
```

原因：

> `append` 可能扩容，扩容后会返回一个全新的切片头。  
> 不接收返回值，新的切片头就丢了。

###  容量足够时：原地写入

```go
s := make([]int, 2, 5) // len=2, cap=5
s[0], s[1] = 1, 2

t := append(s, 3)
fmt.Println(s) // [1 2]，s 的 len 还是 2
fmt.Println(t) // [1 2 3]
```

底层发生了什么：

```text
底层数组: [ 1 | 2 | 3 | _ | _ ]
索引:       0   1   2   3   4

s: len=2, cap=5, 指向索引 0
t: len=3, cap=5, 指向索引 0
```

`append` 直接在索引 2 写入 3，然后把 `t` 的 `len` 改成 3。`s` 的 `len` 还是 2，所以看不到那个 3。上层的切片头遮蔽了底层的数组单元。

但注意：

```go
u := append(s, 9) // 从 s 再 append 一次
fmt.Println(t) // [1 2 9]，t 被覆盖了
```

因为 `t` 和 `u` 共享底层数组，`u` 写入索引 2 时覆盖了 `t` 的位置，这就是经典的append 共享底层数组所产生的问题。在实际编码的时候其实挺难遇到这种情况的，只需要调用 append 同时更新切片头就能避免这类问题。


### range
range 可以用来数组，切片，map 中的元素（在 map 中的遍历稍有不同我们后面再来讲）
```go
for i, v := range s {
    // i 是索引，v 是元素副本
}
```

例子：

```go
s := []int{10, 20, 30}
for i, v := range s {
    fmt.Println(i, v)
}
// 0 10
// 1 20
// 2 30
```
只要索引情况下
```go
for i := range s {
    fmt.Println(i)
}
```

只要元素情况下

```go
for _, v := range s {
    fmt.Println(v)
}
```

注意：`_` 是空白标识符，表示“我不关心这个值”。`v` 是副本，是复制过来的结果，不是指针引用

```go
s := []int{1, 2, 3}
for _, v := range s {
    v = 0 // 改的是副本，不影响 s
}
fmt.Println(s) // [1 2 3]
```

要修改元素，必须用索引：

```go
for i := range s {
    s[i] = 0
}
fmt.Println(s) // [0 0 0]
```

记住：

> `range` 给的 `v` 是元素的一份拷贝。  
> 想改原切片，用 `s[i]`。


`range` 的好处：

- 不用手动管边界；
- 不会越界；
- 代码更短。

但要注意：`range` 表达式在循环开始前只求值一次。

```go
s := []int{1, 2, 3}
for i := range s {
    s = append(s, i) // 不会死循环，range 只遍历最初的 3 个
}
```

---

### 把四个工具串起来，让他们作用于我们的切片形式

```go
func main() {
    // make：创建
    s := make([]int, 0, 4) // len=0, cap=4

    // append：追加
    s = append(s, 1, 2, 3)
    fmt.Println(s, len(s), cap(s)) // [1 2 3] 3 4

    // range：遍历
    for i, v := range s {
        fmt.Println(i, v)
    }

    // copy：复制
    t := make([]int, len(s))
    copy(t, s)
    t[0] = 100
    fmt.Println(s) // [1 2 3]，不受影响
    fmt.Println(t) // [100 2 3]
}
```

## map

切片解决的是“有序、按下标访问”的数据组织。但现实中很多数据不是按索引排列的，而是以两个数据间的对应排列组织的。最典型的就是 DNS：你输入域名 `www.baidu.com`，DNS 帮你查到对应的 IP 地址。这种“键 -> 值”的对应关系，在 Go 里就用 map 来表示。

```go
dns := map[string]string{
    "www.baidu.com": "110.242.68.66",
    "www.google.com": "142.250.72.196",
    "localhost": "127.0.0.1",
}

fmt.Println(dns["www.baidu.com"]) // 110.242.68.66
```

这里 `string` 是键类型，`string` 是值类型。键必须是可比较的类型，值可以是任意类型。你可以把 map 理解成一本字典：给一个键，不需要像切片那样从头遍历。或者从 map 这个英文词的角度来理解：值和键之间形成了映射关系。我们再来举一个实际的代码场景：学号和姓名，姓名和投票的对应统计关系来进行展示，看到 map 数据结构的优势所在。
```go
// 不用 map：结构体切片，手动查找
type Result struct{ Name string; Votes int }
results := []Result{}
for _, name := range votes {
    found := false
    for i := range results {
        if results[i].Name == name {
            results[i].Votes++
            found = true
            break
        }
    }
    if !found {
        results = append(results, Result{name, 1})
    }
}
```

```go
// 用 map：一行搞定
count := map[string]int{}
for _, name := range votes {
    count[name]++
}
```
### 创建 map

```go
var m1 map[string]string        // nil map，不能写
m2 := make(map[string]string)   // 空 map，可写
m3 := map[string]string{        // 字面量
    "a": "1",
    "b": "2",
}
m4 := make(map[string]string, 10) // 预分配容量，提示
```

注意：`var m map[string]string` 得到的是 nil map。这也是我们之前强调尽量使用 new/make 的原因

```go
var m map[string]string
fmt.Println(m["x"]) // ""，读不存在的键返回零值
m["x"] = "1"        // panic: assignment to entry in nil map
```

因此想要真正使用 map 必须使用 new/make 初始化


### 增删改查

```go
m := make(map[string]string)

// 增 / 改
m["www.baidu.com"] = "110.242.68.66"
m["www.baidu.com"] = "110.242.68.67" // 覆盖旧值

// 查
ip := m["www.baidu.com"]        // 不存在返回 ""
ip, ok := m["www.baidu.com"]    // ok 表示键是否存在
if ok {
    fmt.Println("找到:", ip)
} else {
    fmt.Println("没有这个域名")
}

// 删
delete(m, "www.baidu.com")

// 长度
fmt.Println(len(m))
```

重点记住这个惯用法：

```go
v, ok := m[k]
```

因为 `v := m[k]` 在键不存在时返回零值，你无法区分“值本来就是零值”和“键不存在”。 用 `ok` 才能准确判断（bool 值类型）。和range类似，不过 map 的额外附加字段是判断是否存在，其中map 的增删查改更多的是语法上的适应，整体逻辑比较简单，关键是搞清楚键值对的映射关系，如何改变和读取这样的映射关系

### range 在 map 中的不同

WQ23sfrtdgcdv fgb0-p=[]
"?˘0-=]\
 
=-0=786tyfg=9AQSWDFERG T
map 中，`range` 返回的是 `键, 值`：

```go
dns := map[string]string{
    "www.baidu.com": "110.242.68.66",
    "localhost":     "127.0.0.1",
}

for domain, ip := range dns {
    fmt.Println(domain, "->", ip)
}
```

和切片最大的不同：

1. **顺序不保证**。map 遍历顺序是随机的，每次运行可能不一样，不要依赖顺序。这一点在很多地方都需要进行额外处理。
2. 只要键：`for domain := range dns`
3. 只要值：`for _, ip := range dns`
4. `ip` 是值的副本，改它不会影响 map。要改 map 用 `dns[domain] = newIP`

```go
for domain := range dns {
    fmt.Println(domain)
}

for _, ip := range dns {
    fmt.Println(ip)
}
```
map 的 range 是“键值对”，不是“索引元素”，而且无序。

### map 是句柄传递

和切片类似，map 变量本身只是一个句柄。  
把 map 传给函数，函数内修改 map 的内容，外部可见：

```go
func addRecord(m map[string]string) {
    m["www.baidu.com"] = "110.242.68.66"
}

func main() {
    dns := make(map[string]string)
    addRecord(dns)
    fmt.Println(dns["www.baidu.com"]) // 110.242.68.66
}
```

但如果你在函数内让 `m = make(...)`，那只是改了局部副本，外部不受影响。

---

### map 的 key 要求

键必须是可比较的类型。常用的有：

- string
- int、float
- bool
- 指针
- 结构体（字段都可比较）
- 数组

不能做键的：

- slice
- map
- func

因为 map 底层是哈希表，需要能计算哈希和判断相等。

---

### map 不能直接比较

```go
m1 := map[string]string{"a": "1"}
m2 := map[string]string{"a": "1"}

// fmt.Println(m1 == m2) // 编译错误
fmt.Println(m1 == nil)    // 只能和 nil 比较
```

map 只能和 `nil` 比较，不能两个 map 直接 `==`。

---

### 和切片对比

| | 切片 | map |
|---|---|---|
| 组织方式 | 有序，按索引 | 无序，按键 |
| 访问 | `s[i]` | `m[k]` |
| 查找 | 需要遍历 | 平均 O(1) |
| 零值 | nil 切片，append 可用 | nil map，写 panic |
| 复制变量 | 共享底层数组 | 共享底层哈希表 |
| 遍历 | `range` 返回索引、元素 | `range` 返回键、值，顺序随机 |

---

### 小练习

用 map 实现一个简单的 DNS 缓存：

```go
func main() {
    dns := make(map[string]string)

    dns["www.baidu.com"] = "110.242.68.66"
    dns["localhost"] = "127.0.0.1"

    if ip, ok := dns["www.baidu.com"]; ok {
        fmt.Println("百度 IP:", ip)
    }

    delete(dns, "localhost")

    for domain, ip := range dns {
        fmt.Println(domain, "->", ip)
    }
}
```

## 对象

### 一、从宏观角度理解对象

在 C 里，程序的组织方式是：**数据是数据，函数是函数，两者分开。** 结构体是组织数据表示数据的主要方式，而函数用来处理数据，两者完全分开，编写 c 程序只需要关注怎么处理得到我们想要的数据，怎么设计组织相关函数。

```c
struct Student { char *name; int age; };

void birthday(struct Student *s) { s->age++; }
void print(struct Student *s) { printf("%s %d\n", s->name, s->age); }
```

你知道 `birthday` 是给 `Student` 用的，但这是靠**命名约定**和**规范自觉**。语言本身没有把 `birthday` 和 `Student` 绑在一起。任何函数都可以传一个 `Student*` 进去，不管它合不合理。

这种组织方式的问题是：

- 数据和行为分离，代码一多就散；
- 没有“这个操作属于这个类型”的强制约束；
- 类型和操作之间的关系靠人记，不靠语言保证。

于是有了**对象**的思想：

> **把数据和操作数据的行为绑在一起，形成一个整体。**

这个整体就叫对象。

```go
type Student struct {
    Name string
    Age  int
}

func (s *Student) Birthday() {
    s.Age++
}

func (s Student) Print() {
    fmt.Println(s.Name, s.Age)
}
```

现在 `Birthday` 和 `Print` 不是随便哪个函数了，它们明确属于 `Student`。你调 `s.Birthday()`，读起来就是“学生过生日”，而不是“给某个学生调用生日函数”。这就是对象的核心设计思想：

> **数据是“是什么”，方法是“能做什么”，两者绑在一起，就是一个对象。**

---

### 二、对象解决了什么问题

从 C 到 Go，组织代码的方式变了。

**C 的方式：**

```c
struct Student s;
birthday(&s);
print(&s);
```

**Go 的方式：**

```go
s := Student{}
s.Birthday()
s.Print()
```

区别不只是语法。更深层的区别是：

| | C | Go |
|---|---|---|
| 数据 | 结构体 | 结构体 |
| 行为 | 独立函数 | 绑定到类型的方法 |
| 关系 | 靠命名约定 | 语言强制绑定 |
| 调用 | `f(&s)` | `s.f()` |
| 语义 | “对 s 做 f” | “s 执行 f” |

对象让代码从“**对数据做操作**”变成“**对象自己完成行为**”。

这就是面向对象的基本直觉。我们不需要再强迫自己记住那个函数是做什么的，而将目光移到一个对象上，某个对象能干什么，能实现什么。

---

### 三、对象的三层含义

“对象”这个词在不同层面有不同含义，我们把它拆开一步一步分析：

**第一层：数据**

```go
type Student struct {
    Name string
    Age  int
}
```

这是对象的**状态**。它记录“这个对象现在是什么样”。这里就是结构体，可能大家现在没有一个很深的感受，在项目编码中很少使用单独数据结构传参，大量的函数输入和输出都是作为结构体进行的。本身结构体是对零散的数据做了一个聚合一层抽象。

**第二层：方法**

```go
func (s *Student) Birthday() { s.Age++ }
```

这是对象的**行为**。它描述“这个对象能做什么”。这个函数仅仅率属于这个对象，也仅仅只能被这个对象调用。

**第三层：整体**

```go
s := Student{Name: "Tom", Age: 18}
s.Birthday()
```

`Student{...}` 是一个具体的对象实例。它有状态，能执行行为。
> 对象 = 状态 + 行为。  
> 类型定义状态，方法定义行为，对象是两者的结合。


### 四、为什么需要对象

没有对象行不行？当然行，C 就是这么写的。但对象带来了三个好处：

**1. 组织性**

数据和操作它的代码放在一起，不用在几千行里找“哪个函数是处理这个结构体的”。

**2. 封装**

外部不需要知道 `Age` 怎么变，只需要调 `Birthday()`。

```go
s.Birthday() // 你不需要知道内部是 Age++ 还是别的
```

**3. 可扩展**

不同对象可以有相同的行为，这为 interface 打下基础。

```go
type Circle struct{ R float64 }
type Rect struct{ W, H float64 }

func (c Circle) Area() float64 { ... }
func (r Rect) Area() float64   { ... }
```

`Circle` 和 `Rect` 是两个不同的对象，但它们都能 `Area()`。  
这就引出了下一个问题：能不能用同一个类型统一表示它们？  
这就是 interface。

---

### 五、对象和方法的关系

在 Go 里，方法就是绑定到类型上的函数。

```go
// 普通函数
func birthday(s *Student) { s.Age++ }

// 方法
func (s *Student) Birthday() { s.Age++ }
```

两者做的事情一样，区别是：

- 方法有接收者 `(s *Student)`；
- 方法只能通过对象调用：`s.Birthday()`；
- 方法属于类型，函数不属于任何类型。

方法的接收者有两种：

```go
func (s Student) Print()  { ... } // 值接收者：拿到副本
func (s *Student) Inc()   { ... } // 指针接收者：拿到指针，能改原对象
```

这个后面细讲，现在只需要知道：

> 方法是对象的行为，接收者决定方法能不能修改对象。
注意一点，在最开始的时候你可以简单理解方法多了一个参数，你可以去进行读取操作，而不同的接受者决定你能否真正的去修改这个参数，就像我们之前所说的，指针的作用在 go 中主要是保证全局可以写入一个对象，同时减少因为传递值带来的复制开销。

---

### 六、对象和 interface 的关系

现在可以顺理成章地引出 interface 了。

对象有了行为，行为就是方法。  
不同对象可以有相同的方法。  
interface 就是用来描述“**一组方法**”的。

```go
type Shape interface {
    Area() float64
}
```

它的意思不是“Shape 是什么”，而是：

> 谁能 `Area()`，谁就是 Shape。

所以顺序是：

```text
结构体（数据）
  -> 方法（行为）
  -> 对象（数据 + 行为）
  -> 接口（对行为的抽象）
```

没有对象，就没有“行为”的概念。  
没有行为，接口就无从定义。

---

### 七、总结

宏观上，对象是一种组织代码的思想：

> 把数据和操作数据的行为绑在一起。

在 Go 里：

- 结构体定义数据；
- 方法定义行为；
- 接收者把方法绑定到类型上；
- 数据 + 方法 = 对象；
- 对象的行为集合，就是接口的基础。

## interface

切片解决容器问题，是之前在 c 中常有数组的扩展，map 解决键值查找，针对特定类型的数据排布需求。但还有一类问题它们都解决不了：

> 我想写一个函数，它能接收**不同类型**的参数，只要这些类型有某种共同行为。

这就是 interface 要解决的问题。

---

### 为什么需要 interface

先看一个具体场景。假设你要计算不同图形的面积：

```go
type Circle struct{ R float64 }
type Rect struct{ W, H float64 }

func (c Circle) Area() float64 { return 3.14 * c.R * c.R }
func (r Rect) Area() float64   { return r.W * r.H }
```

如果没有 interface，你想写一个“求总面积”的函数，只能每种类型写一个：

```go
func totalCircleArea(cs []Circle) float64 {
    sum := 0.0
    for _, c := range cs {
        sum += c.Area()
    }
    return sum
}

func totalRectArea(rs []Rect) float64 {
    sum := 0.0
    for _, r := range rs {
        sum += r.Area()
    }
    return sum
}
```

代码几乎一样，只是类型不同。每加一种图形，就要再写一遍。

用 interface：

```go
type Shape interface {
    Area() float64
}

func totalArea(shapes []Shape) float64 {
    sum := 0.0
    for _, s := range shapes {
        sum += s.Area()
    }
    return sum
}
```

调用：

```go
shapes := []Shape{
    Circle{R: 1},
    Rect{W: 2, H: 3},
}
fmt.Println(totalArea(shapes)) // 9.14
```

`totalArea` 不关心你传的是圆还是矩形，不关心你函数内部的具体实现，它只关心一件事，你有没有 `Area() float64` 这个方法，这就是 interface 的核心：

> **interface 描述的是“能做什么”，而不是“是什么”。**  
> 只要一个类型实现了接口里的所有方法，它就自动满足这个接口，不需要显式声明。


---

### 接口的定义和实现

```go
type Shape interface {
    Area() float64
}
```

任何类型，只要它有 `Area() float64` 方法，就自动实现了 `Shape`：

```go
type Circle struct{ R float64 }

func (c Circle) Area() float64 {
    return 3.14 * c.R * c.R
}

var s Shape = Circle{R: 1} // 可以，Circle 满足 Shape
```

不需要写 `implements Shape`，不需要继承。这叫**隐式实现**。
值接收者和指针接收者的区别：

```go
type Rect struct{ W, H float64 }

func (r *Rect) Area() float64 { return r.W * r.H }

var s Shape = &Rect{W: 2, H: 3} // 可以
// var s Shape = Rect{W: 2, H: 3} // 不可以，Rect 值类型没有 Area 方法
```
这里的提到的问题是，值和指针虽然共享同一个结构体名字，但是方法归属于域却截然不同，你可以这样理解这种不对称的问题，值的方法只能对结构体进行读取，而指针的方法既对结构体进行读取也能进行写入，这种读取和写入的不对称性是方法域满足的不对称性根源。
> 方法定义在 `*T` 上，只有 `*T` 满足接口。  
> 方法定义在 `T` 上，`T` 和 `*T` 都满足接口。

---

### 底层实现

接口值在底层是一个**两个字段的结构**：

```go
type iface struct {
    tab  *itab          // 类型信息 + 方法表
    data unsafe.Pointer // 指向具体数据
}
```

对于空接口 `interface{}`：

```go
type eface struct {
    _type *_type         // 类型信息
    data  unsafe.Pointer // 指向具体数据
}
```

可以这样理解：

```text
接口值
+-------------------+
| 类型信息 (itab)    |  ->  Circle 的方法表
+-------------------+
| 数据指针 (data)    |  ->  实际的 Circle 值
+-------------------+
```

所以接口值不是“一个值”，而是“**一个类型 + 一个数据指针**”。

这解释了几个现象：

**1. 接口可以存任意类型**

```go
var i interface{} = 42
var j interface{} = "hello"
var k interface{} = Circle{R: 1}
```

因为 `data` 指向实际数据，`_type` 记录它是什么类型。

**2. 接口比较的是类型和值**

```go
var a interface{} = 42
var b interface{} = 42
fmt.Println(a == b) // true，类型和值都相同

var c interface{} = 42
var d interface{} = int64(42)（）
fmt.Println(c == d) // false，类型不同
```

**3. nil 接口和 nil 指针不同**

```go
var p *int
var i interface{} = p
fmt.Println(i == nil) // false
```

因为 `i` 的 `_type` 是 `*int`，`data` 是 nil，但接口本身不是 nil。也许你们可能会对这里有疑惑，为什么这里的 data 是 nil，因为 p 只是一个空指针，没有指向任何地方。而接口是 nil 的要求是 data 是 nil 同时_type也是 nil

---

### 空接口 interface{}

空接口没有任何方法，所以**所有类型都满足它**：

```go
var i interface{}
i = 42
i = "hello"
i = Circle{R: 1}
i = []int{1, 2, 3}
i = map[string]int{"a": 1}
```

`fmt.Println` 的参数就是 `interface{}`：

```go
func Println(a ...interface{}) (n int, err error)
```

所以它能打印任何东西。

但空接口丢失了类型信息，要用**类型断言**取回来：

```go
var i interface{} = "hello"

s, ok := i.(string)
if ok {
    fmt.Println("是字符串:", s)
}

n, ok := i.(int)
if !ok {
    fmt.Println("不是 int")
}
```

也可以用 `switch`：

```go
switch v := i.(type) {
case string:
    fmt.Println("string:", v)
case int:
    fmt.Println("int:", v)
default:
    fmt.Println("其他类型")
}
```

---

### 接口的常见用法

**1. 多态**

```go
type Animal interface {
    Speak() string
}

type Dog struct{}
type Cat struct{}

func (d Dog) Speak() string { return "Woof" }
func (c Cat) Speak() string { return "Meow" }

func main() {
    animals := []Animal{Dog{}, Cat{}}
    for _, a := range animals {
        fmt.Println(a.Speak())
    }
}
```

**2. 标准库中的接口**

```go
type Stringer interface {
    String() string
}

type error interface {
    Error() string
}
```

你实现 `String()` 方法，`fmt.Println` 就会用它：

```go
type Point struct{ X, Y int }

func (p Point) String() string {
    return fmt.Sprintf("(%d, %d)", p.X, p.Y)
}

fmt.Println(Point{1, 2}) // (1, 2)
```

---

### 常见坑

**坑 1：接口里存了 nil 指针，接口本身不是 nil**

```go
var p *int
var i interface{} = p
fmt.Println(i == nil) // false
```

**坑 2：值接收者 vs 指针接收者**

```go
type Counter struct{ N int }

func (c *Counter) Inc() { c.N++ }

type Incer interface{ Inc() }

var i Incer = &Counter{} // 可以
// var i Incer = Counter{}  // 不可以
```

**坑 3：接口比较可能 panic**

```go
var a interface{} = []int{1, 2}
var b interface{} = []int{1, 2}
fmt.Println(a == b) // panic: 切片不能比较
```

接口比较要求底层类型可比较。

---

### 和 C 对比

| | C | Go |
|---|---|---|
| 泛型容器 | `void*` | `interface{}` |
| 多态 | 函数指针 + 手动分发 | 接口 + 隐式实现 |
| 类型安全 | 无 | 有 |
| 类型检查 | 运行时手动 | 类型断言 / switch |
| 方法表 | 手动维护 | itab 自动生成 |

一句话：

> C 用 `void*` 和函数指针实现“泛型”，但类型不安全。  
> Go 用 interface 实现同样的目的，但类型安全、自动绑定方法。

---

### 总结
- interface 描述“能做什么”，不描述“是什么”；
- 隐式实现：有方法就满足，不需要声明；
- 底层是 `(type, data)` 两个字段；
- 空接口 `interface{}` 能存任意类型；
- 用类型断言或 `type switch` 取回具体类型；
- 接口是 Go 实现多态和抽象的核心工具。

## 总结
所以在这个阶段我们再来进行总结，在这节课中主要讲了四个主题，slice，map，对象，interface。前两者在概念上容易理解，因为它们仅仅聚焦数据处理和数据组织，这种思考方式我们在学习 c 的时候已经很习惯了，但是对象和 interface 的概念如果你是初次接触的话可能很难理解，对象以方法和数据聚合的形式重新进行数据组织和排布，而 interface 的接口层，可能需要你们进行实际项目尝试的时候才能体会到其实际意义。在这里我们可以很抽象的说一下。当我们这个项目庞大复杂到一定程度上的时候，很自然的我们会想要去对项目进行概念上乃至实际上的拆分，那么很自然的这样一个问题是肯定要面对的，如何实现层与层的沟通和联系，如何实现单个层改动同时其他层不需要因为这个层的改动而改动，接口就是适配于这种需求，各个层之间不依赖具体的对象具体的函数沟通而是依赖抽象的接口层进行通信，只需要预先针对结构进行抽象和契约就可以实现。
