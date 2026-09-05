"use client";

import Link from "next/link";
import { useLocale } from "next-intl";
import { ComponentProps, ReactNode } from "react";

type Props = Omit<ComponentProps<typeof Link>, "href"> & {
  href: string;
  children: ReactNode;
};

export function LocaleLink({ href, children, ...rest }: Props) {
  const locale = useLocale();
  const isExternal = href.startsWith("http") || href.startsWith("//");
  const localizedHref = isExternal || !href.startsWith("/")
    ? href
    : `/${locale}${href}`;
  return (
    <Link href={localizedHref} {...rest}>
      {children}
    </Link>
  );
}