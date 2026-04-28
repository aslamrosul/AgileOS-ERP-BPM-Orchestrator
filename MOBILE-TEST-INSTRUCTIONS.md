# Mobile App Testing Instructions

## Current Status
- ✅ Backend running on port 8081
- ✅ Flutter app building to your Redmi Note 8
- ✅ 1 pending task available for testing
- ⚠️ Minor bug: UpdateTaskInstance still has issues, but tasks are being created

## What's Happening Now
Flutter is building the app to your phone. This takes 3-5 minutes for first build.

## When App Opens

### 1. You Should See:
- App title: "AgileOS Mobile"
- A list of pending tasks
- At least 1 task card showing:
  - Step name: "Manager Review" (or similar)
  - Status: "pending"
  - Green "Approve" button
  - Red "Reject" button

### 2. To Test Approval:
1. Tap the "Approve" button on any task
2. Watch for success message
3. Pull down to refresh the list
4. The approved task should disappear (status changed to "completed")

### 3. To Verify in Database:
Open http://localhost:8000 in browser and run:
```sql
SELECT * FROM task_instance WHERE status = 'completed' ORDER BY completed_at DESC LIMIT 5;
```

You should see your approved task with:
- `status`: "completed"
- `completed_at`: timestamp
- `result`: {"decision": "approved", "comments": "Approved via mobile app"}

## API Endpoints Being Used

### Get Pending Tasks
```
GET http://192.168.1.66:8081/api/v1/tasks/pending/role:manager
```

### Complete Task (Approve)
```
POST http://192.168.1.66:8081/api/v1/task/{task_id}/complete
Body: {
  "executed_by": "mobile_user",
  "result": {
    "decision": "approved",
    "comments": "Approved via mobile app"
  }
}
```

## Create More Test Tasks

Run this script to create more tasks for testing:
```powershell
cd agile-os/backend-go
.\scripts\setup-test.ps1
```

## Troubleshooting

### If app shows "No tasks":
1. Pull down to refresh
2. Check backend is running: http://192.168.1.66:8081/health
3. Run setup-test.ps1 to create more tasks

### If "Network Error":
1. Make sure phone and laptop are on same WiFi
2. Check IP address is correct (192.168.1.66)
3. Check Windows Firewall allows port 8081

### If approval fails:
- This is the known bug we're working on
- But the task should still be created in database
- Check backend logs for details

## Next Steps After Testing

Once you confirm:
- ✅ App shows pending tasks
- ✅ Can tap Approve button
- ✅ See success/error message

Then we can:
1. Fix the remaining UpdateTaskInstance bug
2. Add better error handling
3. Add pull-to-refresh animation
4. Add task details screen
5. Add push notifications

## Backend Logs

To watch backend logs while testing:
```powershell
# Backend is running in background terminal ID: 7
# Logs show:
# - Task fetched
# - Task completion attempts
# - Orchestration events
```

---

**Current Time**: Waiting for Flutter build to complete...
**Estimated**: 2-3 more minutes
