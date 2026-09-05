# Plan: Multi-language README + v0.2.0 release

## Goal
1. Rename current `README.md` -> `README.zh.md` (中文版)
2. Generate English `README.md` (canonical, default)
3. Add 7 more locales (ja/ko/fr/de/ru/es/it) each as `README.<locale>.md`
4. Every README must have a language picker at top linking all 8 languages
5. Push, tag `v0.2.0`, push tag to trigger goreleaser release

## Why this order
- Renaming `README.md` to `README.zh.md` happens AFTER we have a new English `README.md` ready to replace it; otherwise the repo would have no default README between commits.
- Actually: safer sequence is to commit the 8 README files in one atomic commit (delete old + add 8 new) so Git history stays clean and a single push ships all languages.
- Tag must be created and pushed AFTER the README commit lands, so the released artifacts bundle the right README.

## Step-by-step

### Step 1 — Create README templates
For each locale, produce a README that is the **same structure** (translated), and includes a "Languages" block at top:

```
## 🌐 Languages / 语言
[English](README.md) · [中文](README.zh.md) · [日本語](README.ja.md) · [한국어](README.ko.md) · [Français](README.fr.md) · [Deutsch](README.de.md) · [Русский](README.ru.md) · [Español](README.es.md) · [Italiano](README.it.md)
```

Localize the **narrative text** (title, tagline, intro, features, quick start, command table, project structure, docs links) but keep:
- All CLI commands in English (code blocks, command names, flags)
- All file paths / URLs unchanged
- Badge URLs unchanged
- Code blocks unchanged

8 files total:
- `README.md` (English) — replaces current
- `README.zh.md` (Chinese) — moved from current `README.md`
- `README.ja.md` (Japanese)
- `README.ko.md` (Korean)
- `README.fr.md` (French)
- `README.de.md` (German)
- `README.ru.md` (Russian)
- `README.es.md` (Spanish)
- `README.it.md` (Italian)

### Step 2 — Update `docs/CONSTRAINTS.md` doc nav (if any)
Not required (doc nav lives inside README only).

### Step 3 — Commit & push
Single commit:
```
docs: 8-locale README set (en/zh/ja/ko/fr/de/ru/es/it) with cross-links

- README.md: English canonical
- README.zh.md: Chinese
- README.ja.md: Japanese
- README.ko.md: Korean
- README.fr.md: French
- README.de.md: German
- README.ru.md: Russian
- README.es.md: Spanish
- README.it.md: Italian

Each README includes a "Languages" picker at the top.
```

### Step 4 — Tag & push v0.2.0
```
git tag -a v0.2.0 -m "v0.2.0: J+K+L release (E2E + security hardening + CLI/TUI tests)"
git push origin v0.2.0
```

### Step 5 — Watch release workflow
```
gh run list --limit 1 --workflow release.yml
gh run watch <id> --exit-status
```

## Acceptance criteria
- [ ] 8 README files present on main, each with the 8-language picker at top
- [ ] Single git commit for all 8 files (or 2: rename + add new languages)
- [ ] `git tag v0.2.0` exists locally and pushed to origin
- [ ] GitHub Actions release workflow triggers and produces cross-platform artifacts
- [ ] All artifacts include the localized README files

## Risks
- Translation quality: human-translated where possible; for languages I cannot natively verify, I will use professional-tone machine translation that is technically faithful (preserve all code, paths, commands).
- File encoding: PowerShell WriteAllText UTF-8 no BOM required (per project rule).
- goreleaser `files:` in `.goreleaser.yml` only includes `README.md` (English). Other locales are not bundled in archives by default. We may want to add `README.*.md` to the archive files list to ship all languages with binary releases — propose adding this in the same commit.