import type { Metadata } from 'next';
import { ReactNode } from 'react';
import './globals.css';

export const metadata: Metadata = {
  title: 'AiCodingAgentTeam',
  description: 'A Golang-based AI coding orchestration platform.',
  metadataBase: new URL('https://agentcodinglab.github.io/aicodingagentteam'),
  openGraph: {
    title: 'AiCodingAgentTeam',
    description: 'A Golang-based AI coding orchestration platform.',
    type: 'website',
    locale: 'en_US',
  },
};

export default function RootLayout({ children }: { children: ReactNode }) {
  return children;
}