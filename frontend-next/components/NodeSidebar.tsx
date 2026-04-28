import { CheckCircle, PlayCircle, AlertCircle, Bell, Circle } from "lucide-react";

const nodeTemplates = [
  {
    type: "start",
    label: "Start",
    icon: Circle,
    color: "bg-gray-700",
    description: "Starting point",
  },
  {
    type: "action",
    label: "Action",
    icon: PlayCircle,
    color: "bg-green-500",
    description: "Execute task",
  },
  {
    type: "approval",
    label: "Approval",
    icon: CheckCircle,
    color: "bg-blue-500",
    description: "Requires approval",
  },
  {
    type: "decision",
    label: "Decision",
    icon: AlertCircle,
    color: "bg-yellow-500",
    description: "Conditional branch",
  },
  {
    type: "notify",
    label: "Notify",
    icon: Bell,
    color: "bg-purple-500",
    description: "Send notification",
  },
  {
    type: "end",
    label: "End",
    icon: Circle,
    color: "bg-gray-700",
    description: "End point",
  },
];

export default function NodeSidebar() {
  const onDragStart = (event: React.DragEvent, nodeType: string, label: string) => {
    event.dataTransfer.setData("application/reactflow", nodeType);
    event.dataTransfer.setData("label", label);
    event.dataTransfer.effectAllowed = "move";
  };

  return (
    <div className="w-64 bg-white border-r border-gray-200 p-4 overflow-y-auto">
      <h2 className="text-lg font-bold text-gray-800 mb-4">Node Templates</h2>
      <p className="text-xs text-gray-500 mb-4">
        Drag nodes to canvas to build workflow
      </p>

      <div className="space-y-2">
        {nodeTemplates.map((template) => {
          const Icon = template.icon;
          return (
            <div
              key={template.type}
              draggable
              onDragStart={(e) => onDragStart(e, template.type, template.label)}
              className="flex items-center gap-3 p-3 bg-gray-50 rounded-lg border border-gray-200 cursor-move hover:bg-gray-100 hover:border-gray-300 transition-colors"
            >
              <div className={`p-2 rounded ${template.color}`}>
                <Icon className="w-4 h-4 text-white" />
              </div>
              <div className="flex-1">
                <div className="font-medium text-sm text-gray-800">
                  {template.label}
                </div>
                <div className="text-xs text-gray-500">
                  {template.description}
                </div>
              </div>
            </div>
          );
        })}
      </div>

      <div className="mt-6 p-3 bg-blue-50 rounded-lg border border-blue-200">
        <h3 className="text-sm font-semibold text-blue-900 mb-2">
          💡 Quick Tips
        </h3>
        <ul className="text-xs text-blue-800 space-y-1">
          <li>• Drag nodes to canvas</li>
          <li>• Connect nodes by dragging handles</li>
          <li>• Click node to edit properties</li>
          <li>• Delete: Select + Delete key</li>
        </ul>
      </div>
    </div>
  );
}
