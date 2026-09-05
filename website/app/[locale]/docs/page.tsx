import { redirect } from "next/navigation";

export default function DocsIndex({ params: { locale } }: { params: { locale: string } }) {
  redirect(`/${locale}/docs/requirements`);
}