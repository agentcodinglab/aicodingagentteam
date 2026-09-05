#!/usr/bin/env python3
"""Stage 2: add new i18n keys for dark duotone redesign."""
import json, os
ROOT = r"E:\javaproject\my\2026\agent_team\website\messages"

# titlePrefix / titleHighlight / titleSuffix compose the h1 with the
# middle word highlighted via text-duo (cyan->magenta gradient).
# Keep titlePrefix+titleSuffix combined as the original full title so that
# both splitting and concatenation read naturally.
TITLE = {
    "en": {
        "titlePrefix": "Coordinate a ",
        "titleHighlight": "9-role software team",
        "titleSuffix": " with AI coding CLIs.",
        "meta1": "9 roles",
        "meta2": "4 protocols",
        "meta3": "0 keys held",
        "restart": "restart",
        "scrollCue": "scroll",
    },
    "zh": {
        "titlePrefix": "\u7528 AI \u7f16\u7801 CLI \u8c03\u5ea6\u4e00\u652f",
        "titleHighlight": "9 \u89d2\u8272\u8f6f\u4ef6\u56e2\u961f",
        "titleSuffix": "\u3002",
        "meta1": "9 \u4e2a\u89d2\u8272",
        "meta2": "4 \u5957\u534f\u8bae",
        "meta3": "\u4e0d\u6301\u5bc6\u94a5",
        "restart": "\u91cd\u65b0\u64ad\u653e",
        "scrollCue": "\u4e0b\u62c9",
    },
    "ja": {
        "titlePrefix": "AI \u30b3\u30fc\u30c7\u30a3\u30f3\u30b0 CLI \u3067",
        "titleHighlight": "9 \u30ed\u30fc\u30eb\u958b\u767a\u30c1\u30fc\u30e0",
        "titleSuffix": "\u3092\u7de0\u51fa\u3057\u307e\u3059\u3002",
        "meta1": "9 \u30ed\u30fc\u30eb",
        "meta2": "4 \u30d7\u30ed\u30c8\u30b3\u30eb",
        "meta3": "\u9375\u3092\u4fdd\u6301\u305b\u305a",
        "restart": "\u518d\u751f",
        "scrollCue": "\u30b9\u30af\u30ed\u30fc\u30eb",
    },
    "ko": {
        "titlePrefix": "AI \ucf54\ub529 CLI\ub85c ",
        "titleHighlight": "9\uac1c \uc5ed\ud558\uc758 \uac1c\ubc1c \ud300",
        "titleSuffix": "\uc744 \uc624\ucf08\uc2a4\ud2b8\ub808\uc774\ud2b8\ud558\uc2ed\uc2dc\uc624.",
        "meta1": "9\uac1c \uc5ed\ud558",
        "meta2": "4\uac1c \ud504\ub85c\ud1a0\ucf5c",
        "meta3": "\ud0a4 \ubbf8\ubcf4\uad00",
        "restart": "\uc7ac\uc0dd",
        "scrollCue": "\uc2a4\ud06c\ub864",
    },
    "fr": {
        "titlePrefix": "Orchestrez une ",
        "titleHighlight": "\u00e9quipe logicielle \u00e0 9 r\u00f4les",
        "titleSuffix": " avec des CLI IA.",
        "meta1": "9 r\u00f4les",
        "meta2": "4 protocoles",
        "meta3": "0 cl\u00e9 conserv\u00e9e",
        "restart": "relancer",
        "scrollCue": "d\u00e9filer",
    },
    "de": {
        "titlePrefix": "Orchestrieren Sie ein ",
        "titleHighlight": "9-Rollen-Entwicklungsteam",
        "titleSuffix": " mit KI-Coding-CLIs.",
        "meta1": "9 Rollen",
        "meta2": "4 Protokolle",
        "meta3": "0 Schl\u00fcssel",
        "restart": "neustart",
        "scrollCue": "scrollen",
    },
    "ru": {
        "titlePrefix": "\u041e\u0440\u043a\u0435\u0441\u0442\u0440\u0438\u0440\u0443\u0439\u0442\u0435 ",
        "titleHighlight": "\u043a\u043e\u043c\u0430\u043d\u0434\u0443 \u0438\u0437 9 \u0440\u043e\u043b\u0435\u0439",
        "titleSuffix": " \u0441 \u043f\u043e\u043c\u043e\u0449\u044c\u044e AI-CLI.",
        "meta1": "9 \u0440\u043e\u043b\u0435\u0439",
        "meta2": "4 \u043f\u0440\u043e\u0442\u043e\u043a\u043e\u043b\u0430",
        "meta3": "0 \u043a\u043b\u044e\u0447\u0435\u0439",
        "restart": "\u043f\u0435\u0440\u0435\u0437\u0430\u043f\u0443\u0441\u0442\u0438\u0442\u044c",
        "scrollCue": "\u043f\u0440\u043e\u043a\u0440\u0443\u0442\u0438\u0442\u044c",
    },
    "es": {
        "titlePrefix": "Orquesta un ",
        "titleHighlight": "equipo de software de 9 roles",
        "titleSuffix": " con CLI de IA.",
        "meta1": "9 roles",
        "meta2": "4 protocolos",
        "meta3": "0 claves",
        "restart": "reiniciar",
        "scrollCue": "desplazar",
    },
    "it": {
        "titlePrefix": "Orchestra un ",
        "titleHighlight": "team di sviluppo a 9 ruoli",
        "titleSuffix": " con CLI IA.",
        "meta1": "9 ruoli",
        "meta2": "4 protocolli",
        "meta3": "0 chiavi",
        "restart": "riavvia",
        "scrollCue": "scorri",
    },
}

COMMON = {
    "en": {"backToHome": "Back to home"},
    "zh": {"backToHome": "\u8fd4\u56de\u9996\u9875"},
    "ja": {"backToHome": "\u30db\u30fc\u30e0\u3078"},
    "ko": {"backToHome": "\ud648\uc73c\ub85c"},
    "fr": {"backToHome": "Retour \u00e0 l\u2019accueil"},
    "de": {"backToHome": "Zur\u00fcck zur Startseite"},
    "ru": {"backToHome": "\u041d\u0430 \u0433\u043b\u0430\u0432\u043d\u0443\u044e"},
    "es": {"backToHome": "Volver al inicio"},
    "it": {"backToHome": "Torna alla home"},
}

for loc in TITLE:
    p = os.path.join(ROOT, f"{loc}.json")
    d = json.load(open(p, "r", encoding="utf-8"))
    h = d.setdefault("home", {}).setdefault("hero", {})
    for k, v in TITLE[loc].items():
        h[k] = v
    sc = d.setdefault("home", {}).setdefault("scrollCue", TITLE[loc]["scrollCue"])
    c = d.setdefault("common", {})
    for k, v in COMMON[loc].items():
        c[k] = v
    open(p, "wb").write(json.dumps(d, ensure_ascii=False, indent=2).encode("utf-8") + b"\n")
    print("OK:", loc)