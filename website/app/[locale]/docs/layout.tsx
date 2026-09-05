import { ReactNode } from 'react';
import { Sidebar } from '../../../components/docs/Sidebar';
import { TOC } from '../../../components/docs/TOC';

export default function DocsLayout({ children }: { children: ReactNode }) {
  return (
    <div className="container mx-auto flex max-w-6xl gap-8 px-6 py-10">
      <Sidebar />
      <article className="flex-1 min-w-0">{children}</article>
      <TOC />
    </div>
  );
}