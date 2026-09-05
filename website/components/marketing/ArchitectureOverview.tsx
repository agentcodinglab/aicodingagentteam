import { useTranslations } from 'next-intl';

export function ArchitectureOverview() {
  const t = useTranslations('architecture.layers');
  const items = ['router', 'planner', 'scheduler', 'knowledge', 'qualityGate'] as const;
  return (
    <section className="border-b border-slate-200 py-20 dark:border-slate-800">
      <div className="container mx-auto max-w-5xl px-6">
        <ol className="space-y-4">
          {items.map((k, i) => (
            <li key={k} className="flex gap-4 rounded-lg border border-slate-200 bg-white p-5 dark:border-slate-800 dark:bg-slate-900">
              <span className="flex h-8 w-8 flex-shrink-0 items-center justify-center rounded-full bg-brand-600 text-sm font-semibold text-white">{i + 1}</span>
              <div>
                <div className="text-sm font-semibold uppercase tracking-wide text-brand-600 dark:text-brand-400">{k}</div>
                <div className="mt-1 text-slate-700 dark:text-slate-300">{t(k)}</div>
              </div>
            </li>
          ))}
        </ol>
      </div>
    </section>
  );
}