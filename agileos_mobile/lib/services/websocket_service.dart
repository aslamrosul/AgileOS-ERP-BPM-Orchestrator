import 'dart:async';
import 'dart:convert';
import 'dart:io';
import 'package:flutter/foundation.dart';
import 'package:web_socket_channel/web_socket_channel.dart';

class NotificationMessage {
  final String id;
  final String type;
  final String title;
  final String message;
  final String? userId;
  final String? taskId;
  final String? processId;
  final String? workflowId;
  final String priority;
  final String? actionUrl;
  final Map<String, dynamic>? data;
  final int timestamp;

  NotificationMessage({
    required this.id,
    required this.type,
    required this.title,
    required this.message,
    this.userId,
    this.taskId,
    this.processId,
    this.workflowId,
    required this.priority,
    this.actionUrl,
    this.data,
    required this.timestamp,
  });

  factory NotificationMessage.fromJson(Map<String, dynamic> json) {
    return NotificationMessage(
      id: json['id'] ?? '',
      type: json['type'] ?? '',
      title: json['title'] ?? '',
      message: json['message'] ?? '',
      userId: json['user_id'],
      taskId: json['task_id'],
      processId: json['process_id'],
      workflowId: json['workflow_id'],
      priority: json['priority'] ?? 'low',
      actionUrl: json['action_url'],
      data: json['data'],
      timestamp: json['timestamp'] ?? DateTime.now().millisecondsSinceEpoch,
    );
  }
}

enum ConnectionStatus { disconnected, connecting, connected, error }

class WebSocketService extends ChangeNotifier {
  WebSocketChannel? _channel;
  ConnectionStatus _status = ConnectionStatus.disconnected;
  Timer? _reconnectTimer;
  Timer? _heartbeatTimer;
  int _reconnectAttempts = 0;
  final int _maxReconnectAttempts = 10;
  final Duration _reconnectInterval = const Duration(seconds: 5);
  final Duration _heartbeatInterval = const Duration(seconds: 30);
  
  String? _token;
  final List<NotificationMessage> _notifications = [];
  
  // Streams
  final StreamController<NotificationMessage> _notificationController =
      StreamController<NotificationMessage>.broadcast();
  final StreamController<ConnectionStatus> _statusController =
      StreamController<ConnectionStatus>.broadcast();

  // Getters
  ConnectionStatus get status => _status;
  List<NotificationMessage> get notifications => List.unmodifiable(_notifications);
  Stream<NotificationMessage> get notificationStream => _notificationController.stream;
  Stream<ConnectionStatus> get statusStream => _statusController.stream;
  bool get isConnected => _status == ConnectionStatus.connected;

  // WebSocket URL - use 10.0.2.2 for Android emulator
  String get _webSocketUrl {
    final host = Platform.isAndroid ? '10.0.2.2:8081' : 'localhost:8081';
    return 'ws://$host/ws';
  }

  void setToken(String token) {
    _token = token;
  }

  Future<void> connect() async {
    if (_status == ConnectionStatus.connected || _token == null) {
      return;
    }

    _setStatus(ConnectionStatus.connecting);

    try {
      final uri = Uri.parse('$_webSocketUrl?token=${Uri.encodeComponent(_token!)}');
      _channel = WebSocketChannel.connect(uri);

      // Listen for messages
      _channel!.stream.listen(
        _handleMessage,
        onError: _handleError,
        onDone: _handleDisconnection,
      );

      _setStatus(ConnectionStatus.connected);
      _reconnectAttempts = 0;
      _startHeartbeat();

      if (kDebugMode) {
        print('WebSocket connected successfully');
      }
    } catch (e) {
      if (kDebugMode) {
        print('WebSocket connection error: $e');
      }
      _setStatus(ConnectionStatus.error);
      _scheduleReconnect();
    }
  }

  void disconnect() {
    _reconnectTimer?.cancel();
    _heartbeatTimer?.cancel();
    _channel?.sink.close(status.goingAway);
    _channel = null;
    _setStatus(ConnectionStatus.disconnected);
  }

  void _handleMessage(dynamic message) {
    try {
      final data = jsonDecode(message);
      
      switch (data['type']) {
        case 'pong':
          // Heartbeat response
          break;
        case 'connection_established':
          if (kDebugMode) {
            print('WebSocket connection established');
          }
          break;
        case 'task_assigned':
        case 'task_completed':
        case 'approval_request':
        case 'signature_generated':
        case 'system_notification':
          final notification = NotificationMessage.fromJson(data);
          _addNotification(notification);
          break;
        default:
          if (kDebugMode) {
            print('Unknown message type: ${data['type']}');
          }
      }
    } catch (e) {
      if (kDebugMode) {
        print('Error parsing WebSocket message: $e');
      }
    }
  }

  void _handleError(error) {
    if (kDebugMode) {
      print('WebSocket error: $error');
    }
    _setStatus(ConnectionStatus.error);
    _scheduleReconnect();
  }

  void _handleDisconnection() {
    if (kDebugMode) {
      print('WebSocket disconnected');
    }
    _setStatus(ConnectionStatus.disconnected);
    _heartbeatTimer?.cancel();
    _scheduleReconnect();
  }

  void _setStatus(ConnectionStatus status) {
    _status = status;
    _statusController.add(status);
    notifyListeners();
  }

  void _addNotification(NotificationMessage notification) {
    _notifications.insert(0, notification);
    if (_notifications.length > 50) {
      _notifications.removeLast();
    }
    _notificationController.add(notification);
    notifyListeners();
  }

  void _scheduleReconnect() {
    if (_reconnectAttempts >= _maxReconnectAttempts) {
      if (kDebugMode) {
        print('Max reconnection attempts reached');
      }
      return;
    }

    _reconnectAttempts++;
    _reconnectTimer?.cancel();
    _reconnectTimer = Timer(_reconnectInterval, () {
      if (kDebugMode) {
        print('Attempting to reconnect... ($_reconnectAttempts/$_maxReconnectAttempts)');
      }
      connect();
    });
  }

  void _startHeartbeat() {
    _heartbeatTimer?.cancel();
    _heartbeatTimer = Timer.periodic(_heartbeatInterval, (timer) {
      if (_channel != null && _status == ConnectionStatus.connected) {
        _sendMessage({
          'type': 'ping',
          'data': {},
          'timestamp': DateTime.now().millisecondsSinceEpoch,
        });
      } else {
        timer.cancel();
      }
    });
  }

  void _sendMessage(Map<String, dynamic> message) {
    if (_channel != null && _status == ConnectionStatus.connected) {
      _channel!.sink.add(jsonEncode(message));
    }
  }

  void clearNotifications() {
    _notifications.clear();
    notifyListeners();
  }

  @override
  void dispose() {
    disconnect();
    _notificationController.close();
    _statusController.close();
    super.dispose();
  }
}