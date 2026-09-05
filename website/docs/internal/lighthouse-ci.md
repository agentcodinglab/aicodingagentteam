# Lighthouse CI

Local CI audit of the static docs site. ADR-0016.

## What it does

Runs Lighthouse against 5 URLs (root + en/zh/ja/ko) of the `next export`
output in `out/`, asserts four category scores, and uploads a temporary
public report.

| Category        | Min score | Failure mode |
|-----------------|-----------|--------------|
| performance     | 0.90      | error (CI red) |
| accessibility   | 0.95      | error (CI red) |
| best-practices  | 0.95      | error (CI red) |
| seo             | 0.95      | error (CI red) |

## How it runs

- `.github/workflows/governance.yml` invokes `@lhci/cli autorun`
  after `next build`, using `.lighthouserc.cjs`.
- No server is started; `staticDistDir: ./out` is read directly.
- The action `treosh/lighthouse-ci-action@v9` handles install + run + report.

## Local debug

```bash
npm i -D @lhci/cli
npm run build
npx lhci autorun --config=./.lighthouserc.cjs
```

## Audits skipped and why

- `is-on-https`, `redirects-http`, `uses-http2`: localhost serves http.
- `service-worker`, `installable-manifest`: we ship a static site, no PWA.

## Raising the bar

Adjust `assert.assertions` in `.lighthouserc.cjs` after a baseline run shows
comfortable headroom. The governance workflow records history in GitHub
Actions artifacts (`.lighthouseci/` upload target).
