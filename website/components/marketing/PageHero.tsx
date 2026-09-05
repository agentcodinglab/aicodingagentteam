import { LocaleLink } from "../ui/LocaleLink";
import { ChevronLeft } from "lucide-react";
import { useTranslations } from "next-intl";
import { Reveal } from "../ui/Reveal";

type Props = {
  eyebrow: string;
  title: string;
  subtitle?: string;
};

export function PageHero({ eyebrow, title, subtitle }: Props) {
  const t = useTranslations("common");
  return (
    <section className="relative isolate overflow-hidden border-b border-cyan-line/60 py-20">
      <div
        className="absolute inset-0 -z-10 bg-grid-dark bg-grid opacity-40"
        style={{
          maskImage:
            "radial-gradient(ellipse at top, black 25%, transparent 70%)",
          WebkitMaskImage:
            "radial-gradient(ellipse at top, black 25%, transparent 70%)",
        }}
      />
      <div className="absolute -top-32 left-1/2 -z-10 h-[320px] w-[600px] -translate-x-1/2 rounded-full bg-duo opacity-20 blur-3xl" />
      <div className="container mx-auto max-w-5xl px-6">
        <Reveal>
          <LocaleLink
            href="/"
            className="mb-6 inline-flex items-center gap-1 font-mono text-xs uppercase tracking-[0.18em] text-ink-muted2 transition-colors hover:text-cyan"
          >
            <ChevronLeft className="h-3.5 w-3.5" />
            {t("backToHome")}
          </LocaleLink>
        </Reveal>
        <Reveal delay={80}>
          <p className="mb-3 font-mono text-xs font-semibold uppercase tracking-[0.22em] text-cyan">
            {eyebrow}
          </p>
        </Reveal>
        <Reveal delay={160}>
          <h1 className="font-display text-4xl font-bold leading-[1.05] tracking-tight text-ink sm:text-5xl">
            {title}
          </h1>
        </Reveal>
        {subtitle ? (
          <Reveal delay={240}>
            <p className="mt-5 max-w-2xl text-lg leading-8 text-ink-muted">
              {subtitle}
            </p>
          </Reveal>
        ) : null}
      </div>
    </section>
  );
}