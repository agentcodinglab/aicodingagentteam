import { setRequestLocale } from 'next-intl/server';
import { useTranslations } from 'next-intl';
import { Terminal } from 'lucide-react';

export default function QuickstartPage({ params: { locale } }: { params: { locale: string } }) {
  setRequestLocale(locale);
  return <Content />;
}

function Content() {
  const t = useTranslations('quickstart');
  const steps = useTranslations('quickstart.steps');
  const order = ['build', 'init', 'run', 'verify'] as const;
  const commands: Record<(typeof order)[number], string> = {
    build: 'go build -o bin/aicodingagentteam ./cmd/aicodingagentteam',
    init: './bin/aicodingagentteam init',
    run: './bin/aicodingagentteam run "Build a REST API" --backend codex',
    verify: './bin/aicodingagentteam verify',
  };
  return (
    <section className="py-16">
      <div className="container mx-auto max-w-5xl px-6">
        <h1 className="text-4xl font-bold">{t('title')}</h1>
        <p className="mt-3 text-lg text-slate-600 dark:text-slate-300">{t('subtitle')}</p>
        <div className="mt-10 grid grid-cols-1 gap-6 md:grid-cols-2">
          {order.map((k) => (
            <div key={k} className="rounded-lg border border-slate-200 bg-white p-6 dark:border-slate-800 dark:bg-slate-900">
              <h2 className="text-base font-semibold">{steps(`${k}.title`)}</h2>
              <p className="mt-1 text-sm text-slate-600 dark:text-slate-300">{steps(`${k}.body`)}</p>
              <pre className="mt-4 flex items-start gap-2 overflow-x-auto rounded bg-slate-950 p-3 text-xs text-slate-100">
                <Terminal className="mt-0.5 h-4 w-4 flex-shrink-0 text-slate-400" />
                <code>{commands[k]}</code>
              </pre>
            </div>
          ))}
        </div>
      </div>
    </section>
  );
}