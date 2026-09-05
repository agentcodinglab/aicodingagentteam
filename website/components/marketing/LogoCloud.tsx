import { useTranslations } from "next-intl";

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
    <section className="border-b border-slate-200 bg-slate-50/60 py-12 dark:border-slate-800 dark:bg-slate-900/40">
      <div className="container mx-auto max-w-6xl px-6">
        <p className="text-center text-xs font-semibold uppercase tracking-[0.18em] text-slate-500 dark:text-slate-400">
          {t("trustedBy")}
        </p>
        <div className="mt-6 flex flex-wrap items-center justify-center gap-x-10 gap-y-4">
          {ITEMS.map((name) => (
            <span
              key={name}
              className="font-mono text-sm font-semibold tracking-wide text-slate-500 transition-colors hover:text-slate-900 dark:text-slate-400 dark:hover:text-slate-100"
            >
              {name}
            </span>
          ))}
        </div>
      </div>
    </section>
  );
}