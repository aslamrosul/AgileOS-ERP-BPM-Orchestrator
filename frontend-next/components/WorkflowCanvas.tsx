"use client";

import { useCallback, useRef, useState } from "react";
import ReactFlow, {
  Node,
  Edge,
  Controls,
  Background,
  BackgroundVariant,
  useNodesState,
  useEdgesState,
  addEdge,
  Connection,
  ReactFlowProvider,
  ReactFlowInstance,
} from "reactflow";
import "reactflow/dist/style.css";
import { toast } from "sonner";
import { Save, Download, Upload, Home } from "lucide-react";
import Link from "next/link";

import BPMNode, { BPMNodeData } from "./BPMNode";
import NodeSidebar from "./NodeSidebar";
import { saveWorkflow } from "@/lib/api";

const nodeTypes = {
  bpmNode: BPMNode,
};

let nodeId = 0;
const getId = () => `node_${nodeId++}`;

function WorkflowCanvasInner() {
  const reactFlowWrapper = useRef<HTMLDivElement>(null);
  const [nodes, setNodes, onNodesChange] = useNodesState([]);
  const [edges, setEdges, onEdgesChange] = useEdgesState([]);
  const [reactFlowInstance, setReactFlowInstance] = useState<ReactFlowInstance | null>(null);
  const [workflowName, setWorkflowName] = useState("Untitled Workflow");
  const [isSaving, setIsSaving] = useState(false);

  const onConnect = useCallback(
    (params: Connection) => setEdges((eds) => addEdge(params, eds)),
    [setEdges]
  );

  const onDragOver = useCallback((event: React.DragEvent) => {
    event.preventDefault();
    event.dataTransfer.dropEffect = "move";
  }, []);

  const onDrop = useCallback(
    (event: React.DragEvent) => {
      event.preventDefault();

      const type = event.dataTransfer.getData("application/reactflow");
      const label = event.dataTransfer.getData("label");

      if (typeof type === "undefined" || !type || !reactFlowInstance) {
        return;
      }

      const position = reactFlowInstance.screenToFlowPosition({
        x: event.clientX,
        y: event.clientY,
      });

      const newNode: Node<BPMNodeData> = {
        id: getId(),
        type: "bpmNode",
        position,
        data: {
          label: label || type,
          type: type as BPMNodeData["type"],
          assignedTo: "Unassigned",
          sla: "24h",
        },
      };

      setNodes((nds) => nds.concat(newNode));
    },
    [reactFlowInstance, setNodes]
  );

  const handleSave = async () => {
    if (nodes.length === 0) {
      toast.error("Workflow is empty", {
        description: "Add at least one node to save",
      });
      return;
    }

    setIsSaving(true);
    const toastId = toast.loading("Saving workflow...");

    try {
      const result = await saveWorkflow({
        name: workflowName,
        nodes,
        edges,
      });

      toast.success("Workflow saved successfully!", {
        id: toastId,
        description: `Workflow ID: ${result.workflow_id}`,
      });
    } catch (error: any) {
      toast.error("Failed to save workflow", {
        id: toastId,
        description: error.message || "Please check backend connection",
      });
    } finally {
      setIsSaving(false);
    }
  };

  const handleExport = () => {
    const workflow = {
      name: workflowName,
      nodes,
      edges,
      version: "1.0.0",
      exported_at: new Date().toISOString(),
    };

    const blob = new Blob([JSON.stringify(workflow, null, 2)], {
      type: "application/json",
    });
    const url = URL.createObjectURL(blob);
    const link = document.createElement("a");
    link.href = url;
    link.download = `${workflowName.replace(/\s+/g, "_")}.json`;
    link.click();
    URL.revokeObjectURL(url);

    toast.success("Workflow exported", {
      description: "JSON file downloaded",
    });
  };

  const handleImport = (event: React.ChangeEvent<HTMLInputElement>) => {
    const file = event.target.files?.[0];
    if (!file) return;

    const reader = new FileReader();
    reader.onload = (e) => {
      try {
        const workflow = JSON.parse(e.target?.result as string);
        setWorkflowName(workflow.name || "Imported Workflow");
        setNodes(workflow.nodes || []);
        setEdges(workflow.edges || []);
        toast.success("Workflow imported successfully");
      } catch (error) {
        toast.error("Failed to import workflow", {
          description: "Invalid JSON format",
        });
      }
    };
    reader.readAsText(file);
  };

  return (
    <div className="flex h-screen">
      <NodeSidebar />

      <div className="flex-1 flex flex-col">
        {/* Toolbar */}
        <div className="bg-white border-b border-gray-200 px-4 py-3 flex items-center justify-between">
          <div className="flex items-center gap-4">
            <Link
              href="/"
              className="p-2 hover:bg-gray-100 rounded-lg transition-colors"
              title="Home"
            >
              <Home className="w-5 h-5 text-gray-600" />
            </Link>
            <input
              type="text"
              value={workflowName}
              onChange={(e) => setWorkflowName(e.target.value)}
              className="text-lg font-semibold text-gray-800 bg-transparent border-none focus:outline-none focus:ring-2 focus:ring-indigo-500 rounded px-2"
              placeholder="Workflow Name"
            />
          </div>

          <div className="flex items-center gap-2">
            <label className="flex items-center gap-2 px-4 py-2 bg-gray-100 text-gray-700 rounded-lg hover:bg-gray-200 cursor-pointer transition-colors">
              <Upload className="w-4 h-4" />
              <span className="text-sm font-medium">Import</span>
              <input
                type="file"
                accept=".json"
                onChange={handleImport}
                className="hidden"
              />
            </label>

            <button
              onClick={handleExport}
              className="flex items-center gap-2 px-4 py-2 bg-gray-100 text-gray-700 rounded-lg hover:bg-gray-200 transition-colors"
            >
              <Download className="w-4 h-4" />
              <span className="text-sm font-medium">Export</span>
            </button>

            <button
              onClick={handleSave}
              disabled={isSaving}
              className="flex items-center gap-2 px-4 py-2 bg-indigo-600 text-white rounded-lg hover:bg-indigo-700 disabled:bg-indigo-400 transition-colors"
            >
              <Save className="w-4 h-4" />
              <span className="text-sm font-medium">
                {isSaving ? "Saving..." : "Save"}
              </span>
            </button>
          </div>
        </div>

        {/* Canvas */}
        <div ref={reactFlowWrapper} className="flex-1">
          <ReactFlow
            nodes={nodes}
            edges={edges}
            onNodesChange={onNodesChange}
            onEdgesChange={onEdgesChange}
            onConnect={onConnect}
            onInit={setReactFlowInstance}
            onDrop={onDrop}
            onDragOver={onDragOver}
            nodeTypes={nodeTypes}
            fitView
            className="bg-gray-50"
          >
            <Controls />
            <Background variant={BackgroundVariant.Dots} gap={12} size={1} />
          </ReactFlow>
        </div>

        {/* Status Bar */}
        <div className="bg-white border-t border-gray-200 px-4 py-2 flex items-center justify-between text-xs text-gray-600">
          <div>
            Nodes: {nodes.length} | Edges: {edges.length}
          </div>
          <div>
            AgileOS Workflow Builder v1.0.0
          </div>
        </div>
      </div>
    </div>
  );
}

export default function WorkflowCanvas() {
  return (
    <ReactFlowProvider>
      <WorkflowCanvasInner />
    </ReactFlowProvider>
  );
}
