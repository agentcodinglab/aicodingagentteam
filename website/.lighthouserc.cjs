// .lighthouserc.cjs - Lighthouse CI configuration
//
// ADR-0016: docs/adr/ADR-0016-direction-a-governance.md
// CI: .github/workflows/governance.yml
//
// Strategy: run against the exported static site (next build -> ./out).
// Avoids spinning up a server and removes network noise.

'use strict';

module.exports = {
  ci: {
    collect: {
      // Static export directory produced by `next build`.
      staticDistDir: './out',
      // Test the 4 highest-priority locales + the root redirect target.
      url: [
        'http://localhost/',
        'http://localhost/en/',
        'http://localhost/zh/',
        'http://localhost/ja/',
        'http://localhost/ko/',
      ],
      numberOfRuns: 1,
      settings: {
        // Desktop preset for documentation site (most devs read on desktop).
        preset: 'desktop',
        // Use provided throttling; deterministic across CI runners.
        throttlingMethod: 'provided',
        // Only collect the categories we care about.
        onlyCategories: ['performance', 'accessibility', 'best-practices', 'seo'],
        // Audits that legitimately fail on a static export (PWA / service-worker)
        // or on localhost; don't penalize.
        skipAudits: [
          'is-on-https',
          'redirects-http',
          'uses-http2',
          'service-worker',
          'installable-manifest',
        ],
      },
    },
    assert: {
      // ADR-0016 thresholds.
      // Tighten after first baseline run if scores comfortably exceed minimum.
      assertions: {
        'categories:performance':    ['error', { minScore: 0.90 }],
        'categories:accessibility':  ['error', { minScore: 0.95 }],
        'categories:best-practices': ['error', { minScore: 0.95 }],
        'categories:seo':            ['error', { minScore: 0.95 }],
      },
    },
    upload: {
      target: 'temporary-public-storage',
      reportFilenamePattern: '%%PATHNAME%%-%%DATETIME%%-report',
    },
  },
};
