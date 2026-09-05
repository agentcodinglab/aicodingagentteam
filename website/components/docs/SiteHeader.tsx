import { LocaleLink } from "../ui/LocaleLink";
import { useTranslations } from "next-intl";
import { LocaleSwitcher } from "./LocaleSwitcher";
import { ThemeToggle } from "./ThemeToggle";
import { Github } from "lucide-react";

export function SiteHeader() {
  const t = useTranslations("nav");
  return (
    <header className="sticky top-0 z-40 w-full border-b border-cyan-line/60 bg-bg/80 backdrop-blur supports-[backdrop-filter]:bg-bg/60">
      <div className="container mx-auto flex h-16 max-w-6xl items-center justify-between px-6">
        <LocaleLink href="/" className="flex items-center gap-2 text-sm font-semibold text-ink">
          <span className="inline-flex h-7 w-7 items-center justify-center rounded-md bg-duo font-display text-bg">
            A
          </span>
          <span className="font-display">AiCodingAgentTeam</span>
        </LocaleLink>
        <nav className="hidden items-center gap-6 text-sm md:flex">
          <LocaleLink href="/features" className="text-ink-muted transition-colors hover:text-cyan">
            {t("features")}
          </LocaleLink>
          <LocaleLink href="/architecture" className="text-ink-muted transition-colors hover:text-cyan">
            {t("architecture")}
          </LocaleLink>
          <LocaleLink href="/quickstart" className="text-ink-muted transition-colors hover:text-cyan">
            {t("quickstart")}
          </LocaleLink>
          <LocaleLink href="/docs/requirements" className="text-ink-muted transition-colors hover:text-cyan">
            {t("docs")}
          </LocaleLink>
        </nav>
        <div className="flex items-center gap-1">
          <a
            href="https://github.com/agentcodinglab/aicodingagentteam"
            target="_blank"
            rel="noreferrer"
            className="inline-flex h-9 w-9 items-center justify-center rounded-md text-ink-muted transition-colors hover:bg-bg-2 hover:text-cyan"
            aria-label="GitHub"
          >
            <Github className="h-4 w-4" />
          </a>
          <ThemeToggle />
          <LocaleSwitcher />
        </div>
      </div>
    </header>
  );
}