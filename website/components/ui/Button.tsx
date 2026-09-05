import Link from "next/link";
import { ReactNode } from "react";
import { cn } from "@/lib/utils";

type Variant = "primary" | "secondary" | "ghost" | "duo";

export function Button({
  href,
  children,
  variant = "primary",
  className = "",
  external,
}: {
  href: string;
  children: ReactNode;
  variant?: Variant;
  className?: string;
  external?: boolean;
}) {
  const base =
    "inline-flex items-center justify-center rounded-md px-4 py-2 text-sm font-medium transition-all duration-200 focus:outline-none focus-visible:ring-2 focus-visible:ring-cyan focus-visible:ring-offset-2 focus-visible:ring-offset-bg";
  const variants: Record<Variant, string> = {
    primary:
      "bg-cyan text-bg hover:bg-cyan/90 shadow-[0_18px_36px_-18px_rgba(0,210,255,0.55)]",
    secondary:
      "border border-cyan-line bg-bg-2/60 text-ink hover:border-cyan hover:text-cyan",
    ghost: "text-ink-muted hover:bg-bg-2/60 hover:text-cyan",
    duo: "text-bg shadow-duo hover:brightness-110 [background:var(--duo)]",
  };
  const cls = cn(base, variants[variant], className);
  if (external) {
    return (
      <a href={href} target="_blank" rel="noreferrer" className={cls}>
        {children}
      </a>
    );
  }
  return (
    <Link href={href} className={cls}>
      {children}
    </Link>
  );
}