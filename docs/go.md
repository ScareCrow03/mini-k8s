# Go

### 定义变量

#### 单个变量

```go
//标准格式
var a int = 8
//自动推导
var a = 8
//简短格式
a := 8
```

简短格式只能用于初始化定义，不能用`:=`来进行赋值，并且不能用来定义全局变量，并且不能加var关键字和数据类型

#### 多个变量

##### 连续定义

```go
//定义+赋值
var a1, a2 int = 10
//先定义，再赋值
var a1, a2 int
a1 = 10
a2 = 20
```

##### 变量组

```go
//定义+赋值
var(
	a1 int = 10
	a2 float = 3.14
)
//先定义再赋值
var(
	a1 int
    a2 float
)
a1 = 10
a2 = 3.14
```

### 定义常量

```go
const c int = 10
//连续定义
const c1, c2 int = 10, 20
//常量组
const (
	c1 = 10
	c2 = 20
)
//枚举类型,不直接提供，但加入iota可以实现自增
const (
	male = iota
    female = iota
    yao = iota
)
```

### 类型转换

#### 数值类型转换

```go
var num1 int =10
var num2 int8 = 20
var num3 int16
num3 = int16(num1)
num3 = int8(num2)
```

#### 数值类型和字符串类型转换

##### 标准版

利用Formatxxx函数来从数值类型转化为字符串类型

```go
//FormatInt
//第一个参数为需要转化成字符串的数值，只接受int64类型
//第二个参数是转化成多少进制的数
var num1 int = 10
str1 := strconv.FormatInt(int64(num1), 10) //10进制
str2 := strconv.FormatInt(int64(num1), 2) //2进制
fmt.Println(str1) //10
fmt.Println(Str2) // 1010
```

```go
//FormatFloat
//第一个参数为原数值，只接受float64类型
//第二个参数为转化的格式，'f'代表小数格式，'e'代表指数格式
//第三个参数指表六几位小数，-1表示按照指定类型的有效位保留
//第四个参数指定数据的实际位数，float64就是64，float32就是32
var num1 float64 =  3.1415
str1 = strconv.FormatFloat(num1, 'f', 2, 64)
str2 = strconv.FormatFloat(num1, 'f', -1, 64)
str3 = strconv.FormatFloat(num1, 'e', 2, 64)
fmt.Println(str1) //3.14
fmt.Println(str2) //3.1415
fmt.Println(str3) //3.14e+00
```

```go
//FormatBool
//第一个参数是转化的bool参数
var num bool = true
str = strconv.FormatBool(num)
fmt.Println(str) //true;
```

利用ParseXXX()将字符串类型转数值类型

```go
//ParseInt
//第一个参数是要转换的数据
//第二个参数是被转换的数据原来是几进制
//第三个参数是转换后的数据多少位的整形
//返回值为两个，若发生错误的时候，err不为空，需要检查
str1 := "125"
num1, err := strconv.ParseInt(str1, 10, 8)
if err != nil{
	fmt.Println(err)
}
fmt.Println(num1)
```

```go
//ParseFloat
//第一个参数是需要转换的数据
//第二个参数是转换为多少位小数
str1 := 3.1415926
num1, err := strconv.ParseFloat(str1, 64)
if err != nil{
	fmt.Println(err)
}
fmt.Println(num1)
```

```go
//ParseBool
//传入需要转换的字符串
str1 :=  true;
num1, err := strconv.ParseBool(str1, 64)
if err != nil{
	fmt.Println(err)
}
fmt.Println(num1) //true
```

##### 快速版

```go
//Itoa只接受int类型，返回字符串
str1 := strconv.Itoa(int(20))
//Atoi接受字符串，返回整数和错误信息
num1, err := strconv.Atoi("20")
//利用sprintf将数值转换为字符串
num1 := 20
num2 := 3.14
num3 := true
str1 = fmt.Sprintf("%d", num1)
str1 = fmt.Sprintf("%f", num2)
str1 = fmt.Sprintf("%t", num3)
```

### I/O

#### 输出函数

Printf格式化输出，%d是十进制，%o八进制,%x十六进制,%b二进制,%T输出数据类型,%v全能王

Println类似于cout，每个变量之间自动有空格，最后会自动换行

Print全能王，自动生成格式

Fprintf，Fprintln，Fprint会多一个参数，第一个参数指定输出的位置（os.Stdout标准输出,http.ResponseWriter写入到网络响应）

#### 输入函数

Scanf和C类似

```go
var num1 int
var num2 int
fmt.Scanf("%d%d", &num1, &num2)
```

Scanln类似于cin

```
var num1 int
var num2 int
fmt.Scanln(&num1, &num2)
```

同理也有Fscanf，Fscanln，Fscan

额外的还有Sscanf，Sscanln，Sscan，是从字符串中读取数据，这个字符串是函数的第一个参数

### 获取命令行参数

#### 利用flag包

flag包要求参数前有参数的名字

```go
//flag.xxx
//xxx是这个参数的数据类型
//第一个参数是命令行中参数的名称
//第二个参数是默认值
//第三个参数是对参数的说明
name := flag.String("name", "abc", "姓名")
//执行./main.exe -name=ccc时，name会被赋值为“ccc”

//flag.xxxVar
//类似于上面的函数，获取参数值,多一个参数是存储参数的变量指针
var str string
flag.StringVar(&name, "name", "abc", "姓名")
```

#### 利用os包

os包不允许有名字，并且参数顺序不能乱

```go
name := os.Args[1]
```

### 控制流

#### 选择结构

```go
if 初始化语句;条件表达式{
	//do somthing
}else if 条件表达式{
    //do something
}else{
    //do something
}
```

```go
switch 初始化语句;表达式{
	case 表达式1，表达式2:
		//do something
    case 表达式3:
    	fallthrough //将会继续到下一个case
	default:
		//do something
}
//switch中会直接break，如果想要贯穿，需要加fallthrough语句
//表达式可以是任何语句，而不是只能是数字
//每一个case下的执行语句不用加花括号
```

#### 循环结构

没有while，只有for

```go
for 初始化表达式;循环条件表达式;循环后的操作表达式{
	//do something
}

for 索引，值 := range 被遍历的数据（数组）{
    //do something
}
```

#### 跳转结构

可以利用标签机制，让break和continue在嵌套的for语句中正确跳转

可以对某一个循环贴上标签，然后break和continue中可以指定跳出哪一个循环或者继续哪一个循环

```go
for i := 0; i < 5; i++ {
outer: //outer是后续这个for循环的标签
	for j := 0; j < 3; j++ {
		if i+j == 3 {
			fmt.Print("continue outer loop\n")
			continue outer
		}
		fmt.Printf("i=%d, j=%d\n", i, j)
	}
}
//这里continue就是继续j开头的循环，跟普通的continue功能相同

outer: //outer是后续这个for循环的标签
for i := 0; i < 5; i++ {
	for j := 0; j < 3; j++ {
		if i+j == 3 {
			fmt.Print("continue outer loop\n")
			continue outer
		}
		fmt.Printf("i=%d, j=%d\n", i, j)
	}
}
//这里的continue会继续外层（i开头）的循环，而不是内部循环
```

#### 函数

```go
func 函数名(函数列表) (返回值列表(可以只写类型不写名字，也可以写名字，把这些名字在函数体中赋值直接return也可以))
//可以返回多个值
```

值类型（int，float，bool，string，数组，结构体）是拷贝传递，因为是在栈上分配的空间；在函数体内修改的值不会影响到函数之外的值

引用类型(指针，切片，映射，channel)是引用传递，因为是在堆中分配的空间；在函数体内修改的值会影响到函数外的值

#### 匿名函数

将一个函数定义为函数变量，可以作为返回值，函数参数等

```go
a := func(a, b int){
	fmt.Println(a+b)
}
a(1, 2)
```

#### 延迟调用

加入defer关键词，会在其所在函数执行完后再执行，如果有多个defer，采用后进先出的顺序执行

```go
defer fmt.Println("1")
fmt.Println("2")
defer fmt.Println("3")
//输出为 231
```

### 数组

#### 一维数组

```go
//定义方式
var 数组名 [元素个数]数据类型
var arr [3]int
var arr [3]int = [3]int{1, 2, 3}
var arr [3]int = [3]int{0:1, 2:3} //将第一个元素设为1，第三个元素设为3
arr := [...]int{1,2,3} //自动判断长度

//遍历方式
//传统方法
for i:=0; i<len(arr); i++{
    //do something
}
//range遍历
for i, v := range arr{
    //do something
}
```

#### 多维数组

```go
arr := [2][3]int{
	{1, 2, 3},
	{4, 5, 6}
}
```

### 切片

就是可变长的数组，类似于vector

#### 获取方式

```go
//方法一:从数组中切取,这里的切片指向数组的同一份数据
arr := [5]int{1,3,5,7,9}
slice1 := arr[0:2] //这就是一个切片，值为[1,3]
slice1 = arr[:2] //[1,3]
slice1 = arr[0:] //[1,3,5,7,9]
slice1 = arr[:] //[1,3,5,7,9]

//方法二:用make函数创建
//第一个参数指定切片数据类型
//第二个参数指定切片长度
//第三个参数指定切片容量
sce := make([]int, 3, 5)

//方法三:语法糖，创建数组时不写入长度
sce := []int{1,3,5}
```

#### 操作方式

```go
//append加入新元素
var sce = []int{1,3,5}
sce = append(sce, 2)

//copy复制切片（是拷贝数据，而不是指向同一份数据）
var sce = []int{1,3,5}
var sce2 = make([]int, 5)
copy(sce, sce2) //将sce的数据拷贝到sce2上，类似于vector的assign
```

### map(映射)

对应c++的map

```go
var 变量名 map[key类型]value类型
```

#### 创建方式

```go
//利用语法糖快速创建
dict := map[string]string{"name":"lrh", "age":20, "gender":"male"}

//利用make创建(第二个参数是容量，可有可无)
dict := make(map[string]string,3)
```

#### 操作方式

```go
//增和改直接操作即可
dict[2] = 10 //若没有2的key则创建，否则修改

//删除使用delete,第一个参数是map变量，第二个参数是需要删除的key
delete(dict,"name")

//遍历，使用range
var dict := map[string]string{"name":"lrh", "age":20, "gender":"male"}
for k, v := range dict{
    fmt.Println(k, v)
}
```

### 结构体

#### 定义类型

```go
type 类型名称 struct{
	属性名称 属性类型
	属性名称 属性类型
}

//后续以如下的结构进行介绍
type student struct {
    name string
    age int
}
```

#### 创建结构

```go
//完全初始化
var stu = student{"abc", 12}

//部分初始化，需要指定初始化的属性
var stu = student{name:"abc"}

//匿名结构体
var stu = struct {
    name string
    age int
}{
    name: "abc"
    age: 12
}
```

#### 操作结构

```go
stu := student("abc", 12)
//获得值
fmt.Println(stu.name)
//修改值
stu.name = "bcd"
```

#### 类型转换

只有当属性名，属性类型，属性个数，排列顺序都相同的结构体类型才能转换

### 指针

#### 数组指针

```go
var arr = [3]int{1,3,5}
fmt.Println(&arr) //&arr表示指针，arr不表示指针
```

#### 切片指针和映射指针

```go
//切片指针
sce :=  make([]int, 3)
var p *[]int //指向切片的指针
p = &sce
fmt.Println((*p)[1])

//映射指针
dict := make(map[int][int])
var p *map[int][int] //指向map的指针
p = &dict
```

#### 结构指针

```go
//间接获取指针
stu := student("abc", 13)
p := &stu
fmt.Println((*p)["name"])

//通过new获取指针
p := new(studnet)
```

### 方法

就是c++类中的函数，对应为go中结构的函数

```go
func (结构变量名 结构类型)方法名称(形参列表)(返回值列表){
	方法体
}

//例子
func (p person)say(){
    fmt.Println(p.name, p.age)
}

//形参不会被保存
func (p person)setAge1(age int){
    p.age = age //结束后，传入的person实例的age不会被修改
}

//指针可以被保存
func (p *person)setAge2(age int){
    p.age = age //传入的person会被修改
}

per := person("abc", 12)
per.setAge1(13) //不会被修改
per.setAge2(14) //可以被修改，程序内部会把per转换为地址
```

### 接口

定义一个函数，不同的类有不同的实现方式，即一个抽象类

```go
type Animal interface{
	eat()
}

type Dog struct{
	name string
	age int
}

func (d Dog)eat(){
	fmt.Println("dog eat")
}

type Dog struct{
	name string
	age int
}

func (c Cat)eat(){
	ftm.Println("cat eat")
}

var a Animal
a = Dog{"d", 10}
a.eat()
a = Cat("c", 20)
a.eat()
//把抽象的Animal类转化为具体的类
if cat, ok := a.(Cat); ok{
    fmt.Println(cat.age)
}
```

### 面向对象

#### 继承

```go
//单继承
type Person struct{
	name string
	age int
}

type Student struct{
	Person //继承了Person类
	score int //子类特有的属性
}

stu := Student(Person{"abc", 18}, 99)
//下述两种方法都可以获取name属性，如果有重名，那么就会使用就近原则
fmt.Println(stu.name)
fmt.Println(stu.Person.name)

//多继承同理
type Student struct{
    Person
    Teacher
    //上面继承了两个类
}
```

方法也可以继承，也可以重写

#### 多态

使用接口(interface)实现，同一个函数有不同的实现方式

### 字符串操作

英文字符事实上使用[]byte来存储，含有中文的字符串用[]rune来存储

#### 查找子串出现位置

```go
IndexByte(string, sep) int 返回单个字符出现的第一个位置
IndexRune(string, sep) int 返回单个字符或汉字出现的第一个位置
Index(string, substring) int 返回某个子串(可以是字符也可以是汉字)出现的第一个位置
IndexAny(String ,sepset) int 返回sepset中第一个被查找到的sep位置
LastIndex(string, substring) int 返回最后一次出现的位置
IndexFunc(string, func) int 通过func来查找，如果为true则停止，否则继续
func(r rune) bool //func的形式
```

#### 判断是否包含

```go
Contains(string, substring) bool 子串是否存在
ContainsRune(string, sep) bool 某个字符是否存在
ContainsAny(string, sepset) bool sepset中是否有字符在string中
HasPrefix(string, prefix) bool string是否以某字符串开头
HasSuffix(String, suffix) bool string是否以某字符串结尾
```

#### 比较

```go
Compare(a,b) int 比较ASCII值
EqualFold(a,b) 忽略大小写
```

#### 转换

```
ToUpper
ToLower
```

#### 拆合

```go
Split(string, sep) []string //按照sep分割字符串
SplitN(string, sep ,n) []string //从左到右按sep分割，最多分为n个string
SplitAfter(string, sep) []string //分割字符串，并保留sep到每个字符串的末尾
SplitAfterN(string, sep, n) []string //最多分n个，保留sep
Fields(string) []string //将string按照空格分隔
FieldsFunc(string, func) []string //将string按照func的方式分割
join([]string, sep) string //将多个string用sep连接
repeat(string ,int) string //重复string多次
replace(string, old, new, num) string//把string中的前num个old换为new(num=-1表示全部更换)
```

#### 删除

``` go
Trim(string, sep) //将string左右两端的所有连续的sep(字符串)去掉
TrimFunc(String, func) //按照func把string两端删除
TrimLeft
TrimRight
TrimLeftFunc
TrimRightFunc
TrimSpace(string) //删除两端空格
TrimPrefix(String, prefix) //删除前缀
TrimSuffix(string, suffix) //删除后缀
```

### 文件(待改进)

```go
//打开和关闭
fp, err := os.Open("文件名")
if(err != nil){
	fmt.Println(err)
}else{
	fmt.Println(fp)
}
defer func(){
	err = fp.close()
	if err != nil {
		fmt.Println(err)
	}
}()

//读入
buf := make([]byte, 10)
for{
	count, err := fp.Read(buf)
	fmt.Print(string(buf[:count]))
	if err == io.EOF{
		break
	}
}

//写入
bytes := []byte{"a","b","c","\n"}
fp.Write(bytes)

fp.WriteString("abc\n")

//判断文件是否存在
info, err := os.Stat("文件名")
if err == nil{
	//有
}else if err == os.isNotExist(err){
	//无
}
```

### 并发

#### 多线程

go中的线程事实上是一种协程，比线程更轻量级，每个线程执行不同的任务

```go
func t1(){
	fmt.Println("hello")
}

func t2(){
	fmt.Println("world")
}

func main(){
	go t1() 
	go t2()
	//上面创建了两个协程，分别完成不同的任务
	for{
		;
	}
}
```

#### 管道

```go
//管道类似于消息队列，可以存入数据，其他消费者可以拿走其中的数据
管道名 := make(chan 类型, 容量) //创建一个有一个最大容量的管道，如果不定义容量，则是阻塞(无缓冲)的，当发送信息到管道后，需要等待该信息被接收;如果定义了容量，那么在容量没满之前就不会堵塞
ch := make(chan int, 5)
ch <- 5 //向管道中放入数据
num := <-ch //获取数据，将ch中的一个数据给拿出来
利用管道，可以实现协程之间通信
```

#### 特殊的管道——计时器

```
timer := time.After(time.second*2) //建立了一个timer管道，2s后会向timer中发送信号
ticker := time.NewTicker(time.second*2) //ticker每过一定时间就发一个
for{
	t := <-ticker.c //ticker.c表示管道，跟前面的timer不同
}
```

#### select语句

select是一种通信开关，可以选择某一个协程状态进行操作

```go
select{
case u:=<-ch1:
	do somthing
case v:=<-ch2:
	do something
}
```

如果多个管道可以处理，则随机选择一个即可

如果都堵塞，那么就会等待一个完成

常见的方式是在外面套一层死循环，内部标识完成的管道发送信号后直接结束
