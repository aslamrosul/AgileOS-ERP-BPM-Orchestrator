# AgileOS - Enterprise BPM Platform

Sistem manajemen proses bisnis berbasis cloud yang berjalan di Azure dengan Docker.

## Arsitektur

```
┌─────────────────────────────────────────────────┐
│              AgileOS Platform                    │
├─────────────────────────────────────────────────┤
│  Frontend (Next.js 14)                          │
│  ↓                                               │
│  Backend Engine (Go + Gin)                      │
│  ↓                                               │
│  ┌──────────────┐      ┌──────────────┐        │
│  │  SurrealDB   │      │     NATS     │        │
│  │  (Database)  │      │ (Msg Broker) │        │
│  └──────────────┘      └──────────────┘        │
└─────────────────────────────────────────────────┘
```

## Tech Stack

- **Backend**: Go 1.22 + Gin Framework
- **Frontend**: Next.js 14 + React Flow + Tailwind CSS
- **Database**: SurrealDB (Multi-model database)
- **Message Broker**: NATS (High-performance messaging)
- **Infrastructure**: Docker + Azure

## Quick Start

### Prerequisites
- Docker & Docker Compose
- Go 1.22+ (untuk development)
- Node.js 18+ & npm

### Menjalankan Platform Lengkap

**1. Start Infrastructure:**
```bash
cd agile-os
docker-compose up -d
```

**2. Start Backend:**
```bash
cd backend-go
.\run-local.ps1
```

**3. Start Frontend:**
```bash
cd frontend-next
npm install
npm run dev
```

**4. Open Browser:**
- Landing Page: http://localhost:3000
- Workflow Builder: http://localhost:3000/workflow
- Backend API: http://localhost:8080/health

📖 Detailed guide: See [QUICKSTART.md](QUICKSTART.md)

## Struktur Proyek

```
agile-os/
├── backend-go/          # Go backend engine
│   ├── handlers/        # API handlers
│   │   ├── workflow.go  # Workflow management
│   │   └── task.go      # Task & process execution
│   ├── models/          # Data models
│   │   ├── workflow.go  # Workflow, Step, ProcessInstance
│   │   └── task.go      # TaskInstance
│   ├── database/        # Repository layer
│   │   ├── surreal.go   # SurrealDB operations
│   │   └── seed.surql   # Sample data
│   ├── messaging/       # NATS integration
│   │   ├── nats.go      # Event-driven orchestration
│   │   └── README.md    # Messaging documentation
│   ├── scripts/         # Utility scripts
│   │   ├── test-api.ps1
│   │   └── test-orchestration.ps1
│   ├── Dockerfile       # Multi-stage build
│   ├── main.go          # Entry point
│   └── go.mod           # Dependencies
├── frontend-next/       # Next.js 14 dashboard
│   ├── app/             # App router pages
│   ├── components/      # React components
│   │   ├── WorkflowCanvas.tsx
│   │   ├── BPMNode.tsx
│   │   └── NodeSidebar.tsx
│   ├── lib/             # API client
│   └── package.json
├── deploy/              # Infrastructure configs
│   └── azure/           # Azure deployment files
├── data/                # Persistent data (gitignored)
│   └── surrealdb/       # Database files
├── docker-compose.yml   # Local orchestration
├── QUICKSTART.md        # Quick start guide
├── ORCHESTRATION.md     # Event-driven orchestration guide
└── VERIFICATION.md      # Verification guide
```

## Services

### SurrealDB (Port 8000)
- URL: `http://localhost:8000`
- User: `root`
- Pass: `root`
- Data: Persisted to `./data/surrealdb`

### NATS (Port 4222, 8222)
- Client: `nats://localhost:4222`
- Monitoring: `http://localhost:8222`

### Backend Engine (Port 8080)
- Health: `http://localhost:8080/health`
- API: `http://localhost:8080/api/v1/`

### Frontend Dashboard (Port 3000)
- Landing: `http://localhost:3000`
- Workflow Builder: `http://localhost:3000/workflow`

## Features

### ✅ Implemented (Phase 1-4)

- **Visual Workflow Builder**
  - Drag-and-drop interface
  - Custom BPM nodes (Start, Action, Approval, Decision, Notify, End)
  - Real-time canvas editing
  - Export/Import workflows as JSON

- **Backend API**
  - RESTful API with Gin framework
  - Workflow CRUD operations
  - Graph-based step relations
  - SurrealDB integration
  - Task management endpoints
  - Process execution API

- **Database Architecture**
  - Repository pattern
  - Graph relations for workflow steps
  - Process instance tracking
  - Task instance management
  - Execution history logging

- **Event-Driven Orchestration**
  - NATS message broker integration
  - Automatic workflow progression
  - Task completion triggers next steps
  - Background worker with goroutines
  - Detailed event logging
  - Non-blocking async processing

- **Infrastructure**
  - Docker orchestration
  - SurrealDB for data persistence
  - NATS for messaging
  - CORS-enabled API

### 🚧 Coming Soon (Phase 5+)
- [ ] Real-time process monitoring
- [ ] User authentication & authorization
- [ ] Role-based access control (RBAC)
- [ ] Webhook integrations
- [ ] Email notifications
- [ ] Analytics dashboard
- [ ] Azure deployment
- [ ] CI/CD pipeline

## Development

Stop all services:
```bash
docker-compose down
```

Clean data (reset database):
```bash
docker-compose down -v
rm -rf data/surrealdb/*
```

View logs:
```bash
docker-compose logs -f
```
