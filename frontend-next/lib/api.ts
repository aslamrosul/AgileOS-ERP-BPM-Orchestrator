import axios from "axios";
import { Node, Edge } from "reactflow";
import { BPMNodeData } from "@/components/BPMNode";
import { getAccessToken } from "./auth";

const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";

const api = axios.create({
  baseURL: API_BASE_URL,
  headers: {
    "Content-Type": "application/json",
  },
  timeout: 10000,
});

// Add request interceptor to include auth token
api.interceptors.request.use(
  (config) => {
    const token = getAccessToken();
    if (token) {
      config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
  },
  (error) => {
    return Promise.reject(error);
  }
);

// Export api instance
export { api };

export interface WorkflowPayload {
  name: string;
  nodes: Node<BPMNodeData>[];
  edges: Edge[];
}

export interface WorkflowResponse {
  workflow_id: string;
  steps_created: number;
  relations_created: number;
  message: string;
}

export async function saveWorkflow(payload: WorkflowPayload): Promise<WorkflowResponse> {
  try {
    // Transform React Flow data to backend format
    const steps = payload.nodes.map((node) => ({
      id: node.id,
      name: node.data.label,
      type: node.data.type,
      assigned_to: node.data.assignedTo || "unassigned",
      sla: node.data.sla || "24h",
      position: node.position,
    }));

    const relations = payload.edges.map((edge) => ({
      from: edge.source,
      to: edge.target,
      condition: edge.label ? { label: edge.label } : null,
    }));

    const backendPayload = {
      workflow: {
        name: payload.name,
        version: "1.0.0",
        description: `Workflow created via visual builder`,
        is_active: true,
      },
      steps,
      relations,
    };

    const response = await api.post<WorkflowResponse>("/api/v1/workflow", backendPayload);
    return response.data;
  } catch (error: any) {
    if (error.response) {
      throw new Error(error.response.data.error || "Server error");
    } else if (error.request) {
      throw new Error("Cannot connect to backend. Is it running on port 8080?");
    } else {
      throw new Error(error.message || "Unknown error");
    }
  }
}

export async function getWorkflows() {
  try {
    const response = await api.get("/api/v1/workflows");
    return response.data;
  } catch (error: any) {
    throw new Error(error.response?.data?.error || "Failed to fetch workflows");
  }
}

export async function getWorkflow(id: string) {
  try {
    const response = await api.get(`/api/v1/workflow/${id}`);
    return response.data;
  } catch (error: any) {
    throw new Error(error.response?.data?.error || "Failed to fetch workflow");
  }
}
