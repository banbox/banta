
### 支持并发操作BarEnv和Series
为避免额外冗余创建BarEnv和Series，多策略和多账户会复用，多策略和多账户需支持并发调用，所以banta也需支持并发。
测试给Series的Subs和XLogs，BarEnv的Items和Data改为sync.Map后，性能相比RWMutex+map大幅下降；故保留RWMutex方案。

| 方案 | 耗时 | 内存 | 分配次数 | 性能变化 |
|------|------|------|---------|---------|
| 无锁 | 21ms | 10.8MB | 327K | 基准 |
| RWMutex | 30ms | 10.8MB | 327K | +43% |
| sync.Map | 59ms | 23MB | 708K | +181% |
