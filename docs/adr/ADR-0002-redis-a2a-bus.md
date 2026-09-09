# ADR-0002: 使用 Redis Pub/Sub 作为 A2A 消息总线

> 状态：已接受
> 日期：2026-09-02

## 背景

Agent 间需跨容器通信，候选方案：Redis Pub/Sub、NATS、Kafka、gRPC 直连、in-process channel。

## 决策

选择 **Redis Pub/Sub**。

关键理由：
1. **容器编排友好**：docker-compose 一行启动，K8s 有成熟 Helm chart
2. **简单性**：Pub/Sub 模型直观，消息序列化用 JSON 即可，无需 Kafka 的分区/消费者组复杂度
3. **性能足够**：本地往返 < 10ms，远低于 500ms 阈值（ADR-0008 验证）
4. **可降级**：in-process channel 作为开发期 fallback（ADR-0008），无 Redis 也能开发
5. **生态**：go-redis v9 成熟稳定，支持 context + 超时

淘汰方案：
- Kafka：过重，本项目消息量不需要持久化队列
- gRPC 直连：Agent 容器 IP 不固定，需额外服务发现
- NATS：可行但团队熟悉度低于 Redis

## 后果

- 正面：开发环境门槛低，消息总线可热插拔
- 负面：Redis 单点故障（容器化阶段需配 Redis Sentinel/Cluster）；Pub/Sub 无持久化，Agent 离线消息丢失（需配合任务重试）
- 后续：若消息量增长可平滑迁移到 NATS JetStream