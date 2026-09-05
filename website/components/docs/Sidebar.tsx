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
    <nav className="sticky top-20 hidden h-[calc(100vh-5rem)] w-64 overflow-y-auto border-r border-cyan-line/60 py-8 pr-4 text-sm  lg:block">
      <ul className="space-y-1">
        {docsNav.map((d) => {
          const isActive = current === d.slug;
          return (
            <li key={d.slug}>
              <Link
                href={`/docs/${d.slug}`}
                className={`block rounded px-3 py-1.5 transition-colors ${
                  isActive
                    ? 'bg-cyan/10 font-medium text-cyan  '
                    : 'text-ink-muted hover:bg-bg-2  '
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
