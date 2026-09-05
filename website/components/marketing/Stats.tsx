import { useTranslations } from "next-intl";
import { Reveal } from "../ui/Reveal";

const KEYS = ["roles", "protocols", "locales", "keys"] as const;

export function Stats() {
  const t = useTranslations("home.stats");
  return (
    <Reveal>
      <div className="grid grid-cols-2 gap-px overflow-hidden rounded-2xl border border-cyan-line bg-bg-2/40 sm:grid-cols-4">
        {KEYS.map((k, i) => (
          <div
            key={k}
            className="flex flex-col items-center justify-center gap-1 bg-bg/60 px-6 py-10 text-center"
          >
            <span
              className="font-display text-5xl font-bold tracking-tight text-duo"
              style={{ fontVariationSettings: "'wght' 700" } as React.CSSProperties}
            >
              {t(`${k}.value`)}
            </span>
            <span className="text-xs font-medium uppercase tracking-[0.18em] text-ink-muted2">
              {t(`${k}.label`)}
            </span>
            <span className="mt-1 font-mono text-[10px] text-ink-muted2/60">
              0x{i}
            </span>
          </div>
        ))}
      </div>
    </Reveal>
  );
}