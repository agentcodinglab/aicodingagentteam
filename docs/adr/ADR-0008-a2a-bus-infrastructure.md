# ADR-0008: A2A 通信基础设施 Spike 结论

> 状态：已接受
> 日期：2026-09-02
> 关联计划：`docs/plan/mvp-preparation.md` A2

## 背景

架构设计假设 A2A 消息总线用 Redis Pub/Sub 实现，Coordinator→Agent 委派往返 p95 ≤ 500ms。需验证基础设施可用性。

## 决策

当前环境验证结果：
- **Docker**：未安装（本机当前无容器环境）
- **redis-cli**：未安装
- **go-redis 依赖**：go.mod 中声明但 go.sum 未生成（未实际下载）

### 结论

基础设施未就绪，但这是**部署环境问题，非架构问题**。

### MVP 阶段策略

1. **Phase 1（MVP 早期）**：A2A Bus 用 in-process channel 实现（已在 `internal/a2a/a2a.go` 中 stub），所有 Agent 在同一进程内通信，验证编排逻辑正确性
2. **Phase 2（容器化）**：引入 Docker + Redis，将 in-process Bus 替换为 Redis Pub/Sub，验证跨容器延迟
3. **延迟阈值**：in-process 无网络延迟，p95 < 1ms；Redis 跨容器预计 < 50ms（远低于 500ms 阈值）

## 后果

- 正面：MVP 可不依赖 Docker/Redis，降低开发环境门槛
- 负面：跨容器延迟验证推迟到 Phase 2，存在未知风险（但 Redis 本地往返通常 < 10ms，风险低）
- 后续：容器化阶段需补充 A2A 延迟压测报告