import { describe, it, expect, beforeEach } from "vitest";
import { useStore } from "../store.js";

describe("store", () => {
  beforeEach(() => {
    useStore.getState().reset();
  });

  it("starts with default state", () => {
    const state = useStore.getState();
    expect(state.connected).toBe(false);
    expect(state.score).toBe(0);
    expect(state.passed).toBe(false);
    expect(state.running).toBe(false);
    expect(state.nodes).toEqual([]);
    expect(state.auditLog).toEqual([]);
  });

  it("setConnected updates connection state", () => {
    useStore.getState().setConnected(true);
    expect(useStore.getState().connected).toBe(true);
  });

  it("setAddress updates address", () => {
    useStore.getState().setAddress("localhost:8080");
    expect(useStore.getState().address).toBe("localhost:8080");
  });

  it("setNodes updates the DAG nodes", () => {
    const nodes = [
      { id: "n1", phase: "clarify", role: "coordinator", status: "pending" as const, artifacts: [] },
    ];
    useStore.getState().setNodes(nodes);
    expect(useStore.getState().nodes).toHaveLength(1);
    expect(useStore.getState().nodes[0].id).toBe("n1");
  });

  it("updateNodeStatus updates a single node", () => {
    useStore.getState().setNodes([
      { id: "n1", phase: "clarify", role: "coordinator", status: "pending" as const, artifacts: [] },
      { id: "n2", phase: "research", role: "coordinator", status: "pending" as const, artifacts: [] },
    ]);
    useStore.getState().updateNodeStatus("n1", "running");
    expect(useStore.getState().nodes[0].status).toBe("running");
    expect(useStore.getState().nodes[1].status).toBe("pending");
  });

  it("setCurrentTask updates phase and role", () => {
    useStore.getState().setCurrentTask("research", "coordinator");
    expect(useStore.getState().currentPhase).toBe("research");
    expect(useStore.getState().currentRole).toBe("coordinator");
  });

  it("setProgress updates progress and message", () => {
    useStore.getState().setProgress(50, "halfway");
    expect(useStore.getState().progress).toBe(50);
    expect(useStore.getState().message).toBe("halfway");
  });

  it("setResult updates quality results", () => {
    useStore.getState().setResult(85, true, ["warn"], [], ["file.go"]);
    expect(useStore.getState().score).toBe(85);
    expect(useStore.getState().passed).toBe(true);
    expect(useStore.getState().blocking).toEqual(["warn"]);
    expect(useStore.getState().artifacts).toEqual(["file.go"]);
  });

  it("addAuditEntry appends and caps at 50", () => {
    for (let i = 0; i < 55; i++) {
      useStore.getState().addAuditEntry({
        ts: new Date().toISOString(),
        type: "test",
        agent: "qa",
        task: `t${i}`,
        result: "pass",
      });
    }
    expect(useStore.getState().auditLog).toHaveLength(50);
    expect(useStore.getState().auditLog[0].task).toBe("t5");
  });

  it("setPlanId updates plan ID", () => {
    useStore.getState().setPlanId("plan-123");
    expect(useStore.getState().planId).toBe("plan-123");
  });

  it("setRunning updates running flag", () => {
    useStore.getState().setRunning(true);
    expect(useStore.getState().running).toBe(true);
  });

  it("reset restores initial state", () => {
    useStore.getState().setConnected(true);
    useStore.getState().setResult(100, true, [], [], []);
    useStore.getState().setNodes([{ id: "x", phase: "x", role: "x", status: "completed", artifacts: [] }]);
    useStore.getState().reset();
    const state = useStore.getState();
    expect(state.connected).toBe(false);
    expect(state.score).toBe(0);
    expect(state.nodes).toEqual([]);
  });
});

