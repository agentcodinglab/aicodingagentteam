'use client';
import Link from 'next/link';
import { usePathname } from 'next/navigation';
import { useState, useRef, useEffect } from 'react';
import { Globe, ChevronDown } from 'lucide-react';
import { locales, localeNames, type Locale } from '../../lib/i18n';

export function LocaleSwitcher() {
  const pathname = usePathname() || '/';
  const segs = pathname.split('/').filter(Boolean);
  const current: Locale = (locales.includes(segs[0] as Locale) ? segs[0] : 'en') as Locale;
  const rest = '/' + segs.slice(1).join('/');
  const [open, setOpen] = useState(false);
  const ref = useRef<HTMLDivElement>(null);
  useEffect(() => {
    const onClick = (e: MouseEvent) => {
      if (ref.current && !ref.current.contains(e.target as Node)) setOpen(false);
    };
    document.addEventListener('click', onClick);
    return () => document.removeEventListener('click', onClick);
  }, []);
  return (
    <div ref={ref} className="relative">
      <button
        type="button"
        onClick={() => setOpen((v) => !v)}
        className="inline-flex items-center gap-1.5 rounded-md px-2.5 py-1.5 text-sm text-slate-700 hover:bg-slate-100 dark:text-slate-200 dark:hover:bg-slate-800"
        aria-haspopup="menu"
        aria-expanded={open}
      >
        <Globe className="h-4 w-4" />
        <span>{localeNames[current]}</span>
        <ChevronDown className="h-3 w-3" />
      </button>
      {open && (
        <div role="menu" className="absolute right-0 top-full z-50 mt-1 max-h-80 w-44 overflow-y-auto rounded-md border border-slate-200 bg-white py-1 shadow-lg dark:border-slate-800 dark:bg-slate-900">
          {locales.map((l) => (
            <Link
              key={l}
              href={rest === '/' ? `/${l}/` : `${rest}`}
              onClick={() => {
                if (l !== current) {
                  // replace current locale prefix
                  window.location.href = (l === 'en' ? '' : `/${l}`) + (rest === '/' ? '/' : rest);
                }
              }}
              className={`block px-3 py-1.5 text-sm hover:bg-slate-100 dark:hover:bg-slate-800 ${l === current ? 'font-semibold text-brand-600 dark:text-brand-400' : 'text-slate-700 dark:text-slate-200'}`}
              role="menuitem"
            >
              {localeNames[l]}
            </Link>
          ))}
        </div>
      )}
    </div>
  );
}