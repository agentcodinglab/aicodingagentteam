# Stage 2 — 暗色霓虹双色调站点升级

> 承接 Stage 1（首页最佳实践）后，基于 [ref-project] 参考站的设计语言，对本站做整体视觉升级。

## 0. 目标

保留 Tailwind + 9 语言 i18n 与 GitHub Pages 静态部署不变的前提下，将站点由「亮色蓝单色」升级为「暗色 cyan+magenta 双色霓虹」风格，统一字体、动效与终端交互语言。

## 1. 设计 token

| token | 值 | 用途 |
|---|---|---|
| `--bg` | `#050508` | 页面主背景 |
| `--bg-2` | `#090a10` | 次级面板背景 |
| `--panel` | `rgba(5,8,14,0.88)` | 卡片/终端面板 |
| `--text` | `#f8fcff` | 主文本 |
| `--muted` | `#aebfd0` | 次要文本 |
| `--muted-2` | `#768a9c` | 弱化文本 |
| `--cyan` | `#00d2ff` | 主强调色 |
| `--cyan-2` | `#00aeff` | 强调色变体 |
| `--magenta` | `#ff2a85` | 次强调色 |
| `--magenta-2` | `#c1176b` | 次强调色变体 |
| `--duo` | `linear-gradient(120deg, #00d2ff 0%, #3a7bd5 50%, #ff2a85 100%)` | 渐变 |
| `--green` | `#20e3b2` | 成功/确认 |
| `--gold` | `#f2b84b` | 警告/阶段标签 |
| `--line` | `rgba(0,210,255,0.20)` | 细边框 |
| `--line-strong` | `rgba(0,210,255,0.45)` | 强边框 |

## 2. 字体

- **标题**：Space Grotesk Variable（700 / 600）
- **正文**：Manrope Variable（400 / 500 / 600）
- **代码/终端**：JetBrains Mono Variable（400 / 600）

通过 `@fontsource-variable/*` 自托管，避免 CDN 依赖。

## 3. 组件升级

### 3.1 Hero（终端打字机）
- 三 slide 自动循环：`run / continue / gate`，5s/页
- 每页 8–11 行；typewriter 步进（300–900ms/行）
- 阶段色：prompt=cyan，sys=muted，stage=gold，file=cyan-2，ok=green，done=magenta
- 闪烁光标 + 「RESTART」按钮重置当前 slide
- 右侧 mascot 占位（CSS 渐变块，避免外部 PNG）

### 3.2 ScrollCue
首屏底部向下箭头，呼吸缩放 + 渐变描边。

### 3.3 Reveal / Tilt
`useReveal` hook（IntersectionObserver，threshold 0.15）+ `<Reveal>` 包装器。渐入：`translateY(20px) → 0` + `opacity 0 → 1`，缓动 600ms。Tilt：hover 时 `rotateX/Y` ±1.5° 跟随鼠标。

### 3.4 全局 Section
统一 `bg-[var(--bg)]` + 顶部/底部 `border-[var(--line)]`，间隔 `py-24`。

## 4. 内页升级

`/features` `/architecture` `/quickstart` 共享同一暗色 layout：
- 顶部 hero 紧凑版（h1 + 副标题 + 阶段标签）
- 内容卡片化（同首页 Card 但加 duo 边框）
- 顶部新增「← Back to home」返回链

## 5. 资产

- `public/og.png`（1200×630，霓虹双色 + 站点 logo + 标语）— 程序化生成 PNG
- `public/favicon.svg`（霓虹方块 + 「A」字符，cyan→magenta 渐变）

## 6. 验证

- `tsc --noEmit` exit 0
- `next build` exit 0，93 静态页
- 10 URL（root + 9 语言）均 200
- CI 绿
- 视觉：首页三 slide 终端可见、字体替换生效、暗底生效

## 7. 文件清单（预计）

| 操作 | 文件 |
|---|---|
| 新建 | `components/ui/Reveal.tsx`、`components/ui/ScrollCue.tsx`、`components/marketing/HeroStage.tsx`、`components/marketing/Section.tsx` |
| 改写 | `tailwind.config.ts`、`app/globals.css`、`app/layout.tsx`、`components/marketing/Hero.tsx`、`components/marketing/FeatureGrid.tsx`、`components/marketing/Stats.tsx`、`components/marketing/CodeExample.tsx`、`components/marketing/ArchitectureDiagram.tsx`、`components/marketing/Quickstart.tsx`、`components/marketing/LogoCloud.tsx`、`components/marketing/FinalCTA.tsx`、`app/[locale]/(marketing)/features/page.tsx`、`app/[locale]/(marketing)/architecture/page.tsx`、`app/[locale]/(marketing)/quickstart/page.tsx`、`components/docs/SiteHeader.tsx`、`components/docs/SiteFooter.tsx`、`app/[locale]/page.tsx` |
| 新增 | 9 语言 `home.hero.slide*` keys + 9 语言 `home.scrollCue` + 9 语言 `home.hero.restart` |
| 资产 | `public/og.png`、`public/favicon.svg` |