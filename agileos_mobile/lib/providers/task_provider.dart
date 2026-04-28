import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:http/http.dart' as http;
import 'dart:convert';
import '../models/task.dart';

final apiBaseUrlProvider = Provider<String>((ref) {
  // For Android emulator: 10.0.2.2
  // For iOS simulator: localhost
  // For physical device: Use your computer's IP address
  // Change this to your laptop IP if using physical device
  return 'http://192.168.1.66:8081';
});

class TaskService {
  final String baseUrl;

  TaskService(this.baseUrl);

  Future<List<Task>> fetchPendingTasks(String assignedTo) async {
    try {
      final response = await http.get(
        Uri.parse('$baseUrl/api/v1/tasks/pending/$assignedTo'),
        headers: {'Content-Type': 'application/json'},
      );

      if (response.statusCode == 200) {
        final data = json.decode(response.body);
        final tasks = (data['tasks'] as List)
            .map((taskJson) => Task.fromJson(taskJson))
            .toList();
        return tasks;
      } else {
        throw Exception('Failed to load tasks: ${response.statusCode}');
      }
    } catch (e) {
      throw Exception('Network error: $e');
    }
  }

  Future<void> approveTask(String taskId, String executedBy) async {
    try {
      final response = await http.post(
        Uri.parse('$baseUrl/api/v1/task/$taskId/complete'),
        headers: {'Content-Type': 'application/json'},
        body: json.encode({
          'executed_by': executedBy,
          'result': {
            'decision': 'approved',
            'comments': 'Approved via mobile app',
            'timestamp': DateTime.now().toIso8601String(),
          },
        }),
      );

      if (response.statusCode != 200) {
        throw Exception('Failed to approve task: ${response.statusCode}');
      }
    } catch (e) {
      throw Exception('Network error: $e');
    }
  }

  Future<void> rejectTask(String taskId, String executedBy, String reason) async {
    try {
      final response = await http.post(
        Uri.parse('$baseUrl/api/v1/task/$taskId/complete'),
        headers: {'Content-Type': 'application/json'},
        body: json.encode({
          'executed_by': executedBy,
          'result': {
            'decision': 'rejected',
            'comments': reason,
            'timestamp': DateTime.now().toIso8601String(),
          },
        }),
      );

      if (response.statusCode != 200) {
        throw Exception('Failed to reject task: ${response.statusCode}');
      }
    } catch (e) {
      throw Exception('Network error: $e');
    }
  }
}

final taskServiceProvider = Provider<TaskService>((ref) {
  final baseUrl = ref.watch(apiBaseUrlProvider);
  return TaskService(baseUrl);
});

final assignedToProvider = StateProvider<String>((ref) => 'role:manager');

final pendingTasksProvider = FutureProvider<List<Task>>((ref) async {
  final taskService = ref.watch(taskServiceProvider);
  final assignedTo = ref.watch(assignedToProvider);
  
  try {
    return await taskService.fetchPendingTasks(assignedTo);
  } catch (e) {
    throw Exception('Failed to fetch tasks: $e');
  }
});

final taskActionProvider = StateNotifierProvider<TaskActionNotifier, AsyncValue<void>>(
  (ref) => TaskActionNotifier(ref),
);

class TaskActionNotifier extends StateNotifier<AsyncValue<void>> {
  final Ref ref;

  TaskActionNotifier(this.ref) : super(const AsyncValue.data(null));

  Future<void> approveTask(String taskId) async {
    state = const AsyncValue.loading();
    try {
      final taskService = ref.read(taskServiceProvider);
      await taskService.approveTask(taskId, 'mobile_user');
      state = const AsyncValue.data(null);
      
      // Refresh tasks after approval
      ref.invalidate(pendingTasksProvider);
    } catch (e) {
      state = AsyncValue.error(e, StackTrace.current);
      rethrow;
    }
  }

  Future<void> rejectTask(String taskId, String reason) async {
    state = const AsyncValue.loading();
    try {
      final taskService = ref.read(taskServiceProvider);
      await taskService.rejectTask(taskId, 'mobile_user', reason);
      state = const AsyncValue.data(null);
      
      // Refresh tasks after rejection
      ref.invalidate(pendingTasksProvider);
    } catch (e) {
      state = AsyncValue.error(e, StackTrace.current);
      rethrow;
    }
  }
}
