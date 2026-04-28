# AgileOS Frontend - Workflow Builder

Next.js 14 application dengan React Flow untuk visual workflow builder.

## Tech Stack

- Next.js 14 (App Router)
- React Flow - Visual workflow canvas
- Tailwind CSS - Styling
- Lucide React - Icons
- Sonner - Toast notifications
- Axios - API client

## Getting Started

### 1. Install Dependencies

```bash
cd frontend-next
npm install
```

### 2. Configure Backend URL

File `.env.local` sudah dikonfigurasi:
```
NEXT_PUBLIC_API_URL=http://localhost:8080
```

### 3. Run Development Server

```bash
npm run dev
```

Open [http://localhost:3000](http://localhost:3000)

## Features

### Visual Workflow Builder
- Drag-and-drop node templates
- Connect nodes with edges
- Custom BPM nodes (Approval, Action, Decision, Notify)
- Real-time canvas editing

### Node Types
- **Start**: Workflow starting point
- **Action**: Execute task
- **Approval**: Requires approval
- **Decision**: Conditional branching
- **Notify**: Send notification
- **End**: Workflow end point

### Workflow Management
- Save workflow to backend
- Export workflow as JSON
- Import workflow from JSON
- Real-time validation

## Project Structure

```
frontend-next/
├── app/
│   ├── page.tsx              # Landing page
│   ├── workflow/
│   │   └── page.tsx          # Workflow builder page
│   ├── layout.tsx            # Root layout
│   └── globals.css           # Global styles
├── components/
│   ├── WorkflowCanvas.tsx    # Main canvas component
│   ├── BPMNode.tsx           # Custom node component
│   └── NodeSidebar.tsx       # Node templates sidebar
├── lib/
│   └── api.ts                # Backend API client
└── package.json
```

## Usage

### 1. Create Workflow

1. Open workflow builder: `/workflow`
2. Drag nodes from sidebar to canvas
3. Connect nodes by dragging from output handle to input handle
4. Edit workflow name at top
5. Click "Save" to persist to backend

### 2. Export/Import

- **Export**: Download workflow as JSON file
- **Import**: Upload JSON file to restore workflow

### 3. Backend Integration

The app sends workflow data to backend API:

```typescript
POST /api/v1/workflow
{
  "workflow": {
    "name": "Purchase Approval",
    "version": "1.0.0",
    "description": "...",
    "is_active": true
  },
  "steps": [...],
  "relations": [...]
}
```

## Development

### Build for Production

```bash
npm run build
npm start
```

### Lint

```bash
npm run lint
```

## Connecting to Backend

Pastikan backend Go sudah running di port 8080:

```bash
cd ../backend-go
.\run-local.ps1
```

Test connection:
```bash
curl http://localhost:8080/health
```

## Troubleshooting

### Cannot connect to backend

**Error**: "Cannot connect to backend. Is it running on port 8080?"

**Solution**:
1. Check backend is running: `curl http://localhost:8080/health`
2. Verify `.env.local` has correct URL
3. Check CORS settings in backend

### Nodes not draggable

**Solution**: Make sure React Flow is properly initialized with `ReactFlowProvider`

### Styles not loading

**Solution**: 
```bash
npm install -D tailwindcss postcss autoprefixer
```

## Next Steps

- [ ] Add node property editor panel
- [ ] Implement workflow versioning
- [ ] Add workflow execution view
- [ ] Real-time collaboration
- [ ] Workflow templates library
