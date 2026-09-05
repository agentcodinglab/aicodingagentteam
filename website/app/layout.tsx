import type { Metadata } from "next";
import { ReactNode } from "react";
import "@fontsource-variable/manrope/wght.css";
import "@fontsource-variable/space-grotesk/wght.css";
import "@fontsource-variable/jetbrains-mono/wght.css";
import "./globals.css";

const SITE_URL = "https://agentcodinglab.github.io/aicodingagentteam";

export const metadata: Metadata = {
  metadataBase: new URL(SITE_URL),
  title: {
    default: "AiCodingAgentTeam \u2014 Coordinate a 9-role software team with AI coding CLIs",
    template: "%s \u00b7 AiCodingAgentTeam",
  },
  description:
    "AiCodingAgentTeam is a Go-based orchestrator that dispatches Codex, OpenCode, Claude-Code and DeepSeek-DSH \u2014 without holding any API key \u2014 to build, review, and ship code under a deterministic quality gate.",
  applicationName: "AiCodingAgentTeam",
  keywords: [
    "AiCodingAgentTeam",
    "AI coding",
    "Codex",
    "OpenCode",
    "Claude-Code",
    "DeepSeek-DSH",
    "A2A protocol",
    "MCP",
    "ACP",
    "Go",
    "orchestrator",
    "multi-agent",
  ],
  authors: [{ name: "AiCodingAgentTeam contributors" }],
  icons: {
    icon: "/favicon.svg",
  },
  openGraph: {
    title: "AiCodingAgentTeam \u2014 Coordinate a 9-role software team with AI coding CLIs",
    description:
      "A Go-based orchestrator that dispatches Codex, OpenCode, Claude-Code and DeepSeek-DSH \u2014 without holding any API key \u2014 to build, review, and ship code under a deterministic quality gate.",
    url: SITE_URL,
    siteName: "AiCodingAgentTeam",
    type: "website",
    locale: "en_US",
    images: [
      {
        url: "/og.png",
        width: 1200,
        height: 630,
        alt: "AiCodingAgentTeam \u2014 Coordinate a 9-role software team",
      },
    ],
  },
  twitter: {
    card: "summary_large_image",
    title: "AiCodingAgentTeam",
    description:
      "A Go-based orchestrator that dispatches Codex, OpenCode, Claude-Code and DeepSeek-DSH \u2014 without holding any API key.",
    images: "/og.png",
  },
};

export default function RootLayout({ children }: { children: ReactNode }) {
  return (
    <html lang="en" className="dark" suppressHydrationWarning>
      <body className="dark min-h-screen">{children}</body>
    </html>
  );
}