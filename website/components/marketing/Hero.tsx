import { useTranslations } from 'next-intl';
import { Button } from '../ui/Button';
import { Badge } from '../ui/Badge';

export function Hero() {
  const t = useTranslations('home.hero');
  const tNav = useTranslations('nav');
  return (
    <section className="relative overflow-hidden border-b border-slate-200 bg-gradient-to-b from-brand-50/60 via-white to-white py-24 dark:border-slate-800 dark:from-brand-950/40 dark:via-slate-950 dark:to-slate-950">
      <div className="container mx-auto max-w-5xl px-6 text-center">
        <Badge className="mb-6">{t('eyebrow')}</Badge>
        <h1 className="text-4xl font-bold tracking-tight sm:text-5xl md:text-6xl">{t('title')}</h1>
        <p className="mx-auto mt-6 max-w-2xl text-lg leading-8 text-slate-600 dark:text-slate-300">{t('subtitle')}</p>
        <div className="mt-10 flex flex-wrap items-center justify-center gap-4">
          <Button href="/quickstart" variant="primary">{t('ctaPrimary')}</Button>
          <Button href="/docs/requirements" variant="secondary">{t('ctaSecondary')}</Button>
          <Button href="https://github.com/agentcodinglab/aicodingagentteam" external variant="ghost">
            {tNav('github')} ↗
          </Button>
        </div>
      </div>
    </section>
  );
}