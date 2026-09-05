import { useTranslations } from 'next-intl';
import { Terminal } from 'lucide-react';

export function Quickstart() {
  const t = useTranslations('quickstart.steps');
  const order = ['build', 'init', 'run', 'verify'] as const;
  const commands: Record<(typeof order)[number], string> = {
    build: 'go build -o bin/aicodingagentteam ./cmd/aicodingagentteam',
    init: './bin/aicodingagentteam init',
    run: './bin/aicodingagentteam run "Build a REST API" --backend codex',
    verify: './bin/aicodingagentteam verify',
  };
  return (
    <section className="py-20">
      <div className="container mx-auto max-w-5xl px-6">
        <div className="grid grid-cols-1 gap-6 md:grid-cols-2">
          {order.map((k) => (
            <div key={k} className="rounded-lg border border-slate-200 bg-white p-6 dark:border-slate-800 dark:bg-slate-900">
              <h3 className="text-base font-semibold">{t(`${k}.title`)}</h3>
              <p className="mt-1 text-sm text-slate-600 dark:text-slate-300">{t(`${k}.body`)}</p>
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