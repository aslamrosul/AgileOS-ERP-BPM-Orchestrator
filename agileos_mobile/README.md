# AgileOS Mobile - Real-Time Approval Interface

Flutter mobile application for real-time task approval integrated with AgileOS backend.

## Features

- **Real-time Task List**: View pending tasks assigned to your role
- **Quick Actions**: Approve or reject tasks with one tap
- **Pull-to-Refresh**: Manual refresh for latest tasks
- **Role-based Filtering**: Switch between different roles (Manager, Finance, Procurement, Employee)
- **Time Tracking**: Visual indicators for task urgency
- **Modern UI**: Professional design with Deep Blue & Slate color scheme

## Architecture

```
┌─────────────────────────────────────────────────────────┐
│                    Mobile App (Flutter)                  │
├─────────────────────────────────────────────────────────┤
│  State Management: Riverpod                              │
│  HTTP Client: http package                               │
│  UI Framework: Material 3                                │
│  Fonts: Google Fonts (Inter)                            │
└─────────────────────────────────────────────────────────┘
                          ↓
┌─────────────────────────────────────────────────────────┐
│                    Backend API (Go)                     │
│              http://localhost:8080/api/v1               │
└─────────────────────────────────────────────────────────┘
                          ↓
┌─────────────────────────────────────────────────────────┐
│                    NATS Orchestration                    │
│              Automatic workflow progression              │
└─────────────────────────────────────────────────────────┘
```

## Setup Instructions

### 1. Install Flutter

Make sure you have Flutter installed:
```bash
flutter --version
```

If not installed, download from: https://flutter.dev/docs/get-started/install

### 2. Run Backend First

Before running the mobile app, start the backend:

```bash
cd agile-os/backend-go
.\run-local.ps1
```

Verify backend is running:
```bash
curl http://localhost:8080/health
```

### 3. Configure Network Access

#### For Android Emulator:
- Use `10.0.2.2` to access localhost
- Already configured in `apiBaseUrlProvider`

#### For iOS Simulator:
- Use `localhost`
- Update `apiBaseUrlProvider` in `task_provider.dart`

#### For Physical Device:
- Use your computer's IP address
- Ensure firewall allows port 8080

### 4. Run Mobile App

```bash
cd agile-os/agileos_mobile
flutter pub get
flutter run
```

## API Integration

The app communicates with these backend endpoints:

### GET /api/v1/tasks/pending/:assignedTo
Fetches pending tasks for a specific role/user.

**Example:** `GET /api/v1/tasks/pending/role:manager`

**Response:**
```json
{
  "tasks": [
    {
      "id": "task_instance:abc123",
      "step_name": "Manager Approval",
      "assigned_to": "role:manager",
      "status": "pending",
      "created_at": "2026-04-28T07:26:50.074132500Z",
      "due_at": "2026-04-29T07:26:50.074132500Z",
      "data": {
        "amount": 5000,
        "description": "New laptops"
      }
    }
  ],
  "count": 1
}
```

### POST /api/v1/task/:id/complete
Completes a task (approve/reject).

**Example:** `POST /api/v1/task/task_instance:abc123/complete`

**Request Body:**
```json
{
  "executed_by": "mobile_user",
  "result": {
    "decision": "approved",
    "comments": "Approved via mobile app",
    "timestamp": "2026-04-28T07:30:00Z"
  }
}
```

## State Management

### Providers
- `taskServiceProvider`: HTTP client for API calls
- `pendingTasksProvider`: Fetches and caches pending tasks
- `taskActionProvider`: Handles approve/reject actions
- `assignedToProvider`: Current role filter

### Riverpod Benefits
- **Reactive**: Automatically updates UI when state changes
- **Testable**: Easy to mock dependencies
- **Scalable**: Works well for complex state
- **Type-safe**: Full Dart type checking

## UI Components

### Task Card
Displays:
- Task name and description
- Assigned department
- Creation time
- Time remaining (with urgency indicator)
- Request details (if available)
- Approve/Reject buttons

### Features
- **Pull-to-Refresh**: Swipe down to refresh task list
- **Role Filtering**: Switch between different roles
- **Confirmation Dialogs**: Prevent accidental actions
- **Success/Error Feedback**: Snackbars for user feedback

## Testing

### 1. Seed Database
First, seed the database with sample workflow:

```bash
# Open SurrealDB Dashboard
open http://localhost:8000

# Run seed queries from:
# backend-go/database/seed.surql
```

### 2. Start Process
Create a process instance:

```bash
curl -X POST http://localhost:8080/api/v1/process/start \
  -H "Content-Type: application/json" \
  -d '{
    "workflow_id": "workflow:purchase_approval",
    "initiated_by": "user:test",
    "data": {"amount": 5000}
  }'
```

### 3. Test Mobile App
1. Open the app
2. Select "Manager" role
3. You should see pending tasks
4. Test approve/reject actions

## Troubleshooting

### No Tasks Showing
1. Check backend is running: `curl http://localhost:8080/health`
2. Verify database is seeded with `purchase_approval` workflow
3. Check role filter matches assigned tasks

### Network Connection Error
1. **Emulator**: Use `10.0.2.2` instead of `localhost`
2. **Physical Device**: Use computer's IP address
3. **Firewall**: Allow port 8080

### Build Errors
```bash
flutter clean
flutter pub get
flutter run
```

## Future Enhancements

### Planned Features
- [ ] Push notifications via WebSocket
- [ ] Offline support with local cache
- [ ] Biometric authentication
- [ ] Dark mode
- [ ] Task search and filtering
- [ ] Analytics dashboard
- [ ] Multi-language support

### WebSocket Integration
For real-time updates, add WebSocket support:

```dart
// Future enhancement
final webSocketProvider = StreamProvider.autoDispose((ref) {
  final channel = IOWebSocketChannel.connect(
    'ws://10.0.2.2:8080/ws/tasks',
  );
  return channel.stream;
});
```

## Development Notes

### Code Structure
```
lib/
├── main.dart              # App entry point
├── models/
│   └── task.dart         # Task data model
├── providers/
│   └── task_provider.dart # State management
└── screens/
    └── task_screen.dart  # Main task interface
```

### Design System
- **Primary Color**: Deep Blue (#1E3A8A)
- **Secondary Color**: Slate (#0F172A)
- **Background**: Light Gray (#F8FAFC)
- **Font**: Inter (Google Fonts)
- **Icons**: Material Icons

### Performance Tips
1. Use `const` widgets where possible
2. Implement pagination for large task lists
3. Cache API responses
4. Use `ListView.builder` for efficient scrolling
5. Debounce rapid API calls

## License

AgileOS Mobile is part of the AgileOS BPM Platform.
