'use client';
import Link from 'next/link';
import { useTranslations } from 'next-intl';
import { usePathname } from 'next/navigation';
import { docsNav } from '../../lib/nav';

export function Sidebar() {
  const t = useTranslations('docs.nav');
  const pathname = usePathname() || '';
  const segs = pathname.split('/').filter(Boolean);
  const current = segs[segs.length - 1];
  return (
    <nav className="sticky top-20 hidden h-[calc(100vh-5rem)] w-64 overflow-y-auto border-r border-slate-200 py-8 pr-4 text-sm dark:border-slate-800 lg:block">
      <ul className="space-y-1">
        {docsNav.map((d) => {
          const isActive = current === d.slug;
          return (
            <li key={d.slug}>
              <Link
                href={`/docs/${d.slug}`}
                className={`block rounded px-3 py-1.5 transition-colors ${
                  isActive
                    ? 'bg-brand-50 font-medium text-brand-700 dark:bg-brand-950/40 dark:text-brand-300'
                    : 'text-slate-700 hover:bg-slate-100 dark:text-slate-300 dark:hover:bg-slate-800'
                }`}
              >
                {t(d.titleKey.split('.').pop() as any)}
              </Link>
            </li>
          );
        })}
      </ul>
    </nav>
  );
}