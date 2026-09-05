// docs.spec.ts - a11y tests for the documentation pages.
//
// Strategy: spot-check 3 representative docs pages per locale
// (index, architecture, changelog) using axe-core.

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

const DOCS_PATHS = ['docs', 'docs/architecture', 'docs/changelog'] as const;

for (const locale of LOCALES) {
  for (const path of DOCS_PATHS) {
    test(`docs (${locale.code}/${path}) has no serious a11y violations`, async ({ page }) => {
      const url = `${server.url}/${locale.code}/${path}/`;
      await page.goto(url, { waitUntil: 'networkidle' });

      const results = await new AxeBuilder({ page })
        .withTags(['wcag2a', 'wcag2aa', 'wcag21a', 'wcag21aa'])
        .disableRules(['color-contrast'])
        .analyze();

      const serious = results.violations.filter(
        (v) => v.impact === 'serious' || v.impact === 'critical',
      );

      for (const v of results.violations) {
        // eslint-disable-next-line no-console
        console.log(`[a11y][${locale.code}/${path}] ${v.impact} ${v.id}: ${v.help}`);
      }

      expect(
        serious,
        `${serious.length} serious/critical violations on ${url}: ${serious
          .map((v) => v.id)
          .join(', ')}`,
      ).toEqual([]);
    });
  }
}
