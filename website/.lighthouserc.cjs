// .lighthouserc.cjs - Lighthouse CI configuration
//
// ADR-0016: docs/adr/ADR-0016-direction-a-governance.md
// CI: .github/workflows/governance.yml
//
// Strategy: governance.yml pre-starts a python http server on port
// 3000 serving ./out. We hit URLs relative to that server. We skip
// the root URL because it 302-redirects to a basePath-prefixed URL
// that python http.server cannot resolve.

'use strict';

module.exports = {
  ci: {
    collect: {
      url: [
        'http://localhost:3000/en/',
        'http://localhost:3000/zh/',
        'http://localhost:3000/ja/',
        'http://localhost:3000/ko/',
      ],
      numberOfRuns: 1,
      settings: {
        preset: 'desktop',
        throttlingMethod: 'provided',
        onlyCategories: ['performance', 'accessibility', 'best-practices', 'seo'],
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
