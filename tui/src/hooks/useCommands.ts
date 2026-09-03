import { useCallback } from "react";
import { useStore } from "../store.js";
import type { ClientInterface } from "../grpc/client.js";

const DEFAULT_NODES = [
  { id: "n1-clarify", phase: "clarify", role: "coordinator", status: "pending" as const, artifacts: [] as string[] },
  { id: "n2-research", phase: "research", role: "coordinator", status: "pending" as const, artifacts: [] as string[] },
  { id: "n3-prd", phase: "docs", role: "pm", status: "pending" as const, artifacts: [] as string[] },
  { id: "n4-arch", phase: "docs", role: "architect", status: "pending" as const, artifacts: [] as string[] },
  { id: "n5-spec", phase: "spec", role: "coordinator", status: "pending" as const, artifacts: [] as string[] },
  { id: "n6-frontend", phase: "frontend", role: "frontend", status: "pending" as const, artifacts: [] as string[] },
  { id: "n7-backend", phase: "backend", role: "backend", status: "pending" as const, artifacts: [] as string[] },
  { id: "n8-quality", phase: "quality", role: "qa", status: "pending" as const, artifacts: [] as string[] },
  { id: "n9-delivery", phase: "delivery", role: "coordinator", status: "pending" as const, artifacts: [] as string[] },
];

export function useCommands(client: ClientInterface | null) {
  const store = useStore();

  const execute = useCallback(
    async (input: string) => {
      if (!client) {
        store.setProgress(0, "Connecting to coordinator...");
        return;
      }

      const parts = input.trim().split(/\s+/);
      const cmd = parts[0];
      const arg = parts.slice(1).join(" ");

      switch (cmd) {
        case "/run": {
          if (!arg) {
            store.setProgress(0, "Usage: /run <requirement>");
            return;
          }
          store.reset();
          store.setRunning(true);
          store.setNodes(DEFAULT_NODES.map((n) => ({ ...n })));
          store.setCurrentTask("clarify", "coordinator");
          store.setProgress(10, "Starting pipeline...");
          store.addAuditEntry({
            ts: new Date().toISOString(),
            type: "pipeline.start",
            agent: "coordinator",
            task: arg,
            result: "start",
          });

          try {
            const phases = ["clarify", "research", "docs", "spec", "frontend", "backend", "quality", "delivery"];
            const roles = ["coordinator", "coordinator", "pm", "coordinator", "frontend", "backend", "qa", "coordinator"];
            for (let i = 0; i < phases.length; i++) {
              store.updateNodeStatus(DEFAULT_NODES[i].id, "running");
              store.setCurrentTask(phases[i], roles[i]);
              store.setProgress(10 + (i * 80) / phases.length, `Phase: ${phases[i]}`);
              await new Promise((r) => setTimeout(r, 300));
              store.updateNodeStatus(DEFAULT_NODES[i].id, "completed");
              store.addAuditEntry({
                ts: new Date().toISOString(),
                type: `phase.${phases[i]}`,
                agent: roles[i],
                task: DEFAULT_NODES[i].id,
                result: "pass",
              });
            }

            const resp = await client.runPipeline({
              requirement: arg,
              backend: "codex",
              autoApproveGates: false,
            });

            store.setPlanId(resp.planId);
            store.setResult(resp.score, resp.passed, [], [], resp.artifacts);
            store.setProgress(100, "Pipeline completed");
            store.addAuditEntry({
              ts: new Date().toISOString(),
              type: "pipeline.done",
              agent: "coordinator",
              task: resp.planId,
              result: resp.passed ? "pass" : "fail",
            });
          } catch (err: any) {
            store.setResult(0, false, [err.message], [], []);
            store.setProgress(0, `Error: ${err.message}`);
          } finally {
            store.setRunning(false);
          }
          break;
        }

        case "/quick": {
          if (!arg) {
            store.setProgress(0, "Usage: /quick <description>");
            return;
          }
          store.setRunning(true);
          store.setProgress(50, "Quick editing...");
          try {
            const resp = await client.quickEdit({ description: arg, backend: "codex" });
            store.setResult(resp.passed ? 100 : 0, resp.passed, [], [], resp.filesChanged);
            store.setProgress(100, "Done");
          } catch (err: any) {
            store.setResult(0, false, [err.message], [], []);
          } finally {
            store.setRunning(false);
          }
          break;
        }

        case "/verify": {
          store.setRunning(true);
          store.setProgress(50, "Running quality gate...");
          try {
            const resp = await client.verify({ runtime: true });
            store.setResult(resp.score, resp.passed, resp.blocking, resp.advisory, []);
            store.setProgress(100, `Score: ${resp.score}`);
          } catch (err: any) {
            store.setResult(0, false, [err.message], [], []);
          } finally {
            store.setRunning(false);
          }
          break;
        }

        case "/plan": {
          try {
            const resp = await client.getPlan();
            store.setPlanId(resp.planJson ? "loaded" : "empty");
            store.setProgress(100, `Plan loaded: ${resp.nodeCount} nodes`);
          } catch (err: any) {
            store.setProgress(0, `Error: ${err.message}`);
          }
          break;
        }

        case "/continue": {
          if (!store.planId) {
            store.setProgress(0, "No plan to continue");
            return;
          }
          try {
            const resp = await client.continuePlan({ planId: store.planId });
            store.setProgress(100, `Resumed: ${resp.status}`);
          } catch (err: any) {
            store.setProgress(0, `Error: ${err.message}`);
          }
          break;
        }

        case "/backend": {
          store.setProgress(100, `Backend switched to: ${arg}`);
          break;
        }

        case "/report": {
          store.setProgress(100, `Score: ${store.score} | Blocking: ${store.blocking.join(", ") || "none"} | Advisory: ${store.advisory.join(", ") || "none"}`);
          break;
        }

        case "/exit": {
          process.exit(0); break;
        }

        default: {
          store.setProgress(0, `Unknown command: ${cmd}. See commands panel.`);
        }
      }
    },
    [client, store]
  );

  return { execute };
}
