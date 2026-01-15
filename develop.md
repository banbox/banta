
### 并发操作BarEnv和Series
为避免额外冗余创建BarEnv和Series，多策略和多账户会复用，多策略和多账户需支持并发调用。

原方案使用RWMutex保护Series各方法，但会带来43%的性能损失。现改为在BarEnv中只保留一个通用锁`Lock`，由用户手动调用。

原因：banta的相关函数调用通常是密集且相邻的，所以一个锁就足矣，需要并发时可加锁保护banta代码块。

使用示例：
```go
env.Lock.Lock()
// 进行banta运算
result := banta.EMA(env.Close, 20).Get(0)
env.Lock.Unlock()
```

| 方案 | 耗时 | 内存 | 分配次数 | 性能变化 |
|------|------|------|---------|---------| 
| 无锁 | 21ms | 10.8MB | 327K | 基准 |
| RWMutex | 30ms | 10.8MB | 327K | +43% |
| sync.Map | 59ms | 23MB | 708K | +181% |
