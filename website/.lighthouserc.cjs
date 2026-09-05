// .lighthouserc.cjs - Lighthouse CI configuration
//
// ADR-0016: docs/adr/ADR-0016-direction-a-governance.md
// CI: .github/workflows/governance.yml
//
// Strategy: treosh/lighthouse-ci-action runs lhci collect, which spawns
// Chrome to crawl the URLs. lhci supports staticDistDir, but treosh
// action wraps it with its own static server, so URLs are just paths.

'use strict';

module.exports = {
  ci: {
    collect: {
            // Test the 4 highest-priority locales + the root redirect target.
      // URLs use root-relative paths because lhci+staticDistDir auto-
      // prefixes the dynamic port assigned by treosh's internal server.
      url: [
        'http://localhost:3000/',
        'http://localhost:3000/en/',
        'http://localhost:3000/zh/',
        'http://localhost:3000/ja/',
        'http://localhost:3000/ko/',
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
        // or on localhost; do not penalize.
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





