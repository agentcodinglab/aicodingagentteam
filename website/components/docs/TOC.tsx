"use client";

import { useEffect, useState } from "react";
import { useTranslations } from "next-intl";

type Heading = { id: string; text: string; level: number };

export function TOC() {
  const t = useTranslations("docs");
  const [headings, setHeadings] = useState<Heading[]>([]);
  useEffect(() => {
    const els = Array.from(
      document.querySelectorAll<HTMLElement>(".prose-doc h2, .prose-doc h3"),
    );
    const mapped = els.map((el) => {
      if (!el.id) {
        const slug = (el.textContent || "")
          .trim()
          .toLowerCase()
          .replace(/[^a-z0-9\u4e00-\u9fa5]+/g, "-")
          .replace(/(^-|-$)/g, "");
        el.id = slug;
      }
      return {
        id: el.id,
        text: el.textContent || "",
        level: el.tagName === "H2" ? 2 : 3,
      };
    });
    setHeadings(mapped);
  }, []);
  if (headings.length === 0) return null;
  return (
    <aside className="sticky top-20 hidden h-[calc(100vh-5rem)] w-56 overflow-y-auto py-8 pl-4 text-xs xl:block">
      <div className="mb-2 font-semibold uppercase tracking-wide text-ink-muted2">
        {t("tableOfContents")}
      </div>
      <ul className="space-y-1.5 border-l border-cyan-line/60">
        {headings.map((h) => (
          <li key={h.id} className={h.level === 3 ? "pl-6" : "pl-3"}>
            <a
              href={`#${h.id}`}
              className="text-ink-muted transition-colors hover:text-cyan"
            >
              {h.text}
            </a>
          </li>
        ))}
      </ul>
    </aside>
  );
}