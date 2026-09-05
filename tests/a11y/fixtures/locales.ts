// fixtures/locales.ts - single source of truth for the 9 supported locales.

export type Locale = {
  code: string;
  label: string;
};

export const LOCALES: readonly Locale[] = [
  { code: 'en', label: 'English' },
  { code: 'zh', label: '简体中文' },
  { code: 'ja', label: '日本語' },
  { code: 'ko', label: '한국어' },
  { code: 'fr', label: 'Français' },
  { code: 'de', label: 'Deutsch' },
  { code: 'ru', label: 'Русский' },
  { code: 'es', label: 'Español' },
  { code: 'it', label: 'Italiano' },
] as const;
