import Link from "next/link";
import { useTranslations } from "next-intl";
import { Github } from "lucide-react";

export function SiteFooter() {
  const t = useTranslations("footer");
  const tLinks = useTranslations("footer.links");
  return (
    <footer className="border-t border-cyan-line/60 bg-bg-2/40">
      <div className="container mx-auto max-w-6xl px-6 py-14">
        <div className="grid grid-cols-2 gap-8 md:grid-cols-4">
          <div className="col-span-2 md:col-span-1">
            <div className="flex items-center gap-2 text-sm font-semibold text-ink">
              <span className="inline-flex h-7 w-7 items-center justify-center rounded-md bg-duo font-display text-bg">
                A
              </span>
              <span className="font-display">AiCodingAgentTeam</span>
            </div>
            <p className="mt-3 max-w-xs text-sm leading-6 text-ink-muted">
              {t("releaseNote")}
            </p>
          </div>
          <div>
            <h4 className="font-mono text-xs font-semibold uppercase tracking-[0.18em] text-ink-muted2">
              {t("product")}
            </h4>
            <ul className="mt-3 space-y-2 text-sm">
              <li>
                <Link href="/features" className="text-ink-muted transition-colors hover:text-cyan">
                  {tLinks("features")}
                </Link>
              </li>
              <li>
                <Link href="/architecture" className="text-ink-muted transition-colors hover:text-cyan">
                  {tLinks("architecture")}
                </Link>
              </li>
              <li>
                <Link href="/quickstart" className="text-ink-muted transition-colors hover:text-cyan">
                  {tLinks("quickstart")}
                </Link>
              </li>
            </ul>
          </div>
          <div>
            <h4 className="font-mono text-xs font-semibold uppercase tracking-[0.18em] text-ink-muted2">
              {t("resources")}
            </h4>
            <ul className="mt-3 space-y-2 text-sm">
              <li>
                <Link href="/docs/requirements" className="text-ink-muted transition-colors hover:text-cyan">
                  {tLinks("requirements")}
                </Link>
              </li>
              <li>
                <Link href="/docs/implementation-plan" className="text-ink-muted transition-colors hover:text-cyan">
                  {tLinks("implementation")}
                </Link>
              </li>
              <li>
                <Link href="/docs/quality-constraints" className="text-ink-muted transition-colors hover:text-cyan">
                  {tLinks("quality")}
                </Link>
              </li>
              <li>
                <Link href="/docs/domain-model" className="text-ink-muted transition-colors hover:text-cyan">
                  {tLinks("domain")}
                </Link>
              </li>
            </ul>
          </div>
          <div>
            <h4 className="font-mono text-xs font-semibold uppercase tracking-[0.18em] text-ink-muted2">
              {t("community")}
            </h4>
            <ul className="mt-3 space-y-2 text-sm">
              <li>
                <a
                  href="https://github.com/agentcodinglab/aicodingagentteam"
                  target="_blank"
                  rel="noreferrer"
                  className="inline-flex items-center gap-1.5 text-ink-muted transition-colors hover:text-cyan"
                >
                  <Github className="h-3.5 w-3.5" />
                  {tLinks("github")}
                </a>
              </li>
              <li>
                <a
                  href="https://github.com/agentcodinglab/aicodingagentteam/releases"
                  target="_blank"
                  rel="noreferrer"
                  className="text-ink-muted transition-colors hover:text-cyan"
                >
                  {tLinks("releases")}
                </a>
              </li>
              <li>
                <a
                  href="https://github.com/agentcodinglab/aicodingagentteam/issues"
                  target="_blank"
                  rel="noreferrer"
                  className="text-ink-muted transition-colors hover:text-cyan"
                >
                  {tLinks("issues")}
                </a>
              </li>
              <li>
                <a
                  href="https://github.com/agentcodinglab/aicodingagentteam/discussions"
                  target="_blank"
                  rel="noreferrer"
                  className="text-ink-muted transition-colors hover:text-cyan"
                >
                  {tLinks("discussions")}
                </a>
              </li>
            </ul>
          </div>
        </div>
        <div className="mt-10 flex flex-col items-center justify-between gap-2 border-t border-cyan-line/60 pt-6 text-xs text-ink-muted2 sm:flex-row">
          <span>{t("copyright")}</span>
          <span>{t("releaseNote")}</span>
        </div>
      </div>
    </footer>
  );
}