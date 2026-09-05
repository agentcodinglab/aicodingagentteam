import { ChevronDown } from "lucide-react";

type Props = { label?: string };

export function ScrollCue({ label = "Scroll" }: Props) {
  return (
    <a
      href="#content"
      aria-label={label}
      className="group inline-flex flex-col items-center gap-2 text-xs uppercase tracking-[0.22em] text-ink-muted2 transition-colors hover:text-cyan"
    >
      <span className="relative flex h-10 w-6 items-start justify-center rounded-full border border-cyan-line bg-bg/40 pt-2">
        <span className="block h-2 w-1 animate-scroll-cue rounded-full bg-gradient-to-b from-cyan to-magenta" />
      </span>
      <span className="opacity-70 transition-opacity group-hover:opacity-100">{label}</span>
      <ChevronDown className="h-3.5 w-3.5 animate-scroll-cue text-cyan" aria-hidden />
    </a>
  );
}