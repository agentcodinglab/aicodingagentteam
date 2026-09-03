import * as grpc from "@grpc/grpc-js";
import * as protoLoader from "@grpc/proto-loader";
import path from "path";
import { fileURLToPath } from "url";

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

const PROTO_PATH = path.resolve(__dirname, "../../proto/aicodingagentteam.proto");

export interface RunPipelineRequest {
  requirement: string;
  backend: string;
  autoApproveGates: boolean;
}

export interface RunPipelineResponse {
  planId: string;
  artifacts: string[];
  score: number;
  passed: boolean;
}

export interface ProgressEvent {
  taskId: string;
  phase: string;
  role: string;
  status: string;
  message: string;
}

export interface VerifyResponse {
  score: number;
  passed: boolean;
  blocking: string[];
  advisory: string[];
}

export interface PlanResponse {
  planJson: string;
  nodeCount: number;
  gates: Array<{ id: string; afterNode: string; type: string; status: string }>;
}

export interface QuickResponse {
  filesChanged: string[];
  passed: boolean;
}

export interface ContinueResponse {
  resumed: boolean;
  status: string;
}

export interface ClientInterface {
  runPipeline(req: RunPipelineRequest): Promise<RunPipelineResponse>;
  quickEdit(req: { description: string; backend: string }): Promise<QuickResponse>;
  verify(req: { runtime: boolean }): Promise<VerifyResponse>;
  getPlan(): Promise<PlanResponse>;
  continuePlan(req: { planId: string }): Promise<ContinueResponse>;
}

export class CoordinatorClient implements ClientInterface {
  private client: any;
  private address: string;

  constructor(host?: string, port?: number) {
    const h = host || process.env.AICODINGAGENTTEAM_HOST || "localhost";
    const p = port || parseInt(process.env.AICODINGAGENTTEAM_PORT || "8080", 10);
    this.address = `${h}:${p}`;

    const packageDef = protoLoader.loadSync(PROTO_PATH, {
      keepCase: true,
      longs: String,
      enums: String,
      defaults: true,
      oneofs: true,
    });
    const proto = grpc.loadPackageDefinition(packageDef) as any;
    this.client = new proto.aicodingagentteam.v1.Coordinator(
      this.address,
      grpc.credentials.createInsecure()
    );
  }

  runPipeline(req: RunPipelineRequest): Promise<RunPipelineResponse> {
    return new Promise((resolve, reject) => {
      this.client.runPipeline(req, (err: any, resp: any) => {
        if (err) reject(err);
        else resolve(resp);
      });
    });
  }

  runPipelineStream(req: RunPipelineRequest): grpc.ClientReadableStream<any> {
    return this.client.runPipelineStream(req);
  }

  quickEdit(req: { description: string; backend: string }): Promise<QuickResponse> {
    return new Promise((resolve, reject) => {
      this.client.quickEdit(req, (err: any, resp: any) => {
        if (err) reject(err);
        else resolve(resp);
      });
    });
  }

  verify(req: { runtime: boolean }): Promise<VerifyResponse> {
    return new Promise((resolve, reject) => {
      this.client.verify(req, (err: any, resp: any) => {
        if (err) reject(err);
        else resolve(resp);
      });
    });
  }

  getPlan(): Promise<PlanResponse> {
    return new Promise((resolve, reject) => {
      this.client.getPlan({}, (err: any, resp: any) => {
        if (err) reject(err);
        else resolve(resp);
      });
    });
  }

  continuePlan(req: { planId: string }): Promise<ContinueResponse> {
    return new Promise((resolve, reject) => {
      this.client.continue(req, (err: any, resp: any) => {
        if (err) reject(err);
        else resolve(resp);
      });
    });
  }
}