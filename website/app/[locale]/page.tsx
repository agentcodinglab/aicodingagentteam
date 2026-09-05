import { setRequestLocale } from "next-intl/server";
import { useTranslations } from "next-intl";
import { Hero } from "@/components/marketing/Hero";
import { LogoCloud } from "@/components/marketing/LogoCloud";
import { Section } from "@/components/marketing/Section";
import { FeatureGrid } from "@/components/marketing/FeatureGrid";
import { Stats } from "@/components/marketing/Stats";
import { CodeExample } from "@/components/marketing/CodeExample";
import { ArchitectureDiagram } from "@/components/marketing/ArchitectureDiagram";
import { Quickstart } from "@/components/marketing/Quickstart";
import { FinalCTA } from "@/components/marketing/FinalCTA";

export default function HomePage({
  params: { locale },
}: {
  params: { locale: string };
}) {
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
        eyebrow={t("featuresTitle")}
        title={tFeatures("title")}
        subtitle={tFeatures("subtitle")}
      >
        <FeatureGrid />
      </Section>

      <Section eyebrow="// metrics" title={t("statsTitle")} subtitle={t("statsSubtitle")} alt>
        <Stats />
      </Section>

      <Section eyebrow="$ aicat run" title={t("codeTitle")} subtitle={t("codeSubtitle")}>
        <CodeExample />
      </Section>

      <Section
        eyebrow="// flow"
        title={t("architectureTitle")}
        subtitle={t("architectureSubtitle")}
        alt
      >
        <ArchitectureDiagram />
      </Section>

      <Section eyebrow="// 4 steps" title={t("quickstartTitle")} subtitle={t("quickstartSubtitle")}>
        <Quickstart />
      </Section>

      <FinalCTA />
    </>
  );
}