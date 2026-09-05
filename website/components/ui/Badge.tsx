import { ReactNode } from "react";
import { cn } from "@/lib/utils";

export function Badge({
  children,
  className = "",
}: {
  children: ReactNode;
  className?: string;
}) {
  return (
    <span
      className={cn(
        "inline-flex items-center rounded-full border border-cyan-line bg-bg-2/70 px-3 py-1 text-xs font-medium uppercase text-ink-muted backdrop-blur",
        className,
      )}
    >
      {children}
    </span>
  );
}