import { useTranslations } from 'next-intl';
import { Card } from '../ui/Card';
import { Bot, Container, GitBranch, ShieldCheck, Database, Network, Lock } from 'lucide-react';

const ICONS = [Bot, Container, GitBranch, ShieldCheck, Database, Network, Lock];

export function FeatureGrid() {
  const t = useTranslations('features.items');
  const keys = ['separation', 'containerRoles', 'a2a', 'quality', 'rag', 'protocols', 'local'] as const;
  return (
    <section className="border-b border-slate-200 py-20 dark:border-slate-800">
      <div className="container mx-auto max-w-6xl px-6">
        <div className="grid grid-cols-1 gap-6 md:grid-cols-2 lg:grid-cols-3">
          {keys.map((k, i) => {
            const Icon = ICONS[i];
            return (
              <Card key={k}>
                <Icon className="h-6 w-6 text-brand-600 dark:text-brand-400" />
                <h3 className="mt-4 text-lg font-semibold">{t(`${k}.title`)}</h3>
                <p className="mt-2 text-sm leading-6 text-slate-600 dark:text-slate-300">{t(`${k}.body`)}</p>
              </Card>
            );
          })}
        </div>
      </div>
    </section>
  );
}