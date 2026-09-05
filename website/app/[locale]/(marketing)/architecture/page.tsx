import { setRequestLocale } from 'next-intl/server';
import { useTranslations } from 'next-intl';

export default function ArchitecturePage({ params: { locale } }: { params: { locale: string } }) {
  setRequestLocale(locale);
  return <Content />;
}

function Content() {
  const t = useTranslations('architecture');
  const layers = useTranslations('architecture.layers');
  const keys = ['router', 'planner', 'scheduler', 'knowledge', 'qualityGate'] as const;
  return (
    <section className="py-16">
      <div className="container mx-auto max-w-5xl px-6">
        <h1 className="text-4xl font-bold">{t('title')}</h1>
        <p className="mt-3 text-lg text-slate-600 dark:text-slate-300">{t('subtitle')}</p>
        <ol className="mt-10 space-y-4">
          {keys.map((k, i) => (
            <li key={k} className="flex gap-4 rounded-lg border border-slate-200 bg-white p-5 dark:border-slate-800 dark:bg-slate-900">
              <span className="flex h-8 w-8 flex-shrink-0 items-center justify-center rounded-full bg-brand-600 text-sm font-semibold text-white">{i + 1}</span>
              <div>
                <div className="text-sm font-semibold uppercase tracking-wide text-brand-600 dark:text-brand-400">{k}</div>
                <div className="mt-1 text-slate-700 dark:text-slate-300">{layers(k)}</div>
              </div>
            </li>
          ))}
        </ol>
      </div>
    </section>
  );
}