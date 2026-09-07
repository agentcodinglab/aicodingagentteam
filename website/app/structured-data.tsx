// JSON-LD structured data for SEO.
// Renders SoftwareApplication + BreadcrumbList schema as a <script type="application/ld+json"> tag.
// ADR-0022 P8.3: structured data for Lighthouse SEO score.

interface SoftwareApplication {
  "@context": string;
  "@type": string;
  name: string;
  description: string;
  applicationCategory: string;
  operatingSystem: string;
  url: string;
  author: { "@type": string; name: string };
  keywords: string;
  programmingLanguage: string;
}

interface BreadcrumbItem {
  "@type": string;
  position: number;
  name: string;
  item: string;
}

interface BreadcrumbList {
  "@context": string;
  "@type": string;
  itemListElement: BreadcrumbItem[];
}

const SITE_URL = "https://agentcodinglab.github.io/aicodingagentteam";

export function StructuredData() {
  const app: SoftwareApplication = {
    "@context": "https://schema.org",
    "@type": "SoftwareApplication",
    name: "AiCodingAgentTeam",
    description:
      "A Go-based orchestrator that dispatches Codex, OpenCode, Claude-Code and DeepSeek-DSH — without holding any API key — to build, review, and ship code under a deterministic quality gate.",
    applicationCategory: "DeveloperApplication",
    operatingSystem: "Linux, macOS, Windows",
    url: SITE_URL,
    author: { "@type": "Organization", name: "AiCodingAgentTeam contributors" },
    keywords:
      "AI coding, Codex, OpenCode, Claude-Code, DeepSeek-DSH, A2A protocol, MCP, ACP, Go, orchestrator, multi-agent",
    programmingLanguage: "Go",
  };

  const breadcrumbs: BreadcrumbList = {
    "@context": "https://schema.org",
    "@type": "BreadcrumbList",
    itemListElement: [
      {
        "@type": "ListItem",
        position: 1,
        name: "AiCodingAgentTeam",
        item: SITE_URL,
      },
    ],
  };

  return (
    <>
      <script
        type="application/ld+json"
        dangerouslySetInnerHTML={{ __html: JSON.stringify(app) }}
      />
      <script
        type="application/ld+json"
        dangerouslySetInnerHTML={{ __html: JSON.stringify(breadcrumbs) }}
      />
    </>
  );
}
