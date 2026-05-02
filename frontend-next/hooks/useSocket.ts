'use client';

import { useEffect, useRef, useState, useCallback } from 'react';
import { toast } from 'sonner';
import { getAccessToken } from '@/lib/auth';

export interface NotificationMessage {
  id: string;
  type: string;
  title: string;
  message: string;
  user_id?: string;
  task_id?: string;
  process_id?: string;
  workflow_id?: string;
  priority: 'low' | 'medium' | 'high' | 'urgent';
  action_url?: string;
  data?: Record<string, any>;
  timestamp: number;
}

export interface SocketMessage {
  type: string;
  user_id?: string;
  data: Record<string, any>;
  timestamp: number;
}

interface UseSocketOptions {
  autoConnect?: boolean;
  reconnectInterval?: number;
  maxReconnectAttempts?: number;
  onNotification?: (notification: NotificationMessage) => void;
  onConnect?: () => void;
  onDisconnect?: () => void;
  onError?: (error: Event) => void;
}

export const useSocket = (options: UseSocketOptions = {}) => {
  const {
    autoConnect = true,
    reconnectInterval = 5000,
    maxReconnectAttempts = 10,
    onNotification,
    onConnect,
    onDisconnect,
    onError,
  } = options;

  const [isConnected, setIsConnected] = useState(false);
  const [connectionStatus, setConnectionStatus] = useState<'disconnected' | 'connecting' | 'connected' | 'error'>('disconnected');
  const [lastNotification, setLastNotification] = useState<NotificationMessage | null>(null);
  
  const ws = useRef<WebSocket | null>(null);
  const reconnectAttempts = useRef(0);
  const reconnectTimer = useRef<NodeJS.Timeout | null>(null);
  const isManualDisconnect = useRef(false);

  // Get WebSocket URL based on environment
  const getWebSocketUrl = useCallback(() => {
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const host = process.env.NODE_ENV === 'production' 
      ? window.location.host 
      : 'localhost:8080'; // Backend port
    return `${protocol}//${host}/ws`;
  }, []);

  // Send message to WebSocket
  const sendMessage = useCallback((message: SocketMessage) => {
    if (ws.current && ws.current.readyState === WebSocket.OPEN) {
      ws.current.send(JSON.stringify(message));
    } else {
      console.warn('WebSocket is not connected. Message not sent:', message);
    }
  }, []);

  // Send ping to keep connection alive
  const sendPing = useCallback(() => {
    sendMessage({
      type: 'ping',
      data: {},
      timestamp: Date.now(),
    });
  }, [sendMessage]);

  // Handle incoming messages
  const handleMessage = useCallback((event: MessageEvent) => {
    try {
      const data = JSON.parse(event.data);
      
      // Handle different message types
      switch (data.type) {
        case 'pong':
          // Heartbeat response
          break;
          
        case 'connection_established':
          console.log('WebSocket connection established');
          break;
          
        case 'task_assigned':
          setLastNotification(data);
          toast.success(data.title, {
            description: data.message,
            action: data.action_url ? {
              label: 'View Task',
              onClick: () => window.location.href = data.action_url,
            } : undefined,
          });
          break;
          
        case 'task_completed':
          setLastNotification(data);
          toast.info(data.title, {
            description: data.message,
          });
          break;
          
        case 'approval_request':
          setLastNotification(data);
          toast.warning(data.title, {
            description: data.message,
            action: data.action_url ? {
              label: 'Review',
              onClick: () => window.location.href = data.action_url,
            } : undefined,
          });
          break;
          
        case 'signature_generated':
          setLastNotification(data);
          toast.success(data.title, {
            description: data.message,
          });
          break;
          
        case 'system_notification':
          setLastNotification(data);
          const toastFn = data.priority === 'high' || data.priority === 'urgent' 
            ? toast.error 
            : data.priority === 'medium' 
            ? toast.warning 
            : toast.info;
          
          toastFn(data.title, {
            description: data.message,
          });
          break;
          
        default:
          console.log('Unknown message type:', data.type, data);
      }
      
      // Call custom notification handler
      if (onNotification) {
        onNotification(data);
      }
      
    } catch (error) {
      console.error('Failed to parse WebSocket message:', error);
    }
  }, [onNotification]);

  // Connect to WebSocket
  const connect = useCallback(() => {
    if (ws.current?.readyState === WebSocket.OPEN) {
      return; // Already connected
    }

    const token = getAccessToken();
    if (!token) {
      console.warn('No authentication token available for WebSocket connection');
      setConnectionStatus('error');
      return;
    }

    try {
      setConnectionStatus('connecting');
      const wsUrl = `${getWebSocketUrl()}?token=${encodeURIComponent(token)}`;
      ws.current = new WebSocket(wsUrl);

      ws.current.onopen = () => {
        console.log('WebSocket connected');
        setIsConnected(true);
        setConnectionStatus('connected');
        reconnectAttempts.current = 0;
        isManualDisconnect.current = false;
        
        // Start heartbeat
        const heartbeatInterval = setInterval(() => {
          if (ws.current?.readyState === WebSocket.OPEN) {
            sendPing();
          } else {
            clearInterval(heartbeatInterval);
          }
        }, 30000); // Send ping every 30 seconds
        
        onConnect?.();
      };

      ws.current.onmessage = handleMessage;

      ws.current.onclose = (event) => {
        console.log('WebSocket disconnected:', event.code, event.reason);
        setIsConnected(false);
        setConnectionStatus('disconnected');
        
        onDisconnect?.();
        
        // Attempt to reconnect if not manually disconnected
        if (!isManualDisconnect.current && reconnectAttempts.current < maxReconnectAttempts) {
          reconnectAttempts.current++;
          console.log(`Attempting to reconnect... (${reconnectAttempts.current}/${maxReconnectAttempts})`);
          
          reconnectTimer.current = setTimeout(() => {
            connect();
          }, reconnectInterval);
        } else if (reconnectAttempts.current >= maxReconnectAttempts) {
          console.error('Max reconnection attempts reached');
          setConnectionStatus('error');
          toast.error('Connection Lost', {
            description: 'Unable to maintain real-time connection. Please refresh the page.',
          });
        }
      };

      ws.current.onerror = (error) => {
        console.error('WebSocket error:', error);
        setConnectionStatus('error');
        onError?.(error);
      };

    } catch (error) {
      console.error('Failed to create WebSocket connection:', error);
      setConnectionStatus('error');
    }
  }, [getWebSocketUrl, handleMessage, maxReconnectAttempts, reconnectInterval, onConnect, onDisconnect, onError, sendPing]);

  // Disconnect from WebSocket
  const disconnect = useCallback(() => {
    isManualDisconnect.current = true;
    
    if (reconnectTimer.current) {
      clearTimeout(reconnectTimer.current);
      reconnectTimer.current = null;
    }
    
    if (ws.current) {
      ws.current.close();
      ws.current = null;
    }
    
    setIsConnected(false);
    setConnectionStatus('disconnected');
  }, []);

  // Subscribe to specific notification types
  const subscribe = useCallback((channels: string[]) => {
    sendMessage({
      type: 'subscribe',
      data: { channels },
      timestamp: Date.now(),
    });
  }, [sendMessage]);

  // Auto-connect on mount
  useEffect(() => {
    if (autoConnect) {
      connect();
    }

    return () => {
      disconnect();
    };
  }, [autoConnect, connect, disconnect]);

  // Cleanup on unmount
  useEffect(() => {
    return () => {
      if (reconnectTimer.current) {
        clearTimeout(reconnectTimer.current);
      }
    };
  }, []);

  return {
    isConnected,
    connectionStatus,
    lastNotification,
    connect,
    disconnect,
    sendMessage,
    subscribe,
    sendPing,
  };
};

export default useSocket;