import { useTranslations } from "next-intl";
import { Reveal } from "../ui/Reveal";

const LAYERS = [
  { key: "router", title: "Router", sub: "Intent \u2192 workflow" },
  { key: "planner", title: "Planner", sub: "DAG of writers + reviewers" },
  { key: "scheduler", title: "Scheduler", sub: "Parallel reviewers \u00b7 serial writers" },
  { key: "knowledge", title: "Knowledge", sub: "BM25 RAG + project memory" },
  { key: "qualityGate", title: "Quality Gate", sub: "golangci-lint \u00b7 vet \u00b7 test (machine)" },
] as const;

export function ArchitectureDiagram() {
  const t = useTranslations("architecture.layers");
  return (
    <div className="grid gap-3">
      {LAYERS.map((layer, i) => (
        <Reveal key={layer.key} delay={i * 80}>
          <div className="group relative flex items-stretch overflow-hidden rounded-xl border border-cyan-line bg-bg-panel transition-all transition-all hover:border-cyan/60">
            <div className="flex w-16 flex-shrink-0 items-center justify-center bg-duo text-lg font-bold text-bg">
              {i + 1}
            </div>
            <div className="flex flex-1 flex-col justify-center px-5 py-3">
              <div className="flex flex-wrap items-baseline gap-x-3">
                <span className="font-display text-xs font-semibold uppercase tracking-[0.18em] text-cyan">
                  {layer.title}
                </span>
                <span className="font-mono text-xs text-ink-muted2">
                  {layer.sub}
                </span>
              </div>
              <p className="mt-1 text-sm text-ink-muted">{t(layer.key)}</p>
            </div>
            <div className="pointer-events-none absolute right-4 top-1/2 -translate-y-1/2 font-mono text-[10px] uppercase tracking-wider text-ink-muted2/50">
              layer.{String(i + 1).padStart(2, "0")}
            </div>
          </div>
        </Reveal>
      ))}
    </div>
  );
}