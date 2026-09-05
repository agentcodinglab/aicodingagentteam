import { useTranslations } from "next-intl";

const KEYS = ["roles", "protocols", "locales", "keys"] as const;

export function Stats() {
  const t = useTranslations("home.stats");
  return (
    <div className="grid grid-cols-2 gap-px overflow-hidden rounded-2xl border border-slate-200 bg-slate-200 dark:border-slate-800 dark:bg-slate-800 sm:grid-cols-4">
      {KEYS.map((k) => (
        <div
          key={k}
          className="flex flex-col items-center justify-center gap-1 bg-white px-6 py-10 text-center dark:bg-slate-900"
        >
          <span className="text-5xl font-bold tracking-tight text-brand-600 dark:text-brand-400">
            {t(`${k}.value`)}
          </span>
          <span className="text-xs font-medium uppercase tracking-wider text-slate-500 dark:text-slate-400">
            {t(`${k}.label`)}
          </span>
        </div>
      ))}
    </div>
  );
}