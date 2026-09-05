import Link from "next/link";
import { useTranslations } from "next-intl";
import { LocaleSwitcher } from "./LocaleSwitcher";
import { ThemeToggle } from "./ThemeToggle";
import { Github } from "lucide-react";

export function SiteHeader() {
  const t = useTranslations("nav");
  return (
    <header className="sticky top-0 z-40 w-full border-b border-cyan-line/60 bg-bg/80 backdrop-blur supports-[backdrop-filter]:bg-bg/60">
      <div className="container mx-auto flex h-16 max-w-6xl items-center justify-between px-6">
        <Link href="/" className="flex items-center gap-2 text-sm font-semibold text-ink">
          <span className="inline-flex h-7 w-7 items-center justify-center rounded-md bg-duo font-display text-bg">
            A
          </span>
          <span className="font-display">AiCodingAgentTeam</span>
        </Link>
        <nav className="hidden items-center gap-6 text-sm md:flex">
          <Link href="/features" className="text-ink-muted transition-colors hover:text-cyan">
            {t("features")}
          </Link>
          <Link href="/architecture" className="text-ink-muted transition-colors hover:text-cyan">
            {t("architecture")}
          </Link>
          <Link href="/quickstart" className="text-ink-muted transition-colors hover:text-cyan">
            {t("quickstart")}
          </Link>
          <Link href="/docs/requirements" className="text-ink-muted transition-colors hover:text-cyan">
            {t("docs")}
          </Link>
        </nav>
        <div className="flex items-center gap-1">
          <Link
            href="https://github.com/agentcodinglab/aicodingagentteam"
            target="_blank"
            rel="noreferrer"
            className="inline-flex h-9 w-9 items-center justify-center rounded-md text-ink-muted transition-colors hover:bg-bg-2 hover:text-cyan"
            aria-label="GitHub"
          >
            <Github className="h-4 w-4" />
          </Link>
          <ThemeToggle />
          <LocaleSwitcher />
        </div>
      </div>
    </header>
  );
}