import { ReactNode } from "react";
import { Reveal } from "../ui/Reveal";

type SectionProps = {
  title: string;
  subtitle?: string;
  children: ReactNode;
  alt?: boolean;
  eyebrow?: string;
  id?: string;
};

export function Section({
  title,
  subtitle,
  children,
  alt,
  eyebrow,
  id,
}: SectionProps) {
  return (
    <section
      id={id}
      className={`relative border-b border-cyan-line/60 py-24 ${
        alt ? "bg-bg-2/40" : "bg-bg"
      }`}
    >
      <div className="container mx-auto max-w-6xl px-6">
        <Reveal>
          <div className="mx-auto mb-12 max-w-3xl text-center">
            {eyebrow ? (
              <p className="mb-3 font-mono text-xs font-semibold uppercase tracking-[0.22em] text-cyan">
                {eyebrow}
              </p>
            ) : null}
            <h2 className="font-display text-3xl font-bold tracking-tight text-ink sm:text-4xl">
              {title}
            </h2>
            {subtitle ? (
              <p className="mt-4 text-base leading-7 text-ink-muted">{subtitle}</p>
            ) : null}
          </div>
        </Reveal>
        {children}
      </div>
    </section>
  );
}