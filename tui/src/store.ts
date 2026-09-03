import { create } from "zustand";

export interface TaskNode {
  id: string;
  phase: string;
  role: string;
  status: "pending" | "running" | "completed" | "failed" | "parked";
  artifacts: string[];
}

export interface AuditEntry {
  ts: string;
  type: string;
  agent: string;
  task: string;
  result: string;
}

export interface AppState {
  // Connection
  connected: boolean;
  address: string;

  // Current task
  currentPhase: string;
  currentRole: string;

  // DAG
  nodes: TaskNode[];
  planId: string;

  // Progress
  running: boolean;
  progress: number;
  message: string;

  // Results
  score: number;
  passed: boolean;
  blocking: string[];
  advisory: string[];
  artifacts: string[];

  // Audit
  auditLog: AuditEntry[];

  // Actions
  setConnected: (connected: boolean) => void;
  setAddress: (address: string) => void;
  setNodes: (nodes: TaskNode[]) => void;
  updateNodeStatus: (id: string, status: TaskNode["status"]) => void;
  setCurrentTask: (phase: string, role: string) => void;
  setProgress: (progress: number, message: string) => void;
  setResult: (score: number, passed: boolean, blocking: string[], advisory: string[], artifacts: string[]) => void;
  addAuditEntry: (entry: AuditEntry) => void;
  setPlanId: (id: string) => void;
  setRunning: (running: boolean) => void;
  reset: () => void;
}

const initialState = {
  connected: false,
  address: "",
  currentPhase: "",
  currentRole: "",
  nodes: [],
  planId: "",
  running: false,
  progress: 0,
  message: "",
  score: 0,
  passed: false,
  blocking: [],
  advisory: [],
  artifacts: [],
  auditLog: [],
};

export const useStore = create<AppState>((set) => ({
  ...initialState,

  setConnected: (connected) => set({ connected }),
  setAddress: (address) => set({ address }),
  setNodes: (nodes) => set({ nodes }),
  updateNodeStatus: (id, status) =>
    set((state) => ({
      nodes: state.nodes.map((n) => (n.id === id ? { ...n, status } : n)),
    })),
  setCurrentTask: (currentPhase, currentRole) => set({ currentPhase, currentRole }),
  setProgress: (progress, message) => set({ progress, message }),
  setResult: (score, passed, blocking, advisory, artifacts) =>
    set({ score, passed, blocking, advisory, artifacts }),
  addAuditEntry: (entry) =>
    set((state) => ({ auditLog: [...state.auditLog.slice(-49), entry] })),
  setPlanId: (planId) => set({ planId }),
  setRunning: (running) => set({ running }),
  reset: () => set(initialState),
}));