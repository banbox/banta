
### 指标源码布局

指标实现按“计算语义”分族，并在两个 API 包中保持同名文件：

- 根包 `indicators_*.go` 是逐 K 线推进、带 `Series` 缓存的有状态实现。
- `tav/indicators_*.go` 是一次性处理 `[]float64` 的无状态向量实现。
- `price`、`moving`、`momentum`、`trend`、`volatility`、`volume`、`channels`、`regression` 和 `patterns` 文件分别承载对应族；`indicators_helpers.go` 只保留跨族基础/适配辅助函数。
- `test_helpers_test.go` 提供共享的状态/向量比较器；`testdata/fixture_btc_58.json` 是对照测试的唯一 OHLCV 输入，指标 oracle 文件只保存结果。未被测试引用的汇总报告不纳入版本库。
- 新指标实现规范位于仓库根目录的 `SKILL.md`。新增指标必须同时实现两个 API，并通过 Python 对照脚本和 Go parity 测试。

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
