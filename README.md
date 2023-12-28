# Keen

🔧 工具类库

## 目录

- [fp](./fp/pkg.go)：函数式编程算子
- [multitask](./multitask/pkg.go)：多任务执行工具
- [yhttp](./yhttp/pkg.go)：http 相关工具

## 说明

### fp

### multitask

#### `pubsub`

用法：创建一个`PubChan`实例之后，在多个 worker 中使用其指针来进行`Sub`和`UnSub`操作，在控制 goroutine 中使用指针做`Pub`操作

#### `failtask`

用法：创建一个`FailTasks`实例，将所有的任务（func+args）加入到任务列表，调用该实例的`Run`方法，该方法为非阻塞调用，返回一个 channel，当这些任务在以下运行情况中返回不同的内容

- 全部运行成功，运行结果为成功
- 有任务运行失败，要求回退，运行结果为回退状态（成功/错误），**所有任务执行回退操作**
- 有任务运行失败，不要求回退，运行结果为执行失败

### 任务管理器

#### 功能

任务管理器可以添加任务（单个、批量），每个任务需要一个标识ID，运行时唯一（扩展文件存储），并行/并发或者串行执行这些任务，最大并行度，执行模式包括任意失败全部失败以及失败继续运行，如果有任务内部有同步器，需要自行保证任务失败情况下的释放锁资源，任务的输入输出通过管道（`channel`）来传递，包括任务的输入、输出、进度汇报、错误、报表内容（trace信息，执行日志），所有的`channel`的创建和销毁由任务管理器负责

需要记录任务提交者（id），按照执行顺序返回任务执行结果，添加和执行任务同时进行，goroutine管理

任务 Task

类型：

1. 即刻运行
2. 等待执行
3. 触发执行
   1. 任务触发
   2. 条件触发
4. 循环执行（周期）

任务之间的依赖关系：

一对一依赖
多对一依赖
一对多依赖
多对多依赖

前者的执行依靠后者的执行状态

批量任务 BatchTask

特殊的Task，它有两种类型

1. AllAsOne 任意失败则全部失败
2. All  只是全部执行

任务管理器 TaskManager

Register （生成ID，记录ID，后续提交任务验证）
CommitTask 
CommitBatchTask
Run 运行时

#### 抽象

##### `Task`

```go
type Task interface {
  Do(ch1 chan any, ch2 chan any, errCh chan error, progress chan int, trace chan string)
}
```