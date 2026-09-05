# website

Next.js 14 App Router static site for AiCodingAgentTeam. Renders the marketing landing page and a docs site, with 9-locale support via `next-intl` URL-segment routing.

## Quick reference

```bash
npm ci
npm run dev       # local dev server
npm run build     # static export to ./out
npm run lint
npm run typecheck
```

## Layout

- `app/[locale]/` — locale-scoped routes; `generateStaticParams()` enumerates all 9 locales at build time
- `components/marketing/` — Hero, FeatureGrid, ArchitectureOverview, Quickstart
- `components/docs/` — Sidebar, TOC, LocaleSwitcher, ThemeToggle, SiteHeader, SiteFooter, DocContent
- `content/docs/<locale>/` — markdown sources for the 5 core doc pages
- `messages/<locale>.json` — next-intl UI strings; en + zh fully populated, the other 7 locales have full coverage too (translated from en)
- `lib/i18n.ts` — locale list and `getRequestConfig`
- `middleware.ts` — locale-prefix routing

## Adding a doc

1. Drop a markdown file in `content/docs/en/<slug>.md` (and `zh/<slug>.md` for Chinese)
2. Register the slug in `lib/nav.ts` (`docsNav` array)
3. Add a title key in `messages/en.json` and `messages/zh.json` under `docs.nav`
4. Add the page component at `app/[locale]/docs/<slug>/page.tsx` (use one of the existing 5 as a template)

## Deployment

Pushed on changes to `website/**` or the workflow file. Build → upload artifact → deploy via official `actions/deploy-pages@v4`.

Site lives at `https://agentcodinglab.github.io/aicodingagentteam/` (with `trailingSlash: true` so each locale is at `/en/`, `/zh/`, …).