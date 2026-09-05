// Generate sitemap.xml for the static site.
// Run before `next build` so the file is copied into out/.
import fs from "node:fs";
import path from "node:path";

const BASE = "https://agentcodinglab.github.io/aicodingagentteam";
const LOCALES = ["en", "zh", "ja", "ko", "fr", "de", "ru", "es", "it"];
const STATIC_PAGES = ["", "features/", "architecture/", "quickstart/"];
const DOC_PAGES = [
  "docs/requirements/",
  "docs/architecture/",
  "docs/implementation-plan/",
  "docs/quality-constraints/",
  "docs/domain-model/",
  "docs/changelog/",
];

const urls = [];
urls.push({ loc: BASE + "/", priority: "1.0", changefreq: "weekly" });
for (const loc of LOCALES) {
  for (const page of STATIC_PAGES) {
    urls.push({
      loc: `${BASE}/${loc}/${page}`,
      priority: page === "" ? "1.0" : "0.8",
      changefreq: "weekly",
    });
  }
  for (const page of DOC_PAGES) {
    urls.push({
      loc: `${BASE}/${loc}/${page}`,
      priority: "0.6",
      changefreq: "monthly",
    });
  }
}

const today = new Date().toISOString().split("T")[0];
const xml = [
  '<?xml version="1.0" encoding="UTF-8"?>',
  '<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">',
  ...urls.map(
    (u) =>
      `  <url>\n    <loc>${u.loc}</loc>\n    <lastmod>${today}</lastmod>\n    <changefreq>${u.changefreq}</changefreq>\n    <priority>${u.priority}</priority>\n  </url>`,
  ),
  "</urlset>",
  "",
].join("\n");

const outDir = path.join(process.cwd(), "public");
fs.mkdirSync(outDir, { recursive: true });
const outPath = path.join(outDir, "sitemap.xml");
fs.writeFileSync(outPath, xml, "utf8");
console.log(`sitemap.xml: ${urls.length} URLs, ${fs.statSync(outPath).size} bytes`);