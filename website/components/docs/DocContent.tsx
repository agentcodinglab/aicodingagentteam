import { ReactNode } from 'react';

export function DocContent({ children }: { children: ReactNode }) {
  return <div className="prose-doc">{children}</div>;
}