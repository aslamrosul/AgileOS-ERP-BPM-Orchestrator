# Frontend Features - AgileOS Workflow Builder

## Visual Workflow Builder

### Canvas Features
- **Infinite Canvas**: Pan and zoom for large workflows
- **Grid Background**: Visual alignment guide
- **Minimap**: Navigate large workflows easily (via Controls)
- **Zoom Controls**: Zoom in/out, fit view

### Node Management
- **Drag & Drop**: Drag node templates from sidebar to canvas
- **Custom Nodes**: 6 types of BPM nodes
  - Start (gray) - Workflow entry point
  - Action (green) - Execute tasks
  - Approval (blue) - Require approval
  - Decision (yellow) - Conditional branching
  - Notify (purple) - Send notifications
  - End (gray) - Workflow completion

### Edge Management
- **Visual Connections**: Drag from output handle to input handle
- **Auto-routing**: Edges automatically route around nodes
- **Delete Edges**: Select edge and press Delete key

### Workflow Operations
- **Save**: Persist workflow to backend database
- **Export**: Download workflow as JSON file
- **Import**: Upload JSON file to restore workflow
- **Rename**: Click workflow name to edit

### UI/UX
- **Toast Notifications**: Real-time feedback for actions
- **Loading States**: Visual feedback during save operations
- **Status Bar**: Shows node/edge count and version
- **Responsive**: Works on different screen sizes

## Node Properties

Each node displays:
- **Icon**: Visual indicator of node type
- **Label**: Node name (editable in future)
- **Assigned To**: Role or user (e.g., "role:manager")
- **SLA**: Service Level Agreement time (e.g., "24h")
- **Type**: Node category

## Keyboard Shortcuts

- **Delete**: Remove selected nodes/edges
- **Ctrl/Cmd + Z**: Undo (React Flow built-in)
- **Mouse Wheel**: Zoom in/out
- **Space + Drag**: Pan canvas

## API Integration

### Save Workflow
```typescript
POST /api/v1/workflow
{
  "workflow": {
    "name": "Purchase Approval",
    "version": "1.0.0",
    "description": "...",
    "is_active": true
  },
  "steps": [
    {
      "id": "node_1",
      "name": "Submit Request",
      "type": "action",
      "assigned_to": "role:employee",
      "sla": "1h",
      "position": { "x": 100, "y": 100 }
    }
  ],
  "relations": [
    {
      "from": "node_1",
      "to": "node_2",
      "condition": null
    }
  ]
}
```

### Response
```json
{
  "workflow_id": "workflow:abc123",
  "steps_created": 3,
  "relations_created": 2,
  "message": "Workflow created successfully"
}
```

## Data Flow

```
User Action (Drag Node)
    ↓
React Flow State Update
    ↓
WorkflowCanvas Component
    ↓
Save Button Click
    ↓
Transform to Backend Format (lib/api.ts)
    ↓
POST to Backend API
    ↓
Backend Creates Workflow + Steps + Relations
    ↓
Success Toast Notification
```

## Component Architecture

```
WorkflowCanvas (Main Container)
├── NodeSidebar (Template Library)
│   └── Node Templates (Draggable)
├── ReactFlow (Canvas)
│   ├── BPMNode (Custom Node Component)
│   ├── Controls (Zoom, Fit View)
│   └── Background (Grid)
└── Toolbar
    ├── Workflow Name Input
    ├── Import Button
    ├── Export Button
    └── Save Button
```

## State Management

Uses React Flow's built-in state management:
- `useNodesState`: Manages nodes array
- `useEdgesState`: Manages edges array
- `ReactFlowInstance`: Canvas instance for operations

## Styling

- **Tailwind CSS**: Utility-first styling
- **Custom Classes**: BPM-specific styles
- **React Flow CSS**: Canvas and controls styling

## Future Enhancements

- [ ] Node property editor panel
- [ ] Conditional edge labels
- [ ] Workflow validation
- [ ] Auto-layout algorithm
- [ ] Collaborative editing
- [ ] Version history
- [ ] Template library
- [ ] Search and filter nodes
- [ ] Workflow execution view
- [ ] Real-time updates via WebSocket

## Browser Support

- Chrome/Edge (Recommended)
- Firefox
- Safari
- Opera

Requires modern browser with ES6+ support.

## Performance

- Optimized for workflows up to 100 nodes
- Lazy rendering for large workflows
- Memoized components to prevent re-renders
- Efficient edge routing algorithm

## Accessibility

- Keyboard navigation support
- ARIA labels on interactive elements
- High contrast mode compatible
- Screen reader friendly (basic support)

## Error Handling

- Network errors: Toast notification with retry suggestion
- Validation errors: Inline error messages
- Backend errors: Detailed error descriptions
- Graceful degradation: Works offline for editing (save fails)
