import Link from "next/link";
import { useTranslations } from "next-intl";
import { Github } from "lucide-react";

export function SiteFooter() {
  const t = useTranslations("footer");
  const tLinks = useTranslations("footer.links");
  return (
    <footer className="border-t border-slate-200 bg-slate-50/40 dark:border-slate-800 dark:bg-slate-900/40">
      <div className="container mx-auto max-w-6xl px-6 py-12">
        <div className="grid grid-cols-2 gap-8 md:grid-cols-4">
          <div className="col-span-2 md:col-span-1">
            <div className="flex items-center gap-2 text-sm font-semibold">
              <span className="inline-flex h-7 w-7 items-center justify-center rounded-md bg-brand-600 text-white">A</span>
              <span>AiCodingAgentTeam</span>
            </div>
            <p className="mt-3 max-w-xs text-sm leading-6 text-slate-600 dark:text-slate-400">
              {t("releaseNote")}
            </p>
          </div>
          <div>
            <h4 className="text-xs font-semibold uppercase tracking-wider text-slate-500 dark:text-slate-400">
              {t("product")}
            </h4>
            <ul className="mt-3 space-y-2 text-sm">
              <li><Link href="/features" className="text-slate-700 hover:text-brand-600 dark:text-slate-200 dark:hover:text-brand-400">{tLinks("features")}</Link></li>
              <li><Link href="/architecture" className="text-slate-700 hover:text-brand-600 dark:text-slate-200 dark:hover:text-brand-400">{tLinks("architecture")}</Link></li>
              <li><Link href="/quickstart" className="text-slate-700 hover:text-brand-600 dark:text-slate-200 dark:hover:text-brand-400">{tLinks("quickstart")}</Link></li>
            </ul>
          </div>
          <div>
            <h4 className="text-xs font-semibold uppercase tracking-wider text-slate-500 dark:text-slate-400">
              {t("resources")}
            </h4>
            <ul className="mt-3 space-y-2 text-sm">
              <li><Link href="/docs/requirements" className="text-slate-700 hover:text-brand-600 dark:text-slate-200 dark:hover:text-brand-400">{tLinks("requirements")}</Link></li>
              <li><Link href="/docs/implementation-plan" className="text-slate-700 hover:text-brand-600 dark:text-slate-200 dark:hover:text-brand-400">{tLinks("implementation")}</Link></li>
              <li><Link href="/docs/quality-constraints" className="text-slate-700 hover:text-brand-600 dark:text-slate-200 dark:hover:text-brand-400">{tLinks("quality")}</Link></li>
              <li><Link href="/docs/domain-model" className="text-slate-700 hover:text-brand-600 dark:text-slate-200 dark:hover:text-brand-400">{tLinks("domain")}</Link></li>
            </ul>
          </div>
          <div>
            <h4 className="text-xs font-semibold uppercase tracking-wider text-slate-500 dark:text-slate-400">
              {t("community")}
            </h4>
            <ul className="mt-3 space-y-2 text-sm">
              <li>
                <a href="https://github.com/agentcodinglab/aicodingagentteam" target="_blank" rel="noreferrer" className="inline-flex items-center gap-1.5 text-slate-700 hover:text-brand-600 dark:text-slate-200 dark:hover:text-brand-400">
                  <Github className="h-3.5 w-3.5" />{tLinks("github")}
                </a>
              </li>
              <li><a href="https://github.com/agentcodinglab/aicodingagentteam/releases" target="_blank" rel="noreferrer" className="text-slate-700 hover:text-brand-600 dark:text-slate-200 dark:hover:text-brand-400">{tLinks("releases")}</a></li>
              <li><a href="https://github.com/agentcodinglab/aicodingagentteam/issues" target="_blank" rel="noreferrer" className="text-slate-700 hover:text-brand-600 dark:text-slate-200 dark:hover:text-brand-400">{tLinks("issues")}</a></li>
              <li><a href="https://github.com/agentcodinglab/aicodingagentteam/discussions" target="_blank" rel="noreferrer" className="text-slate-700 hover:text-brand-600 dark:text-slate-200 dark:hover:text-brand-400">{tLinks("discussions")}</a></li>
            </ul>
          </div>
        </div>
        <div className="mt-10 flex flex-col items-center justify-between gap-2 border-t border-slate-200 pt-6 text-xs text-slate-500 dark:border-slate-800 dark:text-slate-400 sm:flex-row">
          <span>{t("copyright")}</span>
          <span>{t("releaseNote")}</span>
        </div>
      </div>
    </footer>
  );
}