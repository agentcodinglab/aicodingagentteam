import { setRequestLocale } from 'next-intl/server';
import { useTranslations } from 'next-intl';
import { Hero } from '../../components/marketing/Hero';
import { FeatureGrid } from '../../components/marketing/FeatureGrid';
import { ArchitectureOverview } from '../../components/marketing/ArchitectureOverview';
import { Quickstart } from '../../components/marketing/Quickstart';

export default function HomePage({ params: { locale } }: { params: { locale: string } }) {
  setRequestLocale(locale);
  return <HomeContent />;
}

function HomeContent() {
  const t = useTranslations('home');
  return (
    <>
      <Hero />
      <Section title={t('featuresTitle')}><FeatureGrid /></Section>
      <Section title={t('architectureTitle')}><ArchitectureOverview /></Section>
      <Section title={t('quickstartTitle')}><Quickstart /></Section>
      <section className="py-12">
        <div className="container mx-auto max-w-5xl px-6 text-center text-sm text-slate-500 dark:text-slate-400">
          {t('footerNote')}
        </div>
      </section>
    </>
  );
}

function Section({ title, children }: { title: string; children: React.ReactNode }) {
  return (
    <section className="py-16">
      <div className="container mx-auto max-w-5xl px-6">
        <h2 className="mb-8 text-2xl font-bold tracking-tight sm:text-3xl">{title}</h2>
        {children}
      </div>
    </section>
  );
}