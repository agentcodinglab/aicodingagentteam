import { useTranslations } from "next-intl";
import { Button } from "../ui/Button";
import { ArrowRight } from "lucide-react";

export function FinalCTA() {
  const t = useTranslations("home");
  return (
    <section className="relative isolate overflow-hidden border-y border-cyan-line py-24">
      <div className="absolute inset-0 -z-10 bg-[#04060a]" />
      <div className="absolute inset-0 -z-10 bg-grid-dark bg-grid opacity-40" />
      <div className="absolute -top-32 left-1/2 -z-10 h-[400px] w-[680px] -translate-x-1/2 rounded-full bg-duo opacity-30 blur-3xl" />
      <div className="absolute -bottom-32 left-1/4 -z-10 h-[280px] w-[420px] rounded-full bg-magenta/30 blur-3xl" />

      <div className="container mx-auto max-w-4xl px-6 text-center">
        <p className="mb-3 font-mono text-xs uppercase tracking-[0.22em] text-cyan">
          ./aicat --ready
        </p>
        <h2 className="font-display text-3xl font-bold tracking-tight text-ink sm:text-5xl">
          <span className="text-ink">{t("ctaTitle")}</span>
        </h2>
        <p className="mx-auto mt-5 max-w-2xl text-base leading-8 text-ink-muted sm:text-lg">
          {t("ctaSubtitle")}
        </p>
        <div className="mt-9 flex flex-wrap items-center justify-center gap-4">
          <Button href="/quickstart" variant="primary" className="px-6 py-3 text-base">
            {t("ctaButton")} <ArrowRight className="ml-2 h-4 w-4" />
          </Button>
          <Button
            href="https://github.com/agentcodinglab/aicodingagentteam"
            external
            variant="ghost"
            className="px-6 py-3 text-base text-ink-muted hover:text-cyan"
          >
            GitHub \u2197
          </Button>
        </div>
      </div>
    </section>
  );
}