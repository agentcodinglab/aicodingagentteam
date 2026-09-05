# Plan: Website (阶段 1) — Next.js 9-lang static site via GitHub Pages

## Goal
Bootstrap `website/` as a Next.js 14 App Router project that:
- Renders a marketing landing page (Hero + Features + Quickstart + Architecture overview)
- Renders a docs site skeleton with sidebar + TOC
- Supports 9 locales (en/zh/ja/ko/fr/de/ru/es/it) via URL-segment i18n
- Is statically exported (`output: 'export'`) and deployed to GitHub Pages via Actions
- Sources 5 core docs (requirements/architecture/implementation-plan/quality-constraints/domain-model) in both English and Chinese

## Non-goals (out of scope for stage 1)
- Full ADR/spec/plan dynamic routing (stage 2)
- 7 remaining locales (ja/ko/fr/de/ru/es/it) full content — only UI string dictionaries ship in stage 1
- Mermaid interactive diagrams, hero animations (stage 4)
- Version-bound docs / mike (stage 5)

## Architecture decisions
- **Static export** — `output: 'export'` in `next.config.mjs`, basePath `/aicodingagentteam` for GitHub Pages user-site URL
- **i18n** — `next-intl` v3 with `[locale]` dynamic segment; UI strings in `messages/<locale>.json`; content in `content/docs/<locale>/...`
- **Content rendering** — `next-mdx-remote` for client-side MDX, `gray-matter` for frontmatter
- **Styling** — Tailwind CSS v3 + custom components (no third-party UI library to keep license surface minimal)
- **Deployment** — GitHub Actions: `actions/checkout@v5` → `actions/setup-node@v5` → `npm ci` → `npm run build` → `actions/upload-pages-artifact@v3` → `actions/deploy-pages@v4`
- **Repo layout** — `website/` subdirectory (monorepo style); does not pollute Go / TUI roots

## Content migration (stage 1)
| Source | Destination |
|---|---|
| `docs/01-需求分析文档.md` (CN) | `website/content/docs/zh/requirements.md` |
| `docs/01-需求分析文档.md` translated (EN) | `website/content/docs/en/requirements.md` |
| `docs/02-技术架构设计.md` (CN) | `website/content/docs/zh/architecture.md` |
| `docs/02-技术架构设计.md` translated (EN) | `website/content/docs/en/architecture.md` |
| `docs/03-系统设计与实施规划.md` (CN) | `website/content/docs/zh/implementation-plan.md` |
| `docs/03-系统设计与实施规划.md` translated (EN) | `website/content/docs/en/implementation-plan.md` |
| `docs/CONSTRAINTS.md` (EN, native) | `website/content/docs/en/quality-constraints.md` |
| `docs/CONSTRAINTS.md` translated (CN) | `website/content/docs/zh/quality-constraints.md` |
| `docs/domain.md` (CN, native) | `website/content/docs/zh/domain-model.md` |
| `docs/domain.md` translated (EN) | `website/content/docs/en/domain-model.md` |

Note: ADR/spec/plan dynamic routes are not wired in stage 1; their content is reachable through the existing `docs/` directory in the repo for now.

## File layout (after stage 1)
```
agent_team/
├── website/
│   ├── app/
│   │   ├── [locale]/
│   │   │   ├── layout.tsx
│   │   │   ├── page.tsx                          ← Marketing landing
│   │   │   ├── (marketing)/
│   │   │   │   ├── features/page.tsx
│   │   │   │   ├── architecture/page.tsx
│   │   │   │   └── quickstart/page.tsx
│   │   │   └── docs/
│   │   │       ├── layout.tsx                    ← Sidebar + TOC
│   │   │       ├── page.tsx                      ← redirect to requirements
│   │   │       └── requirements/page.tsx
│   │   │       └── architecture/page.tsx
│   │   │       └── implementation-plan/page.tsx
│   │   │       └── quality-constraints/page.tsx
│   │   │       └── domain-model/page.tsx
│   ├── components/
│   │   ├── ui/                                   ← Button, Card, Badge
│   │   ├── marketing/                            ← Hero, FeatureGrid, ArchitectureOverview, Quickstart
│   │   └── docs/                                 ← Sidebar, TOC, LocaleSwitcher, ThemeToggle
│   ├── content/
│   │   └── docs/{en,zh}/                         ← 5 core docs each
│   ├── messages/{en,zh,ja,ko,fr,de,ru,es,it}.json
│   ├── lib/i18n.ts
│   ├── lib/nav.ts
│   ├── public/
│   ├── next.config.mjs
│   ├── tailwind.config.ts
│   ├── postcss.config.mjs
│   ├── tsconfig.json
│   ├── package.json
│   └── .gitignore
└── .github/workflows/docs-site.yml
```

## Implementation steps
1. Scaffold `website/` (package.json, tsconfig, next.config.mjs, tailwind, postcss)
2. Install deps: `next@14`, `next-intl`, `react`, `react-dom`, `tailwindcss`, `typescript`, `next-mdx-remote`, `gray-matter`, `remark-gfm`, `rehype-slug`, `rehype-autolink-headings`, `lucide-react`
3. Configure i18n (`lib/i18n.ts`, `middleware.ts` removed — App Router uses `[locale]` segment directly)
4. Build UI primitives: Button, Card, Badge
5. Build marketing components: Hero, FeatureGrid, ArchitectureOverview, Quickstart
6. Build docs components: Sidebar, TOC, LocaleSwitcher, ThemeToggle
7. Author landing page `app/[locale]/page.tsx`
8. Author 4 marketing pages: features/architecture/quickstart/(index under marketing group)
9. Author 5 docs pages with content imports from `content/docs/<locale>/`
10. Write 9 `messages/<locale>.json` (en + zh full, others minimal with `Not translated yet — see English version` for body text)
11. Translate 5 core docs to English (CN originals shipped as-is)
12. Write `.github/workflows/docs-site.yml`
13. Enable GitHub Pages in repo settings (manual step or via gh CLI if permitted)
14. Update root `README.md` to add a Docs badge linking to the deployed site
15. Push and verify build/deploy

## Acceptance criteria
- [ ] `cd website && npm ci && npm run build` produces `out/` with 9 locale subdirs
- [ ] GitHub Actions `docs-site` workflow runs green on push
- [ ] `https://agentcodinglab.github.io/aicodingagentteam/` shows English landing page
- [ ] Top-right language switcher cycles through all 9 locales
- [ ] Each locale `/<locale>/docs/requirements` renders the corresponding markdown
- [ ] Sidebar shows 5 doc pages, TOC anchors work
- [ ] Root README has a Docs badge linking to the deployed site

## Risks
- `npm install` may be slow on CI; mitigated by caching via `actions/setup-node` cache
- Static export + `[locale]` dynamic segment requires `generateStaticParams()` to enumerate all locales
- Mermaid blocks not rendered in stage 1; `docs/04-系统架构图.md` stays in repo, only later stages render it on the website
- Custom domain not in stage 1; site lives at `agentcodinglab.github.io/aicodingagentteam/` with `basePath` set accordingly