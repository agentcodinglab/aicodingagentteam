import { useTranslations } from "next-intl";
import { Card } from "../ui/Card";
import { Reveal } from "../ui/Reveal";
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

export function FeatureGrid() {
  const t = useTranslations("features.items");
  return (
    <div className="grid grid-cols-1 gap-5 md:grid-cols-2 lg:grid-cols-4">
      {KEYS.map((k, i) => {
        const Icon = ICONS[i];
        return (
          <Reveal key={k} delay={i * 60}>
            <Card className="flex h-full flex-col">
              <span className="inline-flex h-10 w-10 items-center justify-center rounded-lg border border-cyan-line bg-cyan/10 text-cyan">
                <Icon className="h-5 w-5" />
              </span>
              <h3 className="mt-5 text-base font-semibold text-ink">
                {t(`${k}.title`)}
              </h3>
              <p className="mt-2 text-sm leading-6 text-ink-muted">
                {t(`${k}.body`)}
              </p>
            </Card>
          </Reveal>
        );
      })}
    </div>
  );
}