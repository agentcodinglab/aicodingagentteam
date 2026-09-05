import { useTranslations } from 'next-intl';

export function SiteFooter() {
  const t = useTranslations('footer');
  return (
    <footer className="border-t border-slate-200 py-10 dark:border-slate-800">
      <div className="container mx-auto max-w-6xl px-6 text-center text-sm text-slate-600 dark:text-slate-400">
        {t('copyright')}
      </div>
    </footer>
  );
}