# AiCodingAgentTeam

## 🌐 Languages / 语言

[English](README.md) · [简体中文](README.zh.md) · [日本語](README.ja.md) · [한국어](README.ko.md) · [Français](README.fr.md) · [Deutsch](README.de.md) · [Русский](README.ru.md) · [Español](README.es.md) · [Italiano](README.it.md)
> Go 言語で実装された AI コーディング・オーケストレーション・プラットフォーム。大規模言語モデル自身は保持せず、外部の AI コーディング CLI（Codex と OpenCode は実駆動、Claude-Code と DeepSeek-DSH はスタブ）に処理を委譲し、A2A プロトコルで協働する 9 ロールのソフトウェア開発チームをシミュレートします。決定論的な品質ゲートとガバナンス監査を備え、コンテナで配布され、gRPC / MCP / ACP / A2A の 4 種プロトコルを外部に公開、TypeScript TUI クライアントも提供します。

[![CI](https://github.com/agentcodinglab/aicodingagentteam/actions/workflows/ci.yml/badge.svg)](https://github.com/agentcodinglab/aicodingagentteam/actions/workflows/ci.yml)
[![Go](https://img.shields.io/badge/Go-1.25-00ADD8)](https://go.dev)
[![License](https://img.shields.io/badge/license-MIT-blue)](LICENSE)

## 主な特徴

- **オーケストレーションと実行の分離** — Coordinator はオーケストレーションのみ行い、コードは書きません。ホスト CLI は実行のみで意思決定は行いません。
- **コンテナーをロール単位に** — 9 つのチームロール（PM / Architect / Frontend / Backend / QA / Security / DevOps など）をそれぞれ独立したコンテナーで実行し、スケールや差し替えが自在です。
- **A2A プロトコルによる協働** — エージェント間の構造化通信。InProc（開発用）と Redis Pub/Sub（コンテナ構成）に対応します。
- **決定論的な品質ゲート** — `golangci-lint` + `go vet` + `go test` を機械的に実行。モデルの自己評価には依存しません。
- **RAG ナレッジベース + プロジェクト記憶** — BM25 検索で文脈に注入。失敗の教訓と過去の手案をメモリに蓄積します。
- **4 種の外部プロトコル** — gRPC（TUI 通信）、MCP（外部ツール統合）、ACP（標準 Agent クライアント）、A2A（インスタンス間メッシュ）。
- **ローカル優先** — コードはコンテナー外に出さず、API キーも保持しません。認証はすべてホスト CLI に委譲します。

## アーキテクチャ概要

```
User → TypeScript TUI (Ink) → gRPC → Coordinator
                                       ├── Router (intent routing)
                                       ├── Planner (DAG plan)
                                       ├── Scheduler (role dispatch)
                                       │    ├── A2A Bus → Reviewer Agents (parallel)
                                       │    └── Single-Writer Lock → Writer Agents (serial)
                                       ├── Knowledge (BM25 RAG retrieval)
                                       ├── Memory (facts/pitfalls/lessons)
                                       └── Quality Gate (deterministic checks)
                                            ↓
Host CLIs: Codex (real) | OpenCode (real) | Claude (stub) | DSH (stub)
```

詳細なアーキテクチャ図は [docs/04-系统架构图.md](docs/04-系统架构图.md) を参照してください。

## クイックスタート

### シングルバイナリ・モード

```bash
# ビルド
go build -o bin/aicodingagentteam ./cmd/aicodingagentteam

# プロジェクトの初期化
./bin/aicodingagentteam init

# パイプライン全体の実行
./bin/aicodingagentteam run "REST API を構築して" --backend codex

# クイック編集
./bin/aicodingagentteam quick "ログインボタンのスタイルを修正"

# 品質検証
./bin/aicodingagentteam verify

# RAG ナレッジベース・デモ
./bin/aicodingagentteam knowledge demo

# Coordinator サービスの起動（gRPC + A2A HTTP）
./bin/aicodingagentteam serve
```

### TUI クライアント

```bash
cd tui && npm install && npm run build

# Coordinator を起動
cd .. && go run ./cmd/aicodingagentteam serve

# TUI を起動（別のターミナルで）
cd tui && node dist/cli.js
# デモモード（Coordinator 不要）
node dist/cli.js --demo
```

### コンテナデプロイ

```bash
docker compose -f deploy/compose/docker-compose.yml up -d
```

## コマンド一覧

| コマンド | 説明 |
|---|---|
| `init` | プロジェクト設定を初期化 |
| `run "requirement"` | パイプライン全体を実行 |
| `quick "description"` | 軽量なクイック編集 |
| `verify` | 品質検証を実行 |
| `govern [--ci] [path]` | ガバナンス・スキャン |
| `report` | コンプライアンス・レポートを出力 |
| `serve` | Coordinator を起動（gRPC + A2A） |
| `knowledge index [dir]` | ディレクトリをナレッジベースにインデックス |
| `knowledge search "query"` | 上位 5 件のナレッジチャンクを取得 |
| `knowledge demo` | RAG + メモリ end-to-end デモ |
| `version` | バージョン表示 |

## プロジェクト構成

```
agent_team/
├── cmd/aicodingagentteam/     # Go バイナリ・エントリ
├── internal/
│   ├── a2a/                   # A2A プロトコル（InProc + Redis）
│   ├── acp/                   # Agent Client Protocol
│   ├── agent/                 # 9 ロールレジストリ
│   ├── audit/                 # 監査ログ
│   ├── config/                # 設定ローダ
│   ├── coordinator/           # オーケストレーション・コア（5 層フロー）
│   ├── governance/            # ガバナンス・ルール（113+ ルール）
│   ├── host/                  # ホストドライバ（codex/opencode/claude/dsh）
│   ├── knowledge/             # RAG 検索エンジン（BM25）
│   ├── mcp/                   # Model Context Protocol
│   ├── memory/                # プロジェクトメモリ（facts/pitfalls/lessons）
│   ├── planner/               # DAG プランビルダ
│   ├── qualitygate/           # 品質ゲート・エンジン
│   ├── router/                # インテント・ルータ
│   ├── scheduler/             # ロールスケジューラ + 単一ライタ・ロック
│   └── types/                 # 共有型
├── pkg/
│   ├── api/                   # gRPC proto + server
│   ├── contracts/             # フロント/バック間契約
│   └── runtime/               # ホストドライバ trait
├── tui/                       # TypeScript TUI クライアント
├── deploy/                    # コンテナデプロイ
├── proto/                     # gRPC proto 定義
└── docs/                      # ドキュメント（要件/アーキ/計画/ADR/spec/plan）
```

## 品質ベースライン

| 観点 | 基準値 | 実績 |
|---|---|---|
| Go テストカバレッジ | ≥ 80%（コア ≥ 90%） | router 100% ・ scheduler 96% ・ governance 96% ・ codex 94.9% |
| Lint | 新規警告 0 | golangci-lint + eslint |
| ビルド | ≤ 5 分 | CI 全工程 ~3 分 |
| セキュリティ | 高危険度 CVE 0 | govulncheck + npm audit |

## 開発

```bash
make all      # lint + vet + test + build
make test     # 全テスト実行
make lint     # golangci-lint
make run      # build + serve
```

## ドキュメント

- [AGENTS.md](AGENTS.md) — プロジェクト規約（ルートドキュメント）
- [docs/01-需求分析文档.md](docs/01-需求分析文档.md) — プロダクト要件
- [docs/02-技术架构设计.md](docs/02-技术架构设计.md) — アーキテクチャ設計
- [docs/03-系统设计与实施规划.md](docs/03-系统设计与实施规划.md) — 実装計画
- [docs/adr/](docs/adr/) — アーキテクチャ決定記録（ADR-0001 ~ 0013）
- [docs/CONSTRAINTS.md](docs/CONSTRAINTS.md) — 品質レッドライン

## ライセンス

[MIT](LICENSE)