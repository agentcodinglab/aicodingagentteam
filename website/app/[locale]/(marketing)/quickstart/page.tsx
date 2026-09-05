import { setRequestLocale } from "next-intl/server";
import { useTranslations } from "next-intl";
import { Terminal } from "lucide-react";
import { Reveal } from "@/components/ui/Reveal";
import { PageHero } from "@/components/marketing/PageHero";

const STEPS = ["build", "init", "run", "verify"] as const;

const COMMANDS: Record<(typeof STEPS)[number], string> = {
  build: "go build -o bin/aicat ./cmd/aicat",
  init: "./bin/aicat init",
  run: "./bin/aicat run \"Build a REST API\" --backend codex",
  verify: "./bin/aicat verify",
};

export default function QuickstartPage({
  params: { locale },
}: {
  params: { locale: string };
}) {
  setRequestLocale(locale);
  return <Content />;
}

function Content() {
  const t = useTranslations("quickstart");
  const tSteps = useTranslations("quickstart.steps");
  return (
    <>
      <PageHero eyebrow="// quickstart" title={t("title")} subtitle={t("subtitle")} />
      <section className="border-b border-cyan-line/60 py-20">
        <div className="container mx-auto max-w-5xl px-6">
          <div className="grid grid-cols-1 gap-5 md:grid-cols-2">
            {STEPS.map((k, i) => (
              <Reveal key={k} delay={i * 80}>
                <div className="flex h-full flex-col rounded-xl border border-cyan-line bg-bg-panel p-6 transition-all hover:border-cyan/60">
                  <div className="mb-3 flex items-center gap-3">
                    <span className="inline-flex h-9 w-9 items-center justify-center rounded-full bg-duo font-display text-sm font-semibold text-bg">
                      {i + 1}
                    </span>
                    <h2 className="font-display text-base font-semibold text-ink">
                      {tSteps(`${k}.title`)}
                    </h2>
                  </div>
                    <p className="text-sm leading-6 text-ink-muted">
                      {tSteps(`${k}.body`)}
                    </p>
                    <pre className="mt-4 flex items-start gap-2 overflow-x-auto rounded-md border border-cyan-line/60 bg-[#04060a] p-3 font-mono text-xs text-ink">
                      <Terminal className="mt-0.5 h-4 w-4 flex-shrink-0 text-cyan" />
                      <code>
                        <span className="text-magenta">$ </span>
                        {COMMANDS[k]}
                      </code>
                    </pre>
                  </div>
              </Reveal>
            ))}
          </div>
        </div>
      </section>
    </>
  );
}