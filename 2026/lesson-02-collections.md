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

`append` 直接在索引 2 写入 3，然后把 `t` 的 `len` 改成 3。`s` 的 `len` 还是 2，所以看不到那个 3。

但注意：

```go
u := append(s, 9) // 从 s 再 append 一次
fmt.Println(t) // [1 2 9]，t 被覆盖了
```

因为 `t` 和 `u` 共享底层数组，`u` 写入索引 2 时覆盖了 `t` 的位置。

这就是经典的“append 共享底层数组陷阱”。

### 4. 容量不足时：分配新数组

```go
s := []int{1, 2, 3} // len=3, cap=3
t := append(s, 4)   // 触发扩容

t[0] = 100
fmt.Println(s) // [1 2 3]，s 不受影响
fmt.Println(t) // [100 2 3 4]
```

因为 `t` 指向了一个全新的底层数组。

### 5. 扩容策略

Go 的 `growslice` 大致策略：

- 旧容量 < 256：新容量翻倍；
- 旧容量 >= 256：按一定比例增长（大约 1.25 倍，逐步过渡）；
- 最后按元素大小做内存对齐。

不用记具体数字，记住：

> 扩容是“新的底层数组 + 复制旧数据 + 返回新切片头”。  
> 所以 `append` 前后，切片可能已经换了底层数组。

### 6. append 常见模式

**做栈：**

```go
stack := []int{}
stack = append(stack, 1) // push
x := stack[len(stack)-1] // top
stack = stack[:len(stack)-1] // pop
```

**删除第 i 个元素：**

```go
s = append(s[:i], s[i+1:]...)
```

**在位置 i 插入 x：**

```go
s = append(s, 0)       // 先扩容一个位置
copy(s[i+1:], s[i:])   // 把 i 后面的元素往右挪
s[i] = x
```

---

## 四、range：遍历切片

### 1. 基本语法

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

### 2. 只要索引

```go
for i := range s {
    fmt.Println(i)
}
```

### 3. 只要元素

```go
for _, v := range s {
    fmt.Println(v)
}
```

注意：`_` 是空白标识符，表示“我不关心这个值”。

### 4. `v` 是副本，不是引用

这是新生最容易犯的错：

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

### 5. 和 C 对比

C 里：

```c
for (int i = 0; i < n; i++) {
    int v = arr[i];
    // ...
}
```

Go 里：

```go
for i, v := range s {
    // ...
}
```

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

## 五、把四个工具串起来

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
