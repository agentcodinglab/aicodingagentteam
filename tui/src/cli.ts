#!/usr/bin/env node

import React from "react";
import { render } from "ink";
import App from "./app.js";

const args = process.argv.slice(2);
const demo = args.includes("--demo") || process.env.AICODINGAGENTTEAM_DEMO === "1";

// Non-TTY: just print help and exit
if (!process.stdout.isTTY) {
  console.log("AiCodingAgentTeam TUI - Terminal UI client");
  console.log("");
  console.log("Usage: aicodingagentteam-tui [--demo]");
  console.log("");
  console.log("Options:");
  console.log("  --demo    Run with mock data (no coordinator needed)");
  console.log("");
  console.log("Environment:");
  console.log("  AICODINGAGENTTEAM_HOST    Coordinator host (default: localhost)");
  console.log("  AICODINGAGENTTEAM_PORT    Coordinator gRPC port (default: 8080)");
  console.log("  AICODINGAGENTTEAM_DEMO=1  Force demo mode");
  process.exit(0);
}

if (demo) {
  console.log("Running in demo mode (mock data, no live coordinator)");
}

render(React.createElement(App, { demo }));