"use client";

import { ReactNode, useEffect, useRef } from "react";
import { cn } from "@/lib/utils";

type TiltProps = {
  children: ReactNode;
  className?: string;
  maxDeg?: number;
};

export function Tilt({ children, className, maxDeg = 1.5 }: TiltProps) {
  const ref = useRef<HTMLDivElement | null>(null);

  useEffect(() => {
    const node = ref.current;
    if (!node) return;

    let raf = 0;
    let rx = 0;
    let ry = 0;
    let tx = 0;
    let ty = 0;

    const tick = () => {
      rx += (tx - rx) * 0.18;
      ry += (ty - ry) * 0.18;
      node.style.transform = `perspective(900px) rotateX(${rx.toFixed(2)}deg) rotateY(${ry.toFixed(2)}deg)`;
      if (Math.abs(tx - rx) > 0.01 || Math.abs(ty - ry) > 0.01) {
        raf = requestAnimationFrame(tick);
      } else {
        raf = 0;
      }
    };

    const onMove = (e: MouseEvent) => {
      const rect = node.getBoundingClientRect();
      const x = (e.clientX - rect.left) / rect.width;
      const y = (e.clientY - rect.top) / rect.height;
      tx = (0.5 - y) * 2 * maxDeg;
      ty = (x - 0.5) * 2 * maxDeg;
      if (!raf) raf = requestAnimationFrame(tick);
    };
    const onLeave = () => {
      tx = 0;
      ty = 0;
      if (!raf) raf = requestAnimationFrame(tick);
    };

    node.addEventListener("mousemove", onMove);
    node.addEventListener("mouseleave", onLeave);
    return () => {
      node.removeEventListener("mousemove", onMove);
      node.removeEventListener("mouseleave", onLeave);
      if (raf) cancelAnimationFrame(raf);
      node.style.transform = "";
    };
  }, [maxDeg]);

  return (
    <div
      ref={ref}
      className={cn("anim-reveal is-visible [transform-style:preserve-3d]", className)}
    >
      {children}
    </div>
  );
}