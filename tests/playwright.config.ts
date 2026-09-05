// playwright.config.ts - governance a11y test config.
//
// Strategy:
// - Chromium only (axe-core works headless without browser matrix overhead).
// - Single worker (deterministic port binding; parallel crawls would need
//   multiple static-server instances).
// - Uses the server fixture defined per-spec.

import { defineConfig } from '@playwright/test';

export default defineConfig({
  testDir: './a11y',
  timeout: 30_000,
  retries: 0,
  workers: 1,
  reporter: process.env.CI ? [['github'], ['list']] : 'list',
  use: {
    baseURL: 'http://127.0.0.1:0', // overridden per-spec by the fixture
    headless: true,
  },
});
