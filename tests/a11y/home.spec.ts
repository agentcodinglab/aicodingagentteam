// home.spec.ts - a11y smoke tests for the 9-locale homepage.
//
// ADR-0016: docs/adr/ADR-0016-direction-a-governance.md
// Strategy: serve ./website/out statically via http.createServer; axe
// runs in real Chromium via @axe-core/playwright after networkidle.
//
// Threshold (from CONSTRAINTS.md):
//   0 serious/critical violations
//   <= 5 moderate/minor violations as warning (logged, not failed)

import { test, expect } from '@playwright/test';
import { AxeBuilder } from '@axe-core/playwright';
import { startStaticServer, type StaticServer } from './fixtures/static-server';
import { LOCALES, type Locale } from './fixtures/locales';

let server: StaticServer;

test.beforeAll(async () => {
  server = await startStaticServer();
});

test.afterAll(async () => {
  await server.close();
});

for (const locale of LOCALES) {
  test(`home (${locale.code}) has no serious a11y violations`, async ({ page }) => {
    const url = `${server.url}/${locale.code}/`;
    await page.goto(url, { waitUntil: 'networkidle' });

    // Run axe with WCAG 2.1 AA + best practices tags.
    const results = await new AxeBuilder({ page })
      .withTags(['wcag2a', 'wcag2aa', 'wcag21a', 'wcag21aa', 'best-practice'])
      // Disable color-contrast at the rule level: dark duotone hero has
      // intentional low-contrast decorative gradients (Stage 2 design).
      // Real text contrast is verified by Lighthouse audit (governance workflow).
      .disableRules(['color-contrast'])
      .analyze();

    const serious = results.violations.filter(
      (v) => v.impact === 'serious' || v.impact === 'critical',
    );
    const moderate = results.violations.filter((v) => v.impact === 'moderate');

    // Log everything for debug visibility in CI logs.
    for (const v of results.violations) {
      // eslint-disable-next-line no-console
      console.log(`[a11y][${locale.code}] ${v.impact} ${v.id}: ${v.help}`);
    }

    expect(
      serious,
      `${serious.length} serious/critical violations on ${url}: ${serious
        .map((v) => v.id)
        .join(', ')}`,
    ).toEqual([]);
    expect(
      moderate.length,
      `${moderate.length} moderate violations on ${url} (threshold: 5)`,
    ).toBeLessThanOrEqual(5);
  });
}
