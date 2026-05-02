# WebSocket Real-Time Notifications Implementation

## Overview

The AgileOS BPM platform now includes a complete real-time notification system using WebSockets. This enables instant notifications for task assignments, approvals, process completions, and system events.

## Architecture

### Backend Components

1. **WebSocket Hub** (`backend-go/internal/ws/hub.go`)
   - Manages active client connections
   - Routes messages to specific users or broadcasts to all
   - Handles client registration/unregistration
   - Thread-safe operations with mutex locks

2. **WebSocket Client** (`backend-go/internal/ws/client.go`)
   - Handles individual WebSocket connections
   - JWT authentication for secure connections
   - Ping/pong heartbeat mechanism
   - Message handling and routing

3. **NATS Notifier** (`backend-go/internal/ws/notifier.go`)
   - Bridges NATS events to WebSocket notifications
   - Subscribes to task and process events
   - Formats and sends real-time notifications

### Frontend Components

1. **useSocket Hook** (`frontend-next/hooks/useSocket.ts`)
   - Custom React hook for WebSocket management
   - Auto-reconnection with exponential backoff
   - JWT token authentication
   - Message handling and toast notifications

2. **WebSocketProvider** (`frontend-next/components/WebSocketProvider.tsx`)
   - React context provider for global WebSocket state
   - Manages connection lifecycle
   - Stores notification history

3. **UI Components**
   - `ConnectionStatus`: Shows real-time connection status
   - `NotificationsPanel`: Displays notification history with actions

## Features

### Real-Time Notifications

- **Task Assignments**: Instant notifications when tasks are assigned
- **Task Completions**: Updates when tasks are completed
- **Approval Requests**: High-priority notifications for approvals
- **Digital Signatures**: Confirmation when signatures are generated
- **System Notifications**: Broadcast messages for system events

### Connection Management

- **Auto-Connect**: Automatic connection on page load
- **Auto-Reconnect**: Intelligent reconnection with retry limits
- **Heartbeat**: Ping/pong to maintain connection health
- **JWT Authentication**: Secure connections with token validation

### User Experience

- **Toast Notifications**: Non-intrusive popup notifications using Sonner
- **Notification History**: Persistent notification panel with actions
- **Connection Status**: Visual indicator of real-time status
- **Priority Handling**: Different styling for urgent vs. normal notifications

## API Endpoints

### WebSocket Connection
```
GET /ws?token=<jwt_token>
```
- Upgrades HTTP connection to WebSocket
- Requires valid JWT token for authentication
- Returns 401 if token is invalid or missing

### Health Checks
```
GET /health
GET /health/live
GET /health/ready
```
- Includes WebSocket hub status in health responses

## Message Formats

### Notification Message
```json
{
  "id": "task_assigned_12345",
  "type": "task_assigned",
  "title": "New Task Assigned",
  "message": "You have been assigned a new task: Invoice Approval",
  "user_id": "user123",
  "task_id": "task456",
  "process_id": "process789",
  "workflow_id": "purchase_approval",
  "priority": "high",
  "action_url": "/tasks/task456",
  "data": {
    "step_name": "Manager Approval",
    "amount": 5000
  },
  "timestamp": 1777378870
}
```

### Client Message
```json
{
  "type": "ping",
  "data": {},
  "timestamp": 1777378870
}
```

## Notification Types

| Type | Description | Priority | Action |
|------|-------------|----------|--------|
| `task_assigned` | New task assignment | medium | View task |
| `task_completed` | Task completion | low | View details |
| `approval_request` | Approval needed | high | Review request |
| `signature_generated` | Digital signature created | medium | View signature |
| `system_notification` | System message | varies | Custom action |
| `connection_established` | WebSocket connected | low | None |

## Configuration

### Environment Variables

```bash
# Backend
PORT=8081
SURREAL_URL=ws://agileos-db:8000/rpc
NATS_URL=nats://agileos-nats:4222

# Frontend (Next.js)
NODE_ENV=production
```

### WebSocket Settings

```typescript
// Frontend configuration
const socketConfig = {
  autoConnect: true,
  reconnectInterval: 5000,
  maxReconnectAttempts: 10,
  heartbeatInterval: 30000
};
```

## Security

### Authentication
- JWT token required for WebSocket connections
- Token validation on connection upgrade
- Automatic disconnection for invalid tokens

### Authorization
- User-specific message routing
- Role-based notification filtering
- Audit logging for all WebSocket events

### Data Protection
- No sensitive data in WebSocket messages
- Reference IDs only (task_id, process_id)
- Secure token transmission via query parameter

## Testing

### Manual Testing
```bash
# Run the WebSocket test script
cd backend-go
./scripts/test-websocket.ps1
```

### Browser Testing
1. Open browser to `http://localhost:3001`
2. Open Developer Tools (F12)
3. Check Console for WebSocket connection logs
4. Look for notification bell icon
5. Create test tasks to trigger notifications

### Expected Behavior
- Connection status shows "Real-time" when connected
- Toast notifications appear for new tasks
- Notification panel shows message history
- Auto-reconnection on connection loss

## Troubleshooting

### Common Issues

1. **Connection Failed**
   - Check if backend is running on port 8081
   - Verify JWT token is valid
   - Check CORS settings

2. **No Notifications**
   - Verify NATS is running and connected
   - Check user authentication
   - Ensure task assignment events are published

3. **Frequent Disconnections**
   - Check network stability
   - Verify heartbeat mechanism
   - Review server logs for errors

### Debug Commands

```bash
# Check backend health
curl http://localhost:8081/health

# Test WebSocket endpoint
curl -i -N -H "Connection: Upgrade" \
     -H "Upgrade: websocket" \
     -H "Sec-WebSocket-Version: 13" \
     -H "Sec-WebSocket-Key: test" \
     http://localhost:8081/ws?token=<jwt_token>

# Monitor NATS messages
docker exec -it agileos-nats nats sub "task.*"
```

## Performance Considerations

### Scalability
- Hub uses efficient map-based client storage
- Concurrent message handling with goroutines
- Memory cleanup for disconnected clients

### Resource Usage
- Minimal memory footprint per connection
- Efficient JSON marshaling/unmarshaling
- Connection pooling for database operations

### Monitoring
- Connection count tracking
- Message delivery metrics
- Error rate monitoring

## Future Enhancements

### Planned Features
- Message persistence for offline users
- Push notifications for mobile apps
- Advanced filtering and subscription management
- Real-time dashboard updates
- WebSocket clustering for horizontal scaling

### Integration Points
- Mobile Flutter WebSocket support
- Email notification fallback
- Slack/Teams integration
- Analytics event tracking

## Mobile Support (Flutter)

The WebSocket implementation is designed to support Flutter mobile apps using the `web_socket_channel` package:

```dart
// Flutter WebSocket connection
final channel = WebSocketChannel.connect(
  Uri.parse('ws://10.0.2.2:8081/ws?token=$token'),
);

// Listen for notifications
channel.stream.listen((message) {
  final notification = jsonDecode(message);
  // Show snackbar or update UI
});
```

## Conclusion

The WebSocket implementation provides a robust, secure, and scalable real-time notification system for the AgileOS BPM platform. It enhances user experience by providing instant feedback on workflow activities and maintains high performance through efficient connection management and message routing.