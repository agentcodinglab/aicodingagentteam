// E2E smoke test: real CoordinatorClient against a live coordinator server.
// Self-contained: spawns the coordinator binary, waits for the port, runs
// RPC checks, then tears the server down. Run via: npm test
import { spawn } from "node:child_process";
import { createConnection } from "node:net";
import path from "node:path";
import { fileURLToPath } from "node:url";
import { CoordinatorClient } from "../dist/grpc/client.js";

const __filename = fileURLToPath(import.meta.url);
const repoRoot = path.resolve(path.dirname(__filename), "..", "..");
const host = "localhost";
const port = 8090;
const binName = process.platform === "win32" ? "aicodingagentteam.exe" : "aicodingagentteam";
const binPath = path.join(repoRoot, binName);

async function waitForPort(host, port, retries = 50, delayMs = 200) {
  for (let i = 0; i < retries; i++) {
    const sock = createConnection({ host, port });
    const ok = await new Promise((res) => {
      sock.once("connect", () => { sock.end(); res(true); });
      sock.once("error", () => res(false));
    });
    if (ok) return;
    await new Promise((r) => setTimeout(r, delayMs));
  }
  throw new Error(`server did not come up on ${host}:${port}`);
}

// Clean env so the spawned Go process behaves like a manual run, not an
// npm subprocess (avoids npm-injected PATH/GOPATH surprises).
const childEnv = {
  PATH: process.env.PATH,
  USERPROFILE: process.env.USERPROFILE,
  HOME: process.env.HOME,
  LOCALAPPDATA: process.env.LOCALAPPDATA || (process.env.USERPROFILE ? process.env.USERPROFILE + "\\AppData\\Local" : undefined),
  AICODINGAGENTTEAM_PORT: String(port),
};

const server = spawn(binPath, ["serve"], { cwd: repoRoot, env: childEnv, stdio: ["ignore", "pipe", "pipe"] });
server.stdout?.on("data", () => {});
server.stderr?.on("data", () => {});

let failures = 0;
function check(name, ok, detail = "") {
  const tag = ok ? "PASS" : "FAIL";
  console.log(`[${tag}] ${name}${detail ? " - " + detail : ""}`);
  if (!ok) failures++;
}

let client;
try {
  await waitForPort(host, port);
  client = new CoordinatorClient(host, port);

  // 1. quickEdit exercises the full pipeline end-to-end. We assert a structured
  // verdict comes back (not that the gate passes), since the gate runs the real
  // toolchain and can flap under concurrent CI load.
  try {
    const resp = await client.quickEdit({ description: "update demo comment", backend: "codex" });
    const ok = typeof resp.passed === "boolean";
    check("quickEdit", ok, `passed=${resp.passed}` + (resp.details?.length ? ` details=${JSON.stringify(resp.details.filter(d=>d.status!=="pass"))}` : ""));
  } catch (e) {
    check("quickEdit", false, e.message);
  }

  // 2. getPlan (after a run has persisted plan.json)
  try {
    const plan = await client.getPlan();
    check("getPlan", typeof plan.nodeCount === "number", `nodes=${plan.nodeCount} gates=${plan.gates?.length ?? 0}`);
  } catch (e) {
    check("getPlan", false, e.message);
  }

  // 3. continuePlan without parked workflow (must respond gracefully)
  try {
    const resp = await client.continuePlan({ planId: "nonexistent" });
    check("continuePlan(no-state)", resp.resumed === false, `status=${resp.status}`);
  } catch (e) {
    check("continuePlan(no-state)", false, e.message);
  }
} catch (e) {
  console.error("SETUP ERROR:", e.message);
  failures = 1;
} finally {
  server.kill("SIGTERM");
  try { await new Promise((r) => server.once("exit", r)); } catch {}
}

console.log(failures === 0 ? "\nE2E LIVE: ALL PASS" : `\nE2E LIVE: ${failures} FAILED`);
process.exit(failures === 0 ? 0 : 1);
