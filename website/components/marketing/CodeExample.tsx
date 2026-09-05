import { Terminal } from "lucide-react";

const LINES: Array<{ prompt?: string; body: string; kind?: "out" | "ok" | "dim" }> = [
  { prompt: "$", body: "./bin/aicodingagentteam run \"Build a REST API\" --backend codex" },
  { kind: "dim", body: "[router]  intent=build_api   workflow=full_team   backend=codex" },
  { kind: "dim", body: "[planner] nodes=9   writers=3   reviewers=6   dag=built (12ms)" },
  { kind: "dim", body: "[sched]   dispatched reviewer:qa + reviewer:security + reviewer:arch (parallel)" },
  { kind: "dim", body: "[sched]   dispatched writer:backend → writer:frontend → writer:devops (serial)" },
  { kind: "out", body: "→ 3 writers · 6 reviewers · 1 quality gate · 1 proof-pack" },
  { kind: "ok", body: "✓ proof-pack.zip  plan.json · verify.jsonl · scorecard.md · delivery-summary.md" },
  { kind: "dim", body: "  gate: 0 lint · 0 vet · 0 test · score=98 / 100" },
];

export function CodeExample() {
  return (
    <div className="overflow-hidden rounded-2xl border border-slate-800 bg-slate-950 shadow-2xl shadow-brand-950/20">
      <div className="flex items-center justify-between border-b border-slate-800 bg-slate-900 px-4 py-2.5 text-xs">
        <div className="flex items-center gap-1.5">
          <span className="h-3 w-3 rounded-full bg-red-500/80" />
          <span className="h-3 w-3 rounded-full bg-yellow-500/80" />
          <span className="h-3 w-3 rounded-full bg-green-500/80" />
        </div>
        <div className="flex items-center gap-1.5 font-mono text-slate-400">
          <Terminal className="h-3.5 w-3.5" />aicodingagentteam run
        </div>
        <span className="w-10" />
      </div>
      <pre className="overflow-x-auto p-6 font-mono text-[13px] leading-7 text-slate-100">
        {LINES.map((l, i) => {
          const cls =
            l.kind === "ok"
              ? "text-emerald-400"
              : l.kind === "out"
              ? "text-brand-300"
              : l.kind === "dim"
              ? "text-slate-400"
              : "text-slate-100";
          return (
            <div key={i} className={cls}>
              {l.prompt ? <span className="text-brand-400">{l.prompt} </span> : null}
              {l.body}
            </div>
          );
        })}
      </pre>
    </div>
  );
}