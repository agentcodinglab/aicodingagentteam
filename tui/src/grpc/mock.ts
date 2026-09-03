import type {
  ClientInterface,
  RunPipelineRequest,
  RunPipelineResponse,
  QuickResponse,
  VerifyResponse,
  PlanResponse,
  ContinueResponse,
} from "./client.js";

const delay = (ms: number) => new Promise((r) => setTimeout(r, ms));

export class MockClient implements ClientInterface {
  async runPipeline(_req: RunPipelineRequest): Promise<RunPipelineResponse> {
    await delay(500);
    return {
      planId: "demo-plan-" + Date.now().toString(36),
      artifacts: [
        "docs/spec/demo-feature.md",
        "docs/plan/demo-feature.md",
        "internal/demo/handler.go",
        "internal/demo/handler_test.go",
      ],
      score: 92,
      passed: true,
    };
  }

  async quickEdit(_req: { description: string; backend: string }): Promise<QuickResponse> {
    await delay(300);
    return {
      filesChanged: ["internal/demo/handler.go", "internal/demo/handler_test.go"],
      passed: true,
    };
  }

  async verify(_req: { runtime: boolean }): Promise<VerifyResponse> {
    await delay(400);
    return {
      score: 88,
      passed: true,
      blocking: [],
      advisory: ["golangci-lint: consider simplifying if-else chain"],
    };
  }

  async getPlan(): Promise<PlanResponse> {
    await delay(200);
    return {
      planJson: JSON.stringify({ nodes: [], gates: [] }),
      nodeCount: 9,
      gates: [
        { id: "g1", afterNode: "n5-spec", type: "human", status: "pending" },
        { id: "g2", afterNode: "n8-quality", type: "auto", status: "pending" },
      ],
    };
  }

  async continuePlan(_req: { planId: string }): Promise<ContinueResponse> {
    await delay(200);
    return { resumed: true, status: "running" };
  }
}