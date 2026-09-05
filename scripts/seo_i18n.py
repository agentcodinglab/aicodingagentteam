#!/usr/bin/env python3
"""Add changelog nav/doc/footer i18n keys to all 9 locales."""
import json, os
ROOT = r"E:\javaproject\my\2026\agent_team\website\messages"

DATA = {
    "en": {"nav": {"changelog": "Changelog"}, "docs": {"nav": {"changelog": "Changelog"}}, "footer": {"links": {"changelog": "Changelog"}}},
    "zh": {"nav": {"changelog": "版本记录"}, "docs": {"nav": {"changelog": "版本记录"}}, "footer": {"links": {"changelog": "版本记录"}}},
    "ja": {"nav": {"changelog": "変更履歴"}, "docs": {"nav": {"changelog": "変更履歴"}}, "footer": {"links": {"changelog": "変更履歴"}}},
    "ko": {"nav": {"changelog": "변경 이력"}, "docs": {"nav": {"changelog": "변경 이력"}}, "footer": {"links": {"changelog": "변경 이력"}}},
    "fr": {"nav": {"changelog": "Journal des versions"}, "docs": {"nav": {"changelog": "Journal des versions"}}, "footer": {"links": {"changelog": "Journal des versions"}}},
    "de": {"nav": {"changelog": "Änderungsprotokoll"}, "docs": {"nav": {"changelog": "Änderungsprotokoll"}}, "footer": {"links": {"changelog": "Änderungsprotokoll"}}},
    "ru": {"nav": {"changelog": "Журнал изменений"}, "docs": {"nav": {"changelog": "Журнал изменений"}}, "footer": {"links": {"changelog": "Журнал изменений"}}},
    "es": {"nav": {"changelog": "Registro de cambios"}, "docs": {"nav": {"changelog": "Registro de cambios"}}, "footer": {"links": {"changelog": "Registro de cambios"}}},
    "it": {"nav": {"changelog": "Registro modifiche"}, "docs": {"nav": {"changelog": "Registro modifiche"}}, "footer": {"links": {"changelog": "Registro modifiche"}}},
}

def deep_merge(a, b):
    for k, v in b.items():
        if isinstance(v, dict) and isinstance(a.get(k), dict):
            deep_merge(a[k], v)
        else:
            a[k] = v
    return a

for loc, add in DATA.items():
    p = os.path.join(ROOT, f"{loc}.json")
    d = json.load(open(p, "r", encoding="utf-8"))
    deep_merge(d, add)
    open(p, "wb").write(json.dumps(d, ensure_ascii=False, indent=2).encode("utf-8") + b"\n")
    print("OK:", loc)