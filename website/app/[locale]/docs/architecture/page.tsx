import { setRequestLocale } from 'next-intl/server';
import { DocContent } from '../../../../components/docs/DocContent';
import fs from 'node:fs/promises';
import path from 'node:path';

type Props = { params: { locale: string } };

export default async function Page({ params: { locale } }: Props) {
  setRequestLocale(locale);
  const file = locale === 'zh' ? 'architecture.md' : 'architecture.md';
  const full = path.join(process.cwd(), 'content', 'docs', locale, file);
  const raw = await fs.readFile(full, 'utf8');
  // Strip leading '# ' title because the layout already shows the page title
  const stripped = raw.replace(/^#\s+.*\n/, '');
  return (
    <DocContent>
      <Markdown source={stripped} />
    </DocContent>
  );
}

function Markdown({ source }: { source: string }) {
  // Minimal markdown renderer (headings, paragraphs, lists, code, tables, blockquote, links)
  const lines = source.split(/\r?\n/);
  const blocks: { type: string; lines: string[] }[] = [];
  let cur: { type: string; lines: string[] } | null = null;
  for (const ln of lines) {
    if (/^#{1,6}\s+/.test(ln)) {
      if (cur) blocks.push(cur);
      cur = { type: 'h' + ln.match(/^#{1,6}/)![0].length, lines: [ln.replace(/^#{1,6}\s+/, '')] };
    } else if (/^`/.test(ln)) {
      if (cur && cur.type === 'code') {
        blocks.push(cur);
        cur = null;
      } else {
        if (cur) blocks.push(cur);
        cur = { type: 'code', lines: [] };
      }
    } else if (/^\s*[-*]\s+/.test(ln)) {
      if (cur && cur.type === 'ul') cur.lines.push(ln.replace(/^\s*[-*]\s+/, ''));
      else { if (cur) blocks.push(cur); cur = { type: 'ul', lines: [ln.replace(/^\s*[-*]\s+/, '')] }; }
    } else if (/^\s*\d+\.\s+/.test(ln)) {
      if (cur && cur.type === 'ol') cur.lines.push(ln.replace(/^\s*\d+\.\s+/, ''));
      else { if (cur) blocks.push(cur); cur = { type: 'ol', lines: [ln.replace(/^\s*\d+\.\s+/, '')] }; }
    } else if (/^\s*\|.*\|\s*$/.test(ln)) {
      if (cur && cur.type === 'table') cur.lines.push(ln.trim());
      else { if (cur) blocks.push(cur); cur = { type: 'table', lines: [ln.trim()] }; }
    } else if (/^>\s?/.test(ln)) {
      if (cur && cur.type === 'quote') cur.lines.push(ln.replace(/^>\s?/, ''));
      else { if (cur) blocks.push(cur); cur = { type: 'quote', lines: [ln.replace(/^>\s?/, '')] }; }
    } else if (ln.trim() === '') {
      if (cur) { blocks.push(cur); cur = null; }
    } else {
      if (cur && cur.type === 'p') cur.lines.push(ln);
      else { if (cur) blocks.push(cur); cur = { type: 'p', lines: [ln] }; }
    }
  }
  if (cur) blocks.push(cur);

  return (
    <>
      {blocks.map((b, i) => {
        const text = (s: string) => s
          .replace(/([^]+)/g, '<code class="font-mono text-sm bg-slate-100 dark:bg-slate-800 px-1.5 py-0.5 rounded"></code>')
          .replace(/\*\*([^*]+)\*\*/g, '<strong></strong>')
          .replace(/\[([^\]]+)\]\(([^)]+)\)/g, '<a href=""></a>');
        if (b.type === 'h1') return <h1 key={i}>{b.lines[0]}</h1>;
        if (b.type === 'h2') return <h2 key={i} id={b.lines[0].toLowerCase().replace(/[^a-z0-9\u4e00-\u9fa5]+/g,'-')}>{b.lines[0]}</h2>;
        if (b.type === 'h3') return <h3 key={i} id={b.lines[0].toLowerCase().replace(/[^a-z0-9\u4e00-\u9fa5]+/g,'-')}>{b.lines[0]}</h3>;
        if (b.type === 'h4') return <h4 key={i}>{b.lines[0]}</h4>;
        if (b.type === 'code') return <pre key={i}><code>{b.lines.join('\n')}</code></pre>;
        if (b.type === 'ul') return <ul key={i}>{b.lines.map((l, j) => <li key={j} dangerouslySetInnerHTML={{ __html: text(l) }} />)}</ul>;
        if (b.type === 'ol') return <ol key={i}>{b.lines.map((l, j) => <li key={j} dangerouslySetInnerHTML={{ __html: text(l) }} />)}</ol>;
        if (b.type === 'quote') return <blockquote key={i}>{b.lines.map((l, j) => <p key={j} dangerouslySetInnerHTML={{ __html: text(l) }} />)}</blockquote>;
        if (b.type === 'table') {
          const rows = b.lines.filter(l => !/^\|?[\s\-:|]+\|?$/.test(l)).map(l => l.replace(/^\||\|$/g,'').split('|').map(c => c.trim()));
          if (rows.length === 0) return null;
          const [head, ...body] = rows;
          return (
            <table key={i}>
              <thead><tr>{head.map((c, j) => <th key={j} dangerouslySetInnerHTML={{ __html: text(c) }} />)}</tr></thead>
              <tbody>{body.map((r, j) => <tr key={j}>{r.map((c, k) => <td key={k} dangerouslySetInnerHTML={{ __html: text(c) }} />)}</tr>)}</tbody>
            </table>
          );
        }
        return <p key={i} dangerouslySetInnerHTML={{ __html: text(b.lines.join(' ')) }} />;
      })}
    </>
  );
}