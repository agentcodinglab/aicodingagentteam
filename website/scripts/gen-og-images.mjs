// Generate static OG (Open Graph) images for each locale homepage.
// Produces SVG files (no external dependencies required).
// Usage: node scripts/gen-og-images.mjs
//
// ADR-0022 P8.2: OG image generation for social sharing.
import fs from "node:fs";
import path from "node:path";

const LOCALES = {
  en: { title: "AiCodingAgentTeam", subtitle: "Coordinate a 9-role software team with AI coding CLIs" },
  zh: { title: "AiCodingAgentTeam", subtitle: "协调 9 角色 AI 编码团队" },
  ja: { title: "AiCodingAgentTeam", subtitle: "AI コーディング CLI で 9 ロールチームを編成" },
  ko: { title: "AiCodingAgentTeam", subtitle: "AI 코딩 CLI로 9역할 팀 조율" },
  fr: { title: "AiCodingAgentTeam", subtitle: "Coordonnez une equipe de 9 roles avec des CLI de codage IA" },
  de: { title: "AiCodingAgentTeam", subtitle: "Steuern Sie ein 9-Rollen-Team mit KI-Coding-CLIs" },
  ru: { title: "AiCodingAgentTeam", subtitle: "Координируйте команду из 9 ролей с ИИ-кодингом" },
  es: { title: "AiCodingAgentTeam", subtitle: "Coordina un equipo de 9 roles con CLIs de codificacion IA" },
  it: { title: "AiCodingAgentTeam", subtitle: "Coordina un team di 9 ruoli con CLI di coding IA" },
};

const W = 1200;
const H = 630;

function escapeXml(s) {
  return s.replace(/&/g, "&amp;").replace(/</g, "&lt;").replace(/>/g, "&gt;").replace(/"/g, "&quot;").replace(/'/g, "&apos;");
}

function svg(locale, { title, subtitle }) {
  return `<svg xmlns="http://www.w3.org/2000/svg" width="${W}" height="${H}" viewBox="0 0 ${W} ${H}">
  <defs>
    <linearGradient id="bg" x1="0" y1="0" x2="1" y2="1">
      <stop offset="0%" stop-color="#0f172a"/>
      <stop offset="100%" stop-color="#1e293b"/>
    </linearGradient>
  </defs>
  <rect width="${W}" height="${H}" fill="url(#bg)"/>
  <rect x="0" y="0" width="8" height="${H}" fill="#3b82f6"/>
  <text x="80" y="280" font-family="'Space Grotesk', sans-serif" font-size="64" font-weight="700" fill="#f8fafc">${escapeXml(title)}</text>
  <text x="80" y="360" font-family="'Manrope', sans-serif" font-size="32" font-weight="400" fill="#94a3b8">${escapeXml(subtitle)}</text>
  <text x="80" y="560" font-family="'JetBrains Mono', monospace" font-size="24" fill="#3b82f6">Go + A2A + MCP + ACP</text>
  <text x="${W - 280}" y="560" font-family="'Manrope', sans-serif" font-size="24" fill="#64748b">${locale.toUpperCase()}</text>
</svg>
`;
}

const outDir = path.join(process.cwd(), "public", "og");
fs.mkdirSync(outDir, { recursive: true });

let count = 0;
for (const [locale, info] of Object.entries(LOCALES)) {
  const outPath = path.join(outDir, `og-${locale}.svg`);
  fs.writeFileSync(outPath, svg(locale, info), "utf8");
  count++;
}

// Also write default og.png as SVG (served as og.svg in layout)
fs.writeFileSync(path.join(outDir, "og-default.svg"), svg("en", LOCALES.en), "utf8");

console.log(`OG images: ${count} locale SVGs + 1 default, written to public/og/`);
