# AiCodingAgentTeam 容器化部署

## 架构概览

```
                    ┌──────────────────────────────────────────────────┐
                    │              docker-compose network               │
                    │                                                    │
                    │  ┌──────────┐    ┌──────────────┐               │
                    │  │ Coordinator │  │   Redis Bus  │               │
                    │  │  :8080 grpc│──│  :6379 pub/sub│              │
                    │  │  :8083 a2a │  └──────────────┘               │
                    │  └─────┬──────┘                                  │
                    │        │ workspace volume                        │
                    │  ┌─────┴──────────────────────────────────┐    │
                    │  │                                        │    │
                    │  │  agent-pm  agent-architect  agent-qa   │    │
                    │  │  agent-security  agent-devops           │    │
                    │  │  host-codex  host-opencode             │    │
                    │  └────────────────────────────────────────┘    │
                    └──────────────────────────────────────────────────┘
```

## 快速启动

```bash
cd deploy/compose
docker compose up -d
```

验证：

```bash
# 检查所有容器状态
docker compose ps

# 检查 Coordinator 就绪
docker compose logs coordinator | head -20

# 测试 gRPC 连接（需安装 grpcurl 或使用 TUI）
# gRPC :8080, A2A HTTP :8083
```

## 服务说明

| 服务 | 端口 | 说明 |
|---|---|---|
| coordinator | 8080 (gRPC), 8081 (MCP), 8082 (ACP), 8083 (A2A HTTP) | 编排引擎入口 |
| a2a-bus | 6379 | Redis Pub/Sub 消息总线 |
| agent-pm | — | PM 评审角色 |
| agent-architect | — | 架构师评审角色 |
| agent-qa | — | QA 评审角色 |
| agent-security | — | 安全评审角色 |
| agent-devops | — | DevOps 评审角色 |
| host-codex | — | Codex CLI 宿主容器 |
| host-opencode | — | OpenCode CLI 宿主容器 |

## A2A 消息总线

容器化模式使用 Redis Pub/Sub 实现跨容器 A2A 通信（ADR-0002）。

- 环境变量 `AICODINGAGENTTEAM_A2A_BUS=redis://a2a-bus:6379` 触发 Redis 模式
- 未设置时自动降级为 in-process 模式（单容器开发）
- Redis 不可用时 fail-open 降级到 in-process（ADR-0006）

## 健康检查

```bash
# Redis 健康检查
docker compose exec a2a-bus redis-cli ping

# Coordinator A2A 端点
curl http://localhost:8083/.well-known/agent.json

# Coordinator gRPC (需 grpcurl)
grpcurl -plaintext localhost:8080 aicodingagentteam.v1.Coordinator/Verify
```

## 开发模式（无 Docker）

不启动容器，直接运行：

```bash
# in-process Bus（无 Redis）
go run ./cmd/aicodingagentteam serve

# TUI 连接
cd tui && node dist/cli.js
```

## 工作区卷

所有容器共享 `workspace` 命名卷，挂载到 `/workspace`。
单写者约束（ADR-0004）通过 scheduler 文件锁保证。

## 故障排查

```bash
# 查看某容器日志
docker compose logs coordinator
docker compose logs agent-qa

# 进入容器调试
docker compose exec coordinator sh

# 重建镜像
docker compose build --no-cache coordinator
```

## 相关 ADR

- ADR-0002: Redis Pub/Sub A2A 总线
- ADR-0003: 容器化部署（每角色独立容器）
- ADR-0004: 单写者模型
- ADR-0005: 不持有密钥
- ADR-0006: fail-open 治理
- ADR-0010: Volume 并发读写