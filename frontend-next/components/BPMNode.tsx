import { memo } from "react";
import { Handle, Position, NodeProps } from "reactflow";
import { CheckCircle, PlayCircle, AlertCircle, Bell } from "lucide-react";

export type BPMNodeData = {
  label: string;
  type: "approval" | "action" | "decision" | "notify" | "start" | "end";
  assignedTo?: string;
  sla?: string;
};

const nodeIcons = {
  approval: CheckCircle,
  action: PlayCircle,
  decision: AlertCircle,
  notify: Bell,
  start: PlayCircle,
  end: CheckCircle,
};

const nodeColors = {
  approval: "bg-blue-500",
  action: "bg-green-500",
  decision: "bg-yellow-500",
  notify: "bg-purple-500",
  start: "bg-gray-700",
  end: "bg-gray-700",
};

function BPMNode({ data, selected }: NodeProps<BPMNodeData>) {
  const Icon = nodeIcons[data.type];

  return (
    <div
      className={`px-4 py-3 shadow-lg rounded-lg border-2 bg-white min-w-[180px] ${
        selected ? "border-indigo-500" : "border-gray-300"
      }`}
    >
      <Handle
        type="target"
        position={Position.Top}
        className="w-3 h-3 !bg-gray-400"
      />

      <div className="flex items-center gap-2 mb-2">
        <div className={`p-1.5 rounded ${nodeColors[data.type]}`}>
          <Icon className="w-4 h-4 text-white" />
        </div>
        <div className="font-semibold text-sm text-gray-800">{data.label}</div>
      </div>

      {data.assignedTo && (
        <div className="text-xs text-gray-600 mb-1">
          👤 {data.assignedTo}
        </div>
      )}

      {data.sla && (
        <div className="text-xs text-gray-500">
          ⏱️ SLA: {data.sla}
        </div>
      )}

      <div className="text-xs text-gray-400 mt-1 capitalize">
        {data.type}
      </div>

      <Handle
        type="source"
        position={Position.Bottom}
        className="w-3 h-3 !bg-gray-400"
      />
    </div>
  );
}

export default memo(BPMNode);
