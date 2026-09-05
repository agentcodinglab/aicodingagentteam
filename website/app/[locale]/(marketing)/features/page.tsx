import { setRequestLocale } from "next-intl/server";
import { useTranslations } from "next-intl";
import { Card } from "@/components/ui/Card";
import { Reveal } from "@/components/ui/Reveal";
import { Tilt } from "@/components/ui/Tilt";
import { PageHero } from "@/components/marketing/PageHero";
import {
  Bot,
  Container,
  GitBranch,
  ShieldCheck,
  Database,
  Network,
  Lock,
  Scaling,
} from "lucide-react";

const ICONS = [Bot, Container, GitBranch, ShieldCheck, Database, Network, Lock, Scaling];
const KEYS = [
  "separation",
  "containerRoles",
  "a2a",
  "quality",
  "rag",
  "protocols",
  "local",
  "scale",
] as const;

export default function FeaturesPage({
  params: { locale },
}: {
  params: { locale: string };
}) {
  setRequestLocale(locale);
  return <Content />;
}

function Content() {
  const t = useTranslations("features");
  const items = useTranslations("features.items");
  return (
    <>
      <PageHero eyebrow="// features" title={t("title")} subtitle={t("subtitle")} />
      <section className="border-b border-cyan-line/60 py-20">
        <div className="container mx-auto max-w-6xl px-6">
          <div className="grid grid-cols-1 gap-5 md:grid-cols-2 lg:grid-cols-3">
            {KEYS.map((k, i) => {
              const Icon = ICONS[i];
              return (
                <Reveal key={k} delay={i * 60}>
                  <Tilt className="h-full">
                    <Card className="flex h-full flex-col">
                      <span className="inline-flex h-10 w-10 items-center justify-center rounded-lg border border-cyan-line bg-cyan/10 text-cyan">
                        <Icon className="h-5 w-5" />
                      </span>
                      <h2 className="mt-5 text-base font-semibold text-ink">
                        {items(`${k}.title`)}
                      </h2>
                      <p className="mt-2 text-sm leading-6 text-ink-muted">
                        {items(`${k}.body`)}
                      </p>
                      <span className="mt-4 font-mono text-[10px] uppercase tracking-wider text-ink-muted2/60">
                        {String(i + 1).padStart(2, "0")} / {KEYS.length}
                      </span>
                    </Card>
                  </Tilt>
                </Reveal>
              );
            })}
          </div>
        </div>
      </section>
    </>
  );
}