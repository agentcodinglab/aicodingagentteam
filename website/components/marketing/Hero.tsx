import { useTranslations } from "next-intl";
import { Button } from "../ui/Button";
import { Badge } from "../ui/Badge";
import { Github, BookOpen } from "lucide-react";
import { HeroStage } from "./HeroStage";
import { ScrollCue } from "../ui/ScrollCue";
import { Reveal } from "../ui/Reveal";

export function Hero() {
  const t = useTranslations("home.hero");
  const tNav = useTranslations("nav");
  const tHome = useTranslations("home");
  return (
    <section className="relative isolate overflow-hidden border-b border-cyan-line pt-20 pb-24">
      <div
        className="absolute inset-0 -z-10 bg-grid-dark bg-grid opacity-50"
        style={{
          maskImage:
            "radial-gradient(ellipse at top, black 30%, transparent 75%)",
          WebkitMaskImage:
            "radial-gradient(ellipse at top, black 30%, transparent 75%)",
        }}
      />
      <div className="absolute -top-40 left-1/2 -z-10 h-[480px] w-[820px] -translate-x-1/2 rounded-full bg-duo opacity-25 blur-3xl" />
      <div className="absolute -top-20 right-0 -z-10 h-[320px] w-[420px] rounded-full bg-magenta/30 blur-3xl" />

      <div className="container mx-auto grid max-w-6xl grid-cols-1 items-center gap-12 px-6 lg:grid-cols-12">
        <div className="lg:col-span-7">
          <Reveal>
            <Badge className="mb-6 border-cyan-line bg-bg-2/60 text-cyan">
              {t("eyebrow")}
            </Badge>
          </Reveal>
          <Reveal delay={120}>
            <h1 className="font-display text-4xl font-bold leading-[1.05] tracking-tight sm:text-5xl md:text-6xl">
              <span className="text-ink">{t("titlePrefix")}</span>{" "}
              <span className="text-duo">{t("titleHighlight")}</span>
              <span className="text-ink">{t("titleSuffix")}</span>
            </h1>
          </Reveal>
          <Reveal delay={220}>
            <p className="mt-6 max-w-2xl text-lg leading-8 text-ink-muted">
              {t("subtitle")}
            </p>
          </Reveal>
          <Reveal delay={320}>
            <div className="mt-10 flex flex-wrap items-center gap-3">
              <Button href="/quickstart" variant="primary">
                {t("ctaPrimary")}
              </Button>
              <Button href="/docs/requirements" variant="secondary">
                <BookOpen className="mr-2 h-4 w-4" /> {t("ctaSecondary")}
              </Button>
              <Button
                href="https://github.com/agentcodinglab/aicodingagentteam"
                external
                variant="ghost"
              >
                <Github className="mr-2 h-4 w-4" /> {tNav("github")}
              </Button>
            </div>
          </Reveal>
          <Reveal delay={440}>
            <div className="mt-10 flex flex-wrap items-center gap-x-6 gap-y-2 text-xs uppercase tracking-[0.18em] text-ink-muted2">
              <span className="text-cyan">{t("meta1")}</span>
              <span>{t("meta2")}</span>
              <span>{t("meta3")}</span>
            </div>
          </Reveal>
        </div>
        <div className="lg:col-span-5">
          <Reveal delay={200}>
            <HeroStage />
          </Reveal>
        </div>
      </div>

      <Reveal delay={700}>
        <div id="content" className="mt-16 flex justify-center">
          <ScrollCue label={tHome("scrollCue")} />
        </div>
      </Reveal>
    </section>
  );
}