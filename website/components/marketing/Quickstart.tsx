import { useTranslations } from "next-intl";
import { Terminal } from "lucide-react";

const STEPS = ["build", "init", "run", "verify"] as const;

const COMMANDS: Record<(typeof STEPS)[number], string> = {
  build: "go build -o bin/aicodingagentteam ./cmd/aicodingagentteam",
  init: "./bin/aicodingagentteam init",
  run: "./bin/aicodingagentteam run \"Build a REST API\" --backend codex",
  verify: "./bin/aicodingagentteam verify",
};

export function Quickstart() {
  const t = useTranslations("quickstart.steps");
  return (
    <div className="grid grid-cols-1 gap-5 md:grid-cols-2">
      {STEPS.map((k, i) => (
        <div
          key={k}
          className="flex flex-col rounded-xl border border-slate-200 bg-white p-6 dark:border-slate-800 dark:bg-slate-900"
        >
          <div className="mb-3 flex items-center gap-3">
            <span className="inline-flex h-8 w-8 items-center justify-center rounded-full bg-brand-600 text-sm font-semibold text-white">
              {i + 1}
            </span>
            <h3 className="text-base font-semibold">{t(`${k}.title`)}</h3>
          </div>
          <p className="text-sm leading-6 text-slate-600 dark:text-slate-300">{t(`${k}.body`)}</p>
          <pre className="mt-4 flex items-start gap-2 overflow-x-auto rounded-md bg-slate-950 p-3 font-mono text-xs text-slate-100">
            <Terminal className="mt-0.5 h-4 w-4 flex-shrink-0 text-slate-400" />
            <code>$ {COMMANDS[k]}</code>
          </pre>
        </div>
      ))}
    </div>
  );
}