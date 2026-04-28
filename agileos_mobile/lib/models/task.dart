import 'package:intl/intl.dart';

class Task {
  final String id;
  final String processInstanceId;
  final String stepId;
  final String stepName;
  final String status;
  final String assignedTo;
  final DateTime createdAt;
  final DateTime dueAt;
  final dynamic data;
  final dynamic result;

  Task({
    required this.id,
    required this.processInstanceId,
    required this.stepId,
    required this.stepName,
    required this.status,
    required this.assignedTo,
    required this.createdAt,
    required this.dueAt,
    this.data,
    this.result,
  });

  factory Task.fromJson(Map<String, dynamic> json) {
    return Task(
      id: json['id'] ?? '',
      processInstanceId: json['process_instance_id'] ?? '',
      stepId: json['step_id'] ?? '',
      stepName: json['step_name'] ?? 'Unnamed Task',
      status: json['status'] ?? 'pending',
      assignedTo: json['assigned_to'] ?? 'Unassigned',
      createdAt: DateTime.parse(json['created_at'] ?? DateTime.now().toIso8601String()),
      dueAt: DateTime.parse(json['due_at'] ?? DateTime.now().toIso8601String()),
      data: json['data'],
      result: json['result'],
    );
  }

  String get formattedCreatedAt {
    return DateFormat('MMM dd, HH:mm').format(createdAt);
  }

  String get formattedDueAt {
    return DateFormat('MMM dd, HH:mm').format(dueAt);
  }

  String get timeRemaining {
    final now = DateTime.now();
    final difference = dueAt.difference(now);
    
    if (difference.inDays > 0) {
      return '${difference.inDays}d ${difference.inHours.remainder(24)}h remaining';
    } else if (difference.inHours > 0) {
      return '${difference.inHours}h ${difference.inMinutes.remainder(60)}m remaining';
    } else if (difference.inMinutes > 0) {
      return '${difference.inMinutes}m remaining';
    } else {
      return 'Overdue';
    }
  }

  String get department {
    if (assignedTo.startsWith('role:')) {
      return assignedTo.substring(5).replaceAll('_', ' ').toUpperCase();
    }
    return assignedTo;
  }

  bool get isUrgent {
    final now = DateTime.now();
    final hoursRemaining = dueAt.difference(now).inHours;
    return hoursRemaining < 24;
  }

  Map<String, dynamic> toJson() {
    return {
      'id': id,
      'process_instance_id': processInstanceId,
      'step_id': stepId,
      'step_name': stepName,
      'status': status,
      'assigned_to': assignedTo,
      'created_at': createdAt.toIso8601String(),
      'due_at': dueAt.toIso8601String(),
      'data': data,
      'result': result,
    };
  }
}
