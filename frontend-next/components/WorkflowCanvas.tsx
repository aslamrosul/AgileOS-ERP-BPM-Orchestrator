"use client";

import { useCallback, useRef, useState, useEffect } from "react";
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
import { Save, Download, Upload, Home, LogOut, User } from "lucide-react";
import Link from "next/link";

import BPMNode, { BPMNodeData } from "./BPMNode";
import NodeSidebar from "./NodeSidebar";
import { saveWorkflow } from "@/lib/api";
import { logout, getUser } from "@/lib/auth";

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
  const [selectedNode, setSelectedNode] = useState<Node<BPMNodeData> | null>(null);
  const [contextMenu, setContextMenu] = useState<{ x: number; y: number; nodeId: string } | null>(null);
  const [editingNode, setEditingNode] = useState<Node<BPMNodeData> | null>(null);
  
  const currentUser = getUser();

  const handleLogout = () => {
    if (confirm("Are you sure you want to logout?")) {
      logout();
    }
  };

  const onConnect = useCallback(
    (params: Connection) => setEdges((eds) => addEdge(params, eds)),
    [setEdges]
  );

  // Handle node click to select/edit
  const onNodeClick = useCallback((_event: React.MouseEvent, node: Node) => {
    setSelectedNode(node as Node<BPMNodeData>);
  }, []);

  // Handle node context menu (right-click)
  const onNodeContextMenu = useCallback(
    (event: React.MouseEvent, node: Node) => {
      event.preventDefault();
      setContextMenu({
        x: event.clientX,
        y: event.clientY,
        nodeId: node.id,
      });
      setSelectedNode(node as Node<BPMNodeData>);
    },
    []
  );

  // Close context menu
  const closeContextMenu = useCallback(() => {
    setContextMenu(null);
  }, []);

  // Handle edit node
  const handleEditNode = useCallback(() => {
    if (selectedNode) {
      setEditingNode(selectedNode);
      closeContextMenu();
    }
  }, [selectedNode, closeContextMenu]);

  // Handle delete node
  const handleDeleteNode = useCallback(() => {
    if (selectedNode) {
      setNodes((nds) => nds.filter((n) => n.id !== selectedNode.id));
      setEdges((eds) => eds.filter((e) => e.source !== selectedNode.id && e.target !== selectedNode.id));
      toast.success("Node deleted", {
        description: `Removed ${selectedNode.data.label}`,
      });
      setSelectedNode(null);
      closeContextMenu();
    }
  }, [selectedNode, setNodes, setEdges, closeContextMenu]);

  // Handle duplicate node
  const handleDuplicateNode = useCallback(() => {
    if (selectedNode) {
      const newNode: Node<BPMNodeData> = {
        ...selectedNode,
        id: getId(),
        position: {
          x: selectedNode.position.x + 50,
          y: selectedNode.position.y + 50,
        },
      };
      setNodes((nds) => nds.concat(newNode));
      toast.success("Node duplicated");
      closeContextMenu();
    }
  }, [selectedNode, setNodes, closeContextMenu]);

  // Save edited node
  const handleSaveNodeEdit = useCallback(
    (updatedData: Partial<BPMNodeData>) => {
      if (editingNode) {
        setNodes((nds) =>
          nds.map((n) =>
            n.id === editingNode.id
              ? { ...n, data: { ...n.data, ...updatedData } }
              : n
          )
        );
        toast.success("Node updated");
        setEditingNode(null);
      }
    },
    [editingNode, setNodes]
  );

  // Handle keyboard shortcuts
  const onKeyDown = useCallback(
    (event: KeyboardEvent) => {
      // Delete selected node with Delete or Backspace key
      if ((event.key === "Delete" || event.key === "Backspace") && selectedNode) {
        setNodes((nds) => nds.filter((n) => n.id !== selectedNode.id));
        setEdges((eds) => eds.filter((e) => e.source !== selectedNode.id && e.target !== selectedNode.id));
        toast.success("Node deleted", {
          description: `Removed ${selectedNode.data.label}`,
        });
        setSelectedNode(null);
      }
    },
    [selectedNode, setNodes, setEdges]
  );

  // Add keyboard event listener
  useEffect(() => {
    document.addEventListener("keydown", onKeyDown);
    document.addEventListener("click", closeContextMenu);
    return () => {
      document.removeEventListener("keydown", onKeyDown);
      document.removeEventListener("click", closeContextMenu);
    };
  }, [onKeyDown, closeContextMenu]);

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
            {/* User Info */}
            {currentUser && (
              <div className="flex items-center gap-2 px-3 py-2 bg-gray-100 rounded-lg mr-2">
                <User className="w-4 h-4 text-gray-600" />
                <div className="text-sm">
                  <div className="font-medium text-gray-800">{currentUser.username}</div>
                  <div className="text-xs text-gray-500">{currentUser.role}</div>
                </div>
              </div>
            )}

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

            <button
              onClick={handleLogout}
              className="flex items-center gap-2 px-4 py-2 bg-red-100 text-red-700 rounded-lg hover:bg-red-200 transition-colors"
              title="Logout"
            >
              <LogOut className="w-4 h-4" />
              <span className="text-sm font-medium">Logout</span>
            </button>
          </div>
        </div>

        {/* Canvas */}
        <div ref={reactFlowWrapper} className="flex-1 relative">
          <ReactFlow
            nodes={nodes}
            edges={edges}
            onNodesChange={onNodesChange}
            onEdgesChange={onEdgesChange}
            onConnect={onConnect}
            onNodeClick={onNodeClick}
            onNodeContextMenu={onNodeContextMenu}
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

          {/* Context Menu */}
          {contextMenu && (
            <div
              className="absolute bg-white rounded-lg shadow-lg border border-gray-200 py-1 z-50"
              style={{ left: contextMenu.x, top: contextMenu.y }}
            >
              <button
                onClick={handleEditNode}
                className="w-full px-4 py-2 text-left text-sm hover:bg-gray-100 flex items-center gap-2"
              >
                <span>✏️</span> Edit Properties
              </button>
              <button
                onClick={handleDuplicateNode}
                className="w-full px-4 py-2 text-left text-sm hover:bg-gray-100 flex items-center gap-2"
              >
                <span>📋</span> Duplicate
              </button>
              <hr className="my-1" />
              <button
                onClick={handleDeleteNode}
                className="w-full px-4 py-2 text-left text-sm hover:bg-red-50 text-red-600 flex items-center gap-2"
              >
                <span>🗑️</span> Delete
              </button>
            </div>
          )}

          {/* Edit Node Modal */}
          {editingNode && (
            <div className="absolute inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
              <div className="bg-white rounded-lg shadow-xl p-6 w-96">
                <h3 className="text-lg font-semibold mb-4">Edit Node Properties</h3>
                <div className="space-y-4">
                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-1">
                      Label
                    </label>
                    <input
                      type="text"
                      defaultValue={editingNode.data.label}
                      onChange={(e) => {
                        editingNode.data.label = e.target.value;
                      }}
                      className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500"
                    />
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-1">
                      Assigned To
                    </label>
                    <input
                      type="text"
                      defaultValue={editingNode.data.assignedTo}
                      onChange={(e) => {
                        editingNode.data.assignedTo = e.target.value;
                      }}
                      className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500"
                    />
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-1">
                      SLA
                    </label>
                    <input
                      type="text"
                      defaultValue={editingNode.data.sla}
                      onChange={(e) => {
                        editingNode.data.sla = e.target.value;
                      }}
                      className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500"
                      placeholder="e.g., 24h, 2d, 1w"
                    />
                  </div>
                </div>
                <div className="flex gap-2 mt-6">
                  <button
                    onClick={() => handleSaveNodeEdit(editingNode.data)}
                    className="flex-1 px-4 py-2 bg-indigo-600 text-white rounded-lg hover:bg-indigo-700"
                  >
                    Save
                  </button>
                  <button
                    onClick={() => setEditingNode(null)}
                    className="flex-1 px-4 py-2 bg-gray-200 text-gray-700 rounded-lg hover:bg-gray-300"
                  >
                    Cancel
                  </button>
                </div>
              </div>
            </div>
          )}
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
