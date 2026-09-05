"use client";

import { useEffect, useMemo, useState } from "react";
import { useLocale, useTranslations } from "next-intl";
import { RotateCcw, Terminal as TerminalIcon } from "lucide-react";

type LineKind = "prompt" | "sys" | "stage" | "file" | "ok" | "done" | "dim";
type Line = { text: string; kind: LineKind };

function slidesFor(locale: string): Line[][] {
  const isZh = locale.startsWith("zh");
  if (isZh) {
    return [
      [
        { text: "$ aicat run \"\u6784\u5efa REST API\" --backend codex", kind: "prompt" },
        { text: "router \u2192 intent=build_api   workflow=full_team", kind: "sys" },
        { text: "phase: docs_confirm", kind: "stage" },
        { text: "output/plan.json written", kind: "file" },
        { text: "output/verify.jsonl written", kind: "file" },
        { text: "scheduler \u2192 dispatching 6 reviewers in parallel", kind: "sys" },
        { text: "scheduler \u2192 dispatching 3 writers (single-writer lock)", kind: "sys" },
        { text: "quality_gate: lint 0 \u00b7 vet 0 \u00b7 tests 0", kind: "ok" },
        { text: "release/proof-pack.zip generated", kind: "file" },
        { text: "\u4ea4\u4ed8\u95ed\u73af\u5b8c\u6210\u3002", kind: "done" },
      ],
      [
        { text: "$ aicat continue --phase docs", kind: "prompt" },
        { text: "doc_agent \u2192 syncing context from PRD\u2026", kind: "sys" },
        { text: "phase: frontend_implement", kind: "stage" },
        { text: "output/uiux.md read", kind: "file" },
        { text: "frontend_writer \u2192 building React components\u2026", kind: "sys" },
        { text: "lint: 0 errors, 0 warnings", kind: "ok" },
        { text: "status: pending_backend", kind: "done" },
      ],
      [
        { text: "$ aicat gate --threshold 90", kind: "prompt" },
        { text: "gate_agent \u2192 running compliance and tests\u2026", kind: "sys" },
        { text: "phase: quality_gate", kind: "stage" },
        { text: "\u2713 build success", kind: "ok" },
        { text: "\u2713 contracts matched", kind: "ok" },
        { text: "\u2713 security scan passed", kind: "ok" },
        { text: "release/scorecard.html generated", kind: "file" },
        { text: "score: 98 / 100", kind: "done" },
      ],
    ];
  }
  return [
    [
      { text: "$ aicat run \"Build a REST API\" --backend codex", kind: "prompt" },
      { text: "router \u2192 intent=build_api   workflow=full_team", kind: "sys" },
      { text: "phase: docs_confirm", kind: "stage" },
      { text: "output/plan.json written", kind: "file" },
      { text: "output/verify.jsonl written", kind: "file" },
      { text: "scheduler \u2192 dispatching 6 reviewers in parallel", kind: "sys" },
      { text: "scheduler \u2192 dispatching 3 writers (single-writer lock)", kind: "sys" },
      { text: "quality_gate: lint 0 \u00b7 vet 0 \u00b7 tests 0", kind: "ok" },
      { text: "release/proof-pack.zip generated", kind: "file" },
      { text: "Delivery complete.", kind: "done" },
    ],
    [
      { text: "$ aicat continue --phase docs", kind: "prompt" },
      { text: "doc_agent \u2192 syncing context from PRD\u2026", kind: "sys" },
      { text: "phase: frontend_implement", kind: "stage" },
      { text: "output/uiux.md read", kind: "file" },
      { text: "frontend_writer \u2192 building React components\u2026", kind: "sys" },
      { text: "lint: 0 errors, 0 warnings", kind: "ok" },
      { text: "status: pending_backend", kind: "done" },
    ],
    [
      { text: "$ aicat gate --threshold 90", kind: "prompt" },
      { text: "gate_agent \u2192 running compliance and tests\u2026", kind: "sys" },
      { text: "phase: quality_gate", kind: "stage" },
      { text: "\u2713 build success", kind: "ok" },
      { text: "\u2713 contracts matched", kind: "ok" },
      { text: "\u2713 security scan passed", kind: "ok" },
      { text: "release/scorecard.html generated", kind: "file" },
      { text: "score: 98 / 100", kind: "done" },
    ],
  ];
}

function colorClass(kind: LineKind): string {
  switch (kind) {
    case "prompt":
      return "text-cyan";
    case "sys":
      return "text-ink-muted";
    case "stage":
      return "text-gold";
    case "file":
      return "text-cyan-2";
    case "ok":
      return "text-ok";
    case "done":
      return "text-magenta";
    case "dim":
    default:
      return "text-ink-muted2";
  }
}

export function HeroStage() {
  const locale = useLocale();
  const t = useTranslations("home.hero");
  const slides = useMemo(() => slidesFor(locale), [locale]);

  const [slideIdx, setSlideIdx] = useState(0);
  const [step, setStep] = useState(0);

  // re-init when locale changes
  useEffect(() => {
    setStep(0);
  }, [locale]);

  // step typewriter
  useEffect(() => {
    const lines = slides[slideIdx] ?? [];
    if (step >= lines.length) return;
    const delay = step === 0 ? 600 : 280 + Math.random() * 380;
    const id = window.setTimeout(() => setStep((s) => s + 1), delay);
    return () => window.clearTimeout(id);
  }, [step, slideIdx, slides]);

  // slide auto-advance
  useEffect(() => {
    const lines = slides[slideIdx] ?? [];
    if (step < lines.length) return;
    const id = window.setTimeout(() => {
      setSlideIdx((i) => (i + 1) % slides.length);
      setStep(0);
    }, 3200);
    return () => window.clearTimeout(id);
  }, [step, slideIdx, slides]);

  const lines = slides[slideIdx] ?? [];
  const slideTitle = ["01", "02", "03"][slideIdx] ?? "01";
  const slideLabel = ["run", "continue", "gate"][slideIdx] ?? "run";

  return (
    <div className="relative overflow-hidden rounded-2xl border border-cyan-line bg-bg-panel shadow-cyan-glow backdrop-blur-sm">
      <div className="pointer-events-none absolute -inset-px -z-10 rounded-2xl bg-duo opacity-30 blur-2xl" />
      <div className="flex items-center justify-between border-b border-cyan-line bg-bg-2/80 px-4 py-2.5 text-xs">
        <div className="flex items-center gap-1.5">
          <span className="h-3 w-3 rounded-full bg-[#ff5f56]/80" />
          <span className="h-3 w-3 rounded-full bg-[#ffbd2e]/80" />
          <span className="h-3 w-3 rounded-full bg-[#27c93f]/80" />
        </div>
        <div className="flex items-center gap-2 font-mono text-ink-muted2">
          <TerminalIcon className="h-3.5 w-3.5" />
          <span>aicat \u2014 bash \u00b7 zsh</span>
        </div>
        <button
          type="button"
          onClick={() => setStep(0)}
          className="inline-flex items-center gap-1 rounded border border-cyan-line px-2 py-1 font-mono text-[10px] uppercase tracking-wider text-ink-muted hover:border-cyan hover:text-cyan"
          aria-label={t("restart")}
        >
          <RotateCcw className="h-3 w-3" />
          {t("restart")}
        </button>
      </div>

      <div className="flex items-center gap-3 border-b border-cyan-line/60 bg-bg-2/40 px-4 py-2 text-[11px] uppercase tracking-[0.18em] text-ink-muted2">
        <span className="font-display text-cyan">{slideTitle}</span>
        <span className="text-ink-muted">/</span>
        <span>slide \u2014 phase: {slideLabel}</span>
      </div>

      <pre className="overflow-x-auto p-5 font-mono text-[13px] leading-7 text-ink">
        {lines.slice(0, step).map((line, i) => (
          <div key={`${slideIdx}-${i}`} className={colorClass(line.kind)}>
            {line.kind === "prompt" ? <span className="select-none text-magenta">$ </span> : null}
            {line.text.replace(/^\$ /, "")}
          </div>
        ))}
        {step < lines.length ? (
          <div className="mt-1 inline-flex items-center gap-1 text-cyan">
            <span className="inline-block h-3.5 w-2 animate-cursor-blink bg-cyan align-middle" />
          </div>
        ) : null}
      </pre>

      <div className="flex items-center justify-between border-t border-cyan-line/60 bg-bg-2/40 px-4 py-2 text-[10px] uppercase tracking-[0.22em] text-ink-muted2">
        <span>step {Math.min(step, lines.length)} / {lines.length}</span>
        <span>{step >= lines.length ? "done" : "typing\u2026"}</span>
      </div>
    </div>
  );
}
