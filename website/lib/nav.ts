export type DocNavItem = { slug: string; titleKey: string };

export const docsNav: DocNavItem[] = [
  { slug: 'requirements', titleKey: 'docs.nav.requirements' },
  { slug: 'architecture', titleKey: 'docs.nav.architecture' },
  { slug: 'implementation-plan', titleKey: 'docs.nav.implementationPlan' },
  { slug: 'quality-constraints', titleKey: 'docs.nav.qualityConstraints' },
  { slug: 'domain-model', titleKey: 'docs.nav.domainModel' },
  { slug: 'changelog', titleKey: 'docs.nav.changelog' },
];

export function getDocBySlug(slug: string): DocNavItem | undefined {
  return docsNav.find((d) => d.slug === slug);
}