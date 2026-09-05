import { useTranslations } from "next-intl";
import { Button } from "../ui/Button";
import { ArrowRight } from "lucide-react";

export function FinalCTA() {
  const t = useTranslations("home");
  return (
    <section className="relative isolate overflow-hidden border-y border-slate-200 bg-slate-950 py-20 dark:border-slate-800">
      <div className="absolute inset-0 -z-10 bg-[radial-gradient(ellipse_at_center,rgba(45,138,255,0.30),transparent_65%)]" />
      <div className="absolute inset-0 -z-10 opacity-30 [mask-image:radial-gradient(ellipse_at_center,black,transparent_70%)]">
        <svg className="h-full w-full" aria-hidden="true">
          <defs>
            <pattern id="cta-grid" width="32" height="32" patternUnits="userSpaceOnUse">
              <path d="M0 32V0h32" fill="none" stroke="rgba(148,163,184,0.18)" />
            </pattern>
          </defs>
          <rect width="100%" height="100%" fill="url(#cta-grid)" />
        </svg>
      </div>
      <div className="container mx-auto max-w-4xl px-6 text-center">
        <h2 className="text-3xl font-bold tracking-tight text-white sm:text-4xl">{t("ctaTitle")}</h2>
        <p className="mx-auto mt-4 max-w-2xl text-lg leading-8 text-slate-300">{t("ctaSubtitle")}</p>
        <div className="mt-8 flex flex-wrap items-center justify-center gap-3">
          <Button href="/quickstart" variant="primary" className="px-6 py-2.5 text-base">
            {t("ctaButton")} <ArrowRight className="ml-2 h-4 w-4" />
          </Button>
          <Button
            href="https://github.com/agentcodinglab/aicodingagentteam"
            external
            variant="ghost"
            className="px-6 py-2.5 text-base text-slate-200 hover:bg-slate-800"
          >
            GitHub ↗
          </Button>
        </div>
      </div>
    </section>
  );
}