import { DocPage } from "@/components/docs/DocPage";

export default function Page({ params: { locale } }: { params: { locale: string } }) {
  return <DocPage locale={locale} slug="architecture" />;
}
