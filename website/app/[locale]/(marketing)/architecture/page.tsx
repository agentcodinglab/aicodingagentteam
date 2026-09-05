import { setRequestLocale } from "next-intl/server";
import { useTranslations } from "next-intl";
import { Reveal } from "@/components/ui/Reveal";
import { PageHero } from "@/components/marketing/PageHero";

const LAYERS = [
  { key: "router", title: "Router", sub: "Intent \u2192 workflow" },
  { key: "planner", title: "Planner", sub: "DAG of writers + reviewers" },
  { key: "scheduler", title: "Scheduler", sub: "Parallel reviewers \u00b7 serial writers" },
  { key: "knowledge", title: "Knowledge", sub: "BM25 RAG + project memory" },
  { key: "qualityGate", title: "Quality Gate", sub: "golangci-lint \u00b7 vet \u00b7 test (machine)" },
] as const;

export default function ArchitecturePage({
  params: { locale },
}: {
  params: { locale: string };
}) {
  setRequestLocale(locale);
  return <Content />;
}

function Content() {
  const t = useTranslations("architecture");
  const tLayers = useTranslations("architecture.layers");
  return (
    <>
      <PageHero eyebrow="// architecture" title={t("title")} subtitle={t("subtitle")} />
      <section className="border-b border-cyan-line/60 py-20">
        <div className="container mx-auto max-w-4xl px-6">
          <div className="grid gap-3">
            {LAYERS.map((layer, i) => (
              <Reveal key={layer.key} delay={i * 80}>
                <div className="group flex items-stretch overflow-hidden rounded-xl border border-cyan-line bg-bg-panel transition-all hover:border-cyan/60">
                  <div className="flex w-16 flex-shrink-0 items-center justify-center bg-duo font-display text-lg font-bold text-bg">
                    {i + 1}
                  </div>
                  <div className="flex flex-1 flex-col justify-center px-5 py-4">
                    <div className="flex flex-wrap items-baseline gap-x-3">
                      <span className="font-display text-xs font-semibold uppercase tracking-[0.18em] text-cyan">
                        {layer.title}
                      </span>
                      <span className="font-mono text-xs text-ink-muted2">
                        {layer.sub}
                      </span>
                    </div>
                    <p className="mt-1 text-sm text-ink-muted">{tLayers(layer.key)}</p>
                  </div>
                </div>
              </Reveal>
            ))}
          </div>
        </div>
      </section>
    </>
  );
}