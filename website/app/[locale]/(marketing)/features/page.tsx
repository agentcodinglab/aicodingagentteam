import { setRequestLocale } from 'next-intl/server';
import { useTranslations } from 'next-intl';
import { Card } from '@/components/ui/Card';
import { Bot, Container, GitBranch, ShieldCheck, Database, Network, Lock } from 'lucide-react';

const ICONS = [Bot, Container, GitBranch, ShieldCheck, Database, Network, Lock];

export default function FeaturesPage({ params: { locale } }: { params: { locale: string } }) {
  setRequestLocale(locale);
  return <Content />;
}

function Content() {
  const t = useTranslations('features');
  const items = useTranslations('features.items');
  const keys = ['separation', 'containerRoles', 'a2a', 'quality', 'rag', 'protocols', 'local'] as const;
  return (
    <section className="py-16">
      <div className="container mx-auto max-w-5xl px-6">
        <h1 className="text-4xl font-bold">{t('title')}</h1>
        <p className="mt-3 text-lg text-slate-600 dark:text-slate-300">{t('subtitle')}</p>
        <div className="mt-10 grid grid-cols-1 gap-6 md:grid-cols-2">
          {keys.map((k, i) => {
            const Icon = ICONS[i];
            return (
              <Card key={k}>
                <Icon className="h-6 w-6 text-brand-600 dark:text-brand-400" />
                <h2 className="mt-4 text-lg font-semibold">{items(`${k}.title`)}</h2>
                <p className="mt-2 text-sm leading-6 text-slate-600 dark:text-slate-300">{items(`${k}.body`)}</p>
              </Card>
            );
          })}
        </div>
      </div>
    </section>
  );
}