# governance tests

Playwright + axe-core a11y smoke tests. ADR-0016.

## What it covers

- `a11y/home.spec.ts`: 9-locale homepage (9 URLs)
- `a11y/docs.spec.ts`: 3 representative docs pages per locale (27 URLs)

Threshold: 0 serious/critical violations; <= 5 moderate per page (warning).

## Local run

```bash
cd tests
npm install
npm run test:a11y:install   # one-time: downloads Chromium
npm run test:a11y
```

Requires `website/out/` to exist (run `npm run build` in `website/` first).

## CI

`.github/workflows/governance.yml` runs these on every push to `main` that
touches `website/**` or `tests/**`.
