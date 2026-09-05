import type { Config } from "tailwindcss";

const config: Config = {
  content: [
    "./app/**/*.{ts,tsx}",
    "./components/**/*.{ts,tsx}",
  ],
  darkMode: "class",
  theme: {
    extend: {
      colors: {
        bg: {
          DEFAULT: "#050508",
          2: "#090a10",
          panel: "rgba(5,8,14,0.88)",
        },
        ink: {
          DEFAULT: "#f8fcff",
          muted: "#aebfd0",
          muted2: "#768a9c",
        },
        cyan: {
          DEFAULT: "#00d2ff",
          2: "#00aeff",
          soft: "rgba(0,210,255,0.16)",
          line: "rgba(0,210,255,0.20)",
          strong: "rgba(0,210,255,0.45)",
        },
        magenta: {
          DEFAULT: "#ff2a85",
          2: "#c1176b",
          soft: "rgba(255,42,133,0.16)",
          line: "rgba(255,42,133,0.32)",
        },
        ok: "#20e3b2",
        gold: "#f2b84b",
      },
      fontFamily: {
        sans: ["Manrope Variable", "ui-sans-serif", "system-ui", "sans-serif"],
        display: ["Space Grotesk Variable", "ui-sans-serif", "system-ui", "sans-serif"],
        mono: ["JetBrains Mono Variable", "ui-monospace", "SFMono-Regular", "Menlo", "monospace"],
      },
      backgroundImage: {
        duo: "linear-gradient(120deg,#00d2ff 0%,#3a7bd5 50%,#ff2a85 100%)",
        "duo-soft": "linear-gradient(120deg,rgba(0,210,255,0.16),rgba(255,42,133,0.16))",
        "grid-dark":
          "linear-gradient(rgba(0,210,255,0.06) 1px,transparent 1px),linear-gradient(90deg,rgba(0,210,255,0.06) 1px,transparent 1px)",
      },
      backgroundSize: {
        grid: "32px 32px",
      },
      boxShadow: {
        "cyan-glow": "0 24px 64px -28px rgba(0,210,255,0.55)",
        "magenta-glow": "0 24px 64px -28px rgba(255,42,133,0.55)",
        duo: "0 24px 64px -28px rgba(0,210,255,0.55),0 24px 64px -28px rgba(255,42,133,0.35)",
      },
      keyframes: {
        "cursor-blink": {
          "0%, 49%": { opacity: "1" },
          "50%, 100%": { opacity: "0" },
        },
        "scroll-cue": {
          "0%, 100%": { transform: "translateY(0) scale(1)", opacity: "0.85" },
          "50%": { transform: "translateY(6px) scale(1.06)", opacity: "1" },
        },
        "pulse-soft": {
          "0%, 100%": { opacity: "0.7" },
          "50%": { opacity: "1" },
        },
        "fade-in-up": {
          "0%": { opacity: "0", transform: "translateY(12px)" },
          "100%": { opacity: "1", transform: "translateY(0)" },
        },
      },
      animation: {
        "cursor-blink": "cursor-blink 1s step-end infinite",
        "scroll-cue": "scroll-cue 1.6s ease-in-out infinite",
        "pulse-soft": "pulse-soft 2.4s ease-in-out infinite",
        "fade-in-up": "fade-in-up 0.6s ease-out both",
      },
    },
  },
  plugins: [],
};

export default config;