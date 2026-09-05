import { setRequestLocale } from "next-intl/server";
import { useTranslations } from "next-intl";
import { Hero } from "@/components/marketing/Hero";
import { LogoCloud } from "@/components/marketing/LogoCloud";
import { FeatureGrid } from "@/components/marketing/FeatureGrid";
import { Stats } from "@/components/marketing/Stats";
import { CodeExample } from "@/components/marketing/CodeExample";
import { ArchitectureDiagram } from "@/components/marketing/ArchitectureDiagram";
import { Quickstart } from "@/components/marketing/Quickstart";
import { FinalCTA } from "@/components/marketing/FinalCTA";

export default function HomePage({ params: { locale } }: { params: { locale: string } }) {
  setRequestLocale(locale);
  return <HomeContent />;
}

function HomeContent() {
  const t = useTranslations("home");
  const tFeatures = useTranslations("features");
  return (
    <>
      <Hero />
      <LogoCloud />

      <Section
        title={tFeatures("title")}
        subtitle={tFeatures("subtitle")}
        eyebrow={t("featuresTitle")}
      >
        <FeatureGrid />
      </Section>

      <Section
        title={t("statsTitle")}
        subtitle={t("statsSubtitle")}
        alt
      >
        <Stats />
      </Section>

      <Section
        title={t("codeTitle")}
        subtitle={t("codeSubtitle")}
      >
        <CodeExample />
      </Section>

      <Section
        title={t("architectureTitle")}
        subtitle={t("architectureSubtitle")}
        alt
      >
        <ArchitectureDiagram />
      </Section>

      <Section
        title={t("quickstartTitle")}
        subtitle={t("quickstartSubtitle")}
      >
        <Quickstart />
      </Section>

      <FinalCTA />
    </>
  );
}

function Section({
  title,
  subtitle,
  children,
  alt,
  eyebrow,
}: {
  title: string;
  subtitle?: string;
  children: React.ReactNode;
  alt?: boolean;
  eyebrow?: string;
}) {
  return (
    <section
      className={`border-b border-slate-200 py-20 dark:border-slate-800 ${
        alt ? "bg-slate-50/60 dark:bg-slate-900/40" : "bg-white dark:bg-slate-950"
      }`}
    >
      <div className="container mx-auto max-w-6xl px-6">
        <div className="mx-auto mb-12 max-w-3xl text-center">
          {eyebrow ? (
            <p className="mb-3 text-xs font-semibold uppercase tracking-[0.18em] text-brand-600 dark:text-brand-400">
              {eyebrow}
            </p>
          ) : null}
          <h2 className="text-3xl font-bold tracking-tight sm:text-4xl">{title}</h2>
          {subtitle ? (
            <p className="mt-4 text-base leading-7 text-slate-600 dark:text-slate-300">{subtitle}</p>
          ) : null}
        </div>
        {children}
      </div>
    </section>
  );
}