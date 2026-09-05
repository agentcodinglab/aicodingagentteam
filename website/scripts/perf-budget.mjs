#!/usr/bin/env node
// perf-budget.mjs - measure Next.js App Router initial JS bundle size.
// Reads .next/build-manifest.json rootMainFiles, sums raw + gzipped,
// compares against budget, exits non-zero on failure.
//
// ADR-0016: docs/adr/ADR-0016-direction-a-governance.md
//
// Usage:
//   node scripts/perf-budget.mjs [--budget-kb=200] [--manifest=.next/build-manifest.json] [--out=.perf-budget.json] [--verbose]

import { readFileSync, writeFileSync, existsSync } from 'node:fs';
import { gzipSync } from 'node:zlib';
import { resolve, dirname, basename } from 'node:path';
import { fileURLToPath } from 'node:url';

const __dirname = dirname(fileURLToPath(import.meta.url));

function arg(name, fallback) {
  const a = process.argv.find((s) => s.startsWith('--' + name + '='));
  return a ? a.slice(name.length + 3) : fallback;
}

const budgetKb = Number(arg('budget-kb', '200'));
const manifestRel = arg('manifest', '.next/build-manifest.json');
const outRel = arg('out', '.perf-budget.json');
const verbose = process.argv.includes('--verbose');

const root = resolve(__dirname, '..');
const manifestPath = resolve(root, manifestRel);

if (!existsSync(manifestPath)) {
  console.error('[perf-budget] build-manifest not found: ' + manifestPath);
  console.error('[perf-budget] run `next build` first.');
  process.exit(2);
}

const manifest = JSON.parse(readFileSync(manifestPath, 'utf8'));
const rootFiles = manifest.rootMainFiles || [];
if (rootFiles.length === 0) {
  console.error('[perf-budget] rootMainFiles empty -- check Next.js version.');
  process.exit(2);
}

const chunksDir = resolve(root, dirname(manifestRel), 'static', 'chunks');
const perFile = rootFiles.map((rel) => {
  const abs = resolve(chunksDir, basename(rel));
  if (!existsSync(abs)) return { file: rel, missing: true, raw: 0, gz: 0 };
  const buf = readFileSync(abs);
  const gz = gzipSync(buf, { level: 9 });
  return { file: rel, raw: buf.length, gz: gz.length };
});

const total = perFile.reduce(
  (acc, c) => ({ raw: acc.raw + c.raw, gz: acc.gz + c.gz, missing: acc.missing + (c.missing ? 1 : 0) }),
  { raw: 0, gz: 0, missing: 0 },
);

const budgetBytes = budgetKb * 1024;
const passed = total.missing === 0 && total.gz <= budgetBytes;

const report = {
  generatedAt: new Date().toISOString(),
  budgetKb,
  budgetBytes,
  totalRaw: total.raw,
  totalGz: total.gz,
  passed,
  chunks: perFile,
};

const outPath = resolve(root, outRel);
writeFileSync(outPath, JSON.stringify(report, null, 2));

const human = (n) => (n / 1024).toFixed(2) + ' KB';
console.log('[perf-budget] initial JS (rootMainFiles: ' + rootFiles.length + ' chunks)');
for (const c of perFile) {
  if (c.missing) console.log('  ! missing ' + c.file);
  else if (verbose) console.log('  - ' + c.file + ': raw=' + human(c.raw) + ' gz=' + human(c.gz));
}
console.log('[perf-budget] total  raw=' + human(total.raw) + '  gz=' + human(total.gz));
console.log('[perf-budget] budget gz<=' + budgetKb + ' KB  ->  ' + (passed ? 'PASS' : 'FAIL'));
console.log('[perf-budget] report: ' + outPath);

process.exit(passed ? 0 : 1);
