import { useTranslations } from "next-intl";
import { Button } from "../ui/Button";
import { Badge } from "../ui/Badge";
import { Github, Terminal } from "lucide-react";

export function Hero() {
  const t = useTranslations("home.hero");
  const tNav = useTranslations("nav");
  const lines = (t.raw("terminalLines") as string[]) ?? [];
  return (
    <section className="relative overflow-hidden border-b border-slate-200 bg-gradient-to-b from-brand-50/70 via-white to-white py-20 dark:border-slate-800 dark:from-brand-950/40 dark:via-slate-950 dark:to-slate-950 sm:py-28">
      <div className="absolute inset-x-0 top-0 -z-10 h-[420px] bg-[radial-gradient(ellipse_at_top,rgba(56,138,255,0.18),transparent_60%)] dark:bg-[radial-gradient(ellipse_at_top,rgba(45,138,255,0.20),transparent_60%)]" />
      <div className="container mx-auto grid max-w-6xl grid-cols-1 items-center gap-12 px-6 lg:grid-cols-12">
        <div className="lg:col-span-7">
          <Badge className="mb-6">{t("eyebrow")}</Badge>
          <h1 className="text-4xl font-bold tracking-tight sm:text-5xl md:text-6xl">{t("title")}</h1>
          <p className="mt-6 max-w-2xl text-lg leading-8 text-slate-600 dark:text-slate-300">{t("subtitle")}</p>
          <div className="mt-10 flex flex-wrap items-center gap-3">
            <Button href="/quickstart" variant="primary">{t("ctaPrimary")}</Button>
            <Button href="/docs/requirements" variant="secondary">{t("ctaSecondary")}</Button>
            <Button href="https://github.com/agentcodinglab/aicodingagentteam" external variant="ghost">
              <Github className="mr-2 h-4 w-4" />{tNav("github")}
            </Button>
          </div>
        </div>
        <div className="lg:col-span-5">
          <div className="overflow-hidden rounded-xl border border-slate-200 bg-slate-950 shadow-2xl shadow-brand-950/20 dark:border-slate-800">
            <div className="flex items-center justify-between border-b border-slate-800 bg-slate-900 px-4 py-2.5 text-xs">
              <div className="flex items-center gap-1.5">
                <span className="h-3 w-3 rounded-full bg-red-500/80" />
                <span className="h-3 w-3 rounded-full bg-yellow-500/80" />
                <span className="h-3 w-3 rounded-full bg-green-500/80" />
              </div>
              <div className="flex items-center gap-1.5 font-mono text-slate-400">
                <Terminal className="h-3.5 w-3.5" />{t("terminalTitle")}
              </div>
              <span className="w-10" />
            </div>
            <pre className="overflow-x-auto p-5 font-mono text-[13px] leading-6 text-slate-100">
              {lines.map((line, i) => (
                <div key={i} className={line.startsWith("→") || line.startsWith("✓") ? "text-emerald-400" : "text-slate-200"}>
                  {!line.startsWith("→") && !line.startsWith("✓") ? <span className="text-brand-400">{t("terminalPrompt")} </span> : null}
                  {line}
                </div>
              ))}
            </pre>
          </div>
        </div>
      </div>
    </section>
  );
}