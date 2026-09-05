import { Terminal } from "lucide-react";
import { Reveal } from "../ui/Reveal";

const LINES: Array<{ prompt?: string; body: string; kind?: "out" | "ok" | "dim" }> = [
  { prompt: "$", body: "./bin/aicat run \"Build a REST API\" --backend codex" },
  { kind: "dim", body: "[router]  intent=build_api   workflow=full_team   backend=codex" },
  { kind: "dim", body: "[planner] nodes=9   writers=3   reviewers=6   dag=built (12ms)" },
  { kind: "dim", body: "[sched]   dispatched reviewer:qa + reviewer:security + reviewer:arch (parallel)" },
  { kind: "dim", body: "[sched]   dispatched writer:backend \u2192 writer:frontend \u2192 writer:devops (serial)" },
  { kind: "out", body: "\u2192 3 writers \u00b7 6 reviewers \u00b7 1 quality gate \u00b7 1 proof-pack" },
  { kind: "ok", body: "\u2713 proof-pack.zip  plan.json \u00b7 verify.jsonl \u00b7 scorecard.md \u00b7 delivery-summary.md" },
  { kind: "dim", body: "  gate: 0 lint \u00b7 0 vet \u00b7 0 test \u00b7 score=98 / 100" },
];

function colorClass(kind?: "out" | "ok" | "dim"): string {
  switch (kind) {
    case "ok":
      return "text-ok";
    case "out":
      return "text-magenta";
    case "dim":
      return "text-ink-muted";
    default:
      return "text-ink";
  }
}

export function CodeExample() {
  return (
    <Reveal>
      <div className="overflow-hidden rounded-2xl border border-cyan-line bg-bg-panel shadow-duo">
        <div className="flex items-center justify-between border-b border-cyan-line bg-bg-2/80 px-4 py-2.5 text-xs">
          <div className="flex items-center gap-1.5">
            <span className="h-3 w-3 rounded-full bg-[#ff5f56]/80" />
            <span className="h-3 w-3 rounded-full bg-[#ffbd2e]/80" />
            <span className="h-3 w-3 rounded-full bg-[#27c93f]/80" />
          </div>
          <div className="flex items-center gap-1.5 font-mono text-ink-muted2">
            <Terminal className="h-3.5 w-3.5" />aicat run
          </div>
          <span className="font-mono text-[10px] uppercase tracking-wider text-ink-muted2/60">
            bash
          </span>
        </div>
        <pre className="overflow-x-auto bg-[#04060a] p-6 font-mono text-[13px] leading-7 text-ink">
          {LINES.map((l, i) => (
            <div key={i} className={colorClass(l.kind)}>
              {l.prompt ? <span className="text-cyan">{l.prompt} </span> : null}
              {l.body}
            </div>
          ))}
          <div className="mt-1 inline-flex items-center gap-1 text-cyan">
            <span className="inline-block h-3.5 w-2 animate-cursor-blink bg-cyan align-middle" />
          </div>
        </pre>
      </div>
    </Reveal>
  );
}