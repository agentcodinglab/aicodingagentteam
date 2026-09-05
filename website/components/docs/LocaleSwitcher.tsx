"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { useState, useRef, useEffect } from "react";
import { Globe, ChevronDown, Check } from "lucide-react";
import { locales, localeNames, type Locale, defaultLocale } from "../../lib/i18n";

export function LocaleSwitcher() {
  const pathname = usePathname() || "/";
  const segs = pathname.split("/").filter(Boolean);
  const current: Locale = (locales.includes(segs[0] as Locale) ? segs[0] : defaultLocale) as Locale;
  const rest = "/" + segs.slice(1).join("/");

  const [open, setOpen] = useState(false);
  const ref = useRef<HTMLDivElement>(null);
  useEffect(() => {
    const onClick = (e: MouseEvent) => {
      if (ref.current && !ref.current.contains(e.target as Node)) setOpen(false);
    };
    document.addEventListener("click", onClick);
    return () => document.removeEventListener("click", onClick);
  }, []);

  function hrefFor(l: Locale): string {
    const suffix = rest === "/" ? "/" : rest;
    return l === defaultLocale ? suffix : `/${l}${suffix}`;
  }

  function remember(l: Locale) {
    try { localStorage.setItem("locale", l); } catch (e) {}
    setOpen(false);
  }

  return (
    <div ref={ref} className="relative">
      <button
        type="button"
        onClick={() => setOpen((v) => !v)}
        className="inline-flex items-center gap-1.5 rounded-md px-2.5 py-1.5 text-sm text-ink-muted hover:bg-bg-2/80"
        aria-haspopup="menu"
        aria-expanded={open}
      >
        <Globe className="h-4 w-4" />
        <span>{localeNames[current]}</span>
        <ChevronDown className="h-3 w-3" />
      </button>
      {open && (
        <div
          role="menu"
          className="absolute right-0 top-full z-50 mt-1 max-h-80 w-44 overflow-y-auto rounded-md border border-cyan-line/60 bg-bg-2 py-1 shadow-lg"
        >
          {locales.map((l) => (
            <Link
              key={l}
              href={hrefFor(l)}
              onClick={() => remember(l)}
              className={`flex items-center justify-between px-3 py-1.5 text-sm hover:bg-bg-2/80 ${
                l === current ? "font-semibold text-cyan" : "text-ink-muted"
              }`}
              role="menuitem"
            >
              <span>{localeNames[l]}</span>
              {l === current ? <Check className="h-3.5 w-3.5" /> : null}
            </Link>
          ))}
        </div>
      )}
    </div>
  );
}