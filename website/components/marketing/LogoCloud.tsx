import { useTranslations } from "next-intl";
import { Reveal } from "../ui/Reveal";

const ITEMS = [
  "Codex",
  "OpenCode",
  "Claude-Code",
  "DeepSeek-DSH",
  "gRPC",
  "MCP",
  "ACP",
  "A2A",
];

export function LogoCloud() {
  const t = useTranslations("home");
  return (
    <section className="border-b border-cyan-line/60 bg-bg-2/40 py-12">
      <Reveal>
        <div className="container mx-auto max-w-6xl px-6">
          <p className="text-center text-xs font-semibold uppercase tracking-[0.22em] text-ink-muted2">
            {t("trustedBy")}
          </p>
          <div className="mt-6 flex flex-wrap items-center justify-center gap-x-10 gap-y-4">
            {ITEMS.map((name, i) => (
              <span
                key={name}
                className="font-mono text-sm font-medium tracking-wide text-ink-muted transition-colors hover:text-duo"
                style={{ transitionDelay: `${i * 30}ms` } as React.CSSProperties}
              >
                {name}
              </span>
            ))}
          </div>
        </div>
      </Reveal>
    </section>
  );
}