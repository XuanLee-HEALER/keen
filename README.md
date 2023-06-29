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
