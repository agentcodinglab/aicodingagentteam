import { describe, it, expect, vi, beforeEach } from "vitest";
import { renderHook, act } from "@testing-library/react";
import { useStore } from "../store.js";
import { useCommands } from "../hooks/useCommands.js";
import type { ClientInterface } from "../grpc/client.js";

function mockClient(overrides: Partial<ClientInterface> = {}): ClientInterface {
  return {
    runPipeline: vi.fn().mockResolvedValue({
      planId: "plan-test",
      artifacts: ["output/app.go"],
      score: 100,
      passed: true,
    }),
    quickEdit: vi.fn().mockResolvedValue({
      filesChanged: ["src/main.go"],
      passed: true,
      score: 100,
    }),
    verify: vi.fn().mockResolvedValue({
      score: 90,
      passed: true,
      blocking: [],
      advisory: ["advisory-1"],
    }),
    getPlan: vi.fn().mockResolvedValue({
      planJson: "{}",
      nodeCount: 5,
    }),
    continuePlan: vi.fn().mockResolvedValue({
      resumed: true,
      status: "resumed",
    }),
    ...overrides,
  } as unknown as ClientInterface;
}

describe("useCommands", () => {
  beforeEach(() => {
    useStore.getState().reset();
  });

  it("returns an execute function", () => {
    const { result } = renderHook(() => useCommands(mockClient()));
    expect(typeof result.current.execute).toBe("function");
  });

  it("/run executes pipeline and sets result", async () => {
    const client = mockClient();
    const { result } = renderHook(() => useCommands(client));
    await act(async () => {
      await result.current.execute("/run build a todo app");
    });
    const state = useStore.getState();
    expect(state.planId).toBe("plan-test");
    expect(state.score).toBe(100);
    expect(state.passed).toBe(true);
    expect(state.running).toBe(false);
  });

  it("/run without argument shows usage", async () => {
    const client = mockClient();
    const { result } = renderHook(() => useCommands(client));
    await act(async () => {
      await result.current.execute("/run");
    });
    expect(useStore.getState().message).toContain("Usage");
  });

  it("/quick executes quick edit", async () => {
    const client = mockClient();
    const { result } = renderHook(() => useCommands(client));
    await act(async () => {
      await result.current.execute("/quick fix typo");
    });
    const state = useStore.getState();
    expect(state.passed).toBe(true);
    expect(state.artifacts).toEqual(["src/main.go"]);
  });

  it("/verify runs quality gate", async () => {
    const client = mockClient();
    const { result } = renderHook(() => useCommands(client));
    await act(async () => {
      await result.current.execute("/verify");
    });
    const state = useStore.getState();
    expect(state.score).toBe(90);
    expect(state.passed).toBe(true);
    expect(state.advisory).toEqual(["advisory-1"]);
  });

  it("/plan loads plan info", async () => {
    const client = mockClient();
    const { result } = renderHook(() => useCommands(client));
    await act(async () => {
      await result.current.execute("/plan");
    });
    expect(useStore.getState().message).toContain("5 nodes");
  });

  it("/backend switches backend", async () => {
    const client = mockClient();
    const { result } = renderHook(() => useCommands(client));
    await act(async () => {
      await result.current.execute("/backend opencode");
    });
    expect(useStore.getState().message).toContain("opencode");
  });

  it("/report shows current score", async () => {
    useStore.getState().setResult(85, true, ["block-1"], [], []);
    const client = mockClient();
    const { result } = renderHook(() => useCommands(client));
    await act(async () => {
      await result.current.execute("/report");
    });
    expect(useStore.getState().message).toContain("85");
  });

  it("unknown command shows error", async () => {
    const client = mockClient();
    const { result } = renderHook(() => useCommands(client));
    await act(async () => {
      await result.current.execute("/nonexistent");
    });
    expect(useStore.getState().message).toContain("Unknown command");
  });

  it("/verify error is handled", async () => {
    const client = mockClient({
      verify: vi.fn().mockRejectedValue(new Error("connection refused")),
    });
    const { result } = renderHook(() => useCommands(client));
    await act(async () => {
      await result.current.execute("/verify");
    });
    const state = useStore.getState();
    expect(state.passed).toBe(false);
    expect(state.blocking).toContain("connection refused");
  });
});