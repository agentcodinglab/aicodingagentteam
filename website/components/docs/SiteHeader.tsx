import Link from 'next/link';
import { useTranslations } from 'next-intl';
import { LocaleSwitcher } from './LocaleSwitcher';
import { ThemeToggle } from './ThemeToggle';
import { Github } from 'lucide-react';

export function SiteHeader() {
  const t = useTranslations('nav');
  return (
    <header className="sticky top-0 z-40 w-full border-b border-slate-200 bg-white/80 backdrop-blur dark:border-slate-800 dark:bg-slate-950/80">
      <div className="container mx-auto flex h-16 max-w-6xl items-center justify-between px-6">
        <Link href="/" className="flex items-center gap-2 text-sm font-semibold">
          <span className="inline-flex h-7 w-7 items-center justify-center rounded-md bg-brand-600 text-white">A</span>
          <span>AiCodingAgentTeam</span>
        </Link>
        <nav className="hidden items-center gap-6 text-sm md:flex">
          <Link href="/features" className="text-slate-700 hover:text-brand-600 dark:text-slate-200 dark:hover:text-brand-400">{t('features')}</Link>
          <Link href="/architecture" className="text-slate-700 hover:text-brand-600 dark:text-slate-200 dark:hover:text-brand-400">{t('architecture')}</Link>
          <Link href="/quickstart" className="text-slate-700 hover:text-brand-600 dark:text-slate-200 dark:hover:text-brand-400">{t('quickstart')}</Link>
          <Link href="/docs/requirements" className="text-slate-700 hover:text-brand-600 dark:text-slate-200 dark:hover:text-brand-400">{t('docs')}</Link>
        </nav>
        <div className="flex items-center gap-1">
          <Link href="https://github.com/agentcodinglab/aicodingagentteam" target="_blank" rel="noreferrer" className="inline-flex h-9 w-9 items-center justify-center rounded-md text-slate-700 hover:bg-slate-100 dark:text-slate-200 dark:hover:bg-slate-800" aria-label="GitHub">
            <Github className="h-4 w-4" />
          </Link>
          <ThemeToggle />
          <LocaleSwitcher />
        </div>
      </div>
    </header>
  );
}