import { ReactNode } from "react";
import { cn } from "@/lib/utils";

export function Card({
  children,
  className = "",
}: {
  children: ReactNode;
  className?: string;
}) {
  return (
    <div
      className={cn(
        "rounded-xl border border-cyan-line bg-bg-panel p-6 backdrop-blur-sm transition-all duration-300 hover:-translate-y-0.5 hover:border-cyan/60 hover:shadow-cyan-glow",
        className,
      )}
    >
      {children}
    </div>
  );
}