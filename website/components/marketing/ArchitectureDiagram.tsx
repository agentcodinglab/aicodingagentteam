import { useTranslations } from "next-intl";

const LAYERS = [
  { key: "router",      title: "Router",      sub: "Intent → workflow" },
  { key: "planner",     title: "Planner",     sub: "DAG of writers + reviewers" },
  { key: "scheduler",   title: "Scheduler",   sub: "Parallel reviewers · serial writers" },
  { key: "knowledge",   title: "Knowledge",   sub: "BM25 RAG + project memory" },
  { key: "qualityGate", title: "Quality Gate", sub: "golangci-lint · vet · test (machine)" },
] as const;

export function ArchitectureDiagram() {
  const t = useTranslations("architecture.layers");
  return (
    <div className="grid gap-3">
      {LAYERS.map((layer, i) => (
        <div
          key={layer.key}
          className="group relative flex items-stretch overflow-hidden rounded-xl border border-slate-200 bg-white dark:border-slate-800 dark:bg-slate-900"
        >
          <div className="flex w-14 flex-shrink-0 items-center justify-center bg-gradient-to-br from-brand-600 to-brand-700 text-lg font-bold text-white">
            {i + 1}
          </div>
          <div className="flex flex-1 flex-col justify-center px-5 py-4">
            <div className="flex flex-wrap items-baseline gap-x-3">
              <span className="text-xs font-semibold uppercase tracking-wider text-brand-600 dark:text-brand-400">
                {layer.title}
              </span>
              <span className="text-xs text-slate-500 dark:text-slate-400">{layer.sub}</span>
            </div>
            <p className="mt-1 text-sm text-slate-700 dark:text-slate-300">{t(layer.key)}</p>
          </div>
          {i < LAYERS.length - 1 ? (
            <div className="pointer-events-none absolute -bottom-3 left-7 h-3 w-px bg-slate-300 dark:bg-slate-700 sm:left-7" />
          ) : null}
        </div>
      ))}
    </div>
  );
}