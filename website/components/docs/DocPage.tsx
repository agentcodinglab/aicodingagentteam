import { setRequestLocale, getTranslations } from "next-intl/server";
import fs from "node:fs/promises";
import path from "node:path";
import ReactMarkdown from "react-markdown";
import remarkGfm from "remark-gfm";
import { DocContent } from "./DocContent";
import { Sidebar } from "./Sidebar";
import { TOC } from "./TOC";

type Props = {
  locale: string;
  slug: string;
};

export async function DocPage({ locale, slug }: Props) {
  setRequestLocale(locale);

  const file = `${slug}.md`;
  const localized = path.join(process.cwd(), "content", "docs", locale, file);
  const en = path.join(process.cwd(), "content", "docs", "en", file);
  let raw: string;
  try {
    raw = await fs.readFile(localized, "utf8");
  } catch {
    raw = await fs.readFile(en, "utf8");
  }
  const stripped = raw.replace(/^#\s+.*\r?\n/, "");

  return (
    <div className="container mx-auto flex max-w-6xl gap-8 px-6 py-10">
      <Sidebar />
      <article className="prose-doc min-w-0 flex-1">
        <DocContent>
          <ReactMarkdown remarkPlugins={[remarkGfm]}>{stripped}</ReactMarkdown>
        </DocContent>
      </article>
      <TOC />
    </div>
  );
}