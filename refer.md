# [ref-project] 技术&架构开发文档
> 项目仓库：https://github.com/[ref-project]/[ref-project]
> 版本基准：main分支，[ref-project]_HOST_SPEC_V1 规范，MIT协议
> [ref-project] 是**基于Rust开发的AI编码编排Agent**，本身不拥有大模型，负责调度5款主流AI编码CLI（Claude‑Code、Codex、OpenCode、Grok‑Build、Kimi‑Code），模拟真实软件开发团队角色，提供完整可审计、带质量门禁的软件交付流水线，项目演进自 super‑dev 项目。

## 目录
1. [项目概述](#1‑项目概述)
    - 1.1 产品定位
    - 1.2 解决行业痛点
    - 1.3 项目演进历史
    - 1.4 核心特性清单
2. [系统整体架构](#2‑系统整体架构)
    - 2.1 高层架构总览
    - 2.2 11个Rust Crates模块详解
    - 2.3 执行流转五层模型
    - 2.4 九角色团队编排模型
    - 2.5 Runtime宿主驱动层（ACP v1 + 私有协议）
3. [核心业务流水线](#3‑核心业务流水线)
    - 3.1 完整交付流程（绿场项目全链路）
    - 3.2 轻量化快速编辑流程
    - 3.3 质量门禁 Quality‑Gate
    - 3.4 Governance治理子系统
4. [知识库与检索子系统](#4‑知识库与检索子系统)
    - 4.1 检索架构（BM25 + 向量混合检索RRF‑HyDE）
    - 4.2 项目记忆与学习机制
5. [数据存储与产物规范](#5‑数据存储与产物规范)
    - 5.1 工作目录结构
    - 5.2 交付物说明
    - 5.3 审计日志格式
6. [配置体系](#6‑配置体系)
    - 6.1 用户全局配置 `~/.[ref-project]/config.toml`
    - 6.2 项目配置 `.[ref-project]rc`
    - 6.3 环境变量全集
7. [命令系统设计](#7‑命令系统设计)
    - 7.1 TUI交互（Slash命令）
    - 7.2 CLI子命令（CI/脚本）
    - 7.3 MCP服务扩展
8. [编译构建与二次开发](#8‑编译构建与二次开发)
    - 8.1 源码编译环境依赖
    - 8.2 编译参数说明（feature开关）
    - 8.3 源码阅读推荐路径
    - 8.4 扩展开发点：新增宿主Backend
9. [关键设计原则](#9‑关键设计原则)
10. [限制与风险说明](#10‑限制与风险说明)

## 1‑项目概述
### 1.1 产品定位
[ref-project] 定位为**AI编码项目总监Agent**：
- 不内置大模型API端点，**复用外部AI编码CLI作为大脑与执行器**；
- 自身负责流程编排、角色调度、计划管理、质量校验、审计留痕；
- 输出完整可交付工件：PRD、架构文档、UIUX设计、API契约、测试报告、审计证据包；
- 单Rust二进制分发，npm包仅作为shell分发器，真正业务逻辑全部在Rust二进制中。

> 比喻：[ref-project]是教练/项目经理，Claude‑Code/Kimi‑Code等是开发人员，教练制定计划、分配角色、做评审门禁，开发人员写代码。

支持5个一等公民宿主后端：
1. `claude‑code`：Claude‑Code私有流协议
2. `codex`：OpenAI Codex JSON‑RPC协议
3. `opencode`：OpenCode HTTP‑SSE协议
4. `grok‑build`：ACP v1标准协议
5. `kimi‑code`：ACP v1标准协议

### 1.2 解决行业痛点
普通AI编码工具普遍问题：
1. 拿到需求直接写代码，缺少PRD、架构、验收标准；
2. 前后端API契约不一致，接口对不上；
3. UI模板化、硬编码颜色、emoji图标等劣质AI生成代码；
4. 大量TODO占位、假数据，却标记任务完成；
5. 需求迭代后上下文丢失，历史决策遗忘；
6. 代码生成完成，但缺少质量报告、交付证据、审计记录；
7. 团队内部规范很难注入AI上下文。

[ref-project]通过**标准化流水线+确定性校验**解决上述问题，不是靠模型自省，而是机器硬校验。

### 1.3 项目演进历史
- 前身项目：`shangyankeji/super‑dev`，早期只是AI代码治理工具，专注拦截坏代码；
- super‑dev阶段核心：阻止AI输出不安全、不规范代码；
- 重写为[ref-project]：
  1. 从单点治理升级为完整全生命周期流水线治理；
  2. 从松散脚本重构为基于 `[ref-project]_HOST_SPEC_V1`（34条规范子句）规范驱动；
  3. 全部使用Rust重写，单二进制，跨平台；
  4. 目标转变：让AI像真实软件团队一样产出**可交付、可审计项目**。

### 1.4 核心特性清单
1. ✅ **基于角色的团队编排**：产品经理、架构师、UIUX、前后端、QA、安全、DevOps + Coordinator协调器，任务规模自适应，小bug不会启动全部角色；
2. ✅ **计划驱动执行**：`.[ref-project]/plan.json` DAG依赖任务图，用户可以干预、增删、调整任务顺序；
3. ✅ **确定性质量门禁**：不依赖模型自评；构建、lint、测试、API契约校验、安全扫描；
4. ✅ **前后端契约自动校验**：解析前端fetch/axios调用，对比OpenAPI契约，路径不匹配直接阻塞；
5. ✅ **本地优先知识库**：内置工程知识库，BM25必选，可选本地向量模型multilingual‑e5‑small；云端embedding必须显式双环境变量开启；
6. ✅ **可审计全链路**：所有评审、工具调用、校验结果写入audit JSONL证据包；
7. ✅ **本地记忆学习**：从失败案例沉淀经验规则，减少重复踩坑；
8. ✅ **TUI终端交互 + CLI脚本CI双接口，支持MCP协议对外暴露治理能力；
9. ✅ 多语言UI：简体中文、繁体中文、英文。

## 2‑系统整体架构
### 2.1 高层架构总览
```
┌─────────────────────────────────────────────────────┐
│  入口层：[ref-project] binary                              │
│  CLI / TUI(Ratatui) / CI Hook / MCP Server / Doctor │
└───────────────────────┬─────────────────────────────┘
                        │
┌───────────────────────▼─────────────────────────────┐
│ [ref-project]‑agent 【核心编排引擎】                        │
│ Router(意图路由) · Plan(DAG计划) · Director调度器     │
│ · Critics角色评审 · Trust信任等级 · Lessons经验库     │
└───┬───────────┬───────────┬──────────┬──────────────┘
    │           │           │          │
┌───▼───┐ ┌─────▼────┐ ┌────▼───┐ ┌────▼────────────┐
│spec   │ │governance│ │knowledge│ │contract        │
│规范定义│ │治理校验 │ │知识库检索│ │API契约校验引擎 │
└───┬───┘ └─────┬────┘ └────┬───┘ └────┬────────────┘
    │           │           │          │
┌───▼───────────▼───────────▼──────────▼────────────┐
│ [ref-project]‑runtime 运行时抽象层 Runtime trait         │
└──────────────────┬────────────────────────────────┘
                   │
    ┌──────────────┴──────────────┐
    ▼                             ▼
┌──────────────┐           ┌──────────────────┐
│[ref-project]‑host   │           │OfflineRuntime    │
│5个宿主驱动    │           │离线模板(测试CI)   │
└──────┬───────┘
       │
       ▼
┌────────────────────────────────────────────────────┐
│外部AI编码CLI子进程：Claude‑Code / Codex / OpenCode /│
│Grok‑Build / Kimi‑Code（真正执行代码生成）           │
└────────────────────────────────────────────────────┘
```

### 2.2 11个Rust Crates模块详解
整个项目是Rust workspace，共11个crates，每个职责边界清晰：

| crate名称 | 核心职责 |
|---|---|
| **[ref-project]** | 二进制入口；clap CLI、TUI启动、CI执行、pre‑write hook、MCP服务、doctor自检；对外暴露全部命令 |
| **[ref-project]‑spec** | [ref-project]_HOST_SPEC_V1规范，34条规范子句，阶段、门禁、运行时类型全部定义为Rust结构体，机器可读 |
| **[ref-project]‑governance** | 治理内核，113项检查规则；UI质量、安全漏洞、代码坏味道；审计日志；SOC2/ISO27001/EU‑AI‑Act合规映射；fail‑open设计（自身异常不会阻断工作） |
| **[ref-project]‑agent** | **最核心编排引擎**；意图路由RoutePlan；DAG任务计划Plan；Coordinator协调器；角色评审Critic；信任等级；证据驱动项目记忆；整个团队模型逻辑全部在此 |
| **[ref-project]‑runtime** | Runtime trait抽象；定义统一宿主接口；OfflineRuntime离线实现；隔离宿主差异 |
| **[ref-project]‑host** | 5个宿主子进程驱动实现；Claude‑Code/Codex/OpenCode私有协议；Grok‑Build/Kimi‑Code ACP‑v1协议；进程会话管理 |
| **[ref-project]‑process** | 跨平台进程生命周期管理；Windows Job‑Object FFI，保证完整杀掉子进程树，隔离unsafe代码 |
| **[ref-project]‑knowledge** | 知识库检索；Markdown分块；纯Rust BM25(CJK bigram分词器)；本地向量candle后端；HyDE查询扩展、RRF结果融合；repo‑map代码库符号索引；本地模型加载 |
| **[ref-project]‑contract** | API契约校验；解析架构文档API表生成OpenAPI json/yaml；扫描前端fetch/axios调用；前后端接口路径交叉校验 |
| **[ref-project]‑tui** | Ratatui终端UI；Markdown渲染；语法高亮；实时diff；构建进度卡片；预览链接；交互快捷键 |
| **[ref-project]‑i18n** | 国际化；zh‑CN / zh‑TW / en；区域语言自动探测 |
| **[ref-project]‑state** | 持久化存储；项目记忆、facts、pitfalls、recipes；捕获/召回策略控制；安全读写隔离 |

### 2.3 执行流转五层模型
用户每一条请求，经过5层处理：
1. **Route意图路由**：判定请求类型：Chat / Explain / QuickEdit / Debug / Build；评估深度、写权限、范围；必要时向用户澄清；在生成代码之前完成。
2. **Plan计划构建**：Build类型任务生成DAG依赖任务计划，保存`.[ref-project]/plan.json`，用户可通过`/plan`命令干预。
3. **Schedule调度执行**：
   - **写角色（前端/后端）**：串行执行，单写者模型，同一时间仅一个角色修改源码；
   - **评审角色（PM、架构、QA、安全等）**：并行启动独立子会话，返回结构化`RoleVerdict`（accept / blocking阻塞 / advisory建议）；角色之间不直接对话，只读写共享工件+输出Verdict评审结果。
4. **Verify自校正验证**：**确定性校验优先于模型自评**；构建、测试、lint、契约校验；阻塞项生成修复方案；指纹快照防止无限循环修复。
5. **Finalize交付产物**：输出对应深度工件；完整Build输出全套PRD/架构/质量报告；小修改输出轻量校验结果；更新项目记忆库。

> 关键点：评审角色失败（超时、解析失败）不会伪造通过，会将任务**park暂停**，等待用户`/continue`重试，不会假装成功完成。

### 2.4 九角色团队编排模型
9个席位：8个专业角色 + Coordinator协调器（技术负责人）

|角色|产出工件|模式|
|---|---|---|
|Product Manager产品经理|*-prd.md 用户故事、EARS验收标准|评审角色（并行子会话）|
|Architect架构师|*-architecture.md + openapi.* 数据模型API|评审角色（并行子会话）|
|UI/UX Designer|*-uiux.md 设计Token、组件状态|评审角色（并行子会话）|
|Frontend Engineer|前端组件页面|写角色，串行主会话|
|Backend Engineer|后端接口业务逻辑|写角色，串行主会话|
|QA Engineer|测试 + runtime‑proof.json运行时探针|评审角色（并行子会话）|
|Security Engineer|威胁模型、SAST安全扫描|评审角色（并行子会话）|
|DevOps|Docker、CI、deploy‑proof.json部署证据|评审角色（并行子会话）|
|**Coordinator协调器**|计划调度、门禁控制、审计日志|主控，不写代码，[ref-project]内部逻辑|

运行约束：
1. **写角色串行**：同一时刻只允许一个写角色修改源代码，避免冲突；
2. **评审角色并行独立会话**：每个评审角色都是全新独立会话，看不到主会话完整聊天记录，只读取共享工件；
3. 角色之间**禁止自由聊天**，通信媒介只有磁盘工件文件 + `RoleVerdict`结构化评审输出；
4. **任务规模自适应**：小bugfix不会启动全部角色；只有完整绿场构建才会拉起全部团队；
5. Coordinator拥有最终决定权，评审意见只是advisory，**质量门禁是确定性硬规则，不是模型主观意见**。

### 2.5 Runtime宿主驱动层（ACP v1 + 私有协议）
宿主分为两类实现：
1. **私有协议驱动**：`claude‑code`、`codex`、`opencode`，对接厂商私有流/RPC协议；
2. **ACP v1标准驱动**：`grok‑build`、`kimi‑code`，基于Agent Client Protocol v1，stdio上JSON‑RPC通信。

> [ref-project]不统一抹平各宿主能力差异；宿主不具备的能力，直接报告，不会模拟伪造。例如部分ACP宿主不支持会话resume持久化，则会新建会话移交上下文。

Runtime trait抽象接口定义行为：
- 会话启动、销毁；
- 发送任务载荷，接收事件流；
- 获取宿主暴露的模型信息；
- 工具调用事件转发；
- 暂停/继续；
- 获取宿主本身鉴权状态。

> [ref-project]**不接管、不存储宿主的API Key与凭证**；鉴权全部交给外部CLI自身完成。

## 3‑核心业务流水线
### 3.1 完整交付流程（绿场项目全链路）
完整Build流水线共9阶段，保存在spec规范；小任务会自动裁剪阶段：
```
clarify澄清 → research调研 → docs文档产出(PRD/架构/UIUX) → docs_confirm【人工门禁】
→ spec生成执行计划 → frontend前端开发 → preview_confirm【预览门禁】
→ backend后端开发 → quality质量门禁 → delivery交付打包
```

|阶段|产出文件|说明|
|---|---|---|
|clarify|output/*‑clarify.md|澄清需求：平台、权限、支付范围等，自动/手动模式|
|research|output/*‑research.md|知识库检索 + 宿主联网搜索，竞品分析|
|docs|prd.md / architecture.md / uiux.md|三大核心设计文档|
|docs_confirm|门禁点|**不写任何代码前，用户评审文档，确认后才继续**|
|spec|plan.json execution‑plan.md|DAG任务依赖计划|
|frontend|前端源码 + frontend‑notes.md|写角色产出前端代码|
|preview_confirm|门禁点|启动前端预览服务，用户确认后再开发后端|
|backend|后端源码 + backend‑notes.md|后端业务逻辑实现|
|quality|quality‑gate.json/md runtime‑proof.json|全套质量校验运行时探针启动应用访问路由|
|delivery|proof‑pack‑*.zip scorecard‑*.html|交付证据包，交付给客户/同事评审|

### 3.2 轻量化快速编辑流程
使用`/quick`命令强制轻量化路径，跳过调研、文档、完整团队编排，适合小修改：
`需求 → 路由判定QuickEdit → 直接调用宿主修改文件 → 轻量治理校验 → 完成`

### 3.3 质量门禁 Quality‑Gate
质量门禁独立运行，**不依赖模型主观判断**，在交付前执行全套校验：
校验清单：
1. PRD需求、验收标准完整性；
2. 架构API、数据模型、鉴权错误处理；
3. UIUX设计Token、组件状态、暗色模式；
4. **前端API调用与OpenAPI契约交叉校验**；
5. UI坏味道：emoji图标、硬编码颜色、AI模板代码；
6. 编译构建、单元测试、lint、类型检查；
7. Dockerfile、CI配置、环境示例`.env.example`；
8. 密钥、密码泄露扫描；
9. 审计日志完整性、合规映射SOC‑2 / ISO27001 / EU‑AI‑Act；
10. Runtime探针：启动应用，访问路由生成`runtime‑proof.json`。

配置阈值：
```toml
[quality]
threshold = 90
skip_checks = []
```
> threshold是得分阈值，超过才允许交付；可跳过指定检查项。

### 3.4 Governance治理子系统
113条治理检查规则，覆盖UI、安全、前后端工程风险；全部可配置关闭、路径排除。
**fail‑open设计**：治理引擎内部异常，不会阻断开发流程，默认返回通过，防止工具本身故障阻塞业务。

治理触发入口：
1. Pre‑write钩子：代码写入文件前（对接宿主CLI钩子）；
2. CI/Pre‑commit；
3. MCP服务对外暴露govern_file；
4. Quality‑gate交付前扫描。

配置样例 `.[ref-project]/rules.toml`
```toml
[disabled]
clauses = []
[exclusions]
paths = ["src/legacy/**","**/*.test.ts"]
```

输出写入 `.[ref-project]/audit/*.jsonl` 审计记录。

## 4‑知识库与检索子系统
### 4.1 检索架构（BM25 + 向量混合检索RRF‑HyDE）
检索引擎`[ref-project]‑knowledge`，**BM25是保底必选，向量是可选增强**。

检索流程：
1. 用户需求+当前阶段 → HyDE查询扩展生成假想文档；
2. 两路并行召回：
   - 路径A：纯Rust BM25（CJK bigram分词），不需要任何模型；
   - 路径B：本地向量模型`multilingual‑e5‑small f16`（224MB），或者**必须双变量开启才生效的云端embedding**；
3. RRF reciprocal rank fusion融合两路召回结果；
4. 返回top‑k块，注入宿主上下文；

> 云端embedding安全限制：
> 必须同时设置：`OPENAI_EMBED_KEY` + `[ref-project]_ALLOW_CLOUD_EMBED=1`，普通`OPENAI_API_KEY`不会触发上传；向量不可用时自动降级为纯BM25，不会报错失败。

本地向量模型：npm包首次运行自动下载校验sha256；源码编译需要手动放置模型文件，开启feature `vector‑local`。

额外能力：`repo‑map`代码库符号索引，对项目源码做符号扫描，计算重要度排序，做上下文预算裁剪。

### 4.2 项目记忆与学习机制
全部记忆保存在项目目录`.[ref-project]/memory/`，**不会跨项目自动共享**。

|存储对象|存储位置|说明|
|---|---|---|
|pitfalls 问题事件库|`.[ref-project]/learned/_raw/dev‑errors.jsonl`|记录运行失败事件；重复发生才会生成待验证规则|
|lessons经验规则|`.[ref-project]/memory/learned‑skills/`|只有修复后，验证器确认修复有效，规则才正式生效|
|facts项目事实|`facts.jsonl`|提取项目环境事实，带来源证据；过期会标记tombstone|
|recipes历史解决方案|`recipes.jsonl`|历史交付方案，严格匹配栈，仅作为建议|
|run‑notes运行时笔记|`.[ref-project]/run‑notes.md`|当前运行阶段临时记录，不能跨运行复用|

记忆捕获、召回可以独立开关命令：
```bash
[ref-project] memory capture off --scope project --store facts
[ref-project] memory recall off --scope project --store recipes
```

## 5‑数据存储与产物规范
### 5.1 工作目录结构
```
your‑project/
├── output/                     # 产出文档
│   ├─ *‑clarify.md
│   ├─ *‑research.md
│   ├─ *‑prd.md
│   ├─ *‑architecture.md
│   ├─ *‑uiux.md
│   ├─ *‑execution‑plan.md
│   ├─ *‑quality‑gate.md/json
│   └─ ...
├── .[ref-project]/
│   ├─ plan.json                # DAG任务计划
│   ├─ workflow‑state.json      # 运行时状态
│   ├─ rules.toml               # 治理规则配置
│   ├─ contracts/
│   │   ├─ openapi.json/yaml    # 生成API契约
│   ├─ audit/
│   │   ├─ tool‑calls.jsonl
│   │   ├─ verify.jsonl
│   │   └─ frontend‑api‑calls.jsonl
│   ├─ memory/                  # facts recipes pitfalls
│   └─ learned/
├── release/
│   ├─ proof‑pack‑*.zip         # 交付证据包
│   └─ scorecard‑*.html         # 质量评分报告
└── .[ref-project]rc                   # 项目配置文件
```

### 5.2 交付物说明
1. **proof‑pack‑*.zip**：完整交付证据包，可以交给客户、评审；包含审计日志、设计文档、校验报告；
2. **scorecard‑*.html**：可视化质量评分单；
3. runtime‑proof.json：探针运行结果，证明服务真实启动访问。

> [ref-project]不会自动git merge/push代码，所有修改留在本地工作区，可以手动提交。

### 5.3 审计日志格式
`.[ref-project]/audit/*.jsonl`，每行一条json；记录工具调用、评审Verdict、校验结果；不可篡改证据链。

## 6‑配置体系
三层配置：
1. 用户全局配置 `~/.[ref-project]/config.toml`；
2. 项目本地配置 `.[ref-project]rc`；
3. 环境变量（优先级最高）。

### 6.1 用户全局配置 `~/.[ref-project]/config.toml`
```toml
backend = "claude‑code"
lang = "zh‑CN"
# 注意：[ref-project]不管理模型；模型需要在对应宿主CLI内部配置
```

### 6.2 项目配置 `.[ref-project]rc`
```toml
[quality]
threshold = 90
skip_checks = []

[pipeline]
skip_phases = []
max_review_rounds = 3
auto_approve_gates = true

[knowledge]
enabled = true
engine = "hybrid"
top_k = 6
```

### 6.3 环境变量全集
|环境变量|作用|
|---|---|
|[ref-project]_*_BIN|覆盖各个宿主二进制路径，[ref-project]_CLAUDE_BIN、[ref-project]_KIMI_BIN|
|[ref-project]_WORKER_TIMEOUT|单任务超时秒数，默认300|
|[ref-project]_VERIFY_TIMEOUT_SECS|校验超时，默认120|
|[ref-project]_NO_GOAL_MODE=1|关闭 /goal目标模式|
|[ref-project]_EMBED_MODEL_DIR|本地向量模型目录|
|OPENAI_EMBED_KEY|远程embedding密钥，需要配合[ref-project]_ALLOW_CLOUD_EMBED=1才生效|
|[ref-project]_ALLOW_CLOUD_EMBED=1|允许云端embedding|

## 7‑命令系统设计
两套命令体系：TUI内Slash命令；外部CLI子命令，两者能力对齐。

### 7.1 TUI交互（Slash命令）
启动直接运行`[ref-project]`进入TUI终端聊天界面，输入`/`唤起命令面板。
分类：切换宿主、流程控制`/run /goal /quick /plan /continue /revise`、预览、检查、知识管理、记忆管理。

### 7.2 CLI子命令（CI脚本自动化）
```bash
[ref-project] init                     # 初始化项目[ref-project].yaml
[ref-project] adopt                    # 导入已有存量项目
[ref-project] run "需求描述" --backend kimi‑code   # 非交互运行完整流水线
[ref-project] quick "小修改"
[ref-project] verify --runtime         # 执行质量校验，运行探针
[ref-project] report                   # 输出合规报告
[ref-project] ci                       # CI治理扫描
[ref-project] mcp serve                # 启动MCP服务
[ref-project] knowledge‑manage add ./docs # 添加自定义知识库
```

### 7.3 MCP服务扩展
`[ref-project] mcp serve`暴露治理能力给其他MCP客户端；提供`govern_file`等工具调用。

## 8‑编译构建与二次开发
### 8.1 源码编译环境依赖
- Rust ≥1.88；
- cargo；
- Node.js ≥18（仅用于编译测试npm分发shell，非必须）。

```bash
git clone https://github.com/[ref-project]/[ref-project].git
cd [ref-project]
# 默认构建，BM25可用，无本地向量
cargo build --release --workspace
# 需要本地向量检索，开启vector‑local feature
cargo build --release --features vector‑local
```

> ⚠️源码编译不会自动下载向量模型文件；需要手动下载`multilingual‑e5‑small f16`放置到目录，配置`[ref-project]_EMBED_MODEL_DIR`。npm包分发才会自动下载模型。

### 8.2 编译feature开关
1. `vector‑local`：启用candle本地向量后端；默认关闭；
2. 默认：仅BM25检索，remote向量仍需要环境变量开关。

### 8.3 源码阅读推荐路径
1. `spec/[ref-project]_HOST_SPEC_V1.md`：顶层规范文档；
2. `crates/[ref-project]‑spec/src/lib.rs`：规范转Rust结构体；
3. `crates/[ref-project]‑agent/src/router.rs`：意图路由逻辑；
4. `crates/[ref-project]‑agent/src/director.rs`：核心调度循环；
5. `crates/[ref-project]‑governance/src/rules.rs`：治理规则；
6. `crates/[ref-project]‑host/`：宿主驱动实现；
7. `crates/[ref-project]/src/main.rs`：二进制入口。

### 8.4 扩展开发点：新增宿主Backend
新增AI编码CLI宿主，需要实现`Runtime` trait：
1. 在`[ref-project]‑host`增加新驱动，实现Runtime trait；
2. 处理子进程生命周期、消息协议转换；
3. 适配会话、工具调用、事件映射；
4. 实现与Agent层交互；
5. 更新spec常量，增加Backend ID；
6. TUI、CLI命令增加新后端选项；
> 必须处理宿主能力差异，宿主不具备的能力，返回不可用，禁止模拟伪造。

## 9‑关键设计原则
1. **模型只是工人，[ref-project]掌握计划与验收标准**：不相信模型自我评估，重要校验全部机器硬执行；
2. **本地优先，默认不上传用户代码/文档到云端**；云端向量必须双显式环境变量开启；
3. **fail‑open治理**：治理组件故障不能阻断开发；
4. **角色通信只通过工件文件+结构化Verdict，禁止自由对话；评审失败不伪造成功；
5. **规模自适应**：小任务轻量流程，大项目完整团队流水线；
6. **不持有密钥**：鉴权全部交给底层宿主CLI；
7. **一切可审计**：关键操作留证据，产出可交付证据包。

## 10‑限制与风险说明
1. 依赖外部AI编码CLI正常安装登录；[ref-project]本身不提供模型能力；
2. 本地向量模型需要约224MB磁盘空间；
3. NFS/SMB网络文件系统，进程锁语义不完全可靠，多机共享工作目录会有锁风险；
4. Windows‑ARM使用x64兼容层运行；
5. 源码构建需要手动准备向量模型文件；npm分发自动处理；
6. 记忆学习机制只是辅助减少重复错误，不能保证完全消除模型错误。

---

如果你需要，我可以基于这份文档进一步输出：
1. 可导出Markdown文档版本；
2. PlantUML架构图文本；
3. 快速上手部署文档；
4. 二次开发API接口文档。

这个项目适合做网页技术文档，工作任务模式可以把架构图、模块说明做成可预览网页，要不要用它继续？